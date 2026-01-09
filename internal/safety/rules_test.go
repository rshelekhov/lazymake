package safety

import (
	"testing"

	"github.com/rshelekhov/lazymake/internal/makefile"
)

func TestRuleMatching(t *testing.T) {
	tests := []struct {
		name         string
		ruleID       string
		recipe       []string
		shouldMatch  bool
		expectedLine string
	}{
		{
			name:         "rm-rf-root matches dangerous rm",
			ruleID:       "rm-rf-root",
			recipe:       []string{"rm -rf /tmp"},
			shouldMatch:  true,
			expectedLine: "rm -rf /tmp",
		},
		{
			name:        "rm-rf-root matches sudo rm",
			ruleID:      "rm-rf-root",
			recipe:      []string{"sudo rm -rf /var/cache"},
			shouldMatch: true,
		},
		{
			name:        "rm-rf-root safe rm does not match",
			ruleID:      "rm-rf-root",
			recipe:      []string{"rm -f build/artifact.o"},
			shouldMatch: false,
		},
		{
			name:        "database-drop matches DROP DATABASE",
			ruleID:      "database-drop",
			recipe:      []string{"psql -c 'DROP DATABASE production;'"},
			shouldMatch: true,
		},
		{
			name:        "database-drop case insensitive",
			ruleID:      "database-drop",
			recipe:      []string{"psql -c 'drop database test;'"},
			shouldMatch: true,
		},
		{
			name:        "git-force-push matches force push",
			ruleID:      "git-force-push",
			recipe:      []string{"git push -f origin main"},
			shouldMatch: true,
		},
		{
			name:        "git-force-push normal push safe",
			ruleID:      "git-force-push",
			recipe:      []string{"git push origin feature-branch"},
			shouldMatch: false,
		},
		{
			name:        "docker-system-prune matches",
			ruleID:      "docker-system-prune",
			recipe:      []string{"docker system prune -f"},
			shouldMatch: true,
		},
		{
			name:        "terraform-destroy matches",
			ruleID:      "terraform-destroy",
			recipe:      []string{"terraform destroy"},
			shouldMatch: true,
		},
		{
			name:        "kubectl-delete matches namespace deletion",
			ruleID:      "kubectl-delete",
			recipe:      []string{"kubectl delete namespace prod"},
			shouldMatch: true,
		},
		// New rule tests: aws-s3-delete
		{
			name:        "aws-s3-delete matches recursive delete",
			ruleID:      "aws-s3-delete",
			recipe:      []string{"aws s3 rm s3://my-bucket/ --recursive"},
			shouldMatch: true,
		},
		{
			name:   "aws-s3-delete matches force remove bucket",
			ruleID: "aws-s3-delete",
		},
		// Cloud provider tests
		{
			name:        "aws-destructive matches cloudformation delete-stack",
			ruleID:      "aws-destructive",
			recipe:      []string{"aws cloudformation delete-stack --stack-name prod-stack"},
			shouldMatch: true,
		},
		{
			name:        "aws-destructive matches s3 rb force",
			ruleID:      "aws-destructive",
			recipe:      []string{"aws s3 rb s3://my-bucket --force"},
			shouldMatch: true,
		},
		{
			name:        "aws-s3-delete safe s3 copy does not match",
			ruleID:      "aws-s3-delete",
			recipe:      []string{"aws s3 cp file.txt s3://bucket/"},
			shouldMatch: false,
		},
		// New rule tests: cloud-instance-terminate
		{
			name:        "cloud-instance-terminate matches AWS EC2",
			ruleID:      "cloud-instance-terminate",
			recipe:      []string{"aws ec2 terminate-instances --instance-ids i-1234567890abcdef0"},
			shouldMatch: true,
		},
		{
			name:        "cloud-instance-terminate matches GCP",
			ruleID:      "cloud-instance-terminate",
			recipe:      []string{"gcloud compute instances delete my-instance"},
			shouldMatch: true,
		},
		{
			name:        "cloud-instance-terminate matches Azure",
			ruleID:      "cloud-instance-terminate",
			recipe:      []string{"az vm delete --resource-group myGroup --name myVM"},
			shouldMatch: true,
		},
		{
			name:        "cloud-instance-terminate safe stop does not match",
			ruleID:      "cloud-instance-terminate",
			recipe:      []string{"aws ec2 stop-instances --instance-ids i-123"},
			shouldMatch: false,
		},
		// New rule tests: curl-pipe-shell
		{
			name:        "curl-pipe-shell matches curl to bash",
			ruleID:      "curl-pipe-shell",
			recipe:      []string{"curl -sSL https://example.com/install.sh | bash"},
			shouldMatch: true,
		},
		{
			name:        "curl-pipe-shell matches wget to sh",
			ruleID:      "curl-pipe-shell",
			recipe:      []string{"wget -qO- https://example.com/script.sh | sh"},
			shouldMatch: true,
		},
		{
			name:        "curl-pipe-shell matches curl to sudo bash",
			ruleID:      "curl-pipe-shell",
			recipe:      []string{"curl https://example.com/install.sh | sudo bash"},
			shouldMatch: true,
		},
		{
			name:        "curl-pipe-shell safe curl to file does not match",
			ruleID:      "curl-pipe-shell",
			recipe:      []string{"curl -o script.sh https://example.com/install.sh"},
			shouldMatch: false,
		},
		// New rule tests: firewall-flush
		{
			name:        "firewall-flush matches iptables flush",
			ruleID:      "firewall-flush",
			recipe:      []string{"iptables -F"},
			shouldMatch: true,
		},
		{
			name:        "aws-destructive safe s3 ls does not match",
			ruleID:      "aws-destructive",
			recipe:      []string{"aws s3 ls s3://my-bucket"},
			shouldMatch: false,
		},
		{
			name:        "gcp-destructive matches projects delete",
			ruleID:      "gcp-destructive",
			recipe:      []string{"gcloud projects delete my-project"},
			shouldMatch: true,
		},
		{
			name:        "azure-destructive matches group delete",
			ruleID:      "azure-destructive",
			recipe:      []string{"az group delete --name my-resource-group"},
			shouldMatch: true,
		},
		{
			name:        "heroku-destructive matches apps destroy",
			ruleID:      "heroku-destructive",
			recipe:      []string{"heroku apps:destroy --app myapp"},
			shouldMatch: true,
		},
		// Database tests
		{
			name:        "redis-flush matches FLUSHALL",
			ruleID:      "redis-flush",
			recipe:      []string{"redis-cli FLUSHALL"},
			shouldMatch: true,
		},
		{
			name:        "redis-flush matches lowercase flushdb",
			ruleID:      "redis-flush",
			recipe:      []string{"redis-cli -h localhost flushdb"},
			shouldMatch: true,
		},
		{
			name:        "redis-flush safe GET does not match",
			ruleID:      "redis-flush",
			recipe:      []string{"redis-cli GET mykey"},
			shouldMatch: false,
		},
		{
			name:        "cassandra-drop matches DROP KEYSPACE",
			ruleID:      "cassandra-drop",
			recipe:      []string{"cqlsh -e 'DROP KEYSPACE production'"},
			shouldMatch: true,
		},
		// System operation tests
		{
			name:        "crontab-remove matches crontab -r",
			ruleID:      "crontab-remove",
			recipe:      []string{"crontab -r"},
			shouldMatch: true,
		},
		{
			name:        "crontab-remove safe crontab -l does not match",
			ruleID:      "crontab-remove",
			recipe:      []string{"crontab -l"},
			shouldMatch: false,
		},
		{
			name:        "iptables-flush matches iptables -F",
			ruleID:      "iptables-flush",
			recipe:      []string{"iptables -F"},
			shouldMatch: true,
		},
		{
			name:        "firewall-flush matches ufw disable",
			ruleID:      "firewall-flush",
			recipe:      []string{"ufw disable"},
			shouldMatch: true,
		},
		{
			name:        "firewall-flush matches iptables --flush",
			ruleID:      "firewall-flush",
			recipe:      []string{"iptables --flush"},
			shouldMatch: true,
		},
		{
			name:        "firewall-flush safe iptables list does not match",
			ruleID:      "firewall-flush",
			recipe:      []string{"iptables -L"},
			shouldMatch: false,
		},
		// New rule tests: process-kill-force
		{
			name:        "process-kill-force matches kill -9",
			ruleID:      "process-kill-force",
			recipe:      []string{"kill -9 12345"},
			shouldMatch: true,
		},
		{
			name:        "process-kill-force matches killall -9",
			ruleID:      "process-kill-force",
			recipe:      []string{"killall -9 nginx"},
			shouldMatch: true,
		},
		{
			name:        "process-kill-force safe kill does not match",
			ruleID:      "process-kill-force",
			recipe:      []string{"kill 12345"},
			shouldMatch: false,
		},
		// New rule tests: helm-delete
		{
			name:        "helm-delete matches helm uninstall",
			ruleID:      "helm-delete",
			recipe:      []string{"helm uninstall my-release"},
			shouldMatch: true,
		},
		{
			name:        "helm-delete matches helm delete",
			ruleID:      "helm-delete",
			recipe:      []string{"helm delete my-release"},
			shouldMatch: true,
		},
		{
			name:        "helm-delete safe helm list does not match",
			ruleID:      "helm-delete",
			recipe:      []string{"helm list"},
			shouldMatch: false,
		},
		// New rule tests: ssh-key-delete
		{
			name:        "ssh-key-delete matches rm .ssh",
			ruleID:      "ssh-key-delete",
			recipe:      []string{"rm -rf ~/.ssh"},
			shouldMatch: true,
		},
		{
			name:        "ssh-key-delete matches rm id_rsa",
			ruleID:      "ssh-key-delete",
			recipe:      []string{"rm ~/.ssh/id_rsa"},
			shouldMatch: true,
		},
		{
			name:        "ssh-key-delete safe ssh-keygen does not match",
			ruleID:      "ssh-key-delete",
			recipe:      []string{"ssh-keygen -t ed25519"},
			shouldMatch: false,
		},
		// New rule tests: env-file-overwrite
		{
			name:        "env-file-overwrite matches redirect to .env",
			ruleID:      "env-file-overwrite",
			recipe:      []string{"echo 'KEY=value' > .env"},
			shouldMatch: true,
		},
		{
			name:        "env-file-overwrite matches cp to .env",
			ruleID:      "env-file-overwrite",
			recipe:      []string{"cp .env.example .env"},
			shouldMatch: true,
		},
		{
			name:        "env-file-overwrite safe cat .env does not match",
			ruleID:      "env-file-overwrite",
			recipe:      []string{"cat .env"},
			shouldMatch: false,
		},
		{
			name:        "iptables-flush matches nft flush",
			ruleID:      "iptables-flush",
			recipe:      []string{"nft flush ruleset"},
			shouldMatch: true,
		},
		// Version control tests
		{
			name:        "git-branch-delete-force matches git branch -D",
			ruleID:      "git-branch-delete-force",
			recipe:      []string{"git branch -D feature-branch"},
			shouldMatch: true,
		},
		{
			name:        "git-branch-delete-force safe -d does not match",
			ruleID:      "git-branch-delete-force",
			recipe:      []string{"git branch -d feature-branch"},
			shouldMatch: false,
		},
		{
			name:        "git-reflog-expire matches reflog expire",
			ruleID:      "git-reflog-expire",
			recipe:      []string{"git reflog expire --all --expire=now"},
			shouldMatch: true,
		},
		// Container orchestration tests
		{
			name:        "docker-swarm-destructive matches stack rm",
			ruleID:      "docker-swarm-destructive",
			recipe:      []string{"docker stack rm mystack"},
			shouldMatch: true,
		},
		{
			name:        "podman-system-reset matches podman system reset",
			ruleID:      "podman-system-reset",
			recipe:      []string{"podman system reset"},
			shouldMatch: true,
		},
		// Package manager tests
		{
			name:        "pip-uninstall-all matches pip uninstall -y",
			ruleID:      "pip-uninstall-all",
			recipe:      []string{"pip uninstall requests -y"},
			shouldMatch: true,
		},
		{
			name:        "pip-uninstall-all safe pip install does not match",
			ruleID:      "pip-uninstall-all",
			recipe:      []string{"pip install requests"},
			shouldMatch: false,
		},
		{
			name:        "go-clean-modcache matches go clean -modcache",
			ruleID:      "go-clean-modcache",
			recipe:      []string{"go clean -modcache"},
			shouldMatch: true,
		},
		// System service tests
		{
			name:        "systemctl-critical-services matches stopping nginx",
			ruleID:      "systemctl-critical-services",
			recipe:      []string{"systemctl stop nginx"},
			shouldMatch: true,
		},
		{
			name:        "systemctl-critical-services matches disabling docker",
			ruleID:      "systemctl-critical-services",
			recipe:      []string{"systemctl disable docker"},
			shouldMatch: true,
		},
		{
			name:        "systemctl-critical-services safe start does not match",
			ruleID:      "systemctl-critical-services",
			recipe:      []string{"systemctl start nginx"},
			shouldMatch: false,
		},
		{
			name:        "killall-force matches killall -9",
			ruleID:      "killall-force",
			recipe:      []string{"killall -9 node"},
			shouldMatch: true,
		},
		{
			name:        "killall-force safe killall does not match",
			ruleID:      "killall-force",
			recipe:      []string{"killall node"},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := GetBuiltinRuleByID(tt.ruleID)
			if rule == nil {
				t.Fatalf("Rule %s not found", tt.ruleID)
			}

			if err := rule.Compile(); err != nil {
				t.Fatalf("Failed to compile rule: %v", err)
			}

			matched, matchedLine := rule.Matches(tt.recipe)

			if matched != tt.shouldMatch {
				t.Errorf("Expected match=%v, got=%v for recipe %v",
					tt.shouldMatch, matched, tt.recipe)
			}

			if tt.shouldMatch && tt.expectedLine != "" && matchedLine != tt.expectedLine {
				t.Errorf("Expected matched line %q, got %q",
					tt.expectedLine, matchedLine)
			}
		})
	}
}

