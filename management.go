package luaskills

import (
	"fmt"
	"net/url"
	"strings"
)

// SkillManagementClient is the ordinary or system lifecycle namespace over JSON FFI management entrypoints.
// SkillManagementClient 是覆盖 JSON FFI 管理入口的普通或 system 生命周期命名空间。
type SkillManagementClient struct {
	client      *Client
	systemPlane bool
	authority   Authority
}

// SystemSkillManagementClient is the system engine namespace with host-injected authority.
// SystemSkillManagementClient 是携带宿主注入权限的 system 引擎命名空间。
type SystemSkillManagementClient struct {
	SkillManagementClient
}

// skillLifecycleAction identifies one fixed skill lifecycle JSON FFI operation.
// skillLifecycleAction 标识一个固定的 skill 生命周期 JSON FFI 操作。
type skillLifecycleAction string

const (
	// skillLifecycleDisableAction disables one managed skill.
	// skillLifecycleDisableAction 停用单个受管理 skill。
	skillLifecycleDisableAction skillLifecycleAction = "disable_skill"
	// skillLifecycleEnableAction enables one managed skill.
	// skillLifecycleEnableAction 启用单个受管理 skill。
	skillLifecycleEnableAction skillLifecycleAction = "enable_skill"
	// skillLifecycleUninstallAction uninstalls one managed skill.
	// skillLifecycleUninstallAction 卸载单个受管理 skill。
	skillLifecycleUninstallAction skillLifecycleAction = "uninstall_skill"
	// skillLifecycleInstallAction installs one managed skill.
	// skillLifecycleInstallAction 安装单个受管理 skill。
	skillLifecycleInstallAction skillLifecycleAction = "install_skill"
	// skillLifecycleUpdateAction updates one managed skill.
	// skillLifecycleUpdateAction 更新单个受管理 skill。
	skillLifecycleUpdateAction skillLifecycleAction = "update_skill"
)

// privateUrlManifestAction identifies one private URL-manifest lifecycle operation.
// privateUrlManifestAction 标识单个私有 URL manifest 生命周期操作。
type privateUrlManifestAction string

const (
	// privateUrlManifestInstallAction identifies the private URL-manifest install operation.
	// privateUrlManifestInstallAction 标识私有 URL manifest 安装操作。
	privateUrlManifestInstallAction privateUrlManifestAction = "install"
	// privateUrlManifestUpdateAction identifies the private URL-manifest update operation.
	// privateUrlManifestUpdateAction 标识私有 URL manifest 更新操作。
	privateUrlManifestUpdateAction privateUrlManifestAction = "update"
)

// RuntimeLeases returns one authority-bound runtime-lease namespace.
// RuntimeLeases 返回一个绑定 authority 的运行时租约命名空间。
func (m *SystemSkillManagementClient) RuntimeLeases() *RuntimeLeaseClient {
	return &RuntimeLeaseClient{
		client:        m.client,
		authority:     m.resolveAuthority(""),
		bindAuthority: true,
	}
}

// ListEntries lists runtime entries visible to the bound authority.
// ListEntries 列出当前绑定 authority 可见的运行时入口。
func (m *SystemSkillManagementClient) ListEntries() ([]map[string]any, error) {
	return m.client.ListEntries(m.resolveAuthority(""))
}

// ListSkillHelp lists runtime help trees visible to the bound authority.
// ListSkillHelp 列出当前绑定 authority 可见的运行时帮助树。
func (m *SystemSkillManagementClient) ListSkillHelp() ([]map[string]any, error) {
	return m.client.ListSkillHelp(m.resolveAuthority(""))
}

// RenderSkillHelpDetail renders one help flow detail visible to the bound authority.
// RenderSkillHelpDetail 渲染当前绑定 authority 可见的单个帮助流程详情。
func (m *SystemSkillManagementClient) RenderSkillHelpDetail(skillID string, flowName string, requestContext any) (map[string]any, error) {
	return m.client.RenderSkillHelpDetail(skillID, flowName, m.resolveAuthority(""), requestContext)
}

// PromptArgumentCompletions queries prompt argument completions visible to the bound authority.
// PromptArgumentCompletions 查询当前绑定 authority 可见的 prompt 参数补全项。
func (m *SystemSkillManagementClient) PromptArgumentCompletions(promptName string, argumentName string) ([]string, error) {
	return m.client.PromptArgumentCompletions(promptName, argumentName, m.resolveAuthority(""))
}

// IsSkill returns whether one canonical tool name resolves to one visible skill entry.
// IsSkill 返回某个 canonical 工具名是否解析为一个可见技能入口。
func (m *SystemSkillManagementClient) IsSkill(toolName string) (bool, error) {
	return m.client.IsSkill(toolName, m.resolveAuthority(""))
}

