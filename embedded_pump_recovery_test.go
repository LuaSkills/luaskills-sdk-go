//go:build cgo

package luaskills

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

// pumpFaultNative intercepts one declared boundary while all unmodified calls and allocations use the actual native library.
// pumpFaultNative 拦截一个声明边界，其余全部调用及分配使用实际原生库。
type pumpFaultNative struct {
	embeddedNative
	// route and mode select the one fault; armed and fired make injection deterministic across coordinator calls.
	// route 和 mode 选择唯一故障；armed 及 fired 使跨协调调用注入保持确定性。
	route string
	mode  string
	armed atomic.Bool
	fired atomic.Bool
	// allocations records only actual results selected for copy or release injection.
	// allocations 仅记录选中用于复制或释放注入的实际结果。
	allocations sync.Map
	// calls counts attempts at the selected mutation, including explicitly allowed completion re-delivery.
	// calls 统计选中变更的尝试，包含显式允许的完成重新交付。
	calls atomic.Uint64
}

// request forwards exact native bytes and selects only a nonempty callback extraction when take is the fault target.
// request 转发精确原生字节；故障目标为领取时，仅选择非空回调提取。
func (n *pumpFaultNative) request(id uint64, frame []byte) (embeddedResult, EmbeddedNativeStatus) {
	decoded, err := DecodeEmbeddedJSON(frame)
	if err != nil {
		panic(err)
	}
	command := decoded.(map[string]any)["command"].(map[string]any)
	matched := false
	if command["type"] == "runtime" {
		matched = command["operation"].(map[string]any)["type"] == n.route
	}
	if matched {
		n.calls.Add(1)
	}
	if matched && n.mode == "before" && n.armed.Load() && n.fired.CompareAndSwap(false, true) {
		return embeddedResult{}, EmbeddedNativeInternal
	}
	result, status := n.embeddedNative.request(id, frame)
	if matched && status == EmbeddedNativeOk && n.armed.Load() && !n.fired.Load() {
		if n.route == "host_requests_take" {
			response, err := n.embeddedNative.copy(result, embeddedTestConfig().MaxResponseBytes)
			if err != nil {
				panic(err)
			}
			batch, err := DecodeEmbeddedOutputRuntimeHostRequestsTakeResponse(response)
			if err != nil || len(batch.Result) == 0 {
				return result, status
			}
		}
		n.allocations.Store(result.id, true)
	}
	return result, status
}

// copy can discard one real delivered response while leaving the ordinary transport to release its actual allocation.
// copy 可以丢弃一个真实已交付响应，同时仍由普通传输释放其实际分配。
func (n *pumpFaultNative) copy(result embeddedResult, limit uint64) ([]byte, error) {
	_, selected := n.allocations.Load(result.id)
	if selected && n.mode == "lost" && n.fired.CompareAndSwap(false, true) {
		return nil, errors.New("controlled lost completion receipt")
	}
	return n.embeddedNative.copy(result, limit)
}

// release rejects one real free without invoking C, then permits explicit recovery of the original descriptor.
// release 不调用 C 而拒绝一次真实释放，然后允许显式恢复原始描述符。
func (n *pumpFaultNative) release(id uint64, result embeddedResult) EmbeddedNativeStatus {
	_, selected := n.allocations.Load(result.id)
	if selected && n.mode == "release" && n.fired.CompareAndSwap(false, true) {
		return EmbeddedNativeBusy
	}
	status := n.embeddedNative.release(id, result)
	if status == EmbeddedNativeOk {
		n.allocations.Delete(result.id)
	}
	return status
}

