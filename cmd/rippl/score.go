package main

import (
	"github.com/spf13/cobra"
)

func newScoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "score <file>",
		Short: "Risk score breakdown for a file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			printNotImplemented(cmd)
		},
	}
}
