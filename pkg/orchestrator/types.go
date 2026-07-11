package orchestrator

import (
	"context"
	"sync"
	"time"

	"ada-love-ai/pkg/providers"
)

// AgentType represents the type of agent
type AgentType string

const (
	AgentTypeOrchestrator AgentType = "orchestrator"
	AgentTypeGoLang       AgentType = "golang"
	AgentTypeReact        AgentType = "react"
	AgentTypeTester       AgentType = "tester"
	AgentTypeCustom       AgentType = "custom"
)

func (a AgentType) String() string {
	return string(a)
}

// AgentCapability represents a capability an agent can have
type AgentCapability string

const (
	CapabilityGoBackend       AgentCapability = "golang_backend"
	CapabilityDatabase        AgentCapability = "database"
	CapabilityAPI             AgentCapability = "api"
	CapabilityConcurrency     AgentCapability = "concurrency"
	CapabilityReactFrontend   AgentCapability = "react_frontend"
	CapabilityUIComponents    AgentCapability = "ui_components"
	CapabilityStateManagement AgentCapability = "state_management"
	CapabilityAPIIntegration  AgentCapability = "api_integration"
	CapabilityTesting         AgentCapability = "testing"
	CapabilityCodeReview      AgentCapability = "code_review"
	CapabilityQualityAssurance AgentCapability = "quality_assurance"
)

// SubAgentConfig holds configuration for a sub-agent
type SubAgentConfig struct {
	Model         string
	SystemPrompt  string
	Capabilities  []AgentCapability
	AllowedTools  []string
	MaxIterations int
	Temperature   float64
}

// PromptLayers represents the 4-layer prompt architecture
type PromptLayers struct {
	// Layer A: System Persona - Fixed per agent type
	SystemPersona string

	// Layer B: Global Context - Project context, architecture, tech stack
	GlobalContext string

	// Layer C: Short-term Memory / State - Previous agent outputs, decisions
	State string

	// Layer D: Current Task - The specific task from orchestrator
	Task string
}

// AgentResult represents the result of an agent execution
type AgentResult struct {
	Success  bool
	Output   string
	Error    error
	Metadata map[string]any
}

// ValidationResult represents the result of a validation
type ValidationResult struct {
	Passed          bool
	Issues          []Issue
	Coverage        float64
	MissingTests    []string
	SecurityIssues  []SecurityIssue
	Report          string
}

// Issue represents a validation issue
type Issue struct {
	Severity    string // Critical, High, Medium, Low
	Category    string // Testing, Security, Performance, CodeStyle, Architecture
	Description string
	File        string
	Line        int
	Suggestion  string
}

// SubTask represents a sub-task in the orchestration
type SubTask struct {
	ID         string    `json:"id"`
	Agent      AgentType `json:"agent"`
	Task       string    `json:"task"`
	DependsOn  []string  `json:"depends_on,omitempty"` // Task IDs this depends on
}

// RoutingDecision represents the orchestrator's routing decision
type RoutingDecision struct {
	Reasoning   string     `json:"reasoning"`
	NextAgent   AgentType  `json:"next_agent"`
	Task        string     `json:"task"`
	RelatedFiles []string  `json:"related_files"`
	RequiresTest bool      `json:"requires_test"`
	SubTasks    []SubTask  `json:"sub_tasks,omitempty"`
}

// SecurityIssue represents a security vulnerability
type SecurityIssue struct {
	Type        string // SQLInjection, XSS, AuthBypass, DataExposure, etc.
	Severity    string // Critical, High, Medium, Low
	Description string
	File        string
	Line        int
	Mitigation  string
}

// ValidationResult represents the orchestration history
type OrchestrationStep struct {
	Timestamp      time.Time
	UserInput      string
	Decision       RoutingDecision
	SubTaskResults []SubTaskResult
	FinalOutput    string
}

// SubTaskResult represents the result of a sub-task
type SubTaskResult struct {
	SubTask SubTask       `json:"sub_task"`
	Result    *AgentResult `json:"result"`
	Duration  time.Duration
	Metadata  map[string]any
}

// AgentResult extends the base with agent-specific metadata
type ExtendedAgentResult struct {
	AgentResult
	AgentType   AgentType
	Duration    time.Duration
	TokensUsed  int
	Iterations  int
}

// SubAgent interface defines the interface for sub-agents
type SubAgent interface {
	Type() AgentType
	Model() string
	SystemPrompt() string
	Capabilities() []AgentCapability
	Tools() []string
	SetProvider(providers.LLMProvider)
	Provider() providers.LLMProvider
	SetModel(string)
	Execute(ctx context.Context, task string, layers PromptLayers) (*AgentResult, error)
	Close() error
}

// OrchestratorConfig holds configuration for the orchestrator
type OrchestratorConfig struct {
	OrchestratorModel string
	AvailableAgents   []AgentType
	DefaultAgent      AgentType
	MaxSubTasks       int
	Timeout           time.Duration
	WorkspaceRoot     string
}

// DefaultOrchestratorConfig returns default configuration
func DefaultOrchestratorConfig() OrchestratorConfig {
	return OrchestratorConfig{
		OrchestratorModel: "gpt-4o",
		AvailableAgents: []AgentType{
			AgentTypeGoLang,
			AgentTypeReact,
			AgentTypeTester,
		},
		DefaultAgent: AgentTypeGoLang,
		MaxSubTasks:  10,
		Timeout:      5 * time.Minute,
	}
}

// AgentRegistry manages available sub-agents
type AgentRegistry struct {
	agents map[AgentType]SubAgent
	mu     sync.RWMutex
}

func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[AgentType]SubAgent),
	}
}

func (r *AgentRegistry) Register(agent SubAgent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[agent.Type()] = agent
}

func (r *AgentRegistry) Get(agentType AgentType) (SubAgent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agent, ok := r.agents[agentType]
	return agent, ok
}

func (r *AgentRegistry) List() []AgentType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]AgentType, 0, len(r.agents))
	for t := range r.agents {
		types = append(types, t)
	}
	return types
}

func (r *AgentRegistry) GetByCapability(targetCap AgentCapability) []SubAgent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var agents []SubAgent
	for _, agent := range r.agents {
		for _, c := range agent.Capabilities() {
			if c == targetCap {
				agents = append(agents, agent)
				break
			}
		}
	}
	return agents
}