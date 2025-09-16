# Lazymake - Development Plan

## Project Overview

**Lazymake** is an interactive TUI application built with Go and Bubbletea that provides a modern, user-friendly interface for running Makefile targets. It solves the common problem of "what targets are available in this Makefile?" by parsing Makefiles and presenting targets in an intuitive, searchable interface.

### Core Problem

Developers often work with complex Makefiles containing dozens of targets, but remembering all available targets and their purposes is difficult. Running `make` without arguments or manually reading Makefiles is inefficient.

### Solution

An interactive terminal UI that:

- Parses Makefiles automatically
- Shows all available targets with descriptions
- Provides search/filter functionality
- Executes selected targets with live output
- Maintains execution history

## Technical Stack

- **Language**: Go 1.21+
- **TUI Framework**: [Bubbletea](https://github.com/charmbracelet/bubbletea)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Additional Libraries**:
  - `cobra` - CLI argument parsing
  - `viper` - Configuration management
  - `testify` - Testing framework

## Project Architecture

```
makefile-runner/
├── cmd/
│   └── root.go              # CLI entry point
├── internal/
│   ├── makefile/
│   │   ├── parser.go        # Makefile parsing logic
│   │   ├── target.go        # Target data structures
│   │   └── parser_test.go   # Parser tests
│   ├── tui/
│   │   ├── model.go         # Bubbletea model
│   │   ├── update.go        # Update functions
│   │   ├── view.go          # View rendering
│   │   ├── commands.go      # Bubbletea commands
│   │   └── styles.go        # Lipgloss styles
│   ├── executor/
│   │   ├── runner.go        # Make command execution
│   │   └── output.go        # Output handling
│   └── config/
│       └── config.go        # Configuration management
├── testdata/
│   └── sample_makefiles/    # Test Makefiles
├── docs/
│   └── README.md
├── go.mod
├── go.sum
├── main.go
└── Makefile                 # Self-hosting example
```

## Development Phases

### Phase 1: MVP Implementation (Week 1-2)

#### 1.1 Project Setup

- [ ] Initialize Go module
- [ ] Set up basic CLI structure with Cobra
- [ ] Configure GitHub repository with proper README
- [ ] Set up basic CI/CD pipeline (GitHub Actions)

#### 1.2 Makefile Parser Implementation

```go
type Target struct {
    Name         string
    Description  string
    Dependencies []string
    Commands     []string
    Line         int
}

type Makefile struct {
    Path    string
    Targets []Target
}
```

**Tasks:**

- [ ] Implement basic Makefile parsing
- [ ] Extract target names and descriptions from comments
- [ ] Handle target dependencies
- [ ] Write comprehensive tests with sample Makefiles
- [ ] Error handling for malformed Makefiles

#### 1.3 Basic TUI Implementation

- [ ] Create Bubbletea model structure
- [ ] Implement target list view
- [ ] Add basic keyboard navigation (up/down arrows)
- [ ] Implement target selection and execution
- [ ] Basic styling with Lipgloss

#### 1.4 Command Execution

- [ ] Implement Make command runner
- [ ] Capture and display command output
- [ ] Handle command errors gracefully
- [ ] Add loading states during execution

### Phase 2: Enhanced Features (Week 3-4)

#### 2.1 Search and Filtering

- [ ] Implement real-time search functionality
- [ ] Filter targets by name/description
- [ ] Highlight search matches
- [ ] Clear search functionality

#### 2.2 Improved UI/UX

- [ ] Add help screen with keyboard shortcuts
- [ ] Implement target descriptions panel
- [ ] Show target dependencies
- [ ] Add status indicators (running, success, error)
- [ ] Improve color scheme and styling

#### 2.3 Configuration System

- [ ] Create configuration file structure (.makerunner.yaml)
- [ ] Configurable keybindings
- [ ] Theme customization
- [ ] Default directory settings

### Phase 3: Advanced Features (Week 5-6)

#### 3.1 History and Bookmarks

- [ ] Track execution history
- [ ] Bookmark frequently used targets
- [ ] History navigation
- [ ] Persistent storage for history/bookmarks

#### 3.2 Multi-Makefile Support

- [ ] Detect Makefiles in subdirectories
- [ ] Makefile switching interface
- [ ] Hierarchical target display
- [ ] Project context awareness

#### 3.3 Enhanced Target Information

- [ ] Parse and display target dependencies visually
- [ ] Show estimated execution time
- [ ] Display last execution status
- [ ] Target categorization/grouping

### Phase 4: Polish and Distribution (Week 7-8)

#### 4.1 Testing and Quality Assurance

- [ ] Achieve 80%+ test coverage
- [ ] Integration tests with real Makefiles
- [ ] Performance testing with large Makefiles
- [ ] Cross-platform testing (Linux, macOS, Windows)

#### 4.2 Documentation

- [ ] Comprehensive README with GIFs/screenshots
- [ ] Usage examples and tutorials
- [ ] Contributing guidelines
- [ ] Changelog maintenance

#### 4.3 Distribution

- [ ] GitHub Releases with binaries
- [ ] Homebrew formula
- [ ] Go module documentation
- [ ] Package for major Linux distributions

## Implementation Details

### Makefile Parser Specifications

**Target Detection Patterns:**

```makefile
# Standard target
target-name: dependencies
	@echo "Running target"

# Target with description
.PHONY: build ## Build the application
build:
	go build -o bin/app

# Target with multiple dependencies
test: deps lint ## Run tests
	go test ./...
```

**Parsing Rules:**

1. Target lines start at column 0, end with `:`
2. Comments starting with `##` are descriptions
3. Commands are indented with tabs
4. `.PHONY` declarations should be tracked
5. Variable substitutions should be preserved

### TUI State Management

**Application States:**

- `StateLoading` - Initial Makefile parsing
- `StateList` - Target list display
- `StateSearch` - Search mode active
- `StateExecuting` - Running make command
- `StateOutput` - Showing command output
- `StateError` - Error display

**Key Bindings:**

- `j/k` or `↓/↑` - Navigate targets
- `/` - Enter search mode
- `Enter` - Execute selected target
- `Esc` - Cancel/go back
- `q` - Quit application
- `?` - Show help
- `r` - Refresh Makefile
- `h` - Show history

### Error Handling Strategy

1. **Graceful Degradation**: If Makefile parsing fails, show available information
2. **User Feedback**: Clear error messages with suggestions
3. **Recovery**: Allow manual Makefile reload
4. **Logging**: Optional debug logging for troubleshooting

### Configuration File Structure

```yaml
# ~/.config/makerunner/config.yaml
keybindings:
  quit: "q"
  search: "/"
  execute: "enter"
  help: "?"

appearance:
  theme: "default"
  show_dependencies: true
  compact_view: false

behavior:
  auto_refresh: true
  save_history: true
  max_history: 100
  confirm_execution: false

directories:
  makefile_paths:
    - "."
    - "./build"
    - "./scripts"
```

## Testing Strategy

### Unit Tests

- Makefile parser with various edge cases
- TUI component testing
- Command execution mocking
- Configuration loading/saving

### Integration Tests

- End-to-end workflows
- Real Makefile parsing
- Cross-platform command execution

### Test Data

Create sample Makefiles covering:

- Simple targets
- Complex dependencies
- Variable usage
- Multi-line commands
- Edge cases (empty targets, comments)

## Performance Considerations

1. **Lazy Loading**: Parse Makefiles on demand
2. **Caching**: Cache parsed results until file changes
3. **Efficient Rendering**: Minimize terminal redraws
4. **Memory Management**: Limit history size and output buffering

## Future Enhancements (Post-MVP)

### Advanced Features

- [ ] Make variable inspection and override
- [ ] Parallel target execution
- [ ] Target execution time tracking
- [ ] Integration with task runners (npm scripts, etc.)
- [ ] Plugin system for custom parsers

### IDE Integration

- [ ] VS Code extension
- [ ] Vim/Neovim plugin
- [ ] Language server protocol support

### Web Interface

- [ ] Optional web UI for team environments
- [ ] Real-time execution sharing
- [ ] Makefile documentation generation

## Success Metrics

### Technical Metrics

- Parse 95% of real-world Makefiles correctly
- Handle Makefiles with 100+ targets efficiently
- Start-up time under 100ms
- Memory usage under 50MB

### User Adoption

- 100+ GitHub stars in first month
- Positive feedback from Go/development communities
- Integration requests from other projects
- Contribution from external developers

## Risk Mitigation

### Technical Risks

- **Complex Makefile syntax**: Start with common patterns, expand gradually
- **Cross-platform compatibility**: Test early and often on all platforms
- **Performance with large Makefiles**: Implement pagination and lazy loading

### Project Risks

- **Scope creep**: Stick to MVP first, then iterate
- **Abandoned dependencies**: Choose well-maintained libraries
- **User adoption**: Focus on solving real problems developers face daily

## Development Workflow

### Daily Workflow

1. Start with tests for new features
2. Implement core functionality
3. Add TUI integration
4. Test with real Makefiles
5. Update documentation

### Weekly Milestones

- Week 1: Basic parsing and TUI
- Week 2: Target execution and search
- Week 3: Polish and configuration
- Week 4: Testing and distribution

### Git Strategy

- `main` branch for stable releases
- `develop` branch for active development
- Feature branches for new functionality
- Conventional commits for clear history

---

## Getting Started

### Prerequisites

- Go 1.21 or later
- Make (for testing with real Makefiles)
- Git

### Initial Setup

```bash
# Clone repository
git clone https://github.com/yourusername/makefile-runner
cd makefile-runner

# Initialize Go module
go mod init github.com/yourusername/makefile-runner

# Install dependencies
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/spf13/cobra@latest

# Create basic structure
mkdir -p {cmd,internal/{makefile,tui,executor,config},testdata/sample_makefiles}

# Start development
go run main.go
```

### First Steps

1. Implement basic Makefile parser
2. Create simple target list TUI
3. Add target execution capability
4. Test with project's own Makefile
5. Iterate based on personal usage

Remember: Start simple, get feedback early, and iterate quickly. The goal is to create a tool that you and other developers will actually want to use daily.
