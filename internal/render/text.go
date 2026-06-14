package render

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/anggasct/rippl/internal/tui"
)

const textAffectedFileLimit = 20

// NewRenderer creates a Renderer for the given format string.
// Supported formats: text, json, mermaid, tui.
// Falls back to text for unknown formats.
// The renderer writes to os.Stdout.
func NewRenderer(format string) (Renderer, error) {
	return NewRendererWithWriter(format, os.Stdout)
}

// NewRendererWithWriter creates a Renderer for the given format string that writes to w.
func NewRendererWithWriter(format string, w io.Writer) (Renderer, error) {
	return NewRendererWithWriterAndColor(format, w, false)
}

// NewRendererWithWriterAndColor creates a Renderer with explicit noColor control.
// Used by callers that need to disable ANSI colors (e.g. --no-color flag).
func NewRendererWithWriterAndColor(format string, w io.Writer, noColor bool) (Renderer, error) {
	switch Format(strings.ToLower(format)) {
	case FormatJSON:
		return &jsonRenderer{out: w}, nil
	case FormatMermaid:
		return &mermaidRenderer{out: w}, nil
	case FormatTUI:
		return &tuiRenderer{out: w, noColor: noColor}, nil
	case FormatText, "":
		return &textRenderer{out: w}, nil
	default:
		return &textRenderer{out: w}, nil
	}
}

// textRenderer writes a human-readable text summary to stdout.
type textRenderer struct {
	out io.Writer
}

func (r *textRenderer) Render(ctx context.Context, out Output) error {
	if _, err := fmt.Fprintf(r.out, "Source: %s\n", out.Source.Path); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(r.out, "Affected: %d files\n", out.Summary.AffectedCount); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(r.out, "  Direct: %d  Indirect: %d\n", out.Summary.DirectCount, out.Summary.IndirectCount); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(r.out, "Without tests: %d  Max risk: %d\n", out.Summary.WithoutTests, out.Summary.MaxRiskScore); err != nil {
		return err
	}

	if len(out.Files) == 0 {
		return nil
	}

	if _, err := fmt.Fprintln(r.out, "\nAffected files:"); err != nil {
		return err
	}

	displayCount := len(out.Files)
	if displayCount > textAffectedFileLimit {
		displayCount = textAffectedFileLimit
	}

	for i := 0; i < displayCount; i++ {
		f := out.Files[i]
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		chain := strings.Join(f.Chain, " -> ")
		if chain == "" {
			chain = f.Path
		}
		testStatus := ""
		if !f.HasTestFile {
			testStatus = " [no test]"
		}

		if _, err := fmt.Fprintf(r.out, "  %d. [%s] %s (depth=%d, risk=%d)%s\n",
			i+1, strings.ToUpper(f.ImpactLevel), chain, f.Depth, f.RiskScore, testStatus); err != nil {
			return err
		}
	}

	if len(out.Files) > textAffectedFileLimit {
		remaining := len(out.Files) - textAffectedFileLimit
		if _, err := fmt.Fprintf(r.out, "  ... and %d more affected files\n", remaining); err != nil {
			return err
		}
	}

	return nil
}

// jsonRenderer writes JSON output to stdout.
type jsonRenderer struct {
	out io.Writer
}

func (r *jsonRenderer) Render(ctx context.Context, out Output) error {
	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// mermaidRenderer writes a Mermaid flowchart to stdout.
type mermaidRenderer struct {
	out io.Writer
}

func (r *mermaidRenderer) Render(ctx context.Context, out Output) error {
	if _, err := fmt.Fprintln(r.out, "graph TD"); err != nil {
		return err
	}

	// Source node with risk info.
	srcNode := mermaidID(out.Source.Path)
	srcLabel := mermaidLabel(out.Source.Path, out.Source.RiskBand, out.Source.RiskScore)
	if _, err := fmt.Fprintf(r.out, "    %s[%q]\n", srcNode, srcLabel); err != nil {
		return err
	}

	// Style source node.
	if _, err := fmt.Fprintf(r.out, "    class %s source;\n", srcNode); err != nil {
		return err
	}

	for i, f := range out.Files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		nodeID := fmt.Sprintf("n%d", i)
		label := mermaidLabel(f.Path, f.RiskBand, f.RiskScore)
		if _, err := fmt.Fprintf(r.out, "    %s[%q]\n", nodeID, label); err != nil {
			return err
		}

		// Edge from source (direct) or from parent (indirect).
		if f.ImpactLevel == "direct" || f.Depth == 1 {
			if _, err := fmt.Fprintf(r.out, "    %s --> %s\n", srcNode, nodeID); err != nil {
				return err
			}
		} else {
			// For indirect files, chain through the previous node.
			parentIdx := i - 1
			if parentIdx >= 0 {
				parentID := fmt.Sprintf("n%d", parentIdx)
				if _, err := fmt.Fprintf(r.out, "    %s -.-> %s\n", parentID, nodeID); err != nil {
					return err
				}
			}
		}

		// Style by risk band.
		switch f.RiskBand {
		case "high":
			if _, err := fmt.Fprintf(r.out, "    class %s riskHigh;\n", nodeID); err != nil {
				return err
			}
		case "medium":
			if _, err := fmt.Fprintf(r.out, "    class %s riskMedium;\n", nodeID); err != nil {
				return err
			}
		case "low":
			if _, err := fmt.Fprintf(r.out, "    class %s riskLow;\n", nodeID); err != nil {
				return err
			}
		}
	}

	// ClassDef for styling.
	if _, err := fmt.Fprintln(r.out, "    classDef source fill:#4a90d9,stroke:#333,color:#fff"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(r.out, "    classDef riskHigh fill:#e74c3c,stroke:#333,color:#fff"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(r.out, "    classDef riskMedium fill:#f39c12,stroke:#333,color:#fff"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(r.out, "    classDef riskLow fill:#27ae60,stroke:#333,color:#fff"); err != nil {
		return err
	}

	return nil
}

// mermaidID converts a file path to a valid Mermaid node ID.
func mermaidID(path string) string {
	return strings.ReplaceAll(strings.ReplaceAll(path, "/", "_"), ".", "_")
}

// mermaidLabel formats a node label with path and risk info.
func mermaidLabel(path, riskBand string, riskScore int) string {
	if riskBand != "" {
		return fmt.Sprintf("%s\\nrisk:%d (%s)", path, riskScore, riskBand)
	}
	return path
}

// tuiRenderer runs an interactive Bubble Tea TUI for browsing affected files.
type tuiRenderer struct {
	out     io.Writer
	noColor bool
}

func (r *tuiRenderer) Render(ctx context.Context, out Output) error {
	tuiOut := tui.TUIOutput{
		Title: fmt.Sprintf("rippl — %s", out.Source.Path),
		Files: make([]tui.FileEntry, 0, len(out.Files)),
	}
	for _, f := range out.Files {
		tuiOut.Files = append(tuiOut.Files, tui.FileEntry{
			Path:        f.Path,
			ImpactLevel: f.ImpactLevel,
			Depth:       f.Depth,
			RiskScore:   f.RiskScore,
			Coverage:    f.Coverage,
			HasTestFile: f.HasTestFile,
			Chain:       f.Chain,
			Reason:      f.Reason,
		})
	}
	return tui.Run(ctx, tuiOut, r.noColor)
}

// Interface satisfaction checks.
var (
	_ Renderer = (*jsonRenderer)(nil)
	_ Renderer = (*mermaidRenderer)(nil)
	_ Renderer = (*tuiRenderer)(nil)
)
