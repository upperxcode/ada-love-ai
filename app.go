package main

import (
	"context"

	"ada-love-ai/backend"
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

// Config
func (a *App) GetAdaConfig() backend.AdaConfig {
	return a.engine.GetAdaConfig()
}
func (a *App) SetAdaConfig(cfg backend.AdaConfig) {
	a.engine.SetAdaConfig(cfg)
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

