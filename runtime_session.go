package luaskills

import "fmt"

// RuntimeLeaseIdentity is the stable runtime-lease identity payload persisted by SDK hosts.
// RuntimeLeaseIdentity 是由 SDK 宿主持久化的稳定运行时租约身份载荷。
type RuntimeLeaseIdentity struct {
	LeaseID    string `json:"lease_id"`
	SID        string `json:"sid"`
	Generation int    `json:"generation"`
}

// RuntimeLeaseCreateOptions configures one runtime-lease creation request.
// RuntimeLeaseCreateOptions 配置单个运行时租约创建请求。
type RuntimeLeaseCreateOptions struct {
	TTLSec        *int
	CWD           *string
	WorkspaceRoot *string
	LuaRoots      []string
	CRoots        []string
	Mounts        any
	SystemPackage *SystemRuntimePackage
}

// SystemRuntimePackage identifies one trusted System Plugin package for a System lease.
// SystemRuntimePackage 标识 System 租约使用的可信 System Plugin 包。
type SystemRuntimePackage struct {
	ID               string `json:"id"`
	Root             string `json:"root"`
	DependenciesFile string `json:"dependencies_file"`
}

// RuntimeLeaseAction is one supported runtime-lease JSON FFI action.
// RuntimeLeaseAction 是一个受支持的运行时租约 JSON FFI 动作。
type RuntimeLeaseAction string

const (
	// RuntimeLeaseCreateAction creates or replaces one runtime lease.
	// RuntimeLeaseCreateAction 创建或替换单个运行时租约。
	RuntimeLeaseCreateAction RuntimeLeaseAction = "create"
	// RuntimeLeaseEvalAction evaluates one Lua chunk inside one runtime lease.
	// RuntimeLeaseEvalAction 在单个运行时租约内执行一段 Lua 代码。
	RuntimeLeaseEvalAction RuntimeLeaseAction = "eval"
	// RuntimeLeaseStatusAction reads one runtime lease status.
	// RuntimeLeaseStatusAction 读取单个运行时租约状态。
	RuntimeLeaseStatusAction RuntimeLeaseAction = "status"
	// RuntimeLeaseListAction lists active runtime leases.
	// RuntimeLeaseListAction 列出活跃运行时租约。
	RuntimeLeaseListAction RuntimeLeaseAction = "list"
	// RuntimeLeaseCloseAction closes one runtime lease.
	// RuntimeLeaseCloseAction 关闭单个运行时租约。
	RuntimeLeaseCloseAction RuntimeLeaseAction = "close"
)

// RuntimeLeaseClient is the stateful runtime-lease namespace over the JSON FFI runtime-lease entrypoints.
// RuntimeLeaseClient 是覆盖 JSON FFI 运行时租约入口的有状态运行时租约命名空间。
type RuntimeLeaseClient struct {
	client        *Client
	authority     Authority
	bindAuthority bool
}

