//go:build cgo

package luaskills

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// driverTestContext bounds test observation only; product operations retain their own explicit budgets.
// driverTestContext 仅限制测试观察；产品操作保留自身显式预算。
func driverTestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// nativeDriverTest borrows a real transport and closes the driver before the caller's native cleanup runs.
// nativeDriverTest 借用真实传输，并在调用方原生清理前关闭驱动器。
func nativeDriverTest(t *testing.T, transport *EmbeddedTransport, config EmbeddedDriverConfig) *EmbeddedCommandDriver {
	t.Helper()
	driver, err := NewEmbeddedCommandDriver(transport, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := driver.Close(ctx); err != nil {
			t.Error(err)
		}
		for _, receipt := range driver.Commands() {
			if err := receipt.Forget(); err != nil {
				t.Error(err)
			}
		}
	})
	return driver
}

// driverSubmit requires an admitted generated command without waiting for native execution.
// driverSubmit 要求生成命令被接纳，不等待原生执行。
func driverSubmit(t *testing.T, driver *EmbeddedCommandDriver, command EmbeddedInputCommand) *EmbeddedCommand {
	t.Helper()
	receipt, err := driver.Submit(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	return receipt
}

// driverResult requires a successful original delivery through a bounded test observer.
// driverResult 通过有界测试观察者要求原始交付成功。
func driverResult(t *testing.T, receipt *EmbeddedCommand) any {
	t.Helper()
	value, err := receipt.Result(driverTestContext(t))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

// driverRequestNative wraps real request entry only, allowing deterministic worker scheduling without replacing C results.
// driverRequestNative 仅包装真实请求入口，允许确定性工作位调度，不替换 C 结果。
type driverRequestNative struct {
	embeddedNative
	// requestFn controls entry around the actual native backend.
	// requestFn 控制实际原生后端周围的入场。
	requestFn func(uint64, []byte) (embeddedResult, EmbeddedNativeStatus)
}

// request delegates the exact identity and frozen frame to the test's controlled actual-native boundary.
// request 将精确身份及冻结帧交给测试控制的真实原生边界。
func (n driverRequestNative) request(id uint64, frame []byte) (embeddedResult, EmbeddedNativeStatus) {
	return n.requestFn(id, frame)
}

// TestEmbeddedDriverNativeCancellation retains queued work and real mutation receipts while independent controls progress.
// TestEmbeddedDriverNativeCancellation 在独立控制推进时保留排队工作及真实变更回执。
func TestEmbeddedDriverNativeCancellation(t *testing.T) {
	transport := nativeEmbeddedTest(t)
	original := transport.native
	entered, allow := make(chan struct{}), make(chan struct{})
	var first, unblock sync.Once
	defer unblock.Do(func() { close(allow) })
	var calls atomic.Uint64
	transport.native = driverRequestNative{original, func(id uint64, frame []byte) (embeddedResult, EmbeddedNativeStatus) {
		decoded, err := DecodeEmbeddedJSON(frame)
		if err != nil {
			panic(err)
		}
		if decoded.(map[string]any)["command"].(map[string]any)["type"] == "runtime_reserve" {
			calls.Add(1)
			first.Do(func() { close(entered); <-allow })
		}
		return original.request(id, frame)
	}}
	driver := nativeDriverTest(t, transport, EmbeddedDriverConfig{1, 2, 2})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	firstReceipt, err := driver.Submit(ctx, EmbeddedInputCommandRuntimeReserve{Type: EmbeddedInputCommandRuntimeReserveTypeRuntimeReserve})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-driverTestContext(t).Done():
		t.Fatal("worker did not enter")
	}
	mutable := &EmbeddedInputCommandRuntimeReserve{Type: EmbeddedInputCommandRuntimeReserveTypeRuntimeReserve}
	second := driverSubmit(t, driver, mutable)
	mutable.Type = "describe"
	cancel()
	if _, err := firstReceipt.Result(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("observer cancellation lost: %v", err)
	}
	if firstReceipt.Done() || second.State() != EmbeddedCommandQueued || firstReceipt.Forget() == nil {
		t.Fatal("cancellation released real work")
	}
	if _, err := driver.Submit(context.Background(), EmbeddedInputCommandRuntimeReserve{Type: EmbeddedInputCommandRuntimeReserveTypeRuntimeReserve}); err == nil {
		t.Fatal("full work receipt quota accepted")
	}
	control := driverSubmit(t, driver, EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe})
	driverResult(t, control)
	if control.Lane() != EmbeddedControlLane {
		t.Fatal("control was misrouted")
	}
	if err := driver.Close(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("close observer cancellation lost: %v", err)
	}
	if transport.Free() == nil || driver.Status().Closed {
		t.Fatal("closed before real workers returned")
	}
	unblock.Do(func() { close(allow) })
	for _, receipt := range []*EmbeddedCommand{firstReceipt, second} {
		result := driverResult(t, receipt).(map[string]any)
		identity := result["runtime_id"].(string)
		result["runtime_id"] = "mutated"
		copy := receipt.ResponseBytes()
		copy[0] = '!'
		if driverResult(t, receipt).(map[string]any)["runtime_id"] != identity {
			t.Fatal("observer mutated retained receipt")
		}
		embeddedRequest(t, transport, map[string]any{"type": "runtime_close", "runtime_id": identity})
		embeddedRequest(t, transport, map[string]any{"type": "runtime_free", "runtime_id": identity})
	}
	if calls.Load() != 2 {
		t.Fatal("command was lost or replayed")
	}
	if err := driver.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	if len(driver.Commands()) != 3 || !driver.Status().Closed {
		t.Fatal("close discarded owned receipt evidence")
	}
	if !slices.Contains(LiveEmbeddedCommandDrivers(), driver) {
		t.Fatal("closed driver lost retained receipt discovery")
	}
	for _, receipt := range driver.Commands() {
		if err := receipt.Forget(); err != nil {
			t.Fatal(err)
		}
	}
	if slices.Contains(LiveEmbeddedCommandDrivers(), driver) {
		t.Fatal("fully released driver remained retained")
	}
}

