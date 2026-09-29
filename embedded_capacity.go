package luaskills

import (
	"context"
	"fmt"
)

// EmbeddedCapacity binds a plugin-owned capacity to its original runtime without replacing native authority.
// EmbeddedCapacity 将插件自有容量绑定原运行时，不替换原生权威。
type EmbeddedCapacity struct {
	// runtime retains the exact namespace for member admission and lifecycle commands.
	// runtime 保留成员入场及生命周期命令的精确命名空间。
	runtime *EmbeddedRuntime
	// identity never changes after binding, even after plugin updates or native errors.
	// identity 绑定后绝不改变，包括插件更新或原生错误之后。
	identity string
}

// Capacity binds nonempty id without probing native state; returns a handle or explicit identity error.
// Capacity 绑定非空 id 而不探测原生状态；返回句柄或明确身份错误。
func (r *EmbeddedRuntime) Capacity(id string) (*EmbeddedCapacity, error) {
	if id == "" {
		return nil, fmt.Errorf("embedded capacity identity must be nonempty")
	}
	return &EmbeddedCapacity{runtime: r, identity: id}, nil
}

// RegisterCapacity freezes pluginID and config under ctx; returns a retained receipt projecting the acknowledged owner.
// RegisterCapacity 在 ctx 下冻结 pluginID 和 config；返回投影已确认所有者的保留回执。
func (r *EmbeddedRuntime) RegisterCapacity(ctx context.Context, pluginID string, config EmbeddedInputEmbeddedCapacityConfig) (*EmbeddedPending[*EmbeddedCapacity], error) {
	if pluginID == "" {
		return nil, fmt.Errorf("embedded plugin identity must be nonempty")
	}
	return submitEmbeddedRuntime(ctx, r, EmbeddedInputRuntimeCommandCapacityRegister{
		Type:     EmbeddedInputRuntimeCommandCapacityRegisterTypeCapacityRegister,
		PluginId: pluginID, Config: config,
	}, func(value any) (*EmbeddedCapacity, error) {
		// The exact native receipt is validated before publishing an immutable capacity handle.
		// 发布不可变容量句柄前校验精确原生回执。
		acknowledged, err := projectEmbeddedResult[EmbeddedOutputCapacityReceipt](value)
		if err != nil {
			return nil, err
		}
		return r.Capacity(acknowledged.CapacityId)
	})
}

// CapacityID returns the original native identity, not a local driver receipt identifier.
// CapacityID 返回原始原生身份，而非本地驱动回执标识。
func (h *EmbeddedCapacity) CapacityID() string { return h.identity }

// Status admits under ctx on the reserved control lane and returns actual physical, queued and cleanup ownership.
// Status 在 ctx 下于预留控制通道入场，返回实际物理、排队及清理归属。
func (h *EmbeddedCapacity) Status(ctx context.Context) (*EmbeddedPending[EmbeddedOutputEmbeddedCapacitySnapshot], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandCapacityStatus{
		Type: EmbeddedInputRuntimeCommandCapacityStatusTypeCapacityStatus, CapacityId: h.identity,
	}, projectEmbeddedResult[EmbeddedOutputEmbeddedCapacitySnapshot])
}

// Policy admits under ctx on the control lane and returns the atomic native revision, policy and convergence.
// Policy 在 ctx 下于控制通道入场，返回原子原生修订、策略及收敛状态。
func (h *EmbeddedCapacity) Policy(ctx context.Context) (*EmbeddedPending[EmbeddedOutputEmbeddedCapacityPolicySnapshot], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandCapacityPolicy{
		Type: EmbeddedInputRuntimeCommandCapacityPolicyTypeCapacityPolicy, CapacityId: h.identity,
	}, projectEmbeddedResult[EmbeddedOutputEmbeddedCapacityPolicySnapshot])
}

