package luaskills

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// skillConfigContractDocument contains the canonical fields consumed by this SDK.
// skillConfigContractDocument 包含此 SDK 使用的规范字段。
type skillConfigContractDocument struct {
	// ContractVersion is the stable wire contract version.
	// ContractVersion 是稳定的线协议版本。
	ContractVersion int `json:"contract_version"`
	// Declaration contains stable enum-like declaration values.
	// Declaration 包含稳定的类枚举声明值。
	Declaration struct {
		// Types contains declared scalar types in contract order.
		// Types 按契约顺序包含已声明标量类型。
		Types []string `json:"types"`
		// Formats contains host rendering formats in contract order.
		// Formats 按契约顺序包含宿主渲染格式。
		Formats []string `json:"formats"`
		// States contains effective-value states in contract order.
		// States 按契约顺序包含有效值状态。
		States []string `json:"states"`
		// DescribeModes contains declaration discovery modes in contract order.
		// DescribeModes 按契约顺序包含声明发现模式。
		DescribeModes []string `json:"describe_modes"`
		// StoreScopes contains routed configuration stores in contract order.
		// StoreScopes 按契约顺序包含路由后的配置存储作用域。
		StoreScopes []string `json:"store_scopes"`
	} `json:"declaration"`
	// Limits contains cross-SDK numeric boundaries.
	// Limits 包含跨 SDK 数值边界。
	Limits struct {
		// MaximumSafeInteger is the common exact integer boundary.
		// MaximumSafeInteger 是公共精确整数边界。
		MaximumSafeInteger int64 `json:"maximum_safe_integer"`
		// MaximumEventPollLimit is the maximum accepted event page.
		// MaximumEventPollLimit 是允许的最大事件分页。
		MaximumEventPollLimit uint64 `json:"maximum_event_poll_limit"`
	} `json:"limits"`
	// Errors contains stable machine-readable error codes in contract order.
	// Errors 按契约顺序包含稳定机器可读错误码。
	Errors []string `json:"errors"`
}

// TestSkillConfigContractConstants verifies exported Go values match the core-generated contract.
// TestSkillConfigContractConstants 验证导出的 Go 值与核心生成契约一致。
//
// t receives assertion failures from the Go test runner.
// t 接收 Go 测试运行器提供的断言失败。
func TestSkillConfigContractConstants(t *testing.T) {
	// contractBytes is the checked-in canonical document from the matching core release.
	// contractBytes 是来自匹配核心版本的已检入规范文档。
	contractBytes, err := os.ReadFile("contracts/skill-config/v1/contract.json")
	if err != nil {
		t.Fatalf("read package-configuration contract: %v", err)
	}
	// contract is the parsed subset used by public Go constants.
	// contract 是公共 Go 常量使用的已解析子集。
	var contract skillConfigContractDocument
	if err := json.Unmarshal(contractBytes, &contract); err != nil {
		t.Fatalf("decode package-configuration contract: %v", err)
	}

	if SkillConfigContractVersion != contract.ContractVersion {
		t.Fatalf("contract version drifted: got %d want %d", SkillConfigContractVersion, contract.ContractVersion)
	}
	if !reflect.DeepEqual(skillConfigTypeStrings(), contract.Declaration.Types) {
		t.Fatalf("configuration types drifted: got %#v want %#v", skillConfigTypeStrings(), contract.Declaration.Types)
	}
	if !reflect.DeepEqual(SkillPackageConfigFormats, contract.Declaration.Formats) {
		t.Fatalf("configuration formats drifted: got %#v want %#v", SkillPackageConfigFormats, contract.Declaration.Formats)
	}
	if !reflect.DeepEqual(skillConfigStateStrings(), contract.Declaration.States) {
		t.Fatalf("configuration states drifted: got %#v want %#v", skillConfigStateStrings(), contract.Declaration.States)
	}
	if !reflect.DeepEqual(skillConfigDescribeModeStrings(), contract.Declaration.DescribeModes) {
		t.Fatalf("configuration describe modes drifted: got %#v want %#v", skillConfigDescribeModeStrings(), contract.Declaration.DescribeModes)
	}
	if !reflect.DeepEqual(skillConfigStoreScopeStrings(), contract.Declaration.StoreScopes) {
		t.Fatalf("configuration store scopes drifted: got %#v want %#v", skillConfigStoreScopeStrings(), contract.Declaration.StoreScopes)
	}
	if !reflect.DeepEqual(skillConfigErrorStrings(), contract.Errors) {
		t.Fatalf("configuration errors drifted: got %#v want %#v", skillConfigErrorStrings(), contract.Errors)
	}
	if SkillConfigMaximumSafeInteger != contract.Limits.MaximumSafeInteger {
		t.Fatalf("safe integer limit drifted: got %d want %d", SkillConfigMaximumSafeInteger, contract.Limits.MaximumSafeInteger)
	}
	if SkillConfigMaximumEventPollLimit != contract.Limits.MaximumEventPollLimit {
		t.Fatalf("event poll limit drifted: got %d want %d", SkillConfigMaximumEventPollLimit, contract.Limits.MaximumEventPollLimit)
	}
}

