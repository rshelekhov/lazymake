# Changelog

All notable changes to the lazymake project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-02-20

### Added

- Fish shell history integration: auto-detection, native `- cmd:` / `when:` format, timestamp support via `include_timestamp`
- Support `{makefile}` and `{dir}` template variables in `shell_integration.format_template`
- Honor `shell_integration.include_timestamp` in runtime behavior: when `false`, always writes plain history entries even if zsh extended history format is detected
- Size-based export rotation via `export.max_file_size_mb` — files exceeding the limit are removed during rotation (previously the config field existed but was never used)
- Regression tests for config defaults parity between code and docs

### Changed

- Unified config merge behavior for `export`, `shell_integration`, and `safety` sections: global (`~/.lazymake.yaml`) and project (`.lazymake.yaml`) configs are now merged consistently — scalars use project-overrides-global, string slices are unioned and deduplicated, struct slices (e.g. `custom_rules`) are appended
- Safety config loading moved from `internal/safety` to centralized `config` package; `Config.Safety` field added to main config struct

### Fixed

- Documentation now matches code defaults for export rotation fields (`max_file_size_mb`, `max_files`, `keep_days` all default to `0`/disabled, not `10`/`50`/`30`)
- Safety rules count updated from 11 to 36 in all docs; complete rule list added
- `makefile` config default documented as auto-detect (`GNUmakefile` → `makefile` → `Makefile`) instead of hardcoded `"Makefile"`
- Config merging docs expanded to cover all sections (`safety`, `export`, `shell_integration`), not just `safety`

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
