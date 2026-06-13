package render

import "context"

// Renderer renders analysis output to a specific format.
//
// Each renderer implements this interface for one of the supported formats
// (text, json, mermaid, tui). The command layer calls Render with the
// prepared Output and writes the result to stdout.
type Renderer interface {
	// Render writes the formatted output for the given result.
	// The context allows cancellation for long-running renderers (e.g. TUI).
	Render(ctx context.Context, out Output) error
}
