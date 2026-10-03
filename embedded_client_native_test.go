//go:build cgo

package luaskills

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEmbeddedClientNativeApplicationIntegerPolicy verifies native rejection before VM initialization and admission.
// TestEmbeddedClientNativeApplicationIntegerPolicy 验证在 VM 初始化及入场前的原生拒绝。
// t owns exact runtime cleanup; accepted safe endpoints and explicit finite floats must echo successfully.
// t 拥有精确运行时清理；已接纳安全端点及显式有限浮点数必须成功回传。
func TestEmbeddedClientNativeApplicationIntegerPolicy(t *testing.T) {
	// The failing initializer makes any accidental business admission observable.
	// 失败初始化器使任何意外业务入场都可观察。
	transport, runtimeID, command, pool := nativeEmbeddedRuntime(t)
	rejectingPool := pool("error('invalid arguments reached initialization')")
	// Exact JSON integer tokens cover positive, negative and recursively nested unsafe values.
	// 精确 JSON 整数 token 覆盖正数、负数及递归嵌套不安全值。
	for _, value := range []any{json.Number("9007199254740992"), json.Number("-9007199254740992"), map[string]any{"nested": []any{json.Number("18446744073709551615")}}} {
		// Bypass the success-only fixture helper to inspect the actual native rejection.
		// 绕过仅成功夹具辅助函数，以检查实际原生拒绝。
		_, err := transport.Request(map[string]any{"type": "runtime", "runtime_id": runtimeID, "operation": map[string]any{"type": "call_submit", "timeout_ms": 10000, "call": map[string]any{"pool_id": rejectingPool, "export": "call", "arguments": value, "context": map[string]any{"request_context": nil, "client_budget": nil, "tool_config": nil}}}})
		var failure *EmbeddedRuntimeError
		if !errors.As(err, &failure) || failure.Code != "invalid_argument" {
			t.Fatalf("unsafe application integer reached admission: %v", err)
		}
	}
	// An empty operation list proves that rejection did not retain any admitted business identity.
	// 空操作列表证明拒绝没有保留任何已入场业务身份。
	page := command(map[string]any{"type": "operation_list", "pool_id": rejectingPool, "after_operation_id": nil, "limit": 16}).(map[string]any)
	if len(page["operation_ids"].([]any)) != 0 {
		t.Fatalf("rejected application values created operations: %#v", page)
	}
	// Decimal Float tokens preserve IEEE754 magnitude independently of the integer admission policy.
	// 十进制 Float token 独立于整数入场政策保留 IEEE754 量级。
	echoPool := pool("return {call=function(a) return a end}")
	for _, value := range []json.Number{"9007199254740991", "-9007199254740991", "9007199254740992.0", "1e100"} {
		// Consume and release each exact operation only after terminal observation.
		// 终态观察后才消费及释放每个精确操作。
		id := embeddedSubmit(command, echoPool, value)
		done := embeddedTerminal(t, command, id)
		actual := done["value"].(json.Number)
		if done["phase"] != "succeeded" || (value != "1e100" && actual != value) {
			t.Fatalf("safe integer or Float changed: %#v", done)
		}
		if value == "1e100" {
			// Compare the declared floating value numerically; exponent spelling is not the application contract.
			// 按数值比较声明浮点值；指数拼写不是应用契约。
			magnitude, err := actual.Float64()
			if err != nil || magnitude != 1e100 {
				t.Fatalf("finite Float magnitude changed: %s %v", actual, err)
			}
		}
		command(map[string]any{"type": "operation_forget", "operation_id": id})
	}
}

