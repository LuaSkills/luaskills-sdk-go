package luaskills

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"
)

// embeddedScopePreEntryRejection recognizes only control failures proven to precede native mutation.
// embeddedScopePreEntryRejection 仅识别已证明发生于原生变更前的控制失败。
// Root close/free prepare receipts before mutation; local transport Busy is returned before the native adapter is entered.
// 根关闭／释放在变更前准备回执；本地传输 Busy 在进入原生适配器前返回。
func embeddedScopePreEntryRejection(err error) bool {
	// failure must identify the exact entrypoint, never just share a numeric status with result release.
	// failure 必须标识精确入口，绝不能仅与结果释放共享数字状态。
	var failure *EmbeddedTransportError
	return errors.As(err, &failure) && ((failure.Function == "luaskills_ffi_embedded_request_v1" && failure.Status == EmbeddedNativeCapacityExceeded) || (failure.Function == "transport" && failure.Status == EmbeddedNativeBusy))
}

// phaseValue reads the coordinator checkpoint under the same lock used by public diagnostics.
// phaseValue 在与公开诊断相同的锁下读取协调器检查点。
func (s *EmbeddedRuntimeScope) phaseValue() EmbeddedScopePhase {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.phase
}

// setPhase publishes only a checkpoint already justified by native or callback evidence.
// setPhase 仅发布已经由原生或回调证据证明的检查点。
func (s *EmbeddedRuntimeScope) setPhase(phase EmbeddedScopePhase) {
	s.mu.Lock()
	s.phase = phase
	s.mu.Unlock()
}

// accept validates the exact generated control response and advances its checkpoint; it never sends a command.
// accept 校验精确生成控制响应并推进检查点；绝不发送命令。
func (s *EmbeddedRuntimeScope) accept(pending *embeddedScopeControl) error {
	if _, err := DecodeEmbeddedResponse(pending.response); err != nil {
		return err
	}
	// next and identity come only from the specific generated command response, not a generic identity fallback.
	// next 和 identity 仅来自特定生成命令响应，不使用通用身份兜底。
	var next EmbeddedScopePhase
	var identity string
	switch pending.kind {
	case "runtime_close":
		response, err := DecodeEmbeddedOutputRootRuntimeCloseResponse(pending.response)
		if err != nil {
			return err
		}
		identity, next = response.Result.RuntimeId, EmbeddedScopeDrainingCallbacks
	case "runtime_status":
		response, err := DecodeEmbeddedOutputRootRuntimeStatusResponse(pending.response)
		if err != nil {
			return err
		}
		if response.Result.Initialization == EmbeddedOutputInitializationPhaseFaulted {
			return fmt.Errorf("faulted initialization prevents proving safe runtime release")
		}
		identity, next = response.Result.RuntimeId, EmbeddedScopeDrainingRuntime
		if response.Result.Closed {
			next = EmbeddedScopeReleasingRuntime
		}
	case "runtime_free":
		response, err := DecodeEmbeddedOutputRootRuntimeFreeResponse(pending.response)
		if err != nil {
			return err
		}
		identity, next = response.Result.RuntimeId, EmbeddedScopeReleased
	default:
		return fmt.Errorf("unknown runtime scope control command")
	}
	if identity != s.runtime.identity {
		return fmt.Errorf("runtime scope control response changed its exact slot identity")
	}
	s.mu.Lock()
	s.phase, pending.consumed = next, true
	s.mu.Unlock()
	return nil
}

