package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test <file>",
		Short: "Run affected tests for a file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "not implemented")
		},
	}
}
