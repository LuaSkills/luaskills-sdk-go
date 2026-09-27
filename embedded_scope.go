package luaskills

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EmbeddedScopePhase records acknowledged shutdown checkpoints, independent of observer context lifetimes.
// EmbeddedScopePhase 记录已确认关闭检查点，独立于观察者上下文寿命。
type EmbeddedScopePhase string

const (
	// EmbeddedScopeOpen owns final release while ordinary runtime admission remains available.
	// EmbeddedScopeOpen 拥有最终释放，同时普通运行时入场保持可用。
	EmbeddedScopeOpen EmbeddedScopePhase = "open"
	// EmbeddedScopeClosingRuntime requests native admission closure exactly once per proven attempt.
	// EmbeddedScopeClosingRuntime 对每次已证明尝试精确请求一次原生入场关闭。
	EmbeddedScopeClosingRuntime EmbeddedScopePhase = "closing_runtime"
	// EmbeddedScopeDrainingCallbacks retains handlers and acknowledgements until the adopted pump exits.
	// EmbeddedScopeDrainingCallbacks 保留处理器及确认，直到已接管泵退出。
	EmbeddedScopeDrainingCallbacks EmbeddedScopePhase = "draining_callbacks"
	// EmbeddedScopeDrainingRuntime polls actual core closure without borrowing ordinary driver quota.
	// EmbeddedScopeDrainingRuntime 轮询实际核心关闭，不借用普通驱动器配额。
	EmbeddedScopeDrainingRuntime EmbeddedScopePhase = "draining_runtime"
	// EmbeddedScopeReleasingRuntime removes the exact closed slot after all native leases have drained.
	// EmbeddedScopeReleasingRuntime 在全部原生租约排空后移除精确已关闭槽。
	EmbeddedScopeReleasingRuntime EmbeddedScopePhase = "releasing_runtime"
	// EmbeddedScopeReleased proves native removal but may still own a response allocation or SDK claim.
	// EmbeddedScopeReleased 证明原生移除，但仍可能拥有响应分配或 SDK 声明。
	EmbeddedScopeReleased EmbeddedScopePhase = "released"
	// EmbeddedScopeClosed proves slot and claim release; successful Close also joins the coordinator.
	// EmbeddedScopeClosed 证明槽及声明释放；成功 Close 还会汇合协调器。
	EmbeddedScopeClosed EmbeddedScopePhase = "closed"
)

// EmbeddedScopeStatus is an independent diagnostic snapshot and makes no native query.
// EmbeddedScopeStatus 是独立诊断快照，不进行原生查询。
type EmbeddedScopeStatus struct {
	// RuntimeID and Phase identify the adopted slot and its last proven checkpoint.
	// RuntimeID 和 Phase 标识已接管槽及最后已证明检查点。
	RuntimeID string
	Phase     EmbeddedScopePhase
	// Running distinguishes an active close attempt from an explicitly retained failure.
	// Running 区分活动关闭尝试与显式保留失败。
	Running bool
	// Retryable permits explicit recovery using retained evidence, never an implicit mutation replay.
	// Retryable 允许使用保留证据显式恢复，绝不表示隐式变更重放。
	Retryable bool
	// NeedsResultRelease records a retained native allocation until explicit release succeeds.
	// NeedsResultRelease 记录保留原生分配，直到显式释放成功。
	NeedsResultRelease bool
	// PendingCommand names the original unconsumed control command, or is empty when none remains.
	// PendingCommand 指明尚未消费的原始控制命令；不存在时为空。
	PendingCommand string
	// Failure is a private copy of the latest attempt's error; nil does not imply native closure.
	// Failure 是最近尝试错误的私有副本；为空不表示原生关闭。
	Failure error
}

// embeddedScopeAttempt retains one cancellation-independent shutdown observation.
// embeddedScopeAttempt 保留一次独立于取消的关闭观察。
type embeddedScopeAttempt struct {
	// done closes only after error publication or successful final cleanup; err is protected by scope.mu.
	// done 仅在错误发布或最终成功清理后关闭；err 由 scope.mu 保护。
	done chan struct{}
	err  error
	// retry requests recovery on the same persistent coordinator instead of creating another worker.
	// retry 在同一持久协调器请求恢复，不创建另一个工作位。
	retry bool
}

// embeddedScopeControl retains the original response and its applied checkpoint across release failure.
// embeddedScopeControl 跨释放失败保留原始响应及已应用检查点。
type embeddedScopeControl struct {
	// kind selects one of the three prevalidated root-control frames.
	// kind 选择三个预校验根控制帧之一。
	kind string
	// response never comes from a retried mutation; returned proves adapter unwind, consumed proves typed application.
	// response 绝不来自重试变更；returned 证明适配器返回，consumed 证明类型化应用。
	response           []byte
	returned, consumed bool
}

