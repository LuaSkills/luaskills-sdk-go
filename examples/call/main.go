package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	luaskills "github.com/LuaSkills/luaskills-sdk-go"
)

// exampleRuntimeRoot resolves the fixture runtime root used by this example.
// exampleRuntimeRoot 解析当前示例使用的夹具 runtime root。
func exampleRuntimeRoot() string {
	if value := os.Getenv("LUASKILLS_EXAMPLE_RUNTIME_ROOT"); value != "" {
		return value
	}
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("failed to resolve current example path")
	}
	return filepath.Join(filepath.Dir(filepath.Dir(currentFile)), "fixture-runtime")
}

// main calls the fixture skill and executes one inline Lua snippet.
// main 调用夹具 skill 并执行一个内联 Lua 片段。
func main() {
	runtimeRoot := exampleRuntimeRoot()
	skillRoots := luaskills.StandardRoots(runtimeRoot)
	toolName := "demo-standard-ffi-skill-ping"
	invocationContext := &luaskills.InvocationContext{
		RequestContext: map[string]any{"transport_name": "go-sdk-example"},
		ClientBudget:   map[string]any{"budget": 1},
		ToolConfig:     map[string]any{"mode": "call-demo"},
	}

	client, err := luaskills.NewClient(luaskills.ClientOptions{
		RuntimeRoot:         runtimeRoot,
		EnsureRuntimeLayout: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if _, err := client.LoadFromRoots(skillRoots); err != nil {
		log.Fatal(err)
	}

	callResult, err := client.CallSkill(toolName, map[string]any{"note": "go-call"}, invocationContext)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Call content:", callResult.Content)

	luaResult, err := client.RunLua(
		"return { note = args.note, transport = vulcan.context.request.transport_name, budget = vulcan.context.client_budget.budget, mode = vulcan.context.tool_config.mode }",
		map[string]any{"note": "go-lua"},
		invocationContext,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Run Lua result:", luaResult)
}
