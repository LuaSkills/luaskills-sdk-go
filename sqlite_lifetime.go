package luaskills

import "sync"

// sqliteScopeMu guards native-owner admission and retired callback generations.
// sqliteScopeMu 保护原生所有者准入及退役回调代次。
var sqliteScopeMu sync.Mutex

// sqliteScopeCount includes legacy engines, embedded transports and pending creation calls.
// sqliteScopeCount 包含传统引擎、嵌入式传输及正在创建的调用。
var sqliteScopeCount int

// sqliteRetired contains cleanup operations for callbacks captured by possible native owners.
// sqliteRetired 包含可能被原生所有者捕获的回调清理操作。
var sqliteRetired []func()

// retainSQLiteProviderScope pins captured callbacks; returns an idempotent release for native destruction.
// retainSQLiteProviderScope 固定已捕获回调；返回原生销毁时使用的幂等释放函数。
func retainSQLiteProviderScope() func() {
	// Once prevents duplicate cleanup from decrementing another native owner's count.
	// once 防止重复清理减少其他原生所有者的计数。
	var once sync.Once
	sqliteScopeMu.Lock()
	sqliteScopeCount++
	sqliteScopeMu.Unlock()
	return func() {
		once.Do(func() {
			// Detached callbacks are destroyed outside the ownership lock.
			// 摘除的回调在所有权锁外销毁。
			var detached []func()
			sqliteScopeMu.Lock()
			sqliteScopeCount--
			if sqliteScopeCount == 0 {
				detached, sqliteRetired = sqliteRetired, nil
			}
			sqliteScopeMu.Unlock()
			for _, dispose := range detached {
				dispose()
			}
		})
	}
}

// retireSQLiteProvider retains dispose while native owners exist, otherwise invokes it immediately.
// retireSQLiteProvider 在原生所有者存在时保留 dispose，否则立即调用它。
func retireSQLiteProvider(dispose func()) {
	sqliteScopeMu.Lock()
	if sqliteScopeCount > 0 {
		sqliteRetired = append(sqliteRetired, dispose)
		sqliteScopeMu.Unlock()
		return
	}
	sqliteScopeMu.Unlock()
	dispose()
}
