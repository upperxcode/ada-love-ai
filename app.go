package main

import (
	"context"
	"encoding/json"
	"fmt"
	gortime "runtime"
	"os"
	"path/filepath"
	"strings"

	"ada-love-ai/backend"
	"ada-love-ai/pkg/commands"
	"ada-love-ai/pkg/config"
	"ada-love-ai/pkg/patterns"
	"ada-love-ai/pkg/providers"
	"ada-love-ai/pkg/registry"
	integration "ada-love-ai/pkg/tools/integration"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v3"
)

type App struct {
	ctx    context.Context
	engine *backend.Engine
}

func NewApp(engine *backend.Engine) *App {
	return &App{engine: engine}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.startEventBridge()
	a.connectQuestionRegistry()
}

func (a *App) startEventBridge() {
	if a.engine == nil {
		return
	}
	a.engine.SubscribeEvents(func(ev backend.Event) {
		if a.ctx == nil {
			return
		}
		switch ev.Kind {
		case backend.EventKindLLMDelta:
			if payload, ok := ev.Payload.(backend.StreamingDeltaPayload); ok {
				runtime.EventsEmit(a.ctx, "chat:delta", map[string]interface{}{
					"session_id": ev.SessionID,
					"content":    payload.Content,
				})
			}
		case backend.EventKindTurnStart:
			runtime.EventsEmit(a.ctx, "chat:turnStart", map[string]interface{}{
				"session_id": ev.SessionID,
			})
		case backend.EventKindTurnEnd:
			runtime.EventsEmit(a.ctx, "chat:turnEnd", map[string]interface{}{
				"session_id": ev.SessionID,
			})
		case backend.EventKindError:
			if payload, ok := ev.Payload.(backend.ErrorPayload); ok {
				runtime.EventsEmit(a.ctx, "chat:error", map[string]interface{}{
					"session_id": ev.SessionID,
					"message":    payload.Message,
				})
			}
		case backend.EventKindStatus:
			if payload, ok := ev.Payload.(backend.StatusPayload); ok {
				runtime.EventsEmit(a.ctx, "chat:status", map[string]interface{}{
					"session_id": ev.SessionID,
					"stage":      payload.Message,
				})
			}
		case backend.EventKindCleared:
			runtime.EventsEmit(a.ctx, "chat:cleared", map[string]interface{}{
				"session_id": ev.SessionID,
			})
		}
	})
}

func (a *App) OpenDirectoryDialog() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Folder",
	})
	if err != nil {
		return ""
	}
	return dir
}

func (a *App) OpenFileDialog() string {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select File",
	})
	if err != nil {
		return ""
	}
	return file
}

// Config
func (a *App) GetAdaConfig() backend.AdaConfig {
	return a.engine.GetAdaConfig()
}
func (a *App) SetAdaConfig(cfg backend.AdaConfig) {
	a.engine.SetAdaConfig(cfg)
}

// Workers
func (a *App) GetWorkers() []backend.WorkerConfig {
	return a.engine.GetWorkers()
}
func (a *App) SetWorkers(workers []backend.WorkerConfig) {
	a.engine.SetWorkers(workers)
}
func (a *App) GetWorkerCategories() []string {
	return a.engine.GetWorkerCategories()
}
func (a *App) SetWorkerCategories(categories []string) {
	a.engine.SetWorkerCategories(categories)
}
func (a *App) GetPredefinedConnections() []backend.ConnectionDefinition {
	return backend.PredefinedConnections()
}
func (a *App) TestConnection(connectionType, connectionName, connectionConfig string) backend.ConnectionTestResult {
	return a.engine.TestConnection(connectionType, connectionName, connectionConfig)
}

// Legacy agent aliases (mapped to workers)
func (a *App) GetAgents() []backend.AgentConfig {
	return a.engine.GetAgents()
}
func (a *App) SetAgents(agents []backend.AgentConfig) {
	a.engine.SetAgents(agents)
}
func (a *App) GetAgentCategories() []string {
	return a.engine.GetAgentCategories()
}
func (a *App) SetAgentCategories(categories []string) {
	a.engine.SetAgentCategories(categories)
}

