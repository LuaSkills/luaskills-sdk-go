# LuaSkills Go SDK 示例

中文示例文档。英文默认文档见 [README.md](README.md)。

LuaSkills 主仓库：[LuaSkills/luaskills](https://github.com/LuaSkills/luaskills)

这些示例使用独立 Go module 路径，适合复制到宿主应用中参考。

## Runtime 准备

Go SDK 会消费共享 SDK runtime manifest，但它本身不下载资产。请使用 TypeScript 或 Python 安装器准备 runtime 资产：

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root .\examples\fixture-runtime
```

或者：

```powershell
luaskills install-runtime --database none --runtime-root .\examples\fixture-runtime
```

运行原生 FFI 示例需要 `CGO_ENABLED=1`、cgo 兼容 C 编译器，以及可被发现的 LuaSkills 动态库。

## 示例索引

`embedded_lifecycle` 是开发用类型化客户端／回调泵／运行时作用域完整示例，已从正式独立示例包排除。须从匹配 SDK 源码检出取得，再通过[候选门禁](../docs/embedded-validation.md) 使用准确本地候选库验收，可加 `--race`。示例不下载资产，须匹配开发库。

`basic` 通过 `luaskills.Version` 查询 JSON FFI 版本。

```powershell
go run .\examples\basic
```

`query` 会加载内置 USER 层夹具 skill，列出委托工具可见入口，检查 `IsSkill`，解析 `SkillNameForTool`，并读取 help/completion 查询面。

```powershell
go run .\examples\query
```

`call` 演示带调用上下文的 `CallSkill` 与 `RunLua`。

```powershell
go run .\examples\call
```

`lifecycle` 演示通过普通 Skills plane 执行 `Disable` 与 `Enable`。

```powershell
go run .\examples\lifecycle
```

`runtime_lease` 演示一个持久运行时租约、绑定 authority 的 system 查询，以及复用同一个交互式子进程句柄的连续 `Eval` 调用。

```powershell
go run .\examples\runtime_lease
```

`provider_callback` 展示 Go provider 与宿主工具 callback API 边界。除非宿主添加受控 cgo callback bridge，否则当前会返回需要宿主桥接的错误。

```powershell
go run .\examples\provider_callback
```

模型 callback 对接见主 [SDK README](../README_cn.md#模型-callback)。Go SDK 暴露类型化模型请求、响应和错误结构；真实 callback 注册仍需要宿主拥有的 cgo bridge。

## Fixture Skill

夹具 skill 位于 `examples/fixture-runtime/user_skills/demo-standard-ffi-skill`。它故意放在 USER 层，这样委托查询示例不需要 System 权限也能看到它。

## 示例发布包

仓库工作流 **Examples Release** 先认证显式原 SDK 候选与成功完成轮次、两条永久签名链及新的冷消费者，再生成 `luaskills-sdk-go-examples-{VERSION}.zip`。它使用独立已验 TypeScript 精确版本、实际核心库字节及新缓存中的公共 Go 模块运行六例。确定性 ZIP 与原 sidecar 在发布前独立签名；恢复仅复用指定原签名制品。

从[示例发布列表](https://github.com/LuaSkills/luaskills-sdk-go/releases?q=examples-v&expanded=true)中匹配的 `examples-v{VERSION}` 条目下载。工作流先向 draft 上传全部资产，再发布此独立 Release；已有 final 资产不覆盖、不追加。
