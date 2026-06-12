package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func EnsureCacheDir(moduleRoot, cacheDir string) error {
	path := cacheDir
	if !filepath.IsAbs(path) {
		path = filepath.Join(moduleRoot, cacheDir)
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create cache dir %q: %w", path, err)
	}

	testFile := filepath.Join(path, ".write-test")
	if err := os.WriteFile(testFile, []byte("ok"), 0o644); err != nil {
		return fmt.Errorf("cache dir %q is not writable: %w", path, err)
	}
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("cleanup cache dir test file: %w", err)
	}

	return nil
}
