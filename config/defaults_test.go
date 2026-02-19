package config

import (
	"testing"

	"github.com/rshelekhov/lazymake/internal/export"
	"github.com/rshelekhov/lazymake/internal/safety"
	"github.com/rshelekhov/lazymake/internal/shell"
)

// These tests ensure code defaults match documented values in docs/guides/configuration.md
// and .lazymake.example.yaml. If a test fails, update BOTH the docs and the code to match.

func TestExportDefaultsMatchDocumented(t *testing.T) {
	d := export.Defaults()

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"enabled", d.Enabled, false},
		{"format", d.Format, "json"},
		{"naming_strategy", d.NamingStrategy, "timestamp"},
		{"max_file_size_mb", d.MaxFileSize, int64(0)},
		{"max_files", d.MaxFiles, 0},
		{"keep_days", d.KeepDays, 0},
		{"success_only", d.SuccessOnly, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("export.%s default = %v, want %v — update docs if default changed", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestShellDefaultsMatchDocumented(t *testing.T) {
	d := shell.Defaults()

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"enabled", d.Enabled, false},
		{"shell", d.Shell, "auto"},
		{"history_file", d.HistoryFile, ""},
		{"include_timestamp", d.IncludeTimestamp, true},
		{"format_template", d.FormatTemplate, "make {target}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("shell.%s default = %v, want %v — update docs if default changed", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestSafetyDefaultsMatchDocumented(t *testing.T) {
	d := safety.DefaultConfig()

	if d.Enabled != true {
		t.Errorf("safety.enabled default = %v, want true", d.Enabled)
	}
	if d.EnabledRules != nil {
		t.Errorf("safety.enabled_rules default = %v, want nil (all rules)", d.EnabledRules)
	}
	if d.ExcludeTargets != nil {
		t.Errorf("safety.exclude_targets default = %v, want nil", d.ExcludeTargets)
	}
	if d.CustomRules != nil {
		t.Errorf("safety.custom_rules default = %v, want nil", d.CustomRules)
	}
}

func TestBuiltinSafetyRulesCount(t *testing.T) {
	count := len(safety.BuiltinRules)
	if count != 36 {
		t.Errorf("expected 36 built-in safety rules, got %d — update docs if rules were added/removed", count)
	}
}
