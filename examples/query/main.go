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

// main loads the fixture skill and prints common delegated-query results.
// main 加载夹具 skill 并输出常见委托查询结果。
func main() {
	runtimeRoot := exampleRuntimeRoot()
	skillRoots := luaskills.StandardRoots(runtimeRoot)
	toolName := "demo-standard-ffi-skill-ping"

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

	entries, err := client.ListEntries(luaskills.AuthorityDelegatedTool)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Entry count:", len(entries))

	visible, err := client.IsSkill(toolName, luaskills.AuthorityDelegatedTool)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Is delegated-visible skill:", visible)

	skillID, err := client.SkillNameForTool(toolName, luaskills.AuthorityDelegatedTool)
	if err != nil {
		log.Fatal(err)
	}
	if skillID != nil {
		fmt.Println("Owning skill id:", *skillID)
	}

	completions, err := client.PromptArgumentCompletions(toolName, "note", luaskills.AuthorityDelegatedTool)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Prompt completion count:", len(completions))

	help, err := client.ListSkillHelp(luaskills.AuthorityDelegatedTool)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Help tree count:", len(help))
}
