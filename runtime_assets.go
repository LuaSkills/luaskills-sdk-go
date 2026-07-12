package luaskills

import (
	"encoding/json"
	"fmt"
	"github.com/LuaSkills/luaskills-sdk-go/internal/runtimeassets"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DefaultLuaSkillsVersion is the release tag used by SDK runtime installation.
// DefaultLuaSkillsVersion 是 SDK 运行时安装使用的 LuaSkills 发布标签。
const DefaultLuaSkillsVersion = "v0.5.1"

// DefaultLuaSkillsPackagesSeries is the release series used by SDK runtime package installation.
// DefaultLuaSkillsPackagesSeries 是 SDK 运行时 package 安装使用的 luaskills-packages 发布协议线。
const DefaultLuaSkillsPackagesSeries = "0.1"

// DefaultVldbControllerVersion is the release tag used by SDK runtime installation.
// DefaultVldbControllerVersion 是 SDK 运行时安装使用的 vldb-controller 发布标签。
const DefaultVldbControllerVersion = "v0.2.1"

// DefaultVldbSQLiteVersion is the release tag used by SDK runtime installation.
// DefaultVldbSQLiteVersion 是 SDK 运行时安装使用的 vldb-sqlite 发布标签。
const DefaultVldbSQLiteVersion = "v0.1.5"

// DefaultVldbLanceDBVersion is the release tag used by SDK runtime installation.
// DefaultVldbLanceDBVersion 是 SDK 运行时安装使用的 vldb-lancedb 发布标签。
const DefaultVldbLanceDBVersion = "v0.1.5"

// DefaultManagedPythonVersion is the managed CPython version used by Lua-driven child runtimes.
// DefaultManagedPythonVersion 是 Lua 调度子运行时使用的受管 CPython 版本。
const DefaultManagedPythonVersion = "3.14.6"

// DefaultManagedUvVersion is the standalone uv version used by managed Python.
// DefaultManagedUvVersion 是受管 Python 使用的独立 uv 版本。
const DefaultManagedUvVersion = "0.11.28"

// DefaultManagedNodeVersion is the managed Node.js version used by Lua-driven child runtimes.
// DefaultManagedNodeVersion 是 Lua 调度子运行时使用的受管 Node.js 版本。
const DefaultManagedNodeVersion = "24.18.0"

// DefaultManagedPnpmVersion is the pnpm version used by managed Node.js.
// DefaultManagedPnpmVersion 是受管 Node.js 使用的 pnpm 版本。
const DefaultManagedPnpmVersion = "11.11.0"

// RuntimeManifestFileName is the manifest name stored under runtime resources.
// RuntimeManifestFileName 是存放在 runtime resources 下的清单文件名。
const RuntimeManifestFileName = "luaskills-sdk-runtime-manifest.json"

// RuntimeDatabasePreset is one SDK-level database integration mode.
// RuntimeDatabasePreset 是单个 SDK 级数据库集成模式。
type RuntimeDatabasePreset string

const (
	// RuntimeDatabaseNone does not install or configure database providers.
	// RuntimeDatabaseNone 不安装也不配置数据库 provider。
	RuntimeDatabaseNone RuntimeDatabasePreset = "none"
	// RuntimeDatabaseVldbController uses vldb-controller through space_controller mode.
	// RuntimeDatabaseVldbController 通过 space_controller 模式使用 vldb-controller。
	RuntimeDatabaseVldbController RuntimeDatabasePreset = "vldb-controller"
	// RuntimeDatabaseVldbDirect uses vldb-sqlite-lib and vldb-lancedb-lib directly.
	// RuntimeDatabaseVldbDirect 直接使用 vldb-sqlite-lib 与 vldb-lancedb-lib。
	RuntimeDatabaseVldbDirect RuntimeDatabasePreset = "vldb-direct"
	// RuntimeDatabaseHostCallback lets the host provide JSON callbacks.
	// RuntimeDatabaseHostCallback 由宿主提供 JSON callback。
	RuntimeDatabaseHostCallback RuntimeDatabasePreset = "host-callback"
)

// RuntimeAssetRole is a logical role for one release asset.
// RuntimeAssetRole 是单个发布资产的逻辑角色。
type RuntimeAssetRole string

const (
	// RuntimeAssetLuaRuntime identifies the Lua runtime archive.
	// RuntimeAssetLuaRuntime 标识 Lua runtime 归档。
	RuntimeAssetLuaRuntime RuntimeAssetRole = "lua_runtime"
	// RuntimeAssetLuaSkillsFFI identifies the LuaSkills FFI SDK archive.
	// RuntimeAssetLuaSkillsFFI 标识 LuaSkills FFI SDK 归档。
	RuntimeAssetLuaSkillsFFI RuntimeAssetRole = "luaskills_ffi"
	// RuntimeAssetVldbController identifies the vldb-controller executable archive.
	// RuntimeAssetVldbController 标识 vldb-controller 可执行文件归档。
	RuntimeAssetVldbController RuntimeAssetRole = "vldb_controller"
	// RuntimeAssetVldbSQLiteLib identifies the vldb-sqlite dynamic library archive.
	// RuntimeAssetVldbSQLiteLib 标识 vldb-sqlite 动态库归档。
	RuntimeAssetVldbSQLiteLib RuntimeAssetRole = "vldb_sqlite_lib"
	// RuntimeAssetVldbLanceDBLib identifies the vldb-lancedb dynamic library archive.
	// RuntimeAssetVldbLanceDBLib 标识 vldb-lancedb 动态库归档。
	RuntimeAssetVldbLanceDBLib RuntimeAssetRole = "vldb_lancedb_lib"
)

// ManagedRuntimeTarget is one SDK-level managed child runtime installation mode.
// ManagedRuntimeTarget 是单个 SDK 级受管子运行时安装模式。
type ManagedRuntimeTarget string

const (
	// ManagedRuntimeNone omits managed child runtimes.
	// ManagedRuntimeNone 省略受管子运行时。
	ManagedRuntimeNone ManagedRuntimeTarget = "none"
	// ManagedRuntimeAll includes Python, uv, Node.js, and pnpm.
	// ManagedRuntimeAll 包含 Python、uv、Node.js 与 pnpm。
	ManagedRuntimeAll ManagedRuntimeTarget = "all"
	// ManagedRuntimePython includes Python and uv.
	// ManagedRuntimePython 包含 Python 与 uv。
	ManagedRuntimePython ManagedRuntimeTarget = "python"
	// ManagedRuntimeNode includes Node.js and pnpm.
	// ManagedRuntimeNode 包含 Node.js 与 pnpm。
	ManagedRuntimeNode ManagedRuntimeTarget = "node"
	// ManagedRuntimePackageManagers includes uv and pnpm, plus Node.js for pnpm.
	// ManagedRuntimePackageManagers 包含 uv 与 pnpm，并包含 pnpm 所需的 Node.js。
	ManagedRuntimePackageManagers ManagedRuntimeTarget = "package-managers"
)

// RuntimePlatformTarget describes release asset naming for one platform.
// RuntimePlatformTarget 描述单个平台的发布资产命名。
type RuntimePlatformTarget struct {
	// PlatformKey is the LuaSkills platform key used by FFI SDK archives.
	// PlatformKey 是 FFI SDK 归档使用的 LuaSkills 平台标识。
	PlatformKey string `json:"platform_key"`
	// TargetTriple is the Rust-style target triple used by VLDB archives.
	// TargetTriple 是 VLDB 归档使用的 Rust 风格 target triple。
	TargetTriple string `json:"target_triple"`
	// ArchiveExt is the archive extension used by this platform.
	// ArchiveExt 是当前平台使用的归档扩展名。
	ArchiveExt string `json:"archive_ext"`
	// ControllerBinaryName is the vldb-controller executable file name.
	// ControllerBinaryName 是 vldb-controller 可执行文件名。
	ControllerBinaryName string `json:"controller_binary_name"`
	// DynamicLibraryExt is the dynamic library extension used by this platform.
	// DynamicLibraryExt 是当前平台使用的动态库扩展名。
	DynamicLibraryExt string `json:"dynamic_library_ext"`
	// LuaSkillsLibraryName is the expected installed LuaSkills library name.
	// LuaSkillsLibraryName 是预期安装后的 LuaSkills 动态库名称。
	LuaSkillsLibraryName string `json:"luaskills_library_name"`
	// SQLiteLibraryName is the expected installed SQLite library name.
	// SQLiteLibraryName 是预期安装后的 SQLite 动态库名称。
	SQLiteLibraryName string `json:"sqlite_library_name"`
	// LanceDBLibraryName is the expected installed LanceDB library name.
	// LanceDBLibraryName 是预期安装后的 LanceDB 动态库名称。
	LanceDBLibraryName string `json:"lancedb_library_name"`
}

// RuntimeAssetDescriptor describes one GitHub Release asset.
// RuntimeAssetDescriptor 描述单个 GitHub Release 资产。
type RuntimeAssetDescriptor struct {
	// Role is the logical asset role.
	// Role 是逻辑资产角色。
	Role RuntimeAssetRole `json:"role"`
	// Repository is the GitHub repository in owner/name form.
	// Repository 是 owner/name 形式的 GitHub 仓库。
	Repository string `json:"repository"`
	// Version is the release tag used by this asset.
	// Version 是当前资产使用的发布标签。
	Version string `json:"version"`
	// AssetName is the exact release asset file name.
	// AssetName 是精确的发布资产文件名。
	AssetName string `json:"asset_name"`
	// SHA256AssetName is the exact SHA-256 sidecar asset file name.
	// SHA256AssetName 是精确的 SHA-256 旁路资产文件名。
	SHA256AssetName string `json:"sha256_asset_name"`
	// DownloadURL is the browser download URL for the archive.
	// DownloadURL 是归档的浏览器下载地址。
	DownloadURL string `json:"download_url"`
	// SHA256URL is the browser download URL for the SHA-256 sidecar.
	// SHA256URL 是 SHA-256 旁路文件的浏览器下载地址。
	SHA256URL string `json:"sha256_url"`
	// InstalledPath is the relative installed executable, library, or marker path.
	// InstalledPath 是已安装可执行文件、动态库或标记文件的相对路径。
	InstalledPath *string `json:"installed_path"`
}

// ManagedRuntimePlatformTarget describes managed child runtime upstream assets.
// ManagedRuntimePlatformTarget 描述受管子运行时上游资产。
type ManagedRuntimePlatformTarget struct {
	// PlatformKey is the stable LuaSkills managed-runtime platform key.
	// PlatformKey 是稳定的 LuaSkills 受管运行时平台键。
	PlatformKey string `json:"platform_key"`
	// UvAssetName is the uv release asset name for this platform.
	// UvAssetName 是当前平台对应的 uv 发布资产名。
	UvAssetName string `json:"uv_asset_name"`
	// NodeAssetTemplate is the Node.js release asset template.
	// NodeAssetTemplate 是 Node.js 发布资产模板。
	NodeAssetTemplate string `json:"node_asset_template"`
	// NodeExtractTemplate is the Node.js archive top-level directory template.
	// NodeExtractTemplate 是 Node.js 归档顶层目录模板。
	NodeExtractTemplate string `json:"node_extract_template"`
	// UvExecutable is the relative uv executable path.
	// UvExecutable 是 uv 可执行文件相对路径。
	UvExecutable string `json:"uv_executable"`
	// NodeExecutable is the relative Node.js executable path.
	// NodeExecutable 是 Node.js 可执行文件相对路径。
	NodeExecutable string `json:"node_executable"`
}

// ManagedRuntimeInstallPlan describes planned managed child runtimes.
// ManagedRuntimeInstallPlan 描述已规划的受管子运行时。
type ManagedRuntimeInstallPlan struct {
	// Target is the requested managed runtime group.
	// Target 是请求的受管运行时分组。
	Target ManagedRuntimeTarget `json:"target"`
	// Platform contains managed runtime platform metadata.
	// Platform 包含受管运行时平台元数据。
	Platform ManagedRuntimePlatformTarget `json:"platform"`
	// PythonVersion is the managed Python version.
	// PythonVersion 是受管 Python 版本。
	PythonVersion string `json:"python_version"`
	// UvVersion is the managed uv version.
	// UvVersion 是受管 uv 版本。
	UvVersion string `json:"uv_version"`
	// NodeVersion is the managed Node.js version.
	// NodeVersion 是受管 Node.js 版本。
	NodeVersion string `json:"node_version"`
	// PnpmVersion is the managed pnpm version.
	// PnpmVersion 是受管 pnpm 版本。
	PnpmVersion string `json:"pnpm_version"`
	// InstalledPaths are relative installation paths under runtime root.
	// InstalledPaths 是 runtime root 下的相对安装路径。
	InstalledPaths map[string]string `json:"installed_paths"`
}

// RuntimeInstallManifest is the shared SDK runtime installation manifest.
// RuntimeInstallManifest 是共享 SDK 运行时安装清单。
type RuntimeInstallManifest struct {
	// SchemaVersion is the manifest schema version.
	// SchemaVersion 是清单结构版本。
	SchemaVersion int `json:"schema_version"`
	// GeneratedAt is the UTC timestamp when the manifest was generated.
	// GeneratedAt 是清单生成时的 UTC 时间戳。
	GeneratedAt string `json:"generated_at"`
	// RuntimeRoot is the runtime root represented by the manifest.
	// RuntimeRoot 是清单表示的 runtime root。
	RuntimeRoot string `json:"runtime_root"`
	// DatabaseMode is the selected database integration mode.
	// DatabaseMode 是选中的数据库集成模式。
	DatabaseMode RuntimeDatabasePreset `json:"database_mode"`
	// Platform is the platform target used by manifest assets.
	// Platform 是清单资产使用的平台目标。
	Platform RuntimePlatformTarget `json:"platform"`
	// Assets are required by the selected runtime mode.
	// Assets 是选中运行时模式所需的资产。
	Assets []RuntimeAssetDescriptor `json:"assets"`
	// HostOptionsPatch is derived from installed runtime assets.
	// HostOptionsPatch 是从已安装运行时资产派生的宿主选项补丁。
	HostOptionsPatch map[string]any `json:"host_options_patch"`
	// ManagedRuntimes describes planned managed Python and Node.js child runtimes.
	// ManagedRuntimes 描述规划中的受管 Python 与 Node.js 子运行时。
	ManagedRuntimes *ManagedRuntimeInstallPlan `json:"managed_runtimes,omitempty"`
}

// RuntimeInstallOptions controls runtime asset planning.
// RuntimeInstallOptions 控制运行时资产规划。
type RuntimeInstallOptions struct {
	// RuntimeRoot receives native assets and the manifest.
	// RuntimeRoot 接收原生资产与清单。
	RuntimeRoot string
	// Database selects the SDK-level database integration mode.
	// Database 选择 SDK 级数据库集成模式。
	Database RuntimeDatabasePreset
	// LuaSkillsVersion is the LuaSkills release tag.
	// LuaSkillsVersion 是 LuaSkills 发布标签。
	LuaSkillsVersion string
	// LuaRuntimeVersion is the runtime packages release tag.
	// LuaRuntimeVersion 是 runtime packages 发布标签。
	LuaRuntimeVersion string
	// LuaRuntimeSeries is the runtime packages release series.
	// LuaRuntimeSeries 是 runtime packages 发布协议线。
	LuaRuntimeSeries string
	// VldbControllerVersion is the vldb-controller release tag.
	// VldbControllerVersion 是 vldb-controller 发布标签。
	VldbControllerVersion string
	// VldbSQLiteVersion is the vldb-sqlite release tag.
	// VldbSQLiteVersion 是 vldb-sqlite 发布标签。
	VldbSQLiteVersion string
	// VldbLanceDBVersion is the vldb-lancedb release tag.
	// VldbLanceDBVersion 是 vldb-lancedb 发布标签。
	VldbLanceDBVersion string
	// SkipLuaSkillsFFI omits the LuaSkills FFI SDK archive from the manifest.
	// SkipLuaSkillsFFI 从清单中省略 LuaSkills FFI SDK 归档。
	SkipLuaSkillsFFI bool
	// SkipLuaRuntime omits the Lua runtime archive from the manifest.
	// SkipLuaRuntime 从清单中省略 Lua runtime 归档。
	SkipLuaRuntime bool
	// LuaSkillsRepo is the GitHub repository that publishes LuaSkills assets.
	// LuaSkillsRepo 是发布 LuaSkills 资产的 GitHub 仓库。
	LuaSkillsRepo string
	// LuaRuntimeRepo is the GitHub repository that publishes runtime packages assets.
	// LuaRuntimeRepo 是发布 runtime packages 资产的 GitHub 仓库。
	LuaRuntimeRepo string
	// VldbControllerRepo is the GitHub repository that publishes vldb-controller assets.
	// VldbControllerRepo 是发布 vldb-controller 资产的 GitHub 仓库。
	VldbControllerRepo string
	// VldbSQLiteRepo is the GitHub repository that publishes vldb-sqlite assets.
	// VldbSQLiteRepo 是发布 vldb-sqlite 资产的 GitHub 仓库。
	VldbSQLiteRepo string
	// VldbLanceDBRepo is the GitHub repository that publishes vldb-lancedb assets.
	// VldbLanceDBRepo 是发布 vldb-lancedb 资产的 GitHub 仓库。
	VldbLanceDBRepo string
	// ManagedRuntimes selects managed child runtimes for SDK installers.
	// ManagedRuntimes 为 SDK 安装器选择受管子运行时。
	ManagedRuntimes ManagedRuntimeTarget
	// ManagedPythonVersion is the managed Python version.
	// ManagedPythonVersion 是受管 Python 版本。
	ManagedPythonVersion string
	// ManagedUvVersion is the managed uv version.
	// ManagedUvVersion 是受管 uv 版本。
	ManagedUvVersion string
	// ManagedNodeVersion is the managed Node.js version.
	// ManagedNodeVersion 是受管 Node.js 版本。
	ManagedNodeVersion string
	// ManagedPnpmVersion is the managed pnpm version.
	// ManagedPnpmVersion 是受管 pnpm 版本。
	ManagedPnpmVersion string
}

// ResolveRuntimePlatformTarget returns the release target for the current Go process.
// ResolveRuntimePlatformTarget 返回当前 Go 进程对应的发布目标。
func ResolveRuntimePlatformTarget() (RuntimePlatformTarget, error) {
	return ResolveRuntimePlatformTargetFor(runtime.GOOS, runtime.GOARCH)
}

// ResolveRuntimePlatformTargetFor returns the release target for explicit platform values.
// ResolveRuntimePlatformTargetFor 返回显式平台值对应的发布目标。
func ResolveRuntimePlatformTargetFor(goos string, goarch string) (RuntimePlatformTarget, error) {
	if goos == "windows" && goarch == "amd64" {
		return RuntimePlatformTarget{
			PlatformKey:          "windows-x64",
			TargetTriple:         "x86_64-pc-windows-msvc",
			ArchiveExt:           ".zip",
			ControllerBinaryName: "vldb-controller.exe",
			DynamicLibraryExt:    ".dll",
			LuaSkillsLibraryName: "luaskills.dll",
			SQLiteLibraryName:    "vldb_sqlite.dll",
			LanceDBLibraryName:   "vldb_lancedb.dll",
		}, nil
	}
	if goos == "darwin" && goarch == "amd64" {
		return darwinRuntimeTarget("x86_64", "macos-x64"), nil
	}
	if goos == "darwin" && goarch == "arm64" {
		return darwinRuntimeTarget("aarch64", "macos-arm64"), nil
	}
	if goos == "linux" && goarch == "amd64" {
		return linuxRuntimeTarget("x86_64", "linux-x64"), nil
	}
	if goos == "linux" && goarch == "arm64" {
		return linuxRuntimeTarget("aarch64", "linux-arm64"), nil
	}
	return RuntimePlatformTarget{}, fmt.Errorf("unsupported runtime platform: %s/%s", goos, goarch)
}

// BuildRuntimeInstallManifest builds one deterministic runtime installation manifest.
// BuildRuntimeInstallManifest 构造一个确定性的运行时安装清单。
func BuildRuntimeInstallManifest(options RuntimeInstallOptions) (*RuntimeInstallManifest, error) {
	target, err := ResolveRuntimePlatformTarget()
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeRuntimeInstallOptions(options)
	if err != nil {
		return nil, err
	}
	assets := buildRuntimeAssetDescriptors(normalized, target)
	managedRuntimes, err := buildManagedRuntimeInstallPlan(normalized)
	if err != nil {
		return nil, err
	}
	manifest := &RuntimeInstallManifest{
		SchemaVersion:    1,
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		RuntimeRoot:      normalizePath(normalized.RuntimeRoot),
		DatabaseMode:     normalized.Database,
		Platform:         target,
		Assets:           assets,
		HostOptionsPatch: buildRuntimeHostOptionsPatch(normalized.RuntimeRoot, normalized.Database, target, assets),
		ManagedRuntimes:  managedRuntimes,
	}
	return manifest, nil
}

// HostOptionsFromRuntimeManifest converts one runtime manifest into host option overrides.
// HostOptionsFromRuntimeManifest 将单个运行时清单转换为宿主选项覆盖。
func HostOptionsFromRuntimeManifest(manifest *RuntimeInstallManifest) (map[string]any, error) {
	if manifest == nil {
		return map[string]any{}, nil
	}
	return sanitizeRuntimeManifestHostOptions(manifest.RuntimeRoot, manifest.HostOptionsPatch)
}

// LoadRuntimeInstallManifest reads one SDK runtime install manifest from a runtime root.
// LoadRuntimeInstallManifest 从单个 runtime root 读取 SDK 运行时安装清单。
func LoadRuntimeInstallManifest(runtimeRoot string) (*RuntimeInstallManifest, error) {
	manifestPath := RuntimeManifestPath(runtimeRoot)
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read runtime install manifest %s: %w", manifestPath, err)
	}
	return decodeRuntimeInstallManifest(manifestPath, raw)
}

