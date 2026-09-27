//go:build cgo

package luaskills

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// embeddedTestConfig provides one explicit bounded native fixture; production has no copied defaults.
// embeddedTestConfig 提供一个显式有界原生夹具；生产不复制这些默认值。
func embeddedTestConfig() EmbeddedTransportConfig {
	return EmbeddedTransportConfig{2, 8, 262144, 32768, 65536}
}

// nativeEmbeddedTest requires deliberate native execution and owns final root cleanup through t.Cleanup.
// nativeEmbeddedTest 要求显式原生执行，并通过 t.Cleanup 拥有最终根清理。
func nativeEmbeddedTest(t *testing.T) *EmbeddedTransport {
	t.Helper()
	if os.Getenv("LUASKILLS_NATIVE_E2E") != "1" {
		t.Skip("LUASKILLS_NATIVE_E2E is not enabled")
	}
	transport, err := NewEmbeddedTransport(embeddedTestConfig())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := transport.ReleaseResults(); err != nil {
			t.Error(err)
		}
		if err := transport.Close(); err != nil {
			t.Error(err)
		}
		if err := transport.Free(); err != nil {
			t.Error(err)
		}
	})
	return transport
}

// embeddedRequest requires one exact successful command and returns its delivered JSON data.
// embeddedRequest 要求一个精确命令成功并返回已交付 JSON 数据。
func embeddedRequest(t *testing.T, transport *EmbeddedTransport, command map[string]any) any {
	t.Helper()
	result, err := transport.Request(command)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// embeddedPoll bounds native fixture waits by wall time while repeatedly reading actual core state.
// embeddedPoll 按墙钟时间限制原生夹具等待，同时重复读取实际核心状态。
func embeddedPoll(t *testing.T, read func() any, done func(any) bool) any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		value := read()
		if done(value) {
			return value
		}
		if time.Now().After(deadline) {
			t.Fatalf("native fixture did not drain: %#v", value)
		}
		time.Sleep(time.Millisecond)
	}
}

