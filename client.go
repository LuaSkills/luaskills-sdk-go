package luaskills

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Client is one high-level LuaSkills SDK client over the public JSON FFI surface.
// Client 是基于公共 JSON FFI 表面的高级 LuaSkills SDK 客户端。
type Client struct {
	// mu guards lifecycle state for the native engine handle.
	// mu 保护原生引擎句柄的生命周期状态。
	mu sync.Mutex
	// stateChanged wakes callers waiting for active FFI calls or close attempts to finish.
	// stateChanged 唤醒等待活跃 FFI 调用或关闭动作完成的调用方。
	stateChanged *sync.Cond
	// engineID is the immutable native engine handle identifier.
	// engineID 是不可变的原生引擎句柄标识符。
	engineID uint64
	// Config exposes the skill config namespace bound to this client.
	// Config 暴露绑定到此客户端的 skill 配置命名空间。
	Config *ConfigClient
	// Skills exposes the skill lifecycle namespace bound to this client.
	// Skills 暴露绑定到此客户端的 skill 生命周期命名空间。
	Skills *SkillManagementClient
	// activeCalls counts FFI calls that currently own the engine handle.
	// activeCalls 统计当前正在持有引擎句柄的 FFI 调用数量。
	activeCalls int
	// closing records whether one Close call is waiting for active calls or freeing the handle.
	// closing 记录是否已有 Close 调用正在等待活跃调用或释放句柄。
	closing bool
	// closed records whether the native engine handle has been released.
	// closed 记录原生引擎句柄是否已经释放。
	closed bool
}

// NewClient creates one native LuaSkills engine and wraps it in a high-level client.
// NewClient 创建一个原生 LuaSkills 引擎并封装为高级客户端。
func NewClient(options ClientOptions) (*Client, error) {
	runtimeRoot := options.RuntimeRoot
	if runtimeRoot == "" {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		runtimeRoot = filepath.Join(workingDirectory, "luaskills-runtime")
	}
	engineOptions := options.EngineOptions
	if engineOptions == nil {
		var err error
		engineOptions, err = CreateEngineOptions(runtimeRoot, options.HostOptions, options.PoolConfig)
		if err != nil {
			return nil, err
		}
		if options.EnsureRuntimeLayout {
			if err := EnsureRuntimeLayout(runtimeRoot, nil); err != nil {
				return nil, err
			}
		}
	}
	var handle struct {
		EngineID uint64 `json:"engine_id"`
	}
	if err := callJSON("luaskills_ffi_engine_new_json", map[string]any{"options": engineOptions}, &handle); err != nil {
		return nil, err
	}
	client := &Client{
		engineID: handle.EngineID,
	}
	client.stateChanged = sync.NewCond(&client.mu)
	client.Config = &ConfigClient{client: client}
	client.Skills = &SkillManagementClient{client: client}
	return client, nil
}

// EngineID returns the immutable native engine handle identifier.
// EngineID 返回不可变的原生引擎句柄标识符。
func (c *Client) EngineID() uint64 {
	return c.engineID
}

// System returns one system-management namespace bound to host-injected authority.
// System 返回绑定到宿主注入权限的 system 管理命名空间。
func (c *Client) System(authority Authority) *SystemSkillManagementClient {
	if authority == "" {
		authority = AuthoritySystem
	}
	return &SystemSkillManagementClient{
		SkillManagementClient: SkillManagementClient{
			client:      c,
			systemPlane: true,
			authority:   authority,
		},
	}
}

// PollManagedSessionEvents destructively drains one bounded engine-level managed-session event batch.
// PollManagedSessionEvents 以破坏性方式排空一批有界的引擎级受管会话事件。
func (c *Client) PollManagedSessionEvents(maxEvents int, authority Authority) (map[string]any, error) {
	if maxEvents <= 0 {
		return nil, fmt.Errorf("max_events must be positive")
	}
	if authority == "" {
		authority = AuthoritySystem
	}
	var result map[string]any
	err := c.call("luaskills_ffi_managed_session_events_poll_json", map[string]any{
		"engine_id":  c.engineID,
		"max_events": maxEvents,
		"authority":  authority,
	}, &result)
	return result, err
}

// WaitManagedSessionEvents waits for and destructively drains one bounded engine-level event batch.
// WaitManagedSessionEvents 等待并以破坏性方式排空一批有界的引擎级事件。
func (c *Client) WaitManagedSessionEvents(maxEvents int, timeoutMS uint64, authority Authority) (map[string]any, error) {
	if maxEvents <= 0 {
		return nil, fmt.Errorf("max_events must be positive")
	}
	if authority == "" {
		authority = AuthoritySystem
	}
	var result map[string]any
	err := c.call("luaskills_ffi_managed_session_events_wait_json", map[string]any{
		"engine_id":  c.engineID,
		"max_events": maxEvents,
		"timeout_ms": timeoutMS,
		"authority":  authority,
	}, &result)
	return result, err
}

