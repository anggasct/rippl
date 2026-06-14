package render

import (
	"fmt"
	"sort"

	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/packages"
	"github.com/anggasct/rippl/internal/scorer"
)

const untestedHighRiskThreshold = 70

// BuildSuggestedActions constructs agent-oriented next steps from a full impact result.
func BuildSuggestedActions(result *graph.ImpactResult, riskScores map[string]scorer.FileRisk) *SuggestedActionsOutput {
	if result == nil {
		return nil
	}

	pkgDirs := packages.UniqueDirs(result)
	actions := &SuggestedActionsOutput{
		PackagesToTest:   packages.ToTestTargets(pkgDirs),
		UntestedHighRisk: []UntestedHighRiskEntry{},
	}

	if len(pkgDirs) > 0 {
		actions.Commands = append(actions.Commands, packages.GoTestCommand(pkgDirs))
	}
	actions.Commands = append(actions.Commands, fmt.Sprintf("rippl score %s --format json", result.Source.Path))

	for _, f := range result.Affected {
		if f.HasTestFile || f.RiskScore < untestedHighRiskThreshold {
			continue
		}
		band := string(riskScores[f.Path].Band)
		actions.UntestedHighRisk = append(actions.UntestedHighRisk, UntestedHighRiskEntry{
			Path:      f.Path,
			RiskScore: f.RiskScore,
			RiskBand:  band,
		})
	}

	sort.Slice(actions.UntestedHighRisk, func(i, j int) bool {
		if actions.UntestedHighRisk[i].RiskScore != actions.UntestedHighRisk[j].RiskScore {
			return actions.UntestedHighRisk[i].RiskScore > actions.UntestedHighRisk[j].RiskScore
		}
		return actions.UntestedHighRisk[i].Path < actions.UntestedHighRisk[j].Path
	})

	return actions
}
