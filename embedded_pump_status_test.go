package luaskills

import (
	"errors"
	"testing"
)

// TestEmbeddedPumpRecoveryPublication distinguishes a returned response awaiting application from a stalled mutation.
// TestEmbeddedPumpRecoveryPublication 区分等待应用的已返回响应与停滞变更。
// This deterministic boundary represents the public observation window between native return and coordinator application.
// 此确定性边界表示原生返回与协调器应用之间的公开观察窗口。
func TestEmbeddedPumpRecoveryPublication(t *testing.T) {
	// mutation has delivered normally; even an older retained failure cannot make this new delivery require recovery.
	// mutation 已正常交付；即使存在历史保留错误，也不能使此新交付需要恢复。
	mutation := &embeddedPumpMutation{kind: "take", returned: true, response: []byte(`{"protocol_version":1,"status":"ok","result":[]}`)}
	pump := &EmbeddedCallbackPump{pending: mutation, failure: errors.New("historical resolved failure")}
	if pump.Status().RecoveryRequired {
		t.Fatal("ordinary returned response was reported as requiring recovery")
	}
	// err is published only when the coordinator cannot apply or safely settle the original mutation.
	// err 仅在协调器无法应用或安全收尾原始变更时发布。
	mutation.err = errors.New("original delivery cannot be applied")
	if !pump.Status().RecoveryRequired {
		t.Fatal("stalled mutation lost its recovery signal")
	}
	mutation.err = nil
	pump.needsRelease = true
	if !pump.Status().RecoveryRequired {
		t.Fatal("retained native allocation lost its recovery signal")
	}
}
