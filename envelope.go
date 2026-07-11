package luaskills

import (
	"encoding/json"
	"fmt"
	"strings"
)

// jsonEnvelope is the standard response wrapper returned by public JSON FFI functions.
// jsonEnvelope 是公共 JSON FFI 函数返回的标准响应包络。
type jsonEnvelope struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

// decodeJSONEnvelopeText decodes one JSON FFI response envelope and annotates protocol failures.
// decodeJSONEnvelopeText 解码单个 JSON FFI 响应包络并标注协议失败。
func decodeJSONEnvelopeText(functionName string, text string, out any) error {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return fmt.Errorf("%s: empty JSON FFI response envelope", functionName)
	}
	if !strings.HasPrefix(trimmed, "{") {
		return fmt.Errorf("%s: JSON FFI response envelope must be one object", functionName)
	}
	var envelope jsonEnvelope
	if err := json.Unmarshal([]byte(trimmed), &envelope); err != nil {
		return fmt.Errorf("%s: invalid JSON FFI response envelope: %w", functionName, err)
	}
	if !envelope.OK {
		if envelope.Error == "" {
			envelope.Error = "unknown LuaSkills FFI error"
		}
		return fmt.Errorf("%s: %s", functionName, envelope.Error)
	}
	if out == nil {
		return nil
	}
	if len(envelope.Result) == 0 {
		return fmt.Errorf("%s: missing JSON FFI result payload", functionName)
	}
	if err := json.Unmarshal(envelope.Result, out); err != nil {
		return fmt.Errorf("%s: invalid JSON FFI result payload: %w", functionName, err)
	}
	return nil
}
