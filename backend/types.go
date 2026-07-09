package backend

import (
	"ada-love-ai/pkg/config"
	"time"
)

type WorkerConfig struct {
	Name             string `json:"name"`
	Persona          string `json:"persona"`
	Language         string `json:"language"`         // idioma de resposta ao usuário (ex: "pt-BR", "en", "es")
	// Conexão (binding) — como o worker se comunica
	ConnectionType   string `json:"connection_type"`   // "ada", "cli", "rest", "mcp"
	ConnectionName   string `json:"connection_name"`   // nome do preset (ex: "Crush", "OpenCode")
	ConnectionConfig string `json:"connection_config"` // JSON com config específica da conexão
	// Inheritência do workspace (toggles)
	InheritFolders   bool   `json:"inherit_folders"`
	InheritKnowledge bool   `json:"inherit_knowledge"`
	InheritSkills    bool   `json:"inherit_skills"`
	InheritTools     bool   `json:"inherit_tools"`
	InheritPersona   bool   `json:"inherit_persona"`
}

// AgentConfig defines a specialized model that executes and/or delegates tasks
// to other models. Unlike Workers (which are chat interfaces), Agents are
// autonomous task executors: they receive a task, work on it (possibly using
// tools or sub-models), and deliver a result.
type AgentConfig struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	// Modelo que o agente usa para executar tarefas
	Provider    string   `json:"provider"`
	Model       string   `json:"model"`
	// Comportamento
	Type        string   `json:"type"`        // "executor", "delegator", "reviewer", "researcher"
	Icon        string   `json:"icon"`
	Color       string   `json:"color"`
	// Configuração de execução
	MaxIterations int    `json:"max_iterations,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
	// Modelos que este agente pode delegar (para type "delegator")
	Delegates   []string `json:"delegates,omitempty"`
	// Sistema de prompt customizado
	SystemPrompt string  `json:"system_prompt,omitempty"`
}

type SkillFullInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	Registry    string   `json:"registry,omitempty"`
	URL         string   `json:"url,omitempty"`
	Markdown    string   `json:"markdown,omitempty"`
	Raw         string   `json:"raw,omitempty"`
	LineCount   int      `json:"line_count,omitempty"`
	CharCount   int      `json:"char_count,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type SearchResult struct {
	Name         string  `json:"name"`
	DisplayName  string  `json:"display_name"`
	RegistryName string  `json:"registry_name"`
	Summary      string  `json:"summary"`
	Description  string  `json:"description"`
	Slug         string  `json:"slug"`
	Version      string  `json:"version"`
	Score        float64 `json:"score"`
}

type WorkspaceConfig struct {
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Path             string   `json:"path"`
	Folders          []string `json:"folders"`
	Personality      string   `json:"personality"`
	Knowledge        []string `json:"knowledge"`
	WorkspaceAgents  []WorkerConfig `json:"workspace_agents"`
	Skills           []string `json:"skills"`
	Tools            []string `json:"tools"`
	Enabled          bool     `json:"enabled"`
	Color            string   `json:"color"`
	Icon             string   `json:"icon"`
	MaxPromptSend    int      `json:"max_prompt_send"`
	CommitChanges    bool     `json:"commit_changes"`
	MaxContextLength int      `json:"max_context_length"`
}

type ToolUIInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Enabled     bool   `json:"enabled"`
}

type ExtraModelConfig struct {
	ContextSize int     `json:"context_size"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	TopP        float64 `json:"top_p"`
}

type ModelConfig struct {
	ModelName string `json:"model_name"`
	Model     string `json:"model"`
	Provider  string `json:"provider"`
	APIBase   string `json:"api_base"`
	Enabled   bool   `json:"enabled"`
}

type AdaConfig struct {
	ActiveWorkspacePath string            `json:"active_workspace_path"`
	ActiveWorkspaceIndex int              `json:"active_workspace_index"`
	Workspaces          []WorkspaceConfig `json:"workspaces"`
	TinyBrain           struct {
		ModelName         string `json:"model_name"`
		Provider          string `json:"provider"`
		EmbeddingModel    string `json:"embedding_model"`
		EmbeddingProvider string `json:"embedding_provider"`
	} `json:"tiny_brain"`
	EmbeddingModel    string `json:"embedding_model"`
	EmbeddingProvider string `json:"embedding_provider"`
	ImageModel        string `json:"image_model"`
	ImageProvider     string `json:"image_provider"`
	SpecModel         string `json:"spec_model"`
	SpecProvider      string `json:"spec_provider"`
	Workers              []WorkerConfig     `json:"workers"`
	WorkerCategories     []string          `json:"worker_categories"`
	Agents               []AgentConfig     `json:"agents"`
	AgentCategories      []string          `json:"agent_categories"`
	ProviderKeys        map[string]string `json:"provider_keys"`
	ProviderBases       map[string]string `json:"provider_bases"`
	ModelSettings       map[string]ExtraModelConfig `json:"model_settings"`
	ModelList           config.SecureModelList      `json:"model_list"`
	Providers           map[string]ProviderConfig   `json:"providers,omitempty"`
	ToolProfiles        []ToolProfile               `json:"tool_profiles,omitempty"`
}

// ProviderConfig represents a unified provider configuration.
type ProviderConfig struct {
	ApiUrl         string                    `json:"api_url"`
	ApiKey         string                    `json:"api_key"`
	TypeConnection string                    `json:"type_connection"`
	Models         map[string]ModelSettings  `json:"models"`
}

// ModelSettings represents per-model settings.
type ModelSettings struct {
	ContextSize int     `json:"context_size,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	Type        string  `json:"type,omitempty"`
	Vision      bool    `json:"vision,omitempty"`
	Embedding   bool    `json:"embedding,omitempty"`
	Tools       bool    `json:"tools,omitempty"`
	Free        bool    `json:"free,omitempty"`
	Thinking    bool    `json:"thinking,omitempty"`
}

// ProviderModel represents a model returned by a provider's /models endpoint,
// enriched with detected capabilities so the UI can filter by them.
type ProviderModel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Vision    bool   `json:"vision,omitempty"`   // accepts image input
	Embedding bool   `json:"embedding,omitempty"` // produces embeddings
	Tools     bool   `json:"tools,omitempty"`     // supports tool/function calling
	Free      bool   `json:"free,omitempty"`      // free / open-weight / no per-token cost
	Thinking  bool   `json:"thinking,omitempty"` // reasoning / chain-of-thought
}

// ProviderTestResult is the outcome of a connection test against a provider's
// /models endpoint. Both Ok and Success are populated (mirroring the frontend's
// ProviderTestResult class) for backwards compatibility.
type ProviderTestResult struct {
	Ok      bool   `json:"ok"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ToolProfile struct {
	ID    int64    `json:"id"`
	Name  string   `json:"name"`
	Color string   `json:"color"`
	Icon  string   `json:"icon"`
	Tools []string `json:"tools"`
}

// --- Eventos da UI ---

type EventKind int

const (
	EventKindLLMDelta EventKind = iota
	EventKindStatus
	EventKindTurnStart
	EventKindTurnEnd
	EventKindToolExecStart
	EventKindToolExecEnd
	EventKindError
	EventKindWorkspaceChanged
	EventKindWorkspaceDeleted
)

type Event struct {
	Kind      EventKind
	SessionID string // ID da sessão vinculada ao evento (vazio para eventos globais)
	Payload   interface{}
	Time      time.Time
}

type StreamingDeltaPayload struct {
	Content string
}

type StatusPayload struct {
	Message string
}

type ToolExecStartPayload struct {
	Tool string
	Args string
}

type ErrorPayload struct {
	Message string
}
