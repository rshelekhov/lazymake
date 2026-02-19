package config

import (
	"github.com/rshelekhov/lazymake/internal/export"
	"github.com/rshelekhov/lazymake/internal/safety"
	"github.com/rshelekhov/lazymake/internal/shell"
	"github.com/spf13/viper"
)

type Config struct {
	MakefilePath     string
	Export           *export.Config
	ShellIntegration *shell.Config
	Safety           *safety.Config
}

func Load() (*Config, error) {
	// Keep Viper config name/path for Cobra flag binding (--file flag).
	// The global Viper instance is still used by Cobra for CLI flags.
	viper.SetConfigName(".lazymake")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")

	// Read config so Cobra flag defaults work (ignore missing file)
	_ = viper.ReadInConfig()

	// Load global and project config files independently
	globalViper := loadViperFromFile(globalConfigPath())
	projectViper := loadViperFromFile(projectConfigPath())

	// Read each section from both files
	globalExport, globalExportSet := readExportConfig(globalViper)
	projectExport, projectExportSet := readExportConfig(projectViper)

	globalShell, globalShellSet := readShellConfig(globalViper)
	projectShell, projectShellSet := readShellConfig(projectViper)

	globalSafety, globalSafetySet := readSafetyConfig(globalViper)
	projectSafety, projectSafetySet := readSafetyConfig(projectViper)

	// Merge each section
	mergedExport := mergeExportConfigs(globalExport, projectExport, globalExportSet, projectExportSet)
	mergedShell := mergeShellConfigs(globalShell, projectShell, globalShellSet, projectShellSet)
	mergedSafety := mergeSafetyConfigs(globalSafety, projectSafety, globalSafetySet, projectSafetySet)

	cfg := &Config{
		Export:           mergedExport,
		ShellIntegration: mergedShell,
		Safety:           mergedSafety,
	}

	// CLI flag override for makefile path
	if viper.IsSet("makefile") {
		cfg.MakefilePath = viper.GetString("makefile")
	}

	return cfg, nil
}
