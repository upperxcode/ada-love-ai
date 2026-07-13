package backend

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// ServedByMeta records, for auditing, how a given assistant message was produced.
// It captures the effective agent/persona, the routing intent, and the model/provider
// actually used to generate the response. Persisted as JSON in messages.served_by.
type ServedByMeta struct {
	Agent    string `json:"agent"`     // resolved agent id (e.g. "golang_agent", "general-worker")
	Persona  string `json:"persona"`   // short label of the persona used (e.g. "worker:Ada-Worker")
	Intent   string `json:"intent"`    // TinyBrain intent (e.g. "GENERAL", "GO_PROGRAMMING")
	Model    string `json:"model"`     // effective model name used
	Provider string `json:"provider"`  // effective provider used
	RoutedBy string `json:"routed_by"` // how the agent was chosen (e.g. "intent:general", "default")
}

type ChatMessage struct {
	ID         int64         `json:"id"`
	Role       string        `json:"role"`
	Content    string        `json:"content"`
	ToolCalls  []ToolCall    `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	Time       time.Time     `json:"time"`
	ServedBy   *ServedByMeta `json:"served_by,omitempty"`
}

type ChatSession struct {
	ID              string `json:"id"`
	WorkspaceID     string `json:"workspace_id"`
	WorkerName      string `json:"worker_name"`
	ParentSessionID string `json:"parent_session_id"`
	Title           string `json:"title"`
	Summary         string `json:"summary"`
	// Config per-chat
	Model    string `json:"model"`
	Provider string `json:"provider"`
	Mode     string `json:"mode"`     // "ask"|"plan"|"auto"|"full"
	Thinking string `json:"thinking"` // "" ou "high"
	// Summarization
	SummarizedContext   string        `json:"summarized_context"`     // Resumo contínuo + últimas N Q&A
	SummarizedAt        time.Time     `json:"summarized_at"`          // Quando foi sumarizado
	LastSummarizedMsgID int64         `json:"last_summarized_msg_id"` // ID da última msg incluída no resumo
	Messages            []ChatMessage `json:"messages"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
	Pinned              bool          `json:"pinned"`
}

type SessionManager struct {
	sessions map[string]*ChatSession
	activeID string
	mu       sync.RWMutex
	db       *Store
}

func NewSessionManager(db *Store) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*ChatSession),
		db:       db,
	}
}
func (s *SessionManager) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions = make(map[string]*ChatSession)
	s.activeID = ""
}

func (s *SessionManager) LoadSessions(sessions []*ChatSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sess := range sessions {
		s.sessions[sess.ID] = sess
	}
}

func (s *SessionManager) CreateSession(title string, workspaceID string, parentSessionID string) *ChatSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	uniqueTitle := UniquifyName(title, func(t string) bool {
		for _, sess := range s.sessions {
			if sess.WorkspaceID == workspaceID && strings.EqualFold(sess.Title, t) {
				return true
			}
		}
		return false
	})

	id := fmt.Sprintf("session_%d", time.Now().UnixNano())
	session := &ChatSession{
		ID:              id,
		WorkspaceID:     workspaceID,
		ParentSessionID: parentSessionID,
		Title:           uniqueTitle,
		Messages:        []ChatMessage{},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	s.sessions[id] = session
	s.activeID = id
	return session
}

func (s *SessionManager) GetActiveID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeID
}

func (s *SessionManager) GetActiveSession() *ChatSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[s.activeID]
}

// LoadSession inserts a session from DB into SessionMgr.
func (s *SessionManager) LoadSession(sess *ChatSession) {
	if sess == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.ID] = sess
}

func (s *SessionManager) GetSession(id string) *ChatSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[id]
}

