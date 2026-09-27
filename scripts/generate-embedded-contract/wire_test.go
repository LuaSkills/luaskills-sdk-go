package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// wireFixture loads only the SDK's packaged contract and returns mutable JSON for deliberate generator drift tests.
// wireFixture 仅加载 SDK 包内契约，并返回可变 JSON 以进行明确生成器漂移测试。
func wireFixture(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "contracts", "embedded", "v1", "contract.json"))
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result
}

// TestEmbeddedWireGeneration verifies deterministic complete source against the committed artifact.
// TestEmbeddedWireGeneration 对照已保存产物验证确定性的完整源码。
func TestEmbeddedWireGeneration(t *testing.T) {
	contract := wireFixture(t)
	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := generateWire(data)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(filepath.Join("..", "..", "embedded_wire_generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(generated, saved) {
		t.Fatal("generated wire source differs")
	}
	// Error envelopes and the standalone core description each require their own decoder.
	// 错误信封和独立核心描述各自需要对应的解码器。
	standalone := []string{"error_response", "core_description"}
	for _, name := range standalone {
		if _, exists := contract[name]; !exists {
			t.Fatalf("missing required independent output schema: %s", name)
		}
	}
	responses := len(standalone) + len(contract["root_responses"].(map[string]any)) + len(contract["runtime_responses"].(map[string]any))
	if strings.Count(string(generated), "func DecodeEmbeddedOutput") != responses {
		t.Fatal("response decoder coverage changed; compare current contract")
	}
}

// TestEmbeddedWireGenerationRejectsDrift ensures unsupported shapes and namespace changes fail before artifact writes.
// TestEmbeddedWireGenerationRejectsDrift 确保不支持形状及命名空间变化在产物写入前失败。
func TestEmbeddedWireGenerationRejectsDrift(t *testing.T) {
	for _, mutation := range []func(map[string]any){
		func(c map[string]any) {
			for _, branch := range c["request"].(map[string]any)["$defs"].(map[string]any)["HostCompletion"].(map[string]any)["anyOf"].([]any) {
				branch.(map[string]any)["additionalProperties"] = true
			}
		},
		func(c map[string]any) { c["request"] = "invalid" },
		func(c map[string]any) { c["generator"].(map[string]any)["schema_draft"] = "unsupported" },
		func(c map[string]any) {
			defs := c["request"].(map[string]any)["$defs"].(map[string]any)
			defs["pool_kind"] = defs["PoolKind"]
		},
		func(c map[string]any) {
			defs := c["request"].(map[string]any)["$defs"].(map[string]any)
			defs["PoolKindShared"] = defs["PoolKind"]
		},
		func(c map[string]any) { c["request"].(map[string]any)["not"] = map[string]any{} },
		func(c map[string]any) {
			c["request"].(map[string]any)["properties"].(map[string]any)["command"].(map[string]any)["$ref"] = "#/$defs/Absent"
		},
		func(c map[string]any) { c["commands"] = []any{"describe"} },
		func(c map[string]any) {
			c["runtime_responses"].(map[string]any)["operation_wait"].(map[string]any)["$defs"].(map[string]any)["OperationSnapshot"].(map[string]any)["description"] = "conflict"
		},
		func(c map[string]any) {
			c["request"].(map[string]any)["$defs"].(map[string]any)["PoolKind"].(map[string]any)["oneOf"].([]any)[0].(map[string]any)["const"] = "not-valid"
		},
	} {
		contract := wireFixture(t)
		mutation(contract)
		data, err := json.Marshal(contract)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := generateWire(data); err == nil {
			t.Fatal("unsupported contract drift was accepted")
		}
	}
}
