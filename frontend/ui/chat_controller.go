package ui

import (
	"ada-love-ai/backend"
	"ada-love-ai/frontend/components"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type msgHistory struct {
	Role string
	Text string
}

type ChatController struct {
	engine  *backend.Engine
	msgList *fyne.Container
	scroll  *container.Scroll
	header  *components.ChatHeader
	details *components.DetailsPanel

	currentAIMessage *components.ChatMessage
	currentAIText    string
	thoughtPanel     *components.ThoughtPanel

	activeAgent *backend.AgentConfig

	// mu protege o estado de streaming para evitar repetições
	mu           sync.Mutex
	isStreaming  bool
	internalMode bool
}

func NewChatController(engine *backend.Engine, msgList *fyne.Container, scroll *container.Scroll, header *components.ChatHeader, details *components.DetailsPanel) *ChatController {
	ctrl := &ChatController{
		engine:       engine,
		msgList:      msgList,
		scroll:       scroll,
		header:       header,
		details:      details,
		thoughtPanel: components.NewThoughtPanel(),
	}

	// Adiciona o thoughtPanel no topo (ele inicia escondido)
	msgList.Add(ctrl.thoughtPanel.CanvasObject)

	// Configura busca no painel de detalhes (aba Conversas)
	if details.SearchEntry != nil {
		details.SearchEntry.OnChanged = func(q string) {
			ctrl.refreshSidebar(q)
		}
	}

	// Define o nome do workspace inicial
	cfg := engine.GetAdaConfig()
	if len(cfg.Workspaces) > 0 && cfg.ActiveWorkspaceIndex < len(cfg.Workspaces) {
		wsName := cfg.Workspaces[cfg.ActiveWorkspaceIndex].Title
		ctrl.header.SetWorkspaceName(wsName)
		ctrl.details.SetWorkspaceName(wsName)
	}

	// Cria a sessão inicial e atualiza a barra lateral de forma assíncrona
	go func() {
		// Pequeno delay para garantir que a janela principal foi criada
		time.Sleep(200 * time.Millisecond)
		fmt.Println("[ChatController] Inicializando sessão no startup...")
		ctrl.InitializeStartupSession()
		// Inscreve-se nos eventos do Picoclaw para streaming
		engine.SubscribeEvents(ctrl.handleEvent)
		fmt.Println("[ChatController] Pronto.")
	}()

	return ctrl
}

func (c *ChatController) SetWorkspaceName(name string) {
	c.header.SetWorkspaceName(name)
	c.details.SetWorkspaceName(name)
}

func (c *ChatController) handleEvent(ev backend.Event) {
	c.mu.Lock()
	isInternal := c.internalMode
	activeID := c.engine.SessionMgr.GetActiveID()
	c.mu.Unlock()

	// Se o evento tem um SessionID e não é o ativo, ignoramos (evita leaks de summarização/título)
	if ev.SessionID != "" && ev.SessionID != activeID {
		return
	}

	// Se estivermos em modo interno (gerando título na sessão ativa), ignoramos atualizações de UI
	if isInternal && ev.Kind != backend.EventKindTurnEnd {
		return
	}

	switch ev.Kind {
	case backend.EventKindLLMDelta:
		if p, ok := ev.Payload.(backend.StreamingDeltaPayload); ok {
			c.mu.Lock()
			delta := p.Content
			if !c.isStreaming {
				c.isStreaming = true
				c.currentAIText = delta
				fyne.Do(func() {
					// Quando começa o streaming da resposta, o pensamento "concluiu"
					c.thoughtPanel.Hide()

					c.currentAIMessage = components.NewChatMessage(c.currentAIText, false)
					c.msgList.Add(c.currentAIMessage.CanvasObject)
					c.msgList.Refresh()
					c.scroll.ScrollToBottom()
				})
			} else {
				c.currentAIText += delta
				textToUpdate := c.currentAIText
				fyne.Do(func() {
					if c.currentAIMessage != nil {
						c.currentAIMessage.UpdateText(textToUpdate)
						c.scroll.ScrollToBottom()
					}
				})
			}
			c.mu.Unlock()
		}
	case backend.EventKindStatus:
		if p, ok := ev.Payload.(backend.StatusPayload); ok {
			msg := p.Message
			fyne.Do(func() {
				c.header.SetStatus(msg)
				c.thoughtPanel.AddStep(msg)
				c.scroll.ScrollToBottom()
				
				go func() {
					time.Sleep(5 * time.Second)
					fyne.Do(func() {
						c.header.SetStatus("")
					})
				}()
			})
		}
	case backend.EventKindTurnStart:
		if !isInternal {
			fyne.Do(func() {
				c.thoughtPanel.Clear()
				// Remove e re-adiciona para garantir que está no final
				c.msgList.Remove(c.thoughtPanel.CanvasObject)
				c.msgList.Add(c.thoughtPanel.CanvasObject)
				
				c.header.SetStatus("Analisando tarefa...")
				c.thoughtPanel.AddStep("Iniciando análise da solicitação...")
				c.msgList.Refresh()
				c.scroll.ScrollToBottom()
			})
		}
	case backend.EventKindToolExecStart:
		if p, ok := ev.Payload.(backend.ToolExecStartPayload); ok {
			fyne.Do(func() {
				statusMsg := fmt.Sprintf("Executando %s...", p.Tool)
				c.header.SetStatus(statusMsg)
				c.thoughtPanel.AddStep(fmt.Sprintf("🔨 Chamando ferramenta: %s", p.Tool))
				c.scroll.ScrollToBottom()
			})
		}
	case backend.EventKindTurnEnd:
		c.mu.Lock()
		c.finalizeAIMessage()
		c.mu.Unlock()
		fyne.Do(func() {
			c.header.SetStatus("")
			c.thoughtPanel.Hide() // Garante que some ao final
		})
	case backend.EventKindError:
		if p, ok := ev.Payload.(backend.ErrorPayload); ok {
			fyne.Do(func() {
				c.header.SetStatus(fmt.Sprintf("Erro: %s", p.Message))
				c.thoughtPanel.AddStep(fmt.Sprintf("❌ Erro: %s", p.Message))
			})
		}
	}
}

func (c *ChatController) SendMessage(text string) {
	if text == "" {
		return
	}

	c.mu.Lock()
	// Só adicionamos e mostramos na UI se não for modo interno
	if !c.internalMode {
		fyne.Do(func() {
			userMsg := components.NewChatMessage(text, true)
			c.msgList.Add(userMsg.CanvasObject)
			c.msgList.Refresh()
			c.scroll.ScrollToBottom()
			go func() {
				time.Sleep(100 * time.Millisecond)
				fyne.Do(func() {
					c.scroll.ScrollToBottom()
				})
			}()
		})
	}
	c.mu.Unlock()

	// Inicia o processamento no backend
	go func() {
		sess := c.engine.SessionMgr.GetActiveSession()
		sessionID := ""
		if sess != nil {
			sessionID = sess.ID
		}

		finalText := text
		c.mu.Lock()
		if c.activeAgent != nil && c.activeAgent.Persona != "" {
			// Adicionamos a persona como uma instrução de sistema se for o início ou reforço
			finalText = fmt.Sprintf("### INSTRUÇÃO DE PERSONA (%s):\n%s\n\n### MENSAGEM DO USUÁRIO:\n%s",
				c.activeAgent.Name, c.activeAgent.Persona, text)
		}
		c.mu.Unlock()

		_, err := c.engine.SendMessage(context.Background(), finalText, sessionID)
		if err != nil {
			fmt.Printf("Erro ao enviar mensagem: %v\n", err)
		}
	}()
}

func (c *ChatController) SetAgent(agent *backend.AgentConfig) {
	c.mu.Lock()
	c.activeAgent = agent
	c.mu.Unlock()

	fyne.Do(func() {
		if agent != nil {
			c.header.SetAgentInfo(agent.Name, agent.Icon)
		} else {
			c.header.SetAgentInfo("", "")
		}
	})
}

func (c *ChatController) InitializeStartupSession() {
	activeID := c.engine.SessionMgr.GetActiveID()
	if activeID != "" {
		c.LoadSession(activeID)
		return
	}

	// Se não tiver ativa, tenta pegar a primeira do workspace
	sessions := c.engine.SessionMgr.ListSessions(c.getActiveWorkspaceID())
	if len(sessions) > 0 {
		c.LoadSession(sessions[0].ID)
		return
	}

	// Se não tiver nada, cria nova
	c.NewChat()
}

func (c *ChatController) getActiveWorkspaceID() string {
	cfg := c.engine.GetAdaConfig()
	if len(cfg.Workspaces) > 0 && cfg.ActiveWorkspaceIndex >= 0 && cfg.ActiveWorkspaceIndex < len(cfg.Workspaces) {
		return cfg.Workspaces[cfg.ActiveWorkspaceIndex].Path
	}
	return "default"
}

func (c *ChatController) NewChat() {
	c.mu.Lock()
	fyne.Do(func() {
		c.msgList.Objects = nil
		c.msgList.Refresh()
		c.header.SetTitle("Nova Conversa")
	})
	c.mu.Unlock()

	c.engine.SessionMgr.CreateSession("Nova Conversa", c.getActiveWorkspaceID())
	c.refreshSidebar("")
}

func (c *ChatController) LoadSession(id string) {
	fmt.Printf("[ChatController] LoadSession iniciado para ID: %s\n", id)
	session := c.engine.SessionMgr.GetSession(id)
	if session == nil {
		fmt.Printf("[ChatController] Sessão %s não encontrada!\n", id)
		return
	}

	c.mu.Lock()
	c.engine.SessionMgr.SetActive(id)

	fyne.Do(func() {
		c.msgList.Objects = nil
		for _, m := range session.Messages {
			isUser := m.Role == "user"
			msg := components.NewChatMessage(m.Content, isUser)
			c.msgList.Add(msg.CanvasObject)
		}
		c.header.SetTitle(session.Title)
		c.msgList.Refresh()
		
		// Scroll imediato e um segundo scroll após o layout assentar
		c.scroll.ScrollToBottom()
		go func() {
			time.Sleep(100 * time.Millisecond)
			fyne.Do(func() {
				c.scroll.ScrollToBottom()
			})
			time.Sleep(200 * time.Millisecond)
			fyne.Do(func() {
				c.scroll.ScrollToBottom()
			})
		}()
	})
	c.mu.Unlock()

	c.refreshSidebar("")
}

func (c *ChatController) refreshSidebar(query string) {
	activeID := c.engine.SessionMgr.GetActiveID()
	fmt.Printf("[ChatController] refreshSidebar: ActiveID=%s, Query='%s'\n", activeID, query)
	sessions := c.engine.SessionMgr.SearchSessions(query, c.getActiveWorkspaceID())
	fyne.Do(func() {
		c.details.UpdateSessions(sessions, activeID)
	})
}

func (c *ChatController) PinSession(id string) {
	c.engine.TogglePin(id)
	c.refreshSidebar(c.details.SearchEntry.Text)
}

func (c *ChatController) DeleteSession(id string) {
	c.engine.DeleteSession(id)

	// Se era a sessão ativa, cria uma nova
	active := c.engine.SessionMgr.GetActiveSession()
	if active == nil {
		c.NewChat()
	}

	c.refreshSidebar(c.details.SearchEntry.Text)
}

func (c *ChatController) RenameSession(id string, newTitle string) {
	c.engine.RenameSession(id, newTitle)

	// Se for a ativa, atualiza o header
	active := c.engine.SessionMgr.GetActiveSession()
	if active != nil && active.ID == id {
		c.header.SetTitle(newTitle)
	}

	c.refreshSidebar(c.details.SearchEntry.Text)
}

func (c *ChatController) suggestTitle() {
	time.Sleep(500 * time.Millisecond)

	c.mu.Lock()
	if c.internalMode {
		c.mu.Unlock()
		return
	}
	c.internalMode = true

	var contextStr strings.Builder
	contextStr.WriteString("CONTEXTO DA CONVERSA:\n")

	sess := c.engine.SessionMgr.GetActiveSession()
	if sess != nil {
		for _, m := range sess.Messages {
			contextStr.WriteString(fmt.Sprintf("%s: %s\n", m.Role, m.Content))
		}
	}
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.internalMode = false
		c.mu.Unlock()
	}()

	prompt := fmt.Sprintf("%s\n\n### INSTRUÇÃO CRÍTICA: Baseado no contexto acima, sugira um título curtíssimo (máximo 4 palavras) para nossa conversa. Responda APENAS o título. NÃO use ferramentas. NÃO dê explicações. NÃO use aspas.", contextStr.String())

	resp, err := c.engine.SendTinyBrainMessage(context.Background(), prompt)
	if err != nil {
		fmt.Printf("[ChatController] Erro ao sugerir título: %v\n", err)
		return
	}

	if resp != "" {
		resp = strings.Trim(resp, "\"'. ")
		fyne.Do(func() {
			c.header.SetTitle(resp)
		})

		sess := c.engine.SessionMgr.GetActiveSession()
		if sess != nil {
			c.engine.RenameSession(sess.ID, resp)
		}
	}
}

// finalizeAIMessage assume que o chamador já possui o lock do Mutex
func (c *ChatController) finalizeAIMessage() {
	isInternal := c.internalMode
	fullText := c.currentAIText

	if !isInternal && fullText != "" {
		sess := c.engine.SessionMgr.GetActiveSession()
		if sess != nil {
			// Não adicionamos mais aqui, pois o Engine.SendMessage já adiciona.
			// Apenas mantemos a lógica de gatilhos baseada no estado atual.
			count := len(sess.Messages)

			// Verifica se precisa sugerir título
			if count >= 6 && count <= 10 && sess.Title == "Nova Conversa" {
				go c.suggestTitle()
			}

			c.refreshSidebar("")
		}
	}

	c.currentAIMessage = nil
	c.currentAIText = ""
	c.isStreaming = false
}
