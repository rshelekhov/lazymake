package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

// Highlighter provides syntax highlighting for code using Chroma
type Highlighter struct {
	style       *chroma.Style
	cache       *Cache
	colorScheme *ColorScheme
}

// ColorScheme defines the colors for syntax highlighting
type ColorScheme struct {
	Keyword  lipgloss.TerminalColor // Keywords (if, for, def, func)
	String   lipgloss.TerminalColor // String literals
	Comment  lipgloss.TerminalColor // Comments
	Function lipgloss.TerminalColor // Function names
	Number   lipgloss.TerminalColor // Numeric literals
	Operator lipgloss.TerminalColor // Operators (+, -, =, |)
	Variable lipgloss.TerminalColor // Variables
	Type     lipgloss.TerminalColor // Types
	Default  lipgloss.TerminalColor // Default text
}

// NewHighlighter creates a new syntax highlighter with default settings
func NewHighlighter() *Highlighter {
	// Use a dark-friendly Chroma style
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	return &Highlighter{
		style:       style,
		cache:       NewCache(1000), // Cache 1000 highlighted code snippets
		colorScheme: defaultColorScheme(),
	}
}

// defaultColorScheme returns a GitHub VS Code theme color scheme with adaptive light/dark support
func defaultColorScheme() *ColorScheme {
	return &ColorScheme{
		Keyword:  lipgloss.AdaptiveColor{Light: "#0000FF", Dark: "#569CD6"}, // Keywords (if, for, func)
		String:   lipgloss.AdaptiveColor{Light: "#A31515", Dark: "#CE9178"}, // String literals
		Comment:  lipgloss.AdaptiveColor{Light: "#008000", Dark: "#6A9955"}, // Comments
		Function: lipgloss.AdaptiveColor{Light: "#795E26", Dark: "#C8C6BF"}, // Function names
		Number:   lipgloss.AdaptiveColor{Light: "#098658", Dark: "#B5CEA8"}, // Numeric literals
		Operator: lipgloss.AdaptiveColor{Light: "#000000", Dark: "#D4D4D4"}, // Operators (+, -, =)
		Variable: lipgloss.AdaptiveColor{Light: "#001080", Dark: "#9CDCFE"}, // Variables
		Type:     lipgloss.AdaptiveColor{Light: "#267F99", Dark: "#4EC9B0"}, // Types, classes
		Default:  lipgloss.AdaptiveColor{Light: "#000000", Dark: "#D4D4D4"}, // Default text
	}
}

// Highlight applies syntax highlighting to code for the given language
func (h *Highlighter) Highlight(code, language string) string {
	// Check cache first
	key := cacheKey(code, language)
	if cached, found := h.cache.Get(key); found {
		return cached
	}

	// Get lexer for language
	lexer := lexers.Get(language)
	if lexer == nil {
		// Try fallback to bash for unknown languages
		lexer = lexers.Get("bash")
	}
	if lexer == nil {
		// Ultimate fallback: return plain text
		return code
	}

	// Ensure lexer is configured
	lexer = chroma.Coalesce(lexer)

	// Tokenize
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		// Fallback to plain text on error
		return code
	}

	// Convert tokens to styled output
	var result strings.Builder
	for _, token := range iterator.Tokens() {
		style := h.tokenToLipgloss(token.Type)
		result.WriteString(style.Render(token.Value))
	}

	highlighted := result.String()

	// Cache result
	h.cache.Set(key, highlighted)

	return highlighted
}

// HighlightLine highlights a single line of code
func (h *Highlighter) HighlightLine(line, language string) string {
	return h.Highlight(line, language)
}

// DetectLanguage is a convenience method that wraps the package-level function
func (h *Highlighter) DetectLanguage(recipe []string, override string) string {
	return DetectLanguage(recipe, override)
}

// tokenToLipgloss converts a Chroma token type to a Lipgloss style
func (h *Highlighter) tokenToLipgloss(tokenType chroma.TokenType) lipgloss.Style {
	color := h.getTokenColor(tokenType)
	return lipgloss.NewStyle().Foreground(color)
}

// getTokenColor returns the appropriate color for a token type
func (h *Highlighter) getTokenColor(tokenType chroma.TokenType) lipgloss.TerminalColor {
	// Check exact token types first
	if color := h.getExactTokenColor(tokenType); color != nil {
		return color
	}

	// Check categories
	if tokenType.InCategory(chroma.Keyword) {
		return h.colorScheme.Keyword
	}
	if tokenType.InCategory(chroma.String) {
		return h.colorScheme.String
	}
	if tokenType.InCategory(chroma.Comment) {
		return h.colorScheme.Comment
	}
	if tokenType.InCategory(chroma.Number) {
		return h.colorScheme.Number
	}
	if tokenType.InCategory(chroma.Operator) {
		return h.colorScheme.Operator
	}
	if tokenType.InCategory(chroma.Name) {
		return h.colorScheme.Function
	}

	return h.colorScheme.Default
}

// getExactTokenColor returns color for exact token type matches
func (h *Highlighter) getExactTokenColor(tokenType chroma.TokenType) lipgloss.TerminalColor {
	switch tokenType {
	case chroma.Keyword:
		return h.colorScheme.Keyword
	case chroma.String:
		return h.colorScheme.String
	case chroma.Comment:
		return h.colorScheme.Comment
	case chroma.NameFunction, chroma.NameBuiltin:
		return h.colorScheme.Function
	case chroma.Number:
		return h.colorScheme.Number
	case chroma.Operator:
		return h.colorScheme.Operator
	case chroma.NameVariable:
		return h.colorScheme.Variable
	case chroma.NameClass, chroma.KeywordType:
		return h.colorScheme.Type
	default:
		return nil
	}
}

// InvalidateCache clears the highlighting cache
func (h *Highlighter) InvalidateCache() {
	h.cache.Clear()
}
