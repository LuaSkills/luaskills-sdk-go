# LuaSkills Go SDK

Go SDK for integrating the LuaSkills runtime through the public JSON FFI surface.

Go SDK，用于通过公共 JSON FFI 接入 LuaSkills 运行时。

The SDK wraps cgo JSON FFI calls, engine lifecycle, formal skill roots, authority-aware management calls, skill config, and runtime manifest helpers.

SDK 封装了 cgo JSON FFI 调用、engine 生命周期、正式 skill root、带权限语义的管理调用、skill config 与 runtime manifest 辅助能力。

## Installation

```bash
go get github.com/LuaSkills/luaskills-sdk-go
```

Runtime calls require `CGO_ENABLED=1`, a C compiler compatible with Go cgo, and a discoverable LuaSkills dynamic library.

运行时调用需要 `CGO_ENABLED=1`、Go cgo 可用的 C 编译器，以及可被链接和加载的 LuaSkills 动态库。

Windows example:

Windows 示例：

```powershell
$env:CGO_ENABLED = "1"
$env:CGO_LDFLAGS = "-LD:\runtime\luaskills\libs"
$env:PATH = "D:\runtime\luaskills\libs;$env:PATH"
```

Linux / macOS example:

Linux / macOS 示例：

```bash
export CGO_ENABLED=1
export CGO_LDFLAGS="-L/opt/luaskills-runtime/libs"
export LD_LIBRARY_PATH="/opt/luaskills-runtime/libs:${LD_LIBRARY_PATH}"
```

## Runtime Assets

The Go SDK plans and consumes the shared SDK runtime manifest. It does not download release assets itself. Use the TypeScript or Python installer, or implement a host installer from the generated manifest.

Go SDK 会规划并消费共享 SDK runtime manifest，但它本身不下载 release 资产。请使用 TypeScript 或 Python 安装器，或基于生成的 manifest 实现宿主自己的安装器。

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root D:\runtime\luaskills
```

```powershell
pip install luaskills-sdk
luaskills install-runtime --database vldb-direct --runtime-root D:\runtime\luaskills
```

Go hosts can inspect the same asset plan:

Go 宿主可以检查同一份资产计划：

```go
manifest, err := luaskills.BuildRuntimeInstallManifest(luaskills.RuntimeInstallOptions{
	RuntimeRoot:    "D:/runtime/luaskills",
	Database:       luaskills.RuntimeDatabaseVldbDirect,
	SkipLuaRuntime: false,
})
if err != nil {
	panic(err)
}

hostOptions := luaskills.HostOptionsFromRuntimeManifest(manifest)
```

`DefaultHostOptions(runtimeRoot)` and `NewClient` automatically read `runtimeRoot/resources/luaskills-sdk-runtime-manifest.json` and merge `host_options_patch` when the manifest exists.

`DefaultHostOptions(runtimeRoot)` 与 `NewClient` 会在 manifest 存在时自动读取 `runtimeRoot/resources/luaskills-sdk-runtime-manifest.json`，并合入 `host_options_patch`。

Database modes:

数据库模式：

- `RuntimeDatabaseNone`: installs the Lua runtime archive and the LuaSkills FFI SDK archive, without database providers.
- `RuntimeDatabaseNone`：安装 Lua runtime 归档与 LuaSkills FFI SDK 归档，但不安装数据库 provider。
- `RuntimeDatabaseVldbController`: uses the `vldb-controller` executable through `space_controller` mode.
- `RuntimeDatabaseVldbController`：通过 `space_controller` 模式使用 `vldb-controller` 可执行文件。
- `RuntimeDatabaseVldbDirect`: uses `vldb-sqlite-lib` and `vldb-lancedb-lib` dynamic libraries.
- `RuntimeDatabaseVldbDirect`：使用 `vldb-sqlite-lib` 与 `vldb-lancedb-lib` 动态库。
- `RuntimeDatabaseHostCallback`: expects the host to provide JSON callbacks.
- `RuntimeDatabaseHostCallback`：由宿主提供 JSON callback。

## Basic Usage

Prepare `runtimeRoot`, then create a client:

准备好 `runtimeRoot` 后创建 client：

```go
package main

