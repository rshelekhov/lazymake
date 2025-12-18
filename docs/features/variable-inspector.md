# Variable Inspector

lazymake helps you understand and track Makefile variables, making it easy to see what values are used and where they come from.

## Two Ways to View Variables

### 1. Full-Screen Inspector (Press `v`)

Browse all variables in your Makefile with detailed information:

```
┌─ Variable Inspector ──────────────────────────────────────┐
│                                                            │
│ 6 variables • 4 used • 2 unused                           │
│                                                            │
│ BINARY_NAME          [:=]  Simply Expanded                │
│   Raw:      lazymake                                       │
│   Expanded: lazymake                                       │
│   Used by:  build, install (2 targets)                    │
│                                                            │
│ VERSION              [=]   Recursive                       │
│   Raw:      1.0.0                                          │
│   Expanded: 1.0.0                                          │
│   Used by:  build (1 target)                              │
│                                                            │
│ GOFLAGS              [=]   Recursive                       │
│   Raw:      -v -race                                       │
│   Expanded: -v -race                                       │
│   Used by:  build, test (2 targets)                       │
│                                                            │
│ LDFLAGS              [=]   Recursive                       │
│   Raw:      -ldflags "-X main.version=$(VERSION)"         │
│   Expanded: -ldflags "-X main.version=1.0.0"              │
│   Used by:  build (1 target)                              │
│                                                            │
│ BUILD_DIR            [?=]  Conditional                     │
│   Raw:      ./bin                                          │
│   Expanded: ./bin                                          │
│   Used by:  build, clean (2 targets)                      │
│                                                            │
│ PATH                                                       │
│   Exported to environment                                  │
│   Not used by any target                                   │
│                                                            │
└────────────────────────────────────────────────────────────┘
  v/esc: return • ↑↓/j/k: navigate • q: quit
```

### 2. Context Panel (Automatic)

When you select a target, the recipe preview shows variables it uses:

```
┌─────────────────────┬────────────────────────────────────────┐
│ ALL TARGETS         │ build:                                 │
│ > build             │                                        │
│   test              │   go build $(GOFLAGS) $(LDFLAGS) \    │
│   clean             │     -o $(BUILD_DIR)/$(BINARY_NAME)    │
│   install           │                                        │
│                     │   💡 Press 'g' to view full graph      │
│                     │                                        │
│                     │   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━   │
│                     │                                        │
│                     │   📦 Variables Used                    │
│                     │                                        │
│                     │     GOFLAGS = -v -race                 │
│                     │     LDFLAGS = -ldflags "-X main..."    │
│                     │     BUILD_DIR = ./bin                  │
│                     │     BINARY_NAME = lazymake             │
│                     │                                        │
│                     │     💡 Press 'v' to view all variables │
└─────────────────────┴────────────────────────────────────────┘
```

## Variable Types Explained

lazymake recognizes all Makefile variable assignment operators:

- **`=` Recursive**: Expanded when used (can reference later variables)
- **`:=` Simply Expanded**: Expanded when defined (like shell variables)
- **`+=` Append**: Adds to existing value
- **`?=` Conditional**: Sets only if not already defined
- **`!=` Shell**: Executes shell command and captures output

## How It Works

1. **Parse Definitions**: Extracts variable assignments from Makefile text
2. **Expand Values**: Runs `make --print-data-base` to get fully expanded values
3. **Track Usage**: Scans all target recipes to find variable references
4. **Display Context**: Shows raw vs expanded values and which targets use them

## Example Makefile

```makefile
# Variable definitions
BINARY_NAME := lazymake
VERSION = 1.0.0
GOFLAGS = -v -race
LDFLAGS = -ldflags "-X main.version=$(VERSION)"
BUILD_DIR ?= ./bin

export PATH

build: ## Build the application
	go build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)

test: ## Run tests
	go test $(GOFLAGS) ./...

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
```

## Benefits

- **Understand configuration**: See what values are actually used
- **Debug issues**: Compare raw vs expanded values to spot errors
- **Track dependencies**: Know which targets are affected by variable changes
- **Onboarding**: New developers can understand build configuration instantly
- **Environment awareness**: Identify which variables are exported

## Navigation

- **`v`**: Open variable inspector from list view
- **`↑/↓` or `j/k`**: Navigate between variables
- **`v` or `esc`**: Return to list view
- **Auto-scroll**: Inspector automatically scrolls to keep selected variable visible

---

[← Back to Documentation](../README.md) | [← Back to Main README](../../README.md)
