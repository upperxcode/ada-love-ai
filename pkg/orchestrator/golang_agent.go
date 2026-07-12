package orchestrator

import (
	"context"
	"fmt"
)

// GoLangAgentConfig holds configuration for the GoLang agent
type GoLangAgentConfig struct {
	Model         string
	WorkspaceRoot string
}

// GoLangAgent specializes in Go backend development
type GoLangAgent struct {
	*BaseSubAgent
}

// NewGoLangAgent creates a new GoLangAgent
func NewGoLangAgent(config GoLangAgentConfig) *GoLangAgent {
	return &GoLangAgent{
		BaseSubAgent: NewBaseSubAgent(AgentTypeGoLang, SubAgentConfig{
			Model:         config.Model,
			SystemPrompt:  goLangSystemPrompt,
			Capabilities:  []AgentCapability{CapabilityGoBackend, CapabilityDatabase, CapabilityAPI, CapabilityConcurrency},
			AllowedTools:  []string{"read_file", "write_file", "edit_file", "list_dir", "glob", "exec"},
			MaxIterations: 10,
			Temperature:   0.2,
		}, config.WorkspaceRoot),
	}
}

const goLangSystemPrompt = "Você é um Engenheiro Go Sênior especialista em:\n" +
	"- Clean Architecture, DDD, Hexagonal Architecture\n" +
	"- Concorrência: goroutines, channels, sync, context, worker pools\n" +
	"- Performance: profiling, benchmarking, memory management\n" +
	"- Banco de dados: SQL (pgx, sqlx, gorm), NoSQL (redis, mongo)\n" +
	"- APIs: REST, gRPC, GraphQL, WebSockets\n" +
	"- Testes: table-driven tests, mocks (gomock/testify), integration tests\n" +
	"- Observabilidade: logging estruturado, metrics, tracing, health checks\n" +
	"- Segurança: auth (JWT/OAuth), rate limiting, input validation, SQL injection prevention\n\n" +
	"REGRAS DE CÓDIGO:\n" +
	"1. Código idiomático Go (gofmt, golint, go vet, staticcheck passam)\n" +
	"2. Tratamento de erros explícito (errors.Is/As, wrapped errors)\n" +
	"3. Context em todas as operações I/O (context.WithTimeout)\n" +
	"4. Interfaces pequenas, implementações desacopladas\n" +
	"5. Dependency injection via construtores\n" +
	"6. Config via structs + env (viper/spf13)\n" +
	"6. Logs estruturados (slog/zap) com níveis apropriados\n" +
	"7. Testes: table-driven, mocks (gomock/testify), race detector limpo\n\n" +
	"PADRÕES DE ARQUITETURA:\n" +
	"- Handler -> Service -> Repository (ou UseCase -> Port -> Adapter)\n" +
	"- Config centralizada, secrets via env\n" +
	"- Middleware: logging, recovery, auth, rate limit, tracing\n" +
	"- Graceful shutdown (signal handling, context cancellation)\n" +
	"- Health checks (liveness/readiness)\n\n" +
	"PADRÕES DE NAMING:\n" +
	"- Arquivos: snake_case (user_service.go, user_service_test.go)\n" +
	"- Packages: singular, lowercase (user, order, payment)\n" +
	"- Interfaces: sufixo er (UserRepository, PaymentProcessor)\n" +
	"- Structs: PascalCase (User, OrderService)\n" +
	"- Constantes: SCREAMING_SNAKE_CASE\n" +
	"- Variáveis: camelCase\n\n" +
	"OUTPUT FORMAT:\n" +
	"Sempre forneça arquivos completos, prontos para compilar.\n" +
	"Use blocos markdown com nome do arquivo: ```go // arquivo.go```\n" +
	"Inclua imports, types, funções, testes quando relevante.\n\n" +
	"FERRAMENTAS DISPONÍVEIS:\n" +
	"read_file, write_file, edit_file, list_dir, glob, exec (go build, go test, go vet)\n" +
	"grep_search, view_file_outline, run_tests\n\n" +
	"FERRAMENTA exec: use para compilar, testar, lintar.\n" +
	"- go build ./...\n" +
	"- go test -race -count=1 ./...\n" +
	"- go vet ./...\n" +
	"- staticcheck ./...\n" +
	"- golangci-lint run\n\n" +
	"Para criar arquivos: write_file com path relativo ao workspace root.\n" +
	"Para rodar testes: exec com \"go test -v -race -count=1 ./...\"\n" +
	"Para lint: exec com \"golangci-lint run\""

