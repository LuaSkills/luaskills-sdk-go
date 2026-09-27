package luaskills

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// embeddedJSONVector records one core-owned JSON spelling and its independent semantic expectation.
// embeddedJSONVector 记录一个核心拥有的 JSON 写法及其独立语义期望。
type embeddedJSONVector struct {
	// ID is the stable case name shared with Rust and the other SDKs.
	// ID 是与 Rust 及其他 SDK 共享的稳定用例名称。
	ID string `json:"id"`
	// JSON is exact JSON text, including deliberate syntax errors in negative cases.
	// JSON 是精确 JSON 文本，包含反例中刻意的语法错误。
	JSON string `json:"json"`
	// Hex preserves bytes that cannot appear in a UTF-8 JSON document.
	// Hex 保留无法出现在 UTF-8 JSON 文档中的字节。
	Hex string `json:"hex"`
	// Expected preserves integer strings and hexadecimal IEEE binary64 bits.
	// Expected 保留整数字符串及十六进制 IEEE 双精度浮点位。
	Expected any `json:"expected"`
}

// embeddedJSONFingerprint maps a decoded value to its semantic evidence without rounding integer tokens.
// embeddedJSONFingerprint 将已解码值映射为语义证据，不舍入整数字面量。
func embeddedJSONFingerprint(t *testing.T, value any) any {
	t.Helper()
	switch value := value.(type) {
	case nil:
		return []any{"null"}
	case bool:
		return []any{"boolean", value}
	case string:
		return []any{"string", value}
	case json.Number:
		if !strings.ContainsAny(string(value), ".eE") && value != "-0" {
			return []any{"integer", string(value)}
		}
		number, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			t.Fatal(err)
		}
		return []any{"float", fmt.Sprintf("%016x", math.Float64bits(number))}
	case []any:
		children := make([]any, len(value))
		for index, child := range value {
			children[index] = embeddedJSONFingerprint(t, child)
		}
		return []any{"array", children}
	case map[string]any:
		children := make(map[string]any, len(value))
		for key, child := range value {
			children[key] = embeddedJSONFingerprint(t, child)
		}
		return []any{"object", children}
	default:
		t.Fatalf("unsupported decoded value type %T", value)
		return nil
	}
}

// TestEmbeddedSharedJSONVectors checks the packaged core corpus through public codecs and response envelopes.
// TestEmbeddedSharedJSONVectors 通过公开编码器及响应信封检查包内核心语料。
func TestEmbeddedSharedJSONVectors(t *testing.T) {
	// The digest-checked contract owns these fixtures; no adjacent core checkout is required.
	// 经摘要校验的契约拥有这些夹具；不需要相邻核心检出。
	data, err := os.ReadFile("contracts/embedded/v1/contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract struct {
		// Vectors selects the corpus from the same contract used for SDK type generation.
		// Vectors 从 SDK 类型生成所用的同一契约选择语料。
		Vectors struct {
			// Version identifies the semantic-vector container format.
			// Version 标识语义向量容器格式。
			Version int `json:"version"`
			// Valid contains exact accepted values and independent fingerprints.
			// Valid 包含精确接受值及独立指纹。
			Valid []embeddedJSONVector `json:"valid"`
			// Invalid contains malformed or unrepresentable JSON text.
			// Invalid 包含畸形或不可表示 JSON 文本。
			Invalid []embeddedJSONVector `json:"invalid"`
			// Bytes contains invalid UTF-8 sequences.
			// Bytes 包含无效 UTF-8 序列。
			Bytes []embeddedJSONVector `json:"invalid_bytes"`
			// Envelopes contains structurally invalid protocol envelopes.
			// Envelopes 包含结构无效的协议信封。
			Envelopes []embeddedJSONVector `json:"invalid_envelopes"`
		} `json:"json_vectors"`
	}
	if err = json.Unmarshal(data, &contract); err != nil {
		t.Fatal(err)
	}
	if contract.Vectors.Version != 1 {
		t.Fatal("unsupported shared JSON vector version")
	}
	for _, entry := range contract.Vectors.Valid {
		t.Run("valid/"+entry.ID, func(t *testing.T) {
			value, err := DecodeEmbeddedJSON([]byte(entry.JSON))
			if err != nil {
				t.Fatal(err)
			}
			if actual := embeddedJSONFingerprint(t, value); !reflect.DeepEqual(actual, entry.Expected) {
				t.Fatalf("fingerprint = %#v; want %#v", actual, entry.Expected)
			}
			encoded, err := EncodeEmbeddedJSON(value, 4096)
			if err != nil {
				t.Fatal(err)
			}
			roundtrip, err := DecodeEmbeddedJSON(encoded)
			if err != nil || !reflect.DeepEqual(embeddedJSONFingerprint(t, roundtrip), entry.Expected) {
				t.Fatalf("roundtrip = %v, %v", roundtrip, err)
			}
			if exact, err := EncodeEmbeddedJSON(value, uint64(len(encoded))); err != nil || string(exact) != string(encoded) {
				t.Fatalf("exact byte budget failed: %v", err)
			}
			if _, err := EncodeEmbeddedJSON(value, uint64(len(encoded)-1)); err == nil {
				t.Fatal("insufficient byte budget accepted")
			}
			response, err := DecodeEmbeddedResponse([]byte(`{"protocol_version":1,"status":"ok","result":` + entry.JSON + `}`))
			if err != nil || !reflect.DeepEqual(embeddedJSONFingerprint(t, response), entry.Expected) {
				t.Fatalf("response = %v, %v", response, err)
			}
		})
	}
	for _, entry := range contract.Vectors.Invalid {
		t.Run("invalid/"+entry.ID, func(t *testing.T) {
			if _, err := DecodeEmbeddedJSON([]byte(entry.JSON)); err == nil {
				t.Fatal("invalid JSON accepted")
			}
			if _, err := DecodeEmbeddedResponse([]byte(`{"protocol_version":1,"status":"ok","result":` + entry.JSON + `}`)); err == nil {
				t.Fatal("invalid response value accepted")
			}
		})
	}
	for _, entry := range contract.Vectors.Bytes {
		t.Run("bytes/"+entry.ID, func(t *testing.T) {
			bytes, err := hex.DecodeString(entry.Hex)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = DecodeEmbeddedJSON(bytes); err == nil {
				t.Fatal("invalid UTF-8 accepted")
			}
		})
	}
	for _, entry := range contract.Vectors.Envelopes {
		t.Run("envelope/"+entry.ID, func(t *testing.T) {
			if _, err := DecodeEmbeddedResponse([]byte(entry.JSON)); err == nil {
				t.Fatal("invalid envelope accepted")
			}
		})
	}
}
