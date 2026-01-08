package config

import (
	"github.com/rshelekhov/lazymake/internal/export"
	"github.com/rshelekhov/lazymake/internal/shell"
	"github.com/spf13/viper"
)

type Config struct {
	MakefilePath     string
	Export           *export.Config
	ShellIntegration *shell.Config
}

func Load() (*Config, error) {
	// Set defaults
	viper.SetDefault("makefile", "")

	// Export defaults
	viper.SetDefault("export.enabled", false)
	viper.SetDefault("export.output_dir", "")
	viper.SetDefault("export.format", "json")
	viper.SetDefault("export.naming_strategy", "timestamp")
	viper.SetDefault("export.max_file_size_mb", 0)
	viper.SetDefault("export.max_files", 0)
	viper.SetDefault("export.keep_days", 0)
	viper.SetDefault("export.success_only", false)

	// Shell integration defaults
	viper.SetDefault("shell_integration.enabled", false)
	viper.SetDefault("shell_integration.shell", "auto")
	viper.SetDefault("shell_integration.history_file", "")
	viper.SetDefault("shell_integration.include_timestamp", true)
	viper.SetDefault("shell_integration.format_template", "make {target}")

	// Read config file
	viper.SetConfigName(".lazymake")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")

	// Ignore error if config file doesn't exist
	_ = viper.ReadInConfig()

	// Build export config
	exportConfig := &export.Config{
		Enabled:        viper.GetBool("export.enabled"),
		OutputDir:      viper.GetString("export.output_dir"),
		Format:         viper.GetString("export.format"),
		NamingStrategy: viper.GetString("export.naming_strategy"),
		MaxFileSize:    viper.GetInt64("export.max_file_size_mb"),
		MaxFiles:       viper.GetInt("export.max_files"),
		KeepDays:       viper.GetInt("export.keep_days"),
		SuccessOnly:    viper.GetBool("export.success_only"),
		ExcludeTargets: viper.GetStringSlice("export.exclude_targets"),
	}

	// Build shell integration config
	shellConfig := &shell.Config{
		Enabled:          viper.GetBool("shell_integration.enabled"),
		Shell:            viper.GetString("shell_integration.shell"),
		HistoryFile:      viper.GetString("shell_integration.history_file"),
		IncludeTimestamp: viper.GetBool("shell_integration.include_timestamp"),
		FormatTemplate:   viper.GetString("shell_integration.format_template"),
		ExcludeTargets:   viper.GetStringSlice("shell_integration.exclude_targets"),
	}

	return &Config{
		MakefilePath:     viper.GetString("makefile"),
		Export:           exportConfig,
		ShellIntegration: shellConfig,
	}, nil
}
