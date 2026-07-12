package luaskills

// Authority is the host-injected authority used by query and system management entrypoints.
// Authority 是查询与 system 管理入口使用的宿主注入权限。
type Authority string

const (
	// AuthoritySystem may manage the ROOT layer when used with system entrypoints.
	// AuthoritySystem 搭配 system 入口时可以管理 ROOT 层。
	AuthoritySystem Authority = "system"
	// AuthorityDelegatedTool follows ordinary user-facing visibility and management boundaries.
	// AuthorityDelegatedTool 遵守普通用户侧可见性与管理边界。
	AuthorityDelegatedTool Authority = "delegated_tool"
)

// SkillInstallSourceType is the managed source type used by install and update requests.
// SkillInstallSourceType 是 install 与 update 请求使用的受管来源类型。
type SkillInstallSourceType string

const (
	// SkillInstallSourceGithub resolves one managed skill from GitHub release metadata.
	// SkillInstallSourceGithub 通过 GitHub release 元数据解析受管 skill。
	SkillInstallSourceGithub SkillInstallSourceType = "github"
	// SkillInstallSourceOfficialHub resolves one managed skill from the configured official LuaSkills Hub.
	// SkillInstallSourceOfficialHub 通过已配置的官方 LuaSkills Hub 解析受管 skill。
	SkillInstallSourceOfficialHub SkillInstallSourceType = "official_hub"
	// SkillInstallSourceURL resolves one managed skill from one remote source descriptor URL.
	// SkillInstallSourceURL 通过远程 source 描述文件 URL 解析受管 skill。
	SkillInstallSourceURL SkillInstallSourceType = "url"
	// SkillInstallSourcePrivateURLManifest resolves one managed skill from a host-private URL manifest.
	// SkillInstallSourcePrivateURLManifest 通过宿主私有 URL manifest 解析受管 skill。
	SkillInstallSourcePrivateURLManifest SkillInstallSourceType = "private_url_manifest"
)

// RuntimeSkillRoot is one named runtime skill root in the formal ROOT, PROJECT, USER chain.
// RuntimeSkillRoot 是正式 ROOT、PROJECT、USER 链中的单个命名运行时 skill root。
type RuntimeSkillRoot struct {
	Name      string `json:"name"`
	SkillsDir string `json:"skills_dir"`
}

// InvocationContext is the optional context injected into call_skill and run_lua.
// InvocationContext 是注入 call_skill 与 run_lua 的可选上下文。
type InvocationContext struct {
	RequestContext any `json:"request_context,omitempty"`
	ClientBudget   any `json:"client_budget,omitempty"`
	ToolConfig     any `json:"tool_config,omitempty"`
}

// SkillInstallRequest is one managed install or update request.
// SkillInstallRequest 是单个受管安装或更新请求。
type SkillInstallRequest struct {
	SkillID    *string                `json:"skill_id,omitempty"`
	Source     *string                `json:"source,omitempty"`
	SourceType SkillInstallSourceType `json:"source_type,omitempty"`
}

// SkillUninstallOptions controls optional database cleanup after uninstall.
// SkillUninstallOptions 控制卸载后的可选数据库清理。
type SkillUninstallOptions struct {
	RemoveSQLite  bool `json:"remove_sqlite,omitempty"`
	RemoveLanceDB bool `json:"remove_lancedb,omitempty"`
}

// LifecycleOptions carries optional target-root and authority overrides.
// LifecycleOptions 携带可选 target-root 与 authority 覆盖。
type LifecycleOptions struct {
	TargetRoot *RuntimeSkillRoot `json:"target_root,omitempty"`
	Authority  Authority         `json:"authority,omitempty"`
}

// RuntimeHostResult is the optional structured host_result returned by call_skill.
// RuntimeHostResult 是 call_skill 返回的可选结构化 host_result。
type RuntimeHostResult struct {
	// Kind is the stable host-result kind identifier.
	// Kind 是稳定的宿主结果类型标识。
	Kind string `json:"kind"`
	// Payload is the arbitrary JSON payload consumed by the host. Use RuntimeChangeSetPayload when Kind is change_set.
	// Payload 是由宿主消费的任意 JSON 载荷。当 Kind 为 change_set 时应使用 RuntimeChangeSetPayload。
	Payload any `json:"payload"`
}

// RuntimeChangeSetLine is one canonical change-set line record.
// RuntimeChangeSetLine 是单条 canonical change_set 行记录。
type RuntimeChangeSetLine struct {
	// Line is the 1-based file line number.
	// Line 是从 1 开始的文件行号。
	Line int `json:"line"`
	// Content is the exact line content stored for this record.
	// Content 是当前记录保存的精确行内容。
	Content string `json:"content"`
}