// TestEmbeddedDriverNativeOwnership verifies exclusive driver claims, response reservations and exact receipt forgetting.
// TestEmbeddedDriverNativeOwnership 验证独占驱动器声明、响应预留及精确回执遗忘。
func TestEmbeddedDriverNativeOwnership(t *testing.T) {
	transport := nativeEmbeddedTest(t)
	if _, err := NewEmbeddedCommandDriver(transport, EmbeddedDriverConfig{transport.Config().MaxResultBuffers, transport.Config().MaxResultBuffers, 1}); err == nil {
		t.Fatal("insufficient control response reservation accepted")
	}
	driver := nativeDriverTest(t, transport, EmbeddedDriverConfig{1, 1, 1})
	if _, err := NewEmbeddedCommandDriver(transport, EmbeddedDriverConfig{1, 1, 1}); err == nil {
		t.Fatal("duplicate driver accepted")
	}
	receipt := driverSubmit(t, driver, EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe})
	driverResult(t, receipt)
	if _, err := driver.Submit(context.Background(), EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe}); err == nil {
		t.Fatal("completed receipt did not retain quota")
	}
	copied := *receipt
	if err := copied.Forget(); err == nil {
		t.Fatal("copied identity released original quota")
	}
	if err := receipt.Forget(); err != nil {
		t.Fatal(err)
	}
	if err := receipt.Forget(); err == nil {
		t.Fatal("duplicate forget accepted")
	}
	if transport.Free() == nil {
		t.Fatal("live idle driver lost transport ownership")
	}
	wait := EmbeddedInputCommandRuntime{Type: EmbeddedInputCommandRuntimeTypeRuntime, RuntimeId: "unused", Operation: EmbeddedInputRuntimeCommandOperationWait{Type: EmbeddedInputRuntimeCommandOperationWaitTypeOperationWait, OperationId: "unused", WaitMs: 1}}
	if _, err := driver.Submit(context.Background(), wait); err == nil {
		t.Fatal("blocking native wait accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := driver.Submit(ctx, EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe}); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled pre-admission context accepted")
	}
}

