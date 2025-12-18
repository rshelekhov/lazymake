# lazymake

A beautiful terminal user interface for browsing and executing Makefile targets.

## Table of Contents

- [Context](#context)
- [Problems We Solve](#problems-we-solve)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Keyboard Shortcuts](#keyboard-shortcuts)
- [Writing Self-Documenting Makefiles](#writing-self-documenting-makefiles)
- [Recent History & Smart Navigation](#recent-history--smart-navigation)
- [Variable Inspector: Understanding Your Build Configuration](#variable-inspector-understanding-your-build-configuration)
- [Workspace Management: Working with Multiple Projects](#workspace-management-working-with-multiple-projects)
- [Understanding Dependency Graphs](#understanding-dependency-graphs)
- [Safety Features: Preventing Accidental Disasters](#safety-features-preventing-accidental-disasters)
  - [Visual Indicators](#visual-indicators)
  - [Two-Column Layout](#two-column-layout)
  - [Confirmation Dialog](#confirmation-dialog)
  - [Built-in Dangerous Patterns](#built-in-dangerous-patterns)
  - [Context-Aware Detection](#context-aware-detection)
  - [Configuration](#configuration)
  - [Disabling Safety Checks](#disabling-safety-checks)
- [Export & Shell Integration](#export--shell-integration)
  - [Export Execution Results](#export-execution-results)
  - [Shell Integration](#shell-integration)

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

- **Search & filtering with recent history**: Enhanced productivity for repetitive workflows
  - Real-time fuzzy search: Type `/` to filter targets by name or description
  - Recent targets tracking: Last 5 executed targets appear at the top with ⏱ indicator
  - Per-Makefile history: Each project maintains its own execution history
  - Automatic cleanup: Stale targets (removed from Makefile) are filtered automatically
  - Persistent across sessions: History stored in `~/.cache/lazymake/history.json`
  - Navigation: Arrow keys skip over section headers and separators

- **Performance profiling**: Track execution time and identify performance regressions
  - Real-time execution timer: Updates every 100ms with progress indicators (🟢 on track, 🔵 finishing up, 🔴 slower than usual)
  - Automatic regression detection: Alerts when targets are >25% slower than average
  - Context-aware display: Performance info shown only when relevant (regressed or recent targets)
  - Visual indicators: 📈 for regressed targets, duration badges color-coded by performance
  - Performance history: Stores last 10 executions per target with stats (avg, min, max)
  - Post-execution alerts: Warnings appear after slow runs with actionable insights
  - Persistent tracking: Performance data survives across sessions

- **Dangerous command detection**: Protect against accidental destructive operations
  - Visual indicators: 🚨 for critical, ⚠️ for warning-level dangerous commands
  - Two-column layout: Recipe preview shows exactly what will execute
  - Confirmation dialogs: Critical commands require explicit confirmation
  - Context-aware detection: Smart severity adjustment based on target name and context
  - Built-in rules: 11 dangerous patterns (rm -rf, database drops, terraform destroy, etc.)
  - Configurable: Customize rules, exclude targets, disable globally
  - Safe defaults: Enabled by default to protect newcomers to projects

- **Variable inspector**: Understand and track Makefile variables
  - Full-screen view: Press `v` to browse all variables with navigation
  - Context panel: Shows variables used by the selected target in recipe preview
  - Hybrid parsing: Extracts definitions from Makefile + expands values using make
  - Variable types: Supports all assignment operators (=, :=, +=, ?=, !=)
  - Usage tracking: Shows which targets use each variable
  - Export detection: Identifies variables exported to environment
  - Dual display: Shows both raw values and fully expanded values
  - Smart navigation: Scrollable list with up/down or j/k keys

- **Workspace/Project Management**: Quick switching between different projects
  - Press `w` to open workspace picker with recent and discovered Makefiles
  - Automatic discovery: Scans project tree (3 levels deep) to find all Makefiles
  - Recent workspaces: Last 10 accessed Makefiles with "time ago" display and access count
  - Discovered workspaces: Shows newly found Makefiles in project that haven't been used yet
  - Favorite workspaces: Star frequently used projects (press `f` to toggle favorites)
  - Status bar integration: Shows current Makefile path relative to working directory
  - Per-project history: Each Makefile maintains its own execution history
  - Smart exclusions: Skips `.git`, `node_modules`, `vendor`, build directories automatically
  - Access tracking: Records access count and last accessed time per workspace
  - Persistent storage: Workspace data saved in `~/.cache/lazymake/workspaces.json`
  - Automatic cleanup: Removes workspace entries for deleted Makefiles

### Planned

#### High Priority
- **Better error handling**: Parse and highlight common Makefile errors with helpful suggestions

### Nice to Have
- Multi-language recipe support with syntax highlighting
- Variable runtime overrides through TUI
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
- `v` - View variable inspector
- `w` - Open workspace picker to switch between Makefiles
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

#### Variable Inspector
- `v` or `esc` - Return to list view
- `↑/↓` or `j/k` - Navigate variables
- `q` or `ctrl+c` - Quit

#### Output View
- `↑/↓` or `j/k` - Scroll through output
- `esc` - Return to list view
- `q` or `ctrl+c` - Quit

#### Workspace Picker
- `↑/↓` or `j/k` - Navigate workspace list
- `Enter` - Switch to selected workspace
- `f` - Toggle favorite for selected workspace
- `esc` or `w` - Return to list view
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

## Recent History & Smart Navigation

lazymake tracks your most frequently used targets to speed up repetitive workflows. When you run the same 2-3 targets regularly (like `test`, `build`, `deploy`), they appear at the top of the list for instant access.

### How It Works

- **Automatic tracking**: Every time you execute a target, it's recorded in history
- **Top 5 recent**: The last 5 executed targets appear in the "RECENT" section
- **Visual indicator**: Recent targets are marked with a ⏱ clock emoji
- **Per-project**: Each Makefile maintains separate history
- **Smart cleanup**: Targets removed from the Makefile are automatically filtered out
- **Persistent**: History survives across sessions

### Example Display

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

### Benefits

- **Faster workflows**: No need to scroll or type to find common targets
- **Context awareness**: Different projects show different recent targets
- **Zero configuration**: Works automatically from the first execution

## Variable Inspector: Understanding Your Build Configuration

lazymake helps you understand and track Makefile variables, making it easy to see what values are used and where they come from.

### Two Ways to View Variables

#### 1. Full-Screen Inspector (Press `v`)

Browse all variables in your Makefile with detailed information:

```
┌─ Variable Inspector ──────────────────────────────────────┐
│                                                            │
│ 6 variables • 4 used • 2 unused                           │
│                                                            │
│ BINARY_NAME          [:=]  Simply Expanded                │
│   Raw:      lazymake                                       │
│   Expanded: lazymake                                       │
│   Used by:  build, install (2 targets)                    │
│                                                            │
│ VERSION              [=]   Recursive                       │
│   Raw:      1.0.0                                          │
│   Expanded: 1.0.0                                          │
│   Used by:  build (1 target)                              │
│                                                            │
│ GOFLAGS              [=]   Recursive                       │
│   Raw:      -v -race                                       │
│   Expanded: -v -race                                       │
│   Used by:  build, test (2 targets)                       │
│                                                            │
│ LDFLAGS              [=]   Recursive                       │
│   Raw:      -ldflags "-X main.version=$(VERSION)"         │
│   Expanded: -ldflags "-X main.version=1.0.0"              │
│   Used by:  build (1 target)                              │
│                                                            │
│ BUILD_DIR            [?=]  Conditional                     │
│   Raw:      ./bin                                          │
│   Expanded: ./bin                                          │
│   Used by:  build, clean (2 targets)                      │
│                                                            │
│ PATH                                                       │
│   Exported to environment                                  │
│   Not used by any target                                   │
│                                                            │
└────────────────────────────────────────────────────────────┘
  v/esc: return • ↑↓/j/k: navigate • q: quit
```

#### 2. Context Panel (Automatic)

When you select a target, the recipe preview shows variables it uses:

```
┌─────────────────────┬────────────────────────────────────────┐
│ ALL TARGETS         │ build:                                 │
│ > build             │                                        │
│   test              │   go build $(GOFLAGS) $(LDFLAGS) \    │
│   clean             │     -o $(BUILD_DIR)/$(BINARY_NAME)    │
│   install           │                                        │
│                     │   💡 Press 'g' to view full graph      │
│                     │                                        │
│                     │   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━   │
│                     │                                        │
│                     │   📦 Variables Used                    │
│                     │                                        │
│                     │     GOFLAGS = -v -race                 │
│                     │     LDFLAGS = -ldflags "-X main..."    │
│                     │     BUILD_DIR = ./bin                  │
│                     │     BINARY_NAME = lazymake             │
│                     │                                        │
│                     │     💡 Press 'v' to view all variables │
└─────────────────────┴────────────────────────────────────────┘
```

### Variable Types Explained

lazymake recognizes all Makefile variable assignment operators:

- **`=` Recursive**: Expanded when used (can reference later variables)
- **`:=` Simply Expanded**: Expanded when defined (like shell variables)
- **`+=` Append**: Adds to existing value
- **`?=` Conditional**: Sets only if not already defined
- **`!=` Shell**: Executes shell command and captures output

### How It Works

1. **Parse Definitions**: Extracts variable assignments from Makefile text
2. **Expand Values**: Runs `make --print-data-base` to get fully expanded values
3. **Track Usage**: Scans all target recipes to find variable references
4. **Display Context**: Shows raw vs expanded values and which targets use them

### Example Makefile

```makefile
# Variable definitions
BINARY_NAME := lazymake
VERSION = 1.0.0
GOFLAGS = -v -race
LDFLAGS = -ldflags "-X main.version=$(VERSION)"
BUILD_DIR ?= ./bin

export PATH

build: ## Build the application
	go build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)

test: ## Run tests
	go test $(GOFLAGS) ./...

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
```

### Benefits

- **Understand configuration**: See what values are actually used
- **Debug issues**: Compare raw vs expanded values to spot errors
- **Track dependencies**: Know which targets are affected by variable changes
- **Onboarding**: New developers can understand build configuration instantly
- **Environment awareness**: Identify which variables are exported

### Navigation

- **`v`**: Open variable inspector from list view
- **`↑/↓` or `j/k`**: Navigate between variables
- **`v` or `esc`**: Return to list view
- **Auto-scroll**: Inspector automatically scrolls to keep selected variable visible

## Workspace Management: Working with Multiple Projects

lazymake makes it easy to work with multiple projects and Makefiles. Press `w` to see recent workspaces and automatically discovered Makefiles in your project.

### Workspace Picker (Press `w`)

Access recent and discovered Makefiles with a single keypress:

```
┌─ Switch Workspace ────────────────────────────────────────┐
│                                                            │
│ ⭐ ./Makefile                                              │
│    Last used: 2 minutes ago • 15 times                    │
│                                                            │
│    ../other-project/Makefile                              │
│    Last used: 1 hour ago • 8 times                        │
│                                                            │
│    examples/dangerous.mk                                  │
│    Discovered in project                                  │
│                                                            │
│    tools/Makefile                                         │
│    Discovered in project                                  │
│                                                            │
└────────────────────────────────────────────────────────────┘
  2 recent • 2 discovered
  enter: switch • f: favorite • esc/w: cancel
```

**Features:**
- **Automatic discovery**: Scans your project tree (up to 3 levels deep) to find all Makefiles
- **Recent workspaces**: Shows last 10 accessed Makefiles with access tracking
- **Discovered workspaces**: Displays found Makefiles you haven't used yet
- **Favorites first**: Star frequently used projects with `f` - they appear at the top
- **Access tracking**: Displays "time ago" (e.g., "2 hours ago") and access count for recent workspaces
- **Smart exclusions**: Skips `.git`, `node_modules`, `vendor`, build directories, and other common non-code paths
- **Fast scanning**: 5-second timeout ensures responsiveness even in large projects

### Status Bar Integration

The current workspace is always visible in the status bar:

```
└──────────────────────────────────────────────────────────┘
│ ./Makefile • 12 targets • 2 dangerous    enter: run • q │
└──────────────────────────────────────────────────────────┘
```

The path is displayed relative to your current working directory:
- `./Makefile` - in current directory
- `../other/Makefile` - in sibling directory
- `~/projects/foo/Makefile` - absolute path with `~` expansion

### Per-Project History

Each workspace automatically maintains its own execution history. When you switch between projects, you'll see the recent targets for that specific Makefile:

```
# Working in project A
RECENT
⏱  build-api    Build the API server       3.2s
⏱  test-api     Run API tests              1.5s

# Switch to project B (press 'w')
RECENT
⏱  deploy-prod  Deploy to production       45.1s
⏱  build-web    Build web frontend         8.3s
```

This means:
- Each Makefile remembers its own frequently used targets
- No need to scroll through unrelated targets
- Faster context switching between projects

### Automatic Tracking

lazymake automatically tracks workspace usage:
- **On first use**: Creates workspace entry when you run a target
- **On subsequent uses**: Updates access count and last accessed time
- **On cleanup**: Removes entries for deleted Makefiles automatically
- **Persistent**: Data survives across sessions in `~/.cache/lazymake/workspaces.json`

### How Discovery Works

When you press `w`, lazymake:
1. **Records current Makefile** - Ensures your current file appears in the list
2. **Scans project tree** - Searches up to 3 levels deep from current directory
3. **Finds all Makefiles** - Detects `Makefile`, `makefile`, `GNUmakefile`, `*.mk`, `*.mak`
4. **Applies exclusions** - Skips `.git`, `node_modules`, `vendor`, `build`, `dist`, `.cache`, etc.
5. **Combines results** - Shows recent workspaces first, then newly discovered ones
6. **Fast operation** - 5-second timeout prevents hanging on large projects

### Use Cases

#### 1. Monorepo Development

Working with multiple Makefiles in a large repository:

```
my-monorepo/
├── Makefile              # Root Makefile
├── services/
│   ├── api/Makefile      # API service
│   ├── auth/Makefile     # Auth service
│   └── worker/Makefile   # Background worker
└── frontend/Makefile     # Frontend app
```

Press `w` to see all Makefiles automatically - no manual browsing needed!

#### 2. Multi-Project Development

Switching between different projects:
- Press `w` to see recent projects and discovered Makefiles
- Star your most frequently used projects with `f`
- Favorites always appear at the top of the list

#### 3. Onboarding

New to a project? Press `w`:
- Instantly see all available Makefiles
- Discovered Makefiles show "Discovered in project"
- Select one to start working - it gets added to your recent list

### For Makefiles Outside Discovery Range

If you need a Makefile that's:
- More than 3 levels deep
- In an excluded directory
- Outside your current project

Use the CLI flag:
```bash
lazymake -f path/to/Makefile
```

Once accessed, it appears in your recent workspaces list.

### Navigation

- **`w`**: Open workspace picker from list view
- **`↑/↓` or `j/k`**: Navigate workspaces
- **`f`**: Toggle favorite (star/unstar workspace)
- **`enter`**: Switch to selected workspace
- **`esc` or `w`**: Return to main list view

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

## Safety Features: Preventing Accidental Disasters

lazymake protects you from accidentally running destructive commands by detecting dangerous patterns and requiring confirmation before execution.

### Visual Indicators

Targets are marked with emoji indicators based on danger level:

```
RECENT
🚨  deploy-prod      Deploy to production
⚠️  clean            Clean build artifacts
──────────────────────────────────────────────
ALL TARGETS
    build            Build the application
🚨  nuke-db          Drop production database
⚠️  docker-clean     Clean Docker resources
```

- **🚨 Critical**: Commands that can cause irreversible damage (requires confirmation)
- **⚠️ Warning**: Commands with potential side effects (executes immediately)
- **No indicator**: Safe commands

### Two-Column Layout

The interface shows exactly what will execute:

```
┌─────────────────────┬────────────────────────────────────────┐
│ RECENT              │ Recipe Preview                         │
│ > 🚨 deploy-prod    │                                        │
│   ⚠️  clean         │ deploy-prod:                           │
│                     │   kubectl apply -f k8s/prod/           │
│ ───────────────     │   terraform apply -var-file=prod.tfvars│
│ ALL TARGETS         │                                        │
│   build             │ Danger: terraform-destroy (CRITICAL)   │
│   test              │ This destroys Terraform infrastructure │
│                     │ 💡 Verify workspace first              │
└─────────────────────┴────────────────────────────────────────┘
│ 12 targets • 2 dangerous      enter: run • ?: help • q: quit │
└──────────────────────────────────────────────────────────────┘
```

**Left column**: Target list with indicators
**Right column**: Recipe commands + safety warnings
**Bottom**: Stats and contextual shortcuts

### Confirmation Dialog

Critical commands show a prominent warning dialog:

```
╔════════════════════════════════════════════════════════╗
║                                                        ║
║     🚨 DANGEROUS COMMAND WARNING                       ║
║                                                        ║
║     Target: nuke-db                                    ║
║                                                        ║
║     CRITICAL: database-drop                            ║
║     Command: psql -c 'DROP DATABASE production;'       ║
║                                                        ║
║     Drops databases or truncates tables. This causes   ║
║     permanent data loss.                               ║
║                                                        ║
║     💡 Always backup before destructive database       ║
║        operations. Verify database name.               ║
║                                                        ║
║     [Enter] Continue Anyway     [Esc] Cancel           ║
║                                                        ║
╚════════════════════════════════════════════════════════╝
```

- Press `Esc` to cancel safely
- Press `Enter` to proceed with execution

### Built-in Dangerous Patterns

#### Critical (🚨 + Confirmation Required)

- **rm-rf-root**: Recursive deletion of system paths (`rm -rf /`, `sudo rm -rf`)
- **disk-wipe**: Disk formatting or block device writes (`dd`, `mkfs`)
- **database-drop**: Database/table deletion (`DROP DATABASE`, `TRUNCATE TABLE`)
- **git-force-push**: Force pushing to repositories (`git push -f`)
- **terraform-destroy**: Infrastructure destruction (`terraform destroy`)
- **kubectl-delete**: Kubernetes resource deletion (`kubectl delete namespace`)

#### Warning (⚠️ Only)

- **docker-system-prune**: Docker cleanup operations
- **git-reset-hard**: Discarding uncommitted changes
- **npm-uninstall-all**: Removing all dependencies
- **package-remove**: System package removal
- **chmod-777**: Overly permissive file permissions

### Context-Aware Detection

lazymake intelligently adjusts severity based on context:

**Clean targets** (downgraded severity):
```makefile
clean:  ## Clean build artifacts
	rm -rf build/  # WARNING instead of CRITICAL
```

**Production keywords** (elevated severity):
```makefile
deploy-prod:  ## Deploy to production
	docker system prune  # CRITICAL instead of WARNING
```

**Interactive flags** (downgraded severity):
```makefile
dangerous-op:
	rm -rfi build/  # WARNING instead of CRITICAL (has -i flag)
```

**Development targets** (downgraded for non-prod):
```makefile
test-cleanup:
	terraform destroy  # WARNING instead of CRITICAL (test target)
```

### Configuration

Create `.lazymake.yaml` in your project or home directory:

```yaml
safety:
  # Master switch (default: true)
  enabled: true

  # Exclude specific targets from ALL checks
  exclude_targets:
    - clean
    - distclean
    - reset-dev-db

  # Enable only specific built-in rules (omit to enable all)
  enabled_rules:
    - rm-rf-root
    - database-drop
    - git-force-push
    - terraform-destroy
    - kubectl-delete

  # Add custom rules
  custom_rules:
    - id: "prod-deploy"
      severity: critical  # critical, warning, or info
      patterns:
        - "kubectl apply.*production"
        - "terraform apply.*prod"
      description: "Deploying to production environment"
      suggestion: "Verify environment and get team approval"
```

**Configuration merging**:
- Settings from `~/.lazymake.yaml` (global) and `./.lazymake.yaml` (project) are merged
- `enabled`: project overrides global
- `enabled_rules`, `exclude_targets`, `custom_rules`: union of both

### Disabling Safety Checks

**Globally** (in `~/.lazymake.yaml`):
```yaml
safety:
  enabled: false
```

**Per-project** (in `./.lazymake.yaml`):
```yaml
safety:
  enabled: false
```

**Per-target** (exclude specific targets):
```yaml
safety:
  exclude_targets:
    - my-dangerous-but-safe-target
    - another-excluded-target
```

### Why Default-Enabled?

Safety checks are enabled by default because:
1. **Protect newcomers**: Developers new to a project are most at risk
2. **Non-intrusive**: Visual indicators don't block workflow
3. **Easy to disable**: One config line turns it off if unwanted
4. **Better safe than sorry**: Confirmation dialogs take 1 second; data recovery takes hours

### Real-World Scenarios

**Scenario 1: New developer runs `make clean`**
```
⚠️  clean  # Warning indicator shown
# Recipe preview shows: rm -rf build/
# Executes immediately (warning-level, clean target)
```

**Scenario 2: Accidentally select `make nuke-prod-db`**
```
🚨 nuke-prod-db  # Critical indicator shown
# Presses Enter
# Confirmation dialog appears with full warning
# Presses Esc to cancel safely
```

**Scenario 3: Experienced dev working on deployment**
```yaml
# .lazymake.yaml
safety:
  exclude_targets:
    - deploy-prod
    - deploy-staging
```

## Export & Shell Integration

lazymake can export execution results to JSON/log files and integrate with your shell history, making it easier to track builds, debug issues, and repeat commands.

### Export Execution Results

Export execution results to structured JSON files or human-readable logs for analysis, debugging, and CI/CD integration.

#### Features

- **Multiple formats**: Export to JSON, plain text logs, or both
- **Flexible naming**: Choose between timestamp-based, target-based, or sequential naming
- **Automatic rotation**: Keep storage under control with configurable file limits and age-based cleanup
- **Rich metadata**: Captures exit codes, execution time, output, environment context, and more
- **Filtering**: Export only successful executions or exclude specific targets
- **Async operation**: Non-blocking exports don't slow down your workflow

#### Exported Data

Each execution record includes:
- Target name and Makefile path
- Start/end timestamps and duration
- Exit code and success status
- Complete stdout/stderr output
- Working directory, user, and hostname
- lazymake version

#### Example JSON Export

```json
{
  "timestamp": "2025-12-12T14:30:22.123Z",
  "target_name": "build",
  "makefile_path": "/path/to/Makefile",
  "duration_ms": 2023,
  "success": true,
  "exit_code": 0,
  "output": "go build -o bin/lazymake...",
  "working_dir": "/path/to/project",
  "user": "developer",
  "hostname": "laptop.local"
}
```

#### Example Log Export

```
================================================================================
Lazymake Execution Log
================================================================================
Target:        build
Makefile:      /path/to/Makefile
Timestamp:     2025-12-12 14:30:22
Duration:      2.023s
Exit Code:     0
Status:        SUCCESS
Working Dir:   /path/to/project
User:          developer
Host:          laptop.local
================================================================================

OUTPUT:
go build -o bin/lazymake cmd/lazymake/main.go
Build complete: bin/lazymake

================================================================================
Execution completed successfully in 2.023s
================================================================================
```

#### Configuration

Add to `.lazymake.yaml`:

```yaml
export:
  # Enable export (default: false)
  enabled: true

  # Output directory (default: ~/.cache/lazymake/exports)
  output_dir: ~/.cache/lazymake/exports

  # Format: "json", "log", or "both" (default: "json")
  format: both

  # Naming strategy (default: "timestamp")
  # - timestamp: build_20251212_143022.json
  # - target: build_latest.json (overwrites)
  # - sequential: build_1.json, build_2.json, ...
  naming_strategy: timestamp

  # Rotation settings
  max_files: 50        # Keep max 50 files per target
  keep_days: 30        # Delete files older than 30 days
  max_file_size_mb: 10 # Skip exports larger than 10MB

  # Filtering
  success_only: false  # Export all executions (default)
  exclude_targets:     # Don't export these targets
    - watch
    - dev
```

#### Use Cases

1. **CI/CD Integration**: Export JSON for automated analysis and metrics
2. **Debugging**: Review detailed logs of failed builds
3. **Performance Tracking**: Analyze execution times over time
4. **Audit Trail**: Maintain records of production deployments

### Shell Integration

Add executed make commands to your shell history, making it easy to re-run commands outside of lazymake.

#### Features

- **Automatic detection**: Detects your shell (bash, zsh) from `$SHELL` environment variable
- **Format support**: Handles both standard and extended zsh history formats
- **File locking**: Safe concurrent writes prevent history corruption
- **Custom templates**: Customize the command format added to history
- **Filtering**: Exclude specific targets from history (e.g., `help`, `list`)
- **Async operation**: Non-blocking writes don't slow down execution

#### Supported Shells

- **bash**: Appends to `~/.bash_history`
- **zsh**: Appends to `~/.zsh_history` or `$HISTFILE`
  - Automatically detects extended history format
  - Includes timestamps when `setopt EXTENDED_HISTORY` is enabled
- **fish**: Coming soon

#### How It Works

When you execute a target in lazymake, the command `make <target>` is added to your shell history. Later, you can:
- Use `history` to see all your make commands
- Press `↑` to cycle through recent make commands
- Use `Ctrl+R` to search your make command history

#### Example

```bash
# After running 'build' and 'test' targets in lazymake
$ history | tail -5
  498  make build
  499  make test
  500  make build
  501  make lint
  502  history | tail -5

# Re-run using shell history
$ make build  # Just press ↑ and Enter
```

#### Configuration

Add to `.lazymake.yaml`:

```yaml
shell_integration:
  # Enable shell integration (default: false)
  enabled: true

  # Shell type: "auto", "bash", "zsh", or "none" (default: "auto")
  shell: auto

  # Override history file path (default: use shell default)
  history_file: ""

  # Include timestamp for zsh extended history (default: true)
  include_timestamp: true

  # Custom format template (default: "make {target}")
  # Available variables: {target}, {makefile}, {dir}
  format_template: "make {target}"

  # Exclude targets from history
  exclude_targets:
    - help
    - list
```

#### Custom Format Templates

Customize what gets added to your shell history:

```yaml
# Simple (default)
format_template: "make {target}"
# Result: make build

# Include Makefile path
format_template: "make -f {makefile} {target}"
# Result: make -f /path/to/Makefile build

# Include working directory
format_template: "cd {dir} && make {target}"
# Result: cd /path/to/project && make build
```

#### Benefits

- **Seamless workflow**: Switch between TUI and command line easily
- **Command history**: All make commands available via `history`
- **Shell completion**: Works with your existing shell completion setup
- **No lock-in**: Commands are standard make invocations

#### Configuration Scenarios

**Scenario 1: Enable for CI/CD analysis**
```yaml
export:
  enabled: true
  format: json
  naming_strategy: target  # Keep only latest
  success_only: true
```

**Scenario 2: Debugging with detailed logs**
```yaml
export:
  enabled: true
  format: both  # JSON + logs
  max_files: 20
  keep_days: 7
```

**Scenario 3: Shell integration for bash**
```yaml
shell_integration:
  enabled: true
  shell: bash
```

**Scenario 4: zsh with custom format**
```yaml
shell_integration:
  enabled: true
  shell: zsh
  format_template: "make -f {makefile} {target}"
  exclude_targets:
    - help
    - list
```

