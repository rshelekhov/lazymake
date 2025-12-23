package highlight

import (
	"strings"
	"testing"
)

func TestNewHighlighter(t *testing.T) {
	h := NewHighlighter()

	if h == nil {
		t.Fatal("NewHighlighter should return non-nil highlighter")
	}

	if h.cache == nil {
		t.Error("highlighter should have cache initialized")
	}

	if h.colorScheme == nil {
		t.Error("highlighter should have color scheme initialized")
	}

	if h.style == nil {
		t.Error("highlighter should have chroma style initialized")
	}
}

func TestHighlight(t *testing.T) {
	h := NewHighlighter()

	tests := []struct {
		name     string
		code     string
		language string
	}{
		{
			name:     "Python keywords",
			code:     "def hello():",
			language: "python",
		},
		{
			name:     "Bash commands",
			code:     "echo 'hello world'",
			language: "bash",
		},
		{
			name:     "Go code",
			code:     "package main",
			language: "go",
		},
		{
			name:     "JavaScript",
			code:     "const x = 42;",
			language: "javascript",
		},
		{
			name:     "Docker",
			code:     "FROM alpine:latest",
			language: "docker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Highlight(tt.code, tt.language)

			// Result should not be empty
			if result == "" {
				t.Error("highlight should return non-empty result")
			}

			// Result should contain the original code (possibly with ANSI codes)
			// We check if key terms are still present
			codeWords := strings.Fields(tt.code)
			if len(codeWords) > 0 {
				// At least the first word should be present (possibly styled)
				firstWord := strings.Trim(codeWords[0], ":;(){}")
				if !strings.Contains(result, firstWord) {
					t.Errorf("highlighted output should contain %q", firstWord)
				}
			}
		})
	}
}

func TestHighlightUnknownLanguage(t *testing.T) {
	h := NewHighlighter()

	code := "echo test"
	result := h.Highlight(code, "unknown-language-xyz")

	// Should still produce output (fallback to bash or plain text)
	if result == "" {
		t.Error("highlight should produce output even for unknown language")
	}
}

func TestHighlightLine(t *testing.T) {
	h := NewHighlighter()

	line := "print('hello')"
	result := h.HighlightLine(line, "python")

	if result == "" {
		t.Error("HighlightLine should return non-empty result")
	}

	// Should contain the word "print"
	if !strings.Contains(result, "print") {
		t.Error("highlighted line should contain original content")
	}
}

func TestHighlightCaching(t *testing.T) {
	h := NewHighlighter()

	code := "echo 'test'"
	language := "bash"

	// First call - cache miss
	result1 := h.Highlight(code, language)

	// Second call - cache hit (should return same result)
	result2 := h.Highlight(code, language)

	if result1 != result2 {
		t.Error("cached result should be identical to original")
	}

	// Verify cache is being used
	key := cacheKey(code, language)
	cached, found := h.cache.Get(key)
	if !found {
		t.Error("result should be in cache")
	}
	if cached != result1 {
		t.Error("cached value should match highlighted result")
	}
}

func TestInvalidateCache(t *testing.T) {
	h := NewHighlighter()

	// Add something to cache
	code := "echo 'test'"
	h.Highlight(code, "bash")

	// Verify it's cached
	key := cacheKey(code, "bash")
	_, found := h.cache.Get(key)
	if !found {
		t.Error("result should be in cache before invalidation")
	}

	// Invalidate cache
	h.InvalidateCache()

	// Verify cache is empty
	_, found = h.cache.Get(key)
	if found {
		t.Error("cache should be empty after invalidation")
	}
}

func TestDetectLanguageConvenience(t *testing.T) {
	h := NewHighlighter()

	recipe := []string{"#!/usr/bin/env python3", "print('hello')"}
	language := h.DetectLanguage(recipe, "")

	if language != "python" {
		t.Errorf("expected 'python', got %s", language)
	}

	// Test with override
	language = h.DetectLanguage(recipe, "go")
	if language != "go" {
		t.Errorf("override should work, expected 'go', got %s", language)
	}
}

func TestDefaultColorScheme(t *testing.T) {
	scheme := defaultColorScheme()

	// Verify all colors are set
	if scheme.Keyword == nil {
		t.Error("Keyword color should be set")
	}
	if scheme.String == nil {
		t.Error("String color should be set")
	}
	if scheme.Comment == nil {
		t.Error("Comment color should be set")
	}
	if scheme.Function == nil {
		t.Error("Function color should be set")
	}
	if scheme.Number == nil {
		t.Error("Number color should be set")
	}
	if scheme.Operator == nil {
		t.Error("Operator color should be set")
	}
	if scheme.Variable == nil {
		t.Error("Variable color should be set")
	}
	if scheme.Type == nil {
		t.Error("Type color should be set")
	}
	if scheme.Default == nil {
		t.Error("Default color should be set")
	}
}

func TestHighlightMultipleLanguages(t *testing.T) {
	h := NewHighlighter()

	languages := []string{"python", "go", "bash", "javascript", "rust", "docker"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			code := "test code"
			result := h.Highlight(code, lang)

			if result == "" {
				t.Errorf("highlight should work for %s", lang)
			}
		})
	}
}

func TestHighlightEmptyCode(t *testing.T) {
	h := NewHighlighter()

	result := h.Highlight("", "python")

	// Should handle empty code gracefully
	if result != "" {
		t.Error("empty code should produce empty result")
	}
}

func TestHighlightComplexCode(t *testing.T) {
	h := NewHighlighter()

	tests := []struct {
		name     string
		code     string
		language string
	}{
		{
			name: "Multi-line Python",
			code: `def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n-1)`,
			language: "python",
		},
		{
			name: "Shell script with pipes",
			code: `cat file.txt | grep "pattern" | sort | uniq`,
			language: "bash",
		},
		{
			name: "Go struct",
			code: `type User struct {
    Name  string
    Email string
}`,
			language: "go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Highlight(tt.code, tt.language)

			if result == "" {
				t.Error("highlight should handle complex code")
			}

			// Should preserve line structure (newlines)
			if strings.Count(tt.code, "\n") > 0 {
				if !strings.Contains(result, "\n") {
					t.Error("multi-line code should preserve newlines")
				}
			}
		})
	}
}
