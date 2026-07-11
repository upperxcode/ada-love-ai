package orchestrator

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TesterAgentConfig holds configuration for the Tester agent
type TesterAgentConfig struct {
	Model         string
	WorkspaceRoot string
}

// TesterAgent specializes in testing and quality assurance
type TesterAgent struct {
	*BaseSubAgent
}

const testerSystemPrompt = `Você é um Engenheiro de QA/Testes Sênior especialista em:
- Go: testes unitários, integração, benchmarks, race detector, mocks (gomock/testify)
- React/TypeScript: Vitest + React Testing Library + MSW, userEvent
- Testes de integração: testcontainers, postgres, redis mock
- Contract testing: Pact, OpenAPI validation
- Performance: benchmarks, load testing (k6), profiling (pprof)
- CI/CD: GitHub Actions, GitLab CI, test pipelines
- Mutation testing: go-mutesting, stryker

REGRAS DE TESTES:
1. Teste COMPORTAMENTO, não implementação
2. Table-driven tests (Go) / describe.each + it.each (TS)
3. Mocks APENAS para dependências EXTERNAS (DB, HTTP, time, random)
4. Isolamento TOTAL entre testes (sem estado compartilhado)
5. Nomes: Test<Function>_<Scenario>_<Expected> (Go) / describe/it descriptive (TS)
6. AAA Pattern: Arrange - Act - Assert
7. Coverage: unit >= 80%, integration >= 60%, e2e >= 40%
8. Race detector LIMPO (go test -race)
9. Benchmarks para paths críticos
10. Mutation testing score >= 80%
11. Testes determinísticos (sem flakiness)
12. Isolamento TOTAL: DB limpo, mocks resetados, sem estado global

PADRÕES DE TESTE:
- Unit: funções puras, use cases, utils, hooks
- Integration: DB repositories, HTTP handlers, API endpoints
- Contract: OpenAPI schema validation, Pact consumer-driven
- E2E: Playwright/Cypress para fluxos críticos (login, checkout)
- Performance: k6, pprof, benchmarks Go
- Mutation: go-mutesting, stryker

OUTPUT:
Sempre forneça arquivos de teste completos, prontos para rodar.
Use table-driven tests (Go) / describe.each (TS).
Inclua mocks adequados, setup/teardown, cleanup.

FERRAMENTAS:
Go: testify, gomock, testcontainers, go-fuzz
TS: Vitest, React Testing Library, MSW, @testing-library/user-event, @faker-js/faker
Contract: Pact (Go/TS)
Performance: k6, pprof, go test -bench
Mutation: go-mutesting, stryker`

const testerEngineeringRules = `
REGRAS DE ENGENHARIA DE TESTES (OBRIGATÓRIAS):
1. Teste COMPORTAMENTO, não implementação
2. Table-driven tests (Go) / describe.each + it.each (TS)
3. Mocks APENAS para dependências EXTERNAS (DB, HTTP, time, random)
4. Isolamento TOTAL entre testes (sem estado compartilhado)
5. Nomes: Test<Function>_<Scenario>_<Expected> (Go) / describe/it descriptive (TS)
6. AAA Pattern: Arrange - Act - Assert
7. Coverage: unit >= 80%, integration >= 60%, e2e >= 40%
8. Race detector LIMPO (go test -race)
9. Benchmarks para paths críticos
10. Mutation testing score >= 80%
11. Testes determinísticos (sem flakiness)
12. Isolamento TOTAL: DB limpo, mocks resetados, sem estado global
13. Nomes: Test<Function>_<Scenario>_<Expected> / describe + it
14. Arrange - Act - Assert (AAA)
15. Mocks APENAS para deps externas (DB, HTTP, time, random)
16. Coverage: unit >= 80%, integration >= 60%, e2e >= 40%
17. Race detector LIMPO (go test -race)
18. Benchmarks para paths críticos
19. Mutation testing >= 80%
20. Determinístico (zero flakiness)
21. Isolamento total (DB limpo, mocks resetados, sem estado global)

PADRÕES DE TESTE:
- Unit: funções puras, use cases, utils, hooks
- Integration: DB repositories, HTTP handlers, API endpoints
- Contract: OpenAPI schema validation, Pact consumer-driven
- E2E: Playwright/Cypress para fluxos críticos (login, checkout)
- Performance: k6, pprof, benchmarks Go
- Mutation: go-mutesting, stryker
- Contract: Pact consumer-driven contracts

OUTPUT:
Sempre forneça arquivos de teste completos, prontos para rodar.
Use table-driven tests (Go) / describe.each (TS).
Inclua mocks adequados, setup/teardown, cleanup.

FERRAMENTAS:
Go: testify, gomock, testcontainers, go-fuzz
TS: Vitest, React Testing Library, MSW, @testing-library/user-event, @faker-js/faker
Contract: Pact (Go/TS)
Performance: k6, pprof, go test -bench
Mutation: go-mutesting, stryker`

