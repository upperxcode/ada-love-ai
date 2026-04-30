package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"bytes"
	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/sipeed/picoclaw/pkg/skills"
	"io"
	"net/http"
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

type AdaConfig struct {
	TinyBrain struct {
		ModelName string `json:"model_name"`
		Provider  string `json:"provider"`
	} `json:"tiny_brain"`
	Workspaces      []string      `json:"workspaces"`
	Knowledge       []string      `json:"knowledge"`
	Agents          []AgentConfig `json:"agents"`
	AgentCategories []string      `json:"agent_categories"`
	WorkspaceAgents []string      `json:"workspace_agents"`
	Skills          []string      `json:"skills"`
	Personality     string        `json:"personality"`
}

const (
	SummaryThreshold = 10 // Começa a sumarizar após 10 mensagens
	SummaryKeepLast  = 4  // Mantém as últimas 4 mensagens após sumarizar
)

type Engine struct {
	cfg        *config.Config
	msgBus     *bus.MessageBus
	agentLoop  *agent.AgentLoop
	mu         sync.Mutex
	adaCfg     AdaConfig
	SessionMgr *SessionManager
	skillReg   *skills.RegistryManager
}

func NewEngine() (*Engine, error) {
	// Carrega a configuração do arquivo config.json no diretório raiz do ada-love-ai
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar config: %w", err)
	}

	// Carrega configurações customizadas de um arquivo separado (ada_config.json)
	var adaCfg AdaConfig
	if data, err := os.ReadFile("ada_config.json"); err == nil {
		json.Unmarshal(data, &adaCfg)
		// Migração: se não tem agentes no workspace mas tem agentes globais, seleciona todos
		if len(adaCfg.WorkspaceAgents) == 0 && len(adaCfg.Agents) > 0 {
			for _, a := range adaCfg.Agents {
				adaCfg.WorkspaceAgents = append(adaCfg.WorkspaceAgents, a.Name)
			}
		}
	}

	// Cria o provider baseado na configuração
	provider, modelID, err := providers.CreateProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar provider: %w", err)
	}

	// Garante que o model_name padrão está sincronizado com o resolvido
	if modelID != "" {
		cfg.Agents.Defaults.ModelName = modelID
	}

	msgBus := bus.NewMessageBus()
	e := &Engine{
		cfg:        cfg,
		msgBus:     msgBus,
		adaCfg:     adaCfg,
		SessionMgr: NewSessionManager(),
		skillReg:   skills.NewRegistryManagerFromToolsConfig(cfg.Tools.Skills),
	}
	e.agentLoop = agent.NewAgentLoop(cfg, msgBus, provider)
	return e, nil
}

func (e *Engine) UpdateWorkspaceConfig(fn func(cfg *AdaConfig)) {
	e.mu.Lock()
	fn(&e.adaCfg)
	e.mu.Unlock()
	e.SaveAdaConfig()
	e.ReloadAgentLoop()
}

func (e *Engine) ReloadAgentLoop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Para o loop atual
	if e.agentLoop != nil {
		e.agentLoop.Stop()
	}

	// Recria o provider para garantir que modelos novos/alterados sejam pegos
	provider, modelID, err := providers.CreateProvider(e.cfg)
	if err != nil {
		return err
	}
	if modelID != "" {
		e.cfg.Agents.Defaults.ModelName = modelID
	}

	e.agentLoop = agent.NewAgentLoop(e.cfg, e.msgBus, provider)
	return nil
}

func (e *Engine) SendMessage(ctx context.Context, text string, sessionID string) (string, error) {
	// Picoclaw utiliza chaves de sessão para manter o histórico.
	sessionKey := "ada:default"
	if sessionID != "" {
		sessionKey = "ada:" + sessionID
	}

	// Prepara o prompt com memória de longo prazo se existir
	finalPrompt := text
	if sessionID != "" {
		if sess, ok := e.SessionMgr.sessions[sessionID]; ok && sess.Summary != "" {
			finalPrompt = fmt.Sprintf("MEMÓRIA DE LONGO PRAZO (Resumo de conversas anteriores):\n%s\n\nUSUÁRIO: %s", sess.Summary, text)
		}
	}

	return e.agentLoop.ProcessDirect(ctx, finalPrompt, sessionKey)
}

