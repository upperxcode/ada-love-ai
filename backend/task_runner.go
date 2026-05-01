package backend

import (
	"context"
	"fmt"
	"ada-love-ai/pkg/agent"
)

// PicoclawTaskRunner é uma implementação de TaskRunner que utiliza o AgentLoop do Picoclaw.
type PicoclawTaskRunner struct {
	loop *agent.AgentLoop
}

func NewPicoclawTaskRunner(loop *agent.AgentLoop) *PicoclawTaskRunner {
	return &PicoclawTaskRunner{
		loop: loop,
	}
}

func (p *PicoclawTaskRunner) Name() string {
	return "ada-love"
}

func (p *PicoclawTaskRunner) Execute(ctx context.Context, prompt string, sessionID string) (string, error) {
	if p.loop == nil {
		return "", fmt.Errorf("Ada Love agent loop não inicializado")
	}

	sessionKey := "ada:" + sessionID
	return p.loop.ProcessDirect(ctx, prompt, sessionKey)
}

func (p *PicoclawTaskRunner) ExecuteStream(ctx context.Context, prompt string, sessionID string, onDelta func(string)) (string, error) {
	if p.loop == nil {
		return "", fmt.Errorf("Ada Love agent loop não inicializado")
	}

	// Atualmente o ProcessDirect do Picoclaw não retorna stream diretamente, 
	// mas ele emite eventos no EventBus. O StreamingWrapper cuida disso se estiver ativo.
	// Para o TaskRunner, vamos apenas chamar o ProcessDirect.
	sessionKey := "ada:" + sessionID
	return p.loop.ProcessDirect(ctx, prompt, sessionKey)
}
