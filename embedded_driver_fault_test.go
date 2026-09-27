package luaskills

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

// uncertainRecoveryNative simulates delivery followed by uncertain release without allocating real native memory.
// uncertainRecoveryNative 模拟交付后出现不确定释放，不分配真实原生内存。
type uncertainRecoveryNative struct {
	uncertainDriverNative
	// releases selects initial retained failure followed by an abnormal recovery exit.
	// releases 选择首次保留失败，随后恢复异常退出。
	releases atomic.Uint64
}

// request returns only a synthetic descriptor used by the paired Go-only copier.
// request 仅返回配套纯 Go 复制器使用的合成描述符。
func (*uncertainRecoveryNative) request(uint64, []byte) (embeddedResult, EmbeddedNativeStatus) {
	return embeddedResult{id: 1}, EmbeddedNativeOk
}

// copy returns an owned envelope without dereferencing a native address.
// copy 返回拥有的信封，不解引用原生地址。
func (*uncertainRecoveryNative) copy(embeddedResult, uint64) ([]byte, error) {
	return []byte(`{"protocol_version":1,"status":"ok","result":null}`), nil
}

// release first retains the descriptor, then exits abnormally before a recovery result can be proven.
// release 首先保留描述符，然后在能够证明恢复结果前异常退出。
func (n *uncertainRecoveryNative) release(uint64, embeddedResult) EmbeddedNativeStatus {
	if n.releases.Add(1) == 1 {
		return EmbeddedNativeBusy
	}
	if n.mode == "goexit" {
		runtime.Goexit()
	}
	panic("controlled recovery failure")
}

// TestEmbeddedDriverUncertainRecovery isolates deliberately retained ownership after recovery panic or Goexit.
// TestEmbeddedDriverUncertainRecovery 隔离恢复 panic 或 Goexit 后故意保留的所有权。
func TestEmbeddedDriverUncertainRecovery(t *testing.T) {
	if mode := os.Getenv("LUASKILLS_DRIVER_RECOVERY_FAULT_TEST"); mode != "" {
		backend := &uncertainRecoveryNative{uncertainDriverNative: uncertainDriverNative{mode}}
		transport, err := createEmbeddedTransport(EmbeddedTransportConfig{1, 2, 4096, 2048, 2048}, backend)
		if err != nil {
			t.Fatal(err)
		}
		driver, err := NewEmbeddedCommandDriver(transport, EmbeddedDriverConfig{1, 1, 1})
		if err != nil {
			t.Fatal(err)
		}
		receipt, err := driver.Submit(context.Background(), EmbeddedInputCommandDescribe{Type: EmbeddedInputCommandDescribeTypeDescribe})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var release *EmbeddedResultReleaseError
		if _, err := receipt.Result(ctx); !errors.As(err, &release) {
			t.Fatalf("missing initial retained release: %v", err)
		}
		exited := make(chan struct{})
		go func() {
			defer close(exited)
			_ = driver.ReleaseResults(context.Background())
		}()
		select {
		case <-exited:
		case <-ctx.Done():
			t.Fatal("recovery did not exit")
		}
		var failure *EmbeddedDriverFailure
		if err := driver.Close(ctx); !errors.As(err, &failure) {
			t.Fatalf("uncertain recovery falsely closed: %v", err)
		}
		if driver.Status().Closed || transport.Free() == nil || transport.RetainedResults() != 1 || len(LiveEmbeddedCommandDrivers()) != 1 {
			t.Fatal("uncertain allocation owner was released")
		}
		if !receipt.Done() {
			t.Fatal("recovery rewrote proven original delivery")
		}
		if value, err := receipt.DeliveredResult(); value != nil || err != nil {
			t.Fatalf("original delivery lost: %#v %v", value, err)
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"panic", "goexit"} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		command := exec.CommandContext(ctx, executable, "-test.run=^TestEmbeddedDriverUncertainRecovery$")
		command.Env = append(os.Environ(), "LUASKILLS_DRIVER_RECOVERY_FAULT_TEST="+mode)
		output, err := command.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("isolated %s recovery test: %v\n%s", mode, err, output)
		}
	}
}
