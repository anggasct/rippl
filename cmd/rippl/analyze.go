package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/render"
	"github.com/anggasct/rippl/internal/scorer"
	"github.com/anggasct/rippl/internal/testmap"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	var (
		noCache bool
		top     int
		minRisk int
		compact bool
	)

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

			modulePath, err := config.ModulePath(moduleRoot)
			if err != nil {
				return err
			}

			out := buildOutput(cfg, modulePath, result, riskScores, coverageInfo, top, minRisk)
			return renderAnalyzeOutput(cmd, cfg, out, compact)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Force cold graph build")
	cmd.Flags().BoolVar(&compact, "compact", false, "Omit chain arrays in JSON output (with --format json)")
	cmd.Flags().IntVar(&top, "top", 0, "Max affected entries in output (0 = all)")
	cmd.Flags().IntVar(&minRisk, "min-risk", 0, "Min risk score 0-100 (0 = off)")

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

func buildOutput(
	cfg *config.Config,
	modulePath string,
	result *graph.ImpactResult,
	riskScores map[string]scorer.FileRisk,
	coverageInfo map[string]testmap.FileCoverage,
	top, minRisk int,
) render.Output {
	now := time.Now().UTC()

	srcRisk := riskScores[result.Source.Path]
	srcCov := coverageInfo[result.Source.Path]

	out := render.Output{
		Version:    render.OutputSchemaVersion,
		Command:    "analyze",
		SourceFile: result.Source.Path,
		Module:     modulePath,
		Generated:  now,
		Source: render.SourceOutput{
			Path:      result.Source.Path,
			RiskScore: result.Source.RiskScore,
			RiskBand:  string(srcRisk.Band),
			Coverage:  coveragePct(srcCov),
		},
	}

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
	}

	totalAffected := len(out.Files)
	if render.IsStructuredFormat(cfg.Output.Format) {
		out.SuggestedActions = render.BuildSuggestedActions(result, riskScores)
	}

	if top > 0 || minRisk > 0 {
		out.Files = render.ApplyAnalyzeFilters(out.Files, top, minRisk)
		out.Summary = render.RecomputeSummary(out.Files, totalAffected)
		out.FilterNote = render.BuildFilterNote(len(out.Files), totalAffected, top, minRisk)
	} else {
		out.Summary = summarizeFiles(out.Files)
	}

	return out
}

func summarizeFiles(files []render.FileOutput) render.SummaryOutput {
	summary := render.SummaryOutput{
		AffectedCount: len(files),
	}
	for _, f := range files {
		switch f.ImpactLevel {
		case string(graph.ImpactDirect):
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

func coveragePct(fc testmap.FileCoverage) float64 {
	if fc.CoveragePct != nil {
		return *fc.CoveragePct
	}
	return 0
}

func renderAnalyzeOutput(cmd *cobra.Command, cfg *config.Config, out render.Output, compact bool) error {
	if shouldCompactAnalyze(cfg, compact) {
		out = render.CompactAnalyzeOutput(out)
	}
	if err := renderOutput(cmd, cfg, out); err != nil {
		return err
	}
	return advisoryUntestedHighRisk(cfg, out)
}

func shouldCompactAnalyze(cfg *config.Config, compact bool) bool {
	if strings.EqualFold(cfg.Output.Format, string(render.FormatAgent)) {
		return true
	}
	return compact && strings.EqualFold(cfg.Output.Format, string(render.FormatJSON))
}

func advisoryUntestedHighRisk(cfg *config.Config, out render.Output) error {
	if !render.IsStructuredFormat(cfg.Output.Format) {
		return nil
	}
	if out.SuggestedActions == nil || len(out.SuggestedActions.UntestedHighRisk) == 0 {
		return nil
	}
	return &config.ExitError{Code: 4, Err: errors.New("untested high-risk files present")}
}

func renderOutput(cmd *cobra.Command, cfg *config.Config, out render.Output) error {
	noColor := cfg.Output.Color == "false"
	r, err := render.NewRendererWithWriterAndColor(cfg.Output.Format, cmd.OutOrStdout(), noColor)
	if err != nil {
		return err
	}
	return r.Render(cmd.Context(), out)
}
