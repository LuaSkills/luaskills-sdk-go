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

// ModelJSONCapability is a standard model capability emitted by vulcan.models.*.
// ModelJSONCapability 是 vulcan.models.* 发出的标准模型能力。
type ModelJSONCapability string

const (
	// ModelJSONCapabilityEmbed identifies embedding callbacks.
	// ModelJSONCapabilityEmbed 标识 embedding callback。
	ModelJSONCapabilityEmbed ModelJSONCapability = "embed"
	// ModelJSONCapabilityLLM identifies non-streaming LLM callbacks.
	// ModelJSONCapabilityLLM 标识非流式 LLM callback。
	ModelJSONCapabilityLLM ModelJSONCapability = "llm"
)

// ModelJSONCaller is the caller context delivered to model JSON callbacks.
// ModelJSONCaller 是传递给模型 JSON callback 的调用方上下文。
type ModelJSONCaller struct {
	// SkillID is the skill identifier that owns the active Lua entry.
	// SkillID 是拥有当前 Lua 入口的 skill 标识符。
	SkillID *string `json:"skill_id"`
	// EntryName is the local entry name declared by the owning skill.
	// EntryName 是所属 skill 声明的局部入口名称。
	EntryName *string `json:"entry_name"`
	// CanonicalToolName is the canonical runtime tool name currently executing.
	// CanonicalToolName 是当前正在执行的 canonical 运行时工具名称。
	CanonicalToolName *string `json:"canonical_tool_name"`
	// RootName is the runtime root name that owns the current skill.
	// RootName 是拥有当前 skill 的运行时根名称。
	RootName *string `json:"root_name"`
	// SkillDir is the host-visible absolute skill directory.
	// SkillDir 是对宿主可见的绝对 skill 目录。
	SkillDir *string `json:"skill_dir"`
	// ClientName is the host-provided client name from the current request context.
	// ClientName 是当前请求上下文中的宿主提供客户端名称。
	ClientName *string `json:"client_name"`
	// RequestID is the host-provided request identifier from the current request context.
	// RequestID 是当前请求上下文中的宿主提供请求标识符。
	RequestID *string `json:"request_id"`
}

// ModelJSONUsage is optional token usage metadata returned by a host-managed model provider.
// ModelJSONUsage 是宿主管理模型 provider 返回的可选 token 用量元数据。
type ModelJSONUsage struct {
	// InputTokens is the optional input token count.
	// InputTokens 是可选输入 token 数量。
	InputTokens *uint64 `json:"input_tokens"`
	// OutputTokens is the optional output token count.
	// OutputTokens 是可选输出 token 数量。
	OutputTokens *uint64 `json:"output_tokens"`
}

// ModelEmbedJSONRequest is the request delivered to model embedding callbacks.
// ModelEmbedJSONRequest 是传递给模型 embedding callback 的请求。
type ModelEmbedJSONRequest struct {
	// Text is the single input text requested by Lua.
	// Text 是 Lua 请求的单条输入文本。
	Text string `json:"text"`
	// Caller is the context captured from the active Lua runtime scope.
	// Caller 是从当前 Lua 运行时作用域捕获的调用方上下文。
	Caller ModelJSONCaller `json:"caller"`
}

// ModelEmbedJSONResponse is the response returned by model embedding callbacks.
// ModelEmbedJSONResponse 是模型 embedding callback 返回的响应。
type ModelEmbedJSONResponse struct {
	// Vector is the embedding vector returned by the host-managed model provider.
	// Vector 是宿主管理模型 provider 返回的 embedding 向量。
	Vector []float32 `json:"vector"`
	// Dimensions is the number of vector dimensions reported by the host.
	// Dimensions 是宿主报告的向量维度数量。
	Dimensions int `json:"dimensions"`
	// Usage is optional token usage metadata reported by the host.
	// Usage 是宿主报告的可选 token 用量元数据。
	Usage *ModelJSONUsage `json:"usage,omitempty"`
}

// ModelLLMJSONRequest is the request delivered to model LLM callbacks.
// ModelLLMJSONRequest 是传递给模型 LLM callback 的请求。
type ModelLLMJSONRequest struct {
	// System is the system instruction text supplied by Lua.
	// System 是 Lua 提供的 system 指令文本。
	System string `json:"system"`
	// User is the user message text supplied by Lua.
	// User 是 Lua 提供的 user 消息文本。
	User string `json:"user"`
	// Caller is the context captured from the active Lua runtime scope.
	// Caller 是从当前 Lua 运行时作用域捕获的调用方上下文。
	Caller ModelJSONCaller `json:"caller"`
}

// ModelLLMJSONResponse is the response returned by model LLM callbacks.
// ModelLLMJSONResponse 是模型 LLM callback 返回的响应。
type ModelLLMJSONResponse struct {
	// Assistant is the text returned by the host-managed model provider.
	// Assistant 是宿主管理模型 provider 返回的文本。
	Assistant string `json:"assistant"`
	// Usage is optional token usage metadata reported by the host.
	// Usage 是宿主报告的可选 token 用量元数据。
	Usage *ModelJSONUsage `json:"usage,omitempty"`
}

// ModelJSONErrorCode is a stable LuaSkills-level model error code.
// ModelJSONErrorCode 是稳定的 LuaSkills 级模型错误码。
type ModelJSONErrorCode string

