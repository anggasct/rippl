package config

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
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

// ModulePath returns the module path from go.mod at moduleRoot.
// If go.mod has no module directive, it returns an empty string.
func ModulePath(moduleRoot string) (string, error) {
	data, err := os.ReadFile(filepath.Join(moduleRoot, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	mod, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", fmt.Errorf("parse go.mod: %w", err)
	}
	if mod.Module == nil {
		return "", nil
	}
	return mod.Module.Mod.Path, nil
}
