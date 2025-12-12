# Dev Command Center

A terminal dashboard that shows your complete developer context at a glance. The capstone project for 12 Days of Shipmas - combining meeting-buffer, pr-dashboard, port-watcher, and git status scanning into one TUI.

## Language & Framework

**Rust + Ratatui** - Fast, beautiful terminal UIs, single binary distribution.

## Why This Exists

Context switching is expensive. Before starting any task, you need to know:
- How long until my next meeting?
- Any PRs need my attention?
- What's running on my ports?
- Do I have uncommitted work somewhere?

One glance. One command. Full context.

## User Experience

```
┌─ Dev Command Center ─────────────────────────────────────────────────────┐
│                                                                          │
│  ⏱ MEETINGS            │  📋 PRS                    │  🔌 PORTS          │
│  ──────────────────────│────────────────────────────│────────────────────│
│  🟢 47m until:         │  Needs Review (2):         │  :3000  node       │
│     1:1 with Sarah     │    #421 Add auth  2d       │  :3001  node       │
│                        │    #418 Fix bug   4d 🟠    │  :5432  postgres   │
│  Then:                 │                            │  :6379  redis      │
│    2:00 PM Sprint Plan │  Your PRs (1):             │  :8080  python     │
│    3:30 PM Eng Sync    │    #419 Refactor  1d 🟢    │  :11434 ollama     │
│                        │                            │                    │
├──────────────────────────────────────────────────────────────────────────┤
│  📁 UNCOMMITTED WORK                                                     │
│  ────────────────────────────────────────────────────────────────────────│
│  hayden-cloud/          2 modified, 1 untracked                          │
│  pr-dashboard/          5 untracked files                                │
│  .claude/               staged changes (not committed)                   │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  📦 STASHES                                                              │
│  ────────────────────────────────────────────────────────────────────────│
│  branch-cleaner (2)     stash@{0}: 3d ago - WIP auth refactor            │
│  runbook-gen (1)        stash@{0}: 47d ago - broken experiment 🪦        │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  [r]efresh  [p]rs  [m]eetings  [g]it  [s]tashes  [q]uit     Updated: 12s │
└──────────────────────────────────────────────────────────────────────────┘
```

## Data Sources

| Panel | Source | Command |
|-------|--------|---------|
| Meetings | Calendar.app via osascript | `osascript -e '...'` |
| PRs | GitHub CLI | `gh pr status --json` |
| Ports | lsof | `lsof -i -P -n \| grep LISTEN` |
| Uncommitted | git | `git status --porcelain` |
| Stashes | git | `git stash list` |

All data gathered via subprocess - same pattern as your other Shipmas projects.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Ratatui TUI                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────────┐ │
│  │ Meetings │  │   PRs    │  │  Ports   │  │ Git Status/Stash │ │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────────┬─────────┘ │
└───────┼─────────────┼─────────────┼─────────────────┼───────────┘
        │             │             │                 │
        ▼             ▼             ▼                 ▼
   ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────────────┐
   │osascript│   │ gh cli  │   │  lsof   │   │ git (per repo)  │
   └─────────┘   └─────────┘   └─────────┘   └─────────────────┘
```

## Files

```
dev-command-center/
├── src/
│   ├── main.rs           # Entry point, event loop
│   ├── app.rs            # Application state
│   ├── ui.rs             # Ratatui layout and rendering
│   ├── data/
│   │   ├── mod.rs
│   │   ├── meetings.rs   # Calendar.app via osascript
│   │   ├── prs.rs        # gh pr status parsing
│   │   ├── ports.rs      # lsof parsing
│   │   ├── git.rs        # git status/stash across repos
│   │   └── config.rs     # Configuration
│   └── widgets/
│       ├── mod.rs
│       ├── meetings.rs   # Meeting panel widget
│       ├── prs.rs        # PR panel widget
│       ├── ports.rs      # Ports panel widget
│       └── git.rs        # Git status widget
├── Cargo.toml
├── config.toml           # User configuration
├── ARCHITECTURE.md
└── README.md
```

## Dependencies

```toml
[package]
name = "dcc"
version = "0.1.0"
edition = "2021"

[dependencies]
ratatui = "0.28"
crossterm = "0.28"
tokio = { version = "1", features = ["full"] }
serde = { version = "1", features = ["derive"] }
serde_json = "1"
toml = "0.8"
dirs = "5"
chrono = "0.4"
```

## Configuration

```toml
# ~/.config/dcc/config.toml