// RuntimeManifestPath returns the expected manifest path under one runtime root.
// RuntimeManifestPath 返回单个 runtime root 下的预期清单路径。
func RuntimeManifestPath(runtimeRoot string) string {
	return filepath.Join(runtimeRoot, "resources", RuntimeManifestFileName)
}

// sanitizeRuntimeManifestHostOptions validates runtime-root path fields from one manifest patch.
// sanitizeRuntimeManifestHostOptions 校验单个 manifest patch 中受 runtime-root 约束的路径字段。
func sanitizeRuntimeManifestHostOptions(runtimeRoot string, patch map[string]any) (map[string]any, error) {
	if strings.TrimSpace(runtimeRoot) == "" {
		return nil, fmt.Errorf("runtime manifest runtime_root must be a string path")
	}
	sanitized := mergeMaps(map[string]any{}, patch)
	for _, key := range []string{"sqlite_library_path", "lancedb_library_path"} {
		path, err := sanitizeRuntimeManifestPath(runtimeRoot, sanitized[key], key)
		if err != nil {
			return nil, err
		}
		if path != nil {
			sanitized[key] = *path
		}
	}
	if rawSpaceController, ok := sanitized["space_controller"]; ok && rawSpaceController != nil {
		spaceController, ok := rawSpaceController.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("host_options_patch.space_controller must be one object")
		}
		spaceCopy := mergeMaps(map[string]any{}, spaceController)
		path, err := sanitizeRuntimeManifestPath(runtimeRoot, spaceCopy["executable_path"], "space_controller.executable_path")
		if err != nil {
			return nil, err
		}
		if path != nil {
			spaceCopy["executable_path"] = *path
		}
		sanitized["space_controller"] = spaceCopy
	}
	return sanitized, nil
}