// TestEmbeddedDriverNativeReleaseRecovery pauses a real allocation owner and drains queued work only after explicit recovery.
// TestEmbeddedDriverNativeReleaseRecovery 暂停真实分配所有者，仅在显式恢复后排空排队工作。
func TestEmbeddedDriverNativeReleaseRecovery(t *testing.T) {
	transport := nativeEmbeddedTest(t)
	original := transport.native
	var rejected atomic.Bool
	transport.native = &controlledEmbeddedNative{embeddedNative: original, releaseFn: func(id uint64, result embeddedResult) EmbeddedNativeStatus {
		if rejected.CompareAndSwap(false, true) {
			return EmbeddedNativeBusy
		}
		return original.release(id, result)
	}}
	driver := nativeDriverTest(t, transport, EmbeddedDriverConfig{1, 2, 1})
	receipt := driverSubmit(t, driver, EmbeddedInputCommandRuntimeReserve{Type: EmbeddedInputCommandRuntimeReserveTypeRuntimeReserve})
	_, err := receipt.Result(driverTestContext(t))
	var release *EmbeddedResultReleaseError
	if !errors.As(err, &release) {
		t.Fatalf("release failure lost: %v", err)
	}
	release.Native.Status = EmbeddedNativeInvalidArgument
	_, err = receipt.Result(driverTestContext(t))
	if !errors.As(err, &release) || release.Native.Status != EmbeddedNativeBusy {
		t.Fatal("caller rewrote retained error evidence")
	}
	second := driverSubmit(t, driver, EmbeddedInputCommandRuntimeReserve{Type: EmbeddedInputCommandRuntimeReserveTypeRuntimeReserve})
	if driver.Status().Paused != 1 || transport.RetainedResults() != 1 || second.State() != EmbeddedCommandQueued {
		t.Fatal("failed allocation did not retain its worker reservation")
	}
	control := driverSubmit(t, driver, EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe})
	driverResult(t, control)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := driver.Close(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("close did not retain asynchronous drainage")
	}
	if err := driver.ReleaseResults(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := driver.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	if transport.RetainedResults() != 0 {
		t.Fatal("native allocation was not released")
	}
	firstValue, err := receipt.DeliveredResult()
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range []any{firstValue, driverResult(t, second)} {
		id := result.(map[string]any)["runtime_id"]
		embeddedRequest(t, transport, map[string]any{"type": "runtime_close", "runtime_id": id})
		embeddedRequest(t, transport, map[string]any{"type": "runtime_free", "runtime_id": id})
	}
}

// TestEmbeddedDriverNativeActiveRecovery rejects recovery during a real response reader without blocking the reserved control lane.
// TestEmbeddedDriverNativeActiveRecovery 在真实响应读取期间拒绝恢复，不阻塞预留控制通道。
func TestEmbeddedDriverNativeActiveRecovery(t *testing.T) {
	transport := nativeEmbeddedTest(t)
	original := transport.native
	entered, allow := make(chan struct{}), make(chan struct{})
	var blocked atomic.Bool
	var unblock sync.Once
	defer unblock.Do(func() { close(allow) })
	transport.native = &controlledEmbeddedNative{embeddedNative: original, copyFn: func(result embeddedResult, limit uint64) ([]byte, error) {
		if blocked.CompareAndSwap(false, true) {
			close(entered)
			<-allow
		}
		return original.copy(result, limit)
	}}
	driver := nativeDriverTest(t, transport, EmbeddedDriverConfig{1, 1, 1})
	receipt := driverSubmit(t, driver, EmbeddedInputCommandRuntime{Type: EmbeddedInputCommandRuntimeTypeRuntime, RuntimeId: "missing", Operation: EmbeddedInputRuntimeCommandPluginRegister{Type: EmbeddedInputRuntimeCommandPluginRegisterTypePluginRegister, PluginId: "unused"}})
	select {
	case <-entered:
	case <-driverTestContext(t).Done():
		t.Fatal("reader did not enter")
	}
	if err := driver.ReleaseResults(context.Background()); err == nil {
		t.Fatal("released during active response copying")
	}
	control := driverSubmit(t, driver, EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe})
	driverResult(t, control)
	unblock.Do(func() { close(allow) })
	_, err := receipt.Result(driverTestContext(t))
	var business *EmbeddedRuntimeError
	if !errors.As(err, &business) || business.Code != "not_found" {
		t.Fatalf("native business failure lost: %v", err)
	}
}