func TestContextAwareSeverityAdjustment(t *testing.T) {
	tests := []struct {
		name             string
		targetName       string
		matchedLine      string
		originalSeverity Severity
		expectedSeverity Severity
	}{
		{
			name:             "clean target downgrades critical to warning",
			targetName:       "clean",
			matchedLine:      "rm -rf build/",
			originalSeverity: SeverityCritical,
			expectedSeverity: SeverityWarning,
		},
		{
			name:             "clean target downgrades warning to info",
			targetName:       "distclean",
			matchedLine:      "docker system prune",
			originalSeverity: SeverityWarning,
			expectedSeverity: SeverityInfo,
		},
		{
			name:             "interactive flag downgrades critical",
			targetName:       "dangerous-op", // Use non-clean target name
			matchedLine:      "rm -rfi build/",
			originalSeverity: SeverityCritical,
			expectedSeverity: SeverityWarning,
		},
		{
			name:             "dev target without prod keywords downgrades critical",
			targetName:       "test-cleanup",
			matchedLine:      "terraform destroy",
			originalSeverity: SeverityCritical,
			expectedSeverity: SeverityWarning,
		},
		{
			name:             "production keyword does not upgrade warning",
			targetName:       "deploy",
			matchedLine:      "docker system prune --filter prod",
			originalSeverity: SeverityWarning,
			expectedSeverity: SeverityWarning, // No longer auto-escalates based on keywords
		},
		{
			name:             "normal target keeps critical",
			targetName:       "deploy-prod", // Use non-clean target name
			matchedLine:      "rm -rf /",
			originalSeverity: SeverityCritical,
			expectedSeverity: SeverityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := makefile.Target{
				Name: tt.targetName,
			}

			rule := Rule{
				Severity: tt.originalSeverity,
			}

			adjusted := adjustSeverity(target, rule, tt.matchedLine)

			if adjusted != tt.expectedSeverity {
				t.Errorf("Expected severity %v, got %v",
					tt.expectedSeverity, adjusted)
			}
		})
	}
}

