package main

import (
	"bytes"
	"encoding/json"
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

func TestScoreJSONModulePath(t *testing.T) {
	moduleRoot := minimoduleRoot(t)
	srcFile := filepath.Join(moduleRoot, "pkg", "alpha", "alpha.go")

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"score", "--format", "json", "--no-cache", srcFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var doc struct {
		Module  string            `json:"module"`
		Signals []json.RawMessage `json:"signals"`
		Command string            `json:"command"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, raw: %q", err, out.String())
	}
	if doc.Module != "example.com/minimodule" {
		t.Fatalf("module = %q, want %q", doc.Module, "example.com/minimodule")
	}
	if strings.HasPrefix(doc.Module, "/") {
		t.Fatalf("module = %q, want go.mod path not filesystem path", doc.Module)
	}
	if doc.Command != "score" {
		t.Fatalf("command = %q, want score", doc.Command)
	}
	if len(doc.Signals) != 6 {
		t.Fatalf("signals length = %d, want 6", len(doc.Signals))
	}
	if strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("output contains ANSI escape sequences")
	}
}

func TestAnalyzeJSONModulePath(t *testing.T) {
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
	cmd.SetArgs([]string{"analyze", "--format", "json", srcFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var doc struct {
		Module string `json:"module"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if doc.Module != "example.com/test" {
		t.Fatalf("module = %q, want %q", doc.Module, "example.com/test")
	}
	if strings.HasPrefix(doc.Module, "/") {
		t.Fatalf("module = %q, want go.mod path not filesystem path", doc.Module)
	}
}

func TestAnalyzeInModuleWithRelativeFile(t *testing.T) {
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
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"analyze", "--format", "text", "main.go"})

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

func TestAnalyzeJSONSuggestedActions(t *testing.T) {
	t.Parallel()

	moduleRoot := minimoduleRoot(t)
	srcFile := filepath.Join(moduleRoot, "pkg", "gamma", "gamma.go")

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"analyze", "--format", "json", "--no-cache", srcFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var doc struct {
		Version          string `json:"version"`
		SuggestedActions struct {
			PackagesToTest   []string `json:"packages_to_test"`
			Commands         []string `json:"commands"`
			UntestedHighRisk []struct {
				Path string `json:"path"`
			} `json:"untested_high_risk"`
		} `json:"suggested_actions"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if doc.Version != "1.1" {
		t.Fatalf("version = %q, want 1.1", doc.Version)
	}
	if len(doc.SuggestedActions.PackagesToTest) == 0 {
		t.Fatal("suggested_actions.packages_to_test is empty")
	}
	if len(doc.SuggestedActions.Commands) < 2 {
		t.Fatalf("commands = %v, want at least 2 entries", doc.SuggestedActions.Commands)
	}
	if !strings.HasPrefix(doc.SuggestedActions.Commands[0], "go test ") {
		t.Fatalf("commands[0] = %q, want go test prefix", doc.SuggestedActions.Commands[0])
	}
}

func TestAnalyzeFilterTop(t *testing.T) {
	t.Parallel()

	moduleRoot := minimoduleRoot(t)
	srcFile := filepath.Join(moduleRoot, "pkg", "gamma", "gamma.go")

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"analyze", "--format", "json", "--top", "1", "--no-cache", srcFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var doc struct {
		Affected []json.RawMessage `json:"affected"`
		Summary  struct {
			TotalAffectedCount int `json:"total_affected_count"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(doc.Affected) != 1 {
		t.Fatalf("affected len = %d, want 1", len(doc.Affected))
	}
	if doc.Summary.TotalAffectedCount < 2 {
		t.Fatalf("total_affected_count = %d, want > 1", doc.Summary.TotalAffectedCount)
	}
}

func TestAnalyzeFilterMinRisk(t *testing.T) {
	t.Parallel()

	moduleRoot := minimoduleRoot(t)
	srcFile := filepath.Join(moduleRoot, "pkg", "gamma", "gamma.go")

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"analyze", "--format", "json", "--min-risk", "100", "--no-cache", srcFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var doc struct {
		Affected []json.RawMessage `json:"affected"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(doc.Affected) != 0 {
		t.Fatalf("affected len = %d, want 0 for min-risk 100", len(doc.Affected))
	}
}

func minimoduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "internal", "parser", "testdata", "minimodule", "go.mod")
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Dir(candidate)
		}
		if filepath.Dir(dir) == dir {
			t.Fatal("minimodule testdata not found")
		}
	}
}
