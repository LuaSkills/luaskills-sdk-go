//go:build cgo

package luaskills

import (
	"context"
	"testing"
)

// TestEmbeddedClientNativeFinalization consumes generated closing declarations and both result union alternatives.
// TestEmbeddedClientNativeFinalization 消费生成的关闭声明及两种结果联合分支。
func TestEmbeddedClientNativeFinalization(t *testing.T) {
	runtime, root, _ := nativeTypedRuntime(t)
	// closing is explicitly present; its arguments preserve required JSON null through generated encoding.
	// closing 显式存在；参数通过生成编码保留必需的 JSON 空值。
	closing := &EmbeddedInputModuleFinalizer{Export: "shutdown", Arguments: nil, TimeoutMs: 1000}
	definition := EmbeddedInputModuleDefinition{
		PluginId: "go-typed-test", Generation: "typed-generation-1", PackageRoot: root,
		DependenciesFile: "dependencies.yaml", Mounts: map[string]any{}, SecurityPartition: "go-typed-test",
		Source:    "local called=false; local fail=false; return {call=function(a) called=true; fail=a; return nil end, shutdown=function() assert(called); if fail then error('closing failed') end; return nil end}",
		Exports:   EmbeddedInputModuleDefinitionExports{{Name: "call", InputSchema: true, OutputSchema: true}, {Name: "shutdown", InputSchema: true, OutputSchema: true}},
		Finalizer: &closing,
	}
	policy := EmbeddedInputPluginPoolConfig{Kind: EmbeddedInputPoolKindShared, MaxResidentVms: 2, MaxRunningCalls: 2, MaxQueuedCalls: 4, Reuse: EmbeddedInputInstanceReuseSingleCall, Backend: EmbeddedInputExecutionBackendInProcess}
	registered, err := runtime.RegisterPool(context.Background(), definition, policy, []string{"go.host"}, "typed-v1")
	pool := typedTake(t, registered, err)
	for _, fail := range []bool{false, true} {
		submitted, err := pool.Submit(context.Background(), "call", fail, EmbeddedInputLuaInvocationContext{}, 10000)
		operation := typedTake(t, submitted, err)
		snapshot, err := operation.Wait(driverTestContext(t))
		if err != nil || snapshot.Finalization == nil || *snapshot.Finalization == nil {
			t.Fatalf("closing result missing: %#v, %v", snapshot, err)
		}
		// stages retains explicit null separately from absent closing evidence.
		// stages 将显式空值与缺失关闭证据分开保留。
		stages := *snapshot.Finalization
		business, ok := stages.Business.(EmbeddedOutputOperationOutcomeVariant1)
		if !ok || business.Value != nil || stages.BusinessEffectCount != 0 || stages.Outcome == nil || *stages.Outcome == nil {
			t.Fatalf("business or closing presence lost: %#v", stages)
		}
		if fail {
			outcome, ok := (**stages.Outcome).(EmbeddedOutputOperationOutcomeVariant2)
			if !ok || snapshot.Phase != EmbeddedOutputOperationPhaseFailed || snapshot.Error == nil || *snapshot.Error == nil || outcome.Error != **snapshot.Error {
				t.Fatalf("independent closing failure lost: %#v", snapshot)
			}
		} else {
			outcome, ok := (**stages.Outcome).(EmbeddedOutputOperationOutcomeVariant1)
			if !ok || outcome.Value != nil || snapshot.Phase != EmbeddedOutputOperationPhaseSucceeded || snapshot.Value == nil || *snapshot.Value != nil {
				t.Fatalf("successful closing null lost: %#v", snapshot)
			}
		}
	}
}
