package luaskills

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testLuaSkillsTag is an exact core tag of the SDK series so manifest tests never discover releases over the network.
// testLuaSkillsTag 是 SDK 系列内的精确核心标签，使清单测试绝不通过网络发现发布。
const testLuaSkillsTag = "v" + DefaultLuaSkillsSeries + ".0"

// fakeGitHubReleases serves one fixed release listing in place of the GitHub API and counts listing requests.
// fakeGitHubReleases 以固定发布清单代替 GitHub API，并统计清单请求次数。
func fakeGitHubReleases(t *testing.T, repository string, releases []map[string]any) *int {
	t.Helper()
	requests := new(int)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/repos/"+repository+"/releases" {
			t.Errorf("unexpected release discovery request: %s", request.URL.Path)
			http.NotFound(writer, request)
			return
		}
		*requests++
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(releases); err != nil {
			t.Error(err)
		}
	}))
	previous := githubAPIBaseURL
	githubAPIBaseURL = server.URL
	t.Cleanup(func() {
		githubAPIBaseURL = previous
		server.Close()
	})
	return requests
}

// fakeRelease builds one GitHub release listing entry.
// fakeRelease 构造一个 GitHub 发布清单条目。
func fakeRelease(tag string, draft bool, prerelease bool) map[string]any {
	return map[string]any{"tag_name": tag, "draft": draft, "prerelease": prerelease}
}

// TestLuaSkillsSeriesMatchesEmbeddedContract keeps the installed core series identical to the compatibility series.
// TestLuaSkillsSeriesMatchesEmbeddedContract 保持安装的核心系列与兼容检查系列一致。
func TestLuaSkillsSeriesMatchesEmbeddedContract(t *testing.T) {
	if DefaultLuaSkillsSeries != EmbeddedCoreSeries {
		t.Fatalf("installer series %s differs from embedded contract series %s", DefaultLuaSkillsSeries, EmbeddedCoreSeries)
	}
}

// TestLuaSkillsSeriesResolvesNewestPatch installs the newest published patch of the series from a fake listing.
// TestLuaSkillsSeriesResolvesNewestPatch 从伪造清单中安装该系列最新已发布补丁版本。
func TestLuaSkillsSeriesResolvesNewestPatch(t *testing.T) {
	series := "v" + DefaultLuaSkillsSeries
	requests := fakeGitHubReleases(t, DefaultLuaSkillsRepository, []map[string]any{
		fakeRelease(series+".0", false, false),
		fakeRelease(series+".10", false, false),
		fakeRelease(series+".3", false, false),
		fakeRelease(series+".11", true, false),
		fakeRelease(series+".12", false, true),
		fakeRelease("v"+embeddedNextMinorVersion(), false, false),
		fakeRelease("nightly", false, false),
	})
	manifest, err := BuildRuntimeInstallManifest(RuntimeInstallOptions{RuntimeRoot: t.TempDir(), LuaRuntimeVersion: "v0.1.6"})
	if err != nil {
		t.Fatal(err)
	}
	if *requests != 1 {
		t.Fatalf("expected one release listing request, got %d", *requests)
	}
	for _, asset := range manifest.Assets {
		if asset.Role != RuntimeAssetLuaSkillsFFI {
			continue
		}
		if asset.Version != series+".10" || !strings.Contains(asset.DownloadURL, "/releases/download/"+series+".10/") {
			t.Fatalf("series default did not resolve the newest published patch: %#v", asset)
		}
		return
	}
	t.Fatal("LuaSkills FFI asset is missing from the manifest")
}

