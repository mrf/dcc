use ratatui::{
    layout::Rect,
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, Paragraph},
    Frame,
};

use crate::data::{DirtyRepo, StashInfo};

pub fn draw_git_panel(
    f: &mut Frame,
    area: Rect,
    dirty_repos: &[DirtyRepo],
    selected: bool,
    is_loading: bool,
) {
    let border_color = if is_loading {
        Color::DarkGray
    } else {
        Color::Yellow
    };

    let border_style = if selected {
        Style::default()
            .fg(border_color)
            .add_modifier(Modifier::BOLD)
    } else {
        Style::default().fg(border_color)
    };

    let block = Block::default()
        .title(" UNCOMMITTED WORK ")
        .borders(Borders::ALL)
        .border_style(border_style);

    let lines = if is_loading {
        vec![
            Line::from(""),
            Line::from(vec![Span::styled(
                "  Scanning repositories...",
                Style::default()
                    .fg(Color::DarkGray)
                    .add_modifier(Modifier::ITALIC),
            )]),
        ]
    } else if dirty_repos.is_empty() {
        vec![Line::from(vec![Span::styled(
            "All repos clean!",
            Style::default().fg(Color::Green),
        )])]
    } else {
        let mut lines = Vec::new();

        for repo in dirty_repos.iter().take(6) {
            let mut parts = Vec::new();

            parts.push(Span::styled(
                format!("{:<20} ", truncate(&repo.name, 20)),
                Style::default()
                    .fg(Color::White)
                    .add_modifier(Modifier::BOLD),
            ));

            let mut status_parts = Vec::new();

            if repo.modified > 0 {
                status_parts.push(format!("{} modified", repo.modified));
            }
            if repo.untracked > 0 {
                status_parts.push(format!("{} untracked", repo.untracked));
            }
            if repo.staged > 0 {
                status_parts.push(format!("{} staged", repo.staged));
            }

            let status_color = if repo.staged > 0 {
                Color::Green
            } else if repo.modified > 0 {
                Color::Yellow
            } else {
                Color::Cyan
            };

            parts.push(Span::styled(
                status_parts.join(", "),
                Style::default().fg(status_color),
            ));

            lines.push(Line::from(parts));
        }

        if dirty_repos.len() > 6 {
            lines.push(Line::from(vec![Span::styled(
                format!("  +{} more repos...", dirty_repos.len() - 6),
                Style::default().fg(Color::DarkGray),
            )]));
        }

        lines
    };

    let paragraph = Paragraph::new(lines).block(block);
    f.render_widget(paragraph, area);
}

