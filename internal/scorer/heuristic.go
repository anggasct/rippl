package scorer

import (
	"context"
	"fmt"
	"time"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/git"
	"github.com/anggasct/rippl/internal/graph"
)

type HeuristicScorer struct {
	now time.Time
}

func NewHeuristic() RiskScorer {
	return &HeuristicScorer{now: time.Now().UTC()}
}

func (h *HeuristicScorer) ScoreFiles(
	ctx context.Context,
	moduleRoot string,
	g *graph.Graph,
	files []string,
	coverage CoverageMap,
	cfg *config.Config,
) (map[string]FileRisk, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	if err := validateWeights(cfg.Risk.Weights); err != nil {
		return nil, fmt.Errorf("score files: %w", err)
	}

	gitStats, err := git.CollectFileStats(ctx, moduleRoot, files, cfg)
	if err != nil {
		return nil, fmt.Errorf("score files: %w", err)
	}

	now := h.now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	out := make(map[string]FileRisk, len(files))
	for _, path := range files {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("score files: %w", err)
		}
		stats := gitStats[path]
		signals := collectSignals(g, path, stats, coverage, cfg.Risk.Weights, now)
		score := scoreFromSignals(signals)
		out[path] = FileRisk{
			Path:    path,
			Score:   score,
			Band:    BandForScore(score),
			Signals: signals,
		}
	}
	return out, nil
}

func validateWeights(w config.RiskWeights) error {
	checks := []struct {
		name string
		v    int
	}{
		{"bug_fix_ratio", w.BugFixRatio},
		{"fan_out", w.FanOut},
		{"churn_rate", w.ChurnRate},
		{"author_count", w.AuthorCount},
		{"stale_age", w.StaleAge},
		{"test_coverage", w.TestCoverage},
	}
	for _, c := range checks {
		if c.v < 0 {
			return fmt.Errorf("risk weight %q must be >= 0", c.name)
		}
	}
	return nil
}
