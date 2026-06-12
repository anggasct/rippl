package git

import (
	"context"
	"fmt"
	"time"

	"github.com/anggasct/rippl/internal/config"
)

type FileGitStats struct {
	Path         string
	CommitCount  int
	BugFixCount  int
	AuthorCount  int
	LastModified time.Time
	Churn        int
}

func CollectFileStats(ctx context.Context, moduleRoot string, files []string, cfg *config.Config) (map[string]FileGitStats, error) {
	return collectFileStats(ctx, moduleRoot, files, cfg, ExecRunner{})
}

func collectFileStats(
	ctx context.Context,
	moduleRoot string,
	files []string,
	cfg *config.Config,
	runner Runner,
) (map[string]FileGitStats, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	ok, err := IsRepo(ctx, moduleRoot, runner)
	if err != nil {
		return nil, fmt.Errorf("collect git stats: %w", err)
	}

	out := make(map[string]FileGitStats, len(files))
	if !ok {
		for _, path := range files {
			out[path] = zeroStats(path)
		}
		return out, nil
	}

	patterns, err := compileBugFixPatterns(cfg.Risk.BugFixPatterns)
	if err != nil {
		return nil, fmt.Errorf("collect git stats: %w", err)
	}

	since := cfg.Risk.Since
	for _, path := range files {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("collect git stats: %w", err)
		}
		stats, err := collectFileGitStats(ctx, runner, moduleRoot, path, since, patterns)
		if err != nil {
			return nil, fmt.Errorf("collect git stats: %w", err)
		}
		out[path] = stats
	}
	return out, nil
}

func zeroStats(path string) FileGitStats {
	return FileGitStats{Path: path}
}
