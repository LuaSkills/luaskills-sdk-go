package luaskills

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

// embeddedControlWorkers is the single control-worker reservation used by SDK execution ownership.
// embeddedControlWorkers 是 SDK 执行所有权使用的唯一控制工作位预留值。
const embeddedControlWorkers uint64 = 1

// EmbeddedDriverConfig explicitly bounds worker concurrency and retained receipts separately for both lanes.
// EmbeddedDriverConfig 显式分别限制两个通道的工作并发和保留回执。
type EmbeddedDriverConfig struct {
	// WorkWorkers limits simultaneous business FFI requests; Lua VM limits remain core-owned.
	// WorkWorkers 限制同时进行的业务 FFI 请求；Lua VM 限制仍由核心拥有。
	WorkWorkers uint64
	// MaxWorkCommands includes queued, running, completed and uncertain business receipts.
	// MaxWorkCommands 包含排队、运行、完成及不确定业务回执。
	MaxWorkCommands uint64
	// MaxControlCommands reserves independent lifecycle receipt capacity, including completed receipts.
	// MaxControlCommands 预留独立生命周期回执容量，包含已完成回执。
	MaxControlCommands uint64
}

// EmbeddedDriverStatus is a copied snapshot of SDK ownership, independent from native operation state.
// EmbeddedDriverStatus 是 SDK 所有权的复制快照，独立于原生操作状态。
type EmbeddedDriverStatus struct {
	// Closing indicates fenced command admission; Closed proves all workers exited and the transport claim was released.
	// Closing 表示命令入场已封闭；Closed 证明全部工作位退出且传输声明已释放。
	Closing bool
	Closed  bool
	// Workers counts owned worker goroutines, including those waiting for recovery.
	// Workers 统计拥有的工作协程，包含等待恢复的工作位。
	Workers uint64
	// Queued and Running count actual SDK command ownership.
	// Queued 和 Running 统计实际 SDK 命令所有权。
	Queued  uint64
	Running uint64
	// Paused counts workers retaining their native frame reservation after failed result release.
	// Paused 统计结果释放失败后仍保留原生帧预留的工作位。
	Paused uint64
	// WorkReceipts and ControlReceipts include completed receipts until explicit Forget.
	// WorkReceipts 和 ControlReceipts 包含尚未显式 Forget 的已完成回执。
	WorkReceipts    uint64
	ControlReceipts uint64
	// Failure is an independent infrastructure diagnostic; non-nil never implies native drainage.
	// Failure 是独立基础设施诊断；非空绝不表示原生已排空。
	Failure error
}

// embeddedDrivers retains discoverable owners until actual close and explicit release of every retained receipt.
// embeddedDrivers 保留可发现所有者，直到实际关闭且显式释放全部保留回执。
var embeddedDrivers sync.Map

// EmbeddedCommandDriver owns fixed workers, two bounded queues and a prestarted close coordinator.
// EmbeddedCommandDriver 拥有固定工作位、两个有界队列及预先启动的关闭协调器。
// Do not copy it. Its borrowed transport and native runtimes require separate explicit closure.
// 不要复制。借用的传输及原生运行时需要单独显式关闭。
type EmbeddedCommandDriver struct {
	// mu protects admission, receipt publication and worker lifecycle, never a native call.
	// mu 保护入场、回执发布及工作位生命周期，绝不跨越原生调用。
	mu sync.Mutex
	// changed wakes workers and the close coordinator after ownership changes.
	// changed 在所有权变化后唤醒工作位和关闭协调器。
	changed *sync.Cond
	// transport and config remain immutable after construction.
	// transport 和 config 在构造后保持不可变。
	transport *EmbeddedTransport
	config    EmbeddedDriverConfig
	// commands retains exact receipt objects; queues hold only not-yet-dispatched receipts.
	// commands 保留精确回执对象；queues 仅持有尚未分发的回执。
	commands map[uint64]*EmbeddedCommand
	queues   map[EmbeddedCommandLane][]*EmbeddedCommand
	// counts bounds receipt retention independently per lane.
	// counts 分别限制各通道回执保留。
	counts map[EmbeddedCommandLane]uint64
	// nextID never reuses a local receipt identity; zero marks exhaustion after the full uint64 range.
	// nextID 绝不复用局部回执身份；零表示完整 uint64 范围耗尽。
	nextID uint64
	// workers, running and paused reflect actual worker ownership.
	// workers、running 及 paused 反映实际工作位所有权。
	workers uint64
	running uint64
	paused  uint64
	// closing fences admission; closed requires safe worker exit and claim release.
	// closing 封闭入场；closed 要求工作位安全退出且声明已释放。
	closing bool
	closed  bool
	// maintenance prevents worker dispatch while retained buffers are explicitly recovered.
	// maintenance 在显式恢复保留缓冲期间阻止工作位分发。
	maintenance bool
	// failure preserves uncertain infrastructure state; no automatic mutation replay or forced release follows.
	// failure 保留不确定基础设施状态；其后不自动重放变更或强制释放。
	failure error
	// closeRequested, stopped and fault independently own closure request, terminal observation and unsafe failure.
	// closeRequested、stopped 和 fault 分别拥有关闭请求、终止观察及不安全故障。
	closeRequested chan struct{}
	stopped        chan struct{}
	fault          chan struct{}
	// recovered is replaced only after successful native result recovery, waking paused workers without polling.
	// recovered 仅在原生结果恢复成功后替换，无需轮询即可唤醒暂停工作位。
	recovered chan struct{}
}

