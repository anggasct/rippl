package scorer

import (
	"fmt"
	"time"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/git"
	"github.com/anggasct/rippl/internal/graph"
)

const (
	signalBugFixRatio  = "bug_fix_ratio"
	signalFanOut       = "fan_out"
	signalChurnRate    = "churn_rate"
	signalAuthorCount  = "author_count"
	signalStaleAge     = "stale_age"
	signalTestCoverage = "test_coverage"
)

func clamp100(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func bugFixRatioSignal(stats git.FileGitStats, weight int) Signal {
	norm := 0
	raw := "0 bug-fix commits in 0 commits"
	if stats.CommitCount > 0 {
		norm = clamp100(stats.BugFixCount * 100 / stats.CommitCount)
		raw = fmt.Sprintf("%d bug-fix commits in %d commits", stats.BugFixCount, stats.CommitCount)
	}
	return signal(signalBugFixRatio, raw, norm, weight)
}

func fanOutSignal(g *graph.Graph, path string, weight int) Signal {
	count := 0
	if g != nil {
		seen := make(map[string]struct{})
		for _, edge := range g.Dependents(path) {
			seen[edge.Target] = struct{}{}
		}
		count = len(seen)
	}
	norm := clamp100(count * 15)
	raw := fmt.Sprintf("%d dependent files", count)
	return signal(signalFanOut, raw, norm, weight)
}

func churnRateSignal(stats git.FileGitStats, weight int) Signal {
	norm := clamp100(stats.Churn / 10)
	raw := fmt.Sprintf("%d lines changed", stats.Churn)
	return signal(signalChurnRate, raw, norm, weight)
}

func authorCountSignal(stats git.FileGitStats, weight int) Signal {
	n := stats.AuthorCount
	norm := authorCountNormalized(n)
	raw := fmt.Sprintf("%d unique authors", n)
	return signal(signalAuthorCount, raw, norm, weight)
}

func authorCountNormalized(n int) int {
	switch {
	case n <= 1:
		return 0
	case n >= 6:
		return 100
	default:
		return clamp100(20 * (n - 1))
	}
}

func staleAgeSignal(stats git.FileGitStats, now time.Time, weight int) Signal {
	if stats.LastModified.IsZero() {
		return signal(signalStaleAge, "never modified", 0, weight)
	}
	days := int(now.Sub(stats.LastModified).Hours() / 24)
	if days < 0 {
		days = 0
	}
	norm := clamp100(days * 100 / 180)
	raw := fmt.Sprintf("%d days since last change", days)
	return signal(signalStaleAge, raw, norm, weight)
}

func testCoverageSignal(coverage CoverageMap, path string, weight int) Signal {
	if coverage == nil {
		return signal(signalTestCoverage, "coverage unknown", 50, weight)
	}
	pct, ok := coverage[path]
	if !ok || pct == nil {
		return signal(signalTestCoverage, "coverage unknown", 50, weight)
	}
	norm := clamp100(100 - int(*pct))
	raw := fmt.Sprintf("%.1f%% coverage", *pct)
	return signal(signalTestCoverage, raw, norm, weight)
}

func signal(name, raw string, normalized, weight int) Signal {
	contribution := float64(normalized*weight) / 100
	return Signal{
		Name:         name,
		Raw:          raw,
		Normalized:   normalized,
		Weight:       weight,
		Contribution: contribution,
	}
}

func collectSignals(
	g *graph.Graph,
	path string,
	stats git.FileGitStats,
	coverage CoverageMap,
	weights config.RiskWeights,
	now time.Time,
) []Signal {
	return []Signal{
		bugFixRatioSignal(stats, weights.BugFixRatio),
		fanOutSignal(g, path, weights.FanOut),
		churnRateSignal(stats, weights.ChurnRate),
		authorCountSignal(stats, weights.AuthorCount),
		staleAgeSignal(stats, now, weights.StaleAge),
		testCoverageSignal(coverage, path, weights.TestCoverage),
	}
}

func scoreFromSignals(signals []Signal) int {
	total := 0.0
	for _, s := range signals {
		total += s.Contribution
	}
	return clamp100(int(total + 0.5))
}
