package parser

import (
	"context"
	"testing"
)

func TestParseModuleIntegration(t *testing.T) {
	t.Parallel()

	analyses, err := ParseModule(context.Background(), minimoduleRoot(t), defaultTestConfig())
	if err != nil {
		t.Fatalf("ParseModule() error = %v", err)
	}

	if len(analyses) != 6 {
		t.Fatalf("len(analyses) = %d, want 6", len(analyses))
	}

	for _, analysis := range analyses {
		if analysis.Path == "" || analysis.Package == "" {
			t.Fatalf("analysis missing path/package: %#v", analysis)
		}
	}
}

func TestParseModuleNilConfigUsesDefaults(t *testing.T) {
	t.Parallel()

	_, err := ParseModule(context.Background(), minimoduleRoot(t), nil)
	if err != nil {
		t.Fatalf("ParseModule() error = %v", err)
	}
}
