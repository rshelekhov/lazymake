package safety

import (
	"regexp"
	"strings"

	"github.com/rshelekhov/lazymake/internal/makefile"
)

// adjustSeverity adjusts rule severity based on target context
// Returns the adjusted severity level
func adjustSeverity(target makefile.Target, rule Rule, matchedLine string) Severity {
	severity := rule.Severity

	// Clean targets are expected to be destructive
	// BUT: don't downgrade if target affects databases or production systems
	if isCleanTarget(target.Name) && !affectsCriticalSystems(target.Name, matchedLine) {
		// Downgrade severity for clean targets
		switch severity {
		case SeverityCritical:
			severity = SeverityWarning
		case SeverityWarning:
			severity = SeverityInfo
		}
	}

	// Commands with interactive confirmation flags are safer
	if hasInteractiveFlag(matchedLine) {
		switch severity {
		case SeverityCritical:
			severity = SeverityWarning
		case SeverityWarning:
			severity = SeverityInfo
		}
	}

	// Development/test environments are less risky
	if isDevelopmentTarget(target.Name) && !containsProductionKeywords(matchedLine) {
		// Only downgrade if not explicitly targeting production
		if severity == SeverityCritical {
			severity = SeverityWarning
		}
	}

	// Note: Removed auto-escalation from Warning to Critical based on production keywords
	// Rules should be defined with appropriate base severity instead

	return severity
}

// isCleanTarget checks if target name suggests cleanup/destructive operations
func isCleanTarget(name string) bool {
	cleanKeywords := []string{
		"clean", "distclean", "purge", "reset", "nuke",
		"remove", "delete", "wipe", "clear",
	}
	nameLower := strings.ToLower(name)
	for _, keyword := range cleanKeywords {
		if strings.Contains(nameLower, keyword) {
			return true
		}
	}
	return false
}

// hasInteractiveFlag checks if command has interactive confirmation flag
func hasInteractiveFlag(command string) bool {
	// Common interactive flags: -i, --interactive
	// Examples: rm -i, git add -i, docker rm -i
	interactivePattern := regexp.MustCompile(`\s+-\w*i\w*(\s|$)|--interactive`)
	return interactivePattern.MatchString(command)
}

// isDevelopmentTarget checks if target name suggests development/testing
func isDevelopmentTarget(name string) bool {
	devKeywords := []string{
		"dev", "develop", "development",
		"test", "testing",
		"local", "localhost",
		"docker", "compose",
		"demo", "example", "sample",
	}
	nameLower := strings.ToLower(name)
	for _, keyword := range devKeywords {
		if strings.Contains(nameLower, keyword) {
			return true
		}
	}
	return false
}

// containsProductionKeywords checks if command targets production
func containsProductionKeywords(command string) bool {
	prodKeywords := []string{
		"prod", "production",
		"master", "main", // git branches
		"live",
		"release",
	}
	cmdLower := strings.ToLower(command)
	for _, keyword := range prodKeywords {
		// Match whole words to avoid false positives like "produce"
		pattern := regexp.MustCompile(`\b` + keyword + `\b`)
		if pattern.MatchString(cmdLower) {
			return true
		}
	}
	return false
}

// affectsCriticalSystems checks if target affects databases or production systems
// Used to prevent downgrading severity for "clean-like" targets that touch critical data
func affectsCriticalSystems(targetName, command string) bool {
	criticalKeywords := []string{
		"db", "database",
		"prod", "production",
		"schema", "migration",
		"backup", "restore",
	}

	combined := strings.ToLower(targetName + " " + command)

	for _, keyword := range criticalKeywords {
		if strings.Contains(combined, keyword) {
			return true
		}
	}
	return false
}
