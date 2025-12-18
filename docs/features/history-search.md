# Recent History & Smart Search

lazymake tracks your most frequently used targets to speed up repetitive workflows. When you run the same 2-3 targets regularly (like `test`, `build`, `deploy`), they appear at the top of the list for instant access.

## How It Works

- **Automatic tracking**: Every time you execute a target, it's recorded in history
- **Top 5 recent**: The last 5 executed targets appear in the "RECENT" section
- **Visual indicator**: Recent targets are marked with a ⏱ clock emoji
- **Per-project**: Each Makefile maintains separate history
- **Smart cleanup**: Targets removed from the Makefile are automatically filtered out
- **Persistent**: History survives across sessions

## Example Display

```
RECENT
⏱  test        Run all tests                          8.1s
⏱  build       Build the application                  3.4s
⏱  lint        Run linters                            0.3s
──────────────────────────────────
ALL TARGETS
▶ build        Build the application
  clean        Clean build artifacts
📈 test        Run all tests                          8.1s
  deploy       Deploy to production
```

**Indicators:**
- ⏱ = Recently executed (with duration badge)
- 📈 = Performance regression detected (>25% slower)
- Duration badges color-coded: green (<1s), cyan (normal), orange (regressed)

## Search & Filtering

Press `/` to activate search mode and filter targets by name or description:

- **Real-time fuzzy search**: Type to filter targets instantly
- **Search in names and descriptions**: Matches both target names and documentation
- **Case-insensitive**: No need to worry about capitalization
- **Clear indicator**: Search query shown at top of list
- **Quick clear**: Press `Esc` to clear search and return to full list

## Benefits

- **Faster workflows**: No need to scroll or type to find common targets
- **Context awareness**: Different projects show different recent targets
- **Zero configuration**: Works automatically from the first execution
- **Performance insights**: See execution times and spot regressions

## Storage

History is stored per-Makefile in:
```
~/.cache/lazymake/history.json
```

The file contains:
- Target name
- Makefile path
- Execution timestamps
- Duration history (last 10 executions)
- Performance statistics (average, min, max)

## Keyboard Shortcuts

- **`/`**: Enter search/filter mode
- **`↑/↓` or `j/k`**: Navigate targets (skips section headers)
- **`Esc`**: Clear search and return to full list
- **`Enter`**: Execute selected target

---

[← Back to Documentation](../README.md) | [← Back to Main README](../../README.md)
