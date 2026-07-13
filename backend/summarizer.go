package backend

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"ada-love-ai/pkg/config"
	"ada-love-ai/pkg/providers"
)

// SummarizerWorker gerencia a sumarização assíncrona de sessões de chat
type SummarizerWorker struct {
	engine  *Engine
	mu      sync.Mutex
	running map[string]bool // sessionID -> running
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// NewSummarizerWorker cria um novo worker de sumarização
func NewSummarizerWorker(e *Engine) *SummarizerWorker {
	return &SummarizerWorker{
		engine:  e,
		running: make(map[string]bool),
		stopCh:  make(chan struct{}),
	}
}

// Start inicia o worker
func (sw *SummarizerWorker) Start() {
	sw.wg.Add(1)
	go sw.run()
	log.Printf("[Summarizer] Worker iniciado")
}

// Stop para o worker
func (sw *SummarizerWorker) Stop() {
	close(sw.stopCh)
	sw.wg.Wait()
	log.Printf("[Summarizer] Worker parado")
}

// TriggerSummarization solicita sumarização para uma sessão (não bloqueante)
func (sw *SummarizerWorker) TriggerSummarization(sessionID string) {
	sw.mu.Lock()
	if sw.running[sessionID] {
		sw.mu.Unlock()
		return
	}
	sw.running[sessionID] = true
	sw.mu.Unlock()

	// Executa em background
	sw.wg.Add(1)
	go func() {
		defer sw.wg.Done()
		defer func() {
			sw.mu.Lock()
			delete(sw.running, sessionID)
			sw.mu.Unlock()
		}()
		sw.summarizeSession(sessionID)
	}()
}

// run loop principal - verifica sessões que precisam de sumarização periodicamente
func (sw *SummarizerWorker) run() {
	defer sw.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sw.stopCh:
			return
		case <-ticker.C:
			sw.checkSessions()
		}
	}
}

// checkSessions verifica todas as sessões ativas e dispara sumarização se necessário
func (sw *SummarizerWorker) checkSessions() {
	workspaces := sw.engine.GetAdaConfig().Workspaces
	for _, ws := range workspaces {
		if !ws.Enabled {
			continue
		}
		// MaxContextLength = 0 significa desabilitado (usa default do agente)
		if ws.MaxContextLength <= 0 {
			continue
		}
		sessions, err := sw.engine.db.GetSessions(ws.Path)
		if err != nil {
			continue
		}
		for _, sess := range sessions {
			if sw.shouldSummarize(sess, ws) {
				sw.TriggerSummarization(sess.ID)
			}
		}
	}
}

// shouldSummarize verifica se a sessão precisa ser sumarizada
func (sw *SummarizerWorker) shouldSummarize(sess *ChatSession, ws WorkspaceConfig) bool {
	if len(sess.Messages) == 0 {
		return false
	}
	// Estimativa simples de tokens (4 chars ≈ 1 token para pt-BR)
	estimatedTokens := sw.estimateTokens(sess.Messages)
	return estimatedTokens > ws.MaxContextLength
}

func (sw *SummarizerWorker) estimateTokens(msgs []ChatMessage) int {
	totalChars := 0
	for _, m := range msgs {
		totalChars += len(m.Content)
	}
	return totalChars / 4
}

// summarizeSession executa a sumarização da sessão
func (sw *SummarizerWorker) summarizeSession(sessionID string) {
	ctx := context.Background()

	sess, err := sw.engine.db.GetSession(sessionID)
	if err != nil {
		log.Printf("[Summarizer] Erro ao carregar sessão %s: %v", sessionID, err)
		return
	}
	if sess == nil {
		return
	}

	ws := sw.findWorkspace(sess.WorkspaceID)
	if ws == nil {
		return
	}

	// Configura o modelo de sumarização (usa EmbeddingModel do workspace)
	modelName := ws.EmbeddingModel
	providerName := ws.EmbeddingProvider
	if modelName == "" {
		// Fallback para TinyBrain
		modelName = sw.engine.GetAdaConfig().TinyBrain.EmbeddingModel
		providerName = sw.engine.GetAdaConfig().TinyBrain.EmbeddingProvider
	}
	if modelName == "" {
		log.Printf("[Summarizer] Sessão %s: nenhum modelo de sumarização configurado", sessionID)
		return
	}

	// Cria provider para sumarização
	prov, err := sw.createSummarizerProvider(modelName, providerName)
	if err != nil {
		log.Printf("[Summarizer] Sessão %s: erro ao criar provider: %v", sessionID, err)
		return
	}

	// Constrói prompt de sumarização incremental
	prompt := sw.buildSummarizationPrompt(sess, *ws)

	log.Printf("[Summarizer] Sessão %s: sumarizando (tokens estimados: %d, max: %d)",
		sessionID, sw.estimateTokens(sess.Messages), ws.MaxContextLength)

	// Chama o modelo
	resp, err := prov.Chat(ctx, []providers.Message{
		{Role: "system", Content: "Você é um especialista em sumarização de conversas. Mantenha o contexto importante e decisões técnicas."},
		{Role: "user", Content: prompt},
	}, nil, modelName, map[string]any{
		"max_tokens":  2000,
		"temperature": 0.3,
	})
	if err != nil {
		log.Printf("[Summarizer] Sessão %s: erro na sumarização: %v", sessionID, err)
		return
	}

	if resp == nil || resp.Content == "" {
		log.Printf("[Summarizer] Sessão %s: resposta vazia", sessionID)
		return
	}

	// Atualiza a sessão com o novo contexto sumarizado
	newContext := sw.mergeSummarizedContext(sess.SummarizedContext, resp.Content)
	lastMsgID := sw.getLastMessageID(sess.Messages)

	sess.SummarizedContext = newContext
	sess.SummarizedAt = time.Now()
	sess.LastSummarizedMsgID = lastMsgID

	if err := sw.engine.db.SaveSession(*sess); err != nil {
		log.Printf("[Summarizer] Sessão %s: erro ao salvar: %v", sessionID, err)
		return
	}

	log.Printf("[Summarizer] Sessão %s: sumarização concluída (contexto: %d chars)",
		sessionID, len(newContext))
}

