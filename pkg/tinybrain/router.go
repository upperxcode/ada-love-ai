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
    // Prompt otimizado com Few-Shot e delimitação estrita
    systemPrompt := `You are a strict binary and multi-class intent router for an AI assistant.
Your ONLY job is to output one of these exact tokens: GENERAL, GO_PROGRAMMING, CODE_REVIEW, DEBUGGING, ARCHITECTURE.

CRITICAL RULES:
- If the user asks for factual info, list of cities, geography, math, weather, history, or casual chat, you MUST respond with GENERAL.
- Do NOT write code. Do NOT create software architectures for general questions.

EXAMPLES:
User: me dê as 5 cidades brasileiras com o maior idh
Assistant: GENERAL

User: como tratar erros de forma idiomatica no go?
Assistant: GO_PROGRAMMING

User: crie uma API com Echo para gerenciar condominio
Assistant: GO_PROGRAMMING

User: por que meu ponteiro da nil pointer aqui?
Assistant: DEBUGGING

User: o que achou da estrutura desse struct? alguma boa pratica?
Assistant: CODE_REVIEW

User: qual o valor da moto dominar 400?
Assistant: GENERAL

Respond ONLY with the category token. No markdown, no punctuation, no explanations.`

    messages := []providers.Message{
        {Role: "system", Content: systemPrompt},
        {Role: "user", Content: userInput},
    }

    // Forçamos a temperatura para 0.0 absoluto
    resp, err := r.model.Chat(ctx, messages, nil, "", map[string]any{
        "temperature": 0.0,
        "max_tokens":  6, // Reduzido para 6 pois o maior token é GO_PROGRAMMING (aprox. 4-5 tokens)
    })
    if err != nil {
        return IntentGeneral, fmt.Errorf("tinybrain classification failed: %w", err)
    }

    // Sanitização defensiva da string
    result := strings.ToUpper(strings.TrimSpace(resp.Content))

    // Usando Contains para blindar contra pontuações ou prefixos como "1. GENERAL"
    if strings.Contains(result, "GO_PROGRAMMING") {
        return IntentGoProgramming, nil
    }
    if strings.Contains(result, "CODE_REVIEW") {
        return IntentCodeReview, nil
    }
    if strings.Contains(result, "DEBUGGING") {
        return IntentDebugging, nil
    }
    if strings.Contains(result, "ARCHITECTURE") {
        return IntentArchitecture, nil
    }

    // Qualquer outra coisa (ou falha) cai com segurança no GENERAL
    return IntentGeneral, nil
}
