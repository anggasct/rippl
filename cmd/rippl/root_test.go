package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anggasct/rippl/internal/config"
)

func TestRootHelpListsSubcommands(t *testing.T) {
	t.Parallel()

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	help := out.String()
	for _, sub := range []string{"analyze", "score", "test", "graph"} {
		if !strings.Contains(help, sub) {
			t.Fatalf("help output missing subcommand %q", sub)
		}
	}
}

func TestAnalyzeOutsideModuleReturnsExitError2(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "foo.go")
	if err := os.WriteFile(target, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetErr(os.Stderr)
	cmd.SetOut(os.Stdout)
	cmd.SetArgs([]string{"analyze", target})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want module error")
	}

	var exitErr *config.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Execute() error = %T(%v), want *config.ExitError", err, err)
	}
	if exitErr.Code != 2 {
		t.Fatalf("exit code = %d, want 2", exitErr.Code)
	}
}

func TestVersionDoesNotRequireModule(t *testing.T) {
	dir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if strings.TrimSpace(out.String()) != "dev" {
		t.Fatalf("version output = %q, want dev", out.String())
	}
}

func TestGraphInModuleCreatesCacheDir(t *testing.T) {
	moduleRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(moduleRoot, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})
	if err := os.Chdir(moduleRoot); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"graph"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	cacheDir := filepath.Join(moduleRoot, ".rippl", "cache")
	if _, err := os.Stat(cacheDir); err != nil {
		t.Fatalf("cache dir stat error = %v", err)
	}
}

func TestAnalyzeInvalidFileReturnsExitError2(t *testing.T) {
	t.Parallel()

	cmd := newRootCmd()
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"analyze", "/nonexistent/path/foo.go"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want exit error")
	}

	var exitErr *config.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Execute() error = %T(%v), want *config.ExitError", err, err)
	}
	if exitErr.Code != 2 {
		t.Fatalf("exit code = %d, want 2", exitErr.Code)
	}
}

func TestAnalyzeInModuleWithFile(t *testing.T) {
	moduleRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	srcFile := filepath.Join(moduleRoot, "main.go")
	if err := os.WriteFile(srcFile, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"analyze", "--format", "text", srcFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Source:") {
		t.Fatalf("output missing 'Source:' header, got: %q", output)
	}
}

func TestAnalyzeDirectoryReturnsExitError2(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	cmd := newRootCmd()
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"analyze", dir})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want exit error")
	}

	var exitErr *config.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Execute() error = %T(%v), want *config.ExitError", err, err)
	}
	if exitErr.Code != 2 {
		t.Fatalf("exit code = %d, want 2", exitErr.Code)
	}
}
