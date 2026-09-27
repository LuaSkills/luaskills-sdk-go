package luaskills

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// scopeTestNative models typed lifecycle replies without a DLL; fault paths retain uncertain owners in isolated processes.
// scopeTestNative 不使用 DLL 地模拟类型化生命周期响应；故障路径在隔离进程中保留不确定所有者。
type scopeTestNative struct {
	// uncertainDriverNative supplies synthetic creation and admission closure only.
	// uncertainDriverNative 仅提供合成创建及入场关闭。
	uncertainDriverNative
	// route and mode select an exact root-control boundary; normal responses echo the declared slot identity.
	// route 和 mode 选择精确根控制边界；正常响应回传已声明槽身份。
	route, mode string
	// next and calls account for actual synthetic allocations and selected native entries.
	// next 和 calls 计费实际合成分配及所选原生入口。
	next, calls atomic.Uint64
	fired       atomic.Bool
	// responses and selected preserve exact per-allocation ownership through concurrent driver activity.
	// responses 和 selected 跨并发驱动器活动保留精确分配所有权。
	responses, selected sync.Map
}

// request publishes generated root responses or the selected controlled failure; it never executes Lua.
// request 发布生成根响应或所选受控失败；绝不执行 Lua。
func (n *scopeTestNative) request(_ uint64, frame []byte) (embeddedResult, EmbeddedNativeStatus) {
	// command is an already validated root union from the owned request frame.
	// command 是拥有型请求帧中已经校验的根联合。
	decoded, err := DecodeEmbeddedJSON(frame)
	if err != nil {
		panic(err)
	}
	command := decoded.(map[string]any)["command"].(map[string]any)
	route := command["type"].(string)
	matched := route == n.route
	if matched {
		n.calls.Add(1)
		if n.mode == "panic" {
			panic("controlled scope adapter panic")
		}
		if n.mode == "goexit" {
			runtime.Goexit()
		}
	}
	identity := command["runtime_id"].(string)
	if matched && n.mode == "identity" {
		identity = "foreign-slot"
	}
	var result any
	switch route {
	case "runtime_close", "runtime_free":
		result = EmbeddedOutputRuntimeReceipt{RuntimeId: identity}
	case "runtime_status":
		result = EmbeddedOutputRuntimeSnapshot{RuntimeId: identity, Initialization: EmbeddedOutputInitializationPhaseReady, Closing: true, Closed: true}
		if matched && n.mode == "faulted" {
			result = EmbeddedOutputRuntimeSnapshot{RuntimeId: identity, Initialization: EmbeddedOutputInitializationPhaseFaulted, Closing: true}
		}
	default:
		panic("unexpected synthetic scope command")
	}
	envelope := map[string]any{"protocol_version": EmbeddedProtocolVersion, "status": "ok", "result": result}
	if matched && n.mode == "busy" && n.fired.CompareAndSwap(false, true) {
		envelope = map[string]any{"protocol_version": EmbeddedProtocolVersion, "status": "error", "error": EmbeddedOutputEmbeddedError{Code: EmbeddedOutputEmbeddedErrorCodeBusy, Message: "retained native lease"}}
	}
	response, err := EncodeEmbeddedJSON(envelope, 2048)
	if err != nil {
		panic(err)
	}
	id := n.next.Add(1)
	n.responses.Store(id, response)
	if matched {
		n.selected.Store(id, true)
	}
	return embeddedResult{id: id}, EmbeddedNativeOk
}

// copy preserves ordinary delivery and injects only a selected missing or malformed response.
// copy 保留普通交付，仅注入所选缺失或畸形响应。
func (n *scopeTestNative) copy(result embeddedResult, _ uint64) ([]byte, error) {
	if _, selected := n.selected.Load(result.id); selected {
		if n.mode == "lost" {
			return nil, errors.New("controlled missing scope receipt")
		}
		if n.mode == "malformed" {
			return []byte(`{"protocol_version":1,"status":"ok","result":{}}`), nil
		}
	}
	// response must be the exact still-retained synthetic allocation.
	// response 必须是精确且仍保留的合成分配。
	response, ok := n.responses.Load(result.id)
	if !ok {
		panic("missing scope fixture allocation")
	}
	return bytes.Clone(response.([]byte)), nil
}

