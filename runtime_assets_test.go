package luaskills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuildRuntimeInstallManifestIncludesLuaRuntime verifies the default Lua runtime asset plan.
// TestBuildRuntimeInstallManifestIncludesLuaRuntime 校验默认 Lua runtime 资产规划。
func TestBuildRuntimeInstallManifestIncludesLuaRuntime(t *testing.T) {
	manifest, err := BuildRuntimeInstallManifest(RuntimeInstallOptions{
		RuntimeRoot:       "runtime",
		Database:          RuntimeDatabaseNone,
		LuaRuntimeVersion: "v0.1.6",
	})
	if err != nil {
		t.Fatalf("BuildRuntimeInstallManifest failed: %v", err)
	}
	if len(manifest.Assets) < 2 {
		t.Fatalf("expected lua runtime and FFI assets, got %d", len(manifest.Assets))
	}
	if manifest.Assets[0].Role != RuntimeAssetLuaRuntime {
		t.Fatalf("expected first asset role %q, got %q", RuntimeAssetLuaRuntime, manifest.Assets[0].Role)
	}
	if manifest.Assets[0].AssetName != "lua-runtime-packages-"+manifest.Platform.PlatformKey+".tar.gz" {
		t.Fatalf("unexpected lua runtime asset name: %s", manifest.Assets[0].AssetName)
	}
}

// TestBuildRuntimeInstallManifestSkipsLuaRuntime verifies the explicit skip option.
// TestBuildRuntimeInstallManifestSkipsLuaRuntime 校验显式跳过选项。
func TestBuildRuntimeInstallManifestSkipsLuaRuntime(t *testing.T) {
	manifest, err := BuildRuntimeInstallManifest(RuntimeInstallOptions{
		RuntimeRoot:       "runtime",
		Database:          RuntimeDatabaseNone,
		LuaRuntimeVersion: "v0.1.6",
		SkipLuaRuntime:    true,
	})
	if err != nil {
		t.Fatalf("BuildRuntimeInstallManifest failed: %v", err)
	}
	for _, asset := range manifest.Assets {
		if asset.Role == RuntimeAssetLuaRuntime {
			t.Fatalf("lua runtime asset should be skipped")
		}
	}
}

// TestDecodeRuntimeInstallManifestReportsPathAwareErrors verifies malformed manifests keep file context.
// TestDecodeRuntimeInstallManifestReportsPathAwareErrors 校验畸形清单会保留文件上下文。
func TestDecodeRuntimeInstallManifestReportsPathAwareErrors(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected string
	}{
		{
			name:     "empty",
			raw:      "",
			expected: "runtime install manifest runtime/resources/manifest.json is empty",
		},
		{
			name:     "invalid json",
			raw:      "{",
			expected: "runtime install manifest runtime/resources/manifest.json is invalid JSON",
		},
		{
			name:     "non object",
			raw:      "[1]",
			expected: "runtime install manifest runtime/resources/manifest.json must be one JSON object",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeRuntimeInstallManifest("runtime/resources/manifest.json", []byte(test.raw))
			if err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("unexpected manifest error: %v", err)
			}
		})
	}
}

// TestDefaultHostOptionsAllowsMissingManifest verifies absent manifests keep the base SDK defaults.
// TestDefaultHostOptionsAllowsMissingManifest 校验缺失清单时仍保留 SDK 基础默认值。
func TestDefaultHostOptionsAllowsMissingManifest(t *testing.T) {
	root := t.TempDir()
	options, err := DefaultHostOptions(root)
	if err != nil {
		t.Fatalf("unexpected default host options error: %v", err)
	}
	if options["runtime_root"] != normalizePath(root) {
		t.Fatalf("unexpected runtime root: %v", options["runtime_root"])
	}
}

// TestCreateEngineOptionsPropagatesManifestErrors verifies malformed manifests fail the engine option path.
// TestCreateEngineOptionsPropagatesManifestErrors 校验畸形清单会让引擎选项构造路径失败。
func TestCreateEngineOptionsPropagatesManifestErrors(t *testing.T) {
	root := t.TempDir()
	manifestPath := RuntimeManifestPath(root)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatalf("create manifest directory: %v", err)
	}
	if err := os.WriteFile(manifestPath, []byte("{"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	_, err := CreateEngineOptions(root, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "runtime install manifest") || !strings.Contains(err.Error(), "invalid JSON") {
		t.Fatalf("unexpected engine option error: %v", err)
	}
}

// TestHostOptionsFromRuntimeManifestSanitizesRootScopedPaths verifies manifest paths stay under runtime root.
// TestHostOptionsFromRuntimeManifestSanitizesRootScopedPaths 校验 manifest 路径保持在 runtime root 下。
func TestHostOptionsFromRuntimeManifestSanitizesRootScopedPaths(t *testing.T) {
	root := t.TempDir()
	manifest := &RuntimeInstallManifest{
		RuntimeRoot: root,
		HostOptionsPatch: map[string]any{
			"sqlite_library_path":  "libs/sqlite.dll",
			"lancedb_library_path": filepath.Join(root, "libs", "lancedb.dll"),
			"space_controller": map[string]any{
				"executable_path": "bin/vldb-controller.exe",
			},
		},
	}

	options, err := HostOptionsFromRuntimeManifest(manifest)
	if err != nil {
		t.Fatalf("unexpected manifest host options error: %v", err)
	}
	if options["sqlite_library_path"] != normalizePath(filepath.Join(root, "libs", "sqlite.dll")) {
		t.Fatalf("unexpected sqlite path: %#v", options["sqlite_library_path"])
	}
	spaceController := options["space_controller"].(map[string]any)
	if spaceController["executable_path"] != normalizePath(filepath.Join(root, "bin", "vldb-controller.exe")) {
		t.Fatalf("unexpected controller path: %#v", spaceController["executable_path"])
	}
}

// TestHostOptionsFromRuntimeManifestRejectsEscapingPaths verifies manifest host paths cannot escape runtime root.
// TestHostOptionsFromRuntimeManifestRejectsEscapingPaths 校验 manifest 宿主路径不能逃逸 runtime root。
func TestHostOptionsFromRuntimeManifestRejectsEscapingPaths(t *testing.T) {
	root := t.TempDir()
	manifest := &RuntimeInstallManifest{
		RuntimeRoot: root,
		HostOptionsPatch: map[string]any{
			"sqlite_library_path": "../outside.dll",
		},
	}

	_, err := HostOptionsFromRuntimeManifest(manifest)
	if err == nil || !strings.Contains(err.Error(), "sqlite_library_path") {
		t.Fatalf("unexpected manifest host options error: %v", err)
	}
}
