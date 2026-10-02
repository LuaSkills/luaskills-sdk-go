//go:build cgo

package luaskills

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// TestEmbeddedPumpNativeUnsafeIntegerCompletion retains actual commit evidence after an invalid Lua application result.
// TestEmbeddedPumpNativeUnsafeIntegerCompletion 在无效 Lua 应用结果后保留实际提交证据。
// t owns pump/runtime cleanup; the original request completes once and a subsequent normal callback remains usable.
// t 拥有泵／运行时清理；原请求单次完成且后续正常回调仍可使用。
func TestEmbeddedPumpNativeUnsafeIntegerCompletion(t *testing.T) {
	// Atomic observations bind all completion assertions to exact callbacks across worker goroutines.
	// 原子观察跨工作协程把所有完成断言绑定到精确回调。
	transport, runtimeID, command, pool := nativeEmbeddedRuntime(t)
	pump := nativePumpTest(t, transport, runtimeID, EmbeddedCallbackPumpConfig{1, 2, 1})
	var calls atomic.Uint64
	var firstRequest atomic.Value
	pumpRegister(t, pump, pumpCapability(func(value any, ctx *EmbeddedHostCallbackContext) (any, error) {
		if calls.Add(1) == 1 {
			firstRequest.Store(ctx.RequestID())
		}
		if err := ctx.ReportEffects(EmbeddedInputEffectStateCommitted); err != nil {
			return nil, err
		}
		if value == "unsafe" {
			return json.Number("18446744073709551615"), nil
		}
		return value, nil
	}))
	// Lua returns the full envelope, separating operation success from capability success.
	// Lua 返回完整信封，区分操作成功与能力成功。
	poolID := pool("return {call=function(a) return vulcan.capabilities.call('go.callback',a) end}")
	rejected := embeddedTerminal(t, command, embeddedSubmit(command, poolID, "unsafe"))
	// One callback's committed evidence never determines the entire Lua operation's effects.
	// 单个回调的已提交证据绝不能确定整个 Lua 操作的副作用。
	if rejected["effects"] != "unknown" {
		t.Fatalf("overall Lua effects lost their independent unknown state: %#v", rejected)
	}
	envelope := rejected["value"].(map[string]any)
	if rejected["phase"] != "succeeded" || envelope["ok"] != false || envelope["error"].(map[string]any)["code"] != "invalid_argument" || envelope["effects"] != "committed" {
		t.Fatalf("invalid callback result lost its committed error envelope: %#v", rejected)
	}
	// Exact identity lookup avoids assertions tied to a mutable ledger position.
	// 精确身份查找避免断言绑定可变化的账本位置。
	found := false
	for _, raw := range rejected["host_effects"].([]any) {
		// Original record must prove both handler ownership release and actual effects.
		// 原始记录必须证明处理器所有权释放及实际副作用。
		effect := raw.(map[string]any)
		if effect["request_id"] == firstRequest.Load() {
			found = effect["phase"] == "completed" && effect["effects"] == "committed"
		}
	}
	if !found {
		t.Fatal("original completed committed ledger is missing")
	}
	if normal := embeddedTerminal(t, command, embeddedSubmit(command, poolID, "normal")); normal["value"].(map[string]any)["value"] != "normal" {
		t.Fatalf("callback pump was not reusable: %#v", normal)
	}
	if err := pump.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	// Closed pump ownership and exactly two calls prove that native acknowledgement did not cause replay.
	// 泵关闭所有权及精确两次调用证明原生确认没有导致重播。
	status := pump.Status()
	if calls.Load() != 2 || len(status.RequestIDs) != 0 || len(status.PendingAcknowledgements) != 0 || status.Failure != nil {
		t.Fatalf("unsafe completion leaked or replayed ownership: %#v", status)
	}
}

