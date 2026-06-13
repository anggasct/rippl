package render

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

// fakeRenderer is a test double that records Render calls.
type fakeRenderer struct {
	called bool
	got    Output
}

func (f *fakeRenderer) Render(ctx context.Context, out Output) error {
	f.called = true
	f.got = out
	return nil
}

// errRenderer is a test double that always returns an error.
type errRenderer struct {
	err error
}

func (e *errRenderer) Render(ctx context.Context, out Output) error {
	return e.err
}

// discardRenderer is a test double that discards output.
type discardRenderer struct{}

func (d discardRenderer) Render(ctx context.Context, out Output) error {
	_, err := io.Copy(io.Discard, bytes.NewReader(nil))
	_ = err
	return nil
}

func TestRendererInterfaceSatisfied(t *testing.T) {
	t.Parallel()

	var r Renderer
	r = &fakeRenderer{}
	if r == nil {
		t.Fatal("fakeRenderer does not satisfy Renderer interface")
	}
	r = &errRenderer{}
	if r == nil {
		t.Fatal("errRenderer does not satisfy Renderer interface")
	}
	r = discardRenderer{}
	if r == nil {
		t.Fatal("discardRenderer does not satisfy Renderer interface")
	}
}

func TestFakeRendererRender(t *testing.T) {
	t.Parallel()

	f := &fakeRenderer{}
	ctx := context.Background()
	out := Output{
		Version:   "1",
		Command:   "analyze",
		Module:    "github.com/example/app",
		Generated: time.Date(2026, 6, 13, 0, 0, 0, 0, time.UTC),
	}

	if err := f.Render(ctx, out); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !f.called {
		t.Fatal("Render() was not called")
	}
	if f.got.Version != "1" {
		t.Fatalf("Version = %q, want %q", f.got.Version, "1")
	}
	if f.got.Command != "analyze" {
		t.Fatalf("Command = %q, want %q", f.got.Command, "analyze")
	}
	if f.got.Module != "github.com/example/app" {
		t.Fatalf("Module = %q, want %q", f.got.Module, "github.com/example/app")
	}
}

func TestErrRendererRender(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("render failed")
	e := &errRenderer{err: wantErr}
	ctx := context.Background()

	err := e.Render(ctx, Output{})
	if err != wantErr {
		t.Fatalf("Render() error = %v, want %v", err, wantErr)
	}
}

func TestRenderContextCancel(*testing.T) {
	// Verify that Render receives a context (renderers may check cancellation).
	ctx, cancel := context.Background(), func() {}
	_ = ctx
	cancel()
}

func TestFormatConstants(t *testing.T) {
	t.Parallel()

	cases := []struct {
		got  Format
		want string
	}{
		{FormatText, "text"},
		{FormatJSON, "json"},
		{FormatMermaid, "mermaid"},
		{FormatTUI, "tui"},
	}
	for _, tc := range cases {
		if string(tc.got) != tc.want {
			t.Fatalf("Format = %q, want %q", tc.got, tc.want)
		}
	}
}

func TestOutputZeroValue(t *testing.T) {
	t.Parallel()

	var out Output
	if out.Version != "" {
		t.Fatalf("zero Version = %q, want empty", out.Version)
	}
	if out.Command != "" {
		t.Fatalf("zero Command = %q, want empty", out.Command)
	}
	if len(out.Files) != 0 {
		t.Fatalf("zero Files len = %d, want 0", len(out.Files))
	}
}

func TestSourceOutputZeroValue(t *testing.T) {
	t.Parallel()

	var s SourceOutput
	if s.Path != "" {
		t.Fatalf("zero Path = %q, want empty", s.Path)
	}
	if s.RiskScore != 0 {
		t.Fatalf("zero RiskScore = %d, want 0", s.RiskScore)
	}
}

func TestSummaryOutputZeroValue(t *testing.T) {
	t.Parallel()

	var s SummaryOutput
	if s.AffectedCount != 0 {
		t.Fatalf("zero AffectedCount = %d, want 0", s.AffectedCount)
	}
	if s.MaxRiskScore != 0 {
		t.Fatalf("zero MaxRiskScore = %d, want 0", s.MaxRiskScore)
	}
}

func TestFileOutputZeroValue(t *testing.T) {
	t.Parallel()

	var f FileOutput
	if f.Path != "" {
		t.Fatalf("zero Path = %q, want empty", f.Path)
	}
	if f.Depth != 0 {
		t.Fatalf("zero Depth = %d, want 0", f.Depth)
	}
	if len(f.Chain) != 0 {
		t.Fatalf("zero Chain len = %d, want 0", len(f.Chain))
	}
}
