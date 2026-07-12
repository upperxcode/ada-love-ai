package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"time"

	"ada-love-ai/pkg/commands"
	"ada-love-ai/pkg/bus"
	"ada-love-ai/pkg/config"
	"ada-love-ai/pkg/providers"
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

func (e *Engine) RenameSession(sessionID, newTitle string) *ChatSession {
	fmt.Printf("[Engine] RenameSession: sessionID=%q newTitle=%q\n", sessionID, newTitle)
	// Ensure the session is loaded into SessionMgr (sessions are read from DB
	// on the frontend without passing through SessionMgr, so they may be missing).
	if e.SessionMgr.GetSession(sessionID) == nil && e.db != nil {
		if dbSess, err := e.db.GetSession(sessionID); err == nil && dbSess != nil {
			e.SessionMgr.LoadSession(dbSess)
			fmt.Printf("[Engine] RenameSession: loaded session %q from DB into SessionMgr\n", sessionID)
		}
	}
	e.SessionMgr.RenameSession(sessionID, newTitle)
	if sess := e.SessionMgr.GetSession(sessionID); sess != nil {
		fmt.Printf("[Engine] RenameSession: found session, title=%q\n", sess.Title)
		if e.db != nil {
			if err := e.db.SaveSession(*sess); err != nil {
				fmt.Printf("[Engine] RenameSession: error saving to DB: %v\n", err)
			}
		}
		return sess
	}
	fmt.Printf("[Engine] RenameSession: session %q not found\n", sessionID)
	return nil
}

func (e *Engine) SetSessionConfig(sessionID, model, provider, mode, thinking string) {
	fmt.Printf("[Engine] SetSessionConfig: session=%q model=%q provider=%q mode=%q thinking=%q\n",
		sessionID, model, provider, mode, thinking)
	sess := e.SessionMgr.GetSession(sessionID)
	if sess == nil {
		return
	}
	sess.Model = model
	sess.Provider = provider
	sess.Mode = mode
	sess.Thinking = thinking
	sess.UpdatedAt = time.Now()
	if e.db != nil {
		e.db.SaveSession(*sess)
	}
}

func (e *Engine) SendMessage(ctx context.Context, text string, sessionID string, modelOverride string, thinkingLevel string, mode string, isRetry bool) (result string, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 16384)
			n := runtime.Stack(buf, false)
			log := fmt.Sprintf("[Engine.SendMessage] PANIC RECOVERED: %v\n%s\n", r, buf[:n])
			fmt.Print(log)
			writePanicLog(log)
			result = ""
			retErr = fmt.Errorf("internal panic: %v", r)
		}
	}()

	e.eventBus.Emit(Event{Kind: EventKindTurnStart, SessionID: sessionID, Time: time.Now()})
	defer e.eventBus.Emit(Event{Kind: EventKindTurnEnd, SessionID: sessionID, Time: time.Now()})

	// Injeta o sessionID no contexto
	ctx = context.WithValue(ctx, "session_id", sessionID)

