package graph

import (
	"testing"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/parser"
)

func testGraph(t *testing.T) *Graph {
	t.Helper()
	return Build(parseMinimodule(t))
}

func affectedByPath(t *testing.T, result *ImpactResult) map[string]AffectedFile {
	t.Helper()
	out := make(map[string]AffectedFile, len(result.Affected))
	for _, file := range result.Affected {
		out[file.Path] = file
	}
	return out
}

func TestAnalyzeImpactFromGamma(t *testing.T) {
	t.Parallel()

	result, err := AnalyzeImpact(testGraph(t), "pkg/gamma/gamma.go", 3)
	if err != nil {
		t.Fatalf("AnalyzeImpact() error = %v", err)
	}

	if result.Source.Path != "pkg/gamma/gamma.go" || result.Source.Level != ImpactSource {
		t.Fatalf("source = %#v", result.Source)
	}

	byPath := affectedByPath(t, result)
	beta, ok := byPath["pkg/beta/beta.go"]
	if !ok {
		t.Fatal("missing beta in affected set")
	}
	if beta.Depth != 1 || beta.Level != ImpactDirect {
		t.Fatalf("beta = %#v, want depth 1 direct", beta)
	}

	alpha, ok := byPath["pkg/alpha/alpha.go"]
	if !ok {
		t.Fatal("missing alpha in affected set")
	}
	if alpha.Depth != 2 || alpha.Level != ImpactIndirect {
		t.Fatalf("alpha = %#v, want depth 2 indirect", alpha)
	}
}

func TestAnalyzeImpactFromBeta(t *testing.T) {
	t.Parallel()

	result, err := AnalyzeImpact(testGraph(t), "pkg/beta/beta.go", 3)
	if err != nil {
		t.Fatalf("AnalyzeImpact() error = %v", err)
	}

	byPath := affectedByPath(t, result)
	alpha, ok := byPath["pkg/alpha/alpha.go"]
	if !ok {
		t.Fatal("missing alpha in affected set")
	}
	if alpha.Depth != 1 || alpha.Level != ImpactDirect {
		t.Fatalf("alpha = %#v, want depth 1 direct", alpha)
	}
	if _, ok := byPath["pkg/gamma/gamma.go"]; ok {
		t.Fatal("gamma should not appear when traversing forward from beta")
	}
}

func TestAnalyzeImpactMaxDepthTruncates(t *testing.T) {
	t.Parallel()

	result, err := AnalyzeImpact(testGraph(t), "pkg/gamma/gamma.go", 1)
	if err != nil {
		t.Fatalf("AnalyzeImpact() error = %v", err)
	}

	byPath := affectedByPath(t, result)
	if _, ok := byPath["pkg/beta/beta.go"]; !ok {
		t.Fatal("missing beta at maxDepth=1")
	}
	if _, ok := byPath["pkg/alpha/alpha.go"]; ok {
		t.Fatal("alpha should be excluded at maxDepth=1")
	}
}

func TestAnalyzeImpactUnknownSource(t *testing.T) {
	t.Parallel()

	_, err := AnalyzeImpact(testGraph(t), "pkg/missing/missing.go", 3)
	if err == nil {
		t.Fatal("AnalyzeImpact() error = nil, want unknown source")
	}
}

func TestAnalyzeImpactChainAndReason(t *testing.T) {
	t.Parallel()

	result, err := AnalyzeImpact(testGraph(t), "pkg/gamma/gamma.go", 3)
	if err != nil {
		t.Fatalf("AnalyzeImpact() error = %v", err)
	}

	alpha := affectedByPath(t, result)["pkg/alpha/alpha.go"]
	wantChain := []string{
		"pkg/gamma/gamma.go",
		"pkg/beta/beta.go",
		"pkg/alpha/alpha.go",
	}
	if len(alpha.Chain) != len(wantChain) {
		t.Fatalf("alpha chain = %#v, want %#v", alpha.Chain, wantChain)
	}
	for i, path := range wantChain {
		if alpha.Chain[i] != path {
			t.Fatalf("alpha chain[%d] = %q, want %q (full=%#v)", i, alpha.Chain[i], path, alpha.Chain)
		}
	}
	if alpha.Reason == "" {
		t.Fatalf("alpha reason = empty, want edge type from beta to alpha")
	}
}

func TestAnalyzeImpactSortOrder(t *testing.T) {
	t.Parallel()

	result, err := AnalyzeImpact(testGraph(t), "pkg/gamma/gamma.go", 3)
	if err != nil {
		t.Fatalf("AnalyzeImpact() error = %v", err)
	}

	if len(result.Affected) < 2 {
		t.Fatalf("affected = %#v, want at least 2 entries", result.Affected)
	}
	if result.Affected[0].Level != ImpactDirect {
		t.Fatalf("first affected = %#v, want direct before indirect", result.Affected[0])
	}
	last := result.Affected[len(result.Affected)-1]
	if last.Level != ImpactIndirect {
		t.Fatalf("last affected = %#v, want indirect", last)
	}
}

func TestAnalyzeImpactFromConfigExcludeTests(t *testing.T) {
	t.Parallel()

	g := testGraph(t)
	cfg := config.DefaultConfig()
	cfg.Impact.IncludeTests = false

	result, err := AnalyzeImpactFromConfig(g, "pkg/beta/beta.go", cfg)
	if err != nil {
		t.Fatalf("AnalyzeImpactFromConfig() error = %v", err)
	}

	for _, file := range result.Affected {
		if isTestFile(file.Path) {
			t.Fatalf("affected includes test file %q with IncludeTests=false", file.Path)
		}
	}
}

func TestAnalyzeImpactFromConfigIncludeTests(t *testing.T) {
	t.Parallel()

	g := testGraph(t)
	cfg := config.DefaultConfig()
	cfg.Impact.IncludeTests = true

	result, err := AnalyzeImpactFromConfig(g, "pkg/beta/beta.go", cfg)
	if err != nil {
		t.Fatalf("AnalyzeImpactFromConfig() error = %v", err)
	}

	if _, ok := affectedByPath(t, result)["pkg/beta/beta_test.go"]; !ok {
		t.Fatal("expected pkg/beta/beta_test.go in affected set with IncludeTests=true")
	}
}

func TestAnalyzeImpactFromConfigNilUsesDefaults(t *testing.T) {
	t.Parallel()

	result, err := AnalyzeImpactFromConfig(testGraph(t), "pkg/beta/beta.go", nil)
	if err != nil {
		t.Fatalf("AnalyzeImpactFromConfig() error = %v", err)
	}
	if len(result.Affected) == 0 {
		t.Fatal("expected affected files with default config")
	}
}

func TestAnalyzeImpactReasonUsesParserEdgeType(t *testing.T) {
	t.Parallel()

	result, err := AnalyzeImpact(testGraph(t), "pkg/gamma/gamma.go", 3)
	if err != nil {
		t.Fatalf("AnalyzeImpact() error = %v", err)
	}

	beta := affectedByPath(t, result)["pkg/beta/beta.go"]
	if beta.Reason != parser.EdgeCall && beta.Reason != parser.EdgeImport && beta.Reason != parser.EdgeTypeRef {
		t.Fatalf("beta reason = %q, want parser edge type", beta.Reason)
	}
}
