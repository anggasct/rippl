package main

import (
	"github.com/spf13/cobra"
)

func newGraphCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "graph",
		Short: "Export full dependency graph",
		Run: func(cmd *cobra.Command, args []string) {
			printNotImplemented(cmd)
		},
	}
}
