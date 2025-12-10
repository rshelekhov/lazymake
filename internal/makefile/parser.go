package makefile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CommentType represents the type of comment found
type CommentType int

const (
	CommentNone   CommentType = iota // No comment
	CommentSingle                    // # comment
	CommentDouble                    // ## comment (industry standard for documentation)
)

// Target represents a Makefile target
type Target struct {
	Name         string
	Description  string
	CommentType  CommentType
	Dependencies []string // List of target names this target depends on
	Recipe       []string // Recipe lines (commands to execute)
}

// commentInfo holds information about a comment
type commentInfo struct {
	text        string
	commentType CommentType
}

func Parse(filename string) ([]Target, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Makefile: %w", err)
	}

	defer file.Close()

	var targets []Target
	var lastComment commentInfo
	var currentTargets []*Target // Track current targets for recipe accumulation (multi-target support)
	var recipeLines []string     // Accumulate recipe lines

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Empty line: commit current targets and reset
		if trimmed == "" {
			if len(currentTargets) > 0 {
				for _, target := range currentTargets {
					target.Recipe = recipeLines
				}
				recipeLines = nil
				currentTargets = nil
			}
			lastComment = commentInfo{}
			continue
		}

		// Recipe line (starts with tab)
		if after, ok := strings.CutPrefix(line, "\t"); ok {
			if len(currentTargets) > 0 {
				// Remove leading tab, keep the command
				recipeLines = append(recipeLines, after)
			}
			continue
		}

		// Check for comment (## takes priority over #)
		if comment, found := strings.CutPrefix(trimmed, "##"); found {
			// Commit current targets before processing new comment
			if len(currentTargets) > 0 {
				for _, target := range currentTargets {
					target.Recipe = recipeLines
				}
				recipeLines = nil
				currentTargets = nil
			}

			lastComment = commentInfo{
				text:        strings.TrimSpace(comment),
				commentType: CommentDouble,
			}
			continue
		} else if comment, found := strings.CutPrefix(trimmed, "#"); found {
			// Commit current targets before processing new comment
			if len(currentTargets) > 0 {
				for _, target := range currentTargets {
					target.Recipe = recipeLines
				}
				recipeLines = nil
				currentTargets = nil
			}

			lastComment = commentInfo{
				text:        strings.TrimSpace(comment),
				commentType: CommentSingle,
			}
			continue
		}

		// Check for target definition
		// Recipe lines start with a tab, so check the original line, not trimmed
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "\t") {
			// Commit previous targets before starting new ones
			if len(currentTargets) > 0 {
				for _, target := range currentTargets {
					target.Recipe = recipeLines
				}
				recipeLines = nil
			}

			parts := strings.SplitN(line, ":", 2)
			targetName := strings.TrimSpace(parts[0])

			// Skip special targets
			if strings.HasPrefix(targetName, ".") || strings.Contains(targetName, "=") {
				lastComment = commentInfo{}
				currentTargets = nil
				continue
			}

			// Extract inline comment from the dependency/recipe section (after the colon)
			// Inline comments take priority over preceding comments
			dependencies := parts[1]
			inlineComment := extractInlineComment(dependencies)

			// Extract the dependency list
			// Need to clean the dependencies string by removing comments first
			// Example: "deps compile ## Build" -> want just "deps compile"
			cleanDeps := dependencies
			if idx := strings.Index(dependencies, "#"); idx >= 0 {
				// Remove everything from the first # onwards (the comment part)
				cleanDeps = dependencies[:idx]
			}
			// Parse the cleaned dependency string to get individual target names
			depList := parseDependencies(cleanDeps)

			// Determine final description and comment type
			finalDesc := lastComment.text
			finalType := lastComment.commentType
			if inlineComment.text != "" {
				finalDesc = inlineComment.text
				finalType = inlineComment.commentType
			}

			// Handle multiple targets on one line
			// All targets on the same line share the same recipe
			currentTargets = nil
			startIdx := len(targets) // Remember where we started adding targets
			names := strings.FieldsSeq(targetName)
			for name := range names {
				targets = append(targets, Target{
					Name:         name,
					Description:  finalDesc,
					CommentType:  finalType,
					Dependencies: depList,
					Recipe:       nil, // Will be populated as we read recipe lines
				})
			}

			// Now take pointers AFTER all targets are added (avoids pointer invalidation during slice growth)
			for i := startIdx; i < len(targets); i++ {
				currentTargets = append(currentTargets, &targets[i])
			}

			lastComment = commentInfo{}
			recipeLines = nil
		}
	}

	// Commit final targets if any
	if len(currentTargets) > 0 {
		for _, target := range currentTargets {
			target.Recipe = recipeLines
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Makefile: %w", err)
	}

	return targets, nil
}

