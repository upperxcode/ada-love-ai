package backend

import (
	"path/filepath"
	"strings"
	"time"
)

func (e *Engine) ListWorkspaces() []WorkspaceConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.adaCfg.Workspaces
}

func (e *Engine) AddWorkspace(w WorkspaceConfig) error {
	e.mu.Lock()
	
	// Garantir título único
	w.Title = UniquifyName(w.Title, func(t string) bool {
		for _, ws := range e.adaCfg.Workspaces {
			if strings.EqualFold(ws.Title, t) {
				return true
			}
		}
		return false
	})

	e.adaCfg.Workspaces = append(e.adaCfg.Workspaces, w)
	e.mu.Unlock()
	return e.SaveAdaConfig()
}

func (e *Engine) SetActiveWorkspace(path string) {
	e.mu.Lock()
	e.adaCfg.ActiveWorkspacePath = path
	
	// Sincroniza o índice para o UI
	for i, w := range e.adaCfg.Workspaces {
		if w.Path == path {
			e.adaCfg.ActiveWorkspaceIndex = i
			break
		}
	}
	e.mu.Unlock()

	e.SaveAdaConfig()
	e.ReloadAgentLoop()
	
	// Recarregar sessões do banco de dados para este workspace
	e.RefreshSessions()

	// Notificar que o workspace mudou
	e.eventBus.Emit(Event{
		Kind:    EventKindWorkspaceChanged,
		Payload: path,
		Time:    time.Now(),
	})
}

func (e *Engine) DeleteWorkspace(path string) {
	e.mu.Lock()
	var newList []WorkspaceConfig
	for _, w := range e.adaCfg.Workspaces {
		if w.Path == path {
			continue
		}
		newList = append(newList, w)
	}
	e.adaCfg.Workspaces = newList
	if e.adaCfg.ActiveWorkspacePath == path {
		if len(newList) > 0 {
			e.adaCfg.ActiveWorkspacePath = newList[0].Path
		} else {
			e.adaCfg.ActiveWorkspacePath = ""
		}
	}
	e.mu.Unlock()
	e.SaveAdaConfig()
	
	e.eventBus.Emit(Event{
		Kind:    EventKindWorkspaceChanged, // Usamos changed para forçar refresh
		Payload: e.adaCfg.ActiveWorkspacePath,
		Time:    time.Now(),
	})
}

func (e *Engine) RegisterWorkspaceTools(path string) {
	// Implementação futura para carregar ferramentas específicas do workspace
}

func (e *Engine) GetAvailableTools() []ToolUIInfo {
	// Lista fixa de ferramentas disponíveis no momento
	available := []struct {
		Name        string
		Description string
		Category    string
	}{
		{"read_file", "Lê o conteúdo de um arquivo", "File System"},
		{"write_file", "Cria ou sobrescreve um arquivo", "File System"},
		{"list_dir", "Lista arquivos em um diretório", "File System"},
		{"edit_file", "Edita blocos específicos de um arquivo", "File System"},
		{"web_search", "Pesquisa na web", "Web"},
		{"web_fetch", "Busca conteúdo de uma URL", "Web"},
		{"message", "Envia mensagem para canais", "Communication"},
	}

	workspacePath := e.GetActiveWorkspace()
	enabledTools := make(map[string]bool)
	e.mu.RLock()
	for _, w := range e.adaCfg.Workspaces {
		if w.Path == workspacePath {
			for _, t := range w.Tools {
				enabledTools[t] = true
			}
			break
		}
	}
	e.mu.RUnlock()

	var list []ToolUIInfo
	for _, t := range available {
		list = append(list, ToolUIInfo{
			Name:        t.Name,
			Description: t.Description,
			Category:    t.Category,
			Enabled:     enabledTools[t.Name],
		})
	}
	return list
}

func (e *Engine) ToggleTool(toolName string, enabled bool) {
	workspacePath := e.GetActiveWorkspace()
	e.mu.Lock()
	for i, w := range e.adaCfg.Workspaces {
		if w.Path == workspacePath {
			found := false
			idx := -1
			for j, t := range w.Tools {
				if t == toolName {
					found = true
					idx = j
					break
				}
			}

			if enabled && !found {
				e.adaCfg.Workspaces[i].Tools = append(w.Tools, toolName)
			} else if !enabled && found {
				e.adaCfg.Workspaces[i].Tools = append(w.Tools[:idx], w.Tools[idx+1:]...)
			}
			break
		}
	}
	e.mu.Unlock()
	e.SaveAdaConfig()
}

func (e *Engine) GetActiveWorkspace() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.adaCfg.ActiveWorkspacePath
}

func (e *Engine) GetWorkspaceName(path string) string {
	if path == "" {
		return "Default"
	}
	return filepath.Base(path)
}

func (e *Engine) ToggleWorkspace(path string) {
	e.mu.Lock()
	for i, w := range e.adaCfg.Workspaces {
		if w.Path == path {
			e.adaCfg.Workspaces[i].Enabled = !e.adaCfg.Workspaces[i].Enabled
			break
		}
	}
	e.mu.Unlock()
	e.SaveAdaConfig()
}
