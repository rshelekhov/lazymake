package config

import "github.com/spf13/viper"

type Config struct {
	MakefilePath string
	Theme        string
}

func Load() (*Config, error) {
	// Set defaults
	viper.SetDefault("makefile", "Makefile")
	viper.SetDefault("theme", "default")

	// Read config file
	viper.SetConfigName(".lazymake")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")

	// Ignore error if config file doesn't exist
	_ = viper.ReadInConfig()

	return &Config{
		MakefilePath: viper.GetString("makefile"),
		Theme:        viper.GetString("theme"),
	}, nil
}