// sanitizeRuntimeManifestPath validates one runtime-root-scoped host option path.
// sanitizeRuntimeManifestPath 校验单个受 runtime-root 约束的宿主选项路径。
func sanitizeRuntimeManifestPath(runtimeRoot string, value any, context string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	pathText, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("host_options_patch.%s must be a string path", context)
	}
	if strings.TrimSpace(pathText) == "" || strings.ContainsRune(pathText, '\x00') {
		return nil, fmt.Errorf("host_options_patch.%s must be a path inside runtime root", context)
	}
	rootPath, err := filepath.Abs(runtimeRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve runtime root for host_options_patch.%s: %w", context, err)
	}
	candidatePath := pathText
	if !filepath.IsAbs(candidatePath) {
		if runtimeassets.HasUnsafeRelativeSegment(candidatePath) {
			return nil, fmt.Errorf("host_options_patch.%s must be a path inside runtime root", context)
		}
		candidatePath = filepath.Join(rootPath, candidatePath)
	}
	candidatePath, err = filepath.Abs(candidatePath)
	if err != nil {
		return nil, fmt.Errorf("resolve host_options_patch.%s: %w", context, err)
	}
	if !runtimeassets.IsStrictlyInside(rootPath, candidatePath) {
		return nil, fmt.Errorf("host_options_patch.%s escapes runtime root: %s", context, pathText)
	}
	normalized := normalizePath(candidatePath)
	return &normalized, nil
}

