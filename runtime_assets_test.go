package luaskills

import "testing"

// TestBuildRuntimeInstallManifestIncludesLuaRuntime verifies the default Lua runtime asset plan.
// TestBuildRuntimeInstallManifestIncludesLuaRuntime 校验默认 Lua runtime 资产规划。
func TestBuildRuntimeInstallManifestIncludesLuaRuntime(t *testing.T) {
	manifest, err := BuildRuntimeInstallManifest(RuntimeInstallOptions{
		RuntimeRoot: "runtime",
		Database:    RuntimeDatabaseNone,
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
	if manifest.Assets[0].AssetName != "lua-runtime-"+manifest.Platform.PlatformKey+".tar.gz" {
		t.Fatalf("unexpected lua runtime asset name: %s", manifest.Assets[0].AssetName)
	}
}

// TestBuildRuntimeInstallManifestSkipsLuaRuntime verifies the explicit skip option.
// TestBuildRuntimeInstallManifestSkipsLuaRuntime 校验显式跳过选项。
func TestBuildRuntimeInstallManifestSkipsLuaRuntime(t *testing.T) {
	manifest, err := BuildRuntimeInstallManifest(RuntimeInstallOptions{
		RuntimeRoot:    "runtime",
		Database:       RuntimeDatabaseNone,
		SkipLuaRuntime: true,
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
