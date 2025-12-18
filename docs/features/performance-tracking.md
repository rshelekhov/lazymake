# Performance Profiling

lazymake tracks execution time and identifies performance regressions automatically, helping you catch slow builds before they become problems.

## Features

- **Real-time execution timer**: Updates every 100ms with progress indicators
  - 🟢 On track - executing within expected time
  - 🔵 Finishing up - nearing completion
  - 🔴 Slower than usual - taking longer than average

- **Automatic regression detection**: Alerts when targets are >25% slower than average

- **Context-aware display**: Performance info shown only when relevant (regressed or recent targets)

- **Visual indicators**:
  - 📈 for regressed targets in the list
  - Duration badges color-coded by performance

- **Performance history**: Stores last 10 executions per target with statistics:
  - Average duration
  - Minimum duration
  - Maximum duration

- **Post-execution alerts**: Warnings appear after slow runs with actionable insights

- **Persistent tracking**: Performance data survives across sessions

## Example Display

```
RECENT
⏱  test        Run all tests                          8.1s
⏱  build       Build the application                  3.4s
⏱  lint        Run linters                            0.3s
──────────────────────────────────
ALL TARGETS
▶ build        Build the application                  3.4s
  clean        Clean build artifacts
📈 test        Run all tests                          8.1s
  deploy       Deploy to production
```

**Duration badges**:
- **Green** (<1s): Very fast targets
- **Cyan** (normal): Within expected performance
- **Orange** (regressed): 📈 indicator with slower than average duration

## Regression Detection

lazymake automatically detects performance regressions:

1. **Baseline calculation**: Tracks average execution time over last 10 runs
2. **Threshold detection**: Flags executions >25% slower than average
3. **Visual warning**: Shows 📈 indicator next to affected targets
4. **Post-run alert**: Displays warning after slow execution completes

Example alert:
```
⚠️ Performance Alert: 'test' took 8.1s (avg: 6.0s)
Consider investigating what's causing the slowdown.
```

## Use Cases

1. **Catch regressions early**: Spot slow builds immediately
2. **Optimize builds**: Identify which targets need optimization
3. **Track improvements**: Verify optimization efforts are working
4. **Team awareness**: Share performance expectations via .lazymake.yaml

## Storage

Performance data is stored per-target in:
```
~/.cache/lazymake/history.json
```

The data includes:
- Last 10 execution durations
- Statistical summary (avg, min, max)
- Timestamps of executions

## Configuration

Performance tracking is always enabled and requires no configuration. The data is automatically cleaned up when targets are removed from the Makefile.

## Benefits

- **Proactive monitoring**: Catch slowdowns before they impact the team
- **Zero overhead**: Tracking happens automatically with no performance cost
- **Historical context**: See if a target has always been slow or just recently
- **Actionable insights**: Visual indicators make it obvious when something needs attention

---

[← Back to Documentation](../README.md) | [← Back to Main README](../../README.md)