// decodeRuntimeInstallManifest decodes one runtime install manifest with path-aware diagnostics.
// decodeRuntimeInstallManifest 使用带路径上下文的诊断解码单个运行时安装清单。
func decodeRuntimeInstallManifest(manifestPath string, raw []byte) (*RuntimeInstallManifest, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, fmt.Errorf("runtime install manifest %s is empty", manifestPath)
	}
	var decoded any
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return nil, fmt.Errorf("runtime install manifest %s is invalid JSON: %w", manifestPath, err)
	}
	if _, ok := decoded.(map[string]any); !ok {
		return nil, fmt.Errorf("runtime install manifest %s must be one JSON object", manifestPath)
	}
	var manifest RuntimeInstallManifest
	if err := json.Unmarshal([]byte(trimmed), &manifest); err != nil {
		return nil, fmt.Errorf("runtime install manifest %s has invalid schema: %w", manifestPath, err)
	}
	return &manifest, nil
}

// normalizeRuntimeInstallOptions fills default release repositories and versions.
// normalizeRuntimeInstallOptions 填充默认发布仓库与版本。
func normalizeRuntimeInstallOptions(options RuntimeInstallOptions) (RuntimeInstallOptions, error) {
	if options.Database == "" {
		options.Database = RuntimeDatabaseNone
	}
	if options.LuaSkillsVersion == "" {
		options.LuaSkillsVersion = DefaultLuaSkillsVersion
	}
	if options.LuaRuntimeSeries == "" {
		options.LuaRuntimeSeries = DefaultLuaSkillsPackagesSeries
	}
	if options.LuaRuntimeVersion == "" {
		resolvedTag, err := resolveReleaseTagForSeries(options.LuaRuntimeRepoOrDefault(), options.LuaRuntimeSeries)
		if err != nil {
			return RuntimeInstallOptions{}, err
		}
		options.LuaRuntimeVersion = resolvedTag
	}
	if options.VldbControllerVersion == "" {
		options.VldbControllerVersion = DefaultVldbControllerVersion
	}
	if options.VldbSQLiteVersion == "" {
		options.VldbSQLiteVersion = DefaultVldbSQLiteVersion
	}
	if options.VldbLanceDBVersion == "" {
		options.VldbLanceDBVersion = DefaultVldbLanceDBVersion
	}
	if options.LuaSkillsRepo == "" {
		options.LuaSkillsRepo = "LuaSkills/luaskills"
	}
	if options.LuaRuntimeRepo == "" {
		options.LuaRuntimeRepo = "LuaSkills/luaskills-packages"
	}
	if options.VldbControllerRepo == "" {
		options.VldbControllerRepo = "OpenVulcan/vldb-controller"
	}
	if options.VldbSQLiteRepo == "" {
		options.VldbSQLiteRepo = "OpenVulcan/vldb-sqlite"
	}
	if options.VldbLanceDBRepo == "" {
		options.VldbLanceDBRepo = "OpenVulcan/vldb-lancedb"
	}
	if options.ManagedRuntimes == "" {
		options.ManagedRuntimes = ManagedRuntimeNone
	}
	if err := validateManagedRuntimeTarget(options.ManagedRuntimes); err != nil {
		return RuntimeInstallOptions{}, err
	}
	if options.ManagedPythonVersion == "" {
		options.ManagedPythonVersion = DefaultManagedPythonVersion
	}
	if options.ManagedUvVersion == "" {
		options.ManagedUvVersion = DefaultManagedUvVersion
	}
	if options.ManagedNodeVersion == "" {
		options.ManagedNodeVersion = DefaultManagedNodeVersion
	}
	if options.ManagedPnpmVersion == "" {
		options.ManagedPnpmVersion = DefaultManagedPnpmVersion
	}
	return options, nil
}

