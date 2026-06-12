package graph

import (
	"sort"

	"github.com/anggasct/rippl/internal/parser"
)

type Edge struct {
	Target string
	Type   parser.EdgeType
	Symbol string
}

type Graph struct {
	nodes    map[string]*Node
	forward  map[string][]Edge
	backward map[string][]Edge
}

func Build(analyses []parser.FileAnalysis) *Graph {
	g := &Graph{
		nodes:    make(map[string]*Node, len(analyses)),
		forward:  make(map[string][]Edge),
		backward: make(map[string][]Edge),
	}

	for _, analysis := range analyses {
		g.nodes[analysis.Path] = &Node{
			Path:    analysis.Path,
			Package: analysis.Package,
			Exports: append([]parser.Export(nil), analysis.Exports...),
		}
	}

	for _, analysis := range analyses {
		addParserEdges(g, analysis.Path, analysis.Imports)
		addParserEdges(g, analysis.Path, analysis.Calls)
		addParserEdges(g, analysis.Path, analysis.TypeRefs)
	}

	return g
}

func addParserEdges(g *Graph, source string, edges []parser.Edge) {
	for _, edge := range edges {
		if edge.TargetFile == "" || edge.TargetFile == source {
			continue
		}
		g.addEdge(source, edge.TargetFile, Edge{
			Target: edge.TargetFile,
			Type:   edge.Type,
			Symbol: edge.Symbol,
		})
	}
}

func (g *Graph) addEdge(source, target string, backwardEdge Edge) {
	if g.nodes[source] == nil || g.nodes[target] == nil {
		return
	}

	forwardEdge := Edge{
		Target: source,
		Type:   backwardEdge.Type,
		Symbol: backwardEdge.Symbol,
	}

	if !containsEdge(g.backward[source], backwardEdge) {
		g.backward[source] = append(g.backward[source], backwardEdge)
		sortEdges(g.backward[source])
	}
	if !containsEdge(g.forward[target], forwardEdge) {
		g.forward[target] = append(g.forward[target], forwardEdge)
		sortEdges(g.forward[target])
	}
}

func containsEdge(edges []Edge, want Edge) bool {
	for _, edge := range edges {
		if edge.Target == want.Target && edge.Type == want.Type && edge.Symbol == want.Symbol {
			return true
		}
	}
	return false
}

func sortEdges(edges []Edge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Target == edges[j].Target {
			if edges[i].Symbol == edges[j].Symbol {
				return edges[i].Type < edges[j].Type
			}
			return edges[i].Symbol < edges[j].Symbol
		}
		return edges[i].Target < edges[j].Target
	})
}

func (g *Graph) Dependents(file string) []Edge {
	edges := g.forward[file]
	out := append([]Edge(nil), edges...)
	return out
}

func (g *Graph) Dependencies(file string) []Edge {
	edges := g.backward[file]
	out := append([]Edge(nil), edges...)
	return out
}

func (g *Graph) Node(file string) (*Node, bool) {
	node, ok := g.nodes[file]
	return node, ok
}

func (g *Graph) Files() []string {
	out := make([]string, 0, len(g.nodes))
	for path := range g.nodes {
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

func (g *Graph) snapshot() cacheSnapshot {
	return cacheSnapshot{
		Nodes:    cloneNodeMap(g.nodes),
		Forward:  cloneEdgeMap(g.forward),
		Backward: cloneEdgeMap(g.backward),
	}
}

func graphFromSnapshot(snapshot cacheSnapshot) *Graph {
	return &Graph{
		nodes:    cloneNodeMap(snapshot.Nodes),
		forward:  cloneEdgeMap(snapshot.Forward),
		backward: cloneEdgeMap(snapshot.Backward),
	}
}

func cloneNodeMap(in map[string]*Node) map[string]*Node {
	out := make(map[string]*Node, len(in))
	for path, node := range in {
		exports := append([]parser.Export(nil), node.Exports...)
		out[path] = &Node{
			Path:    node.Path,
			Package: node.Package,
			Exports: exports,
		}
	}
	return out
}

func cloneEdgeMap(in map[string][]Edge) map[string][]Edge {
	out := make(map[string][]Edge, len(in))
	for path, edges := range in {
		out[path] = append([]Edge(nil), edges...)
	}
	return out
}
