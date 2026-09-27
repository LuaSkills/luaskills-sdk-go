//go:build !cgo

package luaskills

// newEmbeddedNative explicitly rejects native construction when cgo is unavailable; codecs remain usable.
// newEmbeddedNative 在 cgo 不可用时明确拒绝原生构造；编码器仍可使用。
func newEmbeddedNative() (embeddedNative, error) { return nil, errCgoRequired() }
