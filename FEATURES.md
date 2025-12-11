## Feature Ideas from Research

### Market Context
- Make dominates with 19% presence in top GitHub repos
- Developers describe Make as "fragile, dated, and anti-human by modern dev ergonomics"
- 70% of development teams struggle with managing dependencies
- A TUI that makes Make more approachable addresses a real need

### High-Impact Features (Priority)

#### ✅ 1. Self-Documenting Help System
- Extract and display inline comments from Makefile targets (e.g., `## This builds the project`)
- Show target descriptions in a sidebar or help panel
- Addresses biggest pain point: poor onboarding experience for new developers

#### ✅ 2. Dependency Graph Visualization
- Visual representation of target dependencies with ASCII tree structure
- Shows execution order numbering `[N]` indicating which targets run first
- Highlights critical path `★` - the longest dependency chain determining minimum build time
- Identifies parallel opportunities `∥` - targets that can run concurrently with `make -j`
- Configurable depth control to show more or fewer dependency levels
- Smart detection: only marks meaningful build chains, not standalone targets
- Cycle detection with clear warnings for circular dependencies
- Addresses: over 60% of compilation delays stem from misconfigured dependencies
- **Implementation complete**: Uses DFS for cycle detection, Kahn's algorithm for topological sort, and memoized depth calculation for critical path analysis

#### ✅ 3. Search & Filtering
- Real-time fuzzy search for targets (type `/` to filter by name or description)
- Filter by recently used targets with visual ⏱ indicator
- Last 5 executed targets per Makefile appear at top of list for instant access
- Per-project history stored in `~/.cache/lazymake/history.json`
- Automatic cleanup of stale targets (removed from Makefile)
- Smart navigation: arrow keys skip over section headers and separators
- Essential for developer productivity, especially for repetitive workflows
- **Implementation complete**: LRU cache with graceful degradation, persistent across sessions
- **Note**: Favorites and tags were skipped in favor of simplicity - fuzzy search + recent history covers most use cases

### UI/UX Design Decisions

#### Two-Column Layout + Status Bar (Hybrid Approach)
**Decision**: Use two-column layout with bottom status bar for main interface
- **Left column (30-40%)**: Target list with emoji indicators (🚨⚠️ for dangerous commands)
  - Recent targets section
  - Fast scanning and navigation
  - Filtering support
- **Right column (60-70%)**: Recipe preview pane
  - Shows commands for selected target
  - Safety warnings with severity, description, and suggestions (current implementation)
  - Future: Syntax-highlighted multi-language recipes (feature #6)
- **Bottom status bar**: Dynamic contextual information
  - Stats: target count, dangerous count (e.g., "12 targets • 2 dangerous")
  - OR contextual shortcuts that change based on selection
  - Avoids redundancy - shows actionable info, not duplicate labels

**Rationale**:
- Proven pattern in successful TUI apps (glow, television, kanban, Bagels)
- Provides detail without overwhelming the list view
- Progressive enhancement: basic preview now, syntax highlighting later
- Right pane serves dual purpose: safety info today, syntax highlighting tomorrow
- Clean separation of concerns: navigation (left) + details (right) + actions (bottom)

**Inspired by**: Glow (color palette and design), television (preview pane), and other Bubble Tea applications

#### 4. Performance Profiling
- Show execution time for each target
- Highlight slow targets for optimization
- Track build time history and trends
- Process overhead can account for up to 15% of build time

#### 5. Better Error Handling
- Parse and highlight common Makefile errors (missing separator, undefined variables)
- Show context around errors with suggestions
- The "missing separator" error is one of the most frustrating issues developers face

### Nice-to-Have Features

#### 6. Multi-Language Recipe Support
- Detect and syntax-highlight embedded languages (bash, python, etc.) in targets
- Inspired by Just's ability to write recipes in any language

#### 7. Workspace/Project Management
- Quick switching between different project Makefiles
- Remember last-used targets per project
- Support monorepo scenarios

#### 8. CI/CD Preview
- Show which targets would run in CI
- Simulate CI environment locally
- Companies with CI/CD report 30-50% increase in deployment frequency

#### 9. Variable Inspector
- Display all Makefile variables and their values
- Show where variables are defined and used
- Allow runtime variable override through TUI

#### 10. Dry-run Preview
- Show what commands will execute before running (make -n)
- Estimated execution time based on history
- Warning for destructive operations (clean, rm, etc.)

#### 11. Smart Keyboard Shortcuts
- Number keys (1-9) for quick access to frequent targets
- Customizable shortcuts per project
- Common in successful TUI applications

#### 12. Integration Features
- Export execution results to JSON/log files
- Watch mode: auto-run targets on file changes
- Shell integration for command history

### Research Sources
- [Makefile Success Stories from Open Source Projects](https://moldstud.com/articles/p-unlocking-makefiles-success-stories-from-leading-open-source-projects)
- [Make: The Ultimate DevOps Automation Tool in 2025](https://suyashbhawsar.com/make-in-2025-the-devops-secret-weapon-you-already-have)
- [Future Trends in Makefiles for Modern Software Development](https://moldstud.com/articles/p-the-future-of-makefiles-evolving-practices-for-modern-developers)
- [Improving Makefile usability for new developers](https://testdouble.com/insights/makefile-usability-tips-for-new-developers)
- [Makefiles are older than Doom why are we still using them?](https://dev.to/dev_tips/makefiles-are-older-than-doom-why-are-we-still-using-them-35jl)
- [Makefile Profiling Tools to Improve Your Development Workflow](https://moldstud.com/articles/p-boost-your-development-workflow-essential-makefile-profiling-tools-you-need-to-know)
- [Just vs. Make: Which Task Runner Stands Up Best?](https://spin.atomicobject.com/2022/09/27/just-task-runner/)
- [Task Runner Census 2025](https://aleyan.com/blog/2025-task-runners-census/)
- [Justfile became my favorite task runner](https://tduyng.com/blog/justfile-my-favorite-task-runner/)
- [taskwarrior-tui: A terminal user interface for taskwarrior](https://github.com/kdheepak/taskwarrior-tui)
