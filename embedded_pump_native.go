package luaskills

import (
	"errors"
	"fmt"
)

// encode freezes one exact generated command under this borrowed transport's request-byte authority.
// encode 按借用传输的请求字节权威冻结一个精确生成命令。
func (p *EmbeddedCallbackPump) encode(command EmbeddedInputCommand) ([]byte, error) {
	return EncodeEmbeddedRequest(EmbeddedInputRequest{ProtocolVersion: EmbeddedProtocolVersion, Command: command}, p.transport.Config().MaxRequestBytes)
}

// runtimeFrame wraps an exact generated runtime operation without probing alternate protocol fields.
// runtimeFrame 包装精确生成运行时操作，不探测其他协议字段。
func (p *EmbeddedCallbackPump) runtimeFrame(operation EmbeddedInputRuntimeCommand) ([]byte, error) {
	return p.encode(EmbeddedInputCommandRuntime{Type: EmbeddedInputCommandRuntimeTypeRuntime, RuntimeId: p.runtimeID, Operation: operation})
}

// fail retains the first diagnostic and fences new admission while leaving explicit drainage and recovery alive.
// fail 保留首次诊断并封闭新入场，同时保持显式排空及恢复活动。
func (p *EmbeddedCallbackPump) fail(err error) {
	p.mu.Lock()
	wake := p.failure == nil || !p.closing
	if p.failure == nil {
		p.failure = cloneEmbeddedFailure(err)
	}
	p.closing = true
	p.mu.Unlock()
	if wake {
		p.signal()
	}
}

// native executes only on the sequential coordinator, preserving copied delivery even when native buffer release fails.
// native 仅在顺序协调器执行，即使原生缓冲释放失败也保留复制交付。
func (p *EmbeddedCallbackPump) native(frame []byte) ([]byte, error) {
	p.mu.Lock()
	paused := p.needsRelease
	p.mu.Unlock()
	if paused {
		return nil, &EmbeddedRuntimeError{"busy", "callback result release requires explicit recovery"}
	}
	response, err := p.transport.requestBytes(frame)
	var release *EmbeddedResultReleaseError
	if errors.As(err, &release) {
		p.mu.Lock()
		p.needsRelease = true
		p.mu.Unlock()
		p.fail(err)
		if response == nil {
			return nil, err
		}
	} else if err != nil {
		return response, err
	}
	_, err = DecodeEmbeddedResponse(response)
	return response, err
}

// read performs one status query on the reserved coordinator; callers additionally decode its exact generated response.
// read 在预留协调器执行一次状态查询；调用方进一步解码精确生成响应。
func (p *EmbeddedCallbackPump) read(operation EmbeddedInputRuntimeCommand) ([]byte, error) {
	frame, err := p.runtimeFrame(operation)
	if err != nil {
		return nil, err
	}
	return p.native(frame)
}

// attempt publishes ownership before a mutation enters native code and applies original evidence at most once.
// attempt 在变更进入原生代码前发布所有权，并至多一次应用原始证据。
func (p *EmbeddedCallbackPump) attempt(mutation *embeddedPumpMutation) error {
	p.mu.Lock()
	if p.pending != nil || p.needsRelease {
		p.mu.Unlock()
		return &EmbeddedRuntimeError{"busy", "callback mutation requires explicit recovery"}
	}
	p.pending = mutation
	if mutation.publication != nil {
		mutation.publication.attempted = true
	}
	p.mu.Unlock()
	response, err := p.native(mutation.frame)
	p.mu.Lock()
	mutation.response, mutation.err, mutation.returned = response, cloneEmbeddedFailure(err), true
	p.mu.Unlock()
	if err == nil {
		err = p.applyMutation(mutation)
	}
	if err == nil {
		return nil
	}
	// Only the request parser's rejection proves non-dispatch; release errors must never rewrite actual completion.
	// 只有请求解析器拒绝能证明未分发；释放错误绝不能改写实际完成。
	var parserError *EmbeddedTransportError
	if mutation.kind == "complete" && errors.As(err, &parserError) && parserError.Function == "luaskills_ffi_embedded_request_v1" && parserError.Status == EmbeddedNativeInvalidArgument && !mutation.request.parserRejected {
		request := mutation.request
		outcome := EmbeddedInputHostCompletionVariant2{Ok: false, Error: EmbeddedInputEmbeddedError{Code: "execution_failed", Message: "Go host callback result was rejected by the native parser"}, Effects: request.effects}
		frame, encodeError := p.completionFrame(request, outcome)
		p.mu.Lock()
		request.parserRejected = true
		request.frame, request.encodeError = frame, cloneEmbeddedFailure(encodeError)
		p.pending = nil
		p.mu.Unlock()
		if encodeError != nil {
			p.fail(encodeError)
			return encodeError
		}
		return p.attempt(&embeddedPumpMutation{kind: "complete", frame: frame, request: request})
	}
	var business *EmbeddedRuntimeError
	var transport *EmbeddedTransportError
	provenRejection := errors.As(err, &business) || errors.As(err, &transport) && transport.Function == "luaskills_ffi_embedded_request_v1" && transport.Status == EmbeddedNativeInvalidArgument
	p.mu.Lock()
	if mutation.kind == "register" {
		p.finishLocked(mutation.publication.observation, err)
		if provenRejection {
			p.removePublicationLocked(mutation.publication)
			p.pending = nil
			p.mu.Unlock()
			return nil
		}
	}
	if mutation.kind == "take" && provenRejection {
		p.pending = nil
	}
	if mutation.kind == "complete" {
		mutation.request.ackFailed = true
		p.pending = nil
	}
	p.mu.Unlock()
	p.fail(err)
	return err
}

