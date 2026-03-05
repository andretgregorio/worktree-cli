# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

`wt` is a Git worktree manager CLI. A Go binary (`wt-bin`) handles logic, and a shell function wrapper (`shell/wt.sh`) captures `__WT_CD__` / `__WT_ENV__` markers from stdout to `cd` and export env vars in the parent shell.

## Build & Test

```bash
go build -o wt-bin .        # build binary
go test ./...                # run all tests
go test ./internal/worktree  # run tests for one package
go test ./... -run TestIntegration  # run only integration tests
go vet ./...                 # lint
```

No linter or formatter beyond `go vet` is configured. Integration tests in `internal/worktree/` create real bare+clone git repos in temp dirs — they need `git` available.

## Architecture

**Two-layer design:** Go binary produces markers on stdout; shell wrapper (`shell/wt.sh`) parses them to change directory and export env vars. All human-readable output goes to stderr so stdout stays clean for marker parsing.

### Packages

- **`cmd/`** — Cobra commands. `root.go` defines the CLI, `create.go` wires the `wt create` flow.
- **`internal/config/`** — Two-tier config: global (`~/.config/wt/config.toml`) sets defaults, per-repo (`~/.config/wt/repos/<name>.toml`) overrides them. Precedence: repo > global > built-in defaults. Expands `~/` paths.
- **`internal/worktree/`** — Git operations: fetch, branch existence check, worktree add. Handles the "branch already exists" case by resetting it with `git branch -f`.
- **`internal/shell/`** — Runs setup commands via `$SHELL -ic` (interactive, so shell functions like `nvm` work). Prints `__WT_CD__` and `__WT_ENV__` markers.

### Key Design Decisions

- Setup commands use `$SHELL -ic` (not `sh -c`) so shell functions like `nvm` are available.
- Config uses `setup = ["cmd1", "cmd2"]` (string array), not `[[setup]]` array-of-tables.
- When a branch already exists, it's force-reset to `origin/<base>` and reused rather than erroring.
- The shell wrapper resolves `wt-bin` relative to its own location (`../wt-bin`), so no PATH changes needed.