const goEngineeringRules = "REGRAS DE ENGENHARIA GO (OBRIGATÓRIAS):\n" +
	"1. SEMPRE trate erros explicitamente (if err != nil)\n" +
	"2. USE context.Context em TODAS operações I/O (DB, HTTP, Redis, etc)\n" +
	"3. USE interfaces para desacoplamento (Repository, Service interfaces)\n" +
	"4. USE dependency injection via construtores (NewService(repo Repo) *Service)\n" +
	"5. USE context.WithTimeout/WithCancel para timeouts\n" +
	"6. USE structured logging (slog) - NUNCA fmt.Println\n" +
	"7. NUNCA ignore erros (_ = err)\n" +
	"8. USE prepared statements (parameterized queries) - NUNCA string concat\n" +
	"9. USE transações para operações multi-query\n" +
	"10. USE mutex/atomic para concorrência segura\n" +
	"11. USE race detector (go test -race) - deve passar limpo\n" +
	"12. USE table-driven tests para múltiplos cenários\n" +
	"13. USE mocks (gomock/testify) para isolamento\n" +
	"14. USE benchmarks para paths críticos\n" +
	"13. USE go vet, staticcheck, golangci-lint - deve passar limpo\n" +
	"13. USE gofmt/goimports - código formatado\n" +
	"14. USE go modules - versão explícita no go.mod\n" +
	"14. USE build tags se necessário (//go:build)\n" +
	"18. USE build constraints para arquivos de teste (_test.go)\n" +
	"19. NUNCA hardcode secrets - USE env vars\n" +
	"19. USE constants para valores mágicos\n" +
	"19. USE generics quando apropriado (Go 1.18+)\n" +
	"20. DOCUMENTE exported types/functions com comentários\n" +
	"20. USE godoc format (Package X provides...)"

func (g *GoLangAgent) Execute(ctx context.Context, task string, layers PromptLayers) (*AgentResult, error) {
	// Cria cópia para não mutar o original
	enhancedLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          layers.Task + "\n\nREGRAS DE ENGENHARIA GO:\n" + goEngineeringRules,
	}

	return g.BaseSubAgent.Execute(ctx, task, enhancedLayers)
}

func (g *GoLangAgent) CreateHandler(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Crie handler HTTP + service + repository para: %s\n\nREQUISITOS:\n- Clean Architecture (Handler->Service->Repository)\n- Validação de input\n- Error handling com errors.Is/As\n- Context em todas operações I/O\n- Logs estruturados (slog)\n- Testes unitários (table-driven + mocks)", spec)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return g.Execute(ctx, task, newLayers)
}

func (g *GoLangAgent) CreateService(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Crie service Go para: %s\n\nREQUISITOS:\n- Interface + implementação\n- Injeção de dependência\n- Transações DB (se aplicável)\n- Validação de regras de negócio\n- Testes unitários com mocks", spec)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return g.Execute(ctx, task, newLayers)
}

func (g *GoLangAgent) CreateRepository(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Crie repository Go para: %s\n\nREQUISITOS:\n- Interface + implementação (pgx/sqlx/gorm)\n- Queries parametrizadas (prepared statements)\n- Context em todas queries\n- Transações (Begin/Tx/Commit/Rollback)\n- Migrações (golang-migrate)\n- Testes com testcontainers ou mock", spec)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return g.Execute(ctx, task, newLayers)
}

func (g *GoLangAgent) CreateModel(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Crie models/structs Go para: %s\n\nREQUISITOS:\n- Structs com tags json, db, validate\n- Validação (validator.v10)\n- JSON serialization correta\n- Database tags (pgx/gorm)\n- Métodos helpers se necessário", spec)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return g.Execute(ctx, task, newLayers)
}

func (g *GoLangAgent) WriteTests(ctx context.Context, code string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Escreva testes unitários completos para:\n%s\n\nREQUISITOS:\n- Table-driven tests\n- Mocks com gomock/testify\n- Coverage >= 80%%\n- Testes de erro, edge cases, concorrência\n- Benchmarks se aplicável\n- race detector limpo", code)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return g.Execute(ctx, task, newLayers)
}

func (g *GoLangAgent) Refactor(ctx context.Context, code string, instructions string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("REFACTOR:\n%s\n\nINSTRUÇÕES:\n%s\n\nCÓDIGO ATUAL:\n%s", instructions, code, code)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return g.Execute(ctx, task, newLayers)
}

func (g *GoLangAgent) ReviewCode(ctx context.Context, code string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("REVIEW DE CÓDIGO GO:\n%s\n\nVERIFIQUE:\n- Idiomatic Go\n- Error handling\n- Performance\n- Security\n- Testability\n- Clean Architecture\n- Naming conventions\n- Documentation", code)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return g.Execute(ctx, task, newLayers)
}