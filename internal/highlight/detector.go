package highlight

import (
	"regexp"
	"strings"
)

// commandPattern represents a pattern to detect language from commands
type commandPattern struct {
	pattern  *regexp.Regexp
	language string
	weight   int // Higher = more confidence
}

var (
	// shebangPatterns maps shebang patterns to languages
	shebangPatterns = map[*regexp.Regexp]string{
		regexp.MustCompile(`^#!/bin/bash`):              "bash",
		regexp.MustCompile(`^#!/bin/sh`):                "sh",
		regexp.MustCompile(`^#!/usr/bin/env\s+bash`):   "bash",
		regexp.MustCompile(`^#!/usr/bin/env\s+sh`):     "sh",
		regexp.MustCompile(`^#!/usr/bin/env\s+python3?`): "python",
		regexp.MustCompile(`^#!/usr/bin/env\s+ruby`):   "ruby",
		regexp.MustCompile(`^#!/usr/bin/env\s+node`):   "javascript",
		regexp.MustCompile(`^#!/usr/bin/env\s+perl`):   "perl",
		regexp.MustCompile(`^#!/usr/bin/env\s+php`):    "php",
	}

	// commandPatterns for detecting language from commands
	commandPatterns = []commandPattern{
		// Go (high confidence)
		{regexp.MustCompile(`^\s*go\s+(build|test|run|mod|get|install|clean)`), "go", 90},

		// Python (high confidence)
		{regexp.MustCompile(`^\s*python3?\s+`), "python", 85},
		{regexp.MustCompile(`^\s*pip3?\s+(install|freeze|list)`), "python", 80},
		{regexp.MustCompile(`^\s*poetry\s+`), "python", 80},

		// Node.js/JavaScript (high confidence)
		{regexp.MustCompile(`^\s*npm\s+(install|run|test|build|start)`), "javascript", 85},
		{regexp.MustCompile(`^\s*yarn\s+(install|run|build|test|add)`), "javascript", 85},
		{regexp.MustCompile(`^\s*node\s+`), "javascript", 80},
		{regexp.MustCompile(`^\s*npx\s+`), "javascript", 80},

		// Rust (high confidence)
		{regexp.MustCompile(`^\s*cargo\s+(build|test|run|check|clippy)`), "rust", 90},

		// C/C++ (moderate confidence)
		{regexp.MustCompile(`^\s*gcc\s+`), "c", 75},
		{regexp.MustCompile(`^\s*g\+\+\s+`), "cpp", 75},
		{regexp.MustCompile(`^\s*clang\s+`), "c", 75},
		{regexp.MustCompile(`^\s*make\s+`), "makefile", 70},
		{regexp.MustCompile(`^\s*cmake\s+`), "cmake", 75},

		// Ruby
		{regexp.MustCompile(`^\s*ruby\s+`), "ruby", 85},
		{regexp.MustCompile(`^\s*bundle\s+(install|exec)`), "ruby", 80},
		{regexp.MustCompile(`^\s*gem\s+(install|list)`), "ruby", 75},

		// Java/JVM
		{regexp.MustCompile(`^\s*javac?\s+`), "java", 85},
		{regexp.MustCompile(`^\s*mvn\s+`), "java", 80},
		{regexp.MustCompile(`^\s*gradle\s+`), "java", 80},

		// PHP
		{regexp.MustCompile(`^\s*php\s+`), "php", 85},
		{regexp.MustCompile(`^\s*composer\s+`), "php", 80},

		// Kubernetes/Cloud
		{regexp.MustCompile(`^\s*kubectl\s+`), "yaml", 75},
		{regexp.MustCompile(`^\s*helm\s+`), "yaml", 75},

		// Shell utilities (lower confidence - might be just shell)
		{regexp.MustCompile(`^\s*curl\s+`), "bash", 50},
		{regexp.MustCompile(`^\s*wget\s+`), "bash", 50},
		{regexp.MustCompile(`^\s*git\s+`), "bash", 50},
	}

	// languageAliases maps common variants to canonical Chroma lexer names
	languageAliases = map[string]string{
		"py":         "python",
		"js":         "javascript",
		"ts":         "typescript",
		"golang":     "go",
		"sh":         "bash",
		"shell":      "bash",
		"yml":        "yaml",
		"dockerfile": "docker",
	}

	// commentOverridePattern matches language override comments
	commentOverridePattern = regexp.MustCompile(`^\s*#\s*(language|lang|syntax):\s*(\w+)`)
)

// DetectLanguage analyzes a recipe to determine the programming language.
// Detection priority:
// 1. Manual override parameter
// 2. Shebang lines (#!/usr/bin/env python)
// 3. Command patterns (docker, npm, go, etc.)
// 4. Falls back to "bash" as default
func DetectLanguage(recipe []string, override string) string {
	// 1. Manual override (highest priority)
	if override != "" {
		return normalizeLanguage(override)
	}

	if len(recipe) == 0 {
		return "bash"
	}

	// 2. Shebang detection (very reliable)
	if lang := detectShebang(recipe); lang != "" {
		return lang
	}

	// 3. Command pattern matching (good heuristic)
	if lang := detectFromCommands(recipe); lang != "" {
		return lang
	}

	// 4. Default fallback
	return "bash"
}

// detectShebang checks the first line for a shebang and returns the language
func detectShebang(recipe []string) string {
	if len(recipe) == 0 {
		return ""
	}

	firstLine := strings.TrimSpace(recipe[0])
	for pattern, lang := range shebangPatterns {
		if pattern.MatchString(firstLine) {
			return lang
		}
	}
	return ""
}

// detectFromCommands scans recipe lines for command patterns and returns
// the language with the highest weighted vote
func detectFromCommands(recipe []string) string {
	votes := make(map[string]int)

	for _, line := range recipe {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		for _, pattern := range commandPatterns {
			if pattern.pattern.MatchString(line) {
				votes[pattern.language] += pattern.weight
			}
		}
	}

	// Return language with highest vote
	return getHighestVote(votes)
}

// getHighestVote returns the key with the highest value from a vote map
func getHighestVote(votes map[string]int) string {
	if len(votes) == 0 {
		return ""
	}

	maxVotes := 0
	winner := ""
	for lang, count := range votes {
		if count > maxVotes {
			maxVotes = count
			winner = lang
		}
	}
	return winner
}

// normalizeLanguage converts language aliases to canonical names
func normalizeLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if alias, found := languageAliases[lang]; found {
		return alias
	}
	return lang
}

// ParseLanguageOverride extracts language override from a comment line
// Supports formats: "# language: python", "# lang: go", "# syntax: rust"
func ParseLanguageOverride(comment string) string {
	matches := commentOverridePattern.FindStringSubmatch(comment)
	if len(matches) > 2 {
		return normalizeLanguage(matches[2])
	}
	return ""
}
