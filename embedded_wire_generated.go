// Code generated from the packaged embedded contract; DO NOT EDIT.
// 从包内嵌入式契约生成；请勿手工编辑。
package luaskills

import "reflect"

// Immutable capability declaration shared by Rust, generated contracts and SDKs.
// Rust、生成契约与 SDK 共享的不可变能力声明。
type EmbeddedInputCapabilityDescriptor struct {
	// English description of the host capability for tool consumers.
	// 面向工具消费者的宿主能力英文描述。
	Description string `json:"description"`
	// Declared mutation category.
	// 声明的变更类别。
	Effects EmbeddedInputCapabilityEffects `json:"effects"`
	// Explicit native or queued dispatch protocol.
	// 显式原生或队列分发协议。
	Execution EmbeddedInputCapabilityExecution `json:"execution"`
	// Explicit host deduplication contract.
	// 显式宿主去重契约。
	Idempotency EmbeddedInputCapabilityIdempotency `json:"idempotency"`
	// Offline input value contract.
	// 离线输入值契约。
	InputSchema any `json:"input_schema"`
	// Per-capability budget capped by the original operation deadline.
	// 受原始操作截止时间约束的单能力预算。
	MaxCallMs uint64 `json:"max_call_ms"`
	// Maximum in-flight handlers, including cancelled handlers that have not stopped.
	// 在途处理器上限，包含已取消但尚未停止的处理器。
	MaxConcurrent uint64 `json:"max_concurrent"`
	// Maximum serialized input bytes within the parent value limit.
	// 父级值上限内的最大序列化输入字节数。
	MaxInputBytes uint64 `json:"max_input_bytes"`
	// Maximum serialized output bytes within the parent value limit.
	// 父级值上限内的最大序列化输出字节数。
	MaxOutputBytes uint64 `json:"max_output_bytes"`
	// Exact namespaced name; discovery exposes only authorized declarations.
	// 精确命名空间名称；发现操作仅暴露已授权声明。
	Name string `json:"name"`
	// Offline output value contract.
	// 离线输出值契约。
	OutputSchema any `json:"output_schema"`
	// Every listed grant must still exist at each admission boundary.
	// 每个入场边界仍必须拥有列出的全部授权。
	Permissions EmbeddedInputCapabilityDescriptorPermissions `json:"permissions"`
	// Required trusted invocation scope.
	// 必需的可信调用作用域。
	Scope EmbeddedInputCapabilityScope `json:"scope"`
	// Semantic interface version, independent from the core library version.
	// 语义接口版本，独立于核心库版本。
	Version string `json:"version"`
}

// Every listed grant must still exist at each admission boundary.
// 每个入场边界仍必须拥有列出的全部授权。
type EmbeddedInputCapabilityDescriptorPermissions []string

// Declared effect category; it does not make external mutations transactional.
// 声明的副作用类别；它不会使外部变更自动具有事务性。
type EmbeddedInputCapabilityEffects string

const (
	// Host contract promises no externally visible mutation.
	// 宿主契约承诺不产生外部可见变更。
	EmbeddedInputCapabilityEffectsReadOnly EmbeddedInputCapabilityEffects = "read_only"
	// Host must report actual commit, rollback or unknown status.
	// 宿主必须报告真实提交、回滚或未知状态。
	EmbeddedInputCapabilityEffectsMutating EmbeddedInputCapabilityEffects = "mutating"
)

// Host execution transport, explicitly selected before a capability is published.
// 宿主执行传输，在能力发布前显式选择。
type EmbeddedInputCapabilityExecution string

const (
	// Short cooperative Rust callback executed on the owning VM thread.
	// 在所属 VM 线程执行的短时协作 Rust 回调。
	EmbeddedInputCapabilityExecutionNative EmbeddedInputCapabilityExecution = "native"
	// Reliable request consumed and completed by an SDK event pump.
	// 由 SDK 事件泵消费并完成的可靠请求。
	EmbeddedInputCapabilityExecutionQueued EmbeddedInputCapabilityExecution = "queued"
)

// Explicit side-effect deduplication support, never inferred from an operation identifier.
// 显式副作用去重支持，绝不从操作标识推断。
type EmbeddedInputCapabilityIdempotency string

const (
	// Automatic replay is forbidden; an uncertain result needs host reconciliation.
	// 禁止自动重放；不确定结果需要宿主对账。
	EmbeddedInputCapabilityIdempotencyNone EmbeddedInputCapabilityIdempotency = "none"
	// Host implementation deduplicates the supplied request identity durably.
	// 宿主实现对提供的请求身份进行持久去重。
	EmbeddedInputCapabilityIdempotencyHostRequest EmbeddedInputCapabilityIdempotency = "host_request"
)

// Scope required from the trusted caller, independent from Lua business arguments.
// 可信调用方必须具备的作用域，独立于 Lua 业务参数。
type EmbeddedInputCapabilityScope string

const (
	// Available during an ordinary operation or session invocation.
	// 在普通操作或会话调用期间可用。
	EmbeddedInputCapabilityScopeInvocation EmbeddedInputCapabilityScope = "invocation"
	// Requires an explicitly bound session identity.
	// 要求显式绑定的会话身份。
	EmbeddedInputCapabilityScopeSession EmbeddedInputCapabilityScope = "session"
)

// Implemented commands only; new runtime commands are advertised when actually wired.
// 仅包含已实现命令；新的运行时命令在实际接通后才公布。
type EmbeddedInputCommand interface {
	// embeddedVariantEmbeddedInputCommand seals this generated union without invoking serialization hooks.
	// embeddedVariantEmbeddedInputCommand 封闭此生成联合，不调用序列化钩子。
	embeddedVariantEmbeddedInputCommand()
}

// embeddedVariantEmbeddedInputCommand marks the exact EmbeddedInputCommandRuntime alternative.
// embeddedVariantEmbeddedInputCommand 标识精确的 EmbeddedInputCommandRuntime 分支。
func (EmbeddedInputCommandRuntime) embeddedVariantEmbeddedInputCommand() {}

// embeddedVariantEmbeddedInputCommand marks the exact EmbeddedInputCommandDescribe alternative.
// embeddedVariantEmbeddedInputCommand 标识精确的 EmbeddedInputCommandDescribe 分支。
func (EmbeddedInputCommandDescribe) embeddedVariantEmbeddedInputCommand() {}

// embeddedVariantEmbeddedInputCommand marks the exact EmbeddedInputCommandRuntimeReserve alternative.
// embeddedVariantEmbeddedInputCommand 标识精确的 EmbeddedInputCommandRuntimeReserve 分支。
func (EmbeddedInputCommandRuntimeReserve) embeddedVariantEmbeddedInputCommand() {}

// embeddedVariantEmbeddedInputCommand marks the exact EmbeddedInputCommandRuntimeInitialize alternative.
// embeddedVariantEmbeddedInputCommand 标识精确的 EmbeddedInputCommandRuntimeInitialize 分支。
func (EmbeddedInputCommandRuntimeInitialize) embeddedVariantEmbeddedInputCommand() {}

// embeddedVariantEmbeddedInputCommand marks the exact EmbeddedInputCommandRuntimeStatus alternative.
// embeddedVariantEmbeddedInputCommand 标识精确的 EmbeddedInputCommandRuntimeStatus 分支。
func (EmbeddedInputCommandRuntimeStatus) embeddedVariantEmbeddedInputCommand() {}

// embeddedVariantEmbeddedInputCommand marks the exact EmbeddedInputCommandRuntimeClose alternative.
// embeddedVariantEmbeddedInputCommand 标识精确的 EmbeddedInputCommandRuntimeClose 分支。
func (EmbeddedInputCommandRuntimeClose) embeddedVariantEmbeddedInputCommand() {}

// embeddedVariantEmbeddedInputCommand marks the exact EmbeddedInputCommandRuntimeFree alternative.
// embeddedVariantEmbeddedInputCommand 标识精确的 EmbeddedInputCommandRuntimeFree 分支。
func (EmbeddedInputCommandRuntimeFree) embeddedVariantEmbeddedInputCommand() {}

// Inspect effective transport limits and implemented protocol commands.
// 查看有效传输限制与已实现协议命令。
type EmbeddedInputCommandDescribe struct {
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputCommandDescribeType `json:"type"`
}

// EmbeddedInputCommandDescribeType is derived from the packaged wire contract.
// EmbeddedInputCommandDescribeType 从包内线契约派生。
type EmbeddedInputCommandDescribeType string

const (
	// EmbeddedInputCommandDescribeTypeDescribe is derived from the packaged wire contract.
	// EmbeddedInputCommandDescribeTypeDescribe 从包内线契约派生。
	EmbeddedInputCommandDescribeTypeDescribe EmbeddedInputCommandDescribeType = "describe"
)

// Execute one typed operation on an exact initialized runtime.
// 在精确已初始化运行时上执行一个类型化操作。
type EmbeddedInputCommandRuntime struct {
	// Typed core operation with no legacy command aliases.
	// 不含旧命令别名的类型化核心操作。
	Operation EmbeddedInputRuntimeCommand `json:"operation"`
	// Exact transport-local runtime identity.
	// 精确传输局部运行时身份。
	RuntimeId string `json:"runtime_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputCommandRuntimeType `json:"type"`
}

// Close existing admission, including a construction attempt that is still running.
// 关闭已有入场，包含仍在运行的构造尝试。
type EmbeddedInputCommandRuntimeClose struct {
	// Exact retained runtime identity in this transport.
	// 此传输中保留的精确运行时身份。
	RuntimeId string `json:"runtime_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputCommandRuntimeCloseType `json:"type"`
}

// EmbeddedInputCommandRuntimeCloseType is derived from the packaged wire contract.
// EmbeddedInputCommandRuntimeCloseType 从包内线契约派生。
type EmbeddedInputCommandRuntimeCloseType string

const (
	// EmbeddedInputCommandRuntimeCloseTypeRuntimeClose is derived from the packaged wire contract.
	// EmbeddedInputCommandRuntimeCloseTypeRuntimeClose 从包内线契约派生。
	EmbeddedInputCommandRuntimeCloseTypeRuntimeClose EmbeddedInputCommandRuntimeCloseType = "runtime_close"
)

// Remove an explicitly closed and fully drained runtime registration.
// 移除显式关闭且完全排空的运行时注册。
type EmbeddedInputCommandRuntimeFree struct {
	// Exact retained runtime identity in this transport.
	// 此传输中保留的精确运行时身份。
	RuntimeId string `json:"runtime_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputCommandRuntimeFreeType `json:"type"`
}

// EmbeddedInputCommandRuntimeFreeType is derived from the packaged wire contract.
// EmbeddedInputCommandRuntimeFreeType 从包内线契约派生。
type EmbeddedInputCommandRuntimeFreeType string

const (
	// EmbeddedInputCommandRuntimeFreeTypeRuntimeFree is derived from the packaged wire contract.
	// EmbeddedInputCommandRuntimeFreeTypeRuntimeFree 从包内线契约派生。
	EmbeddedInputCommandRuntimeFreeTypeRuntimeFree EmbeddedInputCommandRuntimeFreeType = "runtime_free"
)

// Attempt construction exactly once; the existing identity retains the actual outcome for query.
// 精确尝试构造一次；已有身份保留实际结果供查询。
type EmbeddedInputCommandRuntimeInitialize struct {
	// Explicit core engine options, using the existing engine option contract.
	// 显式核心引擎选项，使用现有引擎选项契约。
	EngineOptions EmbeddedInputLuaEngineOptions `json:"engine_options"`
	// Explicit durable storage; absence selects memory-only execution without creating a database.
	// 显式持久存储；缺失表示纯内存执行，不创建数据库。
	Persistence **EmbeddedInputRuntimePersistenceConfig `json:"persistence,omitempty"`
	// Explicit formal runtime budgets validated before worker construction.
	// 工作线程构造前校验的显式正式运行时预算。
	RuntimeConfig EmbeddedInputEmbeddedRuntimeConfig `json:"runtime_config"`
	// Exact identity returned by runtime_reserve in this transport.
	// 此传输中 runtime_reserve 返回的精确身份。
	RuntimeId string `json:"runtime_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputCommandRuntimeInitializeType `json:"type"`
}

// EmbeddedInputCommandRuntimeInitializeType is derived from the packaged wire contract.
// EmbeddedInputCommandRuntimeInitializeType 从包内线契约派生。
type EmbeddedInputCommandRuntimeInitializeType string

const (
	// EmbeddedInputCommandRuntimeInitializeTypeRuntimeInitialize is derived from the packaged wire contract.
	// EmbeddedInputCommandRuntimeInitializeTypeRuntimeInitialize 从包内线契约派生。
	EmbeddedInputCommandRuntimeInitializeTypeRuntimeInitialize EmbeddedInputCommandRuntimeInitializeType = "runtime_initialize"
)

// Allocate a bounded metadata-only runtime identity before construction can begin.
// 在构造能够开始前分配有界且仅含元数据的运行时身份。
type EmbeddedInputCommandRuntimeReserve struct {
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputCommandRuntimeReserveType `json:"type"`
}

// EmbeddedInputCommandRuntimeReserveType is derived from the packaged wire contract.
// EmbeddedInputCommandRuntimeReserveType 从包内线契约派生。
type EmbeddedInputCommandRuntimeReserveType string

const (
	// EmbeddedInputCommandRuntimeReserveTypeRuntimeReserve is derived from the packaged wire contract.
	// EmbeddedInputCommandRuntimeReserveTypeRuntimeReserve 从包内线契约派生。
	EmbeddedInputCommandRuntimeReserveTypeRuntimeReserve EmbeddedInputCommandRuntimeReserveType = "runtime_reserve"
)

