package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rshelekhov/lazymake/internal/export"
	"github.com/rshelekhov/lazymake/internal/safety"
	"github.com/rshelekhov/lazymake/internal/shell"
	"github.com/spf13/viper"
)

// writeYAML is a test helper that writes content to a YAML file.
func writeYAML(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func TestMergeStringSliceUnion(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name:     "both non-empty, no overlap",
			a:        []string{"watch", "dev"},
			b:        []string{"ci", "deploy"},
			expected: []string{"watch", "dev", "ci", "deploy"},
		},
		{
			name:     "overlapping entries deduplicated",
			a:        []string{"watch", "dev"},
			b:        []string{"dev", "ci"},
			expected: []string{"watch", "dev", "ci"},
		},
		{
			name:     "first slice empty",
			a:        []string{},
			b:        []string{"ci"},
			expected: []string{"ci"},
		},
		{
			name:     "second slice empty",
			a:        []string{"watch"},
			b:        []string{},
			expected: []string{"watch"},
		},
		{
			name:     "both empty",
			a:        []string{},
			b:        []string{},
			expected: []string{},
		},
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: []string{},
		},
		{
			name:     "first nil second non-empty",
			a:        nil,
			b:        []string{"ci"},
			expected: []string{"ci"},
		},
		{
			name:     "all duplicates",
			a:        []string{"watch", "dev"},
			b:        []string{"watch", "dev"},
			expected: []string{"watch", "dev"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeStringSliceUnion(tt.a, tt.b)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected len %d, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, v := range tt.expected {
				if result[i] != v {
					t.Errorf("index %d: expected %q, got %q", i, v, result[i])
				}
			}
		})
	}
}