// createSummarizerProvider cria o provider para o modelo de sumarização
func (sw *SummarizerWorker) createSummarizerProvider(modelName, providerName string) (providers.LLMProvider, error) {
	mc := &config.ModelConfig{
		Provider:  providerName,
		ModelName: modelName,
		Model:     modelName,
		Enabled:   true,
	}
	prov, _, err := sw.engine.CreateProviderFromModelConfig(mc)
	if err != nil {
		return nil, err
	}
	llmProv, ok := prov.(providers.LLMProvider)
	if !ok {
		return nil, fmt.Errorf("provider não implementa LLMProvider")
	}
	return llmProv, nil
}

// buildSummarizationPrompt constrói o prompt para sumarização incremental
func (sw *SummarizerWorker) buildSummarizationPrompt(sess *ChatSession, ws WorkspaceConfig) string {
	var sb strings.Builder

	// Contexto anterior (se existir)
	if sess.SummarizedContext != "" {
		sb.WriteString("=== RESUMO ANTERIOR ===\n")
		sb.WriteString(sess.SummarizedContext)
		sb.WriteString("\n\n")
	}

	// Mensagens novas (após o último sumarizado)
	newMessages := sw.getNewMessages(sess)
	if len(newMessages) == 0 {
		return ""
	}

	sb.WriteString("=== NOVAS MENSAGENS PARA SUMARIZAR ===\n")
	for _, msg := range newMessages {
		role := "Usuário"
		if msg.Role == "assistant" {
			role = "Assistente"
		}
		sb.WriteString(fmt.Sprintf("[%s]: %s\n\n", role, msg.Content))
	}

	sb.WriteString("\n=== INSTRUÇÕES ===\n")
	sb.WriteString("Gere um resumo CONCISO e CONTINUADO que combine o resumo anterior (se houver) com as novas mensagens.\n")
	sb.WriteString("Mantenha:\n")
	sb.WriteString("- Decisões técnicas e arquiteturais\n")
	sb.WriteString("- Nomes de arquivos, funções, variáveis importantes\n")
	sb.WriteString("- Problemas identificados e soluções propostas\n")
	sb.WriteString("- Contexto do projeto e objetivos\n")
	sb.WriteString("Remova:\n")
	sb.WriteString("- Saudações, conversas triviais, repetições\n")
	sb.WriteString("- Detalhes verbosos de código (mantenha apenas assinaturas/nomes-chave)\n")
	sb.WriteString("\nFormato: Parágrafos diretos, sem bullet points, em português.\n")

	return sb.String()
}

// getNewMessages retorna mensagens após o último ID sumarizado
func (sw *SummarizerWorker) getNewMessages(sess *ChatSession) []ChatMessage {
	if sess.LastSummarizedMsgID == 0 {
		return sess.Messages
	}
	var result []ChatMessage
	for _, m := range sess.Messages {
		if m.ID > sess.LastSummarizedMsgID {
			result = append(result, m)
		}
	}
	return result
}

// getLastMessageID retorna o ID da última mensagem
func (sw *SummarizerWorker) getLastMessageID(msgs []ChatMessage) int64 {
	if len(msgs) == 0 {
		return 0
	}
	return msgs[len(msgs)-1].ID
}

// mergeSummarizedContext combina o contexto anterior com o novo resumo
func (sw *SummarizerWorker) mergeSummarizedContext(oldContext, newSummary string) string {
	if oldContext == "" {
		return newSummary
	}
	// Evita duplicação: se o novo resumo já contém o antigo, usa só o novo
	if strings.Contains(newSummary, oldContext) {
		return newSummary
	}
	return oldContext + "\n\n--- ATUALIZAÇÃO ---\n\n" + newSummary
}

// findWorkspace encontra workspace pelo path
func (sw *SummarizerWorker) findWorkspace(path string) *WorkspaceConfig {
	for i := range sw.engine.GetAdaConfig().Workspaces {
		if sw.engine.GetAdaConfig().Workspaces[i].Path == path {
			return &sw.engine.GetAdaConfig().Workspaces[i]
		}
	}
	return nil
}

// BuildContextForLLM constrói o contexto que será enviado ao LLM principal
// Usa: SummarizedContext + últimas N Q&A completas (MaxPromptSend)
func (sw *SummarizerWorker) BuildContextForLLM(sessionID string, workspaceID string) []providers.Message {
	sess, err := sw.engine.db.GetSession(sessionID)
	if err != nil || sess == nil {
		return nil
	}
	ws := sw.findWorkspace(sess.WorkspaceID)
	if ws == nil {
		return nil
	}
	var msgs []providers.Message

	// 1. System prompt com contexto sumarizado
	if sess.SummarizedContext != "" {
		msgs = append(msgs, providers.Message{
			Role:    "system",
			Content: "Contexto da conversa (resumo):\n" + sess.SummarizedContext,
		})
	}

	// 2. Últimas N mensagens completas (MaxPromptSend * 2 = Q+A pairs)
	maxPairs := ws.MaxPromptSend
	if maxPairs <= 0 {
		maxPairs = 3 // default
	}
	maxMsgs := maxPairs * 2

	start := len(sess.Messages) - maxMsgs
	if start < 0 {
		start = 0
	}

	for _, m := range sess.Messages[start:] {
		msgs = append(msgs, providers.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	return msgs
}
