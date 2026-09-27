package luaskills

import (
	"bytes"
	"context"
	"errors"
	"fmt"
)

// EmbeddedCommandLane identifies independently bounded work or lifecycle-control admission.
// EmbeddedCommandLane 标识独立受限的工作或生命周期控制入场。
type EmbeddedCommandLane string

const (
	// EmbeddedWorkLane owns potentially slow construction and business submissions.
	// EmbeddedWorkLane 拥有可能缓慢的构造及业务提交。
	EmbeddedWorkLane EmbeddedCommandLane = "work"
	// EmbeddedControlLane reserves a worker for short status, cancellation and cleanup commands.
	// EmbeddedControlLane 为短状态、取消及清理命令预留工作位。
	EmbeddedControlLane EmbeddedCommandLane = "control"
)

// EmbeddedCommandState describes SDK execution evidence, never the phase of a native Lua operation.
// EmbeddedCommandState 描述 SDK 执行证据，绝不表示原生 Lua 操作的阶段。
type EmbeddedCommandState string

const (
	// EmbeddedCommandQueued has a retained receipt and has not entered the native adapter.
	// EmbeddedCommandQueued 已保留回执，尚未进入原生适配器。
	EmbeddedCommandQueued EmbeddedCommandState = "queued"
	// EmbeddedCommandRunning owns the native call, response validation or cleanup.
	// EmbeddedCommandRunning 拥有原生调用、响应校验或清理。
	EmbeddedCommandRunning EmbeddedCommandState = "running"
	// EmbeddedCommandCompleted returned from the adapter, possibly with a retained allocation-release error.
	// EmbeddedCommandCompleted 已从适配器返回，仍可能带有保留分配释放错误。
	EmbeddedCommandCompleted EmbeddedCommandState = "completed"
	// EmbeddedCommandUncertain lacks completion proof after an unexpected worker exit.
	// EmbeddedCommandUncertain 在工作位意外退出后缺少完成证明。
	EmbeddedCommandUncertain EmbeddedCommandState = "uncertain"
)

// EmbeddedDriverFailure reports an infrastructure failure while uncertain native ownership remains retained.
// EmbeddedDriverFailure 报告基础设施故障，同时保留不确定原生所有权。
type EmbeddedDriverFailure struct {
	// Message describes the failure without guessing a business outcome.
	// Message 描述故障，不猜测业务结果。
	Message string
}

// Error returns the infrastructure diagnostic; it does not authorize command replay.
// Error 返回基础设施诊断；不授权重放命令。
func (e *EmbeddedDriverFailure) Error() string { return e.Message }

// EmbeddedCommand retains one frozen request and its original delivery independently of observer contexts.
// EmbeddedCommand 独立于观察者上下文保留一个冻结请求及原始交付。
// Do not copy this receipt; only its exact driver-owned identity can return admission quota.
// 不要复制此回执；只有驱动器拥有的精确身份可以归还入场配额。
type EmbeddedCommand struct {
	// driver owns publication and retention; its mutex protects all mutable receipt fields.
	// driver 拥有发布及保留；其互斥锁保护全部可变回执字段。
	driver *EmbeddedCommandDriver
	// identity is local SDK metadata and is never sent as a native operation identity.
	// identity 是 SDK 局部元数据，绝不作为原生操作身份发送。
	identity uint64
	// lane and request remain immutable after publication.
	// lane 和 request 在发布后保持不可变。
	lane    EmbeddedCommandLane
	request []byte
	// state distinguishes proven completion from an uncertain worker failure.
	// state 区分已证明完成与不确定工作位故障。
	state EmbeddedCommandState
	// response owns copied bytes; failure owns a private diagnostic snapshot.
	// response 拥有复制字节；failure 拥有私有诊断快照。
	response []byte
	failure  error
	// observed closes once a result or infrastructure failure becomes observable.
	// observed 在结果或基础设施故障可以被观察时只关闭一次。
	observed chan struct{}
}

// CommandID returns an exact driver-local receipt identity, not a native handle.
// CommandID 返回精确驱动器局部回执身份，不是原生句柄。
func (c *EmbeddedCommand) CommandID() uint64 { return c.identity }

