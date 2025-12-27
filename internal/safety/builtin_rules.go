package safety

// BuiltinRules contains all default dangerous command patterns
// These rules are always available unless explicitly disabled in config
var BuiltinRules = []Rule{
	// ========== CRITICAL: System-wide destructive operations ==========

	{
		ID:       "rm-rf-root",
		Severity: SeverityCritical,
		Patterns: []string{
			`rm\s+(-\w*f\w*\s+){1,2}/[^/\s]`, // rm -rf /anything (not deep paths)
			`rm\s+(-\w*f\w*\s+){1,2}\$HOME`,  // rm -rf $HOME
			`sudo\s+rm\s+-\w*rf`,             // any sudo rm -rf
			`rm\s+(-\w*f\w*\s+){1,2}~`,       // rm -rf ~
			`rm\s+(-\w*f\w*\s+){1,2}\*`,      // rm -rf * (dangerous wildcard)
		},
		Description: "Removes files with root privileges or system-wide paths. This can permanently delete critical system files or all user data.",
		Suggestion:  "Use specific paths instead of wildcards or root directories. Double-check paths before execution.",
	},

	{
		ID:       "disk-wipe",
		Severity: SeverityCritical,
		Patterns: []string{
			`dd\s+.*of=/dev/(sd|hd|nvme)`, // dd to block device
			`mkfs\.\w+\s+/dev/`,           // filesystem creation
			`fdisk.*-w`,                   // write partition table
			`parted.*-s`,                  // script mode partitioning
		},
		Description: "Formats disks or writes to block devices. This will erase all data on the target device.",
		Suggestion:  "Triple-check device paths. Use 'lsblk' to verify correct device. Consider backing up first.",
	},

	{
		ID:       "database-drop",
		Severity: SeverityCritical,
		Patterns: []string{
			`(?i)drop\s+database`,                   // Case-insensitive DROP DATABASE
			`(?i)truncate\s+table`,                  // TRUNCATE TABLE
			`(?i)delete\s+from.*where\s+(1=1|true)`, // DELETE FROM ... WHERE 1=1
			`psql.*-c.*drop`,                        // psql with DROP command
			`mysql.*-e.*drop`,                       // mysql with DROP command
			`mongo.*dropDatabase`,                   // MongoDB dropDatabase
		},
		Description: "Drops databases or truncates tables. This causes permanent data loss.",
		Suggestion:  "Always backup before destructive database operations. Verify database name (production vs dev).",
	},

	{
		ID:       "git-force-push",
		Severity: SeverityCritical,
		Patterns: []string{
			`git\s+push.*\s+-f(\s|$)`,      // git push -f
			`git\s+push.*\s+--force(\s|$)`, // git push --force
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
			`tofu\s+destroy`, // OpenTofu
		},
		Description: "Destroys Terraform-managed infrastructure. This will tear down all resources (VMs, databases, networks, etc).",
		Suggestion:  "Run 'terraform plan -destroy' first to review changes. Verify workspace/environment. Consider using -target for specific resources.",
	},

	{
		ID:       "kubectl-delete",
		Severity: SeverityCritical,
		Patterns: []string{
			`kubectl\s+delete\s+(namespace|ns)`,  // Delete namespace
			`kubectl\s+delete\s+(pvc|pv)`,        // Delete persistent volumes
			`kubectl\s+delete.*--all`,            // Delete all resources
			`kubectl\s+delete.*-A`,               // All namespaces
			`kubectl\s+delete.*--all-namespaces`, // All namespaces explicit
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
			`git\s+clean\s+-\w*fd`, // git clean -fd or -fdx
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

	// ========== CRITICAL: Cloud infrastructure destruction ==========

	{
		ID:       "aws-s3-delete",
		Severity: SeverityCritical,
		Patterns: []string{
			`aws\s+s3\s+rm\s+.*--recursive`, // aws s3 rm --recursive
			`aws\s+s3\s+rb\s+.*--force`,     // aws s3 rb --force (remove bucket)
			`aws\s+s3api\s+delete-bucket`,   // aws s3api delete-bucket
		},
		Description: "Deletes S3 bucket contents or entire buckets. Data will be permanently lost.",
		Suggestion:  "Verify bucket name carefully. Enable versioning and MFA delete for critical buckets. Use --dryrun first.",
	},

	{
		ID:       "cloud-instance-terminate",
		Severity: SeverityCritical,
		Patterns: []string{
			`aws\s+ec2\s+terminate-instances`,       // AWS EC2 terminate
			`gcloud\s+compute\s+instances\s+delete`, // GCP instance delete
			`az\s+vm\s+delete`,                      // Azure VM delete
		},
		Description: "Terminates cloud compute instances. Running workloads and ephemeral data will be lost.",
		Suggestion:  "Double-check instance IDs. Consider stopping instead of terminating. Verify correct account/project.",
	},

	{
		ID:       "curl-pipe-shell",
		Severity: SeverityCritical,
		Patterns: []string{
			`curl\s+.*\|\s*(ba)?sh`,          // curl ... | sh or bash
			`wget\s+.*\|\s*(ba)?sh`,          // wget ... | sh or bash
			`curl\s+.*\|\s*sudo\s+(ba)?sh`,   // curl ... | sudo sh
			`wget\s+.*-O\s*-\s*\|\s*(ba)?sh`, // wget -O - | sh
		},
		Description: "Pipes remote content directly to shell. Executes arbitrary code from the internet without inspection.",
		Suggestion:  "Download script first, review it, then execute. Use checksums to verify integrity.",
	},

	// ========== WARNING: Additional destructive operations ==========

	{
		ID:       "firewall-flush",
		Severity: SeverityWarning,
		Patterns: []string{
			`iptables\s+(-F|--flush)`,   // iptables flush
			`ufw\s+disable`,             // ufw disable
			`firewall-cmd\s+.*--remove`, // firewalld remove rules
			`nft\s+flush\s+ruleset`,     // nftables flush
		},
		Description: "Flushes or disables firewall rules. System may become exposed to network attacks.",
		Suggestion:  "Backup firewall rules first. Use 'iptables-save' before flushing. Test in staging environment.",
	},

	{
		ID:       "process-kill-force",
		Severity: SeverityWarning,
		Patterns: []string{
			`kill\s+-9`,                 // kill -9 (SIGKILL)
			`killall\s+-9`,              // killall -9
			`pkill\s+-9`,                // pkill -9
			`killall\s+.*-s\s+(KILL|9)`, // killall with SIGKILL
		},
		Description: "Force kills processes without allowing graceful shutdown. May cause data corruption.",
		Suggestion:  "Use SIGTERM (kill -15) first to allow graceful shutdown. Only use -9 if process doesn't respond.",
	},

	{
		ID:       "helm-delete",
		Severity: SeverityWarning,
		Patterns: []string{
			`helm\s+(delete|uninstall)`, // helm delete/uninstall
			`helm\s+.*--purge`,          // helm with --purge flag
		},
		Description: "Deletes Helm releases from Kubernetes. Services and associated resources will be removed.",
		Suggestion:  "Use 'helm list' to verify release name. Check current kubectl context before deletion.",
	},

	{
		ID:       "ssh-key-delete",
		Severity: SeverityWarning,
		Patterns: []string{
			`rm\s+.*\.ssh`,           // rm anything in .ssh
			`rm\s+.*id_rsa`,          // rm id_rsa
			`rm\s+.*id_ed25519`,      // rm id_ed25519
			`rm\s+.*authorized_keys`, // rm authorized_keys
		},
		Description: "Deletes SSH keys or configuration. May lose access to remote servers.",
		Suggestion:  "Backup SSH keys before deletion. Verify you have alternative access to remote systems.",
	},

	{
		ID:       "env-file-overwrite",
		Severity: SeverityWarning,
		Patterns: []string{
			`>\s*\.env(\s|$)`,               // overwrite .env file
			`cp\s+.*\.env\.example\s+\.env`, // copy example to .env (might overwrite)
			`mv\s+.*\.env`,                  // move .env file
		},
		Description: "Overwrites environment configuration files. Existing secrets and settings may be lost.",
		Suggestion:  "Backup .env files before overwriting. Use version control for environment templates.",
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