// NewEmbeddedCommandDriver validates config and claims worst-case native response capacity before starting goroutines.
// NewEmbeddedCommandDriver 在启动协程前校验 config，并声明最坏情况下的原生响应容量。
// transport must be live and have no existing ordinary driver; no native business command runs during construction.
// transport 必须活动且没有既有普通驱动器；构造期间不执行原生业务命令。
func NewEmbeddedCommandDriver(transport *EmbeddedTransport, config EmbeddedDriverConfig) (*EmbeddedCommandDriver, error) {
	maximum := uint64(^uint(0) >> 1)
	for _, limit := range []uint64{config.WorkWorkers, config.MaxWorkCommands, config.MaxControlCommands} {
		if limit == 0 || limit > maximum {
			return nil, fmt.Errorf("embedded driver budgets must fit positive Go integers")
		}
	}
	if config.WorkWorkers > config.MaxWorkCommands || config.WorkWorkers > maximum-embeddedControlWorkers {
		return nil, fmt.Errorf("invalid embedded worker/receipt relationship")
	}
	if transport == nil {
		return nil, fmt.Errorf("embedded driver requires a transport")
	}
	driver := &EmbeddedCommandDriver{transport: transport, config: config, commands: map[uint64]*EmbeddedCommand{}, queues: map[EmbeddedCommandLane][]*EmbeddedCommand{}, counts: map[EmbeddedCommandLane]uint64{}, nextID: 1, workers: config.WorkWorkers + embeddedControlWorkers, closeRequested: make(chan struct{}), stopped: make(chan struct{}), fault: make(chan struct{}), recovered: make(chan struct{})}
	driver.changed = sync.NewCond(&driver.mu)
	if err := transport.claimCommandDriver(driver); err != nil {
		return nil, err
	}
	embeddedDrivers.Store(driver, struct{}{})
	go driver.awaitClose()
	for index := uint64(0); index < config.WorkWorkers; index++ {
		go driver.worker(EmbeddedWorkLane)
	}
	for index := uint64(0); index < embeddedControlWorkers; index++ {
		go driver.worker(EmbeddedControlLane)
	}
	return driver, nil
}

// LiveEmbeddedCommandDrivers returns a copied discovery list; completed retained receipts remain discoverable after close.
// LiveEmbeddedCommandDrivers 返回复制发现清单；已完成保留回执在关闭后仍可发现。
func LiveEmbeddedCommandDrivers() []*EmbeddedCommandDriver {
	result := []*EmbeddedCommandDriver{}
	embeddedDrivers.Range(func(key, value any) bool { result = append(result, key.(*EmbeddedCommandDriver)); return true })
	return result
}

// Config returns a value copy of the immutable admission and worker limits.
// Config 返回不可变入场及工作位限制的值副本。
func (d *EmbeddedCommandDriver) Config() EmbeddedDriverConfig { return d.config }

// Commands returns retained exact receipt objects ordered by their local submission identities.
// Commands 按局部提交身份顺序返回保留的精确回执对象。
func (d *EmbeddedCommandDriver) Commands() []*EmbeddedCommand {
	d.mu.Lock()
	defer d.mu.Unlock()
	result := make([]*EmbeddedCommand, 0, len(d.commands))
	for _, command := range d.commands {
		result = append(result, command)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].identity < result[j].identity })
	return result
}

