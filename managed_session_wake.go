//go:build cgo

package luaskills

/*
#include <stdint.h>
#include <stdlib.h>

typedef struct FfiOwnedBuffer {
    uint8_t *ptr;
    size_t len;
} FfiOwnedBuffer;

typedef int32_t (*FfiManagedSessionWakeCallback)(uint64_t engine_id, void *user_data, FfiOwnedBuffer *error_out);

extern int32_t goManagedSessionWake(uint64_t engine_id, void *user_data, FfiOwnedBuffer *error_out);
int32_t luaskills_ffi_set_managed_session_wake_callback(uint64_t engine_id, FfiManagedSessionWakeCallback callback, void *user_data, FfiOwnedBuffer *error_out);
int32_t luaskills_ffi_buffer_clone(const uint8_t *value, size_t len, FfiOwnedBuffer *buffer_out, FfiOwnedBuffer *error_out);
void luaskills_ffi_buffer_free(FfiOwnedBuffer value);

static int32_t set_go_managed_session_wake_callback(uint64_t engine_id, int enabled, FfiOwnedBuffer *error_out) {
    return luaskills_ffi_set_managed_session_wake_callback(
        engine_id,
        enabled ? goManagedSessionWake : NULL,
        NULL,
        error_out
    );
}
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"
)

// managedSessionWakeCallbacks owns live Go callbacks until native quiescent replacement completes.
// managedSessionWakeCallbacks 在原生静默替换完成前持有存活的 Go 回调。
var managedSessionWakeCallbacks sync.Map

// setNativeManagedSessionWakeCallback registers, replaces, or clears one engine-level wake callback.
// setNativeManagedSessionWakeCallback 注册、替换或清除一个引擎级唤醒回调。
func setNativeManagedSessionWakeCallback(engineID uint64, callback ManagedSessionWakeCallback) error {
	previous, hadPrevious := managedSessionWakeCallbacks.Load(engineID)
	var errorBuffer C.FfiOwnedBuffer
	enabled := C.int(0)
	if callback != nil {
		enabled = 1
		// Store before native registration because pending events may trigger catch-up before return.
		// 在原生注册前保存，因为待处理事件可能在返回前触发补偿唤醒。
		managedSessionWakeCallbacks.Store(engineID, callback)
	}
	status := C.set_go_managed_session_wake_callback(C.uint64_t(engineID), enabled, &errorBuffer)
	if status != 0 {
		if hadPrevious {
			managedSessionWakeCallbacks.Store(engineID, previous)
		} else {
			managedSessionWakeCallbacks.Delete(engineID)
		}
		message := readManagedSessionWakeOwnedBuffer(errorBuffer)
		if message == "" {
			message = "unknown managed-session wake callback registration error"
		}
		return fmt.Errorf("luaskills_ffi_set_managed_session_wake_callback: %s", message)
	}
	if callback == nil {
		managedSessionWakeCallbacks.Delete(engineID)
	}
	return nil
}

// goManagedSessionWake dispatches one native background wake into its registered Go callback.
// goManagedSessionWake 将一次原生后台唤醒分发到已注册的 Go 回调。
//
//export goManagedSessionWake
func goManagedSessionWake(engineID C.uint64_t, _ unsafe.Pointer, errorOut *C.FfiOwnedBuffer) C.int32_t {
	callbackValue, ok := managedSessionWakeCallbacks.Load(uint64(engineID))
	if !ok {
		return 0
	}
	callback := callbackValue.(ManagedSessionWakeCallback)
	if err := callback(uint64(engineID)); err != nil {
		message := []byte(err.Error())
		var cloneError C.FfiOwnedBuffer
		var pointer *C.uint8_t
		if len(message) > 0 {
			pointer = (*C.uint8_t)(unsafe.Pointer(&message[0]))
		}
		if C.luaskills_ffi_buffer_clone(pointer, C.size_t(len(message)), errorOut, &cloneError) != 0 {
			C.luaskills_ffi_buffer_free(cloneError)
		}
		return 1
	}
	return 0
}

// readManagedSessionWakeOwnedBuffer reads and releases one native diagnostic buffer.
// readManagedSessionWakeOwnedBuffer 读取并释放一个原生诊断缓冲。
func readManagedSessionWakeOwnedBuffer(buffer C.FfiOwnedBuffer) string {
	if buffer.ptr == nil {
		return ""
	}
	defer C.luaskills_ffi_buffer_free(buffer)
	return string(C.GoBytes(unsafe.Pointer(buffer.ptr), C.int(buffer.len)))
}
