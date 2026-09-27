package luaskills

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// EmbeddedCallbackPumpConfig bounds handler ownership, accepted publications and core polling cadence explicitly.
// EmbeddedCallbackPumpConfig 显式限制处理器所有权、已接纳发布及核心轮询节奏。
type EmbeddedCallbackPumpConfig struct {
	// MaxConcurrentHandlers includes returned handlers whose original completion still awaits native acknowledgement.
	// MaxConcurrentHandlers 包含已经返回但原始完成仍等待原生确认的处理器。
	MaxConcurrentHandlers uint64
	// MaxPendingCommands includes accepted registration batches with uncertain native delivery.
	// MaxPendingCommands 包含原生交付不确定的已接纳注册批次。
	MaxPendingCommands uint64
	// PollIntervalMS governs cancellation and drainage observations, never the authoritative operation deadline.
	// PollIntervalMS 控制取消及排空观察，绝不决定权威操作截止时间。
	PollIntervalMS uint64
}

// EmbeddedCallbackPumpStatus copies actual SDK ownership diagnostics; it does not infer native completion.
// EmbeddedCallbackPumpStatus 复制实际 SDK 所有权诊断，不推断原生完成。
type EmbeddedCallbackPumpStatus struct {
	// RuntimeID identifies the exact borrowed native runtime.
	// RuntimeID 标识精确借用的原生运行时。
	RuntimeID string
	// Ready, Closing and Closed distinguish startup, admission closure and proven complete drainage.
	// Ready、Closing 和 Closed 区分启动、入场关闭及已证明的完整排空。
	Ready, Closing, Closed bool
	// RegistrationIDs and RequestIDs retain exact identities until both native and SDK ownership drain.
	// RegistrationIDs 和 RequestIDs 保留精确身份，直到原生及 SDK 所有权均排空。
	RegistrationIDs, RequestIDs []string
	// PendingAcknowledgements identifies returned handlers still retaining completion ownership.
	// PendingAcknowledgements 标识已返回但仍保留完成所有权的处理器。
	PendingAcknowledgements []string
	// PendingCommands counts retained publication batches, including uncertain results.
	// PendingCommands 统计保留的发布批次，包含不确定结果。
	PendingCommands uint64
	// RecoveryRequired excludes normally executing work and identifies completed unresolved native evidence.
	// RecoveryRequired 排除正常执行中的工作，标识已完成但未解决的原生证据。
	RecoveryRequired bool
	// Failure preserves the first diagnostic even after explicit recovery permits safe closure.
	// Failure 保留首次诊断，即使显式恢复之后已经允许安全关闭。
	Failure error
}

// embeddedPumpObservation survives observer cancellation; pump.mu protects completion and error publication.
// embeddedPumpObservation 跨观察者取消存活；pump.mu 保护完成及错误发布。
type embeddedPumpObservation struct {
	// done and finished distinguish notification ownership from any observer's cancellation.
	// done 和 finished 区分通知所有权及任意观察者取消。
	done     chan struct{}
	finished bool
	// err is a private diagnostic copied again for each observer.
	// err 是私有诊断，每个观察者再次获得副本。
	err error
}

// embeddedPumpPublication retains frozen native bytes and exact handler identities through uncertain registration.
// embeddedPumpPublication 在不确定注册期间保留冻结原生字节和精确处理器身份。
type embeddedPumpPublication struct {
	// capabilities and frame are frozen together before the publication enters the coordinator queue.
	// capabilities 和 frame 在发布进入协调队列前一起冻结。
	capabilities []EmbeddedHostCapability
	frame        []byte
	// attempted prevents replay; identities hold proven native acknowledgement in the original batch order.
	// attempted 阻止重放；identities 按原批次顺序保存已证明原生确认。
	attempted  bool
	identities []string
	// observation remains owned after the registering caller stops waiting.
	// observation 在注册调用方停止等待后仍被拥有。
	observation *embeddedPumpObservation
}

