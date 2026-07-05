package backend

import (
	"ada-love-ai/pkg/config"
	"time"
)

type AgentConfig struct {
	Name     string `json:"name"`
	Persona  string `json:"persona"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Category string `json:"category"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
}

type SkillFullInfo struct {
	Name        string
	Description string
	Version     string
	Registry    string
	URL         string
	Markdown    string
	Raw         string
	LineCount   int
	CharCount   int
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
	WorkspaceAgents  []string `json:"workspace_agents"`
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
	Agents              []AgentConfig     `json:"agents"`
	AgentCategories     []string          `json:"agent_categories"`
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
