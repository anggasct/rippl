package main

import (
	"fmt"
	"path/filepath"

	"github.com/anggasct/rippl/internal/config"
)

func resolveRelativeFilePath(moduleRoot, fileArg string) (string, error) {
	absPath, err := filepath.Abs(fileArg)
	if err != nil {
		return "", &config.ExitError{Code: 2, Err: fmt.Errorf("resolve absolute path: %w", err)}
	}

	relPath, err := filepath.Rel(moduleRoot, absPath)
	if err != nil {
		return "", &config.ExitError{Code: 2, Err: fmt.Errorf("resolve relative path: %w", err)}
	}

	return relPath, nil
}
