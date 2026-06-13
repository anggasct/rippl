package graph

import (
	"testing"
)

func TestBuildForwardBackwardMaps(t *testing.T) {
	t.Parallel()

	g := Build(parseMinimodule(t))

	alphaDeps := g.Dependencies("pkg/alpha/alpha.go")
	if !hasDependencyTarget(alphaDeps, "pkg/beta/beta.go") {
		t.Fatalf("alpha dependencies = %#v, want edge to beta", alphaDeps)
	}

	betaDependents := g.Dependents("pkg/beta/beta.go")
	if !hasDependentTarget(betaDependents, "pkg/alpha/alpha.go") {
		t.Fatalf("beta dependents = %#v, want edge from alpha", betaDependents)
	}
}

func TestBuildNodeMetadata(t *testing.T) {
	t.Parallel()

	g := Build(parseMinimodule(t))

	beta, ok := g.Node("pkg/beta/beta.go")
	if !ok {
		t.Fatal("missing beta node")
	}
	if beta.Package != "example.com/minimodule/pkg/beta" {
		t.Fatalf("beta package = %q", beta.Package)
	}
	if !hasExport(beta.Exports, "Foo", "func") {
		t.Fatalf("beta exports = %#v, want Foo func", beta.Exports)
	}
	if !hasExport(beta.Exports, "Type", "type") {
		t.Fatalf("beta exports = %#v, want Type type", beta.Exports)
	}
}

func TestGraphFilesSorted(t *testing.T) {
	t.Parallel()

	g := Build(parseMinimodule(t))
	files := g.Files()
	if len(files) != 6 {
		t.Fatalf("files = %#v, want 6 entries", files)
	}
	for i := 1; i < len(files); i++ {
		if files[i-1] >= files[i] {
			t.Fatalf("files not sorted: %#v", files)
		}
	}
}
