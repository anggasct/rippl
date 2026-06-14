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
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/packages"
)

func TestTestCmdMissingFileArg(t *testing.T) {
	t.Parallel()

	cmd := newRootCmd()
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error for missing arg")
	}
}

func TestTestCmdFileNotFound(t *testing.T) {
	t.Parallel()

	cmd := newRootCmd()
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"test", "/nonexistent/path/foo.go"})

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

func TestTestCmdOutsideModule(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "foo.go")
	if err := os.WriteFile(target, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"test", target})

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

func TestTestCmdNoAffectedTests(t *testing.T) {
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
	cmd.SetArgs([]string{"test", srcFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "no affected packages with tests") {
		t.Fatalf("output = %q, want 'no affected packages with tests'", output)
	}
}

func TestTestCmdWithAffectedTests(t *testing.T) {
	moduleRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a package with a source file and a test file
	pkgDir := filepath.Join(moduleRoot, "pkg", "alpha")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "alpha.go"), []byte("package alpha\n\nfunc Hello() string { return \"hello\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "alpha_test.go"), []byte("package alpha\n\nimport \"testing\"\n\nfunc TestHello(t *testing.T) { if Hello() != \"hello\" { t.Fail() } }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a second package that imports alpha (affected)
	pkgBetaDir := filepath.Join(moduleRoot, "pkg", "beta")
	if err := os.MkdirAll(pkgBetaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgBetaDir, "beta.go"), []byte("package beta\n\nimport \"example.com/test/pkg/alpha\"\n\nfunc Greet() string { return alpha.Hello() }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgBetaDir, "beta_test.go"), []byte("package beta\n\nimport \"testing\"\n\nfunc TestGreet(t *testing.T) { if Greet() != \"hello\" { t.Fail() } }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"test", filepath.Join(pkgDir, "alpha.go")})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "packages run:") {
		t.Fatalf("output = %q, want 'packages run:' summary", output)
	}
}

func TestResolveAffectedPackagesEmptyGraph(t *testing.T) {
	t.Parallel()

	// With an empty graph, all packages are skipped (no test files found)
	g := graph.Build(nil)

	result := &graph.ImpactResult{
		Source:   graph.AffectedFile{Path: "pkg/alpha/alpha.go"},
		Affected: []graph.AffectedFile{},
	}

	pkgs, skipped := packages.AffectedWithTests(g, result)
	if len(pkgs) != 0 {
		t.Fatalf("packages = %v, want empty", pkgs)
	}
	if skipped != 1 {
		t.Fatalf("skipped = %d, want 1 (no test files in empty graph)", skipped)
	}
}

func TestTestCmdReportsRunAndSkipped(t *testing.T) {
	moduleRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Package with tests
	pkgAlpha := filepath.Join(moduleRoot, "alpha")
	if err := os.MkdirAll(pkgAlpha, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgAlpha, "alpha.go"), []byte("package alpha\n\nfunc A() int { return 1 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgAlpha, "alpha_test.go"), []byte("package alpha\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) { if A() != 1 { t.Fail() } }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Package without tests (will be affected but skipped)
	pkgBeta := filepath.Join(moduleRoot, "beta")
	if err := os.MkdirAll(pkgBeta, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgBeta, "beta.go"), []byte("package beta\n\nimport \"example.com/test/alpha\"\n\nfunc B() int { return alpha.A() }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"test", filepath.Join(pkgAlpha, "alpha.go")})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "packages run:") {
		t.Fatalf("output missing 'packages run:' summary: %q", output)
	}
	if !strings.Contains(output, "skipped:") {
		t.Fatalf("output missing 'skipped:' count: %q", output)
	}
}

func TestTestJSONNoAffected(t *testing.T) {
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
	cmd.SetArgs([]string{"test", "--format", "json", srcFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !json.Valid(out.Bytes()) {
		t.Fatalf("stdout is not valid JSON: %q", out.String())
	}

	var doc struct {
		Command     string   `json:"command"`
		PackagesRun []string `json:"packages_run"`
		Passed      bool     `json:"passed"`
		ExitCode    int      `json:"exit_code"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if doc.Command != "test" {
		t.Fatalf("command = %q, want test", doc.Command)
	}
	if len(doc.PackagesRun) != 0 {
		t.Fatalf("packages_run = %v, want empty", doc.PackagesRun)
	}
	if !doc.Passed || doc.ExitCode != 0 {
		t.Fatalf("passed=%v exit_code=%d, want passed=true exit_code=0", doc.Passed, doc.ExitCode)
	}
}

func TestTestJSONWithAffected(t *testing.T) {
	moduleRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pkgDir := filepath.Join(moduleRoot, "pkg", "alpha")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "alpha.go"), []byte("package alpha\n\nfunc Hello() string { return \"hello\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "alpha_test.go"), []byte("package alpha\n\nimport \"testing\"\n\nfunc TestHello(t *testing.T) { if Hello() != \"hello\" { t.Fail() } }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pkgBetaDir := filepath.Join(moduleRoot, "pkg", "beta")
	if err := os.MkdirAll(pkgBetaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgBetaDir, "beta.go"), []byte("package beta\n\nimport \"example.com/test/pkg/alpha\"\n\nfunc Greet() string { return alpha.Hello() }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgBetaDir, "beta_test.go"), []byte("package beta\n\nimport \"testing\"\n\nfunc TestGreet(t *testing.T) { if Greet() != \"hello\" { t.Fail() } }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"test", "--format", "json", filepath.Join(pkgDir, "alpha.go")})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !json.Valid(out.Bytes()) {
		t.Fatalf("stdout is not valid JSON: %q", out.String())
	}

	var doc struct {
		Passed      bool     `json:"passed"`
		PackagesRun []string `json:"packages_run"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !doc.Passed {
		t.Fatal("passed = false, want true")
	}
	if len(doc.PackagesRun) == 0 {
		t.Fatal("packages_run empty, want at least one package")
	}

	stderr := errOut.String()
	if !strings.Contains(stderr, "=== RUN") {
		t.Fatalf("stderr missing progress lines: %q", stderr)
	}
}

func TestTestJSONModulePath(t *testing.T) {
	moduleRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte("module example.com/testjson\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pkgDir := filepath.Join(moduleRoot, "pkg", "alpha")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "alpha.go"), []byte("package alpha\n\nfunc A() int { return 1 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "alpha_test.go"), []byte("package alpha\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) { if A() != 1 { t.Fail() } }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"test", "--format", "json", filepath.Join(pkgDir, "alpha.go")})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var doc struct {
		Module string `json:"module"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if doc.Module != "example.com/testjson" {
		t.Fatalf("module = %q, want example.com/testjson", doc.Module)
	}
	if strings.HasPrefix(doc.Module, "/") {
		t.Fatalf("module = %q, want go.mod path not filesystem path", doc.Module)
	}
}

func TestTestJSONFailure(t *testing.T) {
	moduleRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(moduleRoot, "go.mod"), []byte("module example.com/fail\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pkgDir := filepath.Join(moduleRoot, "pkg", "alpha")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "alpha.go"), []byte("package alpha\n\nfunc A() int { return 1 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "alpha_test.go"), []byte("package alpha\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) { t.Fatal(\"boom\") }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(os.Stderr)
	cmd.SetArgs([]string{"test", "--format", "json", filepath.Join(pkgDir, "alpha.go")})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want test failure exit")
	}
	var exitErr *config.ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("Execute() error = %v, want ExitError code 1", err)
	}

	var doc struct {
		Passed   bool `json:"passed"`
		ExitCode int  `json:"exit_code"`
		Failures []struct {
			Package string `json:"package"`
			Summary string `json:"summary"`
		} `json:"failures"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, raw: %q", err, out.String())
	}
	if doc.Passed || doc.ExitCode != 1 {
		t.Fatalf("passed=%v exit_code=%d, want passed=false exit_code=1", doc.Passed, doc.ExitCode)
	}
	if len(doc.Failures) == 0 {
		t.Fatal("failures empty, want at least one entry")
	}
	if doc.Failures[0].Package != "./pkg/alpha" {
		t.Fatalf("failure package = %q, want ./pkg/alpha", doc.Failures[0].Package)
	}
	if doc.Failures[0].Summary == "" {
		t.Fatal("failure summary empty")
	}
}
