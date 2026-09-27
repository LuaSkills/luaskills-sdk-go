// Command generate-embedded-contract emits protocol metadata and complete wire types from the packaged authoritative contract.
// generate-embedded-contract 命令从包内权威契约生成协议元数据及完整线类型。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/LuaSkills/luaskills-sdk-go/internal/embeddedjson"
	"go/format"
	"os"
	"sort"
	"strings"
)

// document selects protocol metadata; the wire generator reads schemas separately from the same exact artifact.
// document 选择协议元数据；线生成器从同一精确产物中单独读取 Schema。
type document struct {
	// ContractVersion controls the artifact container shape.
	// ContractVersion 控制产物容器形状。
	ContractVersion int `json:"contract_version"`
	// ProtocolVersion selects the exact request/response and ABI version.
	// ProtocolVersion 选择精确请求／响应及 ABI 版本。
	ProtocolVersion uint32 `json:"protocol_version"`
	// CoreVersion records the generating core, not a published-library compatibility claim.
	// CoreVersion 记录生成核心，不声称已发布动态库兼容。
	CoreVersion string `json:"core_version"`
	// NativeStatus preserves authoritative ABI numbers.
	// NativeStatus 保留权威 ABI 数值。
	NativeStatus map[string]int32 `json:"native_status"`
	// Commands lists root dispatch names.
	// Commands 列出根分发名称。
	Commands []string `json:"commands"`
	// RuntimeCommands lists nested runtime dispatch names.
	// RuntimeCommands 列出嵌套运行时分发名称。
	RuntimeCommands []string `json:"runtime_commands"`
	// RootResponses is checked against all non-nested root commands.
	// RootResponses 与全部非嵌套根命令核对。
	RootResponses map[string]json.RawMessage `json:"root_responses"`
	// RuntimeResponses is checked against all nested runtime commands.
	// RuntimeResponses 与全部嵌套运行时命令核对。
	RuntimeResponses map[string]json.RawMessage `json:"runtime_responses"`
}