// nativePumpTest starts an actual ready pump and closes it before fixture-owned native runtime cleanup.
// nativePumpTest 启动实际就绪泵，并在夹具拥有的原生运行时清理前关闭。
func nativePumpTest(t *testing.T, transport *EmbeddedTransport, runtimeID string, config EmbeddedCallbackPumpConfig) *EmbeddedCallbackPump {
	t.Helper()
	pump, err := NewEmbeddedCallbackPump(transport, runtimeID, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := pump.Close(driverTestContext(t)); err != nil {
			t.Error(err)
		}
	})
	if err := pump.Ready(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	return pump
}

// pumpCapability declares one real queued callback with explicit native budgets and permissions.
// pumpCapability 声明一个具有显式原生预算及权限的真实队列回调。
func pumpCapability(handler EmbeddedHostHandler) EmbeddedHostCapability {
	return EmbeddedHostCapability{Descriptor: EmbeddedInputCapabilityDescriptor{Name: "go.callback", Version: "1.0.0", Description: "Go callback pump integration", InputSchema: true, OutputSchema: true, Execution: EmbeddedInputCapabilityExecutionQueued, Permissions: EmbeddedInputCapabilityDescriptorPermissions{"go.host"}, Scope: EmbeddedInputCapabilityScopeInvocation, MaxConcurrent: 2, MaxCallMs: 10000, MaxInputBytes: 1024, MaxOutputBytes: 1024, Effects: EmbeddedInputCapabilityEffectsMutating, Idempotency: EmbeddedInputCapabilityIdempotencyNone}, Handler: handler}
}

// pumpRegister requires one actual accepted registration and returns its exact immutable native identity.
// pumpRegister 要求一个实际已接纳注册，并返回其精确不可变原生身份。
func pumpRegister(t *testing.T, pump *EmbeddedCallbackPump, capability EmbeddedHostCapability) string {
	t.Helper()
	identities, err := pump.Register(driverTestContext(t), []EmbeddedHostCapability{capability})
	if err != nil || len(identities) != 1 {
		t.Fatalf("callback registration: %#v %v", identities, err)
	}
	return identities[0]
}

