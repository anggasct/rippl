package graph

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/parser"
)

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repo root not found")
		}
		dir = parent
	}
}

func minimoduleRootPath() (string, error) {
	root, err := findRepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "internal", "parser", "testdata", "minimodule"), nil
}

func minimoduleRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(mustMinimoduleRootPath(t))
	if err != nil {
		t.Fatalf("minimodule root: %v", err)
	}
	return root
}

func mustMinimoduleRootPath(tb testing.TB) string {
	tb.Helper()
	root, err := minimoduleRootPath()
	if err != nil {
		tb.Fatalf("minimodule root: %v", err)
	}
	return root
}

func parseMinimodule(t *testing.T) []parser.FileAnalysis {
	t.Helper()

	analyses, err := parser.ParseModule(context.Background(), minimoduleRoot(t), config.DefaultConfig())
	if err != nil {
		t.Fatalf("ParseModule() error = %v", err)
	}
	return analyses
}

func hasDependencyTarget(edges []Edge, target string) bool {
	for _, edge := range edges {
		if edge.Target == target {
			return true
		}
	}
	return false
}

func hasDependentTarget(edges []Edge, target string) bool {
	for _, edge := range edges {
		if edge.Target == target {
			return true
		}
	}
	return false
}

func hasExport(exports []parser.Export, name, kind string) bool {
	for _, exp := range exports {
		if exp.Name == name && exp.Kind == kind {
			return true
		}
	}
	return false
}
