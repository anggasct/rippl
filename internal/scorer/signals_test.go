// trigger webhook
package scorer

import (
	"testing"
	"time"

	"github.com/anggasct/rippl/internal/git"
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/parser"
)

func TestBugFixRatioSignal(t *testing.T) {
	t.Parallel()
	s := bugFixRatioSignal(git.FileGitStats{CommitCount: 10, BugFixCount: 3}, 25)
	if s.Normalized != 30 {
		t.Fatalf("Normalized = %d, want 30", s.Normalized)
	}
	if s.Contribution != 7.5 {
		t.Fatalf("Contribution = %v, want 7.5", s.Contribution)
	}

	zero := bugFixRatioSignal(git.FileGitStats{}, 25)
	if zero.Normalized != 0 {
		t.Fatalf("zero commits Normalized = %d, want 0", zero.Normalized)
	}
}

func TestAuthorCountNormalized(t *testing.T) {
	t.Parallel()
	tests := []struct {
		n    int
		want int
	}{
		{0, 0},
		{1, 0},
		{2, 20},
		{5, 80},
		{6, 100},
		{10, 100},
	}
	for _, tc := range tests {
		if got := authorCountNormalized(tc.n); got != tc.want {
			t.Errorf("authorCountNormalized(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestStaleAgeSignal(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 13, 0, 0, 0, 0, time.UTC)
	last := now.AddDate(0, -7, 0)
	s := staleAgeSignal(git.FileGitStats{LastModified: last}, now, 10)
	if s.Normalized != 100 {
		t.Fatalf("Normalized = %d, want 100 for 7 months stale", s.Normalized)
	}

	recent := staleAgeSignal(git.FileGitStats{LastModified: now.AddDate(0, 0, -30)}, now, 10)
	if recent.Normalized != 16 {
		t.Fatalf("Normalized = %d, want 16 for 30 days", recent.Normalized)
	}
}

func TestTestCoverageSignal(t *testing.T) {
	t.Parallel()
	unknown := testCoverageSignal(nil, "a.go", 15)
	if unknown.Normalized != 50 {
		t.Fatalf("unknown Normalized = %d, want 50", unknown.Normalized)
	}

	pct := 45.2
	known := testCoverageSignal(CoverageMap{"a.go": &pct}, "a.go", 15)
	if known.Normalized != 55 {
		t.Fatalf("known Normalized = %d, want 55", known.Normalized)
	}
}

func TestChurnRateSignal(t *testing.T) {
	t.Parallel()
	s := churnRateSignal(git.FileGitStats{Churn: 1000}, 15)
	if s.Normalized != 100 {
		t.Fatalf("Normalized = %d, want 100", s.Normalized)
	}
}

func TestFanOutSignal(t *testing.T) {
	t.Parallel()
	g := graph.Build([]parser.FileAnalysis{
		{Path: "a.go", Package: "a"},
		{Path: "b.go", Package: "b", Imports: []parser.Edge{{TargetFile: "a.go", Type: parser.EdgeImport}}},
		{Path: "c.go", Package: "c", Imports: []parser.Edge{{TargetFile: "a.go", Type: parser.EdgeImport}}},
	})
	s := fanOutSignal(g, "a.go", 20)
	if s.Normalized != 30 {
		t.Fatalf("Normalized = %d, want 30 for 2 dependents", s.Normalized)
	}
}

func TestScoreFromSignals(t *testing.T) {
	t.Parallel()
	signals := []Signal{
		{Normalized: 100, Weight: 25, Contribution: 25},
		{Normalized: 0, Weight: 75, Contribution: 0},
	}
	if got := scoreFromSignals(signals); got != 25 {
		t.Fatalf("score = %d, want 25", got)
	}
}

func TestScoreFromSignalsCapsAt100(t *testing.T) {
	t.Parallel()
	signals := []Signal{
		{Normalized: 100, Weight: 100, Contribution: 100},
	}
	if got := scoreFromSignals(signals); got != 100 {
		t.Fatalf("score = %d, want 100", got)
	}
}
