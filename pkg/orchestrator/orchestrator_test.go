package orchestrator

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"ada-love-ai/pkg/providers"
)

// errMockProvider retorna sempre um erro pré-definido.
type errMockProvider struct {
	err error
}

func (m *errMockProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	return nil, m.err
}

func (m *errMockProvider) GetDefaultModel() string {
	return "mock-model"
}

// mockOrchestratorProvider retorna respostas pré-definidas para testes de roteamento.
type mockOrchestratorProvider struct {
	response string
	err      error
}

func (m *mockOrchestratorProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &providers.LLMResponse{
		Content:   m.response,
		ToolCalls: []providers.ToolCall{},
		Usage:     &providers.UsageInfo{TotalTokens: 100},
	}, nil
}

func (m *mockOrchestratorProvider) GetDefaultModel() string {
	return "mock-orchestrator-model"
}

// --- Testes de LLMRoute ---

func TestLLMRoute_ValidJSON(t *testing.T) {
	decision := RoutingDecision{
		Reasoning:   "Tarefa envolve backend Go",
		NextAgent:   AgentTypeGoLang,
		Task:        "Criar handler HTTP para login",
		RelatedFiles: []string{"handler.go"},
		RequiresTest: true,
	}
	jsonBytes, _ := json.Marshal(decision)

	provider := &mockOrchestratorProvider{response: string(jsonBytes)}
	orch := NewOrchestratorWithProvider(
		DefaultOrchestratorConfig(),
		"mock-model",
		"/tmp/test",
		provider,
	)

	result, err := orch.LLMRoute(context.Background(), "Crie um handler de login em Go", "")
	if err != nil {
		t.Fatalf("LLMRoute retornou erro: %v", err)
	}
	if result.NextAgent != AgentTypeGoLang {
		t.Errorf("esperado AgentTypeGoLang, got %q", result.NextAgent)
	}
	if result.Task != "Criar handler HTTP para login" {
		t.Errorf("task inesperada: %q", result.Task)
	}
	if !result.RequiresTest {
		t.Error("esperado RequiresTest=true")
	}
}

func TestLLMRoute_ValidJSONWithMarkdown(t *testing.T) {
	decision := RoutingDecision{
		Reasoning:  "Frontend React",
		NextAgent:  AgentTypeReact,
		Task:       "Criar componente",
	}
	jsonBytes, _ := json.Marshal(decision)
	// Simula LLM que envolve JSON em code block
	wrapped := "```json\n" + string(jsonBytes) + "\n```"

	provider := &mockOrchestratorProvider{response: wrapped}
	orch := NewOrchestratorWithProvider(
		DefaultOrchestratorConfig(),
		"mock-model",
		"/tmp/test",
		provider,
	)

	result, err := orch.LLMRoute(context.Background(), "Crie um componente React", "")
	if err != nil {
		t.Fatalf("LLMRoute retornou erro: %v", err)
	}
	if result.NextAgent != AgentTypeReact {
		t.Errorf("esperado AgentTypeReact, got %q", result.NextAgent)
	}
}

func TestLLMRoute_InvalidJSON_FallbackToHeuristic(t *testing.T) {
	provider := &mockOrchestratorProvider{response: "I think you should use React for this"}
	orch := NewOrchestratorWithProvider(
		DefaultOrchestratorConfig(),
		"mock-model",
		"/tmp/test",
		provider,
	)

	result, err := orch.LLMRoute(context.Background(), "Crie um componente React", "")
	if err == nil {
		t.Error("esperado erro para JSON inválido")
	}
	// Deve fazer fallback para heurística
	if result.NextAgent != AgentTypeReact {
		t.Errorf("esperado fallback para AgentTypeReact, got %q", result.NextAgent)
	}
}

func TestLLMRoute_NilProvider_FallbackToHeuristic(t *testing.T) {
	orch := NewOrchestrator(
		DefaultOrchestratorConfig(),
		"mock-model",
		"/tmp/test",
	)

	result, err := orch.LLMRoute(context.Background(), "Crie testes unitários", "")
	if err != nil {
		t.Fatalf("LLMRoute com provider nil não deveria retornar erro: %v", err)
	}
	if result.NextAgent != AgentTypeTester {
		t.Errorf("esperado AgentTypeTester, got %q", result.NextAgent)
	}
}

