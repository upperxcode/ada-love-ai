package backend

// SpecWizardConfig defines a specification wizard for generating project specs
// with configurable architecture, patterns, and technology stack.
type SpecWizardConfig struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Plugin selecionado (Agnostic, Go, Flutter, React, etc.)
	ExpertLanguagePlugin string `json:"expert_language_plugin,omitempty"`
	// PRD - Escopo do problema
	PRD string `json:"prd,omitempty"`
	// Requisitos Funcionais
	FunctionalRequirements []string `json:"functional_requirements,omitempty"`
	// Requisitos Não Funcionais
	NonFunctionalRequirements []string `json:"non_functional_requirements,omitempty"`
	// Persistência
	Persistence string `json:"persistence,omitempty"` // Custom, Remote, SQL, NoSQL, Mixed
	// Arquitetura
	Architecture string `json:"architecture,omitempty"` // Custom, Flat, Clean, CRUD, EventSourcing, CQRS, MVC
	// Padrões
	EngineeringPhilosophies []string `json:"engineering_philosophies,omitempty"` // DRY, KISS, SOLID, YAGNI
	DesignPatterns          []string `json:"design_patterns,omitempty"`          // Factory, Builder, Singleton, etc.
	DataPatterns            []string `json:"data_patterns,omitempty"`            // DTO, Entity, Repository, etc.
	// Stack Opinativa
	StackConfig []StackItem `json:"stack_config,omitempty"`
	// Negócio
	Business struct {
		StateManagement           string `json:"state_management,omitempty"`
		APIContract               string `json:"api_contract,omitempty"`
		CustomizationDetails      string `json:"customization_details,omitempty"`
		FinalAdjustments          string `json:"final_adjustments,omitempty"`
		ArchitectureRecommendations string `json:"architecture_recommendations,omitempty"`
	} `json:"business,omitempty"`
}

// StackItem representa um item na stack tecnológica
type StackItem struct {
	Name    string `json:"name"`
	Example string `json:"example,omitempty"`
}
