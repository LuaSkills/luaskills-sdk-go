//go:build cgo

package luaskills

/*
#include "luaskills_json_ffi.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// embeddedCgo calls only the exact version-one ABI declared in the packaged upstream headers.
// embeddedCgo 仅调用包内上游头文件声明的精确版本一 ABI。
type embeddedCgo struct{}

// newEmbeddedNative selects the linked cgo implementation; there is no legacy protocol fallback.
// newEmbeddedNative 选择链接的 cgo 实现；不存在旧协议回退。
func newEmbeddedNative() (embeddedNative, error) { return embeddedCgo{}, nil }

// describe reads the required linked bootstrap and returns independent bytes; it never calls a result-free function.
// describe 读取必需链接引导入口并返回独立字节；绝不调用结果释放函数。
func (embeddedCgo) describe() ([]byte, error) {
	var description C.FfiBorrowedBuffer
	status := EmbeddedNativeStatus(C.luaskills_ffi_embedded_describe_v1(&description))
	if status != EmbeddedNativeOk {
		return nil, &EmbeddedTransportError{"luaskills_ffi_embedded_describe_v1", status}
	}
	return copyEmbeddedDescription(unsafe.Pointer(description.ptr), uint64(description.len))
}

// copyEmbeddedDescription checks pointer/length before reading a library-owned buffer and returns Go-owned bytes.
// copyEmbeddedDescription 在读取动态库拥有缓冲前检查指针及长度，并返回 Go 拥有字节。
// Link-time cgo binding retains the native module for the process lifetime; this SDK has no unload operation.
// 链接时 cgo 绑定在进程寿命内保留原生模块；此 SDK 不提供卸载操作。
func copyEmbeddedDescription(pointer unsafe.Pointer, length uint64) ([]byte, error) {
	if pointer == nil || length == 0 || length > EmbeddedDescriptionMaxBytes || length > uint64(^uint(0)>>1) {
		return nil, &EmbeddedCompatibilityError{Message: "invalid native core description buffer"}
	}
	return append([]byte(nil), unsafe.Slice((*byte)(pointer), int(length))...), nil
}

// create copies explicit config into the exact C layout and returns the full-width identity plus native status.
// create 将显式 config 复制到精确 C 布局，并返回完整位宽身份及原生状态。
func (embeddedCgo) create(config EmbeddedTransportConfig) (uint64, EmbeddedNativeStatus) {
	native := C.FfiEmbeddedTransportConfigV1{struct_size: C.uint32_t(C.sizeof_FfiEmbeddedTransportConfigV1), protocol_version: C.uint32_t(EmbeddedProtocolVersion), max_runtimes: C.uint64_t(config.MaxRuntimes), max_result_buffers: C.uint64_t(config.MaxResultBuffers), max_result_bytes: C.uint64_t(config.MaxResultBytes), max_response_bytes: C.uint64_t(config.MaxResponseBytes), max_request_bytes: C.uint64_t(config.MaxRequestBytes)}
	var identity C.uint64_t
	status := C.luaskills_ffi_embedded_transport_new_v1(&native, &identity)
	return uint64(identity), EmbeddedNativeStatus(status)
}

// request borrows immutable pointer-free Go bytes only until C returns; the output owns an independent native buffer.
// request 仅在 C 返回前借用不含指针的不可变 Go 字节；输出拥有独立原生缓冲。
func (embeddedCgo) request(identity uint64, frame []byte) (embeddedResult, EmbeddedNativeStatus) {
	var result C.FfiEmbeddedResultV1
	input := C.FfiBorrowedBuffer{ptr: (*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(frame))), len: C.size_t(len(frame))}
	status := C.luaskills_ffi_embedded_request_v1(C.uint64_t(identity), input, &result)
	runtime.KeepAlive(frame)
	return embeddedResult{uint64(result.allocation_id), result}, EmbeddedNativeStatus(status)
}

// copy validates descriptor bounds before reading and returns a Go-owned copy within limit.
// copy 在读取前校验描述符边界，返回 limit 内的 Go 拥有副本。
func (embeddedCgo) copy(result embeddedResult, limit uint64) ([]byte, error) {
	native := result.descriptor.(C.FfiEmbeddedResultV1)
	if native.ptr == nil || native.len == 0 || native.allocation_id == 0 || uint64(native.len) > limit || uint64(native.len) > uint64(^uint(0)>>1) {
		return nil, fmt.Errorf("invalid embedded native response descriptor")
	}
	return append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(native.ptr)), int(native.len))...), nil
}

// release submits the exact original descriptor, never a reconstructed pointer or a legacy buffer free.
// release 提交精确原始描述符，绝不使用重建指针或旧缓冲释放入口。
func (embeddedCgo) release(identity uint64, result embeddedResult) EmbeddedNativeStatus {
	return EmbeddedNativeStatus(C.luaskills_ffi_embedded_result_free_v1(C.uint64_t(identity), result.descriptor.(C.FfiEmbeddedResultV1)))
}

// close requests admission closure for the exact identity without unloading the library.
// close 为精确身份请求关闭入场，不卸载动态库。
func (embeddedCgo) close(identity uint64) EmbeddedNativeStatus {
	return EmbeddedNativeStatus(C.luaskills_ffi_embedded_transport_close_v1(C.uint64_t(identity)))
}

// free releases the exact root only after core ownership has drained.
// free 仅在核心所有权排空后释放精确根。
func (embeddedCgo) free(identity uint64) EmbeddedNativeStatus {
	return EmbeddedNativeStatus(C.luaskills_ffi_embedded_transport_free_v1(C.uint64_t(identity)))
}