func TestLLMRoute_ProviderError_FallbackToHeuristic(t *testing.T) {
	errProvider := &errMockProvider{err: context.DeadlineExceeded}
	orch := NewOrchestratorWithProvider(
		DefaultOrchestratorConfig(),
		"mock-model",
		"/tmp/test",
		errProvider,
	)

	result, err := orch.LLMRoute(context.Background(), "Crie uma API REST", "")
	if err == nil {
		t.Error("esperado erro para provider error")
	}
	// Fallback para heurística (default é GoLang)
	if result.NextAgent != AgentTypeGoLang {
		t.Errorf("esperado fallback para AgentTypeGoLang, got %q", result.NextAgent)
	}
}

// --- Testes de HeuristicRoute ---

func TestHeuristicRoute_ReactKeywords(t *testing.T) {
	orch := NewOrchestrator(DefaultOrchestratorConfig(), "mock-model", "/tmp/test")

	tests := []struct {
		input string
	}{
		{"Crie um componente React"},
		{"Faça uma tela de login"},
		{"Adicione um hook personalizado"},
		{"Preciso de uma interface de usuário"},
		{"Crie um JSX component"},
		{"Implemente uma página"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := orch.heuristicRoute(tt.input)
			if result.NextAgent != AgentTypeReact {
				t.Errorf("input %q: esperado AgentTypeReact, got %q", tt.input, result.NextAgent)
			}
		})
	}
}

func TestHeuristicRoute_TesterKeywords(t *testing.T) {
	orch := NewOrchestrator(DefaultOrchestratorConfig(), "mock-model", "/tmp/test")

	tests := []struct {
		input string
	}{
		{"Escreva testes unitários"},
		{"Valide o código"},
		{"Faça um review de código"},
		{"Encontre bugs"},
		{"Teste a cobertura"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := orch.heuristicRoute(tt.input)
			if result.NextAgent != AgentTypeTester {
				t.Errorf("input %q: esperado AgentTypeTester, got %q", tt.input, result.NextAgent)
			}
		})
	}
}

func TestHeuristicRoute_DefaultGoLang(t *testing.T) {
	orch := NewOrchestrator(DefaultOrchestratorConfig(), "mock-model", "/tmp/test")

	result := orch.heuristicRoute("Crie um handler HTTP para autenticação")
	if result.NextAgent != AgentTypeGoLang {
		t.Errorf("esperado AgentTypeGoLang, got %q", result.NextAgent)
	}
}

// --- Testes de ExecuteRouting ---

func TestExecuteRouting_SingleAgent(t *testing.T) {
	agentProvider := &mockOrchestratorProvider{response: "Código Go gerado"}
	orch := NewOrchestratorWithProvider(
		DefaultOrchestratorConfig(),
		"mock-model",
		"/tmp/test",
		agentProvider,
	)

	// Configura provider nos agentes registrados
	if agent, ok := orch.registry.Get(AgentTypeGoLang); ok {
		agent.SetProvider(agentProvider)
	}

	decision := RoutingDecision{
		NextAgent: AgentTypeGoLang,
		Task:      "Criar handler HTTP",
	}

	result, err := orch.ExecuteRouting(context.Background(), decision, "", nil)
	if err != nil {
		t.Fatalf("ExecuteRouting falhou: %v", err)
	}
	if result == "" {
		t.Error("resultado vazio")
	}
}

func TestExecuteRouting_WithSubTasks(t *testing.T) {
	provider := &mockOrchestratorProvider{response: "Tarefa concluída"}
	orch := NewOrchestratorWithProvider(
		DefaultOrchestratorConfig(),
		"mock-model",
		"/tmp/test",
		provider,
	)

	// Configura provider nos agentes
	for _, agentType := range []AgentType{AgentTypeGoLang, AgentTypeTester} {
		if agent, ok := orch.registry.Get(agentType); ok {
			agent.SetProvider(provider)
		}
	}

	decision := RoutingDecision{
		NextAgent: AgentTypeGoLang,
		Task:      "Criar API",
		SubTasks: []SubTask{
			{ID: "1", Agent: AgentTypeGoLang, Task: "Criar handler"},
			{ID: "2", Agent: AgentTypeTester, Task: "Testar handler", DependsOn: []string{"1"}},
		},
	}

	result, err := orch.ExecuteRouting(context.Background(), decision, "", nil)
	if err != nil {
		t.Fatalf("ExecuteRouting falhou: %v", err)
	}
	if result == "" {
		t.Error("resultado vazio")
	}
	// Verifica que o histórico foi registrado
	history := orch.GetHistory()
	if len(history) != 1 {
		t.Errorf("esperado 1 item no histórico, got %d", len(history))
	}
}

