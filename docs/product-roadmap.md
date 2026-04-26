# Product Roadmap

This document translates the current state of lazymake into a product roadmap.
It is intentionally opinionated: the goal is not to list every possible idea, but to prioritize the work that most increases product value.

## Current Product Position

lazymake is already more than a target picker:

- It helps users discover Makefile targets quickly
- It explains execution through graphs, recipe previews, and variable inspection
- It reduces risk through dangerous command detection
- It improves repeat workflows with history, performance tracking, workspace switching, export, and shell integration

In product terms, the core promise is:

> "Make unfamiliar or large Makefile-based workflows easier to understand, safer to run, and faster to repeat."

That promise is strong, but there are still two gaps:

1. **Execution control gap**
Users can run targets, but they still lack pre-run controls such as dry-run, runtime flags, environment variables, and reusable presets.

2. **Compatibility gap**
The parser is still mostly static and line-based. That is enough for many Makefiles, but complex real-world Makefiles will be the main adoption ceiling.

## Product Strategy

The next stage of lazymake should focus on four layers, in order:

1. **Safe Execution**
Make execution more previewable, parameterized, and repeatable.

2. **Compatibility**
Support more real-world Makefiles without surprising omissions or incorrect graphs.

3. **Team and CI Workflows**
Make lazymake useful not only in the TUI, but also in docs, onboarding, and automation.

4. **Intelligence**
Turn historical usage and execution data into meaningful guidance, not just storage.

## Roadmap Principles

- Prioritize features that reduce fear and ambiguity before features that add novelty
- Prefer workflows that fit existing `make` habits over introducing custom abstractions
- Treat parser accuracy as a growth bottleneck, not just a technical nicety
- Make every new execution feature testable and explicit in the UI
- Keep graceful degradation as a product rule

## Detailed Roadmap

### Phase 1: `v0.5` Safe Execution

**Theme:** "Understand before you run, and reuse what works."

**Why this phase first**

This is the highest-leverage product move. Users already trust lazymake enough to browse targets. The next step is to make it the preferred place to launch them.

**Primary outcomes**

- Users can preview target behavior before execution
- Users can pass temporary runtime context safely
- Users can quickly repeat known-good executions
- Repeated workflows become a first-class product concept

**Planned features**

1. Dry-run / explain mode
2. Interactive execution parameters (`ENV` variables and make flags)
3. Saved presets per target and Makefile
4. Rerun last execution with previous parameters
5. TUI flow tests for execution variants and side-effect behavior

**Definition of done for the phase**

- Dry-run is visible, safe, and does not pollute history/export/shell integration
- Parameter entry is deterministic and cancelable
- Presets survive restart and are easy to reapply
- Re-run is a one-keystroke path for common workflows
- Core execution flows have automated tests

**Product risks**

- Added pre-run UI can make the TUI feel heavier
- Too much flexibility in flags/env can create confusing edge cases

**Mitigation**

- Keep default Enter behavior fast
- Make advanced options opt-in
- Use explicit view labels and confirmation text

### Phase 2: `v0.6` Compatibility

**Theme:** "If the Makefile is real, lazymake should still be useful."

**Why this phase second**

Once users depend on lazymake for execution, parser limitations become much more visible. Compatibility work unlocks broader adoption in monorepos and mature build systems.

**Primary outcomes**

- Better target discovery and dependency accuracy on complex Makefiles
- Fewer incorrect graphs, missing targets, and misleading previews
- Clearer trust model when static parsing is incomplete

**Planned features**

1. Parser v2 using a hybrid model with `make` metadata
2. Compatibility fixture suite for complex Makefiles
3. Parser confidence warnings and unsupported-construct hints
4. More robust support for `include`, conditionals, pattern rules, and target-specific variables

**Definition of done for the phase**

- Parser v2 improves compatibility on a representative fixture corpus
- Failures degrade clearly instead of silently misleading users
- Compatibility regressions are caught in CI

**Product risks**

- External `make` inspection can be slower or environment-sensitive
- Some Makefiles remain impossible to model perfectly without execution

**Mitigation**

- Keep a fallback static path
- Measure parse time and expose warnings when needed

### Phase 3: `v0.7` Team and CI

**Theme:** "Make lazymake outputs reusable outside the TUI."

**Why this phase third**

After execution and compatibility are stronger, the next best move is to let the tool serve documentation, onboarding, and automation use cases.

**Primary outcomes**

- Graphs and target metadata can move into docs and CI artifacts
- Teams can standardize common executions
- lazymake becomes useful in headless workflows too

**Planned features**

1. Export dependency graph to DOT and Mermaid
2. Headless CLI output for targets, variables, and graph metadata in machine-readable formats
3. Preset import/export for team sharing
4. Copy/share flows for commands and graph views

**Definition of done for the phase**

- Teams can generate graph artifacts without opening the TUI
- CI and docs use cases are documented and stable
- Presets can be shared without manual file surgery

**Product risks**

- CLI scope can sprawl beyond the product's center

**Mitigation**

- Only add headless commands that reinforce the same understanding and execution model

### Phase 4: `v0.8` Intelligence

**Theme:** "Use history to guide better decisions."

**Why this phase last**

This phase compounds value from previous work. It becomes much stronger once executions are parameterized, repeated, and better modeled.

**Primary outcomes**

- Users can see what changed between runs
- Slow or risky targets become easier to notice
- The product starts surfacing actionable advice

**Planned features**

1. Run diff between current and previous executions
2. Smarter performance insights and trend views
3. Recommendations around parallelism and likely bottlenecks
4. Richer target grouping such as frequent, risky, slow, and recently failed

**Definition of done for the phase**

- Historical data helps users make faster choices, not just inspect logs
- Insights are visible but non-intrusive

## Suggested Milestone Map

### `v0.5`

- Dry-run / explain mode
- Execution parameters
- Saved presets
- Rerun last
- TUI flow tests

### `v0.6`

- Parser v2
- Compatibility fixtures
- Parser confidence UI
- Extended Makefile construct support

### `v0.7`

- Graph export to DOT and Mermaid
- Headless CLI outputs
- Preset sharing
- Command and graph sharing

### `v0.8`

- Run diff
- Richer performance analytics
- Parallelism suggestions
- Smart target grouping

## Success Metrics

These metrics are the most useful ones to watch after each phase:

- Share of executions initiated via lazymake instead of manual shell commands
- Reuse rate of presets and rerun flows
- Parser compatibility rate on fixture corpus
- Number of successful graph exports / headless CLI use cases
- Reduction in failed or canceled dangerous executions
- Median time from app launch to target execution

## What Not to Prioritize Yet

- Full Makefile editing inside the TUI
- A plugin ecosystem before the core execution and compatibility story matures
- Heavy visual redesign without a product problem to solve
- Replacing `make` semantics with custom abstractions

## Product Owner Recommendation

If only three bets can be funded next, they should be:

1. Dry-run / explain mode
2. Interactive execution parameters plus presets
3. Parser v2 with a compatibility fixture suite

Together, those three move lazymake from a strong interface into a dependable execution surface for serious projects.