// SetManagedSessionWakeCallback registers, replaces, or clears this engine's wake callback.
// SetManagedSessionWakeCallback 注册、替换或清除当前引擎的唤醒回调。
func (c *Client) SetManagedSessionWakeCallback(callback ManagedSessionWakeCallback) error {
	return SetManagedSessionWakeCallback(c.engineID, callback)
}

// RuntimeLeases returns one runtime-lease namespace over the public JSON FFI surface.
// RuntimeLeases 返回一个基于公共 JSON FFI 接口的运行时租约命名空间。
func (c *Client) RuntimeLeases() *RuntimeLeaseClient {
	return &RuntimeLeaseClient{
		client: c,
	}
}

// LoadFromRoots loads skills from the formal ordered root chain.
// LoadFromRoots 从正式有序 root 链加载 skills。
func (c *Client) LoadFromRoots(skillRoots []RuntimeSkillRoot) (map[string]any, error) {
	var result map[string]any
	err := c.call("luaskills_ffi_load_from_roots_json", map[string]any{
		"engine_id":   c.engineID,
		"skill_roots": skillRoots,
	}, &result)
	return result, err
}

// ReloadFromRoots reloads skills from the formal ordered root chain.
// ReloadFromRoots 从正式有序 root 链重载 skills。
func (c *Client) ReloadFromRoots(skillRoots []RuntimeSkillRoot) (map[string]any, error) {
	var result map[string]any
	err := c.call("luaskills_ffi_reload_from_roots_json", map[string]any{
		"engine_id":   c.engineID,
		"skill_roots": skillRoots,
	}, &result)
	return result, err
}

// ListEntries lists runtime entries visible to the selected authority.
// ListEntries 列出指定权限可见的运行时入口。
func (c *Client) ListEntries(authority Authority) ([]map[string]any, error) {
	if authority == "" {
		authority = AuthorityDelegatedTool
	}
	var result []map[string]any
	err := c.call("luaskills_ffi_list_entries_json", map[string]any{
		"engine_id": c.engineID,
		"authority": authority,
	}, &result)
	return result, err
}

// ListSkillHelp lists runtime help trees visible to the selected authority.
// ListSkillHelp 列出指定权限可见的运行时帮助树。
func (c *Client) ListSkillHelp(authority Authority) ([]map[string]any, error) {
	if authority == "" {
		authority = AuthorityDelegatedTool
	}
	var result []map[string]any
	err := c.call("luaskills_ffi_list_skill_help_json", map[string]any{
		"engine_id": c.engineID,
		"authority": authority,
	}, &result)
	return result, err
}

// RenderSkillHelpDetail renders one help flow detail visible to the selected authority.
// RenderSkillHelpDetail 渲染指定权限可见的单个帮助流程详情。
func (c *Client) RenderSkillHelpDetail(skillID string, flowName string, authority Authority, requestContext any) (map[string]any, error) {
	if flowName == "" {
		flowName = "main"
	}
	if authority == "" {
		authority = AuthorityDelegatedTool
	}
	var result map[string]any
	err := c.call("luaskills_ffi_render_skill_help_detail_json", map[string]any{
		"engine_id":       c.engineID,
		"skill_id":        skillID,
		"flow_name":       flowName,
		"request_context": requestContext,
		"authority":       authority,
	}, &result)
	return result, err
}

// PromptArgumentCompletions queries prompt argument completions visible to the selected authority.
// PromptArgumentCompletions 查询指定权限可见的 prompt 参数补全项。
func (c *Client) PromptArgumentCompletions(promptName string, argumentName string, authority Authority) ([]string, error) {
	if authority == "" {
		authority = AuthorityDelegatedTool
	}
	var result []string
	err := c.call("luaskills_ffi_prompt_argument_completions_json", map[string]any{
		"engine_id":     c.engineID,
		"prompt_name":   promptName,
		"argument_name": argumentName,
		"authority":     authority,
	}, &result)
	return result, err
}

// IsSkill returns whether one canonical tool name is visible as a skill entry.
// IsSkill 返回指定 canonical 工具名是否可见为 skill 入口。
func (c *Client) IsSkill(toolName string, authority Authority) (bool, error) {
	if authority == "" {
		authority = AuthorityDelegatedTool
	}
	var result struct {
		Value bool `json:"value"`
	}
	err := c.call("luaskills_ffi_is_skill_json", map[string]any{
		"engine_id": c.engineID,
		"tool_name": toolName,
		"authority": authority,
	}, &result)
	return result.Value, err
}

