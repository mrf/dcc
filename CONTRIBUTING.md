# Contributing to Dev Command Center

Thank you for your interest in contributing to Dev Command Center! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR-USERNAME/dev-command-center.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests and lints (see below)
6. Commit your changes
7. Push to your fork and submit a Pull Request

## Development Setup

### Prerequisites

- Rust 1.70 or later
- macOS (required for Calendar.app integration)
- GitHub CLI (`gh`) installed and authenticated
- `cargo-fmt` and `cargo-clippy` (included with Rust)

### Building

```bash
# Build in debug mode
cargo build

# Build in release mode
cargo build --release

# Run in development
cargo run
```

### Testing

```bash
# Run all tests
cargo test

# Run tests with output
cargo test -- --nocapture

# Run a specific test
cargo test test_name
```

### Code Quality

Before submitting a PR, ensure your code passes all checks:

```bash
# Format code
cargo fmt

# Check formatting without modifying
cargo fmt --check

# Run linter
cargo clippy -- -D warnings

# Run all checks
cargo fmt --check && cargo clippy -- -D warnings && cargo test
```

## Code Style

- Follow standard Rust conventions and idioms
- Use `cargo fmt` to format code
- Address all `clippy` warnings
- Write descriptive commit messages
- Add tests for new functionality
- Keep functions focused and small

## Project Structure

```
src/
├── main.rs          # Entry point, event loop
├── app.rs           # Application state management
├── ui.rs            # Layout and rendering
├── data/            # Data fetching modules
│   ├── mod.rs
│   ├── config.rs    # Configuration handling
│   ├── meetings.rs  # Calendar.app integration
│   ├── prs.rs       # GitHub PR fetching
│   ├── ports.rs     # Port scanning
│   └── git.rs       # Git status/stash scanning
└── widgets/         # UI widget modules
    ├── mod.rs
    ├── meetings.rs
    ├── prs.rs
    ├── ports.rs
    └── git.rs
```

## Adding a New Panel

1. Create a data module in `src/data/` for fetching the data
2. Create a widget module in `src/widgets/` for rendering
3. Add the panel to `AppState` in `src/app.rs`
4. Add the panel to `Panel` enum for navigation
5. Update the layout in `src/ui.rs`
6. Add configuration options if needed

## Reporting Issues

When reporting issues, please include:

- Your macOS version
- Rust version (`rustc --version`)
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Any error messages

## Pull Request Guidelines

- Keep PRs focused on a single change
- Update documentation if needed
- Add tests for new functionality
- Ensure all CI checks pass
- Write a clear PR description

## Questions?

Feel free to open an issue for questions or discussion about potential changes.
