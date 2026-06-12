package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anggasct/rippl/internal/config"
)

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}

func minimoduleRoot(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "internal", "parser", "testdata", "minimodule")
}

func defaultTestConfig() *config.Config {
	return config.DefaultConfig()
}

func analysisByPath(t *testing.T, analyses []FileAnalysis) map[string]FileAnalysis {
	t.Helper()

	out := make(map[string]FileAnalysis, len(analyses))
	for _, analysis := range analyses {
		out[analysis.Path] = analysis
	}
	return out
}
