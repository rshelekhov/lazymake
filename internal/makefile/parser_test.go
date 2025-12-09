package makefile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	// Create a temporary test Makefile
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Makefile")

	content := `.PHONY: all build test clean

# Single hash comment
build:
	@echo "Building..."

## Double hash comment (industry standard)
test:
	@echo "Testing..."

clean: ## Inline double hash comment
	@echo "Cleaning..."

run: # Inline single hash comment
	@echo "Running..."

## Preceding double hash
format:
	@echo "Formatting..."

install: ## Inline override
	@echo "Installing..."

no-comment:
	@echo "No comment..."

all: build test ## Inline after deps
	@echo "All done!"
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	targets, err := Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test cases
	tests := []struct {
		name            string
		expectedDesc    string
		expectedType    CommentType
	}{
		{"build", "Single hash comment", CommentSingle},
		{"test", "Double hash comment (industry standard)", CommentDouble},
		{"clean", "Inline double hash comment", CommentDouble},
		{"run", "Inline single hash comment", CommentSingle},
		{"format", "Preceding double hash", CommentDouble},
		{"install", "Inline override", CommentDouble},
		{"no-comment", "", CommentNone},
		{"all", "Inline after deps", CommentDouble},
	}

	// Create a map for easier lookup
	targetMap := make(map[string]Target)
	for _, target := range targets {
		targetMap[target.Name] = target
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, found := targetMap[tt.name]
			if !found {
				t.Errorf("Target %s not found", tt.name)
				return
			}

			if target.Description != tt.expectedDesc {
				t.Errorf("Target %s: expected description %q, got %q",
					tt.name, tt.expectedDesc, target.Description)
			}

			if target.CommentType != tt.expectedType {
				t.Errorf("Target %s: expected comment type %v, got %v",
					tt.name, tt.expectedType, target.CommentType)
			}
		})
	}
}

func TestParseInlineCommentPriority(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Makefile")

	// Test that inline comments override preceding comments
	content := `## Preceding comment
target: ## Inline comment
	@echo "test"
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	targets, err := Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(targets) != 1 {
		t.Fatalf("Expected 1 target, got %d", len(targets))
	}

	target := targets[0]
	if target.Description != "Inline comment" {
		t.Errorf("Expected inline comment to override, got %q", target.Description)
	}

	if target.CommentType != CommentDouble {
		t.Errorf("Expected CommentDouble, got %v", target.CommentType)
	}
}

func TestParseMultipleTargets(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Makefile")

	// Test multiple targets on one line
	content := `## Build and test
build test: ## Inline comment
	@echo "test"
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	targets, err := Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(targets) != 2 {
		t.Fatalf("Expected 2 targets, got %d", len(targets))
	}

	// Both targets should have the inline comment
	for _, target := range targets {
		if target.Description != "Inline comment" {
			t.Errorf("Target %s: expected inline comment, got %q",
				target.Name, target.Description)
		}
		if target.CommentType != CommentDouble {
			t.Errorf("Target %s: expected CommentDouble, got %v",
				target.Name, target.CommentType)
		}
	}
}

func TestParseRecipeLinesWithColons(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Makefile")

	// Test that recipe lines containing colons are not parsed as targets
	// This was a bug where recipe lines like '@echo "foo: bar"' were parsed as targets
	content := `release: ## Create a release
	@echo "To create a release:"
	@echo "1. Create and push a tag: git tag -a v0.1.0"
	@echo "2. Push the tag: git push origin v0.1.0"

build: ## Build the app
	go build -o app
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	targets, err := Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should only have 2 targets: release and build
	if len(targets) != 2 {
		t.Fatalf("Expected 2 targets, got %d", len(targets))
	}

	// Verify the target names are correct
	expectedTargets := map[string]bool{"release": false, "build": false}
	for _, target := range targets {
		if _, expected := expectedTargets[target.Name]; expected {
			expectedTargets[target.Name] = true
		} else {
			t.Errorf("Unexpected target found: %s", target.Name)
		}
	}

	for name, found := range expectedTargets {
		if !found {
			t.Errorf("Expected target %s not found", name)
		}
	}
}
