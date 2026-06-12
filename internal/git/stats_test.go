package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anggasct/rippl/internal/config"
)

func TestCollectFileStatsNonRepoReturnsZeros(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := t.TempDir()
	files := []string{"pkg/a.go", "pkg/b.go"}

	stats, err := CollectFileStats(ctx, dir, files, config.DefaultConfig())
	if err != nil {
		t.Fatalf("CollectFileStats() error = %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("len(stats) = %d, want 2", len(stats))
	}
	for _, path := range files {
		s := stats[path]
		if s.Path != path || s.CommitCount != 0 || s.BugFixCount != 0 || s.AuthorCount != 0 || !s.LastModified.IsZero() || s.Churn != 0 {
			t.Fatalf("stats[%q] = %+v, want zero values", path, s)
		}
	}
}

func TestCollectFileStatsMockFixture(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logFixture, err := os.ReadFile(filepath.Join("testdata", "log_alpha.txt"))
	if err != nil {
		t.Fatalf("ReadFile(log) error = %v", err)
	}
	numFixture, err := os.ReadFile(filepath.Join("testdata", "numstat_alpha.txt"))
	if err != nil {
		t.Fatalf("ReadFile(numstat) error = %v", err)
	}

	path := "pkg/alpha/alpha.go"
	logKey := strings.Join(logArgs("12 months", path), " ")
	numKey := strings.Join(numstatArgs("12 months", path), " ")

	mock := &MockRunner{
		Responses: map[string][]byte{
			"rev-parse --is-inside-work-tree": []byte("true\n"),
			logKey:                            logFixture,
			numKey:                            numFixture,
		},
	}

	stats, err := collectFileStats(ctx, "/repo", []string{path}, config.DefaultConfig(), mock)
	if err != nil {
		t.Fatalf("collectFileStats() error = %v", err)
	}
	got := stats[path]
	if got.CommitCount != 3 {
		t.Fatalf("CommitCount = %d, want 3", got.CommitCount)
	}
	if got.AuthorCount != 2 {
		t.Fatalf("AuthorCount = %d, want 2", got.AuthorCount)
	}
	if got.BugFixCount != 1 {
		t.Fatalf("BugFixCount = %d, want 1", got.BugFixCount)
	}
	if got.Churn != 27 {
		t.Fatalf("Churn = %d, want 27", got.Churn)
	}
	if got.LastModified.Unix() != 1712000000 {
		t.Fatalf("LastModified = %v, want unix 1712000000", got.LastModified)
	}
}

func TestCollectFileStatsNilConfigUsesDefaults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := t.TempDir()
	stats, err := collectFileStats(ctx, dir, []string{"x.go"}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("collectFileStats() error = %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("len(stats) = %d, want 1", len(stats))
	}
}

func TestCollectFileStatsContextCancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	path := "pkg/a.go"
	logKey := strings.Join(logArgs("12 months", path), " ")
	mock := &MockRunner{
		Responses: map[string][]byte{
			"rev-parse --is-inside-work-tree": []byte("true\n"),
			logKey:                            []byte(""),
		},
	}
	_, err := collectFileStats(ctx, "/repo", []string{path}, config.DefaultConfig(), mock)
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

func TestCollectFileStatsInvalidPattern(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.Risk.BugFixPatterns = []string{"("}

	mock := &MockRunner{
		Responses: map[string][]byte{
			"rev-parse --is-inside-work-tree": []byte("true\n"),
		},
	}
	_, err := collectFileStats(ctx, "/repo", []string{"a.go"}, cfg, mock)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}
