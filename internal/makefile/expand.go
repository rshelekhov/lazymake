package makefile

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPatternTargets expands pattern targets (e.g., build-%) by scanning the project
// for matching directories (e.g., src/cmd/*-gui/) and generating concrete targets.
// Returns the original targets with pattern targets replaced/augmented with concrete ones.
func ExpandPatternTargets(targets []Target, makefilePath string) ([]Target, error) {
	// Get the directory containing the Makefile
	makefileDir := filepath.Dir(makefilePath)

	// Find all pattern targets and their expansions
	var result []Target

	for _, target := range targets {
		if !target.IsPatternRule {
			// Non-pattern targets are kept as-is
			result = append(result, target)
			continue
		}

		// This is a pattern target - try to expand it
		expanded := expandSinglePattern(target, makefileDir)
		if len(expanded) > 0 {
			// Add all expanded targets
			result = append(result, expanded...)
		} else {
			// Keep the pattern target as-is if no expansion found
			result = append(result, target)
		}
	}

	return result, nil
}

// expandSinglePattern expands a single pattern target by scanning for matching directories
func expandSinglePattern(target Target, makefileDir string) []Target {
	name := target.Name

	// Determine the directory pattern to look for
	dirPattern := detectDirPattern(name)
	if dirPattern == "" {
		return nil
	}

	// Find matching directories
	pattern := filepath.Join(makefileDir, dirPattern)
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return nil
	}

	// Build expanded targets from matching directories
	var expanded []Target
	for _, match := range matches {
		expanded = appendPatternMatch(expanded, target, match, name)
	}

	return expanded
}

// detectDirPattern determines which directory pattern to scan based on target name
func detectDirPattern(name string) string {
	switch {
	case strings.HasSuffix(name, "-%"):
		return "src/cmd/*-gui"
	case strings.HasSuffix(name, "-%-tui"):
		return "src/cmd/*-tui"
	case strings.HasSuffix(name, "%-tui"):
		return "src/cmd/*-tui"
	case strings.Contains(name, "%"):
		return "src/cmd/*"
	default:
		return ""
	}
}

// appendPatternMatch processes a single match and adds expanded target
func appendPatternMatch(expanded []Target, target Target, match string, patternName string) []Target {
	info, err := os.Stat(match)
	if err != nil || !info.IsDir() {
		return expanded
	}

	dirName := filepath.Base(match)
	baseName := detectBaseName(dirName)
	if baseName == "" {
		return expanded
	}

	newTargetName := buildTargetName(patternName, baseName)
	if newTargetName == "" {
		return expanded
	}

	return append(expanded, Target{
		Name:         newTargetName,
		Description: target.Description,
		CommentType:  target.CommentType,
		Dependencies: target.Dependencies,
		Recipe:     target.Recipe,
		IsPatternRule: false,
	})
}

// detectBaseName extracts app name from directory name
func detectBaseName(dirName string) string {
	switch {
	case strings.HasSuffix(dirName, "-gui"):
		return strings.TrimSuffix(dirName, "-gui")
	case strings.HasSuffix(dirName, "-tui"):
		return strings.TrimSuffix(dirName, "-tui")
	default:
		return dirName
	}
}

// buildTargetName creates the target name from pattern and base name
func buildTargetName(pattern, baseName string) string {
	patternLen := len(pattern)
	switch {
	case strings.HasSuffix(pattern, "%"):
		return pattern[:patternLen-1] + baseName
	case strings.HasSuffix(pattern, "%-tui"):
		return pattern[:patternLen-len("%-tui")] + baseName + "-tui"
	case strings.Contains(pattern, "%"):
		idx := strings.Index(pattern, "%")
		if idx >= 0 {
			return pattern[:idx] + baseName
		}
		return ""
	default:
		return ""
	}
}