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
	var currentTargets []*Target
	var recipeLines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Empty line: commit and reset
		if trimmed == "" {
			commitCurrentTargets(currentTargets, recipeLines)
			currentTargets = nil
			recipeLines = nil
			lastComment = commentInfo{}
			continue
		}

		// Recipe line (starts with tab)
		if after, ok := strings.CutPrefix(line, "\t"); ok {
			if len(currentTargets) > 0 {
				recipeLines = append(recipeLines, after)
			}
			continue
		}

		// Check for comment
		if comment, commentType, found := parseCommentLine(trimmed); found {
			commitCurrentTargets(currentTargets, recipeLines)
			currentTargets = nil
			recipeLines = nil
			lastComment = commentInfo{
				text:        comment,
				commentType: commentType,
			}
			continue
		}

		// Check for target definition
		// Skip variable assignments (e.g., VAR := value, VAR = value, VAR ?= value, VAR += value)
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "\t") && !isVariableAssignment(line) {
			currentTargets = processTargetLine(
				line, &targets, currentTargets, recipeLines, lastComment)
			recipeLines = nil
			lastComment = commentInfo{}
		}
	}

	// Commit final targets
	commitCurrentTargets(currentTargets, recipeLines)

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Makefile: %w", err)
	}

	return targets, nil
}

// isVariableAssignment checks if a line is a variable assignment
// Makefile variable assignments use :=, =, ?=, or +=
func isVariableAssignment(line string) bool {
	// Check for common variable assignment operators
	// Note: We need to check := before : to avoid false positives
	return strings.Contains(line, ":=") ||
		strings.Contains(line, "?=") ||
		strings.Contains(line, "+=") ||
		// For simple = assignments, check that = appears before :
		// This handles "VAR = value" while allowing "target: dep = value" (shell assignment in recipe)
		(strings.Contains(line, "=") && strings.Index(line, "=") < strings.Index(line, ":"))
}

// commitCurrentTargets commits recipe lines to current targets
func commitCurrentTargets(currentTargets []*Target, recipeLines []string) {
	if len(currentTargets) > 0 {
		for _, target := range currentTargets {
			target.Recipe = recipeLines
		}
	}
}

// parseCommentLine checks if a line is a comment and extracts it
func parseCommentLine(trimmed string) (text string, commentType CommentType, found bool) {
	// Check for ## comment (takes priority)
	if comment, ok := strings.CutPrefix(trimmed, "##"); ok {
		return strings.TrimSpace(comment), CommentDouble, true
	}

	// Check for # comment
	if comment, ok := strings.CutPrefix(trimmed, "#"); ok {
		return strings.TrimSpace(comment), CommentSingle, true
	}

	return "", CommentNone, false
}

// processTargetLine processes a target definition line
func processTargetLine(line string, targets *[]Target, currentTargets []*Target,
	recipeLines []string, lastComment commentInfo) []*Target {
	// Commit previous targets
	commitCurrentTargets(currentTargets, recipeLines)

	parts := strings.SplitN(line, ":", 2)
	targetName := strings.TrimSpace(parts[0])

	// Skip special targets (like .PHONY, .SILENT, etc.)
	if strings.HasPrefix(targetName, ".") {
		return nil
	}

	// Extract inline comment and dependencies
	dependencies := parts[1]
	inlineComment := extractInlineComment(dependencies)

	// Clean dependencies string
	cleanDeps := dependencies
	if idx := strings.Index(dependencies, "#"); idx >= 0 {
		cleanDeps = dependencies[:idx]
	}
	depList := parseDependencies(cleanDeps)

	// Determine final description and comment type
	finalDesc := lastComment.text
	finalType := lastComment.commentType
	if inlineComment.text != "" {
		finalDesc = inlineComment.text
		finalType = inlineComment.commentType
	}

	// Add targets and track them
	startIdx := len(*targets)
	names := strings.FieldsSeq(targetName)
	for name := range names {
		*targets = append(*targets, Target{
			Name:         name,
			Description:  finalDesc,
			CommentType:  finalType,
			Dependencies: depList,
			Recipe:       nil,
		})
	}

	// Take pointers after all targets are added
	var newCurrentTargets []*Target
	for i := startIdx; i < len(*targets); i++ {
		newCurrentTargets = append(newCurrentTargets, &(*targets)[i])
	}

	return newCurrentTargets
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