// TestEmbeddedPumpNativeCompletionRecovery proves exact reconciliation after pre-dispatch failure, lost delivery and failed free.
// TestEmbeddedPumpNativeCompletionRecovery 证明分发前失败、交付丢失及释放失败后的精确核对。
func TestEmbeddedPumpNativeCompletionRecovery(t *testing.T) {
	for _, mode := range []string{"before", "lost", "release"} {
		t.Run(mode, func(t *testing.T) {
			transport, runtimeID, command, pool := nativeEmbeddedRuntime(t)
			backend := &pumpFaultNative{embeddedNative: transport.native, route: "host_request_complete", mode: mode}
			transport.native = backend
			pump := nativePumpTest(t, transport, runtimeID, EmbeddedCallbackPumpConfig{1, 1, 1})
			var handlers atomic.Uint64
			pumpRegister(t, pump, pumpCapability(func(arguments any, ctx *EmbeddedHostCallbackContext) (any, error) {
				handlers.Add(1)
				return "actual result", ctx.ReportEffects(EmbeddedInputEffectStateCommitted)
			}))
			backend.armed.Store(true)
			id := embeddedSubmit(command, pool("return {call=function(a) return vulcan.capabilities.call('go.callback',a) end}"), nil)
			embeddedPoll(t, func() any { return pump.Status() }, func(value any) bool { return value.(EmbeddedCallbackPumpStatus).RecoveryRequired })
			if handlers.Load() != 1 || transport.Free() == nil {
				t.Fatal("failed completion lost actual ownership")
			}
			if err := pump.RetryAcknowledgements(driverTestContext(t)); err != nil {
				t.Fatal(err)
			}
			result := embeddedTerminal(t, command, id)
			committed := false
			for _, effect := range result["host_effects"].([]any) {
				committed = committed || effect.(map[string]any)["effects"] == "committed"
			}
			attempts := uint64(1)
			if mode == "before" {
				attempts = 2
			}
			if !committed || handlers.Load() != 1 || backend.calls.Load() != attempts {
				t.Fatalf("completion or handler replay: effects=%v handlers=%d commands=%d", committed, handlers.Load(), backend.calls.Load())
			}
			if err := pump.Close(driverTestContext(t)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestEmbeddedPumpNativeMutationRelease retains successful registration, extraction and retirement evidence across failed frees.
// TestEmbeddedPumpNativeMutationRelease 在释放失败期间保留成功注册、领取及退役证据。
func TestEmbeddedPumpNativeMutationRelease(t *testing.T) {
	for _, route := range []string{"capabilities_register", "host_requests_take", "capability_unregister", "capability_forget"} {
		t.Run(route, func(t *testing.T) {
			transport, runtimeID, command, pool := nativeEmbeddedRuntime(t)
			backend := &pumpFaultNative{embeddedNative: transport.native, route: route, mode: "release"}
			transport.native = backend
			pump := nativePumpTest(t, transport, runtimeID, EmbeddedCallbackPumpConfig{1, 1, 1})
			backend.armed.Store(true)
			var handlers atomic.Uint64
			registration := pumpRegister(t, pump, pumpCapability(func(arguments any, ctx *EmbeddedHostCallbackContext) (any, error) {
				handlers.Add(1)
				return nil, ctx.ReportEffects(EmbeddedInputEffectStateCommitted)
			}))
			var operation string
			if route == "host_requests_take" {
				operation = embeddedSubmit(command, pool("return {call=function(a) return vulcan.capabilities.call('go.callback',a) end}"), nil)
			} else if route != "capabilities_register" {
				// Start retirement with a short-lived observer; the accepted intent survives its cancellation.
				// 用短生命周期观察者启动退役；已接纳意图跨观察者取消存活。
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				returned := make(chan error, 1)
				go func() { returned <- pump.Unregister(ctx, registration) }()
				embeddedPoll(t, func() any { return pump.Status() }, func(value any) bool { return value.(EmbeddedCallbackPumpStatus).RecoveryRequired })
				cancel()
				select {
				case err := <-returned:
					if err != nil && !errors.Is(err, context.Canceled) {
						t.Fatal(err)
					}
				case <-driverTestContext(t).Done():
					t.Fatal("unregister observer remained attached")
				}
			}
			embeddedPoll(t, func() any { return pump.Status() }, func(value any) bool { return value.(EmbeddedCallbackPumpStatus).RecoveryRequired })
			if transport.RetainedResults() != 1 || pump.Status().Closed || transport.Free() == nil {
				t.Fatal("failed release lost the pump reservation")
			}
			if err := pump.RetryAcknowledgements(driverTestContext(t)); err != nil {
				t.Fatal(err)
			}
			if operation != "" {
				embeddedTerminal(t, command, operation)
				if handlers.Load() != 1 {
					t.Fatal("delivered handler lost or replayed")
				}
			}
			if err := pump.Close(driverTestContext(t)); err != nil {
				t.Fatal(err)
			}
			if route != "host_requests_take" && backend.calls.Load() != 1 {
				t.Fatalf("mutation was repeated: %s %d", route, backend.calls.Load())
			}
		})
	}
}
