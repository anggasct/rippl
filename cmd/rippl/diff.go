package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/git"
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/render"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	var noCache bool

	cmd := &cobra.Command{
		Use:   "diff <ref>",
		Short: "Analyze impact of changed files in a git ref or range",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			cfg := configForCmd(cmd)

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			moduleRoot, err := config.FindModuleRoot(cwd)
			if err != nil {
				if errors.Is(err, config.ErrNotGoModule) {
					return &config.ExitError{Code: 2, Err: err}
				}
				return err
			}

			changed, err := git.ChangedGoFiles(cmd.Context(), moduleRoot, ref, nil)
			if err != nil {
				if errors.Is(err, git.ErrNotRepository) {
					return &config.ExitError{Code: 3, Err: err}
				}
				return err
			}

			modulePath, err := config.ModulePath(moduleRoot)
			if err != nil {
				return err
			}

			if !render.IsStructuredFormat(cfg.Output.Format) {
				return fmt.Errorf("diff requires --format json or agent")
			}

			g, err := graph.LoadOrBuild(cmd.Context(), moduleRoot, cfg, noCache)
			if err != nil {
				return fmt.Errorf("load graph: %w", err)
			}

			impactResults := make(map[string]*graph.ImpactResult)
			for _, path := range changed {
				if _, ok := g.Node(path); !ok {
					continue
				}
				result, err := graph.AnalyzeImpactFromConfig(g, path, cfg)
				if err != nil {
					return &config.ExitError{Code: 2, Err: err}
				}
				impactResults[path] = result
			}

			union := graph.UnionImpact(impactResults)
			riskScores, coverageInfo, err := scoreUnionFiles(cmd.Context(), moduleRoot, g, union, cfg)
			if err != nil {
				return fmt.Errorf("score union: %w", err)
			}

			files := make([]render.DiffFileOutput, 0, len(union))
			for _, entry := range union {
				fr := riskScores[entry.Path]
				cov := coverageInfo[entry.Path]
				triggeredBy := ""
				if len(entry.TriggeredBy) > 0 {
					triggeredBy = entry.TriggeredBy[0]
				}
				fo := render.DiffFileOutput{
					Path:        entry.Path,
					ImpactLevel: string(entry.Level),
					Depth:       entry.Depth,
					RiskScore:   entry.RiskScore,
					RiskBand:    string(fr.Band),
					Coverage:    coveragePct(cov),
					HasTestFile: entry.HasTestFile,
					Chain:       entry.Chain,
					Reason:      string(entry.Reason),
					TriggeredBy: triggeredBy,
				}
				if shouldCompactAnalyze(cfg, false) {
					fo.Chain = nil
				}
				files = append(files, fo)
			}

			summary := summarizeDiffFiles(changed, files)
			synthetic := syntheticImpactFromUnion(changed, union)
			actions := render.BuildSuggestedActions(synthetic, riskScores)

			out := render.BuildDiffOutput(modulePath, ref, changed, files, summary, actions, time.Now().UTC())
			if err := printJSON(cmd, out); err != nil {
				return err
			}
			return advisoryUntestedHighRiskActions(actions)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Force cold graph build")
	return cmd
}

func summarizeDiffFiles(changed []string, files []render.DiffFileOutput) render.DiffSummaryOutput {
	summary := render.DiffSummaryOutput{
		ChangedCount:       len(changed),
		AffectedCount:      len(files),
		TotalAffectedCount: len(files),
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

func newContextCmd() *cobra.Command {
	var (
		noCache bool
		top     int
	)

	cmd := &cobra.Command{
		Use:   "context <file>",
		Short: "One-shot session bootstrap for agents",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fileArg := args[0]
			if err := validateFileArg(fileArg); err != nil {
				return err
			}

			cfg := configForCmd(cmd)
			format := effectiveContextFormat(cfg)
			cfg.Output.Format = format

			moduleRoot, err := resolveModuleRoot(fileArg)
			if err != nil {
				return err
			}
			relPath, err := resolveRelativeFilePath(moduleRoot, fileArg)
			if err != nil {
				return err
			}

			analysis, _, err := runImpactAnalysis(cmd.Context(), moduleRoot, relPath, cfg, noCache)
			if err != nil {
				return &config.ExitError{Code: 2, Err: err}
			}

			out := buildOutput(cfg, analysis.modulePath, analysis.result, analysis.riskScores, analysis.coverageInfo, top, 0)
			if shouldCompactAnalyze(cfg, false) {
				out = render.CompactAnalyzeOutput(out)
			}

			scoreOut := render.BuildScoreOutput(analysis.modulePath, relPath, analysis.riskScores[relPath], time.Now().UTC())
			ctxOut := render.BuildContextOutput(
				analysis.modulePath,
				relPath,
				out.Summary,
				out.Source,
				render.BuildContextScore(scoreOut),
				out.Files,
				out.SuggestedActions,
				time.Now().UTC(),
			)

			if err := printJSON(cmd, ctxOut); err != nil {
				return err
			}
			return advisoryUntestedHighRiskActions(out.SuggestedActions)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Force cold graph build")
	cmd.Flags().IntVar(&top, "top", 15, "Max affected entries in output (0 = all)")
	return cmd
}

func effectiveContextFormat(cfg *config.Config) string {
	switch strings.ToLower(cfg.Output.Format) {
	case string(render.FormatJSON), string(render.FormatAgent):
		return cfg.Output.Format
	default:
		return string(render.FormatAgent)
	}
}
