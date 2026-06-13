package render

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// NewRenderer creates a Renderer for the given format string.
// Supported formats: text, json, mermaid, tui.
// Falls back to text for unknown formats.
// The renderer writes to os.Stdout.
func NewRenderer(format string) (Renderer, error) {
	return NewRendererWithWriter(format, os.Stdout)
}

// NewRendererWithWriter creates a Renderer for the given format string that writes to w.
func NewRendererWithWriter(format string, w io.Writer) (Renderer, error) {
	switch Format(strings.ToLower(format)) {
	case FormatJSON:
		return &jsonRenderer{out: w}, nil
	case FormatMermaid:
		return &mermaidRenderer{out: w}, nil
	case FormatTUI:
		return &tuiRenderer{out: w}, nil
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

	for i, f := range out.Files {
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
	fmt.Fprintln(r.out, "flowchart TD")
	fmt.Fprintf(r.out, "    source[%s]\n", mermaidNode(out.Source.Path))

	for i, f := range out.Files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		nodeID := fmt.Sprintf("n%d", i)
		fmt.Fprintf(r.out, "    %s[%s]\n", nodeID, mermaidNode(f.Path))
		fmt.Fprintf(r.out, "    source --> %s\n", nodeID)
	}

	return nil
}

func mermaidNode(path string) string {
	return strings.ReplaceAll(path, ".", "_")
}

// tuiRenderer is a placeholder for the TUI renderer (WP-02).
type tuiRenderer struct {
	out io.Writer
}

func (r *tuiRenderer) Render(ctx context.Context, out Output) error {
	return ErrTUINotImplemented
}

// ErrTUINotImplemented is returned when the selected format has no concrete renderer yet.
var ErrTUINotImplemented = fmt.Errorf("tui renderer not implemented")

// Interface satisfaction checks.
var (
	_ Renderer = (*jsonRenderer)(nil)
	_ Renderer = (*mermaidRenderer)(nil)
	_ Renderer = (*tuiRenderer)(nil)
)