// Workspaces
func (a *App) GetWorkspaces() []backend.WorkspaceConfig {
	return a.engine.ListWorkspaces()
}
func (a *App) AddWorkspace(title, path, personality string) error {
	w := backend.WorkspaceConfig{
		Title:       title,
		Path:        path,
		Personality: personality,
		Tools:       []string{},
		Enabled:     true,
	}
	return a.engine.AddWorkspace(w)
}
func (a *App) DeleteWorkspace(title string) {
	a.engine.DeleteWorkspace(title)
}
func (a *App) SetActiveWorkspace(title string) {
	a.engine.SetActiveWorkspace(title)
}
func (a *App) ToggleWorkspace(title string) {
	a.engine.ToggleWorkspace(title)
}
func (a *App) UpdateWorkspace(originalTitle string, ws backend.WorkspaceConfig) {
	a.engine.UpdateWorkspace(originalTitle, ws)
}
func (a *App) AddToolToWorkspace(workspaceTitle, toolName string) bool {
	return a.engine.AddToolToWorkspace(workspaceTitle, toolName)
}
func (a *App) RemoveToolFromWorkspace(workspaceTitle, toolName string) bool {
	return a.engine.RemoveToolFromWorkspace(workspaceTitle, toolName)
}
func (a *App) SetWorkspaceTools(workspaceTitle string, toolNames []string) {
	a.engine.SetWorkspaceTools(workspaceTitle, toolNames)
}

// Tools & Profiles
func (a *App) GetToolProfiles() []backend.ToolProfile {
	return a.engine.GetToolProfiles()
}
func (a *App) CreateToolProfile(name, color, icon string) backend.ToolProfile {
	return a.engine.CreateToolProfile(name, color, icon)
}
func (a *App) DeleteToolProfile(id int64) bool {
	return a.engine.DeleteToolProfile(id)
}
func (a *App) ToggleProfileTool(profileID int64, toolName string, enabled bool) bool {
	return a.engine.ToggleProfileTool(profileID, toolName, enabled)
}
func (a *App) GetToolProfile(id int64) *backend.ToolProfile {
	return a.engine.GetToolProfile(id)
}
func (a *App) AddToolsToProfile(profileID int64, toolNames []string) bool {
	return a.engine.AddToolsToProfile(profileID, toolNames)
}
func (a *App) RemoveToolsFromProfile(profileID int64, toolNames []string) bool {
	return a.engine.RemoveToolsFromProfile(profileID, toolNames)
}
func (a *App) GetAvailableTools() []backend.ToolUIInfo {
	return a.engine.GetAvailableTools()
}
func (a *App) ToggleTool(toolName string, enabled bool) {
	a.engine.ToggleTool(toolName, enabled)
}

// Models
func (a *App) RemoveModel(name, provider string) {
	a.engine.RemoveModel(name, provider)
}
func (a *App) GetProvidersConfig() map[string]backend.ProviderConfig {
	return a.engine.GetProvidersConfig()
}
func (a *App) SaveProvidersConfig() {
	a.engine.SaveProvidersConfig()
}
func (a *App) GetProviders() []string {
	return a.engine.GetProviders()
}

// ListChatProviders expõe os providers de chat configurados para a UI de Agents.
func (a *App) ListChatProviders() []string {
	return a.engine.GetProviders()
}

// SaveDBProvider salva um único provider no DB e na memória.
func (a *App) SaveDBProvider(name string, cfg backend.ProviderConfig) error {
	return a.engine.SaveDBProvider(name, cfg)
}

// DeleteDBProvider remove um provider do DB e da memória.
func (a *App) DeleteDBProvider(name string) error {
	return a.engine.DeleteDBProvider(name)
}

// FetchProviderModels queries a provider's /models endpoint and returns the
// list enriched with detected capabilities (vision/embedding), keyed by
// connectionType protocol ("openai" | "anthropic" | "gemini").
func (a *App) FetchProviderModels(name, apiKey, apiBase, connectionType string) ([]backend.ProviderModel, error) {
	return a.engine.FetchProviderModels(name, apiKey, apiBase, connectionType)
}

// TestProviderConnection validates an API key (literal or env-var reference)
// against the provider's /models endpoint.
func (a *App) TestProviderConnection(name, apiKey, apiBase, connectionType string) (backend.ProviderTestResult, error) {
	return a.engine.TestProviderConnection(name, apiKey, apiBase, connectionType)
}

