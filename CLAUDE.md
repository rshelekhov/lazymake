# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

lazymake is an interactive TUI for browsing and executing Makefile targets, built in Go with Bubble Tea.

## Commands

```bash
make build              # Build binary → ./lazymake
make test               # Run all tests
make lint               # Run golangci-lint
make lint-fix           # Auto-fix lint issues
go test ./internal/shell/ -v              # Run one package verbose
go test ./internal/shell/ -run TestZsh    # Run tests matching pattern
go test -race ./...                       # Tests with race detector (CI uses this)
```

## Architecture

**Entry point:** `cmd/lazymake/main.go` — Cobra CLI → `config.Load()` → `tui.NewModel(cfg)` → Bubble Tea event loop.

**TUI (`internal/tui/`)** is the core. `Model` holds all app state; `AppState` enum drives which view renders:
- `StateList` → target browser with fuzzy filtering
- `StateExecuting` → streaming output via `executor.ExecuteStreaming()`
- `StateOutput` → scrollable result viewport
- `StateGraph` → dependency tree visualization
- `StateVariables` → variable inspector
- `StateConfirmDangerous` → safety confirmation dialog
- `StateWorkspace` → multi-Makefile picker

**Execution flow when a target runs:**
1. Safety check → prompt if dangerous
2. `executor.ExecuteStreaming()` → streams output chunks via channel
3. On completion: record in `history`, export via `export.Exporter`, write to `shell.Integration`

**Key packages:**
- `config/` — Viper-based config from `.lazymake.yaml` (project) and `~/.lazymake.yaml` (global), merged
- `internal/makefile/` — Parses targets, dependencies, recipes; handles `define/endef` blocks and `##` comments
- `internal/safety/` — Regex-based dangerous command detection with built-in + custom rules
- `internal/shell/` — `HistoryWriter` interface with `BashWriter` and `ZshWriter` implementations; file locking via platform-specific code (`filelock_unix.go`/`filelock_windows.go`)
- `internal/graph/` — Dependency graph with cycle detection, topological sort, critical path
- `internal/history/` — JSON-persisted execution history with performance regression detection (>25% slower)
- `internal/export/` — JSON/log export with rotation (by count, age, size)
- `internal/highlight/` — Chroma-based syntax highlighting with LRU cache
- `internal/workspace/` — Multi-Makefile workspace tracking

## Conventions

- Features degrade gracefully — if history/workspace/shell-integration fails, the app continues silently
- Config fields often exist in YAML/struct before runtime code uses them; check actual usage when wiring new config
- Shell integration has platform-specific file locking; new shell writers must implement the `HistoryWriter` interface
- Table-driven tests with `t.TempDir()` for file operations
- Commit messages: `feat:`, `fix:`, `docs:` prefixes (conventional commits style)
- **Branching:** Always create a feature branch before starting work — never commit directly to main. Branch format: `feat/short-description` (e.g. `feat/honor-include-timestamp`)
- **Post-implementation:** Always run `make lint` after implementing a feature, before committing. Fix any issues before the commit

## Linting

golangci-lint config is in `.golangci.yml`. Notable settings:
- `gocyclo` min-complexity: 15
- `gosec` excludes G104 (duplicate of errcheck) and G304 (file path taint — too strict for this project)
- Test files have relaxed rules for gocyclo, errcheck, dupl, gosec
