package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/render"
	"github.com/anggasct/rippl/internal/scorer"
	"github.com/anggasct/rippl/internal/testmap"
	"github.com/spf13/cobra"
)

type impactAnalysis struct {
	modulePath   string
	moduleRoot   string
	result       *graph.ImpactResult
	riskScores   map[string]scorer.FileRisk
	coverageInfo map[string]testmap.FileCoverage
}

func runImpactAnalysis(
	ctx context.Context,
	moduleRoot, relPath string,
	cfg *config.Config,
	noCache bool,
) (*impactAnalysis, *graph.Graph, error) {
	g, err := graph.LoadOrBuild(ctx, moduleRoot, cfg, noCache)
	if err != nil {
		return nil, nil, fmt.Errorf("load graph: %w", err)
	}

	result, err := graph.AnalyzeImpactFromConfig(g, relPath, cfg)
	if err != nil {
		return nil, g, err
	}

	allFiles := make([]string, 0, len(result.Affected)+1)
	allFiles = append(allFiles, result.Source.Path)
	for _, f := range result.Affected {
		allFiles = append(allFiles, f.Path)
	}

	coverageInfo, err := testmap.MapFileTests(g, allFiles, "")
	if err != nil {
		return nil, g, fmt.Errorf("collect coverage: %w", err)
	}
	coverage := testmap.ToScorerCoverage(coverageInfo)
	testmap.ApplyToImpact(result, coverageInfo)

	riskScores, err := scorer.NewHeuristic().ScoreFiles(ctx, moduleRoot, g, allFiles, coverage, cfg)
	if err != nil {
		return nil, g, fmt.Errorf("score files: %w", err)
	}
	scoreMap := make(map[string]int, len(riskScores))
	for path, fr := range riskScores {
		scoreMap[path] = fr.Score
	}
	graph.ApplyRiskScores(result, scoreMap)

	modulePath, err := config.ModulePath(moduleRoot)
	if err != nil {
		return nil, g, err
	}

	return &impactAnalysis{
		modulePath:   modulePath,
		moduleRoot:   moduleRoot,
		result:       result,
		riskScores:   riskScores,
		coverageInfo: coverageInfo,
	}, g, nil
}

func syntheticImpactFromUnion(changed []string, union []graph.UnionEntry) *graph.ImpactResult {
	sourcePath := ""
	if len(changed) > 0 {
		sourcePath = changed[0]
	}
	out := &graph.ImpactResult{
		Source: graph.AffectedFile{Path: sourcePath},
	}
	for _, entry := range union {
		out.Affected = append(out.Affected, entry.AffectedFile)
	}
	return out
}

func scoreUnionFiles(
	ctx context.Context,
	moduleRoot string,
	g *graph.Graph,
	union []graph.UnionEntry,
	cfg *config.Config,
) (map[string]scorer.FileRisk, map[string]testmap.FileCoverage, error) {
	paths := make([]string, 0, len(union))
	for _, entry := range union {
		paths = append(paths, entry.Path)
	}
	coverageInfo, err := testmap.MapFileTests(g, paths, "")
	if err != nil {
		return nil, nil, err
	}
	coverage := testmap.ToScorerCoverage(coverageInfo)
	riskScores, err := scorer.NewHeuristic().ScoreFiles(ctx, moduleRoot, g, paths, coverage, cfg)
	if err != nil {
		return nil, nil, err
	}
	return riskScores, coverageInfo, nil
}

func printJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func advisoryUntestedHighRiskActions(actions *render.SuggestedActionsOutput) error {
	if actions == nil || len(actions.UntestedHighRisk) == 0 {
		return nil
	}
	return &config.ExitError{Code: 4, Err: errors.New("untested high-risk files present")}
}
