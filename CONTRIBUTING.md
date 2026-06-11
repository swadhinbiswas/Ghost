# Contributing to Ghost

Thanks for your interest in improving Ghost! This guide covers everything you need to get productive quickly.

## Prerequisites

- **Go 1.26+** (the version in `go.mod` is authoritative)
- **Git**
- **[Bun](https://bun.sh/)** — only needed if you work on the website in `site/`
- **Docker** (optional) — enables the sandboxed shell execution path

## Getting started

```bash
git clone https://github.com/swadhinbiswas/Ghost.git
cd Ghost
go build ./...
go test ./...
```

Run the CLI from source:

```bash
go run ./main.go
```

## Project layout

| Path | What lives here |
| --- | --- |
| `main.go` | Entry point — delegates to `internal/cmd`. |
| `internal/cmd` | CLI commands (`run`, `free`, `login`, `models`, …). |
| `internal/agent` | The agent loop and its tools (`bash`, `edit`, `grep`, MCP, …). |
| `internal/config` | Config loading, providers, model catalogs. |
| `internal/ui` | The terminal UI (Bubble Tea / Lipgloss). |
| `internal/db` | Generated data layer (sqlc) — do not edit by hand. |
| `site/` | Astro website + docs (built with Bun). |
| `nvidia_nim_models.json` | Remotely-updatable Nvidia NIM model catalog. |

## Development workflow

1. Create a branch from `main`: `git checkout -b feat/short-description`.
2. Make your change. Keep the diff focused.
3. Run the local checks below — they mirror CI.
4. Open a pull request against `main` with a clear description of what and why.

### Local checks (must pass)

```bash
go build ./...
go vet ./...
go test ./...
gofmt -l .            # should print nothing
```

If you have [golangci-lint](https://golangci-lint.run/) installed:

```bash
golangci-lint run
```

### Website checks

```bash
cd site
bun install
bun run build
```

## Coding guidelines

- Match the surrounding code style; prefer the standard library and the existing dependencies.
- Use secure-by-default patterns: validate input, handle errors, never log secrets or API keys.
- Add tests for new behavior and bug fixes — especially in `internal/agent`, `internal/config`, and `internal/shell`.
- Keep commits small and messages descriptive (Conventional Commits encouraged: `feat:`, `fix:`, `docs:`, `chore:`).

## Adding or updating models

The Nvidia NIM catalog is driven by `nvidia_nim_models.json` at the repo root. Edit that file and open a PR — the change ships to users without a new binary release once merged. The embedded fallback in `internal/config/local_providers.go` should be kept in sync.

## Reporting bugs & requesting features

Use the GitHub issue templates. For security issues, **do not** open a public issue — see [SECURITY.md](SECURITY.md).

## License

By contributing, you agree that your contributions are licensed under the [MIT License](LICENSE).
