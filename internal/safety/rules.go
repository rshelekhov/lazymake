package safety

import (
	"fmt"
	"regexp"
)

// Severity represents the danger level of a matched rule
type Severity int

const (
	SeverityInfo Severity = iota // No UI indication (logged only)
	SeverityWarning                  // Show ⚠️ emoji
	SeverityCritical                 // Show 🚨 emoji + confirmation dialog
)

// String returns human-readable severity name
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Rule defines a pattern to match dangerous commands
type Rule struct {
	ID          string   // Unique identifier: "rm-rf-root", "docker-prune"
	Severity    Severity // Critical, Warning, Info
	Patterns    []string // Regex patterns to match in recipe commands
	Description string   // User-friendly explanation of the danger
	Suggestion  string   // Optional: safer alternative command

	// Compiled patterns (cached for performance)
	compiledPatterns []*regexp.Regexp
}

// Compile compiles the regex patterns for this rule
// Returns error if any pattern is invalid
func (r *Rule) Compile() error {
	r.compiledPatterns = make([]*regexp.Regexp, len(r.Patterns))
	for i, pattern := range r.Patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("rule %s: invalid pattern %q: %w", r.ID, pattern, err)
		}
		r.compiledPatterns[i] = re
	}
	return nil
}

// Matches checks if any recipe line matches this rule
// Returns: (matched bool, matched line string)
func (r *Rule) Matches(recipeLines []string) (bool, string) {
	for _, line := range recipeLines {
		for _, re := range r.compiledPatterns {
			if re.MatchString(line) {
				return true, line
			}
		}
	}
	return false, ""
}

// MatchResult represents a rule match for a specific target
type MatchResult struct {
	Target      string   // Target name that matched
	Rule        Rule     // Matched rule
	MatchedLine string   // Specific command line that triggered the rule
	Severity    Severity // Final severity (may be adjusted by context)
}

// SafetyCheckResult represents the complete safety check for a target
type SafetyCheckResult struct {
	TargetName  string
	IsDangerous bool         // Has any matched safety rules
	DangerLevel Severity     // Highest severity level
	Matches     []MatchResult // All matched rules for this target
}