pub fn draw_stash_panel(
    f: &mut Frame,
    area: Rect,
    stashes: &[StashInfo],
    selected: bool,
    is_loading: bool,
) {
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
        .title(" STASHES ")
        .borders(Borders::ALL)
        .border_style(border_style);

    let lines = if is_loading {
        vec![
            Line::from(""),
            Line::from(vec![Span::styled(
                "  Looking for forgotten work...",
                Style::default()
                    .fg(Color::DarkGray)
                    .add_modifier(Modifier::ITALIC),
            )]),
        ]
    } else if stashes.is_empty() {
        vec![Line::from(vec![Span::styled(
            "No stashes",
            Style::default().fg(Color::DarkGray),
        )])]
    } else {
        let mut lines = Vec::new();

        // Group stashes by repo using BTreeMap for stable ordering
        let mut repo_stashes: std::collections::BTreeMap<&str, Vec<&StashInfo>> =
            std::collections::BTreeMap::new();
        for stash in stashes {
            repo_stashes.entry(&stash.repo).or_default().push(stash);
        }

        for (repo, stash_list) in repo_stashes.iter().take(4) {
            let count = stash_list.len();
            let first = stash_list.first().unwrap();

            let age_indicator = if first.age_days > 30 {
                Span::styled(" (ancient)", Style::default().fg(Color::Red))
            } else if first.age_days > 7 {
                Span::styled(
                    format!(" ({}d)", first.age_days),
                    Style::default().fg(Color::Yellow),
                )
            } else {
                Span::styled(
                    format!(" ({}d)", first.age_days),
                    Style::default().fg(Color::DarkGray),
                )
            };

            lines.push(Line::from(vec![
                Span::styled(
                    format!("{} ", truncate(repo, 15)),
                    Style::default()
                        .fg(Color::White)
                        .add_modifier(Modifier::BOLD),
                ),
                Span::styled(format!("({})", count), Style::default().fg(Color::Cyan)),
                Span::raw(" "),
                Span::styled(
                    truncate(&first.message, 25),
                    Style::default().fg(Color::Gray),
                ),
                age_indicator,
            ]));
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

/// Groups stashes by repository name, returning repos in stable alphabetical order.
/// This prevents the "scrolling" effect that would occur with HashMap's random order.
#[cfg(test)]
pub fn group_stashes_by_repo(stashes: &[StashInfo]) -> Vec<(&str, Vec<&StashInfo>)> {
    use std::collections::BTreeMap;

    let mut repo_stashes: BTreeMap<&str, Vec<&StashInfo>> = BTreeMap::new();
    for stash in stashes {
        repo_stashes.entry(&stash.repo).or_default().push(stash);
    }

    repo_stashes.into_iter().collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_stash(repo: &str, message: &str, age_days: i64) -> StashInfo {
        StashInfo {
            repo: repo.to_string(),
            index: 0,
            message: message.to_string(),
            age_days,
        }
    }

    #[test]
    fn test_group_stashes_by_repo_returns_alphabetical_order() {
        let stashes = vec![
            make_stash("zebra-repo", "stash 1", 1),
            make_stash("alpha-repo", "stash 2", 2),
            make_stash("middle-repo", "stash 3", 3),
            make_stash("alpha-repo", "stash 4", 4),
        ];

        let grouped = group_stashes_by_repo(&stashes);
        let repo_names: Vec<&str> = grouped.iter().map(|(name, _)| *name).collect();

        // Should be alphabetically sorted
        assert_eq!(repo_names, vec!["alpha-repo", "middle-repo", "zebra-repo"]);
    }

    #[test]
    fn test_group_stashes_by_repo_groups_correctly() {
        let stashes = vec![
            make_stash("repo-a", "stash 1", 1),
            make_stash("repo-b", "stash 2", 2),
            make_stash("repo-a", "stash 3", 3),
        ];

        let grouped = group_stashes_by_repo(&stashes);

        assert_eq!(grouped.len(), 2);

        // repo-a should have 2 stashes
        let repo_a = grouped.iter().find(|(name, _)| *name == "repo-a").unwrap();
        assert_eq!(repo_a.1.len(), 2);

        // repo-b should have 1 stash
        let repo_b = grouped.iter().find(|(name, _)| *name == "repo-b").unwrap();
        assert_eq!(repo_b.1.len(), 1);
    }

    #[test]
    fn test_group_stashes_by_repo_stable_across_multiple_calls() {
        let stashes = vec![
            make_stash("charlie", "s1", 1),
            make_stash("alpha", "s2", 2),
            make_stash("bravo", "s3", 3),
        ];

        // Call multiple times and ensure order is always the same
        let result1 = group_stashes_by_repo(&stashes);
        let result2 = group_stashes_by_repo(&stashes);
        let result3 = group_stashes_by_repo(&stashes);

        let names1: Vec<&str> = result1.iter().map(|(n, _)| *n).collect();
        let names2: Vec<&str> = result2.iter().map(|(n, _)| *n).collect();
        let names3: Vec<&str> = result3.iter().map(|(n, _)| *n).collect();

        assert_eq!(names1, names2);
        assert_eq!(names2, names3);
        assert_eq!(names1, vec!["alpha", "bravo", "charlie"]);
    }

    #[test]
    fn test_group_stashes_empty() {
        let stashes: Vec<StashInfo> = vec![];
        let grouped = group_stashes_by_repo(&stashes);
        assert!(grouped.is_empty());
    }

    #[test]
    fn test_truncate_short_string() {
        assert_eq!(truncate("hello", 10), "hello");
    }

    #[test]
    fn test_truncate_exact_length() {
        assert_eq!(truncate("hello", 5), "hello");
    }

    #[test]
    fn test_truncate_long_string() {
        assert_eq!(truncate("hello world", 8), "hello...");
    }

    #[test]
    fn test_truncate_very_short_max() {
        assert_eq!(truncate("hello", 3), "...");
    }
}