// ResolveManagedRuntimePlatformTarget returns the managed-runtime target for the current Go process.
// ResolveManagedRuntimePlatformTarget 返回当前 Go 进程对应的受管运行时目标。
func ResolveManagedRuntimePlatformTarget() (ManagedRuntimePlatformTarget, error) {
	return ResolveManagedRuntimePlatformTargetFor(runtime.GOOS, runtime.GOARCH)
}

// ResolveManagedRuntimePlatformTargetFor returns the managed-runtime target for explicit platform values.
// ResolveManagedRuntimePlatformTargetFor 返回显式平台值对应的受管运行时目标。
func ResolveManagedRuntimePlatformTargetFor(goos string, goarch string) (ManagedRuntimePlatformTarget, error) {
	if goos == "windows" && goarch == "amd64" {
		return ManagedRuntimePlatformTarget{
			PlatformKey:         "windows-x64",
			UvAssetName:         "uv-x86_64-pc-windows-msvc.zip",
			NodeAssetTemplate:   "node-v{version}-win-x64.zip",
			NodeExtractTemplate: "node-v{version}-win-x64",
			UvExecutable:        "uv.exe",
			NodeExecutable:      "node.exe",
		}, nil
	}
	if goos == "darwin" && goarch == "amd64" {
		return managedUnixRuntimeTarget("macos-x64", "x86_64", "x64", "darwin", ".tar.gz"), nil
	}
	if goos == "darwin" && goarch == "arm64" {
		return managedUnixRuntimeTarget("macos-arm64", "aarch64", "arm64", "darwin", ".tar.gz"), nil
	}
	if goos == "linux" && goarch == "amd64" {
		return managedUnixRuntimeTarget("linux-x64", "x86_64", "x64", "linux", ".tar.xz"), nil
	}
	if goos == "linux" && goarch == "arm64" {
		return managedUnixRuntimeTarget("linux-arm64", "aarch64", "arm64", "linux", ".tar.xz"), nil
	}
	return ManagedRuntimePlatformTarget{}, fmt.Errorf("unsupported managed runtime platform: %s/%s", goos, goarch)
}

