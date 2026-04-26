# GitHub Issues Backlog

This document contains ready-to-create issue drafts for the next product phases of lazymake.
The `v0.4.0` backlog has already been delivered in `0.4.0` and `0.4.1`; the issues below focus on open product work.

All issue drafts are written in English so they can be pasted directly into GitHub.

## `v0.5.0` Safe Execution

### 1) Add dry-run / explain mode for target execution

- Milestone: `v0.5.0`
- Labels: `enhancement`, `area:tui`, `area:executor`, `priority:P0`

### Summary
Provide a session-wide dry-run mode so users can preview what `make` would execute without actually running commands.

### Why this matters
Dry-run is the clearest next step for lazymake's safety and onboarding story. It reduces fear when exploring unfamiliar repositories and fits naturally with dangerous target detection.

### Acceptance Criteria
- User can start lazymake in dry-run mode via CLI and config.
- Dry-run execution uses `make -n`.
- UI clearly indicates dry-run mode in list, executing, and output views.
- Dry-run does not write history, exports, or shell integration entries.
- Dangerous targets still show confirmation before dry-run execution.

### Dependencies
- None

### 2) Add interactive execution parameters (environment variables and make flags)

- Milestone: `v0.5.0`
- Labels: `enhancement`, `area:tui`, `area:executor`, `priority:P0`

### Summary
Add a pre-run flow that lets users pass temporary environment variables and make flags per execution.

### Why this matters
Many real workflows are not just `make build`; they are `ENV=prod make deploy -j4`. Without this, users still need to drop back to the shell for important cases.

### Acceptance Criteria
- User can enter key/value environment variables before execution.
- User can enter make flags such as `-j4`, `-k`, or `--always-make`.
- Command construction is safe and deterministic.
- Cancel path returns to the list without execution.
- Output and history clearly reflect the chosen parameters.

### Dependencies
- Should share execution plumbing with dry-run where possible

### 3) Add saved run presets per target and Makefile

- Milestone: `v0.5.0`
- Labels: `enhancement`, `area:tui`, `area:history`, `priority:P1`

### Summary
Allow users to save named presets for a target, including environment variables and make flags.

### Why this matters
Presets turn lazymake from a browser into a repeatable execution console for real project workflows.

### Acceptance Criteria
- Presets are stored per Makefile and target.
- User can create, select, update, and delete presets.
- Last-used preset is easy to reuse.
- Presets survive restart.
- Missing or renamed targets are handled gracefully.

### Dependencies
- Interactive execution parameters

### 4) Add rerun-last flow with previous parameters

- Milestone: `v0.5.0`
- Labels: `enhancement`, `area:tui`, `area:history`, `priority:P1`

### Summary
Add a shortcut to rerun the most recent execution with the same runtime parameters.

### Why this matters
This is the fastest path to making lazymake part of daily development loops.

### Acceptance Criteria
- Shortcut is available from the main list view.
- Rerun includes previous environment variables and make flags.
- UI confirms what is about to run.
- Works after restart when history exists.

### Dependencies
- Interactive execution parameters

### 5) Add TUI flow tests for execution modes and side effects

- Milestone: `v0.5.0`
- Labels: `testing`, `area:tui`, `priority:P1`

### Summary
Add automated tests for the most important execution flows, especially around dry-run and parameterized runs.

### Why this matters
Execution behavior is now central product logic. It should not rely only on manual testing.

### Acceptance Criteria
- Tests cover normal run, dry-run, cancel, and dangerous-target confirmation flows.
- Tests verify side effects are skipped when expected.
- Test structure makes future execution features easier to add safely.

### Dependencies
- Dry-run
- Interactive execution parameters

## `v0.6.0` Compatibility

### 6) Parser v2: hybrid model using static parse plus make metadata

- Milestone: `v0.6.0`
- Labels: `enhancement`, `area:parser`, `priority:P0`

### Summary
Introduce a hybrid parsing strategy that combines static parsing with `make` metadata such as `make -qp`.

### Why this matters
Parser accuracy is the biggest adoption ceiling for more complex or mature repositories.

### Acceptance Criteria
- Parser resolves more real-world targets and dependencies than the current static-only approach.
- Fallback path exists when `make` metadata inspection fails.
- Performance remains acceptable on medium and large Makefiles.
- The trust model is documented.

### Dependencies
- None

### 7) Add compatibility fixture suite for complex Makefiles

- Milestone: `v0.6.0`
- Labels: `testing`, `area:parser`, `priority:P1`

### Summary
Create a fixture-based compatibility test suite from real Makefile patterns.

### Why this matters
Parser work needs product-level regression protection, not only unit-level correctness.