// Interceptação do orquestrador: ativa apenas para tarefas de desenvolvimento
		if e.orchestrator != nil && !isRetry {
			personality := e.resolveWorkspacePersonality(sessionID)
			if personality != "" && isDevTask(text) {
				fmt.Printf("[SendMessage] Orquestrador ativo para sessionID=%q\n", sessionID)
				return e.ProcessOrchestrated(ctx, text, sessionID, modelOverride)
			}
		}

	// Obtem informações da sessão para logging
	var sess *ChatSession
	if e.SessionMgr != nil {
		sess = e.SessionMgr.GetSession(sessionID)
	}

	// If session not in SessionMgr, load from DB (handles sessions from previous app runs)
	if sess == nil && sessionID != "" && e.db != nil {
		fmt.Printf("[SendMessage] session %q not in SessionMgr, loading from DB\n", sessionID)
		if dbSess, err := e.db.GetSession(sessionID); err == nil && dbSess != nil {
			e.SessionMgr.LoadSession(dbSess)
			sess = dbSess
			fmt.Printf("[SendMessage] loaded from DB: messages=%d\n", len(dbSess.Messages))
		} else {
			fmt.Printf("[SendMessage] session %q not found in DB either: %v\n", sessionID, err)
		}
	}

	// Sync the agent's folders/personality with the session's workspace WITHOUT
	// reloading the agent loop (reloading mid-send crashes the app because
	// goroutines from the old loop are still running). Instead we patch
	// cfg.Agents.Defaults and the live ContextBuilder in-place.
	if sess != nil && sess.WorkspaceID != "" {
		e.syncWorkspaceForTurn(sess.WorkspaceID)
	}

	// Variáveis para rastrear modelo e provider
	var resolvedModelID string
	var cached any

	// Se há override de modelo, resolve o provider correto e cacheia
	if modelOverride != "" {
		e.providerMu.RLock()
		cached, ok := e.providerCache[modelOverride]
		e.providerMu.RUnlock()
		if !ok {
			adaCfg := e.GetAdaConfig()
			fmt.Printf("[Engine] Model override=%q, searching %d models in model_list\n", modelOverride, len(adaCfg.ModelList))

			// Step 1: search model_list
		modelListLoop:
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
						break modelListLoop
					} else {
						fmt.Printf("[Engine] Provider creation FAILED: %v\n", err)
					}
				}
			}

			// Step 2: if not found in model_list, search providers config
			if cached == nil {
				parts := strings.SplitN(modelOverride, "/", 2)
				if len(parts) == 2 {
					providerName := parts[0]
					modelName := parts[1]
					fmt.Printf("[Engine] Searching providers config for provider=%q model=%q\n", providerName, modelName)
					if provCfg, ok := adaCfg.Providers[providerName]; ok {
						if _, exists := provCfg.Models[modelName]; exists {
							// Build a synthetic ModelConfig from provider config
							apiBase := provCfg.ApiUrl
							if apiBase == "" {
								apiBase = defaultAPIBaseFor(providerName, provCfg.TypeConnection)
							}
							synthetic := &config.ModelConfig{
								Provider:  providerName,
								ModelName: modelName,
								Model:     modelName,
								APIBase:   apiBase,
								Enabled:   true,
							}
							if apiKey := adaCfg.GetProviderAPIKey(providerName); apiKey != "" {
								synthetic.APIKeys = config.SimpleSecureStrings(apiKey)
							}
							if len(synthetic.APIKeys) == 0 && len(provCfg.ApiKeys) > 0 {
								if k := provCfg.GetAPIKey(); k != "" {
									synthetic.APIKeys = config.SimpleSecureStrings(k)
								}
							}
							fmt.Printf("[Engine] Built synthetic ModelConfig: apiBase=%q apiKey=%q\n", synthetic.APIBase, synthetic.APIKeys)
							p, _, err := e.CreateProviderFromModelConfig(synthetic)
							if err == nil && p != nil {
								cached = p
								resolvedModelID = modelName
								e.providerMu.Lock()
								e.providerCache[modelOverride] = cached
								e.providerMu.Unlock()
								e.overrideModelMu.Lock()
								e.overrideModelIDs[modelOverride] = modelName
								e.overrideModelMu.Unlock()
								fmt.Printf("[Engine] Provider OK from providers config, modelID=%q\n", modelName)
							} else {
								fmt.Printf("[Engine] Provider creation FAILED from providers config: %v\n", err)
							}
						} else {
							fmt.Printf("[Engine] Model %q not found in provider %q models\n", modelName, providerName)
						}
					} else {
						fmt.Printf("[Engine] Provider %q not found in providers config\n", providerName)
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
		// Ensure the override provider is wrapped in StreamingWrapper so deltas are emitted.
		if cached != nil {
			if _, alreadyWrapped := cached.(*StreamingWrapper); !alreadyWrapped {
				if lp, ok := cached.(providers.LLMProvider); ok {
					wrapper := NewStreamingWrapper(lp)
					wrapper.SetEventBus(e.eventBus)
					cached = wrapper
					e.providerMu.Lock()
					e.providerCache[modelOverride] = cached
					e.providerMu.Unlock()
				}
			}
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

	// 1. Adiciona mensagem do usuário ao histórico (in-memory only — persiste só após sucesso)
	isCommand := commands.HasCommandPrefix(text)
	if !isRetry && !isCommand {
		e.SessionMgr.AddMessage(sessionID, "user", text)
	}

	// Handle /clear directly without going through ProcessDirect (avoids building context with all messages)
	if isCommand && (strings.HasPrefix(text, "/clear") || strings.HasPrefix(text, "!clear")) {
		fmt.Printf("[SendMessage] /clear — short-circuit, sessionID=%q\n", sessionID)
		if e.agentLoop != nil {
			fmt.Printf("[SendMessage] /clear — calling ClearSession\n")
			e.agentLoop.ClearSession(ctx, sessionKey)
			fmt.Printf("[SendMessage] /clear — ClearSession done\n")
		}
		fmt.Printf("[SendMessage] /clear — calling ClearMessages\n")
		e.SessionMgr.ClearMessages(sessionID, 0)
		fmt.Printf("[SendMessage] /clear — ClearMessages done\n")
		if e.db != nil {
			if sess := e.SessionMgr.GetSession(sessionID); sess != nil {
				fmt.Printf("[SendMessage] /clear — saving to DB, messages=%d\n", len(sess.Messages))
				e.db.SaveSession(*sess)
				fmt.Printf("[SendMessage] /clear — DB save done\n")
			}
		}
		if e.eventBus != nil {
			fmt.Printf("[SendMessage] /clear — emitting EventKindCleared\n")
			e.eventBus.Emit(Event{Kind: EventKindCleared, SessionID: sessionID, Time: time.Now()})
			fmt.Printf("[SendMessage] /clear — event emitted\n")
		}
		fmt.Printf("[SendMessage] /clear — returning\n")
		return "Chat history cleared!", nil
	}

	fmt.Printf("[SendMessage] step=pre-log sessionID=%q sess=%v\n", sessionID, sess != nil)
	// Log do contexto sendo enviado (antes de ProcessDirect)
	// Obtem o histórico da sessão para logging
	historyForLog := []ChatMessage{}
	var logWorkspaceID, logAgentName string
	if sess != nil {
		historyForLog = sess.Messages
		logWorkspaceID = sess.WorkspaceID
	}
	
	// Determina o modelo e provider finais para logging
	logModel := modelOverride
	logProvider := ""
	if resolvedModelID != "" {
		logModel = resolvedModelID
	}
	if cached != nil {
		// Try to get provider info from cached provider
		if prov, ok := cached.(interface{ Name() string }); ok {
			logProvider = prov.Name()
		}
	}
	
	// Log do contexto completo
	LogChatContextWithHistory(
		sessionID,
		logWorkspaceID,
		logAgentName,
		logModel,
		logProvider,
		mode,
		thinkingLevel,
		text,
		finalPrompt,
		historyForLog,
		50, // max history entries to log
	)

	// Track the pending sessionID so the event bridge can map opaque keys
	e.setPendingSessionID(sessionID)

	fmt.Printf("[SendMessage] step=pre-ProcessDirect sessionID=%q agentLoop=%v\n", sessionID, e.agentLoop != nil)
	resp, err := e.agentLoop.ProcessDirect(ctx, finalPrompt, sessionKey)
	fmt.Printf("[SendMessage] step=post-ProcessDirect sessionID=%q err=%v resp_len=%d\n", sessionID, err, len(resp))
	if err != nil {
		// Rollback: remove mensagem do usuário que acabou de adicionar (skip for commands)
		if !isRetry && !isCommand {
			e.SessionMgr.RemoveLastMessage(sessionID)
		}
		e.eventBus.Emit(Event{Kind: EventKindError, SessionID: sessionID, Payload: ErrorPayload{Message: err.Error()}, Time: time.Now()})
		return "", err
	}

	// 2. Adiciona resposta do assistente ao histórico
	if !isCommand {
		e.SessionMgr.AddMessage(sessionID, "assistant", resp)

		// 3. Só agora persiste TUDO no DB (user + assistant)
		if sess, ok := e.SessionMgr.sessions[sessionID]; ok && e.db != nil {
			e.db.SaveSession(*sess)
		}

		// 4. Gerencia memória de curto prazo (sumarização)
		e.CheckAndSummarize(sessionID)
	}

	// Comandos tratados (OutcomeHandled) têm sua saída exibida em um painel
	// dedicado via evento chat:commandResult. Retornamos vazio para evitar
	// que a saída vire uma bolha de chat (que someria após o reload da sessão,
	// já que comandos não são persistidos no histórico).
	if isCommand {
		return "", nil
	}

	return resp, nil
}

func (e *Engine) CheckAndSummarize(sessionID string) {
	if sessionID == "" {
		return
	}
	// Trigger async summarization via worker
	if e.summarizer != nil {
		e.summarizer.TriggerSummarization(sessionID)
	}
}

func (e *Engine) SendTinyBrainMessage(ctx context.Context, prompt string) (string, error) {
	if e.adaCfg.TinyBrain.ModelName == "" {
		return e.SendMessage(ctx, prompt, "", "", "", "", false)
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
	fmt.Printf("[Engine] RefreshSessions: workspacePath=%q\n", workspacePath)
	if e.db == nil {
		fmt.Printf("[Engine] RefreshSessions: db is nil, skipping\n")
		return
	}

	sessions, err := e.db.GetSessions(workspacePath)
	if err != nil {
		fmt.Printf("[Engine] RefreshSessions: error loading sessions: %v\n", err)
		return
	}

	fmt.Printf("[Engine] RefreshSessions: loaded %d sessions for %q\n", len(sessions), workspacePath)
	e.SessionMgr.Reset()
	e.SessionMgr.LoadSessions(sessions)
}

// devTaskRe detects messages that are development-related tasks.
var devTaskRe = regexp.MustCompile(`(?i)(criar|crie|implemente|implementar|faca|fazer|desenvolva|codifique|adicione|adicionar|altere|modifique|corrija|consertar|crie\s+.*(api|rota|handler|servico|service|banco|tabela|query|migration|teste|test|componente|hook|pagina|tela|interface|formulario|form|modal|botao|layout|estilo|css|rota|middleware|model|struct|funcao|function|arquivo|cli|comando|endpoint|controller|use.?case|repositorio|repository|provider|config|docker|deploy|pipeline|action|workflow))|backend|frontend|api\s+rest|graphql|grpc|sql|banco\s+de\s+dados|teste\s+unitario|teste\s+de\s+integracao|e2e|test\s|bdd|tdd|refatorar|refatoracao|refactor|migrate|migracao|deploy|docker|kubernetes|k8s|ci/cd|pipeline|github\s+actions|gitlab\s+ci`)

// isDevTask returns true if the message appears to be a development task.
func isDevTask(text string) bool {
	return devTaskRe.MatchString(text)
}
