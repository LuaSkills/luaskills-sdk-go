package luaskills

// ManagedSessionWakeCallback is scheduled when one engine has readable managed-session events.
// ManagedSessionWakeCallback 在某个引擎存在可读受管会话事件时被调度。
type ManagedSessionWakeCallback func(engineID uint64) error

// SetManagedSessionWakeCallback registers, replaces, or clears one engine-level wake callback.
// SetManagedSessionWakeCallback 注册、替换或清除一个引擎级唤醒回调。
func SetManagedSessionWakeCallback(engineID uint64, callback ManagedSessionWakeCallback) error {
	return setNativeManagedSessionWakeCallback(engineID, callback)
}
