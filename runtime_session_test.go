package luaskills

import (
	"encoding/json"
	"testing"
)

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

func TestRuntimeLeaseFunctionNameUsesDedicatedSystemEntrypoints(t *testing.T) {
	client := &RuntimeLeaseClient{bindAuthority: true}
	name, err := client.runtimeLeaseFunctionName("status")
	if err != nil {
		t.Fatalf("resolve system runtime-lease function name: %v", err)
	}
	if name != "luaskills_ffi_system_runtime_lease_status_json" {
		t.Fatalf("unexpected system runtime-lease function name: %s", name)
	}

	publicClient := &RuntimeLeaseClient{bindAuthority: false}
	publicName, err := publicClient.runtimeLeaseFunctionName("status")
	if err != nil {
		t.Fatalf("resolve public runtime-lease function name: %v", err)
	}
	if publicName != "luaskills_ffi_runtime_lease_status_json" {
		t.Fatalf("unexpected public runtime-lease function name: %s", publicName)
	}
}
