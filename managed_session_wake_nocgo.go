//go:build !cgo

package luaskills

import "fmt"

// setNativeManagedSessionWakeCallback reports the native bridge requirement when cgo is disabled.
// setNativeManagedSessionWakeCallback 在禁用 cgo 时报告原生桥接要求。
func setNativeManagedSessionWakeCallback(_ uint64, callback ManagedSessionWakeCallback) error {
	if callback == nil {
		return nil
	}
	return fmt.Errorf("managed-session wake callbacks require cgo")
}