// TestEmbeddedPumpNativeDelivery runs actual Lua through automatic Go callbacks and preserves copied identity and effect evidence.
// TestEmbeddedPumpNativeDelivery 通过自动 Go 回调运行真实 Lua，并保留复制身份及副作用证据。
func TestEmbeddedPumpNativeDelivery(t *testing.T) {
	transport, runtimeID, command, pool := nativeEmbeddedRuntime(t)
	pump := nativePumpTest(t, transport, runtimeID, EmbeddedCallbackPumpConfig{2, 2, 1})
	var lastContext atomic.Pointer[EmbeddedHostCallbackContext]
	var calls atomic.Uint64
	capability := pumpCapability(func(arguments any, ctx *EmbeddedHostCallbackContext) (any, error) {
		calls.Add(1)
		lastContext.Store(ctx)
		caller := ctx.Caller()
		if caller.PluginId != "go-embedded-test" || caller.OperationId == "" || ctx.RegistrationID() == "" || ctx.RequestID() == "" {
			return nil, errors.New("trusted identity missing")
		}
		caller.PluginId = "forged"
		if ctx.Caller().PluginId != "go-embedded-test" {
			return nil, errors.New("caller alias changed authority")
		}
		if err := ctx.ReportEffects(EmbeddedInputEffectStateCommitted); err != nil {
			return nil, err
		}
		return arguments, nil
	})
	registration := pumpRegister(t, pump, capability)
	capability.Descriptor.Name = "mutated"
	capability.Descriptor.Permissions[0] = "mutated"
	poolID := pool("return {call=function(a) return vulcan.capabilities.call('go.callback',a) end}")
	for _, value := range []any{nil, false, json.Number("0.25"), "中文"} {
		id := embeddedSubmit(command, poolID, value)
		result := embeddedTerminal(t, command, id)
		envelope := result["value"].(map[string]any)
		if result["phase"] != "succeeded" || envelope["ok"] != true || !reflect.DeepEqual(envelope["value"], value) {
			t.Fatalf("automatic callback delivery changed: %#v", result)
		}
		if err := lastContext.Load().ReportEffects(EmbeddedInputEffectStateRolledBack); err == nil {
			t.Fatal("returned handler rewrote sealed evidence")
		}
		// Read the typed historical caller by exact request identity after the real callback has returned.
		// 真实回调返回后，按精确请求身份读取类型化历史调用方。
		found := false
		// Find evidence through the request identity instead of relying on array positions.
		// 通过请求身份查找证据，不依赖数组位置。
		for _, raw := range result["host_effects"].([]any) {
			// Project the fixture effect object for exact request identity comparison.
			// 投影夹具副作用对象，以比较精确请求身份。
			effect := raw.(map[string]any)
			if effect["request_id"] != lastContext.Load().RequestID() {
				continue
			}
			// Decode through the generated public caller type; compare every original authority field.
			// 通过生成的公开调用方类型解码；比较每个原始权威字段。
			encoded, err := json.Marshal(effect["caller"])
			if err != nil {
				t.Fatal(err)
			}
			// Keep the decoded public caller separate from the still-retained callback context.
			// 将解码后的公开调用方与仍保留的回调上下文分开。
			var historical EmbeddedOutputCapabilityCaller
			if err := json.Unmarshal(encoded, &historical); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(historical, lastContext.Load().Caller()) {
				t.Fatalf("historical caller changed: %#v", historical)
			}
			found = true
		}
		if !found {
			t.Fatal("original callback request evidence is missing; fixture identity may have changed")
		}
		command(map[string]any{"type": "operation_forget", "operation_id": id})
	}
	if calls.Load() != 4 {
		t.Fatal("callback lost or replayed")
	}
	if err := pump.Unregister(driverTestContext(t), registration); err != nil {
		t.Fatal(err)
	}
	if len(pump.Status().RegistrationIDs) != 0 || len(pump.Status().RequestIDs) != 0 {
		t.Fatal("unregister did not drain actual ownership")
	}
}

// TestEmbeddedPumpNativeCancellation proves cancellation and close retain real handler capacity until late effects are acknowledged.
// TestEmbeddedPumpNativeCancellation 证明取消及关闭在迟到副作用确认前保留实际处理器容量。
func TestEmbeddedPumpNativeCancellation(t *testing.T) {
	transport, runtimeID, command, pool := nativeEmbeddedRuntime(t)
	driver := nativeDriverTest(t, transport, EmbeddedDriverConfig{1, 2, 2})
	pump := nativePumpTest(t, transport, runtimeID, EmbeddedCallbackPumpConfig{1, 1, 1})
	entered, cancelled, allow := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var unblock sync.Once
	defer unblock.Do(func() { close(allow) })
	var calls atomic.Uint64
	registration := pumpRegister(t, pump, pumpCapability(func(arguments any, ctx *EmbeddedHostCallbackContext) (any, error) {
		calls.Add(1)
		if _, err := driver.Submit(ctx, EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe}); err == nil {
			return nil, errors.New("nested callback command accepted")
		}
		derived, cancel := context.WithCancel(ctx)
		defer cancel()
		if err := pump.Ready(derived); err == nil {
			return nil, errors.New("derived callback context bypassed dependency guard")
		}
		close(entered)
		<-ctx.Done()
		if ctx.Cancellation() == nil {
			return nil, errors.New("core cancellation diagnostic lost")
		}
		close(cancelled)
		<-allow
		return "late result", ctx.ReportEffects(EmbeddedInputEffectStateCommitted)
	}))
	poolID := pool("return {call=function(a) return vulcan.capabilities.call('go.callback',a) end}")
	first := embeddedSubmit(command, poolID, nil)
	select {
	case <-entered:
	case <-driverTestContext(t).Done():
		t.Fatal("handler did not start")
	}
	second := embeddedSubmit(command, poolID, nil)
	command(map[string]any{"type": "operation_cancel", "operation_id": first})
	select {
	case <-cancelled:
	case <-driverTestContext(t).Done():
		t.Fatal("core cancellation did not reach handler")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := pump.Close(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("close observation changed: %v", err)
	}
	status := pump.Status()
	if status.Closed || len(status.RequestIDs) != 1 || calls.Load() != 1 || transport.Free() == nil {
		t.Fatal("cancelled handler prematurely released ownership")
	}
	unblock.Do(func() { close(allow) })
	firstResult := embeddedTerminal(t, command, first)
	secondResult := embeddedTerminal(t, command, second)
	if firstResult["phase"] != "cancelled" || secondResult["value"].(map[string]any)["ok"] != false {
		t.Fatalf("native cancellation/retirement changed: %#v %#v", firstResult, secondResult)
	}
	committed := false
	for _, effect := range firstResult["host_effects"].([]any) {
		entry := effect.(map[string]any)
		committed = committed || entry["registration_id"] == registration && entry["effects"] == "committed"
	}
	if !committed || calls.Load() != 1 {
		t.Fatal("late committed effects lost or waiting callback was executed after retirement")
	}
	if err := pump.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
}

