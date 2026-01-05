<img src="./assets/lazymake-logo.svg" alt="LAZYMAKE" width="600">

</br>

[![CI](https://img.shields.io/github/actions/workflow/status/rshelekhov/lazymake/ci.yml?branch=main&label=CI)](https://github.com/rshelekhov/lazymake/actions/workflows/ci.yml)
[![GitHub release](https://img.shields.io/github/release/rshelekhov/lazymake.svg)](https://github.com/rshelekhov/lazymake/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/rshelekhov/lazymake)](https://goreportcard.com/report/github.com/rshelekhov/lazymake)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub stars](https://img.shields.io/github/stars/rshelekhov/lazymake?style=social)](https://github.com/rshelekhov/lazymake/stargazers)

A terminal UI for browsing and running Makefile targets.

<img src="docs/assets/demo.gif" alt="Demo" width="100%" style="max-width: 909px;">

## Quick Start

```bash
# Install
brew install rshelekhov/tap/lazymake

# Run in any directory with a Makefile
lazymake
```

## What It Does

lazymake provides an interactive interface for Makefiles with:

- **Target browser** with fuzzy search and execution history
- **Dependency graph visualization** showing what runs when you execute a target
- **Variable inspector** for debugging complex variable expansions
- **Syntax highlighting** for recipes (detects Python, Go, shell scripts, etc.)
- **Safety warnings** for destructive commands (configurable)
- **Performance tracking** to identify slow targets


## Why?

In large projects, Makefiles can have dozens or hundreds of targets. Finding the right one means either memorizing names or grepping through the file. This tool gives you a searchable list with descriptions and shows you exactly what will run before you execute anything.

The dependency graph is particularly useful for understanding complex build systems—you can see the full chain of prerequisites, identify bottlenecks, and spot circular dependencies.

## Installation

### Homebrew (macOS/Linux)
```bash
brew install rshelekhov/tap/lazymake
```

### Go
```bash
go install github.com/rshelekhov/lazymake/cmd/lazymake@latest
```

## Usage

```bash
# Use Makefile in current directory
lazymake

# Specify path
lazymake -f path/to/Makefile
```

### Keyboard shortcuts

- `↑/↓` or `j/k` - Navigate
- `Enter` - Execute selected target
- `g` - Show dependency graph
- `v` - Open variable inspector
- `w` - Switch between Makefiles (workspace picker)
- `/` - Search/filter
- `?` - Help
- `q` - Quit

## Configuration

Optional `.lazymake.yaml` for customization:

```yaml
safety:
  enabled: true
  exclude_targets:
    - clean  # Don't warn about these targets

export:
  enabled: true
  format: json

shell_integration:
  enabled: true
```

Place in `~/.lazymake.yaml` (global) or `./.lazymake.yaml` (project-specific).

See [configuration guide](docs/guides/configuration.md) for all options.

## Makefile Documentation

If your targets have comments starting with `##`, they'll appear as descriptions in the UI:

```makefile
build: ## Build the application
	go build -o bin/app

test: ## Run test suite
	go test ./...
```

This is optional—lazymake works fine without comments.

## Features

### Dependency Graph Visualization

![Dependency Graph](docs/assets/dependency-graph.png)

Press `g` on any target to see its dependency tree with execution order and parallel opportunities. Useful for understanding what `make deploy` actually does.

[Full documentation](docs/features/dependency-graphs.md)

### Variable Inspector

![Variable Inspector](docs/assets/variable-inspector.png)

Press `v` to browse all variables, see their expanded values, and find out which targets use them. Helpful when debugging complex variable substitutions or figuring out where `LDFLAGS` is defined.

[Full documentation](docs/features/variable-inspector.md)

### Dangerous Command Detection

![Safety Features](docs/assets/safety-features.png)

Helps prevent accidental execution of destructive commands. lazymake scans targets for potentially dangerous operations (like `rm -rf`, `DROP DATABASE`, etc.) and prompts for confirmation.

[Full documentation](docs/features/safety-features.md)

### Workspace Management

![Workspace Management](docs/assets/workspace-management.png)

Press `w` to switch between Makefiles in your project. Automatically discovers Makefiles in subdirectories and remembers recent files.

[Full documentation](docs/features/workspace-management.md)

## FAQ

**Does it work with my existing Makefile?**  
Yes. No changes required.

**Which Make implementations are supported?**  
All of them. lazymake parses the file and uses your system's `make` for execution.

**Does it modify my Makefile?**  
No. It's read-only.

**Can I disable safety warnings?**  
Yes. Add target names to `exclude_targets` in `.lazymake.yaml`.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT - see [LICENSE](LICENSE).

## Acknowledgments

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lipgloss](https://github.com/charmbracelet/lipgloss), [Chroma](https://github.com/alecthomas/chroma), and [Cobra](https://github.com/spf13/cobra).

Inspired by [lazygit](https://github.com/jesseduffield/lazygit) and [lazydocker](https://github.com/jesseduffield/lazydocker).

---

**Made with ❤️ for developers.**