// Status returns a synchronized copied snapshot without inferring native lifecycle or cancelling work.
// Status 返回同步复制快照，不推断原生生命周期，也不取消工作。
func (d *EmbeddedCommandDriver) Status() EmbeddedDriverStatus {
	d.mu.Lock()
	defer d.mu.Unlock()
	return EmbeddedDriverStatus{Closing: d.closing, Closed: d.closed, Workers: d.workers, Queued: uint64(len(d.queues[EmbeddedWorkLane])) + uint64(len(d.queues[EmbeddedControlLane])), Running: d.running, Paused: d.paused, WorkReceipts: d.counts[EmbeddedWorkLane], ControlReceipts: d.counts[EmbeddedControlLane], Failure: cloneEmbeddedFailure(d.failure)}
}

// Submit freezes command and publishes its receipt before one worker can enter C; ctx only governs pre-admission cancellation.
// Submit 在工作位能够进入 C 前冻结 command 并发布回执；ctx 仅治理入场前取消。
// Once accepted, caller cancellation never cancels execution or discards the receipt; operation_wait is unsupported.
// 接纳后，调用方取消绝不取消执行或丢弃回执；不支持 operation_wait。
func (d *EmbeddedCommandDriver) Submit(ctx context.Context, command EmbeddedInputCommand) (*EmbeddedCommand, error) {
	if err := checkEmbeddedObserver(ctx); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	d.mu.Lock()
	closing := d.closing
	d.mu.Unlock()
	if closing {
		return nil, &EmbeddedRuntimeError{"closed", "embedded command driver is closing"}
	}
	frame, err := EncodeEmbeddedRequest(EmbeddedInputRequest{ProtocolVersion: EmbeddedProtocolVersion, Command: command}, d.transport.Config().MaxRequestBytes)
	if err != nil {
		return nil, err
	}
	decoded, err := DecodeEmbeddedJSON(frame)
	if err != nil {
		return nil, err
	}
	frozen := decoded.(map[string]any)["command"].(map[string]any)
	lane, err := embeddedCommandLane(frozen)
	if err != nil {
		return nil, err
	}
	// Context implementations are caller-owned; never invoke their methods while holding the driver lock.
	// 上下文实现由调用方拥有；绝不在持有驱动器锁时调用其方法。
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closing {
		return nil, &EmbeddedRuntimeError{"closed", "embedded command driver is closing"}
	}
	limit := d.config.MaxWorkCommands
	if lane == EmbeddedControlLane {
		limit = d.config.MaxControlCommands
	}
	if d.counts[lane] >= limit {
		return nil, &EmbeddedRuntimeError{"capacity_exceeded", "embedded receipt quota is full; forget completed receipts explicitly"}
	}
	if d.nextID == 0 {
		return nil, &EmbeddedRuntimeError{"capacity_exceeded", "embedded receipt identity range is exhausted"}
	}
	// Scope adoption holds the same admission lock, so a queued free can never race into an adopted lifetime.
	// 作用域接管持有相同入场锁，因此排队释放绝不会竞态进入已接管寿命。
	if frozen["type"] == "runtime_free" {
		if err := d.transport.checkUnmanagedRuntime(frozen["runtime_id"].(string)); err != nil {
			return nil, err
		}
	}
	receipt := &EmbeddedCommand{driver: d, identity: d.nextID, lane: lane, request: frame, state: EmbeddedCommandQueued, observed: make(chan struct{})}
	d.nextID++
	d.commands[receipt.identity] = receipt
	d.counts[lane]++
	d.queues[lane] = append(d.queues[lane], receipt)
	d.changed.Broadcast()
	return receipt, nil
}

// embeddedCommandLane classifies only implemented frozen routes; callers cannot place slow work in the control lane.
// embeddedCommandLane 仅分类已实现冻结路由；调用方不能将缓慢工作放入控制通道。
func embeddedCommandLane(command map[string]any) (EmbeddedCommandLane, error) {
	switch command["type"] {
	case "describe", "runtime_status", "runtime_close", "runtime_free":
		return EmbeddedControlLane, nil
	case "runtime_reserve", "runtime_initialize":
		return EmbeddedWorkLane, nil
	case "runtime":
		operation, ok := command["operation"].(map[string]any)
		if !ok {
			return "", fmt.Errorf("invalid embedded runtime command")
		}
		switch operation["type"] {
		case "plugin_register", "pool_register", "call_submit", "session_open", "session_submit", "capabilities_register":
			return EmbeddedWorkLane, nil
		case "plugin_status", "plugin_close", "plugin_forget", "pool_status", "pool_close", "pool_forget", "pool_revoke_permission", "session_status", "session_close", "session_forget", "operation_status", "operation_cancel", "operation_forget", "capabilities_list", "capability_status", "capability_unregister", "capability_forget", "host_requests_take", "host_request_status", "host_request_complete":
			return EmbeddedControlLane, nil
		case "operation_wait":
			return "", &EmbeddedRuntimeError{"unsupported", "poll operation_status instead of occupying a native command worker"}
		}
	}
	return "", fmt.Errorf("unsupported embedded command route")
}