// TestEmbeddedPumpNativeHandlerFailures converts errors, panics and invalid values into safe completions with preserved effects.
// TestEmbeddedPumpNativeHandlerFailures 将错误、panic 及无效值转换为安全完成，并保留副作用。
func TestEmbeddedPumpNativeHandlerFailures(t *testing.T) {
	for _, mode := range []string{"error", "panic", "invalid", "deep", "goexit"} {
		t.Run(mode, func(t *testing.T) {
			transport, runtimeID, command, pool := nativeEmbeddedRuntime(t)
			pump := nativePumpTest(t, transport, runtimeID, EmbeddedCallbackPumpConfig{1, 1, 1})
			pumpRegister(t, pump, pumpCapability(func(arguments any, ctx *EmbeddedHostCallbackContext) (any, error) {
				if err := ctx.ReportEffects(EmbeddedInputEffectStateCommitted); err != nil {
					return nil, err
				}
				switch mode {
				case "panic":
					panic("secret handler detail")
				case "invalid":
					return make(chan struct{}), nil
				case "deep":
					var value any
					for index := 0; index < 150; index++ {
						value = []any{value}
					}
					return value, nil
				case "goexit":
					runtime.Goexit()
				}
				return nil, errors.New("secret handler detail")
			}))
			id := embeddedSubmit(command, pool("return {call=function(a) return vulcan.capabilities.call('go.callback',a) end}"), nil)
			result := embeddedTerminal(t, command, id)
			encoded, err := EncodeEmbeddedJSON(result, transport.Config().MaxResponseBytes)
			if err != nil || strings.Contains(string(encoded), "secret handler detail") || result["value"].(map[string]any)["ok"] != false {
				t.Fatalf("unsafe handler failure: %#v %v", result, err)
			}
			committed := false
			for _, effect := range result["host_effects"].([]any) {
				committed = committed || effect.(map[string]any)["effects"] == "committed"
			}
			if !committed {
				t.Fatalf("error discarded committed effect: %#v", result)
			}
		})
	}
}

