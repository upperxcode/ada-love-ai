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
	"github.com/sipeed/picoclaw/pkg/agent"
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

	activeAgent *backend.AgentConfig

	// mu protege o estado de streaming para evitar repetições
	mu           sync.Mutex
	isStreaming  bool
	internalMode bool
}

func NewChatController(engine *backend.Engine, msgList *fyne.Container, scroll *container.Scroll, header *components.ChatHeader, details *components.DetailsPanel) *ChatController {
	ctrl := &ChatController{
		engine:  engine,
		msgList: msgList,
		scroll:  scroll,
		header:  header,
		details: details,
	}

	// Configura busca no painel de detalhes (aba Conversas)
	if details.SearchEntry != nil {
		details.SearchEntry.OnChanged = func(q string) {
			ctrl.refreshSidebar(q)
		}
	}

	// Cria a sessão inicial
	ctrl.NewChat()

	// Inscreve-se nos eventos do Picoclaw para streaming
	engine.SubscribeEvents(ctrl.handleEvent)

	return ctrl
}

func (c *ChatController) handleEvent(ev agent.Event) {
	c.mu.Lock()
	isInternal := c.internalMode
	c.mu.Unlock()

	// Se estivermos em modo interno (gerando título), ignoramos atualizações de UI
	if isInternal && ev.Kind != agent.EventKindTurnEnd {
		return
	}

	switch ev.Kind {
	case agent.EventKindLLMDelta:
		if payload, ok := ev.Payload.(agent.LLMDeltaPayload); ok {
			c.mu.Lock()
			delta := payload.Content
			if !c.isStreaming {
				c.isStreaming = true
				c.currentAIText = delta
				fyne.Do(func() {
					c.currentAIMessage = components.NewChatMessage(c.currentAIText, false)
					c.msgList.Add(c.currentAIMessage.CanvasObject)
					c.scroll.ScrollToBottom()
				})
			} else {
				if len(delta) > len(c.currentAIText) && strings.HasPrefix(delta, c.currentAIText) {
					c.currentAIText = delta
				} else if delta != "" && !strings.HasSuffix(c.currentAIText, delta) {
					c.currentAIText += delta
				}

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
	case agent.EventKindTurnEnd:
		c.mu.Lock()
		c.finalizeAIMessage()
		c.mu.Unlock()
	}
}

func (c *ChatController) SendMessage(text string) {
	if text == "" {
		return
	}

	c.mu.Lock()
	// Só adicionamos e mostramos na UI se não for modo interno
	if !c.internalMode {
		sess := c.engine.SessionMgr.GetActiveSession()
		if sess != nil {
			c.engine.SessionMgr.AddMessage(sess.ID, "user", text)
			fyne.Do(func() {
				userMsg := components.NewChatMessage(text, true)
				c.msgList.Add(userMsg.CanvasObject)
				c.scroll.ScrollToBottom()
			})
		}
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

func (c *ChatController) NewChat() {
	c.mu.Lock()
	fyne.Do(func() {
		c.msgList.Objects = nil
		c.msgList.Refresh()
		c.header.SetTitle("Nova Conversa")
	})
	c.mu.Unlock()

	c.engine.SessionMgr.CreateSession("Nova Conversa")
	c.refreshSidebar("")
}

func (c *ChatController) LoadSession(id string) {
	session := c.engine.SessionMgr.GetSession(id)
	if session == nil {
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
		c.scroll.ScrollToBottom()
	})
	c.mu.Unlock()
}

func (c *ChatController) refreshSidebar(query string) {
	sessions := c.engine.SessionMgr.SearchSessions(query)
	fyne.Do(func() {
		c.details.UpdateSessions(sessions)
	})
}

func (c *ChatController) PinSession(id string) {
	c.engine.SessionMgr.TogglePin(id)
	c.refreshSidebar(c.details.SearchEntry.Text)
}

func (c *ChatController) DeleteSession(id string) {
	c.engine.SessionMgr.DeleteSession(id)

	// Se era a sessão ativa, cria uma nova
	active := c.engine.SessionMgr.GetActiveSession()
	if active == nil {
		c.NewChat()
	}

	c.refreshSidebar(c.details.SearchEntry.Text)
}

func (c *ChatController) RenameSession(id string, newTitle string) {
	c.engine.SessionMgr.RenameSession(id, newTitle)

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
			sess.Title = resp
			c.refreshSidebar("")
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
			count, _ := c.engine.SessionMgr.AddMessage(sess.ID, "assistant", fullText)

			// Verifica se precisa sugerir título
			if count >= 6 && count <= 10 && sess.Title == "Nova Conversa" {
				go c.suggestTitle()
			}

			// Verifica se precisa sumarizar (Memória)
			if count >= backend.SummaryThreshold {
				fmt.Printf("[ChatController] Gatilho de sumarização para sessão %s\n", sess.ID)
				c.engine.SummarizeSession(sess.ID)
			}

			c.refreshSidebar("")
		}
	}

	c.currentAIMessage = nil
	c.currentAIText = ""
	c.isStreaming = false
}
