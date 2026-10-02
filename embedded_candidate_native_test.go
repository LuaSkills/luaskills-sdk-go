//go:build cgo

package luaskills

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"testing"
)

// TestEmbeddedCandidateEvidence compares the linked native description and staged bytes with explicit frozen inputs.
// TestEmbeddedCandidateEvidence 将已链接原生描述及暂存字节与显式冻结输入比较。
// t owns the transport; ordinary native runs without candidate inputs retain their existing opt-in boundary.
// t 拥有传输；未提供候选输入的普通原生执行保留现有显式启用边界。
func TestEmbeddedCandidateEvidence(t *testing.T) {
	// descriptionPath is supplied only by the local candidate gate, never inferred from release assets.
	// descriptionPath 仅由本地候选门禁提供，绝不根据发布资产推断。
	descriptionPath := os.Getenv("LUASKILLS_CANDIDATE_DESCRIPTION")
	if descriptionPath == "" {
		t.Skip("explicit candidate description is not configured")
	}
	if os.Getenv("LUASKILLS_NATIVE_E2E") != "1" {
		t.Fatal("candidate evidence requires LUASKILLS_NATIVE_E2E=1")
	}
	// bytes come from the frozen description file and pass the existing generated strict decoder.
	// bytes 来自冻结描述文件，并通过现有生成严格解码器。
	bytes, err := os.ReadFile(descriptionPath)
	if err != nil {
		t.Fatal(err)
	}
	// expected retains every build field; a matching contract hash alone is insufficient candidate evidence.
	// expected 保留每个构建字段；仅匹配契约摘要不足以构成候选证据。
	expected, err := validateEmbeddedDescription(bytes)
	if err != nil {
		t.Fatal(err)
	}
	// libraryBytes identify the exact immutable staged library used by the local link/load configuration.
	// libraryBytes 标识本地链接／加载配置使用的精确不可变暂存库。
	libraryBytes, err := os.ReadFile(os.Getenv("LUASKILLS_CANDIDATE_LIBRARY"))
	if err != nil {
		t.Fatal(err)
	}
	// hash preserves the binary digest separately from selected build-input hashes in CoreDescription.
	// hash 将二进制摘要与 CoreDescription 中选定构建输入摘要分别保存。
	hash := sha256.Sum256(libraryBytes)
	if hex.EncodeToString(hash[:]) != os.Getenv("LUASKILLS_CANDIDATE_SHA256") {
		t.Fatal("candidate library SHA-256 mismatch")
	}
	// transport discovers metadata through the real cgo adapter before native allocation.
	// transport 在原生分配前通过真实 cgo 适配器发现元数据。
	transport := nativeEmbeddedTest(t)
	if !reflect.DeepEqual(transport.CoreDescription(), expected) {
		t.Fatalf("linked candidate description differs from frozen evidence: actual=%#v expected=%#v", transport.CoreDescription(), expected)
	}
	if err := compareEmbeddedCandidateDescription(transport, bytes); err != nil {
		t.Fatal(err)
	}
}

// compareEmbeddedCandidateDescription compares every raw JSON field against the same retained native adapter.
// compareEmbeddedCandidateDescription 将每个原始 JSON 字段与同一个保留原生适配器比较。
// transport supplies the already linked library; frozen contains strict original JSON; return mismatch/parse/read errors.
// transport 提供已链接库；frozen 包含严格原始 JSON；返回不匹配／解析／读取错误。
func compareEmbeddedCandidateDescription(transport *EmbeddedTransport, frozen []byte) error {
	// expected preserves open-schema extension fields that typed projection intentionally does not expose.
	// expected 保留类型投影有意不暴露的开放 Schema 扩展字段。
	expected, err := DecodeEmbeddedJSON(frozen)
	if err != nil {
		return err
	}
	// raw comes from the exact bootstrap adapter retained by this transport, without loading another library.
	// raw 来自此传输保留的精确引导适配器，不加载其他库。
	raw, err := transport.native.describe()
	if err != nil {
		return err
	}
	// actual retains the entire original object and exact integer representation.
	// actual 保留整个原始对象及精确整数表示。
	actual, err := DecodeEmbeddedJSON(raw)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf("raw linked candidate description differs from frozen evidence: actual=%#v expected=%#v", actual, expected)
	}
	return nil
}

// TestEmbeddedCandidateDescriptionExtensions rejects mismatched root/build extension fields in frozen native evidence.
// TestEmbeddedCandidateDescriptionExtensions 拒绝冻结原生证据中不匹配的根／构建扩展字段。
// t owns the real transport; ordinary SDK open-schema compatibility remains unchanged.
// t 拥有真实传输；普通 SDK 开放 Schema 兼容性保持不变。
func TestEmbeddedCandidateDescriptionExtensions(t *testing.T) {
	// transport and original share the same linked candidate and library-owned description source.
	// transport 和 original 共享同一个已链接候选及库拥有的描述来源。
	transport := nativeEmbeddedTest(t)
	original, err := transport.native.describe()
	if err != nil {
		t.Fatal(err)
	}
	for _, location := range []string{"root", "build"} {
		t.Run(location, func(t *testing.T) {
			// modified is independently decoded for each exact documented extension location.
			// modified 对每个精确文档化扩展位置独立解码。
			modified, err := DecodeEmbeddedJSON(original)
			if err != nil {
				t.Fatal(err)
			}
			// description is the root JSON object; build is the generated contract's exact build-identity field.
			// description 是根 JSON 对象；build 是生成契约的精确构建身份字段。
			description := modified.(map[string]any)
			if location == "root" {
				description["candidate_extension"] = "different"
			} else {
				description["build"].(map[string]any)["candidate_extension"] = "different"
			}
			// frozen tests candidate-only full-object equality while remaining accepted by ordinary compatibility.
			// frozen 在普通兼容性仍接纳时验证仅候选所用的完整对象相等。
			frozen, err := EncodeEmbeddedJSON(description, EmbeddedDescriptionMaxBytes)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := validateEmbeddedDescription(frozen); err != nil {
				t.Fatal("ordinary compatibility unexpectedly rejected an open-schema extension", err)
			}
			if err := compareEmbeddedCandidateDescription(transport, frozen); err == nil {
				t.Fatal("frozen raw description extension mismatch was ignored")
			}
		})
	}
}
