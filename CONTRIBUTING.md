# Contributing

Thanks for your interest in contributing.

## Development setup

1. Install Go 1.24+ and Git.
2. Clone the repo.
3. Run:

```bash
go build ./...
go test ./...
```

## Workflow

1. Create a branch from `main`.
2. Make focused changes.
3. Add or update tests where relevant.
4. Run `go test ./...` and `go build ./...`.
5. Open a pull request.

## Commit style

Conventional commits are preferred:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation
- `test:` for tests
- `chore:` for maintenance

## Pull requests

Please include:
- What changed
- Why it changed
- How it was tested

Small, focused PRs are easier to review.
