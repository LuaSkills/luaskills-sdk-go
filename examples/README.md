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

`provider_callback` shows the Go callback API boundary. It currently returns `ErrProviderCallbacksRequireHostBridge` unless the host adds a controlled cgo callback bridge.

```powershell
go run .\examples\provider_callback
```

## Fixture Skill

The fixture skill is stored at `examples/fixture-runtime/user_skills/demo-standard-ffi-skill`. It intentionally lives in USER so delegated-query examples can see it without System authority.

## Release Package

The repository workflow **Examples Release** creates `luaskills-sdk-go-examples-{VERSION}.zip` after the matching Go module tag is available. The workflow verifies `github.com/LuaSkills/luaskills-sdk-go@v{VERSION}`, installs LuaSkills runtime assets through the published TypeScript installer, and runs the examples before uploading the asset.

The release tag is `examples-v{VERSION}` so example assets do not interfere with Go module semver tags.
