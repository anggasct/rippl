package main

import (
	"fmt"
	"path/filepath"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/scorer"
	"github.com/anggasct/rippl/internal/testmap"
	"github.com/spf13/cobra"
)

func newScoreCmd() *cobra.Command {
	var noCache bool

	cmd := &cobra.Command{
		Use:   "score <file>",
		Short: "Risk score breakdown for a file",
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

			relPath, err := filepath.Rel(moduleRoot, fileArg)
			if err != nil {
				return &config.ExitError{Code: 2, Err: fmt.Errorf("resolve relative path: %w", err)}
			}

			g, err := graph.LoadOrBuild(cmd.Context(), moduleRoot, cfg, noCache)
			if err != nil {
				return fmt.Errorf("load graph: %w", err)
			}

			coverageInfo, err := testmap.MapFileTests(g, []string{relPath}, "")
			if err != nil {
				return fmt.Errorf("collect coverage: %w", err)
			}
			coverage := testmap.ToScorerCoverage(coverageInfo)

			results, err := scorer.NewHeuristic().ScoreFiles(
				cmd.Context(),
				moduleRoot,
				g,
				[]string{relPath},
				coverage,
				cfg,
			)
			if err != nil {
				return fmt.Errorf("score file: %w", err)
			}

			result, ok := results[relPath]
			if !ok {
				return fmt.Errorf("scorer returned no result for %q", relPath)
			}

			return printScoreBreakdown(cmd, relPath, result)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Force cold graph build")

	return cmd
}
