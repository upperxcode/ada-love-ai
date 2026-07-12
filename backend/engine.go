package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ada-love-ai/pkg/agent"
	"ada-love-ai/pkg/agent/interfaces"
	"ada-love-ai/pkg/bus"
	"ada-love-ai/pkg/config"
	"ada-love-ai/pkg/orchestrator"
	"ada-love-ai/pkg/providers"
	"ada-love-ai/pkg/skills"
	adatools "ada-love-ai/pkg/tools"
	"ada-love-ai/pkg/tools/integration"
)

const (
	SummaryThreshold = 10 // Começa a sumarizar após 10 mensagens
	SummaryKeepLast  = 4  // Mantém as últimas 4 mensagens após sumarizar
)

type Engine struct {
	cfg        *config.Config
	msgBus     *bus.MessageBus
	agentLoop  *agent.AgentLoop
	mu         sync.RWMutex
	adaCfg     AdaConfig
	SessionMgr *SessionManager
	skillReg   *skills.RegistryManager
	// sessionKeyMap tracks agent opaque session keys (sk_v1_...) back to the
	// original sessionID used by the frontend. Keyed by the opaque session key.
	sessionKeyMap   map[string]string
	sessionKeyMapMu sync.RWMutex
	eventBus        *EventBus
	db              *Store
	toolReg         *adatools.ToolRegistry
	questionReg     *integrationtools.QuestionRegistry
	approvalReg     *integrationtools.ApprovalRegistry
	providerCache   map[string]any
	providerMu      sync.RWMutex
	// overrideModelIDs maps frontend model key (e.g. "OpenRouter/nvidia/...") to the
	// actual model field expected by the provider API (e.g. "nvidia/...").
	overrideModelIDs map[string]string
	overrideModelMu  sync.RWMutex
	// Summarization
	summarizer *SummarizerWorker
	// Orchestrator for multi-agent routing
	orchestrator *orchestrator.Orchestrator
}

