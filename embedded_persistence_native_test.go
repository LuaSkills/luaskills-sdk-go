//go:build cgo

package luaskills

import (
	"context"
	"errors"
	"reflect"
	"testing"
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
