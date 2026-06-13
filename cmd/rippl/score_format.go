package main

import (
	"fmt"
	"strings"

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
		if _, err := fmt.Fprintf(out, "\n%s\n", signalDisplayName(s.Name)); err != nil {
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

	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	return nil
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

func signalDisplayName(name string) string {
	switch name {
	case "bug_fix_ratio":
		return "Bug-Fix Ratio"
	case "fan_out":
		return "Fan-Out"
	case "churn_rate":
		return "Churn Rate"
	case "author_count":
		return "Author Count"
	case "stale_age":
		return "Stale Age"
	case "test_coverage":
		return "Test Coverage"
	default:
		return name
	}
}
