package safety

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents safety checker configuration
type Config struct {
	Enabled        bool     // Master switch for safety checks
	EnabledRules   []string // Which built-in rules to enable (empty = all)
	ExcludeTargets []string // Targets to skip checking
	CustomRules    []Rule   // User-defined rules
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:        true, // Enabled by default
		EnabledRules:   nil,  // nil = all built-in rules enabled
		ExcludeTargets: nil,
		CustomRules:    nil,
	}
}

// LoadConfig loads safety configuration from global and project configs
// Merges them according to the merge strategy
func LoadConfig() (*Config, error) {
	globalConfig := loadGlobalConfig()
	projectConfig := loadProjectConfig()

	merged := mergeConfigs(globalConfig, projectConfig)
	return merged, nil
}

// loadGlobalConfig loads from ~/.lazymake.yaml
func loadGlobalConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfig()
	}

	globalPath := filepath.Join(homeDir, ".lazymake.yaml")
	return loadConfigFromFile(globalPath)
}

// loadProjectConfig loads from ./.lazymake.yaml
func loadProjectConfig() *Config {
	return loadConfigFromFile(".lazymake.yaml")
}

// loadConfigFromFile loads configuration from a specific file
func loadConfigFromFile(path string) *Config {
	v := viper.New()
	v.SetConfigFile(path)

	// Ignore error if file doesn't exist
	if err := v.ReadInConfig(); err != nil {
		return DefaultConfig()
	}

	config := DefaultConfig()

	// Read safety configuration
	if v.IsSet("safety.enabled") {
		config.Enabled = v.GetBool("safety.enabled")
	}

	if v.IsSet("safety.enabled_rules") {
		config.EnabledRules = v.GetStringSlice("safety.enabled_rules")
	}

	if v.IsSet("safety.exclude_targets") {
		config.ExcludeTargets = v.GetStringSlice("safety.exclude_targets")
	}

	// Load custom rules
	if v.IsSet("safety.custom_rules") {
		var customRules []map[string]interface{}
		if err := v.UnmarshalKey("safety.custom_rules", &customRules); err == nil {
			config.CustomRules = parseCustomRules(customRules)
		}
	}

	return config
}

// parseCustomRules converts YAML map to Rule structs
func parseCustomRules(rulesMaps []map[string]interface{}) []Rule {
	var rules []Rule

	for _, ruleMap := range rulesMaps {
		rule := Rule{
			ID:          getString(ruleMap, "id"),
			Description: getString(ruleMap, "description"),
			Suggestion:  getString(ruleMap, "suggestion"),
			Patterns:    getStringSlice(ruleMap, "patterns"),
		}

		// Parse severity
		severityStr := getString(ruleMap, "severity")
		switch severityStr {
		case "critical":
			rule.Severity = SeverityCritical
		case "warning":
			rule.Severity = SeverityWarning
		case "info":
			rule.Severity = SeverityInfo
		default:
			rule.Severity = SeverityWarning // Default to warning
		}

		rules = append(rules, rule)
	}

	return rules
}

// mergeConfigs merges global and project configs
// Merge strategy:
//   - enabled: project overrides global
//   - enabled_rules: union (both sets apply)
//   - exclude_targets: union (both sets apply)
//   - custom_rules: union (both sets apply)
func mergeConfigs(global, project *Config) *Config {
	result := &Config{}

	// Project overrides global for enabled flag
	// If both have defaults (true), use project's value
	switch {
	case project != nil:
		result.Enabled = project.Enabled
	case global != nil:
		result.Enabled = global.Enabled
	default:
		result.Enabled = true // Default
	}

	// Union of enabled rules
	ruleSet := make(map[string]bool)
	if global != nil {
		for _, rule := range global.EnabledRules {
			ruleSet[rule] = true
		}
	}
	if project != nil {
		for _, rule := range project.EnabledRules {
			ruleSet[rule] = true
		}
	}
	for rule := range ruleSet {
		result.EnabledRules = append(result.EnabledRules, rule)
	}

	// Union of exclusions
	if global != nil {
		result.ExcludeTargets = append(result.ExcludeTargets, global.ExcludeTargets...)
	}
	if project != nil {
		result.ExcludeTargets = append(result.ExcludeTargets, project.ExcludeTargets...)
	}

	// Union of custom rules
	if global != nil {
		result.CustomRules = append(result.CustomRules, global.CustomRules...)
	}
	if project != nil {
		result.CustomRules = append(result.CustomRules, project.CustomRules...)
	}

	return result
}

// Helper functions for parsing YAML maps

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if val, ok := m[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, len(slice))
			for i, item := range slice {
				if str, ok := item.(string); ok {
					result[i] = str
				}
			}
			return result
		}
	}
	return nil
}