// Lane returns the immutable lane selected from the frozen command's declared route.
// Lane 返回从冻结命令声明路由选出的不可变通道。
func (c *EmbeddedCommand) Lane() EmbeddedCommandLane { return c.lane }

// State returns actual SDK ownership state under the driver's publication lock.
// State 在驱动器发布锁下返回实际 SDK 所有权状态。
func (c *EmbeddedCommand) State() EmbeddedCommandState {
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	return c.state
}

// Done reports proven adapter return, including explicit buffer-release failure; uncertainty is never completion.
// Done 报告已证明的适配器返回，包含显式缓冲释放失败；不确定状态绝不算完成。
func (c *EmbeddedCommand) Done() bool { return c.State() == EmbeddedCommandCompleted }

// RequestBytes returns an independent copy of the exact admitted protocol envelope.
// RequestBytes 返回精确已接纳协议信封的独立副本。
func (c *EmbeddedCommand) RequestBytes() []byte { return bytes.Clone(c.request) }

// ResponseBytes returns copied delivery, or nil if delivery has not been proven; it never waits or replays.
// ResponseBytes 返回复制交付，未证明交付时返回 nil；绝不等待或重放。
func (c *EmbeddedCommand) ResponseBytes() []byte {
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	return bytes.Clone(c.response)
}

// Result observes completion until ctx ends and returns fresh decoded data or an independent retained error.
// Result 观察完成直到 ctx 结束，并返回新解码数据或独立保留错误。
// Cancellation affects only this wait; native work, receipt quota and eventual evidence remain owned.
// 取消仅影响本次等待；原生工作、回执配额及最终证据仍被拥有。
func (c *EmbeddedCommand) Result(ctx context.Context) (any, error) {
	if err := waitEmbeddedObservation(ctx, c.observed); err != nil {
		return nil, err
	}
	c.driver.mu.Lock()
	response, failure := c.response, cloneEmbeddedFailure(c.failure)
	c.driver.mu.Unlock()
	if failure != nil {
		return nil, failure
	}
	return DecodeEmbeddedResponse(response)
}

// DeliveredResult decodes original response evidence even after cleanup failed, without changing ownership.
// DeliveredResult 即使清理失败也解码原始响应证据，不改变所有权。
func (c *EmbeddedCommand) DeliveredResult() (any, error) {
	response := c.ResponseBytes()
	if response == nil {
		return nil, fmt.Errorf("embedded command has no proven response delivery")
	}
	return DecodeEmbeddedResponse(response)
}

// Forget releases only this exact completed receipt's SDK quota; native operations need separate explicit cleanup.
// Forget 仅释放此精确已完成回执的 SDK 配额；原生操作需要单独显式清理。
func (c *EmbeddedCommand) Forget() error { return c.driver.forget(c) }

// waitEmbeddedObservation never propagates observer cancellation into the owned completion channel.
// waitEmbeddedObservation 绝不将观察者取消传播到拥有型完成通道。
// A nil context is invalid; an already cancelled observer fails even when a receipt is already available.
// 空上下文无效；已经取消的观察者即使已有回执也会失败。
func waitEmbeddedObservation(ctx context.Context, ready <-chan struct{}) error {
	if err := checkEmbeddedObserver(ctx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ready:
		return nil
	}
}

// cloneEmbeddedFailure prevents caller mutation of public error fields from rewriting another observer's evidence.
// cloneEmbeddedFailure 防止调用方修改公开错误字段，进而改写其他观察者的证据。
func cloneEmbeddedFailure(err error) error {
	switch value := err.(type) {
	case nil:
		return nil
	case *EmbeddedTransportError:
		copied := *value
		return &copied
	case *EmbeddedRuntimeError:
		copied := *value
		return &copied
	case *EmbeddedDriverFailure:
		copied := *value
		return &copied
	case *EmbeddedResultReleaseError:
		return &EmbeddedResultReleaseError{Native: cloneEmbeddedFailure(value.Native).(*EmbeddedTransportError), response: bytes.Clone(value.response), cause: cloneEmbeddedFailure(value.cause)}
	default:
		return errors.New(err.Error())
	}
}
