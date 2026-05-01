package backend

import (
	"context"
	"fmt"

	"ada-love-ai/pkg/agent"
	"ada-love-ai/pkg/logger"
)

// AdaHook implementa os ganchos do Picoclaw para o Ada-Love
type AdaHook struct {
	engine *Engine
}

// OnEvent captura eventos de ciclo de vida do agente
func (h *AdaHook) OnEvent(ctx context.Context, e agent.Event) error {
	switch e.Kind {
	case agent.EventKindTurnStart:
		logger.InfoCF("ada-hooks", "Turn started", map[string]any{
			"agent_id": e.Meta.AgentID,
			"turn_id":  e.Meta.TurnID,
		})
		// Exemplo: Salvar log de início de turno no banco
		if h.engine.db != nil {
			h.engine.db.AddMessageToSession(e.Meta.SessionKey, "system", fmt.Sprintf("Turn started: %s", e.Meta.AgentID))
		}
	case agent.EventKindTurnEnd:
		logger.InfoCF("ada-hooks", "Turn complete", map[string]any{
			"agent_id": e.Meta.AgentID,
			"turn_id":  e.Meta.TurnID,
		})
	}
	return nil
}

// BeforeLLM permite interceptar ou modificar requisições ao LLM
func (h *AdaHook) BeforeLLM(ctx context.Context, req *agent.LLMHookRequest) (*agent.LLMHookRequest, agent.HookDecision, error) {
	return req, agent.HookDecision{Action: agent.HookActionContinue}, nil
}

// AfterLLM permite processar a resposta do LLM
func (h *AdaHook) AfterLLM(ctx context.Context, res *agent.LLMHookResponse) (*agent.LLMHookResponse, agent.HookDecision, error) {
	return res, agent.HookDecision{Action: agent.HookActionContinue}, nil
}

// SetupHooks registra os ganchos do Ada no Engine
func (e *Engine) SetupHooks() {
	// Picoclaw hooks desabilitados temporariamente para focar no chat
	/*
		if e.agentLoop == nil {
			return
		}

		h := &AdaHook{engine: e}

		err := e.agentLoop.MountHook(agent.HookRegistration{
			Name:     "ada_core",
			Priority: 10,
			Source:   agent.HookSourceInProcess,
			Hook:     h,
		})

		if err != nil {
			fmt.Printf("[Engine] Erro ao montar hook: %v\n", err)
		} else {
			fmt.Printf("[Engine] Hooks registrados com sucesso\n")
		}
	*/
	fmt.Printf("[Engine] Hooks desabilitados por solicitação\n")
}
