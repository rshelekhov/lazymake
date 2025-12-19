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
	Keyword  lipgloss.Color // Keywords (if, for, def, func)
	String   lipgloss.Color // String literals
	Comment  lipgloss.Color // Comments
	Function lipgloss.Color // Function names
	Number   lipgloss.Color // Numeric literals
	Operator lipgloss.Color // Operators (+, -, =, |)
	Variable lipgloss.Color // Variables
	Type     lipgloss.Color // Types
	Default  lipgloss.Color // Default text
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

// defaultColorScheme returns a monokai-inspired color scheme
func defaultColorScheme() *ColorScheme {
	return &ColorScheme{
		Keyword:  lipgloss.Color("#FF79C6"), // Pink
		String:   lipgloss.Color("#E6DB74"), // Yellow
		Comment:  lipgloss.Color("#75715E"), // Gray
		Function: lipgloss.Color("#A6E22E"), // Green
		Number:   lipgloss.Color("#AE81FF"), // Purple
		Operator: lipgloss.Color("#F92672"), // Red/Pink
		Variable: lipgloss.Color("#FD971F"), // Orange
		Type:     lipgloss.Color("#66D9EF"), // Cyan
		Default:  lipgloss.Color("#F8F8F2"), // White
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
func (h *Highlighter) getTokenColor(tokenType chroma.TokenType) lipgloss.Color {
	// Check exact token types first
	if color := h.getExactTokenColor(tokenType); color != "" {
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
func (h *Highlighter) getExactTokenColor(tokenType chroma.TokenType) lipgloss.Color {
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
		return ""
	}
}

// InvalidateCache clears the highlighting cache
func (h *Highlighter) InvalidateCache() {
	h.cache.Clear()
}
