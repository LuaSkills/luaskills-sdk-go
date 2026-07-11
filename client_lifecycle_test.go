package luaskills

import (
	"strings"
	"testing"
	"time"
)

// TestClientEngineIDReturnsImmutableHandle verifies the read-only engine identifier accessor.
// TestClientEngineIDReturnsImmutableHandle 校验只读引擎标识访问器。
func TestClientEngineIDReturnsImmutableHandle(t *testing.T) {
	client := &Client{engineID: 42}

	if client.EngineID() != 42 {
		t.Fatalf("unexpected engine id: %d", client.EngineID())
	}
}

// TestClientCallRejectsClosedEngineBeforeDispatch verifies closed clients never reach FFI dispatch.
// TestClientCallRejectsClosedEngineBeforeDispatch 校验已关闭客户端不会进入 FFI 分发。
func TestClientCallRejectsClosedEngineBeforeDispatch(t *testing.T) {
	client := &Client{engineID: 42, closed: true}
	var result map[string]any

	err := client.call("luaskills_ffi_unsupported_test_json", map[string]any{}, &result)
	if err == nil {
		t.Fatal("expected closed engine error")
	}
	if !strings.Contains(err.Error(), "LuaSkills engine 42 is already closed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestClientCallRejectsClosingEngineBeforeDispatch verifies close-in-progress blocks new calls.
// TestClientCallRejectsClosingEngineBeforeDispatch 校验关闭中的客户端会阻止新调用。
func TestClientCallRejectsClosingEngineBeforeDispatch(t *testing.T) {
	client := &Client{engineID: 42, closing: true}
	var result map[string]any

	err := client.call("luaskills_ffi_unsupported_test_json", map[string]any{}, &result)
	if err == nil {
		t.Fatal("expected closing engine error")
	}
	if !strings.Contains(err.Error(), "LuaSkills engine 42 is closing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestClientBeginCloseWaitsForActiveCalls verifies close waits for active FFI reservations.
// TestClientBeginCloseWaitsForActiveCalls 校验关闭流程会等待活跃 FFI 占用释放。
func TestClientBeginCloseWaitsForActiveCalls(t *testing.T) {
	client := &Client{engineID: 42}
	if err := client.beginCall(); err != nil {
		t.Fatalf("beginCall failed: %v", err)
	}

	released := make(chan uint64, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		engineID, shouldClose := client.beginClose()
		if !shouldClose {
			return
		}
		released <- engineID
		client.finishClose(nil)
	}()

	select {
	case engineID := <-released:
		t.Fatalf("close did not wait for active call, got engine id %d", engineID)
	case <-time.After(10 * time.Millisecond):
	}

	client.endCall()

	select {
	case engineID := <-released:
		if engineID != 42 {
			t.Fatalf("unexpected engine id: %d", engineID)
		}
	case <-time.After(time.Second):
		t.Fatal("close did not continue after active call ended")
	}
	<-done
}
