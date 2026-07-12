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
	db.SetMaxOpenConns(5) // SQLite lida bem com poucas conexões
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	s := &Store{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}
func (s *Store) init() error {
	// PRAGMAs de performance e integridade relacional
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

	// Fresh start: drop todas as tabelas e recria normalizado.
	s.db.Exec(`
		PRAGMA foreign_keys=ON;
	`)

	// Ensure workspace columns exist for extended fields
	// Attempt to add new columns with ALTER TABLE; if they already exist SQLite will error — we ignore errors.
	// This is a best-effort migration for older DBs.
	s.db.Exec(`ALTER TABLE workspaces ADD COLUMN summary TEXT DEFAULT ''`)
	s.db.Exec(`ALTER TABLE workspaces ADD COLUMN enabled BOOL NOT NULL DEFAULT 1`)
	s.db.Exec(`ALTER TABLE workspaces ADD COLUMN max_prompt_send INTEGER NOT NULL DEFAULT 0`)
	s.db.Exec(`ALTER TABLE workspaces ADD COLUMN commit_changes BOOL NOT NULL DEFAULT 0`)
	s.db.Exec(`ALTER TABLE workspaces ADD COLUMN max_context_length INTEGER NOT NULL DEFAULT 0`)
	s.db.Exec(`ALTER TABLE workspaces ADD COLUMN embedding_model TEXT DEFAULT ''`)
	s.db.Exec(`ALTER TABLE workspaces ADD COLUMN embedding_provider TEXT DEFAULT ''`)
	s.db.Exec(`ALTER TABLE workspaces ADD COLUMN routing_rules TEXT DEFAULT ''`)

	queries := []string{ // core tables

		`CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nome TEXT NOT NULL,
			description TEXT,
			path TEXT UNIQUE,
			max_prompt INTEGER NOT NULL DEFAULT 4096,
			max_content INTEGER NOT NULL DEFAULT 8192,
			"commit" BOOL NOT NULL DEFAULT 1,
			spec_provider TEXT,
			spec_wizard_id TEXT REFERENCES spec_wizards(id) ON DELETE SET NULL,
			personality TEXT,
			routing_rules TEXT,
			color TEXT DEFAULT '',
			icon TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS workers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			persona TEXT,
			response_language TEXT DEFAULT 'portuguese',
			connection_type TEXT NOT NULL,
			command TEXT,
			arguments TEXT,
			environment TEXT,
			inheritance_folders BOOL NOT NULL DEFAULT 0,
			inheritance_skills BOOL NOT NULL DEFAULT 0,
			inheritance_persona BOOL NOT NULL DEFAULT 0,
			inheritance_knowledge BOOL NOT NULL DEFAULT 0,
			inheritance_tools BOOL NOT NULL DEFAULT 0,
			color TEXT DEFAULT '',
			icon TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS agents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL CHECK(type IN ('executor','delegator','reviewer','research')),
			provider_id INTEGER,
			model_id INTEGER,
			max_iteration INTEGER NOT NULL DEFAULT 10,
			temperature REAL NOT NULL DEFAULT 0.7,
			system_prompt TEXT,
			color TEXT DEFAULT '',
			icon TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS skills (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			tags TEXT,
			content TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS spec_wizards (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			expert_language_plugin TEXT,
			prd TEXT,
			functional_requirements TEXT,
			non_functional_requirements TEXT,
			persistence TEXT,
			architecture TEXT,
			engineering_philosophies TEXT,
			design_patterns TEXT,
			data_patterns TEXT,
			stack_config TEXT,
			business_state_management TEXT,
			business_api_contract TEXT,
			business_customization_details TEXT,
			business_final_adjustments TEXT,
			business_architecture_recommendations TEXT,
			color TEXT DEFAULT '#3b82f6',
			icon TEXT DEFAULT '📝',
			architecture_health INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS mcps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nome TEXT NOT NULL,
			connect_type TEXT NOT NULL CHECK(connect_type IN ('websocket','url','cli_command')),
			command TEXT,
			arguments TEXT,
			environment TEXT,
			color TEXT DEFAULT '',
			icon TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS providers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			api_url TEXT,
			connection_types TEXT,
			color TEXT DEFAULT '',
			icon TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS provider_apikeys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			provider_id INTEGER NOT NULL,
			apikey TEXT NOT NULL,
			UNIQUE(provider_id, apikey),
			FOREIGN KEY(provider_id) REFERENCES providers(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS provider_models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			provider_id INTEGER NOT NULL,
			model TEXT NOT NULL,
			free BOOL NOT NULL DEFAULT 0,
			thinking BOOL NOT NULL DEFAULT 0,
			tool BOOL NOT NULL DEFAULT 0,
			embedding BOOL NOT NULL DEFAULT 0,
			vision BOOL NOT NULL DEFAULT 0,
			health INTEGER NOT NULL DEFAULT 100,
			UNIQUE(provider_id, model),
			FOREIGN KEY(provider_id) REFERENCES providers(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS fixedmodels (
			embedding_provider TEXT NOT NULL,
			embedding_model TEXT NOT NULL,
			image_provider TEXT NOT NULL,
			image_model TEXT NOT NULL,
			spec_provider TEXT NOT NULL,
			spec_model TEXT NOT NULL,
			PRIMARY KEY (embedding_provider, embedding_model, image_provider, image_model, spec_provider, spec_model)
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_workers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			worker_id INTEGER NOT NULL,
			enabled BOOL NOT NULL DEFAULT 1,
			FOREIGN KEY(workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
			FOREIGN KEY(worker_id) REFERENCES workers(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_agents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			agent_id INTEGER NOT NULL,
			enabled BOOL NOT NULL DEFAULT 1,
			FOREIGN KEY(workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
			FOREIGN KEY(agent_id) REFERENCES agents(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_skills (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			skill_id INTEGER NOT NULL,
			enabled BOOL NOT NULL DEFAULT 1,
			FOREIGN KEY(workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
			FOREIGN KEY(skill_id) REFERENCES skills(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_tools (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			tool_name TEXT NOT NULL,
			enabled BOOL NOT NULL DEFAULT 1,
			FOREIGN KEY(workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_folders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			folder_path TEXT NOT NULL,
			FOREIGN KEY(workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_knowledge (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			knowledge_item TEXT NOT NULL,
			FOREIGN KEY(workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			workspace_path TEXT,
			title TEXT,
			pinned INTEGER DEFAULT 0,
			embedding BLOB,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			worker_name TEXT DEFAULT '',
			parent_session_id TEXT DEFAULT '',
			model TEXT DEFAULT '',
			provider TEXT DEFAULT '',
			mode TEXT DEFAULT 'ask',
			thinking TEXT DEFAULT '',
			summary TEXT,
			summarized_context TEXT DEFAULT '',
			summary_token_count INTEGER DEFAULT 0,
			summarized_at DATETIME,
			last_summarized_msg_id INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('user','assistant','system','tool')),
			content TEXT NOT NULL,
			tokens INTEGER DEFAULT 0,
			time DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS memories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_path TEXT,
			content TEXT NOT NULL,
			importance INTEGER DEFAULT 1,
			embedding BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS spec_wizards (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			expert_language_plugin TEXT,
			prd TEXT,
			functional_requirements TEXT,
			non_functional_requirements TEXT,
			persistence TEXT,
			architecture TEXT,
			engineering_philosophies TEXT,
			design_patterns TEXT,
			data_patterns TEXT,
			stack_config TEXT,
			business_state_management TEXT,
			business_api_contract TEXT,
			business_customization_details TEXT,
			business_final_adjustments TEXT,
			business_architecture_recommendations TEXT,
			color TEXT DEFAULT '#3b82f6',
			icon TEXT DEFAULT '📝',
			architecture_health INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_templates (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE,
				description TEXT,
				personality TEXT NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
		`CREATE TABLE IF NOT EXISTS tool_profiles (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE,
				description TEXT,
				color TEXT DEFAULT '',
				icon TEXT DEFAULT ''
			)`,
		`CREATE TABLE IF NOT EXISTS tool_profile_tools (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				profile_id INTEGER NOT NULL REFERENCES tool_profiles(id) ON DELETE CASCADE,
				tool_name TEXT NOT NULL
			)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("erro ao criar tabela: %v\nQuery: %s", err, q)
		}
	}

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_workspaces_nome ON workspaces(nome)`,
		`CREATE INDEX IF NOT EXISTS idx_workspaces_path ON workspaces(path)`,
		`CREATE INDEX IF NOT EXISTS idx_workers_name ON workers(name)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_name ON agents(name)`,
		`CREATE INDEX IF NOT EXISTS idx_skills_name ON skills(name)`,
		`CREATE INDEX IF NOT EXISTS idx_mcps_nome ON mcps(nome)`,
		`CREATE INDEX IF NOT EXISTS idx_providers_name ON providers(name)`,
		`CREATE INDEX IF NOT EXISTS idx_provider_models_model ON provider_models(model)`,
		`CREATE INDEX IF NOT EXISTS idx_config_key ON config(key)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_workspace_updated ON sessions(workspace_path, updated_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_parent ON sessions(parent_session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id, time)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_time ON messages(time)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_workspace ON memories(workspace_path, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_spec_wizards_name ON spec_wizards(name)`,
		`CREATE TABLE IF NOT EXISTS fixed_models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			provider TEXT,
			model TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS fixed_model_tools (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				fixed_model_id INTEGER NOT NULL REFERENCES fixed_models(id) ON DELETE CASCADE,
				tool TEXT NOT NULL
			)`,
		`CREATE TABLE IF NOT EXISTS mcps (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				nome TEXT NOT NULL,
				connect_type TEXT NOT NULL,
				command TEXT,
				arguments TEXT,
				environment TEXT,
				color TEXT DEFAULT '',
				icon TEXT DEFAULT ''
			)`,
		`CREATE TABLE IF NOT EXISTS tool_profiles (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE,
				description TEXT,
				color TEXT DEFAULT '',
				icon TEXT DEFAULT ''
			)`,
		`CREATE TABLE IF NOT EXISTS tool_profile_tools (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				profile_id INTEGER NOT NULL REFERENCES tool_profiles(id) ON DELETE CASCADE,
				tool_name TEXT NOT NULL
			)`,
	}
	for _, idx := range indexes {
		if _, err := s.db.Exec(idx); err != nil {
			return fmt.Errorf("erro ao criar índice %q: %v", idx, err)
		}
	}

	// Add spec_wizard_id column to workspaces table (migration for existing DBs)
	s.db.Exec(`ALTER TABLE workspaces ADD COLUMN spec_wizard_id TEXT REFERENCES spec_wizards(id) ON DELETE SET NULL`)

	// Ensure migrations table exists to track applied migrations (versioning by integer)
	s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)

	// Read current schema version (max applied)
	var curVersion int
	if err := s.db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&curVersion); err != nil {
		// non-fatal: log and continue with version 0
		fmt.Printf("[DB] Warn: failed to read schema version: %v\n", err)
		curVersion = 0
	}

	// Helper: detect legacy config table and read JSON values from it
	var _cfgTbl string
	configExists := false
	if err := s.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='config'").Scan(&_cfgTbl); err == nil {
		configExists = true
	}

	readConfigString := func(key string) (string, bool) {
		if !configExists {
			return "", false
		}
		var raw sql.NullString
		if err := s.db.QueryRow(`SELECT value FROM config WHERE key = ?`, key).Scan(&raw); err != nil {
			return "", false
		}
		if !raw.Valid {
			return "", false
		}
		// Try to unmarshal JSON string value into a plain string
		var out string
		if err := json.Unmarshal([]byte(raw.String), &out); err == nil {
			return out, true
		}
		// Fallback: return raw as-is
		return raw.String, true
	}

	readConfigStringArray := func(key string) ([]string, bool) {
		if !configExists {
			return nil, false
		}
		var raw sql.NullString
		if err := s.db.QueryRow(`SELECT value FROM config WHERE key = ?`, key).Scan(&raw); err != nil {
			return nil, false
		}
		if !raw.Valid {
			return nil, false
		}
		var out []string
		if err := json.Unmarshal([]byte(raw.String), &out); err == nil {
			return out, true
		}
		return nil, false
	}

	// Migration 1: move legacy global_config entries into fixed_models rows + legacy fixedmodels table migration
	const migrationV1 = 1
	if curVersion < migrationV1 {
		// Perform migration (kept similar to previous behavior). If this fails we abort and do not record the version.
		migrationPerformed := false

		// 1) Spec
		if specModel, ok := readConfigString("spec_model"); ok && specModel != "" {
			migrationPerformed = true
			specProvider, _ := readConfigString("spec_provider")
			// Insert or update fixed_models row for 'spec'
			s.db.Exec(`INSERT OR REPLACE INTO fixed_models (name, provider, model) VALUES ('spec', ?, ?)`, specProvider, specModel)
			// Persist tools if present
			if specTools, ok2 := readConfigStringArray("spec_tools"); ok2 {
				var id int64
				if err := s.db.QueryRow(`SELECT id FROM fixed_models WHERE name = 'spec'`).Scan(&id); err == nil {
					// clear current tools and insert
					s.db.Exec(`DELETE FROM fixed_model_tools WHERE fixed_model_id = ?`, id)
					for _, t := range specTools {
						s.db.Exec(`INSERT INTO fixed_model_tools (fixed_model_id, tool) VALUES (?, ?)`, id, t)
					}
				}
			}
		}

		// 2) TinyBrain
		if tinyRaw, ok := readConfigString("tiny_brain"); ok {
			// tiny_brain may be stored as JSON object; try to unmarshal into struct
			var tb struct {
				ModelName string   `json:"model_name"`
				Provider  string   `json:"provider"`
				Tools     []string `json:"tools"`
			}
			if err := json.Unmarshal([]byte(tinyRaw), &tb); err == nil {
				if tb.ModelName != "" {
					migrationPerformed = true
					s.db.Exec(`INSERT OR REPLACE INTO fixed_models (name, provider, model) VALUES ('tinybrain', ?, ?)`, tb.Provider, tb.ModelName)
					var id int64
					if err := s.db.QueryRow(`SELECT id FROM fixed_models WHERE name = 'tinybrain'`).Scan(&id); err == nil {
						s.db.Exec(`DELETE FROM fixed_model_tools WHERE fixed_model_id = ?`, id)
						for _, t := range tb.Tools {
							s.db.Exec(`INSERT INTO fixed_model_tools (fixed_model_id, tool) VALUES (?, ?)`, id, t)
						}
					}
				}
			}
		}

		// 3) Embedding & Image
		if embeddingModel, ok := readConfigString("embedding_model"); ok && embeddingModel != "" {
			migrationPerformed = true
			embeddingProvider, _ := readConfigString("embedding_provider")
			s.db.Exec(`INSERT OR REPLACE INTO fixed_models (name, provider, model) VALUES ('embedding', ?, ?)`, embeddingProvider, embeddingModel)
		}
		if imageModel, ok := readConfigString("image_model"); ok && imageModel != "" {
			migrationPerformed = true
			imageProvider, _ := readConfigString("image_provider")
			s.db.Exec(`INSERT OR REPLACE INTO fixed_models (name, provider, model) VALUES ('image', ?, ?)`, imageProvider, imageModel)
		}

		// --- Migration: legacy fixedmodels table (compound table) → fixed_models rows ---
		var tblName string
		if err := s.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='fixedmodels'").Scan(&tblName); err == nil {
			// legacy table exists — inspect columns to confirm legacy schema
			rows, err := s.db.Query("PRAGMA table_info('fixedmodels')")
			if err == nil {
				cols := map[string]bool{}
				for rows.Next() {
					var cid int
					var name, ctype string
					var notnull, pk int
					var dflt sql.NullString
					_ = rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk)
					cols[name] = true
				}
				rows.Close()
				// If legacy columns present, rename old table to a backup and migrate
				if cols["embedding_model"] || cols["spec_model"] || cols["image_model"] {
					backup := fmt.Sprintf("fixedmodels_backup_%d", time.Now().Unix())
					if _, err := s.db.Exec(fmt.Sprintf("ALTER TABLE fixedmodels RENAME TO %s", backup)); err != nil {
						fmt.Printf("[DB] Warn: failed to rename legacy fixedmodels table: %v\n", err)
					} else {
						// Read from backup and insert rows into fixed_models
						var ep, em, ip, im, sp, sm sql.NullString
						row := s.db.QueryRow(fmt.Sprintf(`SELECT embedding_provider, embedding_model, image_provider, image_model, spec_provider, spec_model FROM %s LIMIT 1`, backup))
						if row != nil {
							if err := row.Scan(&ep, &em, &ip, &im, &sp, &sm); err == nil {
								if em.Valid && em.String != "" {
									s.db.Exec(`INSERT OR REPLACE INTO fixed_models (name, provider, model) VALUES ('embedding', ?, ?)`, ep.String, em.String)
								}
								if im.Valid && im.String != "" {
									s.db.Exec(`INSERT OR REPLACE INTO fixed_models (name, provider, model) VALUES ('image', ?, ?)`, ip.String, im.String)
								}
								if sm.Valid && sm.String != "" {
									s.db.Exec(`INSERT OR REPLACE INTO fixed_models (name, provider, model) VALUES ('spec', ?, ?)`, sp.String, sm.String)
								}
							}
						}
						fmt.Printf("[DB] Legacy fixedmodels table renamed to %s and migrated to fixed_models\n", backup)
					}
				}
			}
		}

		if migrationPerformed {
			if _, err := s.db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, migrationV1); err != nil {
				fmt.Printf("[DB] Warn: failed to record migration %d: %v\n", migrationV1, err)
			} else {
				fmt.Printf("[DB] Migration %d applied\n", migrationV1)
			}
		}
	}

	// Migration 2: move tool_profiles and mcp_servers from config (key/value) into normalized tables
	const migrationV2 = 2
	if curVersion < migrationV2 {
		moved := false
		// tool_profiles
		if tpsRaw, ok := readConfigString("tool_profiles"); ok {
			var tps []ToolProfile
			if err := json.Unmarshal([]byte(tpsRaw), &tps); err == nil && len(tps) > 0 {
				if err := s.SaveToolProfiles(tps); err == nil {
					moved = true
				}
			}
		}
		// mcp_servers
		if mcpsRaw, ok := readConfigString("mcp_servers"); ok {
			var mcps map[string]MCPServerUI
			if err := json.Unmarshal([]byte(mcpsRaw), &mcps); err == nil {
				for name, m := range mcps {
					// serialize args and env as JSON strings
					argsJSON := ""
					if len(m.Args) > 0 {
						if b, err := json.Marshal(m.Args); err == nil {
							argsJSON = string(b)
						}
					}
					envJSON := ""
					if len(m.Env) > 0 {
						if b, err := json.Marshal(m.Env); err == nil {
							envJSON = string(b)
						}
					}
					if m.URL != "" {
						// merge URL into env under key __url
						var em map[string]string
						if envJSON != "" {
							json.Unmarshal([]byte(envJSON), &em)
						}
						if em == nil {
							em = map[string]string{}
						}
						em["__url"] = m.URL
						if b, err := json.Marshal(em); err == nil {
							envJSON = string(b)
						}
					}
					if _, err := s.SaveMCP(name, "", m.Command, argsJSON, envJSON, m.Color, m.Icon); err == nil {
						moved = true
					}
				}
			}
		}
		if moved {
			if _, err := s.db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, migrationV2); err != nil {
				fmt.Printf("[DB] Warn: failed to record migration %d: %v\n", migrationV2, err)
			} else {
				fmt.Printf("[DB] Migration %d applied\n", migrationV2)
			}
		}
	}

	fmt.Printf("[DB] Init: schema normalizado criado (fresh start)\n")
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

// SetGlobalConfig and GetGlobalConfig are legacy helpers kept for compatibility with
// older migration logic. New runtime code must use normalized tables (app_state, mcps,
// providers, tool_profiles, etc.) instead of the generic config key/value table.
// The legacy config table is removed from schema; these helpers now return errors to
// prevent accidental runtime use.
func (s *Store) SetGlobalConfig(key string, value interface{}) error {
	return fmt.Errorf("SetGlobalConfig is deprecated; use normalized tables instead (key=%s)", key)
}

func (s *Store) GetGlobalConfig(key string, target interface{}) (bool, error) {
	return false, fmt.Errorf("GetGlobalConfig is deprecated; use normalized tables instead (key=%s)", key)
}

// --- Fixed Models (embedding/image/spec/tinybrain) ---
// fixed_models: id, name, provider, model
// fixed_model_tools: id, fixed_model_id, tool

type FixedModel struct {
	ID       int64  `db:"id" json:"id"`
	Name     string `db:"name" json:"name"` // e.g. "embedding", "image", "spec", "tinybrain"
	Provider string `db:"provider" json:"provider"`
	Model    string `db:"model" json:"model"`
}

// --- New fixed_models row-based helpers (renamed to avoid conflict with legacy API) ---
func (s *Store) SaveFixedModelRow(f FixedModel) (int64, error) {
	if f.Name == "" {
		return 0, fmt.Errorf("fixed model name is required")
	}
	// Upsert by name
	_, err := s.db.Exec(`INSERT INTO fixed_models (name, provider, model) VALUES (?, ?, ?) ON CONFLICT(name) DO UPDATE SET provider=excluded.provider, model=excluded.model`, f.Name, f.Provider, f.Model)
	if err != nil {
		return 0, err
	}
	var id int64
	if err := s.db.QueryRow(`SELECT id FROM fixed_models WHERE name = ?`, f.Name).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) ListFixedModelRows() ([]FixedModel, error) {
		rows, err := s.db.Query(`SELECT id, name, provider, model FROM fixed_models`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		out := []FixedModel{}
		for rows.Next() {
			var f FixedModel
			if err := rows.Scan(&f.ID, &f.Name, &f.Provider, &f.Model); err != nil {
				return nil, err
			}
			out = append(out, f)
		}
		// Debug: print loaded fixed models for visibility during startup/migration
		for _, f := range out {
			fmt.Printf("[DB] fixed_model loaded: id=%d name=%q provider=%q model=%q\n", f.ID, f.Name, f.Provider, f.Model)
		}
		return out, nil
	}

func (s *Store) DeleteFixedModelRowByName(name string) error {
	var id int64
	if err := s.db.QueryRow(`SELECT id FROM fixed_models WHERE name = ?`, name).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM fixed_model_tools WHERE fixed_model_id = ?`, id); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM fixed_models WHERE id = ?`, id)
	return err
}

func (s *Store) SetFixedModelRowTools(fixedModelID int64, tools []string) error {
	if _, err := s.db.Exec(`DELETE FROM fixed_model_tools WHERE fixed_model_id = ?`, fixedModelID); err != nil {
		return err
	}
	for _, t := range tools {
		if _, err := s.db.Exec(`INSERT INTO fixed_model_tools (fixed_model_id, tool) VALUES (?, ?)`, fixedModelID, t); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetFixedModelRowTools(fixedModelID int64) ([]string, error) {
	rows, err := s.db.Query(`SELECT tool FROM fixed_model_tools WHERE fixed_model_id = ?`, fixedModelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// --- Operações de Workspace ---

func (s *Store) SaveWorkspace(ws WorkspaceConfig) (int64, error) {
	nome := ws.Title
	if nome == "" {
		nome = ws.Nome
	}
	if nome == "" {
		nome = "Sem nome"
	}
	path := ws.Path
	if path == "" {
		path = strings.ToLower(strings.ReplaceAll(nome, " ", "_"))
	}
	// Ensure spec_wizard_id references an existing spec_wizard, otherwise use NULL to avoid FK error
	specWizardID := ws.SpecWizardID
	if specWizardID != "" {
		var exists int
		err := s.db.QueryRow(`SELECT 1 FROM spec_wizards WHERE id = ? LIMIT 1`, specWizardID).Scan(&exists)
		if err != nil {
			// not found or error -> unset to avoid FK violation
			specWizardID = ""
		}
	}
	_, err := s.db.Exec(`
				INSERT OR REPLACE INTO workspaces (id, nome, description, path, max_prompt, max_content, "commit", spec_provider, spec_wizard_id, personality, routing_rules, color, icon, summary, enabled, max_prompt_send, commit_changes, max_context_length, embedding_model, embedding_provider)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ws.ID, nome, ws.Description, path, ws.MaxPrompt, ws.MaxContent, ws.Commit, ws.SpecProvider, specWizardID, ws.Personality, ws.RoutingRules, ws.Color, ws.Icon,
		ws.Summary, ws.Enabled, ws.MaxPromptSend, ws.CommitChanges, ws.MaxContextLength, ws.EmbeddingModel, ws.EmbeddingProvider)
	if err != nil {
		return 0, err
	}

	var id int64
	if ws.ID > 0 {
		id = int64(ws.ID)
	} else {
		if rid, err2 := s.db.Exec(`SELECT id FROM workspaces WHERE path = ?`, path); err2 == nil {
			_ = rid
		}
	}
	// Recupera o id gerado
	s.db.QueryRow(`SELECT id FROM workspaces WHERE path = ?`, path).Scan(&id)
	return id, nil
}

