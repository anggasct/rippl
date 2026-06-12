package graph

import "github.com/anggasct/rippl/internal/parser"

type Node struct {
	Path    string
	Package string
	Exports []parser.Export
}
