// Package embeddedjson preserves the embedded protocol's numeric and Unicode semantics without serialization hooks.
// embeddedjson 包保留嵌入式协议的数值及 Unicode 语义，不调用序列化钩子。
package embeddedjson

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// visit identifies one currently traversed map or slice view; shared acyclic values remain valid.
// visit 标识当前遍历的映射或切片视图；非循环共享值仍然有效。
type visit struct {
	// valueType distinguishes an aggregate pointer from a differently typed first-field pointer at the same address.
	// valueType 区分地址相同的聚合对象指针与不同类型的首字段指针。
	valueType reflect.Type
	// pointer is used only for equality while the original Go object remains live.
	// pointer 仅在原始 Go 对象存活时用于相等比较。
	pointer uintptr
	// length distinguishes independent slice views sharing their first element.
	// length 区分共享首元素的独立切片视图。
	length int
}

// Encode freezes value into newly owned JSON bytes within limit; unsupported Go types and cycles return errors.
// Encode 在 limit 内将 value 冻结为新拥有的 JSON 字节；不支持的 Go 类型及循环返回错误。
// Numeric Go integers retain all bits; floats always carry a decimal/exponent marker, including negative zero.
// Go 整数保留全部位；浮点数始终携带小数／指数标记，包括负零。
func Encode(value any, limit uint64) ([]byte, error) {
	if limit == 0 || limit > uint64(^uint(0)>>1) {
		return nil, fmt.Errorf("embedded JSON byte limit must fit a positive Go int")
	}
	var output bytes.Buffer
	ancestors := make(map[visit]bool)
	appendToken := func(token string) error {
		if uint64(len(token)) > limit-uint64(output.Len()) {
			return fmt.Errorf("embedded JSON exceeds byte limit")
		}
		output.WriteString(token)
		return nil
	}
	appendString := func(value string) error {
		if !utf8.ValidString(value) {
			return fmt.Errorf("embedded JSON string is not valid UTF-8")
		}
		if uint64(len(value)) > limit-uint64(output.Len()) {
			return fmt.Errorf("embedded JSON exceeds byte limit")
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return appendToken(string(encoded))
	}
	var write func(any) error
	write = func(value any) error {
		switch value := value.(type) {
		case nil:
			return appendToken("null")
		case bool:
			return appendToken(strconv.FormatBool(value))
		case string:
			return appendString(value)
		case json.Number:
			if err := validateNumber(string(value)); err != nil {
				return err
			}
			return appendToken(string(value))
		case int:
			return appendToken(strconv.FormatInt(int64(value), 10))
		case int8:
			return appendToken(strconv.FormatInt(int64(value), 10))
		case int16:
			return appendToken(strconv.FormatInt(int64(value), 10))
		case int32:
			return appendToken(strconv.FormatInt(int64(value), 10))
		case int64:
			return appendToken(strconv.FormatInt(value, 10))
		case uint:
			return appendToken(strconv.FormatUint(uint64(value), 10))
		case uint8:
			return appendToken(strconv.FormatUint(uint64(value), 10))
		case uint16:
			return appendToken(strconv.FormatUint(uint64(value), 10))
		case uint32:
			return appendToken(strconv.FormatUint(uint64(value), 10))
		case uint64:
			return appendToken(strconv.FormatUint(value, 10))
		case float32:
			return write(float64(value))
		case float64:
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("embedded JSON requires finite floats")
			}
			token := strconv.FormatFloat(value, 'g', -1, 64)
			if !strings.ContainsAny(token, ".eE") {
				token += ".0"
			}
			return appendToken(token)
		case []any:
			if value == nil {
				return appendToken("null")
			}
			key := visit{reflect.TypeOf(value), reflect.ValueOf(value).Pointer(), len(value)}
			if ancestors[key] {
				return fmt.Errorf("embedded JSON contains a cycle")
			}
			ancestors[key] = true
			defer delete(ancestors, key)
			if err := appendToken("["); err != nil {
				return err
			}
			for index, child := range value {
				if index != 0 {
					if err := appendToken(","); err != nil {
						return err
					}
				}
				if err := write(child); err != nil {
					return err
				}
			}
			return appendToken("]")
		case map[string]any:
			if value == nil {
				return appendToken("null")
			}
			key := visit{reflect.TypeOf(value), uintptr(reflect.ValueOf(value).UnsafePointer()), 0}
			if ancestors[key] {
				return fmt.Errorf("embedded JSON contains a cycle")
			}
			ancestors[key] = true
			defer delete(ancestors, key)
			if err := appendToken("{"); err != nil {
				return err
			}
			keys := make([]string, 0, len(value))
			for key := range value {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for index, key := range keys {
				if index != 0 {
					if err := appendToken(","); err != nil {
						return err
					}
				}
				if err := appendString(key); err != nil {
					return err
				}
				if err := appendToken(":"); err != nil {
					return err
				}
				if err := write(value[key]); err != nil {
					return err
				}
			}
			return appendToken("}")
		default:
			// Existing SDK options use tagged structs and typed slices; inspect data directly without executing hooks.
			// 既有 SDK 选项使用带标签结构体及类型化切片；直接检查数据，不执行钩子。
			if _, ok := value.(json.Marshaler); ok {
				return fmt.Errorf("embedded JSON forbids serialization hooks")
			}
			if _, ok := value.(encoding.TextMarshaler); ok {
				return fmt.Errorf("embedded JSON forbids text serialization hooks")
			}
			rv := reflect.ValueOf(value)
			var key visit
			switch rv.Kind() {
			case reflect.Pointer, reflect.Map:
				if rv.IsNil() {
					return appendToken("null")
				}
				key = visit{rv.Type(), uintptr(rv.UnsafePointer()), 0}
			case reflect.Slice:
				if rv.IsNil() {
					return appendToken("null")
				}
				// Empty typed slices may share Go's zero-sized allocation address; they cannot contain a cycle.
				// 空类型化切片可能共享 Go 零大小分配地址；它们不可能包含循环。
				if rv.Len() == 0 && rv.Type().Elem().Kind() != reflect.Uint8 {
					return appendToken("[]")
				}
				key = visit{rv.Type(), rv.Pointer(), rv.Len()}
			}
			if key.pointer != 0 {
				if ancestors[key] {
					return fmt.Errorf("embedded JSON contains a cycle")
				}
				ancestors[key] = true
				defer delete(ancestors, key)
			}
			plain, err := plainValue(rv, limit-uint64(output.Len()))
			if err != nil {
				return err
			}
			return write(plain)
		}
	}
	if err := write(value); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

// plainValue projects one typed Go layer into plain JSON values with explicit tags and no custom method calls.
// plainValue 使用显式标签将一层类型化 Go 数据投影为普通 JSON 值，不调用自定义方法。
// remaining bounds temporary member arrays; recursive traversal and cycle detection belong to Encode.
// remaining 限制临时成员数组；递归遍历及循环检测由 Encode 负责。
func plainValue(value reflect.Value, remaining uint64) (any, error) {
	switch value.Kind() {
	case reflect.Pointer:
		return value.Elem().Interface(), nil
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.String:
		return value.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return value.Float(), nil
	case reflect.Slice, reflect.Array:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			return nil, fmt.Errorf("embedded JSON requires explicit byte representation")
		}
		if uint64(value.Len()) > remaining {
			return nil, fmt.Errorf("embedded JSON exceeds byte limit")
		}
		result := make([]any, value.Len())
		for index := range result {
			result[index] = value.Index(index).Interface()
		}
		return result, nil
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("embedded JSON requires string map keys")
		}
		if uint64(value.Len()) > remaining {
			return nil, fmt.Errorf("embedded JSON exceeds byte limit")
		}
		result := make(map[string]any, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			result[iter.Key().String()] = iter.Value().Interface()
		}
		return result, nil
	case reflect.Struct:
		result := make(map[string]any)
		seen := make(map[string]bool)
		for index := 0; index < value.NumField(); index++ {
			field := value.Type().Field(index)
			tag, present := field.Tag.Lookup("json")
			if tag == "-" {
				continue
			}
			if !present || field.PkgPath != "" || field.Anonymous {
				return nil, fmt.Errorf("embedded JSON requires explicitly tagged exported struct fields")
			}
			parts := strings.Split(tag, ",")
			name := parts[0]
			if name == "" || seen[name] {
				return nil, fmt.Errorf("invalid or duplicate embedded JSON field tag")
			}
			seen[name] = true
			omit := false
			for _, option := range parts[1:] {
				if option != "omitempty" || omit {
					return nil, fmt.Errorf("unsupported embedded JSON field tag option")
				}
				omit = true
			}
			child := value.Field(index)
			if omit && emptyValue(child) {
				continue
			}
			// Only emitted fields consume the lower-bound member budget; omitted fields can still encode as {}.
			// 只有输出字段消耗成员数下界预算；省略字段仍然可以编码为 {}。
			if uint64(len(result)) >= remaining {
				return nil, fmt.Errorf("embedded JSON exceeds byte limit")
			}
			result[name] = child.Interface()
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported embedded JSON Go type %s", value.Type())
	}
}

// emptyValue implements the documented omitempty zero forms; structs remain present as in encoding/json.
// emptyValue 实现已声明的 omitempty 零值形状；结构体与 encoding/json 一样保持存在。
func emptyValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return value.Len() == 0
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Interface, reflect.Pointer:
		return value.IsZero()
	}
	return false
}

