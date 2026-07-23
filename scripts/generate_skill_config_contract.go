package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

// contractDocument contains canonical fields used by generated Go constants.
// contractDocument 包含已生成 Go 常量使用的规范字段。
type contractDocument struct {
	// ContractVersion is the package-configuration wire contract version.
	// ContractVersion 是技能包配置线协议版本。
	ContractVersion int `json:"contract_version"`
	// Declaration contains canonical enum-like declaration values.
	// Declaration 包含规范的类枚举声明值。
	Declaration struct {
		// Types contains declared scalar types in canonical order.
		// Types 按规范顺序包含已声明标量类型。
		Types []string `json:"types"`
		// Formats contains host rendering formats in canonical order.
		// Formats 按规范顺序包含宿主渲染格式。
		Formats []string `json:"formats"`
		// States contains effective-value states in canonical order.
		// States 按规范顺序包含有效值状态。
		States []string `json:"states"`
		// DescribeModes contains declaration discovery modes in canonical order.
		// DescribeModes 按规范顺序包含声明发现模式。
		DescribeModes []string `json:"describe_modes"`
		// StoreScopes contains routed configuration stores in canonical order.
		// StoreScopes 按规范顺序包含路由后的配置存储作用域。
		StoreScopes []string `json:"store_scopes"`
	} `json:"declaration"`
	// Limits contains common cross-SDK numeric boundaries.
	// Limits 包含公共跨 SDK 数值边界。
	Limits struct {
		// MaximumSafeInteger is the shared exact integer boundary.
		// MaximumSafeInteger 是共享精确整数边界。
		MaximumSafeInteger int64 `json:"maximum_safe_integer"`
		// MaximumEventPollLimit is the maximum accepted event page.
		// MaximumEventPollLimit 是允许的最大事件分页。
		MaximumEventPollLimit uint64 `json:"maximum_event_poll_limit"`
	} `json:"limits"`
	// Errors contains stable machine-readable error codes in canonical order.
	// Errors 按规范顺序包含稳定的机器可读错误码。
	Errors []string `json:"errors"`
}

// main generates the Go contract module or verifies its exact checked-in output.
// main 生成 Go 契约模块或验证其已检入输出完全一致。
func main() {
	// Check selects read-only drift verification.
	// Check 选择只读漂移验证。
	check := flag.Bool("check", false, "verify generated output without writing")
	flag.Parse()
	// RepositoryRoot is the module root inferred from the current working directory.
	// RepositoryRoot 是从当前工作目录推导出的模块根目录。
	repositoryRoot, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// ContractPath identifies the canonical core-generated input.
	// ContractPath 标识核心生成的规范输入。
	contractPath := filepath.Join(
		repositoryRoot,
		"contracts",
		"skill-config",
		"v1",
		"contract.json",
	)
	// OutputPath identifies the public generated Go module.
	// OutputPath 标识公共已生成 Go 模块。
	outputPath := filepath.Join(repositoryRoot, "config_contract_generated.go")
	// ContractBytes contains the exact canonical JSON input.
	// ContractBytes 包含精确规范 JSON 输入。
	contractBytes, err := os.ReadFile(contractPath)
	if err != nil {
		panic(err)
	}
	// Contract is the parsed generation model.
	// Contract 是已解析生成模型。
	var contract contractDocument
	if err := json.Unmarshal(contractBytes, &contract); err != nil {
		panic(err)
	}
	// Generated is gofmt-normalized deterministic output.
	// Generated 是经 gofmt 规范化的确定性输出。
	generated, err := format.Source([]byte(generateSource(contract)))
	if err != nil {
		panic(err)
	}
	if *check {
		// Existing is compared byte-for-byte with deterministic output.
		// Existing 与确定性输出逐字节比较。
		existing, err := os.ReadFile(outputPath)
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(existing, generated) {
			panic("config_contract_generated.go is stale; run go generate")
		}
		return
	}
	if err := os.WriteFile(outputPath, generated, 0o644); err != nil {
		panic(err)
	}
}