// TestEmbeddedNativeIdentityAndProtocol verifies the real C layout, 64-bit budgets and independent roots.
// TestEmbeddedNativeIdentityAndProtocol 验证真实 C 布局、64 位预算及独立根。
func TestEmbeddedNativeIdentityAndProtocol(t *testing.T) {
	first := nativeEmbeddedTest(t)
	second := nativeEmbeddedTest(t)
	if first.TransportID() == 0 || first.TransportID() == second.TransportID() {
		t.Fatal("root identities are not independent")
	}
	description := embeddedRequest(t, first, map[string]any{"type": "describe"}).(map[string]any)
	names := []any{}
	for _, name := range EmbeddedRootCommands() {
		names = append(names, name)
	}
	if !reflect.DeepEqual(description["commands"], names) {
		t.Fatalf("core commands disagree with packaged contract: %#v", description)
	}
	if err := first.Free(); err == nil {
		t.Fatal("freed open root")
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	reserved := embeddedRequest(t, second, map[string]any{"type": "runtime_reserve"}).(map[string]any)["runtime_id"]
	// Clean the exact returned registration while the other root remains independently closed.
	// 清理精确返回的注册，同时另一根保持独立关闭。
	embeddedRequest(t, second, map[string]any{"type": "runtime_close", "runtime_id": reserved})
	embeddedRequest(t, second, map[string]any{"type": "runtime_free", "runtime_id": reserved})
}

// TestEmbeddedNativeLargeBudgets validates uint64 JSON output beyond JavaScript's exact integer range.
// TestEmbeddedNativeLargeBudgets 验证超出 JavaScript 精确整数范围的 uint64 JSON 输出。
func TestEmbeddedNativeLargeBudgets(t *testing.T) {
	if os.Getenv("LUASKILLS_NATIVE_E2E") != "1" {
		t.Skip("LUASKILLS_NATIVE_E2E is not enabled")
	}
	config := embeddedTestConfig()
	config.MaxRuntimes = 9007199254740993
	transport, err := NewEmbeddedTransport(config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := transport.Close(); err != nil {
			t.Error(err)
		}
		if err := transport.Free(); err != nil {
			t.Error(err)
		}
	}()
	description := embeddedRequest(t, transport, map[string]any{"type": "describe"}).(map[string]any)
	if description["limits"].(map[string]any)["max_runtimes"] != json.Number("9007199254740993") {
		t.Fatal("native budget rounded")
	}
	copied := transport.Config()
	copied.MaxRuntimes = 1
	if transport.Config().MaxRuntimes != config.MaxRuntimes {
		t.Fatal("configuration alias escaped")
	}
}

// controlledEmbeddedNative injects faults around the actual C adapter without replacing successful native execution.
// controlledEmbeddedNative 在实际 C 适配器周围注入故障，不替换成功的原生执行。
type controlledEmbeddedNative struct {
	embeddedNative
	releaseFn func(uint64, embeddedResult) EmbeddedNativeStatus
	copyFn    func(embeddedResult, uint64) ([]byte, error)
}

// release invokes the configured interception or delegates the exact descriptor to C.
// release 调用已配置拦截，或将精确描述符委托给 C。
func (n *controlledEmbeddedNative) release(id uint64, result embeddedResult) EmbeddedNativeStatus {
	if n.releaseFn != nil {
		return n.releaseFn(id, result)
	}
	return n.embeddedNative.release(id, result)
}

// copy invokes the configured reader interception or copies the actual native allocation.
// copy 调用已配置读取者拦截，或复制实际原生分配。
func (n *controlledEmbeddedNative) copy(result embeddedResult, limit uint64) ([]byte, error) {
	if n.copyFn != nil {
		return n.copyFn(result, limit)
	}
	return n.embeddedNative.copy(result, limit)
}

// TestEmbeddedNativeRetainedResult verifies actual unreleased allocation ownership and original success/error evidence.
// TestEmbeddedNativeRetainedResult 验证实际未释放分配所有权及原始成功／错误证据。
func TestEmbeddedNativeRetainedResult(t *testing.T) {
	transport := nativeEmbeddedTest(t)
	native := transport.native
	controlled := &controlledEmbeddedNative{embeddedNative: native}
	transport.native = controlled
	controlled.releaseFn = func(uint64, embeddedResult) EmbeddedNativeStatus { return EmbeddedNativeBusy }
	_, err := transport.Request(map[string]any{"type": "runtime_reserve"})
	var release *EmbeddedResultReleaseError
	if !errors.As(err, &release) {
		t.Fatalf("missing release error: %v", err)
	}
	delivered, err := release.DeliveredResult()
	if err != nil {
		t.Fatal(err)
	}
	runtimeID := delivered.(map[string]any)["runtime_id"].(string)
	mutated := release.ResponseBytes()
	mutated[0] = '!'
	if _, err := release.DeliveredResult(); err != nil {
		t.Fatal("response evidence was mutable")
	}
	if transport.RetainedResults() != 1 {
		t.Fatal("allocation was not retained")
	}
	if err := transport.Free(); err == nil {
		t.Fatal("free bypassed retained allocation")
	}
	_, err = transport.Request(map[string]any{"type": "runtime_status", "runtime_id": "absent"})
	if !errors.As(err, &release) {
		t.Fatal(err)
	}
	_, err = release.DeliveredResult()
	var business *EmbeddedRuntimeError
	if !errors.As(err, &business) || business.Code != "not_found" {
		t.Fatalf("lost native business failure: %v", err)
	}
	controlled.releaseFn = nil
	if err := transport.ReleaseResults(); err != nil {
		t.Fatal(err)
	}
	if transport.RetainedResults() != 0 {
		t.Fatal("allocations leaked")
	}
	embeddedRequest(t, transport, map[string]any{"type": "runtime_close", "runtime_id": runtimeID})
	embeddedRequest(t, transport, map[string]any{"type": "runtime_free", "runtime_id": runtimeID})
}

// TestEmbeddedNativeConcurrentReader prevents recovery/free from racing a real buffer reader while controls remain live.
// TestEmbeddedNativeConcurrentReader 防止恢复／释放与真实缓冲读取者竞争，同时保持控制可用。
func TestEmbeddedNativeConcurrentReader(t *testing.T) {
	transport := nativeEmbeddedTest(t)
	native := transport.native
	entered := make(chan struct{})
	resume := make(chan struct{})
	done := make(chan error, 1)
	transport.native = &controlledEmbeddedNative{embeddedNative: native, copyFn: func(result embeddedResult, limit uint64) ([]byte, error) {
		close(entered)
		<-resume
		return native.copy(result, limit)
	}}
	go func() { _, err := transport.Request(map[string]any{"type": "describe"}); done <- err }()
	<-entered
	if err := transport.ReleaseResults(); err == nil {
		t.Error("recovery raced reader")
	}
	if err := transport.Free(); err == nil {
		t.Error("free raced reader")
	}
	if err := transport.Close(); err != nil {
		t.Error(err)
	}
	close(resume)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

// nativeEmbeddedRuntime creates one initialized fixture and always removes its exact runtime registration.
// nativeEmbeddedRuntime 创建一个已初始化夹具，并始终移除其精确运行时注册。
func nativeEmbeddedRuntime(t *testing.T) (*EmbeddedTransport, string, func(map[string]any) any, func(string) string) {
	t.Helper()
	transport := nativeEmbeddedTest(t)
	root := t.TempDir()
	packageRoot := filepath.Join(root, "system_lua_lib", "go-embedded-test")
	if err := os.MkdirAll(packageRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packageRoot, "dependencies.yaml"), []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runtimeID := embeddedRequest(t, transport, map[string]any{"type": "runtime_reserve"}).(map[string]any)["runtime_id"].(string)
	t.Cleanup(func() {
		embeddedRequest(t, transport, map[string]any{"type": "runtime_close", "runtime_id": runtimeID})
		embeddedPoll(t, func() any {
			return embeddedRequest(t, transport, map[string]any{"type": "runtime_status", "runtime_id": runtimeID})
		}, func(value any) bool { return value.(map[string]any)["closed"] == true })
		embeddedRequest(t, transport, map[string]any{"type": "runtime_free", "runtime_id": runtimeID})
	})
	options, err := CreateEngineOptions(root, map[string]any{"system_lua_lib_dir": filepath.Join(root, "system_lua_lib"), "allow_network_download": false}, nil)
	if err != nil {
		t.Fatal(err)
	}
	limits := map[string]any{"max_registered_plugins": 4, "max_registered_pools": 4, "max_sessions": 4, "max_registered_capabilities": 4, "max_resident_vms": 2, "max_running_calls": 2, "max_queued_calls": 4, "max_queued_bytes": 4096, "max_operations": 16, "max_effect_records_per_operation": 8, "max_effect_bytes_per_operation": 8192, "max_host_requests": 4, "max_host_request_bytes": 8192, "max_value_bytes": 1024}
	embeddedRequest(t, transport, map[string]any{"type": "runtime_initialize", "runtime_id": runtimeID, "engine_options": options, "runtime_config": limits})
	status := embeddedRequest(t, transport, map[string]any{"type": "runtime_status", "runtime_id": runtimeID}).(map[string]any)
	if status["initialization"] != "ready" {
		t.Fatalf("runtime initialization failed: %#v", status)
	}
	command := func(operation map[string]any) any {
		return embeddedRequest(t, transport, map[string]any{"type": "runtime", "runtime_id": runtimeID, "operation": operation})
	}
	plugin := map[string]any{}
	for _, name := range []string{"max_registered_pools", "max_sessions", "max_resident_vms", "max_running_calls", "max_queued_calls", "max_queued_bytes", "max_operations"} {
		plugin[name] = limits[name]
	}
	command(map[string]any{"type": "plugin_register", "plugin_id": "go-embedded-test", "config": plugin})
	pool := func(source string) string {
		return command(map[string]any{"type": "pool_register", "definition": map[string]any{"plugin_id": "go-embedded-test", "generation": "go-generation-1", "package_root": packageRoot, "dependencies_file": "dependencies.yaml", "workspace_root": nil, "cwd": nil, "mounts": map[string]any{}, "security_partition": "go-test", "source": source, "exports": []any{map[string]any{"name": "call", "input_schema": true, "output_schema": true}}}, "policy": map[string]any{"kind": "shared", "min_resident_vms": 0, "max_resident_vms": 2, "max_running_calls": 2, "max_queued_calls": 4, "reuse": "reusable", "serial": false, "backend": "in_process", "idle_ttl_ms": nil, "max_uses": nil}, "permissions": []any{"go.host"}, "execution_revision": "go-v1"}).(map[string]any)["pool_id"].(string)
	}
	return transport, runtimeID, command, pool
}

// embeddedSubmit builds the exact generated call shape; arguments are application data, never caller authority.
// embeddedSubmit 构造精确生成调用形状；arguments 是应用数据，绝不是调用方权限。
func embeddedSubmit(command func(map[string]any) any, pool string, arguments any) string {
	return command(map[string]any{"type": "call_submit", "timeout_ms": 10000, "call": map[string]any{"pool_id": pool, "export": "call", "arguments": arguments, "context": map[string]any{"request_context": nil, "client_budget": nil, "tool_config": nil}}}).(map[string]any)["operation_id"].(string)
}

// embeddedTerminal waits only for actual terminal core phases and returns the complete original snapshot.
// embeddedTerminal 仅等待实际核心终态，并返回完整原始快照。
func embeddedTerminal(t *testing.T, command func(map[string]any) any, id string) map[string]any {
	return embeddedPoll(t, func() any { return command(map[string]any{"type": "operation_status", "operation_id": id}) }, func(value any) bool {
		phase := value.(map[string]any)["phase"]
		return phase == "succeeded" || phase == "failed" || phase == "cancelled"
	}).(map[string]any)
}

// TestEmbeddedNativeLua verifies real execution, explicit null/false and Lua business failure through the new C surface.
// TestEmbeddedNativeLua 通过新 C 表面验证真实执行、显式空值／假值及 Lua 业务失败。
func TestEmbeddedNativeLua(t *testing.T) {
	_, _, command, pool := nativeEmbeddedRuntime(t)
	poolID := pool("return {call=function(a) return a end}")
	for _, value := range []any{nil, false, json.Number("0.25"), "中文"} {
		id := embeddedSubmit(command, poolID, value)
		result := embeddedTerminal(t, command, id)
		if result["phase"] != "succeeded" || !reflect.DeepEqual(result["value"], value) {
			t.Fatalf("Lua value changed: %#v", result)
		}
		command(map[string]any{"type": "operation_forget", "operation_id": id})
	}
	failed := embeddedTerminal(t, command, embeddedSubmit(command, pool("return {call=function(a) error('go expected Lua error') end}"), nil))
	if failed["phase"] != "failed" || !strings.Contains(failed["error"].(map[string]any)["message"].(string), "go expected") {
		t.Fatalf("Lua failure lost: %#v", failed)
	}
}

// TestEmbeddedNativeQueuedCallback verifies trusted identity, explicit cancellation and late committed host evidence.
// TestEmbeddedNativeQueuedCallback 验证可信身份、显式取消及迟到已提交宿主证据。
func TestEmbeddedNativeQueuedCallback(t *testing.T) {
	transport, _, command, pool := nativeEmbeddedRuntime(t)
	descriptor := map[string]any{"name": "go.callback", "version": "1.0.0", "description": "Go native queue integration", "input_schema": true, "output_schema": true, "execution": "queued", "permissions": []any{"go.host"}, "scope": "invocation", "max_concurrent": 1, "max_call_ms": 10000, "max_input_bytes": 1024, "max_output_bytes": 1024, "effects": "mutating", "idempotency": "none"}
	registration := command(map[string]any{"type": "capabilities_register", "descriptors": []any{descriptor}}).(map[string]any)["registration_ids"].([]any)[0]
	id := embeddedSubmit(command, pool("return {call=function(a) return vulcan.capabilities.call('go.callback',a) end}"), map[string]any{"plugin_id": "forged"})
	requests := embeddedPoll(t, func() any { return command(map[string]any{"type": "host_requests_take", "limit": 1}) }, func(value any) bool { return len(value.([]any)) > 0 }).([]any)
	request := requests[0].(map[string]any)
	caller := request["caller"].(map[string]any)
	if request["registration_id"] != registration || caller["plugin_id"] != "go-embedded-test" || caller["operation_id"] != id || request["arguments"].(map[string]any)["plugin_id"] != "forged" {
		t.Fatalf("caller authority mixed with arguments: %#v", request)
	}
	command(map[string]any{"type": "operation_cancel", "operation_id": id})
	if err := transport.Close(); err != nil {
		t.Fatal(err)
	}
	if err := transport.Free(); err == nil {
		t.Fatal("freed root with active callback")
	}
	command(map[string]any{"type": "host_request_complete", "request_id": request["request_id"], "outcome": map[string]any{"ok": true, "value": nil, "effects": "committed"}})
	result := embeddedTerminal(t, command, id)
	if result["phase"] != "cancelled" {
		t.Fatalf("expected real cancellation: %#v", result)
	}
	committed := false
	for _, effect := range result["host_effects"].([]any) {
		if effect.(map[string]any)["effects"] == "committed" {
			committed = true
		}
	}
	if !committed {
		t.Fatalf("late committed effect disappeared: %#v", result)
	}
}
