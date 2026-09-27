# LuaSkills Go SDK

中文文档。英文默认文档见 [README.md](README.md)。

LuaSkills 主仓库：[LuaSkills/luaskills](https://github.com/LuaSkills/luaskills)

Go SDK，用于通过公共 JSON FFI 接入 LuaSkills 运行时。

`0.5.7` 是当前发布版本。它沿用严格的技能包级配置契约，并将运行时资产默认值设为 LuaSkills core `v0.5.7`、vldb-controller `v0.2.3` 与 vldb-sqlite `v0.1.6`。

SDK 封装了 cgo JSON FFI 调用、engine 生命周期、正式 skill root、带权限语义的管理调用、skill config、provider callback 边界、宿主工具 callback 边界与 runtime manifest 辅助能力。

## 嵌入式运行时开发接口

当前开发源码新增 `NewEmbeddedTransport`，使用独立版本一 C ABI。必须链接包含这些新导出的匹配开发核心；本文不表示已发布的 `0.5.7` 动态库支持它们。正式版本和默认资产将随整个生态验收统一提升。`CGO_ENABLED=0` 仍可使用契约和编码器，原生构造明确返回不支持错误。

传输配置显式声明运行时数量、响应数量、聚合响应字节、单响应字节和请求字节上限，创建后不可变。`Request(map[string]any)` 同步冻结命令并返回真实交付；Go 整数保留全部 64 位，Go 浮点数保留小数／指数标记。解码数值为 `json.Number`，不经过 float64 舍入。输入支持字符串键映射、类型化切片／数组、普通标量及带显式 JSON 标签的结构体；标签仅支持名称、`omitempty` 和排除，不执行自定义序列化钩子，不隐式转成 base64。重复解码键、非法 Unicode、循环、额外信封字段及数值溢出被拒绝。

`EmbeddedTransportError` 与 `EmbeddedRuntimeError` 分别表示 ABI 错误和已交付业务拒绝。`EmbeddedResultReleaseError` 保留复制响应；用 `ResponseBytes()` 获取独立副本，或用 `DeliveredResult()` 读取原结果。调用可能已经执行，不能因为释放失败重放变更。`ReleaseResults()` 只恢复实际保留的缓冲；活动读取者存在时明确拒绝，不能与复制竞争。

`Close()` 请求关闭入场；仍须关闭并移除实际运行时、释放缓冲，最后 `Free()` 才能成功。原生调用期间不持有全局 Go 锁，控制调用可并发推进；GC 不替代关闭，`LiveEmbeddedTransports()` 保留可发现所有者。此同步低层接口不提供观察取消。

`NewEmbeddedCommandDriver(transport, config)` 在借用传输上提供固定业务工作位和一个独立控制工作位。显式配置正数 `WorkWorkers`、`MaxWorkCommands` 与 `MaxControlCommands`；配额包含已完成但尚未 `Forget()` 的回执。构造要求没有活动请求、保留原生结果或既有驱动，且响应数量和累计字节足够容纳每个工作位的最大响应。这项预留覆盖驱动工作位；并发直接使用底层传输的占用需要调用方另行计算。Lua VM 并发仍由核心管理。

`Submit(ctx, 生成命令)` 冻结并校验命令，按路由选择通道，在执行前发布 `EmbeddedCommand`。上下文只控制入场前取消。`Result(ctx)` 仅取消该观察者的等待；已接纳工作及回执仍被拥有，可通过 `Commands()` 查找。每个观察者获得独立结果、响应字节及保留公开错误副本；可使用对应生成响应解码器读取 `ResponseBytes()`。驱动拒绝原生 `operation_wait`，应轮询 `operation_status`，通过独立 `operation_cancel` 请求真实取消。

原生结果释放失败时，所属工作位暂停并保留帧预留，回执返回 `EmbeddedResultReleaseError`。`driver.ReleaseResults(ctx)` 不占回执配额，仅重试保留缓冲释放，不重放命令；活动读取者存在时返回忙错误。恢复接纳后同步执行，上下文取消不能中断。恢复成功唤醒暂停工作位，原回执仍保留释放失败记录，`DeliveredResult()` 可读取原始交付。驱动拥有传输期间应使用驱动级恢复，以通知暂停工作位。

`RequestClose()` 封闭新入场并排空已接纳队列。`Close(ctx)` 即使观察者已取消也会启动同样的排空，并等待真实工作位退出；超时不会销毁预先启动的关闭协调器。释放失败须显式恢复，排空才能继续。关闭驱动不关闭借用的原生运行时或传输；应在驱动关闭前通过控制命令关闭原生运行时，或在之后通过底层传输清理，再单独关闭和释放传输。`LiveEmbeddedCommandDrivers()` 保留仍有回执的已关闭驱动，直到全部已完成回执显式遗忘。工作位或恢复意外退出会返回 `EmbeddedDriverFailure`，保留不确定所有权，不宣称关闭成功、不授权重放；不确定命令绝不报告 `Done()`。

`NewEmbeddedCallbackPump(transport, runtimeID, config)` 现已提供自动 Go 队列回调。显式设置正数 `MaxConcurrentHandlers`、`MaxPendingCommands` 及 `PollIntervalMS`，注册前用 `Ready(ctx)` 观察实际启动。每个泵拥有一个顺序原生协调器及固定处理器协程；已返回但仍等待完成确认的处理器继续占额。驱动与泵共同核算响应预留。应在传输无活动调用及保留结果时创建所有者；每个精确运行时仅允许一个泵，该运行时全部队列注册均应经此泵管理。

注册内容为包含生成队列描述符及 `EmbeddedHostHandler` 的 `EmbeddedHostCapability`。`Register(ctx, capabilities)` 在原生发布前冻结描述符并保留精确处理器；取消仅分离观察者，已接纳注册可通过 `Status()` 查找。处理器分开接收应用参数与 `EmbeddedHostCallbackContext`；后者实现 `context.Context`，提供可信调用方字段副本及首次核心取消。`RemainingMS()` 仅作参考，不伪造本地截止时间。变更型处理器的副作用初始为未知，应通过 `ReportEffects` 报告实际证据，处理器返回后封存。panic、`runtime.Goexit`、无效结果和任意错误转换为有界通用失败；有效 `EmbeddedRuntimeError` 的显式消息视为宿主授权披露。Lua 的 `vulcan.capabilities.call` 接收包含 `ok/value/error/effects` 的信封；正常返回该信封时，回调失败不自动变成外层 Lua 操作失败。

下游调用应传递提供的回调上下文或其派生上下文。受控回调依赖实现前，驱动提交及 SDK 等待拒绝该标记。主动丢弃上下文或直接调用非托管同步传输会绕过此护栏，它不是协程沙箱；非阻塞诊断及关闭请求仍可使用。

`Unregister(ctx, registrationID)` 等待实际处理器返回、原生排空及元数据移除。`Close(ctx)` 退役全部已接纳注册并汇合实际处理器工作位，不强制终止应用代码。错误封闭新入场并保留所有权；`RetryAcknowledgements(ctx)` 显式释放保留缓冲，并根据精确请求、注册及操作证据核对失败确认，绝不重新运行处理器。仅请求解析器的 `INVALID_ARGUMENT` 已证明未分发时，允许一次保留原副作用证据的失败确认替换。注册、领取或退役响应丢失时保留原始所有权，不重放、不假定成功。`LiveEmbeddedCallbackPumps()` 保持可发现，传输 `Free()` 拒绝仍拥有的泵。恢复用于排空失败泵，不重新开放入场。下文运行时作用域负责有序关闭。

`NewEmbeddedClient(driver)` 借用现有驱动器，提供 `EmbeddedRuntime`、`EmbeddedPlugin`、`EmbeddedPool`、`EmbeddedSession` 及 `EmbeddedOperation`。命令方法接受入场上下文，返回 `(*EmbeddedPending[T], error)`；`pending.Result(ctx)` 等待并按生成结果类型校验，`pending.DeliveredResult()` 只投影原始交付，可恢复释放失败前已经创建的句柄，不重放变更。`pending.Receipt()` 返回原驱动回执，`pending.Forget()` 仅归还 SDK 配额；句柄的 `Forget(ctx)` 或运行时 `Free(ctx)` 才操作核心记录。投影失败保留回执及原始字节，结果每次重新解码，不共享可变响应。

用 `client.Reserve(ctx)` 取得实际槽身份，再以生成的 `EmbeddedInputLuaEngineOptions` 和 `EmbeddedInputEmbeddedRuntimeConfig` 调用 `runtime.Initialize(ctx, options, budgets)`。`RuntimeID()` 是 FFI 槽身份，状态里的 `CoreRuntimeId` 是另一层命名空间，不能混用。`client.Runtime(id)` 以及运行时上的 `Plugin/Pool/Session/Operation(id)` 可绑定已知非空精确身份，不探测或重新选择对象。配置由宿主显式提供；新的类型化入口不改变旧 `CreateEngineOptions` 的映射返回类型。

池使用 `RegisterPool(ctx, definition, policy, permissions, executionRevision)` 注册不可变执行域。**先注册所需宿主能力，再注册池**：核心注册池时捕获能力快照，之后发布的新注册不会进入原池；需要使用新注册时创建新执行域并排空旧域。公共／专用、可复用／单次／固定会话模式由生成策略显式声明。`pool.Submit(ctx, export, arguments, invocation, timeoutMS)` 返回独立操作；`pool.OpenSession(ctx, timeoutMS)` 返回 `EmbeddedSessionOpen{Session, Initialization}`，必须单独查询初始化操作结果，之后由固定会话 `Submit` 维持模块状态。

`operation.Wait(ctx)` 使用 `EmbeddedDefaultPollInterval`，或以 `WaitInterval(ctx, 正 time.Duration)` 指定间隔。它通过短状态命令轮询，不占用阻塞原生等待工作位；仅 `succeeded/failed/cancelled` 属于终态，清理或取消意图不是完成。终态失败及取消作为包含副作用证据的快照返回；Go 错误表示入场、交付、投影或观察失败。成功只读轮询自动遗忘 SDK 回执，中断或失败的状态回执保留在 `client.Driver().Commands()`，调用方须观察并显式遗忘后归还配额。上下文超时不发送原生取消；用 `operation.Cancel(ctx)` 请求协作取消，并继续查询实际结束与迟到副作用。

`NewEmbeddedRuntimeScope(runtime, pump, interval)` 现已集中管理原生关闭、回调排空、真实核心关闭及最终槽释放；没有回调泵时传入 `nil`，默认间隔使用 `EmbeddedDefaultPollInterval`。作用域拥有一个预先启动的控制协程及独立最坏响应预留，与普通驱动和泵共同核算数量／字节预算。即使普通驱动回执配额已满或驱动已关闭，作用域仍可独立清理；它借用驱动及传输，不会替其他运行时关闭这些共享对象。

接管前须完成当前驱动命令并等待传输无活动调用或保留结果；已有泵的短轮询可能暂时返回忙错误。先创建泵，再把精确同一个实例传给作用域；不允许遗漏现有泵、跨传输或身份接管、重复作用域，或接管后再添加泵。清理帧在接管前按请求预算校验。驱动命令入场与作用域接管共享锁序，排队但尚未执行的释放也会阻止接管；接管之后，类型化 `Free` 和直接驱动 `runtime_free` 都被拒绝。同步底层传输属于非托管入口，调用方须自行遵守所有权边界。请通过构造器创建作用域且不要复制；GC 不替代清理。

`scope.RequestClose()` 非阻塞启动清理；`scope.Close(ctx)` 即使观察上下文已经取消也启动同一条清理流程，随后等待实际完成和协调器退出。空上下文或回调标记会在启动前被拒绝。观察者超时／取消不释放处理器、运行时、响应预留，也不停止清理；可用 `Status()` 和 `LiveEmbeddedRuntimeScopes()` 查询保留所有者。关闭调用重复观察同一尝试；失败不会触发隐式重试。

`scope.RetryClose(ctx)` 显式恢复保留分配、已接管泵的失败确认或已证明尚未进入变更的控制拒绝。原始成功回执即使释放失败，也先推进检查点；恢复后不会再次关闭或移除同一槽。原生请求入口容量拒绝和传输本地入口忙错误有明确未执行证据；只有核心释放命令的业务 `busy` 会因租约尚存而自动继续轮询。复制交付缺失、类型或身份错误、构造 `faulted`、协调器 panic／`Goexit` 均保留不确定所有权，不伪造成功、不授权重放。回调协调器已退出时不提供普通确认恢复。共享传输上的其他驱动或泵如有自己的暂停状态，仍须通过各自的恢复 API 继续。

作用域成功关闭后再处理普通驱动回执、关闭驱动，并在全部作用域及泵排空后关闭和释放共享传输。未使用作用域的类型句柄仍由调用方显式依次请求原生关闭、排空泵、查询真实关闭、释放槽；每步交付失败须保留原始证据。

包内契约还生成独立的 `EmbeddedInput*` 与 `EmbeddedOutput*` 类型、枚举常量、封闭命令分支及全部已声明响应解码器。`EncodeEmbeddedRequest` 冻结并校验类型化信封；调用方显式填写 `EmbeddedProtocolVersion` 和生成的命令判别值。`DecodeEmbeddedOutput*Response` 校验精确字段名、必需字段、枚举、整数位宽和已声明集合约束，不调用自定义序列化钩子。这是线结构校验；宿主完成的成功语义及运行时预算等业务规则，继续以核心校验为准。

可选字段使用外层指针表示存在性，可空字段使用内层指针；`**T` 因而区分缺失、存在且为空和存在且有值，`*any` 保留成功空值与结果缺失的区别。请使用生成响应解码器保留这些状态，普通 `encoding/json.Unmarshal` 不保留空指针的这一差异。上游允许额外对象字段时，类型投影接纳扩展但仅公开已声明字段；需要扩展字段时保留原始响应字节。封闭形状及精确根信封拒绝额外字段。生成器对不支持的新 Schema 约束、输出定义冲突及生成名称冲突明确失败，不弱化类型。

`contracts/embedded/v1` 和根目录两个 C 头文件均来自同一上游源码的精确副本。`go run ./scripts/generate-embedded-contract --check` 从包内契约验证生成的协议常量、状态码、命令元数据及全部线类型，不依赖相邻仓库。原生测试须显式设置 `LUASKILLS_NATIVE_E2E=1` 并配置匹配动态库。`python scripts/verify_embedded_distribution.py` 通过私有文件代理和空模块缓存验证实际 Go 模块 ZIP、包内生成及嵌入式测试；可用 `--go` 指定已安装工具链。该验证版本不会发布到外部注册表。

Windows cgo 显式链接 `luaskills.dll`，避免链接器在同目录误选 MSVC 静态库；因此链接目录和运行时 PATH 都必须包含匹配 DLL。该行为依据 [GNU ld 的 Windows 动态库链接规则](https://sourceware.org/binutils/docs/ld/WIN32.html)。Go 字节仅在同步 C 请求期间借用，响应复制完成后才释放，遵守 [cgo 指针规则](https://pkg.go.dev/cmd/cgo#hdr-Passing_pointers)。

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

仓库提供统一同步脚本，可直接下载 LuaSkills FFI、Lua runtime packages 与 VLDB，无需借用 Python 或 TypeScript 安装器：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/deps/sync_runtime_assets.ps1 -Target all -Database vldb-controller -RuntimeRoot D:\runtime\luaskills
```

```bash
RUNTIME_ROOT=/opt/luaskills scripts/deps/sync_runtime_assets.sh all vldb-controller
```

目标支持 `all`、`luaskills`、`lua`、`vldb`；VLDB 模式支持 `none`、`vldb-controller`、`vldb-direct`、`host-callback`。脚本默认固定 LuaSkills `v0.5.7`，并允许显式覆盖发布版本。

Go SDK 会规划并消费共享 SDK runtime manifest；上述仓库脚本负责直接下载 release 资产。当前共享 manifest 会指向：

- `LuaSkills/luaskills-packages` 的 `lua-runtime-packages-{platform}.tar.gz`
- `LuaSkills/luaskills` 的 `luaskills-ffi-sdk-{platform}.tar.gz`
- `runtime_root/dependencies/runtimes/...` 下可选的受管 Python、`uv`、Node.js 与 `pnpm` 路径

## 仓库结构

模块根目录是公共 `luaskills` package，保留导出的 API 领域以及 Go 构建标签要求的轻量原生桥。私有实现统一隔离在 `internal`：

```text
internal/protocol/       JSON FFI 响应包络解码
internal/runtimeassets/  runtime manifest 路径校验
internal/sessionwake/    并发受管会话唤醒回调所有权
examples/                可独立构建的 SDK 示例
scripts/                 运行时资产与受管运行时准备工具
```

cgo 与 no-cgo 桥接文件继续留在根 package，因为它们通过互斥构建标签实现同一组私有函数。内部 package 不暴露任何 C 类型，SDK 使用方也无法导入这些实现细节。

受管子运行时支持 Windows x64、Linux x64/ARM64 与 macOS x64/ARM64。Windows ARM 会在任何下载或目标目录创建前被明确拒绝。仓库还分发独立拉取与布局校验工具，供不使用 Python 或 TypeScript 安装器的宿主准备 debug 运行时：

当前受管依赖精确版本为 Python `3.14.6`、uv `0.11.28`、Node.js `24.18.0`、pnpm `11.11.0`。除非宿主有意安装其他受支持版本，否则包内 `dependencies.yaml` 必须声明相同的运行时与包管理器精确版本。

LuaSkills 0.5.1 将 LuaSkills 数据根、只读解释器发行根和可写受管环境根拆分为三个边界。两个显式受管根都必须是绝对路径；未设置时继续使用兼容的 `runtime_root/dependencies/runtimes` 与 `runtime_root/dependencies/envs` 布局。

```go
invokeTimeoutMS := uint64(30_000)
hostOptions := map[string]any{
	"managed_runtime_distribution_root": "D:/VulcanCode/dependencies/runtimes",
	"managed_runtime_environment_root":  "D:/VulcanCodeData/managed-runtime-envs",
	"managed_runtime_config": luaskills.ManagedRuntimeConfig{
		WorkerPoolMaxSizePerEnvironment:                   8,
		WorkerIdleTTLSecs:                                 120,
		PersistentSessionLimitPerEngine:                   128,
		PersistentSessionDefaultBufferLimitBytesPerStream: 2 * 1024 * 1024,
		InvokeDefaultTimeoutMS:                            &invokeTimeoutMS,
	},
}

pythonInstall, err := luaskills.ResolveManagedRuntimeInstall(luaskills.ManagedRuntimeResolveOptions{
	DistributionRoot: "D:/VulcanCode/dependencies/runtimes",
	Runtime:          luaskills.ManagedRuntimeKindPython,
	Version:          "3.14.6",
	Platform:         "windows-x64",
})
if err != nil {
	panic(err)
}
```

`DefaultManagedRuntimeConfig()` 返回稳定引擎默认值：单个精确环境/包所有者池 `4` 个 Worker、空闲 `60` 秒、`256` 个持久会话、每个 Session 输出流 `1 MiB`，且默认 invoke 超时为 nil。需要有限引擎默认超时时，应复制该值并把 `InvokeDefaultTimeoutMS` 设置为正数 `*uint64`。所有配置数值必须为正；单次 `invoke.timeout_ms` 与单个 Session 的 `session.open.buffer_limit_bytes` 只覆盖各自对应的引擎默认值。

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/deps/fetch_managed_runtimes.ps1 -RuntimeRoot D:\runtime\luaskills -Target all
python scripts/debug-tools/managed_runtime_layout_check.py D:\runtime\luaskills

# 拆分宿主管理根
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/deps/fetch_managed_runtimes.ps1 -RuntimeRoot D:\VulcanCodeData\luaskills -DistributionRoot D:\VulcanCode\dependencies\runtimes -Target all
python scripts/debug-tools/managed_runtime_layout_check.py D:\VulcanCodeData\luaskills --distribution-root D:\VulcanCode\dependencies\runtimes --environment-root D:\VulcanCodeData\managed-runtime-envs
```

```bash
RUNTIME_ROOT=/opt/luaskills scripts/deps/fetch_managed_runtimes.sh all
python3 scripts/debug-tools/managed_runtime_layout_check.py /opt/luaskills
```

默认情况下，这份共享 manifest 会把 LuaSkills core 固定到 SDK 对应版本，并从兼容的 `0.1` 协议线中自动解析最新已发布的 runtime packages patch 版本。

## 版本对齐

- 尽量让 SDK 与 LuaSkills core 保持同一条当前发布版本线。
- 当前 SDK 默认指向 LuaSkills core 标签 `v0.5.7`。
- runtime packages 与 native deps 仍然来自拆分后的 `LuaSkills/luaskills-packages` 及相关发布资产。
- SDK 默认 host options 传入 `runtime_root`、两个空的受管根覆盖槽与完整稳定的 `managed_runtime_config`；宿主未显式覆盖时，LuaSkills 会推导固定数据布局。
- 宿主工具直接放在 `runtime_root/bin`，不再放到 `runtime_root/bin/tools`。

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root D:\runtime\luaskills
npx @luaskills/sdk install-runtime --database none --managed-runtimes all --runtime-root D:\runtime\luaskills
```

```powershell
pip install luaskills-sdk
luaskills install-runtime --database vldb-direct --runtime-root D:\runtime\luaskills
```

Go 宿主可以检查同一份资产计划：

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

`DefaultHostOptions(runtimeRoot)` 会返回宿主选项与错误。`DefaultHostOptions` 与 `NewClient` 会在 manifest 存在时自动读取 `runtimeRoot/resources/luaskills-sdk-runtime-manifest.json`，并合入 `host_options_patch`。缺失 manifest 时保留 SDK 基础默认值；畸形 manifest 或逃逸 `runtimeRoot` 的宿主路径会返回带路径上下文的错误。

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
go run .\examples\runtime_lease
go run .\examples\provider_callback
```

`provider_callback` 同时覆盖 JSON provider callback 与 `vulcan.host.*` 宿主工具 callback 边界。在宿主安装自有 cgo callback bridge 前，它会返回需要宿主桥接的错误。

query、lifecycle 与持久 runtime-lease 示例使用内置夹具 skill：`examples/fixture-runtime/user_skills/demo-standard-ffi-skill`。请先使用 TypeScript 或 Python 安装器准备 runtime 资产：

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root .\examples\fixture-runtime
```

完整示例索引与 runtime 注意事项见 [examples/README_cn.md](examples/README_cn.md)。英文示例指南见 [examples/README.md](examples/README.md)。

## 持久运行时租约

普通租约入口请使用 `client.RuntimeLeases()`；如果宿主希望通过最新原生库提供的专用 system runtime-lease 导出固定注入 authority，请使用 `client.System(authority).RuntimeLeases()`。

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

## 迁移说明

- 现有 `client.System(authority)` 生命周期调用保持兼容；返回的 wrapper 现在额外暴露查询辅助方法和 `RuntimeLeases()`。
- `RuntimeLeaseHandle` 会持久化 `lease_id + sid + generation`，并在 `Eval`、`Status`、`Close` 时自动补回身份护栏。
- 绑定 authority 的运行时租约辅助层会直接分发到专用 `luaskills_ffi_system_runtime_lease_*` 入口。
- 当宿主在 `request_context.client_capabilities.host_result` 中显式开启结构化结果后，`CallSkill` 会返回可选 `HostResult`，供 IDE 原生结构化结果消费。
- 当 `HostResult.Kind == "change_set"` 时，宿主应把 `HostResult.Payload` 解析为 `RuntimeChangeSetPayload`。
- canonical `change_set` 现在使用文件生命周期记录；`modify` 通过 hunk 级 `before + delete[] + insert[] + after` 表达具体修改。
- `create` 与 `delete` 文件记录直接携带整文件 `content`，`rename` 记录携带 `old_path` 与 `new_path`。
- 普通租约接受 `cwd`、`workspace_root`、`lua_roots`、`c_roots`、`mounts`。System 租约强制要求 `SystemPackage`，拒绝 `lua_roots/c_roots`，并从可信包清单推导根目录。
- `PollManagedSessionEvents`、`WaitManagedSessionEvents`、`SetManagedSessionWakeCallback` 暴露 0.5.1 受管会话事件接口。
- Go 宿主在使用这些 API 时应部署匹配的最新 LuaSkills 原生动态库，因为 cgo 会直接按当前导出符号集合完成链接。

## 权限与管理

查询类接口建议默认使用 `AuthorityDelegatedTool`，因此委托工具看不到 ROOT skills。

`AuthoritySystem` 只表示宿主可以管理 ROOT；它不表示可以绕过 ROOT 所有权或同名 `skill_id` 冲突规则。

`CallSkill` 与 `RunLua` 是运行时执行面，不作为 ROOT 可见性过滤。

## 技能包配置

配置归属于 `skill_id` 标识的有效技能包，而不是包内单个入口。宿主必须在 `ClientOptions.HostOptions` 中把 `skill_config_root` 设置为绝对用户级目录；普通技能包与 ROOT 所属技能包分别保存到 `skills/config.json` 和 `system-skills/config.json`。每条原始 `List` 记录都包含 `StoreScope`，因此两个文件中同名技能包的保留记录仍可明确区分。严格版本化文档使用十进制字符串修订号、跨进程伴随锁、原子替换、缓存快照与文件监听重载；旧的无版本文档会被拒绝。

```go
schema, err := client.Config.Describe(luaskills.SkillPackageConfigDescribeOptions{
	SkillID: "example.settings",
})
status, err := client.Config.Validate("example.settings")
write, err := client.Config.SetValues(
	"example.settings",
	map[string]any{"api_key": "value", "retry_count": 3},
	"",
)
_, err = client.Config.Set(
	"example.settings",
	"retry_count",
	4,
	write.Revision,
)
```

`Describe` 返回参数名、五种稳定类型、约束、UI 提示、技能包作者提供的枚举元数据、有效值状态与完整性；`DescribeInstalled` 可在不执行 Lua 的情况下发现全部物理技能包。人类可读字段由技能包作者自行选择一种语言，建议广泛分发时使用英文，但不强制。

默认不返回配置值。`IncludeValues: true` 返回未遮罩有效值；宿主必须对读取和修改执行允许、拒绝、强制覆盖或用户授权策略，LuaSkills 与 SDK 不实现该策略。Lua 代码只能修改自身技能包配置，宿主 SDK 调用则有意不加跨包限制。

`SetValues` 是规范的原子批量操作，`Set` 把单个键包装到同一事务。修订号参数为写入与删除启用比较并交换。`PollEvents`、`WaitEvents`、`WatchEvents` 提供有序的本地写入与外部重载事件。配置缺失时，应展示 `Describe` 结果，并要求用户或已授权 AI 工具提供声明中的参数。技能包卸载后配置仍会保留，显式清理由宿主负责。

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

发布版本记录在 `VERSION`。Go 用户通过 `v0.5.7` 这类 Go module tag 消费 SDK 版本。

如果要做生态统一发布，必须先发布 `LuaSkills/luaskills-packages`，再发布 `LuaSkills/luaskills`；另外 Go 的 examples release 会通过已发布的 TypeScript 包安装 runtime 资产，因此 TypeScript SDK 也要先于 Go 示例工作流发布。

发布前执行：

```powershell
$env:CGO_ENABLED = "0"
go test ./...
```

推送匹配的 Go module tag 即完成 SDK 发布：

```powershell
git tag v0.5.7
git push origin v0.5.7
```

Go module tag 可用后，手动运行 GitHub Actions 里的 **Examples Release** 工作流。它会读取 `VERSION`，校验 `github.com/LuaSkills/luaskills-sdk-go@v{VERSION}`，通过已发布 TypeScript 安装器安装 LuaSkills runtime 资产，运行 Go 示例冒烟测试，然后创建或更新 `examples-v{VERSION}` GitHub Release，并上传：

- `luaskills-sdk-go-examples-{VERSION}.zip`
- `luaskills-sdk-go-examples-{VERSION}.zip.sha256`

示例 release tag 故意使用 `examples-v` 前缀，避免干扰 Go module 的语义版本 tag。

推荐统一发布顺序：`luaskills-packages` -> `luaskills` 核心仓库 -> TypeScript SDK -> Python SDK -> Go SDK -> 各 SDK 的 examples release。
