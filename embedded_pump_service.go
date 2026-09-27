package luaskills

import (
	"fmt"
	"time"
)

// run owns sequential native execution and only releases its claim after every actual handler worker has joined.
// run 拥有顺序原生执行，仅在每个实际处理器工作位汇合后释放声明。
func (p *EmbeddedCallbackPump) run() {
	completed, workersClosed := false, false
	defer func() {
		panicValue := recover()
		if !completed {
			failure := &EmbeddedDriverFailure{Message: fmt.Sprintf("callback coordinator exited unexpectedly (%T); ownership remains retained", panicValue)}
			p.mu.Lock()
			p.fatal = failure
			if p.failure == nil {
				p.failure = failure
			}
			p.closing = true
			p.finishLocked(p.started, failure)
			p.mu.Unlock()
		}
		if !workersClosed {
			close(p.jobs)
		}
		close(p.ended)
	}()
	frame, err := p.encode(EmbeddedInputCommandRuntimeStatus{Type: EmbeddedInputCommandRuntimeStatusTypeRuntimeStatus, RuntimeId: p.runtimeID})
	if err == nil {
		var response []byte
		response, err = p.native(frame)
		if err == nil {
			var status EmbeddedOutputRootRuntimeStatusResponse
			status, err = DecodeEmbeddedOutputRootRuntimeStatusResponse(response)
			if err == nil && (status.Result.Initialization != EmbeddedOutputInitializationPhaseReady || status.Result.Closing) {
				err = fmt.Errorf("callback pump requires an initialized open runtime")
			}
		}
	}
	if err != nil {
		p.fail(err)
	}
	p.mu.Lock()
	if err == nil && p.closing {
		err = &EmbeddedRuntimeError{"closed", "callback pump closed before becoming ready"}
	}
	p.ready = err == nil
	p.finishLocked(p.started, err)
	p.mu.Unlock()
	ticker := time.NewTicker(time.Duration(p.config.PollIntervalMS) * time.Millisecond)
	defer ticker.Stop()
	for {
		p.mu.Lock()
		retry := p.retry
		p.mu.Unlock()
		if retry != nil {
			err := p.recoverDelivery()
			if err != nil {
				p.fail(err)
			}
			p.mu.Lock()
			p.finishLocked(retry, err)
			p.retry = nil
			p.mu.Unlock()
		}
		if !p.nativePaused() {
			if err := p.step(); err != nil {
				p.fail(err)
			}
		}
		p.mu.Lock()
		finished := p.closing && len(p.publications) == 0 && len(p.registrations) == 0 && len(p.requests) == 0 && p.pending == nil && !p.needsRelease && p.retry == nil
		p.mu.Unlock()
		if finished {
			close(p.jobs)
			workersClosed = true
			p.workers.Wait()
			// Keep public recovery fenced across the final short transport ownership transition.
			// 在最终短时传输所有权转换期间持续封闭公开恢复。
			p.mu.Lock()
			if p.retry != nil {
				p.finishLocked(p.retry, &EmbeddedRuntimeError{"closed", "callback pump already drained"})
				p.retry = nil
			}
			if err := p.transport.releaseCallbackPump(p); err != nil {
				p.mu.Unlock()
				panic(err)
			}
			p.closed = true
			p.mu.Unlock()
			embeddedCallbackPumps.Delete(p)
			completed = true
			return
		}
		select {
		case <-p.wake:
		case <-ticker.C:
		}
	}
}

// nativePaused reports unresolved native ownership that only explicit recovery may advance.
// nativePaused 报告只有显式恢复才能推进的未解决原生所有权。
func (p *EmbeddedCallbackPump) nativePaused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.needsRelease || p.pending != nil
}

// requestSnapshot copies exact request objects; mutable fields remain protected by p.mu.
// requestSnapshot 复制精确请求对象；可变字段仍受 p.mu 保护。
func (p *EmbeddedCallbackPump) requestSnapshot() []*embeddedPumpRequest {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]*embeddedPumpRequest, 0, len(p.requests))
	for _, request := range p.requests {
		result = append(result, request)
	}
	return result
}

