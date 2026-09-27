package luaskills

import (
	"context"
	"fmt"
)

// EmbeddedPending projects one retained command into T without issuing another native request.
// EmbeddedPending 将一个保留命令投影为 T，不发出另一个原生请求。
// Obtain values from client methods; a zero value is not a submitted command.
// 从客户端方法取得实例；零值不表示已提交命令。
type EmbeddedPending[T any] struct {
	// receipt owns the frozen request, copied response and explicit SDK quota.
	// receipt 拥有冻结请求、复制响应及显式 SDK 配额。
	receipt *EmbeddedCommand
	// project validates the generated result shape before binding any native identity.
	// project 在绑定任何原生身份前校验生成结果形状。
	project func(any) (T, error)
}

// Receipt returns the exact driver-owned command for diagnostics and interrupted-observer recovery.
// Receipt 返回驱动器拥有的精确命令，供诊断和中断观察恢复。
func (p *EmbeddedPending[T]) Receipt() *EmbeddedCommand { return p.receipt }

// Result waits under ctx and returns a fresh typed value or the original delivery error.
// Result 在 ctx 下等待，返回新的类型化值或原始交付错误。
// Observer cancellation does not cancel native work, forget its identity or return SDK quota.
// 观察者取消不取消原生工作、不遗忘其身份，也不归还 SDK 配额。
func (p *EmbeddedPending[T]) Result(ctx context.Context) (T, error) {
	// value is a fresh response decode; empty is returned only with an explicit error.
	// value 是新解码的响应；empty 只与显式错误一起返回。
	var empty T
	value, err := p.receipt.Result(ctx)
	if err != nil {
		return empty, err
	}
	return p.project(value)
}

// DeliveredResult returns typed original response evidence even after native buffer release failed.
// DeliveredResult 即使原生缓冲释放失败也返回类型化原始响应证据。
// It never waits, replays work or releases the retained allocation; use driver.ReleaseResults separately.
// 绝不等待、重放工作或释放保留分配；须单独使用 driver.ReleaseResults。
func (p *EmbeddedPending[T]) DeliveredResult() (T, error) {
	// value is copied evidence, not a second attempt at the mutation.
	// value 是复制证据，不是第二次变更尝试。
	var empty T
	value, err := p.receipt.DeliveredResult()
	if err != nil {
		return empty, err
	}
	return p.project(value)
}

// Forget returns only this completed command's SDK quota; native handles require their own cleanup methods.
// Forget 仅归还此已完成命令的 SDK 配额；原生句柄需要各自的清理方法。
func (p *EmbeddedPending[T]) Forget() error { return p.receipt.Forget() }

// projectEmbeddedResult validates value against the generated T shape and returns owned typed data or an error.
// projectEmbeddedResult 按生成的 T 形状校验 value，返回拥有型数据或错误。
func projectEmbeddedResult[T any](value any) (T, error) {
	// projected uses the same schema authority as the public generated envelope decoders.
	// projected 使用与公开生成信封解码器相同的 Schema 权威。
	var empty T
	projected, err := projectEmbeddedWire(value, embeddedWireType[T](), "response.result")
	if err != nil {
		return empty, err
	}
	return projected.Interface().(T), nil
}

// submitEmbedded retains one command before returning its typed observer; ctx governs admission only.
// submitEmbedded 在返回类型化观察者前保留一个命令；ctx 仅控制入场。
// project must be pure and validate the declared result before creating any handle.
// project 必须是纯投影，并在创建任何句柄前校验声明结果。
func submitEmbedded[T any](ctx context.Context, client *EmbeddedClient, command EmbeddedInputCommand, project func(any) (T, error)) (*EmbeddedPending[T], error) {
	// receipt remains discoverable in the borrowed driver when a later observation ends.
	// 后续观察结束时 receipt 仍可在借用驱动器中发现。
	receipt, err := client.driver.Submit(ctx, command)
	if err != nil {
		return nil, err
	}
	return &EmbeddedPending[T]{receipt: receipt, project: project}, nil
}

// submitEmbeddedRuntime binds a typed operation to the exact transport-local runtime identity.
// submitEmbeddedRuntime 将类型化操作绑定到精确传输局部运行时身份。
// It returns the retained observer or a pre-admission error and never probes alternative identities.
// 返回保留观察者或入场前错误，绝不探测候选身份。
func submitEmbeddedRuntime[T any](ctx context.Context, runtime *EmbeddedRuntime, operation EmbeddedInputRuntimeCommand, project func(any) (T, error)) (*EmbeddedPending[T], error) {
	return submitEmbedded(ctx, runtime.client, EmbeddedInputCommandRuntime{Type: EmbeddedInputCommandRuntimeTypeRuntime, RuntimeId: runtime.identity, Operation: operation}, project)
}