[general]
refresh_interval_seconds = 30
projects_dir = "~/Projects"

[meetings]
enabled = true
hours_ahead = 8
calendars_exclude = ["Birthdays", "US Holidays", "Siri Suggestions"]
ignore_patterns = ["Focus Time", "Lunch", "OOO"]

[prs]
enabled = true
repos = [
    "mrf/beads-synced",
    "mrf/branch-cleaner",
    "mrf/hayden-cloud",
]
# If empty, uses `gh pr status` for current repo context

[ports]
enabled = true
hide_system = true
hide_ephemeral = true
hidden_processes = ["rapportd", "ControlCenter", "mDNSResponder"]

[git]
enabled = true
scan_depth = 2  # How deep to scan for git repos
ignore_dirs = ["node_modules", ".git", "target", "vendor"]
```

## Core Data Structures

```rust
use chrono::{DateTime, Local};

#[derive(Debug, Clone)]
pub struct AppState {
    pub meetings: MeetingsPanel,
    pub prs: PrsPanel,
    pub ports: PortsPanel,
    pub git: GitPanel,
    pub last_refresh: DateTime<Local>,
    pub selected_panel: Panel,
}

#[derive(Debug, Clone)]
pub struct MeetingsPanel {
    pub next_meeting: Option<Meeting>,
    pub upcoming: Vec<Meeting>,
    pub buffer_minutes: Option<i64>,
    pub status: MeetingStatus,
}

#[derive(Debug, Clone)]
pub struct Meeting {
    pub title: String,
    pub start: DateTime<Local>,
    pub end: DateTime<Local>,
    pub calendar: String,
}

#[derive(Debug, Clone)]
pub enum MeetingStatus {
    Free { minutes_until: i64 },
    InMeeting { ends_in: i64 },
    Clear, // No more meetings today
}

#[derive(Debug, Clone)]
pub struct PrsPanel {
    pub needs_review: Vec<PullRequest>,
    pub your_prs: Vec<PullRequest>,
}

#[derive(Debug, Clone)]
pub struct PullRequest {
    pub number: u32,
    pub title: String,
    pub repo: String,
    pub url: String,
    pub age_days: i64,
    pub review_decision: Option<String>,
}

#[derive(Debug, Clone)]
pub struct PortsPanel {
    pub ports: Vec<PortInfo>,
}

#[derive(Debug, Clone)]
pub struct PortInfo {
    pub port: u16,
    pub process: String,
    pub pid: u32,
}

#[derive(Debug, Clone)]
pub struct GitPanel {
    pub dirty_repos: Vec<DirtyRepo>,
    pub stashes: Vec<StashInfo>,
}

#[derive(Debug, Clone)]
pub struct DirtyRepo {
    pub name: String,
    pub path: String,
    pub modified: usize,
    pub untracked: usize,
    pub staged: usize,
}

#[derive(Debug, Clone)]
pub struct StashInfo {
    pub repo: String,
    pub index: usize,
    pub message: String,
    pub age_days: i64,
}
```

## Data Collection

### Meetings (osascript)

```rust
pub fn fetch_meetings() -> Result<MeetingsPanel> {
    let script = r#"
        tell application "Calendar"
            set now to current date
            set endTime to now + (8 * hours)
            set output to ""
            repeat with cal in calendars
                if name of cal is not in {"Birthdays", "US Holidays"} then
                    set evts to (every event of cal whose start date ≥ now and start date ≤ endTime)
                    repeat with evt in evts
                        set output to output & (summary of evt) & "|||"
                        set output to output & (start date of evt) & "|||"
                        set output to output & (end date of evt) & "|||"
                        set output to output & (name of cal) & "
"
                    end repeat
                end if
            end repeat
            return output
        end tell
    "#;

    let output = Command::new("osascript")
        .arg("-e")
        .arg(script)
        .output()?;

    parse_meetings(&String::from_utf8_lossy(&output.stdout))
}
```

### PRs (gh cli)

```rust
pub fn fetch_prs(repos: &[String]) -> Result<PrsPanel> {
    let mut needs_review = Vec::new();
    let mut your_prs = Vec::new();

    for repo in repos {
        let output = Command::new("gh")
            .args(["pr", "status", "-R", repo, "--json",
                   "number,title,url,createdAt,reviewDecision"])
            .output()?;

        let status: GhPrStatus = serde_json::from_slice(&output.stdout)?;

        for pr in status.needs_review {
            needs_review.push(pr.into());
        }
        for pr in status.created_by {
            your_prs.push(pr.into());
        }
    }

    Ok(PrsPanel { needs_review, your_prs })
}
```

### Ports (lsof)

```rust
pub fn fetch_ports() -> Result<PortsPanel> {
    let output = Command::new("lsof")
        .args(["-i", "-P", "-n"])
        .output()?;

    let stdout = String::from_utf8_lossy(&output.stdout);
    let ports = stdout
        .lines()
        .filter(|line| line.contains("LISTEN"))
        .filter_map(parse_lsof_line)
        .filter(|p| !is_system_port(p))
        .collect();

    Ok(PortsPanel { ports })
}

