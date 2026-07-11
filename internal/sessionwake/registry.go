// Package sessionwake owns process-local managed-session wake callbacks.
// Package sessionwake 持有进程级受管会话唤醒回调。
package sessionwake

import "sync"

// Callback is one engine-level managed-session wake callback.
// Callback 是单个引擎级受管会话唤醒回调。
type Callback func(engineID uint64) error

// Registry keeps callbacks alive across native registration and dispatch.
// Registry 在原生注册与分发期间保持回调存活。
type Registry struct {
	callbacks sync.Map
}

// Load returns the callback registered for one engine.
// Load 返回为单个引擎注册的回调。
//
// engineID identifies the native engine.
// engineID 标识原生引擎。
// The boolean reports whether a callback exists.
// 布尔值表示回调是否存在。
func (r *Registry) Load(engineID uint64) (Callback, bool) {
	value, ok := r.callbacks.Load(engineID)
	if !ok {
		return nil, false
	}
	return value.(Callback), true
}

// Store installs one callback before native registration can emit a catch-up wake.
// Store 在原生注册可能发出补偿唤醒前安装回调。
//
// engineID identifies the native engine.
// engineID 标识原生引擎。
// callback is the callback retained by the registry.
// callback 是注册表保留的回调。
func (r *Registry) Store(engineID uint64, callback Callback) {
	r.callbacks.Store(engineID, callback)
}

// Delete removes the callback for one engine after native quiescence.
// Delete 在原生侧静默后删除单个引擎的回调。
//
// engineID identifies the native engine.
// engineID 标识原生引擎。
func (r *Registry) Delete(engineID uint64) {
	r.callbacks.Delete(engineID)
}
