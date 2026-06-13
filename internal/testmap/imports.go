package testmap

import (
	"github.com/anggasct/rippl/internal/graph"
	"github.com/anggasct/rippl/internal/parser"
)

func buildPackageSources(g *graph.Graph) map[string][]string {
	out := make(map[string][]string)
	if g == nil {
		return out
	}
	for _, path := range g.Files() {
		if isTestFile(path) {
			continue
		}
		node, ok := g.Node(path)
		if !ok || node.Package == "" {
			continue
		}
		out[node.Package] = append(out[node.Package], path)
	}
	return out
}

func applyImportAssociations(g *graph.Graph, out map[string]FileCoverage) {
	if g == nil {
		return
	}
	pkgSources := buildPackageSources(g)
	for _, path := range g.Files() {
		if !isTestFile(path) {
			continue
		}
		if !isExternalTestFile(path, g) {
			continue
		}
		for _, edge := range g.Dependencies(path) {
			if edge.Type != parser.EdgeImport {
				continue
			}
			targetNode, ok := g.Node(edge.Target)
			if !ok {
				continue
			}
			for _, srcPath := range pkgSources[targetNode.Package] {
				appendTestFile(out, srcPath, path)
			}
		}
	}
}

// isExternalTestFile reports whether FR-T02 import mapping applies.
// In-package tests (X_test.go → X.go exists) use naming only; package-level import
// edges would otherwise attribute production imports to the test file.
func isExternalTestFile(testPath string, g *graph.Graph) bool {
	sourcePath, ok := sourcePathForTest(testPath)
	if !ok {
		return false
	}
	_, exists := g.Node(sourcePath)
	return !exists
}