// Sessions / Chat
func (a *App) CreateSession(workspaceID, workerName string) *backend.ChatSession {
	fmt.Printf("[App] CreateSession: workspaceID=%q workerName=%q\n", workspaceID, workerName)
	sess := a.engine.SessionMgr.CreateSession("Nova Conversa", workspaceID, workerName, "")
	fmt.Printf("[App] CreateSession: created sessionID=%q title=%q workspaceID=%q workerName=%q parent=%q\n",
		sess.ID, sess.Title, sess.WorkspaceID, sess.WorkerName, sess.ParentSessionID)
	a.engine.SaveSessionDB(sess.ID)
	return sess
}

func (a *App) CreateSummarizedSession(workspaceID, workerName, sourceSessionID string) *backend.ChatSession {
	fmt.Printf("[App] CreateSummarizedSession: workspaceID=%q workerName=%q sourceSessionID=%q\n", workspaceID, workerName, sourceSessionID)
	sess := a.engine.SessionMgr.CreateSession("Resumo • "+workerName, workspaceID, workerName, sourceSessionID)
	fmt.Printf("[App] CreateSummarizedSession: created sessionID=%q parent=%q\n", sess.ID, sess.ParentSessionID)
	a.engine.SaveSessionDB(sess.ID)
	return sess
}

// GetSessions retorna sessões de um workspace específico do banco de dados
func (a *App) GetSessions(workspaceID string) []*backend.ChatSession {
	fmt.Printf("[App] GetSessions: workspaceID=%q — querying DB\n", workspaceID)
	if a.engine.DB() != nil {
		sessions, err := a.engine.DB().GetSessions(workspaceID)
		if err != nil {
			fmt.Printf("[App] GetSessions: DB error: %v\n", err)
			return nil
		}
		fmt.Printf("[App] GetSessions: workspaceID=%q → %d sessions from DB\n", workspaceID, len(sessions))
		for _, s := range sessions {
			fmt.Printf("[App]   session=%q title=%q worker=%q messages=%d parent=%q\n", s.ID, s.Title, s.WorkerName, len(s.Messages), s.ParentSessionID)
		}
		return sessions
	}
	// Fallback para SessionMgr se DB não disponível
	sessions := a.engine.SessionMgr.ListSessions(workspaceID)
	fmt.Printf("[App] GetSessions: fallback to SessionMgr → %d sessions\n", len(sessions))
	return sessions
}
func (a *App) DeleteSession(id string) {
	a.engine.DeleteSession(id)
}
func (a *App) RenameSession(id, newTitle string) *backend.ChatSession {
	return a.engine.RenameSession(id, newTitle)
}
func (a *App) SendMessage(sessionID, text, modelOverride, thinkingLevel, mode string) (result string, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 16384)
			n := gortime.Stack(buf, false)
			log := fmt.Sprintf("[App.SendMessage] PANIC RECOVERED: %v\n%s\n", r, buf[:n])
			fmt.Print(log)
			writePanicLog(log)
			result = ""
			retErr = fmt.Errorf("internal panic: %v", r)
		}
	}()
	fmt.Printf("[App.SendMessage] sessionID=%q modelOverride=%q thinkingLevel=%q mode=%q text=%q\n",
		sessionID, modelOverride, thinkingLevel, mode, text[:min(len(text), 50)])
	return a.engine.SendMessage(a.ctx, text, sessionID, modelOverride, thinkingLevel, mode, false)
}

func (a *App) RetryMessage(sessionID, text string) (string, error) {
	fmt.Printf("[App.RetryMessage] sessionID=%q text=%q\n", sessionID, text[:min(len(text), 50)])
	return a.engine.SendMessage(a.ctx, text, sessionID, "", "", "", true)
}

func (a *App) AnswerQuestion(sessionID, answer string) {
	a.engine.AnswerQuestion(sessionID, answer)
}
func (a *App) AnswerApproval(requestID string, approved bool, reason string) {
	a.engine.AnswerApproval(requestID, approved, reason)
}
func (a *App) StopGeneration(sessionID string) {
	a.engine.StopGeneration(sessionID)
}
func (a *App) TogglePin(sessionID string) {
	a.engine.TogglePin(sessionID)
}
func (a *App) SetSessionConfig(sessionID, model, provider, mode, thinking string) {
	a.engine.SetSessionConfig(sessionID, model, provider, mode, thinking)
}

// CommandInfo mirrors commands.Definition for JSON serialization to the frontend.
type CommandInfo struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Usage       string            `json:"usage"`
	Aliases     []string          `json:"aliases"`
	SubCommands []SubCommandInfo  `json:"sub_commands"`
}

