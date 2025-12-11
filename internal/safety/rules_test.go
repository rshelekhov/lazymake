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
			name:             "production keyword upgrades warning",
			targetName:       "deploy",
			matchedLine:      "docker system prune --filter prod",
			originalSeverity: SeverityWarning,
			expectedSeverity: SeverityCritical,
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
