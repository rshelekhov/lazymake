# Dependency Graph Visualization

lazymake visualizes your Makefile's dependency structure to help you understand execution flow and optimize build times.

## Example Graph

```
all [3] ★
├── build [2] ★ ∥
│   └── deps [1] ★
└── test [2] ★ ∥
    └── deps [1] (see above)
```

## Annotations Explained

- **`[N]` Execution Order**: Numbers indicate the order targets will run
  - `[1]` runs first (dependencies with no deps)
  - `[2]` runs second (targets depending on `[1]`)
  - `[3]` runs last (top-level targets)
  - Targets with the same number can run in parallel

- **`★` Critical Path**: Marks the longest chain of dependencies
  - This is the minimum time needed to complete the build
  - Optimizing these targets has the biggest impact on build time
  - Only shown for targets that are part of dependency chains

- **`∥` Parallel Opportunities**: Targets that can run concurrently
  - Make can execute these simultaneously with `-j` flag
  - Example: `make -j4` runs up to 4 targets in parallel
  - Only shown for targets with actual dependencies to coordinate

## Smart Detection

lazymake intelligently identifies meaningful patterns:
- **Standalone targets** (like `clean`, `lint`) are shown without markers
- **Circular dependencies** are detected and displayed with warnings
- **Shared dependencies** are marked with `(see above)` to avoid duplication

## Use Cases

1. **Onboarding**: New developers can see the build structure instantly
2. **Optimization**: Identify bottlenecks in your build process
3. **Debugging**: Understand why certain targets run before others
4. **Parallelization**: Find opportunities to speed up builds with `-j`

## Keyboard Shortcuts

- **`g`**: View dependency graph for selected target (from main list view)
- **`+` or `=`**: Show more dependency levels
- **`-` or `_`**: Show fewer dependency levels
- **`o`**: Toggle execution order numbers `[N]`
- **`c`**: Toggle critical path markers `★`
- **`p`**: Toggle parallel opportunity markers `∥`
- **`g` or `esc`**: Return to list view

---

[← Back to Documentation](../README.md) | [← Back to Main README](../../README.md)