// Read construction outcome and actual worker closure evidence.
// 读取构造结果与实际工作线程关闭证据。
type EmbeddedInputCommandRuntimeStatus struct {
	// Exact retained runtime identity in this transport.
	// 此传输中保留的精确运行时身份。
	RuntimeId string `json:"runtime_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputCommandRuntimeStatusType `json:"type"`
}

// EmbeddedInputCommandRuntimeStatusType is derived from the packaged wire contract.
// EmbeddedInputCommandRuntimeStatusType 从包内线契约派生。
type EmbeddedInputCommandRuntimeStatusType string

const (
	// EmbeddedInputCommandRuntimeStatusTypeRuntimeStatus is derived from the packaged wire contract.
	// EmbeddedInputCommandRuntimeStatusTypeRuntimeStatus 从包内线契约派生。
	EmbeddedInputCommandRuntimeStatusTypeRuntimeStatus EmbeddedInputCommandRuntimeStatusType = "runtime_status"
)

// EmbeddedInputCommandRuntimeType is derived from the packaged wire contract.
// EmbeddedInputCommandRuntimeType 从包内线契约派生。
type EmbeddedInputCommandRuntimeType string

const (
	// EmbeddedInputCommandRuntimeTypeRuntime is derived from the packaged wire contract.
	// EmbeddedInputCommandRuntimeTypeRuntime 从包内线契约派生。
	EmbeddedInputCommandRuntimeTypeRuntime EmbeddedInputCommandRuntimeType = "runtime"
)

// Host-reported effect outcome, independent from execution success or cancellation.
// 宿主报告的副作用结果，独立于执行成功或取消。
type EmbeddedInputEffectState string

const (
	// No business execution has started.
	// 尚未开始业务执行。
	EmbeddedInputEffectStateNotStarted EmbeddedInputEffectState = "not_started"
	// The declared operation has no externally visible mutations.
	// 声明的操作不包含外部可见变更。
	EmbeddedInputEffectStateNotApplicable EmbeddedInputEffectState = "not_applicable"
	// A trusted host transaction explicitly confirmed its commit.
	// 可信宿主事务明确确认其提交。
	EmbeddedInputEffectStateCommitted EmbeddedInputEffectState = "committed"
	// A trusted host transaction explicitly confirmed its rollback.
	// 可信宿主事务明确确认其回滚。
	EmbeddedInputEffectStateRolledBack EmbeddedInputEffectState = "rolled_back"
	// Effects may have occurred; retries require host-specific reconciliation.
	// 副作用可能已发生；重试需要宿主特定的对账。
	EmbeddedInputEffectStateUnknown EmbeddedInputEffectState = "unknown"
)

// Owned structured request admitted under one original deadline.
// 在单个原始截止时间下接纳的拥有所有权的结构化请求。
type EmbeddedInputEmbeddedCall struct {
	// Application value whose encoded size is checked before admission.
	// 入场前检查编码大小的应用值。
	Arguments any `json:"arguments"`
	// Trusted host context retained and charged with the queued request.
	// 与排队请求一并保留和计费的可信宿主上下文。
	Context EmbeddedInputLuaInvocationContext `json:"context"`
	// Exact declared module export, never an evaluated code fragment.
	// 精确声明的模块导出，绝不是求值代码片段。
	Export string `json:"export"`
	// Exact immutable pool identity returned by this runtime.
	// 此运行时返回的精确不可变池身份。
	PoolId string `json:"pool_id"`
}

// Structured error; `code` is stable and `message` is an English diagnostic.
// 结构化错误；`code` 稳定，`message` 为英文诊断信息。
type EmbeddedInputEmbeddedError struct {
	// Stable classification consumed by SDKs instead of parsing text.
	// 供 SDK 使用的稳定分类，避免解析文案。
	Code EmbeddedInputEmbeddedErrorCode `json:"code"`
	// Human-readable detail without credentials or plugin input dumps.
	// 不含凭据或插件输入转储的可读详情。
	Message string `json:"message"`
}

// Stable machine-readable failures shared by Rust and the versioned FFI protocol.
// Rust 与版本化 FFI 协议共享的稳定机器可读错误。
type EmbeddedInputEmbeddedErrorCode string

const (
	// The supplied configuration or request violates its declared contract.
	// 提供的配置或请求违反其声明的契约。
	EmbeddedInputEmbeddedErrorCodeInvalidArgument EmbeddedInputEmbeddedErrorCode = "invalid_argument"
	// The requested identity does not exist or its retained record has expired.
	// 请求的身份不存在，或其保留记录已经过期。
	EmbeddedInputEmbeddedErrorCodeNotFound EmbeddedInputEmbeddedErrorCode = "not_found"
	// The caller supplied an identity from an inactive generation.
	// 调用方提供了非活动代次的身份。
	EmbeddedInputEmbeddedErrorCodeStaleGeneration EmbeddedInputEmbeddedErrorCode = "stale_generation"
	// A bounded resource cannot admit more work.
	// 有界资源无法接纳更多工作。
	EmbeddedInputEmbeddedErrorCodeCapacityExceeded EmbeddedInputEmbeddedErrorCode = "capacity_exceeded"
	// The requested state transition conflicts with live work.
	// 请求的状态变更与仍在运行的工作冲突。
	EmbeddedInputEmbeddedErrorCodeBusy EmbeddedInputEmbeddedErrorCode = "busy"
	// This exact request already has a completion owner or retained terminal result.
	// 此精确请求已具有完成所有者或保留的终态结果。
	EmbeddedInputEmbeddedErrorCodeAlreadyCompleted EmbeddedInputEmbeddedErrorCode = "already_completed"
	// The runtime or registration no longer accepts work.
	// 运行时或注册项已停止接纳工作。
	EmbeddedInputEmbeddedErrorCodeClosed EmbeddedInputEmbeddedErrorCode = "closed"
	// The caller requested cooperative cancellation.
	// 调用方请求了协作取消。
	EmbeddedInputEmbeddedErrorCodeCancelled EmbeddedInputEmbeddedErrorCode = "cancelled"
	// The original end-to-end execution budget expired.
	// 原始端到端执行预算已经耗尽。
	EmbeddedInputEmbeddedErrorCodeDeadlineExceeded EmbeddedInputEmbeddedErrorCode = "deadline_exceeded"
	// The trusted host did not grant the requested capability.
	// 可信宿主未授予所请求的能力。
	EmbeddedInputEmbeddedErrorCodePermissionDenied EmbeddedInputEmbeddedErrorCode = "permission_denied"
	// The requested backend or protocol feature is unavailable.
	// 请求的后端或协议功能不可用。
	EmbeddedInputEmbeddedErrorCodeUnsupported EmbeddedInputEmbeddedErrorCode = "unsupported"
	// Plugin initialization or execution failed.
	// 插件初始化或执行失败。
	EmbeddedInputEmbeddedErrorCodeExecutionFailed EmbeddedInputEmbeddedErrorCode = "execution_failed"
	// Resource teardown failed and still owns its capacity.
	// 资源清理失败且仍占有容量。
	EmbeddedInputEmbeddedErrorCodeCleanupFailed EmbeddedInputEmbeddedErrorCode = "cleanup_failed"
	// An internal invariant failed; the caller must not retry mutations blindly.
	// 内部不变量失败，调用方不得盲目重试有副作用的操作。
	EmbeddedInputEmbeddedErrorCodeInternal EmbeddedInputEmbeddedErrorCode = "internal"
)

// Immutable host-approved aggregate budgets across every generation and execution domain of one plugin.
// 一个插件的全部代次与执行域共享的不可变宿主批准聚合预算。
type EmbeddedInputEmbeddedPluginConfig struct {
	// Maximum retained operations, including completed results not explicitly forgotten.
	// 保留操作的数量上限，包含尚未显式遗忘的已完成结果。
	MaxOperations uint64 `json:"max_operations"`
	// Maximum exact serialized queued request bytes across this plugin.
	// 此插件全部排队请求精确序列化字节数上限。
	MaxQueuedBytes uint64 `json:"max_queued_bytes"`
	// Maximum accepted queued calls across all domains and sessions.
	// 全部域和会话已接纳排队调用的合计上限。
	MaxQueuedCalls uint64 `json:"max_queued_calls"`
	// Maximum retained pool identities, including closed generations awaiting explicit removal.
	// 保留池身份的数量上限，包含等待显式移除的已关闭代次。
	MaxRegisteredPools uint64 `json:"max_registered_pools"`
	// Maximum actual resident VMs plus other domains' unused dedicated reservations.
	// 实际常驻 VM 与其他域未使用专用预留的合计上限。
	MaxResidentVms uint64 `json:"max_resident_vms"`
	// Maximum dispatched operations, retained through initialization, host waiting and cleanup.
	// 已分发操作上限，计费覆盖初始化、宿主等待及清理。
	MaxRunningCalls uint64 `json:"max_running_calls"`
	// Maximum retained sessions, including closed session records.
	// 保留会话的数量上限，包含已关闭会话记录。
	MaxSessions uint64 `json:"max_sessions"`
}

// Explicit parent budgets; hosts resolve defaults once before construction.
// 显式父级预算；宿主在构造前一次性解析默认值。
type EmbeddedInputEmbeddedRuntimeConfig struct {
	// Maximum serialized module context and effect metadata bytes retained by one operation.
	// 单次操作保留的模块上下文及副作用元数据序列化字节上限。
	MaxEffectBytesPerOperation uint64 `json:"max_effect_bytes_per_operation"`
	// Maximum host effect records retained by one operation, including completed callbacks.
	// 单次操作保留的宿主副作用记录上限，包含已完成回调。
	MaxEffectRecordsPerOperation uint64 `json:"max_effect_records_per_operation"`
	// Maximum request bytes and reserved application output bytes, including dispatched calls.
	// 请求字节与预留应用输出字节上限，包含已分发调用。
	// Fixed protocol error metadata is separately bounded by the retained request count.
	// 固定协议错误元数据由保留请求数量独立约束。
	MaxHostRequestBytes uint64 `json:"max_host_request_bytes"`
	// Maximum pending host requests across all plugin instances.
	// 所有插件实例待完成宿主请求的数量上限。
	MaxHostRequests uint64 `json:"max_host_requests"`
	// Maximum operation records retained, including unfinished operations.
	// 操作记录保留数量上限，包含未完成操作。
	MaxOperations uint64 `json:"max_operations"`
	// Maximum serialized bytes retained by queued requests.
	// 排队请求保留的序列化字节数上限。
	MaxQueuedBytes uint64 `json:"max_queued_bytes"`
	// Maximum accepted requests waiting for an execution slot.
	// 等待执行许可的已接纳请求数量上限。
	MaxQueuedCalls uint64 `json:"max_queued_calls"`
	// Maximum retained capability registrations, including draining or unforgotten retired entries.
	// 能力注册保留上限，包含正在排空或尚未遗忘的已退役条目。
	MaxRegisteredCapabilities uint64 `json:"max_registered_capabilities"`
	// Maximum retained plugin registrations, including closed entries awaiting explicit removal.
	// 保留插件注册的数量上限，包含等待显式移除的已关闭条目。
	MaxRegisteredPlugins uint64 `json:"max_registered_plugins"`
	// Maximum concurrently registered pool declarations, including draining generations.
	// 同时注册的池声明数量上限，包含正在排空的代次。
	MaxRegisteredPools uint64 `json:"max_registered_pools"`
	// Maximum resident VMs, including creation and pending teardown.
	// 最大常驻 VM 数，包含创建中与等待清理的实例。
	MaxResidentVms uint64 `json:"max_resident_vms"`
	// Maximum concurrent calls, including calls waiting for host results.
	// 最大并发调用数，包含等待宿主结果的调用。
	MaxRunningCalls uint64 `json:"max_running_calls"`
	// Maximum retained session identities, including closed sessions awaiting explicit removal.
	// 保留会话身份的数量上限，包含等待显式移除的已关闭会话。
	MaxSessions uint64 `json:"max_sessions"`
	// Maximum serialized application value bytes; fixed protocol error metadata is separate.
	// 应用值的最大序列化字节数；固定协议错误元数据独立计算。
	MaxValueBytes uint64 `json:"max_value_bytes"`
}

// Declared backend; unavailable variants are rejected instead of downgraded.
// 声明的执行后端；不可用的取值直接拒绝，不降级。
type EmbeddedInputExecutionBackend string

const (
	// Execute in owned Lua VMs inside the current host process.
	// 在当前宿主进程内的受管 Lua VM 中执行。
	EmbeddedInputExecutionBackendInProcess EmbeddedInputExecutionBackend = "in_process"
	// Reserved protocol identity; no worker backend is advertised yet.
	// 预留的协议身份；目前尚未声明工作进程后端可用。
	EmbeddedInputExecutionBackendWorkerProcess EmbeddedInputExecutionBackend = "worker_process"
)

// Exact historical cursor; its fields come from the original durable record, not a newly opened runtime.
// 精确历史游标；字段来自原持久记录，不来自新打开的运行时。
type EmbeddedInputHistoryCursor struct {
	// Original operation identity within that namespace.
	// 该命名空间中的原始操作身份。
	OperationId string `json:"operation_id"`
	// Original core runtime namespace from the returned history record.
	// 返回历史记录中的原始核心运行时命名空间。
	RuntimeId string `json:"runtime_id"`
}

// Strict host completion shapes match CapabilityOutcome::to_json, preserving successful JSON null.
// 严格宿主完成形状匹配 CapabilityOutcome::to_json，保留成功 JSON 空值。
type EmbeddedInputHostCompletion interface {
	// embeddedVariantEmbeddedInputHostCompletion seals this generated union without invoking serialization hooks.
	// embeddedVariantEmbeddedInputHostCompletion 封闭此生成联合，不调用序列化钩子。
	embeddedVariantEmbeddedInputHostCompletion()
}

// embeddedVariantEmbeddedInputHostCompletion marks the exact EmbeddedInputHostCompletionVariant1 alternative.
// embeddedVariantEmbeddedInputHostCompletion 标识精确的 EmbeddedInputHostCompletionVariant1 分支。
func (EmbeddedInputHostCompletionVariant1) embeddedVariantEmbeddedInputHostCompletion() {}

// embeddedVariantEmbeddedInputHostCompletion marks the exact EmbeddedInputHostCompletionVariant2 alternative.
// embeddedVariantEmbeddedInputHostCompletion 标识精确的 EmbeddedInputHostCompletionVariant2 分支。
func (EmbeddedInputHostCompletionVariant2) embeddedVariantEmbeddedInputHostCompletion() {}

// Exactly the success shape; ok must be true and value remains required even when null.
// 精确成功形状；ok 必须为真，value 即使为空值也必须存在。
type EmbeddedInputHostCompletionVariant1 struct {
	// Actual host effect evidence.
	// 实际宿主副作用证据。
	Effects EmbeddedInputEffectState `json:"effects"`
	// Explicit success discriminator.
	// 显式成功判别。
	Ok bool `json:"ok"`
	// Actual application result.
	// 实际应用结果。
	Value any `json:"value"`
}

// Exactly the failure shape; ok must be false.
// 精确失败形状；ok 必须为假。
type EmbeddedInputHostCompletionVariant2 struct {
	// Actual host effect evidence, even if a commit preceded the error.
	// 实际宿主副作用证据，即使错误前已发生提交。
	Effects EmbeddedInputEffectState `json:"effects"`
	// Actual structured host error.
	// 实际结构化宿主错误。
	Error EmbeddedInputEmbeddedError `json:"error"`
	// Explicit failure discriminator.
	// 显式失败判别。
	Ok bool `json:"ok"`
}

// Resolution for one exact original host effect; neither registration nor caller identity can be supplied anew.
// 一个精确原宿主副作用的结论；不得重新提供注册或调用方身份。
type EmbeddedInputHostEffectReconciliation struct {
	// Exact effect identity from the original snapshot, in the same order as its original records.
	// 原始快照中的精确副作用身份，顺序与其原始记录相同。
	EffectId string `json:"effect_id"`
	// Proven final outcome of this original effect, never a retry's outcome.
	// 此原始副作用的已证实最终结果，绝非重试结果。
	Effects EmbeddedInputResolvedEffectState `json:"effects"`
	// Nonempty host audit or transaction-query reference; credentials and business payloads do not belong here.
	// 非空宿主审计或事务查询引用；此处不应包含凭证及业务载荷。
	Evidence string `json:"evidence"`
}

// Explicit module-state lifetime selected by a validated plugin contract.
// 由已校验插件契约选择的显式模块状态寿命。
type EmbeddedInputInstanceReuse string

const (
	// Destroy the instance after one invocation.
	// 一次调用后销毁实例。
	EmbeddedInputInstanceReuseSingleCall EmbeddedInputInstanceReuse = "single_call"
	// Reuse only within the same immutable generation and security partition.
	// 仅在相同不可变代次与安全分区内复用。
	EmbeddedInputInstanceReuseReusable EmbeddedInputInstanceReuse = "reusable"
	// Retain one instance for an explicitly opened session.
	// 为显式打开的会话保留一个固定实例。
	EmbeddedInputInstanceReuseSession EmbeddedInputInstanceReuse = "session"
)

// Construction options used by the host to create one LuaSkills runtime engine.
// 宿主创建单个 LuaSkills 运行时引擎时使用的构造选项。
type EmbeddedInputLuaEngineOptions struct {
	// Host-owned runtime paths and external library locations.
	// 宿主拥有的运行时路径与外部动态库位置配置。
	HostOptions EmbeddedInputLuaRuntimeHostOptions `json:"host_options"`
	// Pool sizing configuration for reusable Lua virtual machines.
	// 可复用 Lua 虚拟机池的容量配置。
	PoolConfig EmbeddedInputLuaVmPoolConfig `json:"pool_config"`
}

// Host-injected invocation context delivered alongside one skill or runlua call.
// 宿主在单次 skill 或 runlua 调用时一并注入的调用上下文。
type EmbeddedInputLuaInvocationContext struct {
	// Host-resolved client budget object injected into `vulcan.context.client_budget`.
	// 宿主解析后的客户端预算对象，将被注入到 `vulcan.context.client_budget`。
	ClientBudget any `json:"client_budget"`
	// Optional transport/request metadata preserved for Lua consumption.
	// 供 Lua 消费的可选传输层/请求层元数据。
	RequestContext **EmbeddedInputRuntimeRequestContext `json:"request_context,omitempty"`
	// Host-resolved tool configuration object injected into `vulcan.context.tool_config`.
	// 宿主解析后的工具配置对象，将被注入到 `vulcan.context.tool_config`。
	ToolConfig any `json:"tool_config"`
}

// Host-controlled toggles for optional Lua-exposed runtime bridges.
// 宿主控制的可选 Lua 暴露运行时桥接开关集合。
type EmbeddedInputLuaRuntimeCapabilityOptions struct {
	// Whether luaexec and runtime sessions replace Lua's global `io` table with managed IO.
	// luaexec 与持久运行时会话是否使用托管 IO 替换 Lua 全局 `io` 表。
	EnableManagedIoCompat bool `json:"enable_managed_io_compat"`
	// Whether `vulcan.runtime.skills.*` management bridges are exposed to Lua.
	// 是否将 `vulcan.runtime.skills.*` 管理桥接暴露给 Lua。
	EnableSkillManagementBridge *bool `json:"enable_skill_management_bridge,omitempty"`
}

// Callback transport mode used when the database provider mode is `host_callback`.
// 当数据库 provider 模式为 `host_callback` 时所使用的回调传输模式。
type EmbeddedInputLuaRuntimeDatabaseCallbackMode string

const (
	// The library uses the structured standard callback ABI.
	// 由库使用结构化标准回调 ABI。
	EmbeddedInputLuaRuntimeDatabaseCallbackModeStandard EmbeddedInputLuaRuntimeDatabaseCallbackMode = "standard"
	// The library uses the JSON callback ABI.
	// 由库使用 JSON 回调 ABI。
	EmbeddedInputLuaRuntimeDatabaseCallbackModeJson EmbeddedInputLuaRuntimeDatabaseCallbackMode = "json"
)

// Database access mode used by one host-facing runtime backend.
// 单个宿主侧运行时后端所使用的数据库访问模式。
type EmbeddedInputLuaRuntimeDatabaseProviderMode string

const (
	// The library loads and calls the local dynamic-library backend directly.
	// 由库直接加载并调用本地动态库后端。
	EmbeddedInputLuaRuntimeDatabaseProviderModeDynamicLibrary EmbeddedInputLuaRuntimeDatabaseProviderMode = "dynamic_library"
	// The library forwards database operations into one host-registered callback bridge.
	// 由库把数据库操作转发给宿主已注册的回调桥接。
	EmbeddedInputLuaRuntimeDatabaseProviderModeHostCallback EmbeddedInputLuaRuntimeDatabaseProviderMode = "host_callback"
	// The library forwards database operations into one external space controller.
	// 由库把数据库操作转发给外部空间控制器。
	EmbeddedInputLuaRuntimeDatabaseProviderModeSpaceController EmbeddedInputLuaRuntimeDatabaseProviderMode = "space_controller"
)

// Host-provided filesystem and runtime paths consumed by the LuaSkills library.
// 宿主提供给 LuaSkills 库消费的文件系统与运行时路径集合。
type EmbeddedInputLuaRuntimeHostOptions struct {
	// Whether the runtime is allowed to perform network downloads while installing dependencies.
	// 运行时在安装依赖时是否允许执行网络下载。
	AllowNetworkDownload bool `json:"allow_network_download"`
	// Host-provided transient cache policy consumed by `vulcan.cache`.
	// 由宿主提供并供 `vulcan.cache` 消费的临时缓存策略。
	CacheConfig **EmbeddedInputToolCacheConfig `json:"cache_config,omitempty"`
	// Host-controlled optional runtime capability toggles.
	// 由宿主控制的可选运行时能力开关集合。
	Capabilities *EmbeddedInputLuaRuntimeCapabilityOptions `json:"capabilities,omitempty"`
	// Fixed sibling directory name used under one skill-root parent to store skill databases.
	// 在单个技能根父目录下存放技能数据库时使用的固定兄弟目录名称。
	DatabaseDirName *string `json:"database_dir_name,omitempty"`
	// Optional default text encoding label used by managed IO and process APIs.
	// 托管 IO 与进程 API 使用的可选默认文本编码标签。
	DefaultTextEncoding **string `json:"default_text_encoding,omitempty"`
	// Fixed sibling directory name used under one skill-root parent to store dependencies.
	// 在单个技能根父目录下存放依赖时使用的固定兄弟目录名称。
	DependencyDirName *string `json:"dependency_dir_name,omitempty"`
	// Host-managed cache directory used for downloaded archives and remote manifests.
	// 宿主管理的下载缓存目录，用于归档文件和远程清单缓存。
	DownloadCacheRoot **string `json:"download_cache_root,omitempty"`
	// Whether trusted system operations may install from private URL manifests.
	// 可信 system 操作是否允许从私有 URL manifest 安装。
	EnablePrivateUrlSkillInstall *bool `json:"enable_private_url_skill_install,omitempty"`
	// Optional GitHub API base URL override used to resolve release metadata.
	// 可选的 GitHub API 基址覆盖，用于解析 release 元数据。
	GithubApiBaseUrl **string `json:"github_api_base_url,omitempty"`
	// Optional GitHub site base URL override used to rewrite browser download URLs.
	// 可选的 GitHub 站点基址覆盖，用于重写浏览器下载地址。
	GithubBaseUrl **string `json:"github_base_url,omitempty"`
	// Host-managed root directory used only to probe host-provided FFI/native dependencies.
	// 仅用于探测宿主提供 FFI/原生依赖的宿主管理根目录。
	HostProvidedFfiRoot **string `json:"host_provided_ffi_root,omitempty"`
	// Host-managed root directory used only to probe host-provided Lua package dependencies.
	// 仅用于探测宿主提供 Lua 包依赖的宿主管理根目录。
	HostProvidedLuaRoot **string `json:"host_provided_lua_root,omitempty"`
	// Host-managed root directory used only to probe host-provided tool dependencies.
	// 仅用于探测宿主提供工具依赖的宿主管理根目录。
	HostProvidedToolRoot **string `json:"host_provided_tool_root,omitempty"`
	// Host-forced skill identifiers that must be skipped before dependency or database setup.
	// 宿主强制跳过的技能标识符列表，会在依赖或数据库初始化前生效。
	IgnoredSkillIds *EmbeddedInputLuaRuntimeHostOptionsIgnoredSkillIds `json:"ignored_skill_ids,omitempty"`
	// LanceDB callback transport mode selected by the host when provider mode is `host_callback`.
	// 当 provider 模式为 `host_callback` 时，宿主为 LanceDB 选择的回调传输模式。
	LancedbCallbackMode *EmbeddedInputLuaRuntimeDatabaseCallbackMode `json:"lancedb_callback_mode,omitempty"`
	// Explicit LanceDB dynamic-library path owned by the host.
	// 由宿主显式提供的 LanceDB 动态库路径。
	LancedbLibraryPath **string `json:"lancedb_library_path,omitempty"`
	// LanceDB database provider mode selected by the host.
	// 宿主为 LanceDB 数据库选择的 provider 模式。
	LancedbProviderMode *EmbeddedInputLuaRuntimeDatabaseProviderMode `json:"lancedb_provider_mode,omitempty"`
	// Optional lua_packages root used to build `package.path` and `package.cpath`.
	// 用于拼接 `package.path` 与 `package.cpath` 的可选 lua_packages 根目录。
	LuaPackagesDir **string `json:"lua_packages_dir,omitempty"`
	// Host-selected managed Python/Node worker and persistent-session resource policy.
	// 宿主选择的受管 Python/Node Worker 与持久会话资源策略。
	ManagedRuntimeConfig *EmbeddedInputLuaRuntimeManagedRuntimeConfig `json:"managed_runtime_config,omitempty"`
	// Optional host-configured read-only root containing managed Python and Node distributions.
	// 可选的宿主配置只读根目录，包含受管 Python 与 Node 发行包。
	ManagedRuntimeDistributionRoot **string `json:"managed_runtime_distribution_root,omitempty"`
	// Optional host-configured writable root containing reusable managed environments.
	// 可选的宿主配置可写根目录，包含可复用受管环境。
	ManagedRuntimeEnvironmentRoot **string `json:"managed_runtime_environment_root,omitempty"`
	// Optional official LuaSkills Hub base URL used by managed Hub installs.
	// 受管 Hub 安装使用的可选官方 LuaSkills Hub 基址。
	OfficialSkillHubBaseUrl **string `json:"official_skill_hub_base_url,omitempty"`
	// Host-controlled URL prefixes allowed for private skill manifests.
	// 宿主管控的私有技能 manifest 允许 URL 前缀。
	PrivateSkillSourceAllowlist *EmbeddedInputLuaRuntimeHostOptionsPrivateSkillSourceAllowlist `json:"private_skill_source_allowlist,omitempty"`
	// Host-reserved public entry names that LuaSkills canonical name generation must never occupy directly.
	// 宿主保留的公开入口名称集合，LuaSkills 在生成 canonical 名称时必须直接避开这些名称。
	ReservedEntryNames EmbeddedInputLuaRuntimeHostOptionsReservedEntryNames `json:"reserved_entry_names"`
	// Optional host-managed resources directory exposed to Lua as `vulcan.runtime.resources_dir`.
	// 以 `vulcan.runtime.resources_dir` 形式暴露给 Lua 的可选宿主管理资源目录。
	ResourcesDir **string `json:"resources_dir,omitempty"`
	// Optional dedicated pool configuration for isolated `vulcan.runtime.lua.exec` VMs.
	// 供隔离 `vulcan.runtime.lua.exec` 虚拟机使用的可选独立池配置。
	RunluaPoolConfig **EmbeddedInputLuaRuntimeRunLuaPoolConfig `json:"runlua_pool_config,omitempty"`
	// Optional canonical LuaSkills runtime root used to derive the fixed runtime layout.
	// 用于推导固定运行时布局的可选规范 LuaSkills 运行时根目录。
	RuntimeRoot **string `json:"runtime_root,omitempty"`
	// Optional cross-process configuration lock timeout in milliseconds.
	// 可选的配置跨进程锁超时毫秒数。
	SkillConfigLockTimeoutMs **uint64 `json:"skill_config_lock_timeout_ms,omitempty"`
	// Explicit user-level root containing normal and system skill configuration stores.
	// 包含普通技能与系统技能配置存储的显式用户级根目录。
	SkillConfigRoot **string `json:"skill_config_root,omitempty"`
	// Optional configuration file watcher debounce interval in milliseconds.
	// 可选的配置文件监听防抖毫秒数。
	SkillConfigWatchDebounceMs **uint64 `json:"skill_config_watch_debounce_ms,omitempty"`
	// Shared controller client options used when one database backend selects `space_controller`.
	// 当数据库后端选择 `space_controller` 时所使用的共享控制器客户端选项。
	SpaceController *EmbeddedInputLuaRuntimeSpaceControllerOptions `json:"space_controller,omitempty"`
	// SQLite callback transport mode selected by the host when provider mode is `host_callback`.
	// 当 provider 模式为 `host_callback` 时，宿主为 SQLite 选择的回调传输模式。
	SqliteCallbackMode *EmbeddedInputLuaRuntimeDatabaseCallbackMode `json:"sqlite_callback_mode,omitempty"`
	// Explicit SQLite dynamic-library path owned by the host.
	// 由宿主显式提供的 SQLite 动态库路径。
	SqliteLibraryPath **string `json:"sqlite_library_path,omitempty"`
	// SQLite database provider mode selected by the host.
	// 宿主为 SQLite 数据库选择的 provider 模式。
	SqliteProviderMode *EmbeddedInputLuaRuntimeDatabaseProviderMode `json:"sqlite_provider_mode,omitempty"`
	// Fixed sibling directory name used under one skill-root parent to store skill state.
	// 在单个技能根父目录下存放技能状态时使用的固定兄弟目录名称。
	StateDirName *string `json:"state_dir_name,omitempty"`
	// Optional fixed host-owned system Lua library directory used by `system_lua_lib` leases.
	// 供 `system_lua_lib` 租约使用的可选固定宿主系统 Lua 库目录。
	SystemLuaLibDir **string `json:"system_lua_lib_dir,omitempty"`
	// Host-managed temporary directory used by luaexec spill files and similar transient artifacts.
	// 宿主管理的临时目录，供 luaexec 请求文件等短生命周期产物使用。
	TempDir **string `json:"temp_dir,omitempty"`
}

// Host-forced skill identifiers that must be skipped before dependency or database setup.
// 宿主强制跳过的技能标识符列表，会在依赖或数据库初始化前生效。
type EmbeddedInputLuaRuntimeHostOptionsIgnoredSkillIds []string

// Host-controlled URL prefixes allowed for private skill manifests.
// 宿主管控的私有技能 manifest 允许 URL 前缀。
type EmbeddedInputLuaRuntimeHostOptionsPrivateSkillSourceAllowlist []string

// Host-reserved public entry names that LuaSkills canonical name generation must never occupy directly.
// 宿主保留的公开入口名称集合，LuaSkills 在生成 canonical 名称时必须直接避开这些名称。
type EmbeddedInputLuaRuntimeHostOptionsReservedEntryNames []string

// Host-selected resource policy for managed Python and Node workers and persistent sessions.
// 宿主为受管 Python 与 Node Worker 及持久会话选择的资源策略。
type EmbeddedInputLuaRuntimeManagedRuntimeConfig struct {
	// Default positive invoke timeout in milliseconds; absent means unlimited.
	// 默认正数 invoke 超时毫秒数；缺失表示无限制。
	InvokeDefaultTimeoutMs **uint64 `json:"invoke_default_timeout_ms,omitempty"`
	// Default retained byte limit for each persistent-session stdout or stderr stream.
	// 每个持久会话 stdout 或 stderr 流默认保留的字节上限。
	PersistentSessionDefaultBufferLimitBytesPerStream uint64 `json:"persistent_session_default_buffer_limit_bytes_per_stream"`
	// Maximum launching or live persistent sessions retained by one engine.
	// 单个引擎允许保留的启动中或活动持久会话最大数量。
	PersistentSessionLimitPerEngine uint64 `json:"persistent_session_limit_per_engine"`
	// Idle seconds after which an unused worker may be retired.
	// 未使用 Worker 可被回收前的空闲秒数。
	WorkerIdleTtlSecs uint64 `json:"worker_idle_ttl_secs"`
	// Maximum live workers for one exact environment and package-owner pool key.
	// 单个精确环境与包所有者池键允许的最大活动 Worker 数量。
	WorkerPoolMaxSizePerEnvironment uint64 `json:"worker_pool_max_size_per_environment"`
}

// Host-provided pool configuration for isolated runlua VMs.
// 宿主提供的隔离 runlua 虚拟机池配置。
type EmbeddedInputLuaRuntimeRunLuaPoolConfig struct {
	// Idle TTL in seconds before one excess isolated runlua VM may be retired.
	// 多余隔离 runlua 虚拟机在空闲多少秒后允许回收。
	IdleTtlSecs uint64 `json:"idle_ttl_secs"`
	// Maximum number of isolated runlua VMs allowed in the pool.
	// 隔离 runlua 虚拟机池允许存在的最大数量。
	MaxSize uint64 `json:"max_size"`
	// Minimum number of isolated runlua VMs kept warm.
	// 隔离 runlua 虚拟机需要常驻保温的最小数量。
	MinSize uint64 `json:"min_size"`
}

// Host-provided controller client options used when one database backend chooses `space_controller`.
// 当数据库后端选择 `space_controller` 时使用的宿主侧控制器客户端选项。
type EmbeddedInputLuaRuntimeSpaceControllerOptions struct {
	// Whether the runtime may auto-spawn the controller when the endpoint is unavailable.
	// 当控制器端点不可用时，运行时是否允许自动唤起控制器。
	AutoSpawn bool `json:"auto_spawn"`
	// Transport connect timeout in seconds used by the controller client proxy.
	// 控制器客户端代理使用的传输连接超时秒数。
	ConnectTimeoutSecs uint64 `json:"connect_timeout_secs"`
	// Default lease TTL in seconds passed to one auto-spawned controller.
	// 传递给自动唤起控制器的默认租约 TTL 秒数。
	DefaultLeaseTtlSecs uint64 `json:"default_lease_ttl_secs"`
	// Optional explicit controller endpoint; when omitted the shared default endpoint is used.
	// 可选的显式控制器端点；缺失时使用共享默认端点。
	Endpoint **string `json:"endpoint,omitempty"`
	// Optional local executable path copied and managed by the host.
	// 由宿主复制并管理的可选本地可执行文件路径。
	ExecutablePath **string `json:"executable_path,omitempty"`
	// Idle timeout in seconds passed to one auto-spawned controller.
	// 传递给自动唤起控制器的空闲超时秒数。
	IdleTimeoutSecs uint64 `json:"idle_timeout_secs"`
	// Lease renew interval in seconds used by the background controller client task.
	// 后台控制器客户端任务使用的租约续约间隔秒数。
	LeaseRenewIntervalSecs uint64 `json:"lease_renew_interval_secs"`
	// Minimum uptime in seconds passed to one auto-spawned controller.
	// 传递给自动唤起控制器的最小存活秒数。
	MinimumUptimeSecs uint64 `json:"minimum_uptime_secs"`
	// Process mode used when auto-spawning one controller process.
	// 自动唤起控制器进程时使用的进程模式。
	ProcessMode EmbeddedInputLuaRuntimeSpaceControllerProcessMode `json:"process_mode"`
	// Startup retry interval in milliseconds used while polling one spawned controller.
	// 轮询已唤起控制器时使用的启动重试间隔毫秒数。
	StartupRetryIntervalMs uint64 `json:"startup_retry_interval_ms"`
	// Startup timeout in seconds used while waiting for one spawned controller to become ready.
	// 等待已唤起控制器就绪时使用的启动超时秒数。
	StartupTimeoutSecs uint64 `json:"startup_timeout_secs"`
}

// Process mode used when the runtime auto-spawns one local space controller process.
// 运行时自动拉起本地空间控制器进程时使用的进程模式。
type EmbeddedInputLuaRuntimeSpaceControllerProcessMode string

const (
	// Service mode keeps the controller process alive until an external stop happens.
	// Service 模式会让控制器进程持续存活，直到外部显式停止。
	EmbeddedInputLuaRuntimeSpaceControllerProcessModeService EmbeddedInputLuaRuntimeSpaceControllerProcessMode = "service"
	// Managed mode allows the controller process to stop itself after idle timeouts.
	// Managed 模式允许控制器进程在空闲超时后自行停止。
	EmbeddedInputLuaRuntimeSpaceControllerProcessModeManaged EmbeddedInputLuaRuntimeSpaceControllerProcessMode = "managed"
)

// Pool sizing configuration for Lua virtual machines.
// Lua 虚拟机池的容量配置。
type EmbeddedInputLuaVmPoolConfig struct {
	// Idle TTL in seconds before an excess VM can be retired.
	// 多余虚拟机在空闲多少秒后允许回收。
	IdleTtlSecs uint64 `json:"idle_ttl_secs"`
	// Maximum number of VMs allowed in the pool.
	// 池内允许存在的最大虚拟机数量。
	MaxSize uint64 `json:"max_size"`
	// Minimum number of VMs that should stay warm.
	// 需要常驻保温的最小虚拟机数量。
	MinSize uint64 `json:"min_size"`
}

// Immutable source and trusted path declaration supplied during module activation.
// 模块激活时提供的不可变源码与可信路径声明。
type EmbeddedInputModuleDefinition struct {
	// Logical working directory; absent selects the existing package-root rule.
	// 逻辑工作目录；省略时使用既有包根目录规则。
	Cwd **string `json:"cwd,omitempty"`
	// Package-relative dependency manifest, validated by the existing package loader.
	// 由既有包加载器校验的包相对依赖清单。
	DependenciesFile string `json:"dependencies_file"`
	// Exact public exports and value contracts validated before invocation.
	// 调用前校验的精确公开导出及值契约。
	Exports EmbeddedInputModuleDefinitionExports `json:"exports"`
	// Host-assigned immutable code and dependency generation.
	// 宿主分配的不可变代码与依赖代次。
	Generation string `json:"generation"`
	// Trusted mount metadata; must be a JSON object.
	// 可信挂载元数据，必须为 JSON 对象。
	Mounts any `json:"mounts"`
	// Absolute plugin root inside the configured System trust root.
	// 位于已配置 System 信任根内的绝对插件根目录。
	PackageRoot string `json:"package_root"`
	// Host-assigned stable plugin identity.
	// 宿主分配的稳定插件身份。
	PluginId string `json:"plugin_id"`
	// Host-authenticated security partition used for instance matching.
	// 用于实例匹配且由宿主认证的安全分区。
	SecurityPartition string `json:"security_partition"`
	// Source evaluated once; it must return a table of declared functions.
	// 仅求值一次的源码；必须返回已声明函数的表。
	Source string `json:"source"`
	// Explicitly authorized workspace root, absent for package-only execution.
	// 显式授权的工作区根目录；仅在包内执行时省略。
	WorkspaceRoot **string `json:"workspace_root,omitempty"`
}

// Exact public exports and value contracts validated before invocation.
// 调用前校验的精确公开导出及值契约。
type EmbeddedInputModuleDefinitionExports []EmbeddedInputModuleExport

// One named export with explicit input and output schemas.
// 具有显式输入及输出 Schema 的单个具名导出。
type EmbeddedInputModuleExport struct {
	// Offline Draft 2020-12 schema for structured invocation arguments.
	// 结构化调用参数的离线 Draft 2020-12 Schema。
	InputSchema any `json:"input_schema"`
	// Exact Lua table key captured when the module is loaded.
	// 模块加载时捕获的精确 Lua 表键。
	Name string `json:"name"`
	// Offline Draft 2020-12 schema for structured return values.
	// 结构化返回值的离线 Draft 2020-12 Schema。
	OutputSchema any `json:"output_schema"`
}

// Explicit retention budgets; SQLite journal/cache overhead is separate from the database-file cap.
// 显式保留预算；SQLite 日志及缓存开销与数据库文件上限分开计算。
type EmbeddedInputOperationJournalConfig struct {
	// Maximum main database bytes, rounded down to whole SQLite pages.
	// 主数据库最大字节数，向下取整至完整 SQLite 页。
	MaxDatabaseBytes uint64 `json:"max_database_bytes"`
	// Maximum UTF-8 JSON bytes for one complete stored record, including identities and revision.
	// 单条完整存储记录的最大 UTF-8 JSON 字节数，包含身份及修订号。
	MaxRecordBytes uint64 `json:"max_record_bytes"`
	// Maximum retained operations across all runtime namespaces; no automatic eviction occurs.
	// 所有运行时命名空间合计保留的最大操作数；不自动淘汰。
	MaxRecords uint64 `json:"max_records"`
}

// Explicit budgets include queued, executing and caller-retained completed write receipts.
// 显式预算包含排队、执行中及调用方仍保留的已完成写入回执。
type EmbeddedInputOperationJournalWorkerConfig struct {
	// Cumulative JSON request bytes retained across all admitted write attempts.
	// 所有已接纳写入尝试合计保留的 JSON 请求字节数。
	MaxPendingBytes uint64 `json:"max_pending_bytes"`
	// Maximum admitted write attempts until their last actual receipt owner releases them.
	// 最后一个真实回执所有者释放之前，最多接纳的写入尝试数。
	MaxPendingWrites uint64 `json:"max_pending_writes"`
}

// One bounded, final, host-authored attestation covering execution closure and every retained effect.
// 一份有界、最终且由宿主编写的证明，覆盖执行关闭及每个保留副作用。
// This API does not authenticate the attestation; the embedding host must authorize the resolver and verify evidence.
// 此 API 不认证证明；嵌入宿主必须授权对账者并核验证据。
type EmbeddedInputOperationReconciliation struct {
	// Resolved aggregate covering both recorded callbacks and any other effects from the original Lua execution.
	// 已解决的聚合结论，覆盖记录回调及原 Lua 执行的其他副作用。
	Effects EmbeddedInputResolvedEffectState `json:"effects"`
	// Nonempty evidence reference proving owner closure and the whole operation's external-effect conclusion.
	// 非空证据引用，证明所有者关闭及整个操作的外部副作用结论。
	Evidence string `json:"evidence"`
	// Closure evidence consistent with the unchanged original execution phase.
	// 与未改变原执行阶段一致的关闭证据。
	Execution EmbeddedInputReconciledExecution `json:"execution"`
	// Exactly one resolution per original effect, preserving original order and known outcomes.
	// 每个原始副作用精确一个结论，保留原始顺序及已知结果。
	HostEffects EmbeddedInputOperationReconciliationHostEffects `json:"host_effects"`
	// Stable host-assigned resolution identity, retained unchanged across observation or storage retries.
	// 宿主分配的稳定对账身份，跨观测或存储重试保持不变。
	ResolutionId string `json:"resolution_id"`
	// Authorized host resolver identity, not a plugin-supplied authority claim or an authentication credential.
	// 已授权宿主对账者身份，不是插件提供的权限声明或认证凭证。
	Resolver string `json:"resolver"`
}

// Exactly one resolution per original effect, preserving original order and known outcomes.
// 每个原始副作用精确一个结论，保留原始顺序及已知结果。
type EmbeddedInputOperationReconciliationHostEffects []EmbeddedInputHostEffectReconciliation

// Immutable capacity policy for one host-assigned plugin execution group.
// 单个宿主分配的插件执行分组的不可变容量策略。
type EmbeddedInputPluginPoolConfig struct {
	// Requested execution backend, validated before activation.
	// 激活前校验的所请求执行后端。
	Backend EmbeddedInputExecutionBackend `json:"backend"`
	// Idle retirement threshold; absent explicitly disables idle retirement.
	// 空闲退役阈值；省略明确表示关闭空闲退役。
	IdleTtlMs **uint64 `json:"idle_ttl_ms,omitempty"`
	// Shared or dedicated ownership of resident capacity.
	// 常驻容量的公共或专用归属。
	Kind EmbeddedInputPoolKind `json:"kind"`
	// Maximum pending calls in this group.
	// 当前分组等待调用的数量上限。
	MaxQueuedCalls uint64 `json:"max_queued_calls"`
	// Maximum resident instances in this exact immutable execution domain.
	// 此精确不可变执行域的最大常驻实例数。
	MaxResidentVms uint64 `json:"max_resident_vms"`
	// Maximum simultaneously executing calls in this group.
	// 当前分组同时执行的调用数上限。
	MaxRunningCalls uint64 `json:"max_running_calls"`
	// Maximum successful uses before retirement; absent disables this limit.
	// 退役前成功使用次数上限；省略表示关闭此上限。
	MaxUses **uint64 `json:"max_uses,omitempty"`
	// Non-lendable reservation; only dedicated groups may reserve capacity.
	// 不可出借的预留；仅专用分组可以预留容量。
	MinResidentVms uint64 `json:"min_resident_vms"`
	// Module state lifetime; legacy stateless calls select single-call mode.
	// 模块状态寿命；旧无状态调用选择单次模式。
	Reuse EmbeddedInputInstanceReuse `json:"reuse"`
	// Whether all calls in this group require FIFO serialization.
	// 当前分组的全部调用是否要求先进先出的串行执行。
	Serial bool `json:"serial"`
}

// Capacity ownership, separate from instance reuse and ordering requirements.
// 容量归属，与实例复用及顺序要求相互独立。
type EmbeddedInputPoolKind string

const (
	// Capacity shared across independently keyed plugin instances.
	// 在具有独立匹配键的插件实例之间共享容量。
	EmbeddedInputPoolKindShared EmbeddedInputPoolKind = "shared"
	// Host-approved capacity with a non-lendable minimum reservation.
	// 宿主批准且具有不可出借最小预留的容量。
	EmbeddedInputPoolKindDedicated EmbeddedInputPoolKind = "dedicated"
)

// Historical execution closure asserted by the trusted host after actual owners have stopped.
// 实际所有者停止后，由可信宿主断言的历史执行关闭。
type EmbeddedInputReconciledExecution string

const (
	// The unchanged original snapshot already contains an observed terminal execution result.
	// 未改变的原始快照已包含观测到的终态执行结果。
	EmbeddedInputReconciledExecutionObservedTerminal EmbeddedInputReconciledExecution = "observed_terminal"
	// Actual owners stopped without a durable terminal result; the original nonterminal snapshot stays unchanged.
	// 实际所有者停止但没有持久终态结果；原非终态快照保持不变。
	EmbeddedInputReconciledExecutionStoppedWithoutResult EmbeddedInputReconciledExecution = "stopped_without_result"
)

// Strict versioned request; unknown fields and commands are explicit protocol errors.
// 严格版本化请求；未知字段与命令是明确协议错误。
type EmbeddedInputRequest struct {
	// One typed operation; no legacy envelope aliases are inferred.
	// 一个类型化操作；不推断旧信封别名。
	Command EmbeddedInputCommand `json:"command"`
	// Explicit wire version, checked before dispatch.
	// 分发前检查的显式线协议版本。
	ProtocolVersion uint32 `json:"protocol_version"`
}

// Explicit resolved effects; unknown outcomes remain unreconciled instead of being coerced into success.
// 显式已解决副作用；未知结果保持未对账，不强制转为成功。
type EmbeddedInputResolvedEffectState string

const (
	// Trusted evidence proves the business effect never started.
	// 可信证据证明业务副作用从未开始。
	EmbeddedInputResolvedEffectStateNotStarted EmbeddedInputResolvedEffectState = "not_started"
	// Trusted evidence proves no externally visible mutation applies.
	// 可信证据证明不存在适用的外部可见变更。
	EmbeddedInputResolvedEffectStateNotApplicable EmbeddedInputResolvedEffectState = "not_applicable"
	// The original external transaction is proven committed.
	// 原外部事务已证实提交。
	EmbeddedInputResolvedEffectStateCommitted EmbeddedInputResolvedEffectState = "committed"
	// The original external transaction is proven rolled back.
	// 原外部事务已证实回滚。
	EmbeddedInputResolvedEffectStateRolledBack EmbeddedInputResolvedEffectState = "rolled_back"
)

// Generic host-side client identity information passed into the LuaSkills runtime.
// 传入 LuaSkills 运行时的通用宿主客户端身份信息。
type EmbeddedInputRuntimeClientInfo struct {
	// Stable host-defined client kind, such as `mcp`, `ide`, or `desktop`.
	// 宿主定义的稳定客户端类型，例如 `mcp`、`ide` 或 `desktop`。
	Kind **string `json:"kind,omitempty"`
	// Human-readable client name reported by the host.
	// 由宿主上报的人类可读客户端名称。
	Name **string `json:"name,omitempty"`
	// Optional client version string.
	// 可选的客户端版本字符串。
	Version **string `json:"version,omitempty"`
}

// Typed commands for one already initialized runtime; every identity stays bound to that runtime.
// 一个已初始化运行时的类型化命令；每个身份始终绑定该运行时。
type EmbeddedInputRuntimeCommand interface {
	// embeddedVariantEmbeddedInputRuntimeCommand seals this generated union without invoking serialization hooks.
	// embeddedVariantEmbeddedInputRuntimeCommand 封闭此生成联合，不调用序列化钩子。
	embeddedVariantEmbeddedInputRuntimeCommand()
}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandOperationPersistenceFailure alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandOperationPersistenceFailure 分支。
func (EmbeddedInputRuntimeCommandOperationPersistenceFailure) embeddedVariantEmbeddedInputRuntimeCommand() {
}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandOperationRetryCheckpoint alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandOperationRetryCheckpoint 分支。
func (EmbeddedInputRuntimeCommandOperationRetryCheckpoint) embeddedVariantEmbeddedInputRuntimeCommand() {
}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandStorageStatus alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandStorageStatus 分支。
func (EmbeddedInputRuntimeCommandStorageStatus) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandStorageRecover alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandStorageRecover 分支。
func (EmbeddedInputRuntimeCommandStorageRecover) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandHistoryGet alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandHistoryGet 分支。
func (EmbeddedInputRuntimeCommandHistoryGet) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandHistoryNext alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandHistoryNext 分支。
func (EmbeddedInputRuntimeCommandHistoryNext) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandHistoryReconcile alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandHistoryReconcile 分支。
func (EmbeddedInputRuntimeCommandHistoryReconcile) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandHistoryForget alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandHistoryForget 分支。
func (EmbeddedInputRuntimeCommandHistoryForget) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandPluginRegister alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandPluginRegister 分支。
func (EmbeddedInputRuntimeCommandPluginRegister) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandPluginStatus alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandPluginStatus 分支。
func (EmbeddedInputRuntimeCommandPluginStatus) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandPluginClose alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandPluginClose 分支。
func (EmbeddedInputRuntimeCommandPluginClose) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandPluginForget alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandPluginForget 分支。
func (EmbeddedInputRuntimeCommandPluginForget) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandPoolRegister alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandPoolRegister 分支。
func (EmbeddedInputRuntimeCommandPoolRegister) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandPoolStatus alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandPoolStatus 分支。
func (EmbeddedInputRuntimeCommandPoolStatus) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandPoolClose alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandPoolClose 分支。
func (EmbeddedInputRuntimeCommandPoolClose) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandPoolForget alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandPoolForget 分支。
func (EmbeddedInputRuntimeCommandPoolForget) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandPoolRevokePermission alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandPoolRevokePermission 分支。
func (EmbeddedInputRuntimeCommandPoolRevokePermission) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandCallSubmit alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandCallSubmit 分支。
func (EmbeddedInputRuntimeCommandCallSubmit) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandSessionOpen alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandSessionOpen 分支。
func (EmbeddedInputRuntimeCommandSessionOpen) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandSessionSubmit alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandSessionSubmit 分支。
func (EmbeddedInputRuntimeCommandSessionSubmit) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandSessionStatus alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandSessionStatus 分支。
func (EmbeddedInputRuntimeCommandSessionStatus) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandSessionClose alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandSessionClose 分支。
func (EmbeddedInputRuntimeCommandSessionClose) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandSessionForget alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandSessionForget 分支。
func (EmbeddedInputRuntimeCommandSessionForget) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandOperationStatus alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandOperationStatus 分支。
func (EmbeddedInputRuntimeCommandOperationStatus) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandOperationWait alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandOperationWait 分支。
func (EmbeddedInputRuntimeCommandOperationWait) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandOperationCancel alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandOperationCancel 分支。
func (EmbeddedInputRuntimeCommandOperationCancel) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandOperationForget alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandOperationForget 分支。
func (EmbeddedInputRuntimeCommandOperationForget) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandCapabilitiesRegister alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandCapabilitiesRegister 分支。
func (EmbeddedInputRuntimeCommandCapabilitiesRegister) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandCapabilitiesList alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandCapabilitiesList 分支。
func (EmbeddedInputRuntimeCommandCapabilitiesList) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandCapabilityStatus alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandCapabilityStatus 分支。
func (EmbeddedInputRuntimeCommandCapabilityStatus) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandCapabilityUnregister alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandCapabilityUnregister 分支。
func (EmbeddedInputRuntimeCommandCapabilityUnregister) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandCapabilityForget alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandCapabilityForget 分支。
func (EmbeddedInputRuntimeCommandCapabilityForget) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandHostRequestsTake alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandHostRequestsTake 分支。
func (EmbeddedInputRuntimeCommandHostRequestsTake) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandHostRequestStatus alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandHostRequestStatus 分支。
func (EmbeddedInputRuntimeCommandHostRequestStatus) embeddedVariantEmbeddedInputRuntimeCommand() {}

// embeddedVariantEmbeddedInputRuntimeCommand marks the exact EmbeddedInputRuntimeCommandHostRequestComplete alternative.
// embeddedVariantEmbeddedInputRuntimeCommand 标识精确的 EmbeddedInputRuntimeCommandHostRequestComplete 分支。
func (EmbeddedInputRuntimeCommandHostRequestComplete) embeddedVariantEmbeddedInputRuntimeCommand() {}

// Admit an asynchronous ordinary invocation.
// 接纳异步普通调用。
type EmbeddedInputRuntimeCommandCallSubmit struct {
	// Typed ordinary call bound to one exact pool.
	// 绑定一个精确池的类型化普通调用。
	Call EmbeddedInputEmbeddedCall `json:"call"`
	// Original end-to-end execution budget in milliseconds.
	// 原始端到端执行预算毫秒数。
	TimeoutMs uint64 `json:"timeout_ms"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandCallSubmitType `json:"type"`
}

