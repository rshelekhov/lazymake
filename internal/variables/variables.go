package variables

// Variable represents a Makefile variable with its definition and usage information
type Variable struct {
	Name          string   // Variable name (e.g., "GOFLAGS", "CC")
	RawValue      string   // Value as written in Makefile
	ExpandedValue string   // Value after expansion by make
	Type          VarType  // How the variable is defined (=, :=, +=, ?=, !=)
	DefinedAt     int      // Line number in Makefile where defined
	IsExported    bool     // Whether the variable is exported to environment
	UsedByTargets []string // Names of targets that use this variable
}

// VarType represents the type of variable assignment in a Makefile
type VarType int

const (
	VarRecursive   VarType = iota // VAR = value (recursively expanded)
	VarSimple                      // VAR := value (simply expanded)
	VarAppend                      // VAR += value (append)
	VarConditional                 // VAR ?= value (conditional assignment)
	VarShell                       // VAR != command (shell command expansion)
	VarEnvironment                 // Variable from environment
	VarAutomatic                   // Automatic variable ($@, $<, etc.)
	VarUnknown                     // Unknown/other type
)

// String returns a human-readable string representation of the variable type
func (vt VarType) String() string {
	switch vt {
	case VarRecursive:
		return "Recursive"
	case VarSimple:
		return "Simply Expanded"
	case VarAppend:
		return "Append"
	case VarConditional:
		return "Conditional"
	case VarShell:
		return "Shell"
	case VarEnvironment:
		return "Environment"
	case VarAutomatic:
		return "Automatic"
	default:
		return "Unknown"
	}
}

// Symbol returns the assignment operator symbol for the variable type
func (vt VarType) Symbol() string {
	switch vt {
	case VarRecursive:
		return "="
	case VarSimple:
		return ":="
	case VarAppend:
		return "+="
	case VarConditional:
		return "?="
	case VarShell:
		return "!="
	default:
		return ""
	}
}
