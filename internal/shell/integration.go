package shell

import (
	"path/filepath"
	"strings"
)

// ExecutionInfo holds the data needed to record a shell history entry.
type ExecutionInfo struct {
	Target       string
	MakefilePath string
}

// HistoryWriter interface for writing to shell history
type HistoryWriter interface {
	Append(entry string) error
}

// Integration manages shell history integration
type Integration struct {
	config *Config
	writer HistoryWriter
}

// NewIntegration creates a new shell integration instance
func NewIntegration(config *Config) (*Integration, error) {
	if config == nil || !config.Enabled {
		return nil, nil // Disabled
	}

	// Determine shell type
	shellType := config.Shell
	if shellType == "auto" {
		shellType = DetectShell()
	}

	// Get history file path
	historyFile := config.HistoryFile
	if historyFile == "" {
		var err error
		historyFile, err = GetHistoryFile(shellType)
		if err != nil {
			return nil, nil // Can't determine history file, disable gracefully
		}
	}

	// Create appropriate writer
	var writer HistoryWriter
	switch shellType {
	case "bash":
		writer = NewBashWriter(historyFile)
	case "zsh":
		writer = NewZshWriter(historyFile, config.IncludeTimestamp)
	case "fish":
		writer = NewFishWriter(historyFile, config.IncludeTimestamp)
	default:
		return nil, nil // Unsupported shell, disable gracefully
	}

	return &Integration{
		config: config,
		writer: writer,
	}, nil
}

// RecordExecution records a target execution in shell history
func (i *Integration) RecordExecution(info ExecutionInfo) error {
	if i == nil || i.writer == nil {
		return nil // Disabled
	}

	// Check if target is excluded
	for _, excluded := range i.config.ExcludeTargets {
		if info.Target == excluded {
			return nil
		}
	}

	// Format entry using template
	entry := formatEntry(i.config.FormatTemplate, info)

	// Write to history file
	return i.writer.Append(entry)
}

// formatEntry formats a history entry using the template
func formatEntry(template string, info ExecutionInfo) string {
	if template == "" {
		template = "make {target}"
	}

	entry := strings.ReplaceAll(template, "{target}", info.Target)
	entry = strings.ReplaceAll(entry, "{makefile}", info.MakefilePath)
	entry = strings.ReplaceAll(entry, "{dir}", filepath.Dir(info.MakefilePath))

	return entry
}
