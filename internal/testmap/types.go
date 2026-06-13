package testmap

type CoverageStatus string

const (
	StatusNoTest  CoverageStatus = "no_test"
	StatusUnknown CoverageStatus = "unknown"
	StatusPercent CoverageStatus = "percent"
)

type FileCoverage struct {
	Path        string
	HasTestFile bool
	TestFiles   []string
	Status      CoverageStatus
	CoveragePct *float64
}
