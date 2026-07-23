package luaskills

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestSkillPackageConfigPayloads verifies exact request ownership and explicit value disclosure.
// TestSkillPackageConfigPayloads 校验精确请求归属和显式值披露。
//
// t receives assertion failures from the Go test runner.
// t 接收 Go 测试运行器提供的断言失败。
func TestSkillPackageConfigPayloads(t *testing.T) {
	// defaultPayload verifies absent optional controls are not converted into ambiguous empty wire values.
	// defaultPayload 校验缺失的可选控制项不会转换为含义模糊的空线协议值。
	defaultPayload := skillPackageConfigDescribePayload(7, SkillPackageConfigDescribeOptions{})
	defaultExpected := map[string]any{
		"engine_id":      uint64(7),
		"include_values": false,
		"mode":           "effective",
	}
	if !reflect.DeepEqual(defaultPayload, defaultExpected) {
		t.Fatalf("unexpected default describe payload: %#v", defaultPayload)
	}

	// explicitPayload verifies package and unmasked-value controls use the core FFI field names.
	// explicitPayload 校验技能包和未脱敏值控制项使用核心 FFI 字段名。
	explicitPayload := skillPackageConfigDescribePayload(9, SkillPackageConfigDescribeOptions{
		SkillID:       "example.settings",
		IncludeValues: true,
	})
	explicitExpected := map[string]any{
		"engine_id":      uint64(9),
		"skill_id":       "example.settings",
		"include_values": true,
		"mode":           "effective",
	}
	if !reflect.DeepEqual(explicitPayload, explicitExpected) {
		t.Fatalf("unexpected explicit describe payload: %#v", explicitPayload)
	}

	// validatePayload verifies validation is scoped to one exact skill package identifier.
	// validatePayload 校验完整性校验限定到一个精确技能包标识。
	validatePayload := skillPackageConfigValidatePayload(11, "example.settings")
	validateExpected := map[string]any{
		"engine_id": uint64(11),
		"skill_id":  "example.settings",
	}
	if !reflect.DeepEqual(validatePayload, validateExpected) {
		t.Fatalf("unexpected validate payload: %#v", validatePayload)
	}
}

// TestSkillPackageConfigDescriptorDecoding verifies typed responses preserve integer constraint precision.
// TestSkillPackageConfigDescriptorDecoding 校验类型化响应能够保留整数约束精度。
//
// t receives assertion failures from the Go test runner.
// t 接收 Go 测试运行器提供的断言失败。
func TestSkillPackageConfigDescriptorDecoding(t *testing.T) {
	// rawDescriptor represents the public JSON FFI response, including a value disclosed by host choice.
	// rawDescriptor 表示公共 JSON FFI 响应，其中包含由宿主选择披露的值。
	rawDescriptor := []byte(`{
		"skill_id":"example.settings",
		"skill_version":"1.2.3",
		"complete":true,
		"orphaned_count":0,
		"items":[{
			"key":"retry_count",
			"type":"integer",
			"required":true,
			"sensitive":false,
			"description":"Retry count",
			"constraints":{"minimum":-9007199254740991,"maximum":9007199254740991},
			"options":[],
			"default":3,
			"advanced":false,
			"restart_required":false,
			"deprecated":false,
			"state":"configured",
			"satisfied":true,
			"value":"4"
		}]
	}`)
	// descriptor is the strongly typed SDK representation decoded from the response.
	// descriptor 是从响应解码得到的 SDK 强类型表示。
	var descriptor SkillPackageConfigDescriptor
	if err := json.Unmarshal(rawDescriptor, &descriptor); err != nil {
		t.Fatalf("decode descriptor: %v", err)
	}
	if descriptor.Items[0].Type != SkillPackageConfigTypeInteger {
		t.Fatalf("unexpected item type: %s", descriptor.Items[0].Type)
	}
	if got := descriptor.Items[0].Constraints.Minimum.String(); got != "-9007199254740991" {
		t.Fatalf("minimum precision changed: %s", got)
	}
	if got := descriptor.Items[0].Constraints.Maximum.String(); got != "9007199254740991" {
		t.Fatalf("maximum precision changed: %s", got)
	}
	if descriptor.Items[0].Value == nil || *descriptor.Items[0].Value != "4" {
		t.Fatalf("unexpected disclosed value: %#v", descriptor.Items[0].Value)
	}
}

// TestSkillConfigNumericValidation distinguishes bounded integers from finite binary64 floats.
// TestSkillConfigNumericValidation 区分有界整数与有限 binary64 浮点数。
//
// t receives assertion failures from the Go test runner.
// t 接收 Go 测试运行器提供的断言失败。
func TestSkillConfigNumericValidation(t *testing.T) {
	if err := validateSkillConfigValue("integer", int64(SkillConfigMaximumSafeInteger)+1); err == nil {
		t.Fatal("unsafe integer must be rejected")
	}
	if err := validateSkillConfigValue("float", float64(1e20)); err != nil {
		t.Fatalf("finite large float must remain valid: %v", err)
	}
	if err := validateSkillConfigValue("float", json.Number("1e20")); err != nil {
		t.Fatalf("exponent-form JSON float must remain valid: %v", err)
	}
	if err := validateSkillConfigValue("integer", json.Number("9007199254740992")); err == nil {
		t.Fatal("out-of-range integer-form JSON number must be rejected")
	}
}
