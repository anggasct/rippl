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

	byPath := analysisByPath(t, analyses)
	alpha := byPath["pkg/alpha/alpha.go"]
	beta := byPath["pkg/beta/beta.go"]

	if !symbolNamed(alpha.Calls, EdgeCall, "Foo") {
		t.Fatalf("alpha calls = %#v, want Foo call edge", alpha.Calls)
	}
	if !containsPath(edgeTargets(alpha.Calls, EdgeCall), "pkg/beta/beta.go") {
		t.Fatalf("alpha call targets = %#v, want pkg/beta/beta.go", alpha.Calls)
	}
	if !symbolNamed(alpha.Calls, EdgeCall, "Method") {
		t.Fatalf("alpha calls = %#v, want Method call edge", alpha.Calls)
	}
	if !symbolNamed(alpha.TypeRefs, EdgeTypeRef, "Type") {
		t.Fatalf("alpha type refs = %#v, want Type type_ref edge", alpha.TypeRefs)
	}
	if containsPath(edgeTargets(alpha.Calls, EdgeCall), "pkg/alpha/alpha.go") {
		t.Fatalf("alpha calls = %#v, want no same-file call edges", alpha.Calls)
	}
	if containsPath(edgeTargets(alpha.TypeRefs, EdgeTypeRef), "pkg/alpha/alpha.go") {
		t.Fatalf("alpha type refs = %#v, want no same-file type_ref edges", alpha.TypeRefs)
	}
	if symbolNamed(alpha.TypeRefs, EdgeTypeRef, "Foo") {
		t.Fatalf("alpha type refs = %#v, want no type_ref for call-only Foo selector", alpha.TypeRefs)
	}

	if !containsPath(edgeTargets(beta.Calls, EdgeCall), "pkg/beta/helper.go") {
		t.Fatalf("beta calls = %#v, want cross-file same-package call to helper.go", beta.Calls)
	}
	if !containsPath(edgeTargets(beta.Calls, EdgeCall), "pkg/gamma/gamma.go") {
		t.Fatalf("beta calls = %#v, want call edge to gamma.go", beta.Calls)
	}
}