// Revise compares expectedRevision and replaces complete config under ctx; returns a retained committed-token receipt.
// Revise 在 ctx 下比较 expectedRevision 并替换完整 config；返回保留的已提交令牌回执。
// Native conflicts, execution pressure and closure remain explicit; the token is never refreshed or retried automatically.
// 原生冲突、执行压力及关闭保持显式；绝不自动刷新或重试令牌。
func (h *EmbeddedCapacity) Revise(ctx context.Context, expectedRevision string, config EmbeddedInputEmbeddedCapacityConfig) (*EmbeddedPending[string], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandCapacityRevise{
		Type: EmbeddedInputRuntimeCommandCapacityReviseTypeCapacityRevise, CapacityId: h.identity,
		ExpectedRevision: expectedRevision, Config: config,
	}, projectEmbeddedResult[string])
}

// RequestClose admits under ctx and requests member drainage; acknowledgement does not prove completion.
// RequestClose 在 ctx 下入场并请求成员排空；确认不证明完成。
func (h *EmbeddedCapacity) RequestClose(ctx context.Context) (*EmbeddedPending[*EmbeddedJSONNull], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandCapacityClose{
		Type: EmbeddedInputRuntimeCommandCapacityCloseTypeCapacityClose, CapacityId: h.identity,
	}, projectEmbeddedResult[*EmbeddedJSONNull])
}

// Forget admits under ctx and requests actual removal; every member must be explicitly forgotten first.
// Forget 在 ctx 下入场并请求实际移除；必须先显式遗忘全部成员。
func (h *EmbeddedCapacity) Forget(ctx context.Context) (*EmbeddedPending[*EmbeddedJSONNull], error) {
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandCapacityForget{
		Type: EmbeddedInputRuntimeCommandCapacityForgetTypeCapacityForget, CapacityId: h.identity,
	}, projectEmbeddedResult[*EmbeddedJSONNull])
}

// RegisterPool freezes definition, policy, permissions and executionRevision under ctx in this exact capacity.
// RegisterPool 在 ctx 下将 definition、policy、permissions 和 executionRevision 冻结到此精确容量。
// Return the acknowledged member; foreign owners or conflicting budgets fail without independent-placement fallback.
// 返回已确认成员；外来所有者或冲突预算失败，不回退独立归属。
func (h *EmbeddedCapacity) RegisterPool(ctx context.Context, definition EmbeddedInputModuleDefinition, policy EmbeddedInputPluginPoolConfig, permissions []string, executionRevision string) (*EmbeddedPending[*EmbeddedPool], error) {
	return h.RegisterPoolWithInitializationCapabilities(ctx, definition, policy, permissions, executionRevision, nil)
}

// RegisterPoolWithInitializationCapabilities freezes definition, policy, permissions and executionRevision under ctx in this capacity.
// RegisterPoolWithInitializationCapabilities 在 ctx 下将定义、策略、权限和执行修订冻结到此容量。
// initializationCapabilities only narrows source callbacks: nil inherits, an empty non-nil slice denies all.
// initializationCapabilities 仅收窄源码回调：nil 继承，非 nil 空切片全部拒绝。
// Return the acknowledged member without changing exact capacity ownership or ordinary business grants.
// 返回已确认成员，不改变精确容量归属或普通业务授权。
func (h *EmbeddedCapacity) RegisterPoolWithInitializationCapabilities(ctx context.Context, definition EmbeddedInputModuleDefinition, policy EmbeddedInputPluginPoolConfig, permissions []string, executionRevision string, initializationCapabilities []string) (*EmbeddedPending[*EmbeddedPool], error) {
	// A present non-null wire value selects mandatory exact capacity membership.
	// 存在且非空的线值选择强制精确容量成员关系。
	capacityID := &h.identity
	return submitEmbeddedRuntime(ctx, h.runtime, EmbeddedInputRuntimeCommandPoolRegister{
		Type: EmbeddedInputRuntimeCommandPoolRegisterTypePoolRegister, CapacityId: &capacityID,
		Definition: definition, Policy: policy, Permissions: permissions, ExecutionRevision: executionRevision,
		InitializationCapabilities: embeddedInitializationCapabilities(initializationCapabilities),
	}, func(value any) (*EmbeddedPool, error) {
		// Project only the original delivered receipt; observation never registers a second pool.
		// 仅投影原始已交付回执；观测绝不注册第二个池。
		acknowledged, err := projectEmbeddedResult[EmbeddedOutputPoolReceipt](value)
		if err != nil {
			return nil, err
		}
		return h.runtime.Pool(acknowledged.PoolId)
	})
}