// typedTake observes pending with a bounded test context and forgets only its successful SDK receipt.
// typedTake 使用有界测试上下文观察 pending，仅遗忘成功的 SDK 回执。
// submitErr is the original admission outcome; failures preserve the receipt for fixture cleanup.
// submitErr 是原始入场结果；失败时保留回执供夹具清理。
func typedTake[T any](t *testing.T, pending *EmbeddedPending[T], submitErr error) T {
	t.Helper()
	if submitErr != nil {
		t.Fatal(submitErr)
	}
	// result is a typed native acknowledgement, never an inferred completion.
	// result 是类型化原生确认，绝非推断完成。
	result, err := pending.Result(driverTestContext(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := pending.Forget(); err != nil {
		t.Fatal(err)
	}
	return result
}

// nativeTypedRuntime uses the public typed API and returns the runtime, package root and explicit cleanup-transfer function.
// nativeTypedRuntime 使用公开类型化 API，返回运行时、包根目录及显式清理转移函数。
// Cleanup drains the exact runtime before the borrowed driver and transport are closed by earlier fixtures.
// 清理先排空精确运行时，再由较早夹具关闭借用驱动器及传输。
func nativeTypedRuntime(t *testing.T) (*EmbeddedRuntime, string, func()) {
	t.Helper()
	return nativeTypedRuntimeWithPersistence(t, false)
}

// nativeTypedRuntimeWithPersistence selects explicit storage before construction and retains the same cleanup authority.
// nativeTypedRuntimeWithPersistence 在构造前选择显式存储，并保留同一清理权威。
func nativeTypedRuntimeWithPersistence(t *testing.T, persistent bool) (*EmbeddedRuntime, string, func()) {
	t.Helper()
	return nativeTypedRuntimeWithJournalLimit(t, persistent, 16)
}

// nativeTypedRuntimeWithJournalLimit constructs explicit storage with maxRecords and returns runtime, package and cleanup transfer.
// nativeTypedRuntimeWithJournalLimit 以 maxRecords 构造显式存储，返回运行时、包及清理转移函数。
func nativeTypedRuntimeWithJournalLimit(t *testing.T, persistent bool, maxRecords uint64) (*EmbeddedRuntime, string, func()) {
	t.Helper()
	// Register cleanup before native owners, then resolve the owned directory for SQLite NOFOLLOW.
	// 在原生所有者前注册清理，再为 SQLite NOFOLLOW 解析自有目录的物理路径。
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// transport, driver and client establish separate native and SDK ownership boundaries.
	// transport、driver 和 client 建立独立的原生与 SDK 所有权边界。
	transport := nativeEmbeddedTest(t)
	driver := nativeDriverTest(t, transport, EmbeddedDriverConfig{1, 2, 2})
	client, err := NewEmbeddedClient(driver)
	if err != nil {
		t.Fatal(err)
	}
	pending, err := client.Reserve(context.Background())
	runtime := typedTake(t, pending, err)
	// adopted explicitly transfers final native cleanup to a scope created by the calling test.
	// adopted 将最终原生清理显式转移给调用测试创建的作用域。
	adopted := false
	t.Cleanup(func() {
		if adopted {
			return
		}
		embeddedRequest(t, transport, map[string]any{"type": "runtime_close", "runtime_id": runtime.RuntimeID()})
		embeddedPoll(t, func() any {
			return embeddedRequest(t, transport, map[string]any{"type": "runtime_status", "runtime_id": runtime.RuntimeID()})
		}, func(value any) bool { return value.(map[string]any)["closed"] == true })
		embeddedRequest(t, transport, map[string]any{"type": "runtime_free", "runtime_id": runtime.RuntimeID()})
	})
	// root is isolated from user configuration; the formal package is explicitly authorized outside System.
	// root 与用户配置隔离；正式包在 System 外被显式授权。
	system := filepath.Join(root, "system_lua_lib")
	// All typed client and persistent scenarios consume this exact external package through the real core.
	// 所有类型化客户端及持久场景通过真实核心消费此精确外部包。
	packageRoot := filepath.Join(root, "plugin-generations", "go-typed-test")
	if err := os.MkdirAll(packageRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packageRoot, "dependencies.yaml"), []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// options use generated fields directly, including outer presence pointers for nullable paths.
	// options 直接使用生成字段，包含可空路径的外层存在性指针。
	rootPointer, systemPointer := &root, &system
	options := EmbeddedInputLuaEngineOptions{HostOptions: EmbeddedInputLuaRuntimeHostOptions{RuntimeRoot: &rootPointer, SystemLuaLibDir: &systemPointer, ReservedEntryNames: EmbeddedInputLuaRuntimeHostOptionsReservedEntryNames{}, AllowNetworkDownload: false}, PoolConfig: EmbeddedInputLuaVmPoolConfig{MinSize: 0, MaxSize: 2, IdleTtlSecs: 60}}
	budgets := EmbeddedInputEmbeddedRuntimeConfig{MaxRegisteredPlugins: 4, MaxRegisteredPools: 4, MaxSessions: 4, MaxRegisteredCapabilities: 4, MaxResidentVms: 2, MaxRunningCalls: 2, MaxQueuedCalls: 4, MaxQueuedBytes: 4096, MaxOperations: 16, MaxEffectRecordsPerOperation: 8, MaxEffectBytesPerOperation: 8192, MaxHostRequests: 4, MaxHostRequestBytes: 8192, MaxValueBytes: 1024}
	// The selected initialization route is explicit; persistent failures never fall back to memory mode.
	// 初始化路径显式选择；持久初始化失败绝不回退到内存模式。
	var initializing *EmbeddedPending[EmbeddedOutputRuntimeReceipt]
	if persistent {
		initializing, err = runtime.InitializePersistent(context.Background(), options, budgets, EmbeddedInputRuntimePersistenceConfig{Path: filepath.Join(root, "operations.db"), Journal: EmbeddedInputOperationJournalConfig{MaxRecords: maxRecords, MaxRecordBytes: 32768, MaxDatabaseBytes: 262144}, Worker: EmbeddedInputOperationJournalWorkerConfig{MaxPendingWrites: 8, MaxPendingBytes: 131072}})
	} else {
		initializing, err = runtime.Initialize(context.Background(), options, budgets)
	}
	typedTake(t, initializing, err)
	status, err := runtime.Status(context.Background())
	// initialized preserves the actual completed construction result and its original retained error.
	// initialized 保留实际已完成构造的结果及原始保留错误。
	initialized := typedTake(t, status, err)
	if persistent {
		t.Logf("persistent runtime initialization: root=%s phase=%s error=%+v", root, initialized.Initialization, initialized.Error)
	}
	if initialized.Initialization != EmbeddedOutputInitializationPhaseReady {
		t.Fatalf("typed runtime failed to initialize: phase=%s error=%+v", initialized.Initialization, initialized.Error)
	}
	registration, err := runtime.RegisterPlugin(context.Background(), "go-typed-test", EmbeddedInputEmbeddedPluginConfig{MaxRegisteredPools: budgets.MaxRegisteredPools, MaxSessions: budgets.MaxSessions, MaxResidentVms: budgets.MaxResidentVms, MaxRunningCalls: budgets.MaxRunningCalls, MaxQueuedCalls: budgets.MaxQueuedCalls, MaxQueuedBytes: budgets.MaxQueuedBytes, MaxOperations: budgets.MaxOperations})
	typedTake(t, registration, err)
	return runtime, packageRoot, func() { adopted = true }
}

// typedPool registers source with explicit shared or dedicated-session policy and returns its acknowledged handle.
// typedPool 以显式公共或专用会话策略注册 source，并返回已确认句柄。
func typedPool(t *testing.T, runtime *EmbeddedRuntime, root, source string, session bool) *EmbeddedPool {
	t.Helper()
	// policy fixes the execution domain; session mode reserves one non-lendable VM.
	// policy 固定执行域；会话模式预留一个不可出借 VM。
	policy := EmbeddedInputPluginPoolConfig{Kind: EmbeddedInputPoolKindShared, MaxResidentVms: 2, MaxRunningCalls: 2, MaxQueuedCalls: 4, Reuse: EmbeddedInputInstanceReuseReusable, Backend: EmbeddedInputExecutionBackendInProcess}
	if session {
		policy.Kind, policy.Reuse, policy.MinResidentVms = EmbeddedInputPoolKindDedicated, EmbeddedInputInstanceReuseSession, 1
	}
	definition := EmbeddedInputModuleDefinition{PluginId: "go-typed-test", Generation: "typed-generation-1", PackageRoot: root, DependenciesFile: "dependencies.yaml", Mounts: map[string]any{}, SecurityPartition: "go-typed-test", Source: source, Exports: EmbeddedInputModuleDefinitionExports{{Name: "call", InputSchema: true, OutputSchema: true}}}
	pending, err := runtime.RegisterPool(context.Background(), definition, policy, []string{"go.host"}, "typed-v1")
	return typedTake(t, pending, err)
}

// TestEmbeddedClientNativeCalls checks real null/boolean/decimal/text values, failures and separate native/SDK quotas.
// TestEmbeddedClientNativeCalls 检查真实空值／布尔／小数／文本、失败及独立原生／SDK 配额。
func TestEmbeddedClientNativeCalls(t *testing.T) {
	// runtime owns all call handles; each completed operation is explicitly forgotten.
	// runtime 拥有全部调用句柄；每个已完成操作显式遗忘。
	runtime, root, _ := nativeTypedRuntime(t)
	pool := typedPool(t, runtime, root, "return {call=function(a) if a=='fail' then error('expected') end; return a end}", false)
	for _, value := range []any{nil, false, json.Number("0.25"), "中文"} {
		pending, err := pool.Submit(context.Background(), "call", value, EmbeddedInputLuaInvocationContext{}, 10000)
		operation := typedTake(t, pending, err)
		finished, err := operation.Wait(driverTestContext(t))
		if err != nil || finished.Phase != EmbeddedOutputOperationPhaseSucceeded || finished.Value == nil || !reflect.DeepEqual(*finished.Value, value) {
			t.Fatalf("typed result mismatch: %#v, %v", finished, err)
		}
		// The generated module alternative must be present even though this Lua call has no host effects.
		// 即使此 Lua 调用没有宿主副作用，生成的模块分支也必须存在。
		binding, ok := finished.Context.(EmbeddedOutputOperationContextVariant2)
		if !ok || binding.Kind != EmbeddedOutputOperationContextVariant2KindModule || binding.PoolId != pool.PoolID() || binding.Caller.OperationId != operation.OperationID() || binding.Export == nil || *binding.Export != "call" || len(finished.HostEffects) != 0 {
			t.Fatalf("typed module context missing or changed: %#v", finished.Context)
		}
		forgotten, err := operation.Forget(context.Background())
		typedTake(t, forgotten, err)
	}
	pending, err := pool.Submit(context.Background(), "call", "fail", EmbeddedInputLuaInvocationContext{}, 10000)
	operation := typedTake(t, pending, err)
	finished, err := operation.Wait(driverTestContext(t))
	if err != nil || finished.Phase != EmbeddedOutputOperationPhaseFailed || finished.Error == nil || *finished.Error == nil || finished.Value != nil {
		t.Fatalf("typed failure lost: %#v, %v", finished, err)
	}
	forgotten, err := operation.Forget(context.Background())
	typedTake(t, forgotten, err)
	if len(runtime.client.Driver().Commands()) != 0 {
		t.Fatal("successful polling leaked SDK quota")
	}
	// revoked reports actual live grant removal instead of guessing from the immutable declaration.
	// revoked 报告实际实时授权移除，不根据不可变声明猜测。
	revoked, err := pool.RevokePermission(context.Background(), "go.host")
	if !typedTake(t, revoked, err) {
		t.Fatal("expected granted permission to be revoked")
	}
	closing, err := pool.RequestClose(context.Background())
	typedTake(t, closing, err)
	embeddedPoll(t, func() any { p, err := pool.Status(context.Background()); return typedTake(t, p, err) }, func(value any) bool { return value.(EmbeddedOutputPoolUsage).Resident == 0 })
	removed, err := pool.Forget(context.Background())
	typedTake(t, removed, err)
}

// TestEmbeddedClientNativeSession proves fixed VM state, independent initialization and explicit session/plugin drainage.
// TestEmbeddedClientNativeSession 证明固定 VM 状态、独立初始化及显式会话／插件排空。
func TestEmbeddedClientNativeSession(t *testing.T) {
	// opening retains both original identities from the single session reservation.
	// opening 保留单次会话预留的两个原始身份。
	runtime, root, _ := nativeTypedRuntime(t)
	pool := typedPool(t, runtime, root, "local count=0; return {call=function(a) count=count+1; return count end}", true)
	pending, err := pool.OpenSession(context.Background(), 10000)
	opening := typedTake(t, pending, err)
	initialized, err := opening.Initialization.Wait(driverTestContext(t))
	if err != nil || initialized.Phase != EmbeddedOutputOperationPhaseSucceeded {
		t.Fatalf("session initialization failed: %#v, %v", initialized, err)
	}
	// Session initialization has a bound session but no requested export.
	// 会话初始化拥有绑定会话，但没有请求导出。
	initialContext, ok := initialized.Context.(EmbeddedOutputOperationContextVariant2)
	if !ok || initialContext.Export != nil || initialContext.Caller.SessionId == nil || *initialContext.Caller.SessionId != opening.Session.SessionID() {
		t.Fatalf("typed session opening context changed: %#v", initialized.Context)
	}
	status, err := opening.Session.Status(context.Background())
	if typedTake(t, status, err).PoolId != pool.PoolID() {
		t.Fatal("session binding changed pools")
	}
	for _, expected := range []json.Number{"1", "2"} {
		call, err := opening.Session.Submit(context.Background(), "call", nil, EmbeddedInputLuaInvocationContext{}, 10000)
		operation := typedTake(t, call, err)
		result, err := operation.Wait(driverTestContext(t))
		if err != nil || result.Value == nil || *result.Value != expected {
			t.Fatalf("session state changed: %#v, %v", result, err)
		}
		// Later calls retain the pinned session while receiving their own original operation identity.
		// 后续调用保持固定会话，同时获得各自原始操作身份。
		binding, ok := result.Context.(EmbeddedOutputOperationContextVariant2)
		if !ok || binding.Caller.SessionId == nil || *binding.Caller.SessionId != opening.Session.SessionID() || binding.Caller.OperationId != operation.OperationID() || binding.Export == nil || *binding.Export != "call" {
			t.Fatalf("typed session invocation context changed: %#v", result.Context)
		}
		forgotten, err := operation.Forget(context.Background())
		typedTake(t, forgotten, err)
	}
	forgotten, err := opening.Initialization.Forget(context.Background())
	typedTake(t, forgotten, err)
	closing, err := opening.Session.RequestClose(context.Background())
	typedTake(t, closing, err)
	embeddedPoll(t, func() any { p, err := opening.Session.Status(context.Background()); return typedTake(t, p, err) }, func(value any) bool {
		return value.(EmbeddedOutputEmbeddedSessionSnapshot).Phase == EmbeddedOutputEmbeddedSessionPhaseClosed
	})
	removed, err := opening.Session.Forget(context.Background())
	typedTake(t, removed, err)
	closing, err = pool.RequestClose(context.Background())
	typedTake(t, closing, err)
	embeddedPoll(t, func() any { p, err := pool.Status(context.Background()); return typedTake(t, p, err) }, func(value any) bool { return value.(EmbeddedOutputPoolUsage).Resident == 0 })
	removed, err = pool.Forget(context.Background())
	typedTake(t, removed, err)
	plugin, err := runtime.Plugin("go-typed-test")
	if err != nil {
		t.Fatal(err)
	}
	pluginStatus, err := plugin.Status(context.Background())
	if typedTake(t, pluginStatus, err).RetainedPools != 0 {
		t.Fatal("pool metadata remained")
	}
	closing, err = plugin.RequestClose(context.Background())
	typedTake(t, closing, err)
	removed, err = plugin.Forget(context.Background())
	typedTake(t, removed, err)
}

// TestEmbeddedClientNativeRecovery proves an operation created before buffer-release failure is projected without replay.
// TestEmbeddedClientNativeRecovery 证明缓冲释放失败前创建的操作可以不重放地投影恢复。
func TestEmbeddedClientNativeRecovery(t *testing.T) {
	// original owns real native buffers; only the first post-fixture release is rejected.
	// original 拥有真实原生缓冲；仅拒绝夹具建立后的首次释放。
	runtime, root, _ := nativeTypedRuntime(t)
	pool := typedPool(t, runtime, root, "return {call=function(a) return a end}", false)
	transport := runtime.client.driver.transport
	original := transport.native
	var rejected atomic.Bool
	// requests counts actual native entries through recovery, before subsequent status polling starts.
	// requests 计数恢复期间的实际原生入口，计数截止于之后状态轮询开始前。
	var requests atomic.Int32
	counted := driverRequestNative{embeddedNative: original, requestFn: func(id uint64, frame []byte) (embeddedResult, EmbeddedNativeStatus) {
		requests.Add(1)
		return original.request(id, frame)
	}}
	transport.native = &controlledEmbeddedNative{embeddedNative: counted, releaseFn: func(id uint64, result embeddedResult) EmbeddedNativeStatus {
		if rejected.CompareAndSwap(false, true) {
			return EmbeddedNativeBusy
		}
		return original.release(id, result)
	}}
	pending, err := pool.Submit(context.Background(), "call", "once", EmbeddedInputLuaInvocationContext{}, 10000)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pending.Result(driverTestContext(t))
	var release *EmbeddedResultReleaseError
	if !errors.As(err, &release) {
		t.Fatalf("expected release error: %v", err)
	}
	operation, err := pending.DeliveredResult()
	if err != nil {
		t.Fatal(err)
	}
	second, err := pending.DeliveredResult()
	if err != nil || second.OperationID() != operation.OperationID() {
		t.Fatal("recovery changed operation identity")
	}
	if err := runtime.client.driver.ReleaseResults(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := pending.Forget(); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 1 {
		t.Fatal("recovery replayed submission")
	}
	finished, err := operation.Wait(driverTestContext(t))
	if err != nil || finished.Value == nil || *finished.Value != "once" {
		t.Fatalf("lost created operation: %#v, %v", finished, err)
	}
	forgotten, err := operation.Forget(context.Background())
	typedTake(t, forgotten, err)
}

// TestEmbeddedClientNativeCancelledWait preserves actual operation lifetime and late committed callback evidence.
// TestEmbeddedClientNativeCancelledWait 保留实际操作寿命及迟到的已提交回调证据。
func TestEmbeddedClientNativeCancelledWait(t *testing.T) {
	// started and finish keep the host function alive beyond a cancelled observer and cancellation request.
	// started 和 finish 使宿主函数跨越已取消观察及取消请求继续存活。
	runtime, root, _ := nativeTypedRuntime(t)
	pump := nativePumpTest(t, runtime.client.driver.transport, runtime.RuntimeID(), EmbeddedCallbackPumpConfig{1, 2, 1})
	started, finish := make(chan struct{}), make(chan struct{})
	// release permits cleanup to signal the blocked handler even when an assertion fails.
	// release 允许清理在断言失败时仍向阻塞处理器发信号。
	var release sync.Once
	defer release.Do(func() { close(finish) })
	pumpRegister(t, pump, pumpCapability(func(arguments any, callback *EmbeddedHostCallbackContext) (any, error) {
		close(started)
		<-finish
		if err := callback.ReportEffects(EmbeddedInputEffectStateCommitted); err != nil {
			return nil, err
		}
		return arguments, nil
	}))
	// Pool registration captures the current capability snapshot; later publications do not rewrite the binding.
	// 池注册捕获当前能力快照；后续发布不改写绑定。
	pool := typedPool(t, runtime, root, "return {call=function(a) return vulcan.capabilities.call('go.callback',a) end}", false)
	pending, err := pool.Submit(context.Background(), "call", "late", EmbeddedInputLuaInvocationContext{}, 10000)
	operation := typedTake(t, pending, err)
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		status, statusErr := operation.Status(context.Background())
		snapshot := typedTake(t, status, statusErr)
		t.Fatalf("callback did not start; pump=%+v operation=%+v response=%s", pump.Status(), snapshot, status.Receipt().ResponseBytes())
	}
	// observer cancellation is deliberately independent of the later explicit native cancellation.
	// 观察者取消刻意独立于之后的显式原生取消。
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := operation.Wait(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("observer cancellation lost: %v", err)
	}
	status, err := operation.Status(context.Background())
	if typedTake(t, status, err).CancellationRequested {
		t.Fatal("observer cancelled native execution")
	}
	cancelling, err := operation.Cancel(context.Background())
	if !typedTake(t, cancelling, err) {
		t.Fatal("native cancellation was not accepted")
	}
	// The real host completion, not observer cancellation, returns the handler capacity.
	// 实际宿主完成归还处理器容量，而非观察者取消。
	release.Do(func() { close(finish) })
	finished, err := operation.Wait(driverTestContext(t))
	if err != nil || finished.Phase != EmbeddedOutputOperationPhaseCancelled {
		t.Fatalf("cancelled terminal missing: %#v, %v", finished, err)
	}
	committed := false
	for _, effect := range finished.HostEffects {
		committed = committed || effect.Effects == EmbeddedOutputEffectStateCommitted
	}
	if !committed {
		t.Fatal("late committed callback evidence disappeared")
	}
	forgotten, err := operation.Forget(context.Background())
	typedTake(t, forgotten, err)
}

// TestEmbeddedClientNativeLifecycle checks the exact reserved-slot lifecycle and separates handles from core namespaces.
// TestEmbeddedClientNativeLifecycle 检查精确预留槽生命周期，并区分句柄与核心命名空间。
func TestEmbeddedClientNativeLifecycle(t *testing.T) {
	// client borrows a live driver; the transport fixture owns final root release.
	// client 借用活动驱动器；传输夹具拥有最终根释放。
	transport := nativeEmbeddedTest(t)
	driver := nativeDriverTest(t, transport, EmbeddedDriverConfig{1, 1, 1})
	client, err := NewEmbeddedClient(driver)
	if err != nil {
		t.Fatal(err)
	}
	description, err := client.Describe(context.Background())
	typedTake(t, description, err)
	pending, err := client.Reserve(context.Background())
	runtime := typedTake(t, pending, err)
	// freed prevents failure cleanup from replaying an already acknowledged slot removal.
	// freed 防止失败清理重放已经确认的槽移除。
	freed := false
	t.Cleanup(func() {
		if !freed {
			embeddedRequest(t, transport, map[string]any{"type": "runtime_close", "runtime_id": runtime.RuntimeID()})
			embeddedRequest(t, transport, map[string]any{"type": "runtime_free", "runtime_id": runtime.RuntimeID()})
		}
	})
	status, err := runtime.Status(context.Background())
	snapshot := typedTake(t, status, err)
	if snapshot.RuntimeId != runtime.RuntimeID() || snapshot.CoreRuntimeId != nil || snapshot.Initialization != EmbeddedOutputInitializationPhaseReserved {
		t.Fatalf("reserved slot gained a fabricated core identity: %#v", snapshot)
	}
	premature, err := runtime.Free(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_, err = premature.Result(driverTestContext(t))
	var failure *EmbeddedRuntimeError
	if !errors.As(err, &failure) || failure.Code != "busy" {
		t.Fatalf("premature removal was not rejected: %v", err)
	}
	if err := premature.Forget(); err != nil {
		t.Fatal(err)
	}
	closing, err := runtime.RequestClose(context.Background())
	if typedTake(t, closing, err).RuntimeId != runtime.RuntimeID() {
		t.Fatal("closure changed slot identity")
	}
	status, err = runtime.Status(context.Background())
	if !typedTake(t, status, err).Closed {
		t.Fatal("metadata-only slot did not close")
	}
	removing, err := runtime.Free(context.Background())
	if typedTake(t, removing, err).RuntimeId != runtime.RuntimeID() {
		t.Fatal("removal changed identity")
	}
	freed = true
	status, err = runtime.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_, err = status.Result(driverTestContext(t))
	if !errors.As(err, &failure) || failure.Code != "not_found" {
		t.Fatalf("removed handle silently rebound: %v", err)
	}
	if err := status.Forget(); err != nil {
		t.Fatal(err)
	}
}
