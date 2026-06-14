package render

import "testing"

func TestCompactAnalyzeOutput(t *testing.T) {
	t.Parallel()

	out := Output{
		Files: []FileOutput{
			{Path: "a.go", Chain: []string{"src.go", "a.go"}},
		},
	}
	compact := CompactAnalyzeOutput(out)
	if len(compact.Files[0].Chain) != 0 {
		t.Fatalf("Chain = %v, want nil/empty", compact.Files[0].Chain)
	}
}

func TestIsStructuredFormat(t *testing.T) {
	t.Parallel()

	if !IsStructuredFormat("json") || !IsStructuredFormat("agent") {
		t.Fatal("expected json and agent to be structured formats")
	}
	if IsStructuredFormat("text") {
		t.Fatal("text should not be structured")
	}
}