// SkillNameForTool resolves the visible owning skill id for one canonical tool name when available.
// SkillNameForTool 在可见时解析某个 canonical 工具名所属的技能标识。
func (m *SystemSkillManagementClient) SkillNameForTool(toolName string) (*string, error) {
	return m.client.SkillNameForTool(toolName, m.resolveAuthority(""))
}

// InstallPrivateUrlManifest installs one host-approved private URL-manifest skill through the system-private FFI endpoint.
// InstallPrivateUrlManifest 通过 system 私有 FFI 入口安装单个宿主已批准的私有 URL manifest 技能。
func (m *SystemSkillManagementClient) InstallPrivateUrlManifest(skillRoots []RuntimeSkillRoot, skillID string, manifestURL string, targetRoot *RuntimeSkillRoot) (map[string]any, error) {
	return m.privateUrlManifest(privateUrlManifestInstallAction, skillRoots, skillID, manifestURL, targetRoot)
}

// UpdatePrivateUrlManifest updates one host-approved private URL-manifest skill through the system-private FFI endpoint.
// UpdatePrivateUrlManifest 通过 system 私有 FFI 入口更新单个宿主已批准的私有 URL manifest 技能。
func (m *SystemSkillManagementClient) UpdatePrivateUrlManifest(skillRoots []RuntimeSkillRoot, skillID string, manifestURL string, targetRoot *RuntimeSkillRoot) (map[string]any, error) {
	return m.privateUrlManifest(privateUrlManifestUpdateAction, skillRoots, skillID, manifestURL, targetRoot)
}

// Disable disables one skill through formal root-chain lifecycle state.
// Disable 通过正式 root 链生命周期状态停用单个 skill。
func (m *SkillManagementClient) Disable(skillRoots []RuntimeSkillRoot, skillID string, reason string) (map[string]any, error) {
	var result map[string]any
	payload := map[string]any{
		"engine_id":   m.client.engineID,
		"skill_roots": skillRoots,
		"skill_id":    skillID,
		"reason":      nil,
	}
	if reason != "" {
		payload["reason"] = reason
	}
	m.addAuthority(payload, "")
	err := m.callLifecycle(skillLifecycleDisableAction, payload, &result)
	return result, err
}

// Enable enables one skill through formal root-chain lifecycle state.
// Enable 通过正式 root 链生命周期状态启用单个 skill。
func (m *SkillManagementClient) Enable(skillRoots []RuntimeSkillRoot, skillID string) (map[string]any, error) {
	var result map[string]any
	payload := map[string]any{
		"engine_id":   m.client.engineID,
		"skill_roots": skillRoots,
		"skill_id":    skillID,
	}
	m.addAuthority(payload, "")
	err := m.callLifecycle(skillLifecycleEnableAction, payload, &result)
	return result, err
}

// Install installs one managed skill through the current lifecycle namespace.
// Install 通过当前生命周期命名空间安装单个受管 skill。
func (m *SkillManagementClient) Install(skillRoots []RuntimeSkillRoot, request SkillInstallRequest, options LifecycleOptions) (map[string]any, error) {
	return m.apply(skillLifecycleInstallAction, skillRoots, request, options)
}

// Update updates one managed skill through the current lifecycle namespace.
// Update 通过当前生命周期命名空间更新单个受管 skill。
func (m *SkillManagementClient) Update(skillRoots []RuntimeSkillRoot, request SkillInstallRequest, options LifecycleOptions) (map[string]any, error) {
	return m.apply(skillLifecycleUpdateAction, skillRoots, request, options)
}

// Uninstall uninstalls one skill and optionally cleans its databases.
// Uninstall 卸载单个 skill，并可选清理其数据库。
func (m *SkillManagementClient) Uninstall(skillRoots []RuntimeSkillRoot, skillID string, uninstallOptions SkillUninstallOptions, lifecycleOptions LifecycleOptions) (map[string]any, error) {
	var result map[string]any
	payload := map[string]any{
		"engine_id":   m.client.engineID,
		"skill_roots": skillRoots,
		"skill_id":    skillID,
		"options":     uninstallOptions,
		"target_root": lifecycleOptions.TargetRoot,
	}
	m.addAuthority(payload, lifecycleOptions.Authority)
	err := m.callLifecycle(skillLifecycleUninstallAction, payload, &result)
	return result, err
}

// apply executes one install or update JSON FFI action.
// apply 执行单个 install 或 update JSON FFI 动作。
func (m *SkillManagementClient) apply(actionName skillLifecycleAction, skillRoots []RuntimeSkillRoot, request SkillInstallRequest, options LifecycleOptions) (map[string]any, error) {
	if err := validateSkillInstallRequest(actionName, request); err != nil {
		return nil, err
	}
	var result map[string]any
	payload := map[string]any{
		"engine_id":   m.client.engineID,
		"skill_roots": skillRoots,
		"request":     request,
		"target_root": options.TargetRoot,
	}
	m.addAuthority(payload, options.Authority)
	err := m.callLifecycle(actionName, payload, &result)
	return result, err
}

