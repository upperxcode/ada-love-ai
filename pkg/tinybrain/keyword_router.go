package tinybrain

import (
	"context"
	"strings"
)

// KeywordRouter is a lightweight deterministic fallback classifier used when
// an external HTTP classifier is unavailable. It uses simple keyword heuristics
// to distinguish programming/debugging intents from general conversation.
// This is intentionally conservative and deterministic: it prefers GENERAL when
// unsure to avoid triggering the orchestrator incorrectly.
type KeywordRouter struct{
}

func NewKeywordRouter() *KeywordRouter {
	return &KeywordRouter{}
}

func (r *KeywordRouter) DetectIntent(ctx context.Context, text string) (Intent, error) {
	if text == "" {
		return IntentGeneral, nil
	}
	l := strings.ToLower(text)

	// Simple heuristics: if text contains code markers or programming terms,
	// classify as GO_PROGRAMMING. Otherwise return GENERAL.
	codeMarkers := []string{
		"```", // code block
		"func ", "var ", "package ", "import ", "fmt.",
		"nil pointer", "null pointer", "panic", "stack trace",
		"segfault", "segmentation fault", "goroutine", "mutex", "deadlock",
		"compile", "compilar", "compilador", "build", "go run",
		"refator", "refactor", "lint", "gofmt", "golang", "go fmt",
		"pointer", "ponteiro", "erro de compilacao", "erro de compilação",
		"exception", "traceback", "debug", "trace",
	}

	techWords := []string{
		"api", "endpoint", "json", "http", "request", "response", "server", "client",
		"database", "db", "sql", "postgres", "mongodb", "redis", "query",
	}

	// if we see explicit code markers, strongly prefer programming
	for _, m := range codeMarkers {
		if strings.Contains(l, m) {
			return IntentGoProgramming, nil
		}
	}

	// If the text contains several technical words together, treat as programming
	techCount := 0
	for _, w := range techWords {
		if strings.Contains(l, w) {
			techCount++
		}
	}
	if techCount >= 2 {
		return IntentGoProgramming, nil
	}

	// Otherwise default to GENERAL
	return IntentGeneral, nil
}
