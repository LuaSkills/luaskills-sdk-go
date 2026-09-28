//go:build cgo

package luaskills

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// TestEmbeddedCapacityNativePolicyRevision verifies pinned state, exact CAS and control admission through the real core.
// TestEmbeddedCapacityNativePolicyRevision 通过真实核心验证固定状态、精确比较交换及控制入场。
func TestEmbeddedCapacityNativePolicyRevision(t *testing.T) {
	// Existing fixture retains and closes the actual runtime, driver and native transport.
	// 既有夹具保留并关闭实际运行时、驱动器及原生传输。
	runtime, root, _ := nativeTypedRuntime(t)
	// Read the original plugin policy instead of inventing another fixture ceiling.
	// 读取原插件策略，不编造另一套夹具上限。
	plugin, err := runtime.Plugin("go-typed-test")
	if err != nil {
		t.Fatal(err)
	}
	// Current native plugin limits are the sole source for inherited capacity bounds.
	// 当前原生插件限制是继承容量边界的唯一来源。
	pluginStatus, err := plugin.Status(context.Background())
	// No test-side quota accounting replaces this native snapshot.
	// 不用测试端额度记账替代此原生快照。
	limits := typedTake(t, pluginStatus, err).Config
	// The original shared owner permits all fixture sessions and one execution at a time.
	// 原公共所有者允许全部夹具会话及逐次单个执行。
	config := EmbeddedInputEmbeddedCapacityConfig{
		Resources: EmbeddedInputVmCapacityConfig{Kind: EmbeddedInputPoolKindShared,
			MaxResidentVms: limits.MaxResidentVms, MaxRunningCalls: 1},
		MaxQueuedCalls: limits.MaxQueuedCalls, MaxQueuedBytes: limits.MaxQueuedBytes,
	}
	// Exact retained acknowledgement creates the typed native handle.
	// 精确保留确认创建类型化原生句柄。
	registration, err := runtime.RegisterCapacity(context.Background(), "go-typed-test", config)
	// This handle never switches native capacity after errors.
	// 此句柄绝不在错误后切换原生容量。
	capacity := typedTake(t, registration, err)
	// Fixed VMs retain private counters across the aggregate policy revision.
	// 固定 VM 跨聚合策略修订保留私有计数器。
	definition := EmbeddedInputModuleDefinition{
		PluginId: "go-typed-test", Generation: "policy-session", PackageRoot: root,
		DependenciesFile: "dependencies.yaml", Mounts: map[string]any{}, SecurityPartition: "go-typed-test",
		Source: `-- Retain private state in this fixed VM.
-- 在此固定 VM 中保留私有状态。
local count=0
return {
-- Advance this instance's counter.
-- 递增此实例的计数器。
call=function() count=count+1; return count end}`,
		Exports: EmbeddedInputModuleDefinitionExports{{Name: "call", InputSchema: true, OutputSchema: true}},
	}
	// The immutable member declaration reserves no separate guarantee.
	// 不可变成员声明不预留独立保证。
	policy := EmbeddedInputPluginPoolConfig{Kind: EmbeddedInputPoolKindShared,
		MaxResidentVms: config.Resources.MaxResidentVms, MaxRunningCalls: config.Resources.MaxRunningCalls,
		MaxQueuedCalls: config.MaxQueuedCalls, Reuse: EmbeddedInputInstanceReuseSession,
		Backend: EmbeddedInputExecutionBackendInProcess}
	// Member admission does not run the Lua source.
	// 成员入场不运行 Lua 源码。
	registering, err := capacity.RegisterPool(context.Background(), definition, policy, []string{}, "policy-v1")
	// Preserve the original member through the revision.
	// 跨修订保留原成员。
	pool := typedTake(t, registering, err)
	// Retain each original fixed-session identity independently.
	// 独立保留各原固定会话身份。
	sessions := []*EmbeddedSession{}
	for index := uint64(0); index < config.Resources.MaxResidentVms; index++ {
		// Observe actual initialization separately from reservation.
		// 独立于预留观测实际初始化。
		pending, err := pool.OpenSession(context.Background(), 5000)
		// The paired result binds the session and its original initialization operation.
		// 配对结果绑定会话及其原初始化操作。
		opening := typedTake(t, pending, err)
		// A reserved but failed initialization must never count as a live test session.
		// 已预留但初始化失败的结果绝不能算作活动测试会话。
		initialized, err := opening.Initialization.Wait(driverTestContext(t))
		if err != nil || initialized.Phase != "succeeded" {
			t.Fatalf("initialization: %#v, %v", initialized, err)
		}
		// Release only the completed initialization history.
		// 仅释放已完成初始化历史。
		forgotten, err := opening.Initialization.Forget(context.Background())
		typedTake(t, forgotten, err)
		sessions = append(sessions, opening.Session)
	}
	// requireCode checks the exact native or driver error without accepting unrelated failures.
	// requireCode 检查精确原生或驱动错误，不接受无关失败。
	// failure and code are the actual error and required stable code; it reports through the test and returns no value.
	// failure 及 code 为实际错误及所需稳定错误码；通过测试报告，无返回值。
	requireCode := func(failure error, code string) {
		t.Helper()
		// The typed failure retains native evidence independently from observer lifetime.
		// 类型化失败独立于观察者寿命保留原生证据。
		var native *EmbeddedRuntimeError
		if !errors.As(failure, &native) || native.Code != code {
			t.Fatalf("expected %s, got %v", code, failure)
		}
	}
	// Keep failed work receipts until the configured lane is genuinely full.
	// 保留失败工作回执，直到配置通道确实已满。
	held := []*EmbeddedPending[EmbeddedSessionOpen]{}
	for index := uint64(0); index < runtime.client.driver.config.MaxWorkCommands; index++ {
		// Physical pressure rejects opening without consuming an additional VM.
		// 物理压力拒绝开启，不消费额外 VM。
		rejected, err := pool.OpenSession(context.Background(), 5000)
		if err != nil {
			t.Fatal(err)
		}
		_, err = rejected.Result(driverTestContext(t))
		requireCode(err, "capacity_exceeded")
		held = append(held, rejected)
	}
	_, err = pool.OpenSession(context.Background(), 5000)
	requireCode(err, "capacity_exceeded")
	// Reserved control admission remains available with all work receipts retained.
	// 全部工作回执保留时，预留控制入场仍可用。
	query, err := capacity.Policy(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if query.Receipt().Lane() != EmbeddedControlLane {
		t.Fatal("policy query lost control lane")
	}
	// This exact opaque predecessor must be echoed without numeric conversion.
	// 此精确不透明前驱必须原样回传，不进行数值转换。
	before := typedTake(t, query, nil)
	// The original fixed sessions exceed the smaller dedicated target.
	// 原固定会话超过更小的专用目标。
	target := config
	target.Resources.Kind = EmbeddedInputPoolKindDedicated
	target.Resources.MinResidentVms = 1
	target.Resources.MaxResidentVms = 1
	// The actual mutation also uses the reserved short control lane.
	// 实际变更同样使用预留短时控制通道。
	change, err := capacity.Revise(context.Background(), before.Revision, target)
	if err != nil {
		t.Fatal(err)
	}
	if change.Receipt().Lane() != EmbeddedControlLane {
		t.Fatal("policy revision lost control lane")
	}
	// Native acknowledgement alone proves the successful policy commit.
	// 仅原生确认能证明策略提交成功。
	revision := typedTake(t, change, nil)
	if revision == before.Revision {
		t.Fatal("changed policy retained stale revision")
	}
	// A stale predecessor must fail without automatic retry.
	// 过期前驱必须失败，不自动重试。
	stale, err := capacity.Revise(context.Background(), before.Revision, config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = stale.Result(driverTestContext(t))
	requireCode(err, "busy")
	if err := stale.Forget(); err != nil {
		t.Fatal(err)
	}
	// Atomic status includes real occupancy and the exact newly committed target.
	// 原子状态包含真实占用及精确新提交目标。
	current, err := capacity.Policy(context.Background())
	// These counters originate solely from the native runtime.
	// 这些计数仅来自原生运行时。
	snapshot := typedTake(t, current, err)
	if snapshot.Revision != revision || !snapshot.PendingConvergence ||
		snapshot.Capacity.Resources.Resident != uint64(len(sessions)) ||
		snapshot.Capacity.Config.Resources.MaxResidentVms != target.Resources.MaxResidentVms {
		t.Fatalf("revision lost real ownership: %#v", snapshot)
	}
	for _, receipt := range held {
		if err := receipt.Forget(); err != nil {
			t.Fatal(err)
		}
	}
	for _, session := range sessions {
		for _, expected := range []json.Number{"1", "2"} {
			// Exact original state must remain usable while resident convergence is pending.
			// 常驻收敛待完成时，精确原状态必须保持可用。
			pending, err := session.Submit(context.Background(), "call", nil, EmbeddedInputLuaInvocationContext{}, 5000)
			// Keep the actual business operation through its terminal observation.
			// 保留实际业务操作直到终态观测。
			operation := typedTake(t, pending, err)
			// The returned counter proves continuity in the original fixed VM.
			// 返回计数证明原固定 VM 内的连续性。
			result, err := operation.Wait(driverTestContext(t))
			if err != nil || result.Value == nil || *result.Value != expected {
				t.Fatalf("fixed state: %#v, %v", result, err)
			}
			// Completed operation history is released explicitly.
			// 显式释放已完成操作历史。
			forgotten, err := operation.Forget(context.Background())
			typedTake(t, forgotten, err)
		}
		// Closure is requested only after the original business calls have finished.
		// 仅在原业务调用完成后请求关闭。
		closing, err := session.RequestClose(context.Background())
		typedTake(t, closing, err)
	}
	// Bound observation using one deadline rather than renewing the wait each time.
	// 使用单个截止时间限制观测，不逐次续期。
	deadline := time.Now().Add(5 * time.Second)
	for {
		// Query the same native policy owner until real physical release.
		// 查询同一原生策略所有者，直到真实物理释放。
		observed, err := capacity.Policy(context.Background())
		snapshot = typedTake(t, observed, err)
		if snapshot.Capacity.Resources.Resident == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("fixed sessions failed to drain")
		}
		time.Sleep(time.Millisecond)
	}
	if snapshot.PendingConvergence {
		t.Fatal("released capacity still reports excess usage")
	}
	// Runtime closure keeps status readable but permanently fences policy mutations.
	// 运行时关闭保持状态可读，但永久封闭策略变更。
	closing, err := runtime.RequestClose(context.Background())
	typedTake(t, closing, err)
	current, err = capacity.Policy(context.Background())
	if !typedTake(t, current, err).Capacity.Closing {
		t.Fatal("parent closure missing from policy")
	}
	// The current predecessor cannot authorize reopening a closed parent.
	// 当前前驱不能授权重开已关闭父级。
	closed, err := capacity.Revise(context.Background(), revision, config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = closed.Result(driverTestContext(t))
	requireCode(err, "closed")
	if err := closed.Forget(); err != nil {
		t.Fatal(err)
	}
}

// TestEmbeddedCapacityNativeMembers validates generated commands, exact native identities and real member retirement.
// TestEmbeddedCapacityNativeMembers 验证生成命令、精确原生身份及真实成员退役。
func TestEmbeddedCapacityNativeMembers(t *testing.T) {
	// The existing fixture owns actual runtime, driver and transport cleanup.
	// 既有夹具拥有实际运行时、驱动器及传输清理。
	runtime, root, _ := nativeTypedRuntime(t)
	// One capacity reservation covers both members without duplicating their minimum.
	// 一个容量预留覆盖两个成员，不重复其最小值。
	config := EmbeddedInputEmbeddedCapacityConfig{
		Resources: EmbeddedInputVmCapacityConfig{
			Kind: EmbeddedInputPoolKindDedicated, MinResidentVms: 1, MaxResidentVms: 2, MaxRunningCalls: 1,
		},
		MaxQueuedCalls: 4, MaxQueuedBytes: 4096,
	}
	// Registration returns an acknowledged native capacity handle.
	// 注册返回已确认原生容量句柄。
	registration, err := runtime.RegisterCapacity(context.Background(), "go-typed-test", config)
	// The projection must preserve the identity used by every later command.
	// 投影必须保留每个后续命令使用的身份。
	capacity := typedTake(t, registration, err)
	// Rebinding uses the exact runtime and original capacity identity without probing.
	// 重绑定使用精确运行时及原容量身份，不探测。
	bound, err := runtime.Capacity(capacity.CapacityID())
	if err != nil || bound.CapacityID() != capacity.CapacityID() {
		t.Fatalf("capacity identity changed: %v", err)
	}
	// Status uses independently reserved control admission.
	// 状态使用独立预留控制入场。
	status, err := capacity.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Receipt().Lane() != EmbeddedControlLane {
		t.Fatal("capacity status lost control admission")
	}
	if typedTake(t, status, nil).CommittedResidentVms != 1 {
		t.Fatal("empty capacity lost its guarantee")
	}
	// Retain both immutable members for ordered native closure.
	// 保留两个不可变成员，用于有序原生关闭。
	pools := []*EmbeddedPool{}
	for _, generation := range []string{"first", "second"} {
		// Module-local state makes accidental cross-member sharing observable.
		// 模块局部状态使意外跨成员共享可观察。
		definition := EmbeddedInputModuleDefinition{
			PluginId: "go-typed-test", Generation: generation, PackageRoot: root, DependenciesFile: "dependencies.yaml",
			Mounts: map[string]any{}, SecurityPartition: "go-typed-test",
			Source: `-- Keep state private to this exact module.
-- 将状态保持在此精确模块内。
local count=0
-- Return the next count without arguments or external effects.
-- 返回下一个计数，不使用参数或产生外部副作用。
return {call=function() count=count+1; return count end}`,
			Exports: EmbeddedInputModuleDefinitionExports{{Name: "call", InputSchema: true, OutputSchema: true}},
		}
		// Members declare no separate reservation and stay inside capacity limits.
		// 成员不声明独立预留，并保持在容量限制内。
		policy := EmbeddedInputPluginPoolConfig{
			Kind: EmbeddedInputPoolKindDedicated, MaxResidentVms: 1, MaxRunningCalls: 1, MaxQueuedCalls: 4,
			Reuse: EmbeddedInputInstanceReuseReusable, Backend: EmbeddedInputExecutionBackendInProcess,
		}
		// The generated nullable presence field carries the exact capacity identity.
		// 生成的可空存在性字段携带精确容量身份。
		pending, err := capacity.RegisterPool(context.Background(), definition, policy, []string{}, generation)
		pools = append(pools, typedTake(t, pending, err))
	}
	for _, pool := range pools {
		for _, expected := range []json.Number{"1", "2"} {
			// The actual operation result is observed independently of its submission acknowledgement.
			// 实际操作结果独立于提交确认被观测。
			pending, err := pool.Submit(context.Background(), "call", nil, EmbeddedInputLuaInvocationContext{}, 5000)
			// Preserve the admitted operation until its native result is terminal.
			// 保留已入场操作，直到其原生结果达到终态。
			operation := typedTake(t, pending, err)
			// Lua counters must remain independent in each exact member.
			// Lua 计数器必须在每个精确成员中保持独立。
			result, err := operation.Wait(driverTestContext(t))
			if err != nil || result.Value == nil || *result.Value != expected {
				t.Fatalf("member state mismatch: %#v, %v", result, err)
			}
			// Explicit forgetting releases the terminal operation record.
			// 显式遗忘释放终态操作记录。
			forgotten, err := operation.Forget(context.Background())
			typedTake(t, forgotten, err)
		}
	}
	// Closure prevents new member work but does not prematurely drop reservations.
	// 关闭阻止新成员工作，但不提前丢弃预留。
	closing, err := capacity.RequestClose(context.Background())
	typedTake(t, closing, err)
	// Busy rejects capacity deletion while exact member metadata remains.
	// 精确成员元数据仍在时，忙碌拒绝容量删除。
	blocked, err := capacity.Forget(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_, err = blocked.Result(driverTestContext(t))
	// Only the declared native busy failure proves member ownership was retained.
	// 只有声明的原生忙碌错误才能证明成员归属仍被保留。
	var failure *EmbeddedRuntimeError
	if !errors.As(err, &failure) || failure.Code != "busy" {
		t.Fatalf("capacity member removal did not report busy: %v", err)
	}
	if err := blocked.Forget(); err != nil {
		t.Fatal(err)
	}
	// Wait for actual native VM destruction before removing member metadata.
	// 移除成员元数据前等待实际原生 VM 销毁。
	deadline := time.Now().Add(5 * time.Second)
	for {
		// Each snapshot reads the sole native capacity owner.
		// 每个快照读取唯一原生容量所有者。
		observed, err := capacity.Status(context.Background())
		if typedTake(t, observed, err).Resources.Resident == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("capacity members did not retire")
		}
		time.Sleep(time.Millisecond)
	}
	for _, pool := range pools {
		// Physical completion permits explicit member metadata release.
		// 物理完成允许显式释放成员元数据。
		forgotten, err := pool.Forget(context.Background())
		typedTake(t, forgotten, err)
	}
	// Empty capacity ownership retains its guarantee until this exact removal.
	// 空容量归属在此精确移除前保留其保证。
	observed, err := capacity.Status(context.Background())
	if typedTake(t, observed, err).CommittedResidentVms != 1 {
		t.Fatal("guarantee released before forgetting")
	}
	// The final acknowledgement returns the capacity reservation to its plugin.
	// 最终确认将容量预留归还所属插件。
	forgotten, err := capacity.Forget(context.Background())
	typedTake(t, forgotten, err)
}
