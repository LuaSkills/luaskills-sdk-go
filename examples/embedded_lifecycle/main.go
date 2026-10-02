// Command embedded_lifecycle runs typed embedded Lua with a queued Go callback and ordered resource shutdown.
// embedded_lifecycle 命令执行类型化嵌入式 Lua 及队列 Go 回调，并按序关闭资源。
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	luaskills "github.com/LuaSkills/luaskills-sdk-go"
)

// owners retains each actual resource as soon as construction succeeds; no GC finalizer owns shutdown.
// owners 在构造成功时立即保留每个实际资源；关闭不依赖 GC 终结器。
type owners struct {
	// transport and driver are shared infrastructure explicitly closed after the scope.
	// transport 和 driver 是共享基础设施，在作用域之后显式关闭。
	transport *luaskills.EmbeddedTransport
	driver    *luaskills.EmbeddedCommandDriver
	// reserved retains the original typed admission before any observer can cancel and lose its projected runtime.
	// reserved 在任何观察者取消并丢失投影运行时前保留原始类型化入场。
	reserved *luaskills.EmbeddedPending[*luaskills.EmbeddedRuntime]
	// runtime and pump become the exact scope's borrowed owners after successful adoption.
	// runtime 和 pump 在接管成功后成为精确作用域借用的所有者。
	runtime *luaskills.EmbeddedRuntime
	pump    *luaskills.EmbeddedCallbackPump
	// scope removes the native slot only after Lua and callbacks actually drain.
	// scope 仅在 Lua 及回调实际排空后移除原生槽。
	scope *luaskills.EmbeddedRuntimeScope
}

// adopt waits only for the documented quiescent boundary of the exact existing pump and driver.
// adopt 仅等待精确现有泵及驱动器的已定义静止边界。
// ctx bounds observation; return an adopted scope or the original non-busy/cancellation error.
// ctx 限制观察；返回已接管作用域或原始非忙碌／取消错误。
func (o *owners) adopt(ctx context.Context) error {
	for {
		// scope is published only after its exact ownership claim succeeds.
		// scope 仅在精确所有权声明成功后发布。
		scope, err := luaskills.NewEmbeddedRuntimeScope(o.runtime, o.pump, luaskills.EmbeddedDefaultPollInterval)
		if err == nil {
			o.scope = scope
			return nil
		}
		// busy is the existing API's explicit quiescence rejection, not a candidate path fallback.
		// busy 是现有 API 的显式静止边界拒绝，不是候选路径兜底。
		var busy *luaskills.EmbeddedRuntimeError
		if !errors.As(err, &busy) || busy.Code != "busy" {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(luaskills.EmbeddedDefaultPollInterval):
		}
	}
}

// close observes accepted commands, closes the scope, joins the driver, forgets receipts, and releases the transport.
// close 观察已接纳命令、关闭作用域、汇合驱动器、遗忘回执，并释放传输。
// ctx is a fresh cleanup observer; errors retain discoverable owners and never authorize business replay.
// ctx 是新清理观察者；错误保留可发现所有者，绝不授权业务重放。
func (o *owners) close(ctx context.Context) error {
	if o.runtime == nil && o.reserved != nil {
		// runtime comes only from the same accepted reservation, never a replacement Reserve mutation.
		// runtime 仅来自同一个已接纳预留，绝不是替代 Reserve 变更。
		runtime, err := o.reserved.Result(ctx)
		if err != nil {
			return fmt.Errorf("retained reservation requires host recovery: %w", err)
		}
		o.runtime = runtime
	}
	if o.driver != nil {
		for _, command := range o.driver.Commands() {
			if _, err := command.Result(ctx); err != nil {
				return fmt.Errorf("retained command %d requires host recovery: %w", command.CommandID(), err)
			}
		}
	}
	if o.runtime != nil {
		if o.scope == nil {
			if err := o.adopt(ctx); err != nil {
				return err
			}
		}
		if err := o.scope.Close(ctx); err != nil {
			return fmt.Errorf("retained scope requires host recovery: %w", err)
		}
	}
	if o.driver != nil {
		if err := o.driver.Close(ctx); err != nil {
			return err
		}
		for _, command := range o.driver.Commands() {
			if err := command.Forget(); err != nil {
				return err
			}
		}
	}
	if o.transport != nil {
		if err := o.transport.ReleaseResults(); err != nil {
			return err
		}
		if err := o.transport.Close(); err != nil {
			return err
		}
		return o.transport.Free()
	}
	return nil
}

