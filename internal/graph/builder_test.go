package graph

import (
	"context"
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
	// Restore mtime so the on-disk cache entry remains valid while content changed.
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
