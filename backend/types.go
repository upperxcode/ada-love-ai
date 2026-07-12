package backend

import (
	"ada-love-ai/pkg/config"
	"os"
	"strings"
	"time"
)

type WorkerConfig struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Persona          string `json:"persona"`
	ResponseLanguage string `json:"response_language"` // idioma de resposta ao usuário (ex: "pt-BR", "en", "es")
	Icon             string `json:"icon"`
	Color            string `json:"color"`
	// Conexão (binding) — como o worker se comunica
	ConnectionType string `json:"connection_type"` // "websocket", "http_command", "cli", "mcp"
	Command        string `json:"command"`         // comando/preset (ex: "Crush", "OpenCode")
	Arguments      string `json:"arguments"`       // JSON array de argumentos
	Environment    string `json:"environment"`     // JSON array de variaveis/paths
	// Campos legados mantidos para compatibilidade de API (mapeados p/ command/arguments)
	ConnectionName   string `json:"connection_name,omitempty"`
	ConnectionConfig string `json:"connection_config,omitempty"`
	Language         string `json:"language,omitempty"`
	// Herança do workspace (toggles)
	InheritFolders   bool `json:"inherit_folders"`
	InheritKnowledge bool `json:"inherit_knowledge"`
	InheritSkills    bool `json:"inherit_skills"`
	InheritTools     bool `json:"inherit_tools"`
	InheritPersona   bool `json:"inherit_persona"`
}

