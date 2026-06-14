package render

import "time"

// ContextSchemaVersion is the context JSON output schema version.
const ContextSchemaVersion = "1.1"

// ContextOutput is the JSON contract for rippl context.
type ContextOutput struct {
	Version          string                  `json:"version"`
	Command          string                  `json:"command"`
	Module           string                  `json:"module"`
	SourceFile       string                  `json:"source_file"`
	Generated        time.Time               `json:"generated_at"`
	Summary          SummaryOutput           `json:"summary"`
	Source           SourceOutput            `json:"source"`
	Score            ContextScoreOutput      `json:"score"`
	Files            []FileOutput            `json:"affected"`
	SuggestedActions *SuggestedActionsOutput `json:"suggested_actions,omitempty"`
}

// ContextScoreOutput embeds score breakdown for the source file.
type ContextScoreOutput struct {
	Score    int           `json:"score"`
	RiskBand string        `json:"risk_band"`
	Signals  []ScoreSignal `json:"signals"`
}

// BuildContextOutput assembles context JSON.
func BuildContextOutput(
	modulePath string,
	sourceFile string,
	summary SummaryOutput,
	source SourceOutput,
	score ContextScoreOutput,
	files []FileOutput,
	actions *SuggestedActionsOutput,
	generated time.Time,
) ContextOutput {
	if files == nil {
		files = []FileOutput{}
	}
	if score.Signals == nil {
		score.Signals = []ScoreSignal{}
	}
	return ContextOutput{
		Version:          ContextSchemaVersion,
		Command:          "context",
		Module:           modulePath,
		SourceFile:       sourceFile,
		Generated:        generated,
		Summary:          summary,
		Source:           source,
		Score:            score,
		Files:            files,
		SuggestedActions: actions,
	}
}

// BuildContextScore maps a file risk result to embedded score output.
func BuildContextScore(result ScoreOutput) ContextScoreOutput {
	return ContextScoreOutput{
		Score:    result.Score,
		RiskBand: result.RiskBand,
		Signals:  result.Signals,
	}
}
