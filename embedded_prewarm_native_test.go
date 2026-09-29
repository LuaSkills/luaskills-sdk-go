//go:build cgo

package luaskills

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEmbeddedPrewarmNative proves typed prewarming allocates distinct VMs without invoking business exports.
// TestEmbeddedPrewarmNative 证明类型化预热分配不同 VM，且不调用业务导出。
func TestEmbeddedPrewarmNative(t *testing.T) {
	// The fixture supplies exact native ownership and one bounded reusable pool.
	// 夹具提供精确原生归属及单个有界可复用池。
	runtime, root, _ := nativeTypedRuntime(t)
	pool := typedPool(t, runtime, root, "local count=0; return {call=function() count=count+1; return count end}", false)
	// Cold observation must preserve exact authority and allocate no instance.
	// 冷观测必须保留精确权威，且不分配实例。
	initial := reusableReadiness(t, pool)
	if initial.PoolId != pool.PoolID() || initial.Ready != 0 || initial.Physical.Resident != 0 || initial.MaxResidentVms != 2 {
		t.Fatalf("cold readiness changed pool state: %#v", initial)
	}
	instances := make(map[string]bool)
	for range 2 {
		// Delivery acknowledges only the operation; initialization is observed independently.
		// 交付仅确认操作；初始化独立观察。
		pending, err := pool.PrewarmInstance(context.Background(), EmbeddedInputLuaInvocationContext{}, 5000)
		if err != nil {
			t.Fatal(err)
		}
		if pending.Receipt().Lane() != EmbeddedWorkLane {
			t.Fatal("prewarm must not consume the reserved control lane")
		}
		operation := typedTake(t, pending, err)
		result, err := operation.Wait(driverTestContext(t))
		if err != nil || result.Phase != EmbeddedOutputOperationPhaseSucceeded || result.Value == nil {
			t.Fatalf("prewarm failed: %#v, %v", result, err)
		}
		binding, ok := result.Context.(EmbeddedOutputOperationContextVariant2)
		if !ok || binding.Prewarm == nil || !*binding.Prewarm || binding.Export != nil || binding.PoolId != pool.PoolID() {
			t.Fatalf("prewarm context lost: %#v", result.Context)
		}
		value, ok := (*result.Value).(map[string]any)
		if !ok {
			t.Fatalf("prewarm result is not an object: %#v", result.Value)
		}
		identity, ok := value["instance_id"].(string)
		if !ok || identity == "" || instances[identity] {
			t.Fatalf("prewarm reused an instance or omitted its identity: %#v", value)
		}
		instances[identity] = true
		forgotten, err := operation.Forget(context.Background())
		typedTake(t, forgotten, err)
		if snapshot := reusableReadiness(t, pool); snapshot.Ready != uint64(len(instances)) {
			t.Fatalf("confirmed readiness lost initialized instances: %#v", snapshot)
		}
	}
	status, err := pool.Status(context.Background())
	if typedTake(t, status, err).Resident != uint64(len(instances)) {
		t.Fatal("physical occupancy does not match distinct initialized instances")
	}
	// A failed additional allocation must not block later business reuse of confirmed warm state.
	// 额外分配失败不能阻塞后续业务复用已确认预热状态。
	pending, err := pool.PrewarmInstance(context.Background(), EmbeddedInputLuaInvocationContext{}, 5000)
	rejected := typedTake(t, pending, err)
	failure, err := rejected.Wait(driverTestContext(t))
	if err != nil || failure.Phase != EmbeddedOutputOperationPhaseFailed || failure.Error == nil || *failure.Error == nil || (*failure.Error).Code != EmbeddedOutputEmbeddedErrorCodeCapacityExceeded {
		t.Fatalf("full pool failure was lost: %#v, %v", failure, err)
	}
	forgotten, err := rejected.Forget(context.Background())
	typedTake(t, forgotten, err)
	for _, count := range []json.Number{"1", "2"} {
		pending, err := pool.Submit(context.Background(), "call", nil, EmbeddedInputLuaInvocationContext{}, 5000)
		business := typedTake(t, pending, err)
		result, err := business.Wait(driverTestContext(t))
		if err != nil || result.Value == nil || *result.Value != count {
			t.Fatalf("prewarm invoked business or failed reuse: %#v, %v", result, err)
		}
		forgotten, err := business.Forget(context.Background())
		typedTake(t, forgotten, err)
	}
	closing, err := pool.RequestClose(context.Background())
	typedTake(t, closing, err)
	if retired := reusableReadiness(t, pool); !retired.Closing || retired.Ready != 0 || retired.PoolId != pool.PoolID() {
		t.Fatalf("closed readiness lost original authority: %#v", retired)
	}
	if len(runtime.client.Driver().Commands()) != 0 {
		t.Fatal("prewarm leaked command receipt ownership")
	}
}