// NewTesterAgent creates a new TesterAgent
func NewTesterAgent(config TesterAgentConfig) *TesterAgent {
	return &TesterAgent{
		BaseSubAgent: NewBaseSubAgent(AgentTypeTester, SubAgentConfig{
			Model:         config.Model,
			SystemPrompt:  testerSystemPrompt,
			Capabilities:  []AgentCapability{CapabilityTesting, CapabilityCodeReview, CapabilityQualityAssurance},
			AllowedTools:  []string{"read_file", "write_file", "edit_file", "list_dir", "glob", "exec"},
			MaxIterations: 10,
			Temperature:   0.1,
		}, config.WorkspaceRoot),
	}
}

func (t *TesterAgent) Execute(ctx context.Context, task string, layers PromptLayers) (*AgentResult, error) {
	enhancedLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          layers.Task + "\n\n" + testerEngineeringRules,
	}

	return t.BaseSubAgent.Execute(ctx, task, enhancedLayers)
}

func (t *TesterAgent) WriteUnitTests(ctx context.Context, code string, language string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Escreva testes unitários completos para código %s:\n%s\n\nREQUISITOS:\n- Table-driven tests / parametrized tests\n- Coverage >= 80%%\n- Mocks adequados (gomock/testify for Go, vi.fn/MSW for TS)\n- Edge cases + error paths\n- Race detector clean (Go)\n- Nomes descritivos: Test<Function>_<Scenario>_<Expected>", language, code)

	newLayers := PromptLayers{
		SystemPersona:  layers.SystemPersona,
		GlobalContext:  layers.GlobalContext,
		State:          layers.State,
		Task:           task,
	}

	return t.Execute(ctx, task, newLayers)
}

func (t *TesterAgent) WriteIntegrationTests(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Escreva testes de integração para: %s\n\nREQUISITOS:\n- Testcontainers para DB/Redis\n- Setup/teardown automático\n- Isolamento entre testes\n- Testes de API endpoints\n- Testes de repositórios DB\n- Cleanup automático entre testes", spec)

	newLayers := PromptLayers{
		SystemPersona:  layers.SystemPersona,
		GlobalContext:  layers.GlobalContext,
		State:          layers.State,
		Task:           task,
	}

	return t.Execute(ctx, task, newLayers)
}

func (t *TesterAgent) WriteE2ETests(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Escreva testes E2E para: %s\n\nREQUISITOS:\n- Playwright (Go) ou Playwright/Cypress (TS)\n- Fluxos críticos de usuário\n- Login, fluxos críticos\n- Dados de teste isolados\n- Screenshots/video on failure\n- CI/CD integration", spec)

	newLayers := PromptLayers{
		SystemPersona:  layers.SystemPersona,
		GlobalContext:  layers.GlobalContext,
		State:          layers.State,
		Task:           task,
	}

	return t.Execute(ctx, task, newLayers)
}

func (t *TesterAgent) WriteContractTests(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Escreva testes de contrato para: %s\n\nREQUISITOS:\n- Pact consumer-driven contracts\n- Provider verification\n- Schema validation (OpenAPI)\n- Schema registry (se aplicável)\n- CI/CD integration", spec)

	newLayers := PromptLayers{
		SystemPersona:  layers.SystemPersona,
		GlobalContext:  layers.GlobalContext,
		State:          layers.State,
		Task:           task,
	}

	return t.Execute(ctx, task, newLayers)
}

func (t *TesterAgent) WriteBenchmarks(ctx context.Context, code string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Escreva benchmarks para:\n%s\n\nREQUISITOS:\n- go test -bench=.\n- Benchmark functions com b.N\n- b.ResetTimer() / b.StopTimer()\n- b.ReportAllocs()\n- Compare implementations\n- Memory allocations (b.ReportAllocs)", code)

	newLayers := PromptLayers{
		SystemPersona:  layers.SystemPersona,
		GlobalContext:  layers.GlobalContext,
		State:          layers.State,
		Task:           task,
	}

	return t.Execute(ctx, task, newLayers)
}

func (t *TesterAgent) RunTests(ctx context.Context, path string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Execute testes e relate resultados para: %s\n\nEXECUTE:\n- go test -v -race -coverprofile=coverage.out ./...\n- go test -bench=. -benchmem\n- go test -coverprofile=coverage.out\n- go tool cover -html=coverage.out\n- Relate: pass/fail, coverage %%, race issues, benchmarks", path)

	newLayers := PromptLayers{
		SystemPersona:  layers.SystemPersona,
		GlobalContext:  layers.GlobalContext,
		State:          layers.State,
		Task:           task,
	}

	return t.Execute(ctx, task, newLayers)
}

