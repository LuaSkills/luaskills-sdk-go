package luaskills

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestEmbeddedWireInputVariants validates sealed generated commands and keeps successful null callback values present.
// TestEmbeddedWireInputVariants 校验封闭生成命令，并保持成功空值回调结果存在。
func TestEmbeddedWireInputVariants(t *testing.T) {
	request := EmbeddedInputRequest{ProtocolVersion: EmbeddedProtocolVersion, Command: EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe}}
	bytes, err := EncodeEmbeddedRequest(request, 1024)
	if err != nil || string(bytes) != `{"command":{"type":"describe"},"protocol_version":1}` {
		t.Fatalf("typed describe: %s %v", bytes, err)
	}
	request.Command = EmbeddedInputCommandRuntime{Type: EmbeddedInputCommandRuntimeTypeRuntime, RuntimeId: "runtime", Operation: EmbeddedInputRuntimeCommandHostRequestComplete{Type: EmbeddedInputRuntimeCommandHostRequestCompleteTypeHostRequestComplete, RequestId: "request", Outcome: EmbeddedInputHostCompletionVariant1{Effects: EmbeddedInputEffectStateCommitted, Ok: true, Value: nil}}}
	bytes, err = EncodeEmbeddedRequest(request, 4096)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeEmbeddedJSON(bytes)
	if err != nil {
		t.Fatal(err)
	}
	outcome := decoded.(map[string]any)["command"].(map[string]any)["operation"].(map[string]any)["outcome"].(map[string]any)
	if value, present := outcome["value"]; !present || value != nil || outcome["ok"] != true {
		t.Fatalf("null completion changed: %#v", outcome)
	}
	for _, command := range []EmbeddedInputCommand{nil, EmbeddedInputCommandDescribe{}, EmbeddedInputCommandRuntime{Type: EmbeddedInputCommandRuntimeTypeRuntime}, EmbeddedInputCommandDescribe{Type: "unknown"}} {
		request.Command = command
		if _, err := EncodeEmbeddedRequest(request, 4096); err == nil {
			t.Fatalf("invalid typed command accepted: %T", command)
		}
	}
	request.ProtocolVersion = 2
	if _, err := EncodeEmbeddedRequest(request, 4096); err == nil {
		t.Fatal("unknown protocol accepted")
	}
}

// TestEmbeddedWireOptionalPresence distinguishes absence, explicit null, false and zero in real operation snapshots.
// TestEmbeddedWireOptionalPresence 在真实操作快照形状中区分缺失、显式空值、假值和零。
func TestEmbeddedWireOptionalPresence(t *testing.T) {
	base := `{"protocol_version":1,"status":"ok","result":{"operation_id":"op","phase":"succeeded","cancellation_requested":false,"effects":"not_applicable","host_effects":[]`
	for _, value := range []struct {
		suffix   string
		present  bool
		expected any
	}{{"", false, nil}, {`,"value":null,"error":null`, true, nil}, {`,"value":false`, true, false}, {`,"value":0`, true, json.Number("0")}} {
		response, err := DecodeEmbeddedOutputRuntimeOperationStatusResponse([]byte(base + value.suffix + `}}`))
		if err != nil {
			t.Fatal(err)
		}
		if (response.Result.Value != nil) != value.present {
			t.Fatal("optional presence lost")
		}
		if value.present && !reflect.DeepEqual(*response.Result.Value, value.expected) {
			t.Fatalf("optional value changed: %#v", *response.Result.Value)
		}
		if value.suffix == `,"value":null,"error":null` && (response.Result.Error == nil || *response.Result.Error != nil) {
			t.Fatal("explicit nullable error lost")
		}
		encoded, err := EncodeEmbeddedJSON(response, 4096)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeEmbeddedJSON(encoded)
		if err != nil {
			t.Fatal(err)
		}
		result := decoded.(map[string]any)["result"].(map[string]any)
		if _, present := result["value"]; present != value.present {
			t.Fatal("re-encoding changed field presence")
		}
	}
}

