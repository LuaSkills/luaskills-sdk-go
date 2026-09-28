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

// TestEmbeddedCallbackRequestIdentity uses t to verify all optional identity states and immutable callback snapshots; it returns no value.
// TestEmbeddedCallbackRequestIdentity 使用 t 验证可选身份的全部状态及不可变回调快照；无返回值。
func TestEmbeddedCallbackRequestIdentity(t *testing.T) {
	// Keep the host correlation separate from the queue delivery identifier.
	// 将宿主关联号与队列投递标识保持分离。
	original := "host-request"
	// Represent a present string with the generated optional-nullable wire type.
	// 使用生成的可选可空线协议类型表示已提供字符串。
	originalPointer := &original
	// Freeze authority before any caller can mutate the source object.
	// 在调用方能修改来源对象之前冻结权威。
	ctx := newEmbeddedHostContext(EmbeddedOutputHostRequest{RequestId: "queue-request", Caller: EmbeddedOutputCapabilityCaller{RequestId: &originalPointer}}, EmbeddedInputCapabilityEffectsReadOnly)
	original = "source-forged"
	originalPointer = nil
	// Fetch an independent snapshot, including both pointer layers.
	// 获取包含两层指针的独立快照。
	caller := ctx.Caller()
	if caller.RequestId == nil || *caller.RequestId == nil || **caller.RequestId != "host-request" {
		t.Fatal("source mutation changed frozen request identity")
	}
	**caller.RequestId = "consumer-forged"
	*caller.RequestId = nil
	if next := ctx.Caller(); next.RequestId == nil || *next.RequestId == nil || **next.RequestId != "host-request" || ctx.RequestID() != "queue-request" {
		t.Fatal("returned caller aliases request identity or conflates queue identity")
	}
	// Preserve absent and explicit-null wire states without inventing a request.
	// 保留缺省及显式空值线协议状态，不虚构请求。
	var explicitNull *string
	// Freeze the explicit-null outer pointer before the source changes.
	// 在来源变化前冻结显式空值的外层指针。
	nullContext := newEmbeddedHostContext(EmbeddedOutputHostRequest{Caller: EmbeddedOutputCapabilityCaller{RequestId: &explicitNull}}, EmbeddedInputCapabilityEffectsReadOnly)
	explicitNull = &original
	// Mutation of a returned null snapshot must not change future reads.
	// 修改返回的空值快照不得改变后续读取。
	nullCaller := nullContext.Caller()
	if nullCaller.RequestId == nil || *nullCaller.RequestId != nil {
		t.Fatal("explicit-null request presence changed")
	}
	*nullCaller.RequestId = &original
	if next := nullContext.Caller(); next.RequestId == nil || *next.RequestId != nil {
		t.Fatal("returned caller aliases explicit-null presence")
	}
	if absent := newEmbeddedHostContext(EmbeddedOutputHostRequest{}, EmbeddedInputCapabilityEffectsReadOnly).Caller(); absent.RequestId != nil {
		t.Fatal("absent request identity became present")
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
