# Safety Features: Preventing Accidental Disasters

lazymake protects you from accidentally running destructive commands by detecting dangerous patterns and requiring confirmation before execution.

## Visual Indicators

Targets are marked with colored circle indicators based on danger level:

```
RECENT
○  deploy-prod      Deploy to production         (yellow - Warning)
○  clean            Clean build artifacts        (yellow - Warning)
──────────────────────────────────────────────
ALL TARGETS
   build            Build the application
○  nuke-db          Drop production database     (red - Critical)
○  docker-clean     Clean Docker resources       (blue - Info)
```

- **○ Critical (red)**: Commands that can cause irreversible damage (requires confirmation)
- **○ Warning (yellow)**: Commands with potential side effects (executes immediately)
- **○ Info (blue)**: Informational warnings about best practices
- **No indicator**: Safe commands

## Two-Column Layout

The interface shows exactly what will execute:

```
┌─────────────────────┬────────────────────────────────────────┐
│ RECENT              │ Recipe Preview                         │
│ > ○ deploy-prod     │                                        │
│   ○ clean           │ deploy-prod:                           │
│                     │   kubectl apply -f k8s/prod/           │
│ ───────────────     │   terraform apply -var-file=prod.tfvars│
│ ALL TARGETS         │                                        │
│   build             │ ─────────────────────────────────────  │
│   test              │                                        │
│                     │ ┌─────────────────────────────────┐    │
│                     │ │ ○ Warning deployment-commands   │    │
│                     │ │                                 │    │
│                     │ │ Matched: terraform apply ...    │    │
│                     │ │                                 │    │
│                     │ │ Deploys infrastructure changes  │    │
│                     │ │                                 │    │
│                     │ │ ℹ Verify target environment     │    │
│                     │ └─────────────────────────────────┘    │
└─────────────────────┴────────────────────────────────────────┘
│ 12 targets • 2 dangerous • 1 critical  enter: run • ?: help  │
└──────────────────────────────────────────────────────────────┘
```

**Left column**: Target list with colored circle indicators
**Right column**: Recipe commands + safety warnings in bordered boxes
**Bottom**: Stats with separate counts for dangerous/critical targets

## Confirmation Dialog

Critical commands show a prominent warning dialog:

```
╔════════════════════════════════════════════════════════╗
║                                                        ║
║     ○ CRITICAL COMMAND WARNING                         ║
║                                                        ║
║     Target: nuke-db                                    ║
║                                                        ║
║     ○ Critical database-drop                           ║
║     Command: psql -c 'DROP DATABASE production;'       ║
║                                                        ║
║     Drops databases or truncates tables. This causes   ║
║     permanent data loss.                               ║
║                                                        ║
║     ℹ Always backup before destructive database        ║
║        operations. Verify database name.               ║
║                                                        ║
║     [Enter] Continue Anyway     [Esc] Cancel           ║
║                                                        ║
╚════════════════════════════════════════════════════════╝
```

- Press `Esc` to cancel safely
- Press `Enter` to proceed with execution

## Built-in Dangerous Patterns

### Critical (○ red + Confirmation Required)

- **rm-rf-root**: Recursive deletion of system paths (`rm -rf /`, `sudo rm -rf`)
- **disk-wipe**: Disk formatting or block device writes (`dd`, `mkfs`)
- **database-drop**: Database/table deletion (`DROP DATABASE`, `TRUNCATE TABLE`)
- **git-force-push**: Force pushing to repositories (`git push -f`)
- **terraform-destroy**: Infrastructure destruction (`terraform destroy`)
- **kubectl-delete**: Kubernetes resource deletion (`kubectl delete namespace`)

### Warning (○ yellow)

- **deployment-commands**: Deployment operations (`kubectl apply`, `terraform apply`)
- **docker-system-prune**: Docker cleanup operations
- **git-reset-hard**: Discarding uncommitted changes
- **npm-uninstall-all**: Removing all dependencies
- **package-remove**: System package removal
- **chmod-777**: Overly permissive file permissions

### Info (○ blue)

- Informational warnings about commands that may have side effects but are generally safe
- Clean targets with destructive commands are typically downgraded to Info level

## Context-Aware Detection

lazymake intelligently adjusts severity based on context:

**Clean targets** (downgraded severity):
```makefile
clean:  ## Clean build artifacts
	rm -rf build/  # Warning instead of Critical
```

**Database/production clean targets** (severity NOT downgraded):
```makefile
nuke-db:  ## Drop production database
	psql -c 'DROP DATABASE production;'  # Critical (not downgraded despite "nuke" keyword)
```

**Interactive flags** (downgraded severity):
```makefile
dangerous-op:
	rm -rfi build/  # Warning instead of Critical (has -i flag)
```

**Development targets** (downgraded for non-prod):
```makefile
test-cleanup:
	terraform destroy  # Warning instead of Critical (test target)
```

## Configuration

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

## Disabling Safety Checks

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

## Why Default-Enabled?

Safety checks are enabled by default because:
1. **Protect newcomers**: Developers new to a project are most at risk
2. **Non-intrusive**: Visual indicators don't block workflow
3. **Easy to disable**: One config line turns it off if unwanted
4. **Better safe than sorry**: Confirmation dialogs take 1 second; data recovery takes hours

## Real-World Scenarios

**Scenario 1: New developer runs `make clean`**
```
○ clean  # Yellow warning indicator shown
# Recipe preview shows: rm -rf build/
# Executes immediately (warning-level, clean target)
```

**Scenario 2: Accidentally select `make nuke-prod-db`**
```
○ nuke-prod-db  # Red critical indicator shown
# Presses Enter
# Confirmation dialog appears with full warning in bordered box
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

---

[← Back to Documentation](../README.md) | [← Back to Main README](../../README.md)
