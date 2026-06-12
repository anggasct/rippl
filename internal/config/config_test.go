package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigMatchesPRD(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	if cfg.Version != 1 {
		t.Fatalf("Version = %d, want 1", cfg.Version)
	}
	if len(cfg.Languages) != 1 || cfg.Languages[0] != "go" {
		t.Fatalf("Languages = %#v, want [go]", cfg.Languages)
	}
	if cfg.Impact.MaxDepth != 3 {
		t.Fatalf("Impact.MaxDepth = %d, want 3", cfg.Impact.MaxDepth)
	}
	if cfg.Risk.Since != "12 months" {
		t.Fatalf("Risk.Since = %q, want 12 months", cfg.Risk.Since)
	}
	if cfg.Cache.Dir != ".rippl/cache" {
		t.Fatalf("Cache.Dir = %q, want .rippl/cache", cfg.Cache.Dir)
	}
}

func TestLoadMissingConfigUsesDefaults(t *testing.T) {
	t.Parallel()

	moduleRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, _, err := Load(moduleRoot, ".rippl.yaml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	defaults := DefaultConfig()
	if cfg.Version != defaults.Version || cfg.Impact.MaxDepth != defaults.Impact.MaxDepth {
		t.Fatalf("Load() = %#v, want defaults", cfg)
	}
}

func TestLoadMissingConfigSecondReturnFalse(t *testing.T) {
	t.Parallel()

	moduleRoot := t.TempDir()
	_, loaded, err := Load(moduleRoot, ".rippl.yaml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded {
		t.Fatal("loaded = true, want false")
	}
}

func TestLoadCustomConfigPath(t *testing.T) {
	t.Parallel()

	moduleRoot := t.TempDir()
	configPath := filepath.Join(moduleRoot, "custom.yaml")
	if err := os.WriteFile(configPath, []byte(`impact:
  max_depth: 7
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, loaded, err := Load(moduleRoot, "custom.yaml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !loaded {
		t.Fatal("loaded = false, want true")
	}
	if cfg.Impact.MaxDepth != 7 {
		t.Fatalf("Impact.MaxDepth = %d, want 7", cfg.Impact.MaxDepth)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	t.Parallel()

	moduleRoot := t.TempDir()
	configPath := filepath.Join(moduleRoot, ".rippl.yaml")
	if err := os.WriteFile(configPath, []byte("impact:\n  max_depth: not-a-number\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := Load(moduleRoot, ".rippl.yaml")
	if err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
}
