package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/packages"
	"github.com/anggasct/rippl/internal/render"
	"github.com/spf13/cobra"
)

func runTestsJSON(
	ctx context.Context,
	cmd *cobra.Command,
	moduleRoot, modulePath, sourceFile string,
	runDirs, skippedDirs []string,
) error {
	start := time.Now()

	packagesRun := packages.ToTestTargets(runDirs)
	packagesSkipped := packages.ToTestTargets(skippedDirs)
	failures := make([]render.TestFailure, 0)
	failed := false

	errOut := cmd.ErrOrStderr()

	for _, pkgDir := range runDirs {
		relPkgDir, err := filepath.Rel(moduleRoot, pkgDir)
		if err != nil {
			relPkgDir = pkgDir
		}
		target := "./" + filepath.ToSlash(relPkgDir) + "/..."
		pkgTarget := "./" + filepath.ToSlash(relPkgDir)

		if _, err := fmt.Fprintf(errOut, "=== RUN  %s\n", target); err != nil {
			return err
		}

		var capture bytes.Buffer
		testCmd := exec.CommandContext(ctx, "go", "test", target)
		testCmd.Dir = moduleRoot
		testCmd.Stdout = io.MultiWriter(errOut, &capture)
		testCmd.Stderr = io.MultiWriter(errOut, &capture)
		if err := testCmd.Run(); err != nil {
			if _, ferr := fmt.Fprintf(errOut, "--- FAIL: %s\n", target); ferr != nil {
				return ferr
			}
			failures = append(failures, render.TestFailure{
				Package: pkgTarget,
				Summary: failureSummary(capture.Bytes(), target),
			})
			failed = true
		}
	}

	exitCode := 0
	if failed {
		exitCode = 1
	}

	result := render.TestRunResult{
		Module:          modulePath,
		SourceFile:      sourceFile,
		PackagesRun:     packagesRun,
		PackagesSkipped: packagesSkipped,
		DurationMS:      time.Since(start).Milliseconds(),
		ExitCode:        exitCode,
		Failures:        failures,
	}

	if err := printTestJSON(cmd, result); err != nil {
		return err
	}

	if failed {
		return &config.ExitError{Code: 1, Err: errors.New("test failure")}
	}
	return nil
}

func printTestJSON(cmd *cobra.Command, result render.TestRunResult) error {
	out := render.BuildTestOutput(result, time.Now().UTC())
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func printTestJSONEmpty(cmd *cobra.Command, modulePath, sourceFile string, skippedDirs []string) error {
	result := render.TestRunResult{
		Module:          modulePath,
		SourceFile:      sourceFile,
		PackagesRun:     []string{},
		PackagesSkipped: packages.ToTestTargets(skippedDirs),
		DurationMS:      0,
		ExitCode:        0,
		Failures:        []render.TestFailure{},
	}
	return printTestJSON(cmd, result)
}

func failureSummary(output []byte, target string) string {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}
	return fmt.Sprintf("--- FAIL: %s", target)
}
