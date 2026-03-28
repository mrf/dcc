package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrf/dcc/internal/ui"
)

const statusBarHeight = 3

// View implements tea.Model
func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	if m.FocusMode {
		return m.viewFocusMode()
	}

	return m.viewFull()
}

func (m Model) viewFocusMode() string {
	panelHeight := m.Height - statusBarHeight

	meetingsPanel := ui.RenderMeetingsPanel(
		m.Meetings,
		m.Width-2,
		panelHeight-2,
		true,
		m.IsLoading,
	)

	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, meetingsPanel, statusBar)
}

func (m Model) viewFull() string {
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
	)

	prsPanel := ui.RenderPrsPanel(
		m.Prs,
		prsWidth-2,
		topRowHeight-2,
		m.SelectedPanel == PanelPrs,
		m.IsLoading,
	)

	portsPanel := ui.RenderPortsPanel(
		m.Ports,
		portsWidth-2,
		topRowHeight-2,
		m.SelectedPanel == PanelPorts,
		m.IsLoading,
	)

	gitPanel := ui.RenderGitPanel(
		m.Git,
		m.Width-2,
		middleRowHeight-2,
		m.SelectedPanel == PanelGit,
		m.IsLoading,
	)

	stashesPanel := ui.RenderStashesPanel(
		m.Git,
		m.Width-2,
		bottomRowHeight-2,
		m.SelectedPanel == PanelStashes,
		m.IsLoading,
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
	var focusLabel string
	if m.FocusMode {
		focusLabel = "dashboard"
	} else {
		focusLabel = "focus"
	}

	shortcuts := []struct {
		key   string
		label string
	}{
		{"f", focusLabel},
		{"r", "refresh"},
		{"p", "prs"},
		{"m", "meetings"},
		{"g", "git"},
		{"q", "quit"},
	}

	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorCyan)
	parts := make([]string, len(shortcuts))
	for i, s := range shortcuts {
		parts[i] = keyStyle.Render("["+s.key+"]") + s.label
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

	return ui.StatusBarStyle.Render(
		shortcutsStr + fmt.Sprintf("%*s", padding, "") + refreshStr,
	)
}
