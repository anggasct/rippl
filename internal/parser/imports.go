package parser

import (
	"go/ast"
	"go/token"
	"sort"

	"golang.org/x/tools/go/packages"
)

func extractImportsAndExports(mod *loadedModule, ignorePatterns []string) map[string]*FileAnalysis {
	byPath := make(map[string]*FileAnalysis, len(mod.files))
	for _, file := range mod.files {
		byPath[file.relPath] = &FileAnalysis{
			Path:    file.relPath,
			Package: file.pkg.PkgPath,
		}
	}

	pkgByPath := mod.pkgByPath()
	for _, file := range mod.files {
		analysis := byPath[file.relPath]
		analysis.Exports = extractExports(file)
		analysis.Imports = extractImportEdges(mod.moduleRoot, file, pkgByPath, ignorePatterns)
	}

	return byPath
}

func extractImportEdges(moduleRoot string, file fileRef, pkgByPath map[string]*packages.Package, ignorePatterns []string) []Edge {
	if file.pkg == nil || file.pkg.Types == nil {
		return nil
	}

	seen := make(map[string]Edge)
	for _, imp := range file.pkg.Imports {
		if imp == nil {
			continue
		}
		targetPkg, ok := pkgByPath[imp.PkgPath]
		if !ok || targetPkg.PkgPath == file.pkg.PkgPath {
			continue
		}
		for _, targetFile := range packageSourceFiles(moduleRoot, targetPkg, ignorePatterns) {
			key := string(EdgeImport) + "\x00" + targetFile
			seen[key] = Edge{
				Type:       EdgeImport,
				TargetFile: targetFile,
			}
		}
	}

	return sortedEdges(seen)
}

func extractExports(file fileRef) []Export {
	if file.pkg == nil || file.pkg.Fset == nil || file.pkg.Syntax == nil {
		return nil
	}

	var exports []Export
	for _, syntaxFile := range file.pkg.Syntax {
		if syntaxFile == nil {
			continue
		}
		syntaxPath := file.pkg.Fset.File(syntaxFile.Pos()).Name()
		if normalizePath(syntaxPath) != normalizePath(file.absPath) {
			continue
		}

		for _, decl := range syntaxFile.Decls {
			switch node := decl.(type) {
			case *ast.FuncDecl:
				if node.Name != nil && node.Name.IsExported() {
					kind := "func"
					if node.Recv != nil {
						kind = "method"
					}
					exports = append(exports, Export{Name: node.Name.Name, Kind: kind})
				}
			case *ast.GenDecl:
				for _, spec := range node.Specs {
					switch specNode := spec.(type) {
					case *ast.TypeSpec:
						if specNode.Name != nil && specNode.Name.IsExported() {
							kind := "type"
							if _, ok := specNode.Type.(*ast.InterfaceType); ok {
								kind = "interface"
							}
							exports = append(exports, Export{Name: specNode.Name.Name, Kind: kind})
						}
					case *ast.ValueSpec:
						for _, name := range specNode.Names {
							if name.IsExported() {
								kind := "const"
								if node.Tok == token.VAR {
									kind = "var"
								}
								exports = append(exports, Export{Name: name.Name, Kind: kind})
							}
						}
					}
				}
			}
		}
	}

	sort.Slice(exports, func(i, j int) bool {
		if exports[i].Name == exports[j].Name {
			return exports[i].Kind < exports[j].Kind
		}
		return exports[i].Name < exports[j].Name
	})
	return exports
}

func sortedEdges(seen map[string]Edge) []Edge {
	edges := make([]Edge, 0, len(seen))
	for _, edge := range seen {
		edges = append(edges, edge)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].TargetFile == edges[j].TargetFile {
			if edges[i].Symbol == edges[j].Symbol {
				return edges[i].Type < edges[j].Type
			}
			return edges[i].Symbol < edges[j].Symbol
		}
		return edges[i].TargetFile < edges[j].TargetFile
	})
	return edges
}

func hasExport(exports []Export, name, kind string) bool {
	for _, exp := range exports {
		if exp.Name == name && (kind == "" || exp.Kind == kind) {
			return true
		}
	}
	return false
}

func edgeTargets(edges []Edge, edgeType EdgeType) []string {
	var out []string
	for _, edge := range edges {
		if edge.Type == edgeType {
			out = append(out, edge.TargetFile)
		}
	}
	sort.Strings(out)
	return out
}

func containsPath(paths []string, target string) bool {
	for _, path := range paths {
		if path == target {
			return true
		}
	}
	return false
}

func symbolNamed(edges []Edge, edgeType EdgeType, symbol string) bool {
	for _, edge := range edges {
		if edge.Type == edgeType && edge.Symbol == symbol {
			return true
		}
	}
	return false
}
