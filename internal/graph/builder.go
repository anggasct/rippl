package graph

import (
	"context"
	"fmt"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/parser"
)

// LoadOrBuild loads a cached graph when valid, otherwise parses the module and builds a fresh graph.
// nil cfg uses config.DefaultConfig(), matching parser.ParseModule.
func LoadOrBuild(ctx context.Context, moduleRoot string, cfg *config.Config, noCache bool) (*Graph, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	if noCache {
		return buildAndMaybeSave(ctx, moduleRoot, cfg, false)
	}

	cached, cachedMTimes, err := Load(ctx, moduleRoot, cfg.Cache.Dir)
	if err == nil {
		currentMTimes, statErr := CollectMTimes(ctx, moduleRoot, cached.Files())
		if statErr == nil && IsValid(cachedMTimes, currentMTimes) {
			return cached, nil
		}
	}

	return buildAndMaybeSave(ctx, moduleRoot, cfg, true)
}

func buildAndMaybeSave(ctx context.Context, moduleRoot string, cfg *config.Config, save bool) (*Graph, error) {
	analyses, err := parser.ParseModule(ctx, moduleRoot, cfg)
	if err != nil {
		return nil, fmt.Errorf("build graph: %w", err)
	}

	graph := Build(analyses)
	if !save {
		return graph, nil
	}

	// Cache dir resolution follows config.EnsureCacheDir / cfg.Cache.Dir semantics.
	if err := config.EnsureCacheDir(moduleRoot, cfg.Cache.Dir); err != nil {
		return nil, fmt.Errorf("build graph: %w", err)
	}

	mtimes, err := CollectMTimes(ctx, moduleRoot, graph.Files())
	if err != nil {
		return nil, fmt.Errorf("build graph: %w", err)
	}

	if err := Save(moduleRoot, cfg.Cache.Dir, graph, mtimes); err != nil {
		return nil, fmt.Errorf("build graph: %w", err)
	}

	return graph, nil
}
