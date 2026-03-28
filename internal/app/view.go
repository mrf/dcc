package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrf/dcc/internal/ui"
)

// View implements tea.Model
func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	// Calculate layout dimensions
	// Top row: 40% height for 3 panels
	// Middle row: 30% height for git
	// Bottom row: 20% height for stashes
	// Status bar: 3 lines

	statusBarHeight := 3
	topRowHeight := (m.Height - statusBarHeight) * 40 / 100
	middleRowHeight := (m.Height - statusBarHeight) * 30 / 100
	bottomRowHeight := m.Height - statusBarHeight - topRowHeight - middleRowHeight

	// Top row: Meetings (30%), PRs (40%), Ports (30%)
	meetingsWidth := m.Width * 30 / 100
	prsWidth := m.Width * 40 / 100
	portsWidth := m.Width - meetingsWidth - prsWidth

	// Render panels
	meetingsPanel := ui.RenderMeetingsPanel(
		m.Meetings,
		meetingsWidth-2,
		topRowHeight-2,
		m.SelectedPanel == PanelMeetings,
		m.IsLoading,
		m.Cursors[PanelMeetings],
	)

	prsPanel := ui.RenderPrsPanel(
		m.Prs,
		prsWidth-2,
		topRowHeight-2,
		m.SelectedPanel == PanelPrs,
		m.IsLoading,
		m.Cursors[PanelPrs],
	)

	portsPanel := ui.RenderPortsPanel(
		m.Ports,
		portsWidth-2,
		topRowHeight-2,
		m.SelectedPanel == PanelPorts,
		m.IsLoading,
		m.Cursors[PanelPorts],
	)

	gitPanel := ui.RenderGitPanel(
		m.Git,
		m.Width-2,
		middleRowHeight-2,
		m.SelectedPanel == PanelGit,
		m.IsLoading,
		m.Cursors[PanelGit],
	)

	stashesPanel := ui.RenderStashesPanel(
		m.Git,
		m.Width-2,
		bottomRowHeight-2,
		m.SelectedPanel == PanelStashes,
		m.IsLoading,
		m.Cursors[PanelStashes],
	)

	// Compose layout
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, meetingsPanel, prsPanel, portsPanel)

	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left,
		topRow,
		gitPanel,
		stashesPanel,
		statusBar,
	)
}

func (m Model) renderStatusBar() string {
	// Show keyboard shortcuts and last refresh time
	shortcuts := []struct {
		key   string
		label string
	}{
		{"tab", "panel"},
		{"\u2191\u2193", "select"},
		{"\u23ce", "open"},
		{"r", "refresh"},
		{"q", "quit"},
	}

	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorCyan)
	var parts []string
	for _, s := range shortcuts {
		parts = append(parts, keyStyle.Render("["+s.key+"]")+s.label)
	}

	shortcutsStr := strings.Join(parts, " ")

	// Calculate time since last refresh
	var refreshStr string
	if !m.LastRefresh.IsZero() {
		elapsed := time.Since(m.LastRefresh)
		if elapsed < time.Minute {
			refreshStr = fmt.Sprintf("Updated: %ds ago", int(elapsed.Seconds()))
		} else {
			refreshStr = fmt.Sprintf("Updated: %dm ago", int(elapsed.Minutes()))
		}
	}

	// Pad to fill width
	padding := m.Width - lipgloss.Width(shortcutsStr) - len(refreshStr) - 4
	if padding < 0 {
		padding = 0
	}

	statusLine := shortcutsStr + fmt.Sprintf("%*s", padding, "") + refreshStr

	// Show notification bar if there are active notifications
	if len(m.Notifications) > 0 {
		notifStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorYellow)
		notifText := strings.Join(m.Notifications, " | ")
		notifText = ui.Truncate(notifText, m.Width-4)
		notifLine := notifStyle.Render("▸ " + notifText)
		return ui.StatusBarStyle.Render(notifLine + "\n" + statusLine)
	}

	return ui.StatusBarStyle.Render(statusLine)
}