// TestEmbeddedDriverNativeLua proves real Lua delivery and manually acknowledged callbacks through both driver lanes.
// TestEmbeddedDriverNativeLua 通过两个驱动通道证明真实 Lua 交付及手动确认回调。
func TestEmbeddedDriverNativeLua(t *testing.T) {
	transport, runtimeID, _, pool := nativeEmbeddedRuntime(t)
	driver := nativeDriverTest(t, transport, EmbeddedDriverConfig{1, 1, 1})
	// send owns one local receipt until its original result is copied; native objects remain separately owned.
	// send 拥有一个局部回执直到原结果复制；原生对象保持单独所有权。
	send := func(operation EmbeddedInputRuntimeCommand) any {
		receipt := driverSubmit(t, driver, EmbeddedInputCommandRuntime{Type: EmbeddedInputCommandRuntimeTypeRuntime, RuntimeId: runtimeID, Operation: operation})
		result := driverResult(t, receipt)
		if err := receipt.Forget(); err != nil {
			t.Fatal(err)
		}
		return result
	}
	// terminal polls the declared nonblocking status route instead of occupying the control worker with native wait.
	// terminal 轮询声明的非阻塞状态路由，不用原生等待占用控制工作位。
	terminal := func(id string) map[string]any {
		return embeddedPoll(t, func() any {
			return send(EmbeddedInputRuntimeCommandOperationStatus{Type: EmbeddedInputRuntimeCommandOperationStatusTypeOperationStatus, OperationId: id})
		}, func(value any) bool {
			phase := value.(map[string]any)["phase"]
			return phase == "succeeded" || phase == "failed" || phase == "cancelled"
		}).(map[string]any)
	}
	poolID := pool("return {call=function(a) return a end}")
	for _, value := range []any{nil, false, json.Number("0.25"), "中文"} {
		receipt := driverSubmit(t, driver, EmbeddedInputCommandRuntime{Type: EmbeddedInputCommandRuntimeTypeRuntime, RuntimeId: runtimeID, Operation: EmbeddedInputRuntimeCommandCallSubmit{Type: EmbeddedInputRuntimeCommandCallSubmitTypeCallSubmit, TimeoutMs: 1000, Call: EmbeddedInputEmbeddedCall{PoolId: poolID, Export: "call", Arguments: value, Context: EmbeddedInputLuaInvocationContext{}}}})
		driverResult(t, receipt)
		typed, err := DecodeEmbeddedOutputRuntimeCallSubmitResponse(receipt.ResponseBytes())
		if err != nil {
			t.Fatal(err)
		}
		if err := receipt.Forget(); err != nil {
			t.Fatal(err)
		}
		finished := terminal(typed.Result.OperationId)
		if actual, present := finished["value"]; !present || !reflect.DeepEqual(actual, value) || finished["phase"] != "succeeded" {
			t.Fatalf("driver changed real Lua result: %#v", finished)
		}
		send(EmbeddedInputRuntimeCommandOperationForget{Type: EmbeddedInputRuntimeCommandOperationForgetTypeOperationForget, OperationId: typed.Result.OperationId})
	}
	descriptor := EmbeddedInputCapabilityDescriptor{Name: "go.callback", Version: "1.0.0", Description: "Go driver callback integration", InputSchema: true, OutputSchema: true, Execution: EmbeddedInputCapabilityExecutionQueued, Permissions: EmbeddedInputCapabilityDescriptorPermissions{"go.host"}, Scope: EmbeddedInputCapabilityScopeInvocation, MaxConcurrent: 1, MaxCallMs: 10000, MaxInputBytes: 1024, MaxOutputBytes: 1024, Effects: EmbeddedInputCapabilityEffectsMutating, Idempotency: EmbeddedInputCapabilityIdempotencyNone}
	registration := send(EmbeddedInputRuntimeCommandCapabilitiesRegister{Type: EmbeddedInputRuntimeCommandCapabilitiesRegisterTypeCapabilitiesRegister, Descriptors: EmbeddedInputRuntimeCommandCapabilitiesRegisterDescriptors{descriptor}}).(map[string]any)["registration_ids"].([]any)[0]
	callbackPool := pool("return {call=function(a) return vulcan.capabilities.call('go.callback',a) end}")
	operationID := send(EmbeddedInputRuntimeCommandCallSubmit{Type: EmbeddedInputRuntimeCommandCallSubmitTypeCallSubmit, TimeoutMs: 10000, Call: EmbeddedInputEmbeddedCall{PoolId: callbackPool, Export: "call", Arguments: map[string]any{"plugin_id": "forged"}, Context: EmbeddedInputLuaInvocationContext{}}}).(map[string]any)["operation_id"].(string)
	requests := embeddedPoll(t, func() any {
		return send(EmbeddedInputRuntimeCommandHostRequestsTake{Type: EmbeddedInputRuntimeCommandHostRequestsTakeTypeHostRequestsTake, Limit: 1})
	}, func(value any) bool { return len(value.([]any)) > 0 }).([]any)
	request := requests[0].(map[string]any)
	caller := request["caller"].(map[string]any)
	if request["registration_id"] != registration || caller["plugin_id"] != "go-embedded-test" || caller["operation_id"] != operationID {
		t.Fatalf("driver changed trusted callback identity: %#v", request)
	}
	send(EmbeddedInputRuntimeCommandOperationCancel{Type: EmbeddedInputRuntimeCommandOperationCancelTypeOperationCancel, OperationId: operationID})
	send(EmbeddedInputRuntimeCommandHostRequestComplete{Type: EmbeddedInputRuntimeCommandHostRequestCompleteTypeHostRequestComplete, RequestId: request["request_id"].(string), Outcome: EmbeddedInputHostCompletionVariant1{Ok: true, Value: nil, Effects: EmbeddedInputEffectStateCommitted}})
	finished := terminal(operationID)
	committed := false
	for _, effect := range finished["host_effects"].([]any) {
		committed = committed || effect.(map[string]any)["effects"] == "committed"
	}
	if finished["phase"] != "cancelled" || !committed {
		t.Fatalf("late callback evidence lost: %#v", finished)
	}
	send(EmbeddedInputRuntimeCommandOperationForget{Type: EmbeddedInputRuntimeCommandOperationForgetTypeOperationForget, OperationId: operationID})
}

