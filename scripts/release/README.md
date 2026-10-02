# Go SDK 正式发布与恢复验证

本目录只协调 Go SDK。核心平台、资产归档、描述身份及实际 Cargo 消费的权威是固定 `core_commit`
检出的 `candidate.py`、`sdk_prerequisites.py`；精确 attempt、制品下载、签名绑定和全成员清单由
同一检出的 `SDK_RECOVERY.md`、`sdk_recovery.py` 负责。Go 核对已提交原字节后直接调用，
不复制平台表或恢复算法。上游工具变化后必须重读源码并重新冻结、消费实际模块 ZIP。

`VERSION` 是 Go SDK 唯一正式版本，必须与 `EmbeddedCoreVersion` 镜像一致，默认核心资产标签匹配
显式 `core_tag`。NPM/PyPI 版本独立输入并实际消费，示例安装命令使用已验 TS 版本。历史开发快照
`VERSION=0.5.7`、嵌入版本 `0.5.9` 因身份不一致而被正式冻结拒绝。当前源码面向 `0.6.0`
发布线；实际正式状态以发布证据为准。此流程不修改版本或覆盖旧标签。

## 两条签名链

SDK 工作流默认 `mode=artifact-only`。冻结唯一实际模块 ZIP，动态全部核心平台在空缓存/private proxy
中消费该 ZIP，执行真实 cgo、race、取消后的同 receipt 恢复和完整关闭。严格 aggregate 完成后，
独立 `candidate-evidence` 作业重新聚合原始日志并签名候选 subjects，上传唯一
`candidate-evidence-rRUNID-aATTEMPT` 制品。此步骤在任何 tag 或公开 Release 写入前完成。

原候选身份包括 SDK 源码 SHA、run ID、attempt、实际 artifact ID、官方签名及全部物理字节。
恢复不把原 artifact-only 改成发布授权，也不改写原整体 failure。候选精确 attempt 允许实际
in_progress/null 或 completed 的真实 conclusion；freeze、全部 native、aggregate、candidate-evidence
仍须各自 completed/success。下载仅使用显式 artifact ID，过期、缺成员、跨 attempt、源 SHA 不符、
原 bundle/manifest/签名摘要变化均失败，不寻找 latest 或其他制品。

当前显式 `mode=publish` 才继续本次候选；`mode=recover` 指定原候选 run、attempt、artifact ID。
当前 completion 必须来自精确 production dispatch、同 SDK/工作流完整 SHA 和 OIDC 环境。
新完整核心 Cargo 重验和 TS/PY 正式双链消费者先于 Go tag。主 `vVERSION` Release 只保存认证原候选
固定资产；相同完整字节可复用，final 缺资产或不同字节拒绝，绝不增补或 clobber。

随后冷消费公开 Go proxy，逐成员比较冻结 ZIP，再实际 cgo/race/native 生命周期验收。
新独立 `recovery-vVERSION-rCOMPLETIONRUNID-aATTEMPT` Release 保存单独签名完成证据，绑定原候选
bundle、manifest、官方签名、binding 摘要及主 Release 实际 ID、源码、完整资产清单。
初版 completion 源码 SHA 等于原 SDK SHA，二者独立记录。正式下游只接受指定的实际
completed/success completion attempt，原 whole-attempt 状态保留。

主 Release 固定资产：

- go-module.zip、go-aggregate.json、go-core-prerequisites.json
- go-candidate-proof.tar.gz、go-candidate.json
- recovery-binding.json、go-candidate-attestation.jsonl

独立完成 Release 固定资产：

- go-completion.json、go-completion-proof.tar.gz、go-completion-consumer.json
- go-completion-core-prerequisites.json、go-completion-prerequisites.json
- go-completion-attestation.jsonl

原 bundle 保存源码成员清单、实际核心凭证及全部原生日志。完成 bundle 保存实际新核心、前置
SDK 消费文件和公开 Go 原生日志。正式下游从两个公开 Release 验签，无需原 Actions 制品永久保留；
继续恢复发布仍要求指定原制品未过期。

## CLI 与独立前置权威

`python scripts/release/go_release.py COMMAND --help` 提供精确必需参数，不保留含混 `--run-id`。

| 命令 | 实际边界 |
| --- | --- |
| bootstrap | fresh github-only 与公共 toolchain_inputs，输出精确稳定 Rust toolchain；complete=false |
| freeze | 干净提交、SDK/工作流 SHA、版本、独立 npm/pypi 双身份，冻结一个 ZIP 并实际完整核心消费 |
| native / aggregate | 动态全部核心平台消费同 ZIP、核对实际日志；禁止 skip、source replace、备用库 |
| candidate-stage / candidate-finalize | 重聚合原始日志、schema2 binding、验官方签名，上传固定根成员 |
| candidate-consume | 显式 candidate-run-id/candidate-run-attempt/candidate-artifact-id；验字节、签名和精确 attempt |
| completion-prepare / completion-publish | intent=publish/recover，新前置消费→唯一 tag→原主资产→公开冷消费→独立签名完成 Release |
| formal-proof | 双链永久证据、精确成功完成 attempt、当前新核心和 TS/PY/Go 冷消费；失败不生成 accepted.json |
| examples | 六个已发布示例，独立 TS 安装器、精确核心、公开 Go 模块新缓存、确定性 ZIP |
| examples-stage / examples-finalize / examples-publish | 独立签名原 ZIP/sidecar，显式原示例 run/attempt/artifact ID，恢复保持原字节 |

