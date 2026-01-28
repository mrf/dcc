# Contributing to Dev Command Center

Thank you for your interest in contributing to dcc!

## Development Setup

1. **Prerequisites**
   - Go 1.21 or later
   - GitHub CLI (`gh`) for testing PR features
   - macOS for testing Calendar integration (optional)

2. **Clone and build**
   ```bash
   git clone https://github.com/mrf/dcc.git
   cd dcc
   make build
   ```

3. **Run in development mode**
   ```bash
   make dev
   ```

## Code Style

- Run `go fmt ./...` before committing
- Run `go vet ./...` to check for issues
- Keep functions small and focused
- Add tests for parsing logic

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/data/...
```

## Project Structure

```
cmd/dcc/           # Entry point
internal/
  app/             # Bubbletea model, update, view
  config/          # Configuration loading
  data/            # Data fetching (meetings, prs, ports, git)
  ui/              # Lipgloss styles and panel rendering
```

## Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit with a clear message
6. Push and open a PR

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