// CallRaw dispatches one raw runtime-lease JSON request without applying success checks.
// CallRaw 分发单个原始运行时租约 JSON 请求而不附加成功校验。
func (c *RuntimeLeaseClient) CallRaw(action RuntimeLeaseAction, payload map[string]any) (map[string]any, error) {
	functionName, err := c.runtimeLeaseFunctionName(action)
	if err != nil {
		return nil, err
	}
	requestPayload := mergeMaps(payload, map[string]any{
		"engine_id": c.client.engineID,
	})
	if c.bindAuthority {
		requestPayload["authority"] = c.authority
	}
	var result map[string]any
	if err := c.client.call(functionName, requestPayload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Create creates or replaces one persistent runtime lease.
// Create 创建或替换一个持久运行时租约。
func (c *RuntimeLeaseClient) Create(sid string, ttlSec int, replace bool) (map[string]any, error) {
	var options *RuntimeLeaseCreateOptions
	if ttlSec > 0 {
		options = &RuntimeLeaseCreateOptions{TTLSec: &ttlSec}
	}
	return c.CreateWithOptions(sid, replace, options)
}

// CreateWithOptions creates or replaces one persistent runtime lease with explicit host-owned path options.
// CreateWithOptions 使用显式宿主路径选项创建或替换一个持久运行时租约。
func (c *RuntimeLeaseClient) CreateWithOptions(
	sid string,
	replace bool,
	options *RuntimeLeaseCreateOptions,
) (map[string]any, error) {
	payload := map[string]any{
		"sid":     sid,
		"replace": replace,
	}
	if options != nil {
		if c.bindAuthority && (len(options.LuaRoots) > 0 || len(options.CRoots) > 0) {
			return nil, fmt.Errorf("system runtime lease create does not accept lua_roots or c_roots")
		}
		if options.TTLSec != nil && *options.TTLSec > 0 {
			payload["ttl_sec"] = *options.TTLSec
		}
		if options.CWD != nil && *options.CWD != "" {
			payload["cwd"] = *options.CWD
		}
		if options.WorkspaceRoot != nil && *options.WorkspaceRoot != "" {
			payload["workspace_root"] = *options.WorkspaceRoot
		}
		if len(options.LuaRoots) > 0 {
			payload["lua_roots"] = options.LuaRoots
		}
		if len(options.CRoots) > 0 {
			payload["c_roots"] = options.CRoots
		}
		if options.Mounts != nil {
			payload["mounts"] = options.Mounts
		}
		if c.bindAuthority && options.SystemPackage != nil {
			payload["system_package"] = options.SystemPackage
		}
	}
	if c.bindAuthority {
		if options == nil || options.SystemPackage == nil {
			return nil, fmt.Errorf("system runtime lease create requires system_package")
		}
		if options.SystemPackage.ID == "" || options.SystemPackage.Root == "" || options.SystemPackage.DependenciesFile == "" {
			return nil, fmt.Errorf("system_package requires id, root, and dependencies_file")
		}
	}
	result, err := c.CallRaw(RuntimeLeaseCreateAction, payload)
	if err != nil {
		return nil, err
	}
	return requireRuntimeLeaseOK(result, "runtime lease create")
}

// CreateHandle creates one runtime-lease handle object from one fresh create response.
// CreateHandle 基于一份新的 create 响应创建一个运行时租约句柄对象。
func (c *RuntimeLeaseClient) CreateHandle(sid string, ttlSec int, replace bool) (*RuntimeLeaseHandle, error) {
	var options *RuntimeLeaseCreateOptions
	if ttlSec > 0 {
		options = &RuntimeLeaseCreateOptions{TTLSec: &ttlSec}
	}
	return c.CreateHandleWithOptions(sid, replace, options)
}

// CreateHandleWithOptions creates one runtime-lease handle object from one fresh create response with explicit path options.
// CreateHandleWithOptions 使用显式路径选项基于新的 create 响应创建一个运行时租约句柄对象。
func (c *RuntimeLeaseClient) CreateHandleWithOptions(
	sid string,
	replace bool,
	options *RuntimeLeaseCreateOptions,
) (*RuntimeLeaseHandle, error) {
	result, err := c.CreateWithOptions(sid, replace, options)
	if err != nil {
		return nil, err
	}
	return BindRuntimeLeaseHandle(c, result)
}

// BindHandle rebuilds one runtime-lease handle object from one persisted payload.
// BindHandle 基于一份已持久化载荷重建一个运行时租约句柄对象。
func (c *RuntimeLeaseClient) BindHandle(payload map[string]any) (*RuntimeLeaseHandle, error) {
	return BindRuntimeLeaseHandle(c, payload)
}

// Eval evaluates one Lua chunk inside one persistent runtime lease.
// Eval 在一个持久运行时租约中执行单个 Lua 代码块。
func (c *RuntimeLeaseClient) Eval(leaseID string, code string, args map[string]any, timeoutMs int, sid string, generation int) (map[string]any, error) {
	return c.EvalWithContext(leaseID, code, args, timeoutMs, sid, generation, nil)
}

// EvalWithContext evaluates one Lua chunk while attaching optional invocation context.
// EvalWithContext 执行单个 Lua 代码块并附带可选调用上下文。
func (c *RuntimeLeaseClient) EvalWithContext(
	leaseID string,
	code string,
	args map[string]any,
	timeoutMs int,
	sid string,
	generation int,
	invocationContext *InvocationContext,
) (map[string]any, error) {
	if args == nil {
		args = map[string]any{}
	}
	if timeoutMs == 0 {
		timeoutMs = 60000
	}
	payload := map[string]any{
		"lease_id":           leaseID,
		"code":               code,
		"args":               args,
		"timeout_ms":         timeoutMs,
		"invocation_context": normalizeInvocationContext(invocationContext),
	}
	if sid != "" {
		payload["sid"] = sid
	}
	if generation != 0 {
		payload["generation"] = generation
	}
	result, err := c.CallRaw(RuntimeLeaseEvalAction, payload)
	if err != nil {
		return nil, err
	}
	return requireRuntimeLeaseOK(result, "runtime lease eval")
}

// Status reads one runtime lease status payload with optional identity guards.
// Status 读取单个运行时租约状态载荷，并可附带可选身份护栏。
func (c *RuntimeLeaseClient) Status(leaseID string, sid string, generation int) (map[string]any, error) {
	payload := map[string]any{
		"lease_id": leaseID,
	}
	if sid != "" {
		payload["sid"] = sid
	}
	if generation != 0 {
		payload["generation"] = generation
	}
	return c.CallRaw(RuntimeLeaseStatusAction, payload)
}

// List lists active runtime leases and optionally filters by one SID.
// List 列出活跃运行时租约，并可按单个 SID 过滤。
func (c *RuntimeLeaseClient) List(sid string) (map[string]any, error) {
	payload := map[string]any{}
	if sid != "" {
		payload["sid"] = sid
	}
	return c.CallRaw(RuntimeLeaseListAction, payload)
}

// ListHandles lists active runtime-lease handles rebuilt from the current lease listing payload.
// ListHandles 基于当前租约列表载荷重建活跃运行时租约句柄列表。
func (c *RuntimeLeaseClient) ListHandles(sid string) ([]*RuntimeLeaseHandle, error) {
	result, err := c.List(sid)
	if err != nil {
		return nil, err
	}
	rawLeases, ok := result["leases"].([]any)
	if !ok {
		return nil, fmt.Errorf("runtime lease list payload is missing the leases array")
	}
	handles := make([]*RuntimeLeaseHandle, 0, len(rawLeases))
	for _, rawLease := range rawLeases {
		lease, err := requireJSONMap(rawLease, "runtime lease entry")
		if err != nil {
			return nil, err
		}
		handle, err := c.BindHandle(lease)
		if err != nil {
			return nil, err
		}
		handles = append(handles, handle)
	}
	return handles, nil
}

// FindHandle returns the first active runtime-lease handle for one SID when present.
// FindHandle 返回某个 SID 的第一个活跃运行时租约句柄（如果存在）。
func (c *RuntimeLeaseClient) FindHandle(sid string) (*RuntimeLeaseHandle, error) {
	handles, err := c.ListHandles(sid)
	if err != nil {
		return nil, err
	}
	if len(handles) == 0 {
		return nil, nil
	}
	return handles[0], nil
}

// Close closes one runtime lease and returns its final status payload with optional identity guards.
// Close 关闭单个运行时租约并返回其最终状态载荷，并可附带可选身份护栏。
func (c *RuntimeLeaseClient) Close(leaseID string, sid string, generation int) (map[string]any, error) {
	payload := map[string]any{
		"lease_id": leaseID,
	}
	if sid != "" {
		payload["sid"] = sid
	}
	if generation != 0 {
		payload["generation"] = generation
	}
	return c.CallRaw(RuntimeLeaseCloseAction, payload)
}

// UsesSystemRuntimeLeaseEndpoints returns whether this helper will dispatch requests to dedicated system runtime-lease entrypoints.
// UsesSystemRuntimeLeaseEndpoints 返回当前辅助器是否会把请求分发到专用 system 运行时租约入口。
func (c *RuntimeLeaseClient) UsesSystemRuntimeLeaseEndpoints() (bool, error) {
	if !c.bindAuthority {
		return false, nil
	}
	return true, nil
}

// runtimeLeaseFunctionName resolves the concrete runtime-lease JSON FFI entrypoint name for one logical action.
// runtimeLeaseFunctionName 为单个逻辑动作解析具体的运行时租约 JSON FFI 入口名称。
func (c *RuntimeLeaseClient) runtimeLeaseFunctionName(action RuntimeLeaseAction) (string, error) {
	actionValue, err := runtimeLeaseActionValue(action)
	if err != nil {
		return "", err
	}
	publicName := "luaskills_ffi_runtime_lease_" + actionValue + "_json"
	if !c.bindAuthority {
		return publicName, nil
	}
	return "luaskills_ffi_system_runtime_lease_" + actionValue + "_json", nil
}

// runtimeLeaseActionValue returns the validated raw action string used by native JSON FFI function names.
// runtimeLeaseActionValue 返回原生 JSON FFI 函数名使用的已验证原始动作字符串。
func runtimeLeaseActionValue(action RuntimeLeaseAction) (string, error) {
	switch action {
	case RuntimeLeaseCreateAction,
		RuntimeLeaseEvalAction,
		RuntimeLeaseStatusAction,
		RuntimeLeaseListAction,
		RuntimeLeaseCloseAction:
		return string(action), nil
	default:
		return "", fmt.Errorf("unsupported runtime lease action: %s", action)
	}
}

// RuntimeLeaseHandle is the stable host-side runtime-lease handle that carries lease identity guards automatically.
// RuntimeLeaseHandle 是自动携带租约身份护栏的稳定宿主侧运行时租约句柄。
type RuntimeLeaseHandle struct {
	sessions   *RuntimeLeaseClient
	LeaseID    string
	SID        string
	Generation int
}

// BindRuntimeLeaseHandle constructs one runtime-lease handle from one payload that contains identity fields.
// BindRuntimeLeaseHandle 从包含身份字段的一份载荷中构造一个运行时租约句柄。
func BindRuntimeLeaseHandle(sessions *RuntimeLeaseClient, payload map[string]any) (*RuntimeLeaseHandle, error) {
	leaseID, err := requireRuntimeLeaseStringField(payload, "lease_id")
	if err != nil {
		return nil, err
	}
	sid, err := requireRuntimeLeaseStringField(payload, "sid")
	if err != nil {
		return nil, err
	}
	generation, err := requireRuntimeLeaseIntField(payload, "generation")
	if err != nil {
		return nil, err
	}
	return &RuntimeLeaseHandle{
		sessions:   sessions,
		LeaseID:    leaseID,
		SID:        sid,
		Generation: generation,
	}, nil
}

// IdentityPayload exports the stable lease identity fields for persistence or raw FFI calls.
// IdentityPayload 导出稳定租约身份字段，供持久化或原始 FFI 调用使用。
func (h *RuntimeLeaseHandle) IdentityPayload() RuntimeLeaseIdentity {
	return RuntimeLeaseIdentity{
		LeaseID:    h.LeaseID,
		SID:        h.SID,
		Generation: h.Generation,
	}
}

// Eval evaluates Lua code while automatically attaching the stored lease identity guards.
// Eval 执行 Lua 代码时自动附带已保存的租约身份护栏。
func (h *RuntimeLeaseHandle) Eval(code string, args map[string]any, timeoutMs int) (map[string]any, error) {
	return h.sessions.Eval(h.LeaseID, code, args, timeoutMs, h.SID, h.Generation)
}

// EvalWithContext evaluates Lua code while automatically attaching the stored lease identity guards and invocation context.
// EvalWithContext 执行 Lua 代码时自动附带已保存的租约身份护栏与调用上下文。
func (h *RuntimeLeaseHandle) EvalWithContext(
	code string,
	args map[string]any,
	timeoutMs int,
	invocationContext *InvocationContext,
) (map[string]any, error) {
	return h.sessions.EvalWithContext(h.LeaseID, code, args, timeoutMs, h.SID, h.Generation, invocationContext)
}

// Status reads the current lease status while automatically attaching the stored identity guards.
// Status 读取当前租约状态时自动附带已保存的身份护栏。
func (h *RuntimeLeaseHandle) Status() (map[string]any, error) {
	return h.sessions.Status(h.LeaseID, h.SID, h.Generation)
}

// Close closes the current lease while automatically attaching the stored identity guards.
// Close 关闭当前租约时自动附带已保存的身份护栏。
func (h *RuntimeLeaseHandle) Close() (map[string]any, error) {
	return h.sessions.Close(h.LeaseID, h.SID, h.Generation)
}

// requireRuntimeLeaseOK requires one runtime-lease payload to report success.
// requireRuntimeLeaseOK 要求单个运行时租约载荷报告成功。
func requireRuntimeLeaseOK(payload map[string]any, action string) (map[string]any, error) {
	if ok, _ := payload["ok"].(bool); ok {
		return payload, nil
	}
	return nil, fmt.Errorf(
		"%s failed: %v: %v",
		action,
		firstNonNil(payload["error_code"], "unknown"),
		firstNonNil(payload["message"], "Unknown runtime lease error"),
	)
}

// requireRuntimeLeaseStringField reads one required runtime-lease string field from one payload object.
// requireRuntimeLeaseStringField 从一份载荷对象中读取一个必填的运行时租约字符串字段。
func requireRuntimeLeaseStringField(payload map[string]any, fieldName string) (string, error) {
	value, ok := payload[fieldName].(string)
	if ok && value != "" {
		return value, nil
	}
	return "", fmt.Errorf("runtime lease payload is missing required string field: %s", fieldName)
}

// requireRuntimeLeaseIntField reads one required runtime-lease integer field from one payload object.
// requireRuntimeLeaseIntField 从一份载荷对象中读取一个必填的运行时租约整数字段。
func requireRuntimeLeaseIntField(payload map[string]any, fieldName string) (int, error) {
	switch value := payload[fieldName].(type) {
	case int:
		return value, nil
	case int32:
		return int(value), nil
	case int64:
		return int(value), nil
	case float64:
		return int(value), nil
	default:
		return 0, fmt.Errorf("runtime lease payload is missing required integer field: %s", fieldName)
	}
}

// requireJSONMap requires one arbitrary JSON value to be one plain object map.
// requireJSONMap 要求某个任意 JSON 值必须是普通对象映射。
func requireJSONMap(value any, context string) (map[string]any, error) {
	payload, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be one JSON object", context)
	}
	return payload, nil
}

// firstNonNil returns the first non-nil fallback display value.
// firstNonNil 返回第一个非 nil 的回退展示值。
func firstNonNil(value any, fallback any) any {
	if value != nil {
		return value
	}
	return fallback
}