const (
	// ModelJSONErrorModelUnavailable means no callback is registered for the requested capability.
	// ModelJSONErrorModelUnavailable 表示请求能力没有注册 callback。
	ModelJSONErrorModelUnavailable ModelJSONErrorCode = "model_unavailable"
	// ModelJSONErrorInvalidArgument means Lua supplied invalid arguments.
	// ModelJSONErrorInvalidArgument 表示 Lua 提供了非法参数。
	ModelJSONErrorInvalidArgument ModelJSONErrorCode = "invalid_argument"
	// ModelJSONErrorProviderError means the host model provider returned an error.
	// ModelJSONErrorProviderError 表示宿主模型 provider 返回了错误。
	ModelJSONErrorProviderError ModelJSONErrorCode = "provider_error"
	// ModelJSONErrorTimeout means the host model call timed out.
	// ModelJSONErrorTimeout 表示宿主模型调用超时。
	ModelJSONErrorTimeout ModelJSONErrorCode = "timeout"
	// ModelJSONErrorBudgetExceeded means the host rejected the call because of a budget or limit.
	// ModelJSONErrorBudgetExceeded 表示宿主因预算或限制拒绝本次调用。
	ModelJSONErrorBudgetExceeded ModelJSONErrorCode = "budget_exceeded"
	// ModelJSONErrorInternalError means LuaSkills or the host bridge hit an internal failure.
	// ModelJSONErrorInternalError 表示 LuaSkills 或宿主桥接遇到内部故障。
	ModelJSONErrorInternalError ModelJSONErrorCode = "internal_error"
)

// ModelJSONError is the structured error returned by a model JSON callback.
// ModelJSONError 是模型 JSON callback 返回的结构化错误。
type ModelJSONError struct {
	// Code is the stable LuaSkills-level model error code.
	// Code 是稳定的 LuaSkills 级模型错误码。
	Code ModelJSONErrorCode `json:"code"`
	// Message is the human-readable error summary.
	// Message 是人类可读的错误摘要。
	Message string `json:"message"`
	// ProviderMessage is optional raw provider error text after host-side redaction.
	// ProviderMessage 是宿主侧脱敏后的可选 provider 原始错误文本。
	ProviderMessage *string `json:"provider_message,omitempty"`
	// ProviderCode is the optional raw provider error code.
	// ProviderCode 是可选 provider 原始错误码。
	ProviderCode *string `json:"provider_code,omitempty"`
	// ProviderStatus is an optional provider status such as an HTTP status code.
	// ProviderStatus 是可选 provider 状态，例如 HTTP 状态码。
	ProviderStatus *uint16 `json:"provider_status,omitempty"`
}

// ModelJSONErrorEnvelope is the error envelope returned by a model JSON callback.
// ModelJSONErrorEnvelope 是模型 JSON callback 返回的错误包络。
type ModelJSONErrorEnvelope struct {
	// OK is always false for callback error envelopes.
	// OK 在 callback 错误包络中固定为 false。
	OK bool `json:"ok"`
	// Error is the structured model error payload.
	// Error 是结构化模型错误载荷。
	Error ModelJSONError `json:"error"`
}

// ModelEmbedJSONCallback is the host function shape used by model embedding JSON callbacks.
// ModelEmbedJSONCallback 是模型 embedding JSON callback 使用的宿主函数形状。
type ModelEmbedJSONCallback func(request ModelEmbedJSONRequest) (any, error)

// ModelLLMJSONCallback is the host function shape used by model LLM JSON callbacks.
// ModelLLMJSONCallback 是模型 LLM JSON callback 使用的宿主函数形状。
type ModelLLMJSONCallback func(request ModelLLMJSONRequest) (any, error)

// ErrProviderCallbacksRequireHostBridge explains why Go does not register callbacks directly yet.
// ErrProviderCallbacksRequireHostBridge 说明 Go 当前为何尚不直接注册 callback。
var ErrProviderCallbacksRequireHostBridge = errors.New("luaskills Go SDK provider callbacks require a host-owned cgo callback bridge")

// ErrHostToolCallbacksRequireHostBridge explains why Go does not register host-tool callbacks directly yet.
// ErrHostToolCallbacksRequireHostBridge 说明 Go 当前为何尚不直接注册宿主工具 callback。
var ErrHostToolCallbacksRequireHostBridge = errors.New("luaskills Go SDK host tool callbacks require a host-owned cgo callback bridge")

// ErrModelCallbacksRequireHostBridge explains why Go does not register model callbacks directly yet.
// ErrModelCallbacksRequireHostBridge 说明 Go 当前为何尚不直接注册模型 callback。
var ErrModelCallbacksRequireHostBridge = errors.New("luaskills Go SDK model callbacks require a host-owned cgo callback bridge")

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

// SetModelEmbedJSONCallback reports the required host bridge for model embedding JSON callbacks.
// SetModelEmbedJSONCallback 报告模型 embedding JSON callback 所需的宿主桥接。
func SetModelEmbedJSONCallback(callback ModelEmbedJSONCallback) error {
	return ErrModelCallbacksRequireHostBridge
}

// ClearModelEmbedJSONCallback reports the required host bridge for clearing model embedding callbacks.
// ClearModelEmbedJSONCallback 报告清理模型 embedding callback 所需的宿主桥接。
func ClearModelEmbedJSONCallback() error {
	return SetModelEmbedJSONCallback(nil)
}

// SetModelLLMJSONCallback reports the required host bridge for model LLM JSON callbacks.
// SetModelLLMJSONCallback 报告模型 LLM JSON callback 所需的宿主桥接。
func SetModelLLMJSONCallback(callback ModelLLMJSONCallback) error {
	return ErrModelCallbacksRequireHostBridge
}

// ClearModelLLMJSONCallback reports the required host bridge for clearing model LLM callbacks.
// ClearModelLLMJSONCallback 报告清理模型 LLM callback 所需的宿主桥接。
func ClearModelLLMJSONCallback() error {
	return SetModelLLMJSONCallback(nil)
}
