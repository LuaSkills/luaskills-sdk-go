# LuaSkills Go SDK

English documentation is the default package documentation. For Chinese, see [README_cn.md](README_cn.md).

Main LuaSkills repository: [LuaSkills/luaskills](https://github.com/LuaSkills/luaskills)

Go SDK for integrating the LuaSkills runtime through the public JSON FFI surface.

The SDK wraps cgo JSON FFI calls, engine lifecycle, formal skill roots, authority-aware management calls, skill config, and runtime manifest helpers.

## Installation

```bash
go get github.com/LuaSkills/luaskills-sdk-go
```

Runtime calls require `CGO_ENABLED=1`, a C compiler compatible with Go cgo, and a discoverable LuaSkills dynamic library.

Windows example:

```powershell
$env:CGO_ENABLED = "1"
$env:CGO_LDFLAGS = "-LD:\runtime\luaskills\libs"
$env:PATH = "D:\runtime\luaskills\libs;$env:PATH"
```

Linux / macOS example:

```bash
export CGO_ENABLED=1
export CGO_LDFLAGS="-L/opt/luaskills-runtime/libs"
export LD_LIBRARY_PATH="/opt/luaskills-runtime/libs:${LD_LIBRARY_PATH}"
```

## Runtime Assets

The Go SDK plans and consumes the shared SDK runtime manifest. It does not download release assets itself. Use the TypeScript or Python installer, or implement a host installer from the generated manifest.

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root D:\runtime\luaskills
```

```powershell
pip install luaskills-sdk
luaskills install-runtime --database vldb-direct --runtime-root D:\runtime\luaskills
```

Go hosts can inspect the same asset plan:

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

Database modes:

- `RuntimeDatabaseNone`: installs the Lua runtime archive and the LuaSkills FFI SDK archive, without database providers.
- `RuntimeDatabaseVldbController`: uses the `vldb-controller` executable through `space_controller` mode.
- `RuntimeDatabaseVldbDirect`: uses `vldb-sqlite-lib` and `vldb-lancedb-lib` dynamic libraries.
- `RuntimeDatabaseHostCallback`: expects the host to provide JSON callbacks.

## Basic Usage

Prepare `runtimeRoot`, then create a client:

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

```powershell
go run .\examples\basic
go run .\examples\call
go run .\examples\query
go run .\examples\lifecycle
go run .\examples\provider_callback
```

The query and lifecycle examples use the bundled fixture skill at `examples/fixture-runtime/user_skills/demo-standard-ffi-skill`. Prepare runtime assets with a TypeScript or Python installer first:

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root .\examples\fixture-runtime
```

See [examples/README.md](examples/README.md) for the full example index and runtime notes. The Chinese example guide is [examples/README_cn.md](examples/README_cn.md).

## Authority And Management

Query APIs should use `AuthorityDelegatedTool` by default, so ROOT skills are hidden from delegated tools.

`AuthoritySystem` only means the host may manage ROOT. It does not bypass ROOT ownership or same-`skill_id` conflict rules.

`CallSkill` and `RunLua` are runtime execution surfaces. They are not ROOT visibility filters.

Skill config is a plain `skill_id + key` storage surface. Configuration only affects behavior when the Lua skill reads it.

## JSON Provider Callback

The Go SDK exposes the callback API boundary, but does not install a process-level cgo callback bridge by default.

```go
err := luaskills.SetSQLiteProviderJSONCallback(func(request any) (any, error) {
	return map[string]any{"ok": true, "request": request}, nil
})
```

Currently this returns `ErrProviderCallbacksRequireHostBridge`. Production Go hosts that need `host_callback + json` should implement a controlled cgo callback bridge in the host process, or use the TypeScript / Python SDK for JSON callbacks.

## Verification

Source-tree checks:

```powershell
$env:CGO_ENABLED = "0"
go test ./...
```

Full native FFI checks need `CGO_ENABLED=1` and a cgo-compatible compiler. On Windows, Visual Studio alone is usually not enough for Go cgo; install a MinGW-w64/UCRT64 toolchain or another Go-compatible GCC distribution.

## Publishing

The release version is stored in `VERSION`. Go users consume SDK versions through Go module tags such as `v0.2.4`.

Before publishing:

```powershell
$env:CGO_ENABLED = "0"
go test ./...
```

Publish the SDK by pushing the matching Go module tag:

```powershell
git tag v0.2.4
git push origin v0.2.4
```

After the Go module tag is available, run the GitHub Actions workflow **Examples Release** manually. It reads `VERSION`, verifies `github.com/LuaSkills/luaskills-sdk-go@v{VERSION}`, installs LuaSkills runtime assets through the published TypeScript installer, runs the Go examples, then creates or updates the `examples-v{VERSION}` GitHub Release with:

- `luaskills-sdk-go-examples-{VERSION}.zip`
- `luaskills-sdk-go-examples-{VERSION}.zip.sha256`

The examples release tag intentionally uses the `examples-v` prefix so it does not interfere with Go module semver tags.
