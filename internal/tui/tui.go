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

// Model is the Bubble Tea model for the interactive TUI.
type Model struct {
	files        []FileEntry
	cursor       int
	showDetail   bool
	scrollOffset int
	quitting     bool
	noColor      bool
	width        int
	height       int
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
	return nil
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
			if m.showDetail {
				break
			}
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}

		case "down", "j":
			if m.showDetail {
				break
			}
			if m.cursor < len(m.files)-1 {
				m.cursor++
				if m.cursor >= m.scrollOffset+m.visibleLines() {
					m.scrollOffset++
				}
			}

		case "d":
			if len(m.files) > 0 {
				m.showDetail = !m.showDetail
			}

		case "esc":
			if m.showDetail {
				m.showDetail = false
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
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

	var b strings.Builder

	// Header
	headerStyle := m.styleBoldUnderline()
	b.WriteString(headerStyle.Render("Affected Files"))
	b.WriteString("\n\n")

	// Group files by impact level
	direct, indirect := m.groupFiles()

	if len(direct) > 0 {
		b.WriteString(m.renderGroup("Direct Impact", direct, riskColor("high", m.noColor)))
		b.WriteString("\n")
	}

	if len(indirect) > 0 {
		b.WriteString(m.renderGroup("Indirect Impact", indirect, riskColor("medium", m.noColor)))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.styleFaint().Render("  up/down: navigate  d: detail  q: quit"))

	return b.String()
}

// detailView renders the detail panel for the selected file.
func (m Model) detailView() string {
	if len(m.files) == 0 || m.cursor >= len(m.files) {
		return "No file selected.\n"
	}

	f := m.files[m.cursor]
	var b strings.Builder

	b.WriteString(m.styleBold().Render(fmt.Sprintf("Detail: %s", f.Path)))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", 50))
	b.WriteString("\n\n")

	band := riskBand(f.RiskScore)
	bandColor := riskColor(band, m.noColor)
	bandStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(bandColor))
	if m.noColor {
		bandStyle = lipgloss.NewStyle()
	}

	fmt.Fprintf(&b, "  Impact:    %s (depth %d)\n", f.ImpactLevel, f.Depth)
	fmt.Fprintf(&b, "  Risk:      %d/100\n", f.RiskScore)
	fmt.Fprintf(&b, "  Band:      %s\n", bandStyle.Render(band))
	fmt.Fprintf(&b, "  Coverage:  %.1f%%\n", f.Coverage)
	fmt.Fprintf(&b, "  Test file: %s\n", yesNo(f.HasTestFile))
	fmt.Fprintf(&b, "  Reason:    %s\n", f.Reason)

	if len(f.Chain) > 0 {
		fmt.Fprintf(&b, "  Chain:     %s\n", strings.Join(f.Chain, " -> "))
	}

	b.WriteString("\n")
	b.WriteString(m.styleFaint().Render("  d/esc: close detail  q: quit"))

	return b.String()
}

// groupFiles splits files into direct and indirect groups, preserving order.
func (m Model) groupFiles() (direct, indirect []FileEntry) {
	for _, f := range m.files {
		switch f.ImpactLevel {
		case "direct":
			direct = append(direct, f)
		case "indirect":
			indirect = append(indirect, f)
		}
	}
	return
}

// renderGroup renders a group of files under a colored heading.
func (m Model) renderGroup(title string, files []FileEntry, color string) string {
	var b strings.Builder

	groupStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color))
	if m.noColor {
		groupStyle = lipgloss.NewStyle()
	}
	b.WriteString(groupStyle.Render(title))
	b.WriteString("\n")

	// Calculate visible range based on viewport
	visLines := m.visibleLines()
	end := m.scrollOffset + visLines
	if end > len(files) {
		end = len(files)
	}
	start := m.scrollOffset
	if start > end {
		start = end
	}

	for i := start; i < end; i++ {
		f := files[i]
		globalIdx := m.globalIndex(f)

		cursor := "  "
		if m.cursor == globalIdx {
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

		line := fmt.Sprintf("%s[%s] %s (risk=%d)%s\n",
			cursor,
			riskStyle.Render(padRight(band, 8)),
			truncate(f.Path, 40),
			f.RiskScore,
			testStatus,
		)
		b.WriteString(line)
	}

	return b.String()
}

// globalIndex returns the index of target in m.files.
// Note: uses path comparison; duplicate paths are impossible by construction
// since the engine produces unique file paths per analysis run.
func (m Model) globalIndex(target FileEntry) int {
	for i, f := range m.files {
		if f.Path == target.Path {
			return i
		}
	}
	return 0
}

// visibleLines returns how many file lines can fit in the viewport.
func (m Model) visibleLines() int {
	if m.height > 6 {
		return m.height - 6
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
	)
	_, err := p.Run()
	return err
}
