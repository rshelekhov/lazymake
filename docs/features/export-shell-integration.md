# Export & Shell Integration

lazymake can export execution results to JSON/log files and integrate with your shell history, making it easier to track builds, debug issues, and repeat commands.

## Export Execution Results

Export execution results to structured JSON files or human-readable logs for analysis, debugging, and CI/CD integration.

### Features

- **Multiple formats**: Export to JSON, plain text logs, or both
- **Flexible naming**: Choose between timestamp-based, target-based, or sequential naming
- **Automatic rotation**: Keep storage under control with configurable file limits and age-based cleanup
- **Rich metadata**: Captures exit codes, execution time, output, environment context, and more
- **Filtering**: Export only successful executions or exclude specific targets
- **Async operation**: Non-blocking exports don't slow down your workflow

### Exported Data

Each execution record includes:
- Target name and Makefile path
- Start/end timestamps and duration
- Exit code and success status
- Complete stdout/stderr output
- Working directory, user, and hostname
- lazymake version

### Example JSON Export

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

### Example Log Export

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

### Configuration

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

### Use Cases

1. **CI/CD Integration**: Export JSON for automated analysis and metrics
2. **Debugging**: Review detailed logs of failed builds
3. **Performance Tracking**: Analyze execution times over time
4. **Audit Trail**: Maintain records of production deployments

## Shell Integration

Add executed make commands to your shell history, making it easy to re-run commands outside of lazymake.

### Features

- **Automatic detection**: Detects your shell (bash, zsh) from `$SHELL` environment variable
- **Format support**: Handles both standard and extended zsh history formats
- **File locking**: Safe concurrent writes prevent history corruption
- **Custom templates**: Customize the command format added to history
- **Filtering**: Exclude specific targets from history (e.g., `help`, `list`)
- **Async operation**: Non-blocking writes don't slow down execution

### Supported Shells

- **bash**: Appends to `~/.bash_history`
- **zsh**: Appends to `~/.zsh_history` or `$HISTFILE`
  - Automatically detects extended history format
  - Includes timestamps when `setopt EXTENDED_HISTORY` is enabled
- **fish**: Coming soon

### How It Works

When you execute a target in lazymake, the command `make <target>` is added to your shell history. Later, you can:
- Use `history` to see all your make commands
- Press `↑` to cycle through recent make commands
- Use `Ctrl+R` to search your make command history

### Example

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

### Configuration

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

### Custom Format Templates

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

### Benefits

- **Seamless workflow**: Switch between TUI and command line easily
- **Command history**: All make commands available via `history`
- **Shell completion**: Works with your existing shell completion setup
- **No lock-in**: Commands are standard make invocations

## Configuration Scenarios

### Scenario 1: Enable for CI/CD analysis
```yaml
export:
  enabled: true
  format: json
  naming_strategy: target  # Keep only latest
  success_only: true
```

### Scenario 2: Debugging with detailed logs
```yaml
export:
  enabled: true
  format: both  # JSON + logs
  max_files: 20
  keep_days: 7
```

### Scenario 3: Shell integration for bash
```yaml
shell_integration:
  enabled: true
  shell: bash
```

### Scenario 4: zsh with custom format
```yaml
shell_integration:
  enabled: true
  shell: zsh
  format_template: "make -f {makefile} {target}"
  exclude_targets:
    - help
    - list
```

---

[← Back to Documentation](../README.md) | [← Back to Main README](../../README.md)
