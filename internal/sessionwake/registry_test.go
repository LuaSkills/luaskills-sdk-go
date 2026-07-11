package sessionwake

import "testing"

// TestRegistryStoresReplacesAndDeletes verifies the callback ownership lifecycle.
// TestRegistryStoresReplacesAndDeletes 校验回调所有权生命周期。
func TestRegistryStoresReplacesAndDeletes(t *testing.T) {
	var registry Registry
	first := Callback(func(uint64) error { return nil })
	second := Callback(func(uint64) error { return nil })

	registry.Store(7, first)
	loaded, ok := registry.Load(7)
	if !ok || loaded == nil {
		t.Fatal("expected first callback")
	}

	registry.Store(7, second)
	loaded, ok = registry.Load(7)
	if !ok || loaded == nil {
		t.Fatal("expected replacement callback")
	}

	registry.Delete(7)
	if _, ok := registry.Load(7); ok {
		t.Fatal("expected callback deletion")
	}
}