// take observes one accepted typed command and returns its SDK receipt quota after successful projection.
// take 观察一个已接纳类型化命令，并在投影成功后归还其 SDK 回执配额。
// ctx governs observation, pending identifies exact delivery, and admission is the original submission error.
// ctx 控制观察，pending 标识精确交付，admission 是原始提交错误。
// Return the projected value or the original error; failed receipts remain with the driver.
// 返回投影值或原始错误；失败回执保留在驱动器中。
func take[T any](ctx context.Context, pending *luaskills.EmbeddedPending[T], admission error) (T, error) {
	// empty is returned only on an error, never as fabricated native success.
	// empty 仅在错误时返回，绝不伪造原生成功。
	var empty T
	if admission != nil {
		return empty, admission
	}
	// value is independently decoded from the original receipt.
	// value 从原始回执独立解码。
	value, err := pending.Result(ctx)
	if err != nil {
		return empty, err
	}
	return value, pending.Forget()
}

// echo returns the original arguments through a read-only queued callback, preserving the supplied context boundary.
// echo 通过只读队列回调返回原始参数，保留所提供上下文的边界。
// arguments is the Lua value; callback carries trusted cancellation and caller metadata; return a value or error.
// arguments 是 Lua 值；callback 携带可信取消及调用方元数据；返回值或错误。
func echo(arguments any, callback *luaskills.EmbeddedHostCallbackContext) (any, error) {
	if err := callback.Err(); err != nil {
		return nil, err
	}
	return arguments, nil
}