// EmbeddedInputRuntimeCommandCallSubmitType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandCallSubmitType 从包内线契约派生。
type EmbeddedInputRuntimeCommandCallSubmitType string

const (
	// EmbeddedInputRuntimeCommandCallSubmitTypeCallSubmit is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandCallSubmitTypeCallSubmit 从包内线契约派生。
	EmbeddedInputRuntimeCommandCallSubmitTypeCallSubmit EmbeddedInputRuntimeCommandCallSubmitType = "call_submit"
)

// List declarations authorized by explicit host grants.
// 列出显式宿主授权允许的声明。
type EmbeddedInputRuntimeCommandCapabilitiesList struct {
	// Explicit host grants for this binding or discovery request.
	// 此绑定或发现请求的显式宿主授权。
	Permissions EmbeddedInputRuntimeCommandCapabilitiesListPermissions `json:"permissions"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandCapabilitiesListType `json:"type"`
}

// Explicit host grants for this binding or discovery request.
// 此绑定或发现请求的显式宿主授权。
type EmbeddedInputRuntimeCommandCapabilitiesListPermissions []string

// EmbeddedInputRuntimeCommandCapabilitiesListType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandCapabilitiesListType 从包内线契约派生。
type EmbeddedInputRuntimeCommandCapabilitiesListType string

const (
	// EmbeddedInputRuntimeCommandCapabilitiesListTypeCapabilitiesList is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandCapabilitiesListTypeCapabilitiesList 从包内线契约派生。
	EmbeddedInputRuntimeCommandCapabilitiesListTypeCapabilitiesList EmbeddedInputRuntimeCommandCapabilitiesListType = "capabilities_list"
)

// Publish a queued capability batch atomically.
// 原子发布队列能力批次。
type EmbeddedInputRuntimeCommandCapabilitiesRegister struct {
	// Batch of explicit queued capability declarations.
	// 显式队列能力声明批次。
	Descriptors EmbeddedInputRuntimeCommandCapabilitiesRegisterDescriptors `json:"descriptors"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandCapabilitiesRegisterType `json:"type"`
}

// Batch of explicit queued capability declarations.
// 显式队列能力声明批次。
type EmbeddedInputRuntimeCommandCapabilitiesRegisterDescriptors []EmbeddedInputCapabilityDescriptor

// EmbeddedInputRuntimeCommandCapabilitiesRegisterType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandCapabilitiesRegisterType 从包内线契约派生。
type EmbeddedInputRuntimeCommandCapabilitiesRegisterType string

const (
	// EmbeddedInputRuntimeCommandCapabilitiesRegisterTypeCapabilitiesRegister is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandCapabilitiesRegisterTypeCapabilitiesRegister 从包内线契约派生。
	EmbeddedInputRuntimeCommandCapabilitiesRegisterTypeCapabilitiesRegister EmbeddedInputRuntimeCommandCapabilitiesRegisterType = "capabilities_register"
)

// Forget only a drained registration.
// 仅遗忘已排空注册。
type EmbeddedInputRuntimeCommandCapabilityForget struct {
	// Exact capability registration identity.
	// 精确能力注册身份。
	RegistrationId string `json:"registration_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandCapabilityForgetType `json:"type"`
}

// EmbeddedInputRuntimeCommandCapabilityForgetType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandCapabilityForgetType 从包内线契约派生。
type EmbeddedInputRuntimeCommandCapabilityForgetType string

const (
	// EmbeddedInputRuntimeCommandCapabilityForgetTypeCapabilityForget is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandCapabilityForgetTypeCapabilityForget 从包内线契约派生。
	EmbeddedInputRuntimeCommandCapabilityForgetTypeCapabilityForget EmbeddedInputRuntimeCommandCapabilityForgetType = "capability_forget"
)

// Read actual callback registration lifetime.
// 读取实际回调注册寿命。
type EmbeddedInputRuntimeCommandCapabilityStatus struct {
	// Exact capability registration identity.
	// 精确能力注册身份。
	RegistrationId string `json:"registration_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandCapabilityStatusType `json:"type"`
}

// EmbeddedInputRuntimeCommandCapabilityStatusType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandCapabilityStatusType 从包内线契约派生。
type EmbeddedInputRuntimeCommandCapabilityStatusType string

const (
	// EmbeddedInputRuntimeCommandCapabilityStatusTypeCapabilityStatus is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandCapabilityStatusTypeCapabilityStatus 从包内线契约派生。
	EmbeddedInputRuntimeCommandCapabilityStatusTypeCapabilityStatus EmbeddedInputRuntimeCommandCapabilityStatusType = "capability_status"
)

// Retire one exact registration without rerouting existing calls.
// 退役一个精确注册，不重定向既有调用。
type EmbeddedInputRuntimeCommandCapabilityUnregister struct {
	// Exact capability registration identity.
	// 精确能力注册身份。
	RegistrationId string `json:"registration_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandCapabilityUnregisterType `json:"type"`
}

// EmbeddedInputRuntimeCommandCapabilityUnregisterType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandCapabilityUnregisterType 从包内线契约派生。
type EmbeddedInputRuntimeCommandCapabilityUnregisterType string

const (
	// EmbeddedInputRuntimeCommandCapabilityUnregisterTypeCapabilityUnregister is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandCapabilityUnregisterTypeCapabilityUnregister 从包内线契约派生。
	EmbeddedInputRuntimeCommandCapabilityUnregisterTypeCapabilityUnregister EmbeddedInputRuntimeCommandCapabilityUnregisterType = "capability_unregister"
)

// Forget reconciled history only after any matching live runtime operation has been explicitly forgotten.
// 仅在显式遗忘任何匹配的活动运行时操作后，遗忘已对账历史。
type EmbeddedInputRuntimeCommandHistoryForget struct {
	// Positive original revision required for atomic compare-and-swap removal.
	// 原子比较交换删除所需的原始正修订号。
	ExpectedRevision uint64 `json:"expected_revision"`
	// Original historical runtime namespace.
	// 原始历史运行时命名空间。
	HistoryRuntimeId string `json:"history_runtime_id"`
	// Exact original operation identity.
	// 精确原始操作身份。
	OperationId string `json:"operation_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandHistoryForgetType `json:"type"`
}

