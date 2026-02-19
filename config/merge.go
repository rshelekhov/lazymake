package config

import (
	"os"
	"path/filepath"

	"github.com/rshelekhov/lazymake/internal/export"
	"github.com/rshelekhov/lazymake/internal/safety"
	"github.com/rshelekhov/lazymake/internal/shell"
	"github.com/spf13/viper"
)

// fieldSet tracks which config keys were explicitly set in a file.
type fieldSet map[string]bool

// loadViperFromFile creates a fresh Viper instance and reads the given YAML file.
// Returns nil if the file does not exist or cannot be read.
func loadViperFromFile(path string) *viper.Viper {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil
	}

	return v
}

// globalConfigPath returns the path to the global config file (~/.lazymake.yaml).
func globalConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(homeDir, ".lazymake.yaml")
}

// projectConfigPath returns the path to the project config file (./.lazymake.yaml).
func projectConfigPath() string {
	return ".lazymake.yaml"
}

// mergeStringSliceUnion returns a deduplicated union of two string slices.
func mergeStringSliceUnion(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	result := make([]string, 0, len(a)+len(b))

	for _, s := range a {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}

// readExportConfig reads the export section from a Viper instance.
// Returns the config and a fieldSet of explicitly set keys.
func readExportConfig(v *viper.Viper) (*export.Config, fieldSet) {
	if v == nil {
		return export.Defaults(), nil
	}

	cfg := export.Defaults()
	set := make(fieldSet)

	if v.IsSet("export.enabled") {
		cfg.Enabled = v.GetBool("export.enabled")
		set["enabled"] = true
	}

	if v.IsSet("export.output_dir") {
		cfg.OutputDir = v.GetString("export.output_dir")
		set["output_dir"] = true
	}

	if v.IsSet("export.format") {
		cfg.Format = v.GetString("export.format")
		set["format"] = true
	}

	if v.IsSet("export.naming_strategy") {
		cfg.NamingStrategy = v.GetString("export.naming_strategy")
		set["naming_strategy"] = true
	}

	if v.IsSet("export.max_file_size_mb") {
		cfg.MaxFileSize = v.GetInt64("export.max_file_size_mb")
		set["max_file_size_mb"] = true
	}

	if v.IsSet("export.max_files") {
		cfg.MaxFiles = v.GetInt("export.max_files")
		set["max_files"] = true
	}

	if v.IsSet("export.keep_days") {
		cfg.KeepDays = v.GetInt("export.keep_days")
		set["keep_days"] = true
	}

	if v.IsSet("export.success_only") {
		cfg.SuccessOnly = v.GetBool("export.success_only")
		set["success_only"] = true
	}

	if v.IsSet("export.exclude_targets") {
		cfg.ExcludeTargets = v.GetStringSlice("export.exclude_targets")
		set["exclude_targets"] = true
	}

	return cfg, set
}

// readShellConfig reads the shell_integration section from a Viper instance.
// Returns the config and a fieldSet of explicitly set keys.
func readShellConfig(v *viper.Viper) (*shell.Config, fieldSet) {
	if v == nil {
		return shell.Defaults(), nil
	}

	cfg := shell.Defaults()
	set := make(fieldSet)

	if v.IsSet("shell_integration.enabled") {
		cfg.Enabled = v.GetBool("shell_integration.enabled")
		set["enabled"] = true
	}

	if v.IsSet("shell_integration.shell") {
		cfg.Shell = v.GetString("shell_integration.shell")
		set["shell"] = true
	}

	if v.IsSet("shell_integration.history_file") {
		cfg.HistoryFile = v.GetString("shell_integration.history_file")
		set["history_file"] = true
	}

	if v.IsSet("shell_integration.include_timestamp") {
		cfg.IncludeTimestamp = v.GetBool("shell_integration.include_timestamp")
		set["include_timestamp"] = true
	}

	if v.IsSet("shell_integration.format_template") {
		cfg.FormatTemplate = v.GetString("shell_integration.format_template")
		set["format_template"] = true
	}

	if v.IsSet("shell_integration.exclude_targets") {
		cfg.ExcludeTargets = v.GetStringSlice("shell_integration.exclude_targets")
		set["exclude_targets"] = true
	}

	return cfg, set
}

// readSafetyConfig reads the safety section from a Viper instance.
// Returns the config and a fieldSet of explicitly set keys.
func readSafetyConfig(v *viper.Viper) (*safety.Config, fieldSet) {
	if v == nil {
		return safety.DefaultConfig(), nil
	}

	cfg := safety.DefaultConfig()
	set := make(fieldSet)

	if v.IsSet("safety.enabled") {
		cfg.Enabled = v.GetBool("safety.enabled")
		set["enabled"] = true
	}

	if v.IsSet("safety.enabled_rules") {
		cfg.EnabledRules = v.GetStringSlice("safety.enabled_rules")
		set["enabled_rules"] = true
	}

	if v.IsSet("safety.exclude_targets") {
		cfg.ExcludeTargets = v.GetStringSlice("safety.exclude_targets")
		set["exclude_targets"] = true
	}

	if v.IsSet("safety.custom_rules") {
		var customRules []map[string]interface{}
		if err := v.UnmarshalKey("safety.custom_rules", &customRules); err == nil {
			cfg.CustomRules = parseCustomRules(customRules)
		}
		set["custom_rules"] = true
	}

	return cfg, set
}

// mergeExportConfigs merges global and project export configurations.
// Scalars: project overrides global. Slices: union, deduplicated.
func mergeExportConfigs(global, project *export.Config, globalSet, projectSet fieldSet) *export.Config {
	result := export.Defaults()

	applyBool := func(field *bool, key string, gVal, pVal bool) {
		if projectSet[key] {
			*field = pVal
		} else if globalSet[key] {
			*field = gVal
		}
	}

	applyString := func(field *string, key string, gVal, pVal string) {
		if projectSet[key] {
			*field = pVal
		} else if globalSet[key] {
			*field = gVal
		}
	}

	applyInt := func(field *int, key string, gVal, pVal int) {
		if projectSet[key] {
			*field = pVal
		} else if globalSet[key] {
			*field = gVal
		}
	}

	applyInt64 := func(field *int64, key string, gVal, pVal int64) {
		if projectSet[key] {
			*field = pVal
		} else if globalSet[key] {
			*field = gVal
		}
	}

	applyBool(&result.Enabled, "enabled", global.Enabled, project.Enabled)
	applyString(&result.OutputDir, "output_dir", global.OutputDir, project.OutputDir)
	applyString(&result.Format, "format", global.Format, project.Format)
	applyString(&result.NamingStrategy, "naming_strategy", global.NamingStrategy, project.NamingStrategy)
	applyInt64(&result.MaxFileSize, "max_file_size_mb", global.MaxFileSize, project.MaxFileSize)
	applyInt(&result.MaxFiles, "max_files", global.MaxFiles, project.MaxFiles)
	applyInt(&result.KeepDays, "keep_days", global.KeepDays, project.KeepDays)
	applyBool(&result.SuccessOnly, "success_only", global.SuccessOnly, project.SuccessOnly)

	// Slices: union, deduplicated
	result.ExcludeTargets = mergeStringSliceUnion(global.ExcludeTargets, project.ExcludeTargets)

	return result
}

// mergeShellConfigs merges global and project shell_integration configurations.
// Scalars: project overrides global. Slices: union, deduplicated.
func mergeShellConfigs(global, project *shell.Config, globalSet, projectSet fieldSet) *shell.Config {
	result := shell.Defaults()

	applyBool := func(field *bool, key string, gVal, pVal bool) {
		if projectSet[key] {
			*field = pVal
		} else if globalSet[key] {
			*field = gVal
		}
	}

	applyString := func(field *string, key string, gVal, pVal string) {
		if projectSet[key] {
			*field = pVal
		} else if globalSet[key] {
			*field = gVal
		}
	}

	applyBool(&result.Enabled, "enabled", global.Enabled, project.Enabled)
	applyString(&result.Shell, "shell", global.Shell, project.Shell)
	applyString(&result.HistoryFile, "history_file", global.HistoryFile, project.HistoryFile)
	applyBool(&result.IncludeTimestamp, "include_timestamp", global.IncludeTimestamp, project.IncludeTimestamp)
	applyString(&result.FormatTemplate, "format_template", global.FormatTemplate, project.FormatTemplate)

	// Slices: union, deduplicated
	result.ExcludeTargets = mergeStringSliceUnion(global.ExcludeTargets, project.ExcludeTargets)

	return result
}

// mergeSafetyConfigs merges global and project safety configurations.
// Scalars: project overrides global. Slices: union (deduplicated for strings, appended for structs).
func mergeSafetyConfigs(global, project *safety.Config, globalSet, projectSet fieldSet) *safety.Config {
	result := safety.DefaultConfig()

	// Enabled: project overrides global
	if projectSet["enabled"] {
		result.Enabled = project.Enabled
	} else if globalSet["enabled"] {
		result.Enabled = global.Enabled
	}

	// EnabledRules: union, deduplicated
	result.EnabledRules = mergeStringSliceUnion(global.EnabledRules, project.EnabledRules)
	if len(result.EnabledRules) == 0 {
		result.EnabledRules = nil
	}

	// ExcludeTargets: union, deduplicated
	result.ExcludeTargets = mergeStringSliceUnion(global.ExcludeTargets, project.ExcludeTargets)
	if len(result.ExcludeTargets) == 0 {
		result.ExcludeTargets = nil
	}

	// CustomRules: append (struct slice)
	result.CustomRules = append(result.CustomRules, global.CustomRules...)
	result.CustomRules = append(result.CustomRules, project.CustomRules...)
	if len(result.CustomRules) == 0 {
		result.CustomRules = nil
	}

	return result
}

// parseCustomRules converts YAML map to safety.Rule structs.
func parseCustomRules(rulesMaps []map[string]interface{}) []safety.Rule {
	var rules []safety.Rule

	for _, ruleMap := range rulesMaps {
		rule := safety.Rule{
			ID:          getString(ruleMap, "id"),
			Description: getString(ruleMap, "description"),
			Suggestion:  getString(ruleMap, "suggestion"),
			Patterns:    getStringSlice(ruleMap, "patterns"),
		}

		severityStr := getString(ruleMap, "severity")
		switch severityStr {
		case "critical":
			rule.Severity = safety.SeverityCritical
		case "warning":
			rule.Severity = safety.SeverityWarning
		case "info":
			rule.Severity = safety.SeverityInfo
		default:
			rule.Severity = safety.SeverityWarning
		}

		rules = append(rules, rule)
	}

	return rules
}

// getString extracts a string value from a map.
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getStringSlice extracts a string slice from a map.
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
