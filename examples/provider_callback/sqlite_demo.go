package main

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	luaskills "github.com/LuaSkills/luaskills-sdk-go"
	_ "github.com/mattn/go-sqlite3"
)

// fixtureJSON contains the exact Lua/manifest contract shared by the three demonstrations.
// fixtureJSON 包含三个示例共享的精确 Lua 与清单契约。
//
//go:embed sqlite-demo-fixture.json
var fixtureJSON []byte

// sqliteRequest is the core action/binding/input envelope plus the demo's explicit positional SQL protocol.
// sqliteRequest 是核心 action/binding/input 包络及示例显式的位置参数 SQL 协议。
type sqliteRequest struct {
	// Action selects one of the three implemented operations.
	// Action 选择三个已实现操作之一。
	Action string `json:"action"`
	// Binding is produced by the core; skill input never overrides its database path.
	// Binding 由核心产生；技能输入不能覆盖其数据库路径。
	Binding struct {
		DefaultDatabasePath string `json:"default_database_path"`
	} `json:"binding"`
	// Input carries the host-defined SQL and statement parameters.
	// Input 携带宿主定义的 SQL 及语句参数。
	Input struct {
		SQL        string          `json:"sql"`
		Params     json.RawMessage `json:"params"`
		Statements []struct {
			SQL    string `json:"sql"`
			Params []any  `json:"params"`
		} `json:"statements"`
	} `json:"input"`
}

// sqliteProvider executes the core request with host-owned SQLite and returns rows or an error.
// sqliteProvider 使用宿主 SQLite 执行核心请求，返回数据行或错误。
// Connections close per request; batch failures roll back all preceding statements.
// 每次请求关闭连接；批次失败将回滚此前所有语句。
func sqliteProvider(request any) (any, error) {
	// Decode the confirmed JSON protocol, rejecting unsupported parameter representations.
	// 解码已确认的 JSON 协议，拒绝不支持的参数表示。
	encoded, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	var command sqliteRequest
	if err = json.Unmarshal(encoded, &command); err != nil {
		return nil, err
	}
	if command.Binding.DefaultDatabasePath == "" {
		return nil, fmt.Errorf("missing core database binding")
	}
	if err = os.MkdirAll(filepath.Dir(command.Binding.DefaultDatabasePath), 0755); err != nil {
		return nil, err
	}
	// The driver belongs to this example module, never the SDK dependency graph.
	// 驱动属于此示例模块，绝不进入 SDK 依赖图。
	database, err := sql.Open("sqlite3", command.Binding.DefaultDatabasePath)
	if err != nil {
		return nil, err
	}
	defer database.Close()
	switch command.Action {
	case "execute_script":
		_, err = database.Exec(command.Input.SQL)
		return map[string]any{"ok": err == nil}, err
	case "execute_batch":
		// Defer rollback covers any error before commit, including malformed statements.
		// 延后回滚覆盖提交前的任何错误，包括畸形语句。
		transaction, err := database.Begin()
		if err != nil {
			return nil, err
		}
		defer transaction.Rollback()
		for _, statement := range command.Input.Statements {
			if _, err = transaction.Exec(statement.SQL, statement.Params...); err != nil {
				return nil, err
			}
		}
		err = transaction.Commit()
		return map[string]any{"ok": err == nil}, err
	case "query_json":
		// Lua's empty table is encoded as {}; this explicit protocol maps it to zero arguments.
		// Lua 空表编码为 {}；此显式协议将其映射为零个参数。
		var parameters []any
		if string(command.Input.Params) != "{}" {
			if err = json.Unmarshal(command.Input.Params, &parameters); err != nil {
				return nil, err
			}
		}
		rows, err := database.Query(command.Input.SQL, parameters...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		// Column names are supplied by SQLite; values retain JSON null and numeric semantics.
		// 列名由 SQLite 提供；值保留 JSON 空值及数字语义。
		result := make([]map[string]any, 0)
		for rows.Next() {
			values := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for index := range values {
				pointers[index] = &values[index]
			}
			if err = rows.Scan(pointers...); err != nil {
				return nil, err
			}
			row := make(map[string]any, len(columns))
			for index, column := range columns {
				row[column] = values[index]
			}
			result = append(result, row)
		}
		return map[string]any{"rows": result, "database_path": command.Binding.DefaultDatabasePath}, rows.Err()
	default:
		return nil, fmt.Errorf("unsupported SQLite demo action: %s", command.Action)
	}
}

// runSQLiteDemo creates fresh skill data and verifies two engines, returning the first failure.
// runSQLiteDemo 创建全新技能数据并验证两个引擎，返回首个失败。
// The database is retained for inspection; no existing user files are overwritten.
// 数据库保留供检查；不覆盖任何既有用户文件。
func runSQLiteDemo() error {
	// Embedded fixture bytes keep the example runnable from an installed source archive.
	// 内嵌夹具字节确保示例可从已安装源码包运行。
	var fixture struct {
		Manifest string `json:"manifest"`
		Lua      string `json:"lua"`
	}
	if err := json.Unmarshal(fixtureJSON, &fixture); err != nil {
		return err
	}
	runtimeRoot, err := os.MkdirTemp("", "luaskills-sqlite-go-")
	if err != nil {
		return err
	}
	skill := filepath.Join(runtimeRoot, "user_skills", "sqlite-demo")
	if err = os.MkdirAll(filepath.Join(skill, "runtime"), 0755); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(skill, "skill.yaml"), []byte(fixture.Manifest), 0644); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(skill, "runtime", "run.lua"), []byte(fixture.Lua), 0644); err != nil {
		return err
	}
	defer luaskills.ClearSQLiteProviderJSONCallback()
	for _, write := range []bool{true, false} {
		if err = runSQLiteEngine(runtimeRoot, write); err != nil {
			return err
		}
	}
	return nil
}

// runSQLiteEngine captures one provider, clears the default, and verifies write/read behavior; return errors.
// runSQLiteEngine 捕获一个提供方、清空默认值并验证读写行为；返回错误。
func runSQLiteEngine(runtimeRoot string, write bool) error {
	if err := luaskills.SetSQLiteProviderJSONCallback(sqliteProvider); err != nil {
		return err
	}
	// Engine creation captures the callback generation before the process default is cleared.
	// 引擎创建在进程默认值清空前捕获回调代次。
	client, err := luaskills.NewClient(luaskills.ClientOptions{RuntimeRoot: runtimeRoot, EnsureRuntimeLayout: true, HostOptions: map[string]any{"host_mode": "desktop", "sqlite_callback_mode": "json"}})
	if err != nil {
		return err
	}
	defer client.Close()
	if err = luaskills.ClearSQLiteProviderJSONCallback(); err != nil {
		return err
	}
	if _, err = client.LoadFromRoots(luaskills.StandardRoots(runtimeRoot)); err != nil {
		return err
	}
	result, err := client.CallSkill("sqlite-demo-run", map[string]any{"write": write, "text": "宿主 SQLite: '); DROP TABLE notes; --"}, nil)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(result.Content, "sqlite-demo:ok:") {
		return fmt.Errorf("unexpected demo result: %#v", result)
	}
	fmt.Println(result.Content)
	return nil
}
