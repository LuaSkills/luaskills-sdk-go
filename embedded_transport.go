package luaskills

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// EmbeddedTransportConfig declares positive native budgets copied once at creation; no silent defaults are applied.
// EmbeddedTransportConfig 声明创建时一次性复制的正数原生预算；不静默应用默认值。
type EmbeddedTransportConfig struct {
	// MaxRuntimes bounds reserved, live and draining runtime identities.
	// MaxRuntimes 限制预留、活动及排空中的运行时身份。
	MaxRuntimes uint64
	// MaxResultBuffers counts owned responses and in-flight response reservations.
	// MaxResultBuffers 统计拥有的响应及在途响应预留。
	MaxResultBuffers uint64
	// MaxResultBytes bounds owned bytes plus worst-case in-flight response bytes.
	// MaxResultBytes 限制拥有的字节及最坏情况下的在途响应字节。
	MaxResultBytes uint64
	// MaxResponseBytes bounds one complete response before its native command executes.
	// MaxResponseBytes 在原生命令执行前限制一个完整响应。
	MaxResponseBytes uint64
	// MaxRequestBytes bounds one immutable UTF-8 request frame.
	// MaxRequestBytes 限制一个不可变 UTF-8 请求帧。
	MaxRequestBytes uint64
}

// EmbeddedTransportError reports a transport/ABI status separately from a delivered runtime business error.
// EmbeddedTransportError 将传输／ABI 状态与已交付的运行时业务错误分开报告。
type EmbeddedTransportError struct {
	// Function identifies the exact versioned C entrypoint that rejected the action.
	// Function 标识拒绝动作的精确版本化 C 入口。
	Function string
	// Status preserves the exact native signed status code.
	// Status 保留精确原生有符号状态码。
	Status EmbeddedNativeStatus
}

// Error returns the stable function and native code without interpreting them as a business result.
// Error 返回稳定函数及原生代码，不将其解释为业务结果。
func (e *EmbeddedTransportError) Error() string {
	return fmt.Sprintf("%s failed with native status %d", e.Function, e.Status)
}

// EmbeddedRuntimeError is a delivered business rejection with the original protocol code and message.
// EmbeddedRuntimeError 是已交付业务拒绝，保留原始协议代码及消息。
type EmbeddedRuntimeError struct {
	// Code is the core's stable business error classification.
	// Code 是核心的稳定业务错误分类。
	Code string
	// Message is the core's public diagnostic, separate from host implementation errors.
	// Message 是核心公开诊断，与宿主实现错误分离。
	Message string
}

// Error returns the public business message with its exact stable code.
// Error 返回公开业务消息及其精确稳定代码。
func (e *EmbeddedRuntimeError) Error() string { return e.Code + ": " + e.Message }

// EmbeddedResultReleaseError retains copied delivery after buffer release fails; this never authorizes mutation replay.
// EmbeddedResultReleaseError 在缓冲释放失败后保留复制交付；这绝不授权重放变更。
type EmbeddedResultReleaseError struct {
	// Native is the exact buffer-release rejection.
	// Native 是精确的缓冲释放拒绝。
	Native *EmbeddedTransportError
	// response owns an immutable copy, exposed only through fresh copies.
	// response 拥有不可变副本，仅通过新副本公开。
	response []byte
	// cause retains any simultaneous response-copy validation failure.
	// cause 保留同时发生的响应复制校验失败。
	cause error
}

// Error describes release failure without claiming the native command did not execute.
// Error 描述释放失败，不宣称原生命令未执行。
func (e *EmbeddedResultReleaseError) Error() string {
	return e.Native.Error() + "; delivery may already have occurred"
}

// Unwrap exposes both native release and optional copy errors to errors.Is/errors.As.
// Unwrap 向 errors.Is／errors.As 公开原生释放及可选复制错误。
func (e *EmbeddedResultReleaseError) Unwrap() []error {
	if e.cause == nil {
		return []error{e.Native}
	}
	return []error{e.Native, e.cause}
}

// ResponseBytes returns a fresh copy of delivered bytes, or nil when copying was not proven.
// ResponseBytes 返回交付字节的新副本；未证明已复制时返回 nil。
func (e *EmbeddedResultReleaseError) ResponseBytes() []byte { return bytes.Clone(e.response) }