### Acceptance Criteria
- Fixtures include `include`, conditionals, pattern rules, and target-specific variables.
- Expected target and dependency snapshots are versioned in tests.
- CI runs the fixture suite on parser changes.
- Failures show useful diffs.

### Dependencies
- Parser v2

### 8) Surface parser confidence and unsupported constructs in the UI

- Milestone: `v0.6.0`
- Labels: `enhancement`, `area:parser`, `area:tui`, `priority:P1`

### Summary
Make parser uncertainty visible instead of silently presenting incomplete information as fully trusted.

### Why this matters
Trust is product-critical. Clear warnings are better than polished but misleading output.

### Acceptance Criteria
- UI can show when parsing is partial, degraded, or metadata-assisted.
- Unsupported constructs produce understandable hints.
- Warnings do not block normal workflows unless execution would be misleading.

### Dependencies
- Parser v2

### 9) Improve support for included and generated Makefiles

- Milestone: `v0.6.0`
- Labels: `enhancement`, `area:parser`, `priority:P1`

### Summary
Improve discovery and modeling of targets that come from included files or generated Makefile fragments.

### Why this matters
This is a common source of "missing target" confusion in real projects.

### Acceptance Criteria
- Included files are represented more accurately in parsing results.
- Generated or external dependencies degrade gracefully when full resolution is impossible.
- Docs explain supported and unsupported cases.

### Dependencies
- Parser v2

## `v0.7.0` Team and CI Workflows

### 10) Export dependency graphs to DOT and Mermaid

- Milestone: `v0.7.0`
- Labels: `enhancement`, `area:graph`, `priority:P1`

### Summary
Add graph export formats that teams can reuse in docs, PRs, and CI artifacts.

### Why this matters
It extends lazymake beyond the local TUI and strengthens onboarding and architecture communication.

### Acceptance Criteria
- Export graph to DOT format.
- Export graph to Mermaid format.
- Exports include cycle and missing dependency annotations where relevant.
- CLI and documentation are included.

### Dependencies
- None

### 11) Add headless CLI outputs for targets, variables, and graph metadata

- Milestone: `v0.7.0`
- Labels: `enhancement`, `area:cli`, `area:parser`, `priority:P1`

### Summary
Expose selected lazymake insights through machine-readable CLI commands.

### Why this matters
Teams should be able to use lazymake in scripts, docs generation, and CI checks.

### Acceptance Criteria
- CLI can output target metadata in JSON.
- CLI can output graph metadata in JSON or another machine-readable format.
- CLI can output variable metadata in JSON.
- Commands reuse the same parsing rules as the TUI.

### Dependencies
- Parser v2

### 12) Add preset import/export for team sharing

- Milestone: `v0.7.0`
- Labels: `enhancement`, `area:tui`, `area:history`, `priority:P2`

### Summary
Allow presets to be shared across teammates or checked into project-level configuration.

### Why this matters
This turns personal convenience into team workflow standardization.

### Acceptance Criteria
- Presets can be exported and imported in a documented format.
- Team-shared presets can coexist with personal presets.
- Merge and override behavior is explicit.

### Dependencies
- Saved run presets

## `v0.8.0` Intelligence

### 13) Add run diff between current and previous executions

- Milestone: `v0.8.0`
- Labels: `enhancement`, `area:history`, `area:tui`, `priority:P1`

### Summary
Help users compare the latest run with the previous one, including parameters, duration, and output changes.

### Why this matters
This is a practical debugging feature that increases the value of historical data.

### Acceptance Criteria
- User can compare at least the two most recent executions of a target.
- Diff highlights parameter changes and major output changes.
- Works for both successful and failed runs.

### Dependencies
- Parameterized execution

### 14) Add richer performance insights and bottleneck hints

- Milestone: `v0.8.0`
- Labels: `enhancement`, `area:history`, `area:graph`, `priority:P2`

### Summary
Use historical and graph data to surface slow targets, likely bottlenecks, and parallelization opportunities.

### Why this matters
lazymake already stores useful execution context; the next step is to convert it into guidance.

### Acceptance Criteria
- UI can identify consistently slow targets.
- Product can point out possible parallelism or critical-path bottlenecks.
- Insights are visible but non-intrusive.

### Dependencies
- Parser v2
- Stable execution history

### 15) Add smart target grouping based on usage and risk

- Milestone: `v0.8.0`
- Labels: `enhancement`, `area:tui`, `priority:P2`

### Summary
Group targets dynamically by signals such as frequent use, danger level, recent failure, or regression risk.

### Why this matters
This makes the product feel more adaptive without changing the underlying Makefile model.

### Acceptance Criteria
- User can browse grouped views such as recent, risky, slow, or failed.
- Grouping logic is understandable and does not hide targets.
- Default view remains simple for new users.

### Dependencies
- Stable history and performance data