// AgentConfig defines a specialized model that executes and/or delegates tasks
// to other models. Unlike Workers (which are chat interfaces), Agents are
// autonomous task executors: they receive a task, work on it (possibly using
// tools or sub-models), and deliver a result.
type AgentConfig struct {
	ID            int64            `json:"id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	Provider      string           `json:"provider"` // nome do provider (resolvido p/ provider_id)
	Model         string           `json:"model"`    // nome do modelo (resolvido p/ model_id)
	ProviderID    int64            `json:"provider_id,omitempty"`
	ModelID       int64            `json:"model_id,omitempty"`
	Type          string           `json:"type"` // "executor", "delegator", "reviewer", "research"
	Icon          string           `json:"icon"`
	Color         string           `json:"color"`
	MaxIterations int              `json:"max_iterations,omitempty"`
	Temperature   float64          `json:"temperature,omitempty"`
	Delegates     []string         `json:"delegates,omitempty"`
	SystemPrompt  string           `json:"system_prompt,omitempty"`
	Subagents     *SubagentsConfig `json:"subagents,omitempty"`
}

// SubagentsConfig defines which agents this agent can spawn.
type SubagentsConfig struct {
	AllowAgents []string `json:"allow_agents,omitempty"`
}

// SpecWizardConfig defines a specification wizard for generating project specs
// with configurable architecture, patterns, and technology stack.
type SpecWizardConfig struct {
	ID                                  string      `json:"id"`
	Name                                string      `json:"name"`
	Description                         string      `json:"description,omitempty"`
	ExpertLanguagePlugin                string      `json:"expert_language_plugin,omitempty"`
	PRD                                 string      `json:"prd,omitempty"`
	FunctionalRequirements              []string    `json:"functional_requirements,omitempty"`
	NonFunctionalRequirements           []string    `json:"non_functional_requirements,omitempty"`
	Persistence                         string      `json:"persistence,omitempty"`
	Architecture                        string      `json:"architecture,omitempty"`
	EngineeringPhilosophies             []string    `json:"engineering_philosophies,omitempty"`
	DesignPatterns                      []string    `json:"design_patterns,omitempty"`
	DataPatterns                        []string    `json:"data_patterns,omitempty"`
	StackConfig                         []StackItem `json:"stack_config,omitempty"`
	BusinessStateManagement             string      `json:"business_state_management,omitempty"`
	BusinessAPIContract                 string      `json:"business_api_contract,omitempty"`
	BusinessCustomizationDetails        string      `json:"business_customization_details,omitempty"`
	BusinessFinalAdjustments            string      `json:"business_final_adjustments,omitempty"`
	BusinessArchitectureRecommendations string      `json:"business_architecture_recommendations,omitempty"`
	Color                               string      `json:"color"`
	Icon                                string      `json:"icon"`
	ArchitectureHealth                  int         `json:"architecture_health"`
	CreatedAt                           time.Time   `json:"created_at"`
	UpdatedAt                           time.Time   `json:"updated_at"`
}

// StackItem represents an item in the technology stack
type StackItem struct {
	Name    string `json:"name"`
	Example string `json:"example,omitempty"`
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
	ID                int64    `json:"id" db:"id"`
	Nome              string   `json:"nome" db:"nome"`
	Description       string   `json:"description" db:"description"`
	MaxPrompt         int      `json:"max_prompt" db:"max_prompt"`
	MaxContent        int      `json:"max_content" db:"max_content"`
	Commit            bool     `json:"commit" db:"commit"`
	SpecProvider      string   `json:"spec_provider" db:"spec_provider"`
	SpecWizardID      string   `json:"spec_wizard_id" db:"spec_wizard_id"`
	Personality       string   `json:"personality" db:"personality"`
	RoutingRules      string   `json:"routing_rules" db:"routing_rules"`
	Color             string   `json:"color" db:"color"`
	Icon              string   `json:"icon" db:"icon"`
	Title             string   `json:"title"`
	Summary           string   `json:"summary"`
	Path              string   `json:"path"`
	Folders           []string `json:"folders"`
	Knowledge         []string `json:"knowledge"`
	WorkerNames       []string `json:"worker_names"`
	Agents            []string `json:"agents"`
	Skills            []string `json:"skills"`
	Tools             []string `json:"tools"`
	Enabled           bool     `json:"enabled"`
	MaxPromptSend     int      `json:"max_prompt_send"`
	CommitChanges     bool     `json:"commit_changes"`
	MaxContextLength  int      `json:"max_context_length"`
	SpecWizard        string   `json:"spec_wizard"`
	EmbeddingModel    string   `json:"embedding_model"`
	EmbeddingProvider string   `json:"embedding_provider"`
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
	ActiveWorkspacePath  string            `json:"active_workspace_path"`
	ActiveWorkspaceIndex int               `json:"active_workspace_index"`
	Workspaces           []WorkspaceConfig `json:"workspaces"`
	TinyBrain            struct {
		ModelName         string   `json:"model_name"`
		Provider          string   `json:"provider"`
		EmbeddingModel    string   `json:"embedding_model"`
		EmbeddingProvider string   `json:"embedding_provider"`
		Tools             []string `json:"tools"`
	} `json:"tiny_brain"`
	Classifier struct {
		ModelName string   `json:"model_name"`
		Provider  string   `json:"provider"`
		Tools     []string `json:"tools"`
	} `json:"classifier"`
	EmbeddingModel    string                      `json:"embedding_model"`
	EmbeddingProvider string                      `json:"embedding_provider"`
	ImageModel        string                      `json:"image_model"`
	ImageProvider     string                      `json:"image_provider"`
	SpecModel         string                      `json:"spec_model"`
	SpecProvider      string                      `json:"spec_provider"`
	SpecTools         []string                    `json:"spec_tools"`
	Workers           []WorkerConfig              `json:"workers"`
	WorkerCategories  []string                    `json:"worker_categories"`
	Agents            []AgentConfig               `json:"agents"`
	AgentCategories   []string                    `json:"agent_categories"`
	ProviderKeys      map[string]string           `json:"provider_keys"`
	ProviderBases     map[string]string           `json:"provider_bases"`
	ModelSettings     map[string]ExtraModelConfig `json:"model_settings"`
	ModelList         config.SecureModelList      `json:"model_list"`
	Providers         map[string]ProviderConfig   `json:"providers,omitempty"`
	ToolProfiles      []ToolProfile               `json:"tool_profiles,omitempty"`
	MCPServers        map[string]MCPServerUI      `json:"mcp_servers,omitempty"`
	SpecWizards       []SpecWizardConfig          `json:"spec_wizards"`
	Templates         []WorkspaceTemplate         `json:"templates"`
}

// MCPServerUI extends MCP server config with UI fields (icon, color).
type MCPServerUI struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url"`
	Enabled bool              `json:"enabled"`
	Icon    string            `json:"icon"`
	Color   string            `json:"color"`
}

// ProviderApiKey represents a single API key for a provider.
type ProviderApiKey struct {
	Key     string `json:"key"`
	UserKey string `json:"user_key,omitempty"`
}

