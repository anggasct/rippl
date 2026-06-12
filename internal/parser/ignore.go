package parser

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func normalizePath(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func matchesIgnore(relPath string, patterns []string) bool {
	relPath = normalizePath(relPath)
	for _, pattern := range patterns {
		pattern = normalizePath(pattern)
		if ok, err := doublestar.Match(pattern, relPath); err == nil && ok {
			return true
		}
	}
	return false
}

func isGoFile(path string) bool {
	return strings.HasSuffix(path, ".go")
}
