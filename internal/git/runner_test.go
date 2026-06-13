package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExecRunnerInvokesGit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runner := ExecRunner{}

	out, err := runner.Run(ctx, ".", "rev-parse", "--is-inside-work-tree")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if strings.TrimSpace(string(out)) != "true" {
		t.Fatalf("is-inside-work-tree = %q, want true", out)
	}
}

func TestMockRunnerRecordsCalls(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mock := &MockRunner{
		Responses: map[string][]byte{
			"rev-parse --is-inside-work-tree": []byte("true\n"),
		},
	}

	out, err := mock.Run(ctx, "/tmp/repo", "rev-parse", "--is-inside-work-tree")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if strings.TrimSpace(string(out)) != "true" {
		t.Fatalf("output = %q, want true", out)
	}
	if len(mock.Calls) != 1 {
		t.Fatalf("Calls len = %d, want 1", len(mock.Calls))
	}
}

func TestIsRepo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	if ok, err := IsRepo(ctx, ".", ExecRunner{}); err != nil || !ok {
		t.Fatalf("IsRepo(.) = (%v, %v), want (true, nil)", ok, err)
	}

	dir := t.TempDir()
	ok, err := IsRepo(ctx, dir, ExecRunner{})
	if err != nil {
		t.Fatalf("IsRepo(temp) error = %v", err)
	}
	if ok {
		t.Fatal("IsRepo(temp) = true, want false")
	}
}

func TestIsRepoMockFalse(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mock := &MockRunner{
		Responses: map[string][]byte{},
	}
	ok, err := IsRepo(ctx, "/tmp", mock)
	if err != nil {
		t.Fatalf("IsRepo() error = %v", err)
	}
	if ok {
		t.Fatal("IsRepo() = true, want false on git error")
	}
}

func TestGitSince(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{"12 months", "12 months ago"},
		{"12 months ago", "12 months ago"},
		{"", "12 months ago"},
		{" 6 weeks ", "6 weeks ago"},
	}
	for _, tc := range tests {
		if got := gitSince(tc.in); got != tc.want {
			t.Errorf("gitSince(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestLogArgsIncludeSinceAndPath(t *testing.T) {
	t.Parallel()
	args := logArgs("12 months", "pkg/foo.go")
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--since=12 months ago") {
		t.Fatalf("args = %v, missing normalized since", args)
	}
	if !strings.Contains(joined, "pkg/foo.go") {
		t.Fatalf("args = %v, missing path", args)
	}
}

func TestParseCommitLog(t *testing.T) {
	t.Parallel()
	out, err := os.ReadFile(filepath.Join("testdata", "log_alpha.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	commits := parseCommitLog(out)
	if len(commits) != 3 {
		t.Fatalf("len(commits) = %d, want 3", len(commits))
	}
	if commits[0].author != "Alice" || commits[0].subject != "add feature" {
		t.Fatalf("first commit = %+v", commits[0])
	}
	if commits[1].subject != "fix crash in handler" {
		t.Fatalf("second subject = %q", commits[1].subject)
	}
	want := time.Unix(1712000000, 0).UTC()
	if !commits[2].when.Equal(want) {
		t.Fatalf("third when = %v, want %v", commits[2].when, want)
	}
}

func TestParseNumstatChurn(t *testing.T) {
	t.Parallel()
	out, err := os.ReadFile(filepath.Join("testdata", "numstat_alpha.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := parseNumstatChurn(out); got != 27 {
		t.Fatalf("churn = %d, want 27", got)
	}
}

func TestIsBugFix(t *testing.T) {
	t.Parallel()
	patterns, err := compileBugFixPatterns([]string{`\bfix(ed|es)?\b`, `\bhotfix\b`})
	if err != nil {
		t.Fatalf("compileBugFixPatterns() error = %v", err)
	}
	if !isBugFix("fix crash", "", patterns) {
		t.Fatal("expected bug-fix match on subject")
	}
	if !isBugFix("update", "hotfix for prod", patterns) {
		t.Fatal("expected bug-fix match on body")
	}
	if isBugFix("refactor imports", "", patterns) {
		t.Fatal("did not expect bug-fix match")
	}
}
