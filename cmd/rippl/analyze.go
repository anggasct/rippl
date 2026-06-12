package main

import (
	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze <file>",
		Short: "Impact analysis for changing a file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			printNotImplemented(cmd)
		},
	}
}
