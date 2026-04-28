# LuaSkills Go SDK Examples

These examples use the standalone Go module path and are intended to be copied into host applications.

这些示例使用独立 Go module 路径，适合复制到宿主应用中参考。

## Runtime Preparation

The Go SDK consumes the shared SDK runtime manifest but does not download assets itself. Prepare runtime assets with the TypeScript or Python installer:

Go SDK 会消费共享 SDK runtime manifest，但它本身不下载资产。请使用 TypeScript 或 Python 安装器准备 runtime 资产：

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root .\examples\fixture-runtime
```

or:

或者：

```powershell
luaskills install-runtime --database none --runtime-root .\examples\fixture-runtime
```

Running native FFI examples requires `CGO_ENABLED=1`, a cgo-compatible C compiler, and a discoverable LuaSkills dynamic library.

运行原生 FFI 示例需要 `CGO_ENABLED=1`、cgo 兼容 C 编译器，以及可被发现的 LuaSkills 动态库。

## Example Index

`basic` queries the JSON FFI version through `luaskills.Version`.

`basic` 通过 `luaskills.Version` 查询 JSON FFI 版本。

```powershell
go run .\examples\basic
```

`query` loads the bundled USER-layer fixture skill, lists delegated-visible entries, checks `IsSkill`, resolves `SkillNameForTool`, and reads help/completion surfaces.

`query` 会加载内置 USER 层夹具 skill，列出委托工具可见入口，检查 `IsSkill`，解析 `SkillNameForTool`，并读取 help/completion 查询面。

```powershell
go run .\examples\query
```

`call` demonstrates `CallSkill` and `RunLua` with an invocation context.

`call` 演示带调用上下文的 `CallSkill` 与 `RunLua`。

```powershell
go run .\examples\call
```

`lifecycle` demonstrates `Disable` and `Enable` through the ordinary Skills plane.

`lifecycle` 演示通过普通 Skills plane 执行 `Disable` 与 `Enable`。

```powershell
go run .\examples\lifecycle
```

`provider_callback` shows the Go callback API boundary. It currently returns `ErrProviderCallbacksRequireHostBridge` unless the host adds a controlled cgo callback bridge.

`provider_callback` 展示 Go callback API 边界。除非宿主添加受控 cgo callback bridge，否则当前会返回 `ErrProviderCallbacksRequireHostBridge`。

```powershell
go run .\examples\provider_callback
```

## Fixture Skill

The fixture skill is stored at `examples/fixture-runtime/user_skills/demo-standard-ffi-skill`. It intentionally lives in USER so delegated-query examples can see it without System authority.

夹具 skill 位于 `examples/fixture-runtime/user_skills/demo-standard-ffi-skill`。它故意放在 USER 层，这样委托查询示例不需要 System 权限也能看到它。
