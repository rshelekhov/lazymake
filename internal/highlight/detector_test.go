package highlight

import (
	"testing"
)

func TestDetectLanguageShebang(t *testing.T) {
	tests := []struct {
		name     string
		recipe   []string
		expected string
	}{
		{
			name:     "Python shebang",
			recipe:   []string{"#!/usr/bin/env python3", "print('hello')"},
			expected: "python",
		},
		{
			name:     "Bash shebang",
			recipe:   []string{"#!/bin/bash", "echo hello"},
			expected: "bash",
		},
		{
			name:     "Ruby shebang",
			recipe:   []string{"#!/usr/bin/env ruby", "puts 'hello'"},
			expected: "ruby",
		},
		{
			name:     "Node shebang",
			recipe:   []string{"#!/usr/bin/env node", "console.log('hello')"},
			expected: "javascript",
		},
		{
			name:     "Sh shebang",
			recipe:   []string{"#!/bin/sh", "echo hello"},
			expected: "sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguage(tt.recipe, "")
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDetectLanguageCommands(t *testing.T) {
	tests := []struct {
		name     string
		recipe   []string
		expected string
	}{
		{
			name:     "Go commands",
			recipe:   []string{"go build -o app", "go test ./..."},
			expected: "go",
		},
		{
			name:     "Python commands",
			recipe:   []string{"python3 setup.py build", "pip install -r requirements.txt"},
			expected: "python",
		},
		{
			name:     "NPM commands",
			recipe:   []string{"npm install", "npm run build"},
			expected: "javascript",
		},
		{
			name:     "Yarn commands",
			recipe:   []string{"yarn install", "yarn build"},
			expected: "javascript",
		},
		{
			name:     "Cargo commands",
			recipe:   []string{"cargo build", "cargo test"},
			expected: "rust",
		},
		{
			name:     "GCC commands",
			recipe:   []string{"gcc -o app main.c"},
			expected: "c",
		},
		{
			name:     "G++ commands",
			recipe:   []string{"g++ -o app main.cpp"},
			expected: "cpp",
		},
		{
			name:     "Ruby commands",
			recipe:   []string{"ruby script.rb", "bundle install"},
			expected: "ruby",
		},
		{
			name:     "Java commands",
			recipe:   []string{"javac Main.java", "java Main"},
			expected: "java",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguage(tt.recipe, "")
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDetectLanguageManualOverride(t *testing.T) {
	tests := []struct {
		name     string
		recipe   []string
		override string
		expected string
	}{
		{
			name:     "Override to Python",
			recipe:   []string{"echo 'test'"},
			override: "python",
			expected: "python",
		},
		{
			name:     "Override to Go",
			recipe:   []string{"docker build ."},
			override: "go",
			expected: "go",
		},
		{
			name:     "Override with alias (py -> python)",
			recipe:   []string{"echo 'test'"},
			override: "py",
			expected: "python",
		},
		{
			name:     "Override with alias (js -> javascript)",
			recipe:   []string{"echo 'test'"},
			override: "js",
			expected: "javascript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguage(tt.recipe, tt.override)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDetectLanguageFallback(t *testing.T) {
	tests := []struct {
		name     string
		recipe   []string
		expected string
	}{
		{
			name:     "Empty recipe",
			recipe:   []string{},
			expected: "bash",
		},
		{
			name:     "Generic shell commands",
			recipe:   []string{"echo 'hello'", "ls -la"},
			expected: "bash",
		},
		{
			name:     "Comments only",
			recipe:   []string{"# This is a comment"},
			expected: "bash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguage(tt.recipe, "")
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDetectShebang(t *testing.T) {
	tests := []struct {
		shebang  string
		expected string
	}{
		{"#!/bin/bash", "bash"},
		{"#!/usr/bin/env python3", "python"},
		{"#!/usr/bin/env python", "python"},
		{"#!/usr/bin/env ruby", "ruby"},
		{"#!/bin/sh", "sh"},
		{"#!/usr/bin/env node", "javascript"},
		{"#!/usr/bin/env perl", "perl"},
		{"#!/usr/bin/env php", "php"},
		{"#not a shebang", ""},
		{"echo hello", ""},
	}

	for _, tt := range tests {
		t.Run(tt.shebang, func(t *testing.T) {
			result := detectShebang([]string{tt.shebang})
			if result != tt.expected {
				t.Errorf("for %q expected %s, got %s", tt.shebang, tt.expected, result)
			}
		})
	}
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"python", "python"},
		{"Python", "python"},
		{"PYTHON", "python"},
		{"py", "python"},
		{"js", "javascript"},
		{"golang", "go"},
		{"sh", "bash"},
		{"shell", "bash"},
		{"yml", "yaml"},
		{"dockerfile", "docker"},
		{"rust", "rust"}, // No alias, unchanged
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeLanguage(tt.input)
			if result != tt.expected {
				t.Errorf("for %q expected %s, got %s", tt.input, tt.expected, result)
			}
		})
	}
}

func TestParseLanguageOverride(t *testing.T) {
	tests := []struct {
		comment  string
		expected string
	}{
		{"# language: python", "python"},
		{"# lang: go", "go"},
		{"# syntax: rust", "rust"},
		{"# language: JavaScript", "javascript"},
		{"#language:python", "python"},
		{"  # lang: ruby  ", "ruby"},
		{"# not a language override", ""},
		{"language: python", ""}, // Missing #
		{"# foo: bar", ""},        // Wrong keyword
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			result := ParseLanguageOverride(tt.comment)
			if result != tt.expected {
				t.Errorf("for %q expected %s, got %s", tt.comment, tt.expected, result)
			}
		})
	}
}

func TestDetectFromCommandsWeighting(t *testing.T) {
	// Test that higher-weight patterns win
	recipe := []string{
		"curl https://example.com", // bash, weight 50
		"go build -o app",           // go, weight 90
		"go test ./...",             // go, weight 90
	}

	result := detectFromCommands(recipe)
	if result != "go" {
		t.Errorf("expected 'go' (higher weight), got %s", result)
	}
}

func TestDetectLanguageMultipleHeuristics(t *testing.T) {
	// Test that shebang takes priority over commands
	recipe := []string{
		"#!/usr/bin/env python3",
		"go build -o app", // Would normally detect as go
	}

	result := DetectLanguage(recipe, "")
	if result != "python" {
		t.Errorf("shebang should take priority over commands, got %s", result)
	}
}

func TestDetectLanguagePriority(t *testing.T) {
	recipe := []string{
		"#!/bin/bash",
		"go build .",
	}

	// Manual override should take highest priority
	result := DetectLanguage(recipe, "python")
	if result != "python" {
		t.Errorf("manual override should take highest priority, got %s", result)
	}

	// Without override, shebang should win
	result = DetectLanguage(recipe, "")
	if result != "bash" {
		t.Errorf("shebang should take priority over commands, got %s", result)
	}
}
