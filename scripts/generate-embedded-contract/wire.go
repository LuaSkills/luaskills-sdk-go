package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

// wireGenerator owns deterministic declarations and projection rules for one complete contract.
// wireGenerator 拥有一个完整契约的确定性声明及投影规则。
type wireGenerator struct {
	// declarations stores source by globally unique generated name.
	// declarations 按全局唯一生成名称保存源码。
	declarations map[string]string
	// schemas detects collisions before any generated file is written.
	// schemas 在写入任何生成文件前检测冲突。
	schemas map[string]map[string]any
	// rules describes only generated types; arbitrary user methods are never called.
	// rules 仅描述生成类型；绝不调用任意用户方法。
	rules map[string]string
	// names maps exact root-local references into the current input/output namespace.
	// names 将精确根局部引用映射到当前输入／输出命名空间。
	names map[string]string
}

// wireIdentifier converts declared ASCII contract names to exported Go identifiers, rejecting ambiguous characters.
// wireIdentifier 将已声明 ASCII 契约名称转换为导出 Go 标识符，并拒绝歧义字符。
func wireIdentifier(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty wire identifier")
	}
	var result strings.Builder
	upper := true
	for _, char := range name {
		if char == '_' {
			if upper {
				return "", fmt.Errorf("ambiguous wire identifier %q", name)
			}
			upper = true
			continue
		}
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9' && result.Len() != 0)) {
			return "", fmt.Errorf("unsupported wire identifier %q", name)
		}
		if upper {
			char = unicode.ToUpper(char)
			upper = false
		}
		result.WriteRune(char)
	}
	if upper {
		return "", fmt.Errorf("trailing separator in wire identifier %q", name)
	}
	return result.String(), nil
}

// wireKeys returns sorted map keys so source and collision checks are independent from Go map iteration.
// wireKeys 返回排序后的映射键，使源码及冲突检查独立于 Go 映射遍历顺序。
func wireKeys[T any](values map[string]T) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

// wireDoc preserves the authoritative bilingual description, or emits a bilingual generated declaration explanation.
// wireDoc 保留权威双语描述，或输出双语生成声明说明。
func wireDoc(schema map[string]any, name string) string {
	if description, ok := schema["description"].(string); ok && description != "" {
		return "// " + strings.ReplaceAll(description, "\n", "\n// ") + "\n"
	}
	return fmt.Sprintf("// %s is derived from the packaged wire contract.\n// %s 从包内线契约派生。\n", name, name)
}

