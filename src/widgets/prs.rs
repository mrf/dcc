use ratatui::{
    layout::Rect,
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, Paragraph},
    Frame,
};

use crate::data::PrsPanel;

pub fn draw_prs_panel(f: &mut Frame, area: Rect, prs: &PrsPanel, selected: bool, is_loading: bool) {
    let border_color = if is_loading {
        Color::DarkGray
    } else {
        Color::Magenta
    };

    let border_style = if selected {
        Style::default()
            .fg(border_color)
            .add_modifier(Modifier::BOLD)
    } else {
        Style::default().fg(border_color)
    };

    let block = Block::default()
        .title(" PRS ")
        .borders(Borders::ALL)
        .border_style(border_style);

    let lines = if is_loading {
        vec![
            Line::from(""),
            Line::from(vec![Span::styled(
                "  Fetching PRs from GitHub...",
                Style::default()
                    .fg(Color::DarkGray)
                    .add_modifier(Modifier::ITALIC),
            )]),
        ]
    } else {
        let mut lines = Vec::new();

        // Needs Review section
        if !prs.needs_review.is_empty() {
            lines.push(Line::from(vec![Span::styled(
                format!("Needs Review ({}):", prs.needs_review.len()),
                Style::default()
                    .fg(Color::Yellow)
                    .add_modifier(Modifier::BOLD),
            )]));

            for pr in prs.needs_review.iter().take(4) {
                let age_color = age_to_color(pr.age_days);
                let age_indicator = if pr.age_days > 0 {
                    format!("{}d", pr.age_days)
                } else {
                    "new".to_string()
                };

                lines.push(Line::from(vec![
                    Span::styled(
                        format!("  #{} ", pr.number),
                        Style::default().fg(Color::Cyan),
                    ),
                    Span::styled(truncate(&pr.title, 20), Style::default().fg(Color::White)),
                    Span::raw(" "),
                    Span::styled(age_indicator, Style::default().fg(age_color)),
                ]));
            }

            lines.push(Line::from(""));
        }

        // Your PRs section
        if !prs.your_prs.is_empty() {
            lines.push(Line::from(vec![Span::styled(
                format!("Your PRs ({}):", prs.your_prs.len()),
                Style::default()
                    .fg(Color::Green)
                    .add_modifier(Modifier::BOLD),
            )]));

            for pr in prs.your_prs.iter().take(4) {
                let age_color = age_to_color(pr.age_days);
                let age_indicator = if pr.age_days > 0 {
                    format!("{}d", pr.age_days)
                } else {
                    "new".to_string()
                };

                let status_indicator = match pr.review_decision.as_deref() {
                    Some("APPROVED") => Span::styled(" ✓", Style::default().fg(Color::Green)),
                    Some("CHANGES_REQUESTED") => {
                        Span::styled(" ✗", Style::default().fg(Color::Red))
                    }
                    Some("REVIEW_REQUIRED") => {
                        Span::styled(" ○", Style::default().fg(Color::Yellow))
                    }
                    _ => Span::raw(""),
                };

                lines.push(Line::from(vec![
                    Span::styled(
                        format!("  #{} ", pr.number),
                        Style::default().fg(Color::Cyan),
                    ),
                    Span::styled(truncate(&pr.title, 18), Style::default().fg(Color::White)),
                    Span::raw(" "),
                    Span::styled(age_indicator, Style::default().fg(age_color)),
                    status_indicator,
                ]));
            }
        }

        if prs.needs_review.is_empty() && prs.your_prs.is_empty() {
            lines.push(Line::from(vec![Span::styled(
                "No PRs to show",
                Style::default().fg(Color::DarkGray),
            )]));
        }

        lines
    };

    let paragraph = Paragraph::new(lines).block(block);
    f.render_widget(paragraph, area);
}

fn age_to_color(age_days: i64) -> Color {
    match age_days {
        0..=1 => Color::Green,
        2..=4 => Color::Yellow,
        5..=6 => Color::Rgb(255, 165, 0), // Orange
        _ => Color::Red,
    }
}

fn truncate(s: &str, max_len: usize) -> String {
    if s.len() <= max_len {
        s.to_string()
    } else {
        format!("{}...", &s[..max_len.saturating_sub(3)])
    }
}
