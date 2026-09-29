//go:build cgo

package luaskills

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

// TestEmbeddedInitializationPolicyNative proves actual callback authority through typed registration in both placements.
// TestEmbeddedInitializationPolicyNative 通过两种归属的类型化注册证明真实回调权威。
// t owns fresh native fixtures for inherited, denied and explicit lists; all original receipts are retained until observed.
// t 为继承、拒绝及显式列表拥有新原生夹具；全部原回执保留至被观测。
func TestEmbeddedInitializationPolicyNative(t *testing.T) {
	for _, names := range [][]string{nil, {}, {"go.callback"}} {
		for _, grouped := range []bool{false, true} {
			t.Run(fmt.Sprintf("names-%v-inherited-%t-grouped-%t", names, names == nil, grouped), func(t *testing.T) {
				// The existing fixture owns shutdown; the callback pump reports actual entries from this exact runtime.
				// 既有夹具拥有关闭；回调泵报告来自此精确运行时的真实进入。
				runtime, root, _ := nativeTypedRuntime(t)
				pump := nativePumpTest(t, runtime.client.driver.transport, runtime.RuntimeID(), EmbeddedCallbackPumpConfig{1, 1, 1})
				var mutex sync.Mutex
				calls := []string{}
				pumpRegister(t, pump, pumpCapability(func(value any, callback *EmbeddedHostCallbackContext) (any, error) {
					mutex.Lock()
					calls = append(calls, value.(string))
					mutex.Unlock()
					return value, callback.ReportEffects(EmbeddedInputEffectStateCommitted)
				}))
				allowed := names == nil || len(names) > 0
				source := fmt.Sprintf("local init=vulcan.host.call('go.callback','initialization'); assert(init.ok == %t); return {call=function() return vulcan.host.call('go.callback','business').value end}", allowed)
				definition := EmbeddedInputModuleDefinition{PluginId: "go-typed-test", Generation: "init-generation", PackageRoot: root, DependenciesFile: "dependencies.yaml", Mounts: map[string]any{}, SecurityPartition: "go-typed-test", Source: source, Exports: EmbeddedInputModuleDefinitionExports{{Name: "call", InputSchema: true, OutputSchema: true}}}
				policy := EmbeddedInputPluginPoolConfig{Kind: EmbeddedInputPoolKindShared, MaxResidentVms: 2, MaxRunningCalls: 2, MaxQueuedCalls: 4, Reuse: EmbeddedInputInstanceReuseReusable, Backend: EmbeddedInputExecutionBackendInProcess}
				// Bind one exact registration method; grouped authority must never fall back to independent placement.
				// 绑定一个精确注册方法；分组权威绝不能回退到独立归属。
				register := runtime.RegisterPoolWithInitializationCapabilities
				if grouped {
					pending, err := runtime.RegisterCapacity(context.Background(), definition.PluginId, EmbeddedInputEmbeddedCapacityConfig{
						Resources:      EmbeddedInputVmCapacityConfig{Kind: policy.Kind, MaxResidentVms: policy.MaxResidentVms, MaxRunningCalls: policy.MaxRunningCalls},
						MaxQueuedCalls: policy.MaxQueuedCalls, MaxQueuedBytes: 4096,
					})
					capacity := typedTake(t, pending, err)
					register = capacity.RegisterPoolWithInitializationCapabilities
				}
				pending, err := register(context.Background(), definition, policy, []string{"go.host"}, "init-v1", names)
				pool := typedTake(t, pending, err)
				if grouped {
					warming, err := pool.PrewarmInstance(context.Background(), EmbeddedInputLuaInvocationContext{}, 5000)
					prewarm := typedTake(t, warming, err)
					done, err := prewarm.Wait(driverTestContext(t))
					if err != nil || done.Phase != EmbeddedOutputOperationPhaseSucceeded {
						t.Fatalf("prewarm failed: %#v, %v", done, err)
					}
					forgotten, err := prewarm.Forget(context.Background())
					typedTake(t, forgotten, err)
				}
				pendingCall, err := pool.Submit(context.Background(), "call", nil, EmbeddedInputLuaInvocationContext{}, 5000)
				business := typedTake(t, pendingCall, err)
				done, err := business.Wait(driverTestContext(t))
				if err != nil || done.Phase != EmbeddedOutputOperationPhaseSucceeded || done.Value == nil || *done.Value != "business" {
					t.Fatalf("business authority changed: %#v, %v", done, err)
				}
				expected := []string{"business"}
				if allowed {
					expected = []string{"initialization", "business"}
				}
				mutex.Lock()
				observed := append([]string{}, calls...)
				mutex.Unlock()
				if !reflect.DeepEqual(observed, expected) {
					t.Fatalf("unexpected actual host effects: %v, expected %v", observed, expected)
				}
				forgotten, err := business.Forget(context.Background())
				typedTake(t, forgotten, err)
				// Reject both unknown callbacks and attempts to create missing general permissions before registering a pool.
				// 在注册池前拒绝未知回调及创建缺失通用权限的尝试。
				for _, invalid := range []struct {
					// names carries the exact requested narrowing, and permissions carries the independent host grants.
					// names 携带精确请求收窄集合，permissions 携带独立宿主授权。
					names, permissions []string
				}{{[]string{"missing.callback"}, []string{"go.host"}}, {[]string{"go.callback"}, []string{}}} {
					rejected, err := register(context.Background(), definition, policy, invalid.permissions, "invalid-init", invalid.names)
					if err != nil {
						t.Fatal(err)
					}
					_, err = rejected.Result(driverTestContext(t))
					var failure *EmbeddedRuntimeError
					if !errors.As(err, &failure) || failure.Code != "permission_denied" {
						t.Fatalf("initialization declaration expanded authority: %v", err)
					}
					if err := rejected.Forget(); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}
