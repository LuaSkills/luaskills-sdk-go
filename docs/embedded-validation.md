# Embedded candidate validation / 嵌入式候选验收

The **Embedded Offline Contract** CI uses `CGO_ENABLED=0` to check generation, run `go test -p 4 -parallel 4 -count=1 ./...` and verify a real Go module ZIP through a private file proxy and empty cache. The ZIP verifier explicitly disables cgo and native E2E. `TestEmbeddedNoCgo` proves native construction fails while codecs remain available. Offline success does not establish native acceptance.

**Embedded Offline Contract** CI 使用 `CGO_ENABLED=0` 检查生成、执行 `go test -p 4 -parallel 4 -count=1 ./...`，并通过私有文件代理及空缓存验证真实 Go 模块 ZIP。ZIP 校验器显式禁用 cgo 及原生 E2E。`TestEmbeddedNoCgo` 证明原生构造明确失败而编码器仍可用。离线成功不表示原生验收通过。

Freeze the actual current Go module source first. The destination must be a fresh absolute directory; `module.zip`, `module.mod`, `module.info`, `module-members.json` and `freeze.json` record the exact bytes/version. The version is read from the unchanged `VERSION` file. The ZIP SHA-256, not a source copy or published tag, identifies this local frozen artifact.

先冻结实际当前 Go 模块源码。目标必须是全新绝对目录；`module.zip`、`module.mod`、`module.info`、`module-members.json` 和 `freeze.json` 记录精确字节／版本。版本读取未修改的 `VERSION` 文件；ZIP SHA-256 标识该本地冻结产物，不以源码副本或已发布标签替代。

```powershell
python scripts/freeze_embedded_module.py --output D:/candidate/go-module-freeze
python scripts/verify_embedded_distribution.py --module-zip D:/candidate/go-module-freeze/module.zip --module-zip-sha256 <freeze.json-module_zip_sha256> --module-version <freeze.json-module_version>
python scripts/verify_embedded_candidate.py --library D:/candidate/luaskills.dll --library-sha256 <frozen-library-sha256> --description D:/candidate/core-description.json --module-zip D:/candidate/go-module-freeze/module.zip --module-zip-sha256 <freeze.json-module_zip_sha256> --module-version <freeze.json-module_version> --race
```

`--module-zip`, `--module-zip-sha256` and `--module-version` must all be explicit. The gate validates canonical module/version ZIP paths, duplicate/non-regular/escaped members, required files, the embedded VERSION, go.mod module path and the full ZIP hash before Go execution. Go consumes that ZIP through a private `GOPROXY` and a new empty `GOMODCACHE`; the consumer has no replace directive. Downloaded archive/file bytes and Go's actual package resolution must match. Generation/tests read fixtures from the downloaded module directory, and the external example is built from that same import path. `go-mod-download.json`, `go-list.json` and the per-member manifest retain the evidence.

`--module-zip`、`--module-zip-sha256` 和 `--module-version` 三项必须全部显式提供。门禁在 Go 执行前检查规范模块／版本 ZIP 路径、重复／非普通／逃逸成员、必需文件、包内 VERSION、go.mod 模块路径及完整 ZIP 摘要。Go 通过私有 `GOPROXY` 和全新空 `GOMODCACHE` 消费此 ZIP；消费方没有 replace 指令。已下载归档／文件字节及 Go 实际包解析必须一致。生成／测试从已下载模块目录读取夹具，外部示例从同一导入路径构建。`go-mod-download.json`、`go-list.json` 及逐成员清单保留证据。

`--go` selects an installed executable. `--timeout` sets the native test timeout (default `10m`). `--description` contains the raw frozen `luaskills_ffi_embedded_describe_v1` JSON, not a manifest envelope. Use the platform's exact shared-library name: `luaskills.dll`, `libluaskills.so` or `libluaskills.dylib`.

`--go` 指定已安装可执行文件；`--timeout` 指定原生测试超时，默认 `10m`。`--description` 必须是冻结的 `luaskills_ffi_embedded_describe_v1` 原始 JSON，不是清单信封。库名称须匹配平台：`luaskills.dll`、`libluaskills.so` 或 `libluaskills.dylib`。

Missing files, incorrect binary hashes and malformed description JSON fail before compilation. The existing generated decoder and compatibility validator check the full description in `TestEmbeddedCandidateEvidence`; that test compares every field against metadata from the actual linked library. Linking/loading failures, any selected skipped test, description mismatches or library/module changes fail the gate. Binary hashes and selected build-input hashes remain distinct evidence.

