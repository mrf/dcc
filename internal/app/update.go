package app

import (
	"os/exec"
	"runtime"
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
		m.ClampCursors()
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

	case "up", "k":
		m.CursorUp()
		return m, nil

	case "down", "j":
		m.CursorDown()
		return m, nil

	case "enter":
		return m.handleEnter()

	case "m":
		// Open Calendar app (macOS only)
		if runtime.GOOS == "darwin" {
			go func() {
				_ = exec.Command("open", "-a", "Calendar").Run()
			}()
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	idx := m.Cursors[m.SelectedPanel]
	switch m.SelectedPanel {
	case PanelMeetings:
		go func() {
			_ = exec.Command("open", "-a", "Calendar").Run()
		}()
	case PanelPrs:
		go func() {
			_ = data.OpenPrByIndex(m.Prs, idx)
		}()
	case PanelPorts:
		go func() {
			_ = data.OpenPort(m.Ports, idx)
		}()
	case PanelGit:
		go func() {
			_ = data.OpenDirtyRepoByIndex(m.Git, idx)
		}()
	case PanelStashes:
		go func() {
			_ = data.OpenStashRepoByIndex(m.Git, idx)
		}()
	}
	return m, nil
}
