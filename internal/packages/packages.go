package packages

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/anggasct/rippl/internal/graph"
)

// UniqueDirs returns sorted unique package directories for the source and affected files.
func UniqueDirs(result *graph.ImpactResult) []string {
	if result == nil {
		return nil
	}

	pkgSet := make(map[string]struct{})
	pkgSet[filepath.Dir(result.Source.Path)] = struct{}{}
	for _, f := range result.Affected {
		pkgSet[filepath.Dir(f.Path)] = struct{}{}
	}

	out := make([]string, 0, len(pkgSet))
	for dir := range pkgSet {
		out = append(out, dir)
	}
	sort.Strings(out)
	return out
}

// ToTestTargets converts package directories to ./relative/path targets.
func ToTestTargets(pkgDirs []string) []string {
	out := make([]string, len(pkgDirs))
	for i, dir := range pkgDirs {
		out[i] = "./" + filepath.ToSlash(dir)
	}
	return out
}

// GoTestCommand builds a single go test invocation covering all package directories.
func GoTestCommand(pkgDirs []string) string {
	targets := make([]string, len(pkgDirs))
	for i, dir := range pkgDirs {
		targets[i] = "./" + filepath.ToSlash(dir) + "/..."
	}
	return "go test " + strings.Join(targets, " ")
}

// WithTests returns package directories that contain _test.go files in the graph,
// plus the count of directories skipped for lacking tests.
func WithTests(g *graph.Graph, pkgDirs []string) ([]string, int) {
	tested := make([]string, 0, len(pkgDirs))
	skipped := 0

	for _, pkgDir := range pkgDirs {
		if HasTests(g, pkgDir) {
			tested = append(tested, pkgDir)
		} else {
			skipped++
		}
	}

	sort.Strings(tested)
	return tested, skipped
}

// HasTests reports whether the graph contains any _test.go file under pkgDir.
func HasTests(g *graph.Graph, pkgDir string) bool {
	for _, f := range g.Files() {
		if filepath.Dir(f) == pkgDir && strings.HasSuffix(f, "_test.go") {
			return true
		}
	}
	return false
}

// SkippedDirs returns sorted package directories in all that are not in tested.
func SkippedDirs(all, tested []string) []string {
	testedSet := make(map[string]struct{}, len(tested))
	for _, dir := range tested {
		testedSet[dir] = struct{}{}
	}

	skipped := make([]string, 0)
	for _, dir := range all {
		if _, ok := testedSet[dir]; !ok {
			skipped = append(skipped, dir)
		}
	}
	sort.Strings(skipped)
	return skipped
}

// AffectedWithTests returns tested package directories for an impact result.
func AffectedWithTests(g *graph.Graph, result *graph.ImpactResult) ([]string, int) {
	return WithTests(g, UniqueDirs(result))
}