// verifyWireSchema rejects unsupported combinations and references before generation can weaken the upstream shape.
// verifyWireSchema 在生成可能弱化上游形状之前拒绝不支持的组合及引用。
func verifyWireSchema(schema map[string]any, definitions map[string]any, root bool) error {
	annotations := map[string]bool{"description": true, "title": true, "default": true, "$schema": true, "$defs": true}
	allowed := map[string]bool{}
	for key := range annotations {
		allowed[key] = true
	}
	if _, present := schema["$defs"]; present && !root {
		return fmt.Errorf("nested definition namespace")
	}
	if dialect, present := schema["$schema"]; present && dialect != "https://json-schema.org/draft/2020-12/schema" {
		return fmt.Errorf("unsupported schema dialect")
	}
	selector := ""
	for _, key := range []string{"$ref", "oneOf", "anyOf"} {
		if _, present := schema[key]; present {
			if selector != "" {
				return fmt.Errorf("combined schema selectors")
			}
			selector = key
		}
	}
	if selector != "" {
		allowed[selector] = true
		if selector == "$ref" {
			ref, ok := schema[selector].(string)
			if !ok || !strings.HasPrefix(ref, "#/$defs/") || definitions[strings.TrimPrefix(ref, "#/$defs/")] == nil {
				return fmt.Errorf("unresolved root-local reference %v", schema[selector])
			}
		} else {
			variants, ok := schema[selector].([]any)
			if !ok || len(variants) == 0 {
				return fmt.Errorf("invalid schema alternatives")
			}
			for _, variant := range variants {
				child, ok := variant.(map[string]any)
				if !ok {
					return fmt.Errorf("non-object alternative")
				}
				if err := verifyWireSchema(child, definitions, false); err != nil {
					return err
				}
			}
		}
	} else if constant, present := schema["const"]; present {
		allowed["const"], allowed["type"] = true, true
		if _, ok := constant.(string); !ok || schema["type"] != "string" {
			return fmt.Errorf("unsupported constant shape")
		}
	} else if rawType, present := schema["type"]; present {
		allowed["type"] = true
		kind, single := rawType.(string)
		if !single {
			types, ok := rawType.([]any)
			if !ok || len(types) != 2 || types[1] != "null" || types[0] == "null" {
				return fmt.Errorf("unsupported type union")
			}
			kind, ok = types[0].(string)
			if !ok {
				return fmt.Errorf("invalid type union")
			}
		}
		switch kind {
		case "string", "boolean", "null":
		case "integer":
			allowed["format"], allowed["minimum"] = true, true
			if schema["minimum"] != json.Number("0") || (schema["format"] != "uint" && schema["format"] != "uint32" && schema["format"] != "uint64") {
				return fmt.Errorf("unsupported integer constraints")
			}
		case "array":
			allowed["items"], allowed["uniqueItems"] = true, true
			items, ok := schema["items"].(map[string]any)
			if !ok {
				return fmt.Errorf("array has no item schema")
			}
			if err := verifyWireSchema(items, definitions, false); err != nil {
				return err
			}
			if unique, present := schema["uniqueItems"]; present {
				if _, ok := unique.(bool); !ok {
					return fmt.Errorf("invalid uniqueness flag")
				}
				if unique == true && items["type"] != "string" {
					return fmt.Errorf("unsupported non-string unique array")
				}
			}
		case "object":
			allowed["properties"], allowed["required"], allowed["additionalProperties"] = true, true, true
			properties := map[string]any{}
			if raw, present := schema["properties"]; present {
				var ok bool
				properties, ok = raw.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid properties")
				}
			}
			seen := map[string]bool{}
			if raw, present := schema["required"]; present {
				fields, ok := raw.([]any)
				if !ok {
					return fmt.Errorf("invalid required fields")
				}
				for _, field := range fields {
					name, ok := field.(string)
					if !ok || seen[name] || properties[name] == nil {
						return fmt.Errorf("invalid required field")
					}
					seen[name] = true
				}
			}
			if extra, present := schema["additionalProperties"]; present {
				if _, ok := extra.(bool); !ok {
					return fmt.Errorf("unsupported additional properties schema")
				}
			}
			for _, raw := range properties {
				field, ok := raw.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid property schema")
				}
				if err := verifyWireSchema(field, definitions, false); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unsupported wire type %q", kind)
		}
	}
	for key := range schema {
		if !allowed[key] {
			return fmt.Errorf("unsupported schema keyword or combination %q", key)
		}
	}
	if root {
		for _, raw := range definitions {
			definition, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid definition")
			}
			if err := verifyWireSchema(definition, definitions, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// expression emits nested declarations as needed and returns a type preserving both nullability and optionality.
// expression 按需输出嵌套声明，返回保留可空性及可选性的类型。
func (g *wireGenerator) expression(schema map[string]any, name string) (string, error) {
	if ref, ok := schema["$ref"].(string); ok {
		result, found := g.names[strings.TrimPrefix(ref, "#/$defs/")]
		if !found {
			return "", fmt.Errorf("missing reference %s", ref)
		}
		return result, nil
	}
	if types, ok := schema["type"].([]any); ok {
		child := map[string]any{}
		for key, value := range schema {
			child[key] = value
		}
		child["type"] = types[0]
		value, err := g.expression(child, name+"Value")
		return "*" + value, err
	}
	if variants, ok := schema["anyOf"].([]any); ok && len(variants) == 2 && variants[1].(map[string]any)["type"] == "null" {
		value, err := g.expression(variants[0].(map[string]any), name+"Value")
		return "*" + value, err
	}
	if _, ok := schema["const"]; ok {
		if err := g.declare(name, schema); err != nil {
			return "", err
		}
		return name, nil
	}
	if _, ok := schema["oneOf"]; ok {
		if err := g.declare(name, schema); err != nil {
			return "", err
		}
		return name, nil
	}
	if _, ok := schema["anyOf"]; ok {
		if err := g.declare(name, schema); err != nil {
			return "", err
		}
		return name, nil
	}
	switch schema["type"] {
	case nil:
		return "any", nil
	case "string":
		return "string", nil
	case "boolean":
		return "bool", nil
	case "integer":
		if schema["format"] == "uint32" {
			return "uint32", nil
		}
		return "uint64", nil
	case "null":
		return "*EmbeddedJSONNull", nil
	case "object", "array":
		if err := g.declare(name, schema); err != nil {
			return "", err
		}
		return name, nil
	default:
		return "", fmt.Errorf("unsupported type expression %v", schema["type"])
	}
}

// declare reserves a unique name, emits its exact data representation and records safe reflection projection rules.
// declare 预留唯一名称，输出精确数据表示，并记录安全反射投影规则。
func (g *wireGenerator) declare(name string, schema map[string]any) error {
	if _, exists := g.schemas[name]; exists {
		return fmt.Errorf("generated name collision: %s", name)
	}
	g.schemas[name] = schema
	var output strings.Builder
	output.WriteString(wireDoc(schema, name))
	variants, union := schema["oneOf"].([]any)
	if other, ok := schema["anyOf"].([]any); ok {
		variants, union = other, true
	}
	if _, ok := schema["const"]; ok {
		variants, union = []any{schema}, true
	}
	if union {
		stringsOnly := true
		for _, variant := range variants {
			if _, ok := variant.(map[string]any)["const"].(string); !ok {
				stringsOnly = false
			}
		}
		if stringsOnly {
			fmt.Fprintf(&output, "type %s string\nconst (\n", name)
			values := []string{}
			seen := map[string]bool{}
			for _, variant := range variants {
				branch := variant.(map[string]any)
				value := branch["const"].(string)
				suffix, err := wireIdentifier(value)
				if err != nil {
					return err
				}
				if seen[suffix] {
					return fmt.Errorf("enum collision in %s", name)
				}
				seen[suffix] = true
				constant := name + suffix
				output.WriteString(wireDoc(branch, constant))
				fmt.Fprintf(&output, "%s %s = %q\n", constant, name, value)
				values = append(values, fmt.Sprintf("%q", value))
			}
			output.WriteString(")\n")
			g.rules[name] = "{kind: \"enum\", values: []string{" + strings.Join(values, ",") + "}}"
		} else {
			// anyOf may permit multiple simultaneous matches; sealed Go variants require a provably unique shape.
			// anyOf 可能允许同时匹配多项；封闭 Go 分支要求可证明的唯一形状。
			if _, anyOf := schema["anyOf"]; anyOf {
				for index, first := range variants {
					for _, second := range variants[index+1:] {
						if !wireDisjointObjects(first.(map[string]any), second.(map[string]any)) {
							return fmt.Errorf("overlapping or unproven anyOf alternatives in %s", name)
						}
					}
				}
			}
			method := "embeddedVariant" + name
			fmt.Fprintf(&output, "type %s interface {\n// %s seals this generated union without invoking serialization hooks.\n// %s 封闭此生成联合，不调用序列化钩子。\n%s()\n}\n", name, method, method, method)
			branches := []string{}
			for index, variant := range variants {
				branch := variant.(map[string]any)
				if branch["type"] != "object" {
					return fmt.Errorf("unsupported mixed union %s", name)
				}
				suffix := fmt.Sprintf("Variant%d", index+1)
				if properties, ok := branch["properties"].(map[string]any); ok {
					if raw, ok := properties["type"].(map[string]any); ok {
						if value, ok := raw["const"].(string); ok {
							var err error
							suffix, err = wireIdentifier(value)
							if err != nil {
								return err
							}
						}
					}
				}
				variantName := name + suffix
				if err := g.declare(variantName, branch); err != nil {
					return err
				}
				fmt.Fprintf(&output, "// %s marks the exact %s alternative.\n// %s 标识精确的 %s 分支。\nfunc (%s) %s() {}\n", method, variantName, method, variantName, variantName, method)
				branches = append(branches, "embeddedWireType["+variantName+"]()")
			}
			g.rules[name] = "{kind: \"union\", alternatives: []reflect.Type{" + strings.Join(branches, ",") + "}}"
		}
	} else if schema["type"] == "object" {
		properties, _ := schema["properties"].(map[string]any)
		required := map[string]bool{}
		if fields, ok := schema["required"].([]any); ok {
			for _, field := range fields {
				required[field.(string)] = true
			}
		}
		fmt.Fprintf(&output, "type %s struct {\n", name)
		fields := map[string]bool{}
		for _, key := range wireKeys(properties) {
			field, err := wireIdentifier(key)
			if err != nil {
				return err
			}
			if fields[field] {
				return fmt.Errorf("field collision %s.%s", name, field)
			}
			fields[field] = true
			child := properties[key].(map[string]any)
			value, err := g.expression(child, name+field)
			if err != nil {
				return err
			}
			tag := key
			if !required[key] {
				value = "*" + value
				tag += ",omitempty"
			}
			output.WriteString(wireDoc(child, field))
			fmt.Fprintf(&output, "%s %s `json:%q`\n", field, value, tag)
		}
		output.WriteString("}\n")
		g.rules[name] = fmt.Sprintf("{kind: \"object\", additional: %t}", schema["additionalProperties"] != false)
	} else if schema["type"] == "array" {
		item, err := g.expression(schema["items"].(map[string]any), name+"Item")
		if err != nil {
			return err
		}
		fmt.Fprintf(&output, "type %s []%s\n", name, item)
		g.rules[name] = fmt.Sprintf("{kind: \"array\", unique: %t}", schema["uniqueItems"] == true)
	} else {
		return fmt.Errorf("unsupported declaration %s", name)
	}
	g.declarations[name] = output.String()
	return nil
}

// wireDisjointObjects proves exclusive constant selectors or a required field forbidden by the other closed shape.
// wireDisjointObjects 证明互斥常量选择器，或另一封闭形状禁止的必需字段。
// Both inputs are verified object schemas; false means this generator cannot represent their anyOf unambiguously.
// 两个输入均为已校验对象 Schema；false 表示此生成器无法无歧义表示其 anyOf。
func wireDisjointObjects(first, second map[string]any) bool {
	if first["type"] != "object" || second["type"] != "object" {
		return false
	}
	for _, pair := range [][2]map[string]any{{first, second}, {second, first}} {
		properties, _ := pair[0]["properties"].(map[string]any)
		other, _ := pair[1]["properties"].(map[string]any)
		required, _ := pair[0]["required"].([]any)
		otherRequired := map[string]bool{}
		if fields, ok := pair[1]["required"].([]any); ok {
			for _, field := range fields {
				otherRequired[field.(string)] = true
			}
		}
		for _, raw := range required {
			field := raw.(string)
			if _, present := other[field]; !present && pair[1]["additionalProperties"] == false {
				return true
			}
			left, leftOK := properties[field].(map[string]any)
			right, rightOK := other[field].(map[string]any)
			if leftOK && rightOK && otherRequired[field] {
				a, aOK := left["const"]
				b, bOK := right["const"]
				if aOK && bOK && !reflect.DeepEqual(a, b) {
					return true
				}
			}
		}
	}
	return false
}

// wireObject reads one exact required JSON object; malformed metadata is an error rather than a type-assertion panic.
// wireObject 读取一个精确必需 JSON 对象；元数据格式错误时返回错误，而非类型断言崩溃。
func wireObject(parent map[string]any, key string) (map[string]any, error) {
	value, ok := parent[key].(map[string]any)
	if !ok || value == nil {
		return nil, fmt.Errorf("missing or invalid contract object %s", key)
	}
	return value, nil
}

// generateWire validates independent roots and emits all input/output declarations before touching disk.
// generateWire 校验独立根，并在修改磁盘前输出全部输入／输出声明。
func generateWire(data []byte) ([]byte, error) {
	// document is already losslessly syntax-checked by run; UseNumber preserves numeric schema constraints.
	// document 已由 run 无损检查语法；UseNumber 保留数值 Schema 约束。
	var contract map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if err := decoder.Decode(&contract); err != nil {
		return nil, err
	}
	g := wireGenerator{declarations: map[string]string{}, schemas: map[string]map[string]any{}, rules: map[string]string{}}
	metadata, err := wireObject(contract, "generator")
	if err != nil {
		return nil, err
	}
	if metadata["schema_draft"] != "2020-12" {
		return nil, fmt.Errorf("unsupported contract schema draft")
	}
	inputs, err := wireObject(contract, "request")
	if err != nil {
		return nil, err
	}
	inputDefs, err := wireObject(inputs, "$defs")
	if err != nil {
		return nil, err
	}
	if err := verifyWireSchema(inputs, inputDefs, true); err != nil {
		return nil, err
	}
	for _, entry := range []struct{ definition, inventory string }{{"Command", "commands"}, {"RuntimeCommand", "runtime_commands"}} {
		seen := map[string]bool{}
		definition, err := wireObject(inputDefs, entry.definition)
		if err != nil {
			return nil, err
		}
		variants, ok := definition["oneOf"].([]any)
		if !ok {
			return nil, fmt.Errorf("request enum must declare oneOf")
		}
		for _, raw := range variants {
			properties, err := wireObject(raw.(map[string]any), "properties")
			if err != nil {
				return nil, err
			}
			discriminator, err := wireObject(properties, "type")
			if err != nil {
				return nil, err
			}
			name, ok := discriminator["const"].(string)
			if !ok {
				return nil, fmt.Errorf("request enum requires constant type")
			}
			if seen[name] {
				return nil, fmt.Errorf("duplicate request route")
			}
			seen[name] = true
		}
		commands, ok := contract[entry.inventory].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid request inventory")
		}
		if len(seen) != len(commands) {
			return nil, fmt.Errorf("request route coverage differs")
		}
		for _, name := range commands {
			text, ok := name.(string)
			if !ok || !seen[text] {
				return nil, fmt.Errorf("request route missing")
			}
		}
	}
	errorRoot, err := wireObject(contract, "error_response")
	if err != nil {
		return nil, err
	}
	outputRoots := map[string]map[string]any{"EmbeddedOutputErrorResponse": errorRoot}
	for _, entry := range []struct{ key, prefix string }{{"root_responses", "EmbeddedOutputRoot"}, {"runtime_responses", "EmbeddedOutputRuntime"}} {
		responses, err := wireObject(contract, entry.key)
		if err != nil {
			return nil, err
		}
		for key := range responses {
			suffix, err := wireIdentifier(key)
			if err != nil {
				return nil, err
			}
			response, err := wireObject(responses, key)
			if err != nil {
				return nil, err
			}
			name := entry.prefix + suffix + "Response"
			if _, exists := outputRoots[name]; exists {
				return nil, fmt.Errorf("response name collision %s", name)
			}
			outputRoots[name] = response
		}
	}
	outputDefs := map[string]any{}
	for _, root := range outputRoots {
		definitions, _ := root["$defs"].(map[string]any)
		if _, present := root["$defs"]; present && definitions == nil {
			return nil, fmt.Errorf("invalid output definition namespace")
		}
		if err := verifyWireSchema(root, definitions, true); err != nil {
			return nil, err
		}
		for key, raw := range definitions {
			if prior, exists := outputDefs[key]; exists && !reflect.DeepEqual(prior, raw) {
				return nil, fmt.Errorf("conflicting output definition %s", key)
			}
			outputDefs[key] = raw
		}
	}
	for _, space := range []struct {
		prefix      string
		definitions map[string]any
		roots       map[string]map[string]any
	}{{"EmbeddedInput", inputDefs, map[string]map[string]any{"EmbeddedInputRequest": inputs}}, {"EmbeddedOutput", outputDefs, outputRoots}} {
		g.names = map[string]string{}
		for _, key := range wireKeys(space.definitions) {
			suffix, err := wireIdentifier(key)
			if err != nil {
				return nil, err
			}
			g.names[key] = space.prefix + suffix
		}
		for _, key := range wireKeys(space.definitions) {
			if err := g.declare(g.names[key], space.definitions[key].(map[string]any)); err != nil {
				return nil, err
			}
		}
		for _, name := range wireKeys(space.roots) {
			if err := g.declare(name, space.roots[name]); err != nil {
				return nil, err
			}
		}
	}
	var output strings.Builder
	output.WriteString("// Code generated from the packaged embedded contract; DO NOT EDIT.\n// 从包内嵌入式契约生成；请勿手工编辑。\npackage luaskills\nimport \"reflect\"\n")
	for _, name := range wireKeys(g.declarations) {
		output.WriteString(g.declarations[name])
		output.WriteString("\n")
	}
	output.WriteString("// embeddedWireShapes is immutable metadata for generated data projection.\n// embeddedWireShapes 是生成数据投影的不可变元数据。\nvar embeddedWireShapes = map[reflect.Type]embeddedWireShape{\n")
	for _, name := range wireKeys(g.rules) {
		fmt.Fprintf(&output, "embeddedWireType[%s](): %s,\n", name, g.rules[name])
	}
	output.WriteString("}\n")
	for _, name := range wireKeys(outputRoots) {
		fmt.Fprintf(&output, "// Decode%s validates bytes and preserves required fields, nulls and exact numeric values.\n// Decode%s 校验 bytes，并保留必需字段、空值及精确数值。\n// It returns a typed envelope or a structural/protocol error; application failures remain in the error envelope.\n// 返回类型化信封或结构／协议错误；应用失败保留在错误信封中。\nfunc Decode%s(bytes []byte) (%s,error) { return decodeEmbeddedWireEnvelope[%s](bytes) }\n", name, name, name, name, name)
	}
	// Parser declaration checks catch collisions between types, enum constants and decoder functions before writing.
	// 解析器声明检查在写入前捕获类型、枚举常量及解码函数之间的冲突。
	if _, err := parser.ParseFile(token.NewFileSet(), "embedded_wire_generated.go", output.String(), parser.DeclarationErrors); err != nil {
		return nil, err
	}
	return format.Source([]byte(output.String()))
}
