package main

import (
	"context"

	"ada-love-ai/backend"

	"github.com/wailsapp/wails/v2/pkg/runtime"
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

// Agents
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
func (a *App) CreateSession(workspaceID string) *backend.ChatSession {
	return a.engine.SessionMgr.CreateSession("Nova Conversa", workspaceID)
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
func (a *App) SendMessage(sessionID, text string) (string, error) {
	return a.engine.SendMessage(a.ctx, text, sessionID)
}
func (a *App) TogglePin(sessionID string) {
	a.engine.TogglePin(sessionID)
}
