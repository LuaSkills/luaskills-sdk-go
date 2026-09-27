package luaskills_test

import (
	"testing"

	luaskills "github.com/LuaSkills/luaskills-sdk-go"
)

// TestEmbeddedWirePublicConsumer compiles and consumes the generated sealed union through the public module boundary.
// TestEmbeddedWirePublicConsumer 通过公开模块边界编译并消费生成的封闭联合。
func TestEmbeddedWirePublicConsumer(t *testing.T) {
	// command uses an exported variant while its union marker stays private to the SDK.
	// command 使用导出分支，而联合标记仍为 SDK 私有。
	var command luaskills.EmbeddedInputCommand = luaskills.EmbeddedInputCommandDescribe{Type: luaskills.EmbeddedInputCommandDescribeTypeDescribe}
	bytes, err := luaskills.EncodeEmbeddedRequest(luaskills.EmbeddedInputRequest{ProtocolVersion: luaskills.EmbeddedProtocolVersion, Command: command}, 1024)
	if err != nil || len(bytes) == 0 {
		t.Fatalf("public request did not encode: %v", err)
	}
	response, err := luaskills.DecodeEmbeddedOutputRootRuntimeReserveResponse([]byte(`{"protocol_version":1,"status":"ok","result":{"runtime_id":"runtime"}}`))
	if err != nil || response.Result.RuntimeId != "runtime" {
		t.Fatalf("public response did not decode: %#v %v", response, err)
	}
}
