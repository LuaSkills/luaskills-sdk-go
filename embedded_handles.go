package luaskills

import (
	"context"
	"fmt"
)

// EmbeddedPlugin binds an immutable plugin identity to its exact runtime; native state remains queryable.
// EmbeddedPlugin 将不可变插件身份绑定到精确运行时；原生状态保持可查询。
type EmbeddedPlugin struct {
	// runtime and identity define the complete routing authority and never change after binding.
	// runtime 和 identity 定义完整路由权威，绑定后绝不改变。
	runtime  *EmbeddedRuntime
	identity string
}

// Plugin binds known nonempty id without native probing and returns a handle or an identity error.
// Plugin 绑定已知非空 id 而不探测原生状态，返回句柄或身份错误。
func (r *EmbeddedRuntime) Plugin(id string) (*EmbeddedPlugin, error) {
	if id == "" {
		return nil, fmt.Errorf("embedded plugin identity must be nonempty")
	}
	return &EmbeddedPlugin{runtime: r, identity: id}, nil
}

// PluginID returns the exact immutable native plugin identity, not a driver receipt identity.
// PluginID 返回精确不可变原生插件身份，不是驱动器回执身份。
func (h *EmbeddedPlugin) PluginID() string { return h.identity }

// Status admits under ctx and returns the actual generated status snapshot.
// Status 在 ctx 下入场，返回实际生成状态快照。
func (h *EmbeddedPlugin) Status(ctx context.Context) (*EmbeddedPending[EmbeddedOutputEmbeddedPluginSnapshot], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandPluginStatus{Type: EmbeddedInputRuntimeCommandPluginStatusTypePluginStatus, PluginId: h.identity}, projectEmbeddedResult[EmbeddedOutputEmbeddedPluginSnapshot])
}

// RequestClose admits under ctx and closes admission without proving actual drainage.
// RequestClose 在 ctx 下入场，关闭入场，不表示实际排空。
func (h *EmbeddedPlugin) RequestClose(ctx context.Context) (*EmbeddedPending[*EmbeddedJSONNull], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandPluginClose{Type: EmbeddedInputRuntimeCommandPluginCloseTypePluginClose, PluginId: h.identity}, projectEmbeddedResult[*EmbeddedJSONNull])
}

// Forget admits under ctx and removes native metadata only when the core proves it eligible; SDK receipt quota is separate.
// Forget 在 ctx 下入场，仅在核心证明可移除时遗忘原生元数据；SDK 回执配额独立管理。
func (h *EmbeddedPlugin) Forget(ctx context.Context) (*EmbeddedPending[*EmbeddedJSONNull], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandPluginForget{Type: EmbeddedInputRuntimeCommandPluginForgetTypePluginForget, PluginId: h.identity}, projectEmbeddedResult[*EmbeddedJSONNull])
}

// EmbeddedPool binds an immutable pool identity to its exact runtime; native state remains queryable.
// EmbeddedPool 将不可变池身份绑定到精确运行时；原生状态保持可查询。
type EmbeddedPool struct {
	// runtime and identity define the complete routing authority and never change after binding.
	// runtime 和 identity 定义完整路由权威，绑定后绝不改变。
	runtime  *EmbeddedRuntime
	identity string
}

// Pool binds known nonempty id without native probing and returns a handle or an identity error.
// Pool 绑定已知非空 id 而不探测原生状态，返回句柄或身份错误。
func (r *EmbeddedRuntime) Pool(id string) (*EmbeddedPool, error) {
	if id == "" {
		return nil, fmt.Errorf("embedded pool identity must be nonempty")
	}
	return &EmbeddedPool{runtime: r, identity: id}, nil
}

// PoolID returns the exact immutable native pool identity, not a driver receipt identity.
// PoolID 返回精确不可变原生池身份，不是驱动器回执身份。
func (h *EmbeddedPool) PoolID() string { return h.identity }

// Status admits under ctx and returns the actual generated status snapshot.
// Status 在 ctx 下入场，返回实际生成状态快照。
func (h *EmbeddedPool) Status(ctx context.Context) (*EmbeddedPending[EmbeddedOutputPoolUsage], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandPoolStatus{Type: EmbeddedInputRuntimeCommandPoolStatusTypePoolStatus, PoolId: h.identity}, projectEmbeddedResult[EmbeddedOutputPoolUsage])
}

// RequestClose admits under ctx and closes admission without proving actual drainage.
// RequestClose 在 ctx 下入场，关闭入场，不表示实际排空。
func (h *EmbeddedPool) RequestClose(ctx context.Context) (*EmbeddedPending[*EmbeddedJSONNull], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandPoolClose{Type: EmbeddedInputRuntimeCommandPoolCloseTypePoolClose, PoolId: h.identity}, projectEmbeddedResult[*EmbeddedJSONNull])
}

// Forget admits under ctx and removes native metadata only when the core proves it eligible; SDK receipt quota is separate.
// Forget 在 ctx 下入场，仅在核心证明可移除时遗忘原生元数据；SDK 回执配额独立管理。
func (h *EmbeddedPool) Forget(ctx context.Context) (*EmbeddedPending[*EmbeddedJSONNull], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandPoolForget{Type: EmbeddedInputRuntimeCommandPoolForgetTypePoolForget, PoolId: h.identity}, projectEmbeddedResult[*EmbeddedJSONNull])
}