// embeddedPumpRegistration retains one handler until unregister, native drainage and metadata removal are proven.
// embeddedPumpRegistration 保留一个处理器，直到注销、原生排空及元数据移除均得到证明。
type embeddedPumpRegistration struct {
	// identity selects the exact frozen handler rather than its replaceable public capability name.
	// identity 选择精确冻结处理器，而非可被替换的公开能力名称。
	identity   string
	capability EmbeddedHostCapability
	// retiring is host intent; retired proves native admission was closed.
	// retiring 是宿主意图；retired 证明原生入场已经关闭。
	retiring bool
	retired  bool
	// observation completes only after actual references drain and native metadata is forgotten.
	// observation 仅在实际引用排空且原生元数据遗忘后完成。
	observation *embeddedPumpObservation
}

// embeddedPumpRequest owns a delivered identity and its frozen completion independently from the handler's mutable aliases.
// embeddedPumpRequest 独立于处理器可变别名，拥有已交付身份及冻结完成结果。
type embeddedPumpRequest struct {
	// request, registration and context bind one delivered identity to trusted metadata and its exact handler.
	// request、registration 及 context 将已交付身份绑定到可信元数据及精确处理器。
	request      EmbeddedOutputHostRequest
	registration *embeddedPumpRegistration
	context      *EmbeddedHostCallbackContext
	// finished proves actual handler return; frame freezes the original completion for exact acknowledgement.
	// finished 证明实际处理器返回；frame 冻结原始完成以便精确确认。
	finished bool
	frame    []byte
	// effects seals actual host evidence; parserRejected permits one bounded rejection acknowledgement after proven non-dispatch.
	// effects 封存实际宿主证据；parserRejected 在已证明未分发后允许一次有界拒绝确认。
	effects        EmbeddedInputEffectState
	parserRejected bool
	// encodeError and ackFailed retain separate serialization and native acknowledgement failures.
	// encodeError 和 ackFailed 分别保留序列化及原生确认失败。
	encodeError error
	ackFailed   bool
}

// embeddedPumpMutation retains one original side-effecting native attempt before entering the adapter.
// embeddedPumpMutation 在进入适配器前保留一个原始变更原生尝试。
type embeddedPumpMutation struct {
	// kind selects a coordinator-owned mutation, while frame and response preserve original request and delivery bytes.
	// kind 选择协调器拥有的变更，frame 和 response 保留原始请求及交付字节。
	kind     string
	frame    []byte
	response []byte
	// returned proves adapter return; err is published only after the coordinator cannot settle the retained mutation.
	// returned 证明适配器返回；err 仅在协调器无法收尾保留变更后发布。
	err      error
	returned bool
	// Exactly the owner required by kind is set before native entry; take owns the complete returned batch instead.
	// 原生入场前设置 kind 所要求的精确所有者；领取操作则拥有完整返回批次。
	publication  *embeddedPumpPublication
	registration *embeddedPumpRegistration
	request      *embeddedPumpRequest
}

// embeddedCallbackPumps strongly retains active and failed owners until proven closure.
// embeddedCallbackPumps 强引用保留活动及失败所有者，直到已证明关闭。
var embeddedCallbackPumps sync.Map

