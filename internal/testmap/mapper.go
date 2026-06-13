package testmap

import (
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/scorer"
)

func MapFileTests(g *graph.Graph, files []string, coverProfilePath string) (map[string]FileCoverage, error) {
	out := make(map[string]FileCoverage, len(files))
	for _, path := range files {
		if isTestFile(path) {
			continue
		}
		out[path] = FileCoverage{
			Path:   path,
			Status: StatusNoTest,
		}
	}

	if g != nil {
		for _, path := range g.Files() {
			if !isTestFile(path) {
				continue
			}
			sourcePath, ok := sourcePathForTest(path)
			if !ok {
				continue
			}
			if _, tracked := out[sourcePath]; !tracked {
				continue
			}
			appendTestFile(out, sourcePath, path)
		}
		applyImportAssociations(g, out)
	}

	if coverProfilePath == "" {
		return out, nil
	}

	profiles, err := parseCoverProfile(coverProfilePath)
	if err != nil {
		return nil, err
	}
	if profiles == nil {
		return out, nil
	}

	for path, fc := range out {
		pct, ok := matchProfilePct(profiles, path)
		if !ok {
			continue
		}
		fc.CoveragePct = &pct
		fc.Status = StatusPercent
		out[path] = fc
	}
	return out, nil
}

func ToScorerCoverage(m map[string]FileCoverage) scorer.CoverageMap {
	out := make(scorer.CoverageMap, len(m))
	for path, fc := range m {
		switch fc.Status {
		case StatusPercent:
			if fc.CoveragePct != nil {
				pct := *fc.CoveragePct
				out[path] = &pct
			}
		case StatusNoTest:
			zero := 0.0
			out[path] = &zero
		default:
			out[path] = nil
		}
	}
	return out
}

func HasTestMap(m map[string]FileCoverage) map[string]bool {
	out := make(map[string]bool, len(m))
	for path, fc := range m {
		out[path] = fc.HasTestFile
	}
	return out
}

func ApplyToImpact(result *graph.ImpactResult, info map[string]FileCoverage) {
	graph.ApplyTestInfo(result, HasTestMap(info))
}

func SourceFiles(g *graph.Graph) []string {
	if g == nil {
		return nil
	}
	var out []string
	for _, path := range g.Files() {
		if !isTestFile(path) {
			out = append(out, path)
		}
	}
	return out
}