// callLifecycle resolves one validated lifecycle FFI function name and dispatches the request.
// callLifecycle 解析单个已验证的生命周期 FFI 函数名并分发请求。
func (m *SkillManagementClient) callLifecycle(actionName skillLifecycleAction, payload map[string]any, result *map[string]any) error {
	functionName, err := m.functionName(actionName)
	if err != nil {
		return err
	}
	return m.client.call(functionName, payload, result)
}

// privateUrlManifest executes one host-private install or update operation for a URL manifest skill.
// privateUrlManifest 执行单个 URL manifest 技能的宿主私有安装或更新操作。
func (m *SystemSkillManagementClient) privateUrlManifest(actionName privateUrlManifestAction, skillRoots []RuntimeSkillRoot, skillID string, manifestURL string, targetRoot *RuntimeSkillRoot) (map[string]any, error) {
	if err := validatePrivateUrlManifestInput(skillID, manifestURL); err != nil {
		return nil, err
	}
	var result map[string]any
	payload := privateUrlManifestPayload(m.client.engineID, skillRoots, skillID, manifestURL, targetRoot)
	err := m.client.call(privateUrlManifestFunctionName(actionName), payload, &result)
	return result, err
}

// privateUrlManifestFunctionName returns the exact native JSON FFI function name for one private URL-manifest action.
// privateUrlManifestFunctionName 返回单个私有 URL manifest 操作对应的精确原生 JSON FFI 函数名。
func privateUrlManifestFunctionName(actionName privateUrlManifestAction) string {
	return "luaskills_ffi_system_private_" + string(actionName) + "_skill_from_url_manifest_json"
}

// privateUrlManifestPayload builds the exact JSON FFI payload required by Rust PrivateUrlManifestSkillJsonRequest.
// privateUrlManifestPayload 构造 Rust PrivateUrlManifestSkillJsonRequest 所需的精确 JSON FFI 载荷。
func privateUrlManifestPayload(engineID uint64, skillRoots []RuntimeSkillRoot, skillID string, manifestURL string, targetRoot *RuntimeSkillRoot) map[string]any {
	return map[string]any{
		"engine_id":    engineID,
		"skill_roots":  skillRoots,
		"skill_id":     skillID,
		"manifest_url": manifestURL,
		"target_root":  targetRoot,
		"authority":    AuthoritySystem,
	}
}

// validateSkillInstallRequest rejects malformed SDK-side install and update requests before JSON FFI dispatch.
// validateSkillInstallRequest 在 JSON FFI 分发前拒绝格式错误的 SDK 侧安装与更新请求。
func validateSkillInstallRequest(actionName skillLifecycleAction, request SkillInstallRequest) error {
	if !isSupportedSkillInstallSourceType(request.SourceType) {
		return fmt.Errorf("skill install request source_type must be one of github, official_hub, url, private_url_manifest")
	}
	if request.SkillID != nil {
		if err := validateExactNonBlankString(*request.SkillID, "skill_id"); err != nil {
			return err
		}
	}
	if request.Source != nil {
		if err := validateExactNonBlankString(*request.Source, "source"); err != nil {
			return err
		}
		if request.SourceType == SkillInstallSourceURL || request.SourceType == SkillInstallSourcePrivateURLManifest {
			if err := validateHTTPURL(*request.Source, "source"); err != nil {
				return err
			}
		}
	}
	switch actionName {
	case skillLifecycleInstallAction:
		return validateInstallSourcePresence(request)
	case skillLifecycleUpdateAction:
		return validateUpdateSourcePresence(request)
	default:
		return nil
	}
}

// validateInstallSourcePresence enforces the source fields that Rust install resolution actually consumes.
// validateInstallSourcePresence 强制校验 Rust 安装解析实际会消费的来源字段。
func validateInstallSourcePresence(request SkillInstallRequest) error {
	hasSkillID := request.SkillID != nil
	hasSource := request.Source != nil
	switch request.SourceType {
	case SkillInstallSourceGithub:
		if !hasSource {
			return fmt.Errorf("github install request requires source")
		}
	case SkillInstallSourceOfficialHub:
		if !hasSkillID && !hasSource {
			return fmt.Errorf("official_hub install request requires skill_id or source")
		}
	case SkillInstallSourceURL:
		if !hasSkillID || !hasSource {
			return fmt.Errorf("url install request requires skill_id and source")
		}
	case SkillInstallSourcePrivateURLManifest:
		if !hasSkillID || !hasSource {
			return fmt.Errorf("private_url_manifest install request requires skill_id and source")
		}
	}
	return nil
}