// SubCommandInfo mirrors commands.SubCommand for JSON serialization.
type SubCommandInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ArgsUsage   string `json:"args_usage"`
}

// ListCommands returns all registered command definitions for the frontend
// to display in the slash command autocomplete menu.
func (a *App) ListCommands() []CommandInfo {
	defs := commands.BuiltinDefinitions()
	out := make([]CommandInfo, 0, len(defs))
	for _, def := range defs {
		ci := CommandInfo{
			Name:        def.Name,
			Description: def.Description,
			Usage:       def.EffectiveUsage(),
			Aliases:     def.Aliases,
		}
		if len(def.SubCommands) > 0 {
			ci.SubCommands = make([]SubCommandInfo, 0, len(def.SubCommands))
			for _, sc := range def.SubCommands {
				ci.SubCommands = append(ci.SubCommands, SubCommandInfo{
					Name:        sc.Name,
					Description: sc.Description,
					ArgsUsage:   sc.ArgsUsage,
				})
			}
		}
		out = append(out, ci)
	}
	return out
}

func (a *App) connectQuestionRegistry() {
	// Bridge ask_user questions
	qr := a.engine.QuestionRegistry()
	if qr != nil {
		qr.OnAsk(func(sessionID, question string) {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "chat:question", map[string]interface{}{
					"session_id": sessionID,
					"question":   question,
				})
			}
		})
		qr.OnAnswer(func(sessionID string) {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "chat:questionAnswered", map[string]interface{}{
					"session_id": sessionID,
				})
			}
		})
	}

	// Bridge tool approval requests
	ar := a.engine.ApprovalRegistry()
	if ar != nil {
		ar.OnApprove(func(req integration.ApprovalRequest) {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "chat:toolApproval", map[string]interface{}{
					"id":         req.ID,
					"session_id": req.SessionID,
					"tool":       req.Tool,
					"args":       req.Args,
				})
			}
		})
	}
}

// Skills
func (a *App) SearchSkills(query string) ([]backend.SearchResult, error) {
	return a.engine.SearchSkills(query)
}
func (a *App) InstallSkill(registryName, slug, version string) error {
	return a.engine.InstallSkill(registryName, slug, version)
}
func (a *App) GetInstalledSkills() ([]string, error) {
	return a.engine.GetInstalledSkills()
}
func (a *App) UninstallSkill(name string) error {
	return a.engine.UninstallSkill(name)
}
func (a *App) GetSkillDetails(name string) (string, error) {
	return a.engine.GetSkillDetails(name)
}
func (a *App) GetSkillFullInfo(name string) (*backend.SkillFullInfo, error) {
	return a.engine.GetSkillFullInfo(name)
}
func (a *App) SaveCustomSkill(name, description, tagsCSV, content string) error {
	return a.engine.SaveCustomSkill(name, description, tagsCSV, content)
}

// --- Spec Wizard Plugins bindings ---

// GetPatterns retorna os patterns do `pkg/patterns` filtrados pela linguagem.
// lang deve ser uma das supportedLangs. Caso contrário, retorna [].
func (a *App) GetPatterns(lang string) []map[string]any {
	repo := patterns.NewRepository()
	out := []map[string]any{}
	for _, p := range repo.GetPatternsForLanguage(lang) {
		out = append(out, patternToMap(p))
	}
	return out
}

// architecture representa uma entrada de config/architectures.yaml.
type architecture struct {
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	BestFor     []string `yaml:"best_for" json:"best_for"`
	Aliases     []string `yaml:"aliases" json:"aliases"`
}

// GetArchitectures lê config/architectures.yaml e retorna a lista.
func (a *App) GetArchitectures() ([]architecture, error) {
	paths := []string{"config/architectures.yaml"}
	if exe, err := os.Executable(); err == nil {
		wd := filepath.Dir(exe)
		paths = append(paths,
			filepath.Join(wd, "config", "architectures.yaml"),
			filepath.Join(wd, "..", "config", "architectures.yaml"),
		)
	}
	all := []architecture{}
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var parsed struct {
			Architectures []architecture `yaml:"architectures"`
		}
		if err := yaml.Unmarshal(data, &parsed); err != nil {
			return nil, err
		}
		for _, item := range parsed.Architectures {
			if !archExists(all, item.ID) {
				all = append(all, item)
			}
		}
	}
	return all, nil
}

