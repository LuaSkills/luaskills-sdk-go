package luaskills

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// EmbeddedJSONNull represents a null-only schema; its pointer must be nil, never an allocated empty object.
// EmbeddedJSONNull 表示只能为空值的 Schema；其指针必须为空，绝不能是已分配空对象。
type EmbeddedJSONNull struct{}

// embeddedWireShape records constraints derived from one generated schema type, never from caller-provided guesses.
// embeddedWireShape 记录从一个生成 Schema 类型派生的约束，绝不依赖调用方猜测。
type embeddedWireShape struct {
	// kind selects an exact generated representation.
	// kind 选择精确生成表示。
	kind string
	// values lists the exact string enumeration members.
	// values 列出精确字符串枚举成员。
	values []string
	// alternatives contains the concrete generated members of a sealed interface.
	// alternatives 包含封闭接口的具体生成成员。
	alternatives []reflect.Type
	// additional follows the schema's explicit or default additional-properties rule.
	// additional 遵循 Schema 的显式或默认额外属性规则。
	additional bool
	// unique marks the contract's string-set arrays.
	// unique 标记契约的字符串集合数组。
	unique bool
}

// embeddedWireType returns T's static type without constructing or invoking a value of T.
// embeddedWireType 返回 T 的静态类型，不构造或调用 T 的值。
func embeddedWireType[T any]() reflect.Type { return reflect.TypeOf((*T)(nil)).Elem() }

// projectEmbeddedWire converts already decoded data to one generated type, preserving optional-field presence.
// projectEmbeddedWire 将已解码数据转换为一个生成类型，并保留可选字段存在性。
// destination is a generated data type; path identifies failures without including business value contents.
// destination 是生成数据类型；path 标识失败位置，不包含业务值内容。
func projectEmbeddedWire(value any, destination reflect.Type, path string) (reflect.Value, error) {
	zero := reflect.Zero(destination)
	invalid := func(detail string) (reflect.Value, error) {
		return zero, fmt.Errorf("embedded wire %s: %s", path, detail)
	}
	if destination == embeddedWireType[*EmbeddedJSONNull]() {
		if value != nil {
			return invalid("expected null")
		}
		return zero, nil
	}
	if destination.Kind() == reflect.Pointer {
		if value == nil {
			return zero, nil
		}
		child, err := projectEmbeddedWire(value, destination.Elem(), path)
		if err != nil {
			return zero, err
		}
		result := reflect.New(destination.Elem())
		result.Elem().Set(child)
		return result, nil
	}
	shape, known := embeddedWireShapes[destination]
	if known && shape.kind == "union" {
		var selected reflect.Value
		matches := 0
		for _, alternative := range shape.alternatives {
			child, err := projectEmbeddedWire(value, alternative, path)
			if err == nil {
				selected = child
				matches++
			}
		}
		if matches != 1 {
			return invalid("no unique declared union alternative")
		}
		result := reflect.New(destination).Elem()
		result.Set(selected)
		return result, nil
	}
	if destination.Kind() == reflect.Interface && destination.NumMethod() == 0 {
		if value == nil {
			return zero, nil
		}
		return reflect.ValueOf(value), nil
	}
	if value == nil {
		return invalid("unexpected null")
	}
	result := reflect.New(destination).Elem()
	switch destination.Kind() {
	case reflect.String:
		text, ok := value.(string)
		if !ok {
			return invalid("expected string")
		}
		if known && shape.kind == "enum" {
			found := false
			for _, candidate := range shape.values {
				if text == candidate {
					found = true
					break
				}
			}
			if !found {
				return invalid("unknown enumeration member")
			}
		}
		result.SetString(text)
	case reflect.Bool:
		boolean, ok := value.(bool)
		if !ok {
			return invalid("expected boolean")
		}
		result.SetBool(boolean)
	case reflect.Uint32, reflect.Uint64:
		number, ok := value.(json.Number)
		if !ok {
			return invalid("expected exact unsigned integer")
		}
		integer, err := strconv.ParseUint(string(number), 10, destination.Bits())
		if err != nil {
			return invalid("unsigned integer out of range or non-integer token")
		}
		result.SetUint(integer)
	case reflect.Struct:
		if !known || shape.kind != "object" {
			return invalid("unregistered object type")
		}
		object, ok := value.(map[string]any)
		if !ok {
			return invalid("expected object")
		}
		fields := make(map[string]bool, destination.NumField())
		for index := 0; index < destination.NumField(); index++ {
			field := destination.Field(index)
			tag := strings.Split(field.Tag.Get("json"), ",")
			key := tag[0]
			fields[key] = true
			child, present := object[key]
			optional := len(tag) == 2 && tag[1] == "omitempty"
			if !present {
				if !optional {
					return invalid("missing required field " + key)
				}
				continue
			}
			fieldType := field.Type
			// The outer optional pointer records presence even when the inner nullable value is nil.
			// 外层可选指针记录存在性，即使内层可空值为 nil。
			if optional {
				fieldType = fieldType.Elem()
			}
			projected, err := projectEmbeddedWire(child, fieldType, path+"."+key)
			if err != nil {
				return zero, err
			}
			if optional {
				outer := reflect.New(fieldType)
				outer.Elem().Set(projected)
				result.Field(index).Set(outer)
			} else {
				result.Field(index).Set(projected)
			}
		}
		if !shape.additional {
			for key := range object {
				if !fields[key] {
					return invalid("unknown field " + key)
				}
			}
		}
	case reflect.Slice:
		if !known || shape.kind != "array" {
			return invalid("unregistered array type")
		}
		items, ok := value.([]any)
		if !ok {
			return invalid("expected array")
		}
		result = reflect.MakeSlice(destination, len(items), len(items))
		seen := map[string]bool{}
		for index, item := range items {
			child, err := projectEmbeddedWire(item, destination.Elem(), fmt.Sprintf("%s[%d]", path, index))
			if err != nil {
				return zero, err
			}
			if shape.unique {
				key, ok := item.(string)
				if !ok {
					return invalid("unique array requires strings")
				}
				if seen[key] {
					return invalid("duplicate set member")
				}
				seen[key] = true
			}
			result.Index(index).Set(child)
		}
	default:
		return invalid("unsupported generated destination")
	}
	return result, nil
}

