package luaskills

import "github.com/LuaSkills/luaskills-sdk-go/internal/embeddedjson"

// EncodeEmbeddedJSON returns bounded JSON bytes for data maps, slices, scalars, json.Number and explicitly tagged structs.
// EncodeEmbeddedJSON 为数据映射、切片、标量、json.Number 及显式标签结构体返回有界 JSON 字节。
// value cannot contain custom marshalers; struct tags support explicit names, omitempty and exclusion only.
// value 不能包含自定义序列化器；结构体标签仅支持显式名称、omitempty 及排除。
// maxBytes is an explicit positive byte limit; untagged/embedded fields and implicit base64 bytes are rejected.
// maxBytes 是显式正数字节上限；无标签／嵌入字段及隐式 base64 字节均被拒绝。
func EncodeEmbeddedJSON(value any, maxBytes uint64) ([]byte, error) {
	return embeddedjson.Encode(value, maxBytes)
}

// DecodeEmbeddedJSON returns exact JSON data from bytes; numbers remain json.Number and missing keys stay absent.
// DecodeEmbeddedJSON 从 bytes 返回精确 JSON 数据；数值保持 json.Number，缺失键保持不存在。
// Invalid Unicode, duplicate decoded keys, trailing input and native numeric overflow return errors.
// 非法 Unicode、解码后重复键、尾随输入及原生数值溢出返回错误。
func DecodeEmbeddedJSON(bytes []byte) (any, error) { return embeddedjson.Decode(bytes) }