// validateUpdateSourcePresence enforces the identifiers that Rust update resolution can derive.
// validateUpdateSourcePresence 强制校验 Rust 更新解析能够派生出的标识字段。
func validateUpdateSourcePresence(request SkillInstallRequest) error {
	hasSkillID := request.SkillID != nil
	hasSource := request.Source != nil
	switch request.SourceType {
	case SkillInstallSourceGithub, SkillInstallSourceOfficialHub:
		if !hasSkillID && !hasSource {
			return fmt.Errorf("%s update request requires skill_id or source", request.SourceType)
		}
	case SkillInstallSourceURL:
		if !hasSkillID {
			return fmt.Errorf("url update request requires skill_id")
		}
	case SkillInstallSourcePrivateURLManifest:
		if !hasSkillID {
			return fmt.Errorf("private_url_manifest update request requires skill_id")
		}
	}
	return nil
}

// validatePrivateUrlManifestInput validates the dedicated private URL-manifest shortcut payload.
// validatePrivateUrlManifestInput 校验专用私有 URL manifest 快捷入口载荷。
func validatePrivateUrlManifestInput(skillID string, manifestURL string) error {
	if err := validateExactNonBlankString(skillID, "skill_id"); err != nil {
		return err
	}
	return validateHTTPURL(manifestURL, "manifest_url")
}

// isSupportedSkillInstallSourceType reports whether one source type is part of the native JSON protocol.
// isSupportedSkillInstallSourceType 判断单个来源类型是否属于原生 JSON 协议。
func isSupportedSkillInstallSourceType(sourceType SkillInstallSourceType) bool {
	switch sourceType {
	case SkillInstallSourceGithub,
		SkillInstallSourceOfficialHub,
		SkillInstallSourceURL,
		SkillInstallSourcePrivateURLManifest:
		return true
	default:
		return false
	}
}

// validateExactNonBlankString rejects empty or implicitly trimmed JSON string fields.
// validateExactNonBlankString 拒绝空白或需要隐式裁剪的 JSON 字符串字段。
func validateExactNonBlankString(value string, fieldName string) error {
	if value == "" || strings.TrimSpace(value) != value {
		return fmt.Errorf("skill install request %s must be a non-empty string without surrounding whitespace", fieldName)
	}
	return nil
}

// validateHTTPURL rejects non-HTTP, relative, or credential-bearing URLs before native download resolution.
// validateHTTPURL 在原生下载解析前拒绝非 HTTP、相对路径或携带账号信息的 URL。
func validateHTTPURL(value string, fieldName string) error {
	if err := validateExactNonBlankString(value, fieldName); err != nil {
		return err
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("skill install request %s must be an absolute HTTP or HTTPS URL", fieldName)
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return fmt.Errorf("skill install request %s must use http or https", fieldName)
	}
	if parsed.User != nil {
		return fmt.Errorf("skill install request %s must not include credentials", fieldName)
	}
	return nil
}

// functionName builds the concrete JSON FFI function name for the current namespace.
// functionName 为当前命名空间构造具体 JSON FFI 函数名称。
func (m *SkillManagementClient) functionName(actionName skillLifecycleAction) (string, error) {
	baseName, err := skillLifecycleActionValue(actionName)
	if err != nil {
		return "", err
	}
	if m.systemPlane {
		return "luaskills_ffi_system_" + baseName + "_json", nil
	}
	return "luaskills_ffi_" + baseName + "_json", nil
}

// skillLifecycleActionValue returns the validated native base name for one lifecycle action.
// skillLifecycleActionValue 返回单个生命周期操作对应的已验证原生基础名称。
func skillLifecycleActionValue(actionName skillLifecycleAction) (string, error) {
	switch actionName {
	case skillLifecycleDisableAction,
		skillLifecycleEnableAction,
		skillLifecycleUninstallAction,
		skillLifecycleInstallAction,
		skillLifecycleUpdateAction:
		return string(actionName), nil
	default:
		return "", fmt.Errorf("unsupported skill lifecycle action: %s", actionName)
	}
}

// addAuthority adds authority when the current namespace targets system entrypoints.
// addAuthority 在当前命名空间指向 system 入口时添加权限。
func (m *SkillManagementClient) addAuthority(payload map[string]any, override Authority) {
	if !m.systemPlane {
		return
	}
	payload["authority"] = m.resolveAuthority(override)
}

// resolveAuthority resolves one override or bound authority to the concrete JSON payload value.
// resolveAuthority 将单个覆盖值或绑定 authority 解析为具体 JSON 载荷值。
func (m *SkillManagementClient) resolveAuthority(override Authority) Authority {
	authority := override
	if authority == "" {
		authority = m.authority
	}
	if authority == "" {
		authority = AuthoritySystem
	}
	return authority
}
