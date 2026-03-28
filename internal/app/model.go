package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrf/dcc/internal/config"
	"github.com/mrf/dcc/internal/data"
)

// Panel represents the currently selected panel
type Panel int

const (
	PanelMeetings Panel = iota
	PanelPrs
	PanelPorts
	PanelGit
	PanelStashes
)

// Model represents the application state
type Model struct {
	Meetings      data.MeetingsPanel
	Prs           data.PrsPanel
	Ports         data.PortsPanel
	Git           data.GitPanel
	SelectedPanel Panel
	IsLoading     bool
	LastRefresh   time.Time
	Config        config.Config
	Width         int
	Height        int
	Notifications []string
}

// NewModel creates a new model with the given config
func NewModel(cfg config.Config) Model {
	return Model{
		Meetings:      data.MeetingsPanel{IsLoading: true},
		Prs:           data.PrsPanel{IsLoading: true},
		Ports:         data.PortsPanel{IsLoading: true},
		Git:           data.GitPanel{IsLoading: true},
		SelectedPanel: PanelMeetings,
		IsLoading:     true,
		Config:        cfg,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchAllData(),
		tickCmd(time.Duration(m.Config.General.RefreshIntervalSeconds)*time.Second),
	)
}

// NextPanel cycles to the next panel
func (m *Model) NextPanel() {
	m.SelectedPanel = (m.SelectedPanel + 1) % 5
}

// PrevPanel cycles to the previous panel
func (m *Model) PrevPanel() {
	if m.SelectedPanel == 0 {
		m.SelectedPanel = PanelStashes
	} else {
		m.SelectedPanel--
	}
}

// fetchAllData returns a command that fetches all data
func (m Model) fetchAllData() tea.Cmd {
	return func() tea.Msg {
		meetings := data.FetchMeetings(m.Config.Meetings)
		prs := data.FetchPrs(m.Config.Prs)
		ports := data.FetchPorts(m.Config.Ports)
		git := data.FetchGitStatus(m.Config.Git, m.Config.General.ProjectsDir)

		return DataFetchedMsg{
			Meetings: meetings,
			Prs:      prs,
			Ports:    ports,
			Git:      git,
		}
	}
}

// tickCmd returns a command that sends a TickMsg after the given duration
func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
