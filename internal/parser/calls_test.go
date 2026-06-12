package parser

import (
	"context"
	"testing"
)

func TestExtractCallsAndTypeRefs(t *testing.T) {
	t.Parallel()

	analyses, err := ParseModule(context.Background(), minimoduleRoot(t), defaultTestConfig())
	if err != nil {
		t.Fatalf("ParseModule() error = %v", err)
	}

	alpha := analysisByPath(t, analyses)["pkg/alpha/alpha.go"]
	if !symbolNamed(alpha.Calls, EdgeCall, "Foo") {
		t.Fatalf("alpha calls = %#v, want Foo call edge", alpha.Calls)
	}
	if !containsPath(edgeTargets(alpha.Calls, EdgeCall), "pkg/beta/beta.go") {
		t.Fatalf("alpha call targets = %#v, want pkg/beta/beta.go", alpha.Calls)
	}
	if !symbolNamed(alpha.TypeRefs, EdgeTypeRef, "Type") {
		t.Fatalf("alpha type refs = %#v, want Type type_ref edge", alpha.TypeRefs)
	}
}