// control executes one prevalidated frame or consumes its retained original response; acknowledged mutations never replay.
// control 执行一个预校验帧或消费其保留原始响应；已确认变更绝不重放。
func (s *EmbeddedRuntimeScope) control(kind string) error {
	s.mu.Lock()
	// pending is published before adapter entry so an unexpected exit retains the exact uncertain control owner.
	// pending 在适配器进入前发布，使意外退出保留精确不确定控制所有者。
	pending := s.pending
	fresh := pending == nil
	if fresh {
		pending = &embeddedScopeControl{kind: kind}
		s.pending = pending
	}
	s.mu.Unlock()
	if pending.kind != kind {
		return fmt.Errorf("runtime scope checkpoint does not match its retained command")
	}
	// failure preserves the original transport outcome even if copied success can advance the checkpoint.
	// failure 保留原始传输结果，即使复制成功能够推进检查点。
	var failure error
	if fresh {
		response, err := s.transport.requestBytes(s.frames[kind])
		s.mu.Lock()
		pending.response, pending.returned = bytes.Clone(response), true
		s.mu.Unlock()
		failure = err
	}
	// Native release failure is handled before generic error inspection to prevent nested statuses authorizing replay.
	// 在通用错误检查前处理原生释放失败，防止嵌套状态授权重放。
	var release *EmbeddedResultReleaseError
	if errors.As(failure, &release) {
		s.mu.Lock()
		s.needsRelease = true
		s.mu.Unlock()
		_ = s.accept(pending)
		return failure
	}
	if failure == nil {
		if !pending.returned || pending.response == nil {
			failure = fmt.Errorf("original runtime control delivery is unavailable; replay is forbidden")
		} else {
			failure = s.accept(pending)
		}
	}
	// Only a consumed response or a proven business/pre-entry rejection releases this control checkpoint.
	// 仅已消费响应或已证明业务／入口前拒绝释放此控制检查点。
	var business *EmbeddedRuntimeError
	if failure == nil || errors.As(failure, &business) || embeddedScopePreEntryRejection(failure) {
		s.mu.Lock()
		s.pending = nil
		s.mu.Unlock()
	}
	return failure
}

// drainCallbacks closes the adopted pump and observes real coordinator exit without concealing recovery failures.
// drainCallbacks 关闭已接管泵并观察实际协调器退出，不掩盖恢复失败。
// retry permits only its explicit acknowledgement recovery; observer cancellation never interrupts this owned routine.
// retry 仅允许其显式确认恢复；观察者取消绝不中断此拥有型过程。
func (s *EmbeddedRuntimeScope) drainCallbacks(retry bool) error {
	if s.pump == nil {
		return nil
	}
	if status := s.pump.Status(); retry && !status.Closed && status.RecoveryRequired {
		if err := s.pump.RetryAcknowledgements(context.Background()); err != nil {
			return err
		}
	}
	s.pump.RequestClose()
	// ticker is owned by this attempt and released even when callback recovery interrupts the drain.
	// ticker 由本次尝试拥有，即使回调恢复中断排空也会释放。
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		// An exited coordinator cannot perform recovery, even if it retains an earlier failed acknowledgement.
		// 已退出协调器无法执行恢复，即使它仍保留较早的失败确认。
		select {
		case <-s.pump.ended:
			return s.pump.Close(context.Background())
		default:
		}
		if s.pump.Status().RecoveryRequired {
			return &EmbeddedRuntimeError{"busy", "callback pump requires explicit delivery recovery before runtime release"}
		}
		select {
		case <-s.pump.ended:
			return s.pump.Close(context.Background())
		case <-ticker.C:
		}
	}
}