// extractInlineComment extracts a comment from the dependencies/recipe section
// Prioritizes ## over # comments
func extractInlineComment(dependencies string) commentInfo {
	// Check for ## first (industry standard for documentation)
	if idx := strings.Index(dependencies, "##"); idx >= 0 {
		return commentInfo{
			text:        strings.TrimSpace(dependencies[idx+2:]),
			commentType: CommentDouble,
		}
	}

	// Check for single # (backward compatibility)
	if idx := strings.Index(dependencies, "#"); idx >= 0 {
		return commentInfo{
			text:        strings.TrimSpace(dependencies[idx+1:]),
			commentType: CommentSingle,
		}
	}

	return commentInfo{}
}

// parseDependencies extracts dependency target names from the dependency section
// of a Makefile target line (everything after the colon but before any comment)
//
// The function handles several edge cases:
// - Variables like $VAR or $(VAR) are skipped (Make expands these at runtime)
// - Pattern rules like %.o are skipped (these are templates, not concrete targets)
// - Order-only prerequisites (after |) are separated out
//
// Example inputs and outputs:
//
//	"deps compile"           -> ["deps", "compile"]
//	"deps | order-only"      -> ["deps"]           (ignores order-only)
//	"$VAR target"            -> ["target"]         (skips variable)
//	"%.o: %.c"               -> []                 (skips pattern rule)
func parseDependencies(depStr string) []string {
	if depStr == "" {
		return nil
	}

	trimmed := strings.TrimSpace(depStr)
	if trimmed == "" {
		return nil
	}

	// Handle order-only prerequisites: "normal-deps | order-only-deps"
	// We only care about normal dependencies (before the pipe) for visualization
	// Order-only deps are used to control when things rebuild, but don't affect
	// what needs to run before this target
	if idx := strings.Index(trimmed, "|"); idx >= 0 {
		trimmed = trimmed[:idx] // Keep only the part before |
	}

	// Split by whitespace to get individual dependency names
	fields := strings.Fields(trimmed)

	var deps []string
	for _, field := range fields {
		// Skip Makefile variables (start with $)
		// Example: $(OBJS) or $VAR
		// We can't resolve these without running Make, so we skip them
		if strings.HasPrefix(field, "$") {
			continue
		}

		// Skip pattern rules (contain %)
		// Example: %.o in "%.o: %.c"
		// Pattern rules are templates, not actual targets we can visualize
		if strings.Contains(field, "%") {
			continue
		}

		// Skip file paths (likely files, not targets)
		// This is a heuristic - Makefiles CAN have targets with /, but it's uncommon
		// We filter out:
		//   - Paths with multiple slashes: src/pkg/main.go
		//   - Paths with file extensions: main.go, src/main.go
		if strings.Count(field, "/") > 1 {
			continue
		}

		// Check for file extensions (contains / and has extension like .go, .o, .c)
		if strings.Contains(field, "/") && strings.Contains(field, ".") {
			continue
		}

		deps = append(deps, field)
	}

	return deps
}