// SkillNameForTool resolves the owning skill id for one visible canonical tool name.
// SkillNameForTool 解析单个可见 canonical 工具名称所属的 skill id。
func (c *Client) SkillNameForTool(toolName string, authority Authority) (*string, error) {
	if authority == "" {
		authority = AuthorityDelegatedTool
	}
	var result struct {
		SkillID *string `json:"skill_id"`
	}
	err := c.call("luaskills_ffi_skill_name_for_tool_json", map[string]any{
		"engine_id": c.engineID,
		"tool_name": toolName,
		"authority": authority,
	}, &result)
	return result.SkillID, err
}

// CallSkill calls one active skill entry by canonical tool name.
// CallSkill 按 canonical 工具名称调用单个已激活 skill 入口。
func (c *Client) CallSkill(toolName string, args any, invocationContext *InvocationContext) (*RuntimeInvocationResult, error) {
	if args == nil {
		args = map[string]any{}
	}
	var result RuntimeInvocationResult
	err := c.call("luaskills_ffi_call_skill_json", map[string]any{
		"engine_id":          c.engineID,
		"tool_name":          toolName,
		"args":               args,
		"invocation_context": normalizeInvocationContext(invocationContext),
	}, &result)
	return &result, err
}

// RunLua executes one inline Lua snippet against the active runtime.
// RunLua 针对当前活动运行时执行单段内联 Lua。
func (c *Client) RunLua(code string, args any, invocationContext *InvocationContext) (any, error) {
	if args == nil {
		args = map[string]any{}
	}
	var result any
	err := c.call("luaskills_ffi_run_lua_json", map[string]any{
		"engine_id":          c.engineID,
		"code":               code,
		"args":               args,
		"invocation_context": normalizeInvocationContext(invocationContext),
	}, &result)
	return result, err
}

// Close releases the native engine handle.
// Close 释放原生引擎句柄。
func (c *Client) Close() (map[string]any, error) {
	engineID, shouldClose := c.beginClose()
	if !shouldClose {
		return nil, nil
	}
	var result map[string]any
	callbackErr := SetManagedSessionWakeCallback(engineID, nil)
	freeErr := callJSON("luaskills_ffi_engine_free_json", map[string]any{"engine_id": engineID}, &result)
	c.finishClose(freeErr)
	if err := errors.Join(callbackErr, freeErr); err != nil {
		return nil, err
	}
	return result, nil
}

// call invokes one JSON FFI function after checking the engine state.
// call 检查引擎状态后调用一个 JSON FFI 函数。
func (c *Client) call(functionName string, payload any, out any) error {
	if err := c.beginCall(); err != nil {
		return err
	}
	defer c.endCall()
	return callJSON(functionName, payload, out)
}

// beginCall reserves the native engine handle for one FFI dispatch.
// beginCall 为单次 FFI 分发保留原生引擎句柄。
func (c *Client) beginCall() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("LuaSkills engine %d is already closed", c.engineID)
	}
	if c.closing {
		return fmt.Errorf("LuaSkills engine %d is closing", c.engineID)
	}
	c.activeCalls++
	return nil
}

// endCall releases one active FFI call reservation and wakes pending close calls.
// endCall 释放一个活跃 FFI 调用占用并唤醒等待中的关闭调用。
func (c *Client) endCall() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.activeCalls--
	if c.activeCalls == 0 {
		c.condition().Broadcast()
	}
}

// beginClose starts the exclusive close phase after all active FFI calls finish.
// beginClose 在所有活跃 FFI 调用结束后启动独占关闭阶段。
func (c *Client) beginClose() (uint64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	condition := c.condition()
	for c.closing {
		condition.Wait()
	}
	if c.closed {
		return 0, false
	}
	c.closing = true
	for c.activeCalls > 0 {
		condition.Wait()
	}
	return c.engineID, true
}

// finishClose completes the close phase and publishes the final lifecycle state.
// finishClose 完成关闭阶段并发布最终生命周期状态。
func (c *Client) finishClose(closeErr error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if closeErr == nil {
		c.closed = true
	}
	c.closing = false
	c.condition().Broadcast()
}

// condition returns the lifecycle condition variable, lazily creating it for zero-value tests.
// condition 返回生命周期条件变量，并为零值测试场景按需创建。
func (c *Client) condition() *sync.Cond {
	if c.stateChanged == nil {
		c.stateChanged = sync.NewCond(&c.mu)
	}
	return c.stateChanged
}

// CreateEngineOptions builds complete engine options from SDK defaults and caller overrides.
// CreateEngineOptions 基于 SDK 默认值和调用方覆盖构造完整引擎选项。
func CreateEngineOptions(runtimeRoot string, hostOptions map[string]any, poolConfig map[string]any) (map[string]any, error) {
	defaultHostOptions, err := DefaultHostOptions(runtimeRoot)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"pool_config":  mergeMaps(DefaultPoolConfig(), poolConfig),
		"host_options": mergeHostOptions(defaultHostOptions, hostOptions),
	}, nil
}

