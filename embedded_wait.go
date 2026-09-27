package luaskills

import (
	"context"
	"fmt"
	"time"
)

// EmbeddedDefaultPollInterval is the shared SDK observation interval; it never changes native execution deadlines.
// EmbeddedDefaultPollInterval 是共享 SDK 观察间隔；绝不改变原生执行截止时间。
const EmbeddedDefaultPollInterval = 10 * time.Millisecond

// Wait polls using the default interval until ctx ends or a terminal snapshot is observed.
// Wait 使用默认间隔轮询，直到 ctx 结束或观察到终态快照。
// Failure and cancellation snapshots are returned with their effects; only observation failures become Go errors.
// 失败及取消快照连同副作用返回；只有观察失败成为 Go 错误。
func (h *EmbeddedOperation) Wait(ctx context.Context) (EmbeddedOutputOperationSnapshot, error) {
	return h.WaitInterval(ctx, EmbeddedDefaultPollInterval)
}

// WaitInterval polls with positive interval until ctx ends; it returns any proven terminal snapshot and its effects.
// WaitInterval 以正 interval 轮询直到 ctx 结束；返回任何已证明终态快照及其副作用。
// Only successful read-only polls are forgotten automatically; interrupted or failed receipts remain in Driver.Commands.
// 仅自动遗忘成功的只读轮询；中断或失败回执仍保留在 Driver.Commands 中。
func (h *EmbeddedOperation) WaitInterval(ctx context.Context, interval time.Duration) (EmbeddedOutputOperationSnapshot, error) {
	// empty never fabricates a terminal state on observer or transport failure.
	// empty 绝不在观察者或传输失败时伪造终态。
	var empty EmbeddedOutputOperationSnapshot
	if err := checkEmbeddedObserver(ctx); err != nil {
		return empty, err
	}
	if interval <= 0 {
		return empty, fmt.Errorf("embedded poll interval must be positive")
	}
	for {
		// pending survives a cancelled observation, preserving delivery evidence and its original SDK quota.
		// pending 在观察取消后继续存活，保留交付证据及原始 SDK 配额。
		pending, err := h.Status(ctx)
		if err != nil {
			return empty, err
		}
		snapshot, err := pending.Result(ctx)
		if err != nil {
			return empty, err
		}
		if err := pending.Forget(); err != nil {
			return empty, err
		}
		switch snapshot.Phase {
		case EmbeddedOutputOperationPhaseSucceeded, EmbeddedOutputOperationPhaseFailed, EmbeddedOutputOperationPhaseCancelled:
			return snapshot, nil
		}
		if err := waitEmbeddedPoll(ctx, interval); err != nil {
			return empty, err
		}
	}
}

// waitEmbeddedPoll waits for interval or ctx without allocating a goroutine or propagating cancellation into native work.
// waitEmbeddedPoll 等待 interval 或 ctx，不分配协程，也不将取消传播到原生工作。
func waitEmbeddedPoll(ctx context.Context, interval time.Duration) error {
	// timer is stopped on every exit, including early cancellation on the minimum supported Go version.
	// timer 在每个退出路径停止，包含最低支持 Go 版本上的提前取消。
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
