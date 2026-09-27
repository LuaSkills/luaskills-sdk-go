package luaskills

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// uncertainPumpNative simulates lost mutation delivery with Go-owned buffers only; no real native runtime is created.
// uncertainPumpNative 仅以 Go 拥有的缓冲模拟变更交付丢失；不创建真实原生运行时。
type uncertainPumpNative struct {
	uncertainDriverNative
	// route and armed select one mutation after any prerequisite registration has completed.
	// route 及 armed 在所需前置注册完成后选择一个变更。
	route string
	armed atomic.Bool
	lost  atomic.Bool
	calls atomic.Uint64
	// next and responses own synthetic descriptors and their immutable response bytes.
	// next 及 responses 拥有合成描述符及其不可变响应字节。
	next      atomic.Uint64
	responses sync.Map
	selected  sync.Map
}

// request derives the exact fixture reply from the declared route and counts attempted selected mutations.
// request 按声明路由派生精确夹具响应，并统计选中变更的尝试。
func (n *uncertainPumpNative) request(_ uint64, frame []byte) (embeddedResult, EmbeddedNativeStatus) {
	decoded, err := DecodeEmbeddedJSON(frame)
	if err != nil {
		panic(err)
	}
	command := decoded.(map[string]any)["command"].(map[string]any)
	route := command["type"].(string)
	if route == "runtime" {
		route = command["operation"].(map[string]any)["type"].(string)
	}
	var result any
	switch route {
	case "runtime_status":
		result = EmbeddedOutputRuntimeSnapshot{Initialization: EmbeddedOutputInitializationPhaseReady, RuntimeId: "runtime"}
	case "capabilities_register":
		result = map[string]any{"registration_ids": []string{"registration"}}
	case "host_requests_take":
		result = []any{}
	case "capability_status":
		result = EmbeddedOutputCapabilityRegistrationStatus{RegistrationId: "registration", Name: "go.callback", Drained: true}
	case "capability_unregister", "capability_forget":
		result = nil
	default:
		panic("unexpected uncertain-pump fixture route")
	}
	response, err := EncodeEmbeddedJSON(map[string]any{"protocol_version": EmbeddedProtocolVersion, "status": "ok", "result": result}, 32768)
	if err != nil {
		panic(err)
	}
	id := n.next.Add(1)
	n.responses.Store(id, response)
	if route == n.route {
		n.calls.Add(1)
		if n.armed.Load() {
			n.selected.Store(id, true)
		}
	}
	return embeddedResult{id: id}, EmbeddedNativeOk
}

// copy loses one synthetic acknowledgement after its mutation was attempted, leaving no evidence that authorizes replay.
// copy 在变更已经尝试后丢失一个合成确认，不保留能够授权重放的证据。
func (n *uncertainPumpNative) copy(result embeddedResult, _ uint64) ([]byte, error) {
	if _, selected := n.selected.Load(result.id); selected && n.lost.CompareAndSwap(false, true) {
		return nil, errors.New("controlled missing mutation receipt")
	}
	response, ok := n.responses.Load(result.id)
	if !ok {
		panic("missing synthetic allocation")
	}
	return append([]byte{}, response.([]byte)...), nil
}

// release removes synthetic Go-only buffers; it never touches a native address or guesses the mutation outcome.
// release 移除纯 Go 合成缓冲；绝不接触原生地址或猜测变更结果。
func (n *uncertainPumpNative) release(_ uint64, result embeddedResult) EmbeddedNativeStatus {
	n.responses.Delete(result.id)
	n.selected.Delete(result.id)
	return EmbeddedNativeOk
}

// TestEmbeddedPumpUncertainMutation isolates intentionally unresolved publication, extraction and retirement ownership.
// TestEmbeddedPumpUncertainMutation 隔离故意未解决的发布、领取及退役所有权。
func TestEmbeddedPumpUncertainMutation(t *testing.T) {
	if route := os.Getenv("LUASKILLS_PUMP_UNCERTAIN_TEST"); route != "" {
		backend := &uncertainPumpNative{route: route}
		transport, err := createEmbeddedTransport(EmbeddedTransportConfig{1, 4, 131072, 32768, 65536}, backend)
		if err != nil {
			t.Fatal(err)
		}
		pump, err := NewEmbeddedCallbackPump(transport, "runtime", EmbeddedCallbackPumpConfig{1, 1, 1})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := pump.Ready(ctx); err != nil {
			t.Fatal(err)
		}
		backend.armed.Store(route == "capabilities_register")
		capability := EmbeddedHostCapability{Descriptor: EmbeddedInputCapabilityDescriptor{Name: "go.callback", Version: "1.0.0", Description: "Pure Go ownership fixture", InputSchema: true, OutputSchema: true, Execution: "queued", Permissions: EmbeddedInputCapabilityDescriptorPermissions{}, Scope: "invocation", MaxConcurrent: 1, MaxCallMs: 1, MaxInputBytes: 1, MaxOutputBytes: 1, Effects: "mutating", Idempotency: "none"}, Handler: func(any, *EmbeddedHostCallbackContext) (any, error) { panic("unexpected fixture handler") }}
		identities, registerError := pump.Register(ctx, []EmbeddedHostCapability{capability})
		if route == "capabilities_register" {
			if registerError == nil {
				t.Fatal("lost publication reported success")
			}
		} else {
			if registerError != nil || len(identities) != 1 {
				t.Fatalf("fixture registration failed: %#v %v", identities, registerError)
			}
			backend.armed.Store(true)
			if route != "host_requests_take" {
				go func() { _ = pump.Unregister(ctx, identities[0]) }()
			}
		}
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for !pump.Status().RecoveryRequired {
			select {
			case <-ctx.Done():
				t.Fatal("uncertain mutation was not retained")
			case <-ticker.C:
			}
		}
		attempts := backend.calls.Load()
		if err := pump.RetryAcknowledgements(ctx); err == nil {
			t.Fatal("missing original delivery was treated as recovered")
		}
		if backend.calls.Load() != attempts || transport.Free() == nil || pump.Status().Closed || len(LiveEmbeddedCallbackPumps()) != 1 {
			t.Fatal("uncertain mutation was replayed or its owner was released")
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{"capabilities_register", "host_requests_take", "capability_unregister", "capability_forget"} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		command := exec.CommandContext(ctx, executable, "-test.run=^TestEmbeddedPumpUncertainMutation$")
		command.Env = append(os.Environ(), "LUASKILLS_PUMP_UNCERTAIN_TEST="+route)
		output, err := command.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("isolated %s uncertainty test: %v\n%s", route, err, output)
		}
	}
}
