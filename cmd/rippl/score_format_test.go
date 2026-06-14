package main

import (
	"testing"

	"github.com/anggasct/rippl/internal/render"
)

func TestSignalDisplayName(t *testing.T) {
	t.Parallel()
	if got := render.SignalLabel("test_coverage"); got != "Coverage risk" {
		t.Errorf("SignalLabel(test_coverage) = %q, want %q", got, "Coverage risk")
	}
}
