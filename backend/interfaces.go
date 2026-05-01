package backend

import (
	"context"
)

// TaskRunner define a interface para executores de automação e tarefas.
// O Ada-Love utiliza esta interface para desacoplar a orquestração do chat
// de motores específicos como o PicoClaw.
type TaskRunner interface {
	// Execute executa uma tarefa de forma síncrona.
	Execute(ctx context.Context, prompt string, sessionID string) (string, error)
	
	// ExecuteStream executa uma tarefa e envia deltas de texto via callback.
	ExecuteStream(ctx context.Context, prompt string, sessionID string, onDelta func(string)) (string, error)
	
	// Name retorna o nome do executor (ex: "ada-love").
	Name() string
}
