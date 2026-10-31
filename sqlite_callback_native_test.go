//go:build cgo

package luaskills

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestSQLiteCapturedProviders verifies native generation identity, clear, callback errors and panics.
// TestSQLiteCapturedProviders 验证原生代次身份、清空、回调错误及 panic。
func TestSQLiteCapturedProviders(t *testing.T) {
	if os.Getenv("LUASKILLS_NATIVE_E2E") != "1" {
		t.Skip("requires native core")
	}
	// An isolated skill binds SQLite without linking any database driver into the SDK.
	// 隔离技能绑定 SQLite，不向 SDK 链接任何数据库驱动。
	root := t.TempDir()
	skill := filepath.Join(root, "user_skills", "sqlite-generation")
	if err := os.MkdirAll(filepath.Join(skill, "runtime"), 0755); err != nil {
		t.Fatal(err)
	}
	manifest := "name: sqlite-generation\nversion: 0.1.0\nsqlite:\n  enable: true\nentries:\n  - name: run\n    lua_entry: runtime/run.lua\n    lua_module: sqlite_generation_run\n"
	if err := os.WriteFile(filepath.Join(skill, "skill.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "runtime/run.lua"), []byte("-- Return the captured host provider value.\n-- 返回捕获的宿主提供方值。\nreturn function(args) return vulcan.sqlite.query_json(args).value end\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := ClearSQLiteProviderJSONCallback(); err != nil {
			t.Error(err)
		}
	})
	// Each callback validates the binding supplied by the core before returning its generation.
	// 每个回调在返回其代次前验证核心提供的绑定。
	newEngine := func(value string) *Client {
		t.Helper()
		if err := SetSQLiteProviderJSONCallback(func(request any) (any, error) {
			command := request.(map[string]any)
			if command["action"] != "query_json" || command["binding"].(map[string]any)["skill_id"] != "sqlite-generation" {
				return nil, errors.New("invalid core binding")
			}
			if value == "error" {
				return nil, errors.New("intentional SQLite host failure")
			}
			if value == "panic" {
				panic("intentional SQLite host panic")
			}
			return map[string]any{"value": value}, nil
		}); err != nil {
			t.Fatal(err)
		}
		client, err := NewClient(ClientOptions{RuntimeRoot: root, EnsureRuntimeLayout: true, HostOptions: map[string]any{"sqlite_callback_mode": "json"}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if _, err := client.Close(); err != nil {
				t.Error(err)
			}
		})
		if _, err := client.LoadFromRoots(StandardRoots(root)); err != nil {
			t.Fatal(err)
		}
		return client
	}
	first, second, failed, panicked := newEngine("A"), newEngine("B"), newEngine("error"), newEngine("panic")
	if err := ClearSQLiteProviderJSONCallback(); err != nil {
		t.Fatal(err)
	}
	runtime.GC()
	for _, check := range []struct {
		client *Client
		value  string
	}{{first, "A"}, {second, "B"}} {
		result, err := check.client.CallSkill("sqlite-generation-run", map[string]any{}, nil)
		if err != nil || result.Content != check.value {
			t.Fatalf("captured callback: %#v, %v", result, err)
		}
	}
	for _, client := range []*Client{failed, panicked} {
		if _, err := client.CallSkill("sqlite-generation-run", map[string]any{}, nil); err == nil || !strings.Contains(err.Error(), "intentional SQLite host") {
			t.Fatalf("callback failure did not cross FFI: %v", err)
		}
	}
	if _, err := second.Close(); err != nil {
		t.Fatal(err)
	}
	if result, err := first.CallSkill("sqlite-generation-run", map[string]any{}, nil); err != nil || result.Content != "A" {
		t.Fatalf("earlier generation released: %#v, %v", result, err)
	}
}
