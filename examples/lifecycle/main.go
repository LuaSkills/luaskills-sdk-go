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

// main demonstrates disable and enable through the ordinary Skills plane.
// main 演示通过普通 Skills plane 执行 disable 与 enable。
func main() {
	runtimeRoot := exampleRuntimeRoot()
	skillRoots := luaskills.StandardRoots(runtimeRoot)
	toolName := "demo-standard-ffi-skill-ping"
	skillID := "demo-standard-ffi-skill"

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

	before, err := client.CallSkill(toolName, map[string]any{"note": "before-disable"}, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Call before disable:", before.Content)

	if _, err := client.Skills.Disable(skillRoots, skillID, "example maintenance window"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Skill disabled:", skillID)

	visible, err := client.IsSkill(toolName, luaskills.AuthorityDelegatedTool)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Visible after disable:", visible)

	if _, err := client.CallSkill(toolName, map[string]any{"note": "after-disable"}, nil); err == nil {
		log.Fatal("callSkill unexpectedly succeeded while the skill was disabled")
	} else {
		fmt.Println("Call after disable failed as expected:", err)
	}

	if _, err := client.Skills.Enable(skillRoots, skillID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Skill enabled:", skillID)

	after, err := client.CallSkill(toolName, map[string]any{"note": "after-enable"}, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Call after enable:", after.Content)
}
