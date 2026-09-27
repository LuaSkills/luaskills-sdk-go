//go:build cgo

package luaskills_test

import (
	"context"
	"os"
	"testing"
	"time"

	luaskills "github.com/LuaSkills/luaskills-sdk-go"
)

// TestEmbeddedDriverPublicConsumer exercises actual driver ownership and generated decoding through exported APIs only.
// TestEmbeddedDriverPublicConsumer 仅通过导出接口验证实际驱动所有权及生成解码。
func TestEmbeddedDriverPublicConsumer(t *testing.T) {
	if os.Getenv("LUASKILLS_NATIVE_E2E") != "1" {
		t.Skip("LUASKILLS_NATIVE_E2E is not enabled")
	}
	transport, err := luaskills.NewEmbeddedTransport(luaskills.EmbeddedTransportConfig{MaxRuntimes: 1, MaxResultBuffers: 2, MaxResultBytes: 65536, MaxResponseBytes: 32768, MaxRequestBytes: 1024})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := transport.Close(); err != nil {
			t.Error(err)
		}
		if err := transport.Free(); err != nil {
			t.Error(err)
		}
	})
	driver, err := luaskills.NewEmbeddedCommandDriver(transport, luaskills.EmbeddedDriverConfig{WorkWorkers: 1, MaxWorkCommands: 1, MaxControlCommands: 1})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := driver.Close(ctx); err != nil {
			t.Error(err)
		}
		for _, command := range driver.Commands() {
			if err := command.Forget(); err != nil {
				t.Error(err)
			}
		}
	})
	receipt, err := driver.Submit(context.Background(), luaskills.EmbeddedInputCommandDescribe{Type: luaskills.EmbeddedInputCommandDescribeTypeDescribe})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := receipt.Result(ctx); err != nil {
		t.Fatal(err)
	}
	description, err := luaskills.DecodeEmbeddedOutputRootDescribeResponse(receipt.ResponseBytes())
	if err != nil || description.Result.Limits.MaxResponseBytes != transport.Config().MaxResponseBytes {
		t.Fatalf("public driver did not preserve the native envelope: %#v %v", description, err)
	}
	if !receipt.Done() || receipt.State() != luaskills.EmbeddedCommandCompleted || receipt.Lane() != luaskills.EmbeddedControlLane || receipt.CommandID() == 0 {
		t.Fatal("public receipt ownership mismatch")
	}
}