// EmbeddedRuntimeScope exclusively owns one exact runtime's ordered shutdown and optional existing callback pump.
// EmbeddedRuntimeScope 独占拥有一个精确运行时及可选现有回调泵的有序关闭。
// It borrows the driver and transport; callers must explicitly close them after every scope using them has drained.
// 它借用驱动器及传输；所有使用它们的作用域排空后，调用方必须显式关闭它们。
// Construct with NewEmbeddedRuntimeScope and never copy it; its zero value does not own a runtime.
// 使用 NewEmbeddedRuntimeScope 构造且绝不可复制；其零值不拥有运行时。
type EmbeddedRuntimeScope struct {
	// runtime, transport, pump, interval and frames are fixed before the coordinator starts.
	// runtime、transport、pump、interval 和 frames 在协调器启动前固定。
	runtime   *EmbeddedRuntime
	transport *EmbeddedTransport
	pump      *EmbeddedCallbackPump
	interval  time.Duration
	frames    map[string][]byte
	// mu protects diagnostics and attempt publication; native calls never hold this lock.
	// mu 保护诊断及尝试发布；原生调用绝不持有此锁。
	mu                               sync.Mutex
	phase                            EmbeddedScopePhase
	attempt                          *embeddedScopeAttempt
	running, retryable, needsRelease bool
	failure, fatal                   error
	pending                          *embeddedScopeControl
	// wake coalesces explicit close/recovery requests; ended proves actual coordinator exit.
	// wake 合并显式关闭／恢复请求；ended 证明实际协调器退出。
	wake  chan struct{}
	ended chan struct{}
}

// embeddedRuntimeScopes retains every active or failed scope until exact native and SDK ownership release.
// embeddedRuntimeScopes 保留每个活动或失败作用域，直到精确原生及 SDK 所有权释放。
var embeddedRuntimeScopes sync.Map

// NewEmbeddedRuntimeScope adopts runtime and its exact existing pump with a positive polling interval.
// NewEmbeddedRuntimeScope 以正轮询 interval 接管 runtime 及其精确现有 pump。
// Use EmbeddedDefaultPollInterval for the SDK default. Adoption requires settled driver commands and a quiescent transport.
// 使用 EmbeddedDefaultPollInterval 选择 SDK 默认值。接管要求驱动命令已完成且传输静止。
// Return a scope or a pre-adoption error; construction neither initializes nor closes the native runtime.
// 返回作用域或接管前错误；构造不初始化或关闭原生运行时。
func NewEmbeddedRuntimeScope(runtime *EmbeddedRuntime, pump *EmbeddedCallbackPump, interval time.Duration) (*EmbeddedRuntimeScope, error) {
	if runtime == nil || runtime.client == nil || runtime.client.driver == nil || runtime.identity == "" {
		return nil, fmt.Errorf("runtime scope requires a bound runtime handle")
	}
	if interval <= 0 {
		return nil, fmt.Errorf("runtime scope polling interval must be positive")
	}
	// transport comes from the typed handle's single authority; textual runtime identity alone is insufficient.
	// transport 来自类型句柄的唯一权威；仅文本运行时身份不足以确定归属。
	transport := runtime.client.driver.transport
	if pump != nil && (pump.transport != transport || pump.runtimeID != runtime.identity) {
		return nil, fmt.Errorf("callback pump does not belong to the exact runtime transport")
	}
	scope := &EmbeddedRuntimeScope{runtime: runtime, transport: transport, pump: pump, interval: interval, phase: EmbeddedScopeOpen, frames: make(map[string][]byte), wake: make(chan struct{}, 1), ended: make(chan struct{})}
	// Validate every future cleanup frame before claiming a lifetime that an undersized request budget cannot close.
	// 在接管寿命前校验全部未来清理帧，避免请求预算过小导致无法关闭。
	commands := map[string]EmbeddedInputCommand{
		"runtime_close":  EmbeddedInputCommandRuntimeClose{Type: EmbeddedInputCommandRuntimeCloseTypeRuntimeClose, RuntimeId: runtime.identity},
		"runtime_status": EmbeddedInputCommandRuntimeStatus{Type: EmbeddedInputCommandRuntimeStatusTypeRuntimeStatus, RuntimeId: runtime.identity},
		"runtime_free":   EmbeddedInputCommandRuntimeFree{Type: EmbeddedInputCommandRuntimeFreeTypeRuntimeFree, RuntimeId: runtime.identity},
	}
	for kind, command := range commands {
		frame, err := EncodeEmbeddedRequest(EmbeddedInputRequest{ProtocolVersion: EmbeddedProtocolVersion, Command: command}, transport.Config().MaxRequestBytes)
		if err != nil {
			return nil, err
		}
		scope.frames[kind] = frame
	}
	if err := transport.claimRuntimeScope(scope); err != nil {
		return nil, err
	}
	embeddedRuntimeScopes.Store(scope, struct{}{})
	go scope.run()
	return scope, nil
}

