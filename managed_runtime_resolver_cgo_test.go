//go:build cgo

package luaskills

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestResolveManagedRuntimeInstallCallsPublicJSONFFI verifies the Go wrapper against a real LuaSkills library.
// TestResolveManagedRuntimeInstallCallsPublicJSONFFI 使用真实 LuaSkills 库校验 Go 封装。
func TestResolveManagedRuntimeInstallCallsPublicJSONFFI(t *testing.T) {
	// DistributionRoot is the isolated host-owned managed runtime asset root.
	// DistributionRoot 是隔离的宿主自有受管运行时资产根。
	distributionRoot := filepath.Join(t.TempDir(), "应用 资产", "runtimes")
	// Target supplies the exact normalized LuaSkills platform key for this runner.
	// Target 为当前 runner 提供精确规范化 LuaSkills 平台键。
	target, err := ResolveManagedRuntimePlatformTarget()
	if err != nil {
		t.Fatalf("resolve managed runtime platform: %v", err)
	}
	// ExecutablePath follows the published Node archive layout for this operating system.
	// ExecutablePath 遵循当前操作系统已发布 Node 归档布局。
	executablePath := "bin/node"
	if runtime.GOOS == "windows" {
		executablePath = "node.exe"
	}
	// InstallRoot is the exact directory derived by the Rust resolver contract.
	// InstallRoot 是 Rust 解析器契约派生的精确目录。
	installRoot := filepath.Join(
		distributionRoot,
		"node",
		"node-24.18.0-"+target.PlatformKey,
	)
	if err := os.MkdirAll(filepath.Dir(filepath.Join(installRoot, executablePath)), 0o755); err != nil {
		t.Fatalf("create managed Node install directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, executablePath), []byte("node executable"), 0o755); err != nil {
		t.Fatalf("write managed Node executable: %v", err)
	}
	// Manifest is the exact installation contract consumed by LuaSkills.
	// Manifest 是 LuaSkills 消费的精确安装契约。
	manifest := map[string]any{
		"schema_version": 1,
		"runtime":        "node",
		"version":        "24.18.0",
		"platform":       target.PlatformKey,
		"executable":     executablePath,
	}
	// ManifestBytes is the deterministic JSON payload written beside the executable.
	// ManifestBytes 是写入可执行文件旁的确定性 JSON 载荷。
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal managed Node manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "runtime-manifest.json"), manifestBytes, 0o644); err != nil {
		t.Fatalf("write managed Node manifest: %v", err)
	}

	// Descriptor is decoded from the real public JSON FFI response envelope.
	// Descriptor 从真实公共 JSON FFI 响应包络解码。
	descriptor, err := ResolveManagedRuntimeInstall(ManagedRuntimeResolveOptions{
		DistributionRoot: distributionRoot,
		Runtime:          ManagedRuntimeKindNode,
		Version:          "24.18.0",
		Platform:         target.PlatformKey,
	})
	if err != nil {
		t.Fatalf("resolve managed Node installation: %v", err)
	}
	if descriptor.Runtime != ManagedRuntimeKindNode {
		t.Fatalf("unexpected runtime kind: %s", descriptor.Runtime)
	}
	if descriptor.Version != "24.18.0" || descriptor.Platform != target.PlatformKey {
		t.Fatalf("unexpected runtime identity: %#v", descriptor)
	}
	// DescriptorInfo identifies the canonical path returned by Rust, including Windows verbatim form.
	// DescriptorInfo 标识 Rust 返回的规范路径，包括 Windows 逐字路径形式。
	descriptorInfo, err := os.Stat(descriptor.InstallRoot)
	if err != nil {
		t.Fatalf("stat resolved install root: %v", err)
	}
	// ExpectedInfo identifies the fixture path before Rust canonicalization.
	// ExpectedInfo 标识 Rust 规范化前的夹具路径。
	expectedInfo, err := os.Stat(installRoot)
	if err != nil {
		t.Fatalf("stat expected install root: %v", err)
	}
	if !os.SameFile(descriptorInfo, expectedInfo) {
		t.Fatalf("unexpected install root: %s", descriptor.InstallRoot)
	}
	if len(descriptor.ManifestHash) != 64 || len(descriptor.ExecutableHash) != 64 {
		t.Fatalf("unexpected descriptor hashes: %#v", descriptor)
	}
}
