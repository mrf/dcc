package app

import (
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrf/dcc/internal/data"
)

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case DataFetchedMsg:
		var cmds []tea.Cmd

		// Detect changes if we have previous data
		if !m.LastRefresh.IsZero() {
			notes := detectPrChanges(m.Prs, msg.Prs)
			if meetingNote := detectMeetingSoon(m.Meetings, msg.Meetings); meetingNote != "" {
				notes = append(notes, meetingNote)
			}
			if len(notes) > 0 {
				m.Notifications = notes
				cmds = append(cmds, clearNotificationCmd())
			}
		}

		m.Meetings = msg.Meetings
		m.Prs = msg.Prs
		m.Ports = msg.Ports
		m.Git = msg.Git
		m.IsLoading = false
		m.LastRefresh = time.Now()
		return m, tea.Batch(cmds...)

	case ClearNotificationMsg:
		m.Notifications = nil
		return m, nil

	case TickMsg:
		// Auto-refresh
		m.IsLoading = true
		return m, tea.Batch(
			m.fetchAllData(),
			tickCmd(time.Duration(m.Config.General.RefreshIntervalSeconds)*time.Second),
		)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit

	case "r":
		// Manual refresh
		m.IsLoading = true
		return m, m.fetchAllData()

	case "tab":
		m.NextPanel()
		return m, nil

	case "shift+tab":
		m.PrevPanel()
		return m, nil

	case "p":
		// Open first PR in browser
		go func() {
			_ = data.OpenFirstPr(m.Prs)
		}()
		return m, nil

	case "m":
		// Open Calendar app
		go func() {
			_ = exec.Command("open", "-a", "Calendar").Run()
		}()
		return m, nil

	case "g":
		// Open first dirty repo in VS Code or Finder
		go func() {
			_ = data.OpenFirstDirtyRepo(m.Git)
		}()
		return m, nil
	}

	return m, nil
}