// DeliveredResult decodes the original response without issuing a command or releasing native ownership.
// DeliveredResult 解码原始响应，不发出命令，也不释放原生所有权。
func (e *EmbeddedResultReleaseError) DeliveredResult() (any, error) {
	if e.response == nil {
		return nil, fmt.Errorf("embedded response delivery is unavailable")
	}
	return DecodeEmbeddedResponse(e.response)
}

// embeddedResult holds an exact native allocation; its opaque descriptor is never reconstructed from caller data.
// embeddedResult 持有精确原生分配；绝不从调用方数据重建其不透明描述符。
type embeddedResult struct {
	// id is the full-width allocation identity assigned by the native transport.
	// id 是原生传输分配的完整位宽分配身份。
	id uint64
	// descriptor preserves the original allocation descriptor until release succeeds.
	// descriptor 保留原始分配描述符，直到释放成功。
	descriptor any
}

// embeddedNative separates ABI memory access from synchronized Go ownership and permits controlled fault tests.
// embeddedNative 将 ABI 内存访问与同步 Go 所有权分离，并支持受控故障测试。
type embeddedNative interface {
	// describe copies bounded library-owned metadata without constructing native ownership.
	// describe 复制有界且由动态库拥有的元数据，不构造原生所有权。
	describe() ([]byte, error)
	// create allocates a root with explicit budgets and returns its exact identity and ABI status.
	// create 按显式预算分配根，并返回精确身份及 ABI 状态。
	create(EmbeddedTransportConfig) (uint64, EmbeddedNativeStatus)
	// request borrows frozen bytes for one root and returns an owned result descriptor and ABI status.
	// request 为一个根借用冻结字节，并返回拥有型结果描述符及 ABI 状态。
	request(uint64, []byte) (embeddedResult, EmbeddedNativeStatus)
	// copy validates the descriptor against the byte limit and returns independent bytes or a validation error.
	// copy 按字节上限校验描述符，并返回独立字节或校验错误。
	copy(embeddedResult, uint64) ([]byte, error)
	// release frees only the exact descriptor owned by the given root, returning its ABI status.
	// release 仅释放给定根拥有的精确描述符，并返回 ABI 状态。
	release(uint64, embeddedResult) EmbeddedNativeStatus
	// close closes creation admission for the exact root and returns its ABI status.
	// close 关闭精确根的创建入场，并返回 ABI 状态。
	close(uint64) EmbeddedNativeStatus
	// free removes the exact drained root and returns its ABI status without inferring cleanup.
	// free 移除精确且已排空的根，并返回 ABI 状态，不推测清理完成。
	free(uint64) EmbeddedNativeStatus
}

// embeddedTransports strongly retains native owners until explicit successful release; GC never frees active work.
// embeddedTransports 强引用保留原生所有者直到显式成功释放；GC 绝不释放活动工作。
var embeddedTransports sync.Map

// EmbeddedTransport owns one independent native root; concurrent calls keep short Go locks outside native execution.
// EmbeddedTransport 拥有一个独立原生根；并发调用仅在原生执行之外持有短 Go 锁。
// Do not copy it after use. Close requests admission closure; Free requires actual drainage and released buffers.
// 使用后不可复制。Close 请求关闭入场；Free 要求实际排空及缓冲已释放。
type EmbeddedTransport struct {
	// mu protects identity, active readers and result ownership; never held across a native request.
	// mu 保护身份、活动读取者及结果所有权；绝不跨原生请求持有。
	mu sync.Mutex
	// config remains immutable after construction.
	// config 在构造后保持不可变。
	config EmbeddedTransportConfig
	// native is the exact linked implementation chosen at construction.
	// native 是构造时选定的精确链接实现。
	native embeddedNative
	// description retains validated evidence and owns every nested slice independently of native storage.
	// description 保留已校验证据，并独立于原生存储拥有全部嵌套切片。
	description EmbeddedOutputCoreDescription
	// identity becomes zero only after successful native removal.
	// identity 仅在原生成功移除后变为零。
	identity uint64
	// active includes requests, copying and release, preventing premature Free or recovery.
	// active 包含请求、复制及释放，防止过早 Free 或恢复。
	active uint64
	// exclusive fences admission during root removal or retained-result recovery.
	// exclusive 在根移除或保留结果恢复期间封闭入场。
	exclusive bool
	// results retains exact failed-release allocations until explicit successful recovery.
	// results 保留精确释放失败分配，直到显式恢复成功。
	results map[uint64]embeddedResult
	// commandDriver owns fixed native frame reservations until all of its workers have safely exited.
	// commandDriver 拥有固定原生帧预留，直到其全部工作位安全退出。
	commandDriver *EmbeddedCommandDriver
	// callbackPumps retain exact runtime owners and one independent native frame reservation per pump.
	// callbackPumps 保留精确运行时所有者，并为每个泵预留一个独立原生帧。
	callbackPumps map[string]*EmbeddedCallbackPump
	// runtimeScopes retain exclusive slot-removal owners and their independent control-frame reservations.
	// runtimeScopes 保留独占槽移除所有者及其独立控制帧预留。
	runtimeScopes map[string]*EmbeddedRuntimeScope
}

