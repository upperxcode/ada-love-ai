package orchestrator

import (
	"context"
	"fmt"
	"strings"

	"ada-love-ai/pkg/providers"
)

// BaseSubAgent provides common functionality for sub-agents
type BaseSubAgent struct {
	agentType      AgentType
	model          string
	systemPrompt   string
	capabilities   []AgentCapability
	tools          []string
	maxIterations  int
	temperature    float64
	provider       providers.LLMProvider
	workspaceRoot  string
}

func NewBaseSubAgent(agentType AgentType, config SubAgentConfig, workspaceRoot string) *BaseSubAgent {
	return &BaseSubAgent{
		agentType:     agentType,
		model:         config.Model,
		systemPrompt:  config.SystemPrompt,
		capabilities:  config.Capabilities,
		tools:         config.AllowedTools,
		maxIterations: config.MaxIterations,
		temperature:   config.Temperature,
		workspaceRoot: workspaceRoot,
	}
}

func (b *BaseSubAgent) Type() AgentType {
	return b.agentType
}

func (b *BaseSubAgent) Model() string {
	return b.model
}

func (b *BaseSubAgent) SetModel(model string) {
	b.model = model
}

func (b *BaseSubAgent) SystemPrompt() string {
	return b.systemPrompt
}

func (b *BaseSubAgent) Capabilities() []AgentCapability {
	return b.capabilities
}

func (b *BaseSubAgent) Tools() []string {
	return b.tools
}

func (b *BaseSubAgent) SetProvider(provider providers.LLMProvider) {
	b.provider = provider
}

func (b *BaseSubAgent) Provider() providers.LLMProvider {
	return b.provider
}

func (b *BaseSubAgent) Execute(ctx context.Context, task string, layers PromptLayers) (*AgentResult, error) {
	if b.provider == nil {
		return &AgentResult{
			Success: false,
			Error:   fmt.Errorf("no provider configured for agent %s", b.agentType),
		}, nil
	}

	// Build the full prompt from layers
	fullPrompt := b.buildFullPrompt(layers)

	// Execute with provider
	messages := []providers.Message{
		{Role: "system", Content: layers.SystemPersona},
		{Role: "user", Content: fullPrompt},
	}

	resp, err := b.provider.Chat(ctx, messages, nil, b.model, map[string]any{
		"temperature": b.temperature,
		"max_tokens":  4096,
	})
	if err != nil {
		return &AgentResult{
			Success: false,
			Error:   err,
		}, nil
	}

	return &AgentResult{
		Success: true,
		Output:  resp.Content,
		Metadata: map[string]any{
			"model":        b.model,
			"iterations":   1,
			"tokens_used":  resp.Usage.TotalTokens,
		},
	}, nil
}

func (b *BaseSubAgent) buildFullPrompt(layers PromptLayers) string {
	var parts []string

	if layers.GlobalContext != "" {
		parts = append(parts, "=== CONTEXTO GLOBAL ===\n"+layers.GlobalContext)
	}

	if layers.State != "" {
		parts = append(parts, "=== ESTADO ATUAL ===\n"+layers.State)
	}

	parts = append(parts, "=== TAREFA ===\n"+layers.Task)

	return strings.Join(parts, "\n\n")
}

func (b *BaseSubAgent) Close() error {
	return nil
}