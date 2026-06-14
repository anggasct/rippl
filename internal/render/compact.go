package render

// CompactAnalyzeOutput returns a copy of analyze output with chain arrays removed for agent/compact JSON.
func CompactAnalyzeOutput(out Output) Output {
	compact := out
	compact.Files = make([]FileOutput, len(out.Files))
	for i, f := range out.Files {
		f.Chain = nil
		compact.Files[i] = f
	}
	return compact
}

// IsStructuredFormat reports whether the format emits structured JSON (json or agent).
func IsStructuredFormat(format string) bool {
	switch Format(format) {
	case FormatJSON, FormatAgent:
		return true
	default:
		return false
	}
}