// NewEmbeddedTransport validates config and core compatibility before allocating an independent root.
// NewEmbeddedTransport 在分配独立根前校验 config 及核心兼容性。
// It returns a strongly retained owner or an explicit cgo/native error; it creates no runtime or VM.
// 返回强保留所有者或显式 cgo／原生错误；不创建运行时或 VM。
func NewEmbeddedTransport(config EmbeddedTransportConfig) (*EmbeddedTransport, error) {
	if err := validateEmbeddedConfig(config); err != nil {
		return nil, err
	}
	native, err := newEmbeddedNative()
	if err != nil {
		return nil, err
	}
	return createEmbeddedTransport(config, native)
}

// validateEmbeddedConfig rejects invalid bounds before native allocation and avoids narrowing uint64 to Go int.
// validateEmbeddedConfig 在原生分配前拒绝无效边界，避免将 uint64 截断为 Go int。
func validateEmbeddedConfig(config EmbeddedTransportConfig) error {
	for _, limit := range []uint64{config.MaxRuntimes, config.MaxResultBuffers, config.MaxResultBytes, config.MaxResponseBytes, config.MaxRequestBytes} {
		if limit == 0 || limit > uint64(^uint(0)>>1) {
			return fmt.Errorf("embedded budgets must fit positive native signed sizes")
		}
	}
	if config.MaxResponseBytes > config.MaxResultBytes {
		return fmt.Errorf("embedded response budget exceeds aggregate result bytes")
	}
	return nil
}

// createEmbeddedTransport checks native description before allocation and retains the exact successful identity.
// createEmbeddedTransport 在分配前检查原生描述，并保留精确成功身份。
func createEmbeddedTransport(config EmbeddedTransportConfig, native embeddedNative) (*EmbeddedTransport, error) {
	if err := validateEmbeddedConfig(config); err != nil {
		return nil, err
	}
	encoded, err := native.describe()
	if err != nil {
		return nil, err
	}
	description, err := validateEmbeddedDescription(encoded)
	if err != nil {
		return nil, err
	}
	identity, status := native.create(config)
	if status != EmbeddedNativeOk {
		return nil, &EmbeddedTransportError{"luaskills_ffi_embedded_transport_new_v1", status}
	}
	transport := &EmbeddedTransport{config: config, native: native, description: description, identity: identity, results: make(map[uint64]embeddedResult)}
	embeddedTransports.Store(transport, struct{}{})
	return transport, nil
}

// LiveEmbeddedTransports returns a snapshot of strongly retained native owners in unspecified order.
// LiveEmbeddedTransports 以未指定顺序返回强引用保留的原生所有者快照。
func LiveEmbeddedTransports() []*EmbeddedTransport {
	result := []*EmbeddedTransport{}
	embeddedTransports.Range(func(key, value any) bool { result = append(result, key.(*EmbeddedTransport)); return true })
	return result
}

// Config returns an independent copy of the immutable declared budgets.
// Config 返回不可变声明预算的独立副本。
func (t *EmbeddedTransport) Config() EmbeddedTransportConfig { return t.config }

// TransportID returns the exact uint64 identity, or zero after proven release.
// TransportID 返回精确 uint64 身份；证明释放后返回零。
func (t *EmbeddedTransport) TransportID() uint64 { t.mu.Lock(); defer t.mu.Unlock(); return t.identity }

// RetainedResults returns the number of exact native allocations awaiting successful release.
// RetainedResults 返回等待成功释放的精确原生分配数量。
func (t *EmbeddedTransport) RetainedResults() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.results)
}

