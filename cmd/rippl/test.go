package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
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

			packages, skipped := resolveAffectedPackages(g, result)

			if len(packages) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "no affected packages with tests")
				return err
			}

			return runTests(cmd.Context(), cmd, moduleRoot, packages, skipped)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Force cold graph build")

	return cmd
}

// resolveAffectedPackages computes the unique set of Go package directories
// that contain both affected files and _test.go files. It returns a sorted
// slice of package directory paths and the count of affected packages skipped
// due to having no test files.
func resolveAffectedPackages(g *graph.Graph, result *graph.ImpactResult) ([]string, int) {
	pkgSet := make(map[string]bool)

	// Source file's package
	pkgSet[filepath.Dir(result.Source.Path)] = true

	// All affected files' packages
	for _, f := range result.Affected {
		pkgSet[filepath.Dir(f.Path)] = true
	}

	tested := make([]string, 0, len(pkgSet))
	skipped := 0

	for pkgDir := range pkgSet {
		if pkgHasTests(g, pkgDir) {
			tested = append(tested, pkgDir)
		} else {
			skipped++
		}
	}

	sort.Strings(tested)
	return tested, skipped
}

// pkgHasTests reports whether the graph contains any _test.go file under pkgDir.
func pkgHasTests(g *graph.Graph, pkgDir string) bool {
	for _, f := range g.Files() {
		if filepath.Dir(f) == pkgDir && strings.HasSuffix(f, "_test.go") {
			return true
		}
	}
	return false
}

// runTests executes `go test ./<pkgDir/...>` for each package. It prints
// per-package results and a summary. If any `go test` invocation fails, it
// returns ExitError{Code: 1} to propagate the failure.
func runTests(ctx context.Context, cmd *cobra.Command, moduleRoot string, packages []string, skipped int) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	totalRun := 0
	failed := false

	for _, pkgDir := range packages {
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