func (t *TesterAgent) AnalyzeCoverage(ctx context.Context, coveragePath string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Analise coverage report em: %s\n\nRELATE:\n- Overall coverage %%\n- Per-package coverage\n- Uncovered critical paths\n- Suggestions to improve", coveragePath)

	newLayers := PromptLayers{
		SystemPersona:  layers.SystemPersona,
		GlobalContext:  layers.GlobalContext,
		State:          layers.State,
		Task:           task,
	}

	return t.Execute(ctx, task, newLayers)
}

func (t *TesterAgent) WriteMutationTests(ctx context.Context, code string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Configure mutation testing para:\n%s\n\nREQUISITOS:\n- go-mutesting (Go) ou stryker (TS)\n- Mutation score >= 80%%\n- Identify equivalent mutants\n- CI integration", code)

	newLayers := PromptLayers{
		SystemPersona:  layers.SystemPersona,
		GlobalContext:  layers.GlobalContext,
		State:          layers.State,
		Task:           task,
	}

	return t.Execute(ctx, task, newLayers)
}

func (t *TesterAgent) ValidateAndFix(ctx context.Context, implementation string, spec string, layers PromptLayers) (*AgentResult, error) {
	// First validate
	validation, err := t.ValidateImplementation(ctx, implementation, spec, layers)
	if err != nil {
		return nil, err
	}

	if validation.Passed {
		return &AgentResult{Success: true, Output: "Validation passed"}, nil
	}

	// If failed, try to fix
	fixLayers := PromptLayers{
		SystemPersona:  "Você é um Engenheiro Sênior. Corrija a implementação baseada nos issues encontrados.",
		GlobalContext:  layers.GlobalContext,
		State:          layers.State,
		Task:           fmt.Sprintf("CORRIGIR IMPLEMENTAÇÃO\n\nISSUES ENCONTRADOS:\n%s\n\nIMPLEMENTAÇÃO ATUAL:\n%s\n\nESPECIFICAÇÃO:\n%s\n\nForneça implementação corrigida completa.", validation.Report, implementation, spec),
	}

	return t.Execute(ctx, "fix implementation", fixLayers)
}

func (t *TesterAgent) ValidateImplementation(ctx context.Context, implementation string, spec string, layers PromptLayers) (*ValidationResult, error) {
	validationLayers := PromptLayers{
		SystemPersona:  "Você é um Engenheiro de QA Sênior. Valide a implementação contra a especificação. Identifique gaps, bugs, missing tests, security issues, performance issues, code smells.",
		GlobalContext:  layers.GlobalContext,
		State:          layers.State,
		Task:           fmt.Sprintf("Valide:\nSPEC:\n%s\n\nIMPL:\n%s\n\nForneça relatório detalhado com: issues encontrados (Critical/High/Medium/Low), testes faltando, security issues, performance issues, code smells.", spec, implementation),
	}

	result, err := t.Execute(ctx, "validate", validationLayers)
	if err != nil {
		return nil, err
	}

	return &ValidationResult{
		Passed:          result.Success,
		Issues:          parseIssues(result.Output),
		Coverage:        extractCoverage(result.Output),
		MissingTests:    extractMissingTests(result.Output),
		SecurityIssues:  extractSecurityIssues(result.Output),
		Report:          result.Output,
	}, nil
}

// Helper functions for parsing validation results

var issueRegex = regexp.MustCompile(`(?i)(Critical|High|Medium|Low)[:\s]+([^\n]+)`)

func parseIssues(output string) []Issue {
	var issues []Issue
	matches := issueRegex.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			issues = append(issues, Issue{
				Severity:    match[1],
				Category:    "General",
				Description: strings.TrimSpace(match[2]),
			})
		}
	}
	return issues
}

var coverageRegex = regexp.MustCompile(`(?i)coverage[:\s]+(\d+(?:\.\d+)?)%`)

func extractCoverage(output string) float64 {
	matches := coverageRegex.FindStringSubmatch(output)
	if len(matches) >= 2 {
		val, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return val
		}
	}
	return 0.0
}

var missingTestRegex = regexp.MustCompile(`(?i)(missing|faltando|ausente)[:\s]+([^\n]+)`)

func extractMissingTests(output string) []string {
	var tests []string
	matches := missingTestRegex.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			tests = append(tests, strings.TrimSpace(match[2]))
		}
	}
	return tests
}

var securityRegex = regexp.MustCompile(`(?i)(security|segurança|vulnerability|vulnerabilidade|injection|xss|csrf)[:\s]+([^\n]+)`)

func extractSecurityIssues(output string) []SecurityIssue {
	var issues []SecurityIssue
	matches := securityRegex.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			issues = append(issues, SecurityIssue{
				Type:        match[1],
				Description: strings.TrimSpace(match[2]),
			})
		}
	}
	return issues
}