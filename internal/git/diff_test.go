package git

import (
	"context"
	"testing"
)

func TestChangedGoFiles(t *testing.T) {
	t.Parallel()

	runner := &MockRunner{
		Responses: map[string][]byte{
			"rev-parse --is-inside-work-tree": []byte("true\n"),
			"diff --name-only --diff-filter=ACMR HEAD": []byte("pkg/alpha/alpha.go\npkg/beta/beta.go\nREADME.md\n"),
		},
	}

	got, err := ChangedGoFiles(context.Background(), "/mod", "HEAD", runner)
	if err != nil {
		t.Fatalf("ChangedGoFiles() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ChangedGoFiles() = %v, want 2 go files", got)
	}
	if got[0] != "pkg/alpha/alpha.go" || got[1] != "pkg/beta/beta.go" {
		t.Fatalf("ChangedGoFiles() = %v, unexpected paths", got)
	}
}

func TestChangedGoFilesNotRepo(t *testing.T) {
	t.Parallel()

	runner := &MockRunner{
		Responses: map[string][]byte{
			"rev-parse --is-inside-work-tree": []byte("false\n"),
		},
	}

	_, err := ChangedGoFiles(context.Background(), "/mod", "HEAD", runner)
	if err != ErrNotRepository {
		t.Fatalf("ChangedGoFiles() error = %v, want ErrNotRepository", err)
	}
}
