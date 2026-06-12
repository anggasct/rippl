package git

import (
	"context"
	"strings"
)

func IsRepo(ctx context.Context, dir string, runner Runner) (bool, error) {
	if runner == nil {
		runner = ExecRunner{}
	}
	out, err := runner.Run(ctx, dir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		return false, nil
	}
	return strings.TrimSpace(string(out)) == "true", nil
}
