package variables

import (
	"bufio"
	"os/exec"
	"regexp"
	"strings"
)

var (
	// Matches the source line: # makefile (from 'Makefile', line 3)
	sourcePattern = regexp.MustCompile(`^#\s+(makefile|environment|automatic|default)`)

	// Matches variable assignment in database output: VAR = value or VAR := value
	dbVarPattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*([:+?!]?=)\s*(.*)$`)
)

// ExpandVariables runs `make --print-data-base` to get expanded variable values
// and updates the ExpandedValue field for all variables in the slice
// This function modifies the variables slice in place
func ExpandVariables(makefilePath string, variables []Variable) error {
	if len(variables) == 0 {
		return nil
	}

	// Run make --print-data-base to get all variable values
	cmd := exec.Command("make", "-f", makefilePath, "--print-data-base", "--no-builtin-rules", "--no-builtin-variables")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Graceful degradation: if make fails, just return without expanding
		// Variables will still have their RawValue populated
		return nil
	}

	// Parse the database output
	expandedVars := parseMakeDatabase(string(output))

	// Update the ExpandedValue field for each variable
	for i := range variables {
		if expanded, found := expandedVars[variables[i].Name]; found {
			variables[i].ExpandedValue = expanded
		} else {
			// If not found in database, use raw value as expanded value
			variables[i].ExpandedValue = variables[i].RawValue
		}
	}

	return nil
}

// parseMakeDatabase parses the output of `make --print-data-base`
// and returns a map of variable names to their expanded values
func parseMakeDatabase(output string) map[string]string {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(output))

	inVariablesSection := false
	currentSource := ""

	for scanner.Scan() {
		line := scanner.Text()

		// Look for the "# Variables" section marker
		if strings.HasPrefix(line, "# Variables") {
			inVariablesSection = true
			continue
		}

		// Skip until we're in the variables section
		if !inVariablesSection {
			continue
		}

		// Check if we've reached the end of variables section
		if isEndOfVariablesSection(line) {
			break
		}

		// Track the current source (makefile, environment, etc.)
		if matches := sourcePattern.FindStringSubmatch(line); matches != nil {
			currentSource = matches[1]
			continue
		}

		// Parse and store variable if valid
		if varName, value, ok := parseVariableLine(line, currentSource); ok {
			result[varName] = value
		}
	}

	return result
}

// isEndOfVariablesSection checks if we've reached the end of the variables section
func isEndOfVariablesSection(line string) bool {
	if !strings.HasPrefix(line, "# ") || sourcePattern.MatchString(line) {
		return false
	}
	return strings.Contains(line, "Files") || strings.Contains(line, "Implicit") ||
		strings.Contains(line, "Pattern") || strings.Contains(line, "VPATH")
}

// parseVariableLine parses a single variable line from make database output
// Returns the variable name, value, and true if the line is a valid variable assignment
func parseVariableLine(line, currentSource string) (varName, value string, ok bool) {
	// Skip comment lines and empty lines
	if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
		return "", "", false
	}

	// Parse variable assignment
	matches := dbVarPattern.FindStringSubmatch(line)
	if matches == nil {
		return "", "", false
	}

	varName = matches[1]
	value = matches[3]

	// Only accept variables from makefile or environment
	// Skip automatic and default variables unless they were explicitly defined
	if currentSource != "makefile" && currentSource != "environment" {
		return "", "", false
	}

	return varName, value, true
}
