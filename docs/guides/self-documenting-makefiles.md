# Writing Self-Documenting Makefiles

lazymake supports the industry-standard `##` convention for documenting Makefile targets. This guide shows you how to write Makefiles that document themselves.

## The `##` Convention

Use `##` comments to mark targets for documentation:

```makefile
.PHONY: build test deploy

build: ## Build the application
	go build -o app main.go

test: ## Run all tests
	go test ./...

## Deploy to production
deploy:
	./scripts/deploy.sh
```

## Comment Styles

lazymake recognizes three types of comments:

### 1. Inline `##` Comments (Highest Priority)

Place documentation on the same line as the target:

```makefile
build: ## Build the application
	go build -o app main.go
```

- **Shown in**: Target list with cyan color
- **Use for**: Primary target documentation
- **Priority**: Overrides any preceding comments

### 2. Preceding `##` Comments

Place documentation on the line before the target:

```makefile
## Build the application
build:
	go build -o app main.go
```

- **Shown in**: Target list with cyan color
- **Use for**: Longer descriptions that don't fit inline
- **Priority**: Used if no inline comment exists

### 3. Regular `#` Comments (Backward Compatible)

Standard single-hash comments:

```makefile
# Build the application
build:
	go build -o app main.go
```

- **Shown in**: Target list with gray color
- **Use for**: Internal implementation notes
- **Priority**: Lowest (only used if no `##` comments exist)

## Best Practices

### ✅ DO: Use `##` for Public Targets

```makefile
build: ## Build the application
test: ## Run all tests
deploy: ## Deploy to production
clean: ## Clean build artifacts
```

### ✅ DO: Keep Descriptions Concise

```makefile
# Good - concise and clear
build: ## Build the application

# Avoid - too verbose
build: ## This target will build the application using Go compiler with all the necessary flags
```

### ✅ DO: Document What, Not How

```makefile
# Good - describes the outcome
test: ## Run all tests

# Avoid - describes the implementation
test: ## Run go test ./... command
```

### ✅ DO: Use Inline Comments for Quick Reference

```makefile
build: ## Build the application
	go build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)
```

### ❌ DON'T: Mix `##` and `#` for the Same Target

```makefile
# Avoid this - inconsistent
# This is the build target
build: ## Build the application
	go build -o app main.go
```

Pick one style per target.

## Example: Well-Documented Makefile

```makefile
.PHONY: help build test lint clean deploy

.DEFAULT_GOAL := help

## Variables
BINARY_NAME := myapp
VERSION := 1.0.0
BUILD_DIR := ./bin

## Show this help message
help:
	@echo "Available targets:"
	@echo "  make build   - Build the application"
	@echo "  make test    - Run all tests"
	@echo "  make deploy  - Deploy to production"

build: ## Build the application
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

test: ## Run all tests with coverage
	go test -v -race -coverprofile=coverage.out ./...

lint: ## Run linters (golangci-lint)
	golangci-lint run ./...

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

## Deploy to production environment
deploy: build test
	@echo "Deploying to production..."
	./scripts/deploy.sh production
```

## How lazymake Displays Documentation

When you open this Makefile in lazymake:

1. **Target list** shows all targets with their `##` descriptions
2. **Help view** (`?` key) shows only documented targets (those with `##`)
3. **Color coding**:
   - `##` comments appear in cyan
   - `#` comments appear in gray
4. **Priority**: Inline `##` comments override preceding ones

## Tips for Teams

1. **Establish conventions**: Decide whether your team prefers inline or preceding `##` comments
2. **Document all public targets**: Any target a developer might run should have a `##` comment
3. **Skip internal targets**: Targets starting with `_` or `.` don't need documentation
4. **Update documentation**: Keep comments in sync with what the target actually does
5. **Use lazymake's help view**: Press `?` to see all documented targets at a glance

## Converting Existing Makefiles

To add documentation to an existing Makefile:

1. **Identify public targets**: What do developers actually run?
2. **Add `##` comments**: Start with the most-used targets
3. **Test with lazymake**: Run `lazymake` to see how it looks
4. **Iterate**: Refine descriptions based on team feedback

Example conversion:

```makefile
# Before
build:
	go build -o app main.go

# After
build: ## Build the application
	go build -o app main.go
```

---

[← Back to Documentation](../README.md) | [← Back to Main README](../../README.md)
