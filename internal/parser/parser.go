package parser

import (
	"context"
	"fmt"
	"sort"

	"github.com/anggasct/rippl/internal/config"
)

// ParseModule parses all in-scope Go files under moduleRoot.
// cfg.Ignore controls glob exclusions; uses go/packages per ADR-006.
func ParseModule(ctx context.Context, moduleRoot string, cfg *config.Config) ([]FileAnalysis, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	mod, err := loadModule(ctx, moduleRoot, cfg.Ignore)
	if err != nil {
		return nil, fmt.Errorf("parse module: %w", err)
	}

	byPath := extractImportsAndExports(mod, cfg.Ignore)
	extractCallsAndTypeRefs(mod, byPath)

	out := make([]FileAnalysis, 0, len(byPath))
	for _, analysis := range byPath {
		out = append(out, *analysis)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})

	return out, nil
}
