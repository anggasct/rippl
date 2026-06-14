package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/graph"
	"github.com/spf13/cobra"
)

func newGraphCmd() *cobra.Command {
	var noCache bool
	var pkgFilter string

	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Export full dependency graph",
		Long:  "Export the full module dependency graph as Mermaid or JSON.\nUse --package to filter to a subgraph.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configForCmd(cmd)

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			moduleRoot, err := config.FindModuleRoot(cwd)
			if err != nil {
				if err == config.ErrNotGoModule {
					return &config.ExitError{Code: 2, Err: err}
				}
				return err
			}

			g, err := graph.LoadOrBuild(cmd.Context(), moduleRoot, cfg, noCache)
			if err != nil {
				return fmt.Errorf("load graph: %w", err)
			}

			export, err := graph.ExportFull(g, moduleRoot, pkgFilter)
			if err != nil {
				return &config.ExitError{Code: 1, Err: err}
			}

			format := strings.ToLower(cfg.Output.Format)
			if format == "tui" || format == "text" {
				format = "mermaid"
			}

			switch format {
			case "json":
				return writeGraphJSON(cmd, export)
			case "mermaid":
				return writeGraphMermaid(cmd, export)
			default:
				if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "unknown format %q for graph; using mermaid\n", cfg.Output.Format); err != nil {
					return err
				}
				return writeGraphMermaid(cmd, export)
			}
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Force cold graph build")
	cmd.Flags().StringVar(&pkgFilter, "package", "", "Filter to subgraph rooted at package prefix")

	return cmd
}

func writeGraphJSON(cmd *cobra.Command, export *graph.GraphExport) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(export)
}

func writeGraphMermaid(cmd *cobra.Command, export *graph.GraphExport) error {
	_, err := fmt.Fprint(cmd.OutOrStdout(), graph.MermaidGraph(export))
	return err
}