func TestExecuteRouting_DeadlockDetection(t *testing.T) {
	provider := &mockOrchestratorProvider{response: "ok"}
	orch := NewOrchestratorWithProvider(
		DefaultOrchestratorConfig(),
		"mock-model",
		"/tmp/test",
		provider,
	)

	// Sub-tasks com dependência circular
	decision := RoutingDecision{
		NextAgent: AgentTypeGoLang,
		Task:      "Teste",
		SubTasks: []SubTask{
			{ID: "1", Agent: AgentTypeGoLang, Task: "Tarefa A", DependsOn: []string{"2"}},
			{ID: "2", Agent: AgentTypeTester, Task: "Tarefa B", DependsOn: []string{"1"}},
		},
	}

	_, err := orch.ExecuteRouting(context.Background(), decision, "", nil)
	if err == nil {
		t.Error("esperado erro de deadlock para dependências circulares")
	}
	if !strings.Contains(err.Error(), "deadlock") {
		t.Errorf("esperado erro contendo 'deadlock', got: %v", err)
	}
}

// --- Testes de BuildPromptLayers ---

func TestBuildPromptLayers(t *testing.T) {
	orch := NewOrchestrator(DefaultOrchestratorConfig(), "mock-model", "/tmp/test")

	layers := orch.BuildPromptLayers(AgentTypeGoLang, "Criar handler", "estado anterior", []string{"handler.go"})

	if layers.SystemPersona == "" {
		t.Error("SystemPersona vazio")
	}
	if layers.GlobalContext == "" {
		t.Error("GlobalContext vazio")
	}
	if layers.State != "Estado Atual / Contexto Recente:\nestado anterior" {
		t.Errorf("State inesperado: %q", layers.State)
	}
	if !strings.Contains(layers.Task, "Criar handler") {
		t.Errorf("Task não contém tarefa: %q", layers.Task)
	}
	if !strings.Contains(layers.Task, "handler.go") {
		t.Errorf("Task não contém arquivo: %q", layers.Task)
	}
}

// --- Testes de Registry ---

func TestAgentRegistry_GetAndList(t *testing.T) {
	registry := NewAgentRegistry()

	agent := NewGoLangAgent(GoLangAgentConfig{Model: "test", WorkspaceRoot: "/tmp"})
	registry.Register(agent)

	got, ok := registry.Get(AgentTypeGoLang)
	if !ok {
		t.Fatal("Get retornou false para agente registrado")
	}
	if got.Type() != AgentTypeGoLang {
		t.Errorf("Type inesperado: %q", got.Type())
	}

	types := registry.List()
	if len(types) != 1 {
		t.Errorf("esperado 1 tipo, got %d", len(types))
	}
}

func TestAgentRegistry_GetByCapability(t *testing.T) {
	registry := NewAgentRegistry()

	goAgent := NewGoLangAgent(GoLangAgentConfig{Model: "test", WorkspaceRoot: "/tmp"})
	reactAgent := NewReactAgent(ReactAgentConfig{Model: "test", WorkspaceRoot: "/tmp"})
	registry.Register(goAgent)
	registry.Register(reactAgent)

	goAgents := registry.GetByCapability(CapabilityGoBackend)
	if len(goAgents) != 1 {
		t.Errorf("esperado 1 agente Go, got %d", len(goAgents))
	}

	reactAgents := registry.GetByCapability(CapabilityReactFrontend)
	if len(reactAgents) != 1 {
		t.Errorf("esperado 1 agente React, got %d", len(reactAgents))
	}
}
