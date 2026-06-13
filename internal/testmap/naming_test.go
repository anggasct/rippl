package testmap

import "testing"

func TestSourcePathForTest(t *testing.T) {
	t.Parallel()
	got, ok := sourcePathForTest("pkg/beta/beta_test.go")
	if !ok || got != "pkg/beta/beta.go" {
		t.Fatalf("sourcePathForTest() = (%q, %v), want (pkg/beta/beta.go, true)", got, ok)
	}
	if _, ok := sourcePathForTest("pkg/beta/beta.go"); ok {
		t.Fatal("expected false for non-test path")
	}
}

func TestIsTestFile(t *testing.T) {
	t.Parallel()
	if !isTestFile("a_test.go") {
		t.Fatal("expected test file")
	}
	if isTestFile("a.go") {
		t.Fatal("did not expect test file")
	}
}