import (
	"fmt"

	luaskills "github.com/LuaSkills/luaskills-sdk-go"
)

func main() {
	runtimeRoot := "D:/runtime/luaskills"
	roots := luaskills.StandardRoots(runtimeRoot)

	client, err := luaskills.NewClient(luaskills.ClientOptions{
		RuntimeRoot:         runtimeRoot,
		EnsureRuntimeLayout: true,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	if _, err := client.LoadFromRoots(roots); err != nil {
		panic(err)
	}

	entries, err := client.ListEntries(luaskills.AuthorityDelegatedTool)
	if err != nil {
		panic(err)
	}

	result, err := client.CallSkill("demo-standard-ffi-skill-ping", map[string]any{
		"note": "go-sdk",
	}, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(entries)
	fmt.Println(result.Content)
}
```

## Examples

Detailed source-tree examples live under `examples/`.

详细源码示例位于 `examples/`。

```powershell
go run .\examples\basic
go run .\examples\call
go run .\examples\query
go run .\examples\lifecycle
go run .\examples\provider_callback
```

The query and lifecycle examples use the bundled fixture skill at `examples/fixture-runtime/user_skills/demo-standard-ffi-skill`. Prepare runtime assets with a TypeScript or Python installer first:

query 与 lifecycle 示例使用内置夹具 skill：`examples/fixture-runtime/user_skills/demo-standard-ffi-skill`。请先使用 TypeScript 或 Python 安装器准备 runtime 资产：

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root .\examples\fixture-runtime
```

See `examples/README.md` for the full example index and runtime notes.

完整示例索引与 runtime 注意事项见 `examples/README.md`。

## Authority And Management

Query APIs should use `AuthorityDelegatedTool` by default, so ROOT skills are hidden from delegated tools.

查询类接口建议默认使用 `AuthorityDelegatedTool`，因此委托工具看不到 ROOT skills。

`AuthoritySystem` only means the host may manage ROOT. It does not bypass ROOT ownership or same-`skill_id` conflict rules.

`AuthoritySystem` 只表示宿主可以管理 ROOT；它不表示可以绕过 ROOT 所有权或同名 `skill_id` 冲突规则。

`CallSkill` and `RunLua` are runtime execution surfaces. They are not ROOT visibility filters.

`CallSkill` 与 `RunLua` 是运行时执行面，不作为 ROOT 可见性过滤。

Skill config is a plain `skill_id + key` storage surface. Configuration only affects behavior when the Lua skill reads it.

skill config 是普通的 `skill_id + key` 配置存储面。配置只有在 Lua skill 主动读取时才会影响行为。

## JSON Provider Callback

The Go SDK exposes the callback API boundary, but does not install a process-level cgo callback bridge by default.

Go SDK 暴露 callback API 边界，但默认不在包内安装进程级 cgo callback bridge。

```go
err := luaskills.SetSQLiteProviderJSONCallback(func(request any) (any, error) {
	return map[string]any{"ok": true, "request": request}, nil
})
```

Currently this returns `ErrProviderCallbacksRequireHostBridge`. Production Go hosts that need `host_callback + json` should implement a controlled cgo callback bridge in the host process, or use the TypeScript / Python SDK for JSON callbacks.

当前该 API 会返回 `ErrProviderCallbacksRequireHostBridge`。需要 `host_callback + json` 的正式 Go 宿主，应在宿主进程内实现受控 cgo callback bridge，或先使用 TypeScript / Python SDK 接 JSON callback。

## Verification

Source-tree checks:

源码环境检查：

```powershell
$env:CGO_ENABLED = "0"
go test ./...
```

Full native FFI checks need `CGO_ENABLED=1` and a cgo-compatible compiler. On Windows, Visual Studio alone is usually not enough for Go cgo; install a MinGW-w64/UCRT64 toolchain or another Go-compatible GCC distribution.

完整原生 FFI 检查需要 `CGO_ENABLED=1` 与 cgo 兼容的 C 编译器。在 Windows 上，仅有 Visual Studio 通常不足以满足 Go cgo；请安装 MinGW-w64/UCRT64 工具链或其他 Go 兼容的 GCC 发行版。
