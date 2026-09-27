package luaskills

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"testing"
)

// forbiddenEmbeddedMarshaler panics if generic serialization hooks are accidentally invoked.
// forbiddenEmbeddedMarshaler 在误调用通用序列化钩子时引发 panic。
type forbiddenEmbeddedMarshaler struct{}

// MarshalJSON must never run through the plain-data embedded encoder.
// MarshalJSON 绝不能通过纯数据嵌入式编码器执行。
func (forbiddenEmbeddedMarshaler) MarshalJSON() ([]byte, error) { panic("serialization hook invoked") }

// TestEmbeddedJSONNumbers verifies complete native integer widths and explicit float wire intent.
// TestEmbeddedJSONNumbers 验证完整原生整数位宽及显式浮点线意图。
func TestEmbeddedJSONNumbers(t *testing.T) {
	values := []any{int64(math.MinInt64), uint64(math.MaxUint64), uint64(9007199254740993), float64(1), math.Copysign(0, -1), float64(1e100), json.Number("1.00")}
	encoded, err := EncodeEmbeddedJSON(values, 1024)
	if err != nil {
		t.Fatal(err)
	}
	expected := `[-9223372036854775808,18446744073709551615,9007199254740993,1.0,-0.0,1e+100,1.00]`
	if string(encoded) != expected {
		t.Fatalf("numeric wire changed: %s", encoded)
	}
	decoded, err := DecodeEmbeddedJSON(encoded)
	if err != nil {
		t.Fatal(err)
	}
	roundtrip, err := EncodeEmbeddedJSON(decoded, 1024)
	if err != nil || !bytes.Equal(encoded, roundtrip) {
		t.Fatalf("number tokens lost: %s %v", roundtrip, err)
	}
	for _, value := range []any{math.NaN(), math.Inf(1), json.Number("18446744073709551616"), json.Number("-9223372036854775809"), json.Number("1e400"), json.Number("01"), json.Number("null")} {
		if _, err := EncodeEmbeddedJSON(value, 1024); err == nil {
			t.Fatalf("accepted invalid number %v", value)
		}
	}
}

