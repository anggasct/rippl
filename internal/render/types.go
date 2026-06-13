package render

import "time"

// Format identifies a renderer format.
type Format string

const (
	FormatText    Format = "text"
	FormatJSON    Format = "json"
	FormatMermaid Format = "mermaid"
	FormatTUI     Format = "tui"
)

// Output is the shared output contract for all renderers.
// It represents the result of an analysis command in a format-agnostic way.
type Output struct {
	Version     string    `json:"version"`
	Command     string    `json:"command"`
	SourceFile  string    `json:"source_file"`
	Module      string    `json:"module"`
	Generated   time.Time `json:"generated_at"`

	Source  SourceOutput  `json:"source"`
	Summary SummaryOutput `json:"summary"`
	Files   []FileOutput  `json:"affected"`
}

// SourceOutput describes the source file that was analyzed.
type SourceOutput struct {
	Path      string  `json:"path"`
	RiskScore int     `json:"risk_score"`
	RiskBand  string  `json:"risk_band"`
	Coverage  float64 `json:"coverage_pct"`
}

// SummaryOutput aggregates impact counts.
type SummaryOutput struct {
	AffectedCount int `json:"affected_count"`
	DirectCount   int `json:"direct_count"`
	IndirectCount int `json:"indirect_count"`
	WithoutTests  int `json:"without_tests"`
	MaxRiskScore  int `json:"max_risk_score"`
}

// FileOutput describes a single affected file entry.
type FileOutput struct {
	Path        string   `json:"path"`
	ImpactLevel string   `json:"impact_level"`
	Depth       int      `json:"depth"`
	RiskScore   int      `json:"risk_score"`
	RiskBand    string   `json:"risk_band"`
	Coverage    float64  `json:"coverage_pct"`
	HasTestFile bool     `json:"has_test_file"`
	Chain       []string `json:"chain"`
	Reason      string   `json:"reason"`
}
