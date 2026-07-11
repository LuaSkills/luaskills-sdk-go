package luaskills

import (
	"encoding/json"
	"errors"
	"testing"
)

// TestSkillOperationProgressEventMarshalsToRustFields verifies progress event JSON field names.
// TestSkillOperationProgressEventMarshalsToRustFields 校验进度事件 JSON 字段名。
func TestSkillOperationProgressEventMarshalsToRustFields(t *testing.T) {
	skillID := "demo.skill"
	rootName := "PROJECT"
	sourceType := SkillInstallSourceOfficialHub
	sourceLocator := "demo.skill"
	bytesDone := uint64(32)
	bytesTotal := uint64(64)
	percent := 50.0
	message := "downloading"

	event := SkillOperationProgressEvent{
		OperationID:   "operation-1",
		Sequence:      7,
		Plane:         SkillOperationProgressPlaneSkills,
		Action:        SkillOperationProgressActionInstall,
		Phase:         "downloading_archive",
		Status:        "progress",
		SkillID:       &skillID,
		RootName:      &rootName,
		SourceType:    &sourceType,
		SourceLocator: &sourceLocator,
		BytesDone:     &bytesDone,
		BytesTotal:    &bytesTotal,
		Percent:       &percent,
		Message:       &message,
	}

	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal skill operation progress event: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal skill operation progress event: %v", err)
	}
	if payload["operation_id"] != "operation-1" {
		t.Fatalf("unexpected operation_id: %#v", payload["operation_id"])
	}
	if payload["plane"] != "Skills" {
		t.Fatalf("unexpected plane: %#v", payload["plane"])
	}
	if payload["action"] != "Install" {
		t.Fatalf("unexpected action: %#v", payload["action"])
	}
	if payload["source_type"] != "official_hub" {
		t.Fatalf("unexpected source_type: %#v", payload["source_type"])
	}
	if payload["bytes_done"] != float64(32) {
		t.Fatalf("unexpected bytes_done: %#v", payload["bytes_done"])
	}
	if payload["percent"] != 50.0 {
		t.Fatalf("unexpected percent: %#v", payload["percent"])
	}
}

// TestSkillOperationProgressCallbacksRequireHostBridge verifies Go keeps the host-owned bridge boundary.
// TestSkillOperationProgressCallbacksRequireHostBridge 校验 Go 保持宿主拥有的桥接边界。
func TestSkillOperationProgressCallbacksRequireHostBridge(t *testing.T) {
	err := SetSkillOperationProgressJSONCallback(func(event SkillOperationProgressEvent) error {
		return nil
	})
	if !errors.Is(err, ErrSkillOperationProgressCallbacksRequireHostBridge) {
		t.Fatalf("unexpected progress callback error: %v", err)
	}
	err = ClearSkillOperationProgressJSONCallback()
	if !errors.Is(err, ErrSkillOperationProgressCallbacksRequireHostBridge) {
		t.Fatalf("unexpected progress callback clear error: %v", err)
	}
}

// TestClearJSONCallbacksReportsAllHostBridgeRequirements verifies unified cleanup preserves each boundary error.
// TestClearJSONCallbacksReportsAllHostBridgeRequirements 校验统一清理会保留每类桥接边界错误。
func TestClearJSONCallbacksReportsAllHostBridgeRequirements(t *testing.T) {
	err := ClearJSONCallbacks()
	if !errors.Is(err, ErrProviderCallbacksRequireHostBridge) {
		t.Fatalf("expected provider bridge error, got: %v", err)
	}
	if !errors.Is(err, ErrHostToolCallbacksRequireHostBridge) {
		t.Fatalf("expected host-tool bridge error, got: %v", err)
	}
	if !errors.Is(err, ErrSkillOperationProgressCallbacksRequireHostBridge) {
		t.Fatalf("expected progress bridge error, got: %v", err)
	}
	if !errors.Is(err, ErrModelCallbacksRequireHostBridge) {
		t.Fatalf("expected model bridge error, got: %v", err)
	}
}
