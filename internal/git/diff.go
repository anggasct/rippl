package git

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ErrNotRepository is returned when the module root is not inside a git repository.
var ErrNotRepository = errors.New("not a git repository")

// ChangedGoFiles lists module-relative .go files changed for the given git ref or range.
func ChangedGoFiles(ctx context.Context, moduleRoot, ref string, runner Runner) ([]string, error) {
	if runner == nil {
		runner = ExecRunner{}
	}

	ok, err := IsRepo(ctx, moduleRoot, runner)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotRepository
	}

	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("git diff: empty ref")
	}

	out, err := runner.Run(ctx, moduleRoot, "diff", "--name-only", "--diff-filter=ACMR", ref)
	if err != nil {
		return nil, fmt.Errorf("git diff %s: %w", ref, err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	seen := make(map[string]struct{})
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasSuffix(line, ".go") {
			continue
		}
		rel := filepath.ToSlash(line)
		if !filepath.IsLocal(rel) {
			continue
		}
		abs := filepath.Join(moduleRoot, filepath.FromSlash(rel))
		if _, err := filepath.Rel(moduleRoot, abs); err != nil {
			continue
		}
		if _, ok := seen[rel]; ok {
			continue
		}
		seen[rel] = struct{}{}
		files = append(files, rel)
	}
	return files, nil
}