// generateSource renders one complete Go module from the canonical contract.
// generateSource 从规范契约渲染一个完整 Go 模块。
//
// contract is the parsed canonical input and the return value is unformatted Go source.
// contract 是已解析规范输入，返回值是未格式化 Go 源码。
func generateSource(contract contractDocument) string {
	// TypeConstants contains typed constant names in canonical order.
	// TypeConstants 按规范顺序包含类型化常量名称。
	typeConstants := mapValues(contract.Declaration.Types, typeConstant)
	// TypeDeclarations contains generated typed constant assignments.
	// TypeDeclarations 包含已生成的类型化常量赋值。
	typeDeclarations := mapValues(contract.Declaration.Types, func(value string) string {
		// Identifier is the exported Go name derived from the canonical wire value.
		// Identifier 是从规范线协议值派生的导出 Go 名称。
		identifier := typeConstant(value)
		return fmt.Sprintf(
			"// %s identifies package-configuration type %q.\n// %s 标识技能包配置类型 %q。\n%s SkillPackageConfigType = %q",
			identifier,
			value,
			identifier,
			value,
			identifier,
			value,
		)
	})
	// StateConstants contains typed state constant names in canonical order.
	// StateConstants 按规范顺序包含类型化状态常量名称。
	stateConstants := mapValues(contract.Declaration.States, stateConstant)
	// StateDeclarations contains generated typed state constant assignments.
	// StateDeclarations 包含已生成的类型化状态常量赋值。
	stateDeclarations := mapValues(contract.Declaration.States, func(value string) string {
		// Identifier is the exported Go name derived from the canonical wire value.
		// Identifier 是从规范线协议值派生的导出 Go 名称。
		identifier := stateConstant(value)
		return fmt.Sprintf(
			"// %s identifies package-configuration state %q.\n// %s 标识技能包配置状态 %q。\n%s SkillPackageConfigItemState = %q",
			identifier,
			value,
			identifier,
			value,
			identifier,
			value,
		)
	})
	// DescribeModeConstants contains typed discovery-mode constant names in canonical order.
	// DescribeModeConstants 按规范顺序包含类型化发现模式常量名称。
	describeModeConstants := mapValues(contract.Declaration.DescribeModes, describeModeConstant)
	// DescribeModeDeclarations contains generated discovery-mode constant assignments.
	// DescribeModeDeclarations 包含已生成的发现模式常量赋值。
	describeModeDeclarations := mapValues(contract.Declaration.DescribeModes, func(value string) string {
		// Identifier is the exported Go name derived from the canonical wire value.
		// Identifier 是从规范线协议值派生的导出 Go 名称。
		identifier := describeModeConstant(value)
		return fmt.Sprintf(
			"// %s identifies package-configuration discovery mode %q.\n// %s 标识技能包配置发现模式 %q。\n%s SkillPackageConfigDescribeMode = %q",
			identifier,
			value,
			identifier,
			value,
			identifier,
			value,
		)
	})
	// StoreScopeConstants contains typed store-scope constant names in canonical order.
	// StoreScopeConstants 按规范顺序包含类型化存储作用域常量名称。
	storeScopeConstants := mapValues(contract.Declaration.StoreScopes, storeScopeConstant)
	// StoreScopeDeclarations contains generated store-scope constant assignments.
	// StoreScopeDeclarations 包含已生成的存储作用域常量赋值。
	storeScopeDeclarations := mapValues(contract.Declaration.StoreScopes, func(value string) string {
		// Identifier is the exported Go name derived from the canonical wire value.
		// Identifier 是从规范线协议值派生的导出 Go 名称。
		identifier := storeScopeConstant(value)
		return fmt.Sprintf(
			"// %s identifies package-configuration store scope %q.\n// %s 标识技能包配置存储作用域 %q。\n%s SkillConfigStoreScope = %q",
			identifier,
			value,
			identifier,
			value,
			identifier,
			value,
		)
	})
	// ErrorConstants contains typed error constant names in canonical order.
	// ErrorConstants 按规范顺序包含类型化错误常量名称。
	errorConstants := mapValues(contract.Errors, errorConstant)
	// ErrorDeclarations contains generated typed error constant assignments.
	// ErrorDeclarations 包含已生成的类型化错误常量赋值。
	errorDeclarations := mapValues(contract.Errors, func(value string) string {
		// Identifier is the exported Go name derived from the canonical error code.
		// Identifier 是从规范错误码派生的导出 Go 名称。
		identifier := errorConstant(value)
		return fmt.Sprintf(
			"// %s identifies package-configuration error %q.\n// %s 标识技能包配置错误 %q。\n%s SkillConfigErrorCode = %q",
			identifier,
			value,
			identifier,
			value,
			identifier,
			value,
		)
	})
	// FormatValues contains quoted rendering formats in canonical order.
	// FormatValues 按规范顺序包含已引用渲染格式。
	formatValues := mapValues(contract.Declaration.Formats, quote)
	return fmt.Sprintf(`// Code generated by scripts/generate_skill_config_contract.go; DO NOT EDIT.
// 此文件由 scripts/generate_skill_config_contract.go 生成；请勿手工编辑。

%s

package luaskills

// SkillPackageConfigType is the stable wire type of one declared package configuration item.
// SkillPackageConfigType 是单个已声明技能包配置项的稳定线协议类型。
type SkillPackageConfigType string

const (
	%s
)

// SkillPackageConfigItemState identifies the unambiguous runtime state of one declared item.
// SkillPackageConfigItemState 标识单个已声明项的无歧义运行时状态。
type SkillPackageConfigItemState string

const (
	%s
)

// SkillPackageConfigDescribeMode identifies one declaration discovery mode.
// SkillPackageConfigDescribeMode 标识单个声明发现模式。
type SkillPackageConfigDescribeMode string

const (
	%s
)

// SkillConfigStoreScope identifies one routed configuration store.
// SkillConfigStoreScope 标识单个路由后的配置存储。
type SkillConfigStoreScope string

const (
	%s
)

// SkillConfigErrorCode identifies one stable package-configuration failure.
// SkillConfigErrorCode 标识单个稳定技能包配置失败。
type SkillConfigErrorCode string

const (
	%s
)

// SkillConfigContractVersion is the package-configuration wire contract version.
// SkillConfigContractVersion 是技能包配置线协议版本。
const SkillConfigContractVersion = %d

// SkillConfigMaximumSafeInteger is the largest exactly representable integer shared by every supported SDK.
// SkillConfigMaximumSafeInteger 是所有受支持 SDK 共同精确表示的最大整数。
const SkillConfigMaximumSafeInteger int64 = %d

// SkillConfigMaximumEventPollLimit is the largest event page accepted by the core contract.
// SkillConfigMaximumEventPollLimit 是核心契约允许的最大事件分页大小。
const SkillConfigMaximumEventPollLimit uint64 = %d

// SkillPackageConfigTypes contains the stable declared scalar types in contract order.
// SkillPackageConfigTypes 按契约顺序包含稳定的已声明标量类型。
var SkillPackageConfigTypes = []SkillPackageConfigType{
	%s,
}

// SkillPackageConfigFormats contains the stable host rendering formats in contract order.
// SkillPackageConfigFormats 按契约顺序包含稳定的宿主渲染格式。
var SkillPackageConfigFormats = []string{
	%s,
}

// SkillPackageConfigStates contains the stable effective-value states in contract order.
// SkillPackageConfigStates 按契约顺序包含稳定的有效值状态。
var SkillPackageConfigStates = []SkillPackageConfigItemState{
	%s,
}

// SkillPackageConfigDescribeModes contains stable declaration discovery modes in contract order.
// SkillPackageConfigDescribeModes 按契约顺序包含稳定的声明发现模式。
var SkillPackageConfigDescribeModes = []SkillPackageConfigDescribeMode{
	%s,
}

// SkillConfigStoreScopes contains stable routed configuration stores in contract order.
// SkillConfigStoreScopes 按契约顺序包含稳定的路由配置存储作用域。
var SkillConfigStoreScopes = []SkillConfigStoreScope{
	%s,
}

// SkillConfigErrorCodes contains stable package-configuration errors in contract order.
// SkillConfigErrorCodes 按契约顺序包含稳定的技能包配置错误码。
var SkillConfigErrorCodes = []SkillConfigErrorCode{
	%s,
}
`,
		"//go:generate go run ./scripts/generate_skill_config_contract.go",
		strings.Join(typeDeclarations, "\n\t"),
		strings.Join(stateDeclarations, "\n\t"),
		strings.Join(describeModeDeclarations, "\n\t"),
		strings.Join(storeScopeDeclarations, "\n\t"),
		strings.Join(errorDeclarations, "\n\t"),
		contract.ContractVersion,
		contract.Limits.MaximumSafeInteger,
		contract.Limits.MaximumEventPollLimit,
		strings.Join(typeConstants, ",\n\t"),
		strings.Join(formatValues, ",\n\t"),
		strings.Join(stateConstants, ",\n\t"),
		strings.Join(describeModeConstants, ",\n\t"),
		strings.Join(storeScopeConstants, ",\n\t"),
		strings.Join(errorConstants, ",\n\t"),
	)
}

