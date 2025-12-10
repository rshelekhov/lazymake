# lazymake

A beautiful terminal user interface for browsing and executing Makefile targets.

## Context

Make dominates build automation with 19% presence in top GitHub repos, but developers describe it as "fragile, dated, and anti-human by modern dev ergonomics." While Make is powerful and ubiquitous, its poor developer experience creates friction, especially for teams onboarding new developers or working with complex Makefiles.

## Problems We Solve

- **Poor discoverability**: Finding and understanding available Makefile targets requires reading the entire Makefile
- **Dependency confusion**: 70% of development teams struggle with managing dependencies; over 60% of compilation delays stem from misconfigured dependencies
- **Bad onboarding**: New developers face a steep learning curve with undocumented Makefile targets
- **Lack of visibility**: No easy way to see execution time, dependencies, or what commands will actually run
- **Frustrating errors**: Common issues like "missing separator" are cryptic and hard to debug

## Features

### ✅ Implemented
- **Self-documenting help system**: Automatically extracts and displays comments from Makefile targets
  - Supports industry-standard `##` comments for documentation
  - Backward compatible with single `#` comments
  - Inline comments (e.g., `build: ## Build the app`) take priority
  - Press `?` to toggle help view showing all documented targets
  - Visual distinction: cyan for `##` documented targets, gray for regular comments

- **Dependency graph visualization**: Interactive ASCII tree showing target dependencies
  - Press `g` on any target to view its dependency graph
  - Execution order numbering `[N]` - shows the order targets will run
  - Critical path markers `★` - highlights the longest dependency chain
  - Parallel opportunities `∥` - identifies targets that can run concurrently
  - Configurable depth control with `+/-` keys
  - Toggle annotations on/off: `o` (order), `c` (critical), `p` (parallel)
  - Smart detection: only marks meaningful build chains, not standalone targets
  - Cycle detection with clear warnings for circular dependencies

### Planned

#### High Priority
- **Search & filtering**: Real-time fuzzy search for targets, filter by recently used or favorites
- **Performance profiling**: Track execution time and build history to identify slow targets
- **Better error handling**: Parse and highlight common Makefile errors with helpful suggestions

### Nice to Have
- Multi-language recipe support with syntax highlighting
- Workspace/project management for monorepos
- Variable inspector and runtime overrides
- Dry-run preview with warnings for destructive operations
- Watch mode and CI/CD integration

## Installation

### macOS/Linux (Homebrew)
```bash
brew install rshelekhov/tap/lazymake
```

### Linux (apt)
```bash
# Download the .deb from releases page
wget https://github.com/rshelekhov/lazymake/releases/download/v0.1.0/lazymake_0.1.0_Linux_x86_64.deb
sudo dpkg -i lazymake_0.1.0_Linux_x86_64.deb
```

### Linux (yum/rpm)
```bash
# Download the .rpm from releases page
wget https://github.com/rshelekhov/lazymake/releases/download/v0.1.0/lazymake_0.1.0_Linux_x86_64.rpm
sudo rpm -i lazymake_0.1.0_Linux_x86_64.rpm
```

### Go Developers
```bash
go install github.com/rshelekhov/lazymake/cmd/lazymake@latest
```

### From Source
```bash
git clone https://github.com/rshelekhov/lazymake.git
cd lazymake
make install  # Installs to $GOPATH/bin
```

Or install system-wide:
```bash
make install-system  # Installs to /usr/local/bin (requires sudo)
```

## Usage

```bash
# Run with default Makefile
lazymake

# Specify a different Makefile
lazymake -f path/to/Makefile

# Use a custom theme
lazymake -t <theme-name>
```

### Keyboard Shortcuts

#### Main List View
- `↑/↓` or `j/k` - Navigate targets
- `Enter` - Execute selected target
- `g` - View dependency graph for selected target
- `?` - Toggle help view
- `/` - Filter/search targets
- `q` or `ctrl+c` - Quit

#### Graph View
- `g` or `esc` - Return to list view
- `+` or `=` - Show more dependency levels
- `-` or `_` - Show fewer dependency levels
- `o` - Toggle execution order numbers `[N]`
- `c` - Toggle critical path markers `★`
- `p` - Toggle parallel opportunity markers `∥`
- `q` or `ctrl+c` - Quit

#### Output View
- `↑/↓` or `j/k` - Scroll through output
- `esc` - Return to list view
- `q` or `ctrl+c` - Quit

Configuration can be set via `.lazymake.yaml` in your project directory.

## Writing Self-Documenting Makefiles

lazymake supports the industry-standard `##` convention for documenting Makefile targets. Use `##` comments to mark targets for documentation:

```makefile
.PHONY: build test deploy

build: ## Build the application
	go build -o app main.go

test: ## Run all tests
	go test ./...

## Deploy to production
deploy:
	./scripts/deploy.sh
```

**Comment styles:**
- `##` - Industry standard for documentation (shown in cyan)
- `#` - Regular comments (shown in gray, backward compatible)
- Inline comments (after `:`) override preceding comments

**Best practices:**
- Use `##` for targets you want to document for other developers
- Use `#` for internal implementation notes
- Keep descriptions concise (one line)
- Place inline comments for quick reference: `target: ## Description`

## Understanding Dependency Graphs

lazymake visualizes your Makefile's dependency structure to help you understand execution flow and optimize build times.

### Example Graph

```
all [3] ★
├── build [2] ★ ∥
│   └── deps [1] ★
└── test [2] ★ ∥
    └── deps [1] (see above)
```

### Annotations Explained

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

### Smart Detection

lazymake intelligently identifies meaningful patterns:
- **Standalone targets** (like `clean`, `lint`) are shown without markers
- **Circular dependencies** are detected and displayed with warnings
- **Shared dependencies** are marked with `(see above)` to avoid duplication

### Use Cases

1. **Onboarding**: New developers can see the build structure instantly
2. **Optimization**: Identify bottlenecks in your build process
3. **Debugging**: Understand why certain targets run before others
4. **Parallelization**: Find opportunities to speed up builds with `-j`