// EmbeddedSession binds an immutable session identity to its exact runtime; native state remains queryable.
// EmbeddedSession 将不可变会话身份绑定到精确运行时；原生状态保持可查询。
type EmbeddedSession struct {
	// runtime and identity define the complete routing authority and never change after binding.
	// runtime 和 identity 定义完整路由权威，绑定后绝不改变。
	runtime  *EmbeddedRuntime
	identity string
}

// Session binds known nonempty id without native probing and returns a handle or an identity error.
// Session 绑定已知非空 id 而不探测原生状态，返回句柄或身份错误。
func (r *EmbeddedRuntime) Session(id string) (*EmbeddedSession, error) {
	if id == "" {
		return nil, fmt.Errorf("embedded session identity must be nonempty")
	}
	return &EmbeddedSession{runtime: r, identity: id}, nil
}

// SessionID returns the exact immutable native session identity, not a driver receipt identity.
// SessionID 返回精确不可变原生会话身份，不是驱动器回执身份。
func (h *EmbeddedSession) SessionID() string { return h.identity }

// Status admits under ctx and returns the actual generated status snapshot.
// Status 在 ctx 下入场，返回实际生成状态快照。
func (h *EmbeddedSession) Status(ctx context.Context) (*EmbeddedPending[EmbeddedOutputEmbeddedSessionSnapshot], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandSessionStatus{Type: EmbeddedInputRuntimeCommandSessionStatusTypeSessionStatus, SessionId: h.identity}, projectEmbeddedResult[EmbeddedOutputEmbeddedSessionSnapshot])
}

// RequestClose admits under ctx and closes admission without proving actual drainage.
// RequestClose 在 ctx 下入场，关闭入场，不表示实际排空。
func (h *EmbeddedSession) RequestClose(ctx context.Context) (*EmbeddedPending[*EmbeddedJSONNull], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandSessionClose{Type: EmbeddedInputRuntimeCommandSessionCloseTypeSessionClose, SessionId: h.identity}, projectEmbeddedResult[*EmbeddedJSONNull])
}

// Forget admits under ctx and removes native metadata only when the core proves it eligible; SDK receipt quota is separate.
// Forget 在 ctx 下入场，仅在核心证明可移除时遗忘原生元数据；SDK 回执配额独立管理。
func (h *EmbeddedSession) Forget(ctx context.Context) (*EmbeddedPending[*EmbeddedJSONNull], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandSessionForget{Type: EmbeddedInputRuntimeCommandSessionForgetTypeSessionForget, SessionId: h.identity}, projectEmbeddedResult[*EmbeddedJSONNull])
}

// EmbeddedOperation binds an immutable operation identity to its exact runtime; native state remains queryable.
// EmbeddedOperation 将不可变操作身份绑定到精确运行时；原生状态保持可查询。
type EmbeddedOperation struct {
	// runtime and identity define the complete routing authority and never change after binding.
	// runtime 和 identity 定义完整路由权威，绑定后绝不改变。
	runtime  *EmbeddedRuntime
	identity string
}

// Operation binds known nonempty id without native probing and returns a handle or an identity error.
// Operation 绑定已知非空 id 而不探测原生状态，返回句柄或身份错误。
func (r *EmbeddedRuntime) Operation(id string) (*EmbeddedOperation, error) {
	if id == "" {
		return nil, fmt.Errorf("embedded operation identity must be nonempty")
	}
	return &EmbeddedOperation{runtime: r, identity: id}, nil
}

// OperationID returns the exact immutable native operation identity, not a driver receipt identity.
// OperationID 返回精确不可变原生操作身份，不是驱动器回执身份。
func (h *EmbeddedOperation) OperationID() string { return h.identity }

// Status admits under ctx and returns the actual generated status snapshot.
// Status 在 ctx 下入场，返回实际生成状态快照。
func (h *EmbeddedOperation) Status(ctx context.Context) (*EmbeddedPending[EmbeddedOutputOperationSnapshot], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandOperationStatus{Type: EmbeddedInputRuntimeCommandOperationStatusTypeOperationStatus, OperationId: h.identity}, projectEmbeddedResult[EmbeddedOutputOperationSnapshot])
}

// Cancel admits under ctx and requests cooperative cancellation without proving completion or rollback.
// Cancel 在 ctx 下入场，请求协作取消，不表示完成或回滚。
func (h *EmbeddedOperation) Cancel(ctx context.Context) (*EmbeddedPending[bool], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandOperationCancel{Type: EmbeddedInputRuntimeCommandOperationCancelTypeOperationCancel, OperationId: h.identity}, projectEmbeddedResult[bool])
}

