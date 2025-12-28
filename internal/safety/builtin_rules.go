package safety

// BuiltinRules contains all default dangerous command patterns
// These rules are always available unless explicitly disabled in config
var BuiltinRules = []Rule{
	// ========== CRITICAL: System-wide destructive operations ==========

	{
		ID:       "rm-rf-root",
		Severity: SeverityCritical,
		Patterns: []string{
			`rm\s+(-\w*f\w*\s+){1,2}/[^/\s]`,  // rm -rf /anything (not deep paths)
			`rm\s+(-\w*f\w*\s+){1,2}\$HOME`,   // rm -rf $HOME
			`sudo\s+rm\s+-\w*rf`,               // any sudo rm -rf
			`rm\s+(-\w*f\w*\s+){1,2}~`,        // rm -rf ~
			`rm\s+(-\w*f\w*\s+){1,2}\*`,       // rm -rf * (dangerous wildcard)
		},
		Description: "Removes files with root privileges or system-wide paths. This can permanently delete critical system files or all user data.",
		Suggestion:  "Use specific paths instead of wildcards or root directories. Double-check paths before execution.",
	},

	{
		ID:       "disk-wipe",
		Severity: SeverityCritical,
		Patterns: []string{
			`dd\s+.*of=/dev/(sd|hd|nvme)`,  // dd to block device
			`mkfs\.\w+\s+/dev/`,             // filesystem creation
			`fdisk.*-w`,                     // write partition table
			`parted.*-s`,                    // script mode partitioning
		},
		Description: "Formats disks or writes to block devices. This will erase all data on the target device.",
		Suggestion:  "Triple-check device paths. Use 'lsblk' to verify correct device. Consider backing up first.",
	},

	{
		ID:       "database-drop",
		Severity: SeverityCritical,
		Patterns: []string{
			`(?i)drop\s+database`,                     // Case-insensitive DROP DATABASE
			`(?i)truncate\s+table`,                    // TRUNCATE TABLE
			`(?i)delete\s+from.*where\s+(1=1|true)`,  // DELETE FROM ... WHERE 1=1
			`psql.*-c.*drop`,                          // psql with DROP command
			`mysql.*-e.*drop`,                         // mysql with DROP command
			`mongo.*dropDatabase`,                     // MongoDB dropDatabase
		},
		Description: "Drops databases or truncates tables. This causes permanent data loss.",
		Suggestion:  "Always backup before destructive database operations. Verify database name (production vs dev).",
	},

	{
		ID:       "git-force-push",
		Severity: SeverityCritical,
		Patterns: []string{
			`git\s+push.*\s+-f(\s|$)`,        // git push -f
			`git\s+push.*\s+--force(\s|$)`,   // git push --force
		},
		Description: "Force pushes to git repository, potentially overwriting others' work and losing history.",
		Suggestion:  "Coordinate with team before force pushing. Use --force-with-lease for safer alternative. Verify branch name.",
	},

	{
		ID:       "terraform-destroy",
		Severity: SeverityCritical,
		Patterns: []string{
			`terraform\s+destroy`,
			`terraform\s+apply.*-destroy`,
			`tofu\s+destroy`,  // OpenTofu
		},
		Description: "Destroys Terraform-managed infrastructure. This will tear down all resources (VMs, databases, networks, etc).",
		Suggestion:  "Run 'terraform plan -destroy' first to review changes. Verify workspace/environment. Consider using -target for specific resources.",
	},

	{
		ID:       "kubectl-delete",
		Severity: SeverityCritical,
		Patterns: []string{
			`kubectl\s+delete\s+(namespace|ns)`,       // Delete namespace
			`kubectl\s+delete\s+(pvc|pv)`,             // Delete persistent volumes
			`kubectl\s+delete.*--all`,                  // Delete all resources
			`kubectl\s+delete.*-A`,                     // All namespaces
			`kubectl\s+delete.*--all-namespaces`,      // All namespaces explicit
		},
		Description: "Deletes Kubernetes resources, namespaces, or persistent volumes. Data in PVs will be lost.",
		Suggestion:  "Use 'kubectl get' first to verify resources. Check current context with 'kubectl config current-context'.",
	},

	// ========== WARNING: Project-level destructive operations ==========

	{
		ID:       "docker-system-prune",
		Severity: SeverityWarning,
		Patterns: []string{
			`docker\s+system\s+prune`,
			`docker\s+volume\s+(prune|rm).*-f`,
			`docker\s+image\s+prune.*-a`,
			`docker\s+container\s+prune`,
		},
		Description: "Removes Docker volumes, images, or containers. May delete data or require lengthy rebuilds.",
		Suggestion:  "Use specific container/volume names instead of prune. Consider impact on local development.",
	},

	{
		ID:       "git-reset-hard",
		Severity: SeverityWarning,
		Patterns: []string{
			`git\s+reset\s+--hard`,
			`git\s+clean\s+-\w*fd`,  // git clean -fd or -fdx
		},
		Description: "Discards uncommitted changes permanently. Untracked files will be deleted.",
		Suggestion:  "Stash changes with 'git stash' for recovery. Review changes with 'git status' and 'git diff' first.",
	},

	{
		ID:       "npm-uninstall-all",
		Severity: SeverityWarning,
		Patterns: []string{
			`rm\s+-\w*rf\w*\s+node_modules`,
			`npm\s+uninstall.*-g`,
			`pnpm\s+uninstall.*-g`,
			`yarn\s+global\s+remove`,
		},
		Description: "Removes all Node.js dependencies or global packages. Requires reinstall.",
		Suggestion:  "Use 'npm ci' or 'pnpm install --frozen-lockfile' to reinstall from lock file.",
	},

	{
		ID:       "package-remove",
		Severity: SeverityWarning,
		Patterns: []string{
			`apt(-get)?\s+remove`,
			`yum\s+remove`,
			`dnf\s+remove`,
			`brew\s+uninstall`,
			`pacman\s+-R`,
		},
		Description: "Removes system packages. May break system dependencies.",
		Suggestion:  "Verify package names before removal. Consider using package manager's simulation mode first.",
	},

	{
		ID:       "chmod-777",
		Severity: SeverityWarning,
		Patterns: []string{
			`chmod\s+(-R\s+)?777`,
			`chmod\s+(-R\s+)?a\+rwx`,
		},
		Description: "Sets overly permissive file permissions (777 = world-writable). Security risk.",
		Suggestion:  "Use more restrictive permissions. Typically 755 for executables, 644 for files.",
	},

	{
		ID:       "deployment-commands",
		Severity: SeverityWarning,
		Patterns: []string{
			`kubectl\s+apply`,
			`terraform\s+apply`,
			`tofu\s+apply`,  // OpenTofu
			`helm\s+install`,
			`helm\s+upgrade`,
		},
		Description: "Deploys or applies infrastructure changes. May affect running systems.",
		Suggestion:  "Review changes with plan/diff first. Verify target environment. Consider using staging before production.",
	},
}

// GetBuiltinRuleByID returns a built-in rule by its ID, or nil if not found
func GetBuiltinRuleByID(id string) *Rule {
	for i := range BuiltinRules {
		if BuiltinRules[i].ID == id {
			return &BuiltinRules[i]
		}
	}
	return nil
}

// Contributing New Rules:
//
// 1. Add your rule to the BuiltinRules slice above
// 2. Use specific regex patterns to minimize false positives
// 3. Provide clear description explaining WHY it's dangerous
// 4. Add helpful suggestion for safer alternatives
// 5. Test against common Makefiles to verify pattern accuracy
// 6. Consider context: clean targets, dev vs prod environments, etc.
// 7. Choose appropriate severity:
//    - Critical: Irreversible system-wide damage (data loss, infra destruction)
//    - Warning:  Reversible or project-scoped issues (can rebuild, restore from git)
//    - Info:     Educational only (no UI indicator)
//
// Examples of good patterns:
//   ✓ `rm\s+-rf\s+/[^/]`          - Specific, catches dangerous cases
//   ✗ `rm`                         - Too broad, many false positives
//
// Examples of good descriptions:
//   ✓ "Drops production database without backup"
//   ✗ "Deletes stuff"  // Too vague
