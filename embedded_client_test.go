package luaskills

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// typedClientNative is a pure Go response boundary, used only to test observer ownership and malformed deliveries.
// typedClientNative 是纯 Go 响应边界，仅用于测试观察者所有权及畸形交付。
type typedClientNative struct {
	// uncertainDriverNative supplies only synthetic create and close; all execution and release methods are overridden.
	// uncertainDriverNative 仅提供合成创建及关闭；全部执行和释放方法被覆盖。
	uncertainDriverNative
	// response is immutable evidence; entered and allow optionally hold one request beyond observer cancellation.
	// response 是不可变证据；entered 和 allow 可选地使一个请求跨观察取消保持执行。
	response []byte
	entered  chan struct{}
	allow    chan struct{}
	// calls proves repeated typed observations do not issue another native command.
	// calls 证明重复类型化观察不发出另一个原生命令。
	calls atomic.Uint64
}

// request returns an exact synthetic allocation after the optional barrier; frame is intentionally not interpreted.
// request 在可选屏障后返回精确合成分配；刻意不解释 frame。
func (n *typedClientNative) request(_ uint64, _ []byte) (embeddedResult, EmbeddedNativeStatus) {
	// identity uniquely distinguishes successive fixture allocations.
	// identity 唯一区分连续夹具分配。
	identity := n.calls.Add(1)
	if n.entered != nil {
		close(n.entered)
		<-n.allow
	}
	return embeddedResult{id: identity, descriptor: n.response}, EmbeddedNativeOk
}

// copy returns an independent response bounded by maxBytes or a visible fixture error.
// copy 返回受 maxBytes 限制的独立响应，或可见夹具错误。
func (n *typedClientNative) copy(result embeddedResult, maxBytes uint64) ([]byte, error) {
	// response comes exclusively from the original retained allocation descriptor.
	// response 唯一来自原始保留分配描述符。
	response := result.descriptor.([]byte)
	if uint64(len(response)) > maxBytes {
		return nil, errors.New("fixture response exceeds native limit")
	}
	return bytes.Clone(response), nil
}

// release acknowledges the Go-only descriptor without touching native memory.
// release 确认纯 Go 描述符，不接触原生内存。
func (*typedClientNative) release(uint64, embeddedResult) EmbeddedNativeStatus {
	return EmbeddedNativeOk
}

// free acknowledges only the already-drained synthetic transport.
// free 仅确认已经排空的合成传输。
func (*typedClientNative) free(uint64) EmbeddedNativeStatus { return EmbeddedNativeOk }

