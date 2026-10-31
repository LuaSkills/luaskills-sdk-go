//go:build cgo

package luaskills

/*
#include "luaskills_json_ffi.h"
#include <stdlib.h>

// Bridge one immutable Go callback handle into the core JSON provider ABI.
// 将一个不可变 Go 回调句柄桥接到核心 JSON 提供方 ABI。
extern int32_t goSQLiteProviderJSON(FfiBorrowedBuffer request, void *context, FfiOwnedBuffer *response, FfiOwnedBuffer *error);
static int32_t register_go_sqlite(uintptr_t handle, FfiOwnedBuffer *error) {
    return luaskills_ffi_set_sqlite_provider_json_callback(handle ? goSQLiteProviderJSON : NULL, (void *)handle, error);
}
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"runtime/cgo"
	"sync"
	"unsafe"
)

// sqliteRegistrationMu serializes process-default replacements without holding native-owner locks.
// sqliteRegistrationMu 串行化进程默认值替换，不持有原生所有者锁。
var sqliteRegistrationMu sync.Mutex

// sqliteRegistration is the current immutable callback handle; zero means no default callback.
// sqliteRegistration 是当前不可变回调句柄；零表示没有默认回调。
var sqliteRegistration cgo.Handle

// setNativeSQLiteProviderJSONCallback registers callback for future engines and preserves prior snapshots.
// setNativeSQLiteProviderJSONCallback 为未来引擎注册 callback 并保留既有快照。
// Returns the exact native registration error; nil callback clears only the process default.
// 返回精确的原生注册错误；nil 回调仅清空进程默认值。
func setNativeSQLiteProviderJSONCallback(callback JSONProviderCallback) error {
	sqliteRegistrationMu.Lock()
	defer sqliteRegistrationMu.Unlock()
	// Next owns a Go value through an integer token, never a Go pointer retained by C.
	// next 通过整数令牌拥有 Go 值，绝不让 C 保留 Go 指针。
	var next cgo.Handle
	if callback != nil {
		next = cgo.NewHandle(callback)
	}
	// ErrorBuffer is allocated by the core and freed with its matching buffer API.
	// errorBuffer 由核心分配，并使用其匹配缓冲 API 释放。
	var errorBuffer C.FfiOwnedBuffer
	if status := C.register_go_sqlite(C.uintptr_t(next), &errorBuffer); status != 0 {
		if next != 0 {
			next.Delete()
		}
		defer C.luaskills_ffi_buffer_free(errorBuffer)
		return fmt.Errorf("SQLite callback registration failed (%d): %s", status, C.GoStringN((*C.char)(unsafe.Pointer(errorBuffer.ptr)), C.int(errorBuffer.len)))
	}
	C.luaskills_ffi_buffer_free(errorBuffer)
	// Previous remains usable by engines that captured it before this registration.
	// previous 仍可供本次注册前捕获它的引擎使用。
	previous := sqliteRegistration
	sqliteRegistration = next
	if previous != 0 {
		retireSQLiteProvider(previous.Delete)
	}
	return nil
}

// cloneSQLiteCallbackBytes clones bytes into core-owned output memory and returns native status.
// cloneSQLiteCallbackBytes 将字节克隆到核心拥有的输出内存，并返回原生状态。
func cloneSQLiteCallbackBytes(bytes []byte, output *C.FfiOwnedBuffer) C.int32_t {
	// Storage remains borrowed only until the core finishes copying.
	// storage 仅在核心完成复制前保持借用。
	var storage unsafe.Pointer
	if len(bytes) > 0 {
		storage = C.CBytes(bytes)
		defer C.free(storage)
	}
	// CloneError belongs to this helper and never escapes to the callback caller.
	// cloneError 属于此辅助函数，绝不逃逸到回调调用方。
	var cloneError C.FfiOwnedBuffer
	status := C.luaskills_ffi_buffer_clone((*C.uint8_t)(storage), C.size_t(len(bytes)), output, &cloneError)
	C.luaskills_ffi_buffer_free(cloneError)
	return status
}

// goSQLiteProviderJSON invokes the captured callback with the exact request and owned response/error buffers.
// goSQLiteProviderJSON 使用精确请求及拥有型响应或错误缓冲调用被捕获的回调。
// Panics never cross C; immutable cgo handles remain live until all SDK-native owners are freed.
// panic 绝不跨越 C；不可变 cgo 句柄保留到所有 SDK 原生所有者释放。
//
//export goSQLiteProviderJSON
func goSQLiteProviderJSON(request C.FfiBorrowedBuffer, context unsafe.Pointer, response *C.FfiOwnedBuffer, errorOut *C.FfiOwnedBuffer) (status C.int32_t) {
	defer func() {
		if recovered := recover(); recovered != nil {
			cloneSQLiteCallbackBytes([]byte(fmt.Sprintf("SQLite host callback panic: %v", recovered)), errorOut)
			status = 1
		}
	}()
	if request.len > 2147483647 || (request.len > 0 && request.ptr == nil) {
		cloneSQLiteCallbackBytes([]byte("invalid SQLite callback request buffer"), errorOut)
		return 1
	}
	// Payload is the documented action/binding/input envelope, decoded without changing its fields.
	// payload 是已文档化的 action/binding/input 包络，解码时不改变字段。
	var payload any
	if err := json.Unmarshal(C.GoBytes(unsafe.Pointer(request.ptr), C.int(request.len)), &payload); err != nil {
		cloneSQLiteCallbackBytes([]byte(err.Error()), errorOut)
		return 1
	}
	// Callback is selected by the captured generation token, never the current global registration.
	// callback 由捕获的代次令牌选择，绝不使用当前全局注册。
	callback := cgo.Handle(uintptr(context)).Value().(JSONProviderCallback)
	result, err := callback(payload)
	if err != nil {
		cloneSQLiteCallbackBytes([]byte(err.Error()), errorOut)
		return 1
	}
	// Encoded is copied into the core allocator before this Go frame returns.
	// encoded 在此 Go 栈帧返回前复制到核心分配器。
	encoded, err := json.Marshal(result)
	if err != nil {
		cloneSQLiteCallbackBytes([]byte(err.Error()), errorOut)
		return 1
	}
	return cloneSQLiteCallbackBytes(encoded, response)
}
