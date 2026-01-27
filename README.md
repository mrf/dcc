# Dev Command Center (dcc)

A terminal dashboard that shows your complete developer context at a glance. One command, full context.

![Go](https://img.shields.io/badge/go-1.21+-00ADD8)
![License](https://img.shields.io/badge/license-MIT-blue)

## Why

Context switching is expensive. Before starting any task, you need to know:

- How long until my next meeting?
- Any PRs need my attention?
- What's running on my ports?
- Do I have uncommitted work somewhere?

**One glance. One command. Full context.**

## Screenshot

```
┌─ MEETINGS ──────────┬─ PRS ────────────────────────┬─ PORTS ──────────┐
│ ● 47 min until:     │ Needs Review (2):            │ :3000  node      │
│   1:1 with Sarah    │   #421 Add auth  2d          │ :5432  postgres  │
│                     │   #418 Fix bug   4d          │ :6379  redis     │
│ Then:               │                              │ :8080  python    │
│   2:00 PM Sprint    │ Your PRs (1):                │                  │
│   3:30 PM Eng Sync  │   #419 Refactor  1d ✓        │                  │
├─────────────────────┴──────────────────────────────┴──────────────────┤
│ UNCOMMITTED WORK                                                      │
│ cloud                2 modified, 1 untracked                          │
│ pr-dashboard         5 untracked files                                │
├───────────────────────────────────────────────────────────────────────┤
│ STASHES                                                               │
│ branch-cleaner (2)   WIP auth refactor (3d)                           │
│ runbook-gen (1)      broken experiment (ancient)                      │
├───────────────────────────────────────────────────────────────────────┤
│ [r]efresh  [p]rs  [m]eetings  [g]it  [Tab] switch  [q]uit  Updated: 5s│
└───────────────────────────────────────────────────────────────────────┘
```

## Installation

### From source

```bash
# Clone the repository
git clone https://github.com/mrf/dcc.git
cd dcc

# Build
make build

# Or install to ~/bin
make install

# Or run directly
make run
```

### Requirements

- **Go** 1.21+
- **macOS** (Calendar integration uses AppleScript; other platforms show empty meetings panel)
- **GitHub CLI** (`gh`) - for PR status
- **git** - for repository scanning

## Usage

```bash
# Run the dashboard
dcc

# Or with make
make run

# Keyboard shortcuts
q / Esc    - Quit
r          - Refresh all panels
p          - Open first PR in browser
m          - Open Calendar app
g          - Open first dirty repo in VS Code
Tab        - Switch between panels
Shift+Tab  - Switch panels (reverse)
```

## Configuration

Create `~/.config/dcc/config.toml`:

```toml
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

[ports]
enabled = true
hide_system = true
hide_ephemeral = true
hidden_processes = ["rapportd", "ControlCenter", "mDNSResponder"]

[git]
enabled = true
scan_depth = 2
ignore_dirs = ["node_modules", ".git", "target", "vendor"]
```

## Data Sources

| Panel | Source | Command |
|-------|--------|---------|
| Meetings | Calendar.app | `osascript` |
| PRs | GitHub CLI | `gh search prs` |
| Ports | lsof | `lsof -i -P -n` |
| Git Status | git | `git status --porcelain` |
| Stashes | git | `git stash list` |

## Color Coding

### Meeting Buffer
- Green: > 60 minutes
- Yellow: 30-60 minutes
- Orange: 10-30 minutes
- Red: < 10 minutes
- Blue: Currently in meeting

### PR Age
- Green: < 2 days
- Yellow: 2-5 days
- Orange: 5-7 days
- Red: > 7 days

### Stash Age
- Gray: < 7 days
- Yellow: 7-30 days
- Red: > 30 days (ancient)

## Development

```bash
# Run in development mode
make dev

# Run tests
go test ./...

# Build release
make build

# Format code
go fmt ./...

# Lint
go vet ./...
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
