# Go SDK 仓库结构优化计划

## 任务目标

在保持公共导入路径 `github.com/LuaSkills/luaskills-sdk-go` 与现有公共 API 不变的前提下，将 JSON FFI 协议、运行时资产底层能力和会话唤醒注册实现下沉到 `internal`，降低根目录职责密度并明确 cgo/no-cgo 边界。

## 详细执行步骤

1. 建立根 package 文件、公共符号、内部调用和构建标签依赖图。
2. 抽取 JSON 响应包络解析到 `internal/protocol`，保留根 package 的调用适配。
3. 抽取运行时资产的无状态底层能力到 `internal/runtimeassets`，包括平台解析、Release 元数据与路径校验中不依赖公共类型的部分。
4. 抽取 managed-session wake callback 注册表到 `internal/sessionwake`，保留根 package 公共 API 与 cgo 桥接。
5. 将真实 cgo 集成测试与普通单元测试的职责和命名收口，补充仓库结构文档。
6. 运行 `gofmt`、`go vet`、cgo/no-cgo 测试、示例构建与全量测试，循环审核修复。

## 技术选型

- 根目录继续作为唯一公共 package，避免改变用户导入路径。
- 使用 Go `internal` 机制强制实现细节不成为外部 API。
- 内部包只定义基础值对象和无状态函数，公共领域类型继续由根 package 持有，防止循环依赖。
- cgo 文件继续通过构建标签隔离，不把 C 类型泄漏到内部纯 Go 包。

## 验收标准

- 根目录公共 API 与 `go doc` 输出保持一致。
- 根目录实现职责明显减少，新增内部包边界清晰且无循环依赖。
- `CGO_ENABLED=0` 与当前原生 cgo 测试均通过。
- `go vet ./...`、`go test ./...`、全部示例构建通过。
- README 中说明仓库目录与内部边界。

## 执行变更总结

### 1. 核心修复与调整概述

- 保持根模块公共导入路径和全部导出 API 不变，新增 `internal/protocol`、`internal/runtimeassets`、`internal/sessionwake` 三个私有实现包。
- 将 JSON FFI 响应包络解码及测试从根 package 完整迁移到协议内部包。
- 将 runtime manifest 路径片段和根目录包含关系校验下沉到运行时资产内部包。
- 将 managed-session wake callback 的并发注册表下沉到会话内部包，cgo 桥只负责原生注册和错误缓冲转换。
- 更新中英文 README，明确根 package、内部包、构建标签、示例和脚本的职责边界。

### 2. 📂文件变更清单

- 新增：`internal/protocol/envelope.go`、`internal/protocol/envelope_test.go`。
- 新增：`internal/runtimeassets/paths.go`、`internal/runtimeassets/paths_test.go`。
- 新增：`internal/sessionwake/registry.go`、`internal/sessionwake/registry_test.go`。
- 修改：`ffi.go`、`runtime_assets.go`、`managed_session_wake.go`。
- 修改：`README.md`、`README_cn.md`。
- 删除：根目录 `envelope.go`、`envelope_test.go`。

### 3. 💻关键代码调整详情

- `protocol.DecodeEnvelope` 成为唯一 JSON FFI 包络解码实现，错误文本和诊断上下文保持原语义。
- `runtimeassets.HasUnsafeRelativeSegment` 与 `runtimeassets.IsStrictlyInside` 集中承载 manifest 路径安全不变量。
- `sessionwake.Registry` 使用私有 `sync.Map` 管理 callback 生命周期，原生注册失败时仍由根桥恢复前一回调。
- 内部包不导入 `C`，C 类型继续限定在带 cgo 构建标签的根桥文件中。

### 4. ⚠️遗留问题与注意事项

- Go 规范要求同一个 package 位于同一目录，因此公共类型和方法仍保留在根目录；本次没有为了减少文件数而制造额外公共子 package。
- `runtime_assets.go` 仍包含较多公共领域类型和 manifest 编排逻辑，后续若继续增长，可在不改变公共类型的前提下继续抽取 Release 查询与平台映射。
- 本机默认没有 GCC，真实 cgo 验证使用已安装的 Zig 0.15.2 作为 C 编译器，并链接最新构建的 LuaSkills release 动态库。
- 最终连续五轮审核通过：格式与补丁、`go vet`、no-cgo 全量测试、真实 cgo 全量测试、公共 API 与 C 类型边界检查均无问题。
