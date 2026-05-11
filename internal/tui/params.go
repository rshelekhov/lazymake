package tui

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ExecutionParams captures one form submission for a target. EnvRaw and
// FlagsRaw are kept verbatim so the user can edit the same string the
// next time the form opens; Env and Flags are the parsed equivalents
// passed down to the executor.
type ExecutionParams struct {
	EnvRaw   string
	FlagsRaw string
	Env      map[string]string
	Flags    []string
}

// IsEmpty reports whether no params were specified. A run with empty
// params behaves identically to a plain `make <target>` invocation.
func (p ExecutionParams) IsEmpty() bool {
	return len(p.Env) == 0 && len(p.Flags) == 0
}

// parseEnvLine parses a whitespace-separated list of KEY=VALUE pairs.
// Double-quoted values may contain spaces:
//
//	FOO=bar BAZ="hello world" EMPTY=
//
// Returns an error on the first malformed token. The error message
// names the offending token so users can correct it in place.
func parseEnvLine(s string) (map[string]string, error) {
	tokens, err := shellSplit(s)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, nil
	}

	out := make(map[string]string, len(tokens))
	for _, tok := range tokens {
		eq := strings.IndexByte(tok, '=')
		if eq <= 0 {
			return nil, fmt.Errorf("env: expected KEY=VALUE, got %q", tok)
		}
		key := tok[:eq]
		if !isValidEnvKey(key) {
			return nil, fmt.Errorf("env: invalid key %q (use letters, digits, underscore; not starting with a digit)", key)
		}
		out[key] = tok[eq+1:]
	}
	return out, nil
}

// parseFlagsLine parses a whitespace-separated list of make flags.
// Every token must start with '-' so we can reject obvious user
// mistakes like putting variables here instead of in the env field.
//
//	-j4 -k --always-make
//
// Quoted values are accepted but unusual for flags; we still support
// them for consistency with parseEnvLine.
func parseFlagsLine(s string) ([]string, error) {
	tokens, err := shellSplit(s)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, nil
	}
	for _, tok := range tokens {
		if !strings.HasPrefix(tok, "-") {
			if strings.Contains(tok, "=") {
				return nil, fmt.Errorf("flags: %q looks like a variable; put it in the env field instead", tok)
			}
			return nil, fmt.Errorf("flags: %q does not start with '-'", tok)
		}
	}
	return tokens, nil
}

// isValidEnvKey returns true if s is a portable env var name:
//
//	[A-Za-z_][A-Za-z0-9_]*
//
// This is the POSIX shell rule. We intentionally do not accept names
// starting with a digit, even though some make versions tolerate it.
func isValidEnvKey(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			continue
		}
		if i > 0 && r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}

// shellSplit splits s on whitespace, honoring double quotes so values
// like `MSG="hello world"` survive as a single token. It is a small,
// dependency-free subset of POSIX shell tokenization: it does not
// handle single quotes, escapes, or environment-variable expansion.
//
// Returns an error if a quote is left unterminated.
func shellSplit(s string) ([]string, error) {
	var tokens []string
	var cur strings.Builder
	inQuote := false
	hasContent := false

	flush := func() {
		if hasContent {
			tokens = append(tokens, cur.String())
		}
		cur.Reset()
		hasContent = false
	}

	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
			hasContent = true // an empty quoted string is still a token
		case !inQuote && (r == ' ' || r == '\t'):
			flush()
		default:
			cur.WriteRune(r)
			hasContent = true
		}
	}
	if inQuote {
		return nil, errors.New("unterminated double quote")
	}
	flush()
	return tokens, nil
}

// formatEnvRaw serializes a parsed env map back into the raw textinput
// format. Keys are sorted to make the result deterministic and stable
// across saves; values containing whitespace are double-quoted to
// round-trip through parseEnvLine.
//
// Returns "" for an empty/nil map so the textinput shows its
// placeholder instead of a blank string of quotes.
func formatEnvRaw(env map[string]string) string {
	if len(env) == 0 {
		return ""
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := env[k]
		if strings.ContainsAny(v, " \t\"") {
			// Escape embedded double quotes by stripping them — the
			// textinput parser doesn't support escapes, so this is
			// the closest we can do without breaking round-trip.
			v = strings.ReplaceAll(v, `"`, "")
			parts = append(parts, k+`="`+v+`"`)
			continue
		}
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, " ")
}

// formatFlagsRaw serializes a parsed flags slice back into the raw
// textinput format. Flags don't carry whitespace today, so a plain
// space-join suffices; the helper exists for symmetry with
// formatEnvRaw and to centralize this rule.
func formatFlagsRaw(flags []string) string {
	if len(flags) == 0 {
		return ""
	}
	return strings.Join(flags, " ")
}