// TestEmbeddedWireMalformedOutput rejects required-field, enum, case and null violations instead of inventing Go zero values.
// TestEmbeddedWireMalformedOutput 拒绝必需字段、枚举、大小写及空值违规，不凭空生成 Go 零值。
func TestEmbeddedWireMalformedOutput(t *testing.T) {
	for _, body := range []string{`{}`, `{"runtime_id":null}`, `{"Runtime_Id":"x"}`, `{"runtime_id":1}`, `{"runtime_id":"x","extra":true}`} {
		// RuntimeReceipt allows additional properties in the upstream schema; only this exact extension is valid.
		// 上游 RuntimeReceipt 允许额外属性；这里只允许此精确扩展用例。
		_, err := DecodeEmbeddedOutputRootRuntimeReserveResponse([]byte(`{"protocol_version":1,"status":"ok","result":` + body + `}`))
		if body == `{"runtime_id":"x","extra":true}` {
			if err != nil {
				t.Fatal(err)
			}
		} else if err == nil {
			t.Fatalf("invalid receipt accepted: %s", body)
		}
	}
	for _, body := range []string{`{"protocol_version":1,"status":"other","result":null}`, `{"protocol_version":1,"status":"ok","result":{}}`, `{"protocol_version":1.0,"status":"ok","result":null}`, `{"protocol_version":1,"status":"ok","result":null,"extra":true}`} {
		if _, err := DecodeEmbeddedOutputRuntimePluginRegisterResponse([]byte(body)); err == nil {
			t.Fatalf("invalid null response accepted: %s", body)
		}
	}
	if _, err := DecodeEmbeddedOutputRuntimePluginRegisterResponse([]byte(`{"protocol_version":1,"status":"ok","result":null}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeEmbeddedOutputRuntimeOperationStatusResponse([]byte(`{"protocol_version":1,"status":"ok","result":{"operation_id":"op","phase":"invented","cancellation_requested":false,"effects":"unknown","host_effects":[]}}`)); err == nil {
		t.Fatal("unknown operation phase accepted")
	}
}

// TestEmbeddedWireUnsignedRanges projects full-width values and rejects signed, fractional and overflowing integer tokens.
// TestEmbeddedWireUnsignedRanges 投影完整位宽值，并拒绝有符号、分数及溢出整数词元。
func TestEmbeddedWireUnsignedRanges(t *testing.T) {
	for _, value := range []string{"0", "9007199254740993", "18446744073709551615"} {
		projected, err := projectEmbeddedWire(json.Number(value), embeddedWireType[uint64](), "budget")
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := EncodeEmbeddedJSON(projected.Interface(), 128)
		if err != nil || string(encoded) != value {
			t.Fatalf("integer changed: %s %v", encoded, err)
		}
	}
	for _, value := range []string{"-1", "1.0", "1e1", "18446744073709551616"} {
		if _, err := projectEmbeddedWire(json.Number(value), embeddedWireType[uint64](), "budget"); err == nil {
			t.Fatalf("invalid integer accepted %s", value)
		}
	}
	if _, err := projectEmbeddedWire(json.Number("4294967296"), embeddedWireType[uint32](), "version"); err == nil {
		t.Fatal("uint32 overflow accepted")
	}
}

// TestEmbeddedWireUniqueArrays enforces declared set semantics without treating all ordinary arrays as sets.
// TestEmbeddedWireUniqueArrays 执行已声明集合语义，不把所有普通数组当作集合。
func TestEmbeddedWireUniqueArrays(t *testing.T) {
	values := []any{"same", "same"}
	if _, err := projectEmbeddedWire(values, embeddedWireType[EmbeddedInputRuntimeCommandCapabilitiesListPermissions](), "permissions"); err == nil {
		t.Fatal("duplicate permission accepted")
	}
	if _, err := projectEmbeddedWire(values, embeddedWireType[EmbeddedOutputTransportDescriptionCommands](), "commands"); err != nil {
		t.Fatal(err)
	}
}

// TestEmbeddedWireErrorEnvelope preserves structured business errors without treating them as native ABI failures.
// TestEmbeddedWireErrorEnvelope 保留结构化业务错误，不将其当作原生 ABI 失败。
func TestEmbeddedWireErrorEnvelope(t *testing.T) {
	bytes := []byte(`{"protocol_version":1,"status":"error","error":{"code":"busy","message":"still active"}}`)
	response, err := DecodeEmbeddedOutputErrorResponse(bytes)
	if err != nil || response.Error.Code != EmbeddedOutputEmbeddedErrorCodeBusy {
		t.Fatalf("error envelope changed: %#v %v", response, err)
	}
	if _, err := DecodeEmbeddedOutputRootRuntimeReserveResponse(bytes); err == nil {
		t.Fatal("business error projected as a successful reservation")
	}
}
