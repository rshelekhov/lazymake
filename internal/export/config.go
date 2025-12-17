package export

// Config holds export configuration options
type Config struct {
	// Master switch
	Enabled bool `yaml:"enabled"`

	// Output directory (supports ~ expansion and environment variables)
	OutputDir string `yaml:"output_dir"`

	// Format options: "json", "log", or "both"
	Format string `yaml:"format"`

	// File naming strategy: "timestamp", "target", or "sequential"
	NamingStrategy string `yaml:"naming_strategy"`

	// Rotation settings (0 = unlimited/disabled)
	MaxFileSize int64 `yaml:"max_file_size_mb"` // In MB
	MaxFiles    int   `yaml:"max_files"`        // Max files per target
	KeepDays    int   `yaml:"keep_days"`        // Days to keep files

	// Filtering
	SuccessOnly    bool     `yaml:"success_only"`    // Only export successful executions
	ExcludeTargets []string `yaml:"exclude_targets"` // Don't export these targets
}

// Defaults returns a Config with sensible default values
func Defaults() *Config {
	return &Config{
		Enabled:        false,
		OutputDir:      "", // Will be set to ~/.cache/lazymake/exports in NewExporter
		Format:         "json",
		NamingStrategy: "timestamp",
		MaxFileSize:    0, // Unlimited
		MaxFiles:       0, // Unlimited
		KeepDays:       0, // Forever
		SuccessOnly:    false,
		ExcludeTargets: []string{},
	}
}