// removePublicationLocked drops only the exact batch after proven installation or rejection; caller holds p.mu.
// removePublicationLocked 仅在已证明安装或拒绝后丢弃精确批次；调用方持有 p.mu。
func (p *EmbeddedCallbackPump) removePublicationLocked(publication *embeddedPumpPublication) {
	for index, candidate := range p.publications {
		if candidate == publication {
			copy(p.publications[index:], p.publications[index+1:])
			p.publications[len(p.publications)-1] = nil
			p.publications = p.publications[:len(p.publications)-1]
			return
		}
	}
}

// applyMutation validates the exact generated acknowledgement before changing any owner map.
// applyMutation 在改变任何所有者映射前校验精确生成确认。
func (p *EmbeddedCallbackPump) applyMutation(mutation *embeddedPumpMutation) error {
	switch mutation.kind {
	case "register":
		response, err := DecodeEmbeddedOutputRuntimeCapabilitiesRegisterResponse(mutation.response)
		if err != nil {
			return err
		}
		identities := response.Result.RegistrationIds
		p.mu.Lock()
		defer p.mu.Unlock()
		if len(identities) != len(mutation.publication.capabilities) {
			return fmt.Errorf("core callback registration count differs from the frozen batch")
		}
		seen := make(map[string]bool, len(identities))
		for _, id := range identities {
			if id == "" || seen[id] || p.registrations[id] != nil {
				return fmt.Errorf("core callback registration identities are inconsistent")
			}
			seen[id] = true
		}
		for index, id := range identities {
			p.registrations[id] = &embeddedPumpRegistration{identity: id, capability: mutation.publication.capabilities[index], observation: newEmbeddedPumpObservation()}
		}
		mutation.publication.identities = append([]string{}, identities...)
		p.removePublicationLocked(mutation.publication)
		p.finishLocked(mutation.publication.observation, nil)
		p.pending = nil
	case "take":
		response, err := DecodeEmbeddedOutputRuntimeHostRequestsTakeResponse(mutation.response)
		if err != nil {
			return err
		}
		p.mu.Lock()
		if uint64(len(response.Result)) > p.config.MaxConcurrentHandlers-uint64(len(p.requests)) {
			p.mu.Unlock()
			return fmt.Errorf("core callback batch exceeds available handler ownership")
		}
		seen := make(map[string]bool, len(response.Result))
		for _, request := range response.Result {
			if request.RequestId == "" || seen[request.RequestId] || p.requests[request.RequestId] != nil {
				p.mu.Unlock()
				return fmt.Errorf("core callback request identities are inconsistent")
			}
			seen[request.RequestId] = true
		}
		batch := make([]*embeddedPumpRequest, 0, len(response.Result))
		for _, request := range response.Result {
			registration := p.registrations[request.RegistrationId]
			record := &embeddedPumpRequest{request: request, registration: registration}
			if registration != nil {
				record.context = newEmbeddedHostContext(request, registration.capability.Descriptor.Effects)
			}
			p.requests[request.RequestId] = record
			batch = append(batch, record)
		}
		p.pending = nil
		p.mu.Unlock()
		// The whole delivered batch is retained before any handler can execute or fail.
		// 在任何处理器能够执行或失败之前，保留整个已交付批次。
		for _, request := range batch {
			p.jobs <- request
		}
	case "unregister":
		if _, err := DecodeEmbeddedOutputRuntimeCapabilityUnregisterResponse(mutation.response); err != nil {
			return err
		}
		p.mu.Lock()
		mutation.registration.retired = true
		p.pending = nil
		p.mu.Unlock()
	case "forget":
		if _, err := DecodeEmbeddedOutputRuntimeCapabilityForgetResponse(mutation.response); err != nil {
			return err
		}
		p.mu.Lock()
		delete(p.registrations, mutation.registration.identity)
		p.finishLocked(mutation.registration.observation, nil)
		p.pending = nil
		p.mu.Unlock()
	case "complete":
		if _, err := DecodeEmbeddedOutputRuntimeHostRequestCompleteResponse(mutation.response); err != nil {
			return err
		}
		p.mu.Lock()
		delete(p.requests, mutation.request.request.RequestId)
		p.pending = nil
		p.mu.Unlock()
	default:
		return fmt.Errorf("unknown callback mutation owner")
	}
	return nil
}