fn parse_lsof_line(line: &str) -> Option<PortInfo> {
    let parts: Vec<&str> = line.split_whitespace().collect();
    if parts.len() < 9 { return None; }

    let process = parts[0].to_string();
    let pid = parts[1].parse().ok()?;
    let name = parts.get(8)?;

    let port = name
        .split(':')
        .last()?
        .trim_end_matches("(LISTEN)")
        .trim()
        .parse()
        .ok()?;

    Some(PortInfo { port, process, pid })
}
```

### Git Status (scanning)

```rust
pub fn fetch_git_status(projects_dir: &Path, depth: usize) -> Result<GitPanel> {
    let mut dirty_repos = Vec::new();
    let mut stashes = Vec::new();

    for repo_path in find_git_repos(projects_dir, depth)? {
        // Check for uncommitted changes
        let status = Command::new("git")
            .args(["status", "--porcelain"])
            .current_dir(&repo_path)
            .output()?;

        let changes = parse_git_status(&String::from_utf8_lossy(&status.stdout));
        if changes.has_changes() {
            dirty_repos.push(DirtyRepo {
                name: repo_path.file_name().unwrap().to_string_lossy().into(),
                path: repo_path.display().to_string(),
                modified: changes.modified,
                untracked: changes.untracked,
                staged: changes.staged,
            });
        }

        // Check for stashes
        let stash_output = Command::new("git")
            .args(["stash", "list"])
            .current_dir(&repo_path)
            .output()?;

        for stash in parse_stash_list(&String::from_utf8_lossy(&stash_output.stdout)) {
            stashes.push(StashInfo {
                repo: repo_path.file_name().unwrap().to_string_lossy().into(),
                ..stash
            });
        }
    }

    Ok(GitPanel { dirty_repos, stashes })
}
```

## UI Layout (Ratatui)

```rust
use ratatui::{
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Style},
    widgets::{Block, Borders, Paragraph, Row, Table},
    Frame,
};

pub fn draw(f: &mut Frame, app: &AppState) {
    // Main layout: top row (3 panels) + bottom section (git)
    let main_chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Percentage(45),  // Top panels
            Constraint::Percentage(35),  // Uncommitted work
            Constraint::Percentage(15),  // Stashes
            Constraint::Min(3),          // Status bar
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

    draw_meetings_panel(f, top_chunks[0], &app.meetings);
    draw_prs_panel(f, top_chunks[1], &app.prs);
    draw_ports_panel(f, top_chunks[2], &app.ports);
    draw_git_panel(f, main_chunks[1], &app.git.dirty_repos);
    draw_stash_panel(f, main_chunks[2], &app.git.stashes);
    draw_status_bar(f, main_chunks[3], app);
}

fn draw_meetings_panel(f: &mut Frame, area: Rect, meetings: &MeetingsPanel) {
    let color = match &meetings.status {
        MeetingStatus::Free { minutes_until } if *minutes_until > 60 => Color::Green,
        MeetingStatus::Free { minutes_until } if *minutes_until > 30 => Color::Yellow,
        MeetingStatus::Free { minutes_until } if *minutes_until > 10 => Color::Rgb(255, 165, 0),
        MeetingStatus::Free { .. } => Color::Red,
        MeetingStatus::InMeeting { .. } => Color::Blue,
        MeetingStatus::Clear => Color::Green,
    };

    let block = Block::default()
        .title("⏱ MEETINGS")
        .borders(Borders::ALL)
        .border_style(Style::default().fg(color));

    let text = match &meetings.status {
        MeetingStatus::Free { minutes_until } => {
            let next = meetings.next_meeting.as_ref().unwrap();
            format!("🟢 {}m until:\n   {}", minutes_until, next.title)
        }
        MeetingStatus::InMeeting { ends_in } => {
            format!("🔵 IN MEETING\n   Ends in {}m", ends_in)
        }
        MeetingStatus::Clear => "🟢 Clear for the day!".to_string(),
    };

    let paragraph = Paragraph::new(text).block(block);
    f.render_widget(paragraph, area);
}
```

## Event Loop

```rust
use crossterm::event::{self, Event, KeyCode};
use std::time::{Duration, Instant};

