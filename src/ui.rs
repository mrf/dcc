use ratatui::{
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Style},
    text::{Line, Span},
    widgets::{Block, Borders, Paragraph},
    Frame,
};

use crate::app::{AppState, Panel};
use crate::widgets::{
    draw_git_panel, draw_meetings_panel, draw_ports_panel, draw_prs_panel, draw_stash_panel,
};

pub fn draw(f: &mut Frame, app: &AppState) {
    // Main layout: top row (3 panels) + bottom section (git)
    let main_chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Percentage(40), // Top panels
            Constraint::Percentage(30), // Uncommitted work
            Constraint::Percentage(20), // Stashes
            Constraint::Min(3),         // Status bar
        ])
        .split(f.area());

    // Top row: meetings | PRs | ports
    let top_chunks = Layout::default()
        .direction(Direction::Horizontal)
        .constraints([
            Constraint::Percentage(30),
            Constraint::Percentage(40),
            Constraint::Percentage(30),
        ])
        .split(main_chunks[0]);

    draw_meetings_panel(
        f,
        top_chunks[0],
        &app.meetings,
        app.selected_panel == Panel::Meetings,
        app.is_loading,
    );
    draw_prs_panel(
        f,
        top_chunks[1],
        &app.prs,
        app.selected_panel == Panel::Prs,
        app.is_loading,
    );
    draw_ports_panel(
        f,
        top_chunks[2],
        &app.ports,
        app.selected_panel == Panel::Ports,
        app.is_loading,
    );
    draw_git_panel(
        f,
        main_chunks[1],
        &app.git.dirty_repos,
        app.selected_panel == Panel::Git,
        app.is_loading,
    );
    draw_stash_panel(
        f,
        main_chunks[2],
        &app.git.stashes,
        app.selected_panel == Panel::Stashes,
        app.is_loading,
    );
    draw_status_bar(f, main_chunks[3], app);
}

fn draw_status_bar(f: &mut Frame, area: Rect, app: &AppState) {
    let block = Block::default()
        .borders(Borders::TOP)
        .border_style(Style::default().fg(Color::DarkGray));

    let seconds_ago = app.seconds_since_refresh();
    let refresh_text = if seconds_ago < 60 {
        format!("{}s ago", seconds_ago)
    } else {
        format!("{}m ago", seconds_ago / 60)
    };

    let shortcuts = Line::from(vec![
        Span::styled("[r]", Style::default().fg(Color::Cyan)),
        Span::styled("efresh  ", Style::default().fg(Color::DarkGray)),
        Span::styled("[p]", Style::default().fg(Color::Cyan)),
        Span::styled("rs  ", Style::default().fg(Color::DarkGray)),
        Span::styled("[m]", Style::default().fg(Color::Cyan)),
        Span::styled("eetings  ", Style::default().fg(Color::DarkGray)),
        Span::styled("[g]", Style::default().fg(Color::Cyan)),
        Span::styled("it  ", Style::default().fg(Color::DarkGray)),
        Span::styled("[Tab]", Style::default().fg(Color::Cyan)),
        Span::styled(" switch  ", Style::default().fg(Color::DarkGray)),
        Span::styled("[q]", Style::default().fg(Color::Cyan)),
        Span::styled("uit", Style::default().fg(Color::DarkGray)),
        Span::raw("     "),
        Span::styled(
            format!("Updated: {}", refresh_text),
            Style::default().fg(Color::DarkGray),
        ),
    ]);

    let paragraph = Paragraph::new(shortcuts).block(block);
    f.render_widget(paragraph, area);
}