// ProviderConfig represents a unified provider configuration.
type ProviderConfig struct {
	Icon           string                   `json:"icon"`
	Color          string                   `json:"color"`
	ApiUrl         string                   `json:"api_url"`
	ApiKey         string                   `json:"api_key,omitempty"`  // Legacy single key
	ApiKeys        []ProviderApiKey         `json:"api_keys,omitempty"` // New format: array of keys
	TypeConnection string                   `json:"type_connection"`
	Models         map[string]ModelSettings `json:"models"`
}

// GetAPIKey returns the first API key from the provider config.
// Supports both legacy (api_key) and new (api_keys) formats.
func (p *ProviderConfig) GetAPIKey() string {
	if len(p.ApiKeys) > 0 {
		return p.ApiKeys[0].Key
	}
	return p.ApiKey
}

// GetAllAPIKeys returns all API keys as a string slice.
func (p *ProviderConfig) GetAllAPIKeys() []string {
	if len(p.ApiKeys) > 0 {
		keys := make([]string, 0, len(p.ApiKeys))
		for _, k := range p.ApiKeys {
			if k.Key != "" {
				keys = append(keys, k.Key)
			}
		}
		return keys
	}
	if p.ApiKey != "" {
		return []string{p.ApiKey}
	}
	return nil
}

// GetProviderAPIKey resolves the API key for a provider name.
// Checks: 1) ProviderConfig.api_keys, 2) ProviderConfig.api_key, 3) AdaConfig.ProviderKeys map.
// 4) Environment variable (e.g., OPENROUTER_API_KEY)
func (c *AdaConfig) GetProviderAPIKey(providerName string) string {
	lower := strings.ToLower(providerName)
	for key, provider := range c.Providers {
		if strings.ToLower(key) == lower {
			if k := provider.GetAPIKey(); k != "" {
				return k
			}
		}
	}
	for key, val := range c.ProviderKeys {
		if strings.ToLower(key) == lower {
			return val
		}
	}
	envKey := strings.ToUpper(strings.ReplaceAll(providerName, "-", "_")) + "_API_KEY"
	return os.Getenv(envKey)
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
	Vision    bool   `json:"vision,omitempty"`    // accepts image input
	Embedding bool   `json:"embedding,omitempty"` // produces embeddings
	Tools     bool   `json:"tools,omitempty"`     // supports tool/function calling
	Free      bool   `json:"free,omitempty"`      // free / open-weight / no per-token cost
	Thinking  bool   `json:"thinking,omitempty"`  // reasoning / chain-of-thought
}

// WorkspaceCount returns the number of workspaces.
func (c *AdaConfig) WorkspaceCount() int {
	return len(c.Workspaces)
}

type ToolProfile struct {
	ID    int64    `json:"id"`
	Name  string   `json:"name"`
	Color string   `json:"color"`
	Icon  string   `json:"icon"`
	Tools []string `json:"tools"`
}

// WorkspaceTemplate holds a pre-configured personality template for workspace creation.
type WorkspaceTemplate struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Personality string `json:"personality" db:"personality"`
	CreatedAt   string `json:"created_at" db:"created_at"`
}

// ProviderTestResult is the outcome of a connection test against a provider's
// /models endpoint. Both Ok and Success are populated (mirroring the frontend's
// ProviderTestResult class) for backwards compatibility.
type ProviderTestResult struct {
	Ok      bool   `json:"ok"`
	Success bool   `json:"success"`
	Message string `json:"message"`
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
	EventKindCleared
	EventKindOrchestratorDecision
	EventKindSubTaskStart
	EventKindSubTaskComplete
	EventKindSubTaskError
	EventKindCommandResult
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

type OrchestratorDecisionPayload struct {
	Reasoning    string   `json:"reasoning"`
	NextAgent    string   `json:"next_agent"`
	Task         string   `json:"task"`
	SubTasks     int      `json:"sub_tasks"`
	AgentCount   int      `json:"agent_count"`
	RelatedFiles []string `json:"related_files"`
}

type SubTaskPayload struct {
	ID     string `json:"id"`
	Agent  string `json:"agent"`
	Task   string `json:"task"`
	Status string `json:"status,omitempty"` // "started", "completed", "error"
	Error  string `json:"error,omitempty"`
}

// CommandResultPayload carries the structured output of a handled slash
// command so the frontend can render it in a dedicated panel instead of the
// chat bubble stream.
type CommandResultPayload struct {
	Command string `json:"command"`
	Output  string `json:"output"`
}
