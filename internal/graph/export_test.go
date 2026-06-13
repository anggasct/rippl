package graph

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExportFull_AllNodes(t *testing.T) {
	t.Parallel()

	g := Build(parseMinimodule(t))
	export, err := ExportFull(g, "github.com/minimodule", "")
	if err != nil {
		t.Fatalf("ExportFull() error = %v", err)
	}

	if export.Module != "github.com/minimodule" {
		t.Fatalf("Module = %q, want %q", export.Module, "github.com/minimodule")
	}
	if len(export.Nodes) != 6 {
		t.Fatalf("Nodes = %d, want 6", len(export.Nodes))
	}
	if len(export.Edges) == 0 {
		t.Fatal("Edges is empty, want non-empty")
	}
}

func TestExportFull_PackageFilter(t *testing.T) {
	t.Parallel()

	g := Build(parseMinimodule(t))
	export, err := ExportFull(g, "github.com/minimodule", "pkg/alpha")
	if err != nil {
		t.Fatalf("ExportFull() error = %v", err)
	}

	if export.Package != "pkg/alpha" {
		t.Fatalf("Package = %q, want %q", export.Package, "pkg/alpha")
	}

	for _, node := range export.Nodes {
		if !strings.HasPrefix(node.Path, "pkg/alpha") {
			t.Fatalf("node %q does not match filter pkg/alpha", node.Path)
		}
	}

	for _, edge := range export.Edges {
		if !strings.HasPrefix(edge.Source, "pkg/alpha") {
			t.Fatalf("edge source %q does not match filter pkg/alpha", edge.Source)
		}
		if !strings.HasPrefix(edge.Target, "pkg/alpha") {
			t.Fatalf("edge target %q does not match filter pkg/alpha", edge.Target)
		}
	}
}

func TestExportFull_NilGraph(t *testing.T) {
	t.Parallel()

	_, err := ExportFull(nil, "github.com/minimodule", "")
	if err == nil {
		t.Fatal("ExportFull(nil) should return error")
	}
}

func TestExportFull_EmptyPackageFilter(t *testing.T) {
	t.Parallel()

	g := Build(parseMinimodule(t))
	export, err := ExportFull(g, "github.com/minimodule", "")
	if err != nil {
		t.Fatalf("ExportFull() error = %v", err)
	}

	// With no filter, all nodes should be present
	files := g.Files()
	if len(export.Nodes) != len(files) {
		t.Fatalf("Nodes = %d, want %d", len(export.Nodes), len(files))
	}
}

func TestExportFull_NodeExportsSorted(t *testing.T) {
	t.Parallel()

	g := Build(parseMinimodule(t))
	export, err := ExportFull(g, "github.com/minimodule", "")
	if err != nil {
		t.Fatalf("ExportFull() error = %v", err)
	}

	for _, node := range export.Nodes {
		for i := 1; i < len(node.Exports); i++ {
			if node.Exports[i-1] >= node.Exports[i] {
				t.Fatalf("exports not sorted for %s: %v", node.Path, node.Exports)
			}
		}
	}
}

func TestExportFull_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	g := Build(parseMinimodule(t))
	export, err := ExportFull(g, "github.com/minimodule", "")
	if err != nil {
		t.Fatalf("ExportFull() error = %v", err)
	}

	data, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded GraphExport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Module != export.Module {
		t.Fatalf("Module = %q, want %q", decoded.Module, export.Module)
	}
	if len(decoded.Nodes) != len(export.Nodes) {
		t.Fatalf("Nodes = %d, want %d", len(decoded.Nodes), len(export.Nodes))
	}
	if len(decoded.Edges) != len(export.Edges) {
		t.Fatalf("Edges = %d, want %d", len(decoded.Edges), len(export.Edges))
	}
}

func TestMermaidGraph_Output(t *testing.T) {
	t.Parallel()

	g := Build(parseMinimodule(t))
	export, err := ExportFull(g, "github.com/minimodule", "")
	if err != nil {
		t.Fatalf("ExportFull() error = %v", err)
	}

	out := MermaidGraph(export)
	lines := strings.Split(strings.TrimSpace(out), "\n")

	if lines[0] != "flowchart TD" {
		t.Fatalf("first line = %q, want %q", lines[0], "flowchart TD")
	}

	// Must have node lines and edge lines
	hasNode := false
	hasEdge := false
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "[") && strings.Contains(trimmed, "]") {
			hasNode = true
		}
		if strings.Contains(trimmed, "-->") {
			hasEdge = true
		}
	}
	if !hasNode {
		t.Fatal("Mermaid output has no node lines")
	}
	if !hasEdge {
		t.Fatal("Mermaid output has no edge lines")
	}
}

func TestMermaidGraph_PackageFilter(t *testing.T) {
	t.Parallel()

	g := Build(parseMinimodule(t))
	export, err := ExportFull(g, "github.com/minimodule", "pkg/beta")
	if err != nil {
		t.Fatalf("ExportFull() error = %v", err)
	}

	out := MermaidGraph(export)
	lines := strings.Split(strings.TrimSpace(out), "\n")

	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		// All node/edge references should contain pkg_beta (filtered)
		if strings.Contains(trimmed, "pkg_alpha") {
			t.Fatalf("Mermaid output contains pkg_alpha reference with beta filter: %s", trimmed)
		}
	}
}

func TestMermaidNodeID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{"pkg/alpha/alpha.go", "pkg_alpha_alpha_go"},
		{"main.go", "main_go"},
		{"a/b/c.go", "a_b_c_go"},
	}
	for _, tc := range cases {
		got := mermaidNodeID(tc.input)
		if got != tc.want {
			t.Fatalf("mermaidNodeID(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
