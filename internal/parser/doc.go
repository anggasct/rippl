// Package parser extracts file-level dependency data from Go modules.
//
// Edge types: import, call, type_ref. Implicit implements edges are skipped (FR-P08 / ADR-008).
//
// Call edges come from *ast.CallExpr; type_ref edges come from selector expressions that
// reference types, vars, consts, or func values. Selectors used as the callee of a call
// (e.g. beta.Foo()) emit a call edge only, not a duplicate type_ref.
//
// *_test.go files are in parse scope by default (packages.Config{Tests: true}); edges from
// test files are included in FileAnalysis for downstream graph consumers.
package parser
