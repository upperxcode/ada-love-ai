package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ada-love-ai/pkg/bus"
)

func (e *Engine) TogglePin(sessionID string) {
	e.SessionMgr.TogglePin(sessionID)
}

func (e *Engine) DeleteSession(sessionID string) {
	e.SessionMgr.DeleteSession(sessionID)
	if e.db != nil {
		e.db.DeleteSession(sessionID)
	}
}

func (e *Engine) RenameSession(sessionID, newTitle string) {
	e.SessionMgr.RenameSession(sessionID, newTitle)
	if sess := e.SessionMgr.GetSession(sessionID); sess != nil && e.db != nil {
		e.db.SaveSession(*sess)
	}
}

func (e *Engine) SendMessage(ctx context.Context, text string, sessionID string, modelOverride string, thinkingLevel string, mode string) (string, error) {
	e.eventBus.Emit(Event{Kind: EventKindTurnStart, SessionID: sessionID, Time: time.Now()})
	defer e.eventBus.Emit(Event{Kind: EventKindTurnEnd, SessionID: sessionID, Time: time.Now()})

	// Injeta o sessionID no contexto
	ctx = context.WithValue(ctx, "session_id", sessionID)

	// Se há override de modelo, resolve o provider correto e cacheia
	var resolvedModelID string
	if modelOverride != "" {
		e.providerMu.RLock()
		cached, ok := e.providerCache[modelOverride]
		e.providerMu.RUnlock()
		if !ok {
			adaCfg := e.GetAdaConfig()
			fmt.Printf("[Engine] Model override=%q, searching %d models\n", modelOverride, len(adaCfg.ModelList))
			for i, mc := range adaCfg.ModelList {
				if mc == nil {
					continue
				}
				provider := strings.TrimSpace(mc.Provider)
				modelName := strings.TrimSpace(mc.ModelName)
				modelField := strings.TrimSpace(mc.Model)
				fullKey := provider + "/" + modelName
				if modelName == modelOverride || modelField == modelOverride || fullKey == modelOverride {
					fmt.Printf("[Engine] Match at index %d: provider=%q modelField=%q\n", i, provider, modelField)
					p, _, err := e.CreateProviderFromModelConfig(mc)
					if err == nil && p != nil {
						cached = p
						resolvedModelID = modelField
						e.providerMu.Lock()
						e.providerCache[modelOverride] = cached
						e.providerMu.Unlock()
						e.overrideModelMu.Lock()
						e.overrideModelIDs[modelOverride] = modelField
						e.overrideModelMu.Unlock()
						fmt.Printf("[Engine] Provider OK, modelID=%q\n", modelField)
						break
					} else {
						fmt.Printf("[Engine] Provider creation FAILED: %v\n", err)
					}
				}
			}
			if cached == nil {
				fmt.Printf("[Engine] NO MATCH for override=%q\n", modelOverride)
			}
		} else {
			// Re-read the correct model ID from cache.
			e.overrideModelMu.RLock()
			resolvedModelID = e.overrideModelIDs[modelOverride]
			e.overrideModelMu.RUnlock()
		}
		// Pass both the frontend key and the resolved model ID.
		ctx = bus.WithOverrides(ctx, modelOverride, resolvedModelID, thinkingLevel, cached, mode)
	} else if thinkingLevel != "" || mode != "" {
		ctx = bus.WithOverrides(ctx, "", "", thinkingLevel, nil, mode)
	}

	// Ada-Love utiliza chaves de sessão para manter o histórico.
	sessionKey := "ada:default"
	if sessionID != "" {
		sessionKey = "ada:" + sessionID
	}

	// Prepara o prompt com memória de longo prazo se existir
	finalPrompt := text

	// 1. Adiciona mensagem do usuário ao histórico
	e.SessionMgr.AddMessage(sessionID, "user", text)
	if sess, ok := e.SessionMgr.sessions[sessionID]; ok && e.db != nil {
		e.db.SaveSession(*sess)
	}

	// Track the pending sessionID so the event bridge can map opaque keys
	e.setPendingSessionID(sessionID)

	resp, err := e.agentLoop.ProcessDirect(ctx, finalPrompt, sessionKey)
	if err != nil {
		e.eventBus.Emit(Event{Kind: EventKindError, SessionID: sessionID, Payload: ErrorPayload{Message: err.Error()}, Time: time.Now()})
		return "", err
	}

	// 2. Adiciona resposta do assistente ao histórico
	e.SessionMgr.AddMessage(sessionID, "assistant", resp)
	if sess, ok := e.SessionMgr.sessions[sessionID]; ok && e.db != nil {
		e.db.SaveSession(*sess)
	}

	// 3. Gerencia memória de curto prazo (sumarização)
	e.CheckAndSummarize(sessionID)

	return resp, nil
}

func (e *Engine) CheckAndSummarize(sessionID string) {
	if sessionID == "" {
		return
	}

	sess, ok := e.SessionMgr.sessions[sessionID]
	if !ok || len(sess.Messages) < SummaryThreshold {
		return
	}

	fmt.Printf("[Engine] Iniciando sumarização para sessão %s (%d mensagens)...\n", sessionID, len(sess.Messages))

	// Prepara o contexto para sumarizar
	var sb strings.Builder
	for _, m := range sess.Messages {
		sb.WriteString(fmt.Sprintf("%s: %s\n", m.Role, m.Content))
	}

	prompt := fmt.Sprintf("Por favor, resuma a conversa abaixo de forma concisa, mantendo os pontos principais e decisões tomadas:\n\n%s", sb.String())

	summary, err := e.SendTinyBrainMessage(context.Background(), prompt)
	if err != nil {
		fmt.Printf("[Engine] Erro ao sumarizar: %v\n", err)
		return
	}

	e.SessionMgr.SetSummary(sessionID, summary)
	e.SessionMgr.ClearMessages(sessionID, SummaryKeepLast)

	if e.db != nil {
		e.db.SaveSession(*sess)
	}
}

func (e *Engine) SendTinyBrainMessage(ctx context.Context, prompt string) (string, error) {
	if e.adaCfg.TinyBrain.ModelName == "" {
		return e.SendMessage(ctx, prompt, "", "", "", "")
	}

	// Tenta encontrar a URL base para o provider no model_list
	apiBase := ""
	apiKey := ""
	for _, m := range e.adaCfg.ModelList {
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

	return "", fmt.Errorf("nenhuma resposta do provedor")
}

func (e *Engine) RefreshSessions() {
	workspacePath := e.GetActiveWorkspace()
	if e.db == nil {
		return
	}

	sessions, err := e.db.GetSessions(workspacePath)
	if err != nil {
		fmt.Printf("[Engine] Erro ao carregar sessões: %v\n", err)
		return
	}

	e.SessionMgr.Reset()
	e.SessionMgr.LoadSessions(sessions)
}