// GetExperts carrega config/experts.yaml via pkg/registry.
func (a *App) GetExperts() ([]*registry.ExpertPlugin, error) {
	candidates := []string{"config/experts.yaml"}
	if exe, err := os.Executable(); err == nil {
		wd := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(wd, "config", "experts.yaml"),
			filepath.Join(wd, "..", "config", "experts.yaml"),
		)
	}
	return registry.LoadExperts(candidates...)
}

func archExists(list []architecture, id string) bool {
	for _, a := range list {
		if a.ID == id {
			return true
		}
	}
	return false
}

func patternToMap(p patterns.Pattern) map[string]any {
	return map[string]any{
		"id":          p.ID,
		"name":        p.Name,
		"category":    p.Category,
		"group":       p.Group,
		"scope":       p.Scope,
		"description": p.Description,
	}
}

// GetStacks retorna os stack templates para a linguagem especificada.
func (a *App) GetStacks(lang string) []map[string]any {
	repo := patterns.NewRepository()
	out := []map[string]any{}
	for _, s := range repo.GetStacksForLanguage(lang) {
		libs := []map[string]any{}
		for _, lib := range s.Libraries {
			libs = append(libs, map[string]any{
				"name":          lib.Name,
				"mandatory":     lib.Mandatory,
				"usage_example": lib.UsageExample,
			})
		}
		out = append(out, map[string]any{
			"id":        s.ID,
			"name":      s.Name,
			"libraries": libs,
		})
	}
	return out
}

// SuggestFieldValue usa o LLM para sugerir um valor para um campo do SpecWizard.
// Recebe o estado completo do wizard e gera um prompt específico por campo.
func (a *App) SuggestFieldValue(fieldName, wizardStateJSON, currentValue string) (string, error) {
	if a.engine == nil {
		return "", fmt.Errorf("engine not initialized")
	}

	adaCfg := a.engine.GetAdaConfig()
	specProvider := adaCfg.SpecProvider
	specModel := adaCfg.SpecModel

	if specProvider == "" || specModel == "" {
		return "", fmt.Errorf("no Spec Model configured. Please set a Spec Provider and Spec Model in Models settings.")
	}

	// Parse wizard state
	var ws map[string]interface{}
	if err := json.Unmarshal([]byte(wizardStateJSON), &ws); err != nil {
		return "", fmt.Errorf("invalid wizard state: %w", err)
	}

	// Build context string from non-empty fields
	var contextParts []string
	if v, ok := ws["language"].(string); ok && v != "" {
		contextParts = append(contextParts, "Language: "+v)
	}
	if v, ok := ws["architecture"].(string); ok && v != "" {
		contextParts = append(contextParts, "Architecture: "+v)
	}
	if v, ok := ws["persistence"].(string); ok && v != "" {
		contextParts = append(contextParts, "Persistence: "+v)
	}
	if v, ok := ws["stack"].(string); ok && v != "" {
		contextParts = append(contextParts, "Stack: "+v)
	}
	if v, ok := ws["prd"].(string); ok && v != "" {
		contextParts = append(contextParts, "PRD: "+v)
	}
	if v, ok := ws["description"].(string); ok && v != "" {
		contextParts = append(contextParts, "Description: "+v)
	}
	if v, ok := ws["engineering_philosophies"].([]interface{}); ok && len(v) > 0 {
		contextParts = append(contextParts, "Engineering Philosophies: "+strings.Join(stringSlice(v), ", "))
	}
	if v, ok := ws["design_patterns"].([]interface{}); ok && len(v) > 0 {
		contextParts = append(contextParts, "Design Patterns: "+strings.Join(stringSlice(v), ", "))
	}

	contextStr := strings.Join(contextParts, "\n")

	// Get field-specific prompt
	systemPrompt, tokenLimit := getFieldPrompt(fieldName)

	// Build user prompt
	userPrompt := fmt.Sprintf("%s\n\nContext:\n%s\n", systemPrompt, contextStr)
	if currentValue != "" {
		userPrompt += fmt.Sprintf("\nCurrent value (improve/refine if needed):\n%s\n", currentValue)
	}
	userPrompt += fmt.Sprintf("\nGenerate a value for: %s\nMax %d lines.", fieldName, tokenLimit)

	// Find and create provider
	var provider any
	var resolvedModel string

	for _, mc := range adaCfg.ModelList {
		if mc == nil {
			continue
		}
		p := strings.TrimSpace(mc.Provider)
		mn := strings.TrimSpace(mc.ModelName)
		mf := strings.TrimSpace(mc.Model)

		if mn == specModel || mf == specModel {
			if strings.EqualFold(p, specProvider) {
				p2, _, err := a.engine.CreateProviderFromModelConfig(mc)
				if err == nil && p2 != nil {
					provider = p2
					resolvedModel = mf
					if resolvedModel == "" {
						resolvedModel = specModel
					}
					break
				}
			}
		}
	}

	if provider == nil {
		providerCfg := config.ModelConfig{
			Provider:  specProvider,
			ModelName: specModel,
			Model:     specModel,
		}
		p2, modelID, err := a.engine.CreateProviderFromModelConfig(&providerCfg)
		if err == nil && p2 != nil {
			provider = p2
			resolvedModel = modelID
			if resolvedModel == "" {
				resolvedModel = specModel
			}
		}
	}

	if provider == nil {
		return "", fmt.Errorf("failed to create provider for model '%s'", specModel)
	}

	llmProvider, ok := provider.(providers.LLMProvider)
	if !ok {
		return "", fmt.Errorf("provider does not implement LLMProvider interface")
	}

	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := llmProvider.Chat(a.ctx, messages, nil, resolvedModel, nil)
	if err != nil {
		return "", fmt.Errorf("LLM request failed: %w", err)
	}

	if response == nil || response.Content == "" {
		return "", fmt.Errorf("no response from LLM")
	}

	return response.Content, nil
}

