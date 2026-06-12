package graph

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/anggasct/rippl/internal/config"
)

func TestLoadOrBuildUsesCache(t *testing.T) {
	moduleRoot := minimoduleRoot(t)
	cacheDir := filepath.Join(t.TempDir(), ".rippl", "cache")
	cfg := config.DefaultConfig()
	cfg.Cache.Dir = cacheDir
	ctx := context.Background()

	if _, err := LoadOrBuild(ctx, moduleRoot, cfg, false); err != nil {
		t.Fatalf("warm cache: %v", err)
	}

	gomod := filepath.Join(moduleRoot, "go.mod")
	backup := gomod + ".bak"
	if err := os.Rename(gomod, backup); err != nil {
		t.Fatalf("Rename(go.mod) error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Rename(backup, gomod)
	})

	if _, err := LoadOrBuild(ctx, moduleRoot, cfg, false); err != nil {
		t.Fatalf("LoadOrBuild(cache hit) error = %v, want success without go.mod", err)
	}
}

func TestLoadOrBuildRespectsContextOnCacheHit(t *testing.T) {
	moduleRoot := minimoduleRoot(t)
	cacheDir := filepath.Join(t.TempDir(), ".rippl", "cache")
	cfg := config.DefaultConfig()
	cfg.Cache.Dir = cacheDir
	ctx := context.Background()

	if _, err := LoadOrBuild(ctx, moduleRoot, cfg, false); err != nil {
		t.Fatalf("warm cache: %v", err)
	}

	cancelled, cancel := context.WithCancel(ctx)
	cancel()

	if _, err := LoadOrBuild(cancelled, moduleRoot, cfg, false); !errors.Is(err, context.Canceled) {
		t.Fatalf("LoadOrBuild() error = %v, want context.Canceled", err)
	}
}

func TestLoadOrBuildNoCacheBypassesRead(t *testing.T) {
	moduleRoot := minimoduleRoot(t)
	cacheDir := filepath.Join(t.TempDir(), ".rippl", "cache")
	cfg := config.DefaultConfig()
	cfg.Cache.Dir = cacheDir
	ctx := context.Background()

	if _, err := LoadOrBuild(ctx, moduleRoot, cfg, false); err != nil {
		t.Fatalf("warm cache: %v", err)
	}

	target := filepath.Join(moduleRoot, "pkg", "beta", "beta.go")
	orig, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.WriteFile(target, orig, 0o644)
	})

	modified := append(append(orig, '\n'), []byte("func CachedStale() {}\n")...)
	if err := os.WriteFile(target, modified, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Chtimes(target, info.ModTime(), info.ModTime()); err != nil {
		t.Fatalf("Chtimes() error = %v", err)
	}

	cached, err := LoadOrBuild(ctx, moduleRoot, cfg, false)
	if err != nil {
		t.Fatalf("LoadOrBuild(cached) error = %v", err)
	}
	betaCached, ok := cached.Node("pkg/beta/beta.go")
	if !ok {
		t.Fatal("missing beta node in cached graph")
	}
	if hasExport(betaCached.Exports, "CachedStale", "func") {
		t.Fatalf("cached graph should not include post-touch export; exports=%#v", betaCached.Exports)
	}

	fresh, err := LoadOrBuild(ctx, moduleRoot, cfg, true)
	if err != nil {
		t.Fatalf("LoadOrBuild(noCache) error = %v", err)
	}
	betaFresh, ok := fresh.Node("pkg/beta/beta.go")
	if !ok {
		t.Fatal("missing beta node in fresh graph")
	}
	if !hasExport(betaFresh.Exports, "CachedStale", "func") {
		t.Fatalf("noCache graph should include post-touch export; exports=%#v", betaFresh.Exports)
	}
}

func TestLoadOrBuildNoCacheSkipsWrite(t *testing.T) {
	moduleRoot := minimoduleRoot(t)
	cacheDir := filepath.Join(t.TempDir(), ".rippl", "cache")
	cfg := config.DefaultConfig()
	cfg.Cache.Dir = cacheDir
	ctx := context.Background()

	if _, err := LoadOrBuild(ctx, moduleRoot, cfg, false); err != nil {
		t.Fatalf("warm cache: %v", err)
	}

	cacheFile, err := cachePath(moduleRoot, cacheDir)
	if err != nil {
		t.Fatalf("cachePath() error = %v", err)
	}
	before, err := os.Stat(cacheFile)
	if err != nil {
		t.Fatalf("Stat(cache) error = %v", err)
	}

	if _, err := LoadOrBuild(ctx, moduleRoot, cfg, true); err != nil {
		t.Fatalf("LoadOrBuild(noCache) error = %v", err)
	}

	after, err := os.Stat(cacheFile)
	if err != nil {
		t.Fatalf("Stat(cache) after noCache error = %v", err)
	}
	if !before.ModTime().Equal(after.ModTime()) {
		t.Fatalf("cache modtime changed: before=%v after=%v", before.ModTime(), after.ModTime())
	}
}

func BenchmarkLoadOrBuild(b *testing.B) {
	root, err := filepath.Abs(mustMinimoduleRootPath(b))
	if err != nil {
		b.Fatalf("Abs() error = %v", err)
	}
	cfg := config.DefaultConfig()
	ctx := context.Background()

	b.Run("cold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			cacheDir := filepath.Join(b.TempDir(), ".rippl", "cache")
			cfg.Cache.Dir = cacheDir
			b.StartTimer()

			if _, err := LoadOrBuild(ctx, root, cfg, true); err != nil {
				b.Fatalf("LoadOrBuild(cold) error = %v", err)
			}
		}
	})

	b.Run("cached", func(b *testing.B) {
		cacheDir := filepath.Join(b.TempDir(), ".rippl", "cache")
		cfg.Cache.Dir = cacheDir
		if _, err := LoadOrBuild(ctx, root, cfg, false); err != nil {
			b.Fatalf("warm cache: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := LoadOrBuild(ctx, root, cfg, false); err != nil {
				b.Fatalf("LoadOrBuild(cached) error = %v", err)
			}
		}
	})
}
