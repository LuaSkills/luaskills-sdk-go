//go:build cgo

package luaskills

import (
	"context"
	"testing"
)

// TestEmbeddedClientNativeSessionFinalization consumes the generated session closing identity and independent outcome.
// TestEmbeddedClientNativeSessionFinalization 消费生成的会话关闭身份及独立结果。
func TestEmbeddedClientNativeSessionFinalization(t *testing.T) {
	runtime, root, _ := nativeTypedRuntime(t)
	closing := &EmbeddedInputModuleFinalizer{Export: "shutdown", Arguments: nil, TimeoutMs: 1000}
	definition := EmbeddedInputModuleDefinition{
		PluginId: "go-typed-test", Generation: "typed-generation-1", PackageRoot: root,
		DependenciesFile: "dependencies.yaml", Mounts: map[string]any{}, SecurityPartition: "go-typed-test",
		Source:    "local n=0; return {call=function() n=n+1; return tostring(n) end, shutdown=function() return tostring(n) end}",
		Exports:   EmbeddedInputModuleDefinitionExports{{Name: "call", InputSchema: true, OutputSchema: true}, {Name: "shutdown", InputSchema: true, OutputSchema: true}},
		Finalizer: &closing,
	}
	policy := EmbeddedInputPluginPoolConfig{Kind: EmbeddedInputPoolKindShared, MaxResidentVms: 2, MaxRunningCalls: 2, MaxQueuedCalls: 4, Reuse: EmbeddedInputInstanceReuseSession, Backend: EmbeddedInputExecutionBackendInProcess}
	registered, err := runtime.RegisterPool(context.Background(), definition, policy, []string{"go.host"}, "typed-v1")
	pool := typedTake(t, registered, err)
	pending, err := pool.OpenSession(context.Background(), 10000)
	opening := typedTake(t, pending, err)
	initialized, err := opening.Initialization.Wait(driverTestContext(t))
	if err != nil || initialized.Phase != EmbeddedOutputOperationPhaseSucceeded {
		t.Fatalf("initialization: %#v, %v", initialized, err)
	}
	submitted, err := opening.Session.Submit(context.Background(), "call", nil, EmbeddedInputLuaInvocationContext{}, 10000)
	business := typedTake(t, submitted, err)
	result, err := business.Wait(driverTestContext(t))
	if err != nil || result.Value == nil || *result.Value != "1" {
		t.Fatalf("business: %#v, %v", result, err)
	}
	requested, err := opening.Session.RequestClose(context.Background())
	typedTake(t, requested, err)
	var session EmbeddedOutputEmbeddedSessionSnapshot
	embeddedPoll(t, func() any {
		pending, err := opening.Session.Status(context.Background())
		session = typedTake(t, pending, err)
		return session
	}, func(value any) bool {
		return value.(EmbeddedOutputEmbeddedSessionSnapshot).Phase == EmbeddedOutputEmbeddedSessionPhaseClosed
	})
	if session.FinalizationOperation == nil || *session.FinalizationOperation == business.OperationID() {
		t.Fatalf("independent closing identity: %#v", session)
	}
	operation, err := runtime.Operation(*session.FinalizationOperation)
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := operation.Wait(driverTestContext(t))
	if err != nil || outcome.Phase != EmbeddedOutputOperationPhaseSucceeded || outcome.Finalization == nil || *outcome.Finalization == nil {
		t.Fatalf("closing: %#v, %v", outcome, err)
	}
	stages := *outcome.Finalization
	if stages.Outcome == nil || *stages.Outcome == nil {
		t.Fatalf("closing outcome missing: %#v", stages)
	}
	value, ok := (**stages.Outcome).(EmbeddedOutputOperationOutcomeVariant1)
	if !ok || value.Value != "1" {
		t.Fatalf("same VM outcome: %#v", value)
	}
	again, err := business.Wait(driverTestContext(t))
	if err != nil || again.Value == nil || *again.Value != "1" || again.Finalization != nil {
		t.Fatalf("business rewritten: %#v, %v", again, err)
	}
}