// begin starts one real native call while rejecting released or exclusively maintained transports.
// begin 启动一个真实原生调用，同时拒绝已释放或独占维护的传输。
func (t *EmbeddedTransport) begin() (uint64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.identity == 0 {
		return 0, &EmbeddedTransportError{"transport", EmbeddedNativeClosed}
	}
	if t.exclusive {
		return 0, &EmbeddedTransportError{"transport", EmbeddedNativeBusy}
	}
	t.active++
	return t.identity, nil
}

// end releases one actual call's reader claim, after result copying and release have finished.
// end 在结果复制及释放完成后，解除一个实际调用的读取者声明。
func (t *EmbeddedTransport) end() { t.mu.Lock(); t.active--; t.mu.Unlock() }

// Request freezes command into the exact protocol envelope and returns delivered data or a stable error.
// Request 将 command 冻结为精确协议信封，返回已交付数据或稳定错误。
// This low-level call has no observer cancellation; callers must not mutate input while it is being frozen.
// 此低层调用不提供观察取消；调用方不得在输入冻结期间修改输入。
func (t *EmbeddedTransport) Request(command map[string]any) (any, error) {
	frame, err := EncodeEmbeddedJSON(map[string]any{"protocol_version": uint32(EmbeddedProtocolVersion), "command": command}, t.config.MaxRequestBytes)
	if err != nil {
		return nil, err
	}
	response, err := t.requestBytes(frame)
	if err != nil {
		return nil, err
	}
	return DecodeEmbeddedResponse(response)
}

// requestBytes owns exact frozen frame execution and copies before release; native failures never trigger replay.
// requestBytes 拥有精确冻结帧的执行并在释放前复制；原生失败绝不触发重放。
func (t *EmbeddedTransport) requestBytes(frame []byte) ([]byte, error) {
	identity, err := t.begin()
	if err != nil {
		return nil, err
	}
	defer t.end()
	result, status := t.native.request(identity, frame)
	if status != EmbeddedNativeOk {
		return nil, &EmbeddedTransportError{"luaskills_ffi_embedded_request_v1", status}
	}
	t.mu.Lock()
	t.results[result.id] = result
	t.mu.Unlock()
	response, copyError := t.native.copy(result, t.config.MaxResponseBytes)
	status = t.native.release(identity, result)
	if status != EmbeddedNativeOk {
		return response, &EmbeddedResultReleaseError{&EmbeddedTransportError{"luaskills_ffi_embedded_result_free_v1", status}, bytes.Clone(response), copyError}
	}
	t.mu.Lock()
	delete(t.results, result.id)
	t.mu.Unlock()
	return response, copyError
}

// Close permanently closes native creation admission without pretending active runtimes or calls are released.
// Close 永久关闭原生创建入场，不假装活动运行时或调用已释放。
// Cleanup/status requests remain available; returns the exact native error if closure was rejected.
// 清理／状态请求仍可用；关闭被拒绝时返回精确原生错误。
func (t *EmbeddedTransport) Close() error {
	identity, err := t.begin()
	if err != nil {
		return err
	}
	defer t.end()
	if status := t.native.close(identity); status != EmbeddedNativeOk {
		return &EmbeddedTransportError{"luaskills_ffi_embedded_transport_close_v1", status}
	}
	return nil
}

// ReleaseResults retries only retained buffer releases after all readers leave; it never repeats native commands.
// ReleaseResults 在全部读取者离开后仅重试保留缓冲释放；绝不重复原生命令。
// Concurrent admission is fenced; errors retain their exact allocation descriptors for another explicit attempt.
// 并发入场被封闭；错误为下一次显式尝试保留精确分配描述符。
func (t *EmbeddedTransport) ReleaseResults() error {
	t.mu.Lock()
	if t.identity == 0 {
		t.mu.Unlock()
		return &EmbeddedTransportError{"transport", EmbeddedNativeClosed}
	}
	if t.active != 0 || t.exclusive {
		t.mu.Unlock()
		return &EmbeddedTransportError{"transport", EmbeddedNativeBusy}
	}
	t.exclusive = true
	identity := t.identity
	results := make([]embeddedResult, 0, len(t.results))
	for _, result := range t.results {
		results = append(results, result)
	}
	t.mu.Unlock()
	defer func() { t.mu.Lock(); t.exclusive = false; t.mu.Unlock() }()
	var failures []error
	for _, result := range results {
		if status := t.native.release(identity, result); status != EmbeddedNativeOk {
			failures = append(failures, &EmbeddedTransportError{"luaskills_ffi_embedded_result_free_v1", status})
		} else {
			t.mu.Lock()
			delete(t.results, result.id)
			t.mu.Unlock()
		}
	}
	return errors.Join(failures...)
}

