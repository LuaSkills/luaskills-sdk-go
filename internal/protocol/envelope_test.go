package protocol

import (
	"strings"
	"testing"
)

// TestDecodeEnvelopeAnnotatesMalformedEnvelope verifies malformed native envelopes keep FFI context.
// TestDecodeEnvelopeAnnotatesMalformedEnvelope 校验畸形原生包络会保留 FFI 上下文。
func TestDecodeEnvelopeAnnotatesMalformedEnvelope(t *testing.T) {
	var result map[string]any

	err := DecodeEnvelope("luaskills_ffi_demo_json", "", &result)
	if err == nil || !strings.Contains(err.Error(), "luaskills_ffi_demo_json: empty JSON FFI response envelope") {
		t.Fatalf("unexpected empty envelope error: %v", err)
	}

	err = DecodeEnvelope("luaskills_ffi_demo_json", "[1]", &result)
	if err == nil || !strings.Contains(err.Error(), "luaskills_ffi_demo_json: JSON FFI response envelope must be one object") {
		t.Fatalf("unexpected non-object envelope error: %v", err)
	}

	err = DecodeEnvelope("luaskills_ffi_demo_json", "{", &result)
	if err == nil || !strings.Contains(err.Error(), "luaskills_ffi_demo_json: invalid JSON FFI response envelope") {
		t.Fatalf("unexpected invalid envelope error: %v", err)
	}
}

// TestDecodeEnvelopeAnnotatesMalformedResult verifies malformed result payloads keep FFI context.
// TestDecodeEnvelopeAnnotatesMalformedResult 校验畸形结果载荷会保留 FFI 上下文。
func TestDecodeEnvelopeAnnotatesMalformedResult(t *testing.T) {
	var result map[string]any

	err := DecodeEnvelope("luaskills_ffi_demo_json", `{"ok":true}`, &result)
	if err == nil || !strings.Contains(err.Error(), "luaskills_ffi_demo_json: missing JSON FFI result payload") {
		t.Fatalf("unexpected missing result error: %v", err)
	}

	err = DecodeEnvelope("luaskills_ffi_demo_json", `{"ok":true,"result":[]}`, &result)
	if err == nil || !strings.Contains(err.Error(), "luaskills_ffi_demo_json: invalid JSON FFI result payload") {
		t.Fatalf("unexpected invalid result error: %v", err)
	}
}
