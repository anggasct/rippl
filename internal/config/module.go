package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindModuleRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotGoModule
		}
		dir = parent
	}
}

func FindModuleRootFromPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	startDir := abs
	info, err := os.Stat(abs)
	if err == nil && !info.IsDir() {
		startDir = filepath.Dir(abs)
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("stat path: %w", err)
	}

	return FindModuleRoot(startDir)
}