// TestLuaSkillsSeriesWithoutPublishedPatchFails rejects a series with no published patch instead of choosing another series.
// TestLuaSkillsSeriesWithoutPublishedPatchFails 在系列没有已发布补丁时拒绝，而不选择其他系列。
func TestLuaSkillsSeriesWithoutPublishedPatchFails(t *testing.T) {
	series := "v" + DefaultLuaSkillsSeries
	fakeGitHubReleases(t, DefaultLuaSkillsRepository, []map[string]any{
		fakeRelease(series+".1", true, false),
		fakeRelease(series+".2", false, true),
		fakeRelease("v"+embeddedNextMinorVersion(), false, false),
	})
	_, err := BuildRuntimeInstallManifest(RuntimeInstallOptions{RuntimeRoot: t.TempDir(), LuaRuntimeVersion: "v0.1.6"})
	if err == nil || !strings.Contains(err.Error(), DefaultLuaSkillsSeries) {
		t.Fatalf("series without a published patch was not rejected: %v", err)
	}
}

// TestLuaSkillsExplicitTag accepts an exact tag of the series without discovery and rejects every tag outside it.
// TestLuaSkillsExplicitTag 无需发现即接受系列内的精确标签，并拒绝系列外的所有标签。
func TestLuaSkillsExplicitTag(t *testing.T) {
	requests := fakeGitHubReleases(t, DefaultLuaSkillsRepository, nil)
	explicit := "v" + DefaultLuaSkillsSeries + ".4"
	manifest, err := BuildRuntimeInstallManifest(RuntimeInstallOptions{RuntimeRoot: t.TempDir(), LuaRuntimeVersion: "v0.1.6", LuaSkillsVersion: explicit})
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Assets[len(manifest.Assets)-1].Version != explicit || *requests != 0 {
		t.Fatalf("explicit tag was not used exactly: %#v, requests=%d", manifest.Assets, *requests)
	}
	for _, tag := range []string{"v" + embeddedNextMinorVersion(), "v0.0.1", DefaultLuaSkillsSeries + ".1", "v" + DefaultLuaSkillsSeries,
		"v" + DefaultLuaSkillsSeries + ".01", "v" + DefaultLuaSkillsSeries + ".1-rc.1", "latest"} {
		if _, err := BuildRuntimeInstallManifest(RuntimeInstallOptions{RuntimeRoot: t.TempDir(), LuaRuntimeVersion: "v0.1.6", LuaSkillsVersion: tag}); err == nil {
			t.Errorf("explicit tag outside series %s was accepted: %s", DefaultLuaSkillsSeries, tag)
		}
	}
	if *requests != 0 {
		t.Fatalf("explicit tags triggered release discovery: %d", *requests)
	}
}

