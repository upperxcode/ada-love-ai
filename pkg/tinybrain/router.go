package tinybrain

import (
	"ada-love-ai/pkg/providers"
	"context"
	"fmt"
	"strings"
)

type Intent string

const (
	IntentGeneral       Intent = "GENERAL"
	IntentGoProgramming Intent = "GO_PROGRAMMING"
	IntentCodeReview    Intent = "CODE_REVIEW"
	IntentDebugging     Intent = "DEBUGGING"
	IntentArchitecture  Intent = "ARCHITECTURE"
)

type TinyBrainRouter struct {
	model providers.LLMProvider
}

func NewTinyBrainRouter(model providers.LLMProvider) *TinyBrainRouter {
	return &TinyBrainRouter{
		model: model,
	}
}

func (r *TinyBrainRouter) DetectIntent(ctx context.Context, userInput string) (Intent, error) {
	systemPrompt := `You are an intent classifier for a coding assistant. Analyze the user's message and classify it into ONE of these categories:

1. GENERAL - General questions, greetings, casual chat, non-coding topics
2. GO_PROGRAMMING - Writing, creating, or implementing Go code (functions, structs, APIs, CLI tools, etc.)
3. CODE_REVIEW - Asking for code review, improvements, best practices on existing code
4. DEBUGGING - Help with errors, bugs, crashes, test failures, stack traces
5. ARCHITECTURE - System design, patterns, project structure, scalability decisions

Respond ONLY with the category name (e.g., GO_PROGRAMMING). No explanation.`

	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userInput},
	}

	resp, err := r.model.Chat(ctx, messages, nil, "", map[string]any{
		"temperature": 0.1,
		"max_tokens":  10,
	})
	if err != nil {
		return IntentGeneral, fmt.Errorf("tinybrain classification failed: %w", err)
	}

	result := strings.ToUpper(strings.TrimSpace(resp.Content))
	switch result {
	case "GO_PROGRAMMING":
		return IntentGoProgramming, nil
	case "CODE_REVIEW":
		return IntentCodeReview, nil
	case "DEBUGGING":
		return IntentDebugging, nil
	case "ARCHITECTURE":
		return IntentArchitecture, nil
	default:
		return IntentGeneral, nil
	}
}
