package luaskills

import (
	"encoding/json"
	"testing"
)

func TestRuntimeSessionIdentityMarshalsToFFIKeys(t *testing.T) {
	payload := RuntimeSessionIdentity{
		LeaseID:    "lease-001",
		SID:        "sid-001",
		Generation: 7,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal runtime session identity: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal runtime session identity: %v", err)
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

func TestRuntimeSessionFunctionNameUsesDedicatedSystemEntrypoints(t *testing.T) {
	client := &RuntimeSessionClient{bindAuthority: true}
	name, err := client.runtimeSessionFunctionName("status")
	if err != nil {
		t.Fatalf("resolve system runtime-session function name: %v", err)
	}
	if name != "luaskills_ffi_system_runtime_session_status_json" {
		t.Fatalf("unexpected system runtime-session function name: %s", name)
	}

	publicClient := &RuntimeSessionClient{bindAuthority: false}
	publicName, err := publicClient.runtimeSessionFunctionName("status")
	if err != nil {
		t.Fatalf("resolve public runtime-session function name: %v", err)
	}
	if publicName != "luaskills_ffi_runtime_session_status_json" {
		t.Fatalf("unexpected public runtime-session function name: %s", publicName)
	}
}
