package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/anggasct/rippl/internal/config"
)

func main() {
	os.Exit(run())
}

func run() int {
	if err := newRootCmd().Execute(); err != nil {
		var exitErr *config.ExitError
		if errors.As(err, &exitErr) {
			fmt.Fprintln(os.Stderr, exitErr.Error())
			return exitErr.Code
		}
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