// worker executes FIFO work for exactly one lane and retains its frame reservation while release recovery is pending.
// worker 为精确一个通道执行先进先出工作，并在等待释放恢复时保留帧预留。
func (d *EmbeddedCommandDriver) worker(lane EmbeddedCommandLane) {
	// active remains set until completion publication, exposing unexpected exit without inventing a result.
	// active 保持到完成发布，用于暴露意外退出而不凭空生成结果。
	var active *EmbeddedCommand
	defer func() {
		panicValue := recover()
		if panicValue != nil {
			d.fail(active, fmt.Sprintf("embedded native worker panicked (%T); ownership remains uncertain", panicValue))
		} else if active != nil {
			d.fail(active, "embedded native worker exited before publishing completion; ownership remains uncertain")
		}
		d.mu.Lock()
		d.workers--
		d.changed.Broadcast()
		d.mu.Unlock()
	}()
	for {
		d.mu.Lock()
		for d.failure == nil && (d.maintenance || len(d.queues[lane]) == 0) && !(d.closing && len(d.queues[lane]) == 0 && !d.maintenance) {
			d.changed.Wait()
		}
		if d.failure != nil || (d.closing && len(d.queues[lane]) == 0) {
			d.mu.Unlock()
			return
		}
		active = d.queues[lane][0]
		d.queues[lane][0] = nil
		d.queues[lane] = d.queues[lane][1:]
		if len(d.queues[lane]) == 0 {
			d.queues[lane] = nil
		}
		active.state = EmbeddedCommandRunning
		d.running++
		d.mu.Unlock()
		response, err := d.transport.requestBytes(active.request)
		if err == nil {
			_, err = DecodeEmbeddedResponse(response)
		}
		failure := cloneEmbeddedFailure(err)
		var releaseError *EmbeddedResultReleaseError
		paused := errors.As(err, &releaseError)
		d.mu.Lock()
		d.running--
		active.response = bytes.Clone(response)
		active.failure = failure
		active.state = EmbeddedCommandCompleted
		recovered := d.recovered
		if paused {
			d.paused++
		}
		close(active.observed)
		active = nil
		d.changed.Broadcast()
		d.mu.Unlock()
		if paused {
			select {
			case <-recovered:
			case <-d.fault:
			}
			d.mu.Lock()
			d.paused--
			d.changed.Broadcast()
			d.mu.Unlock()
		}
	}
}

// fail fences future dispatch, fails only never-dispatched commands and retains all uncertain native ownership.
// fail 封闭后续分发，仅直接失败从未分发的命令，并保留全部不确定原生所有权。
func (d *EmbeddedCommandDriver) fail(active *EmbeddedCommand, message string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.failure == nil {
		d.failure = &EmbeddedDriverFailure{message}
		close(d.fault)
	}
	if active != nil && active.state == EmbeddedCommandRunning {
		d.running--
		active.state = EmbeddedCommandUncertain
		active.failure = cloneEmbeddedFailure(d.failure)
		close(active.observed)
	}
	for lane, queue := range d.queues {
		for _, command := range queue {
			command.state = EmbeddedCommandCompleted
			command.failure = cloneEmbeddedFailure(d.failure)
			close(command.observed)
		}
		d.queues[lane] = nil
	}
	d.requestCloseLocked()
	d.changed.Broadcast()
}

// forget releases only exact proven-complete local ownership; it never invokes a native operation-forget command.
// forget 仅释放精确且已证明完成的局部所有权；绝不调用原生操作遗忘命令。
func (d *EmbeddedCommandDriver) forget(command *EmbeddedCommand) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.commands[command.identity] != command {
		return fmt.Errorf("embedded receipt is not retained by this driver")
	}
	if command.state != EmbeddedCommandCompleted {
		return &EmbeddedRuntimeError{"busy", "embedded command has not proven completion"}
	}
	delete(d.commands, command.identity)
	d.counts[command.lane]--
	if d.closed && len(d.commands) == 0 {
		embeddedDrivers.Delete(d)
	}
	return nil
}

