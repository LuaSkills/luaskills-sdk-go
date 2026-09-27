//go:build cgo

package luaskills

import (
	"encoding/json"
	"reflect"
	"testing"
)

// embeddedTypedNative sends exactly one generated request and returns its copied native envelope without replay.
// embeddedTypedNative 精确发送一个生成请求，并返回复制的原生信封，不进行重放。
func embeddedTypedNative(t *testing.T, transport *EmbeddedTransport, command EmbeddedInputCommand) []byte {
	t.Helper()
	frame, err := EncodeEmbeddedRequest(EmbeddedInputRequest{ProtocolVersion: EmbeddedProtocolVersion, Command: command}, transport.Config().MaxRequestBytes)
	if err != nil {
		t.Fatal(err)
	}
	response, err := transport.requestBytes(frame)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

// TestEmbeddedWireNativeRoundTrip consumes real description, lifecycle, Lua results and business errors through generated types.
// TestEmbeddedWireNativeRoundTrip 通过生成类型消费真实描述、生命周期、Lua 结果及业务错误。
func TestEmbeddedWireNativeRoundTrip(t *testing.T) {
	transport, runtimeID, command, pool := nativeEmbeddedRuntime(t)
	description, err := DecodeEmbeddedOutputRootDescribeResponse(embeddedTypedNative(t, transport, EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe}))
	if err != nil || description.Result.Limits.MaxResultBytes != transport.Config().MaxResultBytes {
		t.Fatalf("native description projection: %#v %v", description, err)
	}
	status, err := DecodeEmbeddedOutputRootRuntimeStatusResponse(embeddedTypedNative(t, transport, EmbeddedInputCommandRuntimeStatus{Type: EmbeddedInputCommandRuntimeStatusTypeRuntimeStatus, RuntimeId: runtimeID}))
	if err != nil || status.Result.Initialization != EmbeddedOutputInitializationPhaseReady || status.Result.CoreRuntimeId == nil {
		t.Fatalf("native status projection: %#v %v", status, err)
	}
	poolID := pool("return {call=function(a) return a end}")
	for _, value := range []any{nil, false, json.Number("0.25"), "中文"} {
		call := EmbeddedInputRuntimeCommandCallSubmit{Type: EmbeddedInputRuntimeCommandCallSubmitTypeCallSubmit, TimeoutMs: 1000, Call: EmbeddedInputEmbeddedCall{PoolId: poolID, Export: "call", Arguments: value, Context: EmbeddedInputLuaInvocationContext{}}}
		receipt, err := DecodeEmbeddedOutputRuntimeCallSubmitResponse(embeddedTypedNative(t, transport, EmbeddedInputCommandRuntime{Type: EmbeddedInputCommandRuntimeTypeRuntime, RuntimeId: runtimeID, Operation: call}))
		if err != nil {
			t.Fatal(err)
		}
		finished := embeddedPoll(t, func() any {
			response, err := DecodeEmbeddedOutputRuntimeOperationStatusResponse(embeddedTypedNative(t, transport, EmbeddedInputCommandRuntime{Type: EmbeddedInputCommandRuntimeTypeRuntime, RuntimeId: runtimeID, Operation: EmbeddedInputRuntimeCommandOperationStatus{Type: EmbeddedInputRuntimeCommandOperationStatusTypeOperationStatus, OperationId: receipt.Result.OperationId}}))
			if err != nil {
				t.Fatal(err)
			}
			return response.Result
		}, func(value any) bool {
			phase := value.(EmbeddedOutputOperationSnapshot).Phase
			return phase == EmbeddedOutputOperationPhaseSucceeded || phase == EmbeddedOutputOperationPhaseFailed || phase == EmbeddedOutputOperationPhaseCancelled
		}).(EmbeddedOutputOperationSnapshot)
		if finished.Phase != EmbeddedOutputOperationPhaseSucceeded || finished.Value == nil || !reflect.DeepEqual(*finished.Value, value) {
			t.Fatalf("native Lua result changed: %#v", finished)
		}
		command(map[string]any{"type": "operation_forget", "operation_id": receipt.Result.OperationId})
	}
	failure, err := DecodeEmbeddedOutputErrorResponse(embeddedTypedNative(t, transport, EmbeddedInputCommandRuntimeStatus{Type: EmbeddedInputCommandRuntimeStatusTypeRuntimeStatus, RuntimeId: "missing-runtime"}))
	if err != nil || failure.Error.Code != EmbeddedOutputEmbeddedErrorCodeNotFound {
		t.Fatalf("native business error changed: %#v %v", failure, err)
	}
}