// validateNumber accepts one exact signed/unsigned 64-bit integer or finite IEEE-754 JSON float token.
// validateNumber 接受一个精确有符号／无符号 64 位整数或有限 IEEE-754 JSON 浮点词元。
// token is validated before output or publication; out-of-range integer text is never converted to float.
// token 在输出或发布前校验；超范围整数文本绝不转换为浮点数。
func validateNumber(token string) error {
	if !json.Valid([]byte(token)) || len(token) == 0 || (token[0] != '-' && (token[0] < '0' || token[0] > '9')) {
		return fmt.Errorf("invalid embedded JSON number")
	}
	var err error
	if strings.ContainsAny(token, ".eE") {
		var value float64
		value, err = strconv.ParseFloat(token, 64)
		if math.IsInf(value, 0) || math.IsNaN(value) {
			return fmt.Errorf("embedded JSON number is not finite")
		}
	} else if token[0] == '-' {
		_, err = strconv.ParseInt(token, 10, 64)
	} else {
		_, err = strconv.ParseUint(token, 10, 64)
	}
	if err != nil {
		return fmt.Errorf("embedded JSON number outside native range: %w", err)
	}
	return nil
}

// Decode returns maps, slices and json.Number values without losing integer width, float intent or explicit null.
// Decode 返回映射、切片及 json.Number，不丢失整数位宽、浮点意图或显式空值。
// bytes must contain exactly one valid UTF-8 value; duplicate decoded keys and unpaired surrogates are errors.
// bytes 必须包含恰好一个有效 UTF-8 值；解码后重复键及未配对代理项均报错。
func Decode(data []byte) (any, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("embedded JSON is not valid UTF-8")
	}
	if err := validateSurrogates(data); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var read func() (any, error)
	read = func() (any, error) {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch token := token.(type) {
		case json.Number:
			if err := validateNumber(string(token)); err != nil {
				return nil, err
			}
			return token, nil
		case json.Delim:
			switch token {
			case '{':
				value := make(map[string]any)
				for decoder.More() {
					keyToken, err := decoder.Token()
					if err != nil {
						return nil, err
					}
					key, ok := keyToken.(string)
					if !ok {
						return nil, fmt.Errorf("embedded JSON object key is not a string")
					}
					if _, duplicate := value[key]; duplicate {
						return nil, fmt.Errorf("duplicate embedded JSON object key %q", key)
					}
					child, err := read()
					if err != nil {
						return nil, err
					}
					value[key] = child
				}
				ending, err := decoder.Token()
				if err != nil || ending != json.Delim('}') {
					return nil, fmt.Errorf("invalid embedded JSON object ending: %v", err)
				}
				return value, nil
			case '[':
				value := make([]any, 0)
				for decoder.More() {
					child, err := read()
					if err != nil {
						return nil, err
					}
					value = append(value, child)
				}
				ending, err := decoder.Token()
				if err != nil || ending != json.Delim(']') {
					return nil, fmt.Errorf("invalid embedded JSON array ending: %v", err)
				}
				return value, nil
			default:
				return nil, fmt.Errorf("unexpected embedded JSON delimiter")
			}
		default:
			return token, nil
		}
	}
	value, err := read()
	if err != nil {
		return nil, err
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("embedded JSON has trailing content")
	}
	return value, nil
}

