# Dev Toolkit TUI

A terminal-first developer toolkit built with Go and Bubble Tea.

Current tools:
- Regex Vault: test regex patterns against input strings with live validation.
- Git Analytics: scan a local directory for Git repositories and compute commit activity insights.

## Status

Early alpha (`v0.x`). APIs and keybindings may evolve.

## Features

- Tabbed multi-view TUI
- Responsive layout that adapts to terminal resize
- Async repository scanning (non-blocking UI)
- Live regex compile and match highlighting

## Requirements

- Go 1.24+
- Git installed and available in `PATH`

## Run

```bash
go run main.go
```

## Build

```bash
go build ./...
```

## Test

```bash
go test ./...
go test -race ./...
```

## Keybindings

Global:
- `Tab`: next tab
- `Shift+Tab`: previous tab
- `Ctrl+C` / `Ctrl+X`: quit

Regex Vault:
- `Up` / `Down`: switch between regex and test-string input

Git Analytics:
- Type an absolute directory path
- `Enter`: start async scan

## Open Source

- License: MIT (see `LICENSE`)
- Contributions welcome (see `CONTRIBUTING.md`)
- Code of Conduct: see `CODE_OF_CONDUCT.md`
- Security policy: see `SECURITY.md`

## Roadmap

- Cancelable scans
- Export analytics (JSON/CSV)
- Richer repository metrics
- Improved keybinding help overlay