// decodeEmbeddedWireEnvelope validates exact protocol identity before projecting a generated success or error envelope.
// decodeEmbeddedWireEnvelope 在投影生成的成功或错误信封前校验精确协议身份。
// bytes remain caller-owned; the returned value owns all decoded data and invokes no custom unmarshal methods.
// bytes 仍由调用方拥有；返回值拥有全部解码数据，不调用自定义反序列化方法。
func decodeEmbeddedWireEnvelope[T any](bytes []byte) (T, error) {
	var empty T
	decoded, err := DecodeEmbeddedJSON(bytes)
	if err != nil {
		return empty, err
	}
	envelope, ok := decoded.(map[string]any)
	if !ok || len(envelope) != 3 || envelope["protocol_version"] != json.Number(strconv.FormatUint(uint64(EmbeddedProtocolVersion), 10)) {
		return empty, fmt.Errorf("invalid embedded response protocol envelope")
	}
	result, err := projectEmbeddedWire(decoded, embeddedWireType[T](), "response")
	if err != nil {
		return empty, err
	}
	return result.Interface().(T), nil
}

// decodeEmbeddedWireValue projects a standalone decoded value through the same generated structural rules.
// decodeEmbeddedWireValue 通过相同生成结构规则投影独立已解码值。
// It returns owned data or a shape error; compatibility policy is evaluated separately before native creation.
// 返回拥有型数据或形状错误；兼容策略在原生创建前单独判断。
func decodeEmbeddedWireValue[T any](bytes []byte) (T, error) {
	var empty T
	decoded, err := DecodeEmbeddedJSON(bytes)
	if err != nil {
		return empty, err
	}
	result, err := projectEmbeddedWire(decoded, embeddedWireType[T](), "description")
	if err != nil {
		return empty, err
	}
	return result.Interface().(T), nil
}

// EncodeEmbeddedRequest freezes a generated input request within maxBytes and validates its exact declared shape.
// EncodeEmbeddedRequest 在 maxBytes 内冻结生成输入请求，并校验其精确声明形状。
// Optional nil pointers mean absence; a non-nil outer pointer can preserve an explicit inner null.
// 可选空指针表示缺失；非空外层指针可保留显式内层空值。
func EncodeEmbeddedRequest(request EmbeddedInputRequest, maxBytes uint64) ([]byte, error) {
	if request.ProtocolVersion != EmbeddedProtocolVersion {
		return nil, fmt.Errorf("unsupported embedded request protocol version")
	}
	bytes, err := EncodeEmbeddedJSON(request, maxBytes)
	if err != nil {
		return nil, err
	}
	decoded, err := DecodeEmbeddedJSON(bytes)
	if err != nil {
		return nil, err
	}
	if _, err = projectEmbeddedWire(decoded, embeddedWireType[EmbeddedInputRequest](), "request"); err != nil {
		return nil, err
	}
	return bytes, nil
}