// validateSurrogates prevents encoding/json from silently replacing malformed UTF-16 escapes in input.
// validateSurrogates 阻止 encoding/json 静默替换输入中格式错误的 UTF-16 转义。
// data is scanned without modification; ordinary JSON syntax remains the decoder's responsibility.
// 扫描 data 而不修改；普通 JSON 语法仍由解码器负责。
func validateSurrogates(data []byte) error {
	inString := false
	for index := 0; index < len(data); index++ {
		if data[index] == '"' {
			inString = !inString
			continue
		}
		if !inString || data[index] != '\\' {
			continue
		}
		index++
		if index >= len(data) {
			return fmt.Errorf("unfinished embedded JSON escape")
		}
		if data[index] != 'u' {
			continue
		}
		if index+4 >= len(data) {
			return fmt.Errorf("short embedded JSON Unicode escape")
		}
		unit, err := strconv.ParseUint(string(data[index+1:index+5]), 16, 16)
		if err != nil {
			return fmt.Errorf("invalid embedded JSON Unicode escape")
		}
		index += 4
		if unit >= 0xdc00 && unit <= 0xdfff {
			return fmt.Errorf("unpaired embedded JSON low surrogate")
		}
		if unit < 0xd800 || unit > 0xdbff {
			continue
		}
		if index+6 >= len(data) || data[index+1] != '\\' || data[index+2] != 'u' {
			return fmt.Errorf("unpaired embedded JSON high surrogate")
		}
		low, err := strconv.ParseUint(string(data[index+3:index+7]), 16, 16)
		if err != nil || low < 0xdc00 || low > 0xdfff {
			return fmt.Errorf("invalid embedded JSON surrogate pair")
		}
		index += 6
	}
	return nil
}
