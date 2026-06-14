package render

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestWP01JSONSchemaStability verifies that --format json produces stable,
// machine-readable output matching the PRD §12 JSON schema.
// This is the WP-01 acceptance test for CAP-203.
func TestWP01JSONSchemaStability(t *testing.T) {
	t.Parallel()

	out := Output{
		Version:   OutputSchemaVersion,
		Command:   "analyze",
		Module:    "github.com/example/app",
		Generated: time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		Source: SourceOutput{
			Path:      "internal/auth/jwt.go",
			RiskScore: 71,
			RiskBand:  "medium",
			Coverage:  45.2,
		},
		Summary: SummaryOutput{
			AffectedCount: 2,
			DirectCount:   1,
			IndirectCount: 1,
			WithoutTests:  1,
			MaxRiskScore:  82,
		},
		Files: []FileOutput{
			{
				Path:        "handler/admin.go",
				ImpactLevel: "direct",
				Depth:       1,
				RiskScore:   82,
				RiskBand:    "high",
				Coverage:    12.0,
				HasTestFile: true,
				Chain:       []string{"internal/auth/jwt.go", "handler/admin.go"},
				Reason:      "import",
			},
			{
				Path:        "handler/user.go",
				ImpactLevel: "indirect",
				Depth:       2,
				RiskScore:   45,
				RiskBand:    "medium",
				Coverage:    0,
				HasTestFile: false,
				Chain:       []string{"internal/auth/jwt.go", "handler/admin.go", "handler/user.go"},
				Reason:      "call",
			},
		},
		SuggestedActions: &SuggestedActionsOutput{
			PackagesToTest: []string{"./handler"},
			Commands: []string{
				"go test ./handler/...",
				"rippl score internal/auth/jwt.go --format json",
			},
			UntestedHighRisk: []UntestedHighRiskEntry{},
		},
	}

	var buf bytes.Buffer
	r := &jsonRenderer{out: &buf}
	if err := r.Render(context.Background(), out); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	raw := buf.Bytes()

	// 1. Must be valid JSON.
	var doc map[string]interface{}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// 2. Top-level required fields.
	requiredTop := []string{"version", "command", "source_file", "module", "generated_at", "summary", "source", "affected", "suggested_actions"}
	for _, key := range requiredTop {
		if _, ok := doc[key]; !ok {
			t.Errorf("missing top-level field %q", key)
		}
	}

	// 3. source object fields.
	src := doc["source"].(map[string]interface{})
	requiredSource := []string{"path", "risk_score", "risk_band", "coverage_pct"}
	for _, key := range requiredSource {
		if _, ok := src[key]; !ok {
			t.Errorf("missing source field %q", key)
		}
	}

	// 4. summary object fields.
	sum := doc["summary"].(map[string]interface{})
	requiredSummary := []string{"affected_count", "direct_count", "indirect_count", "without_tests", "max_risk_score"}
	for _, key := range requiredSummary {
		if _, ok := sum[key]; !ok {
			t.Errorf("missing summary field %q", key)
		}
	}

	// 5. affected array — each entry has required fields.
	affected := doc["affected"].([]interface{})
	if len(affected) != 2 {
		t.Fatalf("affected length = %d, want 2", len(affected))
	}
	requiredAffected := []string{"path", "impact_level", "depth", "risk_score", "risk_band", "coverage_pct", "has_test_file", "chain", "reason"}
	for i, a := range affected {
		entry := a.(map[string]interface{})
		for _, key := range requiredAffected {
			if _, ok := entry[key]; !ok {
				t.Errorf("affected[%d] missing field %q", i, key)
			}
		}
	}

	// 6. suggested_actions object fields.
	actions := doc["suggested_actions"].(map[string]interface{})
	for _, key := range []string{"packages_to_test", "commands", "untested_high_risk"} {
		if _, ok := actions[key]; !ok {
			t.Errorf("missing suggested_actions field %q", key)
		}
	}

	// 7. Risk bands must be valid enum values.
	validBands := map[string]bool{"high": true, "medium": true, "low": true, "minimal": true}
	band, _ := src["risk_band"].(string)
	if !validBands[band] {
		t.Errorf("source risk_band = %q, want one of [high medium low minimal]", band)
	}
	for i, a := range affected {
		entry := a.(map[string]interface{})
		b, _ := entry["risk_band"].(string)
		if !validBands[b] {
			t.Errorf("affected[%d] risk_band = %q, want one of [high medium low minimal]", i, b)
		}
	}

	// 8. No ANSI escape codes in output.
	if strings.Contains(buf.String(), "\x1b[") {
		t.Error("JSON output contains ANSI escape codes")
	}

	// 9. Deterministic: re-encode and compare.
	var buf2 bytes.Buffer
	r2 := &jsonRenderer{out: &buf2}
	if err := r2.Render(context.Background(), out); err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if buf.String() != buf2.String() {
		t.Error("JSON output is not deterministic between renders")
	}
}

