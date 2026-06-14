package render

import (
	"math"
	"time"

	"github.com/anggasct/rippl/internal/scorer"
)

// ScoreSchemaVersion is the score JSON output schema version.
const ScoreSchemaVersion = "1"

// ScoreOutput is the JSON contract for rippl score --format json.
type ScoreOutput struct {
	Version    string        `json:"version"`
	Command    string        `json:"command"`
	Module     string        `json:"module"`
	SourceFile string        `json:"source_file"`
	Generated  time.Time     `json:"generated_at"`
	Score      int           `json:"score"`
	RiskBand   string        `json:"risk_band"`
	Signals    []ScoreSignal `json:"signals"`
}

// ScoreSignal is one risk signal in score JSON output.
type ScoreSignal struct {
	Name           string  `json:"name"`
	Label          string  `json:"label"`
	RawValue       float64 `json:"raw_value"`
	Weight         float64 `json:"weight"`
	Contribution   int     `json:"contribution"`
	Interpretation string  `json:"interpretation"`
}

// BuildScoreOutput maps scorer results to the score JSON contract.
func BuildScoreOutput(modulePath, sourceFile string, result scorer.FileRisk, generated time.Time) ScoreOutput {
	signals := make([]ScoreSignal, 0, len(result.Signals))
	for _, s := range result.Signals {
		signals = append(signals, ScoreSignal{
			Name:           s.Name,
			Label:          SignalLabel(s.Name),
			RawValue:       rawValueForSignal(s),
			Weight:         float64(s.Weight) / 100,
			Contribution:   int(math.Round(s.Contribution)),
			Interpretation: s.Raw,
		})
	}

	return ScoreOutput{
		Version:    ScoreSchemaVersion,
		Command:    "score",
		Module:     modulePath,
		SourceFile: sourceFile,
		Generated:  generated,
		Score:      result.Score,
		RiskBand:   scoreBandLabel(result.Band),
		Signals:    signals,
	}
}

// SignalLabel returns a human-readable label for a scorer signal name.
func SignalLabel(name string) string {
	switch name {
	case "bug_fix_ratio":
		return "Bug-fix ratio"
	case "fan_out":
		return "Fan-out"
	case "churn_rate":
		return "Churn rate"
	case "author_count":
		return "Author count"
	case "stale_age":
		return "Stale age"
	case "test_coverage":
		return "Coverage risk"
	default:
		return name
	}
}

func scoreBandLabel(b scorer.RiskBand) string {
	switch b {
	case scorer.BandHigh:
		return "high"
	case scorer.BandMedium:
		return "medium"
	case scorer.BandLow:
		return "low"
	case scorer.BandMinimal:
		return "minimal"
	default:
		return string(b)
	}
}

func rawValueForSignal(s scorer.Signal) float64 {
	switch s.Name {
	case "bug_fix_ratio":
		return float64(s.Normalized) / 100
	case "fan_out":
		return float64(s.Normalized) / 15
	case "churn_rate":
		return float64(s.Normalized) * 10
	case "author_count":
		return authorCountFromNormalized(s.Normalized)
	case "stale_age":
		return float64(s.Normalized) * 180 / 100
	case "test_coverage":
		if s.Raw == "coverage unknown" {
			return 0
		}
		return 100 - float64(s.Normalized)
	default:
		return float64(s.Normalized)
	}
}

// authorCountFromNormalized inverts scorer.authorCountNormalized.
func authorCountFromNormalized(norm int) float64 {
	switch {
	case norm <= 0:
		return 1
	case norm >= 100:
		return 6
	default:
		return float64(norm/20 + 1)
	}
}