// EmbeddedInputRuntimeCommandHistoryForgetType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandHistoryForgetType 从包内线契约派生。
type EmbeddedInputRuntimeCommandHistoryForgetType string

const (
	// EmbeddedInputRuntimeCommandHistoryForgetTypeHistoryForget is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandHistoryForgetTypeHistoryForget 从包内线契约派生。
	EmbeddedInputRuntimeCommandHistoryForgetTypeHistoryForget EmbeddedInputRuntimeCommandHistoryForgetType = "history_forget"
)

// Read historical evidence by its original namespace, without adopting it as a live operation.
// 按原命名空间读取历史证据，不将其接管为活动操作。
type EmbeddedInputRuntimeCommandHistoryGet struct {
	// Original core runtime namespace, distinct from the containing FFI slot identity.
	// 原核心运行时命名空间，区别于外层 FFI 槽身份。
	HistoryRuntimeId string `json:"history_runtime_id"`
	// Exact original operation identity.
	// 精确原始操作身份。
	OperationId string `json:"operation_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandHistoryGetType `json:"type"`
}

// EmbeddedInputRuntimeCommandHistoryGetType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandHistoryGetType 从包内线契约派生。
type EmbeddedInputRuntimeCommandHistoryGetType string

const (
	// EmbeddedInputRuntimeCommandHistoryGetTypeHistoryGet is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandHistoryGetTypeHistoryGet 从包内线契约派生。
	EmbeddedInputRuntimeCommandHistoryGetTypeHistoryGet EmbeddedInputRuntimeCommandHistoryGetType = "history_get"
)

// Read at most one historical row after an explicit cursor; absence starts enumeration.
// 在显式游标后至多读取一条历史；缺失表示开始枚举。
type EmbeddedInputRuntimeCommandHistoryNext struct {
	// Original history key returned by a prior row, with no inferred current-runtime substitution.
	// 前一行返回的原始历史键，不推断替换为当前运行时。
	After **EmbeddedInputHistoryCursor `json:"after,omitempty"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandHistoryNextType `json:"type"`
}

// EmbeddedInputRuntimeCommandHistoryNextType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandHistoryNextType 从包内线契约派生。
type EmbeddedInputRuntimeCommandHistoryNextType string

const (
	// EmbeddedInputRuntimeCommandHistoryNextTypeHistoryNext is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandHistoryNextTypeHistoryNext 从包内线契约派生。
	EmbeddedInputRuntimeCommandHistoryNextTypeHistoryNext EmbeddedInputRuntimeCommandHistoryNextType = "history_next"
)

// Attach final trusted-host evidence after all original execution owners have stopped; never replay execution.
// 全部原执行所有者停止后附加最终可信宿主证据；绝不重放执行。
type EmbeddedInputRuntimeCommandHistoryReconcile struct {
	// Positive original revision; exact retries must retain this predecessor and all resolution fields.
	// 原始正修订号；精确重试必须保留此前驱及全部对账字段。
	ExpectedRevision uint64 `json:"expected_revision"`
	// Original historical runtime namespace.
	// 原始历史运行时命名空间。
	HistoryRuntimeId string `json:"history_runtime_id"`
	// Exact original operation identity.
	// 精确原始操作身份。
	OperationId string `json:"operation_id"`
	// Complete host-authorized evidence; this API does not authenticate supplied resolver names.
	// 完整宿主授权证据；此 API 不认证所提供的对账者名称。
	Resolution EmbeddedInputOperationReconciliation `json:"resolution"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandHistoryReconcileType `json:"type"`
}

// EmbeddedInputRuntimeCommandHistoryReconcileType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandHistoryReconcileType 从包内线契约派生。
type EmbeddedInputRuntimeCommandHistoryReconcileType string

const (
	// EmbeddedInputRuntimeCommandHistoryReconcileTypeHistoryReconcile is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandHistoryReconcileTypeHistoryReconcile 从包内线契约派生。
	EmbeddedInputRuntimeCommandHistoryReconcileTypeHistoryReconcile EmbeddedInputRuntimeCommandHistoryReconcileType = "history_reconcile"
)

// Acknowledge actual host completion and preserve effect evidence.
// 确认实际宿主完成并保留副作用证据。
type EmbeddedInputRuntimeCommandHostRequestComplete struct {
	// Actual host result and effect evidence, including late completion.
	// 实际宿主结果与副作用证据，包含迟到完成。
	Outcome EmbeddedInputHostCompletion `json:"outcome"`
	// Exact host request identity to query or acknowledge.
	// 用于查询或确认的精确宿主请求身份。
	RequestId string `json:"request_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandHostRequestCompleteType `json:"type"`
}

// EmbeddedInputRuntimeCommandHostRequestCompleteType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandHostRequestCompleteType 从包内线契约派生。
type EmbeddedInputRuntimeCommandHostRequestCompleteType string

const (
	// EmbeddedInputRuntimeCommandHostRequestCompleteTypeHostRequestComplete is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandHostRequestCompleteTypeHostRequestComplete 从包内线契约派生。
	EmbeddedInputRuntimeCommandHostRequestCompleteTypeHostRequestComplete EmbeddedInputRuntimeCommandHostRequestCompleteType = "host_request_complete"
)

// Read cancellation while retaining actual handler ownership.
// 读取取消状态，同时保留实际处理器所有权。
type EmbeddedInputRuntimeCommandHostRequestStatus struct {
	// Exact host request identity to query or acknowledge.
	// 用于查询或确认的精确宿主请求身份。
	RequestId string `json:"request_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandHostRequestStatusType `json:"type"`
}

// EmbeddedInputRuntimeCommandHostRequestStatusType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandHostRequestStatusType 从包内线契约派生。
type EmbeddedInputRuntimeCommandHostRequestStatusType string

const (
	// EmbeddedInputRuntimeCommandHostRequestStatusTypeHostRequestStatus is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandHostRequestStatusTypeHostRequestStatus 从包内线契约派生。
	EmbeddedInputRuntimeCommandHostRequestStatusTypeHostRequestStatus EmbeddedInputRuntimeCommandHostRequestStatusType = "host_request_status"
)

// Deliver one bounded callback batch with pre-dispatch encoding.
// 通过分发前编码投递一个有界回调批次。
type EmbeddedInputRuntimeCommandHostRequestsTake struct {
	// Maximum host requests in this one bounded batch.
	// 此单个有界批次的宿主请求数量上限。
	Limit uint64 `json:"limit"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandHostRequestsTakeType `json:"type"`
}

// EmbeddedInputRuntimeCommandHostRequestsTakeType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandHostRequestsTakeType 从包内线契约派生。
type EmbeddedInputRuntimeCommandHostRequestsTakeType string

const (
	// EmbeddedInputRuntimeCommandHostRequestsTakeTypeHostRequestsTake is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandHostRequestsTakeTypeHostRequestsTake 从包内线契约派生。
	EmbeddedInputRuntimeCommandHostRequestsTakeTypeHostRequestsTake EmbeddedInputRuntimeCommandHostRequestsTakeType = "host_requests_take"
)

// Request cooperative cancellation without declaring completion.
// 请求协作取消，不宣称完成。
type EmbeddedInputRuntimeCommandOperationCancel struct {
	// Exact retained operation identity.
	// 精确保留操作身份。
	OperationId string `json:"operation_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandOperationCancelType `json:"type"`
}

// EmbeddedInputRuntimeCommandOperationCancelType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandOperationCancelType 从包内线契约派生。
type EmbeddedInputRuntimeCommandOperationCancelType string

const (
	// EmbeddedInputRuntimeCommandOperationCancelTypeOperationCancel is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandOperationCancelTypeOperationCancel 从包内线契约派生。
	EmbeddedInputRuntimeCommandOperationCancelTypeOperationCancel EmbeddedInputRuntimeCommandOperationCancelType = "operation_cancel"
)

// Forget retained terminal evidence explicitly.
// 显式遗忘保留的终态证据。
type EmbeddedInputRuntimeCommandOperationForget struct {
	// Exact retained operation identity.
	// 精确保留操作身份。
	OperationId string `json:"operation_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandOperationForgetType `json:"type"`
}

// EmbeddedInputRuntimeCommandOperationForgetType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandOperationForgetType 从包内线契约派生。
type EmbeddedInputRuntimeCommandOperationForgetType string

const (
	// EmbeddedInputRuntimeCommandOperationForgetTypeOperationForget is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandOperationForgetTypeOperationForget 从包内线契约派生。
	EmbeddedInputRuntimeCommandOperationForgetTypeOperationForget EmbeddedInputRuntimeCommandOperationForgetType = "operation_forget"
)

// Read retained checkpoint failure without disk I/O or retry.
// 读取保留检查点故障，不进行磁盘 I/O 或重试。
type EmbeddedInputRuntimeCommandOperationPersistenceFailure struct {
	// Exact live operation identity in this runtime.
	// 此运行时中的精确活动操作身份。
	OperationId string `json:"operation_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandOperationPersistenceFailureType `json:"type"`
}

// EmbeddedInputRuntimeCommandOperationPersistenceFailureType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandOperationPersistenceFailureType 从包内线契约派生。
type EmbeddedInputRuntimeCommandOperationPersistenceFailureType string

const (
	// EmbeddedInputRuntimeCommandOperationPersistenceFailureTypeOperationPersistenceFailure is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandOperationPersistenceFailureTypeOperationPersistenceFailure 从包内线契约派生。
	EmbeddedInputRuntimeCommandOperationPersistenceFailureTypeOperationPersistenceFailure EmbeddedInputRuntimeCommandOperationPersistenceFailureType = "operation_persistence_failure"
)

// Request one retry of the original immutable checkpoint, never another business execution.
// 请求重试原不可变检查点一次，绝不再次执行业务。
type EmbeddedInputRuntimeCommandOperationRetryCheckpoint struct {
	// Exact live operation identity retaining the failed candidate.
	// 保留失败候选的精确活动操作身份。
	OperationId string `json:"operation_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandOperationRetryCheckpointType `json:"type"`
}

// EmbeddedInputRuntimeCommandOperationRetryCheckpointType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandOperationRetryCheckpointType 从包内线契约派生。
type EmbeddedInputRuntimeCommandOperationRetryCheckpointType string

const (
	// EmbeddedInputRuntimeCommandOperationRetryCheckpointTypeOperationRetryCheckpoint is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandOperationRetryCheckpointTypeOperationRetryCheckpoint 从包内线契约派生。
	EmbeddedInputRuntimeCommandOperationRetryCheckpointTypeOperationRetryCheckpoint EmbeddedInputRuntimeCommandOperationRetryCheckpointType = "operation_retry_checkpoint"
)

// Read current operation outcome and effect evidence.
// 读取当前操作结果及副作用证据。
type EmbeddedInputRuntimeCommandOperationStatus struct {
	// Exact retained operation identity.
	// 精确保留操作身份。
	OperationId string `json:"operation_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandOperationStatusType `json:"type"`
}

// EmbeddedInputRuntimeCommandOperationStatusType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandOperationStatusType 从包内线契约派生。
type EmbeddedInputRuntimeCommandOperationStatusType string

const (
	// EmbeddedInputRuntimeCommandOperationStatusTypeOperationStatus is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandOperationStatusTypeOperationStatus 从包内线契约派生。
	EmbeddedInputRuntimeCommandOperationStatusTypeOperationStatus EmbeddedInputRuntimeCommandOperationStatusType = "operation_status"
)

// Wait for terminal state within an independent observer budget.
// 在独立观察者预算内等待终态。
type EmbeddedInputRuntimeCommandOperationWait struct {
	// Exact retained operation identity.
	// 精确保留操作身份。
	OperationId string `json:"operation_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandOperationWaitType `json:"type"`
	// Finite observer wait in milliseconds, independent of execution cancellation.
	// 有限观察者等待毫秒数，独立于执行取消。
	WaitMs uint64 `json:"wait_ms"`
}

// EmbeddedInputRuntimeCommandOperationWaitType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandOperationWaitType 从包内线契约派生。
type EmbeddedInputRuntimeCommandOperationWaitType string

const (
	// EmbeddedInputRuntimeCommandOperationWaitTypeOperationWait is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandOperationWaitTypeOperationWait 从包内线契约派生。
	EmbeddedInputRuntimeCommandOperationWaitTypeOperationWait EmbeddedInputRuntimeCommandOperationWaitType = "operation_wait"
)

// Close one plugin's admission and pools.
// 关闭一个插件的入场及池。
type EmbeddedInputRuntimeCommandPluginClose struct {
	// Exact host-assigned plugin identity.
	// 精确宿主分配的插件身份。
	PluginId string `json:"plugin_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandPluginCloseType `json:"type"`
}

// EmbeddedInputRuntimeCommandPluginCloseType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandPluginCloseType 从包内线契约派生。
type EmbeddedInputRuntimeCommandPluginCloseType string

const (
	// EmbeddedInputRuntimeCommandPluginCloseTypePluginClose is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandPluginCloseTypePluginClose 从包内线契约派生。
	EmbeddedInputRuntimeCommandPluginCloseTypePluginClose EmbeddedInputRuntimeCommandPluginCloseType = "plugin_close"
)

// Forget only a fully released plugin registration.
// 仅遗忘完全释放的插件注册。
type EmbeddedInputRuntimeCommandPluginForget struct {
	// Exact host-assigned plugin identity.
	// 精确宿主分配的插件身份。
	PluginId string `json:"plugin_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandPluginForgetType `json:"type"`
}

// EmbeddedInputRuntimeCommandPluginForgetType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandPluginForgetType 从包内线契约派生。
type EmbeddedInputRuntimeCommandPluginForgetType string

const (
	// EmbeddedInputRuntimeCommandPluginForgetTypePluginForget is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandPluginForgetTypePluginForget 从包内线契约派生。
	EmbeddedInputRuntimeCommandPluginForgetTypePluginForget EmbeddedInputRuntimeCommandPluginForgetType = "plugin_forget"
)

// Register aggregate plugin budgets.
// 注册插件聚合预算。
type EmbeddedInputRuntimeCommandPluginRegister struct {
	// Explicit aggregate plugin budgets.
	// 显式插件聚合预算。
	Config EmbeddedInputEmbeddedPluginConfig `json:"config"`
	// Exact host-assigned plugin identity.
	// 精确宿主分配的插件身份。
	PluginId string `json:"plugin_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandPluginRegisterType `json:"type"`
}

// EmbeddedInputRuntimeCommandPluginRegisterType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandPluginRegisterType 从包内线契约派生。
type EmbeddedInputRuntimeCommandPluginRegisterType string

const (
	// EmbeddedInputRuntimeCommandPluginRegisterTypePluginRegister is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandPluginRegisterTypePluginRegister 从包内线契约派生。
	EmbeddedInputRuntimeCommandPluginRegisterTypePluginRegister EmbeddedInputRuntimeCommandPluginRegisterType = "plugin_register"
)

// Query live aggregate plugin ownership.
// 查询实时插件聚合所有权。
type EmbeddedInputRuntimeCommandPluginStatus struct {
	// Exact host-assigned plugin identity.
	// 精确宿主分配的插件身份。
	PluginId string `json:"plugin_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandPluginStatusType `json:"type"`
}

// EmbeddedInputRuntimeCommandPluginStatusType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandPluginStatusType 从包内线契约派生。
type EmbeddedInputRuntimeCommandPluginStatusType string

const (
	// EmbeddedInputRuntimeCommandPluginStatusTypePluginStatus is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandPluginStatusTypePluginStatus 从包内线契约派生。
	EmbeddedInputRuntimeCommandPluginStatusTypePluginStatus EmbeddedInputRuntimeCommandPluginStatusType = "plugin_status"
)

// Close an exact pool and begin actual retirement.
// 关闭精确池并开始实际退役。
type EmbeddedInputRuntimeCommandPoolClose struct {
	// Exact immutable pool identity.
	// 精确不可变池身份。
	PoolId string `json:"pool_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandPoolCloseType `json:"type"`
}

// EmbeddedInputRuntimeCommandPoolCloseType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandPoolCloseType 从包内线契约派生。
type EmbeddedInputRuntimeCommandPoolCloseType string

const (
	// EmbeddedInputRuntimeCommandPoolCloseTypePoolClose is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandPoolCloseTypePoolClose 从包内线契约派生。
	EmbeddedInputRuntimeCommandPoolCloseTypePoolClose EmbeddedInputRuntimeCommandPoolCloseType = "pool_close"
)

// Forget only a pool whose ownership has drained.
// 仅遗忘所有权已排空的池。
type EmbeddedInputRuntimeCommandPoolForget struct {
	// Exact immutable pool identity.
	// 精确不可变池身份。
	PoolId string `json:"pool_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandPoolForgetType `json:"type"`
}

// EmbeddedInputRuntimeCommandPoolForgetType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandPoolForgetType 从包内线契约派生。
type EmbeddedInputRuntimeCommandPoolForgetType string

const (
	// EmbeddedInputRuntimeCommandPoolForgetTypePoolForget is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandPoolForgetTypePoolForget 从包内线契约派生。
	EmbeddedInputRuntimeCommandPoolForgetTypePoolForget EmbeddedInputRuntimeCommandPoolForgetType = "pool_forget"
)

// Register immutable source and capability authority without executing Lua.
// 注册不可变源码及能力权威，不执行 Lua。
type EmbeddedInputRuntimeCommandPoolRegister struct {
	// Immutable package and module declaration.
	// 不可变包与模块声明。
	Definition EmbeddedInputModuleDefinition `json:"definition"`
	// Immutable host initialization and configuration revision.
	// 不可变宿主初始化及配置修订。
	ExecutionRevision string `json:"execution_revision"`
	// Explicit host grants for this binding or discovery request.
	// 此绑定或发现请求的显式宿主授权。
	Permissions EmbeddedInputRuntimeCommandPoolRegisterPermissions `json:"permissions"`
	// Explicit immutable VM pool policy.
	// 显式不可变 VM 池策略。
	Policy EmbeddedInputPluginPoolConfig `json:"policy"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandPoolRegisterType `json:"type"`
}

// Explicit host grants for this binding or discovery request.
// 此绑定或发现请求的显式宿主授权。
type EmbeddedInputRuntimeCommandPoolRegisterPermissions []string

// EmbeddedInputRuntimeCommandPoolRegisterType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandPoolRegisterType 从包内线契约派生。
type EmbeddedInputRuntimeCommandPoolRegisterType string

const (
	// EmbeddedInputRuntimeCommandPoolRegisterTypePoolRegister is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandPoolRegisterTypePoolRegister 从包内线契约派生。
	EmbeddedInputRuntimeCommandPoolRegisterTypePoolRegister EmbeddedInputRuntimeCommandPoolRegisterType = "pool_register"
)

// Revoke a grant on the existing live binding.
// 撤销既有实时绑定上的授权。
type EmbeddedInputRuntimeCommandPoolRevokePermission struct {
	// Exact live permission to revoke.
	// 需要撤销的精确实时权限。
	Permission string `json:"permission"`
	// Exact immutable pool identity.
	// 精确不可变池身份。
	PoolId string `json:"pool_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandPoolRevokePermissionType `json:"type"`
}

// EmbeddedInputRuntimeCommandPoolRevokePermissionType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandPoolRevokePermissionType 从包内线契约派生。
type EmbeddedInputRuntimeCommandPoolRevokePermissionType string

const (
	// EmbeddedInputRuntimeCommandPoolRevokePermissionTypePoolRevokePermission is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandPoolRevokePermissionTypePoolRevokePermission 从包内线契约派生。
	EmbeddedInputRuntimeCommandPoolRevokePermissionTypePoolRevokePermission EmbeddedInputRuntimeCommandPoolRevokePermissionType = "pool_revoke_permission"
)

// Read actual resource accounting for an exact pool.
// 读取精确池的实际资源计数。
type EmbeddedInputRuntimeCommandPoolStatus struct {
	// Exact immutable pool identity.
	// 精确不可变池身份。
	PoolId string `json:"pool_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandPoolStatusType `json:"type"`
}

// EmbeddedInputRuntimeCommandPoolStatusType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandPoolStatusType 从包内线契约派生。
type EmbeddedInputRuntimeCommandPoolStatusType string

const (
	// EmbeddedInputRuntimeCommandPoolStatusTypePoolStatus is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandPoolStatusTypePoolStatus 从包内线契约派生。
	EmbeddedInputRuntimeCommandPoolStatusTypePoolStatus EmbeddedInputRuntimeCommandPoolStatusType = "pool_status"
)

// Close and cancel a fixed session without migrating its state.
// 关闭并取消固定会话，不迁移其状态。
type EmbeddedInputRuntimeCommandSessionClose struct {
	// Exact fixed-instance session identity.
	// 精确固定实例会话身份。
	SessionId string `json:"session_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandSessionCloseType `json:"type"`
}

// EmbeddedInputRuntimeCommandSessionCloseType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandSessionCloseType 从包内线契约派生。
type EmbeddedInputRuntimeCommandSessionCloseType string

const (
	// EmbeddedInputRuntimeCommandSessionCloseTypeSessionClose is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandSessionCloseTypeSessionClose 从包内线契约派生。
	EmbeddedInputRuntimeCommandSessionCloseTypeSessionClose EmbeddedInputRuntimeCommandSessionCloseType = "session_close"
)

// Forget only a closed session with no live ownership.
// 仅遗忘没有活动所有权的已关闭会话。
type EmbeddedInputRuntimeCommandSessionForget struct {
	// Exact fixed-instance session identity.
	// 精确固定实例会话身份。
	SessionId string `json:"session_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandSessionForgetType `json:"type"`
}

// EmbeddedInputRuntimeCommandSessionForgetType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandSessionForgetType 从包内线契约派生。
type EmbeddedInputRuntimeCommandSessionForgetType string

const (
	// EmbeddedInputRuntimeCommandSessionForgetTypeSessionForget is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandSessionForgetTypeSessionForget 从包内线契约派生。
	EmbeddedInputRuntimeCommandSessionForgetTypeSessionForget EmbeddedInputRuntimeCommandSessionForgetType = "session_forget"
)

// Reserve a fixed instance and submit initialization.
// 预留固定实例并提交初始化。
type EmbeddedInputRuntimeCommandSessionOpen struct {
	// Exact immutable pool identity.
	// 精确不可变池身份。
	PoolId string `json:"pool_id"`
	// Original end-to-end execution budget in milliseconds.
	// 原始端到端执行预算毫秒数。
	TimeoutMs uint64 `json:"timeout_ms"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandSessionOpenType `json:"type"`
}

// EmbeddedInputRuntimeCommandSessionOpenType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandSessionOpenType 从包内线契约派生。
type EmbeddedInputRuntimeCommandSessionOpenType string

const (
	// EmbeddedInputRuntimeCommandSessionOpenTypeSessionOpen is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandSessionOpenTypeSessionOpen 从包内线契约派生。
	EmbeddedInputRuntimeCommandSessionOpenTypeSessionOpen EmbeddedInputRuntimeCommandSessionOpenType = "session_open"
)

