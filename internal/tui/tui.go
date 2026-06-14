// Package tui provides a Bubble Tea interactive TUI for browsing affected files.
package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileEntry represents a single affected file for the TUI.
type FileEntry struct {
	Path        string
	ImpactLevel string
	Depth       int
	RiskScore   int
	Coverage    float64
	HasTestFile bool
	Chain       []string
	Reason      string
}

// TUIOutput is the data contract for the TUI renderer.
type TUIOutput struct {
	Title string
	Files []FileEntry
}

type listRowKind int

const (
	rowHeader listRowKind = iota
	rowFile
)

type listRow struct {
	kind    listRowKind
	title   string
	color   string
	fileIdx int
}

// Model is the Bubble Tea model for the interactive TUI.
type Model struct {
	files              []FileEntry
	cursor             int
	showDetail         bool
	scrollOffset       int
	detailScrollOffset int
	quitting           bool
	noColor            bool
	width              int
	height             int
}

// NewModel creates a new TUI model from the given output.
func NewModel(out TUIOutput, noColor bool) Model {
	return Model{
		files:   out.Files,
		cursor:  0,
		noColor: noColor,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.WindowSize()
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			m.moveCursor(-1)

		case "down", "j":
			m.moveCursor(1)

		case "pgup", "ctrl+u":
			m.scrollViewport(-m.pageScroll())

		case "pgdown", "ctrl+d":
			m.scrollViewport(m.pageScroll())

		case "home", "g":
			if m.showDetail {
				m.detailScrollOffset = 0
			} else {
				m.scrollOffset = 0
			}

		case "end", "G":
			if m.showDetail {
				m.detailScrollOffset = m.maxDetailScroll()
			} else {
				m.scrollOffset = m.maxListScroll()
			}

		case "d":
			if len(m.files) > 0 {
				m.showDetail = !m.showDetail
				m.detailScrollOffset = 0
			}

		case "esc":
			if m.showDetail {
				m.showDetail = false
				m.detailScrollOffset = 0
			}
		}

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonWheelUp {
			m.scrollViewport(-3)
		}
		if msg.Button == tea.MouseButtonWheelDown {
			m.scrollViewport(3)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureCursorVisible()
	}

	return m, nil
}

func (m *Model) moveCursor(delta int) {
	if len(m.files) == 0 {
		return
	}
	next := m.cursor + delta
	if next < 0 || next >= len(m.files) {
		return
	}
	m.cursor = next
	if m.showDetail {
		m.detailScrollOffset = 0
	} else {
		m.ensureCursorVisible()
	}
}

func (m *Model) pageScroll() int {
	vis := m.visibleLines()
	if vis < 1 {
		return 1
	}
	return vis
}

func (m *Model) maxListScroll() int {
	rows := m.buildListRows()
	vis := m.visibleLines()
	maxOffset := len(rows) - vis
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func (m *Model) scrollViewport(delta int) {
	if m.showDetail {
		m.detailScrollOffset += delta
		max := m.maxDetailScroll()
		if m.detailScrollOffset > max {
			m.detailScrollOffset = max
		}
		if m.detailScrollOffset < 0 {
			m.detailScrollOffset = 0
		}
		return
	}

	m.scrollOffset += delta
	max := m.maxListScroll()
	if m.scrollOffset > max {
		m.scrollOffset = max
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m *Model) maxDetailScroll() int {
	lines := m.detailLineCount()
	vis := m.visibleLines()
	maxOffset := lines - vis
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func (m *Model) detailLineCount() int {
	if len(m.files) == 0 || m.cursor >= len(m.files) {
		return 0
	}
	f := m.files[m.cursor]
	lines := 8 // title, rule, blank, impact, risk, band, coverage, test
	if f.Reason != "" {
		lines++
	}
	if len(f.Chain) > 0 {
		lines++
	}
	return lines
}

func (m *Model) buildListRows() []listRow {
	if len(m.files) == 0 {
		return nil
	}

	rows := make([]listRow, 0, len(m.files)+2)
	var currentGroup string

	for i, f := range m.files {
		group := impactGroupTitle(f.ImpactLevel)
		if group != currentGroup {
			color := riskColor("medium", m.noColor)
			if f.ImpactLevel == "direct" {
				color = riskColor("high", m.noColor)
			}
			rows = append(rows, listRow{
				kind:  rowHeader,
				title: group,
				color: color,
			})
			currentGroup = group
		}
		rows = append(rows, listRow{
			kind:    rowFile,
			fileIdx: i,
		})
	}
	return rows
}

func impactGroupTitle(level string) string {
	switch level {
	case "direct":
		return "Direct Impact"
	case "indirect":
		return "Indirect Impact"
	default:
		return strings.ToUpper(level)
	}
}

func (m *Model) cursorRowIndex(rows []listRow) int {
	for i, row := range rows {
		if row.kind == rowFile && row.fileIdx == m.cursor {
			return i
		}
	}
	return 0
}

func (m *Model) ensureCursorVisible() {
	rows := m.buildListRows()
	if len(rows) == 0 {
		m.scrollOffset = 0
		return
	}

	cursorRow := m.cursorRowIndex(rows)
	vis := m.visibleLines()
	if vis < 1 {
		vis = 1
	}

	if cursorRow < m.scrollOffset {
		m.scrollOffset = cursorRow
	}
	if cursorRow >= m.scrollOffset+vis {
		m.scrollOffset = cursorRow - vis + 1
	}

	maxOffset := len(rows) - vis
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.showDetail {
		return m.detailView()
	}

	return m.listView()
}

// listView renders the grouped file list.
func (m Model) listView() string {
	if len(m.files) == 0 {
		return "No affected files.\n"
	}

	rows := m.buildListRows()

	var b strings.Builder

	headerStyle := m.styleBoldUnderline()
	b.WriteString(headerStyle.Render("Affected Files"))
	b.WriteString("\n\n")

	vis := m.visibleLines()
	end := m.scrollOffset + vis
	if end > len(rows) {
		end = len(rows)
	}

	for i := m.scrollOffset; i < end; i++ {
		row := rows[i]
		switch row.kind {
		case rowHeader:
			groupStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(row.color))
			if m.noColor {
				groupStyle = lipgloss.NewStyle()
			}
			b.WriteString(groupStyle.Render(row.title))
			b.WriteString("\n")
		case rowFile:
			b.WriteString(m.renderFileLine(m.files[row.fileIdx], row.fileIdx == m.cursor))
		}
	}

	b.WriteString("\n")
	footer := fmt.Sprintf("  %d/%d  ↑↓: navigate  PgUp/PgDn: scroll  d: detail  q: quit",
		m.cursor+1, len(m.files))
	b.WriteString(m.styleFaint().Render(footer))

	return b.String()
}

func (m Model) renderFileLine(f FileEntry, selected bool) string {
	cursor := "  "
	if selected {
		cursorStyle := lipgloss.NewStyle().Bold(true)
		if m.noColor {
			cursorStyle = lipgloss.NewStyle()
		}
		cursor = cursorStyle.Render("> ")
	}

	band := riskBand(f.RiskScore)
	riskColorName := riskColor(band, m.noColor)
	riskStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(riskColorName))
	if m.noColor {
		riskStyle = lipgloss.NewStyle()
	}

	testStatus := ""
	if !f.HasTestFile {
		testStatus = " [no test]"
	}

	return fmt.Sprintf("%s[%s] %s (risk=%d)%s\n",
		cursor,
		riskStyle.Render(padRight(band, 8)),
		truncate(f.Path, 40),
		f.RiskScore,
		testStatus,
	)
}

// detailView renders the detail panel for the selected file.
func (m Model) detailView() string {
	if len(m.files) == 0 || m.cursor >= len(m.files) {
		return "No file selected.\n"
	}

	f := m.files[m.cursor]
	lines := m.detailContentLines(f)
	vis := m.visibleLines()
	scroll := m.detailScrollOffset
	maxScroll := len(lines) - vis
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}

	end := scroll + vis
	if end > len(lines) {
		end = len(lines)
	}

	var b strings.Builder
	for i := scroll; i < end; i++ {
		b.WriteString(lines[i])
		b.WriteString("\n")
	}

	b.WriteString("\n")
	footer := fmt.Sprintf("  %d/%d  ↑↓: prev/next  PgUp/PgDn: scroll  d/esc: close  q: quit",
		m.cursor+1, len(m.files))
	b.WriteString(m.styleFaint().Render(footer))

	return b.String()
}

