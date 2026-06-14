package render

import "time"

// OutputSchemaVersion is the analyze JSON output schema version.
const OutputSchemaVersion = "1.1"

// Format identifies a renderer format.
type Format string

const (
	FormatText    Format = "text"
	FormatJSON    Format = "json"
	FormatAgent   Format = "agent"
	FormatMermaid Format = "mermaid"
	FormatTUI     Format = "tui"
)

// Output is the shared output contract for all renderers.
// It represents the result of an analysis command in a format-agnostic way.
type Output struct {
	Version    string    `json:"version"`
	Command    string    `json:"command"`
	SourceFile string    `json:"source_file"`
	Module     string    `json:"module"`
	Generated  time.Time `json:"generated_at"`

	Source           SourceOutput            `json:"source"`
	Summary          SummaryOutput           `json:"summary"`
	Files            []FileOutput            `json:"affected"`
	SuggestedActions *SuggestedActionsOutput `json:"suggested_actions,omitempty"`
	FilterNote       string                  `json:"-"`
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
	AffectedCount      int `json:"affected_count"`
	DirectCount        int `json:"direct_count"`
	IndirectCount      int `json:"indirect_count"`
	WithoutTests       int `json:"without_tests"`
	MaxRiskScore       int `json:"max_risk_score"`
	TotalAffectedCount int `json:"total_affected_count,omitempty"`
}

// SuggestedActionsOutput provides agent-oriented next steps for analyze JSON.
type SuggestedActionsOutput struct {
	PackagesToTest   []string                `json:"packages_to_test"`
	Commands         []string                `json:"commands"`
	UntestedHighRisk []UntestedHighRiskEntry `json:"untested_high_risk"`
}

// UntestedHighRiskEntry describes a high-risk affected file without tests.
type UntestedHighRiskEntry struct {
	Path      string `json:"path"`
	RiskScore int    `json:"risk_score"`
	RiskBand  string `json:"risk_band"`
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
	Chain       []string `json:"chain,omitempty"`
	Reason      string   `json:"reason"`
}
