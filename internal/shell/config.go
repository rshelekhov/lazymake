package shell

// Config holds shell integration configuration options
type Config struct {
	// Master switch
	Enabled bool `yaml:"enabled"`

	// Shell selection: "auto", "bash", "zsh", or "none"
	Shell string `yaml:"shell"`

	// Override default history file path
	HistoryFile string `yaml:"history_file"`

	// Include timestamp in history entry (for zsh extended history)
	IncludeTimestamp bool `yaml:"include_timestamp"`

	// Custom format template for history entries
	// Available variables: {target}, {makefile}, {dir}
	FormatTemplate string `yaml:"format_template"`

	// Don't add these targets to shell history
	ExcludeTargets []string `yaml:"exclude_targets"`
}