// release retires one exact synthetic allocation without invoking real native memory management.
// release 退役一个精确合成分配，不调用实际原生内存管理。
func (n *scopeTestNative) release(_ uint64, result embeddedResult) EmbeddedNativeStatus {
	n.responses.Delete(result.id)
	n.selected.Delete(result.id)
	return EmbeddedNativeOk
}

// free acknowledges only a drained synthetic root; SDK claims prevent this entry while a scope remains alive.
// free 仅确认已排空合成根；作用域仍存活时 SDK 声明会阻止进入此入口。
func (*scopeTestNative) free(uint64) EmbeddedNativeStatus { return EmbeddedNativeOk }

// scopeFixture creates an ordinary borrowed driver with explicit budgets and registers ownership-aware cleanup.
// scopeFixture 以显式预算创建普通借用驱动器，并注册感知所有权的清理。
func scopeFixture(t *testing.T, config EmbeddedTransportConfig, native *scopeTestNative) (*EmbeddedRuntime, *EmbeddedTransport) {
	t.Helper()
	// transport owns synthetic result frames and the driver owns its separate receipt quota.
	// transport 拥有合成结果帧，driver 拥有其独立回执配额。
	transport, err := createEmbeddedTransport(config, native)
	if err != nil {
		t.Fatal(err)
	}
	driver, err := NewEmbeddedCommandDriver(transport, EmbeddedDriverConfig{1, 1, 1})
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewEmbeddedClient(driver)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := client.Runtime("slot")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := driver.Close(ctx); err != nil {
			t.Error(err)
			return
		}
		for _, command := range driver.Commands() {
			if err := command.Forget(); err != nil {
				t.Error(err)
			}
		}
		if err := transport.Close(); err != nil {
			t.Error(err)
		}
		if err := transport.Free(); err != nil {
			t.Error(err)
		}
	})
	return handle, transport
}

