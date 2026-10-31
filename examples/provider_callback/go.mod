module github.com/LuaSkills/luaskills-sdk-go/examples/provider_callback

go 1.22

require (
	github.com/LuaSkills/luaskills-sdk-go v0.7.3
	github.com/mattn/go-sqlite3 v1.14.52
)

// Use this checkout while developing; standalone packaging removes this replacement.
// 开发时使用当前检出；独立打包时移除此替换。
replace github.com/LuaSkills/luaskills-sdk-go => ../..