// drain resumes acknowledged checkpoints and returns only after actual slot removal, or a retained explicit failure.
// drain 恢复已确认检查点，仅在实际槽移除或保留显式失败时返回。
func (s *EmbeddedRuntimeScope) drain(retry bool) error {
	s.mu.Lock()
	// release is an allocation checkpoint, separate from whether its native mutation already succeeded.
	// release 是分配检查点，与其原生变更是否已经成功分开。
	release := s.needsRelease
	s.mu.Unlock()
	if release {
		if err := s.transport.ReleaseResults(); err != nil {
			return err
		}
		s.mu.Lock()
		s.needsRelease = false
		if s.pending != nil && s.pending.consumed {
			s.pending = nil
		}
		s.mu.Unlock()
	}
	if s.phaseValue() == EmbeddedScopeClosingRuntime {
		if err := s.control("runtime_close"); err != nil {
			return err
		}
	}
	if s.phaseValue() == EmbeddedScopeDrainingCallbacks {
		if err := s.drainCallbacks(retry); err != nil {
			return err
		}
		s.setPhase(EmbeddedScopeDrainingRuntime)
	}
	for s.phaseValue() == EmbeddedScopeDrainingRuntime {
		if err := s.control("runtime_status"); err != nil {
			return err
		}
		if s.phaseValue() == EmbeddedScopeDrainingRuntime {
			if err := waitEmbeddedPoll(context.Background(), s.interval); err != nil {
				return err
			}
		}
	}
	for s.phaseValue() == EmbeddedScopeReleasingRuntime {
		if err := s.control("runtime_free"); err != nil {
			// Core Busy rejects removal before mutation while leases remain; this is the sole automatic mutation retry.
			// 核心 Busy 在租约仍存在时于变更前拒绝移除；这是唯一自动变更重试。
			business, rejected := err.(*EmbeddedRuntimeError)
			if !rejected || business.Code != "busy" {
				return err
			}
			if err := waitEmbeddedPoll(context.Background(), s.interval); err != nil {
				return err
			}
		}
	}
	return nil
}

// run preowns every close/recovery attempt on one persistent coordinator and preserves uncertain ownership on panic/Goexit.
// run 在一个持久协调器上预先拥有每次关闭／恢复尝试，并在 panic／Goexit 时保留不确定所有权。
func (s *EmbeddedRuntimeScope) run() {
	// completed is set only after exact native removal and transport claim release both succeed.
	// completed 仅在精确原生移除及传输声明释放均成功后设置。
	completed := false
	defer func() {
		panicValue := recover()
		if !completed {
			s.mu.Lock()
			s.fatal = &EmbeddedDriverFailure{Message: fmt.Sprintf("runtime scope coordinator exited unexpectedly (%T); ownership remains retained", panicValue)}
			s.failure, s.retryable = s.fatal, false
			if s.running {
				s.attempt.err = s.fatal
				s.running = false
				close(s.attempt.done)
			}
			s.mu.Unlock()
		}
		close(s.ended)
	}()
	for range s.wake {
		s.mu.Lock()
		// attempt is immutable after publication; observer cancellation cannot replace a running attempt.
		// attempt 发布后不可变；观察者取消不能替换正在执行的尝试。
		attempt := s.attempt
		s.mu.Unlock()
		err := s.drain(attempt.retry)
		if err == nil {
			if s.phaseValue() != EmbeddedScopeReleased {
				panic("runtime scope drain returned without slot removal")
			}
			if err := s.transport.releaseRuntimeScope(s); err != nil {
				panic(err)
			}
			embeddedRuntimeScopes.Delete(s)
			s.mu.Lock()
			s.phase, s.running, s.failure, s.retryable = EmbeddedScopeClosed, false, nil, false
			close(attempt.done)
			s.mu.Unlock()
			completed = true
			return
		}
		// Callback diagnostics are read outside the scope lock; the pump owns its independent failure publication.
		// 回调诊断在作用域锁外读取；泵独立拥有其失败发布。
		callbackRecovery := s.pump != nil && s.pump.Status().RecoveryRequired
		if s.pump != nil {
			select {
			case <-s.pump.ended:
				callbackRecovery = false
			default:
			}
		}
		s.mu.Lock()
		s.failure, attempt.err = cloneEmbeddedFailure(err), cloneEmbeddedFailure(err)
		s.retryable = s.needsRelease || embeddedScopePreEntryRejection(err) || (s.phase == EmbeddedScopeDrainingCallbacks && callbackRecovery)
		s.running = false
		close(attempt.done)
		s.mu.Unlock()
	}
}
