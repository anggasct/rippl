package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindModuleRoot(t *testing.T) {
	t.Parallel()

	moduleRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	subdir := filepath.Join(moduleRoot, "internal", "pkg")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := FindModuleRoot(subdir)
	if err != nil {
		t.Fatalf("FindModuleRoot() error = %v", err)
	}
	if got != moduleRoot {
		t.Fatalf("FindModuleRoot() = %q, want %q", got, moduleRoot)
	}
}

func TestFindModuleRootFromPath(t *testing.T) {
	t.Parallel()

	moduleRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(moduleRoot, "cmd", "app", "main.go")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := FindModuleRootFromPath(target)
	if err != nil {
		t.Fatalf("FindModuleRootFromPath() error = %v", err)
	}
	if got != moduleRoot {
		t.Fatalf("FindModuleRootFromPath() = %q, want %q", got, moduleRoot)
	}
}

func TestFindModuleRootNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := FindModuleRoot(dir)
	if err != ErrNotGoModule {
		t.Fatalf("FindModuleRoot() error = %v, want ErrNotGoModule", err)
	}
}
