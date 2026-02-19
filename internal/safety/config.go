package safety

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
