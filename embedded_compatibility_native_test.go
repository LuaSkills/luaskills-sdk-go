//go:build cgo

package luaskills

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"unsafe"
)

// TestEmbeddedCompatibilityNative compares actual bootstrap evidence with the existing native describe command.
// TestEmbeddedCompatibilityNative 将实际引导证据与既有原生描述命令比较。
func TestEmbeddedCompatibilityNative(t *testing.T) {
	transport := nativeEmbeddedTest(t)
	bootstrap := transport.CoreDescription()
	bytes, err := json.Marshal(bootstrap)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := DecodeEmbeddedJSON(bytes)
	if err != nil {
		t.Fatal(err)
	}
	actual := embeddedRequest(t, transport, map[string]any{"type": "describe"}).(map[string]any)
	// The established transport schema shares only these five fields with standalone discovery.
	// 既有传输 Schema 仅与独立发现共享这五个字段。
	for _, name := range []string{"core_version", "protocol_version", "abi_structure_version", "commands", "runtime_commands"} {
		value := expected.(map[string]any)[name]
		if !reflect.DeepEqual(actual[name], value) {
			t.Fatalf("bootstrap differs from transport description: %s: actual=%#v expected=%#v", name, actual[name], value)
		}
	}
	first, err := transport.native.describe()
	if err != nil {
		t.Fatal(err)
	}
	first[0] = 0
	second, err := transport.native.describe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := validateEmbeddedDescription(second); err != nil {
		t.Fatal("borrowed native metadata was mutated", err)
	}
}

// TestEmbeddedCompatibilityBufferBounds rejects unsafe bounds before memory reads and copies valid bytes independently.
// TestEmbeddedCompatibilityBufferBounds 在内存读取前拒绝不安全边界，并独立复制有效字节。
func TestEmbeddedCompatibilityBufferBounds(t *testing.T) {
	storage := []byte{1}
	for _, bounds := range []struct {
		pointer unsafe.Pointer
		length  uint64
	}{
		{nil, 1}, {unsafe.Pointer(&storage[0]), 0},
		{unsafe.Pointer(&storage[0]), EmbeddedDescriptionMaxBytes + 1},
		{unsafe.Pointer(&storage[0]), ^uint64(0)},
	} {
		bytes, err := copyEmbeddedDescription(bounds.pointer, bounds.length)
		var failure *EmbeddedCompatibilityError
		if bytes != nil || !errors.As(err, &failure) {
			t.Fatal("invalid bounds were not rejected", err)
		}
	}
	bytes, err := copyEmbeddedDescription(unsafe.Pointer(&storage[0]), uint64(len(storage)))
	if err != nil {
		t.Fatal(err)
	}
	bytes[0] = 2
	if storage[0] != 1 {
		t.Fatal("description copy aliases original storage")
	}
}