// writeFFIArchive writes one gzip tar archive containing the given members.
// writeFFIArchive 写入包含给定成员的 gzip tar 归档。
func writeFFIArchive(t *testing.T, members map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "luaskills-ffi-sdk-test.tar.gz")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	compressed := gzip.NewWriter(file)
	archive := tar.NewWriter(compressed)
	for name, content := range members {
		if err := archive.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := archive.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestVerifyLuaSkillsFFIArchiveContract accepts only an archive packaging this SDK's exact contract digest.
// TestVerifyLuaSkillsFFIArchiveContract 只接受打包本 SDK 精确契约摘要的归档。
func TestVerifyLuaSkillsFFIArchiveContract(t *testing.T) {
	sidecar := EmbeddedContractSHA256 + "  contract.json\n"
	matching := writeFFIArchive(t, map[string]string{"lib/libluaskills.so": "binary", LuaSkillsFFIContractDigestPath: sidecar})
	if err := VerifyLuaSkillsFFIArchiveContract(matching); err != nil {
		t.Fatal("matching contract digest was rejected:", err)
	}
	other := strings.Repeat("0", 64)
	mismatched := writeFFIArchive(t, map[string]string{"lib/libluaskills.so": "binary", LuaSkillsFFIContractDigestPath: other + "  contract.json\n"})
	if err := VerifyLuaSkillsFFIArchiveContract(mismatched); err == nil || !strings.Contains(err.Error(), other) || !strings.Contains(err.Error(), EmbeddedContractSHA256) {
		t.Fatalf("contract digest mismatch was not rejected: %v", err)
	}
	for name, members := range map[string]map[string]string{
		"missing": {"lib/libluaskills.so": "binary"},
		"empty":   {LuaSkillsFFIContractDigestPath: ""},
		"nested":  {"package/" + LuaSkillsFFIContractDigestPath: sidecar},
	} {
		if err := VerifyLuaSkillsFFIArchiveContract(writeFFIArchive(t, members)); err == nil {
			t.Errorf("%s contract digest was accepted", name)
		}
	}
	if err := VerifyLuaSkillsFFIArchiveContract(filepath.Join(t.TempDir(), "absent.tar.gz")); err == nil {
		t.Error("absent archive was accepted")
	}
}

// TestHostCallbackAssets verifies host-only database ownership and explicit migration errors without network access.
// TestHostCallbackAssets 在不访问网络的情况下验证数据库仅由宿主拥有，以及显式迁移错误。
func TestHostCallbackAssets(t *testing.T) {
	// Pin Lua assets so no release discovery can hide a database dependency.
	// 固定 Lua 资产，避免发布发现掩盖数据库依赖。
	root := t.TempDir()
	manifest, err := BuildRuntimeInstallManifest(RuntimeInstallOptions{RuntimeRoot: root, Database: RuntimeDatabaseHostCallback, LuaRuntimeVersion: "v0.1.6", LuaSkillsVersion: testLuaSkillsTag})
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Assets) != 2 || manifest.Assets[0].Role != RuntimeAssetLuaRuntime || manifest.Assets[1].Role != RuntimeAssetLuaSkillsFFI {
		t.Fatalf("unexpected assets: %#v", manifest.Assets)
	}
	if len(manifest.HostOptionsPatch) != 1 || manifest.HostOptionsPatch["sqlite_callback_mode"] != "json" {
		t.Fatalf("unexpected host patch: %#v", manifest.HostOptionsPatch)
	}
	options, err := DefaultHostOptions(root)
	if err != nil || options["host_mode"] != "console" {
		t.Fatalf("unexpected default mode: %#v, %v", options, err)
	}
	for _, preset := range []RuntimeDatabasePreset{"vldb-controller", "vldb-direct"} {
		if _, err := BuildRuntimeInstallManifest(RuntimeInstallOptions{RuntimeRoot: root, Database: preset, LuaRuntimeVersion: "v0.1.6", LuaSkillsVersion: testLuaSkillsTag}); err == nil {
			t.Fatalf("accepted retired preset: %s", preset)
		}
	}
	for _, key := range []string{"sqlite_library_path", "sqlite_provider_mode", "lancedb_library_path", "lancedb_provider_mode", "lancedb_callback_mode", "space_controller"} {
		if _, err := HostOptionsFromRuntimeManifest(&RuntimeInstallManifest{RuntimeRoot: root, HostOptionsPatch: map[string]any{key: nil}}); err == nil || !strings.Contains(err.Error(), key) {
			t.Fatalf("accepted retired field %s: %v", key, err)
		}
	}
}