// RuntimeChangeSetHunk is one canonical change-set modify hunk.
// RuntimeChangeSetHunk 是单个 canonical change_set modify hunk。
type RuntimeChangeSetHunk struct {
	// Before is the contiguous context immediately before the changed block.
	// Before 是紧贴修改块之前的连续上下文。
	Before string `json:"before"`
	// Delete contains deleted old-file lines in ascending order.
	// Delete 包含按升序排列的旧文件删除行。
	Delete []RuntimeChangeSetLine `json:"delete"`
	// Insert contains inserted new-file lines in ascending order.
	// Insert 包含按升序排列的新文件插入行。
	Insert []RuntimeChangeSetLine `json:"insert"`
	// After is the contiguous context immediately after the changed block.
	// After 是紧贴修改块之后的连续上下文。
	After string `json:"after"`
}

// RuntimeChangeSetDiagnostic is one canonical change-set diagnostic record.
// RuntimeChangeSetDiagnostic 是单条 canonical change_set 诊断记录。
type RuntimeChangeSetDiagnostic struct {
	// Level is the structured diagnostic level.
	// Level 是结构化诊断级别。
	Level string `json:"level"`
	// Message is the human-readable diagnostic message.
	// Message 是人类可读诊断消息。
	Message string `json:"message"`
}

// RuntimeChangeSetFile is one canonical change-set file record.
// RuntimeChangeSetFile 是单个 canonical change_set 文件记录。
type RuntimeChangeSetFile struct {
	// Change is the file lifecycle change kind.
	// Change 是文件生命周期变更类型。
	Change string `json:"change"`
	// Path is the absolute file path used by create, modify, and delete records.
	// Path 是 create、modify、delete 记录使用的绝对文件路径。
	Path string `json:"path,omitempty"`
	// OldPath is the absolute old path used by rename records.
	// OldPath 是 rename 记录使用的旧绝对路径。
	OldPath string `json:"old_path,omitempty"`
	// NewPath is the absolute new path used by rename records.
	// NewPath 是 rename 记录使用的新绝对路径。
	NewPath string `json:"new_path,omitempty"`
	// Content is the full-file content used by create and delete records.
	// Content 是 create 与 delete 记录使用的整文件内容。
	Content string `json:"content,omitempty"`
	// Hunks contains explicit modify hunks used by modify records.
	// Hunks 包含 modify 记录使用的显式修改 hunk 列表。
	Hunks []RuntimeChangeSetHunk `json:"hunks,omitempty"`
	// Patch is one optional human-readable patch mirror.
	// Patch 是可选的人类可读 patch 镜像。
	Patch *string `json:"patch,omitempty"`
}

// RuntimeChangeSetPayload is the canonical change-set payload consumed by IDE-aware hosts.
// RuntimeChangeSetPayload 是 IDE 感知宿主消费的 canonical change_set 载荷。
type RuntimeChangeSetPayload struct {
	// Mode reports whether the result is preview or applied.
	// Mode 表示当前结果是预览态还是已应用态。
	Mode string `json:"mode"`
	// Summary is the optional high-level change summary.
	// Summary 是可选的高层变更摘要。
	Summary *string `json:"summary,omitempty"`
	// Files contains the required file lifecycle records.
	// Files 包含必填的文件生命周期记录列表。
	Files []RuntimeChangeSetFile `json:"files"`
	// Diagnostics contains optional diagnostics returned alongside the change-set.
	// Diagnostics 包含随 change_set 一并返回的可选诊断列表。
	Diagnostics []RuntimeChangeSetDiagnostic `json:"diagnostics,omitempty"`
}

// RuntimeInvocationResult is the JSON FFI result returned by call_skill.
// RuntimeInvocationResult 是 call_skill 返回的 JSON FFI 结果。
type RuntimeInvocationResult struct {
	Content      string             `json:"content"`
	OverflowMode *string            `json:"overflow_mode"`
	TemplateHint *string            `json:"template_hint"`
	ContentBytes int                `json:"content_bytes"`
	ContentLines int                `json:"content_lines"`
	HostResult   *RuntimeHostResult `json:"host_result"`
}

// ManagedRuntimeKind identifies the host-selected managed interpreter family.
// ManagedRuntimeKind 标识宿主选择的受管解释器类型。
type ManagedRuntimeKind string

