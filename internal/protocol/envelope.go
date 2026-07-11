// Package protocol implements the private JSON FFI wire protocol.
// Package protocol 实现私有 JSON FFI 线协议。
package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
)

// envelope is the standard response wrapper returned by public JSON FFI functions.
// envelope 是公共 JSON FFI 函数返回的标准响应包络。
type envelope struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

// DecodeEnvelope decodes one JSON FFI response envelope and annotates protocol failures.
// DecodeEnvelope 解码单个 JSON FFI 响应包络并标注协议失败。
//
// functionName identifies the native function for diagnostic context.
// functionName 标识用于诊断上下文的原生函数。
// text is the exact UTF-8 response returned by the native boundary.
// text 是原生边界返回的精确 UTF-8 响应。
// out receives the decoded result when non-nil.
// out 非空时接收解码后的结果。
func DecodeEnvelope(functionName string, text string, out any) error {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return fmt.Errorf("%s: empty JSON FFI response envelope", functionName)
	}
	if !strings.HasPrefix(trimmed, "{") {
		return fmt.Errorf("%s: JSON FFI response envelope must be one object", functionName)
	}
	var response envelope
	if err := json.Unmarshal([]byte(trimmed), &response); err != nil {
		return fmt.Errorf("%s: invalid JSON FFI response envelope: %w", functionName, err)
	}
	if !response.OK {
		if response.Error == "" {
			response.Error = "unknown LuaSkills FFI error"
		}
		return fmt.Errorf("%s: %s", functionName, response.Error)
	}
	if out == nil {
		return nil
	}
	if len(response.Result) == 0 {
		return fmt.Errorf("%s: missing JSON FFI result payload", functionName)
	}
	if err := json.Unmarshal(response.Result, out); err != nil {
		return fmt.Errorf("%s: invalid JSON FFI result payload: %w", functionName, err)
	}
	return nil
}
