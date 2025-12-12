package variables

import (
	"regexp"

	"github.com/rshelekhov/lazymake/internal/makefile"
)

var (
	// Matches variable references: $(VAR_NAME) or $(VAR)
	parenVarPattern = regexp.MustCompile(`\$\(([A-Za-z_][A-Za-z0-9_]*)\)`)

	// Matches single-character variable references: $X or $@
	// We only want to capture uppercase letters and underscores for user variables
	singleVarPattern = regexp.MustCompile(`\$([A-Z_])`)
)

// AnalyzeUsage scans target recipes to find which variables are used
// and updates the UsedByTargets field for all variables
// This function modifies the variables slice in place
func AnalyzeUsage(variables []Variable, targets []makefile.Target) {
	if len(variables) == 0 || len(targets) == 0 {
		return
	}

	// Create a map for quick lookup of variable names
	varMap := make(map[string]*Variable)
	for i := range variables {
		varMap[variables[i].Name] = &variables[i]
	}

	// Scan each target's recipe for variable references
	for _, target := range targets {
		if len(target.Recipe) == 0 {
			continue
		}

		// Track which variables are used by this target (use map to avoid duplicates)
		usedVars := make(map[string]bool)

		for _, recipeLine := range target.Recipe {
			// Find all variable references in this recipe line
			refs := extractVariableReferences(recipeLine)
			for _, ref := range refs {
				usedVars[ref] = true
			}
		}

		// Update the UsedByTargets field for each variable found
		for varName := range usedVars {
			if variable, found := varMap[varName]; found {
				variable.UsedByTargets = append(variable.UsedByTargets, target.Name)
			}
		}
	}
}

// extractVariableReferences finds all variable references in a string
// Returns a slice of variable names (without the $() or $ prefix)
func extractVariableReferences(text string) []string {
	var refs []string
	seen := make(map[string]bool)

	// Find $(VAR) style references
	matches := parenVarPattern.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		varName := match[1]
		if !seen[varName] {
			refs = append(refs, varName)
			seen[varName] = true
		}
	}

	// Find $X style references (single uppercase letter or underscore)
	matches = singleVarPattern.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		varName := match[1]
		if !seen[varName] {
			refs = append(refs, varName)
			seen[varName] = true
		}
	}

	return refs
}

// GetVariablesForTarget returns all variables used by a specific target
// Returns a new slice containing only the variables used by the target
func GetVariablesForTarget(targetName string, variables []Variable) []Variable {
	var result []Variable

	for _, variable := range variables {
		for _, usedTarget := range variable.UsedByTargets {
			if usedTarget == targetName {
				result = append(result, variable)
				break
			}
		}
	}

	return result
}