// EmbeddedCallbackPump owns one runtime's queue, a sequential native coordinator and fixed handler goroutines.
// EmbeddedCallbackPump 拥有一个运行时队列、顺序原生协调器及固定处理器协程。
// Register all queued handlers for this runtime through this pump; close it before freeing its borrowed transport.
// 通过此泵注册此运行时的全部队列处理器；释放借用传输前必须关闭此泵。
type EmbeddedCallbackPump struct {
	// mu protects publication and observation only; native and application calls always run outside it.
	// mu 仅保护发布及观察；原生和应用调用始终在锁外执行。
	mu sync.Mutex
	// transport, runtimeID and config remain immutable after construction.
	// transport、runtimeID 及 config 在构造后保持不可变。
	transport *EmbeddedTransport
	runtimeID string
	config    EmbeddedCallbackPumpConfig
	// publications, registrations and requests retain exact owners across every asynchronous boundary.
	// publications、registrations 及 requests 跨每个异步边界保留精确所有者。
	publications  []*embeddedPumpPublication
	registrations map[string]*embeddedPumpRegistration
	requests      map[string]*embeddedPumpRequest
	// pending records uncertain mutations; needsRelease fences further native calls until explicit buffer recovery.
	// pending 记录不确定变更；needsRelease 在显式缓冲恢复前封闭后续原生调用。
	pending      *embeddedPumpMutation
	needsRelease bool
	// ready, closing, closed and failure describe actual SDK lifecycle; fatal means the coordinator cannot safely continue.
	// ready、closing、closed 及 failure 描述实际 SDK 生命周期；fatal 表示协调器不能安全继续。
	ready, closing, closed bool
	failure, fatal         error
	// started and retry own cancellation-independent observations; ended closes only after coordinator exit.
	// started 及 retry 拥有独立于取消的观察；ended 仅在协调器退出后关闭。
	started *embeddedPumpObservation
	retry   *embeddedPumpObservation
	ended   chan struct{}
	// wake coalesces advisory changes; jobs and workers bound actual handler infrastructure.
	// wake 合并参考变化；jobs 及 workers 限制实际处理器基础设施。
	wake    chan struct{}
	jobs    chan *embeddedPumpRequest
	workers sync.WaitGroup
}

// NewEmbeddedCallbackPump claims bounded ownership and starts infrastructure without waiting for native initialization checks.
// NewEmbeddedCallbackPump 声明有界所有权并启动基础设施，不等待原生初始化检查。
// transport is borrowed, runtimeID must be exact and nonempty, and Ready observes actual startup separately.
// transport 为借用对象，runtimeID 必须精确且非空，Ready 单独观察实际启动。
func NewEmbeddedCallbackPump(transport *EmbeddedTransport, runtimeID string, config EmbeddedCallbackPumpConfig) (*EmbeddedCallbackPump, error) {
	maximum := uint64(^uint(0) >> 1)
	if transport == nil || runtimeID == "" {
		return nil, fmt.Errorf("callback pump requires a transport and exact runtime identity")
	}
	if config.MaxConcurrentHandlers == 0 || config.MaxConcurrentHandlers > maximum || config.MaxPendingCommands == 0 || config.MaxPendingCommands > maximum || config.PollIntervalMS == 0 || config.PollIntervalMS > uint64((1<<63-1)/int64(time.Millisecond)) {
		return nil, fmt.Errorf("callback pump limits must fit positive worker counts and timer durations")
	}
	pump := &EmbeddedCallbackPump{transport: transport, runtimeID: runtimeID, config: config, registrations: make(map[string]*embeddedPumpRegistration), requests: make(map[string]*embeddedPumpRequest), started: newEmbeddedPumpObservation(), ended: make(chan struct{}), wake: make(chan struct{}, 1), jobs: make(chan *embeddedPumpRequest, int(config.MaxConcurrentHandlers))}
	if err := transport.claimCallbackPump(pump); err != nil {
		return nil, err
	}
	embeddedCallbackPumps.Store(pump, struct{}{})
	for index := uint64(0); index < config.MaxConcurrentHandlers; index++ {
		pump.workers.Add(1)
		go pump.handlerWorker()
	}
	go pump.run()
	return pump, nil
}

// newEmbeddedPumpObservation creates one pump-owned completion whose lifecycle is independent of caller contexts.
// newEmbeddedPumpObservation 创建一个泵拥有的完成对象，其生命周期独立于调用方上下文。
func newEmbeddedPumpObservation() *embeddedPumpObservation {
	return &embeddedPumpObservation{done: make(chan struct{})}
}

// LiveEmbeddedCallbackPumps returns an independent discovery list, including failed owners awaiting explicit recovery.
// LiveEmbeddedCallbackPumps 返回独立发现清单，包含等待显式恢复的失败所有者。
func LiveEmbeddedCallbackPumps() []*EmbeddedCallbackPump {
	result := []*EmbeddedCallbackPump{}
	embeddedCallbackPumps.Range(func(key, value any) bool { result = append(result, key.(*EmbeddedCallbackPump)); return true })
	return result
}