// stringSlice converts []interface{} to []string.
func stringSlice(in []interface{}) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// getFieldPrompt returns a tailored system prompt and token limit for each field.
func getFieldPrompt(fieldName string) (string, int) {
	switch fieldName {
	case "PRD":
		return `You are a product analyst. Write a concise Product Requirements Document (PRD) for a software project.

RULES:
- Write in ENGLISH
- Max 8 lines total
- Structure: Problem Statement, Target Users, Core Features (3-5 bullet points), Success Criteria
- Be specific to the given language and stack
- Use plain text, no markdown, no code blocks
- If a PRD is already provided, refine and improve it (don't repeat)`, 8

	case "Functional Requirements":
		return `You are a software architect. List functional requirements for this project.

RULES:
- Write in ENGLISH
- Max 8 lines total
- One requirement per line, starting with a verb (e.g., "Allow users to...")
- Be specific to the given language and stack
- Focus on core business features, not technical implementation
- Use plain text, no markdown, no code blocks`, 8

	case "Non-Functional Requirements":
		return `You are a software architect. List non-functional requirements for this project.

RULES:
- Write in ENGLISH
- Max 6 lines total
- One requirement per line (e.g., "Response time under 200ms for API endpoints")
- Cover: performance, security, scalability, accessibility, availability
- Be realistic for the given stack
- Use plain text, no markdown, no code blocks`, 6

	case "API Contract":
		return `You are a backend architect. Define the API contract for this project.

RULES:
- Write in ENGLISH
- Max 8 lines total
- List 3-5 key REST endpoints with method, path, and brief description
- Format: "GET /api/resource — description"
- Be specific to the given architecture and persistence strategy
- Use plain text, no markdown, no code blocks`, 8

	case "Customization Details":
		return `You are a solution architect. Describe customization and special considerations.

RULES:
- Write in ENGLISH
- Max 5 lines total
- List any non-standard behavior, business rules, or edge cases
- Be specific to the given stack and architecture
- Use plain text, no markdown, no code blocks`, 5

	case "Final Adjustments":
		return `You are a technical advisor. Suggest final adjustments before implementation.

RULES:
- Write in ENGLISH
- Max 5 lines total
- List 2-3 specific action items or warnings
- Focus on what the developer should verify or configure first
- Be practical and actionable
- Use plain text, no markdown, no code blocks`, 5

	case "Architecture Recommendations":
		return `You are a senior software architect. Provide architecture recommendations for this project.

RULES:
- Write in ENGLISH
- Max 6 lines total
- One recommendation per line, starting with a bullet point (•)
- Cover: code organization, error handling, testing strategy, deployment
- Be specific to the chosen language, architecture, and stack
- Use plain text, no markdown, no code blocks`, 6

	default:
		return fmt.Sprintf("You are an expert. Suggest a value for the field '%s'. Be concise. Max 5 lines. Write in ENGLISH.", fieldName), 5
	}
}

func writePanicLog(log string) {
	f, err := os.OpenFile("/tmp/ada-panic.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(log)
}
