use ratatui::{
    layout::Rect,
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, Paragraph},
    Frame,
};

use crate::data::{MeetingStatus, MeetingsPanel};

pub fn draw_meetings_panel(
    f: &mut Frame,
    area: Rect,
    meetings: &MeetingsPanel,
    selected: bool,
    is_loading: bool,
) {
    let border_color = if is_loading {
        Color::DarkGray
    } else {
        match &meetings.status {
            MeetingStatus::Free { minutes_until } if *minutes_until > 60 => Color::Green,
            MeetingStatus::Free { minutes_until } if *minutes_until > 30 => Color::Yellow,
            MeetingStatus::Free { minutes_until } if *minutes_until > 10 => Color::Rgb(255, 165, 0),
            MeetingStatus::Free { .. } => Color::Red,
            MeetingStatus::InMeeting { .. } => Color::Blue,
            MeetingStatus::Clear => Color::Green,
        }
    };

    let border_style = if selected {
        Style::default()
            .fg(border_color)
            .add_modifier(Modifier::BOLD)
    } else {
        Style::default().fg(border_color)
    };

    let block = Block::default()
        .title(" MEETINGS ")
        .borders(Borders::ALL)
        .border_style(border_style);

    let lines = if is_loading {
        vec![
            Line::from(""),
            Line::from(vec![Span::styled(
                "  Checking calendar...",
                Style::default()
                    .fg(Color::DarkGray)
                    .add_modifier(Modifier::ITALIC),
            )]),
        ]
    } else {
        let mut lines = Vec::new();

        let (status_color, status_text) = match &meetings.status {
            MeetingStatus::Free { minutes_until } => {
                (Color::Green, format!("{} min until:", minutes_until))
            }
            MeetingStatus::InMeeting { ends_in } => {
                (Color::Blue, format!("IN MEETING - ends in {} min", ends_in))
            }
            MeetingStatus::Clear => (Color::Green, "Clear for the day!".to_string()),
        };

        lines.push(Line::from(vec![
            Span::styled("● ", Style::default().fg(status_color)),
            Span::styled(status_text, Style::default().fg(Color::White)),
        ]));

        if let Some(next) = &meetings.next_meeting {
            lines.push(Line::from(vec![Span::styled(
                format!("  {}", truncate(&next.title, 25)),
                Style::default()
                    .fg(Color::White)
                    .add_modifier(Modifier::BOLD),
            )]));
            lines.push(Line::from(""));
        }

        if !meetings.upcoming.is_empty() {
            lines.push(Line::from(vec![Span::styled(
                "Then:",
                Style::default().fg(Color::DarkGray),
            )]));

            for meeting in meetings.upcoming.iter().take(3) {
                let time = meeting.start.format("%l:%M %p").to_string();
                lines.push(Line::from(vec![
                    Span::styled(
                        format!("  {} ", time.trim()),
                        Style::default().fg(Color::Cyan),
                    ),
                    Span::styled(
                        truncate(&meeting.title, 18),
                        Style::default().fg(Color::Gray),
                    ),
                ]));
            }
        }

        lines
    };

    let paragraph = Paragraph::new(lines).block(block);
    f.render_widget(paragraph, area);
}

fn truncate(s: &str, max_len: usize) -> String {
    if s.len() <= max_len {
        s.to_string()
    } else {
        format!("{}...", &s[..max_len.saturating_sub(3)])
    }
}
