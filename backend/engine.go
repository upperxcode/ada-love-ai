package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ada-love-ai/pkg/agent"
	"ada-love-ai/pkg/agent/interfaces"
	"ada-love-ai/pkg/bus"
	"ada-love-ai/pkg/config"
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
	eventBus   *EventBus
	db         *Store
	toolReg    *adatools.ToolRegistry
	questionReg *integrationtools.QuestionRegistry
	approvalReg *integrationtools.ApprovalRegistry
	providerCache map[string]any
	providerMu    sync.RWMutex
	// overrideModelIDs maps frontend model key (e.g. "OpenRouter/nvidia/...") to the
	// actual model field expected by the provider API (e.g. "nvidia/...").
	overrideModelIDs map[string]string
	overrideModelMu sync.RWMutex
	// Summarization
	summarizer *SummarizerWorker
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

	// --- Migração one-shot: se DB vazio, lê JSON antigo e semeia ---
	if db != nil {
		providers, _ := db.GetProviders()
		if len(providers) == 0 {
			// DB vazio → tenta migrar de ada_config.json
			adaConfigPath := filepath.Join(configDir, "ada_config.json")
			if data, err := os.ReadFile(adaConfigPath); err == nil {
				var legacy AdaConfig
				json.Unmarshal(data, &legacy)
				if len(legacy.Providers) > 0 {
					db.SaveProviders(legacy.Providers)
					adaCfg.Providers = legacy.Providers
					fmt.Printf("[Engine] Migrados %d providers do JSON legado para DB\n", len(legacy.Providers))
				}
				if len(legacy.Workspaces) > 0 {
					for _, ws := range legacy.Workspaces {
						db.SaveWorkspace(ws)
					}
					fmt.Printf("[Engine] Migrados %d workspaces do JSON legado para DB\n", len(legacy.Workspaces))
				}
				if len(legacy.Workers) > 0 {
					db.SetGlobalConfig("workers", legacy.Workers)
					fmt.Printf("[Engine] Migrados %d workers do JSON legado para DB\n", len(legacy.Workers))
				}
				if len(legacy.Agents) > 0 {
					db.SetGlobalConfig("agents", legacy.Agents)
					fmt.Printf("[Engine] Migrados %d agents do JSON legado para DB\n", len(legacy.Agents))
				}
				if legacy.TinyBrain.ModelName != "" || legacy.TinyBrain.Provider != "" {
					db.SetGlobalConfig("tiny_brain", legacy.TinyBrain)
				}
				if legacy.EmbeddingModel != "" || legacy.EmbeddingProvider != "" {
					db.SetGlobalConfig("embedding_model", legacy.EmbeddingModel)
					db.SetGlobalConfig("embedding_provider", legacy.EmbeddingProvider)
				}
				if legacy.ImageModel != "" || legacy.ImageProvider != "" {
					db.SetGlobalConfig("image_model", legacy.ImageModel)
					db.SetGlobalConfig("image_provider", legacy.ImageProvider)
				}
				if legacy.SpecModel != "" || legacy.SpecProvider != "" {
					db.SetGlobalConfig("spec_model", legacy.SpecModel)
					db.SetGlobalConfig("spec_provider", legacy.SpecProvider)
				}
				if legacy.ToolProfiles != nil {
					db.SetGlobalConfig("tool_profiles", legacy.ToolProfiles)
				}
				if len(legacy.MCPServers) > 0 {
					db.SetGlobalConfig("mcp_servers", legacy.MCPServers)
				}
				if legacy.ActiveWorkspacePath != "" {
					db.SetGlobalConfig("active_workspace_path", legacy.ActiveWorkspacePath)
				}
				if legacy.ActiveWorkspaceIndex > 0 {
					db.SetGlobalConfig("active_workspace_index", legacy.ActiveWorkspaceIndex)
				}
			}
		}
	}

	// --- Carrega do DB ---
	if db != nil {
		// Providers
		providers, err := db.GetProviders()
		if err != nil {
			fmt.Printf("[Engine] Erro ao carregar providers: %v\n", err)
		}
		adaCfg.Providers = providers
		fmt.Printf("[Engine] Carregados %d providers do DB\n", len(providers))

		// Workspaces
		dbWorkspaces, _ := db.GetWorkspaces()
		if len(dbWorkspaces) > 0 {
			adaCfg.Workspaces = dbWorkspaces
		}
		fmt.Printf("[Engine] Carregados %d workspaces do DB\n", adaCfg.WorkspaceCount())

		// Workers
		db.GetGlobalConfig("workers", &adaCfg.Workers)
		// Agents
		db.GetGlobalConfig("agents", &adaCfg.Agents)
		// TinyBrain
		db.GetGlobalConfig("tiny_brain", &adaCfg.TinyBrain)
		// Embedding
		db.GetGlobalConfig("embedding_model", &adaCfg.EmbeddingModel)
		db.GetGlobalConfig("embedding_provider", &adaCfg.EmbeddingProvider)
		// Image
		db.GetGlobalConfig("image_model", &adaCfg.ImageModel)
		db.GetGlobalConfig("image_provider", &adaCfg.ImageProvider)
		// Spec
		db.GetGlobalConfig("spec_model", &adaCfg.SpecModel)
		db.GetGlobalConfig("spec_provider", &adaCfg.SpecProvider)
		// ToolProfiles
		db.GetGlobalConfig("tool_profiles", &adaCfg.ToolProfiles)
		// MCPServers
		db.GetGlobalConfig("mcp_servers", &adaCfg.MCPServers)
		// Active workspace
		db.GetGlobalConfig("active_workspace_path", &adaCfg.ActiveWorkspacePath)
		db.GetGlobalConfig("active_workspace_index", &adaCfg.ActiveWorkspaceIndex)
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

	// Popula cfg.ModelList a partir dos providers do DB.
	// Isso garante que CreateProvider (que busca em cfg.ModelList) encontre os modelos.
	for provName, provCfg := range adaCfg.Providers {
		for modelName := range provCfg.Models {
			found := false
			for _, existing := range cfg.ModelList {
				if existing.ModelName == modelName && existing.Provider == provName {
					found = true
					break
				}
			}
			if found {
				continue
			}
			cfg.ModelList = append(cfg.ModelList, &config.ModelConfig{
				ModelName: modelName,
				Provider:  provName,
				Model:     modelName,
				APIBase:   provCfg.ApiUrl,
				Enabled:   true,
			})
		}
	}
	fmt.Printf("[Engine] ModelList populado: %d modelos de %d providers\n", len(cfg.ModelList), len(adaCfg.Providers))

	e := &Engine{
		cfg:           cfg,
		msgBus:        msgBus,
		eventBus:      eventBus,
		adaCfg:        adaCfg,
		SessionMgr:    NewSessionManager(),
		skillReg:   skills.NewRegistryManagerFromToolsConfig(cfg.Tools.Skills),
		db:         db,
		providerCache: make(map[string]any),
		overrideModelIDs: make(map[string]string),
		sessionKeyMap:   make(map[string]string),
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
			fmt.Printf("[Engine]   session=%q title=%q worker=%q messages=%d pinned=%v\n",
				s.ID, s.Title, s.WorkerName, len(s.Messages), s.Pinned)
		}
		e.SessionMgr.LoadSessions(sessions)
	}

	// Migração: sessões sem workspace_path → move para o primeiro workspace
	if e.db != nil {
		// Corrigir workspaces com path vazio e workers nulos
		for i := range e.adaCfg.Workspaces {
			// Fix empty path
			if e.adaCfg.Workspaces[i].Path == "" {
				newPath := strings.ToLower(strings.ReplaceAll(e.adaCfg.Workspaces[i].Title, " ", "_"))
				fmt.Printf("[Engine] Init: fixing workspace %q: path '' → %q\n", e.adaCfg.Workspaces[i].Title, newPath)
				e.adaCfg.Workspaces[i].Path = newPath
				e.db.db.Exec(`UPDATE workspaces SET path = ? WHERE title = ? AND (path = '' OR path IS NULL)`, newPath, e.adaCfg.Workspaces[i].Title)
			}
			// Fix nil workers
			if e.adaCfg.Workspaces[i].Workers == nil {
				e.adaCfg.Workspaces[i].Workers = []WorkerConfig{}
			}
			// Fix nil tools
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
	e.agentLoop = agent.NewAgentLoop(cfg, msgBus, provider, e)
	e.agentLoop.SetSummarizer(e.summarizer)

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
// ContextBuilder so the system prompt reflects the correct folders.
func (e *Engine) syncWorkspaceForTurn(workspacePath string) {
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

	fmt.Printf("[Engine] syncWorkspaceForTurn: title=%q folders=%v → fsPath=%q\n",
		ws.Title, ws.Folders, fsPath)

	// Update cfg defaults so any future agent creation uses the right workspace.
	e.cfg.Agents.Defaults.Workspace = fsPath
	e.cfg.Agents.Defaults.Folders = ws.Folders
	e.cfg.Agents.Defaults.Personality = ws.Personality
	e.cfg.Agents.Defaults.Knowledge = ws.Knowledge

	// Patch the LIVE agent loop's ContextBuilder (no reload) so the system
	// prompt is rebuilt with the new workspace on the next turn.
	if e.agentLoop != nil {
		registry := e.agentLoop.GetRegistry()
		if registry != nil {
			if agent := registry.GetDefaultAgent(); agent != nil && agent.ContextBuilder != nil {
				agent.ContextBuilder.UpdateWorkspace(fsPath, ws.Folders, ws.Personality, ws.Knowledge)
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

	fmt.Printf("[Engine] applyWorkspaceToAgent: title=%q slug=%q folders=%v → fsPath=%q\n",
		ws.Title, ws.Path, ws.Folders, fsPath)

	e.cfg.Agents.Defaults.Workspace = fsPath
	e.cfg.Agents.Defaults.Folders = ws.Folders
	e.cfg.Agents.Defaults.Personality = ws.Personality
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
	// Write-through: persiste providers no DB
	if e.db != nil && len(cfg.Providers) > 0 {
		if err := e.db.SaveProviders(cfg.Providers); err != nil {
			fmt.Printf("[Engine] Erro ao sincronizar providers no DB: %v\n", err)
		}
	}
}

// GetWorkers retorna os workers configurados.
func (e *Engine) GetWorkers() []WorkerConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.adaCfg.Workers
}

// SetWorkers substitui a lista de workers e persiste.
func (e *Engine) SetWorkers(workers []WorkerConfig) {
	e.mu.Lock()
	e.adaCfg.Workers = workers
	e.mu.Unlock()
	e.SaveAdaConfig()
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

// SetAgents substitui a lista de agentes e persiste.
func (e *Engine) SetAgents(agents []AgentConfig) {
	e.mu.Lock()
	e.adaCfg.Agents = agents
	e.mu.Unlock()
	e.SaveAdaConfig()
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

	// Salva providers no DB
	if err := e.db.SaveProviders(e.adaCfg.Providers); err != nil {
		fmt.Printf("[Engine] Erro ao salvar providers no DB: %v\n", err)
	}

	// Salva workspaces no DB
	for _, ws := range e.adaCfg.Workspaces {
		e.db.SaveWorkspace(ws)
	}

	// Salva seções restantes no DB (key-value)
	e.db.SetGlobalConfig("workers", e.adaCfg.Workers)
	e.db.SetGlobalConfig("agents", e.adaCfg.Agents)
	e.db.SetGlobalConfig("tiny_brain", e.adaCfg.TinyBrain)
	e.db.SetGlobalConfig("embedding_model", e.adaCfg.EmbeddingModel)
	e.db.SetGlobalConfig("embedding_provider", e.adaCfg.EmbeddingProvider)
	e.db.SetGlobalConfig("image_model", e.adaCfg.ImageModel)
	e.db.SetGlobalConfig("image_provider", e.adaCfg.ImageProvider)
	e.db.SetGlobalConfig("spec_model", e.adaCfg.SpecModel)
	e.db.SetGlobalConfig("spec_provider", e.adaCfg.SpecProvider)
	e.db.SetGlobalConfig("tool_profiles", e.adaCfg.ToolProfiles)
	e.db.SetGlobalConfig("mcp_servers", e.adaCfg.MCPServers)
	e.db.SetGlobalConfig("active_workspace_path", e.adaCfg.ActiveWorkspacePath)
	e.db.SetGlobalConfig("active_workspace_index", e.adaCfg.ActiveWorkspaceIndex)

	return nil
}

// SaveProvidersConfig saves providers to SQLite (providers table in config.db)
func (e *Engine) SaveProvidersConfig() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.db == nil {
		return fmt.Errorf("banco de dados não disponível")
	}
	return e.db.SaveProviders(e.adaCfg.Providers)
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
		}
	}
}

// SaveSessionDB persiste a sessão atual no SQLite.
func (e *Engine) SaveSessionDB(sessionID string) {
	if sess, ok := e.SessionMgr.sessions[sessionID]; ok && e.db != nil {
		e.db.SaveSession(*sess)
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
	return e.db.SaveMemory(Memory{
		WorkspacePath: workspacePath,
		Content:       content,
		Importance:    importance,
	})
}

func (e *Engine) GetMemories(workspacePath string) ([]interfaces.MemoryEntry, error) {
	memories, err := e.db.GetMemories(workspacePath)
	if err != nil {
		return nil, err
	}

	var entries []interfaces.MemoryEntry
	for _, m := range memories {
		entries = append(entries, interfaces.MemoryEntry{
			Content:    m.Content,
			Importance: m.Importance,
			CreatedAt:  m.CreatedAt,
		})
	}
	return entries, nil
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

			// Usa ID do ada_config.json se fornecido, senão gera a partir do nome
			normalizedID := strings.ToLower(strings.ReplaceAll(agentID, " ", "-"))
			if adaAgent.ID != "" {
				normalizedID = strings.ToLower(strings.ReplaceAll(adaAgent.ID, " ", "-"))
			}

			if existingIDs[normalizedID] {
				continue // Já existe
			}

			// Converte para config.AgentConfig
			cfgAgent := config.AgentConfig{
				ID:         normalizedID,
				Name:       adaAgent.Name,
				Model: &config.AgentModelConfig{
					Primary: adaAgent.Model,
				},
				Provider:  adaAgent.Provider,
				Type:      adaAgent.Type,
				Icon:      adaAgent.Icon,
				Color:     adaAgent.Color,
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
