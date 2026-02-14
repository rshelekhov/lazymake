# GitHub Issues Backlog

This document contains a ready-to-create backlog of GitHub issues for lazymake.
All issues are written in English and include milestone, labels, summary, and acceptance criteria.

## 1) Shell history templates: support {target}, {makefile}, and {dir}

- Milestone: `v0.4.0`
- Labels: `enhancement`, `area:shell`, `priority:P0`

### Summary
Implement full template variable support for shell history entries: `{target}`, `{makefile}`, and `{dir}`.

### Problem
Current behavior only substitutes `{target}`, while docs describe support for `{makefile}` and `{dir}` as well.

### Acceptance Criteria
- `{target}`, `{makefile}`, and `{dir}` are all substituted in history output.
- Missing/empty values are handled safely without panics.
- Unit tests cover all placeholders and mixed template strings.
- Documentation examples match actual behavior.

## 2) Honor shell_integration.include_timestamp in runtime behavior

- Milestone: `v0.4.0`
- Labels: `enhancement`, `area:shell`, `priority:P0`

### Summary
Make `shell_integration.include_timestamp` actually control timestamp writing behavior.

### Problem
Config contains `include_timestamp`, but behavior is not consistently controlled by this flag.

### Acceptance Criteria
- `include_timestamp: true` writes timestamped entries where format supports it.
- `include_timestamp: false` writes non-timestamped entries.
- Behavior is covered by tests for supported shells.
- Docs explicitly describe shell-specific behavior.

## 3) Add fish shell history integration

- Milestone: `v0.4.0`
- Labels: `enhancement`, `area:shell`, `priority:P0`

### Summary
Implement fish history writer and wire it into shell auto-detection flow.

### Problem
Fish is detected but not supported end-to-end for writing execution history.

### Acceptance Criteria
- Fish writer appends entries in valid fish history format.
- Auto mode initializes fish integration when `$SHELL` is fish.
- File locking/concurrency safety matches other shell writers.
- Integration tests verify append behavior and file output.

## 4) Unify config merge behavior for export and shell_integration

- Milestone: `v0.4.0`
- Labels: `enhancement`, `area:config`, `priority:P0`

### Summary
Implement consistent global + project config merge logic for all config sections, not only safety.

### Problem
Docs describe merged config strategy, but runtime behavior is inconsistent across sections.

### Acceptance Criteria
- Global and project configs merge predictably for `export` and `shell_integration`.
- Override/union rules are clearly defined and documented.
- Tests validate merge precedence and list behavior.
- Breaking changes are documented in changelog.

## 5) Config behavior parity audit: docs vs runtime

- Milestone: `v0.4.0`
- Labels: `documentation`, `area:config`, `priority:P1`

### Summary
Audit and align all documented config options with runtime behavior.

### Acceptance Criteria
- Every documented config key is either implemented or removed from docs.
- Defaults in docs match code defaults.
- Example config file is validated against actual parser behavior.
- A regression test checklist is added for future config changes.

## 6) Add dry-run/explain mode for target execution

- Milestone: `v0.5.0`
- Labels: `enhancement`, `area:tui`, `priority:P0`

### Summary
Provide a dry-run mode to preview commands before execution (e.g. via `make -n`).

### Acceptance Criteria
- User can trigger dry-run from the target list view.
- Dry-run output is shown in a dedicated preview/output view.
- No target command is executed in dry-run mode.
- UI communicates clearly whether user is in dry-run vs real execution.

## 7) Interactive execution parameters (ENV vars and make flags)

- Milestone: `v0.5.0`
- Labels: `enhancement`, `area:tui`, `priority:P0`

### Summary
Add a pre-run prompt to pass environment variables and make flags per execution.

### Acceptance Criteria
- User can input key/value variables (e.g. `ENV=prod`, `TAG=v1`).
- User can set run flags (e.g. `-j4`, `--always-make`).
- Command construction is safe and deterministic.
- Cancel path returns to list view without execution.

## 8) Saved run presets per target

- Milestone: `v0.5.0`
- Labels: `enhancement`, `area:history`, `priority:P1`

### Summary
Allow saving and reusing named parameter presets per target and Makefile.

### Acceptance Criteria
- Presets are stored per workspace/Makefile.
- User can create, select, update, and delete presets.
- Last-used preset is quickly reusable.
- Preset storage survives restart and handles missing targets gracefully.

## 9) Export dependency graph to DOT and Mermaid

- Milestone: `v0.5.0`
- Labels: `enhancement`, `area:graph`, `priority:P1`

### Summary
Add graph export options for documentation and CI artifacts.

### Acceptance Criteria
- Export graph to DOT format.
- Export graph to Mermaid format.
- Exports include cycle and missing dependency annotations when present.
- CLI and/or TUI entry points are documented.

## 10) Rerun last target with previous parameters

- Milestone: `v0.5.0`
- Labels: `enhancement`, `area:tui`, `priority:P1`

### Summary
Add shortcut to re-run the most recent execution with same parameters.

### Acceptance Criteria
- Shortcut is available from main list view.
- Re-run includes previous env vars and flags.
- UI confirms what will be re-run before starting.
- Works across app restart when history exists.

## 11) Parser v2: hybrid model using make metadata

- Milestone: `v0.6.0`
- Labels: `enhancement`, `area:parser`, `priority:P0`

### Summary
Introduce a hybrid parsing strategy combining static parse with `make` metadata (e.g. `make -qp`) for better compatibility.

### Acceptance Criteria
- Parser resolves more real-world targets/dependencies than current static-only behavior.
- Fallback path exists when external `make` metadata fails.
- Performance remains acceptable on medium/large Makefiles.
- Compatibility metrics are tracked in tests.

## 12) Compatibility fixture suite for complex Makefiles

- Milestone: `v0.6.0`
- Labels: `testing`, `area:parser`, `priority:P1`

### Summary
Add fixture-based tests covering complex Makefile patterns from real projects.

### Acceptance Criteria
- Fixtures include includes, conditional blocks, pattern rules, and target-specific vars.
- Expected target/dependency snapshots are versioned in tests.
- CI runs fixture suite on every parser-related change.
- Failures show clear diff for target graph regressions.
