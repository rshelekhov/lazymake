package makefile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Target represents a Makefile target
type Target struct {
	Name        string
	Description string
}

func Parse(filename string) ([]Target, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Makefile: %w", err)
	}

	defer file.Close()

	var targets []Target
	var lastComment string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			lastComment = ""
			continue
		}

		// Check for comment
		if comment, found := strings.CutPrefix(trimmed, "#"); found {
			lastComment = strings.TrimSpace(comment)
			continue
		}

		// Check for target definition
		if strings.Contains(line, ":") && !strings.HasPrefix(trimmed, "\t") {
			parts := strings.SplitN(line, ":", 2)
			targetName := strings.TrimSpace(parts[0])

			// Skip special targets
			if strings.HasPrefix(targetName, ".") || strings.Contains(targetName, "=") {
				lastComment = ""
				continue
			}

			// Handle multiple targets on one line
			names := strings.FieldsSeq(targetName)
			for name := range names {
				targets = append(targets, Target{
					Name:        name,
					Description: lastComment,
				})
			}

			lastComment = ""
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Makefile: %w", err)
	}

	return targets, nil
}
