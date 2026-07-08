package integrationtools

import (
	"context"
	"fmt"
	"sync"
	"time"

	toolshared "ada-love-ai/pkg/tools/shared"
)

// QuestionRegistry manages pending ask_user questions and their response channels.
type QuestionRegistry struct {
	mu       sync.Mutex
	pending  map[string]chan string // sessionID -> response channel
	onAsk    func(sessionID, question string)
	onAnswer func(sessionID string)
	// resolver converts an agent opaque session key (sk_v1_...) to a frontend sessionID.
	resolver func(opaqueKey string) string
}

// NewQuestionRegistry creates a new QuestionRegistry.
func NewQuestionRegistry() *QuestionRegistry {
	return &QuestionRegistry{
		pending: make(map[string]chan string),
	}
}

// SetResolver sets the function used to resolve opaque session keys to frontend sessionIDs.
func (qr *QuestionRegistry) SetResolver(fn func(opaqueKey string) string) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	qr.resolver = fn
}

// resolveSessionKey converts a session key (opaque or ada:-prefixed) to a frontend sessionID.
func (qr *QuestionRegistry) resolveSessionKey(sessionKey string) string {
	qr.mu.Lock()
	resolver := qr.resolver
	qr.mu.Unlock()
	if resolver != nil {
		return resolver(sessionKey)
	}
	return sessionKeyToID(sessionKey)
}

// OnAsk registers a callback for when a question is asked.
func (qr *QuestionRegistry) OnAsk(fn func(sessionID, question string)) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	qr.onAsk = fn
}

// OnAnswer registers a callback for after an answer is received.
func (qr *QuestionRegistry) OnAnswer(fn func(sessionID string)) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	qr.onAnswer = fn
}

// WaitForAnswer blocks until the user responds or the timeout/context expires.
func (qr *QuestionRegistry) WaitForAnswer(ctx context.Context, sessionID, question string, timeout time.Duration) (string, bool) {
	ch := make(chan string, 1)

	qr.mu.Lock()
	qr.pending[sessionID] = ch
	onAsk := qr.onAsk
	qr.mu.Unlock()

	// Notify frontend
	if onAsk != nil {
		onAsk(sessionID, question)
	}

	defer func() {
		qr.mu.Lock()
		delete(qr.pending, sessionID)
		qr.mu.Unlock()
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case answer := <-ch:
		qr.mu.Lock()
		onAnswer := qr.onAnswer
		qr.mu.Unlock()
		if onAnswer != nil {
			onAnswer(sessionID)
		}
		return answer, true
	case <-timer.C:
		return "", false
	case <-ctx.Done():
		return "", false
	}
}

// Respond delivers the user's answer to the waiting tool.
// sessionID can be either a frontend sessionID or an opaque key (resolved automatically).
func (qr *QuestionRegistry) Respond(sessionID, answer string) bool {
	qr.mu.Lock()
	ch, ok := qr.pending[sessionID]
	if !ok && qr.resolver != nil {
		// Try resolving as an opaque key
		resolved := qr.resolver(sessionID)
		if resolved != "" {
			ch, ok = qr.pending[resolved]
		}
	}
	qr.mu.Unlock()

	if !ok {
		return false
	}

	select {
	case ch <- answer:
		return true
	default:
		return false
	}
}

// HasPending returns true if there is a pending question for the session.
func (qr *QuestionRegistry) HasPending(sessionID string) bool {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	_, ok := qr.pending[sessionID]
	return ok
}

// AskUserTool asks a question to the user and waits for their response.
type AskUserTool struct {
	registry *QuestionRegistry
	timeout  time.Duration
}

// NewAskUserTool creates a new AskUserTool.
func NewAskUserTool(registry *QuestionRegistry, timeout time.Duration) *AskUserTool {
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	return &AskUserTool{
		registry: registry,
		timeout:  timeout,
	}
}

func (t *AskUserTool) Name() string {
	return "ask_user"
}

func (t *AskUserTool) Description() string {
	return "Ask the user a question and wait for their response. Use this when you need clarification or information from the user before proceeding."
}

func (t *AskUserTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"question": map[string]any{
				"type":        "string",
				"description": "The question to ask the user",
			},
		},
		"required": []string{"question"},
	}
}

func (t *AskUserTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	question, ok := args["question"].(string)
	if !ok || question == "" {
		return &toolshared.ToolResult{
			ForLLM:  "question is required",
			IsError: true,
		}
	}

	sessionKey := toolshared.ToolSessionKey(ctx)
	if sessionKey == "" {
		return &toolshared.ToolResult{
			ForLLM:  "cannot determine session for ask_user",
			IsError: true,
		}
	}
	sessionID := t.registry.resolveSessionKey(sessionKey)

	answer, answered := t.registry.WaitForAnswer(ctx, sessionID, question, t.timeout)

	if !answered {
		return &toolshared.ToolResult{
			ForLLM:  "The user did not respond in time. Proceed with the information you have.",
			ForUser: "Question timed out.",
			Silent:  true,
		}
	}

	return &toolshared.ToolResult{
		ForLLM:  fmt.Sprintf("User responded: %s", answer),
		ForUser: fmt.Sprintf("User answered: %s", answer),
		Silent:  true,
	}
}

// sessionKeyToID extracts the session ID from a session key (strips "ada:" prefix).
func sessionKeyToID(key string) string {
	if len(key) > 4 && key[:4] == "ada:" {
		return key[4:]
	}
	return key
}
