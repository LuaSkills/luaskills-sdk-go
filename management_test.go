package luaskills

import "testing"

// TestSkillLifecycleFunctionNameMatchesNativeExports verifies managed skill lifecycle FFI names.
// TestSkillLifecycleFunctionNameMatchesNativeExports 校验受管理 skill 生命周期 FFI 函数名。
func TestSkillLifecycleFunctionNameMatchesNativeExports(t *testing.T) {
	tests := []struct {
		name        string
		systemPlane bool
		actionName  skillLifecycleAction
		expected    string
	}{
		{
			name:        "public disable",
			systemPlane: false,
			actionName:  skillLifecycleDisableAction,
			expected:    "luaskills_ffi_disable_skill_json",
		},
		{
			name:        "system disable",
			systemPlane: true,
			actionName:  skillLifecycleDisableAction,
			expected:    "luaskills_ffi_system_disable_skill_json",
		},
		{
			name:        "public enable",
			systemPlane: false,
			actionName:  skillLifecycleEnableAction,
			expected:    "luaskills_ffi_enable_skill_json",
		},
		{
			name:        "system uninstall",
			systemPlane: true,
			actionName:  skillLifecycleUninstallAction,
			expected:    "luaskills_ffi_system_uninstall_skill_json",
		},
		{
			name:        "public install",
			systemPlane: false,
			actionName:  skillLifecycleInstallAction,
			expected:    "luaskills_ffi_install_skill_json",
		},
		{
			name:        "system update",
			systemPlane: true,
			actionName:  skillLifecycleUpdateAction,
			expected:    "luaskills_ffi_system_update_skill_json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &SkillManagementClient{systemPlane: test.systemPlane}
			actual, err := client.functionName(test.actionName)
			if err != nil {
				t.Fatalf("resolve skill lifecycle function name: %v", err)
			}
			if actual != test.expected {
				t.Fatalf("unexpected skill lifecycle function name: %s", actual)
			}
		})
	}
}

// TestSkillLifecycleFunctionNameRejectsUnsupportedAction verifies unknown lifecycle actions fail.
// TestSkillLifecycleFunctionNameRejectsUnsupportedAction 校验未知生命周期动作会失败。
func TestSkillLifecycleFunctionNameRejectsUnsupportedAction(t *testing.T) {
	client := &SkillManagementClient{}
	_, err := client.functionName(skillLifecycleAction("delete_skill"))
	if err == nil {
		t.Fatalf("expected unsupported skill lifecycle action error")
	}
}

// TestPrivateUrlManifestFunctionNameMatchesNativeExports verifies private URL-manifest FFI function names.
// TestPrivateUrlManifestFunctionNameMatchesNativeExports 校验私有 URL manifest FFI 函数名。
func TestPrivateUrlManifestFunctionNameMatchesNativeExports(t *testing.T) {
	tests := []struct {
		name       string
		actionName privateUrlManifestAction
		expected   string
	}{
		{
			name:       "install",
			actionName: privateUrlManifestInstallAction,
			expected:   "luaskills_ffi_system_private_install_skill_from_url_manifest_json",
		},
		{
			name:       "update",
			actionName: privateUrlManifestUpdateAction,
			expected:   "luaskills_ffi_system_private_update_skill_from_url_manifest_json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := privateUrlManifestFunctionName(test.actionName); actual != test.expected {
				t.Fatalf("unexpected private URL manifest function name: %s", actual)
			}
		})
	}
}