// run constructs explicit budgets, invokes Lua through the typed client, and proves complete successful resource closure.
// run 构造显式预算、通过类型化客户端调用 Lua，并证明成功路径的完整资源关闭。
// Return business or cleanup errors; failed cleanup preserves its temporary runtime directory for host recovery.
// 返回业务或清理错误；清理失败时保留临时运行时目录供宿主恢复。
func run() (result error) {
	// root is an isolated local runtime; initialization cannot download dependencies or alter user assets.
	// root 是隔离本地运行时；初始化不能下载依赖或更改用户资产。
	root, err := os.MkdirTemp("", "luaskills-go-embedded-")
	if err != nil {
		return err
	}
	// owned holds every successfully constructed owner for cleanup, including partial initialization.
	// owned 保存每个成功构造的所有者供清理，包含部分初始化。
	owned := &owners{}
	defer func() {
		// cleanup never reuses a cancelled business observer.
		// cleanup 绝不复用已取消业务观察者。
		cleanup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := owned.close(cleanup); err != nil {
			result = errors.Join(result, fmt.Errorf("cleanup failed; retained directory %s: %w", root, err))
			return
		}
		result = errors.Join(result, os.RemoveAll(root))
	}()
	// ctx limits this demonstration's observers; core execution gets its own explicit timeout below.
	// ctx 限制此演示的观察者；核心执行在下方取得独立显式超时。
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	owned.transport, err = luaskills.NewEmbeddedTransport(luaskills.EmbeddedTransportConfig{MaxRuntimes: 1, MaxResultBuffers: 4, MaxResultBytes: 131072, MaxResponseBytes: 32768, MaxRequestBytes: 65536})
	if err != nil {
		return err
	}
	owned.driver, err = luaskills.NewEmbeddedCommandDriver(owned.transport, luaskills.EmbeddedDriverConfig{WorkWorkers: 1, MaxWorkCommands: 4, MaxControlCommands: 4})
	if err != nil {
		return err
	}
	// client borrows the exact driver; Reserve returns the real FFI slot identity.
	// client 借用精确驱动器；Reserve 返回真实 FFI 槽身份。
	client, err := luaskills.NewEmbeddedClient(owned.driver)
	if err != nil {
		return err
	}
	// Preserve typed admission before observing it, so cancellation cannot orphan an accepted native slot.
	// 观察前保存类型化入场，使取消无法遗弃已接纳原生槽。
	owned.reserved, err = client.Reserve(ctx)
	owned.runtime, err = take(ctx, owned.reserved, err)
	if err != nil {
		return err
	}
	// system and packageRoot explicitly separate interpreter assets from the authorized external package.
	// system 和 packageRoot 显式分开解释器资产与授权外部包。
	system, packageRoot := filepath.Join(root, "system_lua_lib"), filepath.Join(root, "plugin-generations", "go-example")
	if err := os.MkdirAll(packageRoot, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(packageRoot, "dependencies.yaml"), []byte("{}\n"), 0644); err != nil {
		return err
	}
	// Nullable generated path fields retain explicit presence instead of relying on SDK defaults.
	// 生成可空路径字段保留显式存在性，不依赖 SDK 默认值。
	rootPointer, systemPointer := &root, &system
	// options and budgets are complete host-owned policy, copied from the same current typed contract as native tests.
	// options 和 budgets 是完整宿主策略，与原生测试使用同一当前类型化契约。
	options := luaskills.EmbeddedInputLuaEngineOptions{HostOptions: luaskills.EmbeddedInputLuaRuntimeHostOptions{RuntimeRoot: &rootPointer, SystemLuaLibDir: &systemPointer, ReservedEntryNames: luaskills.EmbeddedInputLuaRuntimeHostOptionsReservedEntryNames{}, AllowNetworkDownload: false}, PoolConfig: luaskills.EmbeddedInputLuaVmPoolConfig{MinSize: 0, MaxSize: 2, IdleTtlSecs: 60}}
	budgets := luaskills.EmbeddedInputEmbeddedRuntimeConfig{MaxRegisteredPlugins: 1, MaxRegisteredPools: 1, MaxSessions: 1, MaxRegisteredCapabilities: 1, MaxResidentVms: 2, MaxRunningCalls: 2, MaxQueuedCalls: 4, MaxQueuedBytes: 4096, MaxOperations: 16, MaxEffectRecordsPerOperation: 8, MaxEffectBytesPerOperation: 8192, MaxHostRequests: 4, MaxHostRequestBytes: 8192, MaxValueBytes: 1024}
	// initializing acknowledges initialization; Status below checks the actual outcome separately.
	// initializing 确认初始化；下方 Status 单独检查实际结果。
	initializing, err := owned.runtime.Initialize(ctx, options, budgets)
	if _, err = take(ctx, initializing, err); err != nil {
		return err
	}
	for {
		// status is an actual native snapshot, never inferred from an initialization receipt.
		// status 是实际原生快照，绝不根据初始化回执推断。
		statusPending, submitErr := owned.runtime.Status(ctx)
		status, statusErr := take(ctx, statusPending, submitErr)
		if statusErr != nil {
			return statusErr
		}
		if status.Initialization == luaskills.EmbeddedOutputInitializationPhaseReady {
			break
		}
		if status.Initialization != luaskills.EmbeddedOutputInitializationPhaseInitializing {
			return fmt.Errorf("runtime initialization ended in %s", status.Initialization)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(luaskills.EmbeddedDefaultPollInterval):
		}
	}
	// plugin records native quotas independently from the SDK receipt quotas.
	// plugin 独立于 SDK 回执配额记录原生配额。
	plugin, err := owned.runtime.RegisterPlugin(ctx, "go-example", luaskills.EmbeddedInputEmbeddedPluginConfig{MaxRegisteredPools: budgets.MaxRegisteredPools, MaxSessions: budgets.MaxSessions, MaxResidentVms: budgets.MaxResidentVms, MaxRunningCalls: budgets.MaxRunningCalls, MaxQueuedCalls: budgets.MaxQueuedCalls, MaxQueuedBytes: budgets.MaxQueuedBytes, MaxOperations: budgets.MaxOperations})
	if _, err = take(ctx, plugin, err); err != nil {
		return err
	}
	owned.pump, err = luaskills.NewEmbeddedCallbackPump(owned.transport, owned.runtime.RuntimeID(), luaskills.EmbeddedCallbackPumpConfig{MaxConcurrentHandlers: 1, MaxPendingCommands: 1, PollIntervalMS: 10})
	if err != nil {
		return err
	}
	if err = owned.pump.Ready(ctx); err != nil {
		return err
	}
	if err = owned.adopt(ctx); err != nil {
		return err
	}
	// capability publication precedes pool registration because the core freezes authority at that boundary.
	// 能力发布先于池注册，因为核心在该边界冻结权威。
	capability := luaskills.EmbeddedHostCapability{Descriptor: luaskills.EmbeddedInputCapabilityDescriptor{Name: "go.echo", Version: "1.0.0", Description: "Return the supplied value through a read-only Go callback", InputSchema: true, OutputSchema: true, Execution: luaskills.EmbeddedInputCapabilityExecutionQueued, Permissions: luaskills.EmbeddedInputCapabilityDescriptorPermissions{"go.echo"}, Scope: luaskills.EmbeddedInputCapabilityScopeInvocation, MaxConcurrent: 1, MaxCallMs: 5000, MaxInputBytes: 1024, MaxOutputBytes: 1024, Effects: luaskills.EmbeddedInputCapabilityEffectsReadOnly, Idempotency: luaskills.EmbeddedInputCapabilityIdempotencyNone}, Handler: echo}
	if _, err = owned.pump.Register(ctx, []luaskills.EmbeddedHostCapability{capability}); err != nil {
		return err
	}
	// definition uses the real Lua callback envelope, requiring callback success explicitly.
	// definition 使用真实 Lua 回调信封，显式要求回调成功。
	definition := luaskills.EmbeddedInputModuleDefinition{PluginId: "go-example", Generation: "example-1", PackageRoot: packageRoot, DependenciesFile: "dependencies.yaml", Mounts: map[string]any{}, SecurityPartition: "go-example", Source: "return {call=function(a) local r=vulcan.capabilities.call('go.echo',a); assert(r.ok); return r.value end}", Exports: luaskills.EmbeddedInputModuleDefinitionExports{{Name: "call", InputSchema: true, OutputSchema: true}}}
	// policy explicitly chooses reusable in-process execution with no hidden worker-process fallback.
	// policy 显式选择可复用进程内执行，不隐式回退到工作进程。
	policy := luaskills.EmbeddedInputPluginPoolConfig{Kind: luaskills.EmbeddedInputPoolKindShared, MaxResidentVms: budgets.MaxResidentVms, MaxRunningCalls: budgets.MaxRunningCalls, MaxQueuedCalls: budgets.MaxQueuedCalls, Reuse: luaskills.EmbeddedInputInstanceReuseReusable, Backend: luaskills.EmbeddedInputExecutionBackendInProcess}
	poolPending, err := owned.runtime.RegisterPool(ctx, definition, policy, []string{"go.echo"}, "example-v1")
	pool, err := take(ctx, poolPending, err)
	if err != nil {
		return err
	}
	// operation is independent from its admission receipt; Wait observes native completion without cancelling execution.
	// operation 独立于入场回执；Wait 观察原生完成，不取消执行。
	// greeting is the exact expected callback value used to prove actual delivery.
	// greeting 是证明实际交付所用的精确预期回调值。
	const greeting = "hello from embedded Go"
	operationPending, err := pool.Submit(ctx, "call", greeting, luaskills.EmbeddedInputLuaInvocationContext{}, 5000)
	operation, err := take(ctx, operationPending, err)
	if err != nil {
		return err
	}
	// snapshot contains the real result and effect evidence, including failure or cancellation phases.
	// snapshot 包含真实结果及副作用证据，包括失败或取消阶段。
	snapshot, err := operation.Wait(ctx)
	if err != nil {
		return err
	}
	if snapshot.Phase != luaskills.EmbeddedOutputOperationPhaseSucceeded {
		return fmt.Errorf("Lua operation ended in %s: %#v", snapshot.Phase, snapshot)
	}
	if snapshot.Value == nil || *snapshot.Value != greeting {
		return fmt.Errorf("Lua callback result differs from original arguments: %#v", snapshot.Value)
	}
	fmt.Printf("Lua callback result: %v\n", *snapshot.Value)
	// forgotten removes the native operation; take separately releases the associated SDK command receipt.
	// forgotten 移除原生操作；take 单独释放关联 SDK 命令回执。
	forgotten, err := operation.Forget(ctx)
	_, err = take(ctx, forgotten, err)
	return err
}

// main reports business or retained-owner failure and exits nonzero; successful run proves full shutdown.
// main 报告业务或保留所有者失败并非零退出；成功执行证明完整关闭。
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Runtime scope, callback pump, driver receipts and transport closed.")
}
