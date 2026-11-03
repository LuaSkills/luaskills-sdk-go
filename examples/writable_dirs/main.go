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

// main keeps temp files and downloads outside runtime_root so the runtime root can stay read-only.
// main 将临时文件与下载内容放在 runtime_root 之外，使运行时根可以保持只读。
func main() {
	runtimeRoot := exampleRuntimeRoot()
	writableRoot, err := os.MkdirTemp("", "luaskills-writable-dirs-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(writableRoot)
	// The host creates temp_dir; the engine creates download_cache_root when it first needs it.
	// 宿主创建 temp_dir；引擎在首次需要时创建 download_cache_root。
	tempDir := filepath.Join(writableRoot, "temp")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		log.Fatal(err)
	}

	// EnsureRuntimeLayout stays off because it would create directories below runtime_root.
	// EnsureRuntimeLayout 保持关闭，因为它会在 runtime_root 下创建目录。
	client, err := luaskills.NewClient(luaskills.ClientOptions{
		RuntimeRoot: runtimeRoot,
		HostOptions: map[string]any{
			"preserve_host_writable_dirs": true,
			"temp_dir":                    tempDir,
			"download_cache_root":         filepath.Join(writableRoot, "downloads"),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if _, err := client.LoadFromRoots(luaskills.StandardRoots(runtimeRoot)); err != nil {
		log.Fatal(err)
	}
	result, err := client.RunLua("return vulcan.runtime.temp_dir", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Runtime root:", runtimeRoot)
	fmt.Println("Skill temp_dir:", result)
}