func (s *Store) GetWorkspaces() ([]WorkspaceConfig, error) {
	rows, err := s.db.Query(`SELECT id, nome, description, path, max_prompt, max_content, "commit", spec_provider, spec_wizard_id, personality, routing_rules, color, icon, summary, enabled, max_prompt_send, commit_changes, max_context_length, embedding_model, embedding_provider FROM workspaces`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []WorkspaceConfig
	for rows.Next() {
		var ws WorkspaceConfig
		var nome, description, path, specProvider, specWizardID, personality, routingRules, color, icon, summary, embModel, embProvider sql.NullString
		var maxPrompt, maxContent, maxPromptSend, maxContextLen sql.NullInt64
		var commit, enabled sql.NullBool
		if err := rows.Scan(&ws.ID, &nome, &description, &path, &maxPrompt, &maxContent, &commit, &specProvider, &specWizardID, &personality, &routingRules, &color, &icon, &summary, &enabled, &maxPromptSend, &commit, &maxContextLen, &embModel, &embProvider); err != nil {
			return nil, err
		}
		ws.Title = nome.String
		ws.Nome = nome.String
		ws.Description = description.String
		ws.Path = path.String
		ws.MaxPrompt = int(maxPrompt.Int64)
		ws.MaxContent = int(maxContent.Int64)
		ws.Commit = !commit.Valid || commit.Bool
		ws.SpecProvider = specProvider.String
		ws.SpecWizardID = specWizardID.String
		ws.Personality = personality.String
		ws.RoutingRules = routingRules.String
		ws.Color = color.String
		ws.Icon = icon.String
		ws.Summary = summary.String
		ws.Enabled = !enabled.Valid || enabled.Bool
		ws.MaxPromptSend = int(maxPromptSend.Int64)
		ws.CommitChanges = !commit.Valid || commit.Bool
		ws.MaxContextLength = int(maxContextLen.Int64)
		ws.EmbeddingModel = embModel.String
		ws.EmbeddingProvider = embProvider.String
		// Resolve nomes de workers/agents via junction
		ws.WorkerNames = s.GetWorkspaceWorkerNames(ws.ID)
		ws.Agents = s.GetWorkspaceAgentNames(ws.ID)
		ws.Folders = s.GetWorkspaceFolders(ws.ID)
		ws.Knowledge = s.GetWorkspaceKnowledge(ws.ID)
		// Load skills (IDs -> names)
		// GetWorkspaceSkillIDs not implemented, use direct query
		rowsSkills, err := s.db.Query(`SELECT s.name FROM skills s JOIN workspace_skills ws ON s.id = ws.skill_id WHERE ws.workspace_id = ? AND ws.enabled = 1`, ws.ID)
		if err == nil {
			for rowsSkills.Next() {
				var sname string
				rowsSkills.Scan(&sname)
				ws.Skills = append(ws.Skills, sname)
			}
			rowsSkills.Close()
		}
		// Load tools
		ws.Tools = s.GetWorkspaceToolsByID(ws.ID)
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

func (s *Store) GetWorkspaceByPath(path string) (*WorkspaceConfig, error) {
	rows, err := s.db.Query(`SELECT id, nome, description, path, max_prompt, max_content, "commit", spec_provider, spec_wizard_id, personality, routing_rules, color, icon, summary, enabled, max_prompt_send, commit_changes, max_context_length, embedding_model, embedding_provider FROM workspaces WHERE path = ?`, path)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var ws WorkspaceConfig
	var nome, description, p, specProvider, specWizardID, personality, routingRules, color, icon, summary, embModel, embProvider sql.NullString
	var maxPrompt, maxContent, maxPromptSend, maxContextLen sql.NullInt64
	var commit, enabled sql.NullBool
	if err := rows.Scan(&ws.ID, &nome, &description, &p, &maxPrompt, &maxContent, &commit, &specProvider, &specWizardID, &personality, &routingRules, &color, &icon, &summary, &enabled, &maxPromptSend, &commit, &maxContextLen, &embModel, &embProvider); err != nil {
		return nil, err
	}
	ws.Title = nome.String
	ws.Nome = nome.String
	ws.Description = description.String
	ws.Path = p.String
	ws.MaxPrompt = int(maxPrompt.Int64)
	ws.MaxContent = int(maxContent.Int64)
	ws.Commit = !commit.Valid || commit.Bool
	ws.SpecProvider = specProvider.String
	ws.SpecWizardID = specWizardID.String
	ws.Personality = personality.String
	ws.RoutingRules = routingRules.String
	ws.Color = color.String
	ws.Icon = icon.String
	ws.Summary = summary.String
	ws.Enabled = !enabled.Valid || enabled.Bool
	ws.MaxPromptSend = int(maxPromptSend.Int64)
	ws.CommitChanges = !commit.Valid || commit.Bool
	ws.MaxContextLength = int(maxContextLen.Int64)
	ws.EmbeddingModel = embModel.String
	ws.EmbeddingProvider = embProvider.String
	ws.WorkerNames = s.GetWorkspaceWorkerNames(ws.ID)
	ws.Agents = s.GetWorkspaceAgentNames(ws.ID)
	ws.Folders = s.GetWorkspaceFolders(ws.ID)
	ws.Knowledge = s.GetWorkspaceKnowledge(ws.ID)
	// Load skills
	rowsSkills, err := s.db.Query(`SELECT s.name FROM skills s JOIN workspace_skills ws ON s.id = ws.skill_id WHERE ws.workspace_id = ? AND ws.enabled = 1`, ws.ID)
	if err == nil {
		for rowsSkills.Next() {
			var sname string
			rowsSkills.Scan(&sname)
			ws.Skills = append(ws.Skills, sname)
		}
		rowsSkills.Close()
	}
	// Load tools
	ws.Tools = s.GetWorkspaceToolsByID(ws.ID)
	return &ws, nil
}

func (s *Store) DeleteWorkspaceByPath(path string) error {
	_, err := s.db.Exec(`DELETE FROM workspaces WHERE path = ?`, path)
	return err
}

// --- Operações de Worker ---

func (s *Store) SaveWorker(w WorkerConfig) (int64, error) {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO workers (id, name, persona, response_language, connection_type, command, arguments, environment,
			inheritance_folders, inheritance_skills, inheritance_persona, inheritance_knowledge, inheritance_tools, color, icon)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		w.ID, w.Name, w.Persona, w.ResponseLanguage, w.ConnectionType, w.Command, w.Arguments, w.Environment,
		w.InheritFolders, w.InheritSkills, w.InheritPersona, w.InheritKnowledge, w.InheritTools, w.Color, w.Icon)
	if err != nil {
		return 0, err
	}
	var id int64
	s.db.QueryRow(`SELECT id FROM workers WHERE name = ?`, w.Name).Scan(&id)
	return id, nil
}

func (s *Store) GetWorkers() ([]WorkerConfig, error) {
	rows, err := s.db.Query(`SELECT id, name, persona, response_language, connection_type, command, arguments, environment,
		inheritance_folders, inheritance_skills, inheritance_persona, inheritance_knowledge, inheritance_tools, color, icon FROM workers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WorkerConfig
	for rows.Next() {
		var w WorkerConfig
		var name, persona, lang, ctype, cmd, args, env, color, icon sql.NullString
		var inhF, inhS, inhP, inhK, inhT sql.NullBool
		if err := rows.Scan(&w.ID, &name, &persona, &lang, &ctype, &cmd, &args, &env,
			&inhF, &inhS, &inhP, &inhK, &inhT, &color, &icon); err != nil {
			return nil, err
		}
		w.Name = name.String
		w.Persona = persona.String
		w.ResponseLanguage = lang.String
		w.ConnectionType = ctype.String
		w.Command = cmd.String
		w.Arguments = args.String
		w.Environment = env.String
		w.InheritFolders = inhF.Bool
		w.InheritSkills = inhS.Bool
		w.InheritPersona = inhP.Bool
		w.InheritKnowledge = inhK.Bool
		w.InheritTools = inhT.Bool
		w.Color = color.String
		w.Icon = icon.String
		out = append(out, w)
	}
	return out, nil
}

func (s *Store) GetWorkerByName(name string) (*WorkerConfig, error) {
	row := s.db.QueryRow(`SELECT id, name, persona, response_language, connection_type, command, arguments, environment,
		inheritance_folders, inheritance_skills, inheritance_persona, inheritance_knowledge, inheritance_tools, color, icon
		FROM workers WHERE name = ?`, name)
	var w WorkerConfig
	var n, persona, lang, ctype, cmd, args, env, color, icon sql.NullString
	var inhF, inhS, inhP, inhK, inhT sql.NullBool
	err := row.Scan(&w.ID, &n, &persona, &lang, &ctype, &cmd, &args, &env,
		&inhF, &inhS, &inhP, &inhK, &inhT, &color, &icon)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	w.Name = n.String
	w.Persona = persona.String
	w.ResponseLanguage = lang.String
	w.ConnectionType = ctype.String
	w.Command = cmd.String
	w.Arguments = args.String
	w.Environment = env.String
	w.InheritFolders = inhF.Bool
	w.InheritSkills = inhS.Bool
	w.InheritPersona = inhP.Bool
	w.InheritKnowledge = inhK.Bool
	w.InheritTools = inhT.Bool
	w.Color = color.String
	w.Icon = icon.String
	return &w, nil
}

func (s *Store) DeleteWorker(id int64) error {
	_, err := s.db.Exec(`DELETE FROM workers WHERE id = ?`, id)
	return err
}

// --- Operações de Agent ---

func (s *Store) SaveAgent(a AgentConfig) (int64, error) {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO agents (id, name, description, type, provider_id, model_id, max_iteration, temperature, system_prompt, color, icon)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.Name, a.Description, a.Type, a.ProviderID, a.ModelID, a.MaxIterations, a.Temperature, a.SystemPrompt, a.Color, a.Icon)
	if err != nil {
		return 0, err
	}
	var id int64
	s.db.QueryRow(`SELECT id FROM agents WHERE name = ?`, a.Name).Scan(&id)
	return id, nil
}

func (s *Store) GetAgents() ([]AgentConfig, error) {
	rows, err := s.db.Query(`SELECT id, name, description, type, provider_id, model_id, max_iteration, temperature, system_prompt, color, icon FROM agents`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AgentConfig
	for rows.Next() {
		var a AgentConfig
		var name, desc, atype, sp, color, icon sql.NullString
		var pid, mid, maxIter sql.NullInt64
		var temp sql.NullFloat64
		if err := rows.Scan(&a.ID, &name, &desc, &atype, &pid, &mid, &maxIter, &temp, &sp, &color, &icon); err != nil {
			return nil, err
		}
		a.Name = name.String
		a.Description = desc.String
		a.Type = atype.String
		a.ProviderID = pid.Int64
		a.ModelID = mid.Int64
		a.MaxIterations = int(maxIter.Int64)
		a.Temperature = temp.Float64
		a.SystemPrompt = sp.String
		a.Color = color.String
		a.Icon = icon.String
		out = append(out, a)
	}
	return out, nil
}

func (s *Store) GetAgentByName(name string) (*AgentConfig, error) {
	row := s.db.QueryRow(`SELECT id, name, description, type, provider_id, model_id, max_iteration, temperature, system_prompt, color, icon FROM agents WHERE name = ?`, name)
	var a AgentConfig
	var nameS, desc, atype, sp, color, icon sql.NullString
	var pid, mid, maxIter sql.NullInt64
	var temp sql.NullFloat64
	err := row.Scan(&a.ID, &nameS, &desc, &atype, &pid, &mid, &maxIter, &temp, &sp, &color, &icon)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.Name = nameS.String
	a.Description = desc.String
	a.Type = atype.String
	a.ProviderID = pid.Int64
	a.ModelID = mid.Int64
	a.MaxIterations = int(maxIter.Int64)
	a.Temperature = temp.Float64
	a.SystemPrompt = sp.String
	a.Color = color.String
	a.Icon = icon.String
	return &a, nil
}

func (s *Store) DeleteAgent(id int64) error {
	_, err := s.db.Exec(`DELETE FROM agents WHERE id = ?`, id)
	return err
}

// --- Operações de Skill ---

func (s *Store) SaveSkill(name, description, tags, content string) (int64, error) {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO skills (id, name, description, tags, content)
		VALUES ((SELECT id FROM skills WHERE name = ?), ?, ?, ?, ?)`,
		name, name, description, tags, content)
	if err != nil {
		return 0, err
	}
	var id int64
	s.db.QueryRow(`SELECT id FROM skills WHERE name = ?`, name).Scan(&id)
	return id, nil
}

func (s *Store) GetSkills() ([]map[string]string, error) {
	rows, err := s.db.Query(`SELECT id, name, description, tags FROM skills`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]string
	for rows.Next() {
		var id int64
		var name, desc, tags sql.NullString
		if err := rows.Scan(&id, &name, &desc, &tags); err != nil {
			return nil, err
		}
		out = append(out, map[string]string{
			"id":          fmt.Sprintf("%d", id),
			"name":        name.String,
			"description": desc.String,
			"tags":        tags.String,
		})
	}
	return out, nil
}

func (s *Store) GetSkillContent(id int64) (string, error) {
	var content string
	err := s.db.QueryRow(`SELECT content FROM skills WHERE id = ?`, id).Scan(&content)
	return content, err
}

func (s *Store) DeleteSkill(id int64) error {
	_, err := s.db.Exec(`DELETE FROM skills WHERE id = ?`, id)
	return err
}

// --- Operações de MCP ---

func (s *Store) SaveMCP(name, connectType, command, arguments, environment, color, icon string) (int64, error) {
	_, err := s.db.Exec(`
			INSERT OR REPLACE INTO mcps (id, nome, connect_type, command, arguments, environment, color, icon)
			VALUES ((SELECT id FROM mcps WHERE nome = ?), ?, ?, ?, ?, ?, ?, ?)`,
		name, name, connectType, command, arguments, environment, color, icon)
	if err != nil {
		return 0, err
	}
	var id int64
	s.db.QueryRow(`SELECT id FROM mcps WHERE nome = ?`, name).Scan(&id)
	return id, nil
}

func (s *Store) GetMCPs() ([]map[string]string, error) {
	rows, err := s.db.Query(`SELECT id, nome, connect_type, command, arguments, environment, color, icon FROM mcps`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]string
	for rows.Next() {
		var id int64
		var nome, ct, cmd, args, env, color, icon sql.NullString
		if err := rows.Scan(&id, &nome, &ct, &cmd, &args, &env, &color, &icon); err != nil {
			return nil, err
		}
		out = append(out, map[string]string{
			"id":           fmt.Sprintf("%d", id),
			"name":         nome.String,
			"connect_type": ct.String,
			"command":      cmd.String,
			"arguments":    args.String,
			"environment":  env.String,
			"color":        color.String,
			"icon":         icon.String,
		})
	}
	return out, nil
}

// GetMCPsMap returns MCP servers as map[name]MCPServerUI
func (s *Store) GetMCPsMap() (map[string]MCPServerUI, error) {
	rows, err := s.db.Query(`SELECT nome, connect_type, command, arguments, environment, color, icon FROM mcps`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]MCPServerUI{}
	for rows.Next() {
		var nome, ct, cmd, args, env, color, icon sql.NullString
		if err := rows.Scan(&nome, &ct, &cmd, &args, &env, &color, &icon); err != nil {
			return nil, err
		}
		m := MCPServerUI{Command: cmd.String, URL: "", Enabled: true, Icon: icon.String, Color: color.String}
		if args.Valid && args.String != "" {
			var arr []string
			if err := json.Unmarshal([]byte(args.String), &arr); err == nil {
				m.Args = arr
			}
		}
		if env.Valid && env.String != "" {
			var em map[string]string
			if err := json.Unmarshal([]byte(env.String), &em); err == nil {
				if u, ok := em["__url"]; ok {
					m.URL = u
					delete(em, "__url")
				}
				m.Env = em
			}
		}
		out[nome.String] = m
	}
	return out, nil
}

// SaveAppState writes the active workspace state (single-row table id=1)
func (s *Store) SaveAppState(activePath string, index int) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO app_state (id, active_workspace_path, active_workspace_index) VALUES (1, ?, ?)`, activePath, index)
	return err
}

// GetAppState returns active workspace path and index
func (s *Store) GetAppState() (string, int, error) {
	var path sql.NullString
	var idx sql.NullInt64
	if err := s.db.QueryRow(`SELECT active_workspace_path, active_workspace_index FROM app_state WHERE id = 1`).Scan(&path, &idx); err != nil {
		if err == sql.ErrNoRows {
			return "", 0, nil
		}
		return "", 0, err
	}
	return path.String, int(idx.Int64), nil
}

func (s *Store) SaveToolProfiles(profiles []ToolProfile) error {
	// transactionally replace all tool profiles
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM tool_profile_tools`); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM tool_profiles`); err != nil {
		return err
	}

	for _, p := range profiles {
		if _, err = tx.Exec(`INSERT INTO tool_profiles (name, color, icon) VALUES (?, ?, ?)`, p.Name, p.Color, p.Icon); err != nil {
			return err
		}
		var pid int64
		if err = tx.QueryRow(`SELECT id FROM tool_profiles WHERE name = ?`, p.Name).Scan(&pid); err != nil {
			return err
		}
		for _, t := range p.Tools {
			if _, err = tx.Exec(`INSERT INTO tool_profile_tools (profile_id, tool_name) VALUES (?, ?)`, pid, t); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *Store) GetToolProfiles() ([]ToolProfile, error) {
	rows, err := s.db.Query(`SELECT id, name, color, icon FROM tool_profiles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ToolProfile
	for rows.Next() {
		var p ToolProfile
		if err := rows.Scan(&p.ID, &p.Name, &p.Color, &p.Icon); err != nil {
			return nil, err
		}
		// load tools
		trows, err := s.db.Query(`SELECT tool_name FROM tool_profile_tools WHERE profile_id = ? ORDER BY id`, p.ID)
		if err == nil {
			for trows.Next() {
				var tn string
				if trows.Scan(&tn) == nil {
					p.Tools = append(p.Tools, tn)
				}
			}
			trows.Close()
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *Store) DeleteMCP(id int64) error {
	_, err := s.db.Exec(`DELETE FROM mcps WHERE id = ?`, id)
	return err
}

// --- Operações de Junction (workspace <-> worker/agent/skill/tool) ---

func (s *Store) SetWorkspaceWorkers(workspaceID int64, workerIDs []int64) error {
	if _, err := s.db.Exec(`DELETE FROM workspace_workers WHERE workspace_id = ?`, workspaceID); err != nil {
		return err
	}
	for _, wid := range workerIDs {
		if _, err := s.db.Exec(`INSERT OR IGNORE INTO workspace_workers (workspace_id, worker_id) VALUES (?, ?)`, workspaceID, wid); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetWorkspaceWorkerIDs(workspaceID int64) []int64 {
	rows, err := s.db.Query(`SELECT worker_id FROM workspace_workers WHERE workspace_id = ?`, workspaceID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func (s *Store) GetWorkspaceWorkerNames(workspaceID int64) []string {
	rows, err := s.db.Query(`SELECT w.name FROM workers w JOIN workspace_workers ww ON ww.worker_id = w.id WHERE ww.workspace_id = ?`, workspaceID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil {
			names = append(names, n)
		}
	}
	return names
}

func (s *Store) SetWorkspaceAgents(workspaceID int64, agentIDs []int64) error {
	if _, err := s.db.Exec(`DELETE FROM workspace_agents WHERE workspace_id = ?`, workspaceID); err != nil {
		return err
	}
	for _, aid := range agentIDs {
		if _, err := s.db.Exec(`INSERT OR IGNORE INTO workspace_agents (workspace_id, agent_id) VALUES (?, ?)`, workspaceID, aid); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetWorkspaceAgentIDs(workspaceID int64) []int64 {
	rows, err := s.db.Query(`SELECT agent_id FROM workspace_agents WHERE workspace_id = ?`, workspaceID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func (s *Store) GetWorkspaceAgentNames(workspaceID int64) []string {
	rows, err := s.db.Query(`SELECT a.name FROM agents a JOIN workspace_agents wa ON wa.agent_id = a.id WHERE wa.workspace_id = ?`, workspaceID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil {
			names = append(names, n)
		}
	}
	return names
}

func (s *Store) SetWorkspaceSkills(workspaceID int64, skillIDs []int64) error {
	if _, err := s.db.Exec(`DELETE FROM workspace_skills WHERE workspace_id = ?`, workspaceID); err != nil {
		return err
	}
	for _, sid := range skillIDs {
		if _, err := s.db.Exec(`INSERT OR IGNORE INTO workspace_skills (workspace_id, skill_id) VALUES (?, ?)`, workspaceID, sid); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SetWorkspaceTools(workspaceID int64, toolNames []string) error {
	if _, err := s.db.Exec(`DELETE FROM workspace_tools WHERE workspace_id = ?`, workspaceID); err != nil {
		return err
	}
	for _, t := range toolNames {
		if _, err := s.db.Exec(`INSERT OR IGNORE INTO workspace_tools (workspace_id, tool_name) VALUES (?, ?)`, workspaceID, t); err != nil {
			return err
		}
	}
	return nil
}

// GetWorkspaceToolsByID returns tool names for a workspace.
func (s *Store) GetWorkspaceToolsByID(workspaceID int64) []string {
	rows, err := s.db.Query(`SELECT tool_name FROM workspace_tools WHERE workspace_id = ? ORDER BY id`, workspaceID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var tools []string
	for rows.Next() {
		var t string
		if rows.Scan(&t) == nil {
			tools = append(tools, t)
		}
	}
	return tools
}

// GetSkillIDByName returns the skill id for a given skill name, or 0 if not found.
func (s *Store) GetSkillIDByName(name string) (int64, error) {
	var id int64
	if err := s.db.QueryRow(`SELECT id FROM skills WHERE name = ?`, name).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return id, nil
}

func (s *Store) SetWorkspaceFolders(workspaceID int64, folders []string) error {
	if _, err := s.db.Exec(`DELETE FROM workspace_folders WHERE workspace_id = ?`, workspaceID); err != nil {
		return err
	}
	for _, f := range folders {
		if _, err := s.db.Exec(`INSERT INTO workspace_folders (workspace_id, folder_path) VALUES (?, ?)`, workspaceID, f); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetWorkspaceFolders(workspaceID int64) []string {
	rows, err := s.db.Query(`SELECT folder_path FROM workspace_folders WHERE workspace_id = ? ORDER BY id`, workspaceID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var paths []string
	for rows.Next() {
		var p string
		if rows.Scan(&p) == nil {
			paths = append(paths, p)
		}
	}
	return paths
}

func (s *Store) SetWorkspaceKnowledge(workspaceID int64, items []string) error {
	if _, err := s.db.Exec(`DELETE FROM workspace_knowledge WHERE workspace_id = ?`, workspaceID); err != nil {
		return err
	}
	for _, k := range items {
		if _, err := s.db.Exec(`INSERT INTO workspace_knowledge (workspace_id, knowledge_item) VALUES (?, ?)`, workspaceID, k); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetWorkspaceKnowledge(workspaceID int64) []string {
	rows, err := s.db.Query(`SELECT knowledge_item FROM workspace_knowledge WHERE workspace_id = ? ORDER BY id`, workspaceID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var k string
		if rows.Scan(&k) == nil {
			items = append(items, k)
		}
	}
	return items
}

// --- Operações de Providers (normalizado) ---

type StoredProvider struct {
	ID              int64
	Name            string
	APIURL          string
	ConnectionTypes string
	Color           string
	Icon            string
	APIKeys         []string
	Models          []StoredModel
}

type StoredModel struct {
	ID        int64
	Model     string
	Free      bool
	Thinking  bool
	Tool      bool
	Embedding bool
	Vision    bool
	Health    int
}

func (s *Store) SaveProviderFull(p StoredProvider) error {
	_, err := s.db.Exec(`
			INSERT OR REPLACE INTO providers (id, name, api_url, connection_types, color, icon)
			VALUES ((SELECT id FROM providers WHERE name = ?), ?, ?, ?, ?, ?)`,
		p.Name, p.Name, p.APIURL, p.ConnectionTypes, p.Color, p.Icon)
	if err != nil {
		return err
	}
	var pid int64
	if err := s.db.QueryRow(`SELECT id FROM providers WHERE name = ?`, p.Name).Scan(&pid); err != nil {
		return err
	}
	// API keys
	if _, err := s.db.Exec(`DELETE FROM provider_apikeys WHERE provider_id = ?`, pid); err != nil {
		return err
	}
	for _, k := range p.APIKeys {
		if _, err := s.db.Exec(`INSERT OR IGNORE INTO provider_apikeys (provider_id, apikey) VALUES (?, ?)`, pid, k); err != nil {
			return err
		}
	}
	// Models
	if _, err := s.db.Exec(`DELETE FROM provider_models WHERE provider_id = ?`, pid); err != nil {
		return err
	}
	for _, m := range p.Models {
		if _, err := s.db.Exec(`
			INSERT OR IGNORE INTO provider_models (provider_id, model, free, thinking, tool, embedding, vision, health)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			pid, m.Model, m.Free, m.Thinking, m.Tool, m.Embedding, m.Vision, m.Health); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetProvidersFull() ([]StoredProvider, error) {
	rows, err := s.db.Query(`SELECT id, name, api_url, connection_types, color, icon FROM providers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []StoredProvider
	for rows.Next() {
		var p StoredProvider
		if err := rows.Scan(&p.ID, &p.Name, &p.APIURL, &p.ConnectionTypes, &p.Color, &p.Icon); err != nil {
			return nil, err
		}
		// keys
		krows, err := s.db.Query(`SELECT apikey FROM provider_apikeys WHERE provider_id = ?`, p.ID)
		if err == nil {
			for krows.Next() {
				var k string
				if krows.Scan(&k) == nil {
					p.APIKeys = append(p.APIKeys, k)
				}
			}
			krows.Close()
		}
		// models
		mrows, err := s.db.Query(`SELECT id, model, free, thinking, tool, embedding, vision, health FROM provider_models WHERE provider_id = ?`, p.ID)
		if err == nil {
			for mrows.Next() {
				var m StoredModel
				if mrows.Scan(&m.ID, &m.Model, &m.Free, &m.Thinking, &m.Tool, &m.Embedding, &m.Vision, &m.Health) == nil {
					p.Models = append(p.Models, m)
				}
			}
			mrows.Close()
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *Store) DeleteProviderFull(name string) error {
	var pid int64
	if err := s.db.QueryRow(`SELECT id FROM providers WHERE name = ?`, name).Scan(&pid); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM providers WHERE id = ?`, pid)
	return err
}

// --- FixedModels ---

type FixedModels struct {
	EmbeddingProvider string
	EmbeddingModel    string
	ImageProvider     string
	ImageModel        string
	SpecProvider      string
	SpecModel         string
}

func (s *Store) SaveFixedModels(f FixedModels) error {
	_, err := s.db.Exec(`DELETE FROM fixedmodels`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		INSERT INTO fixedmodels (embedding_provider, embedding_model, image_provider, image_model, spec_provider, spec_model)
		VALUES (?, ?, ?, ?, ?, ?)`,
		f.EmbeddingProvider, f.EmbeddingModel, f.ImageProvider, f.ImageModel, f.SpecProvider, f.SpecModel)
	return err
}

func (s *Store) GetFixedModels() (FixedModels, error) {
	var f FixedModels
	err := s.db.QueryRow(`SELECT embedding_provider, embedding_model, image_provider, image_model, spec_provider, spec_model FROM fixedmodels LIMIT 1`).
		Scan(&f.EmbeddingProvider, &f.EmbeddingModel, &f.ImageProvider, &f.ImageModel, &f.SpecProvider, &f.SpecModel)
	if err == sql.ErrNoRows {
		return f, nil
	}
	return f, err
}

// adaptProviderConfig converte ProviderConfig (modelo de app) → StoredProvider (modelo de DB).
func adaptProviderConfig(name string, p ProviderConfig) StoredProvider {
	sp := StoredProvider{
		Name:            name,
		APIURL:          p.ApiUrl,
		ConnectionTypes: p.TypeConnection,
		Color:           p.Color,
		Icon:            p.Icon,
	}
	sp.APIKeys = append(sp.APIKeys, p.GetAPIKey())
	for name, m := range p.Models {
		sm := StoredModel{Model: name, Free: m.Free, Thinking: m.Thinking, Tool: m.Tools, Embedding: m.Embedding, Vision: m.Vision}
		sp.Models = append(sp.Models, sm)
	}
	return sp
}

// deadaptProviderConfig converte StoredProvider (DB) → ProviderConfig (app).
func deadaptProviderConfig(sp StoredProvider) ProviderConfig {
	pc := ProviderConfig{
		ApiUrl:         sp.APIURL,
		TypeConnection: sp.ConnectionTypes,
		Color:          sp.Color,
		Icon:           sp.Icon,
	}
	for _, k := range sp.APIKeys {
		pc.ApiKeys = append(pc.ApiKeys, ProviderApiKey{Key: k})
	}
	pc.Models = make(map[string]ModelSettings)
	for _, m := range sp.Models {
		pc.Models[m.Model] = ModelSettings{Free: m.Free, Thinking: m.Thinking, Tools: m.Tool, Embedding: m.Embedding, Vision: m.Vision}
	}
	return pc
}

// --- Operações de Sessão ---

func (s *Store) SaveSession(sess ChatSession) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT OR REPLACE INTO sessions (id, workspace_path, title, pinned, embedding, created_at, updated_at,
			worker_name, parent_session_id, model, provider, mode, thinking, summary, summarized_context,
			summary_token_count, summarized_at, last_summarized_msg_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.WorkspaceID, sess.Title, sess.Pinned, nil,
		sess.CreatedAt, sess.UpdatedAt, sess.WorkerName, sess.ParentSessionID, sess.Model, sess.Provider,
		sess.Mode, sess.Thinking, sess.Summary, sess.SummarizedContext, 0, sess.SummarizedAt, sess.LastSummarizedMsgID)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM messages WHERE session_id = ?`, sess.ID); err != nil {
		return err
	}
	for i, msg := range sess.Messages {
		res, err := tx.Exec(`INSERT INTO messages (session_id, role, content, time) VALUES (?, ?, ?, ?)`,
			sess.ID, msg.Role, msg.Content, msg.Time)
		if err != nil {
			log.Printf("Erro ao salvar mensagem: %v", err)
			continue
		}
		if id, err := res.LastInsertId(); err == nil {
			sess.Messages[i].ID = id
		}
	}
	return tx.Commit()
}

func (s *Store) AddMessageToSession(sessionID string, role string, content string) error {
	_, err := s.db.Exec(`INSERT INTO messages (session_id, role, content, time) VALUES (?, ?, ?, ?)`,
		sessionID, role, content, time.Now())
	return err
}

func (s *Store) GetSessions(workspacePath string) ([]*ChatSession, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_path, title, pinned, worker_name, parent_session_id, model, provider, mode, thinking,
			summary, summarized_context, summarized_at, last_summarized_msg_id, created_at, updated_at
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
		err := rows.Scan(&sess.ID, &sess.WorkspaceID, &sess.Title, &sess.Pinned, &sess.WorkerName,
			&sess.ParentSessionID, &sess.Model, &sess.Provider, &sess.Mode, &sess.Thinking,
			&sess.Summary, &sess.SummarizedContext, &summarizedAt, &sess.LastSummarizedMsgID,
			&sess.CreatedAt, &sess.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if summarizedAt.Valid {
			sess.SummarizedAt = summarizedAt.Time
		}
		sess.Messages = s.loadMessages(sess.ID)
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

func (s *Store) GetSession(id string) (*ChatSession, error) {
	sess := &ChatSession{}
	var summarizedAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT id, workspace_path, title, pinned, worker_name, parent_session_id, model, provider, mode, thinking,
			summary, summarized_context, summarized_at, last_summarized_msg_id, created_at, updated_at
		FROM sessions WHERE id = ?`, id).Scan(
		&sess.ID, &sess.WorkspaceID, &sess.Title, &sess.Pinned, &sess.WorkerName,
		&sess.ParentSessionID, &sess.Model, &sess.Provider, &sess.Mode, &sess.Thinking,
		&sess.Summary, &sess.SummarizedContext, &summarizedAt, &sess.LastSummarizedMsgID,
		&sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if summarizedAt.Valid {
		sess.SummarizedAt = summarizedAt.Time
	}
	sess.Messages = s.loadMessages(id)
	return sess, nil
}

func (s *Store) loadMessages(sessionID string) []ChatMessage {
	rows, err := s.db.Query(`SELECT id, role, content, time FROM messages WHERE session_id = ? ORDER BY time ASC`, sessionID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var msgs []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if rows.Scan(&m.ID, &m.Role, &m.Content, &m.Time) == nil {
			msgs = append(msgs, m)
		}
	}
	return msgs
}

func (s *Store) DeleteSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// --- Operações de Memória ---

func (s *Store) SaveMemory(workspacePath string, content string, importance int) error {
	now := time.Now()
	_, err := s.db.Exec(`
		INSERT INTO memories (workspace_path, content, importance, embedding, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		workspacePath, content, importance, nil, now, now)
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

func (s *Store) SaveSpecWizard(w SpecWizardConfig) error {
	now := time.Now()
	if w.CreatedAt.IsZero() {
		w.CreatedAt = now
	}
	w.UpdatedAt = now

	funcReq, _ := json.Marshal(w.FunctionalRequirements)
	nonFuncReq, _ := json.Marshal(w.NonFunctionalRequirements)
	engPhilosophies, _ := json.Marshal(w.EngineeringPhilosophies)
	designPatterns, _ := json.Marshal(w.DesignPatterns)
	dataPatterns, _ := json.Marshal(w.DataPatterns)
	stackConfig, _ := json.Marshal(w.StackConfig)

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO spec_wizards (id, name, description, expert_language_plugin, prd,
			functional_requirements, non_functional_requirements, persistence, architecture,
			engineering_philosophies, design_patterns, data_patterns, stack_config,
			business_state_management, business_api_contract, business_customization_details,
			business_final_adjustments, business_architecture_recommendations,
			color, icon, architecture_health, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		w.ID, w.Name, w.Description, w.ExpertLanguagePlugin, w.PRD,
		string(funcReq), string(nonFuncReq), w.Persistence, w.Architecture,
		string(engPhilosophies), string(designPatterns), string(dataPatterns), string(stackConfig),
		w.BusinessStateManagement, w.BusinessAPIContract, w.BusinessCustomizationDetails,
		w.BusinessFinalAdjustments, w.BusinessArchitectureRecommendations,
		w.Color, w.Icon, w.ArchitectureHealth, w.CreatedAt, w.UpdatedAt)
	return err
}

func (s *Store) GetSpecWizards() ([]SpecWizardConfig, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, expert_language_plugin, prd,
			functional_requirements, non_functional_requirements, persistence, architecture,
			engineering_philosophies, design_patterns, data_patterns, stack_config,
			business_state_management, business_api_contract, business_customization_details,
			business_final_adjustments, business_architecture_recommendations,
			color, icon, architecture_health, created_at, updated_at
		FROM spec_wizards ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SpecWizardConfig
	for rows.Next() {
		var w SpecWizardConfig
		var funcReq, nonFuncReq, engPhil, designPat, dataPat, stackCfg sql.NullString
		if err := rows.Scan(&w.ID, &w.Name, &w.Description, &w.ExpertLanguagePlugin, &w.PRD,
			&funcReq, &nonFuncReq, &w.Persistence, &w.Architecture,
			&engPhil, &designPat, &dataPat, &stackCfg,
			&w.BusinessStateManagement, &w.BusinessAPIContract, &w.BusinessCustomizationDetails,
			&w.BusinessFinalAdjustments, &w.BusinessArchitectureRecommendations,
			&w.Color, &w.Icon, &w.ArchitectureHealth, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		if funcReq.Valid {
			json.Unmarshal([]byte(funcReq.String), &w.FunctionalRequirements)
		}
		if nonFuncReq.Valid {
			json.Unmarshal([]byte(nonFuncReq.String), &w.NonFunctionalRequirements)
		}
		if engPhil.Valid {
			json.Unmarshal([]byte(engPhil.String), &w.EngineeringPhilosophies)
		}
		if designPat.Valid {
			json.Unmarshal([]byte(designPat.String), &w.DesignPatterns)
		}
		if dataPat.Valid {
			json.Unmarshal([]byte(dataPat.String), &w.DataPatterns)
		}
		if stackCfg.Valid {
			json.Unmarshal([]byte(stackCfg.String), &w.StackConfig)
		}
		if w.Color == "" {
			w.Color = "#3b82f6"
		}
		if w.Icon == "" {
			w.Icon = "📝"
		}
		out = append(out, w)
	}
	return out, nil
}

func (s *Store) GetSpecWizardByID(id string) (*SpecWizardConfig, error) {
	row := s.db.QueryRow(`
		SELECT id, name, description, expert_language_plugin, prd,
			functional_requirements, non_functional_requirements, persistence, architecture,
			engineering_philosophies, design_patterns, data_patterns, stack_config,
			business_state_management, business_api_contract, business_customization_details,
			business_final_adjustments, business_architecture_recommendations,
			color, icon, architecture_health, created_at, updated_at
		FROM spec_wizards WHERE id = ?`, id)
	var w SpecWizardConfig
	var funcReq, nonFuncReq, engPhil, designPat, dataPat, stackCfg sql.NullString
	err := row.Scan(&w.ID, &w.Name, &w.Description, &w.ExpertLanguagePlugin, &w.PRD,
		&funcReq, &nonFuncReq, &w.Persistence, &w.Architecture,
		&engPhil, &designPat, &dataPat, &stackCfg,
		&w.BusinessStateManagement, &w.BusinessAPIContract, &w.BusinessCustomizationDetails,
		&w.BusinessFinalAdjustments, &w.BusinessArchitectureRecommendations,
		&w.Color, &w.Icon, &w.ArchitectureHealth, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if funcReq.Valid {
		json.Unmarshal([]byte(funcReq.String), &w.FunctionalRequirements)
	}
	if nonFuncReq.Valid {
		json.Unmarshal([]byte(nonFuncReq.String), &w.NonFunctionalRequirements)
	}
	if engPhil.Valid {
		json.Unmarshal([]byte(engPhil.String), &w.EngineeringPhilosophies)
	}
	if designPat.Valid {
		json.Unmarshal([]byte(designPat.String), &w.DesignPatterns)
	}
	if dataPat.Valid {
		json.Unmarshal([]byte(dataPat.String), &w.DataPatterns)
	}
	if stackCfg.Valid {
		json.Unmarshal([]byte(stackCfg.String), &w.StackConfig)
	}
	if w.Color == "" {
		w.Color = "#3b82f6"
	}
	if w.Icon == "" {
		w.Icon = "📝"
	}
	return &w, nil
}

func (s *Store) DeleteSpecWizard(id string) error {
	_, err := s.db.Exec(`DELETE FROM spec_wizards WHERE id = ?`, id)
	return err
}

func Float32ToByte(f []float32) []byte {
	buf := make([]byte, len(f)*4)
	for i, v := range f {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}

// --- Workspace Templates ---

func (s *Store) GetWorkspaceTemplates() ([]WorkspaceTemplate, error) {
	rows, err := s.db.Query(`SELECT id, name, description, personality, created_at FROM workspace_templates ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var templates []WorkspaceTemplate
	for rows.Next() {
		var t WorkspaceTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Personality, &t.CreatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (s *Store) SaveWorkspaceTemplate(t WorkspaceTemplate) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO workspace_templates (id, name, description, personality) VALUES (?, ?, ?, ?)`,
		t.ID, t.Name, t.Description, t.Personality,
	)
	return err
}

func (s *Store) DeleteWorkspaceTemplate(id int64) error {
	_, err := s.db.Exec(`DELETE FROM workspace_templates WHERE id = ?`, id)
	return err
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
