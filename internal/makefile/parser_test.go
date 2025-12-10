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
		name         string
		expectedDesc string
		expectedType CommentType
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

// TestParseDependencies tests the new dependency extraction feature
func TestParseDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Makefile")

	// Create a Makefile with various dependency patterns
	content := `
# Simple dependencies
build: deps compile ## Build the project
	@go build

compile: ## Compile source
	@go build -o bin/app

deps: ## Install dependencies
	@go mod download

# Chained dependencies
test: build ## Run tests
	@go test ./...

# Multiple dependencies
all: deps build test ## Do everything
	@echo "Done!"

# No dependencies
clean: ## Clean build files
	@rm -rf bin/

# Pattern rule (should be ignored)
%.o: %.c ## Pattern rule
	@gcc -c $< -o $@

# Order-only prerequisites
order: normal-dep | order-only-dep ## Order-only test
	@echo "test"

# Variables (should be ignored)
var-test: $(DEPS) real-dep ## With variable
	@echo "test"

# File paths (should be filtered)
file-dep: src/main.go regular-target ## File and target mix
	@echo "test"
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	targets, err := Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Create map for easier lookup
	targetMap := make(map[string]Target)
	for _, target := range targets {
		targetMap[target.Name] = target
	}

	// Test cases for dependency extraction
	tests := []struct {
		name         string
		expectedDeps []string
	}{
		{"build", []string{"deps", "compile"}},
		{"compile", nil}, // No dependencies
		{"deps", nil},
		{"test", []string{"build"}},
		{"all", []string{"deps", "build", "test"}},
		{"clean", nil},
		{"order", []string{"normal-dep"}},        // Order-only deps filtered out
		{"var-test", []string{"real-dep"}},       // Variables filtered out
		{"file-dep", []string{"regular-target"}}, // Complex file paths filtered
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, found := targetMap[tt.name]
			if !found {
				t.Errorf("Target %s not found", tt.name)
				return
			}

			// Check dependency count
			if len(target.Dependencies) != len(tt.expectedDeps) {
				t.Errorf("Target %s: expected %d dependencies, got %d: %v",
					tt.name, len(tt.expectedDeps), len(target.Dependencies), target.Dependencies)
				return
			}

			// Check each dependency
			for i, expectedDep := range tt.expectedDeps {
				if target.Dependencies[i] != expectedDep {
					t.Errorf("Target %s: expected dependency[%d]=%s, got %s",
						tt.name, i, expectedDep, target.Dependencies[i])
				}
			}
		})
	}
}

// TestParseDependenciesWithComments verifies dependencies are extracted correctly
// even when there are inline comments
func TestParseDependenciesWithComments(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Makefile")

	// The tricky part is separating dependencies from comments
	// "build: dep1 dep2 ## Comment" -> deps are ["dep1", "dep2"], not including the comment
	content := `
target1: dep1 dep2 ## Inline comment after deps
	@echo "test"

target2: dep3 # Single hash comment
	@echo "test"

target3: dep4 dep5
	@echo "test"
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	targets, err := Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	targetMap := make(map[string]Target)
	for _, target := range targets {
		targetMap[target.Name] = target
	}

	// Test that comments don't interfere with dependency extraction
	tests := []struct {
		name         string
		expectedDeps []string
	}{
		{"target1", []string{"dep1", "dep2"}},
		{"target2", []string{"dep3"}},
		{"target3", []string{"dep4", "dep5"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := targetMap[tt.name]

			if len(target.Dependencies) != len(tt.expectedDeps) {
				t.Errorf("Expected %d deps, got %d: %v",
					len(tt.expectedDeps), len(target.Dependencies), target.Dependencies)
				return
			}

			for i, expected := range tt.expectedDeps {
				if target.Dependencies[i] != expected {
					t.Errorf("Dependency[%d]: expected %s, got %s",
						i, expected, target.Dependencies[i])
				}
			}
		})
	}
}