func TestCheckerIntegration(t *testing.T) {
	// Create test targets
	targets := []makefile.Target{
		{
			Name:   "build",
			Recipe: []string{"go build -o app"},
		},
		{
			Name:   "clean",
			Recipe: []string{"rm -rf /tmp/build", "rm -f app"}, // System path gets flagged
		},
		{
			Name:   "nuke-prod",
			Recipe: []string{"psql -c 'DROP DATABASE production;'"},
		},
		{
			Name:   "safe-target",
			Recipe: []string{"echo 'Hello'"},
		},
	}

	config := DefaultConfig()
	checker, err := NewChecker(config)
	if err != nil {
		t.Fatalf("Failed to create checker: %v", err)
	}

	results := checker.CheckAllTargets(targets)

	// build should be safe
	if _, found := results["build"]; found {
		t.Error("build should not be flagged as dangerous")
	}

	// clean should be flagged (but severity downgraded)
	if result, found := results["clean"]; !found {
		t.Error("clean should be flagged")
	} else if result.DangerLevel != SeverityWarning {
		t.Errorf("clean should be warning, got %v", result.DangerLevel)
	}

	// nuke-prod should be critical
	if result, found := results["nuke-prod"]; !found {
		t.Error("nuke-prod should be flagged")
	} else if result.DangerLevel != SeverityCritical {
		t.Errorf("nuke-prod should be critical, got %v", result.DangerLevel)
	}

	// safe-target should be safe
	if _, found := results["safe-target"]; found {
		t.Error("safe-target should not be flagged")
	}
}

