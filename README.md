# LuaSkills Go SDK

English documentation is the default package documentation. For Chinese, see [README_cn.md](README_cn.md).

Main LuaSkills repository: [LuaSkills/luaskills](https://github.com/LuaSkills/luaskills)

Go SDK for integrating the LuaSkills runtime through the public JSON FFI surface.

The SDK wraps cgo JSON FFI calls, engine lifecycle, formal skill roots, authority-aware management calls, skill config, provider callback boundaries, host-tool callback boundaries, and runtime manifest helpers.

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

The repository includes a unified synchronization script that directly downloads LuaSkills FFI, Lua runtime packages, and VLDB without requiring the Python or TypeScript installer:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/deps/sync_runtime_assets.ps1 -Target all -Database vldb-controller -RuntimeRoot D:\runtime\luaskills
```

```bash
RUNTIME_ROOT=/opt/luaskills scripts/deps/sync_runtime_assets.sh all vldb-controller
```

Supported targets are `all`, `luaskills`, `lua`, and `vldb`. VLDB presets are `none`, `vldb-controller`, `vldb-direct`, and `host-callback`. The scripts pin LuaSkills to `v0.5.0` by default and accept explicit release-version overrides.

The Go SDK plans and consumes the shared SDK runtime manifest, while the repository scripts above directly download release assets. The shared manifest now points at:

- `lua-runtime-packages-{platform}.tar.gz` from `LuaSkills/luaskills-packages`
- `luaskills-ffi-sdk-{platform}.tar.gz` from `LuaSkills/luaskills`
- optional managed Python, `uv`, Node.js, and `pnpm` paths under `runtime_root/dependencies/runtimes/...`

Managed child runtimes support Windows x64, Linux x64/ARM64, and macOS x64/ARM64. Windows ARM is explicitly rejected before any download or target-directory creation. The repository also ships standalone fetch and layout-validation tools for hosts that prepare debug runtimes without a Python or TypeScript installer:

The current exact managed dependency versions are Python `3.12.7`, uv `0.11.17`, Node.js `22.11.0`, and pnpm `9.15.0`. Package `dependencies.yaml` files must declare the same exact runtime and package-manager versions unless the host deliberately installs another supported version.

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/deps/fetch_managed_runtimes.ps1 -RuntimeRoot D:\runtime\luaskills -Target all
python scripts/debug-tools/managed_runtime_layout_check.py D:\runtime\luaskills
```

```bash
RUNTIME_ROOT=/opt/luaskills scripts/deps/fetch_managed_runtimes.sh all
python3 scripts/debug-tools/managed_runtime_layout_check.py /opt/luaskills
```

By default, the shared manifest keeps LuaSkills core aligned with the SDK release and resolves runtime packages from the compatible `0.1` series by selecting the newest published patch automatically.

## Version Alignment

- Keep the SDK and LuaSkills core on the same current release line whenever possible.
- The current SDK defaults to LuaSkills core tag `v0.5.0`.
- Runtime packages and native dependencies still come from the split `LuaSkills/luaskills-packages` and related release assets.
- SDK default host options now pass only `runtime_root`; LuaSkills derives `bin`, `libs`, `lua_packages`, `resources`, `skills`, `temp`, `dependencies`, `state`, `databases`, `config`, and `system_lua_lib`.
- Host tools live directly under `runtime_root/bin`, not `runtime_root/bin/tools`.

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root D:\runtime\luaskills
npx @luaskills/sdk install-runtime --database none --managed-runtimes all --runtime-root D:\runtime\luaskills
```

```powershell
pip install luaskills-sdk
luaskills install-runtime --database vldb-direct --runtime-root D:\runtime\luaskills
```

Go hosts can inspect the same asset plan:

```go
manifest, err := luaskills.BuildRuntimeInstallManifest(luaskills.RuntimeInstallOptions{
	RuntimeRoot:      "D:/runtime/luaskills",
	Database:         luaskills.RuntimeDatabaseVldbDirect,
	SkipLuaRuntime:   false,
	ManagedRuntimes: luaskills.ManagedRuntimeAll,
})
if err != nil {
	panic(err)
}

hostOptions, err := luaskills.HostOptionsFromRuntimeManifest(manifest)
if err != nil {
	panic(err)
}
```

`DefaultHostOptions(runtimeRoot)` returns host options plus an error. `DefaultHostOptions` and `NewClient` automatically read `runtimeRoot/resources/luaskills-sdk-runtime-manifest.json` and merge `host_options_patch` when the manifest exists. A missing manifest keeps SDK base defaults; a malformed manifest or host path that escapes `runtimeRoot` fails with a path-aware error.

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
go run .\examples\runtime_lease
go run .\examples\provider_callback
```

`provider_callback` covers both JSON provider callbacks and the `vulcan.host.*` host-tool callback boundary. It returns bridge-required errors until a host-owned cgo callback bridge is installed.

The query, lifecycle, and persistent runtime-lease examples use the bundled fixture skill at `examples/fixture-runtime/user_skills/demo-standard-ffi-skill`. Prepare runtime assets with a TypeScript or Python installer first:

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root .\examples\fixture-runtime
```

See [examples/README.md](examples/README.md) for the full example index and runtime notes. The Chinese example guide is [examples/README_cn.md](examples/README_cn.md).

## Persistent Runtime Leases

Use `client.RuntimeLeases()` for the public lease endpoints, or `client.System(authority).RuntimeLeases()` when the host wants fixed authority injection through the dedicated system runtime-lease exports provided by the latest native library.

```go
client, err := luaskills.NewClient(luaskills.ClientOptions{RuntimeRoot: "D:/runtime/luaskills"})
if err != nil {
	log.Fatal(err)
}
defer client.Close()

leases := client.System(luaskills.AuthoritySystem).RuntimeLeases()
cwd := "D:/runtime/luaskills/system_lua_lib"
ttlSec := 600
session, err := leases.CreateHandleWithOptions("demo-session", true, &luaskills.RuntimeLeaseCreateOptions{
	TTLSec: &ttlSec,
	CWD:    &cwd,
	Mounts: map[string]any{"channel": "demo"},
	SystemPackage: &luaskills.SystemRuntimePackage{ID: "debug-plugin", Root: "D:/runtime/luaskills/system_lua_lib/debug-plugin", DependenciesFile: "dependencies.json"},
})
if err != nil {
	log.Fatal(err)
}

result, err := session.Eval("counter = (counter or 0) + 1; return { counter = counter }", nil, 60000)
if err != nil {
	log.Fatal(err)
}

fmt.Println(result["result"])
```

## Migration Notes

- Existing `client.System(authority)` lifecycle calls keep working; the returned wrapper now also exposes query helpers and `RuntimeLeases()`.
- `RuntimeLeaseHandle` persists `lease_id + sid + generation` and automatically reattaches identity guards on `Eval`, `Status`, and `Close`.
- Authority-bound runtime-lease helpers dispatch directly to dedicated `luaskills_ffi_system_runtime_lease_*` entrypoints.
- `CallSkill` now returns optional `HostResult` when the host enables `request_context.client_capabilities.host_result` and one Lua tool emits a fourth structured return value.
- When `HostResult.Kind == "change_set"`, hosts should decode `HostResult.Payload` into `RuntimeChangeSetPayload`.
- Canonical `change_set` payloads now use file lifecycle records plus hunk-level `before + delete[] + insert[] + after` blocks for `modify` changes.
- `create` and `delete` file records carry full-file `content`, while `rename` records carry `old_path` and `new_path`.
- Public leases accept `cwd`, `workspace_root`, `lua_roots`, `c_roots`, and `mounts`. System leases require `SystemPackage`, reject `lua_roots/c_roots`, and derive roots from the trusted package manifest.
- `PollManagedSessionEvents`, `WaitManagedSessionEvents`, and `SetManagedSessionWakeCallback` expose the 0.5.0 managed-session event surface.
- Go hosts should deploy the matching latest LuaSkills native library when using these APIs, because cgo links directly against the current exported symbol set.

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

## Host Tool Callback

`vulcan.host.*` uses the fixed host-tool callback registered through `luaskills_ffi_set_host_tool_json_callback`. The Go SDK exposes the typed request shape and registration boundary:

```go
// Register the host-tool callback boundary in hosts that provide a cgo bridge.
// 在提供 cgo 桥的宿主中注册宿主工具 callback 边界。
err := luaskills.SetHostToolJSONCallback(func(request luaskills.HostToolJSONRequest) (any, error) {
	return map[string]any{"ok": true, "value": request.Args}, nil
})
```

Currently this returns `ErrHostToolCallbacksRequireHostBridge`. A production Go host that wants Lua skills to call host tools should implement the controlled cgo bridge in the host process. The callback request contains `action`, `tool_name`, and `args`; `list` returns metadata, `has` returns availability, and `call` returns one complete table-shaped result without streaming.

## Model Callback

`vulcan.models.*` uses fixed callbacks registered through `luaskills_ffi_set_model_embed_json_callback` and `luaskills_ffi_set_model_llm_json_callback`. The Go SDK exposes typed request, response, and error shapes, but still leaves the process-level cgo callback bridge to the host:

Use these types to keep the production bridge contract explicit:

- `ModelEmbedJSONRequest`: receives `Text` and `Caller`.
- `ModelLLMJSONRequest`: receives `System`, `User`, and `Caller`.
- `ModelEmbedJSONResponse`: returns `Vector`, `Dimensions`, and optional `Usage`.
- `ModelLLMJSONResponse`: returns `Assistant` and optional `Usage`.
- `ModelJSONErrorEnvelope`: preserves model errors and optional provider details.

```go
// Register the model callback boundary in hosts that provide a cgo bridge.
// 在提供 cgo 桥的宿主中注册模型 callback 边界。
err := luaskills.SetModelEmbedJSONCallback(func(request luaskills.ModelEmbedJSONRequest) (any, error) {
	return luaskills.ModelEmbedJSONResponse{
		Vector:     []float32{0.1, 0.2, 0.3},
		Dimensions: 3,
	}, nil
})
```

Provider failures should be returned as structured envelopes when callers need diagnostics:

```go
func ptr[T any](value T) *T {
	return &value
}

failure := luaskills.ModelJSONErrorEnvelope{
	OK: false,
	Error: luaskills.ModelJSONError{
		Code:            luaskills.ModelJSONErrorProviderError,
		Message:         "model provider rejected the request",
		ProviderMessage: ptr("raw provider message after host-side redaction"),
		ProviderCode:    ptr("model_not_found"),
		ProviderStatus:  ptr(uint16(404)),
	},
}
```

Currently `SetModelEmbedJSONCallback` and `SetModelLLMJSONCallback` return `ErrModelCallbacksRequireHostBridge`. A production Go host should implement the controlled cgo callback bridge in its own process, forward `{ text, caller }` for embeddings and `{ system, user, caller }` for LLM calls, and return either a bare success payload or `ModelJSONErrorEnvelope`. Lua does not receive or override model configuration.

Go host checklist:

- Keep provider settings in host configuration, not in Lua skill config.
- Redact API keys, Authorization headers, signatures, and request headers before filling provider error fields.
- Use `Caller` for cost attribution, rate limits, audit logs, and per-skill policy.
- Treat thrown Go errors from the bridge as internal bridge failures; use `ModelJSONErrorEnvelope` for provider failures that should reach Lua.

## Verification

Source-tree checks:

```powershell
$env:CGO_ENABLED = "0"
go test ./...
```

Full native FFI checks need `CGO_ENABLED=1` and a cgo-compatible compiler. On Windows, Visual Studio alone is usually not enough for Go cgo; install a MinGW-w64/UCRT64 toolchain or another Go-compatible GCC distribution.

## Publishing

The release version is stored in `VERSION`. Go users consume SDK versions through Go module tags such as `v0.5.0`.

For one unified ecosystem release, publish `LuaSkills/luaskills-packages` first, then publish `LuaSkills/luaskills`, and publish the TypeScript SDK before the Go examples release flow because the Go examples workflow installs runtime assets through the published TypeScript package.

Before publishing:

```powershell
$env:CGO_ENABLED = "0"
go test ./...
```

Publish the SDK by pushing the matching Go module tag:

```powershell
git tag v0.5.0
git push origin v0.5.0
```

After the Go module tag is available, run the GitHub Actions workflow **Examples Release** manually. It reads `VERSION`, verifies `github.com/LuaSkills/luaskills-sdk-go@v{VERSION}`, installs LuaSkills runtime assets through the published TypeScript installer, runs the Go examples, then creates or updates the `examples-v{VERSION}` GitHub Release with:

- `luaskills-sdk-go-examples-{VERSION}.zip`
- `luaskills-sdk-go-examples-{VERSION}.zip.sha256`

The examples release tag intentionally uses the `examples-v` prefix so it does not interfere with Go module semver tags.

Recommended unified publish order: `luaskills-packages` -> `luaskills` core release -> TypeScript SDK -> Python SDK -> Go SDK -> SDK examples releases.
