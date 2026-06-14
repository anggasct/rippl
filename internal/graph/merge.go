package graph

import "sort"

// UnionEntry is an affected file merged from multiple impact analyses.
type UnionEntry struct {
	AffectedFile
	TriggeredBy []string
}

// UnionImpact merges affected files from per-source impact results.
// results maps each changed/trigger file path to its impact result.
func UnionImpact(results map[string]*ImpactResult) []UnionEntry {
	byPath := make(map[string]*UnionEntry)

	for trigger, result := range results {
		if result == nil {
			continue
		}
		for _, f := range result.Affected {
			existing, ok := byPath[f.Path]
			if !ok || f.Depth < existing.Depth {
				byPath[f.Path] = &UnionEntry{
					AffectedFile: f,
					TriggeredBy:  []string{trigger},
				}
				continue
			}
			if f.Depth == existing.Depth {
				existing.TriggeredBy = appendUnique(existing.TriggeredBy, trigger)
			}
		}
	}

	out := make([]UnionEntry, 0, len(byPath))
	for _, entry := range byPath {
		sort.Strings(entry.TriggeredBy)
		out = append(out, *entry)
	}

	sort.Slice(out, func(i, j int) bool {
		ri, rj := levelRank(out[i].Level), levelRank(out[j].Level)
		if ri != rj {
			return ri < rj
		}
		if out[i].Depth != out[j].Depth {
			return out[i].Depth < out[j].Depth
		}
		if out[i].RiskScore != out[j].RiskScore {
			return out[i].RiskScore > out[j].RiskScore
		}
		return out[i].Path < out[j].Path
	})
	return out
}

func appendUnique(items []string, v string) []string {
	for _, s := range items {
		if s == v {
			return items
		}
	}
	return append(items, v)
}
