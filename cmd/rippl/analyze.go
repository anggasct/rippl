package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/render"
	"github.com/anggasct/rippl/internal/scorer"
	"github.com/anggasct/rippl/internal/testmap"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	var noCache bool

	cmd := &cobra.Command{
		Use:   "analyze <file>",
		Short: "Impact analysis for changing a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fileArg := args[0]

			if err := validateFileArg(fileArg); err != nil {
				return err
			}

			cfg := configForCmd(cmd)
			moduleRoot, err := resolveModuleRoot(fileArg)
			if err != nil {
				return err
			}

			relPath, err := resolveRelativeFilePath(moduleRoot, fileArg)
			if err != nil {
				return err
			}

			g, err := graph.LoadOrBuild(cmd.Context(), moduleRoot, cfg, noCache)
			if err != nil {
				return fmt.Errorf("load graph: %w", err)
			}

			result, err := graph.AnalyzeImpactFromConfig(g, relPath, cfg)
			if err != nil {
				return &config.ExitError{Code: 2, Err: err}
			}

			// Collect risk scores and coverage for source + affected files.
			allFiles := make([]string, 0, len(result.Affected)+1)
			allFiles = append(allFiles, result.Source.Path)
			for _, f := range result.Affected {
				allFiles = append(allFiles, f.Path)
			}

			coverageInfo, err := testmap.MapFileTests(g, allFiles, "")
			if err != nil {
				return fmt.Errorf("collect coverage: %w", err)
			}
			coverage := testmap.ToScorerCoverage(coverageInfo)
			testmap.ApplyToImpact(result, coverageInfo)

			riskScores, err := scorer.NewHeuristic().ScoreFiles(
				cmd.Context(),
				moduleRoot,
				g,
				allFiles,
				coverage,
				cfg,
			)
			if err != nil {
				return fmt.Errorf("score files: %w", err)
			}
			scoreMap := make(map[string]int, len(riskScores))
			for path, fr := range riskScores {
				scoreMap[path] = fr.Score
			}
			graph.ApplyRiskScores(result, scoreMap)

			out := buildOutput(cfg, moduleRoot, result, riskScores, coverageInfo)
			return renderOutput(cmd, cfg, out)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Force cold graph build")

	return cmd
}

func validateFileArg(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &config.ExitError{Code: 2, Err: fmt.Errorf("file not found: %s", path)}
		}
		return &config.ExitError{Code: 2, Err: fmt.Errorf("stat file: %w", err)}
	}
	if info.IsDir() {
		return &config.ExitError{Code: 2, Err: fmt.Errorf("path is a directory: %s", path)}
	}
	return nil
}

func resolveModuleRoot(fileArg string) (string, error) {
	root, err := config.FindModuleRootFromPath(fileArg)
	if err != nil {
		if errors.Is(err, config.ErrNotGoModule) {
			return "", &config.ExitError{Code: 2, Err: err}
		}
		return "", err
	}
	return root, nil
}

func buildOutput(cfg *config.Config, moduleRoot string, result *graph.ImpactResult, riskScores map[string]scorer.FileRisk, coverageInfo map[string]testmap.FileCoverage) render.Output {
	now := time.Now().UTC()

	srcRisk := riskScores[result.Source.Path]
	srcCov := coverageInfo[result.Source.Path]

	out := render.Output{
		Version:    version,
		Command:    "analyze",
		SourceFile: result.Source.Path,
		Module:     moduleRoot,
		Generated:  now,
		Source: render.SourceOutput{
			Path:      result.Source.Path,
			RiskScore: result.Source.RiskScore,
			RiskBand:  string(srcRisk.Band),
			Coverage:  coveragePct(srcCov),
		},
	}

	directCount := 0
	indirectCount := 0
	withoutTests := 0
	maxRisk := 0

	for _, f := range result.Affected {
		level := string(f.Level)
		fRisk := riskScores[f.Path]
		fCov := coverageInfo[f.Path]

		out.Files = append(out.Files, render.FileOutput{
			Path:        f.Path,
			ImpactLevel: level,
			Depth:       f.Depth,
			RiskScore:   f.RiskScore,
			RiskBand:    string(fRisk.Band),
			Coverage:    coveragePct(fCov),
			HasTestFile: f.HasTestFile,
			Chain:       f.Chain,
			Reason:      string(f.Reason),
		})

		if f.Level == graph.ImpactDirect {
			directCount++
		} else {
			indirectCount++
		}
		if !f.HasTestFile {
			withoutTests++
		}
		if f.RiskScore > maxRisk {
			maxRisk = f.RiskScore
		}
	}

	out.Summary = render.SummaryOutput{
		AffectedCount: len(result.Affected),
		DirectCount:   directCount,
		IndirectCount: indirectCount,
		WithoutTests:  withoutTests,
		MaxRiskScore:  maxRisk,
	}

	return out
}

func coveragePct(fc testmap.FileCoverage) float64 {
	if fc.CoveragePct != nil {
		return *fc.CoveragePct
	}
	return 0
}

func renderOutput(cmd *cobra.Command, cfg *config.Config, out render.Output) error {
	noColor := cfg.Output.Color == "false"
	r, err := render.NewRendererWithWriterAndColor(cfg.Output.Format, cmd.OutOrStdout(), noColor)
	if err != nil {
		return err
	}
	return r.Render(cmd.Context(), out)
}