// main checks exact artifact bytes and either verifies or writes formatted Go artifacts; invalid input exits nonzero.
// main 校验精确产物字节，并验证或写入格式化 Go 产物；无效输入以非零状态退出。
func main() {
	check := flag.Bool("check", false, "Verify all generated contract artifacts without writing")
	flag.Parse()
	if err := run(*check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run derives metadata and wire-type sources from packaged files; check rejects drift without modifying any file.
// run 从包内文件派生元数据及线类型源码；check 拒绝漂移且不修改任何文件。
func run(check bool) error {
	data, err := os.ReadFile("contracts/embedded/v1/contract.json")
	if err != nil {
		return err
	}
	digest, err := os.ReadFile("contracts/embedded/v1/contract.sha256")
	if err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	if strings.TrimSpace(string(digest)) != hash+"  contract.json" {
		return fmt.Errorf("embedded contract checksum mismatch")
	}
	if _, err = embeddedjson.Decode(data); err != nil {
		return err
	}
	var contract document
	if err = json.Unmarshal(data, &contract); err != nil {
		return err
	}
	if contract.ContractVersion != 1 || contract.ProtocolVersion != 1 || contract.CoreVersion == "" {
		return fmt.Errorf("unsupported embedded contract metadata")
	}
	if err = checkNames(contract.Commands, contract.RootResponses, true); err != nil {
		return err
	}
	if err = checkNames(contract.RuntimeCommands, contract.RuntimeResponses, false); err != nil {
		return err
	}
	names := make([]string, 0, len(contract.NativeStatus))
	seen := make(map[int32]bool)
	for name, status := range contract.NativeStatus {
		if status < 0 || seen[status] {
			return fmt.Errorf("invalid or duplicate native status")
		}
		seen[status] = true
		names = append(names, name)
	}
	if status, ok := contract.NativeStatus["ok"]; !ok || status != 0 {
		return fmt.Errorf("missing native success status")
	}
	sort.Strings(names)
	var output strings.Builder
	output.WriteString("// Code generated from the packaged embedded contract; DO NOT EDIT.\n// 从包内嵌入式契约生成；请勿手工编辑。\npackage luaskills\n\n")
	fmt.Fprintf(&output, "// EmbeddedProtocolVersion is the exact root JSON and C structure protocol version.\n// EmbeddedProtocolVersion 是精确的根 JSON 及 C 结构协议版本。\nconst EmbeddedProtocolVersion uint32 = %d\n", contract.ProtocolVersion)
	fmt.Fprintf(&output, "// EmbeddedContractSHA256 identifies every byte of the packaged contract.\n// EmbeddedContractSHA256 标识包内契约的全部字节。\nconst EmbeddedContractSHA256 = %q\n", hash)
	output.WriteString("// EmbeddedNativeStatus is one exact signed C ABI status code.\n// EmbeddedNativeStatus 是一个精确的有符号 C ABI 状态码。\ntype EmbeddedNativeStatus int32\nconst (\n")
	for _, name := range names {
		identifier := "EmbeddedNative"
		for _, word := range strings.Split(name, "_") {
			if word == "" {
				return fmt.Errorf("invalid native status name")
			}
			for _, r := range word {
				if r < 'a' || r > 'z' {
					return fmt.Errorf("invalid native status name")
				}
			}
			identifier += strings.ToUpper(word[:1]) + word[1:]
		}
		fmt.Fprintf(&output, "// %s preserves the core's %s status.\n// %s 保留核心的 %s 状态。\n%s EmbeddedNativeStatus = %d\n", identifier, name, identifier, name, identifier, contract.NativeStatus[name])
	}
	output.WriteString(")\n")
	for _, entry := range []struct {
		name   string
		values []string
	}{{"EmbeddedRootCommands", contract.Commands}, {"EmbeddedRuntimeCommands", contract.RuntimeCommands}} {
		fmt.Fprintf(&output, "// %s returns an independent copy of the authoritative command names.\n// %s 返回权威命令名称的独立副本。\nfunc %s() []string { return []string{", entry.name, entry.name, entry.name)
		for _, value := range entry.values {
			fmt.Fprintf(&output, "%q,", value)
		}
		output.WriteString("} }\n")
	}
	formatted, err := format.Source([]byte(output.String()))
	if err != nil {
		return err
	}
	// Build every artifact before writing, so schema rejection cannot leave a partially regenerated contract.
	// 在写入前构建全部产物，防止 Schema 拒绝留下部分重新生成的契约。
	wire, err := generateWire(data)
	if err != nil {
		return err
	}
	artifacts := map[string][]byte{"embedded_contract_generated.go": formatted, "embedded_wire_generated.go": wire}
	for _, name := range wireKeys(artifacts) {
		if check {
			current, err := os.ReadFile(name)
			if err != nil {
				return err
			}
			if !bytes.Equal(current, artifacts[name]) {
				return fmt.Errorf("generated Go contract is stale: %s", name)
			}
		} else if err := os.WriteFile(name, artifacts[name], 0644); err != nil {
			return err
		}
	}
	return nil
}

// checkNames rejects duplicate names and metadata/response coverage drift; runtime dispatch has nested responses.
// checkNames 拒绝重复名称及元数据／响应覆盖漂移；runtime 分发使用嵌套响应。
func checkNames(names []string, responses map[string]json.RawMessage, nested bool) error {
	seen := make(map[string]bool)
	expected := 0
	for _, name := range names {
		if name == "" || seen[name] {
			return fmt.Errorf("invalid embedded command metadata")
		}
		seen[name] = true
		if nested && name == "runtime" {
			continue
		}
		if _, ok := responses[name]; !ok {
			return fmt.Errorf("missing response for %s", name)
		}
		expected++
	}
	if expected != len(responses) || len(names) == 0 || (nested && !seen["runtime"]) {
		return fmt.Errorf("embedded response coverage mismatch")
	}
	return nil
}