// TestWP01JSONEmptyAffected verifies JSON output with no affected files.
func TestWP01JSONEmptyAffected(t *testing.T) {
	t.Parallel()

	out := Output{
		Version:   OutputSchemaVersion,
		Command:   "analyze",
		Module:    "github.com/example/app",
		Generated: time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		Source: SourceOutput{
			Path:      "internal/auth/jwt.go",
			RiskScore: 30,
			RiskBand:  "low",
			Coverage:  80.0,
		},
		Summary: SummaryOutput{
			AffectedCount: 0,
			DirectCount:   0,
			IndirectCount: 0,
			WithoutTests:  0,
			MaxRiskScore:  0,
		},
		Files: []FileOutput{},
	}

	var buf bytes.Buffer
	r := &jsonRenderer{out: &buf}
	if err := r.Render(context.Background(), out); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	affected := doc["affected"].([]interface{})
	if len(affected) != 0 {
		t.Errorf("affected length = %d, want 0", len(affected))
	}
}

// TestWP02MermaidExport verifies that --format mermaid produces valid Mermaid
// graph syntax with risk bands and direct/indirect distinction.
// This is the WP-02 acceptance test for CAP-203.
func TestWP02MermaidExport(t *testing.T) {
	t.Parallel()

	out := Output{
		Version:   OutputSchemaVersion,
		Command:   "analyze",
		Module:    "github.com/example/app",
		Generated: time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		Source: SourceOutput{
			Path:      "internal/auth/jwt.go",
			RiskScore: 71,
			RiskBand:  "medium",
			Coverage:  45.2,
		},
		Summary: SummaryOutput{
			AffectedCount: 3,
			DirectCount:   2,
			IndirectCount: 1,
			WithoutTests:  1,
			MaxRiskScore:  82,
		},
		Files: []FileOutput{
			{
				Path:        "handler/admin.go",
				ImpactLevel: "direct",
				Depth:       1,
				RiskScore:   82,
				RiskBand:    "high",
				HasTestFile: true,
				Reason:      "import",
			},
			{
				Path:        "handler/user.go",
				ImpactLevel: "direct",
				Depth:       1,
				RiskScore:   55,
				RiskBand:    "medium",
				HasTestFile: true,
				Reason:      "import",
			},
			{
				Path:        "handler/audit.go",
				ImpactLevel: "indirect",
				Depth:       2,
				RiskScore:   30,
				RiskBand:    "low",
				HasTestFile: false,
				Reason:      "call",
			},
		},
	}

	var buf bytes.Buffer
	r := &mermaidRenderer{out: &buf}
	if err := r.Render(context.Background(), out); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	graph := buf.String()
	lines := strings.Split(strings.TrimSpace(graph), "\n")

	// 1. Must start with "graph TD".
	if len(lines) == 0 {
		t.Fatal("mermaid output is empty")
	}
	if strings.TrimSpace(lines[0]) != "graph TD" {
		t.Errorf("first line = %q, want %q", strings.TrimSpace(lines[0]), "graph TD")
	}

	// 2. Must contain source node.
	srcID := mermaidID(out.Source.Path)
	if !strings.Contains(graph, srcID+"[") {
		t.Errorf("missing source node %q in output", srcID)
	}

	// 3. Must contain all affected file nodes.
	for i := range out.Files {
		nodeID := fmt.Sprintf("n%d", i)
		if !strings.Contains(graph, nodeID+"[") {
			t.Errorf("missing affected node %q in output", nodeID)
		}
	}

	// 4. Direct files must have solid edges from source.
	directCount := 0
	for _, f := range out.Files {
		if f.ImpactLevel == "direct" {
			directCount++
		}
	}
	solidEdgeCount := strings.Count(graph, srcID+" -->")
	if solidEdgeCount != directCount {
		t.Errorf("solid edges from source = %d, want %d (direct files)", solidEdgeCount, directCount)
	}

	// 5. Indirect files must have dashed edges.
	indirectCount := 0
	for _, f := range out.Files {
		if f.ImpactLevel != "direct" {
			indirectCount++
		}
	}
	dashedEdgeCount := strings.Count(graph, "-.->")
	if dashedEdgeCount != indirectCount {
		t.Errorf("dashed edges = %d, want %d (indirect files)", dashedEdgeCount, indirectCount)
	}

	// 6. Must contain classDef for styling.
	requiredClasses := []string{"classDef source", "classDef riskHigh", "classDef riskMedium", "classDef riskLow"}
	for _, cls := range requiredClasses {
		if !strings.Contains(graph, cls) {
			t.Errorf("missing %q in output", cls)
		}
	}

	// 7. Risk band classes applied to nodes.
	for i, f := range out.Files {
		nodeID := fmt.Sprintf("n%d", i)
		var riskClass string
		switch f.RiskBand {
		case "high":
			riskClass = "riskHigh"
		case "medium":
			riskClass = "riskMedium"
		case "low":
			riskClass = "riskLow"
		}
		if riskClass != "" {
			expected := fmt.Sprintf("class %s %s", nodeID, riskClass)
			if !strings.Contains(graph, expected) {
				t.Errorf("missing %q in output", expected)
			}
		}
	}

	// 8. No ANSI escape codes.
	if strings.Contains(graph, "\x1b[") {
		t.Error("Mermaid output contains ANSI escape codes")
	}

	// 9. Source node styled.
	srcClassLine := fmt.Sprintf("class %s source", srcID)
	if !strings.Contains(graph, srcClassLine) {
		t.Errorf("missing source class styling %q", srcClassLine)
	}
}

