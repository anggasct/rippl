package parser

type EdgeType string

const (
	EdgeImport  EdgeType = "import"
	EdgeCall    EdgeType = "call"
	EdgeTypeRef EdgeType = "type_ref"
)

type Edge struct {
	Type       EdgeType
	TargetFile string
	Symbol     string
}

type Export struct {
	Name string
	Kind string
}

type FileAnalysis struct {
	Path     string
	Package  string
	Imports  []Edge
	Calls    []Edge
	TypeRefs []Edge
	Exports  []Export
}
