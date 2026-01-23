# Changelog

All notable changes to the lazymake project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-01-24

### Added

- Real-time streaming output during target execution (#20)
- Viewport for live output display with auto-scroll
- Keyboard navigation during execution (j/k scroll, ctrl+d/u half-page, g/G top/bottom)
- Cancel running execution with ctrl+c
- Example Makefile for testing streaming output (examples/streaming.mk)

### Fixed

- Skip define...endef blocks in Makefile parser (#21)

## [0.2.0] - 2026-01-10

### Added

- Contributor Covenant Code of Conduct
- 14 new dangerous command patterns to safety checker (AWS, Docker, Git, npm operations)
- Additional dangerous command patterns for improved safety detection
- Critical descriptions to safety rules for better user awareness
- Recipe matching to aws-s3-delete rule test case

### Changed

- Use GNU make order of makefiles for better compatibility
- Enhanced safety rules and tests with refined test cases
- Updated padding and border color in UI

### Fixed

- Use correct makefile path when executing make commands (#17)

## [0.1.0] - 2025-12-22

Initial release of lazymake - an interactive TUI for Makefiles.

### Added

- Interactive terminal UI for browsing Makefiles
- Target listing and execution
- Dependency graph visualization
- Variable inspection
- Safety checker for dangerous commands
- Support for multiple Makefile formats
- Keyboard shortcuts and navigation
- Search functionality