// TestEmbeddedScopeOwnership checks both count/byte reservations, frame validation and exact runtime ownership.
// TestEmbeddedScopeOwnership 检查数量／字节预留、帧校验及精确运行时所有权。
func TestEmbeddedScopeOwnership(t *testing.T) {
	for _, config := range []EmbeddedTransportConfig{{1, 2, 4096, 2048, 2048}, {1, 8, 4096, 2048, 2048}} {
		runtime, transport := scopeFixture(t, config, &scopeTestNative{})
		if _, err := NewEmbeddedRuntimeScope(runtime, nil, time.Millisecond); err == nil {
			t.Fatal("scope exceeded aggregate worker frame budget")
		}
		if len(transport.runtimeScopes) != 0 {
			t.Fatal("failed reservation leaked a scope claim")
		}
	}
	// owner remains valid after invalid and duplicate construction attempts.
	// owner 在无效及重复构造尝试后保持有效。
	owner, transport := scopeFixture(t, EmbeddedTransportConfig{1, 4, 8192, 2048, 2048}, &scopeTestNative{})
	if _, err := NewEmbeddedRuntimeScope(nil, nil, time.Millisecond); err == nil {
		t.Fatal("nil runtime accepted")
	}
	if _, err := NewEmbeddedRuntimeScope(owner, nil, 0); err == nil {
		t.Fatal("zero interval accepted")
	}
	tooLong, err := owner.client.Runtime(strings.Repeat("x", 2048))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewEmbeddedRuntimeScope(tooLong, nil, time.Millisecond); err == nil {
		t.Fatal("scope adopted an unencodable cleanup identity")
	}
	scope, err := NewEmbeddedRuntimeScope(owner, nil, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	t.Cleanup(func() {
		if err := scope.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	other, err := owner.client.Runtime("other-slot")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewEmbeddedRuntimeScope(other, nil, time.Millisecond); err == nil {
		t.Fatal("scope count exceeded MaxRuntimes")
	}
	if _, err := owner.Free(ctx); err == nil {
		t.Fatal("typed free bypassed scope ownership")
	}
	if _, err := owner.client.driver.Submit(ctx, EmbeddedInputCommandRuntimeFree{Type: EmbeddedInputCommandRuntimeFreeTypeRuntimeFree, RuntimeId: owner.RuntimeID()}); err == nil {
		t.Fatal("direct managed command bypassed scope ownership")
	}
	if transport.Free() == nil {
		t.Fatal("transport freed scoped lifetime")
	}
	if err := scope.Close(ctx); err != nil {
		t.Fatal(err)
	}
}

// TestEmbeddedScopeQueuedFree blocks adoption before a queued native removal, even with no active transport readers.
// TestEmbeddedScopeQueuedFree 即使传输没有活动读取者，也在排队原生移除前阻止接管。
func TestEmbeddedScopeQueuedFree(t *testing.T) {
	// maintenance gives a deterministic queued-but-not-entered boundary without a native polling race.
	// maintenance 提供确定性的已排队但尚未进入边界，不依赖原生轮询竞态。
	owner, transport := scopeFixture(t, EmbeddedTransportConfig{1, 3, 6144, 2048, 2048}, &scopeTestNative{})
	driver := owner.client.driver
	driver.mu.Lock()
	driver.maintenance = true
	driver.mu.Unlock()
	unblock := func() { driver.mu.Lock(); driver.maintenance = false; driver.changed.Broadcast(); driver.mu.Unlock() }
	defer unblock()
	pending, err := owner.Free(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if pending.Receipt().State() != EmbeddedCommandQueued {
		t.Fatal("fixture failed to retain queued removal")
	}
	if _, err := NewEmbeddedRuntimeScope(owner, nil, time.Millisecond); err == nil {
		t.Fatal("scope raced past queued slot removal")
	}
	if len(transport.runtimeScopes) != 0 {
		t.Fatal("rejected adoption published an owner")
	}
	unblock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := pending.Result(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pending.Forget(); err != nil {
		t.Fatal(err)
	}
}

// TestEmbeddedScopeMixedReservations checks aggregate driver, scope and pump ownership in both construction orders.
// TestEmbeddedScopeMixedReservations 在两种构造顺序中检查驱动器、作用域和泵的累计所有权。
func TestEmbeddedScopeMixedReservations(t *testing.T) {
	for _, pumpFirst := range []bool{false, true} {
		// The manually claimed pump models only frame ownership; no coordinator or callback is started by this fixture.
		// 手工声明的泵仅模拟帧所有权；此夹具不启动协调器或回调。
		owner, transport := scopeFixture(t, EmbeddedTransportConfig{2, 4, 8192, 2048, 2048}, &scopeTestNative{})
		pump := &EmbeddedCallbackPump{transport: transport, runtimeID: "other-slot"}
		if pumpFirst {
			if err := transport.claimCallbackPump(pump); err != nil {
				t.Fatal(err)
			}
		}
		scope, err := NewEmbeddedRuntimeScope(owner, nil, time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		if !pumpFirst {
			if err := transport.claimCallbackPump(pump); err != nil {
				t.Fatal(err)
			}
		}
		other, err := owner.client.Runtime("other-slot")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := NewEmbeddedRuntimeScope(other, nil, time.Millisecond); err == nil {
			t.Fatal("scope ignored the existing pump owner")
		}
		if _, err := NewEmbeddedRuntimeScope(other, pump, time.Millisecond); err == nil {
			t.Fatal("scope exceeded aggregate mixed-owner frames")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := owner.client.driver.Close(ctx); err != nil {
			cancel()
			t.Fatal(err)
		}
		// Recreating a larger ordinary driver must account for claims that outlive the old driver.
		// 重建更大的普通驱动器必须计入寿命跨越旧驱动器的声明。
		if _, err := NewEmbeddedCommandDriver(transport, EmbeddedDriverConfig{2, 2, 1}); err == nil {
			cancel()
			t.Fatal("replacement driver ignored scope and pump reservations")
		}
		if err := transport.releaseCallbackPump(pump); err != nil {
			cancel()
			t.Fatal(err)
		}
		if err := scope.Close(ctx); err != nil {
			cancel()
			t.Fatal(err)
		}
		cancel()
	}
}

// TestEmbeddedScopeBusyEvidence verifies only an explicit core Busy permits automatic slot-removal retry.
// TestEmbeddedScopeBusyEvidence 验证仅显式核心 Busy 允许自动槽移除重试。
func TestEmbeddedScopeBusyEvidence(t *testing.T) {
	// native models a still-retained native lease followed by a successful original-slot removal.
	// native 模拟仍保留的原生租约，随后成功移除原始槽。
	native := &scopeTestNative{route: "runtime_free", mode: "busy"}
	owner, _ := scopeFixture(t, EmbeddedTransportConfig{1, 3, 6144, 2048, 2048}, native)
	scope, err := NewEmbeddedRuntimeScope(owner, nil, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := scope.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if native.calls.Load() != 2 {
		t.Fatalf("expected one refusal and one removal, got %d", native.calls.Load())
	}
}

// TestEmbeddedScopeExitedPumpFailure reports fatal coordinator exit before offering acknowledgement recovery.
// TestEmbeddedScopeExitedPumpFailure 在提供确认恢复前报告协调器致命退出。
func TestEmbeddedScopeExitedPumpFailure(t *testing.T) {
	// ended represents actual coordinator exit while an earlier mutation remains unresolved.
	// ended 表示实际协调器退出，同时较早变更仍未解决。
	ended := make(chan struct{})
	close(ended)
	pump := &EmbeddedCallbackPump{ended: ended, fatal: &EmbeddedDriverFailure{Message: "coordinator exited"}, pending: &embeddedPumpMutation{err: errors.New("earlier uncertain delivery")}}
	scope := &EmbeddedRuntimeScope{pump: pump, interval: time.Millisecond}
	err := scope.drainCallbacks(false)
	var fatal *EmbeddedDriverFailure
	if !errors.As(err, &fatal) {
		t.Fatalf("fatal pump was offered ordinary recovery: %v", err)
	}
}

// TestEmbeddedScopeUncertainControl isolates irrecoverable ownership and proves no replay after lost, invalid or interrupted delivery.
// TestEmbeddedScopeUncertainControl 隔离不可恢复所有权，证明丢失、无效或中断交付后不重放。
func TestEmbeddedScopeUncertainControl(t *testing.T) {
	if selection := os.Getenv("LUASKILLS_SCOPE_UNCERTAIN_TEST"); selection != "" {
		// parts is an explicit test route/mode pair; the subprocess intentionally exits with a retained scope.
		// parts 是显式测试路由／模式对；子进程刻意保留作用域后退出。
		parts := strings.SplitN(selection, "/", 2)
		native := &scopeTestNative{route: parts[0], mode: parts[1]}
		transport, err := createEmbeddedTransport(EmbeddedTransportConfig{1, 3, 6144, 2048, 2048}, native)
		if err != nil {
			t.Fatal(err)
		}
		driver, err := NewEmbeddedCommandDriver(transport, EmbeddedDriverConfig{1, 1, 1})
		if err != nil {
			t.Fatal(err)
		}
		client, err := NewEmbeddedClient(driver)
		if err != nil {
			t.Fatal(err)
		}
		handle, err := client.Runtime("slot")
		if err != nil {
			t.Fatal(err)
		}
		scope, err := NewEmbeddedRuntimeScope(handle, nil, time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := scope.Close(ctx); err == nil {
			t.Fatal("uncertain control fabricated successful shutdown")
		}
		if scope.Status().Retryable || scope.Status().Phase == EmbeddedScopeClosed || scope.Status().PendingCommand != parts[0] {
			t.Fatalf("uncertain control lost checkpoint: %+v", scope.Status())
		}
		attempts := native.calls.Load()
		if err := scope.Close(ctx); err == nil {
			t.Fatal("repeated observation fabricated completion")
		}
		if err := scope.RetryClose(ctx); err == nil {
			t.Fatal("unknown delivery authorized another mutation")
		}
		if native.calls.Load() != attempts || len(LiveEmbeddedRuntimeScopes()) != 1 || transport.Free() == nil {
			t.Fatal("unknown control was replayed or lost its native owner")
		}
		if err := driver.Close(ctx); err != nil {
			t.Fatal(err)
		}
		return
	}
	// executable runs each permanent fault in its own process so intentional retained owners cannot contaminate later tests.
	// executable 在独立进程运行每个永久故障，避免刻意保留所有者污染后续测试。
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{"runtime_close", "runtime_status", "runtime_free"} {
		// Faulted construction is a status-only uncertainty and must not be mistaken for ordinary failed initialization.
		// 构造故障仅属于状态不确定性，不得与普通初始化失败混淆。
		modes := []string{"lost", "malformed", "identity", "panic", "goexit"}
		if route == "runtime_status" {
			modes = append(modes, "faulted")
		}
		for _, mode := range modes {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			command := exec.CommandContext(ctx, executable, "-test.run=^TestEmbeddedScopeUncertainControl$")
			command.Env = append(os.Environ(), "LUASKILLS_SCOPE_UNCERTAIN_TEST="+route+"/"+mode)
			output, err := command.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("isolated %s/%s: %v\n%s", route, mode, err, output)
			}
		}
	}
}
