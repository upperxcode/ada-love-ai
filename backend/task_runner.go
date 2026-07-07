package backend

import (
	"ada-love-ai/pkg/agent"
	"context"
	"fmt"
)

// AdaLoveTaskRunner é uma implementação de TaskRunner que utiliza o AgentLoop do Ada-Love.
type AdaLoveTaskRunner struct {
	loop *agent.AgentLoop
}

func NewAdaLoveTaskRunner(loop *agent.AgentLoop) *AdaLoveTaskRunner {
	return &AdaLoveTaskRunner{
		loop: loop,
	}
}

func (p *AdaLoveTaskRunner) Name() string {
	return "ada-love"
}

func (p *AdaLoveTaskRunner) Execute(ctx context.Context, prompt string, sessionID string) (string, error) {
	if p.loop == nil {
		return "", fmt.Errorf("Ada Love agent loop não inicializado")
	}

	sessionKey := "ada:" + sessionID
	return p.loop.ProcessDirect(ctx, prompt, sessionKey)
}

func (p *AdaLoveTaskRunner) ExecuteStream(ctx context.Context, prompt string, sessionID string, onDelta func(string)) (string, error) {
	if p.loop == nil {
		return "", fmt.Errorf("Ada Love agent loop não inicializado")
	}

	// Atualmente o ProcessDirect do Ada-Love não retorna stream diretamente,
	// mas ele emite eventos no EventBus. O StreamingWrapper cuida disso se estiver ativo.
	// Para o TaskRunner, vamos apenas chamar o ProcessDirect.
	sessionKey := "ada:" + sessionID
	return p.loop.ProcessDirect(ctx, prompt, sessionKey)
}
