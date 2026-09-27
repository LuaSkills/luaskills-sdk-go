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
	// Each independent coordinator reserves its own maximum response before any owner becomes visible.
	// 每个独立协调器在任何所有者可见前预留自身最大响应。
	claimed := uint64(0)
	for _, owners := range []uint64{uint64(len(t.callbackPumps)), uint64(len(t.runtimeScopes))} {
		if owners > (limit-claimed)/embeddedControlWorkers {
			return &EmbeddedRuntimeError{"capacity_exceeded", "existing SDK coordinator claims exceed transport capacity"}
		}
		claimed += owners * embeddedControlWorkers
	}
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
	if t.runtimeScopes[pump.runtimeID] != nil {
		return &EmbeddedRuntimeError{"busy", "create the callback pump before adopting its runtime scope"}
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

// claimRuntimeScope adopts the exact pump and reserves one control frame at a quiescent driver/transport boundary.
// claimRuntimeScope 在驱动器／传输静止边界接管精确泵，并预留一个控制帧。
// Driver then transport is the existing closure lock order; holding both makes adoption atomic with managed free admission.
// 先驱动器后传输是既有关闭锁序；同时持有两锁使接管与受管释放入场互斥。
func (t *EmbeddedTransport) claimRuntimeScope(scope *EmbeddedRuntimeScope) error {
	// driver is the uniquely bound typed runtime's borrowed driver, never inferred from another handle.
	// driver 是类型化运行时唯一绑定的借用驱动器，绝不根据其他句柄推断。
	driver := scope.runtime.client.driver
	driver.mu.Lock()
	defer driver.mu.Unlock()
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.identity == 0 {
		return &EmbeddedTransportError{"transport", EmbeddedNativeClosed}
	}
	if t.commandDriver != nil && t.commandDriver != driver {
		return &EmbeddedRuntimeError{"busy", "scope handle must use the current command driver"}
	}
	if t.active != 0 || t.exclusive || len(t.results) != 0 || driver.maintenance || t.runtimeScopes[scope.runtime.identity] != nil {
		return &EmbeddedRuntimeError{"busy", "runtime scope requires a quiescent transport and no existing scope"}
	}
	for _, command := range driver.commands {
		if command.state != EmbeddedCommandCompleted {
			return &EmbeddedRuntimeError{"busy", "settle existing driver commands before adopting the runtime scope"}
		}
	}
	if t.callbackPumps[scope.runtime.identity] != scope.pump {
		return &EmbeddedRuntimeError{"busy", "runtime scope must adopt its exact existing callback pump"}
	}
	if uint64(len(t.runtimeScopes)) >= t.config.MaxRuntimes {
		return &EmbeddedRuntimeError{"capacity_exceeded", "runtime scope count exceeds transport runtime capacity"}
	}
	if err := t.checkAdditionalFramesLocked(embeddedControlWorkers); err != nil {
		return err
	}
	if t.runtimeScopes == nil {
		t.runtimeScopes = make(map[string]*EmbeddedRuntimeScope)
	}
	t.runtimeScopes[scope.runtime.identity] = scope
	return nil
}

// releaseRuntimeScope returns only the exact owner's claim after proven native slot removal and final coordinator work.
// releaseRuntimeScope 仅在原生槽移除得到证明且协调器工作收尾后归还精确所有者声明。
func (t *EmbeddedTransport) releaseRuntimeScope(scope *EmbeddedRuntimeScope) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.runtimeScopes[scope.runtime.identity] != scope {
		return fmt.Errorf("embedded runtime scope ownership mismatch")
	}
	delete(t.runtimeScopes, scope.runtime.identity)
	return nil
}

// checkUnmanagedRuntime rejects driver removal of a scoped runtime; the caller holds its driver admission lock.
// checkUnmanagedRuntime 拒绝驱动器移除由作用域拥有的运行时；调用方持有其驱动器入场锁。
func (t *EmbeddedTransport) checkUnmanagedRuntime(runtimeID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.runtimeScopes[runtimeID] != nil {
		return &EmbeddedRuntimeError{"busy", "runtime release belongs to its lifecycle scope"}
	}
	return nil
}
