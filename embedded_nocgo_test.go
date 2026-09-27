//go:build !cgo

package luaskills

import (
	"strings"
	"testing"
)

// TestEmbeddedNoCgo keeps codecs available while explicitly rejecting native allocation without cgo.
// TestEmbeddedNoCgo 保持编码器可用，同时在无 cgo 时明确拒绝原生分配。
func TestEmbeddedNoCgo(t *testing.T) {
	if _, err := EncodeEmbeddedJSON(map[string]any{"ok": true}, 1024); err != nil {
		t.Fatal(err)
	}
	if _, err := NewEmbeddedTransport(EmbeddedTransportConfig{1, 1, 1024, 1024, 1024}); err == nil || !strings.Contains(strings.ToLower(err.Error()), "cgo") {
		t.Fatalf("no-cgo allocation did not fail explicitly: %v", err)
	}
}
