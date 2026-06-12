package graph

import (
	"context"
	"fmt"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/parser"
)

func LoadOrBuild(ctx context.Context, moduleRoot string, cfg *config.Config, noCache bool) (*Graph, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	if noCache {
		return buildAndMaybeSave(ctx, moduleRoot, cfg, false)
	}

	cached, cachedMTimes, err := Load(moduleRoot, cfg.Cache.Dir)
	if err == nil {
		currentMTimes, statErr := CollectMTimes(moduleRoot, cached.Files())
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

	if err := config.EnsureCacheDir(moduleRoot, cfg.Cache.Dir); err != nil {
		return nil, fmt.Errorf("build graph: %w", err)
	}

	mtimes, err := CollectMTimes(moduleRoot, graph.Files())
	if err != nil {
		return nil, fmt.Errorf("build graph: %w", err)
	}

	if err := Save(moduleRoot, cfg.Cache.Dir, graph, mtimes); err != nil {
		return nil, fmt.Errorf("build graph: %w", err)
	}

	return graph, nil
}
