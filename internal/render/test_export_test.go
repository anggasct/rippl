package render

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestTestJSONSchemaStability(t *testing.T) {
	t.Parallel()

	result := TestRunResult{
		Module:          "github.com/example/app",
		SourceFile:      "internal/parser/analysis.go",
		PackagesRun:     []string{"./internal/parser", "./internal/graph"},
		PackagesSkipped: []string{"./internal/handler"},
		DurationMS:      16234,
		ExitCode:        0,
		Failures:        []TestFailure{},
	}
	generated := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	out := BuildTestOutput(result, generated)

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

	requiredTop := []string{
		"version", "command", "module", "source_file", "generated_at",
		"packages_run", "packages_skipped", "duration_ms", "exit_code", "passed", "failures",
	}
	for _, key := range requiredTop {
		if _, ok := doc[key]; !ok {
			t.Errorf("missing top-level key %q", key)
		}
	}

	if doc["version"] != TestSchemaVersion {
		t.Errorf("version = %v, want %q", doc["version"], TestSchemaVersion)
	}
	if doc["command"] != "test" {
		t.Errorf("command = %v, want test", doc["command"])
	}
	if doc["passed"] != true {
		t.Errorf("passed = %v, want true", doc["passed"])
	}

	if _, ok := doc["packages_run"].([]interface{}); !ok {
		t.Error("packages_run is not an array")
	}
	if _, ok := doc["packages_skipped"].([]interface{}); !ok {
		t.Error("packages_skipped is not an array")
	}
}

func TestTestJSONFailureShape(t *testing.T) {
	t.Parallel()

	result := TestRunResult{
		Module:      "example.com/mod",
		SourceFile:  "pkg/a.go",
		PackagesRun: []string{"./pkg/a"},
		ExitCode:    1,
		Failures: []TestFailure{
			{Package: "./pkg/a", Summary: "FAIL: TestFoo"},
		},
	}
	out := BuildTestOutput(result, time.Now().UTC())

	if out.Passed {
		t.Fatal("Passed = true, want false for exit code 1")
	}
	if len(out.Failures) != 1 {
		t.Fatalf("Failures len = %d, want 1", len(out.Failures))
	}
	if out.Failures[0].Package != "./pkg/a" || out.Failures[0].Summary != "FAIL: TestFoo" {
		t.Fatalf("Failures[0] = %+v, want package ./pkg/a", out.Failures[0])
	}
}

func TestBuildTestOutputEmptySlices(t *testing.T) {
	t.Parallel()

	out := BuildTestOutput(TestRunResult{
		Module:     "example.com/mod",
		SourceFile: "main.go",
		ExitCode:   0,
	}, time.Now().UTC())

	if out.PackagesRun == nil || len(out.PackagesRun) != 0 {
		t.Fatalf("PackagesRun = %v, want empty slice", out.PackagesRun)
	}
	if out.PackagesSkipped == nil || len(out.PackagesSkipped) != 0 {
		t.Fatalf("PackagesSkipped = %v, want empty slice", out.PackagesSkipped)
	}
	if out.Failures == nil || len(out.Failures) != 0 {
		t.Fatalf("Failures = %v, want empty slice", out.Failures)
	}
}
