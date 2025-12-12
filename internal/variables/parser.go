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

		// Handle line continuations (backslash at end)
		if strings.HasSuffix(strings.TrimRight(line, " \t"), "\\") {
			if continuedLine == "" {
				continuedLineStart = lineNum
			}
			// Remove the backslash and trailing whitespace, append the line
			continuedLine += strings.TrimSuffix(strings.TrimRight(line, " \t"), "\\")
			continue
		}

		// If we were building a continued line, append this final part
		if continuedLine != "" {
			line = continuedLine + line
			lineNum = continuedLineStart
			continuedLine = ""
		}

		// Skip comments and empty lines
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Skip target definitions (lines with ':' that aren't variable assignments)
		if strings.Contains(trimmedLine, ":") && !strings.Contains(trimmedLine, "=") {
			continue
		}

		// Skip recipe lines (start with tab)
		if strings.HasPrefix(line, "\t") {
			continue
		}

		// Check for export with assignment: export VAR = value
		if matches := exportWithAssignPattern.FindStringSubmatch(trimmedLine); matches != nil {
			varName := matches[1]
			operator := matches[2]
			value := strings.TrimSpace(matches[3])

			variables = append(variables, Variable{
				Name:       varName,
				RawValue:   value,
				Type:       operatorToVarType(operator),
				DefinedAt:  lineNum,
				IsExported: true,
			})
			continue
		}

		// Check for export without assignment: export VAR
		if matches := exportPattern.FindStringSubmatch(trimmedLine); matches != nil {
			varName := matches[1]

			// Mark existing variable as exported, or create a new one
			found := false
			for i := range variables {
				if variables[i].Name == varName {
					variables[i].IsExported = true
					found = true
					break
				}
			}

			if !found {
				// Variable exported but not yet defined in this file
				// It might be defined elsewhere or in environment
				variables = append(variables, Variable{
					Name:       varName,
					Type:       VarEnvironment,
					DefinedAt:  lineNum,
					IsExported: true,
				})
			}
			continue
		}

		// Check for regular variable assignment
		if matches := varPattern.FindStringSubmatch(trimmedLine); matches != nil {
			varName := matches[1]
			operator := matches[2]
			value := strings.TrimSpace(matches[3])

			variables = append(variables, Variable{
				Name:      varName,
				RawValue:  value,
				Type:      operatorToVarType(operator),
				DefinedAt: lineNum,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return variables, nil
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
