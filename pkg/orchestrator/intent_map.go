package orchestrator

import (
	"strings"
)

// CandidateLabels returns the default set of labels passed to the intent
// classifier. Keep these human-friendly and capitalized for backwards
// compatibility with a default HTTP classifier that expects such labels.
func CandidateLabels() []string {
	// These are simple defaults; workspace-specific overrides can be
	// introduced later (DB / adaCfg) without changing callers.
	return []string{"React", "Go", "Tester", "Geral"}
}

// ResolveIntentCandidates maps a normalized classifier label to a list of
// candidate agent identifiers (aliases). The returned slice contains
// candidate agent IDs or name fragments that will be normalized and
// matched against the AgentLoop registry.
func ResolveIntentCandidates(label string) []string {
	if strings.TrimSpace(label) == "" {
		return nil
	}

	l := strings.ToLower(strings.TrimSpace(label))

	// Default mapping. Keep it small and explicit so tests are straightforward.
	var defaults = map[string][]string{
		"react":   {"react", "react_agent", "reactjs", "reactjs_agent"},
		"flutter": {"flutter", "flutter_agent"},
		"go":      {"go", "golang", "go_agent", "golang_agent", "galang", "golanger"},
		"golang":  {"go", "golang", "go_agent", "golang_agent", "galang", "golanger"},
		"tester":  {"tester", "test", "tester_agent", "testing"},
		"test":    {"tester", "test", "tester_agent", "testing"},
	}

	// Accept some localized variants
	switch l {
	case "geral", "assunto geral", "general":
		// GENERAL handled specially by caller — return nil so caller can
		// detect and bypass the orchestrator if desired.
		return nil
	}

	if v, ok := defaults[l]; ok {
		return append([]string(nil), v...)
	}

	// Fallback: return the label itself as a candidate so callers may try to
	// match an agent with that name (after normalization).
	return []string{l}
}
