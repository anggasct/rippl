package parser

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/anggasct/rippl/internal/config"
)

func minimoduleRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "testdata", "minimodule"))
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