// TestBuildRuntimeInstallManifestIncludesLuaRuntime verifies the default Lua runtime asset plan.
// TestBuildRuntimeInstallManifestIncludesLuaRuntime 校验默认 Lua runtime 资产规划。
func TestBuildRuntimeInstallManifestIncludesLuaRuntime(t *testing.T) {
	manifest, err := BuildRuntimeInstallManifest(RuntimeInstallOptions{
		RuntimeRoot:       "runtime",
		Database:          RuntimeDatabaseNone,
		LuaRuntimeVersion: "v0.1.6",
		LuaSkillsVersion:  testLuaSkillsTag,
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
		LuaSkillsVersion:  testLuaSkillsTag,
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
	if value, exists := options["managed_runtime_distribution_root"]; !exists || value != nil {
		t.Fatalf("unexpected managed runtime distribution root: %#v", value)
	}
	if value, exists := options["managed_runtime_environment_root"]; !exists || value != nil {
		t.Fatalf("unexpected managed runtime environment root: %#v", value)
	}
	if value, exists := options["managed_runtime_config"]; !exists || value != DefaultManagedRuntimeConfig() {
		t.Fatalf("unexpected managed runtime config: %#v", value)
	}
	if value, exists := options["preserve_host_writable_dirs"]; !exists || value != false {
		t.Fatalf("writable directories must follow runtime_root by default: %#v", value)
	}
}

// TestCreateEngineOptionsPreservesManagedRuntimeConfig verifies B3-B7 host overrides remain exact.
// TestCreateEngineOptionsPreservesManagedRuntimeConfig 验证 B3-B7 宿主覆盖保持精确值。
func TestCreateEngineOptionsPreservesManagedRuntimeConfig(t *testing.T) {
	// TimeoutMS is a positive engine default retained by the pointer-valued JSON field.
	// TimeoutMS 是由指针型 JSON 字段保留的正数引擎默认超时。
	timeoutMS := uint64(15_000)
	// Config uses nondefault values so an accidental SDK fallback remains observable.
	// Config 使用非默认值，使 SDK 意外回退行为保持可观察。
	config := ManagedRuntimeConfig{
		WorkerPoolMaxSizePerEnvironment:                   8,
		WorkerIdleTTLSecs:                                 90,
		PersistentSessionLimitPerEngine:                   128,
		PersistentSessionDefaultBufferLimitBytesPerStream: 2 * 1024 * 1024,
		InvokeDefaultTimeoutMS:                            &timeoutMS,
	}
	// Options is built through the public SDK merge path used by real clients.
	// Options 通过真实客户端使用的公开 SDK 合并路径构造。
	options, err := CreateEngineOptions(
		t.TempDir(),
		map[string]any{"managed_runtime_config": config},
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected engine options error: %v", err)
	}
	// HostOptions is the complete JSON object submitted to the native engine constructor.
	// HostOptions 是提交到原生引擎构造器的完整 JSON 对象。
	hostOptions := options["host_options"].(map[string]any)
	if value := hostOptions["managed_runtime_config"]; value != config {
		t.Fatalf("unexpected managed runtime config: %#v", value)
	}
}

// TestResolveManagedRuntimeInstallRejectsInvalidInputs verifies validation precedes native lookup.
// TestResolveManagedRuntimeInstallRejectsInvalidInputs 验证输入校验先于原生库查找。
func TestResolveManagedRuntimeInstallRejectsInvalidInputs(t *testing.T) {
	// AbsoluteRoot is a platform-native absolute path used to isolate runtime-name validation.
	// AbsoluteRoot 是用于隔离运行时名称校验的平台原生绝对路径。
	absoluteRoot := t.TempDir()
	// TestCases covers an unknown runtime and a relative distribution authority independently.
	// TestCases 分别覆盖未知运行时与相对发行授权。
	testCases := []struct {
		name    string
		options ManagedRuntimeResolveOptions
		want    string
	}{
		{
			name: "unknown runtime",
			options: ManagedRuntimeResolveOptions{
				DistributionRoot: absoluteRoot,
				Runtime:          ManagedRuntimeKind("ruby"),
				Version:          "3.3.0",
				Platform:         "windows-x64",
			},
			want: "runtime must be either python or node",
		},
		{
			name: "relative distribution root",
			options: ManagedRuntimeResolveOptions{
				DistributionRoot: filepath.Join("relative", "runtimes"),
				Runtime:          ManagedRuntimeKindPython,
				Version:          "3.14.4",
				Platform:         "windows-x64",
			},
			want: "distribution root must be an absolute path",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Error is returned before either the cgo bridge or no-cgo fallback can load a library.
			// Error 会在 cgo 桥或 no-cgo 回退加载库之前返回。
			_, err := ResolveManagedRuntimeInstall(testCase.options)
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
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
	// Only host-owned registration settings survive manifest loading.
	// 清单加载仅保留宿主注册相关设置。
	manifest := &RuntimeInstallManifest{RuntimeRoot: t.TempDir(), HostOptionsPatch: map[string]any{"host_mode": "desktop", "sqlite_callback_mode": "json"}}
	options, err := HostOptionsFromRuntimeManifest(manifest)
	if err != nil || options["sqlite_callback_mode"] != "json" || options["host_mode"] != "desktop" {
		t.Fatalf("unexpected host options: %#v, %v", options, err)
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
