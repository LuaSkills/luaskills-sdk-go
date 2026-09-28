// Code generated from the packaged embedded contract; DO NOT EDIT.
// 从包内嵌入式契约生成；请勿手工编辑。
package luaskills

// EmbeddedProtocolVersion is the exact root JSON and C structure protocol version.
// EmbeddedProtocolVersion 是精确的根 JSON 及 C 结构协议版本。
const EmbeddedProtocolVersion uint32 = 1

// EmbeddedContractSHA256 identifies every byte of the packaged contract.
// EmbeddedContractSHA256 标识包内契约的全部字节。
const EmbeddedContractSHA256 = "05171a3cc51160288de8b5249a509293115e7a8d42c9f5f92122108fb18f2f20"

// EmbeddedCoreVersion identifies the core package that generated this contract.
// EmbeddedCoreVersion 标识生成此契约的核心包。
const EmbeddedCoreVersion = "0.5.9"

// EmbeddedDescriptionVersion is the independent borrowed descriptor format.
// EmbeddedDescriptionVersion 是独立借用型描述格式。
const EmbeddedDescriptionVersion uint32 = 1

// EmbeddedDescriptionMaxBytes bounds native descriptor copies before JSON decoding.
// EmbeddedDescriptionMaxBytes 限制 JSON 解码前的原生描述复制。
const EmbeddedDescriptionMaxBytes uint64 = 16384

// EmbeddedNativeStatus is one exact signed C ABI status code.
// EmbeddedNativeStatus 是一个精确的有符号 C ABI 状态码。
type EmbeddedNativeStatus int32

const (
	// EmbeddedNativeBusy preserves the core's busy status.
	// EmbeddedNativeBusy 保留核心的 busy 状态。
	EmbeddedNativeBusy EmbeddedNativeStatus = 3
	// EmbeddedNativeCapacityExceeded preserves the core's capacity_exceeded status.
	// EmbeddedNativeCapacityExceeded 保留核心的 capacity_exceeded 状态。
	EmbeddedNativeCapacityExceeded EmbeddedNativeStatus = 4
	// EmbeddedNativeClosed preserves the core's closed status.
	// EmbeddedNativeClosed 保留核心的 closed 状态。
	EmbeddedNativeClosed EmbeddedNativeStatus = 5
	// EmbeddedNativeInternal preserves the core's internal status.
	// EmbeddedNativeInternal 保留核心的 internal 状态。
	EmbeddedNativeInternal EmbeddedNativeStatus = 6
	// EmbeddedNativeInvalidArgument preserves the core's invalid_argument status.
	// EmbeddedNativeInvalidArgument 保留核心的 invalid_argument 状态。
	EmbeddedNativeInvalidArgument EmbeddedNativeStatus = 1
	// EmbeddedNativeNotFound preserves the core's not_found status.
	// EmbeddedNativeNotFound 保留核心的 not_found 状态。
	EmbeddedNativeNotFound EmbeddedNativeStatus = 2
	// EmbeddedNativeOk preserves the core's ok status.
	// EmbeddedNativeOk 保留核心的 ok 状态。
	EmbeddedNativeOk EmbeddedNativeStatus = 0
	// EmbeddedNativeUnsupported preserves the core's unsupported status.
	// EmbeddedNativeUnsupported 保留核心的 unsupported 状态。
	EmbeddedNativeUnsupported EmbeddedNativeStatus = 7
)

// EmbeddedRootCommands returns an independent copy of this authoritative name inventory.
// EmbeddedRootCommands 返回此权威名称清单的独立副本。
func EmbeddedRootCommands() []string {
	return []string{"describe", "runtime_reserve", "runtime_initialize", "runtime_status", "runtime_close", "runtime_free", "runtime"}
}

// EmbeddedRuntimeCommands returns an independent copy of this authoritative name inventory.
// EmbeddedRuntimeCommands 返回此权威名称清单的独立副本。
func EmbeddedRuntimeCommands() []string {
	return []string{"operation_persistence_failure", "operation_retry_checkpoint", "storage_status", "storage_recover", "storage_worker_recover", "history_get", "history_next", "history_reconcile", "history_forget", "plugin_register", "plugin_status", "plugin_close", "plugin_forget", "pool_register", "pool_status", "pool_close", "pool_forget", "pool_revoke_permission", "call_submit", "session_open", "session_submit", "session_status", "session_close", "session_forget", "operation_status", "operation_wait", "operation_cancel", "operation_forget", "capabilities_register", "capabilities_list", "capability_status", "capability_unregister", "capability_forget", "host_requests_take", "host_request_status", "host_request_complete"}
}

// EmbeddedRequiredCapabilities returns an independent copy of this authoritative name inventory.
// EmbeddedRequiredCapabilities 返回此权威名称清单的独立副本。
func EmbeddedRequiredCapabilities() []string {
	return []string{"bounded_transports_v1", "plugin_budgets_v1", "shared_pools_v1", "dedicated_pools_v1", "fixed_sessions_v1", "host_request_queue_v1", "in_memory_effect_evidence_v1", "durable_operation_history_v1", "historical_effect_reconciliation_v1", "live_storage_recovery_v1", "journal_worker_recovery_v1", "strict_json_v1"}
}
