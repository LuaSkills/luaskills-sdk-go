//go:build !cgo

package luaskills

// setNativeSQLiteProviderJSONCallback rejects callback registration without a native C bridge.
// setNativeSQLiteProviderJSONCallback 在没有原生 C 桥时拒绝回调注册。
// callback is unused because no native owner exists; return the explicit cgo requirement.
// 没有原生所有者因此不使用 callback；返回明确的 cgo 要求。
func setNativeSQLiteProviderJSONCallback(callback JSONProviderCallback) error {
	return errCgoRequired()
}
