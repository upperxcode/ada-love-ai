package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"ada-love-ai/pkg/providers"
	"golang.org/x/sync/errgroup"
)

// Orchestrator handles multi-agent orchestration with 4-layer prompt architecture
type Orchestrator struct {
	config            OrchestratorConfig
	registry          *AgentRegistry
	orchestratorModel string
	history           []OrchestrationStep
	workspaceRoot     string
	provider          providers.LLMProvider
	mu                sync.RWMutex
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(config OrchestratorConfig, model string, workspaceRoot string) *Orchestrator {
	o := &Orchestrator{
		config:            config,
		registry:          NewAgentRegistry(),
		orchestratorModel: model,
		history:           make([]OrchestrationStep, 0),
		workspaceRoot:     workspaceRoot,
	}

	o.registerDefaultAgents()

	return o
}

// NewOrchestratorWithProvider creates an orchestrator with a specific provider
func NewOrchestratorWithProvider(config OrchestratorConfig, model string, workspaceRoot string, provider providers.LLMProvider) *Orchestrator {
	o := &Orchestrator{
		config:            config,
		registry:          NewAgentRegistry(),
		orchestratorModel: model,
		history:           make([]OrchestrationStep, 0),
		workspaceRoot:     workspaceRoot,
		provider:          provider,
	}

	o.registerDefaultAgents()

	return o
}

// registerDefaultAgents registers the default sub-agents
func (o *Orchestrator) registerDefaultAgents() {
	// GoLang Agent
	goAgent := NewGoLangAgent(GoLangAgentConfig{
		Model:         o.orchestratorModel,
		WorkspaceRoot: o.workspaceRoot,
	})
	o.registry.Register(goAgent)

	// React Agent
	reactAgent := NewReactAgent(ReactAgentConfig{
		Model:         o.orchestratorModel,
		WorkspaceRoot: o.workspaceRoot,
	})
	o.registry.Register(reactAgent)

	// Tester Agent
	testerAgent := NewTesterAgent(TesterAgentConfig{
		Model:         o.orchestratorModel,
		WorkspaceRoot: o.workspaceRoot,
	})
	o.registry.Register(testerAgent)

	// Generic Agent (fallback when no specific agent is configured)
	genericAgent := NewGenericAgent(GenericAgentConfig{
		Model:         o.orchestratorModel,
		WorkspaceRoot: o.workspaceRoot,
	})
	o.registry.Register(genericAgent)
}

// SetSubAgentProviders propaga o provider para todos os sub-agentes registrados.
func (o *Orchestrator) SetSubAgentProviders(provider providers.LLMProvider) {
	if provider == nil {
		return
	}
	for _, agentType := range o.registry.List() {
		if agent, ok := o.registry.Get(agentType); ok {
			agent.SetProvider(provider)
		}
	}
}

// SetSubAgentModel propaga o modelo para todos os sub-agentes registrados.
func (o *Orchestrator) SetSubAgentModel(model string) {
	if model == "" {
		return
	}
	for _, agentType := range o.registry.List() {
		if agent, ok := o.registry.Get(agentType); ok {
			agent.SetModel(model)
		}
	}
	o.orchestratorModel = model
}

// BuildPromptLayers builds the 4-layer prompt architecture
func (o *Orchestrator) BuildPromptLayers(agentType AgentType, task string, state string, relatedFiles []string) PromptLayers {
	agent, ok := o.registry.Get(agentType)
	if !ok {
		return PromptLayers{}
	}

	// Layer A: System Persona (Fixed per agent)
	systemPersona := agent.SystemPrompt()

	// Layer B: Global Context (Project context)
	globalContext := fmt.Sprintf(`Contexto do Projeto: Ada-Love AI
Tipo: multi-agent
Diretório Raiz: %s
Modelo Padrão: %s`, o.workspaceRoot, o.orchestratorModel)

	// Layer C: Short-term Memory / State
	stateContext := ""
	if state != "" {
		stateContext = fmt.Sprintf("Estado Atual / Contexto Recente:\n%s", state)
	}

	// Layer D: Current Task
	taskContext := fmt.Sprintf("Tarefa Atual:\n%s", task)
	if len(relatedFiles) > 0 {
		taskContext += "\n\nArquivos Relacionados:\n" + strings.Join(relatedFiles, "\n")
	}

	return PromptLayers{
		SystemPersona: systemPersona,
		GlobalContext: globalContext,
		State:         stateContext,
		Task:          taskContext,
	}
}

// RoutingPrompt creates the prompt for the orchestrator's routing decision.
// If routingRules is non-empty, it replaces the default hardcoded prompt.
func (o *Orchestrator) RoutingPrompt(userInput string, routingRules string) string {
	if routingRules != "" {
		return routingRules + "\n\nEntrada do usuário:\n" + userInput
	}

	agentsDesc := ""
	for _, agentType := range o.config.AvailableAgents {
		if agent, ok := o.registry.Get(agentType); ok {
			agentsDesc += fmt.Sprintf("- %s: %s\n", agentType, agent.SystemPrompt()[:200])
		}
	}

	return fmt.Sprintf(`Você é o Agente Orquestrador de um sistema de desenvolvimento de software.
Sua única responsabilidade é analisar a requisição do usuário e decidir qual sub-agente deve ser acionado.

Agentes Disponíveis:
%s

Regras de Decisão:
1. Se o usuário pedir para criar/alterar backend (Go, APIs, banco de dados, regras de negócio, concorrência), chame o GOLANG_AGENT.
2. Se o usuário pedir interface, tela, componente React, TypeScript, hooks, estado, chame o REACT_AGENT.
3. Se o usuário pedir para validar, debugar código recém-criado ou escrever testes, chame o TESTER_AGENT.
4. Se a tarefa exigir mais de um passo (ex: "Crie a API e a tela"), quebre em uma lista de execução cronológica.
5. Para tarefas que precisam de validação após implementação, defina "requires_test": true.

FORMATO DE SAÍDA OBRIGATÓRIO (Responda APENAS em JSON estrito):
{
  "reasoning": "Análise rápida do que o usuário pediu...",
  "next_agent": "NOME_DO_AGENTE",
  "task": "Descrição detalhada do que o sub-agente deve fazer",
  "related_files": ["lista_de_arquivos_se_houver"],
  "requires_test": true/false,
  "sub_tasks": [
    {"id": "1", "agent": "GOLANG_AGENT", "task": "criar handler de login", "depends_on": []},
    {"id": "2", "agent": "REACT_AGENT", "task": "criar tela de login", "depends_on": ["1"]},
    {"id": "3", "agent": "TESTER_AGENT", "task": "testar login e autenticação", "depends_on": ["1", "2"]}
  ]
}

Entrada do usuário: %s`, agentsDesc, userInput)
}

// SubTaskEventFunc é chamada para cada evento de sub-task (start, complete, error).
type SubTaskEventFunc func(kind string, st SubTask, err error)

// ExecuteRouting executes the routing decision by invoking sub-agents.
// Independent sub-tasks run concurrently via errgroup.
// onEvent é chamada para cada evento de progresso (pode ser nil).
func (o *Orchestrator) ExecuteRouting(ctx context.Context, decision RoutingDecision, state string, onEvent SubTaskEventFunc) (string, error) {
	// Enforce MaxSubTasks
	if o.config.MaxSubTasks > 0 && len(decision.SubTasks) > o.config.MaxSubTasks {
		decision.SubTasks = decision.SubTasks[:o.config.MaxSubTasks]
	}

	// RequiresTest: injeta sub-task de teste automaticamente
	if decision.RequiresTest && len(decision.SubTasks) > 0 {
		testTaskID := fmt.Sprintf("test-%d", len(decision.SubTasks)+1)
		testDependsOn := make([]string, 0, len(decision.SubTasks))
		for _, st := range decision.SubTasks {
			testDependsOn = append(testDependsOn, st.ID)
		}
		decision.SubTasks = append(decision.SubTasks, SubTask{
			ID:        testTaskID,
			Agent:     AgentTypeTester,
			Task:      fmt.Sprintf("Escreva testes unitários e de integração para a tarefa: %s", decision.Task),
			DependsOn: testDependsOn,
		})
	}

	var finalOutput strings.Builder
	var subTaskResults []SubTaskResult

	if len(decision.SubTasks) > 0 {
		results, _, err := o.executeSubTasksConcurrent(ctx, decision.SubTasks, state, onEvent)
		if err != nil {
			return "", err
		}

		for _, st := range decision.SubTasks {
			if r, ok := results[st.ID]; ok {
				subTaskResults = append(subTaskResults, SubTaskResult{
					SubTask: st,
					Result:  &r,
				})
				finalOutput.WriteString(fmt.Sprintf("\n--- Sub-task %s (%s) ---\n%s\n", st.ID, st.Agent, r.Output))
			}
		}
	} else {
		if onEvent != nil {
			onEvent("start", SubTask{ID: "1", Agent: decision.NextAgent, Task: decision.Task}, nil)
		}

		agent, ok := o.registry.Get(decision.NextAgent)
		if !ok {
			// Agente não configurado — fallback para agente genérico com aviso
			fmt.Printf("[Orchestrator] ⚠️ Agent %q não encontrado no registry, usando fallback genérico\n", decision.NextAgent)
			if genericAgent, gok := o.registry.Get(AgentTypeGeneric); gok {
				agent = genericAgent
			} else {
				return "", fmt.Errorf("agent %s not found and no generic fallback available", decision.NextAgent)
			}
		}

		layers := o.BuildPromptLayers(decision.NextAgent, decision.Task, state, decision.RelatedFiles)
		result, err := agent.Execute(ctx, decision.Task, layers)
		if err != nil {
			if onEvent != nil {
				onEvent("error", SubTask{ID: "1", Agent: decision.NextAgent, Task: decision.Task}, err)
			}
			return "", err
		}

		if onEvent != nil {
			onEvent("complete", SubTask{ID: "1", Agent: decision.NextAgent, Task: decision.Task}, nil)
		}

		finalOutput.WriteString(result.Output)
		subTaskResults = append(subTaskResults, SubTaskResult{
			SubTask: SubTask{ID: "1", Agent: decision.NextAgent, Task: decision.Task},
			Result:  result,
		})
	}

	// Record orchestration step
	o.mu.Lock()
	o.history = append(o.history, OrchestrationStep{
		Timestamp:      time.Now(),
		UserInput:      decision.Task,
		Decision:       decision,
		SubTaskResults: subTaskResults,
		FinalOutput:    finalOutput.String(),
	})
	o.mu.Unlock()

	return finalOutput.String(), nil
}

// executeSubTasksConcurrent runs sub-tasks in parallel batches based on dependencies.
// Returns results map, completed map, and any error.
// Acumula resultados como state para sub-tasks dependentes.
func (o *Orchestrator) executeSubTasksConcurrent(
	ctx context.Context,
	subTasks []SubTask,
	state string,
	onEvent SubTaskEventFunc,
) (map[string]AgentResult, map[string]bool, error) {
	results := make(map[string]AgentResult)
	completed := make(map[string]bool)
	mu := sync.Mutex{}

	remaining := make([]SubTask, len(subTasks))
	copy(remaining, subTasks)

	// Acumula state entre batches
	accumulatedState := state

	for len(remaining) > 0 {
		var batch []SubTask
		var nextRemaining []SubTask
		for _, st := range remaining {
			depsMet := true
			for _, dep := range st.DependsOn {
				if !completed[dep] {
					depsMet = false
					break
				}
			}
			if depsMet {
				batch = append(batch, st)
			} else {
				nextRemaining = append(nextRemaining, st)
			}
		}

		if len(batch) == 0 {
			return results, completed, fmt.Errorf("deadlock: sub-tasks com dependências não resolvíveis restantes: %d", len(nextRemaining))
		}

		// Executa batch em paralelo
		g, ctx := errgroup.WithContext(ctx)
		for _, st := range batch {
			st := st
			g.Go(func() error {
				if onEvent != nil {
					onEvent("start", st, nil)
				}

				agent, ok := o.registry.Get(st.Agent)
				if !ok {
					// Attempt capability-based selection when the explicit agent is not registered.
					desired := agentTypeToCapabilities(st.Agent)
					inferred := decideCapsFromTask(st.Task)
					// merge desired + inferred
					if len(inferred) > 0 {
						desired = append(desired, inferred...)
					}
					if len(desired) > 0 {
						if byCap, found := o.chooseAgentByCapabilities(desired); found {
							agent = byCap
							ok = true
							fmt.Printf("[Orchestrator] Selected sub-agent %s by capabilities for subtask %s\n", agent.Type(), st.ID)
						}
					}
					if !ok {
						// Agente não configurado — fallback para genérico
						fmt.Printf("[Orchestrator] ⚠️ Sub-task agent %q não encontrado, usando fallback genérico\n", st.Agent)
						if genericAgent, gok := o.registry.Get(AgentTypeGeneric); gok {
							agent = genericAgent
						} else {
							err := fmt.Errorf("agent %s not found and no generic fallback available", st.Agent)
							if onEvent != nil {
								onEvent("error", st, err)
							}
							return err
						}
					}
				}

				layers := o.BuildPromptLayers(st.Agent, st.Task, accumulatedState, nil)
				result, err := agent.Execute(ctx, st.Task, layers)
				if err != nil {
					if onEvent != nil {
						onEvent("error", st, err)
					}
					return fmt.Errorf("sub-task %s (%s) failed: %w", st.ID, st.Agent, err)
				}

				mu.Lock()
				results[st.ID] = *result
				completed[st.ID] = true
				// Acumula output como state para batches seguintes
				accumulatedState += fmt.Sprintf("\n\n[Sub-task %s (%s) concluída]:\n%s", st.ID, st.Agent, result.Output)
				mu.Unlock()

				if onEvent != nil {
					onEvent("complete", st, nil)
				}
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return results, completed, err
		}

		remaining = nextRemaining
	}

	return results, completed, nil
}

// RouteAndExecute routes the user input and executes the appropriate sub-agent(s)
func (o *Orchestrator) RouteAndExecute(ctx context.Context, userInput string, state string) (string, error) {
	decision, err := o.LLMRoute(ctx, userInput, "")
	if err != nil {
		fmt.Printf("[Orchestrator] Roteamento LLM falhou, usando heurística: %v\n", err)
		decision = o.heuristicRoute(userInput)
	}
	return o.ExecuteRouting(ctx, decision, state, nil)
}

// RouteAndExecuteWithState routes and executes with a given state
func (o *Orchestrator) RouteAndExecuteWithState(ctx context.Context, userInput string, state string) (string, error) {
	decision, err := o.LLMRoute(ctx, userInput, "")
	if err != nil {
		decision = o.heuristicRoute(userInput)
	}
	return o.ExecuteRouting(ctx, decision, state, nil)
}

// LLMRoute uses the LLM provider to make a routing decision.
// personality is the workspace personality used as routing prompt (empty = default).
// Falls back to heuristicRoute on any error.
func (o *Orchestrator) LLMRoute(ctx context.Context, userInput string, routingRules string) (RoutingDecision, error) {
	if o.provider == nil {
		return o.heuristicRoute(userInput), nil
	}

	prompt := o.RoutingPrompt(userInput, routingRules)
	messages := []providers.Message{
		{Role: "user", Content: prompt},
	}

	resp, err := o.provider.Chat(ctx, messages, nil, o.orchestratorModel, map[string]any{
		"temperature": 0.1,
		"max_tokens":  1024,
	})
	if err != nil {
		return o.heuristicRoute(userInput), fmt.Errorf("LLM routing falhou: %w", err)
	}

	if resp == nil || resp.Content == "" {
		return o.heuristicRoute(userInput), fmt.Errorf("LLM routing retornou resposta vazia")
	}

	// Tenta extrair JSON da resposta (pode vir com markdown code block)
	raw := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(raw, "```") {
		// Remove ```json ... ``` wrapper
		if idx := strings.Index(raw, "\n"); idx != -1 {
			raw = raw[idx+1:]
		}
		if idx := strings.LastIndex(raw, "```"); idx != -1 {
			raw = raw[:idx]
		}
		raw = strings.TrimSpace(raw)
	}

	var decision RoutingDecision
	if err := json.Unmarshal([]byte(raw), &decision); err != nil {
		return o.heuristicRoute(userInput), fmt.Errorf("falha ao parsear JSON do LLM: %w (resposta: %s)", err, raw)
	}

	// Se o JSON foi parseado mas os campos estão vazios, tenta com nomes de campos em português
	if decision.NextAgent == "" && decision.Task == "" {
		ptBR := strings.NewReplacer(
			"pensamento", "reasoning",
			"proximo_agente", "next_agent",
			"tarefa_extraida", "task",
			"arquivos_relacionados", "related_files",
			"requer_testes", "requires_test",
			"sub_tarefas", "sub_tasks",
		)
		translated := ptBR.Replace(raw)
		if translated != raw {
			if err := json.Unmarshal([]byte(translated), &decision); err == nil && decision.NextAgent != "" {
				return decision, nil
			}
		}
	}

	// Normaliza nomes de agent (GOLANG_AGENT → golang, etc.)
	decision.NextAgent = normalizeAgentType(string(decision.NextAgent))
	for i := range decision.SubTasks {
		decision.SubTasks[i].Agent = normalizeAgentType(string(decision.SubTasks[i].Agent))
	}

	return decision, nil
}

// normalizeAgentType mapeia variações de nomes de agent (ex: "GOLANG_AGENT") para o AgentType correto do registry.
func normalizeAgentType(raw string) AgentType {
	s := strings.ToLower(strings.TrimSpace(raw))
	// Remove sufixos comuns
	s = strings.TrimSuffix(s, "_agent")
	s = strings.TrimSuffix(s, "agent")
	s = strings.TrimSpace(s)

	switch s {
	case "golang", "go", "go_lang", "golanger":
		return AgentTypeGoLang
	case "react", "reactjs", "react_agent":
		return AgentTypeReact
	case "tester", "test", "testing", "tester_agent":
		return AgentTypeTester
	case "generic", "default":
		return AgentTypeGeneric
	default:
		return AgentType(s)
	}
}

// chooseAgentByCapabilities selects the best registered SubAgent that matches the
// provided desired capabilities. It scores agents by how many desired capabilities
// they expose and returns the highest-scoring agent (first tie wins).
func (o *Orchestrator) chooseAgentByCapabilities(desired []AgentCapability) (SubAgent, bool) {
	if o == nil || o.registry == nil || len(desired) == 0 {
		return nil, false
	}

	best := SubAgent(nil)
	bestScore := 0
	for _, at := range o.registry.List() {
		sub, ok := o.registry.Get(at)
		if !ok || sub == nil {
			continue
		}
		caps := sub.Capabilities()
		score := 0
		for _, want := range desired {
			for _, have := range caps {
				if have == want {
					score++
					break
				}
			}
		}
		if score > bestScore {
			bestScore = score
			best = sub
		}
	}
	if bestScore > 0 {
		return best, true
	}
	return nil, false
}

// agentTypeToCapabilities maps an AgentType hint to a prioritized list of capabilities
// that are appropriate for that agent type.
func agentTypeToCapabilities(a AgentType) []AgentCapability {
	switch a {
	case AgentTypeGoLang:
		return []AgentCapability{CapabilityGoBackend, CapabilityAPI, CapabilityDatabase}
	case AgentTypeReact:
		return []AgentCapability{CapabilityReactFrontend, CapabilityUIComponents}
	case AgentTypeTester:
		return []AgentCapability{CapabilityTesting, CapabilityQualityAssurance}
	default:
		return nil
	}
}

// decideCapsFromTask returns a best-effort list of capabilities inferred from the task text.
func decideCapsFromTask(task string) []AgentCapability {
	if task == "" {
		return nil
	}
	lower := strings.ToLower(task)
	set := make(map[AgentCapability]bool)
	// Frontend/UI
	if strings.Contains(lower, "react") || strings.Contains(lower, "frontend") || strings.Contains(lower, "tela") || strings.Contains(lower, "ui") || strings.Contains(lower, "component") || strings.Contains(lower, "hook") {
		set[CapabilityReactFrontend] = true
		set[CapabilityUIComponents] = true
	}
	// Database
	if strings.Contains(lower, "db") || strings.Contains(lower, "database") || strings.Contains(lower, "sql") || strings.Contains(lower, "postgres") || strings.Contains(lower, "mysql") {
		set[CapabilityDatabase] = true
	}
	// Testing / QA
	if strings.Contains(lower, "test") || strings.Contains(lower, "teste") || strings.Contains(lower, "validate") || strings.Contains(lower, "bug") || strings.Contains(lower, "debug") {
		set[CapabilityTesting] = true
		set[CapabilityQualityAssurance] = true
	}
	// API
	if strings.Contains(lower, "api") || strings.Contains(lower, "endpoint") || strings.Contains(lower, "rest") || strings.Contains(lower, "graphql") || strings.Contains(lower, "grpc") {
		set[CapabilityAPI] = true
	}
	// Concurrency / performance
	if strings.Contains(lower, "concurr") || strings.Contains(lower, "goroutine") || strings.Contains(lower, "mutex") || strings.Contains(lower, "performance") {
		set[CapabilityConcurrency] = true
	}
	// Code review / refactor
	if strings.Contains(lower, "review") || strings.Contains(lower, "refactor") || strings.Contains(lower, "refator") {
		set[CapabilityCodeReview] = true
	}

	if len(set) == 0 {
		// Default to backend capability
		return []AgentCapability{CapabilityGoBackend}
	}
	out := make([]AgentCapability, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	return out
}

// heuristicRoute uses simple keyword heuristics to route when LLM is unavailable.
func (o *Orchestrator) heuristicRoute(userInput string) RoutingDecision {
	input := strings.ToLower(userInput)

	// Check for React/frontend keywords
	reactKeywords := []string{"react", "frontend", "tela", "interface", "componente", "hook", "jsx", "tsx", "css", "ui", "tela de", "página"}
	for _, kw := range reactKeywords {
		if strings.Contains(input, kw) {
			return RoutingDecision{
				Reasoning:    "Detected frontend/UI related request",
				NextAgent:    AgentTypeReact,
				Task:         "Implement frontend feature",
				RelatedFiles: []string{},
				RequiresTest: true,
				SubTasks:     nil,
			}
		}
	}

	// Check for testing keywords
	testKeywords := []string{"test", "teste", "validad", "valide", "qualidad", "cobertura", "bug", "debug", "review"}
	for _, kw := range testKeywords {
		if strings.Contains(input, kw) {
			return RoutingDecision{
				Reasoning:    "Detected testing/QA related request",
				NextAgent:    AgentTypeTester,
				Task:         "Validate implementation and write tests",
				RelatedFiles: []string{},
				RequiresTest: true,
				SubTasks:     nil,
			}
		}
	}

	// Default to GoLang agent for backend tasks
	return RoutingDecision{
		Reasoning:    "Defaulting to Go backend agent for general development task",
		NextAgent:    AgentTypeGoLang,
		Task:         userInput,
		RelatedFiles: []string{},
		RequiresTest: false,
		SubTasks:     nil,
	}
}

// GetHistory returns the orchestration history
func (o *Orchestrator) GetHistory() []OrchestrationStep {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.history
}

// SetProvider sets the LLM provider for the orchestrator
func (o *Orchestrator) SetProvider(provider providers.LLMProvider) {
	o.provider = provider
}

// SetModel sets the model name used for LLM routing and sub-agents.
func (o *Orchestrator) SetModel(model string) {
	o.orchestratorModel = model
	o.SetSubAgentModel(model)
}

// Config returns the orchestrator configuration
func (o *Orchestrator) Config() OrchestratorConfig {
	return o.config
}

// GetRegistry returns the agent registry (for external validation checks).
func (o *Orchestrator) GetRegistry() *AgentRegistry {
	return o.registry
}