// TestParseRecipes tests recipe extraction functionality
func TestParseRecipes(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Makefile")

	// Use explicit tabs for recipe lines
	content := "## Build the app\n" +
		"build:\n" +
		"\tgo build -o app\n" +
		"\tchmod +x app\n" +
		"\n" +
		"## Run tests\n" +
		"test:\n" +
		"\tgo test ./...\n" +
		"\n" +
		"## Clean build artifacts\n" +
		"clean:\n" +
		"\trm -rf build/\n" +
		"\trm -f app\n" +
		"\n" +
		"## Single line recipe\n" +
		"single:\n" +
		"\techo \"one command\"\n" +
		"\n" +
		"## Meta-target with no recipe\n" +
		"all: build test\n" +
		"\n" +
		"## Recipe with special characters\n" +
		"special:\n" +
		"\t@echo \"Testing: special\"\n" +
		"\techo 'single quotes'\n" +
		"\techo \"double quotes\"\n"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	targets, err := Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	targetMap := make(map[string]Target)
	for _, target := range targets {
		targetMap[target.Name] = target
	}

	tests := []struct {
		name           string
		expectedRecipe []string
	}{
		{
			name: "build",
			expectedRecipe: []string{
				"go build -o app",
				"chmod +x app",
			},
		},
		{
			name: "test",
			expectedRecipe: []string{
				"go test ./...",
			},
		},
		{
			name: "clean",
			expectedRecipe: []string{
				"rm -rf build/",
				"rm -f app",
			},
		},
		{
			name: "single",
			expectedRecipe: []string{
				`echo "one command"`,
			},
		},
		{
			name:           "all",
			expectedRecipe: nil, // Meta-target with no recipe
		},
		{
			name: "special",
			expectedRecipe: []string{
				`@echo "Testing: special"`,
				`echo 'single quotes'`,
				`echo "double quotes"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, found := targetMap[tt.name]
			if !found {
				t.Errorf("Target %s not found", tt.name)
				return
			}

			if len(target.Recipe) != len(tt.expectedRecipe) {
				t.Errorf("Target %s: expected %d recipe lines, got %d\nExpected: %v\nGot: %v",
					tt.name, len(tt.expectedRecipe), len(target.Recipe), tt.expectedRecipe, target.Recipe)
				return
			}

			for i, expected := range tt.expectedRecipe {
				if target.Recipe[i] != expected {
					t.Errorf("Target %s recipe[%d]: expected %q, got %q",
						tt.name, i, expected, target.Recipe[i])
				}
			}
		})
	}
}

// TestParseMultiTargetRecipe tests that multiple targets on one line share the same recipe
func TestParseMultiTargetRecipe(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Makefile")

	// Use explicit tabs (cannot rely on editor tab settings with backticks)
	content := "## Build and install\n" +
		"build install: ## Both targets share the recipe\n" +
		"\tgo build -o app\n" +
		"\tcp app /usr/local/bin\n" +
		"\n" +
		"## Separate targets\n" +
		"separate1:\n" +
		"\techo \"first\"\n" +
		"\n" +
		"separate2:\n" +
		"\techo \"second\"\n"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	targets, err := Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	targetMap := make(map[string]Target)
	for _, target := range targets {
		targetMap[target.Name] = target
	}

	// Both build and install should have the same recipe
	buildTarget := targetMap["build"]
	installTarget := targetMap["install"]

	expectedRecipe := []string{
		"go build -o app",
		"cp app /usr/local/bin",
	}

	if len(buildTarget.Recipe) != len(expectedRecipe) {
		t.Errorf("build: expected %d recipe lines, got %d", len(expectedRecipe), len(buildTarget.Recipe))
	}

	if len(installTarget.Recipe) != len(expectedRecipe) {
		t.Errorf("install: expected %d recipe lines, got %d", len(expectedRecipe), len(installTarget.Recipe))
	}

	for i, expected := range expectedRecipe {
		if buildTarget.Recipe[i] != expected {
			t.Errorf("build recipe[%d]: expected %q, got %q", i, expected, buildTarget.Recipe[i])
		}
		if installTarget.Recipe[i] != expected {
			t.Errorf("install recipe[%d]: expected %q, got %q", i, expected, installTarget.Recipe[i])
		}
	}

	// separate1 and separate2 should have different recipes
	separate1 := targetMap["separate1"]
	separate2 := targetMap["separate2"]

	if len(separate1.Recipe) != 1 || separate1.Recipe[0] != `echo "first"` {
		t.Errorf("separate1 has wrong recipe: %v", separate1.Recipe)
	}

	if len(separate2.Recipe) != 1 || separate2.Recipe[0] != `echo "second"` {
		t.Errorf("separate2 has wrong recipe: %v", separate2.Recipe)
	}
}

// TestParseRecipeWithVariables tests recipes containing variables
func TestParseRecipeWithVariables(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Makefile")

	content := "VAR = value\n\n" +
		"target:\n" +
		"\techo $(VAR)\n" +
		"\techo ${VAR}\n" +
		"\techo $$VAR\n"

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
	expectedRecipe := []string{
		"echo $(VAR)",
		"echo ${VAR}",
		"echo $$VAR",
	}

	if len(target.Recipe) != len(expectedRecipe) {
		t.Errorf("Expected %d recipe lines, got %d", len(expectedRecipe), len(target.Recipe))
	}

	for i, expected := range expectedRecipe {
		if target.Recipe[i] != expected {
			t.Errorf("Recipe[%d]: expected %q, got %q", i, expected, target.Recipe[i])
		}
	}
}

// TestParseRecipeSeparation tests that recipes are correctly separated by empty lines and comments
func TestParseRecipeSeparation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Makefile")

	content := "target1:\n" +
		"\techo \"first\"\n" +
		"\n" +
		"target2:\n" +
		"\techo \"second\"\n" +
		"\n" +
		"# Comment between targets\n" +
		"target3:\n" +
		"\techo \"third\"\n" +
		"\n" +
		"## Documentation comment\n" +
		"target4:\n" +
		"\techo \"fourth\"\n"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	targets, err := Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(targets) != 4 {
		t.Fatalf("Expected 4 targets, got %d", len(targets))
	}

	expected := map[string][]string{
		"target1": {`echo "first"`},
		"target2": {`echo "second"`},
		"target3": {`echo "third"`},
		"target4": {`echo "fourth"`},
	}

	for _, target := range targets {
		expectedRecipe := expected[target.Name]
		if len(target.Recipe) != len(expectedRecipe) {
			t.Errorf("Target %s: expected %d recipe lines, got %d: %v",
				target.Name, len(expectedRecipe), len(target.Recipe), target.Recipe)
			continue
		}

		for i, exp := range expectedRecipe {
			if target.Recipe[i] != exp {
				t.Errorf("Target %s recipe[%d]: expected %q, got %q",
					target.Name, i, exp, target.Recipe[i])
			}
		}
	}
}
