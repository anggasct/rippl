package graph

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anggasct/rippl/internal/config"
)

func TestLoadOrBuildUsesCache(t *testing.T) {
	moduleRoot := minimoduleRoot(t)
	cacheDir := filepath.Join(t.TempDir(), ".rippl", "cache")
	cfg := config.DefaultConfig()
	cfg.Cache.Dir = cacheDir

	ctx := context.Background()

	first, err := LoadOrBuild(ctx, moduleRoot, cfg, false)
	if err != nil {
		t.Fatalf("LoadOrBuild() first error = %v", err)
	}

	second, err := LoadOrBuild(ctx, moduleRoot, cfg, false)
	if err != nil {
		t.Fatalf("LoadOrBuild() second error = %v", err)
	}

	if len(first.Files()) != len(second.Files()) {
		t.Fatalf("first files = %#v, second = %#v", first.Files(), second.Files())
	}
}

func TestLoadOrBuildNoCacheBypassesRead(t *testing.T) {
	moduleRoot := minimoduleRoot(t)
	cacheDir := filepath.Join(t.TempDir(), ".rippl", "cache")
	cfg := config.DefaultConfig()
	cfg.Cache.Dir = cacheDir

	ctx := context.Background()
	if _, err := LoadOrBuild(ctx, moduleRoot, cfg, false); err != nil {
		t.Fatalf("LoadOrBuild() error = %v", err)
	}

	target := filepath.Join(moduleRoot, "pkg", "beta", "beta.go")
	orig, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.WriteFile(target, orig, 0o644)
	})

	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(target, append(orig, '\n'), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadOrBuild(ctx, moduleRoot, cfg, true); err != nil {
		t.Fatalf("LoadOrBuild(noCache) error = %v", err)
	}
}

func BenchmarkLoadOrBuildCached(b *testing.B) {
	root, err := filepath.Abs(mustMinimoduleRootPath(b))
	if err != nil {
		b.Fatalf("Abs() error = %v", err)
	}
	cacheDir := filepath.Join(b.TempDir(), ".rippl", "cache")
	cfg := config.DefaultConfig()
	cfg.Cache.Dir = cacheDir
	ctx := context.Background()

	if _, err := LoadOrBuild(ctx, root, cfg, false); err != nil {
		b.Fatalf("warm cache: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := LoadOrBuild(ctx, root, cfg, false); err != nil {
			b.Fatalf("LoadOrBuild() error = %v", err)
		}
	}
}
