//go:build cgo

package luaskills

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

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
