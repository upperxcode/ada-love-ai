package backend

import (
	"ada-love-ai/pkg/logger"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func (e *Engine) ListWorkspaces() []WorkspaceConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	fmt.Printf("[Engine] ListWorkspaces: returning %d workspaces\n", len(e.adaCfg.Workspaces))
	for i, ws := range e.adaCfg.Workspaces {
		fmt.Printf("[Engine]   [%d] title=%q path=%q enabled=%v workers=%d\n",
			i, ws.Title, ws.Path, ws.Enabled, len(ws.WorkerNames))
	}
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

	// Se path não foi definido, usar o título normalizado como path
	if w.Path == "" {
		w.Path = strings.ToLower(strings.ReplaceAll(w.Title, " ", "_"))
	}

	e.adaCfg.Workspaces = append(e.adaCfg.Workspaces, w)
	e.mu.Unlock()
	return e.SaveAdaConfig()
}

func (e *Engine) SetActiveWorkspace(titleOrPath string) {
	e.mu.Lock()
	// Find the workspace by title or path
	for i, w := range e.adaCfg.Workspaces {
		if w.Path == titleOrPath || w.Title == titleOrPath {
			// Set ActiveWorkspacePath to the PATH, not the title
			e.adaCfg.ActiveWorkspacePath = w.Path
			e.adaCfg.ActiveWorkspaceIndex = i
			break
		}
	}
	e.mu.Unlock()

	e.SaveAdaConfig()
	e.ReloadAgentLoop()
	e.RefreshSessions()

	e.eventBus.Emit(Event{
		Kind:    EventKindWorkspaceChanged,
		Payload: titleOrPath,
		Time:    time.Now(),
	})
}

func (e *Engine) DeleteWorkspace(titleOrPath string) {
	e.mu.Lock()
	var newList []WorkspaceConfig
	for _, w := range e.adaCfg.Workspaces {
		// Aceita tanto title quanto path para compatibilidade
		if w.Title == titleOrPath || w.Path == titleOrPath {
			continue
		}
		newList = append(newList, w)
	}
	e.adaCfg.Workspaces = newList
	if e.adaCfg.ActiveWorkspacePath == titleOrPath {
		if len(newList) > 0 {
			e.adaCfg.ActiveWorkspacePath = newList[0].Title
		} else {
			e.adaCfg.ActiveWorkspacePath = ""
		}
	}
	e.mu.Unlock()
	e.SaveAdaConfig()

	e.eventBus.Emit(Event{
		Kind:    EventKindWorkspaceDeleted,
		Payload: titleOrPath,
		Time:    time.Now(),
	})
}

func (e *Engine) RegisterWorkspaceTools(title string) {}