// TestEmbeddedJSONStrictInput checks duplicate decoded keys, UTF-8 and surrogate errors before publishing data.
// TestEmbeddedJSONStrictInput 在发布数据前检查解码后重复键、UTF-8 及代理项错误。
func TestEmbeddedJSONStrictInput(t *testing.T) {
	for _, input := range []string{`{"x":1,"\u0078":2}`, `"\ud800"`, `"\udc00"`, `"\ud800\u0041"`, `[1]null`, "\xef\xbb\xbfnull", "\"\xff\"", `1e400`, `18446744073709551616`, `-9223372036854775809`} {
		if _, err := DecodeEmbeddedJSON([]byte(input)); err == nil {
			t.Fatalf("accepted invalid JSON %q", input)
		}
	}
	decoded, err := DecodeEmbeddedJSON([]byte(`{"pair":"\ud83d\ude00","literal":"\\ud800","null":null,"false":false,"empty":[],"__proto__":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	object := decoded.(map[string]any)
	if object["pair"] != "😀" || object["literal"] != "\\ud800" || object["null"] != nil || object["false"] != false {
		t.Fatalf("lost explicit values: %#v", object)
	}
	if _, present := object["missing"]; present {
		t.Fatal("missing property became explicit")
	}
	if !reflect.DeepEqual(object["empty"], []any{}) {
		t.Fatalf("empty array changed: %#v", object["empty"])
	}
}

// TestEmbeddedJSONPlainDataAndBounds rejects hooks and cycles while preserving shared acyclic values and exact limits.
// TestEmbeddedJSONPlainDataAndBounds 拒绝钩子及循环，同时保留非循环共享值及精确限制。
func TestEmbeddedJSONPlainDataAndBounds(t *testing.T) {
	cycleMap := map[string]any{}
	cycleMap["self"] = cycleMap
	cycleSlice := make([]any, 1)
	cycleSlice[0] = cycleSlice
	for _, value := range []any{forbiddenEmbeddedMarshaler{}, []byte("bytes"), map[int]any{1: nil}, cycleMap, cycleSlice, "\xff"} {
		if _, err := EncodeEmbeddedJSON(value, 1024); err == nil {
			t.Fatalf("accepted unsupported value %T", value)
		}
	}
	shared := map[string]any{"ok": true}
	if _, err := EncodeEmbeddedJSON([]any{shared, shared}, 1024); err != nil {
		t.Fatal(err)
	}
	data, err := EncodeEmbeddedJSON("中文\n", 1024)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := EncodeEmbeddedJSON("中文\n", uint64(len(data))); err != nil {
		t.Fatal(err)
	}
	if _, err := EncodeEmbeddedJSON("中文\n", uint64(len(data)-1)); err == nil {
		t.Fatal("byte limit ignored")
	}
}

// TestEmbeddedResponseEnvelope distinguishes business errors, explicit null and protocol/shape mismatches.
// TestEmbeddedResponseEnvelope 区分业务错误、显式空值及协议／形状不匹配。
func TestEmbeddedResponseEnvelope(t *testing.T) {
	result, err := DecodeEmbeddedResponse([]byte(`{"protocol_version":1,"status":"ok","result":null}`))
	if err != nil || result != nil {
		t.Fatalf("null delivery lost: %v %v", result, err)
	}
	_, err = DecodeEmbeddedResponse([]byte(`{"protocol_version":1,"status":"error","error":{"code":"busy","message":"still active"}}`))
	var business *EmbeddedRuntimeError
	if !errors.As(err, &business) || business.Code != "busy" {
		t.Fatalf("lost business identity: %v", err)
	}
	for _, input := range []string{`{"protocol_version":2,"status":"ok","result":null}`, `{"protocol_version":1.0,"status":"ok","result":null}`, `{"protocol_version":1,"status":"ok"}`, `{"protocol_version":1,"status":"ok","result":null,"extra":false}`, `{"protocol_version":1,"status":"error","error":{"code":"busy"}}`} {
		if _, err := DecodeEmbeddedResponse([]byte(input)); err == nil {
			t.Fatalf("accepted invalid envelope %s", input)
		}
	}
}

// TestEmbeddedJSONSDKOptionShapes verifies tagged engine options, typed slices and pointer absence without false cycles.
// TestEmbeddedJSONSDKOptionShapes 验证标签化引擎选项、类型化切片及指针缺失，不误判循环。
func TestEmbeddedJSONSDKOptionShapes(t *testing.T) {
	options, err := CreateEngineOptions(t.TempDir(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := EncodeEmbeddedJSON(options, 65536)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeEmbeddedJSON(encoded)
	if err != nil {
		t.Fatal(err)
	}
	host := decoded.(map[string]any)["host_options"].(map[string]any)
	if !reflect.DeepEqual(host["ignored_skill_ids"], []any{}) {
		t.Fatalf("empty typed slice changed: %#v", host)
	}
	if host["managed_runtime_config"].(map[string]any)["invoke_default_timeout_ms"] != nil {
		t.Fatal("nil option changed")
	}
	// names exercises a named slice type through the same data traversal.
	// names 通过同一数据遍历覆盖命名切片类型。
	type names []any
	cycle := make(names, 1)
	cycle[0] = cycle
	if _, err := EncodeEmbeddedJSON(cycle, 1024); err == nil {
		t.Fatal("typed slice cycle accepted")
	}
	// record models explicit SDK field tags and pointer optionality.
	// record 表达显式 SDK 字段标签及指针可选性。
	type record struct {
		// Names retains the declared string-array shape.
		// Names 保留已声明字符串数组形状。
		Names []string `json:"names"`
		// Optional is absent when nil.
		// Optional 在为空指针时缺失。
		Optional *uint64 `json:"optional,omitempty"`
		// Empty preserves an explicit empty array.
		// Empty 保留显式空数组。
		Empty []string `json:"empty"`
	}
	if encoded, err := EncodeEmbeddedJSON(record{Names: []string{"a"}, Empty: []string{}}, 1024); err != nil || string(encoded) != `{"empty":[],"names":["a"]}` {
		t.Fatalf("tagged data changed: %s %v", encoded, err)
	}
}

// TestEmbeddedJSONPointerViews keeps distinct typed pointer views that share an address while rejecting real cycles.
// TestEmbeddedJSONPointerViews 保留地址相同但类型不同的指针视图，同时拒绝真实循环。
func TestEmbeddedJSONPointerViews(t *testing.T) {
	// record's first field shares its address with the whole record; that does not make Alias a recursive record.
	// record 的首字段与整个记录地址相同；这不意味着 Alias 是递归记录。
	type record struct {
		// Value remains an ordinary signed integer.
		// Value 保持普通有符号整数。
		Value int `json:"value"`
		// Alias points to Value, without referring back to the record.
		// Alias 指向 Value，不反向引用记录。
		Alias *int `json:"alias"`
	}
	value := &record{Value: 7}
	value.Alias = &value.Value
	encoded, err := EncodeEmbeddedJSON(value, 1024)
	if err != nil || string(encoded) != `{"alias":7,"value":7}` {
		t.Fatalf("acyclic pointer views rejected: %s %v", encoded, err)
	}
	// recursive explicitly links back to itself and must still fail.
	// recursive 显式反向引用自身，仍然必须失败。
	type recursive struct {
		// Next owns no new node and may form a cycle.
		// Next 不拥有新节点，并可能形成循环。
		Next *recursive `json:"next"`
	}
	cycle := &recursive{}
	cycle.Next = cycle
	if _, err := EncodeEmbeddedJSON(cycle, 1024); err == nil {
		t.Fatal("recursive pointer accepted")
	}
}

// TestEmbeddedJSONOmittedFieldBudget applies byte limits to emitted fields, including exact empty-object boundaries.
// TestEmbeddedJSONOmittedFieldBudget 将字节限制应用于输出字段，包含空对象的精确边界。
func TestEmbeddedJSONOmittedFieldBudget(t *testing.T) {
	// record's absent or excluded fields contribute no JSON members.
	// record 的缺失或排除字段不产生 JSON 成员。
	type record struct {
		// First is omitted when empty.
		// First 在为空时省略。
		First string `json:"first,omitempty"`
		// Second is omitted when nil.
		// Second 在为空指针时省略。
		Second *int `json:"second,omitempty"`
		// Internal is excluded even when it has data.
		// Internal 即使有数据也排除。
		Internal string `json:"-"`
	}
	encoded, err := EncodeEmbeddedJSON(record{Internal: "unused"}, 2)
	if err != nil || string(encoded) != `{}` {
		t.Fatalf("omitted fields incorrectly consume byte budget: %s %v", encoded, err)
	}
	if _, err := EncodeEmbeddedJSON(record{}, 1); err == nil {
		t.Fatal("empty object exceeded exact byte budget")
	}
}