// mapValues maps canonical strings through one deterministic renderer.
// mapValues 通过一个确定性渲染器映射规范字符串。
//
// values preserves canonical order and mapper returns one Go expression per value.
// values 保持规范顺序，mapper 为每个值返回一个 Go 表达式。
func mapValues(values []string, mapper func(string) string) []string {
	// Rendered owns expressions without mutating canonical input.
	// Rendered 保存表达式且不修改规范输入。
	rendered := make([]string, len(values))
	for index, value := range values {
		rendered[index] = mapper(value)
	}
	return rendered
}

// quote renders one Go string literal.
// quote 渲染一个 Go 字符串字面量。
//
// value is canonical UTF-8 text and the return value is a quoted Go expression.
// value 是规范 UTF-8 文本，返回值是已引用 Go 表达式。
func quote(value string) string {
	return fmt.Sprintf("%q", value)
}

// typeConstant maps one canonical scalar type to its public Go constant.
// typeConstant 把一个规范标量类型映射到其公共 Go 常量。
//
// value is one contract type and the return value is its exact Go identifier.
// value 是一个契约类型，返回值是其精确 Go 标识符。
func typeConstant(value string) string {
	switch value {
	case "integer":
		return "SkillPackageConfigTypeInteger"
	case "string":
		return "SkillPackageConfigTypeString"
	case "float":
		return "SkillPackageConfigTypeFloat"
	case "enum":
		return "SkillPackageConfigTypeEnum"
	case "boolean":
		return "SkillPackageConfigTypeBoolean"
	default:
		panic(fmt.Sprintf("unsupported package-configuration type %q", value))
	}
}