// recoverDelivery retries buffer release, consumes retained original receipts and reconciles failed acknowledgements by stable IDs.
// recoverDelivery 重试缓冲释放、消费保留原始回执，并按稳定身份核对失败确认。
func (p *EmbeddedCallbackPump) recoverDelivery() error {
	if err := p.transport.ReleaseResults(); err != nil {
		return err
	}
	p.mu.Lock()
	p.needsRelease = false
	pending := p.pending
	p.mu.Unlock()
	if pending != nil {
		if !pending.returned || pending.response == nil {
			return fmt.Errorf("original callback mutation delivery is unavailable; replay is forbidden")
		}
		if _, err := DecodeEmbeddedResponse(pending.response); err != nil {
			return err
		}
		if err := p.applyMutation(pending); err != nil {
			return err
		}
	}
	for _, request := range p.requestSnapshot() {
		p.mu.Lock()
		failed, encodeError := request.ackFailed, request.encodeError
		p.mu.Unlock()
		if encodeError != nil {
			return cloneEmbeddedFailure(encodeError)
		}
		if failed {
			if err := p.reconcileAcknowledgement(request); err != nil {
				return err
			}
		}
		if p.nativePaused() {
			return &EmbeddedRuntimeError{"busy", "callback recovery produced another retained native result"}
		}
	}
	return nil
}

// reconcileAcknowledgement re-delivers only a frozen completion proven still dispatched; handler execution is never repeated.
// reconcileAcknowledgement 仅重新交付已证明仍在分发态的冻结完成；绝不重复处理器执行。
func (p *EmbeddedCallbackPump) reconcileAcknowledgement(request *embeddedPumpRequest) error {
	response, err := p.read(EmbeddedInputRuntimeCommandHostRequestStatus{Type: EmbeddedInputRuntimeCommandHostRequestStatusTypeHostRequestStatus, RequestId: request.request.RequestId})
	completed := false
	if err == nil {
		status, decodeError := DecodeEmbeddedOutputRuntimeHostRequestStatusResponse(response)
		if decodeError != nil {
			return decodeError
		}
		switch status.Result.Phase {
		case EmbeddedOutputHostRequestPhaseCompleted:
			completed = true
		case EmbeddedOutputHostRequestPhaseCompleting:
			return &EmbeddedRuntimeError{"busy", "callback completion is still owned by the core"}
		case EmbeddedOutputHostRequestPhaseDispatched:
		case EmbeddedOutputHostRequestPhaseQueued:
			return fmt.Errorf("owned callback unexpectedly returned to queued state")
		}
	} else {
		var business *EmbeddedRuntimeError
		if !errors.As(err, &business) || business.Code != "not_found" && business.Code != "already_completed" {
			return err
		}
		response, err = p.read(EmbeddedInputRuntimeCommandOperationStatus{Type: EmbeddedInputRuntimeCommandOperationStatusTypeOperationStatus, OperationId: request.request.Caller.OperationId})
		if err != nil {
			return err
		}
		operation, decodeError := DecodeEmbeddedOutputRuntimeOperationStatusResponse(response)
		if decodeError != nil {
			return decodeError
		}
		for _, effect := range operation.Result.HostEffects {
			if effect.RequestId != nil && *effect.RequestId == request.request.RequestId && effect.RegistrationId == request.request.RegistrationId && effect.Phase == EmbeddedOutputHostEffectPhaseCompleted {
				completed = request.request.EffectId == nil || effect.EffectId == *request.request.EffectId
				if completed {
					break
				}
			}
		}
		if !completed {
			return &EmbeddedRuntimeError{"not_found", "callback completion evidence is unavailable; completion cannot be inferred"}
		}
	}
	if completed {
		p.mu.Lock()
		delete(p.requests, request.request.RequestId)
		p.mu.Unlock()
		return nil
	}
	return p.attempt(&embeddedPumpMutation{kind: "complete", frame: request.frame, request: request})
}
