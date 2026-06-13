package scorer

import (
	"context"
	"testing"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
)

func TestHeuristicScorerNonRepoReturnsScores(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := t.TempDir()
	g := graph.Build(nil)

	scorer := NewHeuristic()
	files := []string{"pkg/a.go"}
	risks, err := scorer.ScoreFiles(ctx, dir, g, files, nil, config.DefaultConfig())
	if err != nil {
		t.Fatalf("ScoreFiles() error = %v", err)
	}
	r, ok := risks["pkg/a.go"]
	if !ok {
		t.Fatal("missing risk for pkg/a.go")
	}
	if r.Score < 0 || r.Score > 100 {
		t.Fatalf("Score = %d, out of range", r.Score)
	}
	if len(r.Signals) != 6 {
		t.Fatalf("len(Signals) = %d, want 6", len(r.Signals))
	}
	if r.Band != BandForScore(r.Score) {
		t.Fatalf("Band = %q, want %q", r.Band, BandForScore(r.Score))
	}
}

func TestHeuristicScorerNilConfigUsesDefaults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := t.TempDir()
	scorer := NewHeuristic()
	_, err := scorer.ScoreFiles(ctx, dir, nil, []string{"a.go"}, nil, nil)
	if err != nil {
		t.Fatalf("ScoreFiles() error = %v", err)
	}
}

func TestHeuristicScorerInvalidWeight(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.Risk.Weights.BugFixRatio = -1
	scorer := NewHeuristic()
	_, err := scorer.ScoreFiles(ctx, t.TempDir(), nil, []string{"a.go"}, nil, cfg)
	if err == nil {
		t.Fatal("expected error for negative weight")
	}
}

func TestHeuristicScorerContextCancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	scorer := NewHeuristic()
	_, err := scorer.ScoreFiles(ctx, t.TempDir(), nil, []string{"a.go"}, nil, config.DefaultConfig())
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

func TestValidateWeights(t *testing.T) {
	t.Parallel()
	if err := validateWeights(config.DefaultConfig().Risk.Weights); err != nil {
		t.Fatalf("validateWeights() error = %v", err)
	}
}
