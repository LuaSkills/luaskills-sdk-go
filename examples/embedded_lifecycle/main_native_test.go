//go:build cgo

package main

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	luaskills "github.com/LuaSkills/luaskills-sdk-go"
)

// TestEmbeddedReserveCancelledObserver preserves the exact admitted native slot through cancellation and complete cleanup.
// TestEmbeddedReserveCancelledObserver 跨观察取消及完整清理保留精确已入场原生槽。
// t runs only with the explicitly selected native candidate; no reservation is replayed during recovery.
// t 仅在显式选定原生候选下执行；恢复期间不重放预留。
func TestEmbeddedReserveCancelledObserver(t *testing.T) {
	if os.Getenv("LUASKILLS_NATIVE_E2E") != "1" {
		t.Skip("LUASKILLS_NATIVE_E2E is not enabled")
	}
	// owned retains the same owners used by the executable example.
	// owned 保留与可执行示例相同的所有者。
	owned := &owners{}
	// cleanup bounds only observation, never accepted native work.
	// cleanup 仅限制观察，绝不限制已接纳原生工作。
	cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	// err is the actual construction or submission result.
	// err 是实际构造或提交结果。
	var err error
	owned.transport, err = luaskills.NewEmbeddedTransport(luaskills.EmbeddedTransportConfig{MaxRuntimes: 1, MaxResultBuffers: 3, MaxResultBytes: 98304, MaxResponseBytes: 32768, MaxRequestBytes: 65536})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if owned.transport.TransportID() != 0 {
			// recovery is independent from the test observer's deferred cancellation.
			// recovery 独立于测试观察者的延迟取消。
			recovery, finish := context.WithTimeout(context.Background(), 5*time.Second)
			defer finish()
			if err := owned.close(recovery); err != nil {
				t.Error(err)
			}
		}
	})
	owned.driver, err = luaskills.NewEmbeddedCommandDriver(owned.transport, luaskills.EmbeddedDriverConfig{WorkWorkers: 1, MaxWorkCommands: 2, MaxControlCommands: 2})
	if err != nil {
		t.Fatal(err)
	}
	// client submits exactly one real reservation before the observer is cancelled.
	// client 在观察者取消前精确提交一次真实预留。
	client, err := luaskills.NewEmbeddedClient(owned.driver)
	if err != nil {
		t.Fatal(err)
	}
	owned.reserved, err = client.Reserve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// observer cancellation follows successful admission and fails even if native delivery has already completed.
	// observer 的取消发生于入场成功后，即使原生交付已完成也会失败。
	observer, cancel := context.WithCancel(context.Background())
	cancel()
	owned.runtime, err = take(observer, owned.reserved, nil)
	if !errors.Is(err, context.Canceled) || owned.runtime != nil {
		t.Fatalf("cancelled observation did not preserve the original receipt: %v", err)
	}
	// original reads the retained receipt without transferring ownership into the example's runtime field.
	// original 读取保留回执，不将所有权转移到示例的 runtime 字段。
	original, err := owned.reserved.Result(cleanup)
	if err != nil {
		t.Fatal(err)
	}
	if len(owned.driver.Commands()) != 1 {
		t.Fatal("fixture must contain the one original reservation receipt")
	}
	if err := owned.close(cleanup); err != nil {
		t.Fatal(err)
	}
	if owned.runtime.RuntimeID() != original.RuntimeID() || owned.scope.Status().Phase != luaskills.EmbeddedScopeClosed {
		t.Fatal("cleanup replaced the admitted runtime or failed to close its exact scope")
	}
	// Successful native transport Free proves its runtime registration table is empty, not merely that a waiter exited.
	// 原生传输 Free 成功证明其运行时注册表为空，而非仅有等待者退出。
	if owned.transport.TransportID() != 0 || len(owned.driver.Commands()) != 0 || !owned.driver.Status().Closed {
		t.Fatal("cleanup retained the native root, reservation receipt or driver workers")
	}
}
