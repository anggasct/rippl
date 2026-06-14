package graph

import (
	"fmt"
	"sort"
	"strings"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/parser"
)

type ImpactLevel string

const (
	ImpactSource   ImpactLevel = "source"
	ImpactDirect   ImpactLevel = "direct"
	ImpactIndirect ImpactLevel = "indirect"
)

type AffectedFile struct {
	Path        string
	Depth       int
	Level       ImpactLevel
	Chain       []string
	Reason      parser.EdgeType
	RiskScore   int
	HasTestFile bool
}

type ImpactResult struct {
	Source   AffectedFile
	Affected []AffectedFile
}

type impactOptions struct {
	maxDepth     int
	includeTests bool
}

// AnalyzeImpact runs BFS from sourcePath through forward edges up to maxDepth hops.
// includeTests defaults to true; use AnalyzeImpactFromConfig for config-driven options.
func AnalyzeImpact(g *Graph, sourcePath string, maxDepth int) (*ImpactResult, error) {
	return analyzeImpact(g, sourcePath, impactOptions{
		maxDepth:     maxDepth,
		includeTests: true,
	})
}

// AnalyzeImpactFromConfig runs BFS using cfg.Impact.MaxDepth and cfg.Impact.IncludeTests.
// nil cfg uses config.DefaultConfig().
func AnalyzeImpactFromConfig(g *Graph, sourcePath string, cfg *config.Config) (*ImpactResult, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return analyzeImpact(g, sourcePath, impactOptions{
		maxDepth:     cfg.Impact.MaxDepth,
		includeTests: cfg.Impact.IncludeTests,
	})
}

func analyzeImpact(g *Graph, sourcePath string, opts impactOptions) (*ImpactResult, error) {
	if g == nil {
		return nil, fmt.Errorf("analyze impact: nil graph")
	}
	if _, ok := g.Node(sourcePath); !ok {
		return nil, fmt.Errorf("analyze impact: unknown source %q", sourcePath)
	}
	if opts.maxDepth < 0 {
		return nil, fmt.Errorf("analyze impact: max depth must be >= 0")
	}

	result := &ImpactResult{
		Source: AffectedFile{
			Path:  sourcePath,
			Depth: 0,
			Level: ImpactSource,
			Chain: []string{sourcePath},
		},
	}

	type queueItem struct {
		path  string
		depth int
		chain []string
	}

	queue := []queueItem{{path: sourcePath, depth: 0, chain: []string{sourcePath}}}
	visited := map[string]bool{sourcePath: true}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.depth >= opts.maxDepth {
			continue
		}

		for _, edge := range g.Dependents(item.path) {
			next := edge.Target
			if visited[next] {
				continue
			}
			if !opts.includeTests && isTestFile(next) {
				continue
			}
			if g.nodes[next] == nil {
				continue
			}

			visited[next] = true
			chain := append(append([]string(nil), item.chain...), next)
			depth := item.depth + 1

			result.Affected = append(result.Affected, AffectedFile{
				Path:   next,
				Depth:  depth,
				Level:  impactLevel(depth),
				Chain:  chain,
				Reason: edge.Type,
			})

			queue = append(queue, queueItem{
				path:  next,
				depth: depth,
				chain: chain,
			})
		}
	}

	sortAffected(result.Affected)
	return result, nil
}

// ApplyRiskScores sets RiskScore on the source and affected files from scores, then re-sorts affected.
func ApplyRiskScores(result *ImpactResult, scores map[string]int) {
	if result == nil {
		return
	}
	if score, ok := scores[result.Source.Path]; ok {
		result.Source.RiskScore = score
	}
	for i := range result.Affected {
		if score, ok := scores[result.Affected[i].Path]; ok {
			result.Affected[i].RiskScore = score
		}
	}
	sortAffected(result.Affected)
}

// ApplyTestInfo sets HasTestFile on the source and affected files.
func ApplyTestInfo(result *ImpactResult, hasTest map[string]bool) {
	if result == nil {
		return
	}
	if isTestFile(result.Source.Path) {
		result.Source.HasTestFile = true
	} else if v, ok := hasTest[result.Source.Path]; ok {
		result.Source.HasTestFile = v
	}
	for i := range result.Affected {
		path := result.Affected[i].Path
		if isTestFile(path) {
			result.Affected[i].HasTestFile = true
			continue
		}
		if v, ok := hasTest[path]; ok {
			result.Affected[i].HasTestFile = v
		}
	}
}

func impactLevel(depth int) ImpactLevel {
	switch depth {
	case 0:
		return ImpactSource
	case 1:
		return ImpactDirect
	default:
		return ImpactIndirect
	}
}

func sortAffected(files []AffectedFile) {
	sort.Slice(files, func(i, j int) bool {
		ri, rj := levelRank(files[i].Level), levelRank(files[j].Level)
		if ri != rj {
			return ri < rj
		}
		if files[i].Depth != files[j].Depth {
			return files[i].Depth < files[j].Depth
		}
		if files[i].RiskScore != files[j].RiskScore {
			return files[i].RiskScore > files[j].RiskScore
		}
		return files[i].Path < files[j].Path
	})
}

func levelRank(level ImpactLevel) int {
	switch level {
	case ImpactDirect:
		return 0
	case ImpactIndirect:
		return 1
	default:
		return 2
	}
}

func isTestFile(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}
