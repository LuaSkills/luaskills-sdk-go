package luaskills

import (
	"strings"
	"testing"
)

// TestDecodeJSONEnvelopeTextAnnotatesMalformedEnvelope verifies malformed native envelopes keep FFI context.
// TestDecodeJSONEnvelopeTextAnnotatesMalformedEnvelope 校验畸形原生包络会保留 FFI 上下文。
func TestDecodeJSONEnvelopeTextAnnotatesMalformedEnvelope(t *testing.T) {
	var result map[string]any

	err := decodeJSONEnvelopeText("luaskills_ffi_demo_json", "", &result)
	if err == nil || !strings.Contains(err.Error(), "luaskills_ffi_demo_json: empty JSON FFI response envelope") {
		t.Fatalf("unexpected empty envelope error: %v", err)
	}

	err = decodeJSONEnvelopeText("luaskills_ffi_demo_json", "[1]", &result)
	if err == nil || !strings.Contains(err.Error(), "luaskills_ffi_demo_json: JSON FFI response envelope must be one object") {
		t.Fatalf("unexpected non-object envelope error: %v", err)
	}

	err = decodeJSONEnvelopeText("luaskills_ffi_demo_json", "{", &result)
	if err == nil || !strings.Contains(err.Error(), "luaskills_ffi_demo_json: invalid JSON FFI response envelope") {
		t.Fatalf("unexpected invalid envelope error: %v", err)
	}
}

// TestDecodeJSONEnvelopeTextAnnotatesMalformedResult verifies malformed result payloads keep FFI context.
// TestDecodeJSONEnvelopeTextAnnotatesMalformedResult 校验畸形结果载荷会保留 FFI 上下文。
func TestDecodeJSONEnvelopeTextAnnotatesMalformedResult(t *testing.T) {
	var result map[string]any

	err := decodeJSONEnvelopeText("luaskills_ffi_demo_json", `{"ok":true}`, &result)
	if err == nil || !strings.Contains(err.Error(), "luaskills_ffi_demo_json: missing JSON FFI result payload") {
		t.Fatalf("unexpected missing result error: %v", err)
	}

	err = decodeJSONEnvelopeText("luaskills_ffi_demo_json", `{"ok":true,"result":[]}`, &result)
	if err == nil || !strings.Contains(err.Error(), "luaskills_ffi_demo_json: invalid JSON FFI result payload") {
		t.Fatalf("unexpected invalid result error: %v", err)
	}
}
