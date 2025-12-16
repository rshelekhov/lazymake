package export

import (
	"fmt"
)

// Exporter handles exporting execution results to files
type Exporter struct {
	config *Config
}

// NewExporter creates a new Exporter with the given configuration
func NewExporter(config *Config) (*Exporter, error) {
	if config == nil {
		config = Defaults()
	}

	// Validate format
	if config.Format != "json" && config.Format != "log" && config.Format != "both" {
		return nil, fmt.Errorf("invalid format: %s (must be 'json', 'log', or 'both')", config.Format)
	}

	// Validate naming strategy
	if config.NamingStrategy != "timestamp" && config.NamingStrategy != "target" && config.NamingStrategy != "sequential" {
		return nil, fmt.Errorf("invalid naming_strategy: %s (must be 'timestamp', 'target', or 'sequential')", config.NamingStrategy)
	}

	return &Exporter{
		config: config,
	}, nil
}

// Export exports an execution record to file(s) based on configuration
func (e *Exporter) Export(record *ExecutionRecord) error {
	if e.config == nil || !e.config.Enabled {
		return nil // Export disabled
	}

	// Check if target is excluded
	for _, excluded := range e.config.ExcludeTargets {
		if record.TargetName == excluded {
			return nil // Target excluded from export
		}
	}

	// Check if we should only export successful executions
	if e.config.SuccessOnly && !record.Success {
		return nil // Skip failed executions
	}

	// Export based on format
	var lastErr error

	if e.config.Format == "json" || e.config.Format == "both" {
		if err := e.exportJSON(record); err != nil {
			lastErr = err
		}
	}

	if e.config.Format == "log" || e.config.Format == "both" {
		if err := e.exportLog(record); err != nil {
			lastErr = err
		}
	}

	// Perform rotation after export
	if lastErr == nil {
		_ = RotateFiles(e.config.OutputDir, record.TargetName, e.config)
	}

	return lastErr
}

// exportJSON exports the record as JSON
func (e *Exporter) exportJSON(record *ExecutionRecord) error {
	filename := record.GenerateFilename(e.config.NamingStrategy, "json")
	path := GenerateExportPath(e.config.OutputDir, filename)

	return WriteJSON(record, path)
}

// exportLog exports the record as a plain text log
func (e *Exporter) exportLog(record *ExecutionRecord) error {
	filename := record.GenerateFilename(e.config.NamingStrategy, "log")
	path := GenerateExportPath(e.config.OutputDir, filename)

	return WriteLog(record, path)
}
