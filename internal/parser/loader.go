package parser

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"golang.org/x/tools/go/packages"
)

const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedImports |
	packages.NeedTypes |
	packages.NeedTypesInfo |
	packages.NeedSyntax |
	packages.NeedDeps

type fileRef struct {
	relPath string
	absPath string
	pkg     *packages.Package
}

type loadedModule struct {
	moduleRoot string
	pkgs       []*packages.Package
	files      []fileRef
}

func loadModule(ctx context.Context, moduleRoot string, ignorePatterns []string) (*loadedModule, error) {
	moduleRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve module root: %w", err)
	}

	cfg := &packages.Config{
		Context: ctx,
		Mode:    loadMode,
		Dir:     moduleRoot,
		Tests:   true,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("load packages: no packages found")
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("load packages: typecheck errors")
	}

	loaded := &loadedModule{
		moduleRoot: moduleRoot,
		pkgs:       pkgs,
	}

	seen := make(map[string]fileRef)
	for _, pkg := range pkgs {
		if pkg.IllTyped && len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("typecheck %s: %s", pkg.PkgPath, pkg.Errors[0].Msg)
		}

		for _, absPath := range sourceFiles(pkg) {
			relPath, err := filepath.Rel(moduleRoot, absPath)
			if err != nil {
				return nil, fmt.Errorf("rel path for %q: %w", absPath, err)
			}
			relPath = normalizePath(relPath)
			if !isGoFile(relPath) {
				continue
			}
			if matchesIgnore(relPath, ignorePatterns) {
				continue
			}
			if _, ok := seen[relPath]; ok {
				continue
			}
			seen[relPath] = fileRef{
				relPath: relPath,
				absPath: absPath,
				pkg:     pkg,
			}
		}
	}

	loaded.files = make([]fileRef, 0, len(seen))
	for _, file := range seen {
		loaded.files = append(loaded.files, file)
	}
	sort.Slice(loaded.files, func(i, j int) bool {
		return loaded.files[i].relPath < loaded.files[j].relPath
	})

	return loaded, nil
}

func sourceFiles(pkg *packages.Package) []string {
	if len(pkg.CompiledGoFiles) > 0 {
		return append([]string{}, pkg.CompiledGoFiles...)
	}
	return append([]string{}, pkg.GoFiles...)
}

func (m *loadedModule) pkgByPath() map[string]*packages.Package {
	out := make(map[string]*packages.Package, len(m.pkgs))
	for _, pkg := range m.pkgs {
		if pkg.PkgPath == "" || pkg.PkgPath == "unsafe" {
			continue
		}
		out[pkg.PkgPath] = pkg
	}
	return out
}

// packageSourceFiles returns non-test compiled source paths for import edge targets.
func packageSourceFiles(moduleRoot string, pkg *packages.Package, ignorePatterns []string) []string {
	paths := make([]string, 0, len(pkg.GoFiles))
	for _, absPath := range pkg.GoFiles {
		relPath, err := filepath.Rel(moduleRoot, absPath)
		if err != nil {
			continue
		}
		relPath = normalizePath(relPath)
		if matchesIgnore(relPath, ignorePatterns) {
			continue
		}
		paths = append(paths, relPath)
	}
	sort.Strings(paths)
	return paths
}

func relPathForAbs(moduleRoot, absPath string) string {
	relPath, err := filepath.Rel(moduleRoot, absPath)
	if err != nil {
		return ""
	}
	return normalizePath(relPath)
}
