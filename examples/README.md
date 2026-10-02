# LuaSkills Go SDK Examples

English documentation is the default example documentation. For Chinese, see [README_cn.md](README_cn.md).

Main LuaSkills repository: [LuaSkills/luaskills](https://github.com/LuaSkills/luaskills)

These examples use the standalone Go module path and are intended to be copied into host applications.

## Runtime Preparation

The Go SDK consumes the shared SDK runtime manifest but does not download assets itself. Prepare runtime assets with the TypeScript or Python installer:

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root .\examples\fixture-runtime
```

or:

```powershell
luaskills install-runtime --database none --runtime-root .\examples\fixture-runtime
```

Running native FFI examples requires `CGO_ENABLED=1`, a cgo-compatible C compiler, and a discoverable LuaSkills dynamic library.

## Example Index

`embedded_lifecycle` is the development-only typed client/callback pump/runtime scope example, excluded from the standalone published examples package. Obtain it from a matching SDK source checkout and run it with [the candidate gate](../docs/embedded-validation.md), including `--race` when needed. It performs no downloads and requires the matching development library.

`basic` queries the JSON FFI version through `luaskills.Version`.

```powershell
go run .\examples\basic
```

`query` loads the bundled USER-layer fixture skill, lists delegated-visible entries, checks `IsSkill`, resolves `SkillNameForTool`, and reads help/completion surfaces.

```powershell
go run .\examples\query
```

`call` demonstrates `CallSkill` and `RunLua` with an invocation context.

```powershell
go run .\examples\call
```

`lifecycle` demonstrates `Disable` and `Enable` through the ordinary Skills plane.

```powershell
go run .\examples\lifecycle
```

`runtime_lease` demonstrates one persistent runtime lease, authority-bound system queries, and repeated `Eval` calls that reuse one interactive child-process handle.

```powershell
go run .\examples\runtime_lease
```

`provider_callback` shows the Go provider and host-tool callback API boundaries. They currently return bridge-required errors unless the host adds a controlled cgo callback bridge.

```powershell
go run .\examples\provider_callback
```

Model callback integration is documented in the main [SDK README](../README.md#model-callback). The Go SDK exposes typed model request, response, and error structures, while real callback registration still requires a host-owned cgo bridge.

## Fixture Skill

The fixture skill is stored at `examples/fixture-runtime/user_skills/demo-standard-ffi-skill`. It intentionally lives in USER so delegated-query examples can see it without System authority.

## Release Package

The repository workflow **Examples Release** authenticates the explicit original SDK candidate and successful completion attempts, both permanent signatures and new cold consumers before creating `luaskills-sdk-go-examples-{VERSION}.zip`. It uses the independently verified TypeScript version, exact core library and public Go module in a new cache to run six examples. The deterministic ZIP and original sidecar are independently signed before publication; recovery reuses only that exact signed artifact.

Download from the matching `examples-v{VERSION}` entry in [Examples releases](https://github.com/LuaSkills/luaskills-sdk-go/releases?q=examples-v&expanded=true). The workflow uploads all assets to a draft before publishing this separate Release; existing final assets are never overwritten or extended.
