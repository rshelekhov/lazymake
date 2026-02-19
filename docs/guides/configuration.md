# Configuration Guide

lazymake can be configured using YAML configuration files. This guide covers all available configuration options.

## Configuration File Locations

Place configuration files at:

- **`~/.lazymake.yaml`**: Global configuration for all projects
- **`./.lazymake.yaml`**: Project-specific configuration

### Configuration Merging

When both files exist, they are merged with consistent rules across all sections (`safety`, `export`, `shell_integration`):

- **Scalars** (`enabled`, `format`, `shell`, `max_files`, etc.): Project config overrides global
- **String lists** (`enabled_rules`, `exclude_targets`): Union of both, deduplicated
- **Struct lists** (`custom_rules`): Appended (global rules first, then project rules)

## Basic Settings

```yaml
# Makefile path (default: auto-detect GNUmakefile, makefile, Makefile)
# When empty, lazymake searches in GNU make order: GNUmakefile → makefile → Makefile
makefile: ""
```

## Safety Features

Configure dangerous command detection and confirmation dialogs.

### Master Switch

```yaml
safety:
  # Enable/disable all safety checks (default: true)
  enabled: true
```

### Exclude Targets

Exclude specific targets from ALL safety checks:

```yaml
safety:
  exclude_targets:
    - clean          # Standard cleanup target
    - distclean      # Deep cleanup
    - reset-dev-db   # Known safe development operation
```

### Enable Specific Rules

Enable only certain built-in rules (omit to enable all 36 rules):

```yaml
safety:
  enabled_rules:
    - rm-rf-root
    - database-drop
    - git-force-push
    - terraform-destroy
    - kubectl-delete
```

**Available built-in rules (36 total):**

*Critical — system-wide destructive operations:*
- `rm-rf-root` - Recursive deletion of system paths
- `disk-wipe` - Disk formatting or block device writes
- `database-drop` - Database/table deletion
- `git-force-push` - Force pushing to repositories
- `terraform-destroy` - Infrastructure destruction
- `kubectl-delete` - Kubernetes resource deletion

*Critical — cloud infrastructure destruction:*
- `aws-s3-delete` - AWS S3 bucket/object deletion
- `cloud-instance-terminate` - Cloud compute instance termination
- `curl-pipe-shell` - Piping remote content to shell
- `aws-destructive` - AWS resource destruction (CloudFormation, EC2, RDS)
- `gcp-destructive` - GCP resource destruction (projects, instances, Cloud SQL)
- `azure-destructive` - Azure resource destruction (resource groups, VMs, SQL)
- `heroku-destructive` - Heroku app/addon destruction

*Critical — database operations:*
- `redis-flush` - Redis FLUSHALL/FLUSHDB
- `cassandra-drop` - Cassandra keyspace/table deletion

*Critical — system operations:*
- `crontab-remove` - Removing all cron jobs
- `iptables-flush` - Flushing all firewall rules

*Warning — project-level destructive operations:*
- `docker-system-prune` - Docker cleanup operations
- `git-reset-hard` - Discarding uncommitted changes
- `npm-uninstall-all` - Removing all Node.js dependencies
- `package-remove` - System package removal
- `chmod-777` - Overly permissive file permissions
- `firewall-flush` - Firewall rule removal
- `process-kill-force` - Force killing processes (SIGKILL)
- `helm-delete` - Helm release deletion
- `ssh-key-delete` - SSH key/config deletion
- `env-file-overwrite` - Environment file overwriting

*Warning — version control:*
- `git-branch-delete-force` - Force deleting git branches
- `git-reflog-expire` - Expiring reflog/pruning objects

*Warning — container orchestration:*
- `docker-swarm-destructive` - Docker swarm stack/service removal
- `podman-system-reset` - Podman system reset

*Warning — package managers:*
- `pip-uninstall-all` - Python package uninstall
- `go-clean-modcache` - Go module cache cleanup

