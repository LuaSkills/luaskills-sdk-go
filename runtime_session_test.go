package luaskills

import (
	"encoding/json"
	"testing"
)

// TestRuntimeLeaseIdentityMarshalsToFFIKeys verifies runtime lease identity JSON field names.
// TestRuntimeLeaseIdentityMarshalsToFFIKeys 校验运行时租约身份的 JSON 字段名。
func TestRuntimeLeaseIdentityMarshalsToFFIKeys(t *testing.T) {
	payload := RuntimeLeaseIdentity{
		LeaseID:    "lease-001",
		SID:        "sid-001",
		Generation: 7,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal runtime lease identity: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal runtime lease identity: %v", err)
	}

	if decoded["lease_id"] != "lease-001" {
		t.Fatalf("expected lease_id key, got %#v", decoded)
	}
	if decoded["sid"] != "sid-001" {
		t.Fatalf("expected sid key, got %#v", decoded)
	}
	if decoded["generation"] != float64(7) {
		t.Fatalf("expected generation key, got %#v", decoded)
	}
}

// TestRuntimeLeaseFunctionNameUsesDedicatedSystemEntrypoints verifies system-bound lease FFI names.
// TestRuntimeLeaseFunctionNameUsesDedicatedSystemEntrypoints 校验绑定 system 权限的租约 FFI 函数名。
func TestRuntimeLeaseFunctionNameUsesDedicatedSystemEntrypoints(t *testing.T) {
	client := &RuntimeLeaseClient{bindAuthority: true}
	name, err := client.runtimeLeaseFunctionName(RuntimeLeaseStatusAction)
	if err != nil {
		t.Fatalf("resolve system runtime-lease function name: %v", err)
	}
	if name != "luaskills_ffi_system_runtime_lease_status_json" {
		t.Fatalf("unexpected system runtime-lease function name: %s", name)
	}

	publicClient := &RuntimeLeaseClient{bindAuthority: false}
	publicName, err := publicClient.runtimeLeaseFunctionName(RuntimeLeaseStatusAction)
	if err != nil {
		t.Fatalf("resolve public runtime-lease function name: %v", err)
	}
	if publicName != "luaskills_ffi_runtime_lease_status_json" {
		t.Fatalf("unexpected public runtime-lease function name: %s", publicName)
	}
}

// TestRuntimeLeaseFunctionNameRejectsUnsupportedAction verifies unknown lease actions are rejected.
// TestRuntimeLeaseFunctionNameRejectsUnsupportedAction 校验未知租约动作会被拒绝。
func TestRuntimeLeaseFunctionNameRejectsUnsupportedAction(t *testing.T) {
	client := &RuntimeLeaseClient{}
	_, err := client.runtimeLeaseFunctionName(RuntimeLeaseAction("destroy"))
	if err == nil {
		t.Fatalf("expected unsupported runtime lease action error")
	}
}

// TestSystemRuntimeLeaseCreateRequiresPackage verifies missing trusted package metadata fails before FFI dispatch.
// TestSystemRuntimeLeaseCreateRequiresPackage 校验缺少可信包元数据时会在 FFI 分发前失败。
func TestSystemRuntimeLeaseCreateRequiresPackage(t *testing.T) {
	client := &RuntimeLeaseClient{bindAuthority: true}
	_, err := client.CreateWithOptions("system-session", false, nil)
	if err == nil || err.Error() != "system runtime lease create requires system_package" {
		t.Fatalf("unexpected system package error: %v", err)
	}
}

// TestSystemRuntimePackageMarshalsExactKeys verifies the Rust request field names.
// TestSystemRuntimePackageMarshalsExactKeys 校验 Rust 请求使用的精确字段名。
func TestSystemRuntimePackageMarshalsExactKeys(t *testing.T) {
	raw, err := json.Marshal(SystemRuntimePackage{ID: "debug", Root: "C:/plugins/debug", DependenciesFile: "dependencies.json"})
	if err != nil {
		t.Fatalf("marshal system package: %v", err)
	}
	if string(raw) != `{"id":"debug","root":"C:/plugins/debug","dependencies_file":"dependencies.json"}` {
		t.Fatalf("unexpected system package JSON: %s", raw)
	}
}