func (e *Engine) GetAvailableTools() []ToolUIInfo {
	available := []struct {
		Name        string
		Description string
		Category    string
	}{
		// File System
		{"read_file", "Lê o conteúdo de um arquivo", "File System"},
		{"write_file", "Cria ou sobrescreve um arquivo", "File System"},
		{"list_dir", "Lista arquivos em um diretório", "File System"},
		{"edit_file", "Edita blocos específicos de um arquivo", "File System"},
		{"append_file", "Adiciona conteúdo ao final de um arquivo", "File System"},
		{"send_file", "Envia um arquivo para o agente", "File System"},
		{"load_image", "Carrega e analisa uma imagem", "File System"},
		{"view_file_outline", "Extrai a estrutura de um arquivo", "File System"},
		{"grep_code", "Busca padrões em arquivos", "Code Search"},
		{"find_files", "Localiza arquivos por nome", "Code Search"},
		{"list_dir_tree", "Lista diretório em árvore", "File System"},

		// Git
		{"git_status", "Mostra o status do repositório git", "Git"},
		{"git_diff", "Mostra as diferenças não commitadas", "Git"},
		{"git_log", "Mostra o histórico de commits", "Git"},
		{"git_commit", "Cria um commit com as mudanças staged", "Git"},
		{"git_push", "Envia commits para o repositório remoto", "Git"},
		{"git_pull", "Atualiza o repositório local", "Git"},
		{"git_clone", "Clona um repositório", "Git"},

		// Web
		{"web_search", "Pesquisa na web", "Web"},
		{"web_fetch", "Busca conteúdo de uma URL", "Web"},
		{"http_request", "Envia requisições HTTP (GET, POST, PUT, DELETE)", "Web"},

		// Media
		{"media_cleanup", "Limpa arquivos de mídia temporários", "Media"},

		// MCP (Model Context Protocol)
		{"mcp", "Ferramentas via MCP (Model Context Protocol)", "MCP"},

		// Communication
		{"message", "Envia mensagem para canais", "Communication"},
		{"reaction", "Reage a eventos/mensagens", "Communication"},

		// Testing
		{"run_tests", "Executa os testes do projeto", "Testing"},

		// Build
		{"build_project", "Compila o projeto", "Build"},
		{"install_deps", "Instala dependências do projeto", "Build"},
		{"lint_code", "Executa linter no projeto", "Build"},
		{"code_metrics", "Analisa métricas do código", "Build"},

		// Exec & Tasks (orquestrador)
		{"exec", "Executa comandos shell no sistema", "Shell"},
		{"cron", "Agenda lembretes, tarefas ou comandos", "Scheduled Tasks"},

		// Memory & Knowledge
		{"tool_save_memory", "Salva informações na memória de longo prazo", "Memory"},
		{"get_agent_memory", "Recupera memórias salvas anteriormente", "Memory"},
		{"search_knowledge_base", "Busca na base de conhecimento local", "Knowledge"},

		// Hardware
		{"i2c", "Comunicação I2C com dispositivos", "Hardware"},
		{"spi", "Comunicação SPI com dispositivos", "Hardware"},

		// Skills
		{"find_skills", "Busca skills disponíveis", "Skills"},
		{"install_skill", "Instala uma skill", "Skills"},

		// Agent
		{"spawn", "Cria um sub-agente para tarefa específica", "Agent"},
		{"spawn_status", "Verifica o status de um sub-agente", "Agent"},
		{"subagent", "Gerencia sub-agentes", "Agent"},
		{"send_tts", "Envia texto para síntese de voz", "Agent"},
	}

	workspacePath := e.GetActiveWorkspace()
	enabledTools := make(map[string]bool)
	e.mu.RLock()
	for _, w := range e.adaCfg.Workspaces {
		if w.Title == workspacePath {
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
	activeTitle := e.GetActiveWorkspace()
	e.mu.Lock()
	for i, w := range e.adaCfg.Workspaces {
		if w.Title == activeTitle {
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

func (e *Engine) ToggleWorkspace(title string) {
	e.mu.Lock()
	for i, w := range e.adaCfg.Workspaces {
		if w.Title == title {
			e.adaCfg.Workspaces[i].Enabled = !e.adaCfg.Workspaces[i].Enabled
			break
		}
	}
	e.mu.Unlock()
	e.SaveAdaConfig()
}

func (e *Engine) UpdateWorkspace(originalTitle string, ws WorkspaceConfig) {
	e.mu.Lock()
	fmt.Printf("[Engine] UpdateWorkspace: originalTitle=%q newWorkers=%d newTitle=%q\n", originalTitle, len(ws.WorkerNames), ws.Title)
	for i, w := range e.adaCfg.Workspaces {
		if w.Title == originalTitle {
			fmt.Printf("[Engine] UpdateWorkspace: found at index %d, replacing with %d worker names\n", i, len(ws.WorkerNames))
			e.adaCfg.Workspaces[i] = ws
			break
		}
	}
	e.mu.Unlock()
	e.SaveAdaConfig()
}

// AddToolToWorkspace adds a tool to the active workspace.
// Returns true if the tool was added.
func (e *Engine) AddToolToWorkspace(workspaceTitle, toolName string) bool {
	logger.DebugCF("workspace", "AddToolToWorkspace called",
		map[string]any{"workspace": workspaceTitle, "tool": toolName})
	e.mu.Lock()
	for i, w := range e.adaCfg.Workspaces {
		if w.Title == workspaceTitle {
			for _, t := range w.Tools {
				if t == toolName {
					logger.DebugCF("workspace", "Tool already exists in workspace",
						map[string]any{"workspace": workspaceTitle, "tool": toolName})
					e.mu.Unlock()
					return false
				}
			}
			e.adaCfg.Workspaces[i].Tools = append(w.Tools, toolName)
			e.mu.Unlock()
			e.SaveAdaConfig()
			logger.DebugCF("workspace", "Tool added to workspace",
				map[string]any{"workspace": workspaceTitle, "tool": toolName})
			return true
		}
	}
	logger.DebugCF("workspace", "Workspace not found",
		map[string]any{"workspace": workspaceTitle})
	e.mu.Unlock()
	return false
}

// RemoveToolFromWorkspace removes a tool from the active workspace.
// Returns true if the tool was removed.
func (e *Engine) RemoveToolFromWorkspace(workspaceTitle, toolName string) bool {
	e.mu.Lock()
	for i, w := range e.adaCfg.Workspaces {
		if w.Title == workspaceTitle {
			newTools := make([]string, 0, len(w.Tools))
			removed := false
			for _, t := range w.Tools {
				if t != toolName {
					newTools = append(newTools, t)
				} else {
					removed = true
				}
			}
			e.adaCfg.Workspaces[i].Tools = newTools
			e.mu.Unlock()
			e.SaveAdaConfig()
			return removed
		}
	}
	e.mu.Unlock()
	return false
}

// SetWorkspaceTools replaces all tools for a workspace.
func (e *Engine) SetWorkspaceTools(workspaceTitle string, toolNames []string) {
	e.mu.Lock()
	for i, w := range e.adaCfg.Workspaces {
		if w.Title == workspaceTitle {
			e.adaCfg.Workspaces[i].Tools = toolNames
			break
		}
	}
	e.mu.Unlock()
	e.SaveAdaConfig()
}

func (e *Engine) GetToolProfiles() []ToolProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.adaCfg.ToolProfiles
}

func (e *Engine) CreateToolProfile(name, color, icon string) ToolProfile {
	e.mu.Lock()
	id := int64(1)
	for _, p := range e.adaCfg.ToolProfiles {
		if p.ID >= id {
			id = p.ID + 1
		}
	}
	profile := ToolProfile{
		ID:    id,
		Name:  name,
		Color: color,
		Icon:  icon,
		Tools: []string{},
	}
	e.adaCfg.ToolProfiles = append(e.adaCfg.ToolProfiles, profile)
	e.mu.Unlock()
	e.SaveAdaConfig()
	return profile
}

func (e *Engine) DeleteToolProfile(id int64) bool {
	e.mu.Lock()
	for i, p := range e.adaCfg.ToolProfiles {
		if p.ID == id {
			e.adaCfg.ToolProfiles = append(e.adaCfg.ToolProfiles[:i], e.adaCfg.ToolProfiles[i+1:]...)
			e.mu.Unlock()
			e.SaveAdaConfig()
			return true
		}
	}
	e.mu.Unlock()
	return false
}

func (e *Engine) ToggleProfileTool(profileID int64, toolName string, enabled bool) bool {
	e.mu.Lock()
	for i, p := range e.adaCfg.ToolProfiles {
		if p.ID == profileID {
			if enabled {
				found := false
				for _, t := range p.Tools {
					if t == toolName {
						found = true
						break
					}
				}
				if !found {
					e.adaCfg.ToolProfiles[i].Tools = append(p.Tools, toolName)
				}
			} else {
				newTools := make([]string, 0, len(p.Tools))
				for _, t := range p.Tools {
					if t != toolName {
						newTools = append(newTools, t)
					}
				}
				e.adaCfg.ToolProfiles[i].Tools = newTools
			}
			e.mu.Unlock()
			e.SaveAdaConfig()
			return true
		}
	}
	e.mu.Unlock()
	return false
}

// GetToolProfile returns a tool profile by ID, or nil if not found.
func (e *Engine) GetToolProfile(id int64) *ToolProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for i := range e.adaCfg.ToolProfiles {
		if e.adaCfg.ToolProfiles[i].ID == id {
			return &e.adaCfg.ToolProfiles[i]
		}
	}
	return nil
}

// AddToolsToProfile adds multiple tools to a profile at once.
// Tools that are already in the profile are skipped.
// Returns true if the profile was found and updated.
func (e *Engine) AddToolsToProfile(profileID int64, toolNames []string) bool {
	e.mu.Lock()
	for i, p := range e.adaCfg.ToolProfiles {
		if p.ID == profileID {
			existing := make(map[string]bool)
			for _, t := range p.Tools {
				existing[t] = true
			}
			added := 0
			for _, toolName := range toolNames {
				if !existing[toolName] {
					e.adaCfg.ToolProfiles[i].Tools = append(e.adaCfg.ToolProfiles[i].Tools, toolName)
					added++
				}
			}
			e.mu.Unlock()
			e.SaveAdaConfig()
			return added > 0
		}
	}
	e.mu.Unlock()
	return false
}

// RemoveToolsFromProfile removes multiple tools from a profile at once.
// Returns true if the profile was found and updated.
func (e *Engine) RemoveToolsFromProfile(profileID int64, toolNames []string) bool {
	e.mu.Lock()
	for i, p := range e.adaCfg.ToolProfiles {
		if p.ID == profileID {
			removed := 0
			for _, toolName := range toolNames {
				newTools := make([]string, 0, len(p.Tools))
				for _, t := range p.Tools {
					if t != toolName {
						newTools = append(newTools, t)
					}
				}
				if len(newTools) != len(p.Tools) {
					removed++
				}
				e.adaCfg.ToolProfiles[i].Tools = newTools
			}
			e.mu.Unlock()
			e.SaveAdaConfig()
			return removed > 0
		}
	}
	e.mu.Unlock()
	return false
}
