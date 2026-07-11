// Package runtimeassets implements private runtime-asset validation helpers.
// Package runtimeassets 实现私有运行时资产校验辅助能力。
package runtimeassets

import (
	"path/filepath"
	"strings"
)

// HasUnsafeRelativeSegment reports whether one relative path contains ambiguous segments.
// HasUnsafeRelativeSegment 返回单个相对路径是否包含不明确片段。
func HasUnsafeRelativeSegment(pathText string) bool {
	normalized := strings.ReplaceAll(pathText, "\\", "/")
	for _, segment := range strings.Split(normalized, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return true
		}
	}
	return false
}

// IsStrictlyInside reports whether one path is strictly inside one root directory.
// IsStrictlyInside 返回单个路径是否严格位于 root 目录内部。
func IsStrictlyInside(rootPath string, candidatePath string) bool {
	relativePath, err := filepath.Rel(rootPath, candidatePath)
	if err != nil {
		return false
	}
	return relativePath != "." && relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator))
}
