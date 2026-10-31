package main

import "log"

// main runs the real SQLite host demonstration and exits on failure.
// main 运行真实 SQLite 宿主演示，并在失败时退出。
func main() {
	if err := runSQLiteDemo(); err != nil {
		log.Fatal(err)
	}
}