const (
	// ManagedRuntimeKindPython selects one managed CPython installation.
	// ManagedRuntimeKindPython 选择一个受管 CPython 安装。
	ManagedRuntimeKindPython ManagedRuntimeKind = "python"
	// ManagedRuntimeKindNode selects one managed Node.js installation.
	// ManagedRuntimeKindNode 选择一个受管 Node.js 安装。
	ManagedRuntimeKindNode ManagedRuntimeKind = "node"
)

// ManagedRuntimeConfig is the host-selected managed Worker and persistent-session resource policy.
// ManagedRuntimeConfig 是宿主选择的受管 Worker 与持久会话资源策略。
type ManagedRuntimeConfig struct {
	// WorkerPoolMaxSizePerEnvironment is the maximum live Worker count for one exact pool key.
	// WorkerPoolMaxSizePerEnvironment 是单个精确池键的最大活动 Worker 数量。
	WorkerPoolMaxSizePerEnvironment int `json:"worker_pool_max_size_per_environment"`
	// WorkerIdleTTLSecs is the idle lifetime before one unused Worker may be retired.
	// WorkerIdleTTLSecs 是未使用 Worker 可被回收前的空闲秒数。
	WorkerIdleTTLSecs uint64 `json:"worker_idle_ttl_secs"`
	// PersistentSessionLimitPerEngine bounds launching and live sessions owned by one engine.
	// PersistentSessionLimitPerEngine 限制单个引擎拥有的启动中与活动会话数量。
	PersistentSessionLimitPerEngine int `json:"persistent_session_limit_per_engine"`
	// PersistentSessionDefaultBufferLimitBytesPerStream is the omitted session.open stream limit.
	// PersistentSessionDefaultBufferLimitBytesPerStream 是 session.open 省略时的每流缓冲上限。
	PersistentSessionDefaultBufferLimitBytesPerStream int `json:"persistent_session_default_buffer_limit_bytes_per_stream"`
	// InvokeDefaultTimeoutMS is the omitted invoke timeout; nil means unlimited.
	// InvokeDefaultTimeoutMS 是 invoke 省略时的超时；nil 表示无限制。
	InvokeDefaultTimeoutMS *uint64 `json:"invoke_default_timeout_ms"`
}

// ManagedRuntimeResolveOptions identifies one host-shared managed runtime installation.
// ManagedRuntimeResolveOptions 标识一个由宿主共享的受管运行时安装。
type ManagedRuntimeResolveOptions struct {
	// DistributionRoot is the existing absolute root that directly contains python and node.
	// DistributionRoot 是直接包含 python 与 node 的现有绝对根。
	DistributionRoot string
	// Runtime is the exact managed interpreter family.
	// Runtime 是精确受管解释器类型。
	Runtime ManagedRuntimeKind
	// Version is the exact semantic runtime version.
	// Version 是精确语义化运行时版本。
	Version string
	// Platform is the exact normalized LuaSkills platform key.
	// Platform 是精确规范化 LuaSkills 平台键。
	Platform string
}

// ManagedRuntimeInstallDescriptor is one runtime installation validated by LuaSkills.
// ManagedRuntimeInstallDescriptor 是一份经 LuaSkills 校验的运行时安装描述符。
type ManagedRuntimeInstallDescriptor struct {
	// Runtime is the exact managed interpreter family.
	// Runtime 是精确受管解释器类型。
	Runtime ManagedRuntimeKind `json:"runtime"`
	// Version is the exact semantic runtime version.
	// Version 是精确语义化运行时版本。
	Version string `json:"version"`
	// Platform is the normalized LuaSkills platform key.
	// Platform 是规范化 LuaSkills 平台键。
	Platform string `json:"platform"`
	// InstallRoot is the canonical installation directory.
	// InstallRoot 是规范安装目录。
	InstallRoot string `json:"install_root"`
	// Executable is the canonical interpreter executable path.
	// Executable 是规范解释器可执行文件路径。
	Executable string `json:"executable"`
	// ManifestHash is the SHA-256 digest of runtime-manifest.json.
	// ManifestHash 是 runtime-manifest.json 的 SHA-256 摘要。
	ManifestHash string `json:"manifest_hash"`
	// ExecutableHash is the SHA-256 digest of the interpreter executable.
	// ExecutableHash 是解释器可执行文件的 SHA-256 摘要。
	ExecutableHash string `json:"executable_hash"`
}

// ClientOptions controls creation of one LuaSkills client and native engine.
// ClientOptions 控制单个 LuaSkills 客户端与原生引擎的创建。
type ClientOptions struct {
	RuntimeRoot         string
	EngineOptions       map[string]any
	HostOptions         map[string]any
	PoolConfig          map[string]any
	EnsureRuntimeLayout bool
}