// stateConstant maps one canonical item state to its public Go constant.
// stateConstant 把一个规范配置项状态映射到其公共 Go 常量。
//
// value is one contract state and the return value is its exact Go identifier.
// value 是一个契约状态，返回值是其精确 Go 标识符。
func stateConstant(value string) string {
	switch value {
	case "unset":
		return "SkillPackageConfigItemStateUnset"
	case "missing":
		return "SkillPackageConfigItemStateMissing"
	case "default":
		return "SkillPackageConfigItemStateDefault"
	case "configured":
		return "SkillPackageConfigItemStateConfigured"
	case "invalid":
		return "SkillPackageConfigItemStateInvalid"
	default:
		panic(fmt.Sprintf("unsupported package-configuration state %q", value))
	}
}

// describeModeConstant maps one canonical discovery mode to its public Go constant.
// describeModeConstant 把单个规范发现模式映射到公共 Go 常量。
//
// value is one contract discovery mode and the return value is its exact Go identifier.
// value 是单个契约发现模式，返回值是其精确 Go 标识符。
func describeModeConstant(value string) string {
	switch value {
	case "effective":
		return "SkillPackageConfigDescribeModeEffective"
	case "installed":
		return "SkillPackageConfigDescribeModeInstalled"
	default:
		panic(fmt.Sprintf("unsupported package-configuration discovery mode %q", value))
	}
}

// storeScopeConstant maps one canonical store scope to its public Go constant.
// storeScopeConstant 把单个规范存储作用域映射到公共 Go 常量。
//
// value is one contract store scope and the return value is its exact Go identifier.
// value 是单个契约存储作用域，返回值是其精确 Go 标识符。
func storeScopeConstant(value string) string {
	switch value {
	case "skills":
		return "SkillConfigStoreScopeSkills"
	case "system-skills":
		return "SkillConfigStoreScopeSystemSkills"
	default:
		panic(fmt.Sprintf("unsupported package-configuration store scope %q", value))
	}
}

// errorConstant maps one canonical error code to its public Go constant.
// errorConstant 把一个规范错误码映射到其公共 Go 常量。
//
// value is one CONFIG-prefixed contract code and the return value is its exact Go identifier.
// value 是一个以 CONFIG 开头的契约错误码，返回值是其精确 Go 标识符。
func errorConstant(value string) string {
	if !strings.HasPrefix(value, "CONFIG_") {
		panic(fmt.Sprintf("unsupported package-configuration error code %q", value))
	}
	// Segments contains lowercase error-name components after the stable prefix.
	// Segments 包含稳定前缀之后的小写错误名称片段。
	segments := strings.Split(strings.ToLower(strings.TrimPrefix(value, "CONFIG_")), "_")
	for index, segment := range segments {
		if segment == "" {
			panic(fmt.Sprintf("unsupported package-configuration error code %q", value))
		}
		segments[index] = strings.ToUpper(segment[:1]) + segment[1:]
	}
	return "SkillConfigErrorCode" + strings.Join(segments, "")
}
