package luaskills

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

// TestEmbeddedDriverRouteCoverage requires an explicit lane decision for every generated route and rejects blocking waits.
// TestEmbeddedDriverRouteCoverage 要求为每个生成路由明确选择通道，并拒绝阻塞等待。
func TestEmbeddedDriverRouteCoverage(t *testing.T) {
	for _, name := range EmbeddedRootCommands() {
		if name == "runtime" {
			continue
		}
		if _, err := embeddedCommandLane(map[string]any{"type": name}); err != nil {
			t.Fatalf("root route requires an explicit driver decision: %s: %v", name, err)
		}
	}
	for _, name := range EmbeddedRuntimeCommands() {
		_, err := embeddedCommandLane(map[string]any{"type": "runtime", "operation": map[string]any{"type": name}})
		if name == "operation_wait" {
			var failure *EmbeddedRuntimeError
			if !errors.As(err, &failure) || failure.Code != "unsupported" {
				t.Fatal("blocking wait must be unsupported")
			}
		} else if err != nil {
			t.Fatalf("runtime route requires an explicit driver decision: %s: %v", name, err)
		}
	}
	if _, err := embeddedCommandLane(map[string]any{"type": "unknown"}); err == nil {
		t.Fatal("unknown route accepted")
	}
}

// TestEmbeddedDriverInvalidConfig rejects invalid worker budgets before touching a native transport.
// TestEmbeddedDriverInvalidConfig 在接触原生传输前拒绝无效工作位预算。
func TestEmbeddedDriverInvalidConfig(t *testing.T) {
	for _, config := range []EmbeddedDriverConfig{{}, {0, 1, 1}, {1, 0, 1}, {1, 1, 0}, {2, 1, 1}, {^uint64(0), ^uint64(0), 1}} {
		if _, err := NewEmbeddedCommandDriver(nil, config); err == nil {
			t.Fatal("invalid driver configuration accepted")
		}
	}
	if err := waitEmbeddedObservation(nil, make(chan struct{})); err == nil {
		t.Fatal("nil observer context accepted")
	}
}

// uncertainDriverNative is a Go-only fault boundary; it never allocates or invokes an actual native runtime.
// uncertainDriverNative 是纯 Go 故障边界；绝不分配或调用真实原生运行时。
type uncertainDriverNative struct {
	// mode selects a worker panic or silent goroutine exit for ownership verification.
	// mode 选择工作位 panic 或静默协程退出，以验证所有权。
	mode string
}

// create returns a synthetic fixture identity without allocating native memory.
// create 返回合成夹具身份，不分配原生内存。
func (uncertainDriverNative) create(EmbeddedTransportConfig) (uint64, EmbeddedNativeStatus) {
	return 1, EmbeddedNativeOk
}

// request deliberately exits before delivery; neither path proves a native result.
// request 故意在交付前退出；任何路径都不证明原生结果。
func (n uncertainDriverNative) request(uint64, []byte) (embeddedResult, EmbeddedNativeStatus) {
	if n.mode == "goexit" {
		runtime.Goexit()
	}
	panic("controlled worker failure")
}

// copy is unreachable because this fixture never publishes an allocation.
// copy 不可达，因为此夹具绝不发布分配。
func (uncertainDriverNative) copy(embeddedResult, uint64) ([]byte, error) { panic("unexpected copy") }

// release is unreachable because this fixture owns no allocation.
// release 不可达，因为此夹具不拥有分配。
func (uncertainDriverNative) release(uint64, embeddedResult) EmbeddedNativeStatus {
	panic("unexpected release")
}

// close acknowledges only synthetic admission closure.
// close 仅确认合成入场关闭。
func (uncertainDriverNative) close(uint64) EmbeddedNativeStatus { return EmbeddedNativeOk }

// free must never be reached while the driver retains uncertain ownership.
// free 在驱动器保留不确定所有权时绝不能到达。
func (uncertainDriverNative) free(uint64) EmbeddedNativeStatus {
	panic("uncertain native ownership was released")
}

// TestEmbeddedDriverUncertainExit isolates deliberately unrecoverable SDK ownership in child test processes.
// TestEmbeddedDriverUncertainExit 将故意不可恢复的 SDK 所有权隔离在子测试进程中。
func TestEmbeddedDriverUncertainExit(t *testing.T) {
	if mode := os.Getenv("LUASKILLS_DRIVER_FAULT_TEST"); mode != "" {
		transport, err := createEmbeddedTransport(EmbeddedTransportConfig{1, 2, 4096, 2048, 2048}, uncertainDriverNative{mode})
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
		_, err = receipt.Result(ctx)
		var failure *EmbeddedDriverFailure
		if !errors.As(err, &failure) {
			t.Fatalf("missing infrastructure error: %v", err)
		}
		if err := driver.Close(ctx); !errors.As(err, &failure) {
			t.Fatalf("unsafe closure reported success: %v", err)
		}
		if receipt.Done() || receipt.State() != EmbeddedCommandUncertain || driver.Status().Closed || receipt.Forget() == nil {
			t.Fatal("uncertain work was falsely completed or forgotten")
		}
		if transport.Free() == nil || len(LiveEmbeddedCommandDrivers()) != 1 || len(driver.Commands()) != 1 {
			t.Fatal("uncertain owner became undiscoverable")
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"panic", "goexit"} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		command := exec.CommandContext(ctx, executable, "-test.run=^TestEmbeddedDriverUncertainExit$")
		command.Env = append(os.Environ(), "LUASKILLS_DRIVER_FAULT_TEST="+mode)
		output, err := command.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("isolated %s ownership test: %v\n%s", mode, err, output)
		}
	}
}
