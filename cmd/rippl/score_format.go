package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/anggasct/rippl/internal/render"
	"github.com/anggasct/rippl/internal/scorer"
	"github.com/spf13/cobra"
)

func printScoreBreakdown(cmd *cobra.Command, filePath string, result scorer.FileRisk) error {
	out := cmd.OutOrStdout()

	if _, err := fmt.Fprintf(out, "Risk Score: %d/100 (%s)\n", result.Score, bandLabel(result.Band)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "File: %s\n", filePath); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, strings.Repeat("-", 40)); err != nil {
		return err
	}

	for _, s := range result.Signals {
		if _, err := fmt.Fprintf(out, "\n%s\n", render.SignalLabel(s.Name)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(out, "  Raw:          %s\n", s.Raw); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(out, "  Normalized:   %d/100\n", s.Normalized); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(out, "  Weight:       %d\n", s.Weight); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(out, "  Contribution: %.1f\n", s.Contribution); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(out, "Note: lower test coverage increases the Coverage risk contribution."); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	return nil
}

func printScoreJSON(cmd *cobra.Command, modulePath, filePath string, result scorer.FileRisk) error {
	out := render.BuildScoreOutput(modulePath, filePath, result, time.Now().UTC())
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func bandLabel(b scorer.RiskBand) string {
	switch b {
	case scorer.BandHigh:
		return "high"
	case scorer.BandMedium:
		return "medium"
	case scorer.BandLow:
		return "low"
	case scorer.BandMinimal:
		return "minimal"
	default:
		return string(b)
	}
}
