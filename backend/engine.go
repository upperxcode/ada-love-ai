package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"ada-love-ai/pkg/agent"
	"ada-love-ai/pkg/agent/interfaces"
	"ada-love-ai/pkg/bus"
	"ada-love-ai/pkg/config"
	"ada-love-ai/pkg/providers"
	"ada-love-ai/pkg/skills"
	adatools "ada-love-ai/pkg/tools"
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
	adaConfigPath string
	SessionMgr *SessionManager
	skillReg   *skills.RegistryManager
	eventBus   *EventBus
	db         *Store
	toolReg    *adatools.ToolRegistry
}

func NewEngine() (*Engine, error) {
	configDir := getOSConfigDir()
	fmt.Printf("[Engine] Using config directory: %s\n", configDir)

	// Carrega a configuração do arquivo config.json no diretório config/
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar config.json: %w", err)
	}

	// Carrega configuração persistente do Ada-Love
	var adaCfg AdaConfig
	adaConfigPath := filepath.Join(configDir, "ada_config.json")
	if data, err := os.ReadFile(adaConfigPath); err == nil {
		json.Unmarshal(data, &adaCfg)
	} else {
		// Fallback to local config/ directory
		if data, err := os.ReadFile("config/ada_config.json"); err == nil {
			json.Unmarshal(data, &adaCfg)
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

	// Carrega providers do SQLite
	if db != nil {
		providers, err := db.GetProviders()
		if err != nil {
			fmt.Printf("[Engine] Erro ao carregar providers: %v\n", err)
		} else if len(providers) > 0 {
			adaCfg.Providers = providers
			fmt.Printf("[Engine] Carregados %d providers do SQLite\n", len(providers))
		}
	}

	e := &Engine{
		cfg:           cfg,
		msgBus:        msgBus,
		eventBus:      eventBus,
		adaCfg:        adaCfg,
		adaConfigPath: adaConfigPath,
		SessionMgr:    NewSessionManager(),
		skillReg:   skills.NewRegistryManagerFromToolsConfig(cfg.Tools.Skills),
		db:         db,
	}

	e.connectMessageBus()

	// Carrega dados do banco se disponível
	if db != nil {
		// Tenta carregar workspaces do banco
		dbWorkspaces, _ := db.GetWorkspaces()
		if len(dbWorkspaces) > 0 {
			e.adaCfg.Workspaces = dbWorkspaces
			// Inicializa ferramentas padrão se estiverem nulas
			for i := range e.adaCfg.Workspaces {
				if e.adaCfg.Workspaces[i].Tools == nil {
					e.adaCfg.Workspaces[i].Tools = []string{"read_file", "write_file", "list_dir", "edit_file"}
				}
			}
		}

		// Carrega sessões para o workspace ativo
		workspacePath := e.adaCfg.ActiveWorkspacePath
		if workspacePath == "" {
			workspacePath = "default"
			if len(e.adaCfg.Workspaces) > 0 && e.adaCfg.ActiveWorkspaceIndex < len(e.adaCfg.Workspaces) {
				workspacePath = e.adaCfg.Workspaces[e.adaCfg.ActiveWorkspaceIndex].Path
			}
		}
		sessions, _ := db.GetSessions(workspacePath)
		e.SessionMgr.LoadSessions(sessions)
	}

	// Sincroniza o workspace ativo com a configuração antes de iniciar
	if e.adaCfg.ActiveWorkspacePath != "" {
		e.cfg.Agents.Defaults.Workspace = e.adaCfg.ActiveWorkspacePath

		// Sincroniza as pastas, personalidade e conhecimento do workspace ativo
		if e.adaCfg.ActiveWorkspaceIndex >= 0 && e.adaCfg.ActiveWorkspaceIndex < len(e.adaCfg.Workspaces) {
			ws := e.adaCfg.Workspaces[e.adaCfg.ActiveWorkspaceIndex]
			e.cfg.Agents.Defaults.Folders = ws.Folders
			e.cfg.Agents.Defaults.Personality = ws.Personality
			e.cfg.Agents.Defaults.Knowledge = ws.Knowledge
		}
	}

	// Inicializa o AgentLoop (Ada Love)
	rawProvider, _, err := providers.CreateProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar provider: %w", err)
	}
	provider := e.wrapProvider(rawProvider)
	e.agentLoop = agent.NewAgentLoop(cfg, msgBus, provider, e)

	return e, nil
}

func (e *Engine) wrapProvider(p providers.LLMProvider) providers.LLMProvider {
	wrapper := NewStreamingWrapper(p)
	wrapper.SetEventBus(e.eventBus)
	return wrapper
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

func (e *Engine) SaveAdaConfig() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	data, err := json.MarshalIndent(e.adaCfg, "", "  ")
	if err != nil {
		return err
	}
	// Sempre persiste no mesmo arquivo de onde foi carregado (OS config dir),
	// evitando que os dados sejam salvos num caminho relativo diferente.
	path := e.adaConfigPath
	if path == "" {
		path = "config/ada_config.json"
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	// Salva também no banco para persistência robusta de workspaces
	if e.db != nil {
		for _, ws := range e.adaCfg.Workspaces {
			e.db.SaveWorkspace(ws)
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
	return e.db.SaveProviders(e.adaCfg.Providers)
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
	if e.adaCfg.ActiveWorkspacePath != "" {
		e.cfg.Agents.Defaults.Workspace = e.adaCfg.ActiveWorkspacePath

		// Sincroniza as pastas, personalidade e conhecimento do workspace ativo
		if e.adaCfg.ActiveWorkspaceIndex >= 0 && e.adaCfg.ActiveWorkspaceIndex < len(e.adaCfg.Workspaces) {
			ws := e.adaCfg.Workspaces[e.adaCfg.ActiveWorkspaceIndex]
			e.cfg.Agents.Defaults.Folders = ws.Folders
			e.cfg.Agents.Defaults.Personality = ws.Personality
			e.cfg.Agents.Defaults.Knowledge = ws.Knowledge
		}
	}

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
	return nil
}

func (e *Engine) SubscribeEvents(handler func(Event)) int {
	return e.eventBus.Subscribe(handler)
}

func (e *Engine) UnsubscribeEvents(id int) {
	e.eventBus.Unsubscribe(id)
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
}

// getOSConfigDir returns the OS-specific config directory
func getOSConfigDir() string {
	var configDir string
	switch runtime.GOOS {
	case "linux":
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "ada-love")
	case "darwin":
		configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "ada-love")
	case "windows":
		configDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "ada-love")
	default:
		configDir = "config"
	}
	os.MkdirAll(configDir, 0755)
	return configDir
}