Go formal-proof 必须显式传入 `--sdk-sha --sdk-version --candidate-run-id --candidate-run-attempt
--completion-run-id --completion-run-attempt --completion-source-sha --core-root --core-tag --core-commit
--npm-root --pypi-root --platform --output`。成功写 schema2 accepted.json：原 SDK 与完成源码、
两个 run ID（字符串）/attempt（整数），registry_consumer_file=fresh-formal-consumer.json 及真实摘要。
原永久 issuer 消费字节保留，本次新消费者文件独立。

NPM/PyPI 各需独立版本、原源码 SHA、candidate run/attempt、completion run/attempt、completion 源码 SHA。
Go 只调用固定 checkout 的权威 formal-proof CLI，核对 schema2 公共 accepted 头和实际 fresh 消费文件，
不解析其他 SDK 内部 Release 协议。未配置正式账户或尚未成功时前置失败，不能以本地测试声称已发布。

需要 Cargo 的新 runner 均 fresh github-only→公共 toolchain_inputs→安装返回的
`rustup toolchain install --profile minimal`→complete/recheck，不从 MSRV、另一平台或本机 stable 推断。
`--platform host` 仅选唯一当前宿主，不证明跨平台成功。

## 权限、签名与示例

新工作流先进入公共默认分支，再通过分支/标签 dispatch 冻结完整 SHA；输入、GITHUB_SHA、
GITHUB_WORKFLOW_SHA 与实际 HEAD 必须一致。提交须在默认分支历史中，工作流全子树与默认分支当前树一致。
普通 GITHUB_TOKEN 无可声明 Workflows 写权限；不引入 PAT 绕过新工作流标签限制。
[Git refs API](https://docs.github.com/en/rest/git/refs)，
[工作流权限与 dispatch 输入](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)

官方固定 actions/attest v4.1.0 使用 OIDC/Sigstore。每个实际 subject 执行 gh attestation verify 的固定
repo/signer-workflow/source-digest/signer-digest/deny-self-hosted-runners/bundle 参数，核对实际
certificate.runInvocationURI 和 statement invocationId 到指定 run/attempt。公开自摘要不能代替验签。
[CLI 验证合同](https://cli.github.com/manual/gh_attestation_verify)，
[Sigstore 验证结果](https://github.com/sigstore/sigstore-go/blob/main/docs/verification.md)

SDK、completion、独立 examples-vVERSION Release 均 draft→上传全部→publish。
草稿通过认证完整分页按唯一 tag 获取实际 ID；按 tag 公共端点不能取得草稿。
写入前和编辑公开前递归核对实际 tag，仅初始精确 ref 404 表示不存在，annotation 子对象 404 显式失败。
服务端 immutable 属性如实记录，不要求或修改 Admin 设置。
[GitHub Release API](https://docs.github.com/en/rest/releases/releases)

SDK completion 显式 dispatch 示例 mode=publish，不依赖 token 标签事件。
示例默认 artifact-only：先重新正式验收 SDK 双链并运行六例，再官方签名 ZIP/原 sidecar/binding。
ZIP 固定 1980 时间戳、普通 0644 权限、排序和 ZIP_STORED，恢复不重新制作 ZIP。
publish/recover 重新 formal-proof 后仅消费指定原签名示例制品。开发 embedded_lifecycle 始终排除，
文档相对链接目标随包。
[触发事件](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)

## 离线验收边界

先设置 LUASKILLS_RELEASE_CORE_ROOT 为唯一实际核心 checkout，再执行：

```powershell
$env:LUASKILLS_RELEASE_CORE_ROOT = "D:/projects/vulcan-luaskills"
rtk proxy python -X utf8 -m unittest discover -s scripts/release -p test_*.py -v
rtk proxy python -X utf8 -m unittest discover -s scripts -p test_prepare_examples_release.py -v
```

测试使用实际源码 ZIP、归档及共同恢复 authority，显式 mock API/密码学工具边界，覆盖精确原失败/活动
attempt、部分门禁、过期/篡改、三原摘要、主资产冲突、公开成功回读丢失、独立 SDK 版本、签名证书、
实际 ZIP 元数据重复性、原 sidecar 与父 SDK attempt 替换拒绝。它们不等于实际 OIDC、公开 registry、
全部平台 CI 或远端发布。旧开发 ZIP 的 219 native/race 通过仅对应当时冻结源码；当前恢复源码变更后，
须新正式版本、冻结 ZIP、匹配 DLL 再验收，不复用历史证明。
