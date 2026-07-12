package orchestrator

import "fmt"

// GenericAgent is a fallback agent used when no specific agent is configured
// for the requested type. It uses a generic system prompt and delegates to
// the LLM directly.
type GenericAgent struct {
	*BaseSubAgent
}

// GenericAgentConfig holds configuration for the generic fallback agent.
type GenericAgentConfig struct {
	Model         string
	WorkspaceRoot string
}

// NewGenericAgent creates a new generic fallback agent.
func NewGenericAgent(config GenericAgentConfig) *GenericAgent {
	return &GenericAgent{
		BaseSubAgent: NewBaseSubAgent(AgentTypeGeneric, SubAgentConfig{
			Model: config.Model,
			SystemPrompt: fmt.Sprintf(`Você é um assistente genérico de desenvolvimento de software.
Nenhum agente especializado está configurado para esta tarefa.
Responda à melhor da sua capacidade com base no contexto disponível.
Workspace: %s`, config.WorkspaceRoot),
			Capabilities: []AgentCapability{
				CapabilityCodeReview,
			},
			AllowedTools:  []string{},
			MaxIterations: 10,
			Temperature:   0.7,
		}, config.WorkspaceRoot),
	}
}

// WarningMessage returns a user-facing warning that this is a fallback agent.
func (g *GenericAgent) WarningMessage() string {
	return "⚠️ Agente especializado não configurado. Usando agente genérico como fallback."
}
