package testmap

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/parser"
)

func minimoduleGraph(t *testing.T) *graph.Graph {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "parser", "testdata", "minimodule"))
	if err != nil {
		t.Fatalf("Abs() error = %v", err)
	}
	analyses, err := parser.ParseModule(context.Background(), root, config.DefaultConfig())
	if err != nil {
		t.Fatalf("ParseModule() error = %v", err)
	}
	return graph.Build(analyses)
}

func TestMapFileTestsMinimodule(t *testing.T) {
	t.Parallel()
	g := minimoduleGraph(t)
	files := SourceFiles(g)

	mapped, err := MapFileTests(g, files, "")
	if err != nil {
		t.Fatalf("MapFileTests() error = %v", err)
	}

	beta := mapped["pkg/beta/beta.go"]
	if !beta.HasTestFile || beta.Status != StatusUnknown {
		t.Fatalf("beta = %#v, want HasTestFile with unknown status", beta)
	}
	if len(beta.TestFiles) == 0 {
		t.Fatalf("beta TestFiles = %#v", beta.TestFiles)
	}

	for _, path := range []string{"pkg/beta/helper.go"} {
		helper := mapped[path]
		if !helper.HasTestFile {
			t.Fatalf("%s = %#v, want HasTestFile via alpha_test import of beta", path, helper)
		}
	}

	alpha := mapped["pkg/alpha/alpha.go"]
	if alpha.HasTestFile {
		t.Fatalf("alpha = %#v, want no test (alpha_test imports beta, not alpha)", alpha)
	}

	gamma := mapped["pkg/gamma/gamma.go"]
	if gamma.HasTestFile {
		t.Fatalf("gamma = %#v, want no test", gamma)
	}
}

func TestMapFileTestsWithCoverProfile(t *testing.T) {
	t.Parallel()
	g := minimoduleGraph(t)
	profile := filepath.Join("testdata", "coverage.out")
	mapped, err := MapFileTests(g, SourceFiles(g), profile)
	if err != nil {
		t.Fatalf("MapFileTests() error = %v", err)
	}
	beta := mapped["pkg/beta/beta.go"]
	if beta.Status != StatusPercent || beta.CoveragePct == nil {
		t.Fatalf("beta = %#v, want percent status", beta)
	}
}

func TestToScorerCoverage(t *testing.T) {
	t.Parallel()
	pct := 42.0
	m := map[string]FileCoverage{
		"a.go": {Status: StatusNoTest},
		"b.go": {Status: StatusUnknown, HasTestFile: true},
		"c.go": {Status: StatusPercent, CoveragePct: &pct},
	}
	cov := ToScorerCoverage(m)
	if cov["a.go"] == nil || *cov["a.go"] != 0 {
		t.Fatalf("no test = %#v, want 0", cov["a.go"])
	}
	if cov["b.go"] != nil {
		t.Fatalf("unknown = %#v, want nil", cov["b.go"])
	}
	if cov["c.go"] == nil || *cov["c.go"] != 42 {
		t.Fatalf("percent = %#v, want 42", cov["c.go"])
	}
}

func TestApplyToImpact(t *testing.T) {
	t.Parallel()
	result := &graph.ImpactResult{
		Source:   graph.AffectedFile{Path: "pkg/beta/beta.go"},
		Affected: []graph.AffectedFile{{Path: "pkg/gamma/gamma.go"}},
	}
	info := map[string]FileCoverage{
		"pkg/beta/beta.go":   {HasTestFile: true},
		"pkg/gamma/gamma.go": {HasTestFile: false},
	}
	ApplyToImpact(result, info)
	if !result.Source.HasTestFile {
		t.Fatal("expected source HasTestFile")
	}
	if result.Affected[0].HasTestFile {
		t.Fatal("expected affected HasTestFile false")
	}
}