// Read actual session ownership and closure.
// 读取实际会话所有权及关闭状态。
type EmbeddedInputRuntimeCommandSessionStatus struct {
	// Exact fixed-instance session identity.
	// 精确固定实例会话身份。
	SessionId string `json:"session_id"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandSessionStatusType `json:"type"`
}

// EmbeddedInputRuntimeCommandSessionStatusType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandSessionStatusType 从包内线契约派生。
type EmbeddedInputRuntimeCommandSessionStatusType string

const (
	// EmbeddedInputRuntimeCommandSessionStatusTypeSessionStatus is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandSessionStatusTypeSessionStatus 从包内线契约派生。
	EmbeddedInputRuntimeCommandSessionStatusTypeSessionStatus EmbeddedInputRuntimeCommandSessionStatusType = "session_status"
)

// Submit work to an exact fixed session.
// 向精确固定会话提交工作。
type EmbeddedInputRuntimeCommandSessionSubmit struct {
	// Structured application arguments.
	// 结构化应用参数。
	Arguments any `json:"arguments"`
	// Trusted host invocation context.
	// 可信宿主调用上下文。
	Context EmbeddedInputLuaInvocationContext `json:"context"`
	// Declared module export name.
	// 已声明模块导出名称。
	Export string `json:"export"`
	// Exact fixed-instance session identity.
	// 精确固定实例会话身份。
	SessionId string `json:"session_id"`
	// Original end-to-end execution budget in milliseconds.
	// 原始端到端执行预算毫秒数。
	TimeoutMs uint64 `json:"timeout_ms"`
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandSessionSubmitType `json:"type"`
}

// EmbeddedInputRuntimeCommandSessionSubmitType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandSessionSubmitType 从包内线契约派生。
type EmbeddedInputRuntimeCommandSessionSubmitType string

const (
	// EmbeddedInputRuntimeCommandSessionSubmitTypeSessionSubmit is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandSessionSubmitTypeSessionSubmit 从包内线契约派生。
	EmbeddedInputRuntimeCommandSessionSubmitTypeSessionSubmit EmbeddedInputRuntimeCommandSessionSubmitType = "session_submit"
)

// Reopen and validate failed storage; this synchronous disk command belongs on a work lane.
// 重新打开并校验失败存储；此同步磁盘命令归入工作通道。
type EmbeddedInputRuntimeCommandStorageRecover struct {
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandStorageRecoverType `json:"type"`
}

// EmbeddedInputRuntimeCommandStorageRecoverType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandStorageRecoverType 从包内线契约派生。
type EmbeddedInputRuntimeCommandStorageRecoverType string

const (
	// EmbeddedInputRuntimeCommandStorageRecoverTypeStorageRecover is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandStorageRecoverTypeStorageRecover 从包内线契约派生。
	EmbeddedInputRuntimeCommandStorageRecoverTypeStorageRecover EmbeddedInputRuntimeCommandStorageRecoverType = "storage_recover"
)

// Read actual bounded writer ownership without waiting for disk.
// 读取真实有界写入者所有权，不等待磁盘。
type EmbeddedInputRuntimeCommandStorageStatus struct {
	// Type is derived from the packaged wire contract.
	// Type 从包内线契约派生。
	Type EmbeddedInputRuntimeCommandStorageStatusType `json:"type"`
}

// EmbeddedInputRuntimeCommandStorageStatusType is derived from the packaged wire contract.
// EmbeddedInputRuntimeCommandStorageStatusType 从包内线契约派生。
type EmbeddedInputRuntimeCommandStorageStatusType string

const (
	// EmbeddedInputRuntimeCommandStorageStatusTypeStorageStatus is derived from the packaged wire contract.
	// EmbeddedInputRuntimeCommandStorageStatusTypeStorageStatus 从包内线契约派生。
	EmbeddedInputRuntimeCommandStorageStatusTypeStorageStatus EmbeddedInputRuntimeCommandStorageStatusType = "storage_status"
)

// Explicit host storage selection; omitting this whole object selects the existing memory-only runtime.
// 显式宿主存储选择；省略整个对象表示选择既有纯内存运行时。
type EmbeddedInputRuntimePersistenceConfig struct {
	// Explicit durable retention limits, independent of transient runtime budgets.
	// 显式持久保留上限，独立于瞬态运行时预算。
	Journal EmbeddedInputOperationJournalConfig `json:"journal"`
	// Absolute database path owned and protected by the host, never a plugin-selected location.
	// 由宿主拥有和保护的绝对数据库路径，绝非插件选择位置。
	Path string `json:"path"`
	// Explicit bounded storage-thread receipt limits.
	// 显式有界存储线程回执上限。
	Worker EmbeddedInputOperationJournalWorkerConfig `json:"worker"`
}

// Generic request-scoped context injected by the host into one runtime invocation.
// 宿主在单次运行时调用中注入的通用请求级上下文。
type EmbeddedInputRuntimeRequestContext struct {
	// Optional host-provided raw client capabilities object.
	// 可选的宿主原始客户端能力对象。
	ClientCapabilities *any `json:"client_capabilities,omitempty"`
	// Optional host-side client metadata.
	// 可选的宿主客户端元数据。
	ClientInfo **EmbeddedInputRuntimeClientInfo `json:"client_info,omitempty"`
	// Optional host-defined client name for audit and cost attribution.
	// 可选的宿主客户端名称，用于审计和成本归因。
	ClientName **string `json:"client_name,omitempty"`
	// Optional host-defined request identifier for audit and cost attribution.
	// 可选的宿主请求标识符，用于审计和成本归因。
	RequestId **string `json:"request_id,omitempty"`
	// Optional host-defined session identifier.
	// 可选的宿主会话标识。
	SessionId **string `json:"session_id,omitempty"`
	// Optional host-defined transport name for the current request.
	// 当前请求的可选宿主传输层名称。
	TransportName **string `json:"transport_name,omitempty"`
}

// Runtime configuration for the shared tool cache, controlling capacity and expiration behavior.
// 共享工具缓存的运行时配置，控制容量与过期策略。
type EmbeddedInputToolCacheConfig struct {
	// Default TTL in seconds used when callers omit a TTL.
	// 默认 TTL（秒），调用方未传 TTL 时使用。
	DefaultTtlSecs uint64 `json:"default_ttl_secs"`
	// Maximum number of entries; oldest entries are evicted when the cache exceeds this size.
	// 缓存最大条目数，超出后会按创建顺序淘汰最旧条目。
	MaxEntries uint64 `json:"max_entries"`
	// Maximum TTL in seconds; requested TTL values are clamped to this ceiling.
	// 最大 TTL（秒），请求 TTL 会被限制在该范围内。
	MaxTtlSecs uint64 `json:"max_ttl_secs"`
}

// Host-authenticated caller data copied outside plugin-controlled arguments.
// 在插件可控参数之外复制的宿主认证调用方数据。
type EmbeddedOutputCapabilityCaller struct {
	// Immutable execution and initialization-configuration revision.
	// 不可变执行与初始化配置修订。
	ExecutionRevision string `json:"execution_revision"`
	// Exact operation whose original budget applies.
	// 适用原始预算的精确操作。
	OperationId string `json:"operation_id"`
	// Immutable package and dependency generation.
	// 不可变包与依赖代次。
	PackageGeneration string `json:"package_generation"`
	// Exact activated plugin identity.
	// 精确激活的插件身份。
	PluginId string `json:"plugin_id"`
	// Runtime namespace that owns the registration and operation.
	// 拥有注册及操作的运行时命名空间。
	RuntimeId string `json:"runtime_id"`
	// Trusted user/workspace partition, never copied from a Lua argument.
	// 可信用户与工作区分区，绝不从 Lua 参数复制。
	SecurityPartition string `json:"security_partition"`
	// Optional fixed session required by session-scoped capabilities.
	// 会话作用域能力要求的可选固定会话。
	SessionId *string `json:"session_id"`
	// Explicit authorized workspace; absent means package-only context.
	// 显式授权工作区；省略表示仅包内上下文。
	WorkspaceRoot *string `json:"workspace_root"`
}

// Immutable capability declaration shared by Rust, generated contracts and SDKs.
// Rust、生成契约与 SDK 共享的不可变能力声明。
type EmbeddedOutputCapabilityDescriptor struct {
	// English description of the host capability for tool consumers.
	// 面向工具消费者的宿主能力英文描述。
	Description string `json:"description"`
	// Declared mutation category.
	// 声明的变更类别。
	Effects EmbeddedOutputCapabilityEffects `json:"effects"`
	// Explicit native or queued dispatch protocol.
	// 显式原生或队列分发协议。
	Execution EmbeddedOutputCapabilityExecution `json:"execution"`
	// Explicit host deduplication contract.
	// 显式宿主去重契约。
	Idempotency EmbeddedOutputCapabilityIdempotency `json:"idempotency"`
	// Offline input value contract.
	// 离线输入值契约。
	InputSchema any `json:"input_schema"`
	// Per-capability budget capped by the original operation deadline.
	// 受原始操作截止时间约束的单能力预算。
	MaxCallMs uint64 `json:"max_call_ms"`
	// Maximum in-flight handlers, including cancelled handlers that have not stopped.
	// 在途处理器上限，包含已取消但尚未停止的处理器。
	MaxConcurrent uint64 `json:"max_concurrent"`
	// Maximum serialized input bytes within the parent value limit.
	// 父级值上限内的最大序列化输入字节数。
	MaxInputBytes uint64 `json:"max_input_bytes"`
	// Maximum serialized output bytes within the parent value limit.
	// 父级值上限内的最大序列化输出字节数。
	MaxOutputBytes uint64 `json:"max_output_bytes"`
	// Exact namespaced name; discovery exposes only authorized declarations.
	// 精确命名空间名称；发现操作仅暴露已授权声明。
	Name string `json:"name"`
	// Offline output value contract.
	// 离线输出值契约。
	OutputSchema any `json:"output_schema"`
	// Every listed grant must still exist at each admission boundary.
	// 每个入场边界仍必须拥有列出的全部授权。
	Permissions EmbeddedOutputCapabilityDescriptorPermissions `json:"permissions"`
	// Required trusted invocation scope.
	// 必需的可信调用作用域。
	Scope EmbeddedOutputCapabilityScope `json:"scope"`
	// Semantic interface version, independent from the core library version.
	// 语义接口版本，独立于核心库版本。
	Version string `json:"version"`
}

// Every listed grant must still exist at each admission boundary.
// 每个入场边界仍必须拥有列出的全部授权。
type EmbeddedOutputCapabilityDescriptorPermissions []string

// Declared effect category; it does not make external mutations transactional.
// 声明的副作用类别；它不会使外部变更自动具有事务性。
type EmbeddedOutputCapabilityEffects string

const (
	// Host contract promises no externally visible mutation.
	// 宿主契约承诺不产生外部可见变更。
	EmbeddedOutputCapabilityEffectsReadOnly EmbeddedOutputCapabilityEffects = "read_only"
	// Host must report actual commit, rollback or unknown status.
	// 宿主必须报告真实提交、回滚或未知状态。
	EmbeddedOutputCapabilityEffectsMutating EmbeddedOutputCapabilityEffects = "mutating"
)

// Host execution transport, explicitly selected before a capability is published.
// 宿主执行传输，在能力发布前显式选择。
type EmbeddedOutputCapabilityExecution string

const (
	// Short cooperative Rust callback executed on the owning VM thread.
	// 在所属 VM 线程执行的短时协作 Rust 回调。
	EmbeddedOutputCapabilityExecutionNative EmbeddedOutputCapabilityExecution = "native"
	// Reliable request consumed and completed by an SDK event pump.
	// 由 SDK 事件泵消费并完成的可靠请求。
	EmbeddedOutputCapabilityExecutionQueued EmbeddedOutputCapabilityExecution = "queued"
)

// Explicit side-effect deduplication support, never inferred from an operation identifier.
// 显式副作用去重支持，绝不从操作标识推断。
type EmbeddedOutputCapabilityIdempotency string

const (
	// Automatic replay is forbidden; an uncertain result needs host reconciliation.
	// 禁止自动重放；不确定结果需要宿主对账。
	EmbeddedOutputCapabilityIdempotencyNone EmbeddedOutputCapabilityIdempotency = "none"
	// Host implementation deduplicates the supplied request identity durably.
	// 宿主实现对提供的请求身份进行持久去重。
	EmbeddedOutputCapabilityIdempotencyHostRequest EmbeddedOutputCapabilityIdempotency = "host_request"
)

// Observable lifetime of an exact registration, including unregistration still draining calls.
// 精确注册的可观察生命周期，包含注销后仍在排空的调用。
type EmbeddedOutputCapabilityRegistrationStatus struct {
	// Whether new calls may enter this specific registration.
	// 新调用是否可以进入此特定注册。
	Accepting bool `json:"accepting"`
	// True only after admission closed and all native callback references were released.
	// 仅在入场关闭且全部原生回调引用释放后为真。
	Drained bool `json:"drained"`
	// Actually executing or dispatched handlers, including pending cancellation.
	// 实际执行或已分发的处理器，包含等待取消完成的处理器。
	InFlight uint64 `json:"in_flight"`
	// Exact declared capability name.
	// 精确声明的能力名称。
	Name string `json:"name"`
	// Opaque identity, never a lossy language number.
	// 不透明身份，绝不使用有精度损失的语言数值。
	RegistrationId string `json:"registration_id"`
}

// Scope required from the trusted caller, independent from Lua business arguments.
// 可信调用方必须具备的作用域，独立于 Lua 业务参数。
type EmbeddedOutputCapabilityScope string

const (
	// Available during an ordinary operation or session invocation.
	// 在普通操作或会话调用期间可用。
	EmbeddedOutputCapabilityScopeInvocation EmbeddedOutputCapabilityScope = "invocation"
	// Requires an explicitly bound session identity.
	// 要求显式绑定的会话身份。
	EmbeddedOutputCapabilityScopeSession EmbeddedOutputCapabilityScope = "session"
)

// Recovery state is independent of the operation's business phase and cancellation intent.
// 恢复状态独立于操作业务阶段及取消意愿。
type EmbeddedOutputCheckpointRetryState string

const (
	// No retry will run until the host explicitly requests one.
	// 宿主显式请求之前不会执行重试。
	EmbeddedOutputCheckpointRetryStateWaiting EmbeddedOutputCheckpointRetryState = "waiting"
	// A single retry request is retained for the original checkpoint owner.
	// 为原始检查点所有者保留了单次重试请求。
	EmbeddedOutputCheckpointRetryStateRequested EmbeddedOutputCheckpointRetryState = "requested"
	// The requested retry is being driven; repeated observations cannot create another attempt.
	// 正在推进已请求重试；重复观测不能创建另一次尝试。
	EmbeddedOutputCheckpointRetryStateRetrying EmbeddedOutputCheckpointRetryState = "retrying"
)

// Immutable description of the exact linked core, usable without a transport or runtime.
// 精确链接核心的不可变描述，无需传输或运行时即可使用。
type EmbeddedOutputCoreDescription struct {
	// Exact independent embedded ABI structure version.
	// 精确独立嵌入式 ABI 结构版本。
	AbiStructureVersion uint32 `json:"abi_structure_version"`
	// Build input evidence; release manifests bind it to commits and signed artifact checksums separately.
	// 构建输入证据；发布清单另将其关联到提交及签名产物摘要。
	Build EmbeddedOutputEmbeddedBuildIdentity `json:"build"`
	// Implemented semantic features; a name does not grant host permissions.
	// 已实现语义功能；名称不授予宿主权限。
	Capabilities EmbeddedOutputCoreDescriptionCapabilities `json:"capabilities"`
	// Root command names shared with the exhaustive dispatcher.
	// 与穷尽分发器共享的根命令名称。
	Commands EmbeddedOutputCoreDescriptionCommands `json:"commands"`
	// Cargo package version of this exact core.
	// 此精确核心的 Cargo 包版本。
	CoreVersion string `json:"core_version"`
	// Version of this independent descriptor format.
	// 此独立描述格式的版本。
	DescriptionVersion uint32 `json:"description_version"`
	// Actually implemented execution backends, excluding reserved unsupported variants.
	// 实际已实现执行后端，不包含预留且不支持的取值。
	ExecutionBackends EmbeddedOutputCoreDescriptionExecutionBackends `json:"execution_backends"`
	// Exact embedded JSON protocol version.
	// 精确嵌入式 JSON 协议版本。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Nested command names shared with the exhaustive dispatcher.
	// 与穷尽分发器共享的嵌套命令名称。
	RuntimeCommands EmbeddedOutputCoreDescriptionRuntimeCommands `json:"runtime_commands"`
}

// Implemented semantic features; a name does not grant host permissions.
// 已实现语义功能；名称不授予宿主权限。
type EmbeddedOutputCoreDescriptionCapabilities []string

// Root command names shared with the exhaustive dispatcher.
// 与穷尽分发器共享的根命令名称。
type EmbeddedOutputCoreDescriptionCommands []string

// Actually implemented execution backends, excluding reserved unsupported variants.
// 实际已实现执行后端，不包含预留且不支持的取值。
type EmbeddedOutputCoreDescriptionExecutionBackends []EmbeddedOutputExecutionBackend

// Nested command names shared with the exhaustive dispatcher.
// 与穷尽分发器共享的嵌套命令名称。
type EmbeddedOutputCoreDescriptionRuntimeCommands []string

// Host-reported effect outcome, independent from execution success or cancellation.
// 宿主报告的副作用结果，独立于执行成功或取消。
type EmbeddedOutputEffectState string

const (
	// No business execution has started.
	// 尚未开始业务执行。
	EmbeddedOutputEffectStateNotStarted EmbeddedOutputEffectState = "not_started"
	// The declared operation has no externally visible mutations.
	// 声明的操作不包含外部可见变更。
	EmbeddedOutputEffectStateNotApplicable EmbeddedOutputEffectState = "not_applicable"
	// A trusted host transaction explicitly confirmed its commit.
	// 可信宿主事务明确确认其提交。
	EmbeddedOutputEffectStateCommitted EmbeddedOutputEffectState = "committed"
	// A trusted host transaction explicitly confirmed its rollback.
	// 可信宿主事务明确确认其回滚。
	EmbeddedOutputEffectStateRolledBack EmbeddedOutputEffectState = "rolled_back"
	// Effects may have occurred; retries require host-specific reconciliation.
	// 副作用可能已发生；重试需要宿主特定的对账。
	EmbeddedOutputEffectStateUnknown EmbeddedOutputEffectState = "unknown"
)

// Selected package and compiler input identities; binary authentication remains the release artifact's job.
// 选定包及编译器输入身份；二进制认证仍由发布产物负责。
type EmbeddedOutputEmbeddedBuildIdentity struct {
	// Sorted Cargo feature environment suffixes; they are not reverse-mapped into guessed feature names.
	// 排序后 Cargo 功能环境后缀；不反向映射为猜测功能名。
	CargoFeatures EmbeddedOutputEmbeddedBuildIdentityCargoFeatures `json:"cargo_features"`
	// Exact bundled embedded contract identity, independently checked by SDKs.
	// 精确包内嵌入式契约身份，由 SDK 独立检查。
	ContractSha256 string `json:"contract_sha256"`
	// Cargo's debug-information setting, independent of optimization.
	// Cargo 调试信息设置，独立于优化。
	DebugInfo string `json:"debug_info"`
	// SHA-256 of the exact machine-readable selected-input report emitted by build.rs.
	// build.rs 输出的精确机器可读选定输入报告的 SHA-256。
	InputsSha256 string `json:"inputs_sha256"`
	// Actual Cargo optimization setting, not an inferred profile label.
	// 实际 Cargo 优化设置，不推断配置名称。
	OptLevel string `json:"opt_level"`
	// Bundled package lockfile identity; a consuming Rust workspace may resolve a different dependency graph.
	// 包内锁文件身份；消费它的 Rust 工作区可能解析出不同依赖图。
	PackageLockSha256 string `json:"package_lock_sha256"`
	// Cargo's target pointer width, preserved as its exact textual value.
	// Cargo 目标指针位宽，保留其精确文本值。
	PointerWidth string `json:"pointer_width"`
	// The selected rustc executable's verbose version output.
	// 所选 rustc 可执行文件的详细版本输出。
	Rustc string `json:"rustc"`
	// SHA-256 of Cargo's exact encoded additional compiler flags.
	// Cargo 精确编码额外编译参数的 SHA-256。
	RustflagsSha256 string `json:"rustflags_sha256"`
	// SHA-256 of sorted package-relative input paths and their exact content hashes.
	// 排序后包相对输入路径及其精确内容摘要的 SHA-256。
	SourceSha256 string `json:"source_sha256"`
	// Cargo's target triple for this build.
	// 此构建的 Cargo 目标三元组。
	Target string `json:"target"`
	// Cargo's target architecture identity.
	// Cargo 目标架构身份。
	TargetArch string `json:"target_arch"`
	// Cargo's target operating-system identity.
	// Cargo 目标操作系统身份。
	TargetOs string `json:"target_os"`
}

// Sorted Cargo feature environment suffixes; they are not reverse-mapped into guessed feature names.
// 排序后 Cargo 功能环境后缀；不反向映射为猜测功能名。
type EmbeddedOutputEmbeddedBuildIdentityCargoFeatures []string

// Structured error; `code` is stable and `message` is an English diagnostic.
// 结构化错误；`code` 稳定，`message` 为英文诊断信息。
type EmbeddedOutputEmbeddedError struct {
	// Stable classification consumed by SDKs instead of parsing text.
	// 供 SDK 使用的稳定分类，避免解析文案。
	Code EmbeddedOutputEmbeddedErrorCode `json:"code"`
	// Human-readable detail without credentials or plugin input dumps.
	// 不含凭据或插件输入转储的可读详情。
	Message string `json:"message"`
}

// Stable machine-readable failures shared by Rust and the versioned FFI protocol.
// Rust 与版本化 FFI 协议共享的稳定机器可读错误。
type EmbeddedOutputEmbeddedErrorCode string

const (
	// The supplied configuration or request violates its declared contract.
	// 提供的配置或请求违反其声明的契约。
	EmbeddedOutputEmbeddedErrorCodeInvalidArgument EmbeddedOutputEmbeddedErrorCode = "invalid_argument"
	// The requested identity does not exist or its retained record has expired.
	// 请求的身份不存在，或其保留记录已经过期。
	EmbeddedOutputEmbeddedErrorCodeNotFound EmbeddedOutputEmbeddedErrorCode = "not_found"
	// The caller supplied an identity from an inactive generation.
	// 调用方提供了非活动代次的身份。
	EmbeddedOutputEmbeddedErrorCodeStaleGeneration EmbeddedOutputEmbeddedErrorCode = "stale_generation"
	// A bounded resource cannot admit more work.
	// 有界资源无法接纳更多工作。
	EmbeddedOutputEmbeddedErrorCodeCapacityExceeded EmbeddedOutputEmbeddedErrorCode = "capacity_exceeded"
	// The requested state transition conflicts with live work.
	// 请求的状态变更与仍在运行的工作冲突。
	EmbeddedOutputEmbeddedErrorCodeBusy EmbeddedOutputEmbeddedErrorCode = "busy"
	// This exact request already has a completion owner or retained terminal result.
	// 此精确请求已具有完成所有者或保留的终态结果。
	EmbeddedOutputEmbeddedErrorCodeAlreadyCompleted EmbeddedOutputEmbeddedErrorCode = "already_completed"
	// The runtime or registration no longer accepts work.
	// 运行时或注册项已停止接纳工作。
	EmbeddedOutputEmbeddedErrorCodeClosed EmbeddedOutputEmbeddedErrorCode = "closed"
	// The caller requested cooperative cancellation.
	// 调用方请求了协作取消。
	EmbeddedOutputEmbeddedErrorCodeCancelled EmbeddedOutputEmbeddedErrorCode = "cancelled"
	// The original end-to-end execution budget expired.
	// 原始端到端执行预算已经耗尽。
	EmbeddedOutputEmbeddedErrorCodeDeadlineExceeded EmbeddedOutputEmbeddedErrorCode = "deadline_exceeded"
	// The trusted host did not grant the requested capability.
	// 可信宿主未授予所请求的能力。
	EmbeddedOutputEmbeddedErrorCodePermissionDenied EmbeddedOutputEmbeddedErrorCode = "permission_denied"
	// The requested backend or protocol feature is unavailable.
	// 请求的后端或协议功能不可用。
	EmbeddedOutputEmbeddedErrorCodeUnsupported EmbeddedOutputEmbeddedErrorCode = "unsupported"
	// Plugin initialization or execution failed.
	// 插件初始化或执行失败。
	EmbeddedOutputEmbeddedErrorCodeExecutionFailed EmbeddedOutputEmbeddedErrorCode = "execution_failed"
	// Resource teardown failed and still owns its capacity.
	// 资源清理失败且仍占有容量。
	EmbeddedOutputEmbeddedErrorCodeCleanupFailed EmbeddedOutputEmbeddedErrorCode = "cleanup_failed"
	// An internal invariant failed; the caller must not retry mutations blindly.
	// 内部不变量失败，调用方不得盲目重试有副作用的操作。
	EmbeddedOutputEmbeddedErrorCodeInternal EmbeddedOutputEmbeddedErrorCode = "internal"
)

// Immutable host-approved aggregate budgets across every generation and execution domain of one plugin.
// 一个插件的全部代次与执行域共享的不可变宿主批准聚合预算。
type EmbeddedOutputEmbeddedPluginConfig struct {
	// Maximum retained operations, including completed results not explicitly forgotten.
	// 保留操作的数量上限，包含尚未显式遗忘的已完成结果。
	MaxOperations uint64 `json:"max_operations"`
	// Maximum exact serialized queued request bytes across this plugin.
	// 此插件全部排队请求精确序列化字节数上限。
	MaxQueuedBytes uint64 `json:"max_queued_bytes"`
	// Maximum accepted queued calls across all domains and sessions.
	// 全部域和会话已接纳排队调用的合计上限。
	MaxQueuedCalls uint64 `json:"max_queued_calls"`
	// Maximum retained pool identities, including closed generations awaiting explicit removal.
	// 保留池身份的数量上限，包含等待显式移除的已关闭代次。
	MaxRegisteredPools uint64 `json:"max_registered_pools"`
	// Maximum actual resident VMs plus other domains' unused dedicated reservations.
	// 实际常驻 VM 与其他域未使用专用预留的合计上限。
	MaxResidentVms uint64 `json:"max_resident_vms"`
	// Maximum dispatched operations, retained through initialization, host waiting and cleanup.
	// 已分发操作上限，计费覆盖初始化、宿主等待及清理。
	MaxRunningCalls uint64 `json:"max_running_calls"`
	// Maximum retained sessions, including closed session records.
	// 保留会话的数量上限，包含已关闭会话记录。
	MaxSessions uint64 `json:"max_sessions"`
}

// Plugin-wide observation from exact immutable pool ownership, including draining generations.
// 根据精确不可变池归属形成的插件级观测，包含正在排空的代次。
type EmbeddedOutputEmbeddedPluginSnapshot struct {
	// Dispatched unfinished operations, including initialization and cleanup.
	// 已分发未完成操作，包含初始化与清理。
	ActiveOperations uint64 `json:"active_operations"`
	// Whether new pool and call admission is permanently closed.
	// 新池及新调用入场是否已永久关闭。
	Closing bool `json:"closing"`
	// Actual residents plus all unused dedicated domain guarantees.
	// 实际常驻实例与全部未使用专用域保证的合计。
	CommittedResidentVms uint64 `json:"committed_resident_vms"`
	// Immutable effective aggregate policy.
	// 不可变有效聚合策略。
	Config EmbeddedOutputEmbeddedPluginConfig `json:"config"`
	// Trusted host plugin identity used by module definitions and fair scheduling.
	// 模块定义与公平调度使用的可信宿主插件身份。
	PluginId string `json:"plugin_id"`
	// Exact serialized bytes still held by queued requests.
	// 排队请求仍持有的精确序列化字节数。
	QueuedBytes uint64 `json:"queued_bytes"`
	// Accepted queued calls across every domain and session.
	// 全部域和会话已接纳的排队调用。
	QueuedCalls uint64 `json:"queued_calls"`
	// Actual resident VM counters through confirmed destruction.
	// 持续记账到确认销毁的实际常驻 VM 计数。
	Resources EmbeddedOutputPoolUsage `json:"resources"`
	// Retained operations, including results whose callers dropped their handles.
	// 保留操作，包含调用方已丢弃句柄的结果。
	RetainedOperations uint64 `json:"retained_operations"`
	// Retained pools, including closed metadata not explicitly forgotten.
	// 保留池，包含尚未显式遗忘的已关闭元数据。
	RetainedPools uint64 `json:"retained_pools"`
	// Retained fixed sessions, including closed records.
	// 保留固定会话，包含已关闭记录。
	RetainedSessions uint64 `json:"retained_sessions"`
}

// Live scheduler observations; queue bytes exclude already-dispatched request values.
// 实时调度观测；队列字节不包含已分发的请求值。
type EmbeddedOutputEmbeddedRuntimeUsage struct {
	// Nonqueued unfinished operations, including rejected requests awaiting terminal publication.
	// 不在队列中的未完成操作，包含等待终态发布的已拒绝请求。
	ActiveOperations uint64 `json:"active_operations"`
	// Operations whose execution returned and whose cleanup remains owned.
	// 执行已返回但仍拥有清理的操作。
	CleaningOperations uint64 `json:"cleaning_operations"`
	// Whether new admission has permanently closed.
	// 新入场是否已永久关闭。
	Closing bool `json:"closing"`
	// Exact serialized queued request bytes.
	// 排队请求的精确序列化字节数。
	QueuedBytes uint64 `json:"queued_bytes"`
	// Requests waiting for actual resource admission.
	// 等待实际资源入场的请求。
	QueuedCalls uint64 `json:"queued_calls"`
}

// Observable pinned-session lifecycle; closing never implies that its VM is already destroyed.
// 可观察的固定会话生命周期；正在关闭绝不表示其 VM 已销毁。
type EmbeddedOutputEmbeddedSessionPhase string

const (
	// Creation is queued, initializing, or publishing its operation result.
	// 创建正在排队、初始化或发布操作结果。
	EmbeddedOutputEmbeddedSessionPhaseOpening EmbeddedOutputEmbeddedSessionPhase = "opening"
	// The exact VM is idle and ready for another call.
	// 精确 VM 空闲，能够接收下一次调用。
	EmbeddedOutputEmbeddedSessionPhaseReady EmbeddedOutputEmbeddedSessionPhase = "ready"
	// One call owns the VM through execution and operation cleanup.
	// 一次调用持有 VM，覆盖执行与操作清理。
	EmbeddedOutputEmbeddedSessionPhaseRunning EmbeddedOutputEmbeddedSessionPhase = "running"
	// Admission stopped; actual execution, queued rejection or retirement remains unfinished.
	// 入场已停止；实际执行、队列拒绝或退役尚未完成。
	EmbeddedOutputEmbeddedSessionPhaseClosing EmbeddedOutputEmbeddedSessionPhase = "closing"
	// All queued work and actual VM ownership have drained.
	// 全部排队任务和实际 VM 所有权均已排空。
	EmbeddedOutputEmbeddedSessionPhaseClosed EmbeddedOutputEmbeddedSessionPhase = "closed"
)

// Bounded retained session observation with immutable pool ownership.
// 具有不可变池归属的有界保留会话观测。
type EmbeddedOutputEmbeddedSessionSnapshot struct {
	// Current operation, including initialization and cleanup; absent while idle.
	// 当前操作，包含初始化与清理；空闲时省略。
	ActiveOperation *string `json:"active_operation"`
	// First execution failure that made the session unusable.
	// 导致会话不可用的首次执行错误。
	Error *EmbeddedOutputEmbeddedError `json:"error"`
	// Current lifecycle observation.
	// 当前生命周期观测。
	Phase EmbeddedOutputEmbeddedSessionPhase `json:"phase"`
	// Exact generation and security partition fixed at creation.
	// 创建时固定的精确代次与安全分区。
	PoolId string `json:"pool_id"`
	// Accepted calls waiting behind this session's current owner.
	// 此会话当前所有者之后等待的已接纳调用数。
	QueuedCalls uint64 `json:"queued_calls"`
	// Runtime-issued opaque identity, never reused after forgetting.
	// 运行时签发的不透明身份，遗忘后绝不复用。
	SessionId string `json:"session_id"`
}

// Live structured failure envelope whose shape also drives SDK generation.
// 同时驱动 SDK 生成的实际结构化失败信封。
type EmbeddedOutputErrorResponse struct {
	// Borrowed core error retained until response encoding returns.
	// 保留到响应编码返回的借用核心错误。
	Error EmbeddedOutputEmbeddedError `json:"error"`
	// Accepted embedded protocol version.
	// 接受的嵌入式协议版本。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Exact business-error discriminator.
	// 精确业务错误判别。
	Status EmbeddedOutputErrorStatus `json:"status"`
}

// Single legal business-error discriminator, independent of native transport failures.
// 独立于原生传输失败的唯一合法业务错误判别。
type EmbeddedOutputErrorStatus string

const (
	// Delivered structured core failure.
	// 已交付结构化核心失败。
	EmbeddedOutputErrorStatusError EmbeddedOutputErrorStatus = "error"
)

// Declared backend; unavailable variants are rejected instead of downgraded.
// 声明的执行后端；不可用的取值直接拒绝，不降级。
type EmbeddedOutputExecutionBackend string

const (
	// Execute in owned Lua VMs inside the current host process.
	// 在当前宿主进程内的受管 Lua VM 中执行。
	EmbeddedOutputExecutionBackendInProcess EmbeddedOutputExecutionBackend = "in_process"
	// Reserved protocol identity; no worker backend is advertised yet.
	// 预留的协议身份；目前尚未声明工作进程后端可用。
	EmbeddedOutputExecutionBackendWorkerProcess EmbeddedOutputExecutionBackend = "worker_process"
)

// Actual handler lifecycle, separate from its reported business effect.
// 真实处理器生命周期，独立于其报告的业务副作用。
type EmbeddedOutputHostEffectPhase string

const (
	// Retention reserved before execution.
	// 执行前已预留保留容量。
	EmbeddedOutputHostEffectPhasePrepared EmbeddedOutputHostEffectPhase = "prepared"
	// Handler may execute; cancellation is not completion.
	// 处理器可能执行；取消不等于完成。
	EmbeddedOutputHostEffectPhaseRunning EmbeddedOutputHostEffectPhase = "running"
	// Actual handler and admission ownership have been released.
	// 真实处理器及入场所有权已释放。
	EmbeddedOutputHostEffectPhaseCompleted EmbeddedOutputHostEffectPhase = "completed"
)

// Resolution for one exact original host effect; neither registration nor caller identity can be supplied anew.
// 一个精确原宿主副作用的结论；不得重新提供注册或调用方身份。
type EmbeddedOutputHostEffectReconciliation struct {
	// Exact effect identity from the original snapshot, in the same order as its original records.
	// 原始快照中的精确副作用身份，顺序与其原始记录相同。
	EffectId string `json:"effect_id"`
	// Proven final outcome of this original effect, never a retry's outcome.
	// 此原始副作用的已证实最终结果，绝非重试结果。
	Effects EmbeddedOutputResolvedEffectState `json:"effects"`
	// Nonempty host audit or transaction-query reference; credentials and business payloads do not belong here.
	// 非空宿主审计或事务查询引用；此处不应包含凭证及业务载荷。
	Evidence string `json:"evidence"`
}

// Bounded evidence retained independently from values returned to Lua.
// 独立于返回 Lua 的值保留的有界证据。
type EmbeddedOutputHostEffectRecord struct {
	// Original host-bound caller identity, retained for reconciliation without consulting a newer plugin generation.
	// 原始宿主绑定调用身份；对账保留该身份，不查询较新的插件代次。
	Caller EmbeddedOutputCapabilityCaller `json:"caller"`
	// Public capability name, excluding business arguments and credentials.
	// 公开能力名称，不包含业务参数与凭证。
	CapabilityName string `json:"capability_name"`
	// Interface version of the exact registration.
	// 精确注册的接口版本。
	CapabilityVersion string `json:"capability_version"`
	// Never-reused identity within the original operation.
	// 原始操作内绝不复用的身份。
	EffectId string `json:"effect_id"`
	// Trusted host evidence, independent from caller cancellation.
	// 可信宿主证据，独立于调用方取消。
	Effects EmbeddedOutputEffectState `json:"effects"`
	// Actual handler ownership lifecycle.
	// 真实处理器所有权生命周期。
	Phase EmbeddedOutputHostEffectPhase `json:"phase"`
	// Exact immutable registration identity.
	// 精确不可变注册身份。
	RegistrationId string `json:"registration_id"`
	// SDK request identity for queued execution.
	// 队列执行的 SDK 请求身份。
	RequestId *string `json:"request_id"`
}

// SDK request copied from an admitted, authenticated invocation.
// 从已入场、已认证调用复制的 SDK 请求。
type EmbeddedOutputHostRequest struct {
	// Validated structured business arguments.
	// 已校验的结构化业务参数。
	Arguments any `json:"arguments"`
	// Trusted host identity, separate from business arguments.
	// 可信宿主身份，独立于业务参数。
	Caller EmbeddedOutputCapabilityCaller `json:"caller"`
	// Original operation effect record, absent only for untracked low-level calls.
	// 原始操作副作用记录，仅未跟踪低层调用省略。
	EffectId *string `json:"effect_id"`
	// Declared capability name.
	// 声明的能力名称。
	Name string `json:"name"`
	// Exact registration, independent from later replacement by name.
	// 精确注册，独立于后续按名称替换。
	RegistrationId string `json:"registration_id"`
	// Advisory remaining duration; the core retains the original deadline.
	// 建议剩余时长；核心保留原始截止时间。
	RemainingMs uint64 `json:"remaining_ms"`
	// Never-reused identity used for completion and host-side deduplication.
	// 用于完成与宿主侧去重、绝不复用的身份。
	RequestId string `json:"request_id"`
	// Declared semantic interface version.
	// 声明的语义接口版本。
	Version string `json:"version"`
}

// Observable request phase; cancellation does not imply handler termination.
// 可观察请求阶段；取消不代表处理器已经终止。
type EmbeddedOutputHostRequestPhase string

const (
	// No SDK handler has received this request.
	// 尚无 SDK 处理器收到此请求。
	EmbeddedOutputHostRequestPhaseQueued EmbeddedOutputHostRequestPhase = "queued"
	// Exactly one SDK consumer owns execution.
	// 精确一个 SDK 消费者拥有执行权。
	EmbeddedOutputHostRequestPhaseDispatched EmbeddedOutputHostRequestPhase = "dispatched"
	// An exclusive owner validates the result and releases admission outside locks.
	// 独占所有者在锁外校验结果并释放入场许可。
	EmbeddedOutputHostRequestPhaseCompleting EmbeddedOutputHostRequestPhase = "completing"
	// The actual handler has finished and admission is released.
	// 真实处理器已结束且入场许可已释放。
	EmbeddedOutputHostRequestPhaseCompleted EmbeddedOutputHostRequestPhase = "completed"
)

// Live request status for SDK cancellation and orderly runtime shutdown.
// 用于 SDK 取消与运行时有序关闭的实时请求状态。
type EmbeddedOutputHostRequestStatus struct {
	// Cooperative cancellation reason, preserved until actual completion.
	// 协作取消原因，保留到真实完成。
	Cancellation *EmbeddedOutputEmbeddedError `json:"cancellation"`
	// Actual execution lifecycle.
	// 真实执行生命周期。
	Phase EmbeddedOutputHostRequestPhase `json:"phase"`
	// Exact request identity.
	// 精确请求身份。
	RequestId string `json:"request_id"`
}

// Initialization ownership is separate from the core's execution and closing state.
// 初始化所有权独立于核心的执行及关闭状态。
type EmbeddedOutputInitializationPhase string

const (
	// A known identity exists but no engine or worker has been constructed.
	// 已存在已知身份，但尚未构造引擎或工作线程。
	EmbeddedOutputInitializationPhaseReserved EmbeddedOutputInitializationPhase = "reserved"
	// Exactly one native call owns construction.
	// 精确一个原生调用拥有构造过程。
	EmbeddedOutputInitializationPhaseInitializing EmbeddedOutputInitializationPhase = "initializing"
	// The actual core owner was stored successfully.
	// 实际核心所有者已成功保存。
	EmbeddedOutputInitializationPhaseReady EmbeddedOutputInitializationPhase = "ready"
	// Construction returned an explicit error; retained storage still requires verified drainage.
	// 构造返回明确错误；已保留存储仍需验证排空。
	EmbeddedOutputInitializationPhaseFailed EmbeddedOutputInitializationPhase = "failed"
	// Construction panicked and safe library unloading cannot be proven.
	// 构造发生 panic，无法证明可以安全卸载动态库。
	EmbeddedOutputInitializationPhaseFaulted EmbeddedOutputInitializationPhase = "faulted"
)

// Historical checkpoint, not a live handle and not evidence authorizing execution replay.
// 历史检查点，不是活动句柄，也不是授权执行重放的证据。
type EmbeddedOutputJournalOperation struct {
	// Separate final host attestation; the original snapshot remains unchanged, including unknown results.
	// 独立最终宿主证明；原始快照保持不变，包括未知结果。
	Reconciliation *EmbeddedOutputOperationReconciliation `json:"reconciliation"`
	// Monotonic compare-and-swap revision; positive and bounded by SQLite's signed integer.
	// 单调比较交换修订号；为正数且受 SQLite 有符号整数范围约束。
	Revision uint64 `json:"revision"`
	// Original runtime namespace, never rebound to the namespace of a restarted runtime.
	// 原始运行时命名空间，绝不重新绑定到重启后的命名空间。
	RuntimeId string `json:"runtime_id"`
	// Exact last committed observation; an unfinished phase remains unfinished after restart.
	// 最后提交的精确观测；未结束的阶段在重启后仍保持未结束。
	Snapshot EmbeddedOutputOperationSnapshot `json:"snapshot"`
}

// Explicit operation origin; unbound low-level work is never inferred to belong to a current plugin.
// 明确的操作来源；未绑定的低层工作绝不被推断归属于当前插件。
type EmbeddedOutputOperationContext interface {
	// embeddedVariantEmbeddedOutputOperationContext seals this generated union without invoking serialization hooks.
	// embeddedVariantEmbeddedOutputOperationContext 封闭此生成联合，不调用序列化钩子。
	embeddedVariantEmbeddedOutputOperationContext()
}

// embeddedVariantEmbeddedOutputOperationContext marks the exact EmbeddedOutputOperationContextVariant1 alternative.
// embeddedVariantEmbeddedOutputOperationContext 标识精确的 EmbeddedOutputOperationContextVariant1 分支。
func (EmbeddedOutputOperationContextVariant1) embeddedVariantEmbeddedOutputOperationContext() {}

// embeddedVariantEmbeddedOutputOperationContext marks the exact EmbeddedOutputOperationContextVariant2 alternative.
// embeddedVariantEmbeddedOutputOperationContext 标识精确的 EmbeddedOutputOperationContextVariant2 分支。
func (EmbeddedOutputOperationContextVariant2) embeddedVariantEmbeddedOutputOperationContext() {}

// The low-level host admitted this operation without a module binding.
// 低层宿主接纳此操作时没有模块绑定。
type EmbeddedOutputOperationContextVariant1 struct {
	// Kind is derived from the packaged wire contract.
	// Kind 从包内线契约派生。
	Kind EmbeddedOutputOperationContextVariant1Kind `json:"kind"`
}

// EmbeddedOutputOperationContextVariant1Kind is derived from the packaged wire contract.
// EmbeddedOutputOperationContextVariant1Kind 从包内线契约派生。
type EmbeddedOutputOperationContextVariant1Kind string

const (
	// EmbeddedOutputOperationContextVariant1KindUnbound is derived from the packaged wire contract.
	// EmbeddedOutputOperationContextVariant1KindUnbound 从包内线契约派生。
	EmbeddedOutputOperationContextVariant1KindUnbound EmbeddedOutputOperationContextVariant1Kind = "unbound"
)

// The formal scheduler froze this module context before publishing the operation.
// 正式调度器在发布操作前冻结了此模块上下文。
type EmbeddedOutputOperationContextVariant2 struct {
	// Original trusted caller shared by initialization and this operation's host callbacks.
	// 初始化及此操作宿主回调共同使用的原始可信调用方。
	Caller EmbeddedOutputCapabilityCaller `json:"caller"`
	// Exact capability membership snapshot frozen when the pool was registered.
	// 注册池时冻结的精确能力成员快照。
	CapabilityRevision string `json:"capability_revision"`
	// Requested declared export; absent only for a fixed-session opening operation.
	// 请求的已声明导出；仅固定会话开启操作省略。
	Export *string `json:"export"`
	// Kind is derived from the packaged wire contract.
	// Kind 从包内线契约派生。
	Kind EmbeddedOutputOperationContextVariant2Kind `json:"kind"`
	// Exact retained pool identity, not a lookup of the plugin's newest pool.
	// 精确保留池身份，不查询插件最新的池。
	PoolId string `json:"pool_id"`
}

// EmbeddedOutputOperationContextVariant2Kind is derived from the packaged wire contract.
// EmbeddedOutputOperationContextVariant2Kind 从包内线契约派生。
type EmbeddedOutputOperationContextVariant2Kind string

const (
	// EmbeddedOutputOperationContextVariant2KindModule is derived from the packaged wire contract.
	// EmbeddedOutputOperationContextVariant2KindModule 从包内线契约派生。
	EmbeddedOutputOperationContextVariant2KindModule EmbeddedOutputOperationContextVariant2Kind = "module"
)

// Live worker observations; retained receipts keep quota even after the thread has finished.
// 实时工作线程观测；线程结束后，保留的回执仍占有配额。
type EmbeddedOutputOperationJournalWorkerStatus struct {
	// New attempts are permanently refused while admitted attempts finish normally.
	// 永久拒绝新尝试，而已接纳尝试正常完成。
	Closing bool `json:"closing"`
	// First infrastructure failure; individual database rejection remains on its own receipt.
	// 首个基础设施故障；单独的数据库拒绝仍位于各自回执。
	Failure *EmbeddedOutputEmbeddedError `json:"failure"`
	// Encoded request bytes reserved until each attempt's last owner disappears.
	// 每次尝试最后一个所有者消失前预留的请求编码字节数。
	PendingBytes uint64 `json:"pending_bytes"`
	// Total owned attempts including caller-retained completed receipts.
	// 拥有的尝试总数，包含调用方保留的已完成回执。
	PendingWrites uint64 `json:"pending_writes"`
	// Attempts still owned by the queue.
	// 仍由队列拥有的尝试数。
	QueuedWrites uint64 `json:"queued_writes"`
	// Actual thread termination, observed from its join handle rather than a provisional flag.
	// 从等待句柄观测到的真实线程终止，而非临时标记。
	WorkerExited bool `json:"worker_exited"`
	// Whether a real write is currently owned by the storage thread.
	// 存储线程当前是否拥有真实写入。
	Writing bool `json:"writing"`
}

// A failed checkpoint remains queryable by exact operation ID until that original checkpoint is acknowledged.
// 失败检查点可按精确操作 ID 查询，直至原检查点得到确认。
type EmbeddedOutputOperationPersistenceFailure struct {
	// Retained persistence error; this does not rewrite the original business result.
	// 保留的持久化错误；不改写原始业务结果。
	Error EmbeddedOutputEmbeddedError `json:"error"`
	// Stable original operation identity, never a replacement execution.
	// 稳定的原始操作身份，绝非替代执行。
	OperationId string `json:"operation_id"`
	// Exact candidate phase whose write failed; terminal candidates are not yet publicly terminal.
	// 写入失败的精确候选阶段；终态候选尚不是公开终态。
	Phase EmbeddedOutputOperationPhase `json:"phase"`
	// Explicit host retry coordination, separate from ordinary polling.
	// 显式宿主重试协调，独立于普通轮询。
	Retry EmbeddedOutputCheckpointRetryState `json:"retry"`
}

// Execution phase; cancellation intent is reported separately from actual termination.
// 执行阶段；取消意图与实际终止分开报告。
type EmbeddedOutputOperationPhase string

const (
	// Accepted and waiting for resource admission.
	// 已接纳且正在等待资源入场。
	EmbeddedOutputOperationPhaseQueued EmbeddedOutputOperationPhase = "queued"
	// Creating or loading the selected VM instance.
	// 正在创建或加载选定的 VM 实例。
	EmbeddedOutputOperationPhaseInitializing EmbeddedOutputOperationPhase = "initializing"
	// Executing the requested exported function.
	// 正在执行请求的导出函数。
	EmbeddedOutputOperationPhaseRunning EmbeddedOutputOperationPhase = "running"
	// Still owns its VM while awaiting a host capability result.
	// 等待宿主能力结果期间仍拥有其 VM。
	EmbeddedOutputOperationPhaseWaitingForHost EmbeddedOutputOperationPhase = "waiting_for_host"
	// Execution returned but request-owned cleanup is not yet complete.
	// 执行已返回，但请求所属清理尚未完成。
	EmbeddedOutputOperationPhaseCleaning EmbeddedOutputOperationPhase = "cleaning"
	// Execution and required cleanup completed successfully.
	// 执行及必要清理已成功完成。
	EmbeddedOutputOperationPhaseSucceeded EmbeddedOutputOperationPhase = "succeeded"
	// Execution or cleanup failed and the actual result is available.
	// 执行或清理失败，且实际结果已经可用。
	EmbeddedOutputOperationPhaseFailed EmbeddedOutputOperationPhase = "failed"
	// Cooperative cancellation completed; this does not imply effect rollback.
	// 协作取消已完成；这不表示副作用已回滚。
	EmbeddedOutputOperationPhaseCancelled EmbeddedOutputOperationPhase = "cancelled"
)

// Actual operation admission acknowledgement; it does not imply execution completion.
// 实际操作入场确认；不代表执行完成。
type EmbeddedOutputOperationReceipt struct {
	// Retained queryable operation identity.
	// 保留且可查询的操作身份。
	OperationId string `json:"operation_id"`
}

// One bounded, final, host-authored attestation covering execution closure and every retained effect.
// 一份有界、最终且由宿主编写的证明，覆盖执行关闭及每个保留副作用。
// This API does not authenticate the attestation; the embedding host must authorize the resolver and verify evidence.
// 此 API 不认证证明；嵌入宿主必须授权对账者并核验证据。
type EmbeddedOutputOperationReconciliation struct {
	// Resolved aggregate covering both recorded callbacks and any other effects from the original Lua execution.
	// 已解决的聚合结论，覆盖记录回调及原 Lua 执行的其他副作用。
	Effects EmbeddedOutputResolvedEffectState `json:"effects"`
	// Nonempty evidence reference proving owner closure and the whole operation's external-effect conclusion.
	// 非空证据引用，证明所有者关闭及整个操作的外部副作用结论。
	Evidence string `json:"evidence"`
	// Closure evidence consistent with the unchanged original execution phase.
	// 与未改变原执行阶段一致的关闭证据。
	Execution EmbeddedOutputReconciledExecution `json:"execution"`
	// Exactly one resolution per original effect, preserving original order and known outcomes.
	// 每个原始副作用精确一个结论，保留原始顺序及已知结果。
	HostEffects EmbeddedOutputOperationReconciliationHostEffects `json:"host_effects"`
	// Stable host-assigned resolution identity, retained unchanged across observation or storage retries.
	// 宿主分配的稳定对账身份，跨观测或存储重试保持不变。
	ResolutionId string `json:"resolution_id"`
	// Authorized host resolver identity, not a plugin-supplied authority claim or an authentication credential.
	// 已授权宿主对账者身份，不是插件提供的权限声明或认证凭证。
	Resolver string `json:"resolver"`
}

// Exactly one resolution per original effect, preserving original order and known outcomes.
// 每个原始副作用精确一个结论，保留原始顺序及已知结果。
type EmbeddedOutputOperationReconciliationHostEffects []EmbeddedOutputHostEffectReconciliation

// Bounded operation snapshot suitable for direct serialization to every SDK.
// 适合直接序列化给各 SDK 的有界操作快照。
type EmbeddedOutputOperationSnapshot struct {
	// Whether cooperative cancellation has been requested.
	// 是否已请求协作取消。
	CancellationRequested bool `json:"cancellation_requested"`
	// Admission-time module authority, or an explicit unbound low-level origin; never reconstructed from effects.
	// 入场时模块权威，或明确未绑定的低层来源；绝不从副作用重建。
	Context EmbeddedOutputOperationContext `json:"context"`
	// Explicit effect status; successful execution does not automatically imply commit.
	// 显式副作用状态；执行成功不自动表示提交。
	Effects EmbeddedOutputEffectState `json:"effects"`
	// Structured terminal error; absent while execution is still in progress.
	// 结构化终态错误；执行仍在进行时省略。
	Error **EmbeddedOutputEmbeddedError `json:"error,omitempty"`
	// Exact host callback evidence retained even after Lua failure, cancellation or output rejection.
	// 即使 Lua 失败、取消或输出被拒绝也保留的精确宿主回调证据。
	HostEffects EmbeddedOutputOperationSnapshotHostEffects `json:"host_effects"`
	// Opaque identifier, never a JavaScript floating-point integer.
	// 不透明标识符，绝不使用 JavaScript 浮点整数。
	OperationId string `json:"operation_id"`
	// Current execution phase, including still-running cancelled requests.
	// 当前执行阶段，包含已请求取消但仍在运行的请求。
	Phase EmbeddedOutputOperationPhase `json:"phase"`
	// Successful value; absent before completion or on failure, distinct from JSON null.
	// 成功值；完成前或失败时省略，与 JSON 空值不同。
	Value *any `json:"value,omitempty"`
}

// Exact host callback evidence retained even after Lua failure, cancellation or output rejection.
// 即使 Lua 失败、取消或输出被拒绝也保留的精确宿主回调证据。
type EmbeddedOutputOperationSnapshotHostEffects []EmbeddedOutputHostEffectRecord

// Actual pool registration acknowledgement shared by capacity preparation and publication.
// 容量准备及发布共享的实际池注册确认。
type EmbeddedOutputPoolReceipt struct {
	// Immutable core pool identity.
	// 不可变核心池身份。
	PoolId string `json:"pool_id"`
}

// Current counters for one governor or one execution group.
// 单个治理器或执行分组的当前计数。
type EmbeddedOutputPoolUsage struct {
	// Allocations that have not finished initialization.
	// 尚未完成初始化的分配。
	Creating uint64 `json:"creating"`
	// Initialized allocations that do not currently execute.
	// 已初始化且当前未执行的分配。
	Idle uint64 `json:"idle"`
	// All allocated slots, including creating and retiring instances.
	// 全部分配槽位，包含正在创建与退役的实例。
	Resident uint64 `json:"resident"`
	// Allocations retained until teardown has genuinely completed.
	// 保留到清理真正完成的分配。
	Retiring uint64 `json:"retiring"`
	// Allocations holding a parent and group execution permit together.
	// 同时持有父级与分组执行许可的分配。
	Running uint64 `json:"running"`
}

// Historical execution closure asserted by the trusted host after actual owners have stopped.
// 实际所有者停止后，由可信宿主断言的历史执行关闭。
type EmbeddedOutputReconciledExecution string

const (
	// The unchanged original snapshot already contains an observed terminal execution result.
	// 未改变的原始快照已包含观测到的终态执行结果。
	EmbeddedOutputReconciledExecutionObservedTerminal EmbeddedOutputReconciledExecution = "observed_terminal"
	// Actual owners stopped without a durable terminal result; the original nonterminal snapshot stays unchanged.
	// 实际所有者停止但没有持久终态结果；原非终态快照保持不变。
	EmbeddedOutputReconciledExecutionStoppedWithoutResult EmbeddedOutputReconciledExecution = "stopped_without_result"
)

// Exact identities from one atomic host capability publication.
// 单次原子宿主能力发布的精确身份。
type EmbeddedOutputRegistrationReceipt struct {
	// Identities retain the descriptor batch's original order.
	// 身份保留描述符批次的原始顺序。
	RegistrationIds EmbeddedOutputRegistrationReceiptRegistrationIds `json:"registration_ids"`
}

// Identities retain the descriptor batch's original order.
// 身份保留描述符批次的原始顺序。
type EmbeddedOutputRegistrationReceiptRegistrationIds []string

// Explicit resolved effects; unknown outcomes remain unreconciled instead of being coerced into success.
// 显式已解决副作用；未知结果保持未对账，不强制转为成功。
type EmbeddedOutputResolvedEffectState string

const (
	// Trusted evidence proves the business effect never started.
	// 可信证据证明业务副作用从未开始。
	EmbeddedOutputResolvedEffectStateNotStarted EmbeddedOutputResolvedEffectState = "not_started"
	// Trusted evidence proves no externally visible mutation applies.
	// 可信证据证明不存在适用的外部可见变更。
	EmbeddedOutputResolvedEffectStateNotApplicable EmbeddedOutputResolvedEffectState = "not_applicable"
	// The original external transaction is proven committed.
	// 原外部事务已证实提交。
	EmbeddedOutputResolvedEffectStateCommitted EmbeddedOutputResolvedEffectState = "committed"
	// The original external transaction is proven rolled back.
	// 原外部事务已证实回滚。
	EmbeddedOutputResolvedEffectStateRolledBack EmbeddedOutputResolvedEffectState = "rolled_back"
)

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRootDescribeResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputTransportDescription `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRootRuntimeCloseResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputRuntimeReceipt `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRootRuntimeFreeResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputRuntimeReceipt `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRootRuntimeInitializeResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputRuntimeReceipt `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRootRuntimeReserveResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputRuntimeReceipt `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRootRuntimeStatusResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputRuntimeSnapshot `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeCallSubmitResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputOperationReceipt `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeCapabilitiesListResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputRuntimeCapabilitiesListResponseResult `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed result whose owner lives through serialization.
// 借用结果，其所有者跨序列化存活。
type EmbeddedOutputRuntimeCapabilitiesListResponseResult []EmbeddedOutputCapabilityDescriptor

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeCapabilitiesRegisterResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputRegistrationReceipt `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeCapabilityForgetResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeCapabilityStatusResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputCapabilityRegistrationStatus `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeCapabilityUnregisterResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeHistoryForgetResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeHistoryGetResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedOutputJournalOperation `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeHistoryNextResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedOutputJournalOperation `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeHistoryReconcileResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result uint64 `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeHostRequestCompleteResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeHostRequestStatusResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputHostRequestStatus `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeHostRequestsTakeResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputRuntimeHostRequestsTakeResponseResult `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed result whose owner lives through serialization.
// 借用结果，其所有者跨序列化存活。
type EmbeddedOutputRuntimeHostRequestsTakeResponseResult []EmbeddedOutputHostRequest

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeOperationCancelResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result bool `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeOperationForgetResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeOperationPersistenceFailureResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedOutputOperationPersistenceFailure `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeOperationRetryCheckpointResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result bool `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeOperationStatusResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputOperationSnapshot `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeOperationWaitResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputOperationSnapshot `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimePluginCloseResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimePluginForgetResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimePluginRegisterResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimePluginStatusResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputEmbeddedPluginSnapshot `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimePoolCloseResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimePoolForgetResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimePoolRegisterResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputPoolReceipt `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimePoolRevokePermissionResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result bool `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimePoolStatusResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputPoolUsage `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Exact lifecycle identity returned before or after a runtime control mutation.
// 在运行时控制变更前或后返回的精确生命周期身份。
type EmbeddedOutputRuntimeReceipt struct {
	// Immutable transport-local runtime identity.
	// 不可变传输局部运行时身份。
	RuntimeId string `json:"runtime_id"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeSessionCloseResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeSessionForgetResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result *EmbeddedJSONNull `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeSessionOpenResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputSessionReceipt `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeSessionStatusResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputEmbeddedSessionSnapshot `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeSessionSubmitResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputOperationReceipt `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Queryable construction and core closure evidence; no runtime implementation state is inferred by SDKs.
// 可查询构造与核心关闭证据；SDK 不推断运行时实现状态。
type EmbeddedOutputRuntimeSnapshot struct {
	// True only after construction finishes and all created core and storage workers and receipts drain.
	// 仅在构造结束且全部已创建核心、存储线程与回执排空后为真。
	Closed bool `json:"closed"`
	// Whether this slot has permanently closed admission.
	// 此槽是否已永久关闭入场。
	Closing bool `json:"closing"`
	// Actual runtime namespace, present only after successful construction.
	// 实际运行时命名空间，仅在成功构造后存在。
	CoreRuntimeId *string `json:"core_runtime_id"`
	// Retained failure; a failed construction still owns its bounded registration until explicit release.
	// 保留失败；构造失败仍拥有其有界注册，直到显式释放。
	Error *EmbeddedOutputEmbeddedError `json:"error"`
	// Actual one-shot construction state.
	// 实际单次构造状态。
	Initialization EmbeddedOutputInitializationPhase `json:"initialization"`
	// Actual storage worker status when durable ownership has been created, including failed construction.
	// 持久所有权创建后的实际存储工作线程状态，包含构造失败。
	Persistence *EmbeddedOutputOperationJournalWorkerStatus `json:"persistence"`
	// Live resident and execution accounting directly from the core when available.
	// 可用时直接来自核心的实时常驻与执行计数。
	Resources *EmbeddedOutputPoolUsage `json:"resources"`
	// Exact FFI slot identity used by every control command.
	// 每个控制命令使用的精确 FFI 槽身份。
	RuntimeId string `json:"runtime_id"`
	// Live scheduler observations directly from the core when available.
	// 可用时直接来自核心的实时调度观测。
	Usage *EmbeddedOutputEmbeddedRuntimeUsage `json:"usage"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeStorageRecoverResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result bool `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Borrowed success envelope avoids cloning application output during native response publication.
// 借用成功信封，避免原生响应发布期间克隆应用输出。
type EmbeddedOutputRuntimeStorageStatusResponse struct {
	// Single protocol version authority.
	// 唯一协议版本权威。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Borrowed result whose owner lives through serialization.
	// 借用结果，其所有者跨序列化存活。
	Result EmbeddedOutputOperationJournalWorkerStatus `json:"result"`
	// Exact success discriminator.
	// 精确成功判别。
	Status EmbeddedOutputSuccessStatus `json:"status"`
}

// Fixed-session reservation and its independently queryable initialization operation.
// 固定会话预留及其可独立查询的初始化操作。
type EmbeddedOutputSessionReceipt struct {
	// Initialization operation whose actual outcome must be observed separately.
	// 必须单独观察实际结果的初始化操作。
	OperationId string `json:"operation_id"`
	// Immutable fixed-session identity.
	// 不可变固定会话身份。
	SessionId string `json:"session_id"`
}

// Single legal success discriminator, shared by live encoding and derived contracts.
// 实际编码及派生契约共享的唯一合法成功判别。
type EmbeddedOutputSuccessStatus string

const (
	// Successfully delivered result, including explicit JSON null.
	// 已成功交付结果，包含显式 JSON 空值。
	EmbeddedOutputSuccessStatusOk EmbeddedOutputSuccessStatus = "ok"
)

// Validated native-sized transport configuration, copied once from the host declaration.
// 从宿主声明一次性复制的已校验原生大小传输配置。
type EmbeddedOutputTransportConfig struct {
	// Per-request byte limit.
	// 逐请求字节上限。
	MaxRequestBytes uint64 `json:"max_request_bytes"`
	// Per-response pre-dispatch reservation.
	// 逐响应分发前预留。
	MaxResponseBytes uint64 `json:"max_response_bytes"`
	// Maximum response owners, including pre-dispatch reservations.
	// 响应所有者数量上限，包含分发前预留。
	MaxResultBuffers uint64 `json:"max_result_buffers"`
	// Aggregate response allocation budget.
	// 聚合响应分配预算。
	MaxResultBytes uint64 `json:"max_result_bytes"`
	// Maximum retained runtime identities.
	// 保留运行时身份数量上限。
	MaxRuntimes uint64 `json:"max_runtimes"`
}

// Implemented transport description; every field comes from the running core's own authority.
// 已实现传输描述；每个字段均来自运行核心自身权威。
type EmbeddedOutputTransportDescription struct {
	// Version of the independent embedded ABI structures.
	// 独立嵌入式 ABI 结构版本。
	AbiStructureVersion uint32 `json:"abi_structure_version"`
	// Root commands implemented by the exhaustive dispatcher.
	// 穷尽分发器实现的根命令。
	Commands EmbeddedOutputTransportDescriptionCommands `json:"commands"`
	// Cargo package version of this exact library build.
	// 此精确动态库构建的 Cargo 包版本。
	CoreVersion string `json:"core_version"`
	// Actual validated transport limits supplied at construction.
	// 构造时提供的实际已校验传输边界。
	Limits EmbeddedOutputTransportConfig `json:"limits"`
	// Version of the accepted embedded JSON protocol.
	// 接受的嵌入式 JSON 协议版本。
	ProtocolVersion uint32 `json:"protocol_version"`
	// Runtime operations implemented by the exhaustive dispatcher.
	// 穷尽分发器实现的运行时操作。
	RuntimeCommands EmbeddedOutputTransportDescriptionRuntimeCommands `json:"runtime_commands"`
}

// Root commands implemented by the exhaustive dispatcher.
// 穷尽分发器实现的根命令。
type EmbeddedOutputTransportDescriptionCommands []string

// Runtime operations implemented by the exhaustive dispatcher.
// 穷尽分发器实现的运行时操作。
type EmbeddedOutputTransportDescriptionRuntimeCommands []string

// embeddedWireShapes is immutable metadata for generated data projection.
// embeddedWireShapes 是生成数据投影的不可变元数据。
var embeddedWireShapes = map[reflect.Type]embeddedWireShape{
	embeddedWireType[EmbeddedInputCapabilityDescriptor]():                             {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputCapabilityDescriptorPermissions]():                  {kind: "array", unique: true},
	embeddedWireType[EmbeddedInputCapabilityEffects]():                                {kind: "enum", values: []string{"read_only", "mutating"}},
	embeddedWireType[EmbeddedInputCapabilityExecution]():                              {kind: "enum", values: []string{"native", "queued"}},
	embeddedWireType[EmbeddedInputCapabilityIdempotency]():                            {kind: "enum", values: []string{"none", "host_request"}},
	embeddedWireType[EmbeddedInputCapabilityScope]():                                  {kind: "enum", values: []string{"invocation", "session"}},
	embeddedWireType[EmbeddedInputCommand]():                                          {kind: "union", alternatives: []reflect.Type{embeddedWireType[EmbeddedInputCommandRuntime](), embeddedWireType[EmbeddedInputCommandDescribe](), embeddedWireType[EmbeddedInputCommandRuntimeReserve](), embeddedWireType[EmbeddedInputCommandRuntimeInitialize](), embeddedWireType[EmbeddedInputCommandRuntimeStatus](), embeddedWireType[EmbeddedInputCommandRuntimeClose](), embeddedWireType[EmbeddedInputCommandRuntimeFree]()}},
	embeddedWireType[EmbeddedInputCommandDescribe]():                                  {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputCommandDescribeType]():                              {kind: "enum", values: []string{"describe"}},
	embeddedWireType[EmbeddedInputCommandRuntime]():                                   {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputCommandRuntimeClose]():                              {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputCommandRuntimeCloseType]():                          {kind: "enum", values: []string{"runtime_close"}},
	embeddedWireType[EmbeddedInputCommandRuntimeFree]():                               {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputCommandRuntimeFreeType]():                           {kind: "enum", values: []string{"runtime_free"}},
	embeddedWireType[EmbeddedInputCommandRuntimeInitialize]():                         {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputCommandRuntimeInitializeType]():                     {kind: "enum", values: []string{"runtime_initialize"}},
	embeddedWireType[EmbeddedInputCommandRuntimeReserve]():                            {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputCommandRuntimeReserveType]():                        {kind: "enum", values: []string{"runtime_reserve"}},
	embeddedWireType[EmbeddedInputCommandRuntimeStatus]():                             {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputCommandRuntimeStatusType]():                         {kind: "enum", values: []string{"runtime_status"}},
	embeddedWireType[EmbeddedInputCommandRuntimeType]():                               {kind: "enum", values: []string{"runtime"}},
	embeddedWireType[EmbeddedInputEffectState]():                                      {kind: "enum", values: []string{"not_started", "not_applicable", "committed", "rolled_back", "unknown"}},
	embeddedWireType[EmbeddedInputEmbeddedCall]():                                     {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputEmbeddedError]():                                    {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputEmbeddedErrorCode]():                                {kind: "enum", values: []string{"invalid_argument", "not_found", "stale_generation", "capacity_exceeded", "busy", "already_completed", "closed", "cancelled", "deadline_exceeded", "permission_denied", "unsupported", "execution_failed", "cleanup_failed", "internal"}},
	embeddedWireType[EmbeddedInputEmbeddedPluginConfig]():                             {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputEmbeddedRuntimeConfig]():                            {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputExecutionBackend]():                                 {kind: "enum", values: []string{"in_process", "worker_process"}},
	embeddedWireType[EmbeddedInputHistoryCursor]():                                    {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputHostCompletion]():                                   {kind: "union", alternatives: []reflect.Type{embeddedWireType[EmbeddedInputHostCompletionVariant1](), embeddedWireType[EmbeddedInputHostCompletionVariant2]()}},
	embeddedWireType[EmbeddedInputHostCompletionVariant1]():                           {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputHostCompletionVariant2]():                           {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputHostEffectReconciliation]():                         {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputInstanceReuse]():                                    {kind: "enum", values: []string{"single_call", "reusable", "session"}},
	embeddedWireType[EmbeddedInputLuaEngineOptions]():                                 {kind: "object", additional: true},
	embeddedWireType[EmbeddedInputLuaInvocationContext]():                             {kind: "object", additional: true},
	embeddedWireType[EmbeddedInputLuaRuntimeCapabilityOptions]():                      {kind: "object", additional: true},
	embeddedWireType[EmbeddedInputLuaRuntimeDatabaseCallbackMode]():                   {kind: "enum", values: []string{"standard", "json"}},
	embeddedWireType[EmbeddedInputLuaRuntimeDatabaseProviderMode]():                   {kind: "enum", values: []string{"dynamic_library", "host_callback", "space_controller"}},
	embeddedWireType[EmbeddedInputLuaRuntimeHostOptions]():                            {kind: "object", additional: true},
	embeddedWireType[EmbeddedInputLuaRuntimeHostOptionsIgnoredSkillIds]():             {kind: "array", unique: false},
	embeddedWireType[EmbeddedInputLuaRuntimeHostOptionsPrivateSkillSourceAllowlist](): {kind: "array", unique: false},
	embeddedWireType[EmbeddedInputLuaRuntimeHostOptionsReservedEntryNames]():          {kind: "array", unique: false},
	embeddedWireType[EmbeddedInputLuaRuntimeManagedRuntimeConfig]():                   {kind: "object", additional: true},
	embeddedWireType[EmbeddedInputLuaRuntimeRunLuaPoolConfig]():                       {kind: "object", additional: true},
	embeddedWireType[EmbeddedInputLuaRuntimeSpaceControllerOptions]():                 {kind: "object", additional: true},
	embeddedWireType[EmbeddedInputLuaRuntimeSpaceControllerProcessMode]():             {kind: "enum", values: []string{"service", "managed"}},
	embeddedWireType[EmbeddedInputLuaVmPoolConfig]():                                  {kind: "object", additional: true},
	embeddedWireType[EmbeddedInputModuleDefinition]():                                 {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputModuleDefinitionExports]():                          {kind: "array", unique: false},
	embeddedWireType[EmbeddedInputModuleExport]():                                     {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputOperationJournalConfig]():                           {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputOperationJournalWorkerConfig]():                     {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputOperationReconciliation]():                          {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputOperationReconciliationHostEffects]():               {kind: "array", unique: false},
	embeddedWireType[EmbeddedInputPluginPoolConfig]():                                 {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputPoolKind]():                                         {kind: "enum", values: []string{"shared", "dedicated"}},
	embeddedWireType[EmbeddedInputReconciledExecution]():                              {kind: "enum", values: []string{"observed_terminal", "stopped_without_result"}},
	embeddedWireType[EmbeddedInputRequest]():                                          {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputResolvedEffectState]():                              {kind: "enum", values: []string{"not_started", "not_applicable", "committed", "rolled_back"}},
	embeddedWireType[EmbeddedInputRuntimeClientInfo]():                                {kind: "object", additional: true},
	embeddedWireType[EmbeddedInputRuntimeCommand]():                                   {kind: "union", alternatives: []reflect.Type{embeddedWireType[EmbeddedInputRuntimeCommandOperationPersistenceFailure](), embeddedWireType[EmbeddedInputRuntimeCommandOperationRetryCheckpoint](), embeddedWireType[EmbeddedInputRuntimeCommandStorageStatus](), embeddedWireType[EmbeddedInputRuntimeCommandStorageRecover](), embeddedWireType[EmbeddedInputRuntimeCommandHistoryGet](), embeddedWireType[EmbeddedInputRuntimeCommandHistoryNext](), embeddedWireType[EmbeddedInputRuntimeCommandHistoryReconcile](), embeddedWireType[EmbeddedInputRuntimeCommandHistoryForget](), embeddedWireType[EmbeddedInputRuntimeCommandPluginRegister](), embeddedWireType[EmbeddedInputRuntimeCommandPluginStatus](), embeddedWireType[EmbeddedInputRuntimeCommandPluginClose](), embeddedWireType[EmbeddedInputRuntimeCommandPluginForget](), embeddedWireType[EmbeddedInputRuntimeCommandPoolRegister](), embeddedWireType[EmbeddedInputRuntimeCommandPoolStatus](), embeddedWireType[EmbeddedInputRuntimeCommandPoolClose](), embeddedWireType[EmbeddedInputRuntimeCommandPoolForget](), embeddedWireType[EmbeddedInputRuntimeCommandPoolRevokePermission](), embeddedWireType[EmbeddedInputRuntimeCommandCallSubmit](), embeddedWireType[EmbeddedInputRuntimeCommandSessionOpen](), embeddedWireType[EmbeddedInputRuntimeCommandSessionSubmit](), embeddedWireType[EmbeddedInputRuntimeCommandSessionStatus](), embeddedWireType[EmbeddedInputRuntimeCommandSessionClose](), embeddedWireType[EmbeddedInputRuntimeCommandSessionForget](), embeddedWireType[EmbeddedInputRuntimeCommandOperationStatus](), embeddedWireType[EmbeddedInputRuntimeCommandOperationWait](), embeddedWireType[EmbeddedInputRuntimeCommandOperationCancel](), embeddedWireType[EmbeddedInputRuntimeCommandOperationForget](), embeddedWireType[EmbeddedInputRuntimeCommandCapabilitiesRegister](), embeddedWireType[EmbeddedInputRuntimeCommandCapabilitiesList](), embeddedWireType[EmbeddedInputRuntimeCommandCapabilityStatus](), embeddedWireType[EmbeddedInputRuntimeCommandCapabilityUnregister](), embeddedWireType[EmbeddedInputRuntimeCommandCapabilityForget](), embeddedWireType[EmbeddedInputRuntimeCommandHostRequestsTake](), embeddedWireType[EmbeddedInputRuntimeCommandHostRequestStatus](), embeddedWireType[EmbeddedInputRuntimeCommandHostRequestComplete]()}},
	embeddedWireType[EmbeddedInputRuntimeCommandCallSubmit]():                         {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandCallSubmitType]():                     {kind: "enum", values: []string{"call_submit"}},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilitiesList]():                   {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilitiesListPermissions]():        {kind: "array", unique: true},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilitiesListType]():               {kind: "enum", values: []string{"capabilities_list"}},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilitiesRegister]():               {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilitiesRegisterDescriptors]():    {kind: "array", unique: false},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilitiesRegisterType]():           {kind: "enum", values: []string{"capabilities_register"}},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilityForget]():                   {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilityForgetType]():               {kind: "enum", values: []string{"capability_forget"}},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilityStatus]():                   {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilityStatusType]():               {kind: "enum", values: []string{"capability_status"}},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilityUnregister]():               {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandCapabilityUnregisterType]():           {kind: "enum", values: []string{"capability_unregister"}},
	embeddedWireType[EmbeddedInputRuntimeCommandHistoryForget]():                      {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandHistoryForgetType]():                  {kind: "enum", values: []string{"history_forget"}},
	embeddedWireType[EmbeddedInputRuntimeCommandHistoryGet]():                         {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandHistoryGetType]():                     {kind: "enum", values: []string{"history_get"}},
	embeddedWireType[EmbeddedInputRuntimeCommandHistoryNext]():                        {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandHistoryNextType]():                    {kind: "enum", values: []string{"history_next"}},
	embeddedWireType[EmbeddedInputRuntimeCommandHistoryReconcile]():                   {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandHistoryReconcileType]():               {kind: "enum", values: []string{"history_reconcile"}},
	embeddedWireType[EmbeddedInputRuntimeCommandHostRequestComplete]():                {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandHostRequestCompleteType]():            {kind: "enum", values: []string{"host_request_complete"}},
	embeddedWireType[EmbeddedInputRuntimeCommandHostRequestStatus]():                  {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandHostRequestStatusType]():              {kind: "enum", values: []string{"host_request_status"}},
	embeddedWireType[EmbeddedInputRuntimeCommandHostRequestsTake]():                   {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandHostRequestsTakeType]():               {kind: "enum", values: []string{"host_requests_take"}},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationCancel]():                    {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationCancelType]():                {kind: "enum", values: []string{"operation_cancel"}},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationForget]():                    {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationForgetType]():                {kind: "enum", values: []string{"operation_forget"}},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationPersistenceFailure]():        {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationPersistenceFailureType]():    {kind: "enum", values: []string{"operation_persistence_failure"}},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationRetryCheckpoint]():           {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationRetryCheckpointType]():       {kind: "enum", values: []string{"operation_retry_checkpoint"}},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationStatus]():                    {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationStatusType]():                {kind: "enum", values: []string{"operation_status"}},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationWait]():                      {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandOperationWaitType]():                  {kind: "enum", values: []string{"operation_wait"}},
	embeddedWireType[EmbeddedInputRuntimeCommandPluginClose]():                        {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandPluginCloseType]():                    {kind: "enum", values: []string{"plugin_close"}},
	embeddedWireType[EmbeddedInputRuntimeCommandPluginForget]():                       {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandPluginForgetType]():                   {kind: "enum", values: []string{"plugin_forget"}},
	embeddedWireType[EmbeddedInputRuntimeCommandPluginRegister]():                     {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandPluginRegisterType]():                 {kind: "enum", values: []string{"plugin_register"}},
	embeddedWireType[EmbeddedInputRuntimeCommandPluginStatus]():                       {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandPluginStatusType]():                   {kind: "enum", values: []string{"plugin_status"}},
	embeddedWireType[EmbeddedInputRuntimeCommandPoolClose]():                          {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandPoolCloseType]():                      {kind: "enum", values: []string{"pool_close"}},
	embeddedWireType[EmbeddedInputRuntimeCommandPoolForget]():                         {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandPoolForgetType]():                     {kind: "enum", values: []string{"pool_forget"}},
	embeddedWireType[EmbeddedInputRuntimeCommandPoolRegister]():                       {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandPoolRegisterPermissions]():            {kind: "array", unique: true},
	embeddedWireType[EmbeddedInputRuntimeCommandPoolRegisterType]():                   {kind: "enum", values: []string{"pool_register"}},
	embeddedWireType[EmbeddedInputRuntimeCommandPoolRevokePermission]():               {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandPoolRevokePermissionType]():           {kind: "enum", values: []string{"pool_revoke_permission"}},
	embeddedWireType[EmbeddedInputRuntimeCommandPoolStatus]():                         {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandPoolStatusType]():                     {kind: "enum", values: []string{"pool_status"}},
	embeddedWireType[EmbeddedInputRuntimeCommandSessionClose]():                       {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandSessionCloseType]():                   {kind: "enum", values: []string{"session_close"}},
	embeddedWireType[EmbeddedInputRuntimeCommandSessionForget]():                      {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandSessionForgetType]():                  {kind: "enum", values: []string{"session_forget"}},
	embeddedWireType[EmbeddedInputRuntimeCommandSessionOpen]():                        {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandSessionOpenType]():                    {kind: "enum", values: []string{"session_open"}},
	embeddedWireType[EmbeddedInputRuntimeCommandSessionStatus]():                      {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandSessionStatusType]():                  {kind: "enum", values: []string{"session_status"}},
	embeddedWireType[EmbeddedInputRuntimeCommandSessionSubmit]():                      {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandSessionSubmitType]():                  {kind: "enum", values: []string{"session_submit"}},
	embeddedWireType[EmbeddedInputRuntimeCommandStorageRecover]():                     {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandStorageRecoverType]():                 {kind: "enum", values: []string{"storage_recover"}},
	embeddedWireType[EmbeddedInputRuntimeCommandStorageStatus]():                      {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeCommandStorageStatusType]():                  {kind: "enum", values: []string{"storage_status"}},
	embeddedWireType[EmbeddedInputRuntimePersistenceConfig]():                         {kind: "object", additional: false},
	embeddedWireType[EmbeddedInputRuntimeRequestContext]():                            {kind: "object", additional: true},
	embeddedWireType[EmbeddedInputToolCacheConfig]():                                  {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputCapabilityCaller]():                                {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputCapabilityDescriptor]():                            {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputCapabilityDescriptorPermissions]():                 {kind: "array", unique: true},
	embeddedWireType[EmbeddedOutputCapabilityEffects]():                               {kind: "enum", values: []string{"read_only", "mutating"}},
	embeddedWireType[EmbeddedOutputCapabilityExecution]():                             {kind: "enum", values: []string{"native", "queued"}},
	embeddedWireType[EmbeddedOutputCapabilityIdempotency]():                           {kind: "enum", values: []string{"none", "host_request"}},
	embeddedWireType[EmbeddedOutputCapabilityRegistrationStatus]():                    {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputCapabilityScope]():                                 {kind: "enum", values: []string{"invocation", "session"}},
	embeddedWireType[EmbeddedOutputCheckpointRetryState]():                            {kind: "enum", values: []string{"waiting", "requested", "retrying"}},
	embeddedWireType[EmbeddedOutputCoreDescription]():                                 {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputCoreDescriptionCapabilities]():                     {kind: "array", unique: false},
	embeddedWireType[EmbeddedOutputCoreDescriptionCommands]():                         {kind: "array", unique: false},
	embeddedWireType[EmbeddedOutputCoreDescriptionExecutionBackends]():                {kind: "array", unique: false},
	embeddedWireType[EmbeddedOutputCoreDescriptionRuntimeCommands]():                  {kind: "array", unique: false},
	embeddedWireType[EmbeddedOutputEffectState]():                                     {kind: "enum", values: []string{"not_started", "not_applicable", "committed", "rolled_back", "unknown"}},
	embeddedWireType[EmbeddedOutputEmbeddedBuildIdentity]():                           {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputEmbeddedBuildIdentityCargoFeatures]():              {kind: "array", unique: false},
	embeddedWireType[EmbeddedOutputEmbeddedError]():                                   {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputEmbeddedErrorCode]():                               {kind: "enum", values: []string{"invalid_argument", "not_found", "stale_generation", "capacity_exceeded", "busy", "already_completed", "closed", "cancelled", "deadline_exceeded", "permission_denied", "unsupported", "execution_failed", "cleanup_failed", "internal"}},
	embeddedWireType[EmbeddedOutputEmbeddedPluginConfig]():                            {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputEmbeddedPluginSnapshot]():                          {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputEmbeddedRuntimeUsage]():                            {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputEmbeddedSessionPhase]():                            {kind: "enum", values: []string{"opening", "ready", "running", "closing", "closed"}},
	embeddedWireType[EmbeddedOutputEmbeddedSessionSnapshot]():                         {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputErrorResponse]():                                   {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputErrorStatus]():                                     {kind: "enum", values: []string{"error"}},
	embeddedWireType[EmbeddedOutputExecutionBackend]():                                {kind: "enum", values: []string{"in_process", "worker_process"}},
	embeddedWireType[EmbeddedOutputHostEffectPhase]():                                 {kind: "enum", values: []string{"prepared", "running", "completed"}},
	embeddedWireType[EmbeddedOutputHostEffectReconciliation]():                        {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputHostEffectRecord]():                                {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputHostRequest]():                                     {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputHostRequestPhase]():                                {kind: "enum", values: []string{"queued", "dispatched", "completing", "completed"}},
	embeddedWireType[EmbeddedOutputHostRequestStatus]():                               {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputInitializationPhase]():                             {kind: "enum", values: []string{"reserved", "initializing", "ready", "failed", "faulted"}},
	embeddedWireType[EmbeddedOutputJournalOperation]():                                {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputOperationContext]():                                {kind: "union", alternatives: []reflect.Type{embeddedWireType[EmbeddedOutputOperationContextVariant1](), embeddedWireType[EmbeddedOutputOperationContextVariant2]()}},
	embeddedWireType[EmbeddedOutputOperationContextVariant1]():                        {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputOperationContextVariant1Kind]():                    {kind: "enum", values: []string{"unbound"}},
	embeddedWireType[EmbeddedOutputOperationContextVariant2]():                        {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputOperationContextVariant2Kind]():                    {kind: "enum", values: []string{"module"}},
	embeddedWireType[EmbeddedOutputOperationJournalWorkerStatus]():                    {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputOperationPersistenceFailure]():                     {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputOperationPhase]():                                  {kind: "enum", values: []string{"queued", "initializing", "running", "waiting_for_host", "cleaning", "succeeded", "failed", "cancelled"}},
	embeddedWireType[EmbeddedOutputOperationReceipt]():                                {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputOperationReconciliation]():                         {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputOperationReconciliationHostEffects]():              {kind: "array", unique: false},
	embeddedWireType[EmbeddedOutputOperationSnapshot]():                               {kind: "object", additional: false},
	embeddedWireType[EmbeddedOutputOperationSnapshotHostEffects]():                    {kind: "array", unique: false},
	embeddedWireType[EmbeddedOutputPoolReceipt]():                                     {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputPoolUsage]():                                       {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputReconciledExecution]():                             {kind: "enum", values: []string{"observed_terminal", "stopped_without_result"}},
	embeddedWireType[EmbeddedOutputRegistrationReceipt]():                             {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRegistrationReceiptRegistrationIds]():              {kind: "array", unique: false},
	embeddedWireType[EmbeddedOutputResolvedEffectState]():                             {kind: "enum", values: []string{"not_started", "not_applicable", "committed", "rolled_back"}},
	embeddedWireType[EmbeddedOutputRootDescribeResponse]():                            {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRootRuntimeCloseResponse]():                        {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRootRuntimeFreeResponse]():                         {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRootRuntimeInitializeResponse]():                   {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRootRuntimeReserveResponse]():                      {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRootRuntimeStatusResponse]():                       {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeCallSubmitResponse]():                       {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeCapabilitiesListResponse]():                 {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeCapabilitiesListResponseResult]():           {kind: "array", unique: false},
	embeddedWireType[EmbeddedOutputRuntimeCapabilitiesRegisterResponse]():             {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeCapabilityForgetResponse]():                 {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeCapabilityStatusResponse]():                 {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeCapabilityUnregisterResponse]():             {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeHistoryForgetResponse]():                    {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeHistoryGetResponse]():                       {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeHistoryNextResponse]():                      {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeHistoryReconcileResponse]():                 {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeHostRequestCompleteResponse]():              {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeHostRequestStatusResponse]():                {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeHostRequestsTakeResponse]():                 {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeHostRequestsTakeResponseResult]():           {kind: "array", unique: false},
	embeddedWireType[EmbeddedOutputRuntimeOperationCancelResponse]():                  {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeOperationForgetResponse]():                  {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeOperationPersistenceFailureResponse]():      {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeOperationRetryCheckpointResponse]():         {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeOperationStatusResponse]():                  {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeOperationWaitResponse]():                    {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimePluginCloseResponse]():                      {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimePluginForgetResponse]():                     {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimePluginRegisterResponse]():                   {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimePluginStatusResponse]():                     {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimePoolCloseResponse]():                        {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimePoolForgetResponse]():                       {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimePoolRegisterResponse]():                     {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimePoolRevokePermissionResponse]():             {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimePoolStatusResponse]():                       {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeReceipt]():                                  {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeSessionCloseResponse]():                     {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeSessionForgetResponse]():                    {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeSessionOpenResponse]():                      {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeSessionStatusResponse]():                    {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeSessionSubmitResponse]():                    {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeSnapshot]():                                 {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeStorageRecoverResponse]():                   {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputRuntimeStorageStatusResponse]():                    {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputSessionReceipt]():                                  {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputSuccessStatus]():                                   {kind: "enum", values: []string{"ok"}},
	embeddedWireType[EmbeddedOutputTransportConfig]():                                 {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputTransportDescription]():                            {kind: "object", additional: true},
	embeddedWireType[EmbeddedOutputTransportDescriptionCommands]():                    {kind: "array", unique: false},
	embeddedWireType[EmbeddedOutputTransportDescriptionRuntimeCommands]():             {kind: "array", unique: false},
}

