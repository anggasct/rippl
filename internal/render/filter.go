package render

import "fmt"

// ApplyAnalyzeFilters applies min-risk and top-N limits to affected file output.
func ApplyAnalyzeFilters(files []FileOutput, top, minRisk int) []FileOutput {
	filtered := files
	if minRisk > 0 {
		next := make([]FileOutput, 0, len(filtered))
		for _, f := range filtered {
			if f.RiskScore >= minRisk {
				next = append(next, f)
			}
		}
		filtered = next
	}
	if top > 0 && len(filtered) > top {
		filtered = filtered[:top]
	}
	return filtered
}

// RecomputeSummary rebuilds summary counts from the filtered affected file list.
func RecomputeSummary(files []FileOutput, totalAffected int) SummaryOutput {
	summary := SummaryOutput{
		AffectedCount: len(files),
	}
	if totalAffected > len(files) {
		summary.TotalAffectedCount = totalAffected
	}

	for _, f := range files {
		switch f.ImpactLevel {
		case "direct":
			summary.DirectCount++
		default:
			summary.IndirectCount++
		}
		if !f.HasTestFile {
			summary.WithoutTests++
		}
		if f.RiskScore > summary.MaxRiskScore {
			summary.MaxRiskScore = f.RiskScore
		}
	}

	return summary
}

// BuildFilterNote returns a human-readable note when analyze filters are active.
func BuildFilterNote(filtered, total, top, minRisk int) string {
	if top <= 0 && minRisk <= 0 {
		return ""
	}

	note := fmt.Sprintf("Showing %d of %d affected files", filtered, total)
	if top > 0 && minRisk > 0 {
		return fmt.Sprintf("%s (--top %d, --min-risk %d)", note, top, minRisk)
	}
	if top > 0 {
		return fmt.Sprintf("%s (--top %d)", note, top)
	}
	return fmt.Sprintf("%s (--min-risk %d)", note, minRisk)
}
