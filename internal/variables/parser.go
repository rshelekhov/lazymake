package variables

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var (
	// varPattern matches variable assignments: VAR = value, VAR := value, etc.
	// Captures: (1) variable name, (2) operator, (3) value
	varPattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*([:+?!]?=)\s*(.*)$`)

	// exportPattern matches export declarations: export VAR or export VAR = value
	exportPattern = regexp.MustCompile(`^export\s+([A-Za-z_][A-Za-z0-9_]*)`)

	// exportWithAssignPattern matches: export VAR = value
	exportWithAssignPattern = regexp.MustCompile(`^export\s+([A-Za-z_][A-Za-z0-9_]*)\s*([:+?!]?=)\s*(.*)$`)
)

// ParseVariables extracts variable definitions from a Makefile
// Returns a slice of Variable structs with Name, RawValue, Type, DefinedAt, and IsExported fields populated
// The ExpandedValue and UsedByTargets fields will be empty and should be populated by other functions
func ParseVariables(makefilePath string) ([]Variable, error) {
	file, err := os.Open(makefilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var variables []Variable
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var continuedLine string
	var continuedLineStart int

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Handle line continuations
		if handleLineContinuation(line, &continuedLine, &continuedLineStart, lineNum) {
			continue
		}

		// If we were building a continued line, append this final part
		if continuedLine != "" {
			line = continuedLine + line
			lineNum = continuedLineStart
			continuedLine = ""
		}

		// Skip comments, empty lines, targets, and recipe lines
		if shouldSkipLine(line) {
			continue
		}

		trimmedLine := strings.TrimSpace(line)

		// Try to process as export statement
		if variable, found := processExportStatement(trimmedLine, &variables, lineNum); found {
			if variable.Name != "" { // New variable needs to be added
				variables = append(variables, variable)
			}
			continue
		}

		// Try to process as regular variable assignment
		if variable, found := processVariableAssignment(trimmedLine, lineNum); found {
			variables = append(variables, variable)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return variables, nil
}

// handleLineContinuation manages line continuations (backslash at end)
// Returns true if the line should be continued (and caller should skip processing)
func handleLineContinuation(line string, continuedLine *string, continuedLineStart *int, lineNum int) bool {
	if strings.HasSuffix(strings.TrimRight(line, " \t"), "\\") {
		if *continuedLine == "" {
			*continuedLineStart = lineNum
		}
		// Remove the backslash and trailing whitespace, append the line
		*continuedLine += strings.TrimSuffix(strings.TrimRight(line, " \t"), "\\")
		return true
	}
	return false
}

// shouldSkipLine checks if a line should be skipped during parsing
func shouldSkipLine(line string) bool {
	trimmedLine := strings.TrimSpace(line)

	// Skip comments and empty lines
	if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
		return true
	}

	// Skip target definitions (lines with ':' that aren't variable assignments)
	if strings.Contains(trimmedLine, ":") && !strings.Contains(trimmedLine, "=") {
		return true
	}

	// Skip recipe lines (start with tab)
	if strings.HasPrefix(line, "\t") {
		return true
	}

	return false
}

// processExportStatement handles export declarations
// Returns the variable and true if this is an export statement
// If the variable already exists and was just marked as exported, returns empty variable
func processExportStatement(trimmedLine string, variables *[]Variable, lineNum int) (Variable, bool) {
	// Check for export with assignment: export VAR = value
	if matches := exportWithAssignPattern.FindStringSubmatch(trimmedLine); matches != nil {
		return Variable{
			Name:       matches[1],
			RawValue:   strings.TrimSpace(matches[3]),
			Type:       operatorToVarType(matches[2]),
			DefinedAt:  lineNum,
			IsExported: true,
		}, true
	}

	// Check for export without assignment: export VAR
	if matches := exportPattern.FindStringSubmatch(trimmedLine); matches != nil {
		varName := matches[1]

		// Mark existing variable as exported, or create a new one
		for i := range *variables {
			if (*variables)[i].Name == varName {
				(*variables)[i].IsExported = true
				return Variable{}, true // Return empty variable since we modified existing one
			}
		}

		// Variable exported but not yet defined in this file
		return Variable{
			Name:       varName,
			Type:       VarEnvironment,
			DefinedAt:  lineNum,
			IsExported: true,
		}, true
	}

	return Variable{}, false
}

// processVariableAssignment handles regular variable assignments
// Returns the variable and true if this is a valid assignment
func processVariableAssignment(trimmedLine string, lineNum int) (Variable, bool) {
	matches := varPattern.FindStringSubmatch(trimmedLine)
	if matches == nil {
		return Variable{}, false
	}

	return Variable{
		Name:      matches[1],
		RawValue:  strings.TrimSpace(matches[3]),
		Type:      operatorToVarType(matches[2]),
		DefinedAt: lineNum,
	}, true
}

// operatorToVarType converts a Makefile assignment operator to a VarType
func operatorToVarType(operator string) VarType {
	switch operator {
	case "=":
		return VarRecursive
	case ":=":
		return VarSimple
	case "+=":
		return VarAppend
	case "?=":
		return VarConditional
	case "!=":
		return VarShell
	default:
		return VarUnknown
	}
}
