use anyhow::Result;
use chrono::{DateTime, Local};

use crate::data::{
    fetch_git_status, fetch_meetings, fetch_ports, fetch_prs, Config, GitPanel, MeetingsPanel,
    PortsPanel, PrsPanel,
};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Panel {
    Meetings,
    Prs,
    Ports,
    Git,
    Stashes,
}

impl Panel {
    pub fn next(&self) -> Self {
        match self {
            Panel::Meetings => Panel::Prs,
            Panel::Prs => Panel::Ports,
            Panel::Ports => Panel::Git,
            Panel::Git => Panel::Stashes,
            Panel::Stashes => Panel::Meetings,
        }
    }

    pub fn prev(&self) -> Self {
        match self {
            Panel::Meetings => Panel::Stashes,
            Panel::Prs => Panel::Meetings,
            Panel::Ports => Panel::Prs,
            Panel::Git => Panel::Ports,
            Panel::Stashes => Panel::Git,
        }
    }
}

#[derive(Debug, Clone)]
pub struct AppState {
    pub meetings: MeetingsPanel,
    pub prs: PrsPanel,
    pub ports: PortsPanel,
    pub git: GitPanel,
    pub last_refresh: DateTime<Local>,
    pub selected_panel: Panel,
    pub should_quit: bool,
    pub is_loading: bool,
}

impl Default for AppState {
    fn default() -> Self {
        Self {
            meetings: MeetingsPanel::default(),
            prs: PrsPanel::default(),
            ports: PortsPanel::default(),
            git: GitPanel::default(),
            last_refresh: Local::now(),
            selected_panel: Panel::Meetings,
            should_quit: false,
            is_loading: true, // Start in loading state
        }
    }
}

impl AppState {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn refresh(&mut self, config: &Config) -> Result<()> {
        // Fetch all data
        self.meetings = fetch_meetings(&config.meetings).unwrap_or_default();
        self.prs = fetch_prs(&config.prs).unwrap_or_default();
        self.ports = fetch_ports(&config.ports).unwrap_or_default();
        self.git = fetch_git_status(&config.projects_path(), &config.git).unwrap_or_default();
        self.last_refresh = Local::now();
        self.is_loading = false;

        Ok(())
    }

    pub fn next_panel(&mut self) {
        self.selected_panel = self.selected_panel.next();
    }

    pub fn prev_panel(&mut self) {
        self.selected_panel = self.selected_panel.prev();
    }

    pub fn seconds_since_refresh(&self) -> i64 {
        (Local::now() - self.last_refresh).num_seconds()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_panel_next_cycles_through_all_panels() {
        let mut panel = Panel::Meetings;

        panel = panel.next();
        assert_eq!(panel, Panel::Prs);

        panel = panel.next();
        assert_eq!(panel, Panel::Ports);

        panel = panel.next();
        assert_eq!(panel, Panel::Git);

        panel = panel.next();
        assert_eq!(panel, Panel::Stashes);

        // Should wrap back to Meetings
        panel = panel.next();
        assert_eq!(panel, Panel::Meetings);
    }

    #[test]
    fn test_panel_prev_cycles_backwards() {
        let mut panel = Panel::Meetings;

        // Going backwards from Meetings should go to Stashes
        panel = panel.prev();
        assert_eq!(panel, Panel::Stashes);

        panel = panel.prev();
        assert_eq!(panel, Panel::Git);

        panel = panel.prev();
        assert_eq!(panel, Panel::Ports);

        panel = panel.prev();
        assert_eq!(panel, Panel::Prs);

        panel = panel.prev();
        assert_eq!(panel, Panel::Meetings);
    }

    #[test]
    fn test_app_state_default_starts_on_meetings() {
        let app = AppState::new();
        assert_eq!(app.selected_panel, Panel::Meetings);
        assert!(!app.should_quit);
    }

    #[test]
    fn test_app_state_next_panel() {
        let mut app = AppState::new();
        assert_eq!(app.selected_panel, Panel::Meetings);

        app.next_panel();
        assert_eq!(app.selected_panel, Panel::Prs);

        app.next_panel();
        assert_eq!(app.selected_panel, Panel::Ports);
    }

    #[test]
    fn test_app_state_prev_panel() {
        let mut app = AppState::new();
        assert_eq!(app.selected_panel, Panel::Meetings);

        app.prev_panel();
        assert_eq!(app.selected_panel, Panel::Stashes);

        app.prev_panel();
        assert_eq!(app.selected_panel, Panel::Git);
    }

    #[test]
    fn test_panel_navigation_is_reversible() {
        // Going next then prev should return to same panel
        let mut panel = Panel::Prs;
        panel = panel.next();
        panel = panel.prev();
        assert_eq!(panel, Panel::Prs);

        // Going prev then next should also return to same panel
        panel = panel.prev();
        panel = panel.next();
        assert_eq!(panel, Panel::Prs);
    }

    #[test]
    fn test_full_cycle_next_returns_to_start() {
        let mut panel = Panel::Meetings;

        // Cycle through all 5 panels
        for _ in 0..5 {
            panel = panel.next();
        }

        assert_eq!(panel, Panel::Meetings);
    }

    #[test]
    fn test_full_cycle_prev_returns_to_start() {
        let mut panel = Panel::Meetings;

        // Cycle backwards through all 5 panels
        for _ in 0..5 {
            panel = panel.prev();
        }

        assert_eq!(panel, Panel::Meetings);
    }
}
