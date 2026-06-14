package main

import (
	"time"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/render"
	"github.com/spf13/cobra"
)

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
	switch cfg.Output.Format {
	case string(render.FormatJSON), string(render.FormatAgent):
		return cfg.Output.Format
	default:
		return string(render.FormatAgent)
	}
}
