package luaskills

import "fmt"

// RuntimeSessionIdentity is the stable runtime-session identity payload persisted by SDK hosts.
// RuntimeSessionIdentity 是由 SDK 宿主持久化的稳定运行时会话身份载荷。
type RuntimeSessionIdentity struct {
	LeaseID    string `json:"lease_id"`
	SID        string `json:"sid"`
	Generation int    `json:"generation"`
}

// RuntimeSessionClient is the stateful runtime-session namespace over the JSON FFI runtime-session entrypoints.
// RuntimeSessionClient 是覆盖 JSON FFI 运行时会话入口的有状态运行时会话命名空间。
type RuntimeSessionClient struct {
	client        *Client
	authority     Authority
	bindAuthority bool
}

// CallRaw dispatches one raw runtime-session JSON request without applying success checks.
// CallRaw 分发单个原始运行时会话 JSON 请求而不附加成功校验。
func (c *RuntimeSessionClient) CallRaw(action string, payload map[string]any) (map[string]any, error) {
	functionName, err := c.runtimeSessionFunctionName(action)
	if err != nil {
		return nil, err
	}
	requestPayload := mergeMaps(payload, map[string]any{
		"engine_id": c.client.EngineID,
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
func (c *RuntimeSessionClient) Create(sid string, ttlSec int, replace bool) (map[string]any, error) {
	if ttlSec == 0 {
		ttlSec = 600
	}
	result, err := c.CallRaw("create", map[string]any{
		"sid":     sid,
		"ttl_sec": ttlSec,
		"replace": replace,
	})
	if err != nil {
		return nil, err
	}
	return requireRuntimeSessionOK(result, "runtime session create")
}

// CreateHandle creates one runtime-session handle object from one fresh create response.
// CreateHandle 基于一份新的 create 响应创建一个运行时会话句柄对象。
func (c *RuntimeSessionClient) CreateHandle(sid string, ttlSec int, replace bool) (*RuntimeSessionHandle, error) {
	result, err := c.Create(sid, ttlSec, replace)
	if err != nil {
		return nil, err
	}
	return BindRuntimeSessionHandle(c, result)
}

// BindHandle rebuilds one runtime-session handle object from one persisted payload.
// BindHandle 基于一份已持久化载荷重建一个运行时会话句柄对象。
func (c *RuntimeSessionClient) BindHandle(payload map[string]any) (*RuntimeSessionHandle, error) {
	return BindRuntimeSessionHandle(c, payload)
}

// Eval evaluates one Lua chunk inside one persistent runtime lease.
// Eval 在一个持久运行时租约中执行单个 Lua 代码块。
func (c *RuntimeSessionClient) Eval(leaseID string, code string, args map[string]any, timeoutMs int, sid string, generation int) (map[string]any, error) {
	if args == nil {
		args = map[string]any{}
	}
	if timeoutMs == 0 {
		timeoutMs = 60000
	}
	payload := map[string]any{
		"lease_id":   leaseID,
		"code":       code,
		"args":       args,
		"timeout_ms": timeoutMs,
	}
	if sid != "" {
		payload["sid"] = sid
	}
	if generation != 0 {
		payload["generation"] = generation
	}
	result, err := c.CallRaw("eval", payload)
	if err != nil {
		return nil, err
	}
	return requireRuntimeSessionOK(result, "runtime session eval")
}

// Status reads one runtime lease status payload with optional identity guards.
// Status 读取单个运行时租约状态载荷，并可附带可选身份护栏。
func (c *RuntimeSessionClient) Status(leaseID string, sid string, generation int) (map[string]any, error) {
	payload := map[string]any{
		"lease_id": leaseID,
	}
	if sid != "" {
		payload["sid"] = sid
	}
	if generation != 0 {
		payload["generation"] = generation
	}
	return c.CallRaw("status", payload)
}

// List lists active runtime leases and optionally filters by one SID.
// List 列出活跃运行时租约，并可按单个 SID 过滤。
func (c *RuntimeSessionClient) List(sid string) (map[string]any, error) {
	payload := map[string]any{}
	if sid != "" {
		payload["sid"] = sid
	}
	return c.CallRaw("list", payload)
}

// ListHandles lists active runtime-session handles rebuilt from the current lease listing payload.
// ListHandles 基于当前租约列表载荷重建活跃运行时会话句柄列表。
func (c *RuntimeSessionClient) ListHandles(sid string) ([]*RuntimeSessionHandle, error) {
	result, err := c.List(sid)
	if err != nil {
		return nil, err
	}
	rawLeases, ok := result["leases"].([]any)
	if !ok {
		return nil, fmt.Errorf("runtime session list payload is missing the leases array")
	}
	handles := make([]*RuntimeSessionHandle, 0, len(rawLeases))
	for _, rawLease := range rawLeases {
		lease, err := requireJSONMap(rawLease, "runtime session lease entry")
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

// FindHandle returns the first active runtime-session handle for one SID when present.
// FindHandle 返回某个 SID 的第一个活跃运行时会话句柄（如果存在）。
func (c *RuntimeSessionClient) FindHandle(sid string) (*RuntimeSessionHandle, error) {
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
func (c *RuntimeSessionClient) Close(leaseID string, sid string, generation int) (map[string]any, error) {
	payload := map[string]any{
		"lease_id": leaseID,
	}
	if sid != "" {
		payload["sid"] = sid
	}
	if generation != 0 {
		payload["generation"] = generation
	}
	return c.CallRaw("close", payload)
}

// UsesSystemRuntimeSessionEndpoints returns whether this helper will dispatch requests to dedicated system entrypoints.
// UsesSystemRuntimeSessionEndpoints 返回当前辅助器是否会把请求分发到专用 system 入口。
func (c *RuntimeSessionClient) UsesSystemRuntimeSessionEndpoints() (bool, error) {
	if !c.bindAuthority {
		return false, nil
	}
	return true, nil
}

// runtimeSessionFunctionName resolves the concrete runtime-session JSON FFI entrypoint name for one logical action.
// runtimeSessionFunctionName 为单个逻辑动作解析具体的运行时会话 JSON FFI 入口名称。
func (c *RuntimeSessionClient) runtimeSessionFunctionName(action string) (string, error) {
	publicName := "luaskills_ffi_runtime_session_" + action + "_json"
	if !c.bindAuthority {
		return publicName, nil
	}
	return "luaskills_ffi_system_runtime_session_" + action + "_json", nil
}

// RuntimeSessionHandle is the stable host-side runtime-session handle that carries lease identity guards automatically.
// RuntimeSessionHandle 是自动携带租约身份护栏的稳定宿主侧运行时会话句柄。
type RuntimeSessionHandle struct {
	sessions   *RuntimeSessionClient
	LeaseID    string
	SID        string
	Generation int
}

// BindRuntimeSessionHandle constructs one runtime-session handle from one payload that contains identity fields.
// BindRuntimeSessionHandle 从包含身份字段的一份载荷中构造一个运行时会话句柄。
func BindRuntimeSessionHandle(sessions *RuntimeSessionClient, payload map[string]any) (*RuntimeSessionHandle, error) {
	leaseID, err := requireRuntimeSessionStringField(payload, "lease_id")
	if err != nil {
		return nil, err
	}
	sid, err := requireRuntimeSessionStringField(payload, "sid")
	if err != nil {
		return nil, err
	}
	generation, err := requireRuntimeSessionIntField(payload, "generation")
	if err != nil {
		return nil, err
	}
	return &RuntimeSessionHandle{
		sessions:   sessions,
		LeaseID:    leaseID,
		SID:        sid,
		Generation: generation,
	}, nil
}

// IdentityPayload exports the stable lease identity fields for persistence or raw FFI calls.
// IdentityPayload 导出稳定租约身份字段，供持久化或原始 FFI 调用使用。
func (h *RuntimeSessionHandle) IdentityPayload() RuntimeSessionIdentity {
	return RuntimeSessionIdentity{
		LeaseID:    h.LeaseID,
		SID:        h.SID,
		Generation: h.Generation,
	}
}

// Eval evaluates Lua code while automatically attaching the stored lease identity guards.
// Eval 执行 Lua 代码时自动附带已保存的租约身份护栏。
func (h *RuntimeSessionHandle) Eval(code string, args map[string]any, timeoutMs int) (map[string]any, error) {
	return h.sessions.Eval(h.LeaseID, code, args, timeoutMs, h.SID, h.Generation)
}

// Status reads the current lease status while automatically attaching the stored identity guards.
// Status 读取当前租约状态时自动附带已保存的身份护栏。
func (h *RuntimeSessionHandle) Status() (map[string]any, error) {
	return h.sessions.Status(h.LeaseID, h.SID, h.Generation)
}

// Close closes the current lease while automatically attaching the stored identity guards.
// Close 关闭当前租约时自动附带已保存的身份护栏。
func (h *RuntimeSessionHandle) Close() (map[string]any, error) {
	return h.sessions.Close(h.LeaseID, h.SID, h.Generation)
}

// requireRuntimeSessionOK requires one runtime-session payload to report success.
// requireRuntimeSessionOK 要求单个运行时会话载荷报告成功。
func requireRuntimeSessionOK(payload map[string]any, action string) (map[string]any, error) {
	if ok, _ := payload["ok"].(bool); ok {
		return payload, nil
	}
	return nil, fmt.Errorf(
		"%s failed: %v: %v",
		action,
		firstNonNil(payload["error_code"], "unknown"),
		firstNonNil(payload["message"], "Unknown runtime session error"),
	)
}

// requireRuntimeSessionStringField reads one required runtime-session string field from one payload object.
// requireRuntimeSessionStringField 从一份载荷对象中读取一个必填的运行时会话字符串字段。
func requireRuntimeSessionStringField(payload map[string]any, fieldName string) (string, error) {
	value, ok := payload[fieldName].(string)
	if ok && value != "" {
		return value, nil
	}
	return "", fmt.Errorf("runtime session payload is missing required string field: %s", fieldName)
}

// requireRuntimeSessionIntField reads one required runtime-session integer field from one payload object.
// requireRuntimeSessionIntField 从一份载荷对象中读取一个必填的运行时会话整数字段。
func requireRuntimeSessionIntField(payload map[string]any, fieldName string) (int, error) {
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
		return 0, fmt.Errorf("runtime session payload is missing required integer field: %s", fieldName)
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