func TestMergeExportConfigs(t *testing.T) {
	tests := []struct {
		name       string
		global     *export.Config
		project    *export.Config
		globalSet  fieldSet
		projectSet fieldSet
		check      func(t *testing.T, result *export.Config)
	}{
		{
			name:    "disjoint fields — all values preserved",
			global:  &export.Config{Enabled: true, Format: "json", ExcludeTargets: []string{"watch"}},
			project: &export.Config{OutputDir: "/tmp/exports", MaxFiles: 10, ExcludeTargets: []string{"ci"}},
			globalSet: fieldSet{
				"enabled": true, "format": true, "exclude_targets": true,
			},
			projectSet: fieldSet{
				"output_dir": true, "max_files": true, "exclude_targets": true,
			},
			check: func(t *testing.T, r *export.Config) {
				if !r.Enabled {
					t.Error("expected enabled=true from global")
				}
				if r.Format != "json" {
					t.Errorf("expected format=json, got %s", r.Format)
				}
				if r.OutputDir != "/tmp/exports" {
					t.Errorf("expected output_dir=/tmp/exports, got %s", r.OutputDir)
				}
				if r.MaxFiles != 10 {
					t.Errorf("expected max_files=10, got %d", r.MaxFiles)
				}
				assertSliceEqual(t, r.ExcludeTargets, []string{"watch", "ci"})
			},
		},
		{
			name:    "overlapping scalars — project wins",
			global:  &export.Config{Format: "log", MaxFiles: 50},
			project: &export.Config{Format: "json", MaxFiles: 20},
			globalSet: fieldSet{
				"format": true, "max_files": true,
			},
			projectSet: fieldSet{
				"format": true, "max_files": true,
			},
			check: func(t *testing.T, r *export.Config) {
				if r.Format != "json" {
					t.Errorf("expected format=json (project wins), got %s", r.Format)
				}
				if r.MaxFiles != 20 {
					t.Errorf("expected max_files=20 (project wins), got %d", r.MaxFiles)
				}
			},
		},
		{
			name:    "overlapping exclude_targets — union deduplicated",
			global:  &export.Config{ExcludeTargets: []string{"watch", "dev"}},
			project: &export.Config{ExcludeTargets: []string{"dev", "ci"}},
			globalSet: fieldSet{
				"exclude_targets": true,
			},
			projectSet: fieldSet{
				"exclude_targets": true,
			},
			check: func(t *testing.T, r *export.Config) {
				assertSliceEqual(t, r.ExcludeTargets, []string{"watch", "dev", "ci"})
			},
		},
		{
			name:       "only global — global values used",
			global:     &export.Config{Enabled: true, Format: "log", MaxFiles: 50},
			project:    export.Defaults(),
			globalSet:  fieldSet{"enabled": true, "format": true, "max_files": true},
			projectSet: nil,
			check: func(t *testing.T, r *export.Config) {
				if !r.Enabled {
					t.Error("expected enabled=true from global")
				}
				if r.Format != "log" {
					t.Errorf("expected format=log, got %s", r.Format)
				}
				if r.MaxFiles != 50 {
					t.Errorf("expected max_files=50, got %d", r.MaxFiles)
				}
			},
		},
		{
			name:       "only project — project values used",
			global:     export.Defaults(),
			project:    &export.Config{Enabled: true, Format: "both"},
			globalSet:  nil,
			projectSet: fieldSet{"enabled": true, "format": true},
			check: func(t *testing.T, r *export.Config) {
				if !r.Enabled {
					t.Error("expected enabled=true from project")
				}
				if r.Format != "both" {
					t.Errorf("expected format=both, got %s", r.Format)
				}
			},
		},
		{
			name:       "neither file — all defaults",
			global:     export.Defaults(),
			project:    export.Defaults(),
			globalSet:  nil,
			projectSet: nil,
			check: func(t *testing.T, r *export.Config) {
				defaults := export.Defaults()
				if r.Enabled != defaults.Enabled {
					t.Errorf("expected enabled=%v, got %v", defaults.Enabled, r.Enabled)
				}
				if r.Format != defaults.Format {
					t.Errorf("expected format=%s, got %s", defaults.Format, r.Format)
				}
			},
		},
		{
			name:       "project explicitly sets enabled=false — overrides global true",
			global:     &export.Config{Enabled: true},
			project:    &export.Config{Enabled: false},
			globalSet:  fieldSet{"enabled": true},
			projectSet: fieldSet{"enabled": true},
			check: func(t *testing.T, r *export.Config) {
				if r.Enabled {
					t.Error("expected enabled=false (project explicit override)")
				}
			},
		},
		{
			name:       "project explicitly sets max_files=0 — overrides global 50",
			global:     &export.Config{MaxFiles: 50},
			project:    &export.Config{MaxFiles: 0},
			globalSet:  fieldSet{"max_files": true},
			projectSet: fieldSet{"max_files": true},
			check: func(t *testing.T, r *export.Config) {
				if r.MaxFiles != 0 {
					t.Errorf("expected max_files=0 (project explicit zero), got %d", r.MaxFiles)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeExportConfigs(tt.global, tt.project, tt.globalSet, tt.projectSet)
			tt.check(t, result)
		})
	}
}

func TestMergeShellConfigs(t *testing.T) {
	tests := []struct {
		name       string
		global     *shell.Config
		project    *shell.Config
		globalSet  fieldSet
		projectSet fieldSet
		check      func(t *testing.T, result *shell.Config)
	}{
		{
			name:    "disjoint fields — all values preserved",
			global:  &shell.Config{Enabled: true, Shell: "zsh", ExcludeTargets: []string{"watch"}},
			project: &shell.Config{FormatTemplate: "lazymake {target}", ExcludeTargets: []string{"ci"}},
			globalSet: fieldSet{
				"enabled": true, "shell": true, "exclude_targets": true,
			},
			projectSet: fieldSet{
				"format_template": true, "exclude_targets": true,
			},
			check: func(t *testing.T, r *shell.Config) {
				if !r.Enabled {
					t.Error("expected enabled=true from global")
				}
				if r.Shell != "zsh" {
					t.Errorf("expected shell=zsh, got %s", r.Shell)
				}
				if r.FormatTemplate != "lazymake {target}" {
					t.Errorf("expected format_template=lazymake {target}, got %s", r.FormatTemplate)
				}
				assertSliceEqual(t, r.ExcludeTargets, []string{"watch", "ci"})
			},
		},
		{
			name:    "overlapping scalars — project wins",
			global:  &shell.Config{Shell: "bash", FormatTemplate: "make {target}"},
			project: &shell.Config{Shell: "zsh", FormatTemplate: "lazymake {target}"},
			globalSet: fieldSet{
				"shell": true, "format_template": true,
			},
			projectSet: fieldSet{
				"shell": true, "format_template": true,
			},
			check: func(t *testing.T, r *shell.Config) {
				if r.Shell != "zsh" {
					t.Errorf("expected shell=zsh (project wins), got %s", r.Shell)
				}
				if r.FormatTemplate != "lazymake {target}" {
					t.Errorf("expected format_template=lazymake {target} (project wins), got %s", r.FormatTemplate)
				}
			},
		},
		{
			name:       "project explicitly sets enabled=false — overrides global true",
			global:     &shell.Config{Enabled: true},
			project:    &shell.Config{Enabled: false},
			globalSet:  fieldSet{"enabled": true},
			projectSet: fieldSet{"enabled": true},
			check: func(t *testing.T, r *shell.Config) {
				if r.Enabled {
					t.Error("expected enabled=false (project explicit override)")
				}
			},
		},
		{
			name:       "project explicitly sets include_timestamp=false — overrides default true",
			global:     shell.Defaults(),
			project:    &shell.Config{IncludeTimestamp: false},
			globalSet:  nil,
			projectSet: fieldSet{"include_timestamp": true},
			check: func(t *testing.T, r *shell.Config) {
				if r.IncludeTimestamp {
					t.Error("expected include_timestamp=false (project explicit override)")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeShellConfigs(tt.global, tt.project, tt.globalSet, tt.projectSet)
			tt.check(t, result)
		})
	}
}

func TestMergeSafetyConfigs(t *testing.T) {
	tests := []struct {
		name       string
		global     *safety.Config
		project    *safety.Config
		globalSet  fieldSet
		projectSet fieldSet
		check      func(t *testing.T, result *safety.Config)
	}{
		{
			name: "custom_rules from both files are appended",
			global: &safety.Config{
				Enabled: true,
				CustomRules: []safety.Rule{
					{ID: "global-rule", Description: "from global"},
				},
			},
			project: &safety.Config{
				Enabled: true,
				CustomRules: []safety.Rule{
					{ID: "project-rule", Description: "from project"},
				},
			},
			globalSet:  fieldSet{"enabled": true, "custom_rules": true},
			projectSet: fieldSet{"enabled": true, "custom_rules": true},
			check: func(t *testing.T, r *safety.Config) {
				if len(r.CustomRules) != 2 {
					t.Fatalf("expected 2 custom rules, got %d", len(r.CustomRules))
				}
				if r.CustomRules[0].ID != "global-rule" {
					t.Errorf("expected first rule=global-rule, got %s", r.CustomRules[0].ID)
				}
				if r.CustomRules[1].ID != "project-rule" {
					t.Errorf("expected second rule=project-rule, got %s", r.CustomRules[1].ID)
				}
			},
		},
		{
			name: "enabled_rules from both files are unioned and deduplicated",
			global: &safety.Config{
				Enabled:      true,
				EnabledRules: []string{"rm-rf-root", "git-force-push"},
			},
			project: &safety.Config{
				Enabled:      true,
				EnabledRules: []string{"git-force-push", "docker-system-prune"},
			},
			globalSet:  fieldSet{"enabled": true, "enabled_rules": true},
			projectSet: fieldSet{"enabled": true, "enabled_rules": true},
			check: func(t *testing.T, r *safety.Config) {
				assertSliceEqual(t, r.EnabledRules, []string{"rm-rf-root", "git-force-push", "docker-system-prune"})
			},
		},
		{
			name:    "exclude_targets from both files are unioned",
			global:  &safety.Config{Enabled: true, ExcludeTargets: []string{"clean"}},
			project: &safety.Config{Enabled: true, ExcludeTargets: []string{"clean", "test"}},
			globalSet:  fieldSet{"enabled": true, "exclude_targets": true},
			projectSet: fieldSet{"enabled": true, "exclude_targets": true},
			check: func(t *testing.T, r *safety.Config) {
				assertSliceEqual(t, r.ExcludeTargets, []string{"clean", "test"})
			},
		},
		{
			name:       "project sets enabled=false — overrides default true",
			global:     safety.DefaultConfig(),
			project:    &safety.Config{Enabled: false},
			globalSet:  nil,
			projectSet: fieldSet{"enabled": true},
			check: func(t *testing.T, r *safety.Config) {
				if r.Enabled {
					t.Error("expected enabled=false (project explicit override)")
				}
			},
		},
		{
			name:       "neither file — all defaults",
			global:     safety.DefaultConfig(),
			project:    safety.DefaultConfig(),
			globalSet:  nil,
			projectSet: nil,
			check: func(t *testing.T, r *safety.Config) {
				if !r.Enabled {
					t.Error("expected enabled=true (default)")
				}
				if r.EnabledRules != nil {
					t.Error("expected enabled_rules=nil (default)")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeSafetyConfigs(tt.global, tt.project, tt.globalSet, tt.projectSet)
			tt.check(t, result)
		})
	}
}

func TestReadAndMergeFromYAML(t *testing.T) {
	type viperPair struct {
		global  *viper.Viper
		project *viper.Viper
	}

	tests := []struct {
		name        string
		globalYAML  string
		projectYAML string
		check       func(t *testing.T, vp viperPair)
	}{
		{
			name: "both files — disjoint export fields merged",
			globalYAML: `
export:
  enabled: true
  exclude_targets:
    - watch
`,
			projectYAML: `
export:
  format: both
  exclude_targets:
    - ci
`,
			check: func(t *testing.T, vp viperPair) {
				ge, gs := readExportConfig(vp.global)
				pe, ps := readExportConfig(vp.project)
				r := mergeExportConfigs(ge, pe, gs, ps)
				if !r.Enabled {
					t.Error("expected enabled=true from global")
				}
				if r.Format != "both" {
					t.Errorf("expected format=both from project, got %s", r.Format)
				}
				assertSliceEqual(t, r.ExcludeTargets, []string{"watch", "ci"})
			},
		},
		{
			name: "both files — shell_integration merged",
			globalYAML: `
shell_integration:
  enabled: true
  shell: zsh
  exclude_targets:
    - watch
`,
			projectYAML: `
shell_integration:
  format_template: "lazymake {target}"
  exclude_targets:
    - ci
`,
			check: func(t *testing.T, vp viperPair) {
				gs, gset := readShellConfig(vp.global)
				ps, pset := readShellConfig(vp.project)
				r := mergeShellConfigs(gs, ps, gset, pset)
				if !r.Enabled {
					t.Error("expected enabled=true from global")
				}
				if r.Shell != "zsh" {
					t.Errorf("expected shell=zsh, got %s", r.Shell)
				}
				if r.FormatTemplate != "lazymake {target}" {
					t.Errorf("expected format_template=lazymake {target}, got %s", r.FormatTemplate)
				}
				assertSliceEqual(t, r.ExcludeTargets, []string{"watch", "ci"})
			},
		},
		{
			name: "both files — safety custom_rules appended",
			globalYAML: `
safety:
  enabled: true
  custom_rules:
    - id: global-rule
      description: "global custom rule"
      severity: warning
      patterns:
        - "dangerous-global"
`,
			projectYAML: `
safety:
  custom_rules:
    - id: project-rule
      description: "project custom rule"
      severity: critical
      patterns:
        - "dangerous-project"
`,
			check: func(t *testing.T, vp viperPair) {
				gs, gset := readSafetyConfig(vp.global)
				ps, pset := readSafetyConfig(vp.project)
				r := mergeSafetyConfigs(gs, ps, gset, pset)
				if !r.Enabled {
					t.Error("expected enabled=true")
				}
				if len(r.CustomRules) != 2 {
					t.Fatalf("expected 2 custom rules, got %d", len(r.CustomRules))
				}
				if r.CustomRules[0].ID != "global-rule" {
					t.Errorf("expected first rule id=global-rule, got %s", r.CustomRules[0].ID)
				}
				if r.CustomRules[1].ID != "project-rule" {
					t.Errorf("expected second rule id=project-rule, got %s", r.CustomRules[1].ID)
				}
				if r.CustomRules[1].Severity != safety.SeverityCritical {
					t.Errorf("expected project rule severity=critical, got %v", r.CustomRules[1].Severity)
				}
			},
		},
		{
			name: "only global file — global values used",
			globalYAML: `export:
  enabled: true
  format: log
  max_files: 50
`,
			projectYAML: "",
			check: func(t *testing.T, vp viperPair) {
				ge, gs := readExportConfig(vp.global)
				pe, ps := readExportConfig(vp.project)
				r := mergeExportConfigs(ge, pe, gs, ps)
				if !r.Enabled {
					t.Error("expected enabled=true from global")
				}
				if r.Format != "log" {
					t.Errorf("expected format=log, got %s", r.Format)
				}
				if r.MaxFiles != 50 {
					t.Errorf("expected max_files=50, got %d", r.MaxFiles)
				}
			},
		},
		{
			name:       "only project file — project values used",
			globalYAML: "",
			projectYAML: `export:
  enabled: true
  format: both
`,
			check: func(t *testing.T, vp viperPair) {
				ge, gs := readExportConfig(vp.global)
				pe, ps := readExportConfig(vp.project)
				r := mergeExportConfigs(ge, pe, gs, ps)
				if !r.Enabled {
					t.Error("expected enabled=true from project")
				}
				if r.Format != "both" {
					t.Errorf("expected format=both, got %s", r.Format)
				}
			},
		},
		{
			name:        "neither file — defaults",
			globalYAML:  "",
			projectYAML: "",
			check: func(t *testing.T, vp viperPair) {
				ge, gs := readExportConfig(vp.global)
				pe, ps := readExportConfig(vp.project)
				r := mergeExportConfigs(ge, pe, gs, ps)
				defaults := export.Defaults()
				if r.Enabled != defaults.Enabled {
					t.Errorf("expected enabled=%v, got %v", defaults.Enabled, r.Enabled)
				}
				if r.Format != defaults.Format {
					t.Errorf("expected format=%s, got %s", defaults.Format, r.Format)
				}
			},
		},
		{
			name: "project explicitly sets enabled false — overrides global true",
			globalYAML: `
export:
  enabled: true
`,
			projectYAML: `
export:
  enabled: false
`,
			check: func(t *testing.T, vp viperPair) {
				ge, gs := readExportConfig(vp.global)
				pe, ps := readExportConfig(vp.project)
				r := mergeExportConfigs(ge, pe, gs, ps)
				if r.Enabled {
					t.Error("expected enabled=false (project explicit override)")
				}
			},
		},
		{
			name: "project explicitly sets max_files 0 — overrides global 50",
			globalYAML: `
export:
  max_files: 50
`,
			projectYAML: `
export:
  max_files: 0
`,
			check: func(t *testing.T, vp viperPair) {
				ge, gs := readExportConfig(vp.global)
				pe, ps := readExportConfig(vp.project)
				r := mergeExportConfigs(ge, pe, gs, ps)
				if r.MaxFiles != 0 {
					t.Errorf("expected max_files=0 (project explicit zero), got %d", r.MaxFiles)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			var gv *viper.Viper
			if tt.globalYAML != "" {
				writeYAML(t, dir, "global.yaml", tt.globalYAML)
				gv = loadViperFromFile(filepath.Join(dir, "global.yaml"))
			}

			var pv *viper.Viper
			if tt.projectYAML != "" {
				writeYAML(t, dir, "project.yaml", tt.projectYAML)
				pv = loadViperFromFile(filepath.Join(dir, "project.yaml"))
			}

			tt.check(t, viperPair{global: gv, project: pv})
		})
	}
}

func TestLoadViperFromFile(t *testing.T) {
	t.Run("nonexistent file returns nil", func(t *testing.T) {
		v := loadViperFromFile("/nonexistent/path/file.yaml")
		if v != nil {
			t.Error("expected nil for nonexistent file")
		}
	})

	t.Run("valid file returns viper", func(t *testing.T) {
		dir := t.TempDir()
		writeYAML(t, dir, "test.yaml", "export:\n  enabled: true\n")
		v := loadViperFromFile(filepath.Join(dir, "test.yaml"))
		if v == nil {
			t.Fatal("expected non-nil viper for valid file")
		}
		if !v.GetBool("export.enabled") {
			t.Error("expected export.enabled=true")
		}
	})
}

func TestParseCustomRules(t *testing.T) {
	rulesMaps := []map[string]interface{}{
		{
			"id":          "test-rule",
			"description": "test description",
			"severity":    "critical",
			"patterns":    []interface{}{"pattern1", "pattern2"},
			"suggestion":  "try this instead",
		},
		{
			"id":       "default-severity",
			"severity": "unknown-value",
		},
	}

	rules := parseCustomRules(rulesMaps)
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}

	if rules[0].ID != "test-rule" {
		t.Errorf("expected id=test-rule, got %s", rules[0].ID)
	}
	if rules[0].Severity != safety.SeverityCritical {
		t.Errorf("expected severity=critical, got %v", rules[0].Severity)
	}
	if len(rules[0].Patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(rules[0].Patterns))
	}
	if rules[0].Suggestion != "try this instead" {
		t.Errorf("expected suggestion='try this instead', got %s", rules[0].Suggestion)
	}

	// Unknown severity defaults to warning
	if rules[1].Severity != safety.SeverityWarning {
		t.Errorf("expected default severity=warning, got %v", rules[1].Severity)
	}
}

// assertSliceEqual checks that two string slices have the same elements in the same order.
func assertSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("slice length mismatch: got %v (len %d), want %v (len %d)", got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
		}
	}
}
