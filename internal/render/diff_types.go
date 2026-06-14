package render

import "time"

// DiffSchemaVersion is the diff JSON output schema version.
const DiffSchemaVersion = "1.1"

// DiffOutput is the JSON contract for rippl diff --format json.
type DiffOutput struct {
	Version          string                  `json:"version"`
	Command          string                  `json:"command"`
	Module           string                  `json:"module"`
	Ref              string                  `json:"ref"`
	Generated        time.Time               `json:"generated_at"`
	ChangedFiles     []string                `json:"changed_files"`
	Summary          DiffSummaryOutput       `json:"summary"`
	Files            []DiffFileOutput        `json:"affected"`
	SuggestedActions *SuggestedActionsOutput `json:"suggested_actions,omitempty"`
}

// DiffSummaryOutput aggregates diff impact counts.
type DiffSummaryOutput struct {
	ChangedCount       int `json:"changed_count"`
	AffectedCount      int `json:"affected_count"`
	DirectCount        int `json:"direct_count"`
	IndirectCount      int `json:"indirect_count"`
	WithoutTests       int `json:"without_tests"`
	MaxRiskScore       int `json:"max_risk_score"`
	TotalAffectedCount int `json:"total_affected_count"`
}

// DiffFileOutput describes one affected file in diff output.
type DiffFileOutput struct {
	Path        string   `json:"path"`
	ImpactLevel string   `json:"impact_level"`
	Depth       int      `json:"depth"`
	RiskScore   int      `json:"risk_score"`
	RiskBand    string   `json:"risk_band"`
	Coverage    float64  `json:"coverage_pct"`
	HasTestFile bool     `json:"has_test_file"`
	Chain       []string `json:"chain,omitempty"`
	Reason      string   `json:"reason"`
	TriggeredBy string   `json:"triggered_by"`
}

// BuildDiffOutput assembles diff JSON from merged entries.
func BuildDiffOutput(
	modulePath, ref string,
	changed []string,
	entries []DiffFileOutput,
	summary DiffSummaryOutput,
	actions *SuggestedActionsOutput,
	generated time.Time,
) DiffOutput {
	if changed == nil {
		changed = []string{}
	}
	if entries == nil {
		entries = []DiffFileOutput{}
	}
	return DiffOutput{
		Version:          DiffSchemaVersion,
		Command:          "diff",
		Module:           modulePath,
		Ref:              ref,
		Generated:        generated,
		ChangedFiles:     changed,
		Summary:          summary,
		Files:            entries,
		SuggestedActions: actions,
	}
}