// buildManagedRuntimeInstallPlan builds one managed runtime section for the SDK manifest.
// buildManagedRuntimeInstallPlan 为 SDK 清单构造一个受管运行时片段。
func buildManagedRuntimeInstallPlan(options RuntimeInstallOptions) (*ManagedRuntimeInstallPlan, error) {
	if options.ManagedRuntimes == ManagedRuntimeNone {
		return nil, nil
	}
	target, err := ResolveManagedRuntimePlatformTarget()
	if err != nil {
		return nil, err
	}
	return &ManagedRuntimeInstallPlan{
		Target:        options.ManagedRuntimes,
		Platform:      target,
		PythonVersion: options.ManagedPythonVersion,
		UvVersion:     options.ManagedUvVersion,
		NodeVersion:   options.ManagedNodeVersion,
		PnpmVersion:   options.ManagedPnpmVersion,
		InstalledPaths: managedRuntimeInstalledPaths(
			target,
			options.ManagedPythonVersion,
			options.ManagedUvVersion,
			options.ManagedNodeVersion,
			options.ManagedPnpmVersion,
		),
	}, nil
}

// validateManagedRuntimeTarget rejects unknown managed runtime target values.
// validateManagedRuntimeTarget 拒绝未知的受管运行时目标值。
func validateManagedRuntimeTarget(target ManagedRuntimeTarget) error {
	switch target {
	case ManagedRuntimeNone, ManagedRuntimeAll, ManagedRuntimePython, ManagedRuntimeNode, ManagedRuntimePackageManagers:
		return nil
	default:
		return fmt.Errorf("unsupported managed runtime target: %s", target)
	}
}

// managedUnixRuntimeTarget builds one Unix-like managed runtime target descriptor.
// managedUnixRuntimeTarget 构造一个类 Unix 受管运行时目标描述。
func managedUnixRuntimeTarget(platformKey string, rustArch string, nodeArch string, nodeOS string, nodeArchiveExt string) ManagedRuntimePlatformTarget {
	uvOS := "unknown-linux-gnu"
	if nodeOS == "darwin" {
		uvOS = "apple-darwin"
	}
	return ManagedRuntimePlatformTarget{
		PlatformKey:         platformKey,
		UvAssetName:         fmt.Sprintf("uv-%s-%s.tar.gz", rustArch, uvOS),
		NodeAssetTemplate:   fmt.Sprintf("node-v{version}-%s-%s%s", nodeOS, nodeArch, nodeArchiveExt),
		NodeExtractTemplate: fmt.Sprintf("node-v{version}-%s-%s", nodeOS, nodeArch),
		UvExecutable:        "uv",
		NodeExecutable:      "bin/node",
	}
}

