package luaskills

import "errors"

// JSONProviderCallback is the host function shape used by JSON provider callbacks.
// JSONProviderCallback 是 JSON provider callback 使用的宿主函数形状。
type JSONProviderCallback func(request any) (any, error)

// HostToolJSONAction is a host-tool bridge action emitted by vulcan.host.*.
// HostToolJSONAction 是 vulcan.host.* 发出的宿主工具桥接动作。
type HostToolJSONAction string

const (
	// HostToolJSONActionList requests the current host-visible tool list.
	// HostToolJSONActionList 请求当前对宿主可见的工具列表。
	HostToolJSONActionList HostToolJSONAction = "list"
	// HostToolJSONActionHas requests whether one host tool exists.
	// HostToolJSONActionHas 请求判断某个宿主工具是否存在。
	HostToolJSONActionHas HostToolJSONAction = "has"
	// HostToolJSONActionCall requests one host-tool invocation.
	// HostToolJSONActionCall 请求执行一次宿主工具调用。
	HostToolJSONActionCall HostToolJSONAction = "call"
)

// HostToolJSONRequest is the request delivered to host-tool JSON callbacks.
// HostToolJSONRequest 是传递给宿主工具 JSON callback 的请求。
type HostToolJSONRequest struct {
	// Action is the requested host-tool bridge action.
	// Action 是请求的宿主工具桥接动作。
	Action HostToolJSONAction `json:"action"`
	// ToolName is the optional host tool name for has and call actions.
	// ToolName 是 has 与 call 动作使用的可选宿主工具名称。
	ToolName *string `json:"tool_name"`
	// Args is the JSON payload converted from the Lua table argument.
	// Args 是从 Lua table 参数转换得到的 JSON 载荷。
	Args any `json:"args"`
}

// HostToolJSONCallback is the host function shape used by host-tool JSON callbacks.
// HostToolJSONCallback 是宿主工具 JSON callback 使用的宿主函数形状。
type HostToolJSONCallback func(request HostToolJSONRequest) (any, error)

// ErrProviderCallbacksRequireHostBridge explains why Go does not register callbacks directly yet.
// ErrProviderCallbacksRequireHostBridge 说明 Go 当前为何尚不直接注册 callback。
var ErrProviderCallbacksRequireHostBridge = errors.New("luaskills Go SDK provider callbacks require a host-owned cgo callback bridge")

// ErrHostToolCallbacksRequireHostBridge explains why Go does not register host-tool callbacks directly yet.
// ErrHostToolCallbacksRequireHostBridge 说明 Go 当前为何尚不直接注册宿主工具 callback。
var ErrHostToolCallbacksRequireHostBridge = errors.New("luaskills Go SDK host tool callbacks require a host-owned cgo callback bridge")

// SetSQLiteProviderJSONCallback reports the required host bridge for SQLite JSON callbacks.
// SetSQLiteProviderJSONCallback 报告 SQLite JSON callback 所需的宿主桥接。
func SetSQLiteProviderJSONCallback(callback JSONProviderCallback) error {
	return ErrProviderCallbacksRequireHostBridge
}

// ClearSQLiteProviderJSONCallback reports the required host bridge for clearing SQLite callbacks.
// ClearSQLiteProviderJSONCallback 报告清理 SQLite callback 所需的宿主桥接。
func ClearSQLiteProviderJSONCallback() error {
	return SetSQLiteProviderJSONCallback(nil)
}

// SetLanceDBProviderJSONCallback reports the required host bridge for LanceDB JSON callbacks.
// SetLanceDBProviderJSONCallback 报告 LanceDB JSON callback 所需的宿主桥接。
func SetLanceDBProviderJSONCallback(callback JSONProviderCallback) error {
	return ErrProviderCallbacksRequireHostBridge
}

// ClearLanceDBProviderJSONCallback reports the required host bridge for clearing LanceDB callbacks.
// ClearLanceDBProviderJSONCallback 报告清理 LanceDB callback 所需的宿主桥接。
func ClearLanceDBProviderJSONCallback() error {
	return SetLanceDBProviderJSONCallback(nil)
}

// SetHostToolJSONCallback reports the required host bridge for host-tool JSON callbacks.
// SetHostToolJSONCallback 报告宿主工具 JSON callback 所需的宿主桥接。
func SetHostToolJSONCallback(callback HostToolJSONCallback) error {
	return ErrHostToolCallbacksRequireHostBridge
}

// ClearHostToolJSONCallback reports the required host bridge for clearing host-tool callbacks.
// ClearHostToolJSONCallback 报告清理宿主工具 callback 所需的宿主桥接。
func ClearHostToolJSONCallback() error {
	return SetHostToolJSONCallback(nil)
}
