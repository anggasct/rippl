package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/render"
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

			relPath, err := filepath.Rel(moduleRoot, fileArg)
			if err != nil {
				return &config.ExitError{Code: 2, Err: fmt.Errorf("resolve relative path: %w", err)}
			}

			g, err := graph.LoadOrBuild(cmd.Context(), moduleRoot, cfg, noCache)
			if err != nil {
				return fmt.Errorf("load graph: %w", err)
			}

			result, err := graph.AnalyzeImpactFromConfig(g, relPath, cfg)
			if err != nil {
				return &config.ExitError{Code: 2, Err: err}
			}

			out := buildOutput(cfg, moduleRoot, result)
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

func buildOutput(cfg *config.Config, moduleRoot string, result *graph.ImpactResult) render.Output {
	now := time.Now().UTC()

	out := render.Output{
		Version:   version,
		Command:   "analyze",
		Module:    moduleRoot,
		Generated: now,
		Source: render.SourceOutput{
			Path:      result.Source.Path,
			RiskScore: result.Source.RiskScore,
		},
	}

	directCount := 0
	indirectCount := 0
	withoutTests := 0
	maxRisk := 0

	for _, f := range result.Affected {
		level := string(f.Level)
		out.Files = append(out.Files, render.FileOutput{
			Path:        f.Path,
			ImpactLevel: level,
			Depth:       f.Depth,
			RiskScore:   f.RiskScore,
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

func renderOutput(cmd *cobra.Command, cfg *config.Config, out render.Output) error {
	r, err := render.NewRendererWithWriter(cfg.Output.Format, cmd.OutOrStdout())
	if err != nil {
		return err
	}
	return r.Render(cmd.Context(), out)
}
