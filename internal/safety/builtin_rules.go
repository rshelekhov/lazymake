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

	// ========== CRITICAL: Cloud provider operations ==========

	{
		ID:       "aws-destructive",
		Severity: SeverityCritical,
		Patterns: []string{
			`aws\s+cloudformation\s+delete-stack`,
			`aws\s+s3\s+rb\s+.*--force`,
			`aws\s+s3\s+rm\s+.*--recursive`,
			`aws\s+ec2\s+terminate-instances`,
			`aws\s+rds\s+delete-db-instance`,
			`aws\s+rds\s+delete-db-cluster`,
		},
		Description: "Deletes AWS resources (CloudFormation stacks, S3 buckets, EC2 instances, RDS databases). This will destroy cloud infrastructure and data permanently.",
		Suggestion:  "Verify AWS profile and region. Use 'aws cloudformation describe-stacks' or 'aws s3 ls' to review resources first. Enable deletion protection for critical resources.",
	},

	{
		ID:       "gcp-destructive",
		Severity: SeverityCritical,
		Patterns: []string{
			`gcloud\s+projects\s+delete`,
			`gcloud\s+compute\s+instances\s+delete`,
			`gcloud\s+sql\s+instances\s+delete`,
			`gsutil\s+rb`,
			`gsutil\s+rm\s+.*-r`,
		},
		Description: "Deletes GCP resources (projects, instances, Cloud SQL, GCS buckets). This will destroy cloud infrastructure and data.",
		Suggestion:  "Verify GCP project with 'gcloud config list'. Use 'gcloud compute instances list' to review resources first.",
	},

	{
		ID:       "azure-destructive",
		Severity: SeverityCritical,
		Patterns: []string{
			`az\s+group\s+delete`,
			`az\s+vm\s+delete`,
			`az\s+sql\s+server\s+delete`,
			`az\s+storage\s+account\s+delete`,
		},
		Description: "Deletes Azure resources (resource groups, VMs, SQL servers, storage accounts). This will destroy cloud infrastructure.",
		Suggestion:  "Verify Azure subscription with 'az account show'. Use 'az group list' to review resources first.",
	},

	{
		ID:       "heroku-destructive",
		Severity: SeverityCritical,
		Patterns: []string{
			`heroku\s+apps:destroy`,
			`heroku\s+addons:destroy`,
			`heroku\s+pg:reset`,
		},
		Description: "Destroys Heroku applications, addons, or resets databases. This causes permanent data loss.",
		Suggestion:  "Verify app name with 'heroku apps'. Create a backup with 'heroku pg:backups:capture' first.",
	},

	// ========== CRITICAL: Additional database operations ==========

	{
		ID:       "redis-flush",
		Severity: SeverityCritical,
		Patterns: []string{
			`redis-cli\s+.*FLUSHALL`,
			`redis-cli\s+.*FLUSHDB`,
			`redis-cli\s+.*flushall`,
			`redis-cli\s+.*flushdb`,
		},
		Description: "Flushes all data from Redis (FLUSHALL clears all databases, FLUSHDB clears current database). Data loss is immediate and unrecoverable.",
		Suggestion:  "Verify Redis instance with 'redis-cli INFO'. Consider using 'SCAN' and 'DEL' for targeted deletion. Backup with 'BGSAVE' first.",
	},

	{
		ID:       "cassandra-drop",
		Severity: SeverityCritical,
		Patterns: []string{
			`(?i)cqlsh.*DROP\s+KEYSPACE`,
			`(?i)cqlsh.*DROP\s+TABLE`,
			`(?i)nodetool\s+clearsnapshot`,
		},
		Description: "Drops Cassandra keyspaces or tables, or clears snapshots. This causes permanent data loss.",
		Suggestion:  "Verify keyspace with 'DESCRIBE KEYSPACES'. Create snapshot with 'nodetool snapshot' first.",
	},

	// ========== CRITICAL: System operations ==========

	{
		ID:       "crontab-remove",
		Severity: SeverityCritical,
		Patterns: []string{
			`crontab\s+-r`,
			`crontab\s+-ri`,
		},
		Description: "Removes all cron jobs for the current user. Scheduled tasks will stop running.",
		Suggestion:  "Backup crontab with 'crontab -l > crontab.backup' first. Use 'crontab -e' to edit specific jobs.",
	},

	{
		ID:       "iptables-flush",
		Severity: SeverityCritical,
		Patterns: []string{
			`iptables\s+-F`,
			`iptables\s+--flush`,
			`ip6tables\s+-F`,
			`nft\s+flush\s+ruleset`,
		},
		Description: "Flushes all firewall rules. This may expose services to the network or lock you out of remote servers.",
		Suggestion:  "Save rules with 'iptables-save > rules.backup' first. Test changes in a staging environment.",
	},

	// ========== WARNING: Version control destructive operations ==========

	{
		ID:       "git-branch-delete-force",
		Severity: SeverityWarning,
		Patterns: []string{
			`git\s+branch\s+-D`,
			`git\s+branch\s+--delete\s+--force`,
		},
		Description: "Force deletes a git branch regardless of merge status. Unmerged commits may be lost.",
		Suggestion:  "Use 'git branch -d' (lowercase) for safe deletion that checks merge status. Verify branch with 'git log' first.",
	},

	{
		ID:       "git-reflog-expire",
		Severity: SeverityWarning,
		Patterns: []string{
			`git\s+reflog\s+expire`,
			`git\s+gc\s+--prune=now`,
		},
		Description: "Expires reflog entries or prunes objects immediately. This removes the ability to recover from mistakes.",
		Suggestion:  "Use default 'git gc' which keeps recent history. Reflog is your safety net for recovering lost commits.",
	},

	// ========== WARNING: Container orchestration ==========

	{
		ID:       "docker-swarm-destructive",
		Severity: SeverityWarning,
		Patterns: []string{
			`docker\s+stack\s+rm`,
			`docker\s+swarm\s+leave\s+--force`,
			`docker\s+service\s+rm`,
		},
		Description: "Removes Docker swarm stacks, services, or leaves the swarm cluster. Running services will be stopped.",
		Suggestion:  "Use 'docker stack services' to review running services. Scale down gradually if possible.",
	},

	{
		ID:       "podman-system-reset",
		Severity: SeverityWarning,
		Patterns: []string{
			`podman\s+system\s+reset`,
			`podman\s+volume\s+prune\s+-f`,
		},
		Description: "Resets all Podman data or removes unused volumes. All containers, images, and volumes may be deleted.",
		Suggestion:  "Use 'podman ps -a' and 'podman volume ls' to review resources first.",
	},

	// ========== WARNING: Package managers ==========

	{
		ID:       "pip-uninstall-all",
		Severity: SeverityWarning,
		Patterns: []string{
			`pip\s+uninstall\s+.*-y`,
			`pip3\s+uninstall\s+.*-y`,
			`pip\s+freeze.*xargs.*pip\s+uninstall`,
		},
		Description: "Uninstalls Python packages without confirmation. May break Python environments.",
		Suggestion:  "Use virtual environments (venv/virtualenv). Review packages with 'pip list' first.",
	},

	{
		ID:       "go-clean-modcache",
		Severity: SeverityWarning,
		Patterns: []string{
			`go\s+clean\s+-modcache`,
			`go\s+clean\s+.*-cache.*-modcache`,
		},
		Description: "Removes all downloaded Go modules from the cache. Requires re-downloading on next build.",
		Suggestion:  "Use 'go clean -cache' to clear build cache only. Module cache is shared across projects.",
	},

	// ========== WARNING: Critical service operations ==========

	{
		ID:       "systemctl-critical-services",
		Severity: SeverityWarning,
		Patterns: []string{
			`systemctl\s+(stop|disable)\s+(nginx|apache2|httpd|postgresql|mysql|mariadb|docker|kubelet|sshd)`,
			`service\s+(nginx|apache2|httpd|postgresql|mysql|mariadb|docker|sshd)\s+stop`,
		},
		Description: "Stops or disables critical system services. May cause downtime or lock you out of servers.",
		Suggestion:  "Use 'systemctl status' to check service state. Consider using 'systemctl reload' for config changes.",
	},

	{
		ID:       "killall-force",
		Severity: SeverityWarning,
		Patterns: []string{
			`killall\s+-9`,
			`killall\s+--signal\s+KILL`,
			`pkill\s+-9`,
		},
		Description: "Force kills all processes by name. Processes won't have a chance to cleanup, may cause data corruption.",
		Suggestion:  "Use 'killall' without -9 first to allow graceful shutdown. Check processes with 'pgrep' before killing.",
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