// Status copies ownership and diagnostics under the publication lock; returned slices cannot mutate the pump.
// Status 在发布锁下复制所有权及诊断；返回切片不能修改泵。
func (p *EmbeddedCallbackPump) Status() EmbeddedCallbackPumpStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	status := EmbeddedCallbackPumpStatus{RuntimeID: p.runtimeID, Ready: p.ready, Closing: p.closing, Closed: p.closed, PendingCommands: uint64(len(p.publications)), RecoveryRequired: p.needsRelease || p.pending != nil && p.pending.err != nil, Failure: cloneEmbeddedFailure(p.failure)}
	for id := range p.registrations {
		status.RegistrationIDs = append(status.RegistrationIDs, id)
	}
	for id, request := range p.requests {
		status.RequestIDs = append(status.RequestIDs, id)
		if request.finished {
			status.PendingAcknowledgements = append(status.PendingAcknowledgements, id)
		}
		status.RecoveryRequired = status.RecoveryRequired || request.ackFailed || request.encodeError != nil
	}
	sort.Strings(status.RegistrationIDs)
	sort.Strings(status.RequestIDs)
	sort.Strings(status.PendingAcknowledgements)
	return status
}

// Ready observes the actual native runtime startup check; cancellation detaches only this observer.
// Ready 观察实际原生运行时启动检查；取消仅分离此观察者。
func (p *EmbeddedCallbackPump) Ready(ctx context.Context) error { return p.observe(ctx, p.started) }

// Register freezes one generated queued batch and retains its exact handlers before native publication.
// Register 在原生发布前冻结一个生成队列批次并保留其精确处理器。
// ctx controls admission and observation only; accepted registration remains discoverable after caller cancellation.
// ctx 仅控制入场及观察；调用方取消后，已接纳注册仍可被发现。
func (p *EmbeddedCallbackPump) Register(ctx context.Context, capabilities []EmbeddedHostCapability) ([]string, error) {
	if err := checkEmbeddedObserver(ctx); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	descriptors := make(EmbeddedInputRuntimeCommandCapabilitiesRegisterDescriptors, len(capabilities))
	for index, capability := range capabilities {
		if capability.Handler == nil || capability.Descriptor.Execution != EmbeddedInputCapabilityExecutionQueued {
			return nil, fmt.Errorf("callback registration requires non-nil handlers and explicitly queued descriptors")
		}
		descriptors[index] = capability.Descriptor
	}
	command := EmbeddedInputCommandRuntime{Type: EmbeddedInputCommandRuntimeTypeRuntime, RuntimeId: p.runtimeID, Operation: EmbeddedInputRuntimeCommandCapabilitiesRegister{Type: EmbeddedInputRuntimeCommandCapabilitiesRegisterTypeCapabilitiesRegister, Descriptors: descriptors}}
	frame, err := p.encode(command)
	if err != nil {
		return nil, err
	}
	decoded, err := DecodeEmbeddedJSON(frame)
	if err != nil {
		return nil, err
	}
	frozen, err := projectEmbeddedWire(decoded, embeddedWireType[EmbeddedInputRequest](), "request")
	if err != nil {
		return nil, err
	}
	batch := frozen.Interface().(EmbeddedInputRequest).Command.(EmbeddedInputCommandRuntime).Operation.(EmbeddedInputRuntimeCommandCapabilitiesRegister).Descriptors
	owned := make([]EmbeddedHostCapability, len(batch))
	for index := range batch {
		owned[index] = EmbeddedHostCapability{Descriptor: batch[index], Handler: capabilities[index].Handler}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	publication := &embeddedPumpPublication{capabilities: owned, frame: frame, observation: newEmbeddedPumpObservation()}
	p.mu.Lock()
	if !p.ready || p.closing {
		p.mu.Unlock()
		return nil, &EmbeddedRuntimeError{"closed", "callback pump is not accepting registrations"}
	}
	if uint64(len(p.publications)) >= p.config.MaxPendingCommands {
		p.mu.Unlock()
		return nil, &EmbeddedRuntimeError{"capacity_exceeded", "callback publication capacity is exhausted"}
	}
	p.publications = append(p.publications, publication)
	p.mu.Unlock()
	p.signal()
	if err := p.observe(ctx, publication.observation); err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string{}, publication.identities...), nil
}

