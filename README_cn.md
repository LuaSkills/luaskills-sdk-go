# LuaSkills Go SDK

中文文档。英文默认文档见 [README.md](README.md)。

LuaSkills 主仓库：[LuaSkills/luaskills](https://github.com/LuaSkills/luaskills)

Go SDK，用于通过公共 JSON FFI 接入 LuaSkills 运行时。

SDK 封装了 cgo JSON FFI 调用、engine 生命周期、正式 skill root、带权限语义的管理调用、skill config、provider callback 边界、宿主工具 callback 边界与 runtime manifest 辅助能力。

## 安装

```bash
go get github.com/LuaSkills/luaskills-sdk-go
```

运行时调用需要 `CGO_ENABLED=1`、Go cgo 可用的 C 编译器，以及可被链接和加载的 LuaSkills 动态库。

Windows 示例：

```powershell
$env:CGO_ENABLED = "1"
$env:CGO_LDFLAGS = "-LD:\runtime\luaskills\libs"
$env:PATH = "D:\runtime\luaskills\libs;$env:PATH"
```

Linux / macOS 示例：

```bash
export CGO_ENABLED=1
export CGO_LDFLAGS="-L/opt/luaskills-runtime/libs"
export LD_LIBRARY_PATH="/opt/luaskills-runtime/libs:${LD_LIBRARY_PATH}"
```

## Runtime 资产

Go SDK 会规划并消费共享 SDK runtime manifest，但它本身不下载 release 资产。请使用 TypeScript 或 Python 安装器，或基于生成的 manifest 实现宿主自己的安装器。

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root D:\runtime\luaskills
```

```powershell
pip install luaskills-sdk
luaskills install-runtime --database vldb-direct --runtime-root D:\runtime\luaskills
```

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

`DefaultHostOptions(runtimeRoot)` 与 `NewClient` 会在 manifest 存在时自动读取 `runtimeRoot/resources/luaskills-sdk-runtime-manifest.json`，并合入 `host_options_patch`。

数据库模式：

- `RuntimeDatabaseNone`：安装 Lua runtime 归档与 LuaSkills FFI SDK 归档，但不安装数据库 provider。
- `RuntimeDatabaseVldbController`：通过 `space_controller` 模式使用 `vldb-controller` 可执行文件。
- `RuntimeDatabaseVldbDirect`：使用 `vldb-sqlite-lib` 与 `vldb-lancedb-lib` 动态库。
- `RuntimeDatabaseHostCallback`：由宿主提供 JSON callback。

## 基础用法

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

## 示例

详细源码示例位于 `examples/`。

```powershell
go run .\examples\basic
go run .\examples\call
go run .\examples\query
go run .\examples\lifecycle
go run .\examples\provider_callback
```

`provider_callback` 同时覆盖 JSON provider callback 与 `vulcan.host.*` 宿主工具 callback 边界。在宿主安装自有 cgo callback bridge 前，它会返回需要宿主桥接的错误。

query 与 lifecycle 示例使用内置夹具 skill：`examples/fixture-runtime/user_skills/demo-standard-ffi-skill`。请先使用 TypeScript 或 Python 安装器准备 runtime 资产：

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root .\examples\fixture-runtime
```

完整示例索引与 runtime 注意事项见 [examples/README_cn.md](examples/README_cn.md)。英文示例指南见 [examples/README.md](examples/README.md)。

## 权限与管理

查询类接口建议默认使用 `AuthorityDelegatedTool`，因此委托工具看不到 ROOT skills。

`AuthoritySystem` 只表示宿主可以管理 ROOT；它不表示可以绕过 ROOT 所有权或同名 `skill_id` 冲突规则。

`CallSkill` 与 `RunLua` 是运行时执行面，不作为 ROOT 可见性过滤。

skill config 是普通的 `skill_id + key` 配置存储面。配置只有在 Lua skill 主动读取时才会影响行为。

## JSON Provider Callback

Go SDK 暴露 callback API 边界，但默认不在包内安装进程级 cgo callback bridge。

```go
err := luaskills.SetSQLiteProviderJSONCallback(func(request any) (any, error) {
	return map[string]any{"ok": true, "request": request}, nil
})
```

当前该 API 会返回 `ErrProviderCallbacksRequireHostBridge`。需要 `host_callback + json` 的正式 Go 宿主，应在宿主进程内实现受控 cgo callback bridge，或先使用 TypeScript / Python SDK 接 JSON callback。

## 宿主工具 Callback

`vulcan.host.*` 使用通过 `luaskills_ffi_set_host_tool_json_callback` 注册的固定宿主工具 callback。Go SDK 暴露类型化请求结构与注册边界：

```go
// Register the host-tool callback boundary in hosts that provide a cgo bridge.
// 在提供 cgo 桥的宿主中注册宿主工具 callback 边界。
err := luaskills.SetHostToolJSONCallback(func(request luaskills.HostToolJSONRequest) (any, error) {
	return map[string]any{"ok": true, "value": request.Args}, nil
})
```

当前该 API 会返回 `ErrHostToolCallbacksRequireHostBridge`。如果正式 Go 宿主希望让 Lua skill 调用宿主工具，应在宿主进程内实现受控 cgo bridge。callback 请求包含 `action`、`tool_name` 与 `args`；`list` 返回工具元数据，`has` 返回可用性，`call` 返回一次完整的 table 形态结果，不走 stream。

## 模型 Callback

`vulcan.models.*` 使用通过 `luaskills_ffi_set_model_embed_json_callback` 与 `luaskills_ffi_set_model_llm_json_callback` 注册的固定 callback。Go SDK 暴露类型化请求、响应和错误结构，但进程级 cgo callback bridge 仍由宿主实现：

可使用这些类型保持正式 bridge 契约清晰：

- `ModelEmbedJSONRequest`：接收 `Text` 与 `Caller`。
- `ModelLLMJSONRequest`：接收 `System`、`User` 与 `Caller`。
- `ModelEmbedJSONResponse`：返回 `Vector`、`Dimensions` 与可选 `Usage`。
- `ModelLLMJSONResponse`：返回 `Assistant` 与可选 `Usage`。
- `ModelJSONErrorEnvelope`：保留模型错误与可选 provider 诊断字段。

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

provider 失败且需要让 Lua 侧拿到诊断信息时，应返回结构化错误包络：

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

当前 `SetModelEmbedJSONCallback` 与 `SetModelLLMJSONCallback` 会返回 `ErrModelCallbacksRequireHostBridge`。正式 Go 宿主应在自己的进程中实现受控 cgo callback bridge，转发 embedding 的 `{ text, caller }` 和 LLM 的 `{ system, user, caller }`，并返回裸成功载荷或 `ModelJSONErrorEnvelope`。Lua 侧不会接触或覆盖模型配置。

Go 宿主检查清单：

- 模型 provider 配置放在宿主配置中，不放进 Lua skill config。
- 填充 provider 错误字段前先脱敏 API key、Authorization header、签名和请求头。
- 使用 `Caller` 做成本归因、限流、审计和按 skill 策略控制。
- Go bridge 抛出的错误应视为内部桥接失败；需要透传给 Lua 的 provider 失败应使用 `ModelJSONErrorEnvelope`。

## 验证

源码环境检查：

```powershell
$env:CGO_ENABLED = "0"
go test ./...
```

完整原生 FFI 检查需要 `CGO_ENABLED=1` 与 cgo 兼容的 C 编译器。在 Windows 上，仅有 Visual Studio 通常不足以满足 Go cgo；请安装 MinGW-w64/UCRT64 工具链或其他 Go 兼容的 GCC 发行版。

## 发布

发布版本记录在 `VERSION`。Go 用户通过 `v0.2.6` 这类 Go module tag 消费 SDK 版本。

发布前执行：

```powershell
$env:CGO_ENABLED = "0"
go test ./...
```

推送匹配的 Go module tag 即完成 SDK 发布：

```powershell
git tag v0.2.6
git push origin v0.2.6
```

Go module tag 可用后，手动运行 GitHub Actions 里的 **Examples Release** 工作流。它会读取 `VERSION`，校验 `github.com/LuaSkills/luaskills-sdk-go@v{VERSION}`，通过已发布 TypeScript 安装器安装 LuaSkills runtime 资产，运行 Go 示例冒烟测试，然后创建或更新 `examples-v{VERSION}` GitHub Release，并上传：

- `luaskills-sdk-go-examples-{VERSION}.zip`
- `luaskills-sdk-go-examples-{VERSION}.zip.sha256`

示例 release tag 故意使用 `examples-v` 前缀，避免干扰 Go module 的语义版本 tag。
