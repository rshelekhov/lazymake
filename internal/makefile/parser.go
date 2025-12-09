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
	CommentNone CommentType = iota // No comment
	CommentSingle                   // # comment
	CommentDouble                   // ## comment (industry standard for documentation)
)

// Target represents a Makefile target
type Target struct {
	Name        string
	Description string
	CommentType CommentType
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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			lastComment = commentInfo{}
			continue
		}

		// Check for comment (## takes priority over #)
		if comment, found := strings.CutPrefix(trimmed, "##"); found {
			lastComment = commentInfo{
				text:        strings.TrimSpace(comment),
				commentType: CommentDouble,
			}
			continue
		} else if comment, found := strings.CutPrefix(trimmed, "#"); found {
			lastComment = commentInfo{
				text:        strings.TrimSpace(comment),
				commentType: CommentSingle,
			}
			continue
		}

		// Check for target definition
		// Recipe lines start with a tab, so check the original line, not trimmed
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "\t") {
			parts := strings.SplitN(line, ":", 2)
			targetName := strings.TrimSpace(parts[0])

			// Skip special targets
			if strings.HasPrefix(targetName, ".") || strings.Contains(targetName, "=") {
				lastComment = commentInfo{}
				continue
			}

			// Extract inline comment from the dependency/recipe section (after the colon)
			// Inline comments take priority over preceding comments
			dependencies := parts[1]
			inlineComment := extractInlineComment(dependencies)

			// Determine final description and comment type
			finalDesc := lastComment.text
			finalType := lastComment.commentType
			if inlineComment.text != "" {
				finalDesc = inlineComment.text
				finalType = inlineComment.commentType
			}

			// Handle multiple targets on one line
			names := strings.FieldsSeq(targetName)
			for name := range names {
				targets = append(targets, Target{
					Name:        name,
					Description: finalDesc,
					CommentType: finalType,
				})
			}

			lastComment = commentInfo{}
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