// TestEmbeddedDriverNativeRecoveryClose retains the transport claim while recovery and closure overlap.
// TestEmbeddedDriverNativeRecoveryClose 在恢复与关闭重叠时保留传输声明。
func TestEmbeddedDriverNativeRecoveryClose(t *testing.T) {
	transport := nativeEmbeddedTest(t)
	original := transport.native
	entered, allow := make(chan struct{}), make(chan struct{})
	var unblock sync.Once
	defer unblock.Do(func() { close(allow) })
	var releases atomic.Uint64
	transport.native = &controlledEmbeddedNative{embeddedNative: original, releaseFn: func(id uint64, result embeddedResult) EmbeddedNativeStatus {
		switch releases.Add(1) {
		case 1:
			return EmbeddedNativeBusy
		case 2:
			close(entered)
			<-allow
		}
		return original.release(id, result)
	}}
	driver := nativeDriverTest(t, transport, EmbeddedDriverConfig{1, 1, 1})
	receipt := driverSubmit(t, driver, EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe})
	var release *EmbeddedResultReleaseError
	if _, err := receipt.Result(driverTestContext(t)); !errors.As(err, &release) {
		t.Fatalf("expected retained allocation: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- driver.ReleaseResults(context.Background()) }()
	select {
	case <-entered:
	case <-driverTestContext(t).Done():
		t.Fatal("recovery did not enter")
	}
	driver.RequestClose()
	if driver.Status().Closed || transport.Free() == nil {
		t.Fatal("close released a recovery owner")
	}
	if _, err := driver.Submit(context.Background(), EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe}); err == nil {
		t.Fatal("closing driver admitted work")
	}
	unblock.Do(func() { close(allow) })
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-driverTestContext(t).Done():
		t.Fatal("recovery did not return")
	}
	if err := driver.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	if !driver.Status().Closed || releases.Load() != 2 || transport.RetainedResults() != 0 {
		t.Fatal("recovery did not safely drain")
	}
	// A second exact driver claim may start only after the first has really released its workers.
	// 只有第一个驱动实际释放工作位后，第二个精确驱动声明才能启动。
	nativeDriverTest(t, transport, EmbeddedDriverConfig{1, 1, 1})
}

// TestEmbeddedDriverNativeReservation rejects pre-existing allocations and insufficient aggregate response bytes.
// TestEmbeddedDriverNativeReservation 拒绝既有分配及不足的累计响应字节。
func TestEmbeddedDriverNativeReservation(t *testing.T) {
	transport := nativeEmbeddedTest(t)
	original := transport.native
	var rejected atomic.Bool
	transport.native = &controlledEmbeddedNative{embeddedNative: original, releaseFn: func(id uint64, result embeddedResult) EmbeddedNativeStatus {
		if rejected.CompareAndSwap(false, true) {
			return EmbeddedNativeBusy
		}
		return original.release(id, result)
	}}
	if _, err := transport.Request(map[string]any{"type": "describe"}); err == nil {
		t.Fatal("fixture did not retain an allocation")
	}
	if driver, err := NewEmbeddedCommandDriver(transport, EmbeddedDriverConfig{1, 1, 1}); err == nil {
		if err := driver.Close(driverTestContext(t)); err != nil {
			t.Fatal(err)
		}
		t.Fatal("driver reserved capacity over a pre-existing allocation")
	}
	if err := transport.ReleaseResults(); err != nil {
		t.Fatal(err)
	}
	nativeDriverTest(t, transport, EmbeddedDriverConfig{1, 1, 1})
	config := embeddedTestConfig()
	config.MaxResultBytes = 2*config.MaxResponseBytes - 1
	limited, err := NewEmbeddedTransport(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := limited.Close(); err != nil {
			t.Error(err)
		}
		if err := limited.Free(); err != nil {
			t.Error(err)
		}
	})
	if driver, err := NewEmbeddedCommandDriver(limited, EmbeddedDriverConfig{1, 1, 1}); err == nil {
		if err := driver.Close(driverTestContext(t)); err != nil {
			t.Fatal(err)
		}
		t.Fatal("worker count fit but worst-case response bytes did not")
	}
}
