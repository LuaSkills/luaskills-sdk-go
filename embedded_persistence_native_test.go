//go:build cgo

package luaskills

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

// TestEmbeddedClientNativePersistence verifies the actual DLL's typed storage, history and independent retention fences.
// TestEmbeddedClientNativePersistence 验证实际 DLL 的类型化存储、历史与独立保留屏障。
func TestEmbeddedClientNativePersistence(t *testing.T) {
	// All resources share the existing actual native fixture's ordered cleanup ownership.
	// 全部资源共享既有实际原生夹具的有序清理所有权。
	runtime, root, _ := nativeTypedRuntimeWithPersistence(t, true)
	// Status supplies the original core namespace rather than a guessed FFI identity.
	// 状态提供原核心命名空间，而非猜测 FFI 身份。
	statusPending, err := runtime.Status(context.Background())
	status := typedTake(t, statusPending, err)
	if status.Persistence == nil || status.CoreRuntimeId == nil {
		t.Fatalf("missing durable ownership: %#v", status)
	}
	// A healthy recovery is an explicit no-op, not an execution retry.
	// 健康恢复是显式无操作，不是执行重试。
	recovery, err := runtime.RecoverStorage(context.Background())
	if typedTake(t, recovery, err) {
		t.Fatal("healthy storage reported recovery")
	}
	// A healthy actual writer is not replaced merely because recovery was requested.
	// 健康实际写入者不会仅因恢复请求而被替换。
	workerRecovery, err := runtime.RecoverStorageWorker(context.Background())
	if typedTake(t, workerRecovery, err) {
		t.Fatal("healthy writer was replaced")
	}
	// Writer status uses the independently reserved control lane.
	// 写入者状态使用独立预留的控制通道。
	writer, err := runtime.StorageStatus(context.Background())
	if typedTake(t, writer, err).Closing {
		t.Fatal("writer closed before runtime closure")
	}
	// The Lua result and admission context both cross the real C boundary.
	// Lua 结果和入场上下文均经过真实 C 边界。
	pool := typedPool(t, runtime, root, "return {call=function(a) return a end}", false)
	// Preserve this admitted operation independently of transient SDK command receipts.
	// 独立于瞬态 SDK 命令回执保留此入场操作。
	submission, err := pool.Submit(context.Background(), "call", "持久结果", EmbeddedInputLuaInvocationContext{}, 10000)
	operation := typedTake(t, submission, err)
	// The observed terminal snapshot is acknowledged in durable storage.
	// 观测到的终态快照已在持久存储中确认。
	done, err := operation.Wait(driverTestContext(t))
	if err != nil {
		t.Fatal(err)
	}
	// Failure observation does not create a new checkpoint or business execution.
	// 故障观测不创建新检查点或业务执行。
	failure, err := operation.PersistenceFailure(context.Background())
	if typedTake(t, failure, err) != nil {
		t.Fatal("successful operation has a persistence failure")
	}
	// No failure is an explicit native conflict rather than a successful checkpoint retry.
	// 不存在故障是明确原生冲突，而非成功检查点重试。
	retry, err := operation.RetryCheckpoint(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_, err = retry.Result(driverTestContext(t))
	// Exact business error identity remains available through typed projections.
	// 精确业务错误身份仍可通过类型化投影取得。
	var nativeError *EmbeddedRuntimeError
	if !errors.As(err, &nativeError) || nativeError.Code != "busy" {
		t.Fatalf("unexpected retry result: %v", err)
	}
	if err := retry.Forget(); err != nil {
		t.Fatal(err)
	}
	// Read original evidence without adopting it as another active operation.
	// 读取原证据，不将其接管为另一个活动操作。
	reading, err := runtime.HistoryGet(context.Background(), *status.CoreRuntimeId, operation.OperationID())
	history := typedTake(t, reading, err)
	if history == nil || !reflect.DeepEqual(history.Snapshot, done) {
		t.Fatalf("history changed: %#v", history)
	}
	// Cursor input and nullable output retain the generated wire contract.
	// 游标输入及可空输出保留生成的线契约。
	first, err := runtime.HistoryNext(context.Background(), nil)
	if !reflect.DeepEqual(typedTake(t, first, err), history) {
		t.Fatal("first history row changed")
	}
	end, err := runtime.HistoryNext(context.Background(), &EmbeddedInputHistoryCursor{RuntimeId: history.RuntimeId, OperationId: operation.OperationID()})
	if typedTake(t, end, err) != nil {
		t.Fatal("cursor did not reach enumeration end")
	}
	// The trusted test host inspected this exact pure source; a successful return alone proves no external facts.
	// 可信测试宿主检查了此精确纯源码；仅成功返回不能证明外部事实。
	resolution := EmbeddedInputOperationReconciliation{
		ResolutionId: "go-audit", Resolver: "trusted-test-host", Evidence: "fixture:pure-source-and-stopped-owner",
		Execution: EmbeddedInputReconciledExecutionObservedTerminal, Effects: EmbeddedInputResolvedEffectStateNotApplicable,
		HostEffects: EmbeddedInputOperationReconciliationHostEffects{},
	}
	// Even a terminal retained operation must be explicitly released before final audit mutation.
	// 即使终态保留操作也必须在最终审计变更前显式释放。
	blocked, err := runtime.HistoryReconcile(context.Background(), history.RuntimeId, operation.OperationID(), history.Revision, resolution)
	if err != nil {
		t.Fatal(err)
	}
	_, err = blocked.Result(driverTestContext(t))
	if !errors.As(err, &nativeError) || nativeError.Code != "busy" {
		t.Fatalf("live history reconciled: %v", err)
	}
	if err := blocked.Forget(); err != nil {
		t.Fatal(err)
	}
	// Removing live metadata cannot erase unresolved ordinary Lua effects from storage.
	// 移除活动元数据不能从存储抹除普通 Lua 未决副作用。
	forgotten, err := operation.Forget(context.Background())
	typedTake(t, forgotten, err)
	deletion, err := runtime.HistoryForget(context.Background(), history.RuntimeId, operation.OperationID(), history.Revision)
	if err != nil {
		t.Fatal(err)
	}
	_, err = deletion.Result(driverTestContext(t))
	if !errors.As(err, &nativeError) || nativeError.Code != "busy" {
		t.Fatalf("history unexpectedly deleted: %v", err)
	}
	if err := deletion.Forget(); err != nil {
		t.Fatal(err)
	}
	// Exact proof retries share one durable successor and never replay the original Lua export.
	// 精确证明重试共享一个持久后继，绝不重放原 Lua 导出。
	finalizing, err := runtime.HistoryReconcile(context.Background(), history.RuntimeId, operation.OperationID(), history.Revision, resolution)
	revision := typedTake(t, finalizing, err)
	if revision != history.Revision+1 {
		t.Fatalf("unexpected final revision: %d", revision)
	}
	repeated, err := runtime.HistoryReconcile(context.Background(), history.RuntimeId, operation.OperationID(), history.Revision, resolution)
	if typedTake(t, repeated, err) != revision {
		t.Fatal("exact retry changed final revision")
	}
	// Original observations remain byte-equivalent in meaning beside the separate final attestation.
	// 独立最终证明旁的原观测保持等义。
	reading, err = runtime.HistoryGet(context.Background(), history.RuntimeId, operation.OperationID())
	reconciled := typedTake(t, reading, err)
	if reconciled == nil || !reflect.DeepEqual(reconciled.Snapshot, history.Snapshot) || reconciled.Reconciliation == nil || reconciled.Reconciliation.ResolutionId != resolution.ResolutionId {
		t.Fatalf("history lost original evidence: %#v", reconciled)
	}
	deletion, err = runtime.HistoryForget(context.Background(), history.RuntimeId, operation.OperationID(), revision)
	typedTake(t, deletion, err)
	reading, err = runtime.HistoryGet(context.Background(), history.RuntimeId, operation.OperationID())
	if typedTake(t, reading, err) != nil {
		t.Fatal("resolved history was not removed")
	}
}

// availablePersistenceFailure reads the exact operation once its nonblocking metadata gate is available.
// availablePersistenceFailure 在非阻塞元数据门禁可用后读取精确操作。
// Return actual optional failure; retry only explicit busy queries within the fixture deadline and release every receipt.
// 返回实际可空故障；仅在夹具期限内重试明确忙碌查询，并释放每个回执。
func availablePersistenceFailure(t *testing.T, operation *EmbeddedOperation) *EmbeddedOutputOperationPersistenceFailure {
	t.Helper()
	// The fixture context bounds repeated observations without changing native execution.
	// 夹具上下文约束重复观察，不改变原生执行。
	ctx := driverTestContext(t)
	for {
		// Each observation remains a separate completed command receipt.
		// 每次观察保持为独立完成命令回执。
		pending, err := operation.PersistenceFailure(ctx)
		if err != nil {
			t.Fatal(err)
		}
		// Observe the real result before releasing command capacity.
		// 释放命令容量前观测真实结果。
		failure, err := pending.Result(ctx)
		if forgetErr := pending.Forget(); forgetErr != nil {
			t.Fatal(forgetErr)
		}
		if err == nil {
			return failure
		}
		// Only the native nonblocking observation conflict is retryable here.
		// 此处仅允许重试原生非阻塞观察冲突。
		var nativeError *EmbeddedRuntimeError
		if !errors.As(err, &nativeError) || nativeError.Code != "busy" {
			t.Fatal(err)
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(time.Millisecond):
		}
	}
}

// TestEmbeddedClientNativePersistenceCapacity verifies actual durable Go callbacks and explicit storage-only recovery.
// TestEmbeddedClientNativePersistenceCapacity 验证实际持久 Go 回调及显式仅存储恢复。
func TestEmbeddedClientNativePersistenceCapacity(t *testing.T) {
	// The sole history row is occupied by the first real business operation, without raw database mutation.
	// 唯一历史行由首个真实业务操作占用，不直接变更数据库。
	runtime, root, _ := nativeTypedRuntimeWithJournalLimit(t, true, 1)
	// Count actual host entries independently of operation result and storage observations.
	// 独立于操作结果及存储观测统计实际宿主进入次数。
	var calls atomic.Uint64
	// Actual pump cleanup precedes native runtime cleanup.
	// 实际泵清理先于原生运行时清理。
	pump := nativePumpTest(t, runtime.client.driver.transport, runtime.RuntimeID(), EmbeddedCallbackPumpConfig{1, 2, 1})
	pumpRegister(t, pump, pumpCapability(func(arguments any, callback *EmbeddedHostCallbackContext) (any, error) {
		calls.Add(1)
		if err := callback.ReportEffects(EmbeddedInputEffectStateCommitted); err != nil {
			return nil, err
		}
		return arguments, nil
	}))
	// Source has only the observed callback and no custom finalizers or other external effects.
	// 源码仅含已观测回调，没有自定义终结器或其他外部副作用。
	pool := typedPool(t, runtime, root, "return {call=function(a) local r=vulcan.capabilities.call('go.callback',a); return r.value end}", false)
	// Preserve the original handle separately from transient native command receipts.
	// 独立于瞬态原生命令回执保留原句柄。
	submitting, err := pool.Submit(context.Background(), "call", "committed-once", EmbeddedInputLuaInvocationContext{}, 10000)
	first := typedTake(t, submitting, err)
	// Terminal observation also proves the original callback has actually returned.
	// 终态观察也证明原回调已实际返回。
	firstDone, err := first.Wait(driverTestContext(t))
	if err != nil {
		t.Fatal(err)
	}
	// Core status supplies the original namespace, not the outer FFI slot.
	// 核心状态提供原命名空间，而非外层 FFI 槽。
	statusPending, err := runtime.Status(context.Background())
	status := typedTake(t, statusPending, err)
	if status.CoreRuntimeId == nil {
		t.Fatal("missing core namespace")
	}
	// Retain the complete durable callback evidence after releasing only live metadata.
	// 仅释放活动元数据后，保留完整持久回调证据。
	reading, err := runtime.HistoryGet(context.Background(), *status.CoreRuntimeId, first.OperationID())
	history := typedTake(t, reading, err)
	if history == nil {
		t.Fatal("first callback history missing")
	}
	forgotten, err := first.Forget(context.Background())
	typedTake(t, forgotten, err)
	// The next separate operation is accepted in memory but cannot persist its execution intent.
	// 下个独立操作在内存中被接纳，但无法持久化其执行意图。
	submitting, err = pool.Submit(context.Background(), "call", "must-not-run", EmbeddedInputLuaInvocationContext{}, 10000)
	second := typedTake(t, submitting, err)
	// Poll the independent control lane until the actual retained failure is available.
	// 轮询独立控制通道，直至实际保留故障可用。
	failure := embeddedPoll(t, func() any { return availablePersistenceFailure(t, second) }, func(value any) bool { return value.(*EmbeddedOutputOperationPersistenceFailure) != nil }).(*EmbeddedOutputOperationPersistenceFailure)
	// The host inspected the exact callback-only source and observed its actual handler completion.
	// 宿主检查了精确仅回调源码，并观测其实际处理器完成。
	resolution := EmbeddedInputOperationReconciliation{ResolutionId: "capacity-audit", Resolver: "trusted-test-host", Evidence: "fixture:joined-callback-and-no-other-effects", Execution: EmbeddedInputReconciledExecutionObservedTerminal, Effects: EmbeddedInputResolvedEffectStateCommitted, HostEffects: EmbeddedInputOperationReconciliationHostEffects{}}
	for _, effect := range history.Snapshot.HostEffects {
		resolution.HostEffects = append(resolution.HostEffects, EmbeddedInputHostEffectReconciliation{EffectId: effect.EffectId, Effects: EmbeddedInputResolvedEffectStateCommitted, Evidence: "fixture:actual-handler-completed"})
	}
	// Repair capacity through the public typed interface while the original failed operation is still owned.
	// 原失败操作仍被拥有时，通过公开类型化接口修复容量。
	finalizing, err := runtime.HistoryReconcile(context.Background(), history.RuntimeId, first.OperationID(), history.Revision, resolution)
	revision := typedTake(t, finalizing, err)
	deleting, err := runtime.HistoryForget(context.Background(), history.RuntimeId, first.OperationID(), revision)
	typedTake(t, deleting, err)
	// Capacity repair leaves the same candidate waiting for explicit storage retry.
	// 容量修复使同一候选继续等待显式存储重试。
	repairedFailure := availablePersistenceFailure(t, second)
	// Retry persists the original business failure rather than invoking its callback after the fact.
	// 重试持久化原业务失败，而非事后调用其回调。
	retrying, err := second.RetryCheckpoint(context.Background())
	requested := typedTake(t, retrying, err)
	// Observe the same operation after actual cleanup and final persistence.
	// 实际清理及最终持久化后观测同一操作。
	done, err := second.Wait(driverTestContext(t))
	if err != nil {
		t.Fatal(err)
	}
	if firstDone.Value == nil || *firstDone.Value != "committed-once" || !reflect.DeepEqual(history.Snapshot, firstDone) || len(firstDone.HostEffects) != 1 {
		t.Fatalf("original callback evidence changed: %#v", firstDone)
	}
	for _, effect := range firstDone.HostEffects {
		if effect.Effects != EmbeddedOutputEffectStateCommitted || effect.Caller.OperationId != first.OperationID() {
			t.Fatalf("wrong original callback: %#v", effect)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("callback replayed: %d", calls.Load())
	}
	if failure.OperationId != second.OperationID() || failure.Error.Code != EmbeddedOutputEmbeddedErrorCodeCapacityExceeded || failure.Retry != EmbeddedOutputCheckpointRetryStateWaiting || !reflect.DeepEqual(repairedFailure, failure) || !requested {
		t.Fatalf("original failure/retry changed: %#v %#v %v", failure, repairedFailure, requested)
	}
	// Retiring an already initialized VM conservatively keeps aggregate effects unknown.
	// 退役已初始化 VM 保守地使聚合副作用保持未知。
	if done.Phase != EmbeddedOutputOperationPhaseFailed || done.Error == nil || *done.Error == nil || (*done.Error).Code != EmbeddedOutputEmbeddedErrorCodeCapacityExceeded || done.Effects != EmbeddedOutputEffectStateUnknown || len(done.HostEffects) != 0 || availablePersistenceFailure(t, second) != nil {
		t.Fatalf("failed operation changed or replayed: %#v", done)
	}
	// Verify the original failed result is durable before independently auditing the fixture's retirement.
	// 独立审计夹具退役前，验证原失败结果已持久化。
	reading, err = runtime.HistoryGet(context.Background(), history.RuntimeId, second.OperationID())
	failedHistory := typedTake(t, reading, err)
	if failedHistory == nil || !reflect.DeepEqual(failedHistory.Snapshot, done) {
		t.Fatal("failed checkpoint was not durable")
	}
	forgotten, err = second.Forget(context.Background())
	typedTake(t, forgotten, err)
	// This distinct host proof covers the pure fixture's retirement without claiming successful business execution.
	// 此独立宿主证明覆盖纯夹具退役，不宣称业务成功执行。
	failedResolution := EmbeddedInputOperationReconciliation{ResolutionId: "failed-capacity-audit", Resolver: "trusted-test-host", Evidence: "fixture:retired-vm-without-finalizers-or-new-effects", Execution: EmbeddedInputReconciledExecutionObservedTerminal, Effects: EmbeddedInputResolvedEffectStateNotApplicable, HostEffects: EmbeddedInputOperationReconciliationHostEffects{}}
	finalizing, err = runtime.HistoryReconcile(context.Background(), history.RuntimeId, second.OperationID(), failedHistory.Revision, failedResolution)
	revision = typedTake(t, finalizing, err)
	deleting, err = runtime.HistoryForget(context.Background(), history.RuntimeId, second.OperationID(), revision)
	typedTake(t, deleting, err)
}