*Warning — critical services:*
- `systemctl-critical-services` - Stopping critical system services
- `killall-force` - Force killing all processes by name
- `deployment-commands` - Deployment operations (kubectl apply, terraform apply)

### Custom Rules

Add project-specific dangerous patterns:

```yaml
safety:
  custom_rules:
    # Example: Production deployment detection
    - id: "prod-deploy"
      severity: critical  # critical, warning, or info
      patterns:
        - "kubectl apply.*production"
        - "kubectl apply.*prod"
        - "terraform apply.*prod"
        - "helm.*--namespace=production"
      description: "Deploying to production environment without review"
      suggestion: "Get team approval before production deployments. Use staging first."

    # Example: Database migrations in production
    - id: "prod-migration"
      severity: critical
      patterns:
        - "rails db:migrate.*RAILS_ENV=production"
        - "alembic upgrade head.*production"
        - "migrate.*--env=production"
      description: "Running database migrations in production"
      suggestion: "Backup database first. Test migration in staging. Have rollback plan ready."
```

### Safety Configuration Scenarios

**Scenario 1: Disable safety for experienced team**
```yaml
# Global ~/.lazymake.yaml
safety:
  enabled: false
```

**Scenario 2: Enable only for critical operations**
```yaml
# Project .lazymake.yaml
safety:
  enabled: true
  enabled_rules:
    - rm-rf-root
    - database-drop
    - terraform-destroy
```

**Scenario 3: Trust all cleanup targets**
```yaml
# Project .lazymake.yaml
safety:
  exclude_targets:
    - clean
    - distclean
    - purge
    - reset
    - nuke-dev
    - nuke-test
```

**Scenario 4: Strict safety for production project**
```yaml
# Project .lazymake.yaml
safety:
  enabled: true  # All built-in rules enabled
  custom_rules:
    - id: "any-prod-operation"
      severity: critical
      patterns: [".*prod.*", ".*production.*"]
      description: "Any operation mentioning production"
      suggestion: "Triple-check production operations"
```

## Export Configuration

Export execution results to JSON/log files for analysis and debugging.

### Basic Export Settings

```yaml
export:
  # Enable/disable execution result exports (default: false)
  enabled: false

  # Output directory (default: ~/.cache/lazymake/exports)
  # Supports ~ expansion and environment variables
  output_dir: ~/.cache/lazymake/exports

  # Format: "json", "log", or "both" (default: "json")
  format: json

  # Naming strategy: "timestamp", "target", or "sequential" (default: "timestamp")
  naming_strategy: timestamp
```

### Naming Strategies

**timestamp**: Creates files like `build_20251212_143022.json`
- Best for historical analysis
- Keeps all execution records

**target**: Creates files like `build_latest.json`
- Overwrites previous result for same target
- Best for CI/CD where you only need latest result

**sequential**: Creates files like `build_1.json`, `build_2.json`
- Incremental numbering
- Best for tracking execution series

### Rotation Settings

Control automatic cleanup of old exports:

```yaml
export:
  # Maximum file size in MB (default: 0 = unlimited)
  # Files exceeding this size are removed during rotation
  max_file_size_mb: 0

  # Maximum files per target (default: 0 = unlimited)
  # Older files are automatically deleted
  max_files: 0

  # Keep exports for N days (default: 0 = forever)
  # Files older than N days are cleaned up
  keep_days: 0
```

### Filtering

```yaml
export:
  # Only export successful executions (default: false)
  success_only: false

  # Don't export results for these targets
  exclude_targets:
    - watch
    - dev
    - test-watch
```

### Export Configuration Scenarios

**Scenario 1: Enable for CI/CD integration**
```yaml
export:
  enabled: true
  format: json
  naming_strategy: target  # Overwrite latest result
  success_only: true
```

**Scenario 2: Debugging with logs**
```yaml
export:
  enabled: true
  format: both  # JSON + human-readable logs
  max_files: 20
  keep_days: 7
```

