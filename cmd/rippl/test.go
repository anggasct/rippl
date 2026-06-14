package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/packages"
	"github.com/anggasct/rippl/internal/render"
	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	var noCache bool

	cmd := &cobra.Command{
		Use:   "test <file>",
		Short: "Run affected tests for a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fileArg := args[0]

			if err := validateFileArg(fileArg); err != nil {
				return err
			}

			cfg := configForCmd(cmd)
			if err := rejectAgentFormat(cfg, "test"); err != nil {
				return err
			}
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

			allDirs := packages.UniqueDirs(result)
			runDirs, skippedCount := packages.WithTests(g, allDirs)
			skippedDirs := packages.SkippedDirs(allDirs, runDirs)

			if strings.EqualFold(cfg.Output.Format, string(render.FormatJSON)) {
				modulePath, err := config.ModulePath(moduleRoot)
				if err != nil {
					return err
				}
				if len(runDirs) == 0 {
					return printTestJSONEmpty(cmd, modulePath, relPath, skippedDirs)
				}
				return runTestsJSON(cmd.Context(), cmd, moduleRoot, modulePath, relPath, runDirs, skippedDirs)
			}

			if len(runDirs) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "no affected packages with tests")
				return err
			}

			return runTestsText(cmd.Context(), cmd, moduleRoot, runDirs, skippedCount)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Force cold graph build")

	return cmd
}

// runTestsText executes `go test ./<pkgDir/...>` for each package. It prints
// per-package results and a summary. If any `go test` invocation fails, it
// returns ExitError{Code: 1} to propagate the failure.
func runTestsText(ctx context.Context, cmd *cobra.Command, moduleRoot string, runDirs []string, skipped int) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	totalRun := 0
	failed := false

	for _, pkgDir := range runDirs {
		relPkgDir, err := filepath.Rel(moduleRoot, pkgDir)
		if err != nil {
			relPkgDir = pkgDir
		}
		target := "./" + filepath.ToSlash(relPkgDir) + "/..."

		if _, err := fmt.Fprintf(out, "=== RUN  %s\n", target); err != nil {
			return err
		}

		testCmd := exec.CommandContext(ctx, "go", "test", target)
		testCmd.Dir = moduleRoot
		testCmd.Stdout = out
		testCmd.Stderr = errOut
		if err := testCmd.Run(); err != nil {
			if _, ferr := fmt.Fprintf(errOut, "--- FAIL: %s\n", target); ferr != nil {
				return ferr
			}
			failed = true
		}
		totalRun++
	}

	if _, err := fmt.Fprintf(out, "packages run: %d", totalRun); err != nil {
		return err
	}
	if skipped > 0 {
		if _, err := fmt.Fprintf(out, ", skipped: %d (no test files)", skipped); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}

	if failed {
		return &config.ExitError{Code: 1, Err: errors.New("test failure")}
	}
	return nil
}
