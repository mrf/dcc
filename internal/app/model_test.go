package app

import (
	"testing"

	"github.com/mrf/dcc/internal/config"
)

func TestPanelNextCyclesThroughAllPanels(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	// Start at Meetings
	if m.SelectedPanel != PanelMeetings {
		t.Errorf("expected PanelMeetings, got %v", m.SelectedPanel)
	}

	m.NextPanel()
	if m.SelectedPanel != PanelPrs {
		t.Errorf("expected PanelPrs, got %v", m.SelectedPanel)
	}

	m.NextPanel()
	if m.SelectedPanel != PanelPorts {
		t.Errorf("expected PanelPorts, got %v", m.SelectedPanel)
	}

	m.NextPanel()
	if m.SelectedPanel != PanelGit {
		t.Errorf("expected PanelGit, got %v", m.SelectedPanel)
	}

	m.NextPanel()
	if m.SelectedPanel != PanelStashes {
		t.Errorf("expected PanelStashes, got %v", m.SelectedPanel)
	}

	// Should wrap back to Meetings
	m.NextPanel()
	if m.SelectedPanel != PanelMeetings {
		t.Errorf("expected PanelMeetings after wrap, got %v", m.SelectedPanel)
	}
}

func TestPanelPrevCyclesBackwards(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	// Going backwards from Meetings should go to Stashes
	m.PrevPanel()
	if m.SelectedPanel != PanelStashes {
		t.Errorf("expected PanelStashes, got %v", m.SelectedPanel)
	}

	m.PrevPanel()
	if m.SelectedPanel != PanelGit {
		t.Errorf("expected PanelGit, got %v", m.SelectedPanel)
	}

	m.PrevPanel()
	if m.SelectedPanel != PanelPorts {
		t.Errorf("expected PanelPorts, got %v", m.SelectedPanel)
	}

	m.PrevPanel()
	if m.SelectedPanel != PanelPrs {
		t.Errorf("expected PanelPrs, got %v", m.SelectedPanel)
	}

	m.PrevPanel()
	if m.SelectedPanel != PanelMeetings {
		t.Errorf("expected PanelMeetings, got %v", m.SelectedPanel)
	}
}

func TestAppStateDefaultStartsOnMeetings(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	if m.SelectedPanel != PanelMeetings {
		t.Errorf("expected PanelMeetings, got %v", m.SelectedPanel)
	}
}

func TestAppStateNextPanel(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	if m.SelectedPanel != PanelMeetings {
		t.Errorf("expected PanelMeetings, got %v", m.SelectedPanel)
	}

	m.NextPanel()
	if m.SelectedPanel != PanelPrs {
		t.Errorf("expected PanelPrs, got %v", m.SelectedPanel)
	}

	m.NextPanel()
	if m.SelectedPanel != PanelPorts {
		t.Errorf("expected PanelPorts, got %v", m.SelectedPanel)
	}
}

func TestAppStatePrevPanel(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	if m.SelectedPanel != PanelMeetings {
		t.Errorf("expected PanelMeetings, got %v", m.SelectedPanel)
	}

	m.PrevPanel()
	if m.SelectedPanel != PanelStashes {
		t.Errorf("expected PanelStashes, got %v", m.SelectedPanel)
	}

	m.PrevPanel()
	if m.SelectedPanel != PanelGit {
		t.Errorf("expected PanelGit, got %v", m.SelectedPanel)
	}
}

func TestPanelNavigationIsReversible(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.SelectedPanel = PanelPrs

	// Going next then prev should return to same panel
	m.NextPanel()
	m.PrevPanel()
	if m.SelectedPanel != PanelPrs {
		t.Errorf("expected PanelPrs after next+prev, got %v", m.SelectedPanel)
	}

	// Going prev then next should also return to same panel
	m.PrevPanel()
	m.NextPanel()
	if m.SelectedPanel != PanelPrs {
		t.Errorf("expected PanelPrs after prev+next, got %v", m.SelectedPanel)
	}
}

func TestFullCycleNextReturnsToStart(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	// Cycle through all 5 panels
	for i := 0; i < 5; i++ {
		m.NextPanel()
	}

	if m.SelectedPanel != PanelMeetings {
		t.Errorf("expected PanelMeetings after full cycle, got %v", m.SelectedPanel)
	}
}

func TestFullCyclePrevReturnsToStart(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	// Cycle backwards through all 5 panels
	for i := 0; i < 5; i++ {
		m.PrevPanel()
	}

	if m.SelectedPanel != PanelMeetings {
		t.Errorf("expected PanelMeetings after full cycle, got %v", m.SelectedPanel)
	}
}
