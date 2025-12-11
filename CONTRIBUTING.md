# Contributing to lazymake

Thank you for your interest in contributing to lazymake! This guide will help you get started.

## Table of Contents

- [Contributing Safety Rules](#contributing-safety-rules)
- [Development Setup](#development-setup)
- [Submitting Changes](#submitting-changes)

## Contributing Safety Rules

Safety rules detect dangerous commands in Makefiles and warn users before execution. Follow these guidelines when adding new rules:

### Guidelines

1. **Add your rule to the BuiltinRules slice** in `internal/safety/builtin_rules.go`

2. **Use specific regex patterns** to minimize false positives
   - ✓ Good: `rm\s+-rf\s+/[^/]` (specific, catches dangerous cases)
   - ✗ Bad: `rm` (too broad, many false positives)

3. **Provide clear description** explaining WHY it's dangerous
   - ✓ Good: "Drops production database without backup"
   - ✗ Bad: "Deletes stuff" (too vague)

4. **Add helpful suggestion** for safer alternatives
   - Example: "Use specific paths instead of wildcards. Double-check paths before execution."

5. **Test against common Makefiles** to verify pattern accuracy
   - Test with real-world Makefiles
   - Ensure no false positives
   - Verify dangerous cases are caught

6. **Consider context**: clean targets, dev vs prod environments, etc.
   - Context-aware severity adjustment happens automatically
   - Clean targets get downgraded severity
   - Production keywords elevate severity

7. **Choose appropriate severity**:
   - **Critical**: Irreversible system-wide damage (data loss, infrastructure destruction)
     - Examples: `rm -rf /`, `DROP DATABASE`, `terraform destroy`
   - **Warning**: Reversible or project-scoped issues (can rebuild, restore from git)
     - Examples: `docker system prune`, `git reset --hard`
   - **Info**: Educational only (no UI indicator)
     - For documentation purposes

### Example Rule

```go
{
    ID:       "example-dangerous-op",
    Severity: SeverityCritical,
    Patterns: []string{
        `dangerous-command\s+--force`,
        `risky-operation.*--no-confirm`,
    },
    Description: "Performs irreversible operation without confirmation. Data will be permanently lost.",
    Suggestion:  "Use --dry-run first to preview changes. Backup critical data before proceeding.",
}
```

### Testing Your Rule

Add tests to `internal/safety/rules_test.go`:

```go
{
    name:         "your-rule matches dangerous command",
    ruleID:       "your-rule-id",
    recipe:       []string{"dangerous-command --force"},
    shouldMatch:  true,
    expectedLine: "dangerous-command --force",
}
```

Run tests:
```bash
go test ./internal/safety/ -v
```

---

## Development Setup

*TODO: Add development setup instructions*

## Submitting Changes

*TODO: Add PR submission guidelines*
