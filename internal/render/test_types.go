package render

import "time"

// TestSchemaVersion is the test JSON output schema version.
const TestSchemaVersion = "1"

// TestOutput is the JSON contract for rippl test --format json.
type TestOutput struct {
	Version         string        `json:"version"`
	Command         string        `json:"command"`
	Module          string        `json:"module"`
	SourceFile      string        `json:"source_file"`
	Generated       time.Time     `json:"generated_at"`
	PackagesRun     []string      `json:"packages_run"`
	PackagesSkipped []string      `json:"packages_skipped"`
	DurationMS      int64         `json:"duration_ms"`
	ExitCode        int           `json:"exit_code"`
	Passed          bool          `json:"passed"`
	Failures        []TestFailure `json:"failures"`
}

// TestFailure describes a failed package test run.
type TestFailure struct {
	Package string `json:"package"`
	Summary string `json:"summary"`
}

// TestRunResult holds executor output for BuildTestOutput.
type TestRunResult struct {
	Module          string
	SourceFile      string
	PackagesRun     []string
	PackagesSkipped []string
	DurationMS      int64
	ExitCode        int
	Failures        []TestFailure
}

// BuildTestOutput maps a test run result to the test JSON contract.
func BuildTestOutput(result TestRunResult, generated time.Time) TestOutput {
	packagesRun := result.PackagesRun
	if packagesRun == nil {
		packagesRun = []string{}
	}
	packagesSkipped := result.PackagesSkipped
	if packagesSkipped == nil {
		packagesSkipped = []string{}
	}
	failures := result.Failures
	if failures == nil {
		failures = []TestFailure{}
	}

	return TestOutput{
		Version:         TestSchemaVersion,
		Command:         "test",
		Module:          result.Module,
		SourceFile:      result.SourceFile,
		Generated:       generated,
		PackagesRun:     packagesRun,
		PackagesSkipped: packagesSkipped,
		DurationMS:      result.DurationMS,
		ExitCode:        result.ExitCode,
		Passed:          result.ExitCode == 0,
		Failures:        failures,
	}
}
