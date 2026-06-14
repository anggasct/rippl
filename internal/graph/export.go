package graph

import (
	"fmt"
	"sort"
	"strings"
)

// GraphExport represents the full dependency graph for serialization.
type GraphExport struct {
	Module  string       `json:"module"`
	Nodes   []NodeExport `json:"nodes"`
	Edges   []EdgeExport `json:"edges"`
	Package string       `json:"package,omitempty"`
}

type NodeExport struct {
	Path    string   `json:"path"`
	Package string   `json:"package"`
	Exports []string `json:"exports,omitempty"`
}

type EdgeExport struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
	Symbol string `json:"symbol,omitempty"`
}

// ExportFull exports the full dependency graph, optionally filtered to a package prefix.
// When packageFilter is empty, all nodes and edges are included.
func ExportFull(g *Graph, modulePath, packageFilter string) (*GraphExport, error) {
	if g == nil {
		return nil, fmt.Errorf("export graph: nil graph")
	}

	export := &GraphExport{
		Module:  modulePath,
		Package: packageFilter,
	}

	files := g.Files()
	for _, f := range files {
		node, ok := g.Node(f)
		if !ok {
			continue
		}
		if packageFilter != "" && !strings.HasPrefix(f, packageFilter) {
			continue
		}

		n := NodeExport{
			Path:    node.Path,
			Package: node.Package,
		}
		for _, exp := range node.Exports {
			n.Exports = append(n.Exports, exp.Name)
		}
		sort.Strings(n.Exports)
		export.Nodes = append(export.Nodes, n)
	}

	// Collect edges from the forward map where both source and target pass the filter.
	seenEdges := make(map[string]bool)
	for _, f := range files {
		if packageFilter != "" && !strings.HasPrefix(f, packageFilter) {
			continue
		}
		deps := g.Dependencies(f)
		for _, dep := range deps {
			target := dep.Target
			if packageFilter != "" && !strings.HasPrefix(target, packageFilter) {
				continue
			}
			edgeKey := f + "|" + target + "|" + string(dep.Type) + "|" + dep.Symbol
			if seenEdges[edgeKey] {
				continue
			}
			seenEdges[edgeKey] = true
			export.Edges = append(export.Edges, EdgeExport{
				Source: f,
				Target: target,
				Type:   string(dep.Type),
				Symbol: dep.Symbol,
			})
		}
	}

	sort.Slice(export.Edges, func(i, j int) bool {
		if export.Edges[i].Source != export.Edges[j].Source {
			return export.Edges[i].Source < export.Edges[j].Source
		}
		if export.Edges[i].Target != export.Edges[j].Target {
			return export.Edges[i].Target < export.Edges[j].Target
		}
		return export.Edges[i].Type < export.Edges[j].Type
	})

	return export, nil
}

// MermaidGraph generates a Mermaid flowchart from a GraphExport.
func MermaidGraph(export *GraphExport) string {
	var b strings.Builder
	b.WriteString("flowchart TD\n")

	for _, node := range export.Nodes {
		nodeID := mermaidNodeID(node.Path)
		fmt.Fprintf(&b, "    %s[%s]\n", nodeID, mermaidLabel(node.Path))
	}

	for _, edge := range export.Edges {
		src := mermaidNodeID(edge.Source)
		tgt := mermaidNodeID(edge.Target)
		fmt.Fprintf(&b, "    %s --> %s\n", src, tgt)
	}

	return b.String()
}

func mermaidNodeID(path string) string {
	return strings.ReplaceAll(strings.ReplaceAll(path, "/", "_"), ".", "_")
}

func mermaidLabel(path string) string {
	return path
}
