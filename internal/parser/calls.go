package parser

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/types/typeutil"
)

func extractCallsAndTypeRefs(mod *loadedModule, byPath map[string]*FileAnalysis) {
	for _, file := range mod.files {
		analysis := byPath[file.relPath]
		if file.pkg == nil || file.pkg.TypesInfo == nil {
			continue
		}

		callSeen := make(map[string]Edge)
		typeRefSeen := make(map[string]Edge)

		for _, syntaxFile := range file.pkg.Syntax {
			if syntaxFile == nil {
				continue
			}
			syntaxPath := file.pkg.Fset.File(syntaxFile.Pos()).Name()
			if normalizePath(syntaxPath) != normalizePath(file.absPath) {
				continue
			}

			parents := parentMap(syntaxFile)

			ast.Inspect(syntaxFile, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.CallExpr:
					if fn := typeutil.Callee(file.pkg.TypesInfo, n); fn != nil {
						addObjectEdge(mod, analysis.Path, fn, EdgeCall, callSeen)
					}
				case *ast.SelectorExpr:
					if call, ok := parents[n].(*ast.CallExpr); ok && call.Fun == n {
						break
					}
					if obj := objectForSelector(file.pkg.TypesInfo, n); obj != nil {
						switch obj.(type) {
						case *types.TypeName, *types.Var, *types.Const, *types.Func:
							addObjectEdge(mod, analysis.Path, obj, EdgeTypeRef, typeRefSeen)
						}
					}
				}
				return true
			})
		}

		analysis.Calls = sortedEdges(callSeen)
		analysis.TypeRefs = sortedEdges(typeRefSeen)
	}
}

func parentMap(file *ast.File) map[ast.Node]ast.Node {
	parents := make(map[ast.Node]ast.Node)
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		if node := c.Node(); node != nil {
			if parent := c.Parent(); parent != nil {
				parents[node] = parent
			}
		}
		return true
	})
	return parents
}

func objectForSelector(info *types.Info, sel *ast.SelectorExpr) types.Object {
	if info == nil || sel == nil {
		return nil
	}
	if obj, ok := info.Uses[sel.Sel]; ok {
		return obj
	}
	return nil
}

func (mod *loadedModule) fileForTokenPos(pos token.Pos) string {
	if pos == token.NoPos {
		return ""
	}
	for _, pkg := range mod.pkgs {
		if pkg.Fset == nil {
			continue
		}
		file := pkg.Fset.File(pos)
		if file == nil {
			continue
		}
		if rel := relPathForAbs(mod.moduleRoot, file.Name()); rel != "" {
			return rel
		}
	}
	return ""
}

func addObjectEdge(mod *loadedModule, sourcePath string, obj types.Object, edgeType EdgeType, seen map[string]Edge) {
	if obj == nil {
		return
	}

	targetFile := mod.fileForTokenPos(obj.Pos())
	if targetFile == "" || targetFile == sourcePath {
		return
	}

	edge := Edge{
		Type:       edgeType,
		TargetFile: targetFile,
		Symbol:     obj.Name(),
	}
	key := string(edge.Type) + "\x00" + edge.TargetFile + "\x00" + edge.Symbol
	seen[key] = edge
}