// EmbeddedClient borrows an existing command driver; it does not own transport or runtime shutdown.
// EmbeddedClient 借用已有命令驱动器；不拥有传输或运行时关闭。
// Construct with NewEmbeddedClient; handles keep this client and their exact native identities immutable.
// 使用 NewEmbeddedClient 构造；句柄保持此客户端及其精确原生身份不可变。
type EmbeddedClient struct {
	// driver owns all submitted commands independently of client handle lifetimes.
	// driver 独立于客户端句柄寿命拥有全部已提交命令。
	driver *EmbeddedCommandDriver
}

// NewEmbeddedClient returns a typed facade over driver, or an error for a nil driver; it makes no native call.
// NewEmbeddedClient 返回 driver 的类型化外观，driver 为空时返回错误；不发起原生调用。
func NewEmbeddedClient(driver *EmbeddedCommandDriver) (*EmbeddedClient, error) {
	if driver == nil {
		return nil, fmt.Errorf("embedded client requires a command driver")
	}
	return &EmbeddedClient{driver: driver}, nil
}

// Driver returns the borrowed driver for command discovery and explicit buffer recovery.
// Driver 返回借用驱动器，供命令发现和显式缓冲恢复。
func (c *EmbeddedClient) Driver() *EmbeddedCommandDriver { return c.driver }

// Describe admits under ctx and returns a pending generated transport description.
// Describe 在 ctx 下入场，返回待观察的生成传输描述。
func (c *EmbeddedClient) Describe(ctx context.Context) (*EmbeddedPending[EmbeddedOutputTransportDescription], error) {
	return submitEmbedded(ctx, c, EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe}, projectEmbeddedResult[EmbeddedOutputTransportDescription])
}

// Reserve admits under ctx and returns the actual metadata-only runtime handle after native acknowledgement.
// Reserve 在 ctx 下入场，在原生确认后返回实际仅含元数据的运行时句柄。
func (c *EmbeddedClient) Reserve(ctx context.Context) (*EmbeddedPending[*EmbeddedRuntime], error) {
	return submitEmbedded(ctx, c, EmbeddedInputCommandRuntimeReserve{Type: EmbeddedInputCommandRuntimeReserveTypeRuntimeReserve}, func(value any) (*EmbeddedRuntime, error) {
		// acknowledged is the sole source of the new runtime identity.
		// acknowledged 是新运行时身份的唯一来源。
		acknowledged, err := projectEmbeddedResult[EmbeddedOutputRuntimeReceipt](value)
		if err != nil {
			return nil, err
		}
		return c.Runtime(acknowledged.RuntimeId)
	})
}

// Runtime binds a known nonempty runtimeID without probing existence; subsequent commands expose native errors.
// Runtime 绑定已知非空 runtimeID，不探测存在性；后续命令暴露原生错误。
func (c *EmbeddedClient) Runtime(runtimeID string) (*EmbeddedRuntime, error) {
	if runtimeID == "" {
		return nil, fmt.Errorf("embedded runtime identity must be nonempty")
	}
	return &EmbeddedRuntime{client: c, identity: runtimeID}, nil
}

// EmbeddedRuntime identifies one exact FFI slot, distinct from the core runtime namespace in status snapshots.
// EmbeddedRuntime 标识一个精确 FFI 槽，区别于状态快照中的核心运行时命名空间。
type EmbeddedRuntime struct {
	// client and identity remain immutable; native state is always queried rather than inferred locally.
	// client 和 identity 保持不可变；原生状态始终通过查询取得，不在本地推断。
	client   *EmbeddedClient
	identity string
}

// RuntimeID returns the immutable transport-local identity used by every runtime command.
// RuntimeID 返回每个运行时命令使用的不可变传输局部身份。
func (r *EmbeddedRuntime) RuntimeID() string { return r.identity }

// Initialize admits exact options and budgets under ctx; its acknowledgement does not replace a status query.
// Initialize 在 ctx 下接纳精确 options 和 budgets；其确认不替代状态查询。
// The core allows one construction attempt; retrying observation never retries construction.
// 核心允许一次构造尝试；重试观察绝不重试构造。
func (r *EmbeddedRuntime) Initialize(ctx context.Context, options EmbeddedInputLuaEngineOptions, budgets EmbeddedInputEmbeddedRuntimeConfig) (*EmbeddedPending[EmbeddedOutputRuntimeReceipt], error) {
	return submitEmbedded(ctx, r.client, EmbeddedInputCommandRuntimeInitialize{Type: EmbeddedInputCommandRuntimeInitializeTypeRuntimeInitialize, RuntimeId: r.identity, EngineOptions: options, RuntimeConfig: budgets}, projectEmbeddedResult[EmbeddedOutputRuntimeReceipt])
}

