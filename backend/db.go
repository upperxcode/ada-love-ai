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
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"ada-love-ai/pkg/agent/interfaces"
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

	// Configurações de pool para SQLite (vários leitores, um escritor típico)
	db.SetMaxOpenConns(5)        // SQLite lida bem com poucas conexões
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	s := &Store{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}
func (s *Store) init() error {
	// PRAGMAs de performance (aplicados uma vez por conexão)
	if _, err := s.db.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA synchronous=NORMAL;
		PRAGMA cache_size=-32768;
		PRAGMA mmap_size=268435456;
		PRAGMA foreign_keys=ON;
		PRAGMA temp_store=MEMORY;
	`); err != nil {
		return fmt.Errorf("erro ao aplicar pragmas: %v", err)
	}

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
			workers TEXT, -- JSON
			skills TEXT, -- JSON
			tools TEXT -- JSON
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			workspace_path TEXT,
			worker_name TEXT DEFAULT '',
			parent_session_id TEXT DEFAULT '',
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
			time DATETIME DEFAULT CURRENT_TIMESTAMP,
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

	// Índices (idempotentes)
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_workspaces_slug ON workspaces(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_workspaces_enabled ON workspaces(enabled)`,
		`CREATE INDEX IF NOT EXISTS idx_workspaces_path ON workspaces(path)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_workspace_updated ON sessions(workspace_path, updated_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_worker ON sessions(worker_name)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_parent ON sessions(parent_session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id, time)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_time ON messages(time)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_workspace ON memories(workspace_path, importance DESC, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_importance ON memories(importance DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_config_key ON config(key)`,
		`CREATE INDEX IF NOT EXISTS idx_providers_name ON providers(name)`,
		`CREATE INDEX IF NOT EXISTS idx_skill_tags ON skill_tags(tag)`,
	}
	for _, idx := range indexes {
		if _, err := s.db.Exec(idx); err != nil {
			return fmt.Errorf("erro ao criar índice %q: %v", idx, err)
		}
	}

	// Migrações de esquema (idempotentes / seguras)
	s.db.Exec("ALTER TABLE sessions ADD COLUMN embedding BLOB")                // ignora se já existir
	s.db.Exec("ALTER TABLE memories ADD COLUMN embedding BLOB")                 // ignora se já existir
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN tools TEXT")                   // ignora se já existir
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN description TEXT")             // novo campo
	s.db.Exec("ALTER TABLE workspaces RENAME COLUMN agents TO workers")         // rename
	s.db.Exec("ALTER TABLE sessions ADD COLUMN worker_name TEXT DEFAULT ''")    // novo campo

	// Campos de workspace
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN enabled INTEGER DEFAULT 0")
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN color TEXT DEFAULT ''")
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN icon TEXT DEFAULT ''")
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN max_prompt_send INTEGER DEFAULT 0")
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN commit_changes INTEGER DEFAULT 1")
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN max_context_length INTEGER DEFAULT 0")
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN spec_wizard TEXT DEFAULT ''")
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN agents TEXT DEFAULT '[]'")
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN embedding_model TEXT DEFAULT ''")
	s.db.Exec("ALTER TABLE workspaces ADD COLUMN embedding_provider TEXT DEFAULT ''")

	// Hierarquia de sessões
	s.db.Exec("ALTER TABLE sessions ADD COLUMN parent_session_id TEXT DEFAULT ''")

	// Config per-chat
	s.db.Exec("ALTER TABLE sessions ADD COLUMN model TEXT DEFAULT ''")
	s.db.Exec("ALTER TABLE sessions ADD COLUMN provider TEXT DEFAULT ''")
	s.db.Exec("ALTER TABLE sessions ADD COLUMN mode TEXT DEFAULT 'ask'")
	s.db.Exec("ALTER TABLE sessions ADD COLUMN thinking TEXT DEFAULT ''")

	// Sumarização
	s.db.Exec("ALTER TABLE sessions ADD COLUMN summarized_context TEXT DEFAULT ''")
	s.db.Exec("ALTER TABLE sessions ADD COLUMN summarized_at DATETIME")
	s.db.Exec("ALTER TABLE sessions ADD COLUMN last_summarized_msg_id INTEGER DEFAULT 0")

	// Migração: sessões sem workspace_path → move para o workspace ativo
	s.db.Exec(`
		UPDATE sessions SET workspace_path = (
			SELECT path FROM workspaces ORDER BY id LIMIT 1
		)
		WHERE workspace_path = '' OR workspace_path IS NULL
	`)
	fmt.Printf("[DB] Init: migrated orphan sessions with empty workspace_path\n")

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// Exec executes a query without returning rows.
func (s *Store) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