**Scenario 3: Long-term metrics collection**
```yaml
export:
  enabled: true
  format: json
  naming_strategy: timestamp
  keep_days: 90  # Keep 3 months of data
  exclude_targets:
    - help
    - list
    - clean
```

## Shell Integration

Add executed make commands to your shell history.

### Basic Shell Settings

```yaml
shell_integration:
  # Enable/disable shell history integration (default: false)
  enabled: false

  # Shell type: "auto", "bash", "zsh", "fish", or "none" (default: "auto")
  shell: auto

  # Override shell history file path (default: "")
  # Leave empty to use shell defaults:
  # - bash: ~/.bash_history
  # - zsh: ~/.zsh_history or $HISTFILE
  # - fish: ~/.local/share/fish/fish_history
  history_file: ""

  # Include timestamp in history entry (default: true)
  # - zsh: auto-detects extended history format and writes timestamps when detected
  # - fish: writes the "when:" timestamp field
  # - bash: no effect
  # When false, always writes plain entries regardless of shell.
  include_timestamp: true
```

### Format Templates

Customize the command format added to history:

```yaml
shell_integration:
  # Available variables: {target}, {makefile}, {dir}
  format_template: "make {target}"
```

**Examples:**
- `"make {target}"` → `make build`
- `"make -f {makefile} {target}"` → `make -f /path/to/Makefile build`
- `"cd {dir} && make {target}"` → `cd /path/to/project && make build`

### Exclude Targets from History

```yaml
shell_integration:
  exclude_targets:
    - help
    - list
```

### Shell Integration Scenarios

**Scenario 1: Enable for bash**
```yaml
shell_integration:
  enabled: true
  shell: bash
```

**Scenario 2: Enable for zsh with extended history**
```yaml
shell_integration:
  enabled: true
  shell: zsh
  include_timestamp: true
```

**Scenario 3: Enable for fish with timestamps**
```yaml
shell_integration:
  enabled: true
  shell: fish
  include_timestamp: true
```

**Scenario 4: Custom format with context**
```yaml
shell_integration:
  enabled: true
  format_template: "make -f {makefile} {target}"
  exclude_targets:
    - help
    - list
    - show-%
```

## Complete Example Configuration

Here's a comprehensive example combining multiple features:

```yaml
# Basic settings (makefile defaults to auto-detect if omitted)
# makefile: Makefile

# Safety features
safety:
  enabled: true
  exclude_targets:
    - clean
    - distclean
  custom_rules:
    - id: "prod-deploy"
      severity: critical
      patterns:
        - "kubectl apply.*production"
        - "terraform apply.*prod"
      description: "Deploying to production environment"
      suggestion: "Get team approval before deployment"

# Export results
export:
  enabled: true
  format: both
  naming_strategy: timestamp
  max_files: 50
  keep_days: 30
  exclude_targets:
    - watch
    - dev

# Shell integration
shell_integration:
  enabled: true
  shell: auto
  format_template: "make {target}"
  exclude_targets:
    - help
    - list
```

## Environment Variables

Some settings support environment variable expansion:
- `output_dir` in export configuration
- `history_file` in shell integration

Example:
```yaml
export:
  output_dir: $HOME/.lazymake/exports
```

## Tips

1. **Start minimal**: Begin with default settings and add configuration as needed
2. **Use project config**: Put project-specific rules in `./.lazymake.yaml`
3. **Global preferences**: Put personal preferences in `~/.lazymake.yaml`
4. **Test changes**: Run `lazymake` after changing config to verify behavior
5. **Share project config**: Commit `.lazymake.yaml` to version control for team consistency

## See Also

- [Full example configuration](../../.lazymake.example.yaml) - Comprehensive example with all options
- [Safety Features](../features/safety-features.md) - Detailed safety feature documentation
- [Export & Shell Integration](../features/export-shell-integration.md) - Export and shell integration details

---

[← Back to Documentation](../README.md) | [← Back to Main README](../../README.md)