// Status admits under ctx and returns actual construction, closure and resource evidence.
// Status 在 ctx 下入场，返回实际构造、关闭及资源证据。
func (r *EmbeddedRuntime) Status(ctx context.Context) (*EmbeddedPending[EmbeddedOutputRuntimeSnapshot], error) {
	return submitEmbedded(ctx, r.client, EmbeddedInputCommandRuntimeStatus{Type: EmbeddedInputCommandRuntimeStatusTypeRuntimeStatus, RuntimeId: r.identity}, projectEmbeddedResult[EmbeddedOutputRuntimeSnapshot])
}

// RequestClose admits closure under ctx and returns acknowledgement; it does not wait for actual native drainage.
// RequestClose 在 ctx 下接纳关闭并返回确认；不等待实际原生排空。
func (r *EmbeddedRuntime) RequestClose(ctx context.Context) (*EmbeddedPending[EmbeddedOutputRuntimeReceipt], error) {
	return submitEmbedded(ctx, r.client, EmbeddedInputCommandRuntimeClose{Type: EmbeddedInputCommandRuntimeCloseTypeRuntimeClose, RuntimeId: r.identity}, projectEmbeddedResult[EmbeddedOutputRuntimeReceipt])
}

// Free admits explicit slot removal under ctx; the core rejects a runtime that is not closed and fully drained.
// Free 在 ctx 下接纳显式槽移除；核心拒绝尚未关闭且完全排空的运行时。
func (r *EmbeddedRuntime) Free(ctx context.Context) (*EmbeddedPending[EmbeddedOutputRuntimeReceipt], error) {
	return submitEmbedded(ctx, r.client, EmbeddedInputCommandRuntimeFree{Type: EmbeddedInputCommandRuntimeFreeTypeRuntimeFree, RuntimeId: r.identity}, projectEmbeddedResult[EmbeddedOutputRuntimeReceipt])
}

// RegisterPlugin freezes pluginID and budgets under ctx; returns a handle only after validating the null acknowledgement.
// RegisterPlugin 在 ctx 下冻结 pluginID 和 budgets；仅在校验空值确认后返回句柄。
func (r *EmbeddedRuntime) RegisterPlugin(ctx context.Context, pluginID string, budgets EmbeddedInputEmbeddedPluginConfig) (*EmbeddedPending[*EmbeddedPlugin], error) {
	if pluginID == "" {
		return nil, fmt.Errorf("embedded plugin identity must be nonempty")
	}
	return submitEmbeddedRuntime(ctx, r, EmbeddedInputRuntimeCommandPluginRegister{Type: EmbeddedInputRuntimeCommandPluginRegisterTypePluginRegister, PluginId: pluginID, Config: budgets}, func(value any) (*EmbeddedPlugin, error) {
		if _, err := projectEmbeddedResult[*EmbeddedJSONNull](value); err != nil {
			return nil, err
		}
		return r.Plugin(pluginID)
	})
}

// RegisterPool freezes definition, policy, permissions and executionRevision under ctx and returns the acknowledged pool.
// RegisterPool 在 ctx 下冻结 definition、policy、permissions 和 executionRevision，返回已确认池。
func (r *EmbeddedRuntime) RegisterPool(ctx context.Context, definition EmbeddedInputModuleDefinition, policy EmbeddedInputPluginPoolConfig, permissions []string, executionRevision string) (*EmbeddedPending[*EmbeddedPool], error) {
	return submitEmbeddedRuntime(ctx, r, EmbeddedInputRuntimeCommandPoolRegister{Type: EmbeddedInputRuntimeCommandPoolRegisterTypePoolRegister, Definition: definition, Policy: policy, Permissions: permissions, ExecutionRevision: executionRevision}, func(value any) (*EmbeddedPool, error) {
		// acknowledged is validated before its immutable pool identity becomes usable.
		// acknowledged 在不可变池身份可用前经过校验。
		acknowledged, err := projectEmbeddedResult[EmbeddedOutputPoolReceipt](value)
		if err != nil {
			return nil, err
		}
		return r.Pool(acknowledged.PoolId)
	})
}