// requestCloseLocked fences admission once and wakes the coordinator; caller must hold mu.
// requestCloseLocked 一次性封闭入场并唤醒协调器；调用方必须持有 mu。
func (d *EmbeddedCommandDriver) requestCloseLocked() {
	if !d.closing {
		d.closing = true
		close(d.closeRequested)
	}
	d.changed.Broadcast()
}

// RequestClose starts owned drainage immediately without cancelling queued commands or closing the borrowed transport.
// RequestClose 立即启动拥有型排空，不取消排队命令，也不关闭借用传输。
func (d *EmbeddedCommandDriver) RequestClose() { d.mu.Lock(); d.requestCloseLocked(); d.mu.Unlock() }

// Close starts drainage even for an already cancelled ctx, then observes actual worker shutdown with that context.
// Close 即使 ctx 已取消也启动排空，然后用该上下文观察实际工作位关闭。
// Timeout keeps the prestarted coordinator alive; retained-release workers need explicit ReleaseResults to proceed.
// 超时保持预先启动的协调器存活；保留释放的工作位需要显式 ReleaseResults 才能继续。
func (d *EmbeddedCommandDriver) Close(ctx context.Context) error {
	if err := checkEmbeddedObserver(ctx); err != nil {
		return err
	}
	d.RequestClose()
	if err := waitEmbeddedObservation(ctx, d.stopped); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return cloneEmbeddedFailure(d.failure)
}

// awaitClose waits for real worker exits before releasing the exact transport claim, retaining uncertain failures.
// awaitClose 等待真实工作位退出后才释放精确传输声明，并保留不确定故障。
func (d *EmbeddedCommandDriver) awaitClose() {
	<-d.closeRequested
	d.mu.Lock()
	for d.workers != 0 || d.maintenance {
		d.changed.Wait()
	}
	failure := d.failure
	if failure == nil {
		// Releasing the claim takes only a short Go ownership lock; keep recovery fenced until Closed is published.
		// 释放声明只取得短时 Go 所有权锁；发布 Closed 前持续封闭恢复入场。
		failure = d.transport.releaseCommandDriver(d)
	}
	if failure != nil {
		d.failure = cloneEmbeddedFailure(failure)
	} else {
		d.closed = true
		if len(d.commands) == 0 {
			embeddedDrivers.Delete(d)
		}
	}
	close(d.stopped)
	d.mu.Unlock()
}

// ReleaseResults retries exact retained buffer frees independently from receipt quotas; it never re-executes a command.
// ReleaseResults 独立于回执配额重试精确保留缓冲释放；绝不重新执行命令。
// ctx is checked before recovery starts; active readers cause Busy, and accepted synchronous recovery cannot be cancelled.
// 恢复开始前检查 ctx；活动读取者导致忙错误，已接纳的同步恢复不能被取消。
func (d *EmbeddedCommandDriver) ReleaseResults(ctx context.Context) (err error) {
	if err := checkEmbeddedObserver(ctx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	d.mu.Lock()
	if d.failure != nil {
		err := cloneEmbeddedFailure(d.failure)
		d.mu.Unlock()
		return err
	}
	if d.closed {
		d.mu.Unlock()
		return &EmbeddedRuntimeError{"closed", "embedded driver already closed"}
	}
	if d.running != 0 || d.maintenance {
		d.mu.Unlock()
		return &EmbeddedRuntimeError{"busy", "embedded native readers or recovery are active"}
	}
	d.maintenance = true
	d.mu.Unlock()
	// A recovery adapter panic or Goexit cannot prove whether C released an allocation; retain the entire owner.
	// 恢复适配器 panic 或 Goexit 无法证明 C 是否释放分配；保留整个所有者。
	returned := false
	defer func() {
		panicValue := recover()
		if !returned {
			diagnostic := fmt.Sprintf("embedded result recovery exited unexpectedly (%T); ownership remains uncertain", panicValue)
			d.fail(nil, diagnostic)
			err = &EmbeddedDriverFailure{diagnostic}
		}
		d.mu.Lock()
		d.maintenance = false
		if returned && err == nil {
			close(d.recovered)
			d.recovered = make(chan struct{})
		}
		d.changed.Broadcast()
		d.mu.Unlock()
	}()
	err = d.transport.ReleaseResults()
	returned = true
	return err
}
