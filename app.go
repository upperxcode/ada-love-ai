package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"ada-love-ai/backend"
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
	sess := a.engine.SessionMgr.CreateSession("Nova Conversa", workspaceID, workerName)
	a.engine.SaveSessionDB(sess.ID)
	return sess
}

func (a *App) CreateSummarizedSession(workspaceID, workerName, sourceSessionID string) *backend.ChatSession {
	sess := a.engine.SessionMgr.CreateSession("Resumo • "+workerName, workspaceID, workerName)
	a.engine.SaveSessionDB(sess.ID)
	// Future: copy summary from sourceSessionID.
	_ = sourceSessionID
	return sess
}

func (a *App) GetSessions(workspaceID string) []*backend.ChatSession {
	return a.engine.SessionMgr.ListSessions(workspaceID)
}
func (a *App) DeleteSession(id string) {
	a.engine.DeleteSession(id)
}
func (a *App) RenameSession(id, newTitle string) {
	a.engine.RenameSession(id, newTitle)
}
func (a *App) SendMessage(sessionID, text, modelOverride, thinkingLevel, mode string) (string, error) {
	fmt.Printf("[App.SendMessage] sessionID=%q modelOverride=%q thinkingLevel=%q mode=%q text=%q\n",
		sessionID, modelOverride, thinkingLevel, mode, text[:min(len(text), 50)])
	return a.engine.SendMessage(a.ctx, text, sessionID, modelOverride, thinkingLevel, mode)
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
func (a *App) SuggestFieldValue(fieldName, context, currentValue string) (string, error) {
	fmt.Printf("[App.SuggestFieldValue] fieldName=%q currentValue=%q\n", fieldName, currentValue)
	if a.engine == nil {
		fmt.Println("[App.SuggestFieldValue] ERROR: engine not initialized")
		return "", fmt.Errorf("engine not initialized")
	}

	adaCfg := a.engine.GetAdaConfig()
	specProvider := adaCfg.SpecProvider
	specModel := adaCfg.SpecModel
	fmt.Printf("[App.SuggestFieldValue] specProvider=%q specModel=%q\n", specProvider, specModel)

	if specProvider == "" || specModel == "" {
		fmt.Println("[App.SuggestFieldValue] ERROR: no Spec Model configured")
		return "", fmt.Errorf("no Spec Model configured. Please set a Spec Provider and Spec Model in Models settings.")
	}

	// Find the provider config
	provCfg, ok := adaCfg.Providers[specProvider]
	if !ok {
		fmt.Printf("[App.SuggestFieldValue] ERROR: provider %q not found\n", specProvider)
		return "", fmt.Errorf("provider '%s' not found in configured providers", specProvider)
	}
	fmt.Printf("[App.SuggestFieldValue] provider found: api_url=%q type=%q\n", provCfg.ApiUrl, provCfg.TypeConnection)

	// Create provider from config
	providerCfg := config.ModelConfig{
		Provider:    specProvider,
		ModelName:   specModel,
		Model:       specModel,
		APIBase:     provCfg.ApiUrl,
		APIKeys:     config.SimpleSecureStrings(provCfg.ApiKey),
		ConnectMode: provCfg.TypeConnection,
	}
	fmt.Printf("[App.SuggestFieldValue] creating provider: provider=%q model=%q\n", specProvider, specModel)

	provider, _, err := providers.CreateProviderFromConfig(&providerCfg)
	if err != nil {
		fmt.Printf("[App.SuggestFieldValue] ERROR creating provider: %v\n", err)
		return "", fmt.Errorf("failed to create provider: %w", err)
	}
	fmt.Println("[App.SuggestFieldValue] provider created successfully")

	systemPrompt := "You are an expert software architect. Generate concise, practical suggestions for specification fields. Return ONLY the suggested value, no explanations, no markdown, no formatting."
	userPrompt := fmt.Sprintf("Field: %s\nContext: %s\nCurrent value: %s\n\nSuggest a value for this field:", fieldName, context, currentValue)

	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	fmt.Println("[App.SuggestFieldValue] calling LLM...")
	response, err := provider.Chat(a.ctx, messages, nil, specModel, nil)
	if err != nil {
		fmt.Printf("[App.SuggestFieldValue] ERROR calling LLM: %v\n", err)
		return "", fmt.Errorf("LLM request failed: %w", err)
	}

	if response == nil || response.Content == "" {
		fmt.Println("[App.SuggestFieldValue] ERROR: empty response from LLM")
		return "", fmt.Errorf("no response from LLM")
	}

	fmt.Printf("[App.SuggestFieldValue] response length=%d\n", len(response.Content))
	return response.Content, nil
}
