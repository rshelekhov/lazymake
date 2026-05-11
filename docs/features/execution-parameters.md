# Interactive execution parameters

Lazymake can launch a target with per-execution environment variables and make
flags without forcing you to drop back to the shell. This is useful for
workflows that look like `ENV=prod make deploy -j4` rather than a plain `make
build`.

## Opening the form

From the target list, place the cursor on the target you want to run and press
`e` (Edit & Run). A modal opens with two fields:

- **ENV** — whitespace-separated `KEY=value` pairs. Quote values that contain
  spaces: `MSG="hello world"`.
- **FLAGS** — whitespace-separated make flags, e.g. `-j4 -k --always-make`.
  Every token must start with `-`; if you accidentally type a `KEY=value`
  here, lazymake will tell you to move it to the env field instead.

Switch fields with `Tab` / `Shift+Tab`. Press `Enter` to run, `Esc` to cancel
without any side effects.

## What gets recorded

- The executing screen header shows `make -n` (in dry-run) or `make`, the
  target name, and your env+flags so you can verify the invocation before
  output starts streaming.
- The output view header shows the same params so you can come back to a
  previous run and remember what was applied.
- Execution history (`~/.cache/lazymake/history.json`) stores env and flags
  alongside the timing record under `recent_executions[].env` and
  `recent_executions[].flags`. The fields are omitted when no parameters were
  used, so plain runs serialize exactly as before.
- Export records (when export is enabled) include `env` and `flags` fields in
  the JSON output and `Env:` / `Flags:` lines in the human-readable log.

## Session memory

When you submit the form, lazymake caches the parsed env and flags per target
**for the current session**. Reopening the form on the same target pre-fills
the previous values so you can re-run with a small edit. The cache is in
memory only — restart and the values are forgotten. Persistent presets are
tracked separately under issue #38.

## Interaction with other features

- **Dry-run** (`--dry-run`): if the session is in dry-run mode, your flags are
  combined with `-n`, e.g. `make -n -j4 build`. History, exports, and shell
  integration entries are still skipped, matching the dry-run behavior for
  plain runs.
- **Dangerous targets**: the form opens first, then the existing confirmation
  dialog. This lets you see exactly what env+flags will be applied before
  approving the dangerous command.
- **Quick run**: `Enter` (without the form) still works as before — it runs
  the target with no extra env vars and no flags.

## Safety notes

- Commands are assembled with `exec.CommandContext("make", args...)`. No shell
  is involved, so values are not subject to shell metacharacter
  interpretation. `MSG="hi; rm -rf /"` is delivered to make as a literal
  string, not interpreted.
- Env keys must match `[A-Za-z_][A-Za-z0-9_]*`. Make-style variable overrides
  like `VAR=value` passed on the make CLI are intentionally not supported in
  the flags field — put them in env instead.
- Env keys are sorted before the process is launched, so identical input
  always produces the same `make` invocation, useful for reproducing runs.
