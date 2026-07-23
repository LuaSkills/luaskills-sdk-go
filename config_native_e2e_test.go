package luaskills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestSkillConfigNativeRoundTrip verifies the SDK against one matching native LuaSkills library.
// TestSkillConfigNativeRoundTrip 使用一个匹配的原生 LuaSkills 库验证 SDK。
//
// t receives assertion failures from the Go test runner.
// t 接收 Go 测试运行器提供的断言失败。
func TestSkillConfigNativeRoundTrip(t *testing.T) {
	if os.Getenv("LUASKILLS_NATIVE_E2E") != "1" {
		t.Skip("LUASKILLS_NATIVE_E2E is not enabled")
	}

	// HostRoot isolates runtime layout, skill packages, and persistent configuration.
	// HostRoot 隔离运行时布局、技能包与持久化配置。
	hostRoot := t.TempDir()
	// RuntimeRoot satisfies unrelated fixed engine layout requirements.
	// RuntimeRoot 满足其他固定引擎布局要求。
	runtimeRoot := filepath.Join(hostRoot, "runtime")
	// ConfigRoot is the explicit user-level package-configuration root.
	// ConfigRoot 是显式用户级技能包配置根目录。
	configRoot := filepath.Join(hostRoot, "config")
	// RootSkills verifies dedicated system-skill configuration routing.
	// RootSkills 验证专用系统技能配置路由。
	rootSkills := filepath.Join(hostRoot, "root-skills")
	// UserSkills verifies ordinary package-configuration routing.
	// UserSkills 验证普通技能包配置路由。
	userSkills := filepath.Join(hostRoot, "user-skills")
	writeNativeConfigSkill(t, rootSkills, "system-config-e2e")
	writeNativeConfigSkill(t, userSkills, "user-config-e2e")

	// Client is linked to the matching native library by the cgo test environment.
	// Client 由 cgo 测试环境链接到匹配的原生库。
	client, err := NewClient(ClientOptions{
		RuntimeRoot: runtimeRoot,
		HostOptions: map[string]any{
			"skill_config_root":              configRoot,
			"skill_config_lock_timeout_ms":   5_000,
			"skill_config_watch_debounce_ms": 20,
		},
		EnsureRuntimeLayout: true,
	})
	if err != nil {
		t.Fatalf("create native client: %v", err)
	}
	t.Cleanup(func() {
		if _, closeErr := client.Close(); closeErr != nil {
			t.Errorf("close native client: %v", closeErr)
		}
	})
	if _, err := client.LoadFromRoots([]RuntimeSkillRoot{
		{Name: "ROOT", SkillsDir: rootSkills},
		{Name: "USER", SkillsDir: userSkills},
	}); err != nil {
		t.Fatalf("load native config skills: %v", err)
	}

	// Descriptor proves every common declaration type crosses the native boundary.
	// Descriptor 证明每个常见声明类型均能穿过原生边界。
	descriptors, err := client.Config.Describe(SkillPackageConfigDescribeOptions{
		SkillID: "user-config-e2e",
	})
	if err != nil {
		t.Fatalf("describe native package configuration: %v", err)
	}
	// Types preserves the canonical manifest declaration order.
	// Types 保持规范清单声明顺序。
	types := make([]SkillPackageConfigType, len(descriptors[0].Items))
	for index, item := range descriptors[0].Items {
		types[index] = item.Type
	}
	if !reflect.DeepEqual(types, []SkillPackageConfigType{
		SkillPackageConfigTypeString,
		SkillPackageConfigTypeInteger,
		SkillPackageConfigTypeFloat,
		SkillPackageConfigTypeEnum,
		SkillPackageConfigTypeBoolean,
	}) {
		t.Fatalf("unexpected native declaration types: %#v", types)
	}
	status, err := client.Config.Validate("user-config-e2e")
	if err != nil {
		t.Fatalf("validate incomplete native configuration: %v", err)
	}
	if status.Complete {
		t.Fatal("configuration unexpectedly complete before required token write")
	}

	// Written is one atomic typed batch and supplies the subsequent CAS revision.
	// Written 是一个原子类型化批次并提供后续 CAS 修订号。
	written, err := client.Config.SetValues(
		"user-config-e2e",
		map[string]any{
			"token":   "secret",
			"retries": 4,
			"ratio":   0.5,
			"mode":    "fast",
			"enabled": true,
		},
		"",
	)
	if err != nil {
		t.Fatalf("write native configuration batch: %v", err)
	}
	if !written.Changed {
		t.Fatal("native configuration batch did not report a change")
	}
	if !reflect.DeepEqual(
		written.ChangedKeys,
		[]string{"enabled", "mode", "ratio", "retries", "token"},
	) {
		t.Fatalf("unexpected changed keys: %#v", written.ChangedKeys)
	}
	status, err = client.Config.Validate("user-config-e2e")
	if err != nil || !status.Complete {
		t.Fatalf("configuration did not become complete: status=%#v error=%v", status, err)
	}
	retries, err := client.Config.Get("user-config-e2e", "retries")
	if err != nil || retries.Value == nil || *retries.Value != "4" {
		t.Fatalf("unexpected canonical retry value: value=%#v error=%v", retries, err)
	}
	// UserEntries are raw records routed through the ordinary persisted store.
	// UserEntries 是经普通持久化存储路由的原始记录。
	userEntries, err := client.Config.List("user-config-e2e")
	if err != nil {
		t.Fatalf("list user configuration entries: %v", err)
	}
	if len(userEntries) != 5 {
		t.Fatalf("unexpected user configuration entry count: %d", len(userEntries))
	}
	for _, entry := range userEntries {
		if entry.StoreScope != SkillConfigStoreScopeSkills {
			t.Fatalf("unexpected user entry store scope: %#v", entry)
		}
	}

	// A stale revision must not overwrite the committed snapshot.
	// 过期修订号不得覆盖已提交快照。
	if _, err := client.Config.Set("user-config-e2e", "retries", 5, "0"); err == nil ||
		!strings.Contains(err.Error(), "CONFIG_REVISION_CONFLICT") {
		t.Fatalf("stale CAS was not rejected correctly: %v", err)
	}
	// One invalid member must reject the complete batch atomically.
	// 一个非法成员必须原子拒绝完整批次。
	if _, err := client.Config.SetValues(
		"user-config-e2e",
		map[string]any{"retries": 99, "enabled": false},
		written.Revision,
	); err == nil || !strings.Contains(err.Error(), "CONFIG_VALUE_OUT_OF_RANGE") {
		t.Fatalf("invalid batch was not rejected correctly: %v", err)
	}
	enabled, err := client.Config.Get("user-config-e2e", "enabled")
	if err != nil || enabled.Value == nil || *enabled.Value != "true" {
		t.Fatalf("invalid batch changed committed state: value=%#v error=%v", enabled, err)
	}

	// ROOT-owned values are persisted only in the dedicated system store.
	// ROOT 所属值仅持久化到专用系统存储。
	if _, err := client.Config.SetValues(
		"system-config-e2e",
		map[string]any{
			"token":   "system-secret",
			"retries": 1,
			"ratio":   1.0,
			"mode":    "safe",
			"enabled": false,
		},
		"",
	); err != nil {
		t.Fatalf("write system package configuration: %v", err)
	}
	// SystemDocument is the parsed dedicated system-skills store.
	// SystemDocument 是已解析的专用 system-skills 存储。
	systemDocument := readNativeConfigDocument(
		t,
		filepath.Join(configRoot, "system-skills", "config.json"),
	)
	// NormalDocument is the parsed ordinary skills store.
	// NormalDocument 是已解析的普通 skills 存储。
	normalDocument := readNativeConfigDocument(
		t,
		filepath.Join(configRoot, "skills", "config.json"),
	)
	if got := systemDocument["system-config-e2e"]["token"]; got != "system-secret" {
		t.Fatalf("unexpected system store token: %q", got)
	}
	if _, exists := normalDocument["system-config-e2e"]; exists {
		t.Fatal("system package leaked into the ordinary configuration store")
	}

	// Cursor pagination must expose subsequent transactions without skipping.
	// 游标分页必须公开后续事务且不得跳过。
	firstPage, err := client.Config.PollEvents("", 1)
	if err != nil || len(firstPage.Events) != 1 {
		t.Fatalf("unexpected first event page: page=%#v error=%v", firstPage, err)
	}
	secondPage, err := client.Config.PollEvents(firstPage.NextSequence, 10)
	if err != nil || len(secondPage.Events) < 1 {
		t.Fatalf("unexpected second event page: page=%#v error=%v", secondPage, err)
	}
}

