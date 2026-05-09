package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	luaskills "github.com/LuaSkills/luaskills-sdk-go"
)

// runtimeSessionSID is the stable session id reused by the example lease lifecycle.
// runtimeSessionSID 是示例租约生命周期复用的稳定会话标识。
const runtimeSessionSID = "go-sdk-runtime-lease-demo"

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

// main runs one persistent runtime-lease smoke flow through the high-level Go SDK surface.
// main 通过高级 Go SDK 接口执行一条持久运行时租约烟测链路。
func main() {
	runtimeRoot := exampleRuntimeRoot()
	skillRoots := luaskills.StandardRoots(runtimeRoot)

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

	system := client.System(luaskills.AuthoritySystem)
	entries, err := system.ListEntries()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Visible entry count:", len(entries))

	skillID, err := system.SkillNameForTool("demo-standard-ffi-skill-ping")
	if err != nil {
		log.Fatal(err)
	}
	if skillID != nil {
		fmt.Println("Visible skill ownership:", *skillID)
	}

	sessions := system.RuntimeLeases()
	usesSystemEndpoints, err := sessions.UsesSystemRuntimeLeaseEndpoints()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Uses dedicated system runtime-lease endpoints:", usesSystemEndpoints)

	systemLuaLib := filepath.Join(runtimeRoot, "system_lua_lib")
	ttlSec := 600
	session, err := sessions.CreateHandleWithOptions(runtimeSessionSID, true, &luaskills.RuntimeLeaseCreateOptions{
		TTLSec: &ttlSec,
		CWD:    &systemLuaLib,
		Mounts: map[string]any{"example": "go-runtime-lease"},
	})
	if err != nil {
		log.Fatal(err)
	}
	identity := session.IdentityPayload()
	fmt.Println("Lease created:", identity.LeaseID)

	handles, err := sessions.ListHandles(runtimeSessionSID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Lease handle count:", len(handles))

	opened, err := session.Eval(`
local info = vulcan.os.info()
if not proc then
  local spec
  if info.os == "windows" then
    spec = {
      program = "cmd",
      args = { "/V:ON", "/C", "set /P line=&echo session:!line!" },
      encoding = "utf-8",
    }
  else
    spec = {
      program = "sh",
      args = { "-c", "read line; echo session:$line" },
      encoding = "utf-8",
    }
  end
  proc = vulcan.process.session.open(spec)
end
counter = (counter or 0) + 1
proc:write((args.input or "runtime-lease-demo") .. "\n")
return {
  opened = true,
  counter = counter,
  input = args.input,
}
`, map[string]any{"input": "runtime-lease-demo"}, 60000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Open eval result:", opened["result"])

	readOutput, err := session.Eval(`
counter = (counter or 0) + 1
local output = proc:read({ timeout_ms = 2000, max_bytes = 8192 })
return {
  counter = counter,
  stdout = output.stdout,
  stderr = output.stderr,
  timed_out = output.timed_out,
}
`, nil, 60000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Read eval result:", readOutput["result"])

	status, err := session.Status()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Lease status result:", status)

	closedProcess, err := session.Eval(`
counter = (counter or 0) + 1
local status = proc:close({ timeout_ms = 3000 })
proc = nil
return {
  counter = counter,
  exited = status.exited,
  success = status.success,
}
`, nil, 60000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Close process eval result:", closedProcess["result"])

	closedLease, err := session.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Lease close result:", closedLease)

	postClose, err := sessions.CallRaw("eval", map[string]any{
		"lease_id":   identity.LeaseID,
		"sid":        identity.SID,
		"generation": identity.Generation,
		"timeout_ms": 60000,
		"args":       map[string]any{},
		"code":       "return 1",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Post-close eval result:", postClose)
}
