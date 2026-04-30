package main

import (
	"errors"
	"fmt"
	"log"

	luaskills "github.com/LuaSkills/luaskills-sdk-go"
)

// main shows the current Go JSON callback bridge boundary.
// main 展示当前 Go JSON callback 桥接边界。
func main() {
	// Register the SQLite provider callback boundary for demonstration.
	// 注册 SQLite provider callback 边界用于演示。
	providerErr := luaskills.SetSQLiteProviderJSONCallback(func(request any) (any, error) {
		return map[string]any{"ok": true, "request": request}, nil
	})
	// Register the host-tool callback boundary for demonstration.
	// 注册宿主工具 callback 边界用于演示。
	hostToolErr := luaskills.SetHostToolJSONCallback(func(request luaskills.HostToolJSONRequest) (any, error) {
		return map[string]any{"ok": true, "request": request}, nil
	})
	if errors.Is(providerErr, luaskills.ErrProviderCallbacksRequireHostBridge) &&
		errors.Is(hostToolErr, luaskills.ErrHostToolCallbacksRequireHostBridge) {
		fmt.Println("Go JSON callbacks require a host-owned cgo callback bridge.")
		return
	}
	if providerErr != nil {
		log.Fatal(providerErr)
	}
	defer luaskills.ClearSQLiteProviderJSONCallback()
	if hostToolErr != nil {
		log.Fatal(hostToolErr)
	}
	defer luaskills.ClearHostToolJSONCallback()
}
