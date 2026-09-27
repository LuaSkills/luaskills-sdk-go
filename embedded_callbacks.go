package luaskills

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EmbeddedHostHandler receives application arguments separately from authenticated core context and returns actual host delivery.
// EmbeddedHostHandler 分开接收应用参数和经过认证的核心上下文，并返回实际宿主交付。
// The context supports cooperative cancellation; returning or panicking never implies rollback of external effects.
// 上下文支持协作取消；返回或 panic 绝不表示外部副作用已回滚。
type EmbeddedHostHandler func(arguments any, ctx *EmbeddedHostCallbackContext) (any, error)

// EmbeddedHostCapability binds an explicitly queued generated descriptor to one exact retained Go handler.
// EmbeddedHostCapability 将显式队列生成描述符绑定到一个精确保留的 Go 处理器。
type EmbeddedHostCapability struct {
	// Descriptor is copied and validated before native publication; later aliases cannot change the registration.
	// Descriptor 在原生发布前复制及校验；后续别名不能改变注册。
	Descriptor EmbeddedInputCapabilityDescriptor
	// Handler remains retained until native registration and actual handler ownership both drain.
	// Handler 保留到原生注册及实际处理器所有权均排空。
	Handler EmbeddedHostHandler
}

// embeddedCallbackKey marks context propagation without relying on goroutine identities or process-global callback state.
// embeddedCallbackKey 标记上下文传播，不依赖协程身份或进程全局回调状态。
type embeddedCallbackKey struct{}

// EmbeddedHostCallbackContext implements context.Context with core-observed cancellation, trusted caller data and explicit effects.
// EmbeddedHostCallbackContext 实现 context.Context，提供核心观察取消、可信调用方数据和显式副作用。
// Pass this context to downstream calls; SDK waits reject its marker until controlled callback dependencies are supported.
// 将此上下文传给下游调用；受控回调依赖支持前，SDK 等待拒绝其标记。
type EmbeddedHostCallbackContext struct {
	// context and cancel own one cooperative core cancellation publication, never a fabricated local deadline.
	// context 和 cancel 拥有一次协作核心取消发布，绝不伪造本地截止时间。
	context context.Context
	cancel  context.CancelCauseFunc
	// requestID, registrationID and caller derive only from the copied core request.
	// requestID、registrationID 及 caller 仅来自复制的核心请求。
	requestID      string
	registrationID string
	caller         EmbeddedOutputCapabilityCaller
	// started and budget provide advisory milliseconds without overflowing time.Duration multiplication.
	// started 和 budget 提供参考毫秒数，不进行可能溢出的 time.Duration 乘法。
	started time.Time
	budget  uint64
	// mu protects effect sealing and the first cancellation diagnostic across handler goroutines.
	// mu 跨处理器协程保护副作用封存及首次取消诊断。
	mu           sync.Mutex
	effects      EmbeddedInputEffectState
	sealed       bool
	cancellation *EmbeddedRuntimeError
}

// newEmbeddedHostContext copies authenticated request metadata and selects the descriptor's initial unknown or non-mutating evidence.
// newEmbeddedHostContext 复制认证请求元数据，并按声明选择初始未知或非变更证据。
func newEmbeddedHostContext(request EmbeddedOutputHostRequest, effects EmbeddedInputCapabilityEffects) *EmbeddedHostCallbackContext {
	ctx, cancel := context.WithCancelCause(context.Background())
	result := &EmbeddedHostCallbackContext{context: ctx, cancel: cancel, requestID: request.RequestId, registrationID: request.RegistrationId, caller: copyEmbeddedCaller(request.Caller), started: time.Now(), budget: request.RemainingMs, effects: EmbeddedInputEffectStateUnknown}
	if effects == EmbeddedInputCapabilityEffectsReadOnly {
		result.effects = EmbeddedInputEffectStateNotApplicable
	}
	return result
}

// copyEmbeddedCaller copies optional string ownership so host mutation cannot rewrite authenticated context.
// copyEmbeddedCaller 复制可选字符串所有权，使宿主修改不能改写认证上下文。
func copyEmbeddedCaller(caller EmbeddedOutputCapabilityCaller) EmbeddedOutputCapabilityCaller {
	if caller.SessionId != nil {
		value := *caller.SessionId
		caller.SessionId = &value
	}
	if caller.WorkspaceRoot != nil {
		value := *caller.WorkspaceRoot
		caller.WorkspaceRoot = &value
	}
	return caller
}

// Deadline returns no local deadline; only the core can publish authoritative cancellation.
// Deadline 不返回本地截止时间；只有核心能够发布权威取消。
func (c *EmbeddedHostCallbackContext) Deadline() (time.Time, bool) { return time.Time{}, false }

// Done closes on the first core cancellation, not on observer timeout or inferred handler termination.
// Done 在首次核心取消时关闭，不因观察超时或推断处理器终止而关闭。
func (c *EmbeddedHostCallbackContext) Done() <-chan struct{} { return c.context.Done() }

// Err implements context.Context cancellation; Cancellation returns its original core error code and message.
// Err 实现 context.Context 取消；Cancellation 返回原始核心错误码及消息。
func (c *EmbeddedHostCallbackContext) Err() error { return c.context.Err() }

