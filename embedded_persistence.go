package luaskills

import "context"

// InitializePersistent admits explicit host storage under ctx; query Status for the actual one-shot construction result.
// InitializePersistent 在 ctx 下接纳显式宿主存储；实际单次构造结果须查询 Status。
// options and budgets retain their existing authority; persistence selects the absolute database and bounded writer.
// options 与 budgets 保留既有权威；persistence 选择绝对数据库路径和有界写入者。
func (r *EmbeddedRuntime) InitializePersistent(ctx context.Context, options EmbeddedInputLuaEngineOptions, budgets EmbeddedInputEmbeddedRuntimeConfig, persistence EmbeddedInputRuntimePersistenceConfig) (*EmbeddedPending[EmbeddedOutputRuntimeReceipt], error) {
	// configured encodes an explicitly present non-null declaration using the generated optional representation.
	// configured 使用生成的可选表示编码显式存在且非空的声明。
	configured := &persistence
	return submitEmbedded(ctx, r.client, EmbeddedInputCommandRuntimeInitialize{Type: EmbeddedInputCommandRuntimeInitializeTypeRuntimeInitialize, RuntimeId: r.identity, EngineOptions: options, RuntimeConfig: budgets, Persistence: &configured}, projectEmbeddedResult[EmbeddedOutputRuntimeReceipt])
}

// StorageStatus returns actual writer ownership through the control lane under ctx; memory mode reports unsupported.
// StorageStatus 在 ctx 下通过控制通道返回实际写入者所有权；内存模式报告不支持。
func (r *EmbeddedRuntime) StorageStatus(ctx context.Context) (*EmbeddedPending[EmbeddedOutputOperationJournalWorkerStatus], error) {
	return submitEmbeddedRuntime(ctx, r, EmbeddedInputRuntimeCommandStorageStatus{Type: EmbeddedInputRuntimeCommandStorageStatusTypeStorageStatus}, projectEmbeddedResult[EmbeddedOutputOperationJournalWorkerStatus])
}

// RecoverStorage reopens and validates the same failed database on the work lane; result states whether recovery was needed.
// RecoverStorage 在工作通道重新打开并校验同一故障数据库；结果表示是否需要恢复。
// ctx only bounds admission; this neither retries checkpoints nor executes plugin or host business code.
// ctx 仅约束入场；此操作不重试检查点，也不执行插件或宿主业务代码。
func (r *EmbeddedRuntime) RecoverStorage(ctx context.Context) (*EmbeddedPending[bool], error) {
	return submitEmbeddedRuntime(ctx, r, EmbeddedInputRuntimeCommandStorageRecover{Type: EmbeddedInputRuntimeCommandStorageRecoverTypeStorageRecover}, projectEmbeddedResult[bool])
}

// HistoryGet reads the exact original historyRuntimeID and operationID under ctx; nil does not prove execution never occurred.
// HistoryGet 在 ctx 下读取精确原始 historyRuntimeID 和 operationID；空值不证明从未执行。
func (r *EmbeddedRuntime) HistoryGet(ctx context.Context, historyRuntimeID, operationID string) (*EmbeddedPending[*EmbeddedOutputJournalOperation], error) {
	return submitEmbeddedRuntime(ctx, r, EmbeddedInputRuntimeCommandHistoryGet{Type: EmbeddedInputRuntimeCommandHistoryGetTypeHistoryGet, HistoryRuntimeId: historyRuntimeID, OperationId: operationID}, projectEmbeddedResult[*EmbeddedOutputJournalOperation])
}

// HistoryNext reads one row after the exact original cursor under ctx; nil after starts enumeration and nil result ends it.
// HistoryNext 在 ctx 下读取精确原始游标之后一行；after 为空表示开始枚举，结果为空表示结束。
// Concurrent edits are not a multi-call snapshot; old records never become current operation handles.
// 并发编辑不构成跨调用快照；旧记录绝不会成为当前操作句柄。
func (r *EmbeddedRuntime) HistoryNext(ctx context.Context, after *EmbeddedInputHistoryCursor) (*EmbeddedPending[*EmbeddedOutputJournalOperation], error) {
	return submitEmbeddedRuntime(ctx, r, EmbeddedInputRuntimeCommandHistoryNext{Type: EmbeddedInputRuntimeCommandHistoryNextTypeHistoryNext, After: &after}, projectEmbeddedResult[*EmbeddedOutputJournalOperation])
}

// HistoryForget removes reconciled terminal history at expectedRevision under ctx; first forget any retained live operation.
// HistoryForget 在 ctx 下按 expectedRevision 移除已对账终态历史；需先遗忘仍保留的活动操作。
// Return the work-lane deletion receipt; stale revisions and unresolved effects retain the original record.
// 返回工作通道删除回执；过期修订和未决副作用保留原始记录。
func (r *EmbeddedRuntime) HistoryForget(ctx context.Context, historyRuntimeID, operationID string, expectedRevision uint64) (*EmbeddedPending[*EmbeddedJSONNull], error) {
	return submitEmbeddedRuntime(ctx, r, EmbeddedInputRuntimeCommandHistoryForget{Type: EmbeddedInputRuntimeCommandHistoryForgetTypeHistoryForget, HistoryRuntimeId: historyRuntimeID, OperationId: operationID, ExpectedRevision: expectedRevision}, projectEmbeddedResult[*EmbeddedJSONNull])
}

// PersistenceFailure returns this operation's retained checkpoint failure under ctx without disk waits or implicit retries.
// PersistenceFailure 在 ctx 下返回此操作保留的检查点故障，不等待磁盘，也不隐式重试。
func (h *EmbeddedOperation) PersistenceFailure(ctx context.Context) (*EmbeddedPending[*EmbeddedOutputOperationPersistenceFailure], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandOperationPersistenceFailure{Type: EmbeddedInputRuntimeCommandOperationPersistenceFailureTypeOperationPersistenceFailure, OperationId: h.identity}, projectEmbeddedResult[*EmbeddedOutputOperationPersistenceFailure])
}

// RetryCheckpoint requests one original checkpoint retry under ctx; false means already pending and no failure reports busy.
// RetryCheckpoint 在 ctx 下请求一次原检查点重试；假表示已在等待，不存在故障则报告忙碌。
// Return the admission receipt; no Lua or host callback is replayed and the original result stays retained.
// 返回入场回执；不重放 Lua 或宿主回调，原始结果继续保留。
func (h *EmbeddedOperation) RetryCheckpoint(ctx context.Context) (*EmbeddedPending[bool], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandOperationRetryCheckpoint{Type: EmbeddedInputRuntimeCommandOperationRetryCheckpointTypeOperationRetryCheckpoint, OperationId: h.identity}, projectEmbeddedResult[bool])
}
