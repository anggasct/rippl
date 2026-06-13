package tui

import (
	"testing"
)

func TestRiskBand(t *testing.T) {
	t.Parallel()
	cases := []struct {
		score int
		want  string
	}{
		{0, "minimal"},
		{24, "minimal"},
		{25, "low"},
		{49, "low"},
		{50, "medium"},
		{74, "medium"},
		{75, "high"},
		{100, "high"},
	}
	for _, tc := range cases {
		got := riskBand(tc.score)
		if got != tc.want {
			t.Errorf("riskBand(%d) = %q, want %q", tc.score, got, tc.want)
		}
	}
}

func TestRiskColor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		band    string
		noColor bool
		want    string
	}{
		{"high", false, "1"},
		{"medium", false, "3"},
		{"low", false, "2"},
		{"minimal", false, "8"},
		{"high", true, ""},
		{"medium", true, ""},
	}
	for _, tc := range cases {
		got := riskColor(tc.band, tc.noColor)
		if got != tc.want {
			t.Errorf("riskColor(%q, %v) = %q, want %q", tc.band, tc.noColor, got, tc.want)
		}
	}
}

func TestYesNo(t *testing.T) {
	t.Parallel()
	if yesNo(true) != "yes" {
		t.Errorf("yesNo(true) = %q, want %q", yesNo(true), "yes")
	}
	if yesNo(false) != "no" {
		t.Errorf("yesNo(false) = %q, want %q", yesNo(false), "no")
	}
}

func TestPadRight(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s    string
		w    int
		want string
	}{
		{"hi", 8, "hi      "},
		{"hello", 5, "hello"},
		{"hello world", 5, "hello world"},
	}
	for _, tc := range cases {
		got := padRight(tc.s, tc.w)
		if got != tc.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tc.s, tc.w, got, tc.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"hello world", 5, "he..."},
		{"hello world", 3, "hel"},
		{"hello world", 2, "he"},
		{"", 5, ""},
	}
	for _, tc := range cases {
		got := truncate(tc.s, tc.maxLen)
		if got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.s, tc.maxLen, got, tc.want)
		}
	}
}

func TestNewModel(t *testing.T) {
	t.Parallel()
	out := TUIOutput{
		Title: "test",
		Files: []FileEntry{
			{Path: "a.go", ImpactLevel: "direct", RiskScore: 80},
			{Path: "b.go", ImpactLevel: "indirect", RiskScore: 30},
		},
	}
	m := NewModel(out, false)
	if len(m.files) != 2 {
		t.Fatalf("NewModel files len = %d, want 2", len(m.files))
	}
	if m.cursor != 0 {
		t.Errorf("NewModel cursor = %d, want 0", m.cursor)
	}
	if m.showDetail {
		t.Error("NewModel showDetail should be false")
	}
}

func TestModelGroupFiles(t *testing.T) {
	t.Parallel()
	m := NewModel(TUIOutput{
		Files: []FileEntry{
			{Path: "a.go", ImpactLevel: "direct"},
			{Path: "b.go", ImpactLevel: "indirect"},
			{Path: "c.go", ImpactLevel: "direct"},
		},
	}, false)
	direct, indirect := m.groupFiles()
	if len(direct) != 2 {
		t.Errorf("direct count = %d, want 2", len(direct))
	}
	if len(indirect) != 1 {
		t.Errorf("indirect count = %d, want 1", len(indirect))
	}
}

func TestModelGlobalIndex(t *testing.T) {
	t.Parallel()
	m := NewModel(TUIOutput{
		Files: []FileEntry{
			{Path: "a.go"},
			{Path: "b.go"},
			{Path: "c.go"},
		},
	}, false)
	if m.globalIndex(m.files[1]) != 1 {
		t.Errorf("globalIndex(b.go) = %d, want 1", m.globalIndex(m.files[1]))
	}
}

func TestModelVisibleLines(t *testing.T) {
	t.Parallel()
	m := NewModel(TUIOutput{}, false)
	m.height = 20
	if m.visibleLines() != 14 {
		t.Errorf("visibleLines() = %d, want 14", m.visibleLines())
	}
	m.height = 3
	if m.visibleLines() != 10 {
		t.Errorf("visibleLines() = %d, want 10 (default)", m.visibleLines())
	}
}