// Value propagates the private callback marker and the owned cancellation cause through derived contexts.
// Value 经派生上下文传播私有回调标记及拥有的取消原因。
func (c *EmbeddedHostCallbackContext) Value(key any) any {
	if _, marked := key.(embeddedCallbackKey); marked {
		return c
	}
	return c.context.Value(key)
}

// RequestID returns the exact never-reused native request identity.
// RequestID 返回精确且不复用的原生请求身份。
func (c *EmbeddedHostCallbackContext) RequestID() string { return c.requestID }

// RegistrationID returns the immutable registration that selected this handler.
// RegistrationID 返回选择此处理器的不可变注册身份。
func (c *EmbeddedHostCallbackContext) RegistrationID() string { return c.registrationID }

// Caller returns an independent copy of trusted caller fields, never application argument authority.
// Caller 返回可信调用方字段的独立副本，绝不使用应用参数作为权威。
func (c *EmbeddedHostCallbackContext) Caller() EmbeddedOutputCapabilityCaller {
	return copyEmbeddedCaller(c.caller)
}

// RemainingMS returns nonnegative advisory milliseconds from the delivered uint64 budget without narrowing it.
// RemainingMS 返回从已交付 uint64 预算派生的非负参考毫秒，不缩窄其位宽。
func (c *EmbeddedHostCallbackContext) RemainingMS() uint64 {
	elapsed := uint64(time.Since(c.started) / time.Millisecond)
	if elapsed >= c.budget {
		return 0
	}
	return c.budget - elapsed
}

// Effects returns the latest explicit host evidence; success and cancellation never synthesize a commit or rollback.
// Effects 返回最新显式宿主证据；成功及取消绝不自动生成提交或回滚。
func (c *EmbeddedHostCallbackContext) Effects() EmbeddedInputEffectState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.effects
}

// ReportEffects validates a generated effect state and records it only while the actual handler is active.
// ReportEffects 校验生成副作用状态，仅在实际处理器活动时记录。
func (c *EmbeddedHostCallbackContext) ReportEffects(effects EmbeddedInputEffectState) error {
	if _, err := projectEmbeddedWire(string(effects), embeddedWireType[EmbeddedInputEffectState](), "effects"); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sealed {
		return fmt.Errorf("embedded callback effect evidence is sealed")
	}
	c.effects = effects
	return nil
}

// Cancellation returns a copied first core diagnostic without claiming that the handler has stopped.
// Cancellation 返回首次核心诊断副本，不宣称处理器已停止。
func (c *EmbeddedHostCallbackContext) Cancellation() *EmbeddedRuntimeError {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancellation == nil {
		return nil
	}
	value := *c.cancellation
	return &value
}

// observeCancellation publishes exactly the first observed core reason to cooperative waiters.
// observeCancellation 向协作等待方精确发布首次观察到的核心原因。
func (c *EmbeddedHostCallbackContext) observeCancellation(reason EmbeddedOutputEmbeddedError) {
	c.mu.Lock()
	if c.cancellation != nil {
		c.mu.Unlock()
		return
	}
	c.cancellation = &EmbeddedRuntimeError{Code: string(reason.Code), Message: reason.Message}
	// The cancellation cause has separate ownership from publicly returned diagnostic copies.
	// 取消原因与公开返回的诊断副本分别拥有。
	cause := &EmbeddedRuntimeError{Code: string(reason.Code), Message: reason.Message}
	c.mu.Unlock()
	c.cancel(cause)
}

// seal returns final explicit effect evidence and prevents later context aliases from changing completion.
// seal 返回最终显式副作用证据，并阻止迟到上下文别名改变完成结果。
func (c *EmbeddedHostCallbackContext) seal() EmbeddedInputEffectState {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sealed = true
	return c.effects
}

// checkEmbeddedObserver rejects missing contexts and callback dependencies before any SDK work is admitted or awaited.
// checkEmbeddedObserver 在任何 SDK 工作入场或等待前拒绝缺失上下文及回调依赖。
func checkEmbeddedObserver(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("embedded observation requires a context")
	}
	if ctx.Value(embeddedCallbackKey{}) != nil {
		return &EmbeddedRuntimeError{Code: "unsupported", Message: "nested embedded SDK calls from a host callback require a controlled dependency protocol"}
	}
	return nil
}

// embeddedCallbackFailure preserves explicit protocol errors and redacts arbitrary host exception text.
// embeddedCallbackFailure 保留显式协议错误，并隐去任意宿主异常文本。
func embeddedCallbackFailure(err error) EmbeddedInputEmbeddedError {
	if failure, ok := err.(*EmbeddedRuntimeError); ok {
		if _, invalid := projectEmbeddedWire(failure.Code, embeddedWireType[EmbeddedInputEmbeddedErrorCode](), "callback_error"); invalid == nil {
			return EmbeddedInputEmbeddedError{Code: EmbeddedInputEmbeddedErrorCode(failure.Code), Message: failure.Message}
		}
	}
	return EmbeddedInputEmbeddedError{Code: "execution_failed", Message: "Go host callback failed"}
}