func (m Model) detailContentLines(f FileEntry) []string {
	band := riskBand(f.RiskScore)
	bandColor := riskColor(band, m.noColor)
	bandStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(bandColor))
	if m.noColor {
		bandStyle = lipgloss.NewStyle()
	}

	lines := []string{
		m.styleBold().Render(fmt.Sprintf("Detail: %s", f.Path)),
		strings.Repeat("-", 50),
		"",
		fmt.Sprintf("  Impact:    %s (depth %d)", f.ImpactLevel, f.Depth),
		fmt.Sprintf("  Risk:      %d/100", f.RiskScore),
		fmt.Sprintf("  Band:      %s", bandStyle.Render(band)),
		fmt.Sprintf("  Coverage:  %.1f%%", f.Coverage),
		fmt.Sprintf("  Test file: %s", yesNo(f.HasTestFile)),
		fmt.Sprintf("  Reason:    %s", f.Reason),
	}
	if len(f.Chain) > 0 {
		lines = append(lines, fmt.Sprintf("  Chain:     %s", strings.Join(f.Chain, " -> ")))
	}
	return lines
}

// visibleLines returns how many content rows fit below the list header and above the footer.
func (m Model) visibleLines() int {
	// Fixed chrome: title block (2 lines) + footer (2 lines).
	const chromeLines = 4
	if m.height > chromeLines+1 {
		return m.height - chromeLines
	}
	return 10
}

// riskBand returns the risk band string for a score.
func riskBand(score int) string {
	switch {
	case score >= 75:
		return "high"
	case score >= 50:
		return "medium"
	case score >= 25:
		return "low"
	default:
		return "minimal"
	}
}

// riskColor returns the lipgloss color name for a risk band.
func riskColor(band string, noColor bool) string {
	if noColor {
		return ""
	}
	switch band {
	case "high":
		return "1" // red
	case "medium":
		return "3" // yellow
	case "low":
		return "2" // green
	default:
		return "8" // gray (bright black)
	}
}

// styleBold returns a bold style.
func (m Model) styleBold() lipgloss.Style {
	if m.noColor {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Bold(true)
}

// styleBoldUnderline returns a bold+underline style.
func (m Model) styleBoldUnderline() lipgloss.Style {
	if m.noColor {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Bold(true).Underline(true)
}

// styleFaint returns a faint style.
func (m Model) styleFaint() lipgloss.Style {
	if m.noColor {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Faint(true)
}

// yesNo returns "yes" or "no".
func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

// padRight pads a string to the given width.
func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

// truncate truncates a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Run starts the interactive TUI with the given output.
// It blocks until the user quits.
// Note: tea.WithAltScreen() may leave the terminal in alt-screen mode if the
// process is killed (e.g. SIGKILL). This is a known Bubble Tea caveat —
// graceful quit (q / ctrl+c) restores the terminal normally.
func Run(ctx context.Context, out TUIOutput, noColor bool) error {
	p := tea.NewProgram(
		NewModel(out, noColor),
		tea.WithContext(ctx),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
