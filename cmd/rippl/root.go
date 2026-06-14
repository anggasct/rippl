package main

import (
	"context"
	"os"

	"github.com/anggasct/rippl/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const version = "dev"

type cliFlags struct {
	format   string
	maxDepth int
	since    string
	noCache  bool
	config   string
	noColor  bool
}

func defaultFormat() string {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return "tui"
	}
	return "text"
}

func newRootCmd() *cobra.Command {
	flags := &cliFlags{}

	rootCmd := &cobra.Command{
		Use:   "rippl",
		Short: "Analyze change impact in Go modules",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "version" {
				return nil
			}
			return prepareRuntime(cmd, args, flags)
		},
	}

	rootCmd.PersistentFlags().StringVar(&flags.format, "format", defaultFormat(), "Output format: tui, json, mermaid, text")
	rootCmd.PersistentFlags().IntVar(&flags.maxDepth, "max-depth", 3, "Impact traversal depth")
	rootCmd.PersistentFlags().StringVar(&flags.since, "since", "12 months", "Git history window for risk signals")
	rootCmd.PersistentFlags().BoolVar(&flags.noCache, "no-cache", false, "Force cold graph build")
	rootCmd.PersistentFlags().StringVar(&flags.config, "config", ".rippl.yaml", "Config file path")
	rootCmd.PersistentFlags().BoolVar(&flags.noColor, "no-color", false, "Disable ANSI colors")

	rootCmd.AddCommand(
		newAnalyzeCmd(),
		newScoreCmd(),
		newTestCmd(),
		newGraphCmd(),
		newVersionCmd(),
	)

	return rootCmd
}

func prepareRuntime(cmd *cobra.Command, args []string, flags *cliFlags) error {
	var (
		root string
		err  error
	)

	switch {
	case len(args) > 0 && commandUsesFileArg(cmd.Name()):
		root, err = config.FindModuleRootFromPath(args[0])
	default:
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return cwdErr
		}
		root, err = config.FindModuleRoot(cwd)
	}
	if err != nil {
		if err == config.ErrNotGoModule {
			return &config.ExitError{Code: 2, Err: err}
		}
		return err
	}

	cfg, loadedFromFile, err := config.Load(root, flags.config)
	if err != nil {
		return &config.ExitError{Code: 1, Err: err}
	}

	applyFlagOverrides(cmd, cfg, loadedFromFile, flags)

	if err := config.EnsureCacheDir(root, cfg.Cache.Dir); err != nil {
		return &config.ExitError{Code: 1, Err: err}
	}

	cmd.SetContext(context.WithValue(cmd.Context(), configKey, cfg))

	return nil
}

func applyFlagOverrides(cmd *cobra.Command, cfg *config.Config, loadedFromFile bool, flags *cliFlags) {
	if cmd.Flags().Changed("format") {
		cfg.Output.Format = flags.format
	} else if !loadedFromFile {
		cfg.Output.Format = defaultFormat()
	}

	if cmd.Flags().Changed("max-depth") {
		cfg.Impact.MaxDepth = flags.maxDepth
	}

	if cmd.Flags().Changed("since") {
		cfg.Risk.Since = flags.since
	}

	if cmd.Flags().Changed("no-color") && flags.noColor {
		cfg.Output.Color = "false"
	} else if !cmd.Flags().Changed("no-color") && cfg.Output.Color == "" {
		cfg.Output.Color = "auto"
	}

	_ = flags.noCache
}

func commandUsesFileArg(name string) bool {
	switch name {
	case "analyze", "score", "test":
		return true
	default:
		return false
	}
}

type ctxKey string

const configKey ctxKey = "config"

func configForCmd(cmd *cobra.Command) *config.Config {
	if v := cmd.Context().Value(configKey); v != nil {
		if cfg, ok := v.(*config.Config); ok {
			return cfg
		}
	}
	return config.DefaultConfig()
}
