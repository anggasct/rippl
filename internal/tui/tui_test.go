package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

func TestBuildListRowsIncludesHeaders(t *testing.T) {
	t.Parallel()
	m := NewModel(fixture48(), false)
	rows := m.buildListRows()

	var headers int
	var files int
	for _, row := range rows {
		switch row.kind {
		case rowHeader:
			headers++
		case rowFile:
			files++
		}
	}
	if headers != 2 {
		t.Fatalf("header count = %d, want 2", headers)
	}
	if files != 48 {
		t.Fatalf("file row count = %d, want 48", files)
	}
}

func TestModelNavigateDown(t *testing.T) {
	m := NewModel(fixture48(), false)
	for range 47 {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = next.(Model)
	}
	if m.cursor != 47 {
		t.Fatalf("cursor = %d, want 47", m.cursor)
	}
}

func TestModelNavigateUp(t *testing.T) {
	m := NewModel(fixture48(), false)
	for range 6 {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = next.(Model)
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = next.(Model)
	if m.cursor != 5 {
		t.Fatalf("cursor = %d, want 5", m.cursor)
	}
}

func TestModelNavigateInDetail(t *testing.T) {
	m := NewModel(fixture48(), false)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = next.(Model)
	if !m.showDetail {
		t.Fatal("showDetail should be true after d")
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 in detail mode", m.cursor)
	}
}

func TestListViewShowsCursorOnSelectedFile(t *testing.T) {
	m := NewModel(fixture48(), true)
	m.height = 30
	for range 31 {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = next.(Model)
	}
	view := m.listView()
	if !strings.Contains(view, "> ") {
		t.Fatalf("list view missing cursor marker:\n%s", view)
	}
	if !strings.Contains(view, "f31.go") {
		t.Fatalf("list view should show selected file f31.go:\n%s", view)
	}
}

func TestListViewShowsFilterNote(t *testing.T) {
	m := NewModel(TUIOutput{
		Files:      []FileEntry{{Path: "a.go", RiskScore: 50}},
		FilterNote: "Showing 1 of 10 affected files (--top 1)",
	}, true)
	m.height = 20
	view := m.listView()
	if !strings.Contains(view, "Showing 1 of 10 affected files") {
		t.Fatalf("list view missing filter note:\n%s", view)
	}
}

func fixture48() TUIOutput {
	files := make([]FileEntry, 48)
	for i := range files {
		level := "direct"
		if i >= 31 {
			level = "indirect"
		}
		files[i] = FileEntry{
			Path:        "f" + itoa(i) + ".go",
			ImpactLevel: level,
			RiskScore:   50,
		}
	}
	return TUIOutput{Files: files}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

func TestModelScrollViewport(t *testing.T) {
	m := NewModel(fixture48(), false)
	m.height = 20

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = next.(Model)
	if m.scrollOffset == 0 {
		t.Fatalf("scrollOffset = %d after PgDown, want > 0", m.scrollOffset)
	}

	start := m.scrollOffset
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = next.(Model)
	if m.scrollOffset >= start {
		t.Fatalf("scrollOffset = %d after PgUp, want < %d", m.scrollOffset, start)
	}
}

func TestModelVisibleLines(t *testing.T) {
	t.Parallel()
	m := NewModel(TUIOutput{}, false)
	m.height = 20
	if m.visibleLines() != 16 {
		t.Errorf("visibleLines() = %d, want 16", m.visibleLines())
	}
	m.height = 3
	if m.visibleLines() != 10 {
		t.Errorf("visibleLines() = %d, want 10 (default)", m.visibleLines())
	}
}
