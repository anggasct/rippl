package graph

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheSaveLoadAndValidate(t *testing.T) {
	t.Parallel()

	moduleRoot := minimoduleRoot(t)
	cacheDir := t.TempDir()

	g := Build(parseMinimodule(t))
	mtimes, err := CollectMTimes(moduleRoot, g.Files())
	if err != nil {
		t.Fatalf("CollectMTimes() error = %v", err)
	}

	if err := Save(moduleRoot, cacheDir, g, mtimes); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, loadedMTimes, err := Load(moduleRoot, cacheDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	current, err := CollectMTimes(moduleRoot, loaded.Files())
	if err != nil {
		t.Fatalf("CollectMTimes() error = %v", err)
	}
	if !IsValid(loadedMTimes, current) {
		t.Fatal("expected cached mtimes to be valid")
	}
	if len(loaded.Files()) != len(g.Files()) {
		t.Fatalf("loaded files = %#v, want %#v", loaded.Files(), g.Files())
	}
}

func TestIsValidDetectsNewerMtime(t *testing.T) {
	t.Parallel()

	cached := map[string]int64{"pkg/a.go": time.Now().Add(-time.Hour).UnixNano()}
	current := map[string]int64{"pkg/a.go": time.Now().UnixNano()}
	if IsValid(cached, current) {
		t.Fatal("IsValid() = true, want false for newer mtime")
	}
}

func TestCacheInvalidatesAfterFileTouch(t *testing.T) {
	moduleRoot := minimoduleRoot(t)
	target := filepath.Join(moduleRoot, "pkg", "beta", "beta.go")

	orig, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.WriteFile(target, orig, 0o644)
	})

	cacheDir := t.TempDir()
	g := Build(parseMinimodule(t))
	mtimes, err := CollectMTimes(moduleRoot, g.Files())
	if err != nil {
		t.Fatalf("CollectMTimes() error = %v", err)
	}
	if err := Save(moduleRoot, cacheDir, g, mtimes); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(target, append(orig, '\n'), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	loaded, loadedMTimes, err := Load(moduleRoot, cacheDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	current, err := CollectMTimes(moduleRoot, loaded.Files())
	if err != nil {
		t.Fatalf("CollectMTimes() error = %v", err)
	}
	if IsValid(loadedMTimes, current) {
		t.Fatal("IsValid() = true, want false after file touch")
	}
}