// TestEmbeddedClientNativeReusableFinalization consumes typed publication pagination and same-VM closing outcomes.
// TestEmbeddedClientNativeReusableFinalization 消费类型化发布分页及同 VM 关闭结果。
func TestEmbeddedClientNativeReusableFinalization(t *testing.T) {
	runtime, root, _ := nativeTypedRuntime(t)
	closing := &EmbeddedInputModuleFinalizer{Export: "shutdown", Arguments: nil, TimeoutMs: 1000}
	definition := EmbeddedInputModuleDefinition{
		PluginId: "go-typed-test", Generation: "typed-generation-1", PackageRoot: root,
		DependenciesFile: "dependencies.yaml", Mounts: map[string]any{}, SecurityPartition: "go-typed-test",
		Source:    "local n=0; return {call=function() n=n+1; return tostring(n) end, shutdown=function() return tostring(n) end}",
		Exports:   EmbeddedInputModuleDefinitionExports{{Name: "call", InputSchema: true, OutputSchema: true}, {Name: "shutdown", InputSchema: true, OutputSchema: true}},
		Finalizer: &closing,
	}
	policy := EmbeddedInputPluginPoolConfig{Kind: EmbeddedInputPoolKindShared, MaxResidentVms: 2, MaxRunningCalls: 2, MaxQueuedCalls: 4, Reuse: EmbeddedInputInstanceReuseReusable, Backend: EmbeddedInputExecutionBackendInProcess}
	registered, err := runtime.RegisterPool(context.Background(), definition, policy, []string{"go.host"}, "typed-v1")
	pool := typedTake(t, registered, err)
	// Keep the latest real business identity, which was allocated after the VM's closing reservation.
	// 保留最新真实业务身份，其分配晚于 VM 的关闭预留。
	var business *EmbeddedOperation
	for _, expected := range []string{"1", "2"} {
		submitted, err := pool.Submit(context.Background(), "call", nil, EmbeddedInputLuaInvocationContext{}, 10000)
		business = typedTake(t, submitted, err)
		result, err := business.Wait(driverTestContext(t))
		if err != nil || result.Value == nil || *result.Value != expected {
			t.Fatalf("business: %#v, %v", result, err)
		}
	}
	requested, err := pool.RequestClose(context.Background())
	typedTake(t, requested, err)
	poolID, after := pool.PoolID(), business.OperationID()
	var page EmbeddedOutputOperationPage
	embeddedPoll(t, func() any {
		pending, err := runtime.ListOperations(context.Background(), &poolID, &after, 1)
		page = typedTake(t, pending, err)
		return page
	}, func(value any) bool { return len(value.(EmbeddedOutputOperationPage).OperationIds) != 0 })
	if len(page.OperationIds) != 1 || page.AfterOperationId == nil || *page.AfterOperationId != page.OperationIds[0] {
		t.Fatalf("closing page: %#v", page)
	}
	operation, err := runtime.Operation(page.OperationIds[0])
	if err != nil {
		t.Fatal(err)
	}
	result, err := operation.Wait(driverTestContext(t))
	if err != nil || result.Phase != EmbeddedOutputOperationPhaseSucceeded || result.Finalization == nil || *result.Finalization == nil {
		t.Fatalf("closing: %#v, %v", result, err)
	}
	module, ok := result.Context.(EmbeddedOutputOperationContextVariant2)
	if !ok || module.FinalizationInstanceId == nil || *module.FinalizationInstanceId == nil || **module.FinalizationInstanceId == "" || module.Caller.SessionId != nil {
		t.Fatalf("instance authority: %#v", result.Context)
	}
	stages := *result.Finalization
	if stages.Outcome == nil || *stages.Outcome == nil {
		t.Fatalf("closing outcome: %#v", stages)
	}
	value, ok := (**stages.Outcome).(EmbeddedOutputOperationOutcomeVariant1)
	if !ok || value.Value != "2" {
		t.Fatalf("same VM result: %#v", value)
	}
	again, err := business.Wait(driverTestContext(t))
	if err != nil || again.Value == nil || *again.Value != "2" || again.Finalization != nil {
		t.Fatalf("business rewritten: %#v, %v", again, err)
	}
}

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