func TestExcludeTargets(t *testing.T) {
	targets := []makefile.Target{
		{
			Name:   "dangerous-but-excluded",
			Recipe: []string{"rm -rf /"},
		},
	}

	config := &Config{
		Enabled:        true,
		ExcludeTargets: []string{"dangerous-but-excluded"},
	}

	checker, err := NewChecker(config)
	if err != nil {
		t.Fatalf("Failed to create checker: %v", err)
	}

	results := checker.CheckAllTargets(targets)

	if _, found := results["dangerous-but-excluded"]; found {
		t.Error("Excluded target should not be checked")
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityCritical, "CRITICAL"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestIsCleanTarget(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"clean", true},
		{"distclean", true},
		{"purge", true},
		{"reset", true},
		{"nuke", true},
		{"build", false},
		{"test", false},
		{"cleanup-temp", true}, // contains "clean"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCleanTarget(tt.name); got != tt.expected {
				t.Errorf("isCleanTarget(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestIsDevelopmentTarget(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"dev", true},
		{"test", true},
		{"local", true},
		{"docker", true},
		{"demo", true},
		{"prod", false},
		{"deploy", false},
		{"test-prod", true}, // contains "test"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDevelopmentTarget(tt.name); got != tt.expected {
				t.Errorf("isDevelopmentTarget(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestContainsProductionKeywords(t *testing.T) {
	tests := []struct {
		command  string
		expected bool
	}{
		{"kubectl apply -f prod.yaml", true},
		{"terraform apply production", true},
		{"git push origin main", true},
		{"docker push myapp:latest", false},
		{"echo 'produce output'", false}, // "produce" != "prod"
		{"kubectl apply -f dev.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			if got := containsProductionKeywords(tt.command); got != tt.expected {
				t.Errorf("containsProductionKeywords(%q) = %v, want %v", tt.command, got, tt.expected)
			}
		})
	}
}

func TestHasInteractiveFlag(t *testing.T) {
	tests := []struct {
		command  string
		expected bool
	}{
		{"rm -i file.txt", true},
		{"rm -rfi build/", true},
		{"git add -i", true},
		{"docker rm --interactive container", true},
		{"rm -rf build/", false},
		{"git add .", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			if got := hasInteractiveFlag(tt.command); got != tt.expected {
				t.Errorf("hasInteractiveFlag(%q) = %v, want %v", tt.command, got, tt.expected)
			}
		})
	}
}
