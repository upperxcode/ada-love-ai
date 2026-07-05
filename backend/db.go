package backend

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Memory struct {
	ID            int       `json:"id"`
	WorkspacePath string    `json:"workspace_path"`
	Content       string    `json:"content"`
	Importance    int       `json:"importance"`
	Embedding     []float32 `json:"embedding"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	// Garante que o diretório existe
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			description TEXT,
			path TEXT UNIQUE,
			personality TEXT,
			folders TEXT, -- JSON
			knowledge TEXT, -- JSON
			agents TEXT, -- JSON
			skills TEXT, -- JSON
			tools TEXT -- JSON
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			workspace_path TEXT,
			title TEXT,
			summary TEXT,
			pinned INTEGER DEFAULT 0,
			embedding BLOB,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT,
			role TEXT,
			content TEXT,
			time DATETIME,
			FOREIGN KEY(session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS memories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_path TEXT,
			content TEXT,
			importance INTEGER DEFAULT 1,
			embedding BLOB,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("erro ao criar tabela: %v\nQuery: %s", err, q)
		}
	}

	// Migrações manuais para colunas novas
	s.db.Exec("ALTER TABLE sessions ADD COLUMN embedding BLOB") // Ignora erro se já existir
	s.db.Exec("ALTER TABLE memories ADD COLUMN embedding BLOB") // Ignora erro se já existir
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN tools TEXT")   // Ignora erro se já existir
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN description TEXT") // Migração para novo campo manual

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// --- Operações de Configuração ---

func (s *Store) SetGlobalConfig(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)`, key, string(data))
	return err
}

func (s *Store) GetGlobalConfig(key string, target interface{}) (bool, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM config WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = json.Unmarshal([]byte(value), target)
	return true, err
}

// --- Operações de Workspace ---

func (s *Store) SaveWorkspace(ws WorkspaceConfig) error {
	folders, _ := json.Marshal(ws.Folders)
	knowledge, _ := json.Marshal(ws.Knowledge)
	agents, _ := json.Marshal(ws.WorkspaceAgents)
	skills, _ := json.Marshal(ws.Skills)
	tools, _ := json.Marshal(ws.Tools)

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO workspaces (title, description, path, personality, folders, knowledge, agents, skills, tools)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ws.Title, ws.Description, ws.Path, ws.Personality,
		string(folders), string(knowledge), string(agents), string(skills), string(tools),
	)
	return err
}

func (s *Store) GetWorkspaces() ([]WorkspaceConfig, error) {
	rows, err := s.db.Query(`SELECT title, description, path, personality, folders, knowledge, agents, skills, tools FROM workspaces`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []WorkspaceConfig
	for rows.Next() {
		var ws WorkspaceConfig
		var folders, knowledge, agents, skills, tools string
		err := rows.Scan(&ws.Title, &ws.Description, &ws.Path, &ws.Personality, &folders, &knowledge, &agents, &skills, &tools)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(folders), &ws.Folders)
		json.Unmarshal([]byte(knowledge), &ws.Knowledge)
		json.Unmarshal([]byte(agents), &ws.WorkspaceAgents)
		json.Unmarshal([]byte(skills), &ws.Skills)
		json.Unmarshal([]byte(tools), &ws.Tools)
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

func (s *Store) DeleteWorkspace(path string) error {
	_, err := s.db.Exec(`DELETE FROM workspaces WHERE path = ?`, path)
	return err
}

// --- Operações de Sessão ---

func (s *Store) SaveSession(sess ChatSession) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO sessions (id, workspace_path, title, summary, pinned, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.WorkspaceID, sess.Title, sess.Summary, sess.Pinned, sess.CreatedAt, sess.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Salva mensagens - Para simplicidade, limpamos e reinserimos as mensagens da sessão
	// Em uma implementação mais avançada, faríamos apenas o append.
	_, err = s.db.Exec(`DELETE FROM messages WHERE session_id = ?`, sess.ID)
	if err != nil {
		return err
	}

	for _, msg := range sess.Messages {
		_, err = s.db.Exec(`INSERT INTO messages (session_id, role, content, time) VALUES (?, ?, ?, ?)`,
			sess.ID, msg.Role, msg.Content, msg.Time)
		if err != nil {
			log.Printf("Erro ao salvar mensagem: %v", err)
		}
	}

	return nil
}

func (s *Store) AddMessageToSession(sessionID string, role string, content string) error {
	_, err := s.db.Exec(`INSERT INTO messages (session_id, role, content, time) VALUES (?, ?, ?, ?)`,
		sessionID, role, content, time.Now())
	return err
}

func (s *Store) GetSessions(workspacePath string) ([]*ChatSession, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_path, title, summary, pinned, created_at, updated_at 
		FROM sessions WHERE workspace_path = ?
		ORDER BY pinned DESC, updated_at DESC`, workspacePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*ChatSession
	for rows.Next() {
		sess := &ChatSession{}
		err := rows.Scan(&sess.ID, &sess.WorkspaceID, &sess.Title, &sess.Summary, &sess.Pinned, &sess.CreatedAt, &sess.UpdatedAt)
		if err != nil {
			return nil, err
		}
		
		// Carrega mensagens para esta sessão
		msgRows, err := s.db.Query(`SELECT role, content, time FROM messages WHERE session_id = ? ORDER BY time ASC`, sess.ID)
		if err == nil {
			for msgRows.Next() {
				var msg ChatMessage
				msgRows.Scan(&msg.Role, &msg.Content, &msg.Time)
				sess.Messages = append(sess.Messages, msg)
			}
			msgRows.Close()
		}
		
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

func (s *Store) DeleteSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// --- Operações de Memória ---

func (s *Store) SaveMemory(m Memory) error {
	now := time.Now()
	emb := Float32ToByte(m.Embedding)
	_, err := s.db.Exec(`
		INSERT INTO memories (workspace_path, content, importance, embedding, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		m.WorkspacePath, m.Content, m.Importance, emb, now, now)
	return err
}

func (s *Store) GetMemories(workspacePath string) ([]Memory, error) {
	rows, err := s.db.Query(`SELECT id, content, importance, embedding, created_at FROM memories WHERE workspace_path = ? ORDER BY importance DESC, created_at DESC`, workspacePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		var emb []byte
		if err := rows.Scan(&m.ID, &m.Content, &m.Importance, &emb, &m.CreatedAt); err == nil {
			m.Embedding = ByteToFloat32(emb)
			memories = append(memories, m)
		}
	}
	return memories, nil
}

func (s *Store) DeleteMemory(id int) error {
	_, err := s.db.Exec(`DELETE FROM memories WHERE id = ?`, id)
	return err
}

// --- Operações de Providers ---

func (s *Store) SaveProviders(providers map[string]ProviderConfig) error {
	data, err := json.Marshal(providers)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)`, "providers", string(data))
	return err
}

func (s *Store) GetProviders() (map[string]ProviderConfig, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM config WHERE key = ?`, "providers").Scan(&value)
	if err == sql.ErrNoRows {
		return make(map[string]ProviderConfig), nil
	}
	if err != nil {
		return nil, err
	}
	var providers map[string]ProviderConfig
	if err := json.Unmarshal([]byte(value), &providers); err != nil {
		return nil, err
	}
	return providers, nil
}

// --- Utilitários de Vetor ---

func Float32ToByte(f []float32) []byte {
	buf := make([]byte, len(f)*4)
	for i, v := range f {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}

func ByteToFloat32(b []byte) []float32 {
	if len(b) == 0 {
		return nil
	}
	f := make([]float32, len(b)/4)
	for i := 0; i < len(f); i++ {
		f[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return f
}