// TestWP02MermaidEmptyAffected verifies mermaid output with no affected files.
func TestWP02MermaidEmptyAffected(t *testing.T) {
	t.Parallel()

	out := Output{
		Version:   OutputSchemaVersion,
		Command:   "analyze",
		Module:    "github.com/example/app",
		Generated: time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		Source: SourceOutput{
			Path:      "internal/auth/jwt.go",
			RiskScore: 30,
			RiskBand:  "low",
			Coverage:  80.0,
		},
		Summary: SummaryOutput{
			AffectedCount: 0,
			DirectCount:   0,
			IndirectCount: 0,
			WithoutTests:  0,
			MaxRiskScore:  0,
		},
		Files: []FileOutput{},
	}

	var buf bytes.Buffer
	r := &mermaidRenderer{out: &buf}
	if err := r.Render(context.Background(), out); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	graph := buf.String()
	if !strings.Contains(graph, "graph TD") {
		t.Error("missing graph TD header")
	}
	if strings.Contains(graph, "n0[") {
		t.Error("should not have any affected file nodes")
	}
}

// TestWP02MermaidNodeIDs verifies that mermaidID produces valid node IDs.
func TestWP02MermaidNodeIDs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path string
		want string
	}{
		{"internal/auth/jwt.go", "internal_auth_jwt_go"},
		{"handler/admin.go", "handler_admin_go"},
		{"main.go", "main_go"},
	}
	for _, tc := range cases {
		got := mermaidID(tc.path)
		if got != tc.want {
			t.Errorf("mermaidID(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

// TestWP02MermaidLabel verifies label formatting.
func TestWP02MermaidLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path     string
		riskBand string
		score    int
		want     string
	}{
		{"handler/admin.go", "high", 82, "handler/admin.go\\nrisk:82 (high)"},
		{"handler/user.go", "", 0, "handler/user.go"},
	}
	for _, tc := range cases {
		got := mermaidLabel(tc.path, tc.riskBand, tc.score)
		if got != tc.want {
			t.Errorf("mermaidLabel(%q, %q, %d) = %q, want %q", tc.path, tc.riskBand, tc.score, got, tc.want)
		}
	}
}

func makeTextOutput(fileCount int) Output {
	files := make([]FileOutput, fileCount)
	for i := range files {
		files[i] = FileOutput{
			Path:        fmt.Sprintf("pkg/file%d.go", i),
			ImpactLevel: "direct",
			Depth:       1,
			RiskScore:   50,
			Chain:       []string{"source.go", fmt.Sprintf("pkg/file%d.go", i)},
		}
	}
	return Output{
		Source: SourceOutput{Path: "source.go"},
		Summary: SummaryOutput{
			AffectedCount: fileCount,
			DirectCount:   fileCount,
		},
		Files: files,
	}
}

func TestTextRendererTruncatesLongList(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	r := &textRenderer{out: &buf}
	if err := r.Render(context.Background(), makeTextOutput(25)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	out := buf.String()
	if strings.Count(out, "\n  ") < 20 {
		t.Fatalf("expected at least 20 file lines, got:\n%s", out)
	}
	if !strings.Contains(out, "... and 5 more affected files") {
		t.Fatalf("output missing truncation footer, got:\n%s", out)
	}
	if strings.Contains(out, "file24.go") {
		t.Fatalf("output should not include file beyond limit, got:\n%s", out)
	}
}

func TestTextRendererFullListWhenShort(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	r := &textRenderer{out: &buf}
	if err := r.Render(context.Background(), makeTextOutput(5)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "... and") {
		t.Fatalf("output should not truncate short lists, got:\n%s", out)
	}
	if !strings.Contains(out, "file4.go") {
		t.Fatalf("output missing last file, got:\n%s", out)
	}
}
