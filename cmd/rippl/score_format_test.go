package main

import "testing"

func TestSignalDisplayName(t *testing.T) {
	t.Parallel()
	if got := signalDisplayName("test_coverage"); got != "Coverage risk" {
		t.Errorf("signalDisplayName(test_coverage) = %q, want %q", got, "Coverage risk")
	}
}