// TestPrivateUrlManifestPayloadMatchesRustRequest verifies private URL-manifest request keys.
// TestPrivateUrlManifestPayloadMatchesRustRequest 校验私有 URL manifest 请求键。
func TestPrivateUrlManifestPayloadMatchesRustRequest(t *testing.T) {
	targetRoot := &RuntimeSkillRoot{Name: "ROOT", SkillsDir: "runtime/skills"}
	skillRoots := []RuntimeSkillRoot{{Name: "ROOT", SkillsDir: "runtime/skills"}}

	payload := privateUrlManifestPayload(42, skillRoots, "private.demo", "https://example.test/skill.json", targetRoot)

	if payload["engine_id"] != uint64(42) {
		t.Fatalf("unexpected engine_id: %#v", payload["engine_id"])
	}
	if payload["skill_id"] != "private.demo" {
		t.Fatalf("unexpected skill_id: %#v", payload["skill_id"])
	}
	if payload["manifest_url"] != "https://example.test/skill.json" {
		t.Fatalf("unexpected manifest_url: %#v", payload["manifest_url"])
	}
	if payload["target_root"] != targetRoot {
		t.Fatalf("unexpected target_root: %#v", payload["target_root"])
	}
	if payload["authority"] != AuthoritySystem {
		t.Fatalf("unexpected authority: %#v", payload["authority"])
	}
	roots, ok := payload["skill_roots"].([]RuntimeSkillRoot)
	if !ok || len(roots) != 1 || roots[0].Name != "ROOT" {
		t.Fatalf("unexpected skill_roots: %#v", payload["skill_roots"])
	}
}

// TestValidateSkillInstallRequestRejectsMissingSourceType verifies SDK requests must name the Rust protocol source_type.
// TestValidateSkillInstallRequestRejectsMissingSourceType 校验 SDK 请求必须声明 Rust 协议 source_type。
func TestValidateSkillInstallRequestRejectsMissingSourceType(t *testing.T) {
	source := "LuaSkills/demo-skill"
	request := SkillInstallRequest{Source: &source}

	err := validateSkillInstallRequest(skillLifecycleInstallAction, request)

	if err == nil {
		t.Fatalf("expected missing source_type error")
	}
}

// TestValidateSkillInstallRequestRejectsPrivateInstallWithoutManifestSource verifies private installs need an explicit manifest URL source.
// TestValidateSkillInstallRequestRejectsPrivateInstallWithoutManifestSource 校验私有安装必须携带显式 manifest URL 来源。
func TestValidateSkillInstallRequestRejectsPrivateInstallWithoutManifestSource(t *testing.T) {
	skillID := "private.demo"
	request := SkillInstallRequest{
		SkillID:    &skillID,
		SourceType: SkillInstallSourcePrivateURLManifest,
	}

	err := validateSkillInstallRequest(skillLifecycleInstallAction, request)

	if err == nil {
		t.Fatalf("expected missing private manifest source error")
	}
}

// TestValidateSkillInstallRequestRejectsNonHTTPPrivateSource verifies private manifest sources must be HTTP URLs.
// TestValidateSkillInstallRequestRejectsNonHTTPPrivateSource 校验私有 manifest 来源必须是 HTTP URL。
func TestValidateSkillInstallRequestRejectsNonHTTPPrivateSource(t *testing.T) {
	skillID := "private.demo"
	source := "file:///tmp/private.json"
	request := SkillInstallRequest{
		SkillID:    &skillID,
		Source:     &source,
		SourceType: SkillInstallSourcePrivateURLManifest,
	}

	err := validateSkillInstallRequest(skillLifecycleInstallAction, request)

	if err == nil {
		t.Fatalf("expected non-http private manifest source error")
	}
}

// TestValidateSkillInstallRequestAcceptsGithubUpdateDerivedFromSource verifies update can derive a GitHub skill id from source.
// TestValidateSkillInstallRequestAcceptsGithubUpdateDerivedFromSource 校验更新可从 GitHub 来源派生 skill 标识。
func TestValidateSkillInstallRequestAcceptsGithubUpdateDerivedFromSource(t *testing.T) {
	source := "LuaSkills/demo-skill"
	request := SkillInstallRequest{
		Source:     &source,
		SourceType: SkillInstallSourceGithub,
	}

	if err := validateSkillInstallRequest(skillLifecycleUpdateAction, request); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

// TestValidatePrivateUrlManifestInputRejectsRelativeURL verifies dedicated private manifest helpers reject relative URLs.
// TestValidatePrivateUrlManifestInputRejectsRelativeURL 校验专用私有 manifest 辅助入口拒绝相对 URL。
func TestValidatePrivateUrlManifestInputRejectsRelativeURL(t *testing.T) {
	err := validatePrivateUrlManifestInput("private.demo", "/private/skill.json")

	if err == nil {
		t.Fatalf("expected relative manifest URL error")
	}
}