// DecodeEmbeddedOutputCoreDescription validates standalone borrowed descriptor bytes without assuming an envelope.
// DecodeEmbeddedOutputCoreDescription 校验独立借用型描述字节，不假定信封。
func DecodeEmbeddedOutputCoreDescription(bytes []byte) (EmbeddedOutputCoreDescription, error) {
	return decodeEmbeddedWireValue[EmbeddedOutputCoreDescription](bytes)
}

// DecodeEmbeddedOutputErrorResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputErrorResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputErrorResponse(bytes []byte) (EmbeddedOutputErrorResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputErrorResponse](bytes)
}

// DecodeEmbeddedOutputRootDescribeResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRootDescribeResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRootDescribeResponse(bytes []byte) (EmbeddedOutputRootDescribeResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRootDescribeResponse](bytes)
}

// DecodeEmbeddedOutputRootRuntimeCloseResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRootRuntimeCloseResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRootRuntimeCloseResponse(bytes []byte) (EmbeddedOutputRootRuntimeCloseResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRootRuntimeCloseResponse](bytes)
}

// DecodeEmbeddedOutputRootRuntimeFreeResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRootRuntimeFreeResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRootRuntimeFreeResponse(bytes []byte) (EmbeddedOutputRootRuntimeFreeResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRootRuntimeFreeResponse](bytes)
}