func (e *Engine) SubscribeEvents(handler func(agent.Event)) {
	sub := e.agentLoop.SubscribeEvents(0)
	go func() {
		for ev := range sub.C {
			handler(ev)
		}
	}()
}

func (e *Engine) TinyBrainModel() string {
	return e.adaCfg.TinyBrain.ModelName
}

func (e *Engine) GetAdaConfig() AdaConfig {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.adaCfg
}

func (e *Engine) SetAdaConfig(cfg AdaConfig) error {
	e.mu.Lock()
	e.adaCfg = cfg
	e.mu.Unlock()
	return e.SaveAdaConfig()
}

func (e *Engine) SaveAdaConfig() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	data, err := json.MarshalIndent(e.adaCfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("ada_config.json", data, 0644)
}

// SendTinyBrainMessage envia uma mensagem usando o modelo tiny_brain se configurado,
// caso contrário usa o modelo padrão. Faz a chamada direta ao provedor para ignorar
// restrições de validação do Picoclaw.
func (e *Engine) SendTinyBrainMessage(ctx context.Context, prompt string) (string, error) {
	if e.adaCfg.TinyBrain.ModelName == "" {
		return e.SendMessage(ctx, prompt, "")
	}

	// Tenta encontrar a URL base para o provider no model_list
	apiBase := ""
	apiKey := ""
	for _, m := range e.cfg.ModelList {
		if m.Provider == e.adaCfg.TinyBrain.Provider && m.APIBase != "" {
			apiBase = m.APIBase
			apiKey = m.APIKey()
			break
		}
	}

	// Fallback para localhost se for LM Studio/Ollama e não achou no config
	if apiBase == "" {
		switch e.adaCfg.TinyBrain.Provider {
		case "lmstudio":
			apiBase = "http://127.0.0.1:1234/v1"
		case "ollama":
			apiBase = "http://127.0.0.1:11434/v1"
		default:
			return "", fmt.Errorf("URL base não encontrada para o provider: %s", e.adaCfg.TinyBrain.Provider)
		}
	}

	// Endpoint OpenAI padrão
	url := fmt.Sprintf("%s/chat/completions", apiBase)

	requestBody := map[string]interface{}{
		"model": e.adaCfg.TinyBrain.ModelName,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("falha na requisição ao provedor: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("provedor retornou erro (%d): %s", resp.StatusCode, string(body))
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", err
	}

	if len(openAIResp.Choices) > 0 {
		return openAIResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("resposta vazia do provedor")
}

func (e *Engine) SummarizeSession(sessionID string) {
	e.mu.Lock()
	sess, ok := e.SessionMgr.sessions[sessionID]
	if !ok || len(sess.Messages) < SummaryThreshold {
		e.mu.Unlock()
		return
	}

	// Prepara mensagens para sumarizar (exceto as que vamos manter)
	toSummarize := sess.Messages[:len(sess.Messages)-SummaryKeepLast]
	var text strings.Builder
	for _, m := range toSummarize {
		text.WriteString(fmt.Sprintf("%s: %s\n", m.Role, m.Content))
	}
	oldSummary := sess.Summary
	e.mu.Unlock()

	go func() {
		ctx := context.Background()
		prompt := fmt.Sprintf(`Você é um assistente de memória. Seu trabalho é refinar e consolidar um resumo de conversa.
Abaixo está o resumo antigo (se houver) e as novas mensagens.
Crie um resumo final coeso, mantendo os pontos principais e o contexto técnico.

--- RESUMO ANTIGO ---
%s

--- NOVAS MENSAGENS ---
%s

Responda APENAS o novo resumo consolidado.`, oldSummary, text.String())

		newSummary, err := e.SendTinyBrainMessage(ctx, prompt)
		if err != nil {
			fmt.Printf("[Engine] Erro ao sumarizar: %v\n", err)
			return
		}

		if newSummary != "" {
			e.SessionMgr.SetSummary(sessionID, newSummary)
			e.SessionMgr.ClearMessages(sessionID, SummaryKeepLast)
			fmt.Printf("[Engine] Sessão %s sumarizada com sucesso.\n", sessionID)
		}
	}()
}

func (e *Engine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.agentLoop != nil {
		e.agentLoop.Stop()
		e.agentLoop.Close()
	}
	if e.msgBus != nil {
		e.msgBus.Close()
	}
}

func (e *Engine) GetModelList() config.SecureModelList {
	return e.cfg.ModelList
}
func (e *Engine) SearchSkills(ctx context.Context, query string) ([]skills.SearchResult, error) {
	if e.skillReg == nil {
		return nil, fmt.Errorf("skill registry manager not initialized")
	}
	return e.skillReg.SearchAll(ctx, query, 20)
}

func (e *Engine) InstallSkill(ctx context.Context, registryName, slug, version string) error {
	reg := e.skillReg.GetRegistry(registryName)
	if reg == nil {
		return fmt.Errorf("registry %s not found", registryName)
	}

	// Define o diretório de destino (workspace atual)
	workspace := e.cfg.Agents.Defaults.Workspace
	targetDir := strings.Join([]string{workspace, "skills"}, "/")

	// Resolve o nome do diretório da skill
	dirName, err := reg.ResolveInstallDirName(slug)
	if err != nil {
		return err
	}
	fullTargetDir := strings.Join([]string{targetDir, dirName}, "/")

	_, err = reg.DownloadAndInstall(ctx, slug, version, fullTargetDir)
	return err
}

func (e *Engine) GetInstalledSkills() ([]string, error) {
	workspace := e.cfg.Agents.Defaults.Workspace
	targetDir := strings.Join([]string{workspace, "skills"}, "/")

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var skillsList []string
	for _, entry := range entries {
		if entry.IsDir() {
			skillsList = append(skillsList, entry.Name())
		}
	}
	return skillsList, nil
}

func (e *Engine) UninstallSkill(name string) error {
	workspace := e.cfg.Agents.Defaults.Workspace
	targetDir := strings.Join([]string{workspace, "skills", name}, "/")
	return os.RemoveAll(targetDir)
}

func (e *Engine) GetSkillDetails(name string) (string, error) {
	workspace := e.cfg.Agents.Defaults.Workspace
	filePath := strings.Join([]string{workspace, "skills", name, "SKILL.md"}, "/")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (e *Engine) GetSkillFullInfo(name string) (*SkillFullInfo, error) {
	workspace := e.cfg.Agents.Defaults.Workspace
	skillDir := strings.Join([]string{workspace, "skills", name}, "/")

	info := &SkillFullInfo{
		Name:     name,
		Registry: "local",
		Version:  "0.0.1",
	}

	// 1. Tentar ler SKILL.md
	mdPath := skillDir + "/SKILL.md"
	if data, err := os.ReadFile(mdPath); err == nil {
		info.Markdown = string(data)
		info.Raw = info.Markdown
		info.CharCount = len(data)
		info.LineCount = strings.Count(info.Raw, "\n") + 1
	}

	// 2. Tentar ler SKILL.json (manifesto oficial do Picoclaw)
	jsonPath := skillDir + "/SKILL.json"
	if data, err := os.ReadFile(jsonPath); err == nil {
		var manifest struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Version     string `json:"version"`
			Repository  string `json:"repository"`
		}
		if err := json.Unmarshal(data, &manifest); err == nil {
			if manifest.Description != "" {
				info.Description = manifest.Description
			}
			if manifest.Version != "" {
				info.Version = manifest.Version
			}
			if manifest.Repository != "" {
				info.URL = manifest.Repository
			}
		}
	}

	// 3. Como fallback para a descrição, tentar extrair do MD se ainda estiver vazia
	if info.Description == "" && info.Markdown != "" {
		lines := strings.Split(info.Markdown, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" && !strings.HasPrefix(l, "#") {
				info.Description = l
				break
			}
		}
	}

	// 3. Simulação de metadados técnicos (em uma implementação real, leríamos o manifest da skill)
	info.URL = "https://github.com/sipeed/picoclaw/tree/main/skills/" + name
	
	return info, nil
}
