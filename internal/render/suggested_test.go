package render

import (
	"testing"

	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/scorer"
)

func TestBuildSuggestedActions(t *testing.T) {
	t.Parallel()

	result := &graph.ImpactResult{
		Source: graph.AffectedFile{Path: "internal/parser/analysis.go"},
		Affected: []graph.AffectedFile{
			{Path: "handler/admin.go", RiskScore: 82, HasTestFile: true},
			{Path: "handler/legacy.go", RiskScore: 85, HasTestFile: false},
			{Path: "handler/user.go", RiskScore: 45, HasTestFile: false},
		},
	}
	riskScores := map[string]scorer.FileRisk{
		"internal/parser/analysis.go": {Score: 71, Band: scorer.BandMedium},
		"handler/admin.go":            {Score: 82, Band: scorer.BandHigh},
		"handler/legacy.go":           {Score: 85, Band: scorer.BandHigh},
		"handler/user.go":             {Score: 45, Band: scorer.BandLow},
	}

	got := BuildSuggestedActions(result, riskScores)
	if got == nil {
		t.Fatal("BuildSuggestedActions() = nil")
	}
	if len(got.PackagesToTest) != 2 {
		t.Fatalf("PackagesToTest len = %d, want 2", len(got.PackagesToTest))
	}
	if got.Commands[0] != "go test ./handler/... ./internal/parser/..." {
		t.Fatalf("Commands[0] = %q", got.Commands[0])
	}
	if got.Commands[1] != "rippl score internal/parser/analysis.go --format json" {
		t.Fatalf("Commands[1] = %q", got.Commands[1])
	}
	if len(got.UntestedHighRisk) != 1 || got.UntestedHighRisk[0].Path != "handler/legacy.go" {
		t.Fatalf("UntestedHighRisk = %#v, want only handler/legacy.go", got.UntestedHighRisk)
	}
}

func TestApplyAnalyzeFilters(t *testing.T) {
	t.Parallel()

	files := []FileOutput{
		{Path: "a.go", ImpactLevel: "direct", RiskScore: 90},
		{Path: "b.go", ImpactLevel: "direct", RiskScore: 60},
		{Path: "c.go", ImpactLevel: "indirect", RiskScore: 80},
		{Path: "d.go", ImpactLevel: "indirect", RiskScore: 40},
	}

	got := ApplyAnalyzeFilters(files, 2, 70)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Path != "a.go" || got[1].Path != "c.go" {
		t.Fatalf("filtered = %#v", got)
	}
}

func TestRecomputeSummaryWithTotal(t *testing.T) {
	t.Parallel()

	files := []FileOutput{
		{Path: "a.go", ImpactLevel: "direct", RiskScore: 90, HasTestFile: true},
		{Path: "c.go", ImpactLevel: "indirect", RiskScore: 80, HasTestFile: false},
	}
	summary := RecomputeSummary(files, 4)
	if summary.AffectedCount != 2 || summary.TotalAffectedCount != 4 {
		t.Fatalf("summary = %#v", summary)
	}
	if summary.DirectCount != 1 || summary.IndirectCount != 1 || summary.WithoutTests != 1 || summary.MaxRiskScore != 90 {
		t.Fatalf("summary counts = %#v", summary)
	}
}

func TestBuildFilterNote(t *testing.T) {
	t.Parallel()

	got := BuildFilterNote(10, 48, 10, 70)
	want := "Showing 10 of 48 affected files (--top 10, --min-risk 70)"
	if got != want {
		t.Fatalf("BuildFilterNote() = %q, want %q", got, want)
	}
}
