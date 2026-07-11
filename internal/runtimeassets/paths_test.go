package runtimeassets

import (
	"path/filepath"
	"testing"
)

// TestHasUnsafeRelativeSegment verifies ambiguous manifest segments are rejected.
// TestHasUnsafeRelativeSegment 校验不明确的清单路径片段会被拒绝。
func TestHasUnsafeRelativeSegment(t *testing.T) {
	for _, value := range []string{"../libs/a", "libs/../a", "libs//a", "./libs/a"} {
		if !HasUnsafeRelativeSegment(value) {
			t.Fatalf("expected unsafe path: %s", value)
		}
	}
	if HasUnsafeRelativeSegment("libs/luaskills.dll") {
		t.Fatal("expected safe relative path")
	}
}

// TestIsStrictlyInside verifies root equality and escapes are rejected.
// TestIsStrictlyInside 校验根目录自身与逃逸路径会被拒绝。
func TestIsStrictlyInside(t *testing.T) {
	root := filepath.Join(t.TempDir(), "runtime")
	if IsStrictlyInside(root, root) {
		t.Fatal("root itself must not be accepted")
	}
	if !IsStrictlyInside(root, filepath.Join(root, "libs", "luaskills.dll")) {
		t.Fatal("expected contained path")
	}
	if IsStrictlyInside(root, filepath.Join(root, "..", "outside")) {
		t.Fatal("expected escaped path rejection")
	}
}
