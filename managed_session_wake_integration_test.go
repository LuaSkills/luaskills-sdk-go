//go:build cgo

package luaskills

import (
	"os"
	"testing"
)

// TestManagedSessionWakeNativeRoundTrip verifies native callback registration and event polling.
// TestManagedSessionWakeNativeRoundTrip 校验原生回调注册与事件轮询。
func TestManagedSessionWakeNativeRoundTrip(t *testing.T) {
	if os.Getenv("LUASKILLS_SDK_NATIVE_INTEGRATION") != "1" {
		t.Skip("set LUASKILLS_SDK_NATIVE_INTEGRATION=1 to run native integration")
	}
	client, err := NewClient(ClientOptions{RuntimeRoot: t.TempDir(), EnsureRuntimeLayout: true})
	if err != nil {
		t.Fatalf("create native client: %v", err)
	}
	defer func() {
		if _, closeErr := client.Close(); closeErr != nil {
			t.Errorf("close native client: %v", closeErr)
		}
	}()
	if err := client.SetManagedSessionWakeCallback(func(_ uint64) error { return nil }); err != nil {
		t.Fatalf("register wake callback: %v", err)
	}
	batch, err := client.PollManagedSessionEvents(1, AuthoritySystem)
	if err != nil {
		t.Fatalf("poll managed-session events: %v", err)
	}
	if _, ok := batch["events"]; !ok {
		t.Fatalf("managed-session event batch is missing events: %#v", batch)
	}
	if err := client.SetManagedSessionWakeCallback(nil); err != nil {
		t.Fatalf("clear wake callback: %v", err)
	}
}
