package scorer

import (
	"context"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
)

type RiskBand string

const (
	BandHigh    RiskBand = "high"
	BandMedium  RiskBand = "medium"
	BandLow     RiskBand = "low"
	BandMinimal RiskBand = "minimal"
)

type Signal struct {
	Name         string
	Raw          string
	Normalized   int
	Weight       int
	Contribution float64
}

type FileRisk struct {
	Path    string
	Score   int
	Band    RiskBand
	Signals []Signal
}

// CoverageMap maps file paths to coverage percentage (0–100).
// A nil map value means coverage is unknown (neutral signal until CAP-905).
type CoverageMap map[string]*float64

type RiskScorer interface {
	ScoreFiles(
		ctx context.Context,
		moduleRoot string,
		g *graph.Graph,
		files []string,
		coverage CoverageMap,
		cfg *config.Config,
	) (map[string]FileRisk, error)
}