// Forget admits under ctx and removes native metadata only when the core proves it eligible; SDK receipt quota is separate.
// Forget 在 ctx 下入场，仅在核心证明可移除时遗忘原生元数据；SDK 回执配额独立管理。
func (h *EmbeddedOperation) Forget(ctx context.Context) (*EmbeddedPending[*EmbeddedJSONNull], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandOperationForget{Type: EmbeddedInputRuntimeCommandOperationForgetTypeOperationForget, OperationId: h.identity}, projectEmbeddedResult[*EmbeddedJSONNull])
}

// RevokePermission admits permission revocation under ctx; returns whether the live grant was removed.
// RevokePermission 在 ctx 下接纳 permission 撤销；返回实时授权是否移除。
func (h *EmbeddedPool) RevokePermission(ctx context.Context, permission string) (*EmbeddedPending[bool], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandPoolRevokePermission{Type: EmbeddedInputRuntimeCommandPoolRevokePermissionTypePoolRevokePermission, PoolId: h.identity, Permission: permission}, projectEmbeddedResult[bool])
}

// operationResult validates a submission acknowledgement and returns its exact independently queryable operation.
// operationResult 校验提交确认，返回其精确且可独立查询的操作。
func (r *EmbeddedRuntime) operationResult(value any) (*EmbeddedOperation, error) {
	// acknowledged is the only operation identity source; no local sequence is inferred.
	// acknowledged 是操作身份的唯一来源；不推断本地序号。
	acknowledged, err := projectEmbeddedResult[EmbeddedOutputOperationReceipt](value)
	if err != nil {
		return nil, err
	}
	return r.Operation(acknowledged.OperationId)
}

// Submit freezes export, arguments and invocation under ctx; timeoutMS is the native end-to-end execution budget.
// Submit 在 ctx 下冻结 export、arguments 和 invocation；timeoutMS 是原生端到端执行预算。
// It returns the acknowledged operation, whose outcome must be queried separately from command delivery.
// 返回已确认操作，其结果必须与命令交付分开查询。
func (h *EmbeddedPool) Submit(ctx context.Context, export string, arguments any, invocation EmbeddedInputLuaInvocationContext, timeoutMS uint64) (*EmbeddedPending[*EmbeddedOperation], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandCallSubmit{Type: EmbeddedInputRuntimeCommandCallSubmitTypeCallSubmit, TimeoutMs: timeoutMS, Call: EmbeddedInputEmbeddedCall{PoolId: h.identity, Export: export, Arguments: arguments, Context: invocation}}, h.runtime.operationResult)
}

// Submit freezes export, arguments and invocation under ctx for this fixed session; timeoutMS remains a native budget.
// Submit 在 ctx 下为此固定会话冻结 export、arguments 和 invocation；timeoutMS 保持为原生预算。
// It returns the acknowledged operation; an observer timeout never implies cancellation or slot release.
// 返回已确认操作；观察超时绝不表示取消或槽释放。
func (h *EmbeddedSession) Submit(ctx context.Context, export string, arguments any, invocation EmbeddedInputLuaInvocationContext, timeoutMS uint64) (*EmbeddedPending[*EmbeddedOperation], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandSessionSubmit{Type: EmbeddedInputRuntimeCommandSessionSubmitTypeSessionSubmit, SessionId: h.identity, Export: export, Arguments: arguments, Context: invocation, TimeoutMs: timeoutMS}, h.runtime.operationResult)
}

// EmbeddedSessionOpen retains both identities from the same session reservation acknowledgement.
// EmbeddedSessionOpen 保留同一个会话预留确认中的两个身份。
type EmbeddedSessionOpen struct {
	// Session owns the fixed VM after successful initialization and needs explicit closure.
	// Session 在初始化成功后拥有固定 VM，需要显式关闭。
	Session *EmbeddedSession
	// Initialization is independently queryable; reservation alone does not prove successful construction.
	// Initialization 可独立查询；仅预留不表示构造成功。
	Initialization *EmbeddedOperation
}

// OpenSession admits under ctx with native initialization timeoutMS; returns session and initialization handles together.
// OpenSession 在 ctx 下以原生初始化 timeoutMS 入场；同时返回会话及初始化句柄。
func (h *EmbeddedPool) OpenSession(ctx context.Context, timeoutMS uint64) (*EmbeddedPending[EmbeddedSessionOpen], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandSessionOpen{Type: EmbeddedInputRuntimeCommandSessionOpenTypeSessionOpen, PoolId: h.identity, TimeoutMs: timeoutMS}, func(value any) (EmbeddedSessionOpen, error) {
		// acknowledged keeps the two identities tied to their original atomic reservation.
		// acknowledged 保持两个身份与原始原子预留绑定。
		var empty EmbeddedSessionOpen
		acknowledged, err := projectEmbeddedResult[EmbeddedOutputSessionReceipt](value)
		if err != nil {
			return empty, err
		}
		// session and initialization are bound only after complete acknowledgement validation.
		// session 和 initialization 仅在完整确认校验后绑定。
		session, err := h.runtime.Session(acknowledged.SessionId)
		if err != nil {
			return empty, err
		}
		initialization, err := h.runtime.Operation(acknowledged.OperationId)
		if err != nil {
			return empty, err
		}
		return EmbeddedSessionOpen{Session: session, Initialization: initialization}, nil
	})
}