// Query executes a query and returns rows.
func (s *Store) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(query, args...)
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
	workers, _ := json.Marshal(ws.Workers)
	skills, _ := json.Marshal(ws.Skills)
	tools, _ := json.Marshal(ws.Tools)
	agents, _ := json.Marshal(ws.Agents)

	fmt.Printf("[DB] SaveWorkspace: title=%q path=%q workers=%d\n", ws.Title, ws.Path, len(ws.Workers))
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO workspaces (title, description, path, personality, folders, knowledge, workers, skills, tools, enabled, color, icon, max_prompt_send, commit_changes, max_context_length, spec_wizard, agents, embedding_model, embedding_provider)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ws.Title, ws.Description, ws.Path, ws.Personality,
		string(folders), string(knowledge), string(workers), string(skills), string(tools),
		ws.Enabled, ws.Color, ws.Icon, ws.MaxPromptSend, ws.CommitChanges, ws.MaxContextLength, ws.SpecWizard,
		string(agents),
		ws.EmbeddingModel, ws.EmbeddingProvider,
	)
	return err
}

func (s *Store) GetWorkspaces() ([]WorkspaceConfig, error) {
	fmt.Printf("[DB] GetWorkspaces: querying all workspaces\n")
	rows, err := s.db.Query(`SELECT title, description, path, personality, folders, knowledge, workers, skills, tools, enabled, color, icon, max_prompt_send, commit_changes, max_context_length, spec_wizard, agents, embedding_model, embedding_provider FROM workspaces`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []WorkspaceConfig
	for rows.Next() {
		var ws WorkspaceConfig
		var folders, knowledge, workers, skills, tools, agents string
		var enabled sql.NullBool
		var commitChanges sql.NullBool
		var maxPromptSend, maxContextLength sql.NullInt64
		var color, icon, specWizard, embeddingModel, embeddingProvider sql.NullString
		err := rows.Scan(&ws.Title, &ws.Description, &ws.Path, &ws.Personality, &folders, &knowledge, &workers, &skills, &tools,
			&enabled, &color, &icon, &maxPromptSend, &commitChanges, &maxContextLength, &specWizard, &agents,
			&embeddingModel, &embeddingProvider)
		if err != nil {
			return nil, err
		}
		ws.Enabled = !enabled.Valid || enabled.Bool
		ws.CommitChanges = !commitChanges.Valid || commitChanges.Bool
		ws.MaxPromptSend = int(maxPromptSend.Int64)
		ws.MaxContextLength = int(maxContextLength.Int64)
		if color.Valid {
			ws.Color = color.String
		}
		if icon.Valid {
			ws.Icon = icon.String
		}
		if specWizard.Valid {
			ws.SpecWizard = specWizard.String
		}
		ws.EmbeddingModel = embeddingModel.String
		ws.EmbeddingProvider = embeddingProvider.String
		json.Unmarshal([]byte(folders), &ws.Folders)
		json.Unmarshal([]byte(knowledge), &ws.Knowledge)
		json.Unmarshal([]byte(workers), &ws.Workers)
		json.Unmarshal([]byte(skills), &ws.Skills)
		json.Unmarshal([]byte(tools), &ws.Tools)
		json.Unmarshal([]byte(agents), &ws.Agents)
		// Garantir que path nunca seja vazio
		if ws.Path == "" {
			ws.Path = strings.ToLower(strings.ReplaceAll(ws.Title, " ", "_"))
			s.db.Exec(`UPDATE workspaces SET path = ? WHERE title = ? AND (path = '' OR path IS NULL)`, ws.Path, ws.Title)
		}
		fmt.Printf("[DB] GetWorkspaces: title=%q path=%q workers=%d\n", ws.Title, ws.Path, len(ws.Workers))
		workspaces = append(workspaces, ws)
	}
	fmt.Printf("[DB] GetWorkspaces: total %d workspaces found\n", len(workspaces))
	return workspaces, nil
}

func (s *Store) DeleteWorkspace(path string) error {
	_, err := s.db.Exec(`DELETE FROM workspaces WHERE path = ?`, path)
	return err
}

// --- Operações de Sessão ---

func (s *Store) SaveSession(sess ChatSession) error {
	fmt.Printf("[DB] SaveSession: id=%q workspace=%q worker=%q parent=%q title=%q messages=%d pinned=%v model=%q mode=%q\n",
		sess.ID, sess.WorkspaceID, sess.WorkerName, sess.ParentSessionID, sess.Title, len(sess.Messages), sess.Pinned, sess.Model, sess.Mode)
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO sessions (id, workspace_path, worker_name, parent_session_id, title, summary, model, provider, mode, thinking, summarized_context, summarized_at, last_summarized_msg_id, pinned, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.WorkspaceID, sess.WorkerName, sess.ParentSessionID, sess.Title, sess.Summary,
		sess.Model, sess.Provider, sess.Mode, sess.Thinking,
		sess.SummarizedContext, sess.SummarizedAt, sess.LastSummarizedMsgID,
		sess.Pinned, sess.CreatedAt, sess.UpdatedAt,
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

	for i, msg := range sess.Messages {
		result, err := s.db.Exec(`INSERT INTO messages (session_id, role, content, time) VALUES (?, ?, ?, ?)`,
			sess.ID, msg.Role, msg.Content, msg.Time)
		if err != nil {
			log.Printf("Erro ao salvar mensagem: %v", err)
			continue
		}
		// Captura o ID gerado
		if id, err := result.LastInsertId(); err == nil {
			sess.Messages[i].ID = id
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
	fmt.Printf("[DB] GetSessions: workspacePath=%q\n", workspacePath)
	rows, err := s.db.Query(`
		SELECT id, workspace_path, worker_name, parent_session_id, title, summary, model, provider, mode, thinking, summarized_context, summarized_at, last_summarized_msg_id, pinned, created_at, updated_at 
		FROM sessions WHERE workspace_path = ?
		ORDER BY pinned DESC, updated_at DESC`, workspacePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*ChatSession
	for rows.Next() {
		sess := &ChatSession{}
		var summarizedAt sql.NullTime
		err := rows.Scan(&sess.ID, &sess.WorkspaceID, &sess.WorkerName, &sess.ParentSessionID, &sess.Title, &sess.Summary,
			&sess.Model, &sess.Provider, &sess.Mode, &sess.Thinking,
			&sess.SummarizedContext, &summarizedAt, &sess.LastSummarizedMsgID,
			&sess.Pinned, &sess.CreatedAt, &sess.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if summarizedAt.Valid {
			sess.SummarizedAt = summarizedAt.Time
		}
		
		// Carrega mensagens para esta sessão
		msgRows, err := s.db.Query(`SELECT id, role, content, time FROM messages WHERE session_id = ? ORDER BY time ASC`, sess.ID)
		if err == nil {
			for msgRows.Next() {
				var msg ChatMessage
				msgRows.Scan(&msg.ID, &msg.Role, &msg.Content, &msg.Time)
				sess.Messages = append(sess.Messages, msg)
			}
			msgRows.Close()
		}
		
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

func (s *Store) GetSession(id string) (*ChatSession, error) {
	sess := &ChatSession{}
	var summarizedAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT id, workspace_path, worker_name, parent_session_id, title, summary, model, provider, mode, thinking, summarized_context, summarized_at, last_summarized_msg_id, pinned, created_at, updated_at 
		FROM sessions WHERE id = ?`, id).Scan(
		&sess.ID, &sess.WorkspaceID, &sess.WorkerName, &sess.ParentSessionID, &sess.Title, &sess.Summary,
		&sess.Model, &sess.Provider, &sess.Mode, &sess.Thinking,
		&sess.SummarizedContext, &summarizedAt, &sess.LastSummarizedMsgID,
		&sess.Pinned, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if summarizedAt.Valid {
		sess.SummarizedAt = summarizedAt.Time
	}
	msgRows, err := s.db.Query(`SELECT id, role, content, time FROM messages WHERE session_id = ? ORDER BY time ASC`, id)
	if err == nil {
		for msgRows.Next() {
			var msg ChatMessage
			msgRows.Scan(&msg.ID, &msg.Role, &msg.Content, &msg.Time)
			sess.Messages = append(sess.Messages, msg)
		}
		msgRows.Close()
	}
	return sess, nil
}

func (s *Store) DeleteSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// --- Operações de Memória ---

func (s *Store) SaveMemory(workspacePath string, content string, importance int) error {
	now := time.Now()
	emb := Float32ToByte(nil) // placeholder - mantido compatível; a engine preenche embedding separadamente
	_, err := s.db.Exec(`
		INSERT INTO memories (workspace_path, content, importance, embedding, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		workspacePath, content, importance, emb, now, now)
	return err
}

func (s *Store) GetMemories(workspacePath string) ([]interfaces.MemoryEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, content, importance, created_at
		FROM memories
		WHERE workspace_path = ?
		ORDER BY importance DESC, created_at DESC`, workspacePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []interfaces.MemoryEntry
	for rows.Next() {
		var m interfaces.MemoryEntry
		var id, importance int
		var content string
		var createdAt time.Time
		if err := rows.Scan(&id, &content, &importance, &createdAt); err != nil {
			continue
		}
		m.Content = content
		m.Importance = importance
		m.CreatedAt = createdAt
		memories = append(memories, m)
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

// SaveDBProvider saves or updates a single provider in the DB.
// Reads the full map, upserts the entry, and writes it back.
func (s *Store) SaveDBProvider(name string, cfg ProviderConfig) error {
	providers, err := s.GetProviders()
	if err != nil {
		return err
	}
	if providers == nil {
		providers = make(map[string]ProviderConfig)
	}
	providers[name] = cfg
	return s.SaveProviders(providers)
}

// DeleteDBProvider removes a single provider from the DB.
func (s *Store) DeleteDBProvider(name string) error {
	providers, err := s.GetProviders()
	if err != nil {
		return err
	}
	delete(providers, name)
	return s.SaveProviders(providers)
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
