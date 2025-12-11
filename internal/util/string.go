package util

import "strings"

// WriteString writes a string to a strings.Builder, ignoring the returned error
// since strings.Builder.Write methods never return errors.
func WriteString(b *strings.Builder, s string) {
	_, _ = b.WriteString(s)
}