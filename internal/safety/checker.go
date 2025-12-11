package safety

import (
	"log"

	"github.com/rshelekhov/lazymake/internal/makefile"
)

// Checker performs safety checks on Makefile targets
type Checker struct {
	rules  []Rule
	config *Config
}

// NewChecker creates a new safety checker with the given configuration
func NewChecker(config *Config) (*Checker, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Collect and compile all rules
	rules := collectRules(config)

	return &Checker{
		rules:  rules,
		config: config,
	}, nil
}

// collectRules gathers all rules from config (built-in + custom)
// Only includes rules that compile successfully
func collectRules(config *Config) []Rule {
	var rules []Rule

	// Add built-in rules if enabled
	if config.Enabled {
		for _, rule := range BuiltinRules {
			// Check if this rule is enabled
			if len(config.EnabledRules) == 0 || contains(config.EnabledRules, rule.ID) {
				// Test compile before adding
				if err := rule.Compile(); err != nil {
					log.Printf("Warning: skipping invalid built-in rule %s: %v", rule.ID, err)
					continue
				}
				rules = append(rules, rule)
			}
		}
	}

	// Add custom rules
	for _, rule := range config.CustomRules {
		// Test compile before adding
		if err := rule.Compile(); err != nil {
			log.Printf("Warning: skipping invalid custom rule %s: %v", rule.ID, err)
			continue
		}
		rules = append(rules, rule)
	}

	return rules
}

// CheckTarget performs safety check on a single target
// Returns nil if target is safe or excluded
func (c *Checker) CheckTarget(target makefile.Target) *SafetyCheckResult {
	// Skip if safety checks disabled
	if !c.config.Enabled {
		return nil
	}

	// Skip if target is in exclusion list
	if contains(c.config.ExcludeTargets, target.Name) {
		return nil
	}

	// Skip if target has no recipe (meta-targets, phony targets with only deps)
	if len(target.Recipe) == 0 {
		return nil
	}

	var matches []MatchResult
	highestSeverity := SeverityInfo

	// Check each rule against target's recipe
	for _, rule := range c.rules {
		if matched, matchedLine := rule.Matches(target.Recipe); matched {
			// Adjust severity based on context
			adjustedSeverity := adjustSeverity(target, rule, matchedLine)

			match := MatchResult{
				Target:      target.Name,
				Rule:        rule,
				MatchedLine: matchedLine,
				Severity:    adjustedSeverity,
			}
			matches = append(matches, match)

			// Track highest severity
			if adjustedSeverity > highestSeverity {
				highestSeverity = adjustedSeverity
			}
		}
	}

	// Return nil if no matches
	if len(matches) == 0 {
		return nil
	}

	return &SafetyCheckResult{
		TargetName:  target.Name,
		IsDangerous: true,
		DangerLevel: highestSeverity,
		Matches:     matches,
	}
}

// CheckAllTargets performs safety check on all targets
// Returns map of target name -> result (only includes dangerous targets)
func (c *Checker) CheckAllTargets(targets []makefile.Target) map[string]*SafetyCheckResult {
	results := make(map[string]*SafetyCheckResult)

	for _, target := range targets {
		if result := c.CheckTarget(target); result != nil {
			results[target.Name] = result
		}
	}

	return results
}

// contains checks if slice contains string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