// DefaultPoolConfig returns the SDK default VM pool configuration.
// DefaultPoolConfig 返回 SDK 默认虚拟机池配置。
func DefaultPoolConfig() map[string]any {
	return map[string]any{"min_size": 1, "max_size": 4, "idle_ttl_secs": 60}
}

// DefaultHostOptions returns the SDK default host options for one runtime root.
// DefaultHostOptions 返回单个 runtime root 对应的 SDK 默认宿主选项。
func DefaultHostOptions(runtimeRoot string) (map[string]any, error) {
	root := normalizePath(runtimeRoot)
	baseOptions := map[string]any{
		"runtime_root":            root,
		"temp_dir":                nil,
		"resources_dir":           nil,
		"lua_packages_dir":        nil,
		"host_provided_tool_root": nil,
		"host_provided_lua_root":  nil,
		"host_provided_ffi_root":  nil,
		"system_lua_lib_dir":      nil,
		"download_cache_root":     nil,
		"dependency_dir_name":     "",
		"state_dir_name":          "",
		"database_dir_name":       "",
		"skill_config_file_path":  nil,
		"allow_network_download":  true,
		"github_base_url":         nil,
		"github_api_base_url":     nil,
		"sqlite_library_path":     nil,
		"sqlite_provider_mode":    "dynamic_library",
		"sqlite_callback_mode":    "standard",
		"lancedb_library_path":    nil,
		"lancedb_provider_mode":   "dynamic_library",
		"lancedb_callback_mode":   "standard",
		"space_controller":        DefaultSpaceControllerOptions(),
		"cache_config":            nil,
		"runlua_pool_config":      nil,
		"reserved_entry_names":    []string{},
		"ignored_skill_ids":       []string{},
		"capabilities": map[string]any{
			"enable_skill_management_bridge": false,
			"enable_managed_io_compat":       true,
		},
	}
	manifest, err := LoadRuntimeInstallManifest(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return baseOptions, nil
		}
		return nil, err
	}
	manifestOptions, err := HostOptionsFromRuntimeManifest(manifest)
	if err != nil {
		return nil, err
	}
	return mergeHostOptions(baseOptions, manifestOptions), nil
}

// DefaultSpaceControllerOptions returns the SDK default space-controller options.
// DefaultSpaceControllerOptions 返回 SDK 默认 space-controller 选项。
func DefaultSpaceControllerOptions() map[string]any {
	return map[string]any{
		"endpoint":                  nil,
		"auto_spawn":                false,
		"executable_path":           nil,
		"process_mode":              "managed",
		"minimum_uptime_secs":       300,
		"idle_timeout_secs":         900,
		"default_lease_ttl_secs":    120,
		"connect_timeout_secs":      5,
		"startup_timeout_secs":      15,
		"startup_retry_interval_ms": 250,
		"lease_renew_interval_secs": 30,
	}
}

// mergeHostOptions merges caller host overrides over complete default host options.
// mergeHostOptions 将调用方宿主覆盖合并到完整默认宿主选项上。
func mergeHostOptions(base map[string]any, overrides map[string]any) map[string]any {
	merged := mergeMaps(base, overrides)
	if value, ok := overrides["space_controller"].(map[string]any); ok {
		merged["space_controller"] = mergeMaps(base["space_controller"].(map[string]any), value)
	}
	if value, ok := overrides["capabilities"].(map[string]any); ok {
		merged["capabilities"] = mergeMaps(base["capabilities"].(map[string]any), value)
	}
	return merged
}

// mergeMaps returns one shallow copy of base merged with overrides.
// mergeMaps 返回 base 与 overrides 合并后的浅拷贝。
func mergeMaps(base map[string]any, overrides map[string]any) map[string]any {
	merged := make(map[string]any, len(base)+len(overrides))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range overrides {
		merged[key] = value
	}
	return merged
}

// normalizeInvocationContext converts one optional invocation context into JSON payload form.
// normalizeInvocationContext 将单个可选调用上下文转换为 JSON 载荷形式。
func normalizeInvocationContext(context *InvocationContext) map[string]any {
	if context == nil {
		return nil
	}
	clientBudget := context.ClientBudget
	if clientBudget == nil {
		clientBudget = map[string]any{}
	}
	toolConfig := context.ToolConfig
	if toolConfig == nil {
		toolConfig = map[string]any{}
	}
	return map[string]any{
		"request_context": context.RequestContext,
		"client_budget":   clientBudget,
		"tool_config":     toolConfig,
	}
}