// TestEmbeddedPumpNativePublicationCancellation retains a frozen in-flight registration after its observer leaves.
// TestEmbeddedPumpNativePublicationCancellation 在观察者离开后保留冻结的在途注册。
func TestEmbeddedPumpNativePublicationCancellation(t *testing.T) {
	transport, runtimeID, command, pool := nativeEmbeddedRuntime(t)
	original := transport.native
	entered, allow := make(chan struct{}), make(chan struct{})
	var once, unblock sync.Once
	defer unblock.Do(func() { close(allow) })
	var publications atomic.Uint64
	transport.native = driverRequestNative{original, func(id uint64, frame []byte) (embeddedResult, EmbeddedNativeStatus) {
		decoded, err := DecodeEmbeddedJSON(frame)
		if err != nil {
			panic(err)
		}
		request := decoded.(map[string]any)["command"].(map[string]any)
		if request["type"] == "runtime" && request["operation"].(map[string]any)["type"] == "capabilities_register" {
			publications.Add(1)
			once.Do(func() { close(entered); <-allow })
		}
		return original.request(id, frame)
	}}
	pump := nativePumpTest(t, transport, runtimeID, EmbeddedCallbackPumpConfig{1, 1, 1})
	var handlers atomic.Uint64
	capabilities := []EmbeddedHostCapability{pumpCapability(func(arguments any, ctx *EmbeddedHostCallbackContext) (any, error) {
		handlers.Add(1)
		return "original", nil
	})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	returned := make(chan error, 1)
	go func() { _, err := pump.Register(ctx, capabilities); returned <- err }()
	select {
	case <-entered:
	case <-driverTestContext(t).Done():
		t.Fatal("publication did not enter native boundary")
	}
	capabilities[0].Descriptor.Name = "changed"
	capabilities[0].Handler = func(any, *EmbeddedHostCallbackContext) (any, error) { panic("mutated handler") }
	cancel()
	if err := <-returned; !errors.Is(err, context.Canceled) {
		t.Fatalf("publication observation did not cancel: %v", err)
	}
	if _, err := pump.Register(driverTestContext(t), []EmbeddedHostCapability{pumpCapability(func(any, *EmbeddedHostCallbackContext) (any, error) { return nil, nil })}); err == nil {
		t.Fatal("cancelled observer released publication quota")
	}
	if pump.Status().PendingCommands != 1 || transport.Free() == nil {
		t.Fatal("native publication ownership disappeared")
	}
	unblock.Do(func() { close(allow) })
	embeddedPoll(t, func() any { return pump.Status() }, func(value any) bool { return len(value.(EmbeddedCallbackPumpStatus).RegistrationIDs) == 1 })
	id := embeddedSubmit(command, pool("return {call=function(a) return vulcan.capabilities.call('go.callback',a) end}"), nil)
	result := embeddedTerminal(t, command, id)
	if result["value"].(map[string]any)["value"] != "original" || publications.Load() != 1 || handlers.Load() != 1 {
		t.Fatalf("cancelled observer changed registration delivery: %#v", result)
	}
}

// TestEmbeddedPumpNativeStartupOwnership releases failed startup claims and refuses duplicate queue consumers.
// TestEmbeddedPumpNativeStartupOwnership 释放启动失败声明，并拒绝重复队列消费者。
func TestEmbeddedPumpNativeStartupOwnership(t *testing.T) {
	transport, runtimeID, _, _ := nativeEmbeddedRuntime(t)
	missing, err := NewEmbeddedCallbackPump(transport, "missing", EmbeddedCallbackPumpConfig{1, 1, 1})
	if err != nil {
		t.Fatal(err)
	}
	if missing.Ready(driverTestContext(t)) == nil {
		t.Fatal("unknown runtime became ready")
	}
	if err := missing.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	pump := nativePumpTest(t, transport, runtimeID, EmbeddedCallbackPumpConfig{1, 1, 1})
	if duplicate, err := NewEmbeddedCallbackPump(transport, runtimeID, EmbeddedCallbackPumpConfig{1, 1, 1}); err == nil {
		_ = duplicate.Close(driverTestContext(t))
		t.Fatal("duplicate queue consumer accepted")
	}
	if transport.Free() == nil {
		t.Fatal("live idle pump lost transport ownership")
	}
	if err := pump.Close(driverTestContext(t)); err != nil {
		t.Fatal(err)
	}
	nativePumpTest(t, transport, runtimeID, EmbeddedCallbackPumpConfig{1, 1, 1})
}