// LiveEmbeddedRuntimeScopes returns discoverable owners, including scopes stalled on missing native evidence.
// LiveEmbeddedRuntimeScopes 返回可发现所有者，包含因缺少原生证据而停滞的作用域。
func LiveEmbeddedRuntimeScopes() []*EmbeddedRuntimeScope {
	// scopes is an independent slice; individual scopes remain exact noncopyable lifecycle owners.
	// scopes 是独立切片；各作用域仍是精确且不可复制的生命周期所有者。
	var scopes []*EmbeddedRuntimeScope
	embeddedRuntimeScopes.Range(func(key, _ any) bool { scopes = append(scopes, key.(*EmbeddedRuntimeScope)); return true })
	return scopes
}

// Runtime returns the adopted handle for ordinary commands; final removal is exclusively owned by this scope.
// Runtime 返回已接管句柄用于普通命令；最终移除由此作用域独占拥有。
func (s *EmbeddedRuntimeScope) Runtime() *EmbeddedRuntime { return s.runtime }

// Status returns current SDK checkpoints and copied failure evidence without waiting for native execution.
// Status 返回当前 SDK 检查点及复制失败证据，不等待原生执行。
func (s *EmbeddedRuntimeScope) Status() EmbeddedScopeStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	// status cannot rewrite retained state or an error observed by another caller.
	// status 不能改写保留状态或其他调用方观察到的错误。
	status := EmbeddedScopeStatus{RuntimeID: s.runtime.identity, Phase: s.phase, Running: s.running, Retryable: s.retryable, NeedsResultRelease: s.needsRelease, Failure: cloneEmbeddedFailure(s.failure)}
	if s.pending != nil {
		status.PendingCommand = s.pending.kind
	}
	return status
}

// request publishes one close attempt, or explicitly resumes a retryable failed attempt on the same coordinator.
// request 发布一次关闭尝试，或在同一协调器显式恢复可重试的失败尝试。
// Repeated ordinary close returns the original observation and never resubmits a failed mutation.
// 重复普通关闭返回原观察，绝不重新提交失败变更。
func (s *EmbeddedRuntimeScope) request(retry bool) (*embeddedScopeAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.attempt != nil && (!retry || s.running || s.phase == EmbeddedScopeClosed) {
		return s.attempt, nil
	}
	if s.attempt != nil && (!s.retryable || s.fatal != nil) {
		return nil, fmt.Errorf("runtime scope failure has no safe retry evidence: %v", s.failure)
	}
	s.attempt = &embeddedScopeAttempt{done: make(chan struct{}), retry: retry}
	s.running, s.retryable, s.failure = true, false, nil
	if s.phase == EmbeddedScopeOpen {
		s.phase = EmbeddedScopeClosingRuntime
	}
	select {
	case s.wake <- struct{}{}:
	default:
	}
	return s.attempt, nil
}

// RequestClose starts owned cleanup without waiting; native or callback failures remain observable through Close and Status.
// RequestClose 启动拥有型清理而不等待；原生或回调失败仍可经 Close 和 Status 观察。
func (s *EmbeddedRuntimeScope) RequestClose() { _, _ = s.request(false) }

// Close starts cleanup even for a cancelled ctx, then observes actual completion and coordinator exit.
// Close 即使 ctx 已取消也启动清理，随后观察实际完成及协调器退出。
// A callback-marked or nil context is rejected before starting; observer timeout leaves the same cleanup owner alive.
// 带回调标记或空上下文在启动前被拒绝；观察超时使同一清理所有者继续存活。
func (s *EmbeddedRuntimeScope) Close(ctx context.Context) error {
	return s.observe(ctx, false)
}

// RetryClose explicitly recovers retained buffers or proven pre-dispatch rejection using ctx only for observation.
// RetryClose 显式恢复保留缓冲或已证明分发前拒绝，ctx 仅用于观察。
// It never replays an acknowledged mutation or invents success from missing delivery.
// 绝不重放已确认变更，也不根据缺失交付捏造成功。
func (s *EmbeddedRuntimeScope) RetryClose(ctx context.Context) error {
	return s.observe(ctx, true)
}

// observe checks callback restrictions, starts or resumes owned work and joins its coordinator after proven success.
// observe 检查回调限制，启动或恢复拥有型工作，并在已证明成功后汇合协调器。
func (s *EmbeddedRuntimeScope) observe(ctx context.Context, retry bool) error {
	if err := checkEmbeddedObserver(ctx); err != nil {
		return err
	}
	// attempt remains retained when this observer exits; all result access follows the scope publication lock.
	// 此观察者退出时 attempt 仍被保留；全部结果读取遵循作用域发布锁。
	attempt, err := s.request(retry)
	if err != nil {
		return err
	}
	if err := waitEmbeddedObservation(ctx, attempt.done); err != nil {
		return err
	}
	s.mu.Lock()
	err = cloneEmbeddedFailure(attempt.err)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	return waitEmbeddedObservation(ctx, s.ended)
}