// skillConfigTypeStrings converts exported typed constants into their wire values.
// skillConfigTypeStrings 将导出的类型化常量转换为线协议值。
//
// The returned slice preserves canonical contract order.
// 返回切片保持规范契约顺序。
func skillConfigTypeStrings() []string {
	// values owns the converted wire values without mutating the exported slice.
	// values 保存转换后的线协议值且不修改导出切片。
	values := make([]string, len(SkillPackageConfigTypes))
	for index, value := range SkillPackageConfigTypes {
		values[index] = string(value)
	}
	return values
}

// skillConfigStateStrings converts exported typed constants into their wire values.
// skillConfigStateStrings 将导出的类型化常量转换为线协议值。
//
// The returned slice preserves canonical contract order.
// 返回切片保持规范契约顺序。
func skillConfigStateStrings() []string {
	// values owns the converted wire values without mutating the exported slice.
	// values 保存转换后的线协议值且不修改导出切片。
	values := make([]string, len(SkillPackageConfigStates))
	for index, value := range SkillPackageConfigStates {
		values[index] = string(value)
	}
	return values
}

// skillConfigDescribeModeStrings converts exported typed discovery modes into their wire values.
// skillConfigDescribeModeStrings 将导出的类型化发现模式转换为线协议值。
//
// The returned slice preserves canonical contract order.
// 返回切片保持规范契约顺序。
func skillConfigDescribeModeStrings() []string {
	// Values owns the converted wire values without mutating the exported slice.
	// Values 保存转换后的线协议值且不修改导出切片。
	values := make([]string, len(SkillPackageConfigDescribeModes))
	for index, value := range SkillPackageConfigDescribeModes {
		values[index] = string(value)
	}
	return values
}

// skillConfigStoreScopeStrings converts exported typed store scopes into their wire values.
// skillConfigStoreScopeStrings 将导出的类型化存储作用域转换为线协议值。
//
// The returned slice preserves canonical contract order.
// 返回切片保持规范契约顺序。
func skillConfigStoreScopeStrings() []string {
	// Values owns the converted wire values without mutating the exported slice.
	// Values 保存转换后的线协议值且不修改导出切片。
	values := make([]string, len(SkillConfigStoreScopes))
	for index, value := range SkillConfigStoreScopes {
		values[index] = string(value)
	}
	return values
}

// skillConfigErrorStrings converts exported typed error codes into their wire values.
// skillConfigErrorStrings 将导出的类型化错误码转换为线协议值。
//
// The returned slice preserves canonical contract order.
// 返回切片保持规范契约顺序。
func skillConfigErrorStrings() []string {
	// values owns the converted wire values without mutating the exported slice.
	// values 保存转换后的线协议值且不修改导出切片。
	values := make([]string, len(SkillConfigErrorCodes))
	for index, value := range SkillConfigErrorCodes {
		values[index] = string(value)
	}
	return values
}