// typedFixture creates a bounded pure Go client and drains the optional barrier before closing all owners.
// typedFixture 创建有界纯 Go 客户端，并在关闭全部所有者前排空可选屏障。
func typedFixture(t *testing.T, native *typedClientNative) (*EmbeddedClient, func()) {
	t.Helper()
	// transport and driver reserve exactly one work and one control response frame.
	// transport 和 driver 精确预留一个工作响应帧和一个控制响应帧。
	transport, err := createEmbeddedTransport(EmbeddedTransportConfig{1, 2, 4096, 2048, 2048}, native)
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
	// once ensures success and failure cleanup can both unblock an accepted request safely.
	// once 保证成功和失败清理均可安全解除已接纳请求阻塞。
	var once sync.Once
	unblock := func() {
		once.Do(func() {
			if native.allow != nil {
				close(native.allow)
			}
		})
	}
	t.Cleanup(func() {
		unblock()
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
	return client, unblock
}

// TestEmbeddedClientProjectionFailure preserves each original receipt when its declared identity shape is invalid.
// TestEmbeddedClientProjectionFailure 在声明身份形状无效时保留每个原始回执。
func TestEmbeddedClientProjectionFailure(t *testing.T) {
	for _, result := range []string{`null`, `{"runtime_id":false}`, `{"runtime_id":""}`, `{}`} {
		t.Run(result, func(t *testing.T) {
			// native publishes valid success framing with a deliberately unusable reserve result.
			// native 发布有效成功信封，内含刻意不可用的预留结果。
			native := &typedClientNative{response: []byte(`{"protocol_version":1,"status":"ok","result":` + result + `}`)}
			client, _ := typedFixture(t, native)
			pending, err := client.Reserve(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if _, err := pending.Result(ctx); err == nil {
				t.Fatal("malformed identity accepted")
			}
			if _, err := pending.DeliveredResult(); err == nil {
				t.Fatal("delivered projection weakened the schema")
			}
			if len(client.Driver().Commands()) != 1 || native.calls.Load() != 1 {
				t.Fatal("failed projection lost or replayed its receipt")
			}
			if err := pending.Forget(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestEmbeddedClientInterruptedReceipt keeps a reserved identity recoverable after its observer exits.
// TestEmbeddedClientInterruptedReceipt 在观察者退出后保持预留身份可恢复。
func TestEmbeddedClientInterruptedReceipt(t *testing.T) {
	// entered proves the native adapter is active before cancellation is requested.
	// entered 在请求取消前证明原生适配器处于活动状态。
	native := &typedClientNative{response: []byte(`{"protocol_version":1,"status":"ok","result":{"runtime_id":"exact-slot"}}`), entered: make(chan struct{}), allow: make(chan struct{})}
	client, unblock := typedFixture(t, native)
	pending, err := client.Reserve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-native.entered:
	case <-time.After(time.Second):
		t.Fatal("native adapter did not enter")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := pending.Result(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("observer cancellation changed: %v", err)
	}
	if _, err := pending.DeliveredResult(); err == nil || pending.Receipt().Done() {
		t.Fatal("unreturned request fabricated delivery")
	}
	if pending.Forget() == nil {
		t.Fatal("active request returned SDK quota")
	}
	unblock()
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	first, err := pending.Result(ctx)
	if err != nil {
		t.Fatal(err)
	}
	second, err := pending.DeliveredResult()
	if err != nil || first.RuntimeID() != "exact-slot" || second.RuntimeID() != first.RuntimeID() || native.calls.Load() != 1 {
		t.Fatal("identity was lost or replayed")
	}
	if err := pending.Forget(); err != nil {
		t.Fatal(err)
	}
}

// TestEmbeddedClientInterruptedPoll retains the in-flight status receipt and never converts observer cancellation into native cancel.
// TestEmbeddedClientInterruptedPoll 保留在途状态回执，绝不将观察取消变成原生取消。
func TestEmbeddedClientInterruptedPoll(t *testing.T) {
	// native returns a valid nonterminal snapshot only after the observer has already left.
	// native 仅在观察者已经离开后返回有效非终态快照。
	native := &typedClientNative{response: []byte(`{"protocol_version":1,"status":"ok","result":{"operation_id":"op-exact","phase":"running","effects":"unknown","cancellation_requested":false,"context":{"kind":"unbound"},"host_effects":[]}}`), entered: make(chan struct{}), allow: make(chan struct{})}
	client, unblock := typedFixture(t, native)
	runtime, err := client.Runtime("slot-exact")
	if err != nil {
		t.Fatal(err)
	}
	operation, err := runtime.Operation("op-exact")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan error, 1)
	go func() { _, err := operation.Wait(ctx); finished <- err }()
	select {
	case <-native.entered:
	case <-time.After(time.Second):
		t.Fatal("status request did not enter")
	}
	cancel()
	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled observer kept waiting")
	}
	commands := client.Driver().Commands()
	if len(commands) != 1 || commands[0].Done() || native.calls.Load() != 1 {
		t.Fatal("interrupted poll lost its owner")
	}
	unblock()
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := commands[0].Result(ctx); err != nil {
		t.Fatal(err)
	}
	if err := commands[0].Forget(); err != nil {
		t.Fatal(err)
	}
	if native.calls.Load() != 1 {
		t.Fatal("poll sent a cancellation command")
	}
}

// TestEmbeddedClientInputBoundaries rejects nil owners, empty identities and invalid wait contexts before any request.
// TestEmbeddedClientInputBoundaries 在任何请求前拒绝空所有者、空身份及无效等待上下文。
func TestEmbeddedClientInputBoundaries(t *testing.T) {
	if _, err := NewEmbeddedClient(nil); err == nil {
		t.Fatal("nil driver accepted")
	}
	// native should remain unused for every rejected boundary below.
	// 对下列每个被拒绝边界，native 都应保持未使用。
	native := &typedClientNative{}
	client, _ := typedFixture(t, native)
	if _, err := client.Runtime(""); err == nil {
		t.Fatal("empty runtime accepted")
	}
	runtime, err := client.Runtime("known-slot")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Plugin(""); err == nil {
		t.Fatal("empty plugin accepted")
	}
	if _, err := runtime.Pool(""); err == nil {
		t.Fatal("empty pool accepted")
	}
	if _, err := runtime.Session(""); err == nil {
		t.Fatal("empty session accepted")
	}
	if _, err := runtime.Operation(""); err == nil {
		t.Fatal("empty operation accepted")
	}
	operation, err := runtime.Operation("known-op")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Wait(nil); err == nil {
		t.Fatal("nil observer accepted")
	}
	for _, interval := range []time.Duration{0, -time.Second} {
		if _, err := operation.WaitInterval(context.Background(), interval); err == nil {
			t.Fatal("invalid interval accepted")
		}
	}
	if native.calls.Load() != 0 || len(client.Driver().Commands()) != 0 {
		t.Fatal("invalid input reached native admission")
	}
}
