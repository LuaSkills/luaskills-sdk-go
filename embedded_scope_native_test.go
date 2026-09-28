//go:build cgo

package luaskills

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// nativeScopeTest adopts exact existing owners once native polling reaches a quiescent boundary, then transfers fixture cleanup.
// nativeScopeTest 在原生轮询到达静止边界时接管精确现有所有者，随后转移夹具清理。
func nativeScopeTest(t *testing.T, runtime *EmbeddedRuntime, pump *EmbeddedCallbackPump, adopt func()) *EmbeddedRuntimeScope {
	t.Helper()
	// deadline bounds only adoption retries caused by the independently polling existing pump.
	// deadline 仅限制由独立轮询现有泵导致的接管重试。
	deadline := time.Now().Add(5 * time.Second)
	for {
		scope, err := NewEmbeddedRuntimeScope(runtime, pump, time.Millisecond)
		if err == nil {
			adopt()
			t.Cleanup(func() {
				err := scope.Close(driverTestContext(t))
				if err != nil && scope.Status().Retryable {
					err = scope.RetryClose(driverTestContext(t))
				}
				if err != nil {
					t.Error(err)
				}
			})
			return scope
		}
		var busy *EmbeddedRuntimeError
		if !errors.As(err, &busy) || busy.Code != "busy" || time.Now().After(deadline) {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
	}
}

// TestEmbeddedScopeNativeIndependentControl proves full driver quotas cannot consume the scope's reserved cleanup lane.
// TestEmbeddedScopeNativeIndependentControl 证明驱动器配额耗尽不能占用作用域预留清理通道。
func TestEmbeddedScopeNativeIndependentControl(t *testing.T) {
	// scope owns this slot but continues borrowing the ordinary driver and transport.
	// scope 拥有此槽，同时继续借用普通驱动器及传输。
	runtime, _, adopt := nativeTypedRuntime(t)
	scope := nativeScopeTest(t, runtime, nil, adopt)
	if _, err := NewEmbeddedRuntimeScope(runtime, nil, time.Millisecond); err == nil {
		t.Fatal("duplicate scope accepted")
	}
	if _, err := NewEmbeddedCallbackPump(runtime.client.driver.transport, runtime.RuntimeID(), EmbeddedCallbackPumpConfig{1, 1, 1}); err == nil {
		t.Fatal("late pump bypassed scope adoption")
	}
	if _, err := runtime.Free(context.Background()); err == nil {
		t.Fatal("ordinary handle bypassed scoped removal")
	}
	// first and second deliberately retain completed control receipts until after scope closure.
	// first 和 second 刻意保留已完成控制回执，直到作用域关闭之后。
	first, err := runtime.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Result(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	second, err := runtime.client.Describe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := second.Result(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Status(context.Background()); err == nil {
		t.Fatal("fixture did not exhaust driver quota")
	}
	if err := scope.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	if scope.Status().Phase != EmbeddedScopeClosed || runtime.client.driver.Status().Closed {
		t.Fatal("scope closed its borrowed driver or failed to remove the slot")
	}
	if err := first.Forget(); err != nil {
		t.Fatal(err)
	}
	if err := second.Forget(); err != nil {
		t.Fatal(err)
	}
	// other proves the shared transport still accepts a separate runtime after this scope releases its exact slot.
	// other 证明此作用域释放精确槽后，共享传输仍接纳其他运行时。
	reserved, err := runtime.client.Reserve(context.Background())
	other := typedTake(t, reserved, err)
	otherScope := nativeScopeTest(t, other, nil, func() {})
	if err := otherScope.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
}

// TestEmbeddedScopeNativeClosedDriver proves scope cleanup survives actual ordinary-worker exit.
// TestEmbeddedScopeNativeClosedDriver 证明作用域清理跨越普通工作位实际退出继续有效。
func TestEmbeddedScopeNativeClosedDriver(t *testing.T) {
	runtime, _, adopt := nativeTypedRuntime(t)
	scope := nativeScopeTest(t, runtime, nil, adopt)
	if err := runtime.client.driver.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	if runtime.client.driver.transport.Free() == nil {
		t.Fatal("transport ignored the retained runtime scope")
	}
	if err := scope.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	if scope.Status().Phase != EmbeddedScopeClosed {
		t.Fatal("independent scope did not finish")
	}
}

// TestEmbeddedScopeNativeCallbackDrain preserves actual handler ownership after cancelled and timed-out close observers.
// TestEmbeddedScopeNativeCallbackDrain 在关闭观察者取消和超时后保留实际处理器所有权。
func TestEmbeddedScopeNativeCallbackDrain(t *testing.T) {
	// scope is assigned before the call is published; the callback receives it only through the synchronized handler queue.
	// scope 在调用发布前赋值；回调仅通过同步处理器队列取得它。
	runtime, root, adopt := nativeTypedRuntime(t)
	pump := nativePumpTest(t, runtime.client.driver.transport, runtime.RuntimeID(), EmbeddedCallbackPumpConfig{1, 1, 1})
	var scope *EmbeddedRuntimeScope
	entered, allow := make(chan struct{}), make(chan struct{})
	var release sync.Once
	defer release.Do(func() { close(allow) })
	var guarded atomic.Bool
	var calls atomic.Int32
	pumpRegister(t, pump, pumpCapability(func(value any, callback *EmbeddedHostCallbackContext) (any, error) {
		guarded.Store(scope.Close(callback) != nil)
		if calls.Add(1) == 1 {
			close(entered)
		}
		<-allow
		return value, callback.ReportEffects(EmbeddedInputEffectStateCommitted)
	}))
	closing := &EmbeddedInputModuleFinalizer{Export: "shutdown", Arguments: nil, TimeoutMs: 5000}
	definition := EmbeddedInputModuleDefinition{
		PluginId: "go-typed-test", Generation: "typed-generation-1", PackageRoot: root,
		DependenciesFile: "dependencies.yaml", Mounts: map[string]any{}, SecurityPartition: "go-typed-test",
		Source:  "return {call=function(a) return vulcan.capabilities.call('go.callback',a) end, shutdown=function() local r=vulcan.capabilities.call('go.callback','closing'); assert(r.ok); return r.value end}",
		Exports: EmbeddedInputModuleDefinitionExports{{Name: "call", InputSchema: true, OutputSchema: true}, {Name: "shutdown", InputSchema: true, OutputSchema: true}}, Finalizer: &closing,
	}
	policy := EmbeddedInputPluginPoolConfig{Kind: EmbeddedInputPoolKindShared, MaxResidentVms: 2, MaxRunningCalls: 2, MaxQueuedCalls: 4, Reuse: EmbeddedInputInstanceReuseSingleCall, Backend: EmbeddedInputExecutionBackendInProcess}
	registered, registerErr := runtime.RegisterPool(context.Background(), definition, policy, []string{"go.host"}, "typed-v1")
	pool := typedTake(t, registered, registerErr)
	scope = nativeScopeTest(t, runtime, pump, adopt)
	pending, err := pool.Submit(context.Background(), "call", nil, EmbeddedInputLuaInvocationContext{}, 10000)
	typedTake(t, pending, err)
	select {
	case <-entered:
	case <-driverTestContext(t).Done():
		t.Fatal("callback did not enter")
	}
	if !guarded.Load() || scope.Status().Phase != EmbeddedScopeOpen {
		t.Fatal("callback context started a nested lifecycle wait")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := scope.Close(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled observer changed: %v", err)
	}
	embeddedPoll(t, func() any { return scope.Status() }, func(value any) bool { return value.(EmbeddedScopeStatus).Phase == EmbeddedScopeDrainingRuntime })
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := scope.Close(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("blocked handler did not retain scope: %v", err)
	}
	if !scope.Status().Running || pump.Status().Closed || runtime.client.driver.transport.Free() == nil {
		t.Fatal("observer exit released actual handler ownership")
	}
	release.Do(func() { close(allow) })
	if err := scope.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	if !pump.Status().Closed || scope.Status().Phase != EmbeddedScopeClosed {
		t.Fatal("scope closed before actual callback drainage")
	}
	if calls.Load() != 2 {
		t.Fatal("scope closed the pump before automatic Lua finalization could call the host")
	}
}

// scopeFaultNative injects exact root-control failures while retaining real native result allocations.
// scopeFaultNative 在保留真实原生结果分配的同时注入精确根控制失败。
type scopeFaultNative struct {
	// pumpFaultNative supplies allocation-aware copied-response and release failure injection.
	// pumpFaultNative 提供感知分配的复制响应及释放故障注入。
	*pumpFaultNative
}

// request intercepts the selected root route only; all other calls and successful selected attempts use the real DLL.
// request 仅拦截所选根路由；全部其他调用及成功所选尝试使用真实 DLL。
func (n *scopeFaultNative) request(id uint64, frame []byte) (embeddedResult, EmbeddedNativeStatus) {
	// command comes from the exact decoded root union and is never guessed from nested candidate fields.
	// command 来自精确解码的根联合，绝不根据嵌套候选字段猜测。
	decoded, err := DecodeEmbeddedJSON(frame)
	if err != nil {
		panic(err)
	}
	command := decoded.(map[string]any)["command"].(map[string]any)
	matched := command["type"] == n.route
	if matched {
		n.calls.Add(1)
	}
	if matched && n.mode == "capacity" && n.fired.CompareAndSwap(false, true) {
		return embeddedResult{}, EmbeddedNativeCapacityExceeded
	}
	result, status := n.embeddedNative.request(id, frame)
	if matched && status == EmbeddedNativeOk && !n.fired.Load() {
		n.allocations.Store(result.id, true)
	}
	return result, status
}

// TestEmbeddedScopeNativeRecovery verifies exact close/status/free checkpoints across release failure and pre-entry capacity refusal.
// TestEmbeddedScopeNativeRecovery 验证释放失败及入口前容量拒绝时精确关闭／状态／释放检查点。
func TestEmbeddedScopeNativeRecovery(t *testing.T) {
	for _, route := range []string{"runtime_close", "runtime_status", "runtime_free"} {
		for _, mode := range []string{"release", "capacity"} {
			t.Run(route+"/"+mode, func(t *testing.T) {
				// backend injects after fixture initialization, before any scope control enters the adapter.
				// backend 在夹具初始化后、任何作用域控制进入适配器前注入。
				runtime, _, adopt := nativeTypedRuntime(t)
				transport := runtime.client.driver.transport
				backend := &scopeFaultNative{&pumpFaultNative{embeddedNative: transport.native, route: route, mode: mode}}
				transport.native = backend
				scope := nativeScopeTest(t, runtime, nil, adopt)
				if err := scope.Close(driverTestContext(t)); err == nil {
					t.Fatal("injected failure reported closure")
				}
				before := backend.calls.Load()
				if !scope.Status().Retryable || transport.Free() == nil {
					t.Fatal("failed scope lost recovery ownership")
				}
				if err := scope.Close(driverTestContext(t)); err == nil || backend.calls.Load() != before {
					t.Fatal("ordinary close replayed a failed attempt")
				}
				if err := scope.RetryClose(driverTestContext(t)); err != nil {
					t.Fatal(err)
				}
				if scope.Status().Phase != EmbeddedScopeClosed || transport.RetainedResults() != 0 {
					t.Fatal("recovered scope failed to settle native ownership")
				}
				if route != "runtime_status" {
					expected := uint64(1)
					if mode == "capacity" {
						expected = 2
					}
					if backend.calls.Load() != expected {
						t.Fatalf("mutation replay: %s %s: %d", route, mode, backend.calls.Load())
					}
				}
			})
		}
	}
}

// TestEmbeddedScopeNativePumpRecovery requires explicit recovery of the adopted pump before final slot removal.
// TestEmbeddedScopeNativePumpRecovery 在最终槽移除前要求显式恢复已接管泵。
func TestEmbeddedScopeNativePumpRecovery(t *testing.T) {
	// backend owns the real completion buffer that fails its first release.
	// backend 拥有首次释放失败的真实完成缓冲。
	runtime, root, adopt := nativeTypedRuntime(t)
	transport := runtime.client.driver.transport
	backend := &pumpFaultNative{embeddedNative: transport.native, route: "host_request_complete", mode: "release"}
	transport.native = backend
	pump := nativePumpTest(t, transport, runtime.RuntimeID(), EmbeddedCallbackPumpConfig{1, 1, 1})
	var handlers atomic.Uint64
	pumpRegister(t, pump, pumpCapability(func(value any, callback *EmbeddedHostCallbackContext) (any, error) {
		handlers.Add(1)
		return value, callback.ReportEffects(EmbeddedInputEffectStateCommitted)
	}))
	pool := typedPool(t, runtime, root, "return {call=function(a) return vulcan.capabilities.call('go.callback',a) end}", false)
	scope := nativeScopeTest(t, runtime, pump, adopt)
	backend.armed.Store(true)
	pending, err := pool.Submit(context.Background(), "call", nil, EmbeddedInputLuaInvocationContext{}, 10000)
	typedTake(t, pending, err)
	embeddedPoll(t, func() any { return pump.Status() }, func(value any) bool { return value.(EmbeddedCallbackPumpStatus).RecoveryRequired })
	if err := scope.Close(driverTestContext(t)); err == nil || !scope.Status().Retryable {
		t.Fatal("scope concealed required callback recovery")
	}
	if err := scope.RetryClose(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	if handlers.Load() != 1 || backend.calls.Load() != 1 || !pump.Status().Closed {
		t.Fatal("scope recovery replayed completion or failed to drain the pump")
	}
}