// Free removes only a closed and drained root; failures leave identity and ownership available for cleanup.
// Free 仅移除已关闭且排空的根；失败后身份及所有权仍可用于清理。
// Active calls, readers, retained allocations, drivers or callback pumps reject before entering C.
// 活动调用、读取者、保留分配、驱动器或回调泵在进入 C 前即被拒绝。
func (t *EmbeddedTransport) Free() error {
	t.mu.Lock()
	if t.identity == 0 {
		t.mu.Unlock()
		return &EmbeddedTransportError{"transport", EmbeddedNativeClosed}
	}
	if t.active != 0 || t.exclusive || len(t.results) != 0 || t.commandDriver != nil || len(t.callbackPumps) != 0 || len(t.runtimeScopes) != 0 {
		t.mu.Unlock()
		return &EmbeddedTransportError{"transport", EmbeddedNativeBusy}
	}
	t.exclusive = true
	identity := t.identity
	t.mu.Unlock()
	status := t.native.free(identity)
	t.mu.Lock()
	t.exclusive = false
	if status == EmbeddedNativeOk {
		t.identity = 0
	}
	t.mu.Unlock()
	if status != EmbeddedNativeOk {
		return &EmbeddedTransportError{"luaskills_ffi_embedded_transport_free_v1", status}
	}
	embeddedTransports.Delete(t)
	return nil
}

// claimCommandDriver reserves worst-case frames before driver workers start; counts and multiplication are checked safely.
// claimCommandDriver 在驱动器工作位启动前预留最坏情况帧；安全校验数量及乘法边界。
func (t *EmbeddedTransport) claimCommandDriver(driver *EmbeddedCommandDriver) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.identity == 0 {
		return &EmbeddedTransportError{"transport", EmbeddedNativeClosed}
	}
	if t.active != 0 || t.exclusive || len(t.results) != 0 || t.commandDriver != nil {
		return &EmbeddedRuntimeError{"busy", "embedded driver requires no active requests, retained allocations, driver or exclusive maintenance"}
	}
	if err := t.checkAdditionalFramesLocked(driver.config.WorkWorkers + embeddedControlWorkers); err != nil {
		return err
	}
	t.commandDriver = driver
	return nil
}

// releaseCommandDriver removes only the exact claim after safe worker drainage; a mismatch leaves ownership unchanged.
// releaseCommandDriver 仅在工作位安全排空后移除精确声明；不匹配时保持所有权不变。
func (t *EmbeddedTransport) releaseCommandDriver(driver *EmbeddedCommandDriver) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.commandDriver != driver {
		return fmt.Errorf("embedded command driver ownership mismatch")
	}
	t.commandDriver = nil
	return nil
}

// DecodeEmbeddedResponse validates the exact envelope and returns result including explicit null, or a business error.
// DecodeEmbeddedResponse 校验精确信封并返回包含显式空值的 result，或业务错误。
// Unknown versions, missing results and additional fields are rejected without guessing an alternate protocol.
// 未知版本、缺失结果及额外字段均被拒绝，不猜测备用协议。
func DecodeEmbeddedResponse(data []byte) (any, error) {
	decoded, err := DecodeEmbeddedJSON(data)
	if err != nil {
		return nil, err
	}
	envelope, ok := decoded.(map[string]any)
	if !ok || len(envelope) != 3 || envelope["protocol_version"] != json.Number(fmt.Sprint(EmbeddedProtocolVersion)) {
		return nil, fmt.Errorf("invalid embedded response protocol envelope")
	}
	if envelope["status"] == "ok" {
		if result, present := envelope["result"]; present {
			return result, nil
		}
	}
	if envelope["status"] == "error" {
		failure, ok := envelope["error"].(map[string]any)
		if ok && len(failure) == 2 {
			code, codeOK := failure["code"].(string)
			message, messageOK := failure["message"].(string)
			if codeOK && messageOK {
				return nil, &EmbeddedRuntimeError{code, message}
			}
		}
	}
	return nil, fmt.Errorf("invalid embedded response envelope")
}