// Unregister starts exact registration retirement and waits for actual native and Go drainage without cancelling handlers.
// Unregister 启动精确注册退役，并等待真实原生及 Go 排空，不取消处理器。
func (p *EmbeddedCallbackPump) Unregister(ctx context.Context, registrationID string) error {
	if err := checkEmbeddedObserver(ctx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	p.mu.Lock()
	registration := p.registrations[registrationID]
	if registration == nil {
		p.mu.Unlock()
		return &EmbeddedRuntimeError{"not_found", "callback registration is not owned by this pump"}
	}
	registration.retiring = true
	p.mu.Unlock()
	p.signal()
	return p.observe(ctx, registration.observation)
}

// RetryAcknowledgements explicitly recovers retained native delivery and completion evidence without re-running handlers.
// RetryAcknowledgements 显式恢复保留原生交付及完成证据，不重新运行处理器。
func (p *EmbeddedCallbackPump) RetryAcknowledgements(ctx context.Context) error {
	if err := checkEmbeddedObserver(ctx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	p.mu.Lock()
	if p.closed || p.fatal != nil {
		err := cloneEmbeddedFailure(p.fatal)
		p.mu.Unlock()
		if err == nil {
			err = &EmbeddedRuntimeError{"closed", "callback pump already closed"}
		}
		return err
	}
	if p.retry == nil {
		p.retry = newEmbeddedPumpObservation()
	}
	observation := p.retry
	p.mu.Unlock()
	p.signal()
	return p.observe(ctx, observation)
}

// RequestClose nonblockingly fences new registration and starts cooperative retirement of all accepted ownership.
// RequestClose 非阻塞地封闭新注册，并启动全部已接纳所有权的协作退役。
func (p *EmbeddedCallbackPump) RequestClose() {
	p.mu.Lock()
	p.closing = true
	p.mu.Unlock()
	p.signal()
}

// Close requests drainage even for a cancelled observer; it succeeds only after actual handlers and coordinator work exit.
// Close 即使观察者已取消也请求排空；只有实际处理器及协调工作退出后才成功。
func (p *EmbeddedCallbackPump) Close(ctx context.Context) error {
	if err := checkEmbeddedObserver(ctx); err != nil {
		return err
	}
	p.RequestClose()
	if err := waitEmbeddedObservation(ctx, p.ended); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	return cloneEmbeddedFailure(p.fatal)
}

// signal coalesces advisory wakeups; authoritative queues and maps remain owned until the coordinator handles them.
// signal 合并参考唤醒；权威队列及映射保留到协调器处理。
func (p *EmbeddedCallbackPump) signal() {
	select {
	case p.wake <- struct{}{}:
	default:
	}
}

// finishLocked publishes an observation once; a timed-out observer never alters its original result.
// finishLocked 只发布一次观察；超时观察者绝不改变原结果。
func (p *EmbeddedCallbackPump) finishLocked(observation *embeddedPumpObservation, err error) {
	if !observation.finished {
		observation.err = cloneEmbeddedFailure(err)
		observation.finished = true
		close(observation.done)
	}
}

// observe waits for owned evidence or coordinator failure without propagating context cancellation into either owner.
// observe 等待拥有证据或协调器失败，不将上下文取消传播给任何所有者。
func (p *EmbeddedCallbackPump) observe(ctx context.Context, observation *embeddedPumpObservation) error {
	if err := checkEmbeddedObserver(ctx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-observation.done:
	case <-p.ended:
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if observation.finished {
		return cloneEmbeddedFailure(observation.err)
	}
	if p.fatal != nil {
		return cloneEmbeddedFailure(p.fatal)
	}
	return &EmbeddedRuntimeError{"closed", "callback pump ended before this observation completed"}
}