// registrationSnapshot copies exact handler owners for coordinator-only native mutation.
// registrationSnapshot 复制精确处理器所有者，供仅由协调器执行的原生变更使用。
func (p *EmbeddedCallbackPump) registrationSnapshot() []*embeddedPumpRegistration {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]*embeddedPumpRegistration, 0, len(p.registrations))
	for _, registration := range p.registrations {
		result = append(result, registration)
	}
	return result
}

// step advances one publication, retirement, cancellation, acknowledgement and bounded extraction pass without blocking on handlers.
// step 推进一次发布、退役、取消、确认及有界领取，不阻塞等待处理器。
func (p *EmbeddedCallbackPump) step() error {
	p.mu.Lock()
	var publication *embeddedPumpPublication
	for _, candidate := range p.publications {
		if !candidate.attempted {
			publication = candidate
			break
		}
	}
	p.mu.Unlock()
	if publication != nil {
		if err := p.attempt(&embeddedPumpMutation{kind: "register", frame: publication.frame, publication: publication}); err != nil {
			return err
		}
		if p.nativePaused() {
			return nil
		}
	}
	for _, registration := range p.registrationSnapshot() {
		p.mu.Lock()
		if p.closing {
			registration.retiring = true
		}
		retire := registration.retiring && !registration.retired
		p.mu.Unlock()
		if retire {
			frame, err := p.runtimeFrame(EmbeddedInputRuntimeCommandCapabilityUnregister{Type: EmbeddedInputRuntimeCommandCapabilityUnregisterTypeCapabilityUnregister, RegistrationId: registration.identity})
			if err != nil {
				return err
			}
			if err := p.attempt(&embeddedPumpMutation{kind: "unregister", frame: frame, registration: registration}); err != nil {
				return err
			}
			if p.nativePaused() {
				return nil
			}
		}
	}
	for _, request := range p.requestSnapshot() {
		p.mu.Lock()
		finished, failed, encodeError, frame := request.finished, request.ackFailed, request.encodeError, request.frame
		p.mu.Unlock()
		if !finished {
			response, err := p.read(EmbeddedInputRuntimeCommandHostRequestStatus{Type: EmbeddedInputRuntimeCommandHostRequestStatusTypeHostRequestStatus, RequestId: request.request.RequestId})
			if err != nil {
				return err
			}
			status, err := DecodeEmbeddedOutputRuntimeHostRequestStatusResponse(response)
			if err != nil {
				return err
			}
			if status.Result.Cancellation != nil && request.context != nil {
				request.context.observeCancellation(*status.Result.Cancellation)
			}
		} else if encodeError == nil && !failed {
			if err := p.attempt(&embeddedPumpMutation{kind: "complete", frame: frame, request: request}); err != nil {
				return err
			}
		}
		if p.nativePaused() {
			return nil
		}
	}
	for _, registration := range p.registrationSnapshot() {
		p.mu.Lock()
		retired, active := registration.retired, false
		for _, request := range p.requests {
			active = active || request.registration == registration
		}
		p.mu.Unlock()
		if !retired || active {
			continue
		}
		response, err := p.read(EmbeddedInputRuntimeCommandCapabilityStatus{Type: EmbeddedInputRuntimeCommandCapabilityStatusTypeCapabilityStatus, RegistrationId: registration.identity})
		if err != nil {
			return err
		}
		status, err := DecodeEmbeddedOutputRuntimeCapabilityStatusResponse(response)
		if err != nil {
			return err
		}
		if p.nativePaused() {
			return nil
		}
		if status.Result.Drained {
			frame, err := p.runtimeFrame(EmbeddedInputRuntimeCommandCapabilityForget{Type: EmbeddedInputRuntimeCommandCapabilityForgetTypeCapabilityForget, RegistrationId: registration.identity})
			if err != nil {
				return err
			}
			if err := p.attempt(&embeddedPumpMutation{kind: "forget", frame: frame, registration: registration}); err != nil {
				return err
			}
			if p.nativePaused() {
				return nil
			}
		}
	}
	p.mu.Lock()
	closing := p.closing
	available := p.config.MaxConcurrentHandlers - uint64(len(p.requests))
	p.mu.Unlock()
	if !closing && available != 0 {
		frame, err := p.runtimeFrame(EmbeddedInputRuntimeCommandHostRequestsTake{Type: EmbeddedInputRuntimeCommandHostRequestsTakeTypeHostRequestsTake, Limit: available})
		if err != nil {
			return err
		}
		return p.attempt(&embeddedPumpMutation{kind: "take", frame: frame})
	}
	return nil
}