func NewEngine() (*Engine, error) {
	configDir := getOSConfigDir()
	fmt.Printf("[Engine] Using config directory: %s\n", configDir)

	// Initialize context logger for tracking what each chat sends as context
	logPath := filepath.Join(configDir, "context_logs.jsonl")
	InitContextLogger(logPath, true)
	fmt.Printf("[Engine] Context logger initialized at: %s\n", logPath)

	// Config base: defaults (zero JSON dependency)
	cfg := config.DefaultConfig()

	// Inicializa o Store (SQLite) no diretório de configuração do SO
	dbPath := filepath.Join(configDir, "ada_love.db")
	db, err := NewStore(dbPath)
	if err != nil {
		fmt.Printf("[Engine] Aviso: Erro ao inicializar banco de dados: %v\n", err)
		// Fallback to local path
		db, err = NewStore("config/ada_love.db")
		if err != nil {
			fmt.Printf("[Engine] Erro fatal ao inicializar banco: %v\n", err)
		}
	}

	// Carrega tudo do DB — fonte de verdade, zero JSON
	var adaCfg AdaConfig

	// --- Migração one-shot: se DB vazio, lê JSON antigo e semeia tabelas normalizadas ---
	if db != nil {
		providers, _ := db.GetProvidersFull()
		if len(providers) == 0 {
			adaConfigPath := filepath.Join(configDir, "ada_config.json")
			if data, err := os.ReadFile(adaConfigPath); err == nil {
				var legacy AdaConfig
				json.Unmarshal(data, &legacy)
				for k, p := range legacy.Providers {
					db.SaveProviderFull(adaptProviderConfig(k, p))
				}
				for _, ws := range legacy.Workspaces {
					db.SaveWorkspace(ws)
				}
				for _, w := range legacy.Workers {
					db.SaveWorker(w)
				}
				for _, a := range legacy.Agents {
					db.SaveAgent(a)
				}
				// Persist fixed models from legacy config
				if legacy.EmbeddingModel != "" {
					db.SaveFixedModelRow(FixedModel{Name: "embedding", Provider: legacy.EmbeddingProvider, Model: legacy.EmbeddingModel})
				}
				if legacy.ImageModel != "" {
					db.SaveFixedModelRow(FixedModel{Name: "image", Provider: legacy.ImageProvider, Model: legacy.ImageModel})
				}
				if legacy.SpecModel != "" || legacy.SpecProvider != "" {
					if id, err := db.SaveFixedModelRow(FixedModel{Name: "spec", Provider: legacy.SpecProvider, Model: legacy.SpecModel}); err == nil {
						if len(legacy.SpecTools) > 0 {
							db.SetFixedModelRowTools(id, legacy.SpecTools)
						}
					}
				}
				// tinybrain
				if legacy.TinyBrain.ModelName != "" || legacy.TinyBrain.Provider != "" {
					if id, err := db.SaveFixedModelRow(FixedModel{Name: "tinybrain", Provider: legacy.TinyBrain.Provider, Model: legacy.TinyBrain.ModelName}); err == nil {
						if len(legacy.TinyBrain.Tools) > 0 {
							db.SetFixedModelRowTools(id, legacy.TinyBrain.Tools)
						}
					}
				}
				// tool profiles
				if legacy.ToolProfiles != nil {
					db.SaveToolProfiles(legacy.ToolProfiles)
				}
				// mcp servers
				if len(legacy.MCPServers) > 0 {
					for name, m := range legacy.MCPServers {
						argsJSON := ""
						if len(m.Args) > 0 {
							if b, err := json.Marshal(m.Args); err == nil {
								argsJSON = string(b)
							}
						}
						envJSON := ""
						if len(m.Env) > 0 {
							if b, err := json.Marshal(m.Env); err == nil {
								envJSON = string(b)
							}
						}
						// merge URL
						if m.URL != "" {
							var em map[string]string
							if envJSON != "" {
								json.Unmarshal([]byte(envJSON), &em)
							}
							if em == nil {
								em = map[string]string{}
							}
							em["__url"] = m.URL
							if b, err := json.Marshal(em); err == nil {
								envJSON = string(b)
							}
						}
						db.SaveMCP(name, "", m.Command, argsJSON, envJSON, m.Color, m.Icon)
					}
				}
				// active workspace
				db.SaveAppState(legacy.ActiveWorkspacePath, legacy.ActiveWorkspaceIndex)
				fmt.Printf("[Engine] Migração one-shot do JSON legado concluída\n")
			}
		}
	}

	// --- Carrega do DB (tabelas normalizadas) ---
	if db != nil {
		// Providers (converte StoredProvider → ProviderConfig)
		stored, err := db.GetProvidersFull()
		if err != nil {
			fmt.Printf("[Engine] Erro ao carregar providers: %v\n", err)
		}
		adaCfg.Providers = make(map[string]ProviderConfig)
		for _, sp := range stored {
			adaCfg.Providers[sp.Name] = deadaptProviderConfig(sp)
		}
		fmt.Printf("[Engine] Carregados %d providers do DB\n", len(adaCfg.Providers))

		// Workspaces
		dbWorkspaces, _ := db.GetWorkspaces()
		if len(dbWorkspaces) > 0 {
			adaCfg.Workspaces = dbWorkspaces
		}
		fmt.Printf("[Engine] Carregados %d workspaces do DB\n", adaCfg.WorkspaceCount())

		// Workers
		if ws, err := db.GetWorkers(); err == nil {
			adaCfg.Workers = ws
		}
		// Agents
		if ag, err := db.GetAgents(); err == nil {
			adaCfg.Agents = ag
		}
		// SpecWizards
		if sw, err := db.GetSpecWizards(); err == nil {
			adaCfg.SpecWizards = sw
		}
		// Embedding / Image / Spec / TinyBrain loaded from fixed_models rows (row-based storage)
		if rows, err := db.ListFixedModelRows(); err == nil {
			for _, r := range rows {
				switch strings.ToLower(r.Name) {
				case "embedding":
					adaCfg.EmbeddingModel = r.Model
					adaCfg.EmbeddingProvider = r.Provider
				case "image":
					adaCfg.ImageModel = r.Model
					adaCfg.ImageProvider = r.Provider
				case "spec":
					adaCfg.SpecModel = r.Model
					adaCfg.SpecProvider = r.Provider
					if tools, err := db.GetFixedModelRowTools(r.ID); err == nil {
						adaCfg.SpecTools = tools
					}
				case "tinybrain":
					adaCfg.TinyBrain.ModelName = r.Model
					adaCfg.TinyBrain.Provider = r.Provider
					if tools, err := db.GetFixedModelRowTools(r.ID); err == nil {
						adaCfg.TinyBrain.Tools = tools
					}
				}
			}
		} else {
			// If fixed_models cannot be read, leave embedding/image/spec empty (we don't fallback to JSON files)
		}
		// ToolProfiles: load from normalized table
		if tps, err := db.GetToolProfiles(); err == nil {
			adaCfg.ToolProfiles = tps
		}
		// MCPServers: load from normalized table
		if mcs, err := db.GetMCPsMap(); err == nil {
			adaCfg.MCPServers = mcs
		}
		// Active workspace: load from typed table
		if path, idx, err := db.GetAppState(); err == nil {
			adaCfg.ActiveWorkspacePath = path
			adaCfg.ActiveWorkspaceIndex = idx
		}
	}

	// Migração e saneamento básico
	if adaCfg.ProviderBases == nil {
		adaCfg.ProviderBases = make(map[string]string)
	}
	if adaCfg.ProviderKeys == nil {
		adaCfg.ProviderKeys = make(map[string]string)
	}
	if adaCfg.ModelSettings == nil {
		adaCfg.ModelSettings = make(map[string]ExtraModelConfig)
	}
	if adaCfg.ToolProfiles == nil {
		adaCfg.ToolProfiles = []ToolProfile{}
	}
	// Garante a existência de um perfil "Default" para a UI de Tools.
	hasDefault := false
	for _, p := range adaCfg.ToolProfiles {
		if p.Name == "Default" {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		adaCfg.ToolProfiles = append([]ToolProfile{{
			ID:    1,
			Name:  "Default",
			Color: "#6b7280",
			Icon:  "🔧",
			Tools: []string{},
		}}, adaCfg.ToolProfiles...)
	}

	msgBus := bus.NewMessageBus()
	eventBus := NewEventBus()

	// Popula adaCfg.ModelList a partir dos providers do DB (tabela provider_models).
	// adaCfg é a fonte de verdade consultada por workspaceHealth, chat_manager,
	// resolveProviderForSession e model_manager — por isso o ModelList deve ser
	// derivado para cá, não apenas para o cfg em memória.
	for provName, provCfg := range adaCfg.Providers {
		keys := config.SimpleSecureStrings(provCfg.GetAllAPIKeys()...)
		for modelName := range provCfg.Models {
			found := false
			for _, existing := range adaCfg.ModelList {
				if existing.ModelName == modelName && existing.Provider == provName {
					found = true
					break
				}
			}
			if found {
				continue
			}
			adaCfg.ModelList = append(adaCfg.ModelList, &config.ModelConfig{
				ModelName: modelName,
				Provider:  provName,
				Model:     modelName,
				APIBase:   provCfg.ApiUrl,
				APIKeys:   keys,
				Enabled:   true,
			})
		}
	}
	// Espelha para o cfg em memória (usado por CreateProvider em runtime).
	cfg.ModelList = adaCfg.ModelList
	fmt.Printf("[Engine] ModelList populado: %d modelos de %d providers (do DB)\n", len(adaCfg.ModelList), len(adaCfg.Providers))

	e := &Engine{
		cfg:              cfg,
		msgBus:           msgBus,
		eventBus:         eventBus,
		adaCfg:           adaCfg,
		SessionMgr:       NewSessionManager(db),
		skillReg:         skills.NewRegistryManagerFromToolsConfig(cfg.Tools.Skills),
		db:               db,
		providerCache:    make(map[string]any),
		overrideModelIDs: make(map[string]string),
		sessionKeyMap:    make(map[string]string),
	}

	e.connectMessageBus()

	// Carrega dados do banco se disponível
	if db != nil {
		// Tenta carregar workspaces do banco
		dbWorkspaces, _ := db.GetWorkspaces()
		fmt.Printf("[Engine] Init: %d workspaces loaded from DB\n", len(dbWorkspaces))
		if len(dbWorkspaces) > 0 {
			e.adaCfg.Workspaces = dbWorkspaces
			for i := range e.adaCfg.Workspaces {
				if e.adaCfg.Workspaces[i].Tools == nil {
					e.adaCfg.Workspaces[i].Tools = []string{"read_file", "write_file", "list_dir", "edit_file"}
				}
			}
		} else {
			fmt.Printf("[Engine] Init: no workspaces in DB, using config file (%d)\n", len(e.adaCfg.Workspaces))
		}

		// Carrega sessões para o workspace ativo
		workspacePath := e.adaCfg.ActiveWorkspacePath
		if workspacePath == "" {
			workspacePath = "default"
			if len(e.adaCfg.Workspaces) > 0 && e.adaCfg.ActiveWorkspaceIndex < len(e.adaCfg.Workspaces) {
				workspacePath = e.adaCfg.Workspaces[e.adaCfg.ActiveWorkspaceIndex].Path
			}
		}
		fmt.Printf("[Engine] Init: activeWorkspace=%q activeIndex=%d\n", e.adaCfg.ActiveWorkspacePath, e.adaCfg.ActiveWorkspaceIndex)
		sessions, _ := db.GetSessions(workspacePath)
		fmt.Printf("[Engine] Init: loaded %d sessions for active workspace %q\n", len(sessions), workspacePath)
		for _, s := range sessions {
			fmt.Printf("[Engine]   session=%q title=%q messages=%d pinned=%v\n",
				s.ID, s.Title, len(s.Messages), s.Pinned)
		}
		e.SessionMgr.LoadSessions(sessions)
	}

	// Migração: sessões sem workspace_path → move para o primeiro workspace
	if e.db != nil {
		// Corrigir workspaces com path vazio
		for i := range e.adaCfg.Workspaces {
			// Fix empty path
			if e.adaCfg.Workspaces[i].Path == "" {
				newPath := strings.ToLower(strings.ReplaceAll(e.adaCfg.Workspaces[i].Title, " ", "_"))
				fmt.Printf("[Engine] Init: fixing workspace %q: path '' → %q\n", e.adaCfg.Workspaces[i].Title, newPath)
				e.adaCfg.Workspaces[i].Path = newPath
				e.db.db.Exec(`UPDATE workspaces SET path = ? WHERE nome = ? AND (path = '' OR path IS NULL)`, newPath, e.adaCfg.Workspaces[i].Title)
			}
			// Fix nil slices
			if e.adaCfg.Workspaces[i].WorkerNames == nil {
				e.adaCfg.Workspaces[i].WorkerNames = []string{}
			}
			if e.adaCfg.Workspaces[i].Tools == nil {
				e.adaCfg.Workspaces[i].Tools = []string{}
			}
		}

		// Migrar sessões órfãs
		var count int
		e.db.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE workspace_path = '' OR workspace_path IS NULL`).Scan(&count)
		if count > 0 {
			fmt.Printf("[Engine] Init: migrating %d orphan sessions with empty workspace_path\n", count)
			if len(e.adaCfg.Workspaces) > 0 {
				firstWS := e.adaCfg.Workspaces[0].Path
				if firstWS == "" {
					firstWS = strings.ToLower(strings.ReplaceAll(e.adaCfg.Workspaces[0].Title, " ", "_"))
				}
				e.db.db.Exec(`UPDATE sessions SET workspace_path = ? WHERE workspace_path = '' OR workspace_path IS NULL`, firstWS)
				fmt.Printf("[Engine] Init: migrated %d sessions to workspace %q\n", count, firstWS)
			}
		}
	}

	// Sincroniza o workspace ativo com a configuração antes de iniciar
	e.syncActiveWorkspaceToAgent()

	// Sincroniza agentes do ada_config.json com cfg.Agents.List
	syncAdaAgentsToConfig(e.adaCfg.Agents, &cfg.Agents.List)

	// Inicializa o AgentLoop — provider pode ser nil se nenhum modelo padrão configurado
	rawProvider, _, err := providers.CreateProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar provider: %w", err)
	}
	var provider providers.LLMProvider
	if rawProvider != nil {
		provider = e.wrapProvider(rawProvider)
	}
	e.questionReg = integrationtools.NewQuestionRegistry()
	e.approvalReg = integrationtools.NewApprovalRegistry()
	e.questionReg.SetResolver(e.resolveSessionID)
	e.approvalReg.SetResolver(e.resolveSessionID)
	// Start summarization worker
	e.summarizer = NewSummarizerWorker(e)
	e.summarizer.Start()

	// Inicializa o Orchestrator (usa personality do workspace + provider do chat)
	orchCfg := orchestrator.DefaultOrchestratorConfig()
	orchCfg.WorkspaceRoot = e.GetActiveWorkspace()
	e.orchestrator = orchestrator.NewOrchestrator(
		orchCfg, "", orchCfg.WorkspaceRoot,
	)
	fmt.Println("[Engine] Orchestrator inicializado (personality-based)")

	// Seed workspace templates se a tabela estiver vazia
	if e.db != nil {
		templates, _ := e.db.GetWorkspaceTemplates()
		if len(templates) == 0 {
			defaultTemplates := []WorkspaceTemplate{
				{
					Name:        "fullstack",
					Description: "Fullstack Go + React + Testes",
					Personality: `Você é o Orquestrador deste workspace fullstack.
Agentes disponíveis: GOLANG_AGENT (backend Go), REACT_AGENT (frontend React), TESTER_AGENT (testes).

Regras de orquestração:
1. Se o pedido envolver backend E frontend, quebre em sub-tasks dependentes (backend primeiro, frontend depois).
2. Após qualquer implementação, gere automaticamente a sub-task de teste.
3. Use o modelo e provider do chat ativo para todas as chamadas LLM.
4. Responda APENAS em JSON com o formato obrigatório do orquestrador.
5. Priorize GOLANG_AGENT para APIs, banco de dados, regras de negócio, concorrência.
6. Priorize REACT_AGENT para interfaces, componentes, hooks, estado, estilos.`,
				},
				{
					Name:        "backend",
					Description: "Backend Go + Testes",
					Personality: `Você é o Orquestrador deste workspace backend (Go).
Agentes disponíveis: GOLANG_AGENT (backend Go), TESTER_AGENT (testes).

Regras de orquestração:
1. Roteie tudo para GOLANG_AGENT, exceto pedidos de teste/debug/validação.
2. Após qualquer implementação, injete sub-task de teste automaticamente.
3. Use o modelo e provider do chat ativo para todas as chamadas LLM.
4. Responda APENAS em JSON com o formato obrigatório do orquestrador.
5. Foque em: APIs REST/gRPC, banco de dados, regras de negócio, concorrência, segurança.`,
				},
				{
					Name:        "frontend",
					Description: "Frontend React + Testes",
					Personality: `Você é o Orquestrador deste workspace frontend (React).
Agentes disponíveis: REACT_AGENT (frontend React), TESTER_AGENT (testes).

Regras de orquestração:
1. Roteie tudo para REACT_AGENT, exceto pedidos de teste/debug/validação.
2. Após qualquer implementação, injete sub-task de teste automaticamente.
3. Use o modelo e provider do chat ativo para todas as chamadas LLM.
4. Responda APENAS em JSON com o formato obrigatório do orquestrador.
5. Foque em: componentes, hooks, estado, estilos, acessibilidade, performance.`,
				},
			}
			for _, t := range defaultTemplates {
				if err := e.db.SaveWorkspaceTemplate(t); err != nil {
					fmt.Printf("[Engine] Aviso: falha ao seed template %q: %v\n", t.Name, err)
				}
			}
			fmt.Printf("[Engine] %d workspace templates seedados\n", len(defaultTemplates))
		}
	}
	e.agentLoop = agent.NewAgentLoop(cfg, msgBus, provider, e)
	e.agentLoop.SetSummarizer(e.summarizer)
	e.agentLoop.SetHealthFunc(e.workspaceHealth)
	e.agentLoop.SetTestConnFunc(e.testConnections)

	// Register the context logger so the agent's pipeline can push the
	// COMPLETE LLM context (system prompt + messages + tools) to our JSONL log.
	agent.RegisterContextLogger(func(sessionKey, agentID, model, mode string, messages []providers.Message, toolDefs []providers.ToolDefinition, userMessage string) {
		LogFullContext(sessionKey, agentID, model, mode, messages, toolDefs, userMessage)
	})
	// Register ask_user tool with all agents
	if e.agentLoop != nil {
		e.agentLoop.RegisterToolForAllAgents(integrationtools.NewAskUserTool(e.questionReg, 0))
		// Mount frontend approval hook for tool execution
		e.agentLoop.MountHook(agent.HookRegistration{
			Name:   "frontend_approval",
			Hook:   NewFrontendApprovalHook(e.approvalReg, 0),
			Source: agent.HookSourceInProcess,
		})
	}

	// Bridge eventos do agente para o backend EventBus (para status no frontend)
	go e.bridgeAgentEvents()
	// Track session key mappings (opaque sk_v1_ -> frontend sessionID)
	go e.trackSessionKeys()

	return e, nil
}

func (e *Engine) wrapProvider(p providers.LLMProvider) providers.LLMProvider {
	wrapper := NewStreamingWrapper(p)
	wrapper.SetEventBus(e.eventBus)
	return wrapper
}

// syncActiveWorkspaceToAgent reconciles the agent defaults with the active
// workspace from the database.
func (e *Engine) syncActiveWorkspaceToAgent() {
	activePath := e.adaCfg.ActiveWorkspacePath
	if activePath == "" {
		return
	}
	e.applyWorkspaceToAgent(activePath)
}

// syncActiveWorkspaceToAgentLocked is the lock-holding variant of
// syncActiveWorkspaceToAgent, for use when the caller already holds e.mu.
func (e *Engine) syncActiveWorkspaceToAgentLocked() {
	activePath := e.adaCfg.ActiveWorkspacePath
	if activePath == "" {
		return
	}
	e.applyWorkspaceToAgentLocked(activePath)
}

// ensureWorkspaceSynced makes sure the agent is bound to the given workspace
// before a turn runs. If the workspace differs from the currently active one,
// it updates the active workspace and reloads the agent loop so the
// ContextBuilder and file tools pick up the correct folders.
func (e *Engine) ensureWorkspaceSynced(workspacePath string) {
	if workspacePath == "" {
		return
	}

	e.mu.RLock()
	current := e.adaCfg.ActiveWorkspacePath
	e.mu.RUnlock()

	if current == workspacePath {
		// Same workspace — nothing to do, the agent loop already uses it.
		return
	}

	fmt.Printf("[Engine] ensureWorkspaceSynced: switching %q → %q\n", current, workspacePath)
	e.SetActiveWorkspace(workspacePath)
}

// syncWorkspaceForTurn updates the live agent loop with the session's workspace
// WITHOUT reloading (which would crash the app mid-turn). It patches
// cfg.Agents.Defaults and calls UpdateWorkspace on the default agent's
// ContextBuilder so the system prompt reflects the correct folders and worker persona.
func (e *Engine) syncWorkspaceForTurn(workspacePath string, sessionID string) {
	if workspacePath == "" {
		return
	}

	e.mu.RLock()
	var ws *WorkspaceConfig
	for i := range e.adaCfg.Workspaces {
		w := &e.adaCfg.Workspaces[i]
		if w.Path == workspacePath || w.Title == workspacePath {
			ws = w
			e.adaCfg.ActiveWorkspaceIndex = i
			e.adaCfg.ActiveWorkspacePath = w.Path
			break
		}
	}
	e.mu.RUnlock()

	if ws == nil {
		fmt.Printf("[Engine] syncWorkspaceForTurn: workspace %q not found\n", workspacePath)
		return
	}

	// Resolve the actual filesystem path (first folder = project root).
	fsPath := ""
	if len(ws.Folders) > 0 {
		fsPath = ws.Folders[0]
	}
	if fsPath == "" {
		fsPath = ws.Path
	}

	// Valida que fsPath é um diretório real
	if info, err := os.Stat(fsPath); err != nil || !info.IsDir() {
		fsPath = ""
		for _, f := range ws.Folders {
			if info, err := os.Stat(f); err == nil && info.IsDir() {
				fsPath = f
				break
			}
		}
		if fsPath == "" {
			fmt.Printf("[Engine] syncWorkspaceForTurn: NENHUM folder válido para %q\n", ws.Title)
			return
		}
	}

	fmt.Printf("[Engine] syncWorkspaceForTurn: title=%q folders=%v → fsPath=%q\n",
		ws.Title, ws.Folders, fsPath)

	// Resolve a persona do worker associado à sessão para usar como system prompt.
	workerPersona := e.resolveWorkerPersona(sessionID)

	// Update cfg defaults so any future agent creation uses the right workspace.
	e.cfg.Agents.Defaults.Workspace = fsPath
	e.cfg.Agents.Defaults.Folders = ws.Folders
	e.cfg.Agents.Defaults.Knowledge = ws.Knowledge

	// Injeta a persona do worker como system prompt no agent genérico.
	if workerPersona != "" {
		e.cfg.Agents.Defaults.Personality = workerPersona
	}

	// Patch the LIVE agent loop's ContextBuilder (no reload) so the system
	// prompt is rebuilt with the new workspace on the next turn.
	if e.agentLoop != nil {
		registry := e.agentLoop.GetRegistry()
		if registry != nil {
			if agent := registry.GetDefaultAgent(); agent != nil && agent.ContextBuilder != nil {
				agent.ContextBuilder.UpdateWorkspace(fsPath, ws.Folders, workerPersona, ws.Knowledge)
			}
		}
	}
}

// applyWorkspaceToAgent looks up a workspace by path or title and copies its
// Folders, Personality and Knowledge into cfg.Agents.Defaults. The Workspace
// field is set to the first real folder (the project root) so the agent's file
// tools resolve paths correctly.
func (e *Engine) applyWorkspaceToAgent(pathOrTitle string) {
	if pathOrTitle == "" {
		return
	}

	e.mu.RLock()
	defer e.mu.RUnlock()
	e.applyWorkspaceToAgentLocked(pathOrTitle)
}

// applyWorkspaceToAgentLocked is the lock-holding variant — caller must hold
// e.mu (read or write).
func (e *Engine) applyWorkspaceToAgentLocked(pathOrTitle string) {
	if pathOrTitle == "" {
		return
	}

	var ws *WorkspaceConfig
	for i := range e.adaCfg.Workspaces {
		w := &e.adaCfg.Workspaces[i]
		if w.Path == pathOrTitle || w.Title == pathOrTitle {
			ws = w
			e.adaCfg.ActiveWorkspaceIndex = i
			e.adaCfg.ActiveWorkspacePath = w.Path
			break
		}
	}

	if ws == nil {
		fmt.Printf("[Engine] applyWorkspaceToAgent: workspace %q not found\n", pathOrTitle)
		return
	}

	// cfg.Agents.Defaults.Workspace must be the actual filesystem path so the
	// agent's file tools and context builder resolve paths correctly. Prefer
	// the first real folder (the project root); fall back to the workspace slug.
	fsPath := ""
	if len(ws.Folders) > 0 {
		fsPath = ws.Folders[0]
	}
	if fsPath == "" {
		fsPath = ws.Path
	}

	// Valida que fsPath é um diretório real no filesystem
	if info, err := os.Stat(fsPath); err != nil || !info.IsDir() {
		// Se o path não é um diretório válido, tenta cada folder individualmente
		fsPath = ""
		for _, f := range ws.Folders {
			if info, err := os.Stat(f); err == nil && info.IsDir() {
				fsPath = f
				break
			}
		}
		if fsPath == "" {
			fmt.Printf("[Engine] applyWorkspaceToAgent: NENHUM folder válido encontrado para workspace %q (slug=%q, folders=%v)\n",
				ws.Title, ws.Path, ws.Folders)
			return
		}
	}

	fmt.Printf("[Engine] applyWorkspaceToAgent: title=%q slug=%q folders=%v → fsPath=%q\n",
		ws.Title, ws.Path, ws.Folders, fsPath)

	e.cfg.Agents.Defaults.Workspace = fsPath
	e.cfg.Agents.Defaults.Folders = ws.Folders
	// Personality do worker é injetada por syncWorkspaceForTurn (tem sessionID).
	e.cfg.Agents.Defaults.Knowledge = ws.Knowledge
}

func (e *Engine) connectMessageBus() {
	go func() {
		outbound := e.msgBus.OutboundChan()
		for msg := range outbound {
			kind, _ := msg.Context.Raw["kind"]
			if kind == "tool_feedback" {
				// Extrai o sessionID da chave de sessão (formato "ada:ID")
				sessionID := ""
				if strings.HasPrefix(msg.SessionKey, "ada:") {
					sessionID = strings.TrimPrefix(msg.SessionKey, "ada:")
				}

				e.eventBus.Emit(Event{
					Kind:      EventKindStatus,
					SessionID: sessionID,
					Payload: StatusPayload{
						Message: msg.Content,
					},
					Time: time.Now(),
				})
			}
		}
	}()
}

func (e *Engine) UpdateWorkspaceConfig(fn func(cfg *AdaConfig)) {
	e.mu.Lock()
	fn(&e.adaCfg)
	e.mu.Unlock()
	e.SaveAdaConfig()
	e.ReloadAgentLoop()
}

func (e *Engine) GetAdaConfig() AdaConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.adaCfg
}

func (e *Engine) SetAdaConfig(cfg AdaConfig) {
	e.mu.Lock()
	e.adaCfg = cfg
	e.mu.Unlock()
	e.SaveAdaConfig()
	// Write-through: persiste providers no DB (tabelas normalizadas)
	if e.db != nil && len(cfg.Providers) > 0 {
		for name, pc := range cfg.Providers {
			if err := e.db.SaveProviderFull(adaptProviderConfig(name, pc)); err != nil {
				fmt.Printf("[Engine] Erro ao sincronizar provider %q no DB: %v\n", name, err)
			}
		}
	}
}

// GetWorkers retorna os workers configurados.
func (e *Engine) GetWorkers() []WorkerConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.adaCfg.Workers
}

// --- Workspace Templates ---

func (e *Engine) GetWorkspaceTemplates() ([]WorkspaceTemplate, error) {
	if e.db == nil {
		return nil, fmt.Errorf("banco de dados não inicializado")
	}
	return e.db.GetWorkspaceTemplates()
}

func (e *Engine) SaveWorkspaceTemplate(t WorkspaceTemplate) error {
	if e.db == nil {
		return fmt.Errorf("banco de dados não inicializado")
	}
	return e.db.SaveWorkspaceTemplate(t)
}

func (e *Engine) DeleteWorkspaceTemplate(id int64) error {
	if e.db == nil {
		return fmt.Errorf("banco de dados não inicializado")
	}
	return e.db.DeleteWorkspaceTemplate(id)
}

// SetWorkers substitui a lista de workers e persiste na tabela workers.
func (e *Engine) SetWorkers(workers []WorkerConfig) {
	e.mu.Lock()
	e.adaCfg.Workers = workers
	e.mu.Unlock()
	for _, w := range workers {
		if _, err := e.db.SaveWorker(w); err != nil {
			fmt.Printf("[Engine] Erro ao salvar worker %q: %v\n", w.Name, err)
		}
	}
}

// GetWorkerCategories retorna as categorias de workers.
func (e *Engine) GetWorkerCategories() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.adaCfg.WorkerCategories
}

// SetWorkerCategories substitui as categorias e persiste.
func (e *Engine) SetWorkerCategories(categories []string) {
	e.mu.Lock()
	e.adaCfg.WorkerCategories = categories
	e.mu.Unlock()
	e.SaveAdaConfig()
}

// GetAgents retorna os agentes configurados.
func (e *Engine) GetAgents() []AgentConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.adaCfg.Agents
}

// SetAgents substitui a lista de agentes e persiste na tabela agents.
func (e *Engine) SetAgents(agents []AgentConfig) {
	e.mu.Lock()
	e.adaCfg.Agents = agents
	e.mu.Unlock()
	for _, a := range agents {
		if _, err := e.db.SaveAgent(a); err != nil {
			fmt.Printf("[Engine] Erro ao salvar agent %q: %v\n", a.Name, err)
		}
	}
}

// GetAgentCategories retorna as categorias de agentes.
func (e *Engine) GetAgentCategories() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.adaCfg.AgentCategories
}

// SetAgentCategories substitui as categorias e persiste.
func (e *Engine) SetAgentCategories(categories []string) {
	e.mu.Lock()
	e.adaCfg.AgentCategories = categories
	e.mu.Unlock()
	e.SaveAdaConfig()
}

// FixWorkspacePaths corrige workspaces com path vazio
func (e *Engine) FixWorkspacePaths() {
	e.mu.Lock()
	fixed := false
	for i := range e.adaCfg.Workspaces {
		if e.adaCfg.Workspaces[i].Path == "" {
			newPath := strings.ToLower(strings.ReplaceAll(e.adaCfg.Workspaces[i].Title, " ", "_"))
			fmt.Printf("[Engine] FixWorkspacePaths: fixing %q: '' → %q\n", e.adaCfg.Workspaces[i].Title, newPath)
			e.adaCfg.Workspaces[i].Path = newPath
			fixed = true
		}
	}
	e.mu.Unlock()
	if fixed {
		e.SaveAdaConfig()
	}
}

func (e *Engine) SaveAdaConfig() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.db == nil {
		return fmt.Errorf("banco de dados não disponível")
	}

	// Salva providers no DB (tabelas normalizadas)
	for name, pc := range e.adaCfg.Providers {
		if err := e.db.SaveProviderFull(adaptProviderConfig(name, pc)); err != nil {
			fmt.Printf("[Engine] Erro ao salvar provider %q no DB: %v\n", name, err)
		}
	}

	// Salva spec-wizards nas tabelas próprias (fazer antes de gravar workspaces que referenciam spec_wizard_id)
	for _, sw := range e.adaCfg.SpecWizards {
		if err := e.db.SaveSpecWizard(sw); err != nil {
			fmt.Printf("[Engine] Erro ao salvar spec-wizard %q: %v\n", sw.Name, err)
		}
	}

	// Salva workspaces no DB e sincroniza junctions
	for _, ws := range e.adaCfg.Workspaces {
		id, err := e.db.SaveWorkspace(ws)
		if err != nil {
			fmt.Printf("[Engine] Erro ao salvar workspace %q: %v\n", ws.Title, err)
			continue
		}
		// Resolve IDs de workers/agents pelos nomes e grava junctions
		var wids []int64
		for _, wn := range ws.WorkerNames {
			if w, err := e.db.GetWorkerByName(wn); err == nil && w != nil {
				wids = append(wids, w.ID)
			}
		}
		e.db.SetWorkspaceWorkers(id, wids)
		var aids []int64
		for _, an := range ws.Agents {
			if a, err := e.db.GetAgentByName(an); err == nil && a != nil {
				aids = append(aids, a.ID)
			}
		}
		e.db.SetWorkspaceAgents(id, aids)
		// Persiste folders, knowledge, skills e tools via junction tables
		e.db.SetWorkspaceFolders(id, ws.Folders)
		e.db.SetWorkspaceKnowledge(id, ws.Knowledge)
		// Resolve and persist skills (names -> ids)
		var sids []int64
		for _, sname := range ws.Skills {
			// Try to get existing skill id
			sid, err := e.db.GetSkillIDByName(sname)
			if err != nil || sid == 0 {
				// create a minimal skill record
				if _, err := e.db.SaveSkill(sname, "", "", ""); err != nil {
					continue
				}
				if sid, err = e.db.GetSkillIDByName(sname); err != nil {
					continue
				}
			}
			sids = append(sids, sid)
		}
		if len(sids) > 0 {
			e.db.SetWorkspaceSkills(id, sids)
		}
		// Persist tools (names)
		e.db.SetWorkspaceTools(id, ws.Tools)
	}

	// Salva workers/agents nas tabelas próprias
	for _, w := range e.adaCfg.Workers {
		if _, err := e.db.SaveWorker(w); err != nil {
			fmt.Printf("[Engine] Erro ao salvar worker %q: %v\n", w.Name, err)
		}
	}
	for _, a := range e.adaCfg.Agents {
		if _, err := e.db.SaveAgent(a); err != nil {
			fmt.Printf("[Engine] Erro ao salvar agent %q: %v\n", a.Name, err)
		}
	}

	// Salva seções restantes no DB (key-value / normalizadas)

	// Persist small legacy keys into normalized fixed_models as appropriate
	if e.adaCfg.EmbeddingModel != "" {
		if _, err := e.db.SaveFixedModelRow(FixedModel{Name: "embedding", Provider: e.adaCfg.EmbeddingProvider, Model: e.adaCfg.EmbeddingModel}); err != nil {
			fmt.Printf("[Engine] Warn: failed to persist embedding fixed model: %v\n", err)
		}
	}
	if e.adaCfg.ImageModel != "" {
		if _, err := e.db.SaveFixedModelRow(FixedModel{Name: "image", Provider: e.adaCfg.ImageProvider, Model: e.adaCfg.ImageModel}); err != nil {
			fmt.Printf("[Engine] Warn: failed to persist image fixed model: %v\n", err)
		}
	}
	if e.adaCfg.SpecModel != "" || e.adaCfg.SpecProvider != "" {
		if id, err := e.db.SaveFixedModelRow(FixedModel{Name: "spec", Provider: e.adaCfg.SpecProvider, Model: e.adaCfg.SpecModel}); err == nil {
			if len(e.adaCfg.SpecTools) > 0 {
				e.db.SetFixedModelRowTools(id, e.adaCfg.SpecTools)
			}
		} else {
			fmt.Printf("[Engine] Warn: failed to persist spec fixed model: %v\n", err)
		}
	}
	if len(e.adaCfg.SpecTools) > 0 {
		// spec tools handled above per fixed model
	}

	// Persist tool profiles in normalized tables
	if len(e.adaCfg.ToolProfiles) > 0 {
		if err := e.db.SaveToolProfiles(e.adaCfg.ToolProfiles); err != nil {
			fmt.Printf("[Engine] Warn: failed to persist tool_profiles: %v\n", err)
		}
	}
	// Persist MCP servers in normalized table
	for name, m := range e.adaCfg.MCPServers {
		argsJSON := ""
		if len(m.Args) > 0 {
			if b, err := json.Marshal(m.Args); err == nil {
				argsJSON = string(b)
			}
		}
		envJSON := ""
		if len(m.Env) > 0 {
			if b, err := json.Marshal(m.Env); err == nil {
				envJSON = string(b)
			}
		}
		// merge URL into env under __url
		if m.URL != "" {
			var em map[string]string
			if envJSON != "" {
				json.Unmarshal([]byte(envJSON), &em)
			}
			if em == nil {
				em = map[string]string{}
			}
			em["__url"] = m.URL
			if b, err := json.Marshal(em); err == nil {
				envJSON = string(b)
			}
		}
		if _, err := e.db.SaveMCP(name, "", m.Command, argsJSON, envJSON, m.Color, m.Icon); err != nil {
			fmt.Printf("[Engine] Warn: failed to persist MCP %s: %v\n", name, err)
		}
	}
	// Persist active workspace into normalized app_state table
	if err := e.db.SaveAppState(e.adaCfg.ActiveWorkspacePath, e.adaCfg.ActiveWorkspaceIndex); err != nil {
		fmt.Printf("[Engine] Warn: failed to persist app state: %v\n", err)
	}

	// Also persist fixed_models rows for embedding, image, spec and tinybrain
	if e.db != nil {
		// embedding
		if e.adaCfg.EmbeddingModel != "" {
			if _, err := e.db.SaveFixedModelRow(FixedModel{Name: "embedding", Provider: e.adaCfg.EmbeddingProvider, Model: e.adaCfg.EmbeddingModel}); err != nil {
				fmt.Printf("[Engine] Warn: failed to persist embedding fixed model: %v\n", err)
			}
		}
		// image
		if e.adaCfg.ImageModel != "" {
			if _, err := e.db.SaveFixedModelRow(FixedModel{Name: "image", Provider: e.adaCfg.ImageProvider, Model: e.adaCfg.ImageModel}); err != nil {
				fmt.Printf("[Engine] Warn: failed to persist image fixed model: %v\n", err)
			}
		}
		// spec
		if e.adaCfg.SpecModel != "" || e.adaCfg.SpecProvider != "" {
			if id, err := e.db.SaveFixedModelRow(FixedModel{Name: "spec", Provider: e.adaCfg.SpecProvider, Model: e.adaCfg.SpecModel}); err == nil {
				// set tools if present
				if len(e.adaCfg.SpecTools) > 0 {
					if err := e.db.SetFixedModelRowTools(id, e.adaCfg.SpecTools); err != nil {
						fmt.Printf("[Engine] Warn: failed to persist spec tools: %v\n", err)
					}
				}
			} else {
				fmt.Printf("[Engine] Warn: failed to persist spec fixed model: %v\n", err)
			}
		}
		// tinybrain
		if e.adaCfg.TinyBrain.ModelName != "" || e.adaCfg.TinyBrain.Provider != "" {
			if id, err := e.db.SaveFixedModelRow(FixedModel{Name: "tinybrain", Provider: e.adaCfg.TinyBrain.Provider, Model: e.adaCfg.TinyBrain.ModelName}); err == nil {
				if len(e.adaCfg.TinyBrain.Tools) > 0 {
					if err := e.db.SetFixedModelRowTools(id, e.adaCfg.TinyBrain.Tools); err != nil {
						fmt.Printf("[Engine] Warn: failed to persist tinybrain tools: %v\n", err)
					}
				}
			} else {
				fmt.Printf("[Engine] Warn: failed to persist tinybrain fixed model: %v\n", err)
			}
		}
	}

	// Salva spec-wizards nas tabelas próprias
	for _, sw := range e.adaCfg.SpecWizards {
		if err := e.db.SaveSpecWizard(sw); err != nil {
			fmt.Printf("[Engine] Erro ao salvar spec-wizard %q: %v\n", sw.Name, err)
		}
	}

	return nil
}

// SaveProvidersConfig saves providers to SQLite (providers table in config.db)
func (e *Engine) SaveProvidersConfig() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.db == nil {
		return fmt.Errorf("banco de dados não disponível")
	}
	for name, pc := range e.adaCfg.Providers {
		if err := e.db.SaveProviderFull(adaptProviderConfig(name, pc)); err != nil {
			return err
		}
	}
	return nil
}

// DB retorna o Store para consultas diretas (ex: GetSessions).
func (e *Engine) DB() *Store {
	return e.db
}

// GetProvidersConfig returns the current providers from memory
func (e *Engine) GetProvidersConfig() map[string]ProviderConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.adaCfg.Providers
}

func (e *Engine) ReloadAgentLoop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Sincroniza o workspace ativo com a configuração global do agente
	e.syncActiveWorkspaceToAgentLocked()

	if e.agentLoop != nil {
		e.agentLoop.Stop()
	}

	rawProvider, modelID, err := providers.CreateProvider(e.cfg)
	if err != nil {
		return err
	}
	if modelID != "" {
		e.cfg.Agents.Defaults.ModelName = modelID
	}

	provider := e.wrapProvider(rawProvider)
	e.agentLoop = agent.NewAgentLoop(e.cfg, e.msgBus, provider, e)

	// Re-register health function on the reloaded agent loop
	e.agentLoop.SetHealthFunc(e.workspaceHealth)
	e.agentLoop.SetTestConnFunc(e.testConnections)
	e.agentLoop.SetTestConnFunc(e.testConnections)

	// Re-register ask_user tool and frontend approval hook on the new loop
	if e.agentLoop != nil {
		e.agentLoop.RegisterToolForAllAgents(integrationtools.NewAskUserTool(e.questionReg, 0))
		e.agentLoop.MountHook(agent.HookRegistration{
			Name:   "frontend_approval",
			Hook:   NewFrontendApprovalHook(e.approvalReg, 0),
			Source: agent.HookSourceInProcess,
		})
	}

	// Re-subscribe the agent event bridge to the new loop's event bus
	go e.bridgeAgentEvents()

	return nil
}

func (e *Engine) SubscribeEvents(handler func(Event)) int {
	return e.eventBus.Subscribe(handler)
}

func (e *Engine) UnsubscribeEvents(id int) {
	e.eventBus.Unsubscribe(id)
}

// bridgeAgentEvents traduz eventos do agent loop para o backend EventBus
// para que cheguem ao frontend via Wails runtime.
func (e *Engine) bridgeAgentEvents() {
	if e.agentLoop == nil || e.eventBus == nil {
		return
	}
	sub := e.agentLoop.SubscribeEvents(64)
	for evt := range sub.C {
		sessionID := e.resolveSessionID(evt.Meta.SessionKey)
		if sessionID == "" {
			continue
		}

		switch evt.Kind {
		case agent.EventKindLLMRequest:
			e.eventBus.Emit(Event{
				Kind:      EventKindStatus,
				SessionID: sessionID,
				Payload:   StatusPayload{Message: "thinking"},
			})
		case agent.EventKindToolExecStart:
			if p, ok := evt.Payload.(agent.ToolExecStartPayload); ok {
				e.eventBus.Emit(Event{
					Kind:      EventKindStatus,
					SessionID: sessionID,
					Payload:   StatusPayload{Message: "tool:" + p.Tool},
				})
			}
		case agent.EventKindToolExecEnd:
			e.eventBus.Emit(Event{
				Kind:      EventKindStatus,
				SessionID: sessionID,
				Payload:   StatusPayload{Message: "writing"},
			})
		case agent.EventKindSubTurnSpawn:
			if p, ok := evt.Payload.(agent.SubTurnSpawnPayload); ok {
				// Resolve a friendly agent label from the agent registry if available.
				label := p.Label
				if e.agentLoop != nil {
					if ag, agOK := e.agentLoop.GetRegistry().GetAgent(p.AgentID); agOK && ag.Name != "" {
						label = ag.Name
					}
				}
				fmt.Printf("[Bridge] SubTurn SPAWN: session=%q agent=%q label=%q\n", sessionID, p.AgentID, label)
				e.eventBus.Emit(Event{
					Kind:      EventKindStatus,
					SessionID: sessionID,
					Payload:   StatusPayload{Message: "agent:" + label},
				})
			}
		case agent.EventKindSubTurnEnd:
			if p, ok := evt.Payload.(agent.SubTurnEndPayload); ok {
				label := p.Label
				if label == "" {
					label = p.AgentID
				}
				if e.agentLoop != nil {
					if ag, agOK := e.agentLoop.GetRegistry().GetAgent(p.AgentID); agOK && ag.Name != "" {
						label = ag.Name
					}
				}
				fmt.Printf("[Bridge] SubTurn END: session=%q agent=%q label=%q status=%q\n", sessionID, p.AgentID, label, p.Status)
				status := "writing"
				if p.Status == "error" {
					status = "agent_error"
				} else if p.Status == "completed" {
					status = "agent_done"
				}
				e.eventBus.Emit(Event{
					Kind:      EventKindStatus,
					SessionID: sessionID,
					Payload:   StatusPayload{Message: status},
				})
			}
		case agent.EventKindCommandResult:
			if p, ok := evt.Payload.(agent.CommandResultPayload); ok {
				// Commands resolve before a TurnStart, so the opaque→real session
				// mapping may not exist yet. Fall back to the in-flight
				// SendMessage sessionID when resolution yields the raw key.
				cmdSessionID := sessionID
				if cmdSessionID == "" || strings.HasPrefix(cmdSessionID, "sk_v1_") {
					if sid := e.takePendingSessionID(); sid != "" {
						cmdSessionID = sid
					}
				}
				if cmdSessionID == "" {
					continue
				}
				e.eventBus.Emit(Event{
					Kind:      EventKindCommandResult,
					SessionID: cmdSessionID,
					Payload:   CommandResultPayload{Command: p.Command, Output: p.Output},
				})
			}
		}
	}
}

// SaveSessionDB persiste a sessão atual no SQLite.
func (e *Engine) SaveSessionDB(sessionID string) {
	if sess, ok := e.SessionMgr.sessions[sessionID]; ok && e.db != nil {
		if err := e.db.SaveSession(*sess); err != nil {
			fmt.Printf("[Engine] SaveSessionDB: error saving session %q: %v\n", sessionID, err)
		}
	}
}

// AnswerQuestion delivers the user's answer to a pending ask_user question.
func (e *Engine) AnswerQuestion(sessionID, answer string) {
	if e.questionReg != nil {
		e.questionReg.Respond(sessionID, answer)
	}
}

// trackSessionKeys watches agent events and records opaque session key (sk_v1_...)
// to frontend sessionID mappings. It uses the backend EventBus's TurnStart event
// (which carries the correct SessionID) to correlate with the agent's opaque key.
func (e *Engine) trackSessionKeys() {
	if e.agentLoop == nil {
		return
	}
	sub := e.agentLoop.SubscribeEvents(16)
	for evt := range sub.C {
		if evt.Kind != agent.EventKindTurnStart {
			continue
		}
		opaqueKey := evt.Meta.SessionKey
		if opaqueKey == "" || !strings.HasPrefix(opaqueKey, "sk_v1_") {
			continue
		}
		// The context value "session_id" was set by SendMessage. We can't read it
		// from events, so we rely on the fact that only one SendMessage runs at a
		// time per session. We check pendingSessionTrackers.
		e.sessionKeyMapMu.RLock()
		_, known := e.sessionKeyMap[opaqueKey]
		e.sessionKeyMapMu.RUnlock()
		if known {
			continue
		}
		// Try to find the sessionID from the pending tracker
		if sid := e.takePendingSessionID(); sid != "" {
			e.trackSessionKey(opaqueKey, sid)
		}
	}
}

// pendingSessionID holds the sessionID of the in-flight SendMessage call.
// This is simple and works because the frontend sends messages one at a time.
var pendingSessionID atomic.Value

func (e *Engine) setPendingSessionID(sid string) {
	pendingSessionID.Store(sid)
}

func (e *Engine) takePendingSessionID() string {
	v := pendingSessionID.Load()
	if v == nil {
		return ""
	}
	sid, ok := v.(string)
	if !ok {
		return ""
	}
	return sid
}

// QuestionRegistry returns the question registry for the App to connect callbacks.
func (e *Engine) QuestionRegistry() *integrationtools.QuestionRegistry {
	return e.questionReg
}

// ApprovalRegistry returns the approval registry for the App to connect callbacks.
func (e *Engine) ApprovalRegistry() *integrationtools.ApprovalRegistry {
	return e.approvalReg
}

// AnswerApproval delivers the user's approval decision to a pending tool approval.
func (e *Engine) AnswerApproval(requestID string, approved bool, reason string) {
	if e.approvalReg != nil {
		e.approvalReg.Respond(requestID, approved, reason)
	}
}

// StopGeneration aborts the current turn for the given session.
func (e *Engine) StopGeneration(sessionID string) {
	if e.agentLoop != nil && sessionID != "" {
		// Try the opaque key first (what the agent actually uses internally)
		if opaqueKey := e.resolveOpaqueKey(sessionID); opaqueKey != "" {
			_ = e.agentLoop.HardAbort(opaqueKey)
			return
		}
		_ = e.agentLoop.HardAbort("ada:" + sessionID)
	}
}

// trackSessionKey records the mapping from agent opaque session key to frontend sessionID.
func (e *Engine) trackSessionKey(opaqueKey, sessionID string) {
	if opaqueKey == "" || sessionID == "" {
		return
	}
	e.sessionKeyMapMu.Lock()
	e.sessionKeyMap[opaqueKey] = sessionID
	e.sessionKeyMapMu.Unlock()
}

// resolveSessionID maps an agent opaque session key (sk_v1_...) back to the
// frontend sessionID. Falls back to stripping "ada:" prefix for legacy keys.
func (e *Engine) resolveSessionID(opaqueKey string) string {
	if opaqueKey == "" {
		return ""
	}
	e.sessionKeyMapMu.RLock()
	sessionID, ok := e.sessionKeyMap[opaqueKey]
	e.sessionKeyMapMu.RUnlock()
	if ok {
		return sessionID
	}
	if strings.HasPrefix(opaqueKey, "ada:") {
		return opaqueKey[4:]
	}
	return opaqueKey
}

// resolveOpaqueKey maps a frontend sessionID to the agent opaque key (sk_v1_...).
func (e *Engine) resolveOpaqueKey(sessionID string) string {
	e.sessionKeyMapMu.RLock()
	defer e.sessionKeyMapMu.RUnlock()
	for opaque, sid := range e.sessionKeyMap {
		if sid == sessionID {
			return opaque
		}
	}
	return ""
}

// Implementação de interfaces.MemoryStore para o agente

func (e *Engine) SaveMemory(workspacePath string, content string, importance int) error {
	return e.db.SaveMemory(workspacePath, content, importance)
}

func (e *Engine) GetMemories(workspacePath string) ([]interfaces.MemoryEntry, error) {
	return e.db.GetMemories(workspacePath)
}

func (e *Engine) Close() {
	if e.db != nil {
		e.db.Close()
	}
	// Close the context logger
	if cl := GetContextLogger(); cl != nil {
		cl.Close()
	}
}

// extractProtocol extracts the protocol and model ID from a ModelConfig.
func (e *Engine) extractProtocol(mc *config.ModelConfig) (protocol, modelID string) {
	if mc == nil {
		return "", ""
	}
	// Use the same logic as providers.ExtractProtocol
	model := strings.TrimSpace(mc.Model)
	if model == "" {
		return "", ""
	}
	parts := strings.SplitN(model, "/", 2)
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	provider := strings.TrimSpace(mc.Provider)
	if provider == "" {
		provider = "openai"
	}
	return provider, model
}

// CreateProviderFromModelConfig creates an LLM provider from a ModelConfig,
// enriching it with api_key and api_base from providers config or env vars.
func (e *Engine) CreateProviderFromModelConfig(mc *config.ModelConfig) (any, string, error) {
	if mc == nil {
		return nil, "", fmt.Errorf("nil ModelConfig")
	}
	clone := *mc

	providerName := strings.TrimSpace(clone.Provider)

	// Try to enrich from ada_config providers (case-insensitive lookup).
	if providerName != "" && e.adaCfg.Providers != nil {
		lower := strings.ToLower(providerName)
		for key, provCfg := range e.adaCfg.Providers {
			if strings.ToLower(key) == lower {
				if clone.APIBase == "" && provCfg.ApiUrl != "" {
					clone.APIBase = provCfg.ApiUrl
				}
				if len(clone.APIKeys) == 0 {
					if apiKey := e.adaCfg.GetProviderAPIKey(key); apiKey != "" {
						clone.APIKeys = config.SimpleSecureStrings(apiKey)
					}
				}
				// Use type_connection as the provider protocol when it's a known
				// factory protocol. This lets custom-named providers (e.g. "nararouter")
				// with type_connection="openai" be routed through the OpenAI-compatible
				// code path instead of being rejected as "unknown protocol".
				if tc := strings.TrimSpace(provCfg.TypeConnection); tc != "" {
					switch strings.ToLower(tc) {
					case "openai", "anthropic", "gemini":
						clone.Provider = strings.ToLower(tc)
					}
				}
				break
			}
		}
	}

	// If still no API key, check environment variables.
	hasValidKey := false
	for _, k := range clone.APIKeys {
		if k.String() != "" {
			hasValidKey = true
			break
		}
	}
	if !hasValidKey && providerName != "" {
		envKey := strings.ToUpper(strings.ReplaceAll(providerName, "-", "_")) + "_API_KEY"
		if apiKey := os.Getenv(envKey); apiKey != "" {
			clone.APIKeys = config.SimpleSecureStrings(apiKey)
		}
	}

	// If still no API base, set sensible defaults for known providers.
	if clone.APIBase == "" {
		switch strings.ToLower(providerName) {
		case "openrouter":
			clone.APIBase = "https://openrouter.ai/api/v1"
		case "openai":
			clone.APIBase = "https://api.openai.com/v1"
		case "anthropic":
			clone.APIBase = "https://api.anthropic.com/v1"
		}
	}

	return providers.CreateProviderFromConfig(&clone)
}

// Orchestrator retorna a instância do orquestrador (pode ser nil se desabilitado).
func (e *Engine) Orchestrator() *orchestrator.Orchestrator {
	return e.orchestrator
}

// resolveProviderFromModel resolve um LLMProvider a partir do nome do provider e modelo.
func (e *Engine) resolveProviderFromModel(providerName, modelName string) providers.LLMProvider {
	if providerName == "" || modelName == "" {
		return nil
	}
	// Procura no ModelList
	for _, mc := range e.adaCfg.ModelList {
		if mc == nil {
			continue
		}
		if strings.EqualFold(mc.Provider, providerName) && strings.EqualFold(mc.ModelName, modelName) {
			p, _, err := e.CreateProviderFromModelConfig(mc)
			if err == nil {
				if lp, ok := p.(providers.LLMProvider); ok {
					return lp
				}
			}
		}
	}
	// Fallback: cria ModelConfig sintético
	synthetic := &config.ModelConfig{
		Provider:  providerName,
		ModelName: modelName,
		Model:     modelName,
		Enabled:   true,
	}
	if apiKey := e.adaCfg.GetProviderAPIKey(providerName); apiKey != "" {
		synthetic.APIKeys = config.SimpleSecureStrings(apiKey)
	}
	p, _, err := e.CreateProviderFromModelConfig(synthetic)
	if err == nil {
		if lp, ok := p.(providers.LLMProvider); ok {
			return lp
		}
	}
	return nil
}

// resolveWorkspaceRoutingRules retorna as regras de roteamento do workspace associado à sessão.
// Usado APENAS pelo orquestrador para decidir qual sub-agente acionar.
// A persona do worker é usada pelo agent normal (ContextBuilder), não aqui.
func (e *Engine) resolveWorkspaceRoutingRules(sessionID string) string {
	sess := e.SessionMgr.GetSession(sessionID)
	if sess == nil {
		return ""
	}

	for _, ws := range e.adaCfg.Workspaces {
		if ws.Path == sess.WorkspaceID || ws.Title == sess.WorkspaceID {
			return ws.RoutingRules
		}
	}
	return ""
}

// resolveWorkerPersona retorna a persona completa do worker associado à sessão.
// Usada pelo agent normal (ContextBuilder) como system prompt.
func (e *Engine) resolveWorkerPersona(sessionID string) string {
	sess := e.SessionMgr.GetSession(sessionID)
	if sess == nil || e.db == nil {
		return ""
	}

	if sess.WorkerName != "" {
		if worker, err := e.db.GetWorkerByName(sess.WorkerName); err == nil && worker != nil {
			return FullPersona(*worker)
		}
	}
	return ""
}

// ProcessOrchestrated processa uma requisição através do orquestrador multi-agent.
func (e *Engine) ProcessOrchestrated(ctx context.Context, text string, sessionID string, modelOverride string) (string, error) {
	fmt.Printf("[ProcessOrchestrated] step=start sessionID=%q modelOverride=%q\n", sessionID, modelOverride)

	// 0. Garante que a sessão está no SessionMgr (pode ter vindo direto do DB via frontend)
	if sess := e.SessionMgr.GetSession(sessionID); sess == nil && e.db != nil {
		if dbSess, err := e.db.GetSession(sessionID); err == nil && dbSess != nil {
			e.SessionMgr.LoadSession(dbSess)
			fmt.Printf("[ProcessOrchestrated] sessão %q carregada do DB para SessionMgr\n", sessionID)
		}
	}

	// 1. Extrai o nome do modelo do override (formato "Provider/ModelName")
	modelName := modelOverride
	if idx := strings.Index(modelOverride, "/"); idx != -1 {
		modelName = modelOverride[idx+1:]
	}
	fmt.Printf("[ProcessOrchestrated] modelName=%q\n", modelName)

	// 1. Resolve provider do chat ativo (modelOverride do frontend)
	provider := e.resolveProviderForSession(modelOverride)
	if provider != nil {
		fmt.Printf("[ProcessOrchestrated] provider resolved OK\n")
		e.orchestrator.SetProvider(provider)
		e.orchestrator.SetModel(modelName)
		e.orchestrator.SetSubAgentProviders(provider)
	} else {
		fmt.Printf("[ProcessOrchestrated] provider is NIL - using heuristic routing\n")
	}

	// 2. Resolve routing rules do workspace
	routingRules := e.resolveWorkspaceRoutingRules(sessionID)
	fmt.Printf("[ProcessOrchestrated] routingRules=%q\n", routingRules)

	// 3. Aplica timeout do config
	timeout := e.orchestrator.Config().Timeout
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// 4. Monta estado do histórico da sessão
	state := e.buildOrchestratorState(sessionID)

	// 5. Roteamento via LLM (routingRules como prompt de roteamento)
	decision, err := e.orchestrator.LLMRoute(ctx, text, routingRules)
	if err != nil {
		fmt.Printf("[ProcessOrchestrated] LLM routing falhou: %v\n", err)
	}
	fmt.Printf("[ProcessOrchestrated] decision: nextAgent=%q task=%q subTasks=%d reasoning=%q\n",
		decision.NextAgent, decision.Task, len(decision.SubTasks), decision.Reasoning)

	// 6. Emite decisão de roteamento
	e.eventBus.Emit(Event{
		Kind:      EventKindOrchestratorDecision,
		SessionID: sessionID,
		Payload: OrchestratorDecisionPayload{
			Reasoning:    decision.Reasoning,
			NextAgent:    string(decision.NextAgent),
			Task:         decision.Task,
			SubTasks:     len(decision.SubTasks),
			AgentCount:   len(decision.SubTasks),
			RelatedFiles: decision.RelatedFiles,
		},
	})

	// 7. Callback para emitir eventos de progresso das sub-tasks
	onEvent := func(kind string, st orchestrator.SubTask, err error) {
		switch kind {
		case "start":
			e.eventBus.Emit(Event{
				Kind:      EventKindSubTaskStart,
				SessionID: sessionID,
				Payload: SubTaskPayload{
					ID:     st.ID,
					Agent:  string(st.Agent),
					Task:   st.Task,
					Status: "started",
				},
			})
		case "complete":
			e.eventBus.Emit(Event{
				Kind:      EventKindSubTaskComplete,
				SessionID: sessionID,
				Payload: SubTaskPayload{
					ID:     st.ID,
					Agent:  string(st.Agent),
					Task:   st.Task,
					Status: "completed",
				},
			})
		case "error":
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}
			e.eventBus.Emit(Event{
				Kind:      EventKindSubTaskError,
				SessionID: sessionID,
				Payload: SubTaskPayload{
					ID:     st.ID,
					Agent:  string(st.Agent),
					Task:   st.Task,
					Status: "error",
					Error:  errMsg,
				},
			})
		}
	}

	// 8. Execução (concorrente para sub-tasks)
	fmt.Printf("[ProcessOrchestrated] calling ExecuteRouting: decision={NextAgent=%s Task=%q SubTasks=%d}\n",
		decision.NextAgent, decision.Task, len(decision.SubTasks))
	output, err := e.orchestrator.ExecuteRouting(ctx, decision, state, onEvent)
	fmt.Printf("[ProcessOrchestrated] ExecuteRouting returned: outputLen=%d err=%v\n", len(output), err)
	if err != nil {
		// Se o erro for de "agent not found" (NextAgent vazio), trata como resposta educada
		if decision.NextAgent == "" {
			output = "Olá! Como posso ajudar você hoje? Posso desenvolver backend em Go, criar interfaces em React ou escrever testes automatizados."
			err = nil
		} else {
			// Salva a mensagem do usuário e o erro no histórico antes de retornar
			e.SessionMgr.AddMessage(sessionID, "user", text)
			e.SessionMgr.AddMessage(sessionID, "assistant", fmt.Sprintf("Erro: %v", err))
			if sess := e.SessionMgr.GetSession(sessionID); sess != nil && e.db != nil {
				e.db.SaveSession(*sess)
			}
			return "", fmt.Errorf("execução do orquestrador falhou: %w", err)
		}
	}

	// Emite aviso se o agente genérico foi usado como fallback
	if decision.NextAgent != "" && decision.NextAgent != orchestrator.AgentTypeGeneric {
		// Verifica se o agent realmente existe no registry (pode ter sido substituído por genérico)
		if _, ok := e.orchestrator.GetRegistry().Get(decision.NextAgent); !ok {
			warningMsg := fmt.Sprintf("⚠️ Agente %q não configurado. Usando agente genérico como fallback.", decision.NextAgent)
			e.eventBus.Emit(Event{
				Kind:      EventKindStatus,
				SessionID: sessionID,
				Payload:   StatusPayload{Message: warningMsg},
				Time:      time.Now(),
			})
			output = warningMsg + "\n\n" + output
		}
	}

	// 9. Salva na sessão
	e.SessionMgr.AddMessage(sessionID, "user", text)
	e.SessionMgr.AddMessage(sessionID, "assistant", output)
	if sess := e.SessionMgr.GetSession(sessionID); sess != nil && e.db != nil {
		e.db.SaveSession(*sess)
	}

	return output, nil
}

// resolveProviderForSession resolve o LLMProvider a partir do modelOverride do frontend.
func (e *Engine) resolveProviderForSession(modelOverride string) providers.LLMProvider {
	if modelOverride == "" {
		return nil
	}
	e.providerMu.RLock()
	cached, ok := e.providerCache[modelOverride]
	e.providerMu.RUnlock()
	if ok {
		if lp, ok := cached.(providers.LLMProvider); ok {
			fmt.Printf("[resolveProviderForSession] cache HIT for %q\n", modelOverride)
			return lp
		}
	}

	fmt.Printf("[resolveProviderForSession] cache MISS for %q — creating provider\n", modelOverride)

	// Not in cache — try to create provider from model_list or providers config
	adaCfg := e.GetAdaConfig()
	var cachedProvider any
	var resolvedModelID string

	// Step 1: search model_list
	for _, mc := range adaCfg.ModelList {
		if mc == nil {
			continue
		}
		provider := strings.TrimSpace(mc.Provider)
		modelName := strings.TrimSpace(mc.ModelName)
		modelField := strings.TrimSpace(mc.Model)
		fullKey := provider + "/" + modelName
		fmt.Printf("[resolveProviderForSession] checking: provider=%q modelName=%q fullKey=%q\n", provider, modelName, fullKey)
		if modelName == modelOverride || modelField == modelOverride || fullKey == modelOverride {
			fmt.Printf("[resolveProviderForSession] MATCH found! Provider=%q Model=%q Keys=%d\n", mc.Provider, mc.Model, len(mc.APIKeys))
			p, _, err := e.CreateProviderFromModelConfig(mc)
			if err == nil && p != nil {
				cachedProvider = p
				resolvedModelID = modelField
				break
			}
			fmt.Printf("[resolveProviderForSession] Provider creation FAILED for %q: %v\n", modelOverride, err)
		}
	}

	// Step 2: if not found in model_list, search providers config
	if cachedProvider == nil {
		fmt.Printf("[resolveProviderForSession] not found in model_list, trying providers config\n")
		parts := strings.SplitN(modelOverride, "/", 2)
		if len(parts) == 2 {
			providerName := parts[0]
			modelName := parts[1]
			fmt.Printf("[resolveProviderForSession] looking for providerName=%q modelName=%q\n", providerName, modelName)
			if provCfg, ok := adaCfg.Providers[providerName]; ok {
				fmt.Printf("[resolveProviderForSession] found provider %q with %d models and %d keys\n", providerName, len(provCfg.Models), len(provCfg.ApiKeys))
				if _, exists := provCfg.Models[modelName]; exists {
					apiBase := provCfg.ApiUrl
					if apiBase == "" {
						apiBase = defaultAPIBaseFor(providerName, provCfg.TypeConnection)
					}
					allKeys := provCfg.GetAllAPIKeys()
					fmt.Printf("[resolveProviderForSession] building synthetic ModelConfig: apiBase=%q keys=%v\n", apiBase, allKeys)
					synthetic := &config.ModelConfig{
						Provider:  providerName,
						ModelName: modelName,
						Model:     modelName,
						APIBase:   apiBase,
						APIKeys:   config.SimpleSecureStrings(allKeys...),
						Enabled:   true,
					}
					p, _, err := e.CreateProviderFromModelConfig(synthetic)
					if err == nil && p != nil {
						cachedProvider = p
						resolvedModelID = modelName
					} else {
						fmt.Printf("[resolveProviderForSession] FAILED to create from providers config: %v\n", err)
					}
				} else {
					fmt.Printf("[resolveProviderForSession] model %q not found in provider %q models (available: %v)\n", modelName, providerName, getMapKeys(provCfg.Models))
				}
			} else {
				fmt.Printf("[resolveProviderForSession] provider %q NOT found in adaCfg.Providers (keys: %v)\n", providerName, getMapKeys(adaCfg.Providers))
			}
		}
	}

	// Cache the provider if found
	if cachedProvider != nil {
		fmt.Printf("[resolveProviderForSession] caching provider for %q\n", modelOverride)
		e.providerMu.Lock()
		e.providerCache[modelOverride] = cachedProvider
		e.providerMu.Unlock()
		e.overrideModelMu.Lock()
		e.overrideModelIDs[modelOverride] = resolvedModelID
		e.overrideModelMu.Unlock()

		if lp, ok := cachedProvider.(providers.LLMProvider); ok {
			fmt.Printf("[resolveProviderForSession] returning LLMProvider OK\n")
			return lp
		}
	}

	fmt.Printf("[resolveProviderForSession] FAILED to resolve any provider for %q\n", modelOverride)
	return nil
}

// getMapKeys returns the keys of a map as a slice for debug logging.
func getMapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// buildOrchestratorState monta o estado do histórico para o orquestrador.
func (e *Engine) buildOrchestratorState(sessionID string) string {
	sess := e.SessionMgr.GetSession(sessionID)
	if sess == nil || len(sess.Messages) == 0 {
		return ""
	}

	// Pega as últimas 6 mensagens como contexto
	var parts []string
	start := 0
	if len(sess.Messages) > 6 {
		start = len(sess.Messages) - 6
	}
	for _, msg := range sess.Messages[start:] {
		role := msg.Role
		if role == "assistant" {
			role = "Agente"
		} else if role == "user" {
			role = "Usuário"
		}
		content := msg.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		parts = append(parts, fmt.Sprintf("[%s]: %s", role, content))
	}
	return strings.Join(parts, "\n")
}

// isOrchestratorRequest detecta se uma mensagem deve ser roteada pelo orquestrador.
var (
	orchestratorMultiStepRe = regexp.MustCompile(`(?i)(crie\s+.*\s+e\s+(depois|faça|adicione|implemente)|implemente\s+.*\s+e\s+(depois|faça|adicione|test)|faça\s+.*\s+e\s+(depois|adicione|crie)|backend\s+.*\s+e\s+.*frontend|frontend\s+.*\s+e\s+.*backend|go\s+.*\s+e\s+.*react|react\s+.*\s+e\s+.*go|api\s+.*\s+e\s+.*tela|tela\s+.*\s+e\s+.*api)`)
	orchestratorPrefixRe    = regexp.MustCompile(`(?i)^/(orchestrate|multi)\b`)
	orchestratorExplicitRe  = regexp.MustCompile(`(?i)(golang\s+agent|react\s+agent|tester\s+agent|orquestrador|orquest)`)
)

func isOrchestratorRequest(text string) bool {
	return orchestratorMultiStepRe.MatchString(text) ||
		orchestratorPrefixRe.MatchString(text) ||
		orchestratorExplicitRe.MatchString(text)
}

// getOSConfigDir returns the OS-specific config directory
func getOSConfigDir() string {
	var configDir string
	switch runtime.GOOS {
	case "linux":
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "ada-love-ai")
	case "darwin":
		configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "ada-love-ai")
	case "windows":
		configDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "ada-love-ai")
	default:
		configDir = "config"
	}
	os.MkdirAll(configDir, 0755)
	return configDir
}

// syncAdaAgentsToConfig sincroniza os agentes do ada_config.json com cfg.Agents.List.
// Converte agentes do formato backend.AgentConfig para config.AgentConfig.
func syncAdaAgentsToConfig(adaAgents []AgentConfig, cfgAgents *[]config.AgentConfig) {
	if len(adaAgents) == 0 {
		return
	}

	// Cria um mapa dos agentes existentes para evitar duplicatas
	existingIDs := make(map[string]bool)
	for _, a := range *cfgAgents {
		if a.ID != "" {
			existingIDs[a.ID] = true
		}
	}

	// Adiciona agentes do ada_config.json que ainda não existem
	for _, adaAgent := range adaAgents {
		agentID := adaAgent.Name
		if agentID == "" {
			continue
		}

		// Usa ID numérico do agente se fornecido, senão gera a partir do nome
		normalizedID := strings.ToLower(strings.ReplaceAll(agentID, " ", "-"))
		if adaAgent.ID > 0 {
			normalizedID = fmt.Sprintf("agent-%d", adaAgent.ID)
		}

		if existingIDs[normalizedID] {
			continue // Já existe
		}

		// Converte para config.AgentConfig
		cfgAgent := config.AgentConfig{
			ID:   normalizedID,
			Name: adaAgent.Name,
			Model: &config.AgentModelConfig{
				Primary: adaAgent.Model,
			},
			Provider: adaAgent.Provider,
			Type:     adaAgent.Type,
			Icon:     adaAgent.Icon,
			Color:    adaAgent.Color,
		}

		// Converte delegates para subagents.allow_agents
		if len(adaAgent.Delegates) > 0 {
			cfgAgent.Subagents = &config.SubagentsConfig{
				AllowAgents: adaAgent.Delegates,
			}
		}

		// Usa SystemPrompt como personality se não houver personality definida
		if adaAgent.SystemPrompt != "" {
			cfgAgent.Personality = adaAgent.SystemPrompt
		}

		*cfgAgents = append(*cfgAgents, cfgAgent)
		fmt.Printf("[Engine] Sincronizado agente: %s (type=%s, id=%s)\n", adaAgent.Name, adaAgent.Type, normalizedID)
	}
}

// filterAgentsByWorkspace filtra a lista de agentes para manter apenas os
// selecionados no workspace. Um agente é selecionado se seu ID ou Name estiver
// na lista ws.Agents.
func filterAgentsByWorkspace(cfgAgents *[]config.AgentConfig, selectedAgentNames []string) {
	if len(selectedAgentNames) == 0 {
		return
	}

	// Cria um set de nomes/ID selecionados (case-insensitive, normalizado)
	selectedSet := make(map[string]bool)
	for _, name := range selectedAgentNames {
		normalized := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		selectedSet[normalized] = true
	}

	// Filtra a lista mantendo apenas os agentes selecionados
	var filtered []config.AgentConfig
	for _, agent := range *cfgAgents {
		// Verifica pelo ID normalizado
		idNormalized := strings.ToLower(strings.ReplaceAll(agent.ID, " ", "-"))
		nameNormalized := strings.ToLower(strings.ReplaceAll(agent.Name, " ", "-"))

		if selectedSet[idNormalized] || selectedSet[nameNormalized] {
			filtered = append(filtered, agent)
		}
	}

	*cfgAgents = filtered
	fmt.Printf("[Engine] Filtrados agentes para workspace: %d selecionados de %d\n", len(filtered), len(selectedAgentNames))
}

// GetSummarizedContext retorna o contexto sumarizado da sessão do backend
func (e *Engine) GetSummarizedContext(sessionID string) string {
	if e.db == nil || sessionID == "" {
		return ""
	}
	sess, err := e.db.GetSession(sessionID)
	if err != nil || sess == nil {
		return ""
	}
	return sess.SummarizedContext
}

// workspaceHealth checks the active workspace configuration and returns a health report.
func (e *Engine) workspaceHealth() (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	activePath := e.adaCfg.ActiveWorkspacePath
	if activePath == "" {
		return "", fmt.Errorf("no active workspace")
	}

	var ws *WorkspaceConfig
	for i := range e.adaCfg.Workspaces {
		if e.adaCfg.Workspaces[i].Path == activePath || e.adaCfg.Workspaces[i].Title == activePath {
			ws = &e.adaCfg.Workspaces[i]
			break
		}
	}
	if ws == nil {
		return "", fmt.Errorf("workspace %q not found", activePath)
	}

	var report strings.Builder
	report.WriteString(fmt.Sprintf("🔍 Workspace Health: %s\n", ws.Title))
	report.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	critical := 0
	warnings := 0
	ok := 0

	// 1. Workspace path
	if ws.Path != "" {
		report.WriteString(fmt.Sprintf("✅ Path: %s\n", ws.Path))
		ok++
	} else {
		report.WriteString("❌ Path: NENHUM — workspace sem path\n")
		critical++
	}

	// 2. Folders
	if len(ws.Folders) > 0 {
		report.WriteString(fmt.Sprintf("✅ Folders: %s\n", strings.Join(ws.Folders, ", ")))
		ok++
		// Verifica se as pastas existem no disco
		for _, f := range ws.Folders {
			if _, err := os.Stat(f); os.IsNotExist(err) {
				report.WriteString(fmt.Sprintf("  ❌ Folder não encontrada: %s\n", f))
				critical++
			}
		}
	} else {
		report.WriteString("❌ Folders: NENHUM — chat sessions, skills e file tools não funcionarão corretamente\n")
		critical++
	}

	// 3. Agents, Workers & Orchestrator
	if e.orchestrator != nil {
		report.WriteString("✅ Orchestrator: Ativo\n")
		ok++
	} else {
		report.WriteString("⚠️  Orchestrator: Inativo (multi-agent desabilitado)\n")
		warnings++
	}

	agentCount := len(ws.Agents)
	if agentCount == 0 && len(e.adaCfg.Agents) > 0 {
		agentCount = len(e.adaCfg.Agents) // fallback para agentes globais
	}
	if agentCount > 0 {
		report.WriteString(fmt.Sprintf("✅ Agents: %d configurados\n", agentCount))
		ok++
	} else {
		report.WriteString("⚠️  Agents: Nenhum agente específico no workspace\n")
		warnings++
	}

	if len(ws.WorkerNames) > 0 {
		report.WriteString(fmt.Sprintf("✅ Workers: %s (%d configurados)\n", strings.Join(ws.WorkerNames, ", "), len(ws.WorkerNames)))
		ok++
	} else {
		report.WriteString("⚠️  Workers: NENHUM — orquestrador desativado, apenas chat direto\n")
		warnings++
	}

	// 4. Personality
	if ws.Personality != "" {
		report.WriteString("✅ Personality: Configurada\n")
		ok++
	} else {
		report.WriteString("⚠️  Personality: Não configurada — agente sem personalidade\n")
		warnings++
	}

	// 5. Skills & Tools
	skillCount := len(ws.Skills)
	if skillCount > 0 {
		report.WriteString(fmt.Sprintf("✅ Skills: %d habilitadas\n", skillCount))
		ok++
	} else {
		report.WriteString("⚠️  Skills: Nenhuma skill habilitada no workspace\n")
		warnings++
	}

	toolCount := len(ws.Tools)
	if toolCount > 0 {
		report.WriteString(fmt.Sprintf("✅ Tools: %d habilitadas\n", toolCount))
		ok++
	} else {
		report.WriteString("⚠️  Tools: Nenhuma ferramenta habilitada (read/write file desativados)\n")
		warnings++
	}

	mcpCount := 0
	for _, m := range e.adaCfg.MCPServers {
		if m.Enabled {
			mcpCount++
		}
	}
	if mcpCount > 0 {
		report.WriteString(fmt.Sprintf("✅ MCP: %d servidores ativos\n", mcpCount))
		ok++
	} else {
		report.WriteString("ℹ️  MCP: Nenhum servidor externo configurado\n")
	}

	// 6. Providers & Models
	if len(e.adaCfg.Providers) > 0 {
		provNames := getMapKeys(e.adaCfg.Providers)
		totalKeys := 0
		for _, name := range provNames {
			if p, ok := e.adaCfg.Providers[name]; ok {
				totalKeys += len(p.ApiKeys)
			}
		}
		report.WriteString(fmt.Sprintf("✅ Providers: %d configurados, %d API keys no total\n", len(provNames), totalKeys))
		ok++
	} else {
		report.WriteString("❌ Providers: NENHUM — nenhum modelo disponível para chat\n")
		critical++
	}

	if len(e.adaCfg.ModelList) > 0 {
		report.WriteString(fmt.Sprintf("✅ Modelos: %d disponíveis no model_list\n", len(e.adaCfg.ModelList)))
		ok++
	} else {
		report.WriteString("⚠️  Modelos: NENHUM no model_list — use override de modelo no chat\n")
		warnings++
	}

	// 7. Special Providers
	if e.adaCfg.SpecModel != "" && e.adaCfg.SpecProvider != "" {
		report.WriteString(fmt.Sprintf("✅ Spec Wizard: %s (%s)\n", e.adaCfg.SpecModel, e.adaCfg.SpecProvider))
		ok++
	} else {
		report.WriteString("⚠️  Spec Wizard: Não configurado (geração de specs desativada)\n")
		warnings++
	}

	if e.adaCfg.ImageModel != "" {
		report.WriteString(fmt.Sprintf("✅ Image Gen: %s\n", e.adaCfg.ImageModel))
		ok++
	}

	if e.adaCfg.EmbeddingModel != "" {
		report.WriteString(fmt.Sprintf("✅ Embeddings: %s\n", e.adaCfg.EmbeddingModel))
		ok++
	}

	// 8. TinyBrain
	if e.adaCfg.TinyBrain.ModelName != "" {
		report.WriteString(fmt.Sprintf("✅ TinyBrain: %s (%s)\n", e.adaCfg.TinyBrain.ModelName, e.adaCfg.TinyBrain.Provider))
		ok++
	} else {
		report.WriteString("ℹ️  TinyBrain: Não configurado (usando modelo principal para tarefas leves)\n")
	}

	// 9. Knowledge Base
	if len(ws.Knowledge) > 0 {
		report.WriteString(fmt.Sprintf("✅ Knowledge: %d arquivos\n", len(ws.Knowledge)))
		ok++
		for _, k := range ws.Knowledge {
			if _, err := os.Stat(k); os.IsNotExist(err) {
				report.WriteString(fmt.Sprintf("  ❌ Arquivo não encontrado: %s\n", k))
				critical++
			}
		}
	}

	// 10. Git Integration
	if ws.CommitChanges {
		report.WriteString("✅ Git Auto-commit: Habilitado\n")
		ok++
		for _, f := range ws.Folders {
			gitPath := filepath.Join(f, ".git")
			if _, err := os.Stat(gitPath); os.IsNotExist(err) {
				report.WriteString(fmt.Sprintf("  ⚠️  Folder não é repo Git: %s\n", f))
				warnings++
			}
		}
	}

	// 11. Constraints
	report.WriteString(fmt.Sprintf("⚙️  Constraints: context=%d, prompt=%d\n", ws.MaxContextLength, ws.MaxPrompt))

	report.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	report.WriteString(fmt.Sprintf("%d ❌ critical | %d ⚠️  warning | %d ✅ ok\n", critical, warnings, ok))

	return report.String(), nil
}

// testConnections validates real provider API connectivity for all configured providers.
func (e *Engine) testConnections() (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var report strings.Builder
	report.WriteString("🌐 Provider Connection Test\n")
	report.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	ok := 0
	failed := 0

	for name, p := range e.adaCfg.Providers {
		report.WriteString(fmt.Sprintf("Testing %s... ", name))

		// Use existing TestProviderConnection logic
		res, err := e.TestProviderConnection(name, p.GetAPIKey(), p.ApiUrl, p.TypeConnection)

		if err == nil && res.Ok {
			report.WriteString("✅ OK\n")
			ok++
		} else {
			msg := "FAIL"
			if res.Message != "" {
				msg = res.Message
			} else if err != nil {
				msg = err.Error()
			}
			report.WriteString(fmt.Sprintf("❌ %s\n", msg))
			failed++
		}
	}

	report.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	report.WriteString(fmt.Sprintf("%d ✅ success | %d ❌ failed\n", ok, failed))

	return report.String(), nil
}