// managedRuntimeInstalledPaths builds relative managed runtime installation paths.
// managedRuntimeInstalledPaths 构造受管运行时相对安装路径。
func managedRuntimeInstalledPaths(target ManagedRuntimePlatformTarget, pythonVersion string, uvVersion string, nodeVersion string, pnpmVersion string) map[string]string {
	return map[string]string{
		"python": fmt.Sprintf("dependencies/runtimes/python/cpython-%s-%s", pythonVersion, target.PlatformKey),
		"uv":     fmt.Sprintf("dependencies/runtimes/python/uv-%s-%s", uvVersion, target.PlatformKey),
		"node":   fmt.Sprintf("dependencies/runtimes/node/node-%s-%s", nodeVersion, target.PlatformKey),
		"pnpm":   fmt.Sprintf("dependencies/runtimes/node/pnpm-%s", pnpmVersion),
	}
}

// LuaRuntimeRepoOrDefault returns the configured runtime packages repository or the SDK default.
// LuaRuntimeRepoOrDefault 返回已配置的 runtime packages 仓库或 SDK 默认仓库。
func (options RuntimeInstallOptions) LuaRuntimeRepoOrDefault() string {
	if options.LuaRuntimeRepo != "" {
		return options.LuaRuntimeRepo
	}
	return "LuaSkills/luaskills-packages"
}

// darwinRuntimeTarget builds one macOS runtime platform descriptor.
// darwinRuntimeTarget 构造单个 macOS 运行时平台描述。
func darwinRuntimeTarget(archPrefix string, platformKey string) RuntimePlatformTarget {
	return RuntimePlatformTarget{
		PlatformKey:          platformKey,
		TargetTriple:         archPrefix + "-apple-darwin",
		ArchiveExt:           ".tar.gz",
		ControllerBinaryName: "vldb-controller",
		DynamicLibraryExt:    ".dylib",
		LuaSkillsLibraryName: "libluaskills.dylib",
		SQLiteLibraryName:    "libvldb_sqlite.dylib",
		LanceDBLibraryName:   "libvldb_lancedb.dylib",
	}
}

// linuxRuntimeTarget builds one Linux runtime platform descriptor.
// linuxRuntimeTarget 构造单个 Linux 运行时平台描述。
func linuxRuntimeTarget(archPrefix string, platformKey string) RuntimePlatformTarget {
	return RuntimePlatformTarget{
		PlatformKey:          platformKey,
		TargetTriple:         archPrefix + "-unknown-linux-gnu",
		ArchiveExt:           ".tar.gz",
		ControllerBinaryName: "vldb-controller",
		DynamicLibraryExt:    ".so",
		LuaSkillsLibraryName: "libluaskills.so",
		SQLiteLibraryName:    "libvldb_sqlite.so",
		LanceDBLibraryName:   "libvldb_lancedb.so",
	}
}

// buildRuntimeAssetDescriptors builds every asset required by one manifest.
// buildRuntimeAssetDescriptors 构造单个清单所需的全部资产。
func buildRuntimeAssetDescriptors(options RuntimeInstallOptions, target RuntimePlatformTarget) []RuntimeAssetDescriptor {
	assets := []RuntimeAssetDescriptor{}
	if !options.SkipLuaRuntime {
		assetName := fmt.Sprintf("lua-runtime-packages-%s.tar.gz", target.PlatformKey)
		assets = append(assets, releaseRuntimeAsset(RuntimeAssetLuaRuntime, options.LuaRuntimeRepo, options.LuaRuntimeVersion, assetName, stringPtr("resources/lua-runtime-manifest.json")))
	}
	if !options.SkipLuaSkillsFFI {
		assetName := fmt.Sprintf("luaskills-ffi-sdk-%s.tar.gz", target.PlatformKey)
		assets = append(assets, releaseRuntimeAsset(RuntimeAssetLuaSkillsFFI, options.LuaSkillsRepo, options.LuaSkillsVersion, assetName, stringPtr("libs/"+target.LuaSkillsLibraryName)))
	}
	if options.Database == RuntimeDatabaseVldbController {
		assetName := fmt.Sprintf("vldb-controller-%s-%s%s", options.VldbControllerVersion, target.TargetTriple, target.ArchiveExt)
		assets = append(assets, releaseRuntimeAsset(RuntimeAssetVldbController, options.VldbControllerRepo, options.VldbControllerVersion, assetName, stringPtr("bin/"+target.ControllerBinaryName)))
	}
	if options.Database == RuntimeDatabaseVldbDirect {
		sqliteAsset := fmt.Sprintf("vldb-sqlite-lib-%s-%s%s", options.VldbSQLiteVersion, target.TargetTriple, target.ArchiveExt)
		lancedbAsset := fmt.Sprintf("vldb-lancedb-lib-%s-%s%s", options.VldbLanceDBVersion, target.TargetTriple, target.ArchiveExt)
		assets = append(assets, releaseRuntimeAsset(RuntimeAssetVldbSQLiteLib, options.VldbSQLiteRepo, options.VldbSQLiteVersion, sqliteAsset, stringPtr("libs/"+target.SQLiteLibraryName)))
		assets = append(assets, releaseRuntimeAsset(RuntimeAssetVldbLanceDBLib, options.VldbLanceDBRepo, options.VldbLanceDBVersion, lancedbAsset, stringPtr("libs/"+target.LanceDBLibraryName)))
	}
	return assets
}