// DecodeEmbeddedOutputRootRuntimeInitializeResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRootRuntimeInitializeResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRootRuntimeInitializeResponse(bytes []byte) (EmbeddedOutputRootRuntimeInitializeResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRootRuntimeInitializeResponse](bytes)
}

// DecodeEmbeddedOutputRootRuntimeReserveResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRootRuntimeReserveResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRootRuntimeReserveResponse(bytes []byte) (EmbeddedOutputRootRuntimeReserveResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRootRuntimeReserveResponse](bytes)
}

// DecodeEmbeddedOutputRootRuntimeStatusResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRootRuntimeStatusResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRootRuntimeStatusResponse(bytes []byte) (EmbeddedOutputRootRuntimeStatusResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRootRuntimeStatusResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeCallSubmitResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeCallSubmitResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeCallSubmitResponse(bytes []byte) (EmbeddedOutputRuntimeCallSubmitResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeCallSubmitResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeCapabilitiesListResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeCapabilitiesListResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeCapabilitiesListResponse(bytes []byte) (EmbeddedOutputRuntimeCapabilitiesListResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeCapabilitiesListResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeCapabilitiesRegisterResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeCapabilitiesRegisterResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeCapabilitiesRegisterResponse(bytes []byte) (EmbeddedOutputRuntimeCapabilitiesRegisterResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeCapabilitiesRegisterResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeCapabilityForgetResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeCapabilityForgetResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeCapabilityForgetResponse(bytes []byte) (EmbeddedOutputRuntimeCapabilityForgetResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeCapabilityForgetResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeCapabilityStatusResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeCapabilityStatusResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeCapabilityStatusResponse(bytes []byte) (EmbeddedOutputRuntimeCapabilityStatusResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeCapabilityStatusResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeCapabilityUnregisterResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeCapabilityUnregisterResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeCapabilityUnregisterResponse(bytes []byte) (EmbeddedOutputRuntimeCapabilityUnregisterResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeCapabilityUnregisterResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeHistoryForgetResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeHistoryForgetResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeHistoryForgetResponse(bytes []byte) (EmbeddedOutputRuntimeHistoryForgetResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeHistoryForgetResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeHistoryGetResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeHistoryGetResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeHistoryGetResponse(bytes []byte) (EmbeddedOutputRuntimeHistoryGetResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeHistoryGetResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeHistoryNextResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeHistoryNextResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeHistoryNextResponse(bytes []byte) (EmbeddedOutputRuntimeHistoryNextResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeHistoryNextResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeHistoryReconcileResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeHistoryReconcileResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeHistoryReconcileResponse(bytes []byte) (EmbeddedOutputRuntimeHistoryReconcileResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeHistoryReconcileResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeHostRequestCompleteResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeHostRequestCompleteResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeHostRequestCompleteResponse(bytes []byte) (EmbeddedOutputRuntimeHostRequestCompleteResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeHostRequestCompleteResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeHostRequestStatusResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeHostRequestStatusResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeHostRequestStatusResponse(bytes []byte) (EmbeddedOutputRuntimeHostRequestStatusResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeHostRequestStatusResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeHostRequestsTakeResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeHostRequestsTakeResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeHostRequestsTakeResponse(bytes []byte) (EmbeddedOutputRuntimeHostRequestsTakeResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeHostRequestsTakeResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeOperationCancelResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeOperationCancelResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeOperationCancelResponse(bytes []byte) (EmbeddedOutputRuntimeOperationCancelResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeOperationCancelResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeOperationForgetResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeOperationForgetResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeOperationForgetResponse(bytes []byte) (EmbeddedOutputRuntimeOperationForgetResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeOperationForgetResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeOperationPersistenceFailureResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeOperationPersistenceFailureResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeOperationPersistenceFailureResponse(bytes []byte) (EmbeddedOutputRuntimeOperationPersistenceFailureResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeOperationPersistenceFailureResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeOperationRetryCheckpointResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeOperationRetryCheckpointResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeOperationRetryCheckpointResponse(bytes []byte) (EmbeddedOutputRuntimeOperationRetryCheckpointResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeOperationRetryCheckpointResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeOperationStatusResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeOperationStatusResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeOperationStatusResponse(bytes []byte) (EmbeddedOutputRuntimeOperationStatusResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeOperationStatusResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeOperationWaitResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeOperationWaitResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeOperationWaitResponse(bytes []byte) (EmbeddedOutputRuntimeOperationWaitResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeOperationWaitResponse](bytes)
}

