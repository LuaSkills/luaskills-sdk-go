package luaskills

import (
	"encoding/json"
	"errors"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// embeddedDescriptionFixture returns fresh synthetic metadata for isolated ownership tests, never native evidence.
// embeddedDescriptionFixture 为隔离所有权测试返回新合成元数据，绝不作为原生证据。
func embeddedDescriptionFixture() EmbeddedOutputCoreDescription {
	// digest is intentionally synthetic; compatibility validates shape, not binary authentication.
	// digest 故意使用合成值；兼容检查校验形状，不认证二进制。
	digest := strings.Repeat("a", 64)
	return EmbeddedOutputCoreDescription{
		DescriptionVersion: EmbeddedDescriptionVersion, ProtocolVersion: EmbeddedProtocolVersion,
		AbiStructureVersion: EmbeddedProtocolVersion, CoreVersion: EmbeddedCoreVersion,
		Commands: EmbeddedRootCommands(), RuntimeCommands: EmbeddedRuntimeCommands(),
		Capabilities:      EmbeddedRequiredCapabilities(),
		ExecutionBackends: []EmbeddedOutputExecutionBackend{EmbeddedOutputExecutionBackendInProcess},
		Build: EmbeddedOutputEmbeddedBuildIdentity{
			CargoFeatures: []string{}, ContractSha256: EmbeddedContractSHA256, InputsSha256: digest,
			SourceSha256: digest, PackageLockSha256: digest, RustflagsSha256: digest,
			DebugInfo: "true", OptLevel: "0", PointerWidth: strconv.Itoa(strconv.IntSize),
			Rustc: "synthetic test compiler", Target: "synthetic-test-target", TargetArch: "synthetic",
			TargetOs: map[string]string{"windows": "windows", "linux": "linux", "darwin": "macos"}[runtime.GOOS],
		},
	}
}

// compatibilityNative counts constructor calls and supplies controlled bootstrap evidence.
// compatibilityNative 统计构造调用并提供受控引导证据。
type compatibilityNative struct {
	uncertainDriverNative
	// encoded and failure select exact descriptor delivery or failure.
	// encoded 及 failure 选择精确描述交付或失败。
	encoded []byte
	failure error
	// created proves whether a rejected bootstrap crossed the native allocation boundary.
	// created 证明被拒绝的引导是否跨越原生分配边界。
	created int
}

// describe returns controlled bytes or their exact original error without native ownership.
// describe 返回受控字节或其精确原始错误，不创建原生所有权。
func (n *compatibilityNative) describe() ([]byte, error) { return n.encoded, n.failure }

// create records a synthetic allocation attempt and returns a fixture identity.
// create 记录合成分配尝试并返回夹具身份。
func (n *compatibilityNative) create(EmbeddedTransportConfig) (uint64, EmbeddedNativeStatus) {
	n.created++
	return 1, EmbeddedNativeOk
}

// free releases only the synthetic owner used by successful snapshot tests.
// free 仅释放成功快照测试所使用的合成所有者。
func (*compatibilityNative) free(uint64) EmbeddedNativeStatus { return EmbeddedNativeOk }

// requireCompatibilityRejection verifies typed rejection before allocation or owner publication.
// requireCompatibilityRejection 验证在分配及所有者发布前发生类型化拒绝。
func requireCompatibilityRejection(t *testing.T, encoded []byte) {
	t.Helper()
	native := &compatibilityNative{encoded: encoded}
	transport, err := createEmbeddedTransport(EmbeddedTransportConfig{1, 2, 4096, 2048, 2048}, native)
	var failure *EmbeddedCompatibilityError
	if transport != nil || !errors.As(err, &failure) || native.created != 0 {
		t.Fatalf("invalid metadata allocated ownership: owner=%v created=%d err=%v", transport, native.created, err)
	}
	embeddedTransports.Range(func(key, _ any) bool {
		if key.(*EmbeddedTransport).native == native {
			t.Error("rejected native backend was published")
		}
		return true
	})
}

// TestEmbeddedCompatibilityRejectsMetadata verifies every required field plus semantic mismatches before allocation.
// TestEmbeddedCompatibilityRejectsMetadata 在分配前验证每个必需字段及语义不匹配。
func TestEmbeddedCompatibilityRejectsMetadata(t *testing.T) {
	encoded, err := json.Marshal(embeddedDescriptionFixture())
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	for name := range fields {
		t.Run("missing_"+name, func(t *testing.T) {
			copy := make(map[string]json.RawMessage, len(fields))
			for field, value := range fields {
				if field != name {
					copy[field] = value
				}
			}
			bytes, err := json.Marshal(copy)
			if err != nil {
				t.Fatal(err)
			}
			requireCompatibilityRejection(t, bytes)
		})
	}
	var build map[string]json.RawMessage
	if err := json.Unmarshal(fields["build"], &build); err != nil {
		t.Fatal(err)
	}
	for name := range build {
		t.Run("missing_build_"+name, func(t *testing.T) {
			copy := make(map[string]json.RawMessage, len(build))
			for field, value := range build {
				if field != name {
					copy[field] = value
				}
			}
			bytes, err := json.Marshal(copy)
			if err != nil {
				t.Fatal(err)
			}
			root := make(map[string]json.RawMessage, len(fields))
			for field, value := range fields {
				root[field] = value
			}
			root["build"] = bytes
			bytes, err = json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			requireCompatibilityRejection(t, bytes)
		})
	}
	for name, change := range map[string]func(*EmbeddedOutputCoreDescription){
		"core_version":             func(d *EmbeddedOutputCoreDescription) { d.CoreVersion = "mismatch" },
		"description_version":      func(d *EmbeddedOutputCoreDescription) { d.DescriptionVersion++ },
		"protocol_version":         func(d *EmbeddedOutputCoreDescription) { d.ProtocolVersion++ },
		"abi_structure_version":    func(d *EmbeddedOutputCoreDescription) { d.AbiStructureVersion++ },
		"contract_digest":          func(d *EmbeddedOutputCoreDescription) { d.Build.ContractSha256 = strings.Repeat("b", 64) },
		"build_digest":             func(d *EmbeddedOutputCoreDescription) { d.Build.SourceSha256 = strings.Repeat("A", 64) },
		"empty_compiler":           func(d *EmbeddedOutputCoreDescription) { d.Build.Rustc = "" },
		"operating_system":         func(d *EmbeddedOutputCoreDescription) { d.Build.TargetOs = "mismatch" },
		"pointer_width":            func(d *EmbeddedOutputCoreDescription) { d.Build.PointerWidth = "0" },
		"missing_commands":         func(d *EmbeddedOutputCoreDescription) { d.Commands = []string{} },
		"missing_runtime_commands": func(d *EmbeddedOutputCoreDescription) { d.RuntimeCommands = []string{} },
		"missing_capabilities":     func(d *EmbeddedOutputCoreDescription) { d.Capabilities = []string{} },
		"duplicate_capabilities":   func(d *EmbeddedOutputCoreDescription) { d.Capabilities = append(d.Capabilities, d.Capabilities[0]) },
		"duplicate_commands":       func(d *EmbeddedOutputCoreDescription) { d.Commands = append(d.Commands, d.Commands[0]) },
		"duplicate_runtime_commands": func(d *EmbeddedOutputCoreDescription) {
			d.RuntimeCommands = append(d.RuntimeCommands, d.RuntimeCommands[0])
		},
		"empty_feature":      func(d *EmbeddedOutputCoreDescription) { d.Build.CargoFeatures = []string{""} },
		"duplicate_features": func(d *EmbeddedOutputCoreDescription) { d.Build.CargoFeatures = []string{"LUA", "LUA"} },
		"missing_backend":    func(d *EmbeddedOutputCoreDescription) { d.ExecutionBackends = []EmbeddedOutputExecutionBackend{} },
		"unknown_backend":    func(d *EmbeddedOutputCoreDescription) { d.ExecutionBackends = append(d.ExecutionBackends, "unknown") },
		"duplicate_backends": func(d *EmbeddedOutputCoreDescription) {
			d.ExecutionBackends = append(d.ExecutionBackends, d.ExecutionBackends[0])
		},
	} {
		t.Run(name, func(t *testing.T) {
			description := embeddedDescriptionFixture()
			change(&description)
			bytes, err := json.Marshal(description)
			if err != nil {
				t.Fatal(err)
			}
			requireCompatibilityRejection(t, bytes)
		})
	}
}

// TestEmbeddedCompatibilityRejectsWire tests malformed bytes and explicit float version tokens without allocation.
// TestEmbeddedCompatibilityRejectsWire 测试畸形字节及显式浮点版本记号，不进行分配。
func TestEmbeddedCompatibilityRejectsWire(t *testing.T) {
	encoded, err := json.Marshal(embeddedDescriptionFixture())
	if err != nil {
		t.Fatal(err)
	}
	for _, bytes := range [][]byte{nil, []byte("{"), {0xff}, []byte("null"), []byte("[]"),
		[]byte(strings.Repeat(" ", int(EmbeddedDescriptionMaxBytes)+1)),
		[]byte(strings.Replace(string(encoded), `"protocol_version":1`, `"protocol_version":1.0`, 1)),
		[]byte(strings.Replace(string(encoded), `"protocol_version":1`, `"protocol_version":1,"protocol_version":1`, 1)),
	} {
		requireCompatibilityRejection(t, bytes)
	}
	_, err = validateEmbeddedDescription([]byte("{"))
	var failure *EmbeddedCompatibilityError
	if !errors.As(err, &failure) || errors.Unwrap(failure) == nil || !strings.Contains(failure.Error(), failure.Cause.Error()) {
		t.Fatal("strict decoding cause was not retained", err)
	}
}

// TestEmbeddedCompatibilityNativeFailure preserves exact bootstrap status and skips allocation and buffer release.
// TestEmbeddedCompatibilityNativeFailure 保留精确引导状态，跳过分配及缓冲释放。
func TestEmbeddedCompatibilityNativeFailure(t *testing.T) {
	failure := &EmbeddedTransportError{"luaskills_ffi_embedded_describe_v1", EmbeddedNativeInternal}
	native := &compatibilityNative{failure: failure}
	transport, err := createEmbeddedTransport(EmbeddedTransportConfig{1, 2, 4096, 2048, 2048}, native)
	if transport != nil || !errors.Is(err, failure) || native.created != 0 {
		t.Fatal("bootstrap failure changed", err)
	}
}

// TestEmbeddedCompatibilitySnapshot preserves additive capabilities and isolates every returned slice, even after free.
// TestEmbeddedCompatibilitySnapshot 保留新增能力，并隔离每个返回切片，即使释放后仍如此。
func TestEmbeddedCompatibilitySnapshot(t *testing.T) {
	description := embeddedDescriptionFixture()
	description.Capabilities = append(description.Capabilities, "future_capability_v1")
	description.Build.CargoFeatures = []string{"TEST_FEATURE"}
	encoded, err := json.Marshal(description)
	if err != nil {
		t.Fatal(err)
	}
	native := &compatibilityNative{encoded: encoded}
	transport, err := createEmbeddedTransport(EmbeddedTransportConfig{1, 2, 4096, 2048, 2048}, native)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if transport.TransportID() == 0 {
			return
		}
		if err := transport.Close(); err != nil {
			t.Error(err)
		}
		if err := transport.Free(); err != nil {
			t.Error(err)
		}
	})
	snapshot := transport.CoreDescription()
	snapshot.Commands[0] = "changed"
	snapshot.RuntimeCommands[0] = "changed"
	snapshot.Capabilities[0] = "changed"
	snapshot.ExecutionBackends[0] = "changed"
	snapshot.Build.CargoFeatures[0] = "changed"
	snapshot.Build.Rustc = "changed"
	for index := range encoded {
		encoded[index] = 0
	}
	if !reflect.DeepEqual(transport.CoreDescription(), description) {
		t.Fatal("caller mutated retained compatibility evidence")
	}
	if err := transport.Close(); err != nil {
		t.Fatal(err)
	}
	if err := transport.Free(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(transport.CoreDescription(), description) {
		t.Fatal("release destroyed compatibility evidence")
	}
	// Non-nil empty arrays are part of the strict JSON contract, including snapshot copies.
	// 非 nil 空数组属于严格 JSON 契约，包括快照副本。
	empty := (&EmbeddedTransport{description: embeddedDescriptionFixture()}).CoreDescription()
	if empty.Build.CargoFeatures == nil {
		t.Fatal("empty feature array became null")
	}
}