// releaseRuntimeAsset builds one release asset descriptor from exact naming inputs.
// releaseRuntimeAsset 从精确命名输入构造单个发布资产描述。
func releaseRuntimeAsset(role RuntimeAssetRole, repository string, version string, assetName string, installedPath *string) RuntimeAssetDescriptor {
	baseURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repository, version, assetName)
	return RuntimeAssetDescriptor{
		Role:            role,
		Repository:      repository,
		Version:         version,
		AssetName:       assetName,
		SHA256AssetName: assetName + ".sha256",
		DownloadURL:     baseURL,
		SHA256URL:       baseURL + ".sha256",
		InstalledPath:   installedPath,
	}
}

// releaseSemVer stores one parsed semantic-version tag.
// releaseSemVer 保存单个已解析的语义化版本标签。
type releaseSemVer struct {
	major int
	minor int
	patch int
	tag   string
}

// resolveReleaseTagForSeries resolves the newest published release inside one semantic-version series.
// resolveReleaseTagForSeries 解析单个语义化版本协议线中的最新已发布版本。
func resolveReleaseTagForSeries(repository string, series string) (string, error) {
	major, minor, err := parseReleaseSeries(series)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=100", repository), nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "luaskills-sdk-go")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github releases api returned %d for %s", response.StatusCode, repository)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	var releases []struct {
		TagName    string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
	}
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", err
	}
	candidates := make([]releaseSemVer, 0)
	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}
		parsed, ok := parseReleaseTagSemVer(release.TagName)
		if !ok || parsed.major != major || parsed.minor != minor {
			continue
		}
		candidates = append(candidates, parsed)
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no published release found in series %s for %s", series, repository)
	}
	sort.Slice(candidates, func(left int, right int) bool {
		return candidates[left].patch > candidates[right].patch
	})
	return candidates[0].tag, nil
}

// parseReleaseSeries parses one release series string like 0.1.
// parseReleaseSeries 解析形如 0.1 的发布协议线字符串。
func parseReleaseSeries(series string) (int, int, error) {
	parts := strings.Split(series, ".")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid release series: %s", series)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid release series: %s", series)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid release series: %s", series)
	}
	return major, minor, nil
}

// parseReleaseTagSemVer parses one release tag when it follows vX.Y.Z or X.Y.Z.
// parseReleaseTagSemVer 在发布标签符合 vX.Y.Z 或 X.Y.Z 时解析该标签。
func parseReleaseTagSemVer(tag string) (releaseSemVer, bool) {
	normalized := strings.TrimPrefix(tag, "v")
	parts := strings.Split(normalized, ".")
	if len(parts) != 3 {
		return releaseSemVer{}, false
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return releaseSemVer{}, false
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return releaseSemVer{}, false
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return releaseSemVer{}, false
	}
	return releaseSemVer{
		major: major,
		minor: minor,
		patch: patch,
		tag:   tag,
	}, true
}

// buildRuntimeHostOptionsPatch builds host option overrides for one database mode.
// buildRuntimeHostOptionsPatch 为单个数据库模式构造宿主选项覆盖。
func buildRuntimeHostOptionsPatch(runtimeRoot string, database RuntimeDatabasePreset, target RuntimePlatformTarget, assets []RuntimeAssetDescriptor) map[string]any {
	root := normalizePath(runtimeRoot)
	if database == RuntimeDatabaseHostCallback {
		return map[string]any{
			"sqlite_provider_mode":  "host_callback",
			"sqlite_callback_mode":  "json",
			"lancedb_provider_mode": "host_callback",
			"lancedb_callback_mode": "json",
		}
	}
	if database == RuntimeDatabaseVldbController {
		return map[string]any{
			"sqlite_provider_mode":  "space_controller",
			"lancedb_provider_mode": "space_controller",
			"space_controller": map[string]any{
				"endpoint":                  nil,
				"auto_spawn":                true,
				"executable_path":           normalizePath(filepath.Join(root, "bin", target.ControllerBinaryName)),
				"process_mode":              "managed",
				"minimum_uptime_secs":       300,
				"idle_timeout_secs":         900,
				"default_lease_ttl_secs":    120,
				"connect_timeout_secs":      5,
				"startup_timeout_secs":      15,
				"startup_retry_interval_ms": 250,
				"lease_renew_interval_secs": 30,
			},
		}
	}
	if database == RuntimeDatabaseVldbDirect {
		return map[string]any{
			"sqlite_library_path":   resolveRuntimeInstalledAsset(root, assets, RuntimeAssetVldbSQLiteLib),
			"sqlite_provider_mode":  "dynamic_library",
			"lancedb_library_path":  resolveRuntimeInstalledAsset(root, assets, RuntimeAssetVldbLanceDBLib),
			"lancedb_provider_mode": "dynamic_library",
		}
	}
	return map[string]any{}
}

// resolveRuntimeInstalledAsset resolves the absolute installed path for one asset role.
// resolveRuntimeInstalledAsset 解析单个资产角色的绝对安装路径。
func resolveRuntimeInstalledAsset(runtimeRoot string, assets []RuntimeAssetDescriptor, role RuntimeAssetRole) any {
	for _, asset := range assets {
		if asset.Role == role && asset.InstalledPath != nil {
			return normalizePath(filepath.Join(runtimeRoot, *asset.InstalledPath))
		}
	}
	return nil
}

// stringPtr returns a pointer to one string literal.
// stringPtr 返回单个字符串字面量的指针。
func stringPtr(value string) *string {
	return &value
}