func (s *SessionManager) ListSessions(workspaceID string) []*ChatSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, workspace_path, worker_name, parent_session_id, title, summary, model, provider, mode, thinking, summarized_context, summarized_at, last_summarized_msg_id, pinned, created_at, updated_at
		FROM sessions
		WHERE workspace_path = ?
		ORDER BY pinned DESC, updated_at DESC`, workspaceID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	list := make([]*ChatSession, 0)
	for rows.Next() {
		sess := &ChatSession{}
		var summarizedAt sql.NullTime
		err := rows.Scan(&sess.ID, &sess.WorkspaceID, &sess.WorkerName, &sess.ParentSessionID, &sess.Title, &sess.Summary,
			&sess.Model, &sess.Provider, &sess.Mode, &sess.Thinking,
			&sess.SummarizedContext, &summarizedAt, &sess.LastSummarizedMsgID,
			&sess.Pinned, &sess.CreatedAt, &sess.UpdatedAt)
		if err != nil {
			continue // em production, talvez logar; aqui ignoramos linha com erro
		}
		if summarizedAt.Valid {
			sess.SummarizedAt = summarizedAt.Time
		}
		list = append(list, sess)
	}
	return list
}

func (s *SessionManager) SearchSessions(query string, workspaceID string) []*ChatSession {
	all := s.ListSessions(workspaceID)
	if query == "" {
		return all
	}

	filtered := make([]*ChatSession, 0)
	query = strings.ToLower(query)
	for _, sess := range all {
		if strings.Contains(strings.ToLower(sess.Title), query) {
			filtered = append(filtered, sess)
		}
	}
	return filtered
}

func (s *SessionManager) SetActive(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeID = id
}

func (s *SessionManager) TogglePin(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok {
		sess.Pinned = !sess.Pinned
	}
}

func (s *SessionManager) DeleteSession(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
	if s.activeID == id {
		s.activeID = ""
	}
}

func (s *SessionManager) RenameSession(id string, newTitle string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok {
		sess.Title = UniquifyName(newTitle, func(t string) bool {
			for sid, sptr := range s.sessions {
				if sid == id {
					continue
				}
				if sptr.WorkspaceID == sess.WorkspaceID && strings.EqualFold(sptr.Title, t) {
					return true
				}
			}
			return false
		})
		sess.UpdatedAt = time.Now()
	}
}

// AddMessage adiciona uma mensagem simples e retorna o total de mensagens e se a sessão existe
func (s *SessionManager) AddMessage(id, role, content string) (int, bool) {
	return s.AddRichMessage(id, ChatMessage{
		Role:    role,
		Content: content,
	})
}

// AddMessageWithMeta adiciona uma mensagem de assistente carregando metadados de
// auditoria (ServedBy), registrando qual persona/agente/modelo/provider respondeu.
func (s *SessionManager) AddMessageWithMeta(id, role, content string, servedBy *ServedByMeta) (int, bool) {
	return s.AddRichMessage(id, ChatMessage{
		Role:     role,
		Content:  content,
		ServedBy: servedBy,
	})
}

// AddRichMessage adiciona ou atualiza uma mensagem em uma sessão de forma incremental.
// Em vez de apagar todas e reinserir, faz UPSERT por ID para reduzir bloqueios.
func (s *SessionManager) AddRichMessage(id string, msg ChatMessage) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[id]
	if !ok {
		return 0, false
	}
	if msg.Time.IsZero() {
		msg.Time = time.Now()
	}

	// Só faz upsert se ID for válido (> 0). IDs zero significam mensagem nova
	// sem ID atribuído pelo banco — sem isso, todas as mensagens novas (ID=0)
	// casam entre si e a segunda substitui a primeira.
	if msg.ID > 0 {
		existing := -1
		for i, m := range sess.Messages {
			if m.ID == msg.ID {
				existing = i
				break
			}
		}
		if existing >= 0 {
			sess.Messages[existing] = msg
			sess.UpdatedAt = time.Now()
			return len(sess.Messages), true
		}
	}

	sess.Messages = append(sess.Messages, msg)
	sess.UpdatedAt = time.Now()

	return len(sess.Messages), true
}

// SaveSession atualizada para transação curta e UPSERT incremental de mensagens.
// Mantém bloqueio RWLock apenas pelo tempo necessário.
func (s *SessionManager) SaveSession(sess ChatSession) error {
	s.mu.Lock()
	// Atualiza cabeçalho da sessão (sem apagar mensagens)
	sess.UpdatedAt = time.Now()
	current, ok := s.sessions[sess.ID]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("sessão não encontrada: %s", sess.ID)
	}
	current.Title = sess.Title
	current.Summary = sess.Summary
	current.Model = sess.Model
	current.Provider = sess.Provider
	current.Mode = sess.Mode
	current.Thinking = sess.Thinking
	current.SummarizedContext = sess.SummarizedContext
	current.SummarizedAt = sess.SummarizedAt
	current.LastSummarizedMsgID = sess.LastSummarizedMsgID
	current.Pinned = sess.Pinned
	// Atualiza a sessão no mapa
	s.sessions[sess.ID] = current
	// Libera lock antes de trabalhar com DB para não bloquear outras reads
	s.mu.Unlock()

	// UPSERT incremental de mensagens em transação curta
	for _, msg := range sess.Messages {
		var servedByJSON string
		if msg.ServedBy != nil {
			if b, err := json.Marshal(msg.ServedBy); err == nil {
				servedByJSON = string(b)
			}
		}
		// Tenta UPDATE
		res, err := s.db.Exec(`UPDATE messages SET role=?, content=?, time=?, served_by=? WHERE id=?`,
			msg.Role, msg.Content, msg.Time, servedByJSON, msg.ID)
		if err != nil {
			return err
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			// Insere nova mensagem
			_, err = s.db.Exec(`INSERT INTO messages (id, session_id, role, content, time, served_by) VALUES (?,?,?,?,?,?)`,
				msg.ID, sess.ID, msg.Role, msg.Content, msg.Time, servedByJSON)
			if err != nil {
				return err
			}
		}
	}
	// Opcional: remover mensagens muito antigas aqui (com DELETE LIMIT), evite grandes deletes.
	return nil
}

func (s *SessionManager) SetSummary(id, summary string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok {
		sess.Summary = summary
	}
}

func (s *SessionManager) ClearMessages(id string, keepLast int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok {
		if len(sess.Messages) > keepLast {
			sess.Messages = sess.Messages[len(sess.Messages)-keepLast:]
		}
	}
}

func (s *SessionManager) RemoveLastMessage(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok && len(sess.Messages) > 0 {
		sess.Messages = sess.Messages[:len(sess.Messages)-1]
	}
}
