package luaskills

import "fmt"

// checkAdditionalFramesLocked verifies aggregate SDK frame reservations before publishing another owner; caller holds t.mu.
// checkAdditionalFramesLocked 在发布其他所有者前验证 SDK 累计帧预留；调用方持有 t.mu。
// Division and subtraction precede admission so neither multiplication nor addition can overflow.
// 入场前先做除法及减法，使乘法及加法均不会溢出。
func (t *EmbeddedTransport) checkAdditionalFramesLocked(additional uint64) error {
	limit := t.config.MaxResultBuffers
	if bytesLimit := t.config.MaxResultBytes / t.config.MaxResponseBytes; bytesLimit < limit {
		limit = bytesLimit
	}
	claimed := uint64(len(t.callbackPumps)) * embeddedControlWorkers
	if t.commandDriver != nil {
		ordinary := t.commandDriver.config.WorkWorkers + embeddedControlWorkers
		if ordinary > limit || claimed > limit-ordinary {
			return &EmbeddedRuntimeError{"capacity_exceeded", "existing SDK frame claims exceed transport capacity"}
		}
		claimed += ordinary
	}
	if claimed > limit || additional > limit-claimed {
		return &EmbeddedRuntimeError{"capacity_exceeded", "transport cannot reserve worst-case response frames for all SDK workers"}
	}
	return nil
}

// claimCallbackPump publishes one exact owner before its coordinator starts and rejects ambiguous or unbudgeted ownership.
// claimCallbackPump 在协调器启动前发布一个精确所有者，并拒绝含糊或超预算的所有权。
func (t *EmbeddedTransport) claimCallbackPump(pump *EmbeddedCallbackPump) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.identity == 0 {
		return &EmbeddedTransportError{"transport", EmbeddedNativeClosed}
	}
	if t.active != 0 || t.exclusive || len(t.results) != 0 || t.callbackPumps[pump.runtimeID] != nil {
		return &EmbeddedRuntimeError{"busy", "callback pump requires a quiescent transport and no existing owner for its runtime"}
	}
	if uint64(len(t.callbackPumps)) >= t.config.MaxRuntimes {
		return &EmbeddedRuntimeError{"capacity_exceeded", "callback pump count exceeds transport runtime capacity"}
	}
	if err := t.checkAdditionalFramesLocked(embeddedControlWorkers); err != nil {
		return err
	}
	if t.callbackPumps == nil {
		t.callbackPumps = make(map[string]*EmbeddedCallbackPump)
	}
	t.callbackPumps[pump.runtimeID] = pump
	return nil
}

// releaseCallbackPump removes only the exact owner after all callbacks, acknowledgements and coordinator work drain.
// releaseCallbackPump 仅在全部回调、确认及协调工作排空后移除精确所有者。
func (t *EmbeddedTransport) releaseCallbackPump(pump *EmbeddedCallbackPump) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.callbackPumps[pump.runtimeID] != pump {
		return fmt.Errorf("embedded callback pump ownership mismatch")
	}
	delete(t.callbackPumps, pump.runtimeID)
	return nil
}
