mod app;
mod data;
mod ui;
mod widgets;

use anyhow::Result;
use crossterm::{
    event::{self, Event, KeyCode, KeyEventKind},
    execute,
    terminal::{
        disable_raw_mode, enable_raw_mode, Clear, ClearType, EnterAlternateScreen,
        LeaveAlternateScreen,
    },
};
use ratatui::{backend::CrosstermBackend, Terminal};
use std::io;
use std::time::{Duration, Instant};

use app::AppState;
use data::Config;

fn main() -> Result<()> {
    // Load configuration
    let config = Config::load().unwrap_or_default();

    // Setup terminal
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen, Clear(ClearType::All))?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    // Create app state
    let mut app = AppState::new();

    // Draw initial UI immediately (shows "loading" state with empty data)
    terminal.draw(|f| ui::draw(f, &app))?;

    // Then fetch data
    let _ = app.refresh(&config);

    // Run the main loop
    let result = run_app(&mut terminal, &mut app, &config);

    // Restore terminal
    disable_raw_mode()?;
    execute!(terminal.backend_mut(), LeaveAlternateScreen)?;
    terminal.show_cursor()?;

    if let Err(err) = result {
        eprintln!("Error: {err:?}");
    }

    Ok(())
}

fn run_app(
    terminal: &mut Terminal<CrosstermBackend<io::Stdout>>,
    app: &mut AppState,
    config: &Config,
) -> Result<()> {
    let tick_rate = Duration::from_millis(100); // Fast polling for responsive keyboard
    let refresh_rate = Duration::from_secs(config.general.refresh_interval_seconds);
    let mut last_refresh = Instant::now();

    loop {
        // Draw UI
        terminal.draw(|f| ui::draw(f, app))?;

        // Handle input with timeout
        if event::poll(tick_rate)? {
            match event::read()? {
                Event::Key(key) => {
                    // Only handle key press events (not release)
                    if key.kind == KeyEventKind::Press {
                        match key.code {
                            KeyCode::Char('q') | KeyCode::Esc => {
                                app.should_quit = true;
                            }
                            KeyCode::Char('r') => {
                                let _ = app.refresh(config);
                                last_refresh = Instant::now();
                            }
                            KeyCode::Char('p') => {
                                open_pr_in_browser(app);
                            }
                            KeyCode::Char('m') => {
                                open_calendar();
                            }
                            KeyCode::Char('g') => {
                                open_git_repo(app);
                            }
                            KeyCode::Tab => {
                                app.next_panel();
                            }
                            KeyCode::BackTab => {
                                app.prev_panel();
                            }
                            _ => {}
                        }
                    }
                }
                Event::Resize(_, _) => {
                    // Terminal resized, just redraw on next loop
                }
                _ => {}
            }
        }

        if app.should_quit {
            return Ok(());
        }

        // Auto-refresh
        if last_refresh.elapsed() >= refresh_rate {
            let _ = app.refresh(config);
            last_refresh = Instant::now();
        }
    }
}

fn open_pr_in_browser(app: &AppState) {
    // Open first PR needing review, or first of your PRs
    let url = app
        .prs
        .needs_review
        .first()
        .or(app.prs.your_prs.first())
        .map(|pr| pr.url.as_str());

    if let Some(url) = url {
        let _ = std::process::Command::new("open").arg(url).spawn();
    }
}

fn open_calendar() {
    let _ = std::process::Command::new("open")
        .arg("-a")
        .arg("Calendar")
        .spawn();
}

fn open_git_repo(app: &AppState) {
    // Open first dirty repo in the default terminal/editor
    if let Some(repo) = app.git.dirty_repos.first() {
        // Try to open in VS Code, fall back to Finder
        let result = std::process::Command::new("code").arg(&repo.path).spawn();

        if result.is_err() {
            let _ = std::process::Command::new("open").arg(&repo.path).spawn();
        }
    }
}
