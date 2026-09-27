package luaskills

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

// EmbeddedCompatibilityError rejects native metadata before transport allocation; it never authorizes fallback.
// EmbeddedCompatibilityError 在传输分配前拒绝原生元数据；绝不授权降级。
type EmbeddedCompatibilityError struct {
	// Message identifies the malformed or incompatible field without invoking native work.
	// Message 标识格式错误或不兼容字段，不调用原生工作。
	Message string
	// Cause retains a strict decoding error when one exists.
	// Cause 在存在严格解码错误时保留其原因。
	Cause error
}

// Error returns a readable compatibility diagnostic and optional original cause.
// Error 返回可读兼容诊断及可选原始原因。
func (e *EmbeddedCompatibilityError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap exposes the original parse error to errors.Is and errors.As without losing its identity.
// Unwrap 向 errors.Is 及 errors.As 暴露原始解析错误，不丢失其身份。
func (e *EmbeddedCompatibilityError) Unwrap() error { return e.Cause }

// embeddedBuildDigest validates the core's lowercase SHA-256 input evidence, not binary authenticity.
// embeddedBuildDigest 校验核心小写 SHA-256 输入证据，不认证二进制。
var embeddedBuildDigest = regexp.MustCompile(`^[0-9a-f]{64}$`)

// validateEmbeddedNames requires unique nonempty names and the generated required subset; empty optional lists are valid.
// validateEmbeddedNames 要求唯一非空名称及生成必需子集；可选清单允许为空。
func validateEmbeddedNames(values, required []string, field string) error {
	seen := make(map[string]bool, len(values))
	for _, name := range values {
		if name == "" || seen[name] {
			return &EmbeddedCompatibilityError{Message: "invalid native name inventory: " + field}
		}
		seen[name] = true
	}
	for _, name := range required {
		if !seen[name] {
			return &EmbeddedCompatibilityError{Message: "missing native requirement in " + field + ": " + name}
		}
	}
	return nil
}

// validateEmbeddedDescription decodes bounded owned bytes and returns exact compatible metadata or a typed rejection.
// validateEmbeddedDescription 解码有界拥有型字节，返回精确兼容元数据或类型化拒绝。
// The generated decoder owns structural validation; selected input hashes do not authenticate the loaded binary.
// 生成解码器负责结构校验；选定输入摘要不认证已加载二进制。
func validateEmbeddedDescription(bytes []byte) (EmbeddedOutputCoreDescription, error) {
	var empty EmbeddedOutputCoreDescription
	if len(bytes) == 0 || uint64(len(bytes)) > EmbeddedDescriptionMaxBytes {
		return empty, &EmbeddedCompatibilityError{Message: "invalid native core description size"}
	}
	description, err := DecodeEmbeddedOutputCoreDescription(bytes)
	if err != nil {
		return empty, &EmbeddedCompatibilityError{Message: "invalid native core description", Cause: err}
	}
	for _, version := range []struct {
		name             string
		actual, expected uint32
	}{{"description_version", description.DescriptionVersion, EmbeddedDescriptionVersion},
		{"protocol_version", description.ProtocolVersion, EmbeddedProtocolVersion},
		{"abi_structure_version", description.AbiStructureVersion, EmbeddedProtocolVersion}} {
		if version.actual != version.expected {
			return empty, &EmbeddedCompatibilityError{Message: fmt.Sprintf("native %s mismatch: expected %d, received %d", version.name, version.expected, version.actual)}
		}
	}
	if description.CoreVersion != EmbeddedCoreVersion || description.Build.ContractSha256 != EmbeddedContractSHA256 {
		return empty, &EmbeddedCompatibilityError{Message: "native core version or embedded contract SHA-256 mismatch"}
	}
	for _, inventory := range []struct {
		name             string
		actual, required []string
	}{{"commands", description.Commands, EmbeddedRootCommands()},
		{"runtime_commands", description.RuntimeCommands, EmbeddedRuntimeCommands()},
		{"capabilities", description.Capabilities, EmbeddedRequiredCapabilities()},
		{"cargo_features", description.Build.CargoFeatures, nil}} {
		if err := validateEmbeddedNames(inventory.actual, inventory.required, inventory.name); err != nil {
			return empty, err
		}
	}
	backends := make([]string, len(description.ExecutionBackends))
	for index, backend := range description.ExecutionBackends {
		backends[index] = string(backend)
	}
	if err := validateEmbeddedNames(backends, []string{"in_process"}, "execution_backends"); err != nil {
		return empty, err
	}
	// Field names and string coverage come from the generated build type, not a second handwritten schema.
	// 字段名称及字符串覆盖来自生成构建类型，不使用第二份手写 Schema。
	build := reflect.ValueOf(description.Build)
	for index := 0; index < build.NumField(); index++ {
		value := build.Field(index)
		if value.Kind() != reflect.String {
			continue
		}
		name := build.Type().Field(index).Tag.Get("json")
		if value.String() == "" || (strings.HasSuffix(name, "_sha256") && !embeddedBuildDigest.MatchString(value.String())) {
			return empty, &EmbeddedCompatibilityError{Message: "invalid native build identity: " + name}
		}
	}
	// The linker/loader enforces instruction-set compatibility; compare OS and the Go process's actual pointer width.
	// 链接器／加载器负责指令集兼容；比较操作系统及 Go 进程实际指针位宽。
	operatingSystem, supported := map[string]string{"windows": "windows", "linux": "linux", "darwin": "macos"}[runtime.GOOS]
	if !supported || description.Build.TargetOs != operatingSystem || description.Build.PointerWidth != strconv.Itoa(strconv.IntSize) {
		return empty, &EmbeddedCompatibilityError{Message: "native target operating system or pointer width mismatch"}
	}
	return description, nil
}

// CoreDescription returns an independent immutable-evidence snapshot, including after successful native release.
// CoreDescription 返回不可变证据的独立快照，包括原生释放成功之后。
// Slice clones preserve empty arrays and prevent callers from mutating retained compatibility evidence.
// 切片副本保留空数组，防止调用方修改所保留的兼容证据。
func (t *EmbeddedTransport) CoreDescription() EmbeddedOutputCoreDescription {
	result := t.description
	result.Commands = slices.Clone(result.Commands)
	result.RuntimeCommands = slices.Clone(result.RuntimeCommands)
	result.Capabilities = slices.Clone(result.Capabilities)
	result.ExecutionBackends = slices.Clone(result.ExecutionBackends)
	result.Build.CargoFeatures = slices.Clone(result.Build.CargoFeatures)
	return result
}