// TestEmbeddedPrewarmNativeScope retains an actual initializer callback after its close observer times out.
// TestEmbeddedPrewarmNativeScope 在关闭观察者超时后保留真实初始化回调。
func TestEmbeddedPrewarmNativeScope(t *testing.T) {
	// Keep the pump and exact native slot under one existing lifecycle scope.
	// 将泵与精确原生槽置于一个既有生命周期作用域之下。
	runtime, root, adopt := nativeTypedRuntime(t)
	pump := nativePumpTest(t, runtime.client.driver.transport, runtime.RuntimeID(), EmbeddedCallbackPumpConfig{1, 1, 1})
	entered, allow := make(chan struct{}), make(chan struct{})
	var release sync.Once
	defer release.Do(func() { close(allow) })
	var calls atomic.Int32
	pumpRegister(t, pump, pumpCapability(func(value any, callback *EmbeddedHostCallbackContext) (any, error) {
		calls.Add(1)
		close(entered)
		<-allow
		return value, callback.ReportEffects(EmbeddedInputEffectStateCommitted)
	}))
	pool := typedPool(t, runtime, root, "assert(vulcan.host.call('go.callback','initialization').ok); return {call=function() error('business must not execute') end}", false)
	scope := nativeScopeTest(t, runtime, pump, adopt)
	pending, err := pool.PrewarmInstance(context.Background(), EmbeddedInputLuaInvocationContext{}, 10000)
	operation := typedTake(t, pending, err)
	select {
	case <-entered:
	case <-driverTestContext(t).Done():
		t.Fatal("prewarm initializer callback did not enter")
	}
	if initializing := reusableReadiness(t, pool); initializing.Ready != 0 || initializing.Unavailable != 1 || initializing.Physical.Resident != 1 || initializing.PoolId != pool.PoolID() {
		t.Fatalf("initializer ownership was mistaken for readiness: %#v", initializing)
	}
	// An observer deadline cannot release the native VM or the still-running language callback.
	// 观察截止时间不能释放原生 VM 或仍在运行的语言回调。
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := scope.Close(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("prewarm close did not retain callback ownership: %v", err)
	}
	if !scope.Status().Running || pump.Status().Closed {
		t.Fatal("timed-out prewarm observer released actual owners")
	}
	status, err := pool.Status(context.Background())
	if typedTake(t, status, err).Resident != 1 {
		t.Fatal("prewarm VM disappeared while its callback was active")
	}
	if draining := reusableReadiness(t, pool); !draining.Closing || draining.Ready != 0 || draining.Physical.Resident != 1 {
		t.Fatalf("closing callback ownership was lost: %#v", draining)
	}
	opStatus, err := operation.Status(context.Background())
	snapshot := typedTake(t, opStatus, err)
	binding, ok := snapshot.Context.(EmbeddedOutputOperationContextVariant2)
	if !ok || binding.Prewarm == nil || !*binding.Prewarm {
		t.Fatalf("retained prewarm authority disappeared: %#v", snapshot.Context)
	}
	release.Do(func() { close(allow) })
	if err := scope.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	if scope.Status().Phase != EmbeddedScopeClosed || !pump.Status().Closed || calls.Load() != 1 {
		t.Fatal("prewarm scope failed to drain or replayed initialization")
	}
}

// reusableReadiness consumes one exact pool snapshot through the reserved control lane and returns its value.
// reusableReadiness 通过预留控制通道消费一个精确池快照并返回其值。
// t owns assertion failures and pool supplies the original native identity; no pool is created or replaced.
// t 拥有断言失败，pool 提供原始原生身份；不创建或替换池。
func reusableReadiness(t *testing.T, pool *EmbeddedPool) EmbeddedOutputEmbeddedReusablePoolSnapshot {
	t.Helper()
	// The receipt must stay in the control lane independently of native work saturation.
	// 回执必须留在控制通道，独立于原生工作饱和状态。
	pending, err := pool.ReusableStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if pending.Receipt().Lane() != EmbeddedControlLane {
		t.Fatal("readiness consumed the work lane")
	}
	return typedTake(t, pending, err)
}

// TestEmbeddedReusableReadinessErrors preserves invalid-kind and unknown-identity errors without running source.
// TestEmbeddedReusableReadinessErrors 保留无效类型及未知身份错误，且不运行源码。
// t owns fixture cleanup and failures; the test returns after every failed receipt has been explicitly forgotten.
// t 拥有夹具清理与失败；测试在显式遗忘每个失败回执后返回。
func TestEmbeddedReusableReadinessErrors(t *testing.T) {
	// A registered fixed-session module is deliberately unsuitable for reusable readiness.
	// 已登记固定会话模块故意不适用于可复用就绪查询。
	runtime, root, _ := nativeTypedRuntime(t)
	session := typedPool(t, runtime, root, "error('query must not execute source')", true)
	// Unknown binds one explicit absent identity without probing alternative pools.
	// unknown 绑定一个明确缺失身份，不探测其他池。
	unknown, err := runtime.Pool("unknown-readiness-pool")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		// pool preserves the exact handle whose native error is expected.
		// pool 保留预期原生错误的精确句柄。
		pool *EmbeddedPool
		// code records the declared native error, never a replacement state.
		// code 记录声明的原生错误，绝不是替代状态。
		code string
	}{{session, "invalid_argument"}, {unknown, "not_found"}}
	for _, item := range cases {
		pending, err := item.pool.ReusableStatus(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		_, err = pending.Result(driverTestContext(t))
		var failure *EmbeddedRuntimeError
		if !errors.As(err, &failure) || failure.Code != item.code {
			t.Fatalf("readiness error lost: %v", err)
		}
		if err := pending.Forget(); err != nil {
			t.Fatal(err)
		}
	}
	status, err := session.Status(context.Background())
	if typedTake(t, status, err).Resident != 0 || len(runtime.client.Driver().Commands()) != 0 {
		t.Fatal("readiness executed source or leaked a receipt")
	}
}
