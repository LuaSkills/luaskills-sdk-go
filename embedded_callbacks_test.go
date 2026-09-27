package luaskills

import (
	"context"
	"errors"
	"testing"
)

// TestEmbeddedCallbackContext preserves full-width advisory budgets, copied authority and sealed explicit effects.
// TestEmbeddedCallbackContext 保留完整位宽参考预算、复制权威及已封存显式副作用。
func TestEmbeddedCallbackContext(t *testing.T) {
	session := "session"
	request := EmbeddedOutputHostRequest{RequestId: "request", RegistrationId: "registration", RemainingMs: ^uint64(0), Caller: EmbeddedOutputCapabilityCaller{PluginId: "plugin", SessionId: &session}}
	ctx := newEmbeddedHostContext(request, EmbeddedInputCapabilityEffectsMutating)
	session = "changed"
	caller := ctx.Caller()
	*caller.SessionId = "forged"
	if *ctx.Caller().SessionId != "session" || ctx.RequestID() != "request" || ctx.RegistrationID() != "registration" || ctx.RemainingMS() < 1<<63 {
		t.Fatal("callback identity or uint64 budget changed")
	}
	if _, hasDeadline := ctx.Deadline(); hasDeadline || ctx.Effects() != EmbeddedInputEffectStateUnknown {
		t.Fatal("context invented a native deadline or transaction result")
	}
	if ctx.ReportEffects("invalid") == nil || ctx.ReportEffects(EmbeddedInputEffectStateCommitted) != nil || ctx.seal() != EmbeddedInputEffectStateCommitted || ctx.ReportEffects(EmbeddedInputEffectStateRolledBack) == nil {
		t.Fatal("effect validation or sealing failed")
	}
	ctx.observeCancellation(EmbeddedOutputEmbeddedError{Code: "deadline_exceeded", Message: "original deadline"})
	ctx.observeCancellation(EmbeddedOutputEmbeddedError{Code: "cancelled", Message: "later reason"})
	reason := ctx.Cancellation()
	reason.Code = "forged"
	if !errors.Is(ctx.Err(), context.Canceled) || ctx.Cancellation().Code != "deadline_exceeded" {
		t.Fatal("first core cancellation was overwritten")
	}
	derived, cancel := context.WithCancel(ctx)
	defer cancel()
	if checkEmbeddedObserver(derived) == nil || checkEmbeddedObserver(nil) == nil || checkEmbeddedObserver(context.Background()) != nil {
		t.Fatal("callback context dependency marker lost")
	}
	readOnly := newEmbeddedHostContext(EmbeddedOutputHostRequest{}, EmbeddedInputCapabilityEffectsReadOnly)
	if readOnly.Effects() != EmbeddedInputEffectStateNotApplicable || readOnly.RemainingMS() != 0 {
		t.Fatal("read-only or exhausted advisory budget changed")
	}
}

// TestEmbeddedCallbackErrorProjection exposes only explicitly valid protocol errors, never arbitrary host exception detail.
// TestEmbeddedCallbackErrorProjection 仅暴露显式有效协议错误，绝不暴露任意宿主异常细节。
func TestEmbeddedCallbackErrorProjection(t *testing.T) {
	for _, err := range []error{errors.New("secret"), &EmbeddedRuntimeError{Code: "invalid_code", Message: "secret"}} {
		failure := embeddedCallbackFailure(err)
		if failure.Code != "execution_failed" || failure.Message != "Go host callback failed" {
			t.Fatal("host diagnostic leaked")
		}
	}
	failure := embeddedCallbackFailure(&EmbeddedRuntimeError{Code: "permission_denied", Message: "declared failure"})
	if failure.Code != "permission_denied" || failure.Message != "declared failure" {
		t.Fatal("explicit valid protocol error changed")
	}
}

// TestEmbeddedPumpFrameClaims verifies aggregate reservations in both creation orders without invoking a native backend.
// TestEmbeddedPumpFrameClaims 在不调用原生后端的情况下验证两种创建顺序的累计预留。
func TestEmbeddedPumpFrameClaims(t *testing.T) {
	for _, config := range []EmbeddedTransportConfig{{1, 2, 4096, 2048, 2048}, {1, 3, 6143, 2048, 2048}} {
		// This synthetic ownership-only object is never registered globally and never enters native code.
		// 此合成纯所有权对象绝不全局注册，也绝不进入原生代码。
		transport := &EmbeddedTransport{identity: 1, config: config}
		driver := &EmbeddedCommandDriver{config: EmbeddedDriverConfig{1, 1, 1}}
		pump := &EmbeddedCallbackPump{runtimeID: "runtime"}
		if err := transport.claimCommandDriver(driver); err != nil {
			t.Fatal(err)
		}
		if transport.claimCallbackPump(pump) == nil {
			t.Fatal("pump ignored existing driver reservation")
		}
		if err := transport.releaseCommandDriver(driver); err != nil {
			t.Fatal(err)
		}
		if err := transport.claimCallbackPump(pump); err != nil {
			t.Fatal(err)
		}
		if transport.claimCommandDriver(driver) == nil || transport.claimCallbackPump(&EmbeddedCallbackPump{runtimeID: "runtime"}) == nil {
			t.Fatal("existing callback ownership was ignored")
		}
		if transport.releaseCallbackPump(&EmbeddedCallbackPump{runtimeID: "runtime"}) == nil {
			t.Fatal("foreign identity released callback claim")
		}
		if err := transport.releaseCallbackPump(pump); err != nil {
			t.Fatal(err)
		}
	}
}
