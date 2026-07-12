package utils

import (
	"strings"
	"sync/atomic"
	"unicode"
)

// Global variable to disable truncation
var disableTruncation atomic.Bool

// SetDisableTruncation globally enables or disables string truncation
func SetDisableTruncation(enabled bool) {
	disableTruncation.Store(enabled)
}

// SanitizeMessageContent removes Unicode control characters, format characters (RTL overrides,
// zero-width characters), and other non-graphic characters that could confuse an LLM
// or cause display issues in the agent UI.
func SanitizeMessageContent(input string) string {
	var sb strings.Builder
	// Pre-allocate memory to avoid multiple allocations
	sb.Grow(len(input))

	for _, r := range input {
		// unicode.IsGraphic returns true if the rune is a Unicode graphic character.
		// This includes letters, marks, numbers, punctuation, and symbols.
		// It excludes control characters (Cc), format characters (Cf),
		// surrogates (Cs), and private use (Co).
		if unicode.IsGraphic(r) || r == '\n' || r == '\r' || r == '\t' {
			sb.WriteRune(r)
		}
	}

	return sb.String()
}

// Truncate returns a truncated version of s with at most maxLen runes.
// Handles multi-byte Unicode characters properly.
// If the string is truncated, "..." is appended to indicate truncation.
func Truncate(s string, maxLen int) string {
	// If the no-truncate flag is active, it returns the full string
	if disableTruncation.Load() {
		return s
	}
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	// Reserve 3 chars for "..."
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// DerefStr dereferences a pointer to a string and
// returns the value or a fallback if the pointer is nil.
func DerefStr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

// StripThoughtTags removes reasoning blocks wrapped in <thought>...</thought>
// or [THOUGHT]...[/THOUGHT] tags. This prevents internal reasoning from
// leaking into the final UI response or summaries.
func StripThoughtTags(s string) string {
	// Simple approach: look for tags and remove everything between them
	// including the tags themselves.
	tags := []struct {
		start, end string
	}{
		{"<thought>", "</thought>"},
		{"[THOUGHT]", "[/THOUGHT]"},
	}

	result := s
	for _, tag := range tags {
		for {
			startIdx := strings.Index(result, tag.start)
			if startIdx == -1 {
				break
			}
			endIdx := strings.Index(result[startIdx:], tag.end)
			if endIdx == -1 {
				// Unclosed tag, just remove from start to end of string
				result = result[:startIdx]
				break
			}
			// Add startIdx back because endIdx is relative to result[startIdx:]
			endIdx += startIdx + len(tag.end)
			result = result[:startIdx] + result[endIdx:]
		}
	}
	return strings.TrimSpace(result)
}
