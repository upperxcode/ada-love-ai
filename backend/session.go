package backend

import (
	"fmt"
	"sort"
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

type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Time       time.Time  `json:"time"`
}

type ChatSession struct {
	ID          string        `json:"id"`
	WorkspaceID string        `json:"workspace_id"` // Vínculo com o Workspace
	WorkerName  string        `json:"worker_name"`  // Worker vinculado ao chat
	Title       string        `json:"title"`
	Summary     string        `json:"summary"` // Memória de longo prazo (resumo)
	Messages    []ChatMessage `json:"messages"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Pinned      bool          `json:"pinned"`
}

type SessionManager struct {
	sessions map[string]*ChatSession
	activeID string
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*ChatSession),
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

func (s *SessionManager) CreateSession(title string, workspaceID string, workerName string) *ChatSession {
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
		ID:          id,
		WorkspaceID: workspaceID,
		WorkerName:  workerName,
		Title:       uniqueTitle,
		Messages:    []ChatMessage{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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

func (s *SessionManager) GetSession(id string) *ChatSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[id]
}

func (s *SessionManager) ListSessions(workspaceID string) []*ChatSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*ChatSession, 0)
	for _, sess := range s.sessions {
		if sess.WorkspaceID == workspaceID {
			list = append(list, sess)
		}
	}

	// Ordena por Pinned primeiro, depois por Título (ordem alfabética) e por fim data de atualização
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Pinned != list[j].Pinned {
			return list[i].Pinned // True (pinned) vem antes de False
		}
		titleI := strings.ToLower(list[i].Title)
		titleJ := strings.ToLower(list[j].Title)
		if titleI == "" {
			titleI = "zzz" 
		}
		if titleJ == "" {
			titleJ = "zzz"
		}
		if titleI == titleJ {
			return list[i].UpdatedAt.After(list[j].UpdatedAt)
		}
		return titleI < titleJ
	})

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

// AddRichMessage adiciona uma mensagem completa com metadados
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
	sess.Messages = append(sess.Messages, msg)
	sess.UpdatedAt = time.Now()

	return len(sess.Messages), true
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
