use ratatui::{
    layout::Rect,
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, Paragraph},
    Frame,
};

use crate::data::PortsPanel;

pub fn draw_ports_panel(
    f: &mut Frame,
    area: Rect,
    ports: &PortsPanel,
    selected: bool,
    is_loading: bool,
) {
    let border_color = if is_loading {
        Color::DarkGray
    } else {
        Color::Cyan
    };

    let border_style = if selected {
        Style::default()
            .fg(border_color)
            .add_modifier(Modifier::BOLD)
    } else {
        Style::default().fg(border_color)
    };

    let block = Block::default()
        .title(" PORTS ")
        .borders(Borders::ALL)
        .border_style(border_style);

    let lines = if is_loading {
        vec![
            Line::from(""),
            Line::from(vec![Span::styled(
                "  Scanning ports...",
                Style::default()
                    .fg(Color::DarkGray)
                    .add_modifier(Modifier::ITALIC),
            )]),
        ]
    } else if ports.ports.is_empty() {
        vec![Line::from(vec![Span::styled(
            "No active ports",
            Style::default().fg(Color::DarkGray),
        )])]
    } else {
        let mut lines = Vec::new();

        for port in ports.ports.iter().take(10) {
            let port_color = port_to_color(port.port);

            lines.push(Line::from(vec![
                Span::styled(
                    format!(":{:<5} ", port.port),
                    Style::default().fg(port_color).add_modifier(Modifier::BOLD),
                ),
                Span::styled(
                    truncate(&port.process.to_lowercase(), 12),
                    Style::default().fg(Color::White),
                ),
            ]));
        }

        if ports.ports.len() > 10 {
            lines.push(Line::from(vec![Span::styled(
                format!("  +{} more...", ports.ports.len() - 10),
                Style::default().fg(Color::DarkGray),
            )]));
        }

        lines
    };

    let paragraph = Paragraph::new(lines).block(block);
    f.render_widget(paragraph, area);
}

fn port_to_color(port: u16) -> Color {
    match port {
        80 | 443 | 8080 | 8443 => Color::Green,      // Web servers
        5432 | 3306 | 27017 | 6379 => Color::Yellow, // Databases
        3000..=3999 => Color::Cyan,                  // Dev servers
        _ => Color::White,
    }
}

fn truncate(s: &str, max_len: usize) -> String {
    if s.len() <= max_len {
        s.to_string()
    } else {
        format!("{}...", &s[..max_len.saturating_sub(3)])
    }
}
