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
			if render.IsStructuredFormat(cfg.Output.Format) {
				if err := printJSON(cmd, out); err != nil {
					return err
				}
				return advisoryUntestedHighRiskActions(actions)
			}
			return printDiffText(cmd, modulePath, ref, changed, files, summary, actions)
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

func printDiffText(cmd *cobra.Command, modulePath, ref string, changed []string, files []render.DiffFileOutput, summary render.DiffSummaryOutput, actions *render.SuggestedActionsOutput) error {
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Diff: %s @ %s\n", modulePath, ref); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Changed: %d files\n", summary.ChangedCount); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Affected: %d files (direct: %d, indirect: %d)\n", summary.AffectedCount, summary.DirectCount, summary.IndirectCount); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Without tests: %d  Max risk: %d\n", summary.WithoutTests, summary.MaxRiskScore); err != nil {
		return err
	}

	if len(files) > 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "\nAffected files:"); err != nil {
			return err
		}
		for i, f := range files {
			chain := strings.Join(f.Chain, " -> ")
			if chain == "" {
				chain = f.Path
			}
			testStatus := ""
			if !f.HasTestFile {
				testStatus = " [no test]"
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "  %d. [%s] %s (depth=%d, risk=%d, coverage=%.0f%%)%s\n",
				i+1, strings.ToUpper(f.ImpactLevel), chain, f.Depth, f.RiskScore, f.Coverage, testStatus); err != nil {
				return err
			}
		}
	}

	if actions != nil && len(actions.UntestedHighRisk) > 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "\nHigh-risk untested files:"); err != nil {
			return err
		}
		for _, u := range actions.UntestedHighRisk {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "  ! %s (risk=%d, band=%s)\n", u.Path, u.RiskScore, u.RiskBand); err != nil {
				return err
			}
		}
	}

	return nil
}