缺文件、二进制摘要不符及畸形描述 JSON 在编译前失败。`TestEmbeddedCandidateEvidence` 使用现有生成解码器及兼容性校验器检查完整描述，并将每个字段与实际链接库的元数据比较。链接／加载失败、任何选中测试跳过、描述不符或库／模块发生变化，均使门禁失败。二进制摘要与选定构建输入摘要分别记录。

The gate forces `LUASKILLS_NATIVE_E2E=1`, disables remote dependency/toolchain downloads and clears ambient library search flags. The hash-checked library is copied into private staging; Windows tests and the example execute beside that DLL. Successful evidence identifies `mode=frozen-module-zip`, ZIP hash/version, every member, actual cache import directory and `replace=false`. For source development only, explicitly use `--source-development` without ZIP arguments; its result is labeled `mode=source-development` and `replace=true`, never ZIP acceptance. There is no automatic mode fallback. Evidence remains under `.temp/embedded-candidate-*`. This does not establish a published module tag, release build or five-platform acceptance. It neither changes VERSION nor authorizes republishing an already published version. This local validation entry point does not trigger publication; actual publication uses the repository SDK and standalone examples release workflows.

门禁强制 `LUASKILLS_NATIVE_E2E=1`、禁止远端依赖／工具链下载，并清除外部库搜索标志。已校验库复制到私有暂存目录；Windows 测试与示例在该 DLL 旁执行。成功证据标明 `mode=frozen-module-zip`、ZIP 摘要／版本、每个成员、实际缓存导入目录及 `replace=false`。仅开发源码验证时须显式使用 `--source-development` 且不传 ZIP 参数，其结果标为 `mode=source-development`、`replace=true`，不代表 ZIP 验收；不存在自动模式回退。证据保存在 `.temp/embedded-candidate-*`。这不代表正式模块标签、发布构建或五平台验收，不修改 VERSION，也不授权重新发布已存在版本。此本地验证入口本身不触发发布；实际发布使用仓库的 SDK 与独立示例发布工作流。

`examples/embedded_lifecycle` demonstrates the typed client, automatic queued Go callback, actual operation observation, native operation forgetting, scope-owned runtime/pump closure, driver join/receipt forgetting and transport release. Its isolated runtime forbids downloads. Failed cleanup reports retained owners and preserves the runtime directory; observer cancellation never counts as actual closure. An accepted reservation retains its typed pending receipt before observation; cleanup recovers the same runtime from that receipt without issuing Reserve again. The candidate gate runs this cancellation regression against the actual selected library.

`examples/embedded_lifecycle` 完整演示类型化客户端、自动队列 Go 回调、真实操作观察、原生操作遗忘、作用域拥有的运行时／泵关闭、驱动汇合／回执遗忘及传输释放。隔离运行时禁止下载；清理失败明确报告保留所有者并保存目录，观察者取消不代表实际关闭。已接纳预留在观察前保留类型化 pending 回执；清理从同一回执恢复原运行时，不再次发出 Reserve。候选门禁使用实际选定库执行此取消回归。

The published examples package includes only the six examples in `scripts/prepare_examples_release.py` and their fixture sources. `embedded_lifecycle` is excluded. Extend the published inventory only alongside a matching SDK module version and its actual smoke test. The workflow verifies the final ZIP rejects unpublished directories and checks every relative Markdown file target. `python scripts/prepare_examples_release.py --check --npm-version 1.2.3` performs the same inspection on a real temporary package/ZIP without downloading modules or runtime assets. The offline installer version is an explicit fixture; formal packaging takes the independently authenticated TypeScript version. The standalone package ships this document so root/example README links remain valid. The candidate scripts and development example are available only in a matching [SDK source checkout](https://github.com/LuaSkills/luaskills-sdk-go).

正式示例包仅包含 `scripts/prepare_examples_release.py` 中声明的六个示例及夹具源码，排除 `embedded_lifecycle`。扩展正式清单必须同时匹配 SDK 模块版本及实际冒烟测试。工作流检查最终 ZIP，拒绝未发布目录并检查每个相对 Markdown 文件目标。`python scripts/prepare_examples_release.py --check --npm-version 1.2.3` 对真实临时包／ZIP执行相同检查，不下载模块或运行时资产。离线安装器版本是显式夹具；正式打包使用独立认证的 TypeScript 版本。独立包包含本文档，使根／示例 README 链接保持有效；候选脚本及开发示例仅在匹配的 [SDK 源码检出](https://github.com/LuaSkills/luaskills-sdk-go) 中提供。