// writeNativeConfigSkill creates one valid package containing every common declaration type.
// writeNativeConfigSkill 创建一个包含所有常见声明类型的合法技能包。
//
// t receives setup failures, skillsRoot owns package directories, and skillID is the stable package identifier.
// t 接收准备失败，skillsRoot 保存技能包目录，skillID 是稳定技能包标识符。
func writeNativeConfigSkill(t *testing.T, skillsRoot string, skillID string) {
	t.Helper()
	// PackageRoot owns the generated manifest and one inert Lua entry.
	// PackageRoot 保存生成的清单与一个无副作用 Lua 入口。
	packageRoot := filepath.Join(skillsRoot, skillID)
	if err := os.MkdirAll(filepath.Join(packageRoot, "runtime"), 0o755); err != nil {
		t.Fatalf("create native config skill directory: %v", err)
	}
	// Manifest declares all common scalar types and one required secret.
	// Manifest 声明所有常见标量类型与一个必填密钥。
	manifest := fmt.Sprintf(`name: %s
version: 1.0.0
enable: true
debug: false
config:
  - key: token
    type: string
    required: true
    sensitive: true
    description: Access token
    constraints:
      min_length: 1
      max_length: 128
  - key: retries
    type: integer
    default: 3
    description: Retry count
    constraints:
      minimum: 0
      maximum: 10
  - key: ratio
    type: float
    default: 0.25
    description: Sampling ratio
    constraints:
      minimum: 0.0
      maximum: 1.0
  - key: mode
    type: enum
    default: safe
    description: Execution mode
    options:
      - value: safe
        label: Safe
        description: Conservative mode
      - value: fast
        label: Fast
        description: Fast mode
  - key: enabled
    type: boolean
    default: false
    description: Feature switch
entries:
  - name: ping
    description: Return a stable response.
    lua_entry: runtime/main.lua
    lua_module: %s.main
`, skillID, skillID)
	if err := os.WriteFile(filepath.Join(packageRoot, "skill.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write native config skill manifest: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(packageRoot, "runtime", "main.lua"),
		[]byte("return function() return 'ok' end\n"),
		0o644,
	); err != nil {
		t.Fatalf("write native config skill entry: %v", err)
	}
}

// readNativeConfigDocument decodes the package map from one strict persisted store.
// readNativeConfigDocument 从一个严格持久化存储解码技能包映射。
//
// t receives read failures and path identifies the exact store document.
// t 接收读取失败，path 标识精确存储文档。
func readNativeConfigDocument(t *testing.T, path string) map[string]map[string]string {
	t.Helper()
	// DocumentBytes contains the exact persisted JSON document.
	// DocumentBytes 包含精确持久化 JSON 文档。
	documentBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read native configuration document: %v", err)
	}
	// Document contains only the persisted fields required by this assertion.
	// Document 仅包含当前断言所需的持久化字段。
	var document struct {
		// Skills maps stable package identifiers to canonical string values.
		// Skills 把稳定技能包标识符映射到规范字符串值。
		Skills map[string]map[string]string `json:"skills"`
	}
	if err := json.Unmarshal(documentBytes, &document); err != nil {
		t.Fatalf("decode native configuration document: %v", err)
	}
	return document.Skills
}
