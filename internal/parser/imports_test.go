package parser

import (
	"context"
	"testing"
)

func TestExtractImportsAndExports(t *testing.T) {
	t.Parallel()

	analyses, err := ParseModule(context.Background(), minimoduleRoot(t), defaultTestConfig())
	if err != nil {
		t.Fatalf("ParseModule() error = %v", err)
	}

	byPath := analysisByPath(t, analyses)
	alpha, ok := byPath["pkg/alpha/alpha.go"]
	if !ok {
		t.Fatal("missing alpha analysis")
	}
	beta, ok := byPath["pkg/beta/beta.go"]
	if !ok {
		t.Fatal("missing beta analysis")
	}

	if !containsPath(edgeTargets(alpha.Imports, EdgeImport), "pkg/beta/beta.go") {
		t.Fatalf("alpha imports = %#v, want edge to pkg/beta/beta.go", alpha.Imports)
	}
	if !hasExport(beta.Exports, "Foo", "func") {
		t.Fatalf("beta exports = %#v, want Foo func", beta.Exports)
	}
	if !hasExport(beta.Exports, "Type", "type") {
		t.Fatalf("beta exports = %#v, want Type type", beta.Exports)
	}
}
