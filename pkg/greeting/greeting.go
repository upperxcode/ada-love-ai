package greeting

import (
	"database/sql"
	"log"
	"strings"
	"unicode"

	_ "modernc.org/sqlite"
)

type GreetingSystem struct {
	db *sql.DB
}

func NewGreetingSystem(dbPath string) (*GreetingSystem, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS greetings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		keyword TEXT UNIQUE NOT NULL,
		language TEXT NOT NULL,
		response TEXT NOT NULL
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	return &GreetingSystem{db: db}, nil
}

func (s *GreetingSystem) Seed() {
	greetings := []struct {
		keyword  string
		lang     string
		response string
	}{
		{"ola", "pt", "Olá! Sou o assistente virtual. Como posso ajudar com seus sistemas ou códigos hoje?"},
		{"oi", "pt", "Olá! Tudo bem? Em que posso te ajudar hoje?"},
		{"bomdia", "pt", "Bom dia! Como posso ser útil hoje?"},
		{"boatarde", "pt", "Boa tarde! Como posso ajudar você agora?"},
		{"boanoite", "pt", "Boa noite! Em que posso te ajudar nesta noite?"},
		{"hello", "en", "Hello! I am your virtual assistant. How can I help you with your project today?"},
		{"hi", "en", "Hi there! How can I help you today?"},
		{"hey", "en", "Hey! What can I do for you today?"},
		{"goodmorning", "en", "Good morning! How can I assist you today?"},
		{"goodafternoon", "en", "Good afternoon! How can I help you?"},
		{"goodevening", "en", "Good evening! How can I assist you tonight?"},
	}

	tx, err := s.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, _ := tx.Prepare("INSERT OR IGNORE INTO greetings (keyword, language, response) VALUES (?, ?, ?)")
	defer stmt.Close()

	for _, g := range greetings {
		_, _ = stmt.Exec(g.keyword, g.lang, g.response)
	}
	_ = tx.Commit()
}

func normalizeText(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))

	replacer := strings.NewReplacer("!", "", "?", "", ".", "", ",", "", "-", "")
	text = replacer.Replace(text)

	var result strings.Builder
	for _, r := range text {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		switch r {
		case 'á', 'à', 'ã', 'â':
			result.WriteRune('a')
		case 'é', 'ê':
			result.WriteRune('e')
		case 'í':
			result.WriteRune('i')
		case 'ó', 'ô', 'õ':
			result.WriteRune('o')
		case 'ú':
			result.WriteRune('u')
		default:
			result.WriteRune(r)
		}
	}

	return strings.ReplaceAll(result.String(), " ", "")
}

func (s *GreetingSystem) CheckGreeting(userInput string) (string, bool) {
	cleaned := normalizeText(userInput)

	var response string
	query := "SELECT response FROM greetings WHERE keyword = ?"
	err := s.db.QueryRow(query, cleaned).Scan(&response)

	if err == sql.ErrNoRows {
		return "", false
	} else if err != nil {
		log.Printf("Erro ao consultar o banco: %v", err)
		return "", false
	}

	return response, true
}

func (s *GreetingSystem) Close() error {
	return s.db.Close()
}
