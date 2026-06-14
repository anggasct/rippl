package render

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/anggasct/rippl/internal/scorer"
)

func TestScoreJSONSchemaStability(t *testing.T) {
	t.Parallel()

	result := scorer.FileRisk{
		Path:  "internal/auth/jwt.go",
		Score: 71,
		Band:  scorer.BandMedium,
		Signals: []scorer.Signal{
			{Name: "bug_fix_ratio", Raw: "3 bug-fix commits in 10 commits", Normalized: 30, Weight: 25, Contribution: 7.5},
			{Name: "fan_out", Raw: "2 dependent files", Normalized: 30, Weight: 20, Contribution: 6},
			{Name: "churn_rate", Raw: "150 lines changed", Normalized: 100, Weight: 15, Contribution: 15},
			{Name: "author_count", Raw: "3 unique authors", Normalized: 40, Weight: 15, Contribution: 6},
			{Name: "stale_age", Raw: "45 days since last change", Normalized: 25, Weight: 10, Contribution: 2.5},
			{Name: "test_coverage", Raw: "54.8% coverage", Normalized: 45, Weight: 15, Contribution: 6.75},
		},
	}

	generated := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	out := BuildScoreOutput("github.com/example/app", "internal/auth/jwt.go", result, generated)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	requiredTop := []string{"version", "command", "module", "source_file", "generated_at", "score", "risk_band", "signals"}
	for _, key := range requiredTop {
		if _, ok := doc[key]; !ok {
			t.Errorf("missing top-level key %q", key)
		}
	}

	if doc["version"] != ScoreSchemaVersion {
		t.Errorf("version = %v, want %q", doc["version"], ScoreSchemaVersion)
	}
	if doc["command"] != "score" {
		t.Errorf("command = %v, want score", doc["command"])
	}

	signals, ok := doc["signals"].([]interface{})
	if !ok {
		t.Fatal("signals is not an array")
	}
	if len(signals) != 6 {
		t.Fatalf("signals length = %d, want 6", len(signals))
	}

	requiredSignal := []string{"name", "label", "raw_value", "weight", "contribution", "interpretation"}
	for i, sig := range signals {
		m, ok := sig.(map[string]interface{})
		if !ok {
			t.Fatalf("signals[%d] is not an object", i)
		}
		for _, key := range requiredSignal {
			if _, ok := m[key]; !ok {
				t.Errorf("signals[%d] missing key %q", i, key)
			}
		}
	}

	// Coverage risk label (CAP-201 IMP-001).
	last := signals[5].(map[string]interface{})
	if last["label"] != "Coverage risk" {
		t.Errorf("test_coverage label = %v, want Coverage risk", last["label"])
	}
}

func TestAuthorCountFromNormalized(t *testing.T) {
	t.Parallel()

	cases := []struct {
		norm int
		want float64
	}{
		{0, 1},
		{20, 2},
		{40, 3},
		{80, 5},
		{100, 6},
	}

	for _, tc := range cases {
		got := authorCountFromNormalized(tc.norm)
		if got != tc.want {
			t.Errorf("authorCountFromNormalized(%d) = %v, want %v", tc.norm, got, tc.want)
		}
	}
}

func TestRawValueForSignalTestCoverageUnknown(t *testing.T) {
	t.Parallel()

	s := scorer.Signal{Name: "test_coverage", Raw: "coverage unknown", Normalized: 50}
	if got := rawValueForSignal(s); got != 0 {
		t.Errorf("rawValueForSignal(coverage unknown) = %v, want 0", got)
	}
}

func TestScoreJSONEncode(t *testing.T) {
	t.Parallel()

	out := BuildScoreOutput("example.com/mod", "pkg/a.go", scorer.FileRisk{
		Score: 10,
		Band:  scorer.BandMinimal,
		Signals: []scorer.Signal{
			{Name: "bug_fix_ratio", Raw: "none", Normalized: 0, Weight: 25, Contribution: 0},
		},
	}, time.Now().UTC())

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if !json.Valid(buf.Bytes()) {
		t.Fatalf("output is not valid JSON: %q", buf.String())
	}
}