pub async fn run(mut app: AppState, config: Config) -> Result<()> {
    let mut terminal = setup_terminal()?;
    let tick_rate = Duration::from_secs(1);
    let refresh_rate = Duration::from_secs(config.general.refresh_interval_seconds);
    let mut last_refresh = Instant::now();

    loop {
        terminal.draw(|f| ui::draw(f, &app))?;

        // Handle input with timeout
        if event::poll(tick_rate)? {
            if let Event::Key(key) = event::read()? {
                match key.code {
                    KeyCode::Char('q') => break,
                    KeyCode::Char('r') => {
                        app = refresh_all(&config).await?;
                        last_refresh = Instant::now();
                    }
                    KeyCode::Char('p') => open_pr_in_browser(&app.prs)?,
                    KeyCode::Char('m') => open_calendar()?,
                    KeyCode::Char('g') => open_git_repo(&app.git)?,
                    _ => {}
                }
            }
        }

        // Auto-refresh
        if last_refresh.elapsed() >= refresh_rate {
            app = refresh_all(&config).await?;
            last_refresh = Instant::now();
        }

        // Update "time ago" displays every tick
        app.update_relative_times();
    }

    restore_terminal()?;
    Ok(())
}
```

## Color Coding

| Element | Condition | Color |
|---------|-----------|-------|
| Meeting buffer | > 60m | 🟢 Green |
| Meeting buffer | 30-60m | 🟡 Yellow |
| Meeting buffer | 10-30m | 🟠 Orange |
| Meeting buffer | < 10m | 🔴 Red |
| Meeting buffer | In meeting | 🔵 Blue |
| PR age | < 2 days | 🟢 Green |
| PR age | 2-5 days | 🟡 Yellow |
| PR age | > 5 days | 🟠 Orange |
| PR age | > 7 days | 🔴 Red |
| Stash age | < 7 days | Normal |
| Stash age | 7-30 days | 🟡 Yellow |
| Stash age | > 30 days | 🪦 Ancient |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `q` | Quit |
| `r` | Refresh all panels |
| `p` | Open selected PR in browser |
| `m` | Open Calendar app |
| `g` | Open selected repo in terminal/editor |
| `s` | Focus stashes panel |
| `↑/↓` | Navigate within panel |
| `Tab` | Switch panels |

## MVP Scope (2 hours)

1. ✅ Basic Ratatui layout with 4 panels
2. ✅ Meeting data from osascript
3. ✅ PR data from gh cli (single repo)
4. ✅ Port data from lsof
5. ✅ Git status scanning for dirty repos
6. ✅ Basic color coding
7. ✅ Refresh on `r` key
8. ⏭️ Stash detection (stretch goal)
9. ⏭️ Config file (stretch - hardcode first)
10. ⏭️ Keyboard navigation (stretch)

## Building

```bash
# Development
cargo run

# Release (optimized)
cargo build --release

# Install
cargo install --path .

# Binary at
./target/release/dcc
```

## Why This Is The Capstone

This project combines:
- **meeting-buffer** (Day 9) - Calendar integration
- **pr-dashboard** (Day 4) - PR status via gh
- **port-watcher** (Day 8) - lsof parsing
- **Uncommitted scanner** - New, addresses real pain point
- **Stash finder** - New, solves forgotten work problem

One command, full developer context. The perfect finale for 12 Days of Shipmas.

## Future Ideas

- [ ] Customizable panel layout
- [ ] Notifications when state changes
- [ ] History/trending (are PRs aging faster?)
- [ ] Integration with IDE (open file at line)
- [ ] Team mode (see team's PRs too)
- [ ] Export daily summary to Slack/Discord
