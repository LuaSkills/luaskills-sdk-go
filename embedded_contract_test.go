package luaskills

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

// TestEmbeddedContractIdentity compares all exact packaged bytes and generated command metadata.
// TestEmbeddedContractIdentity 比较全部精确包内字节及生成命令元数据。
func TestEmbeddedContractIdentity(t *testing.T) {
	data, err := os.ReadFile("contracts/embedded/v1/contract.json")
	if err != nil {
		t.Fatal(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	if hash != EmbeddedContractSHA256 {
		t.Fatal("packaged embedded contract digest differs from generated Go identity")
	}
	digest, err := os.ReadFile("contracts/embedded/v1/contract.sha256")
	if err != nil || strings.TrimSpace(string(digest)) != hash+"  contract.json" {
		t.Fatalf("packaged digest changed: %v", err)
	}
	decoded, err := DecodeEmbeddedJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	document := decoded.(map[string]any)
	if document["protocol_version"] != json.Number(fmt.Sprint(EmbeddedProtocolVersion)) {
		t.Fatal("protocol identity changed")
	}
	for _, entry := range []struct {
		name   string
		values []string
	}{{"commands", EmbeddedRootCommands()}, {"runtime_commands", EmbeddedRuntimeCommands()}} {
		values := []any{}
		for _, value := range entry.values {
			values = append(values, value)
		}
		if !reflect.DeepEqual(document[entry.name], values) {
			t.Fatalf("metadata drift: %s", entry.name)
		}
	}
	names := EmbeddedRootCommands()
	names[0] = "changed"
	if EmbeddedRootCommands()[0] == "changed" {
		t.Fatal("mutable command authority escaped")
	}
}

// TestEmbeddedConfigValidation rejects overflow and inconsistent budgets before allocation.
// TestEmbeddedConfigValidation 在分配前拒绝溢出及不一致预算。
func TestEmbeddedConfigValidation(t *testing.T) {
	for _, config := range []EmbeddedTransportConfig{{}, {1, 1, 1, 2, 1}, {1, 1, 1, 1, ^uint64(0)}} {
		if _, err := NewEmbeddedTransport(config); err == nil {
			t.Fatalf("accepted invalid budgets: %#v", config)
		}
	}
}
