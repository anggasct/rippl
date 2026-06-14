package packages

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/parser"
)

func TestUniqueDirs(t *testing.T) {
	t.Parallel()

	result := &graph.ImpactResult{
		Source: graph.AffectedFile{Path: "pkg/alpha/alpha.go"},
		Affected: []graph.AffectedFile{
			{Path: "pkg/beta/beta.go"},
			{Path: "pkg/beta/helper.go"},
			{Path: "pkg/gamma/gamma.go"},
		},
	}

	got := UniqueDirs(result)
	want := []string{"pkg/alpha", "pkg/beta", "pkg/gamma"}
	if len(got) != len(want) {
		t.Fatalf("UniqueDirs() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("UniqueDirs()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestGoTestCommand(t *testing.T) {
	t.Parallel()

	got := GoTestCommand([]string{"internal/parser", "internal/graph"})
	want := "go test ./internal/parser/... ./internal/graph/..."
	if got != want {
		t.Fatalf("GoTestCommand() = %q, want %q", got, want)
	}
}

func TestToTestTargets(t *testing.T) {
	t.Parallel()

	got := ToTestTargets([]string{"pkg/beta", "pkg/alpha"})
	want := []string{"./pkg/beta", "./pkg/alpha"}
	if len(got) != len(want) {
		t.Fatalf("ToTestTargets() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ToTestTargets()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func parseMinimodule(t *testing.T) []parser.FileAnalysis {
	t.Helper()
	root := filepath.Join("..", "parser", "testdata", "minimodule")
	analyses, err := parser.ParseModule(context.Background(), root, config.DefaultConfig())
	if err != nil {
		t.Fatalf("ParseModule() error = %v", err)
	}
	return analyses
}

func TestAffectedWithTests(t *testing.T) {
	t.Parallel()

	g := graph.Build(parseMinimodule(t))
	result, err := graph.AnalyzeImpact(g, "pkg/gamma/gamma.go", 3)
	if err != nil {
		t.Fatalf("AnalyzeImpact() error = %v", err)
	}

	pkgs, skipped := AffectedWithTests(g, result)
	if len(pkgs) == 0 {
		t.Fatal("expected at least one package with tests")
	}
	foundBeta := false
	for _, pkg := range pkgs {
		if pkg == "pkg/beta" {
			foundBeta = true
		}
	}
	if !foundBeta {
		t.Fatalf("expected pkg/beta in %v", pkgs)
	}
	if skipped < 0 {
		t.Fatalf("skipped = %d, want >= 0", skipped)
	}
}

func TestSkippedDirs(t *testing.T) {
	t.Parallel()

	all := []string{"pkg/alpha", "pkg/beta", "pkg/gamma"}
	tested := []string{"pkg/alpha", "pkg/gamma"}
	got := SkippedDirs(all, tested)
	want := []string{"pkg/beta"}
	if len(got) != len(want) {
		t.Fatalf("SkippedDirs() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SkippedDirs()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