// handlerWorker executes fixed handler capacity and fences new admission if application Goexit removes a worker.
// handlerWorker 执行固定处理器容量；应用 Goexit 移除工作位时封闭新入场。
func (p *EmbeddedCallbackPump) handlerWorker() {
	normalExit := false
	defer func() {
		panicValue := recover()
		if !normalExit {
			p.fail(&EmbeddedDriverFailure{Message: fmt.Sprintf("callback handler worker exited unexpectedly (%T)", panicValue)})
		}
		p.workers.Done()
	}()
	for request := range p.jobs {
		p.handle(request)
	}
	normalExit = true
}

// handle seals effects after all actual handler defers unwind, then freezes its completion before publishing readiness.
// handle 在实际处理器全部延迟清理完成后封存副作用，再冻结完成结果并发布就绪。
func (p *EmbeddedCallbackPump) handle(request *embeddedPumpRequest) {
	var value any
	var err error
	returned := false
	defer func() {
		panicValue := recover()
		if !returned {
			err = fmt.Errorf("host handler exited unexpectedly (%T)", panicValue)
			if panicValue == nil {
				// Goexit will remove this fixed worker; fence extraction before publishing its returned ownership.
				// Goexit 将移除此固定工作位；发布已返回所有权前先封闭领取。
				p.fail(err)
			}
		}
		effects := EmbeddedInputEffectStateNotStarted
		if request.context != nil {
			effects = request.context.seal()
		}
		var outcome EmbeddedInputHostCompletion
		if err == nil {
			outcome = EmbeddedInputHostCompletionVariant1{Ok: true, Value: value, Effects: effects}
		} else {
			outcome = EmbeddedInputHostCompletionVariant2{Ok: false, Error: embeddedCallbackFailure(err), Effects: effects}
		}
		frame, encodeError := p.completionFrame(request, outcome)
		if encodeError != nil {
			outcome = EmbeddedInputHostCompletionVariant2{Ok: false, Error: EmbeddedInputEmbeddedError{Code: "execution_failed", Message: "Go host callback produced an invalid or oversized result"}, Effects: effects}
			frame, encodeError = p.completionFrame(request, outcome)
		}
		p.mu.Lock()
		request.frame, request.encodeError, request.finished, request.effects = frame, cloneEmbeddedFailure(encodeError), true, effects
		request.request.Arguments = nil
		p.mu.Unlock()
		if encodeError != nil {
			p.fail(encodeError)
		}
		p.signal()
	}()
	if request.registration == nil {
		err = &EmbeddedRuntimeError{Code: "internal", Message: "Go callback registration owner is absent"}
		p.fail(err)
	} else {
		value, err = request.registration.capability.Handler(request.request.Arguments, request.context)
	}
	returned = true
}

// completionFrame encodes an exact completion envelope; arbitrary host serializers and mutable response aliases never enter native code.
// completionFrame 编码精确完成信封；任意宿主序列化器及可变响应别名绝不进入原生代码。
func (p *EmbeddedCallbackPump) completionFrame(request *embeddedPumpRequest, outcome EmbeddedInputHostCompletion) ([]byte, error) {
	return p.runtimeFrame(EmbeddedInputRuntimeCommandHostRequestComplete{Type: EmbeddedInputRuntimeCommandHostRequestCompleteTypeHostRequestComplete, RequestId: request.request.RequestId, Outcome: outcome})
}
