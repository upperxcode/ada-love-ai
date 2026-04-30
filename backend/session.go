package backend

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type ChatMessage struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

type ChatSession struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Summary   string        `json:"summary"` // Memória de longo prazo (resumo)
	Messages  []ChatMessage `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Pinned    bool          `json:"pinned"`
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

func (s *SessionManager) CreateSession(title string) *ChatSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("session_%d", time.Now().UnixNano())
	session := &ChatSession{
		ID:        id,
		Title:     title,
		Messages:  []ChatMessage{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.sessions[id] = session
	s.activeID = id
	return session
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

func (s *SessionManager) ListSessions() []*ChatSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*ChatSession, 0, len(s.sessions))
	for _, sess := range s.sessions {
		list = append(list, sess)
	}

	// Ordena por Pinned primeiro, depois por data de atualização (mais recentes primeiro)
	sort.Slice(list, func(i, j int) bool {
		if list[i].Pinned != list[j].Pinned {
			return list[i].Pinned // true vem antes de false no sort descending? Não, temos que retornar se i deve vir antes de j
		}
		return list[i].UpdatedAt.After(list[j].UpdatedAt)
	})

	return list
}

func (s *SessionManager) SearchSessions(query string) []*ChatSession {
	all := s.ListSessions()
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
		sess.Title = newTitle
		sess.UpdatedAt = time.Now()
	}
}

// AddMessage adiciona uma mensagem e retorna o total de mensagens e se a sessão existe
func (s *SessionManager) AddMessage(id, role, content string) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[id]
	if !ok {
		return 0, false
	}

	sess.Messages = append(sess.Messages, ChatMessage{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	})
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
