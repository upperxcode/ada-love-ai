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

func (e *Engine) SetActiveWorkspace(title string) {
	e.mu.Lock()
	e.adaCfg.ActiveWorkspacePath = title
	
	for i, w := range e.adaCfg.Workspaces {
		if w.Path == title || w.Title == title {
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
		Payload: title,
		Time:    time.Now(),
	})
}

func (e *Engine) DeleteWorkspace(title string) {
	e.mu.Lock()
	var newList []WorkspaceConfig
	for _, w := range e.adaCfg.Workspaces {
		if w.Title == title {
			continue
		}
		newList = append(newList, w)
	}
	e.adaCfg.Workspaces = newList
	if e.adaCfg.ActiveWorkspacePath == title {
		if len(newList) > 0 {
			e.adaCfg.ActiveWorkspacePath = newList[0].Title
		} else {
			e.adaCfg.ActiveWorkspacePath = ""
		}
	}
	e.mu.Unlock()
	e.SaveAdaConfig()
	
	e.eventBus.Emit(Event{
		Kind:    EventKindWorkspaceChanged,
		Payload: e.adaCfg.ActiveWorkspacePath,
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
	for i, w := range e.adaCfg.Workspaces {
		if w.Title == originalTitle {
			e.adaCfg.Workspaces[i] = ws
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

