package main

import "testing"

// TestSQLiteDemo executes the real SQLite driver through two native engines and returns failures to Go tests.
// TestSQLiteDemo 通过两个原生引擎执行真实 SQLite 驱动，并向 Go 测试报告失败。
func TestSQLiteDemo(t *testing.T) {
	if err := runSQLiteDemo(); err != nil {
		t.Fatal(err)
	}
}