// DecodeEmbeddedOutputRuntimePluginCloseResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimePluginCloseResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimePluginCloseResponse(bytes []byte) (EmbeddedOutputRuntimePluginCloseResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimePluginCloseResponse](bytes)
}

// DecodeEmbeddedOutputRuntimePluginForgetResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimePluginForgetResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimePluginForgetResponse(bytes []byte) (EmbeddedOutputRuntimePluginForgetResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimePluginForgetResponse](bytes)
}

// DecodeEmbeddedOutputRuntimePluginRegisterResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimePluginRegisterResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimePluginRegisterResponse(bytes []byte) (EmbeddedOutputRuntimePluginRegisterResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimePluginRegisterResponse](bytes)
}

// DecodeEmbeddedOutputRuntimePluginStatusResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimePluginStatusResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimePluginStatusResponse(bytes []byte) (EmbeddedOutputRuntimePluginStatusResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimePluginStatusResponse](bytes)
}

// DecodeEmbeddedOutputRuntimePoolCloseResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimePoolCloseResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimePoolCloseResponse(bytes []byte) (EmbeddedOutputRuntimePoolCloseResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimePoolCloseResponse](bytes)
}

// DecodeEmbeddedOutputRuntimePoolForgetResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimePoolForgetResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimePoolForgetResponse(bytes []byte) (EmbeddedOutputRuntimePoolForgetResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimePoolForgetResponse](bytes)
}

// DecodeEmbeddedOutputRuntimePoolRegisterResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimePoolRegisterResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimePoolRegisterResponse(bytes []byte) (EmbeddedOutputRuntimePoolRegisterResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimePoolRegisterResponse](bytes)
}

// DecodeEmbeddedOutputRuntimePoolRevokePermissionResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimePoolRevokePermissionResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimePoolRevokePermissionResponse(bytes []byte) (EmbeddedOutputRuntimePoolRevokePermissionResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimePoolRevokePermissionResponse](bytes)
}

// DecodeEmbeddedOutputRuntimePoolStatusResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimePoolStatusResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimePoolStatusResponse(bytes []byte) (EmbeddedOutputRuntimePoolStatusResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimePoolStatusResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeSessionCloseResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeSessionCloseResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeSessionCloseResponse(bytes []byte) (EmbeddedOutputRuntimeSessionCloseResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeSessionCloseResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeSessionForgetResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeSessionForgetResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeSessionForgetResponse(bytes []byte) (EmbeddedOutputRuntimeSessionForgetResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeSessionForgetResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeSessionOpenResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeSessionOpenResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeSessionOpenResponse(bytes []byte) (EmbeddedOutputRuntimeSessionOpenResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeSessionOpenResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeSessionStatusResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeSessionStatusResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeSessionStatusResponse(bytes []byte) (EmbeddedOutputRuntimeSessionStatusResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeSessionStatusResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeSessionSubmitResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeSessionSubmitResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeSessionSubmitResponse(bytes []byte) (EmbeddedOutputRuntimeSessionSubmitResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeSessionSubmitResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeStorageRecoverResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeStorageRecoverResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeStorageRecoverResponse(bytes []byte) (EmbeddedOutputRuntimeStorageRecoverResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeStorageRecoverResponse](bytes)
}

// DecodeEmbeddedOutputRuntimeStorageStatusResponse validates bytes and preserves required fields, nulls and exact numeric values.
// DecodeEmbeddedOutputRuntimeStorageStatusResponse 校验 bytes，并保留必需字段、空值及精确数值。
// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.
// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。
func DecodeEmbeddedOutputRuntimeStorageStatusResponse(bytes []byte) (EmbeddedOutputRuntimeStorageStatusResponse, error) {
	return decodeEmbeddedWireEnvelope[EmbeddedOutputRuntimeStorageStatusResponse](bytes)
}
