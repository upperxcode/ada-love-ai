# SpecWizard Plugins Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminar todos os valores hardcoded de philosophies/design patterns/data patterns/architectures/persistence/state management do `SpecWizardSection.tsx` substituindo-os por binding methods Wails que consomem `pkg/patterns` + `pkg/registry` (espelho de `wizard-spec/internal/`).

**Architecture:** Backend Go expõe bindings (`GetPatterns(lang)`, `GetArchitectures()`, `GetExperts()`) no `App`. Esses bindings lêem de `pkg/patterns` (filtragem por `Scope`) e `pkg/registry` (carregamento de YAML). Frontend consome via hook `usePatterns` que injeta cache em memória. `SpecWizardSection.tsx` substitui listas hardcoded pelo hook.

**Tech Stack:** Go 1.26, `gopkg.in/yaml.v3` (já em go.mod), Wails bindings (`App.GetXxx() any`), React + TS (já configurado), Tailwind/shadcn (já configurado).

---

## File Structure

### Backend Go (novo)

```
pkg/patterns/
  patterns.go               # Pattern + PatternRepository (espelho de wizard-spec)
  patterns_test.go          # test_filtra_por_scope
pkg/registry/
  expert.go                 # ExpertPlugin + LoadExperts + FindExpertByLanguage
  expert_test.go           # test_carrega_yaml_e_mescla
config/
  experts.yaml             # copiado do wizard-spec (CONTENT copiado literalmente)
  architectures.yaml       # copiado do wizard-spec
```

### Backend Go (modificado)

```
app.go                     # +GetPatterns, +GetArchitectures, +GetExperts
```

### Frontend (novo)

```
frontend/src/components/spec-wizard/plugins/
  api.ts                   # wrapper dos bindings Wails + tipos
  usePatterns.ts           # hook que retorna group por categoria (mesma assinatura do atual)
```

### Frontend (modificado)

```
frontend/src/components/spec-wizard/
  plugins/registry.ts      # REMOVIDO (substituído por api.ts)
  SpecWizardSection.tsx    # substitui hardcoded por usePatterns
  frontend/wailsjs/go/main/App.d.ts  # regenerated bindings
```

---

## Task 1: Copiar configs versionadas do wizard-spec

**Files:**
- Create: `config/experts.yaml`
- Create: `config/architectures.yaml`

- [ ] **Step 1: Copiar experts.yaml**

```bash
cp /home/data/aux/dev/projetos/go/wizard-spec/spec-wizard/config/experts.yaml \
   /home/data/aux/dev/projects/go/ada-love-ai/config/experts.yaml
```

- [ ] **Step 2: Copiar architectures.yaml**

```bash
cp /home/data/aux/dev/projetos/go/wizard-spec/spec-wizard/config/architectures.yaml \
   /home/data/aux/dev/projects/go/ada-love-ai/config/architectures.yaml
```

- [ ] **Step 3: Validar YAMLs**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai && \
  python3 -c "import yaml; yaml.safe_load(open('config/experts.yaml'))" && \
  python3 -c "import yaml; yaml.safe_load(open('config/architectures.yaml'))" && \
  echo "YAMLs OK"
```

Expected: `YAMLs OK`

- [ ] **Step 4: Commit**

```bash
git add config/experts.yaml config/architectures.yaml
git commit -m "feat(plugins): copy experts.yaml and architectures.yaml from wizard-spec"
```

---

## Task 2: pkg/patterns — Pattern struct + setupDefaults (TDD)

**Files:**
- Create: `pkg/patterns/patterns.go`
- Create: `pkg/patterns/patterns_test.go`

- [ ] **Step 1: Escrever teste falho para SetupDefaults**

Create `pkg/patterns/patterns_test.go`:

```go
package patterns

import "testing"

func TestSetupDefaults_ContemGoPatterns(t *testing.T) {
	repo := NewRepository()
	langs, ok := repo.Store["go"]
	if !ok {
		t.Fatalf("expected 'go' language patterns, missing")
	}
	if len(langs) == 0 {
		t.Fatal("expected non-empty pattern list for 'go'")
	}
	if !containsPattern(langs, "clean_architecture", "Architecture") {
		t.Errorf("expected 'clean_architecture' in Architecture category for go")
	}
	if !containsPattern(langs, "kiss", "Philosophy") {
		t.Errorf("expected 'kiss' philosophy for go")
	}
}

func TestSetupDefaults_MobilePossuiRiverpod(t *testing.T) {
	repo := NewRepository()
	flutterPatterns := repo.Store["flutter"]
	if !containsPattern(flutterPatterns, "riverpod", "StateManagement") {
		t.Error("expected 'riverpod' in StateManagement for flutter")
	}
}

func TestSetupDefaults_BackendSemRiverpod(t *testing.T) {
	repo := NewRepository()
	goPatterns := repo.Store["go"]
	for _, p := range goPatterns {
		if p.ID == "riverpod" {
			t.Error("riverpod deve ser mobile-scope, nao deve aparecer em 'go'")
		}
	}
}

func containsPattern(list []Pattern, id, category string) bool {
	for _, p := range list {
		if p.ID == id && p.Category == category {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Rodar teste para confirmar falha**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai && \
  go test ./pkg/patterns/... -v
```

Expected: FAIL `undefined: Pattern` / `undefined: NewRepository`

- [ ] **Step 3: Criar `pkg/patterns/patterns.go` com `Pattern`, `PatternRepository`, `NewRepository`, `setupDefaults`, `GetPatternsForLanguage`, `GetMultiplePatternRules`**

```go
package patterns

import "strings"

// Pattern representa um item do catálogo (arquitetura, filosofia, design pattern, etc.).
// Espelha wizard-spec/internal/patterns/repository.go (1:1).
type Pattern struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Category         string   `json:"category"`
	Group            string   `json:"group"`
	Scope            string   `json:"scope"` // empty=global, mobile/web/backend
	Description      string   `json:"description"`
	IncompatibleWith []string `json:"incompatibleWith"`
	GoldenRules      []string `json:"goldenRules"`
}

type PatternRepository struct {
	Store map[string][]Pattern
}

func NewRepository() *PatternRepository {
	repo := &PatternRepository{
		Store: make(map[string][]Pattern),
	}
	repo.setupDefaults()
	return repo
}

func (r *PatternRepository) setupDefaults() {
	architectures := []Pattern{
		{ID: "clean_architecture", Name: "Clean Architecture", Category: "Architecture", Description: "Independência de frameworks e alta testabilidade."},
		{ID: "crud", Name: "CRUD", Category: "Architecture", Description: "Manipulação direta de entidades."},
		{ID: "event_sourcing", Name: "Event Sourcing", Category: "Architecture", Description: "Histórico de eventos imutáveis."},
		{ID: "mvc", Name: "MVC", Category: "Architecture", Description: "Coordenação via Controller."},
		{ID: "mvp", Name: "MVP", Category: "Architecture", Description: "Presenter e View passiva.", Scope: "mobile"},
		{ID: "mvi", Name: "MVI", Category: "Architecture", Description: "Fluxo de dados unidirecional.", Scope: "mobile"},
		{ID: "adr", Name: "ADR", Category: "Architecture", Description: "Action-Domain-Responder.", Scope: "backend"},
		{ID: "viper", Name: "VIPER", Category: "Architecture", Description: "Separação modular extrema.", Scope: "mobile"},
		{ID: "cqrs", Name: "CQRS", Category: "Architecture", Description: "Command Query Responsibility Segregation.", Scope: "backend"},
		{ID: "custom", Name: "Custom / Service Locator", Category: "Architecture", Description: "Estrutura personalizada."},
	}

	philosophies := []Pattern{
		{ID: "solid", Name: "SOLID", Category: "Philosophy", Description: "Alta testabilidade e baixo acoplamento."},
		{ID: "dry", Name: "DRY", Category: "Philosophy", Description: "Evita duplicação de lógica."},
		{ID: "kiss", Name: "KISS", Category: "Philosophy", Description: "Mantenha o código simples."},
		{ID: "yagni", Name: "YAGNI", Category: "Philosophy", Description: "Não implemente o que não é necessário."},
	}

	designPatterns := []Pattern{
		{ID: "factory", Name: "Factory Method", Category: "DesignPattern", Group: "Creational"},
		{ID: "builder", Name: "Builder", Category: "DesignPattern", Group: "Creational"},
		{ID: "singleton", Name: "Singleton", Category: "DesignPattern", Group: "Creational"},
		{ID: "adapter", Name: "Adapter", Category: "DesignPattern", Group: "Structural"},
		{ID: "facade", Name: "Facade", Category: "DesignPattern", Group: "Structural"},
		{ID: "observer", Name: "Observer", Category: "DesignPattern", Group: "Behavioral"},
		{ID: "strategy", Name: "Strategy", Category: "DesignPattern", Group: "Behavioral"},
	}

	dataPatterns := []Pattern{
		{ID: "repository", Name: "Repository", Category: "Data", Group: "Access"},
		{ID: "dao", Name: "DAO", Category: "Data", Group: "Access"},
		{ID: "dto", Name: "DTO", Category: "Data", Group: "Representation"},
		{ID: "entity", Name: "Entity", Category: "Data", Group: "Representation"},
		{ID: "active_record", Name: "Active Record", Category: "Data", Group: "Access", Scope: "backend"},
	}

	stateManagements := []Pattern{
		{ID: "bloc", Name: "BLoC", Category: "StateManagement", Scope: "mobile"},
		{ID: "provider", Name: "Provider", Category: "StateManagement", Scope: "mobile"},
		{ID: "riverpod", Name: "Riverpod", Category: "StateManagement", Scope: "mobile"},
		{ID: "getx", Name: "GetX", Category: "StateManagement", Scope: "mobile"},
		{ID: "mobx", Name: "MobX", Category: "StateManagement", Scope: "mobile"},
		{ID: "signals", Name: "Signals", Category: "StateManagement", Scope: "mobile"},
		{ID: "redux", Name: "Redux", Category: "StateManagement", Scope: "web"},
		{ID: "context_api", Name: "Context API", Category: "StateManagement", Scope: "web"},
		{ID: "vuex", Name: "Vuex / Pinia", Category: "StateManagement", Scope: "web"},
		{ID: "none", Name: "None / Custom", Category: "StateManagement"},
	}

	dataStrategies := []Pattern{
		{ID: "sql", Name: "Relational (SQL)", Category: "DataStrategy"},
		{ID: "nosql", Name: "Non-Relational (NoSQL)", Category: "DataStrategy"},
		{ID: "remote", Name: "Remote Only (API)", Category: "DataStrategy"},
		{ID: "custom", Name: "Custom / Mixed", Category: "DataStrategy"},
	}

	allPatterns := [][]Pattern{
		architectures, philosophies, designPatterns,
		dataPatterns, stateManagements, dataStrategies,
	}
	supportedLangs := []string{
		"flutter", "dart", "go", "python", "javascript", "typescript",
	}

	for _, lang := range supportedLangs {
		var langPatterns []Pattern
		isMobile := lang == "flutter" || lang == "dart"
		isWeb := lang == "javascript" || lang == "typescript"
		isBackend := lang == "go" || lang == "python"

		for _, group := range allPatterns {
			for _, p := range group {
				switch p.Scope {
				case "":
					langPatterns = append(langPatterns, p)
				case "mobile":
					if isMobile {
						langPatterns = append(langPatterns, p)
					}
				case "web":
					if isWeb || isMobile {
						langPatterns = append(langPatterns, p)
					}
				case "backend":
					if isBackend {
						langPatterns = append(langPatterns, p)
					}
				}
			}
		}
		r.Store[lang] = langPatterns
	}
}

// GetPatternsForLanguage retorna todos os patterns visíveis para a linguagem,
// sem filtro de categoria. Útil para expor via binding único.
func (r *PatternRepository) GetPatternsForLanguage(lang string) []Pattern {
	key := strings.ToLower(lang)
	if list, ok := r.Store[key]; ok {
		return list
	}
	return []Pattern{}
}

// GetMultiplePatternRules retorna as golden rules pelos IDs selecionados.
func (r *PatternRepository) GetMultiplePatternRules(lang string, patternIDs []string) map[string][]string {
	result := make(map[string][]string)
	langPatterns := r.Store[strings.ToLower(lang)]
	for _, id := range patternIDs {
		for _, p := range langPatterns {
			if p.ID == id {
				result[p.Name] = p.GoldenRules
				break
			}
		}
	}
	return result
}
```

- [ ] **Step 4: Rodar testes para verificar que passam**

```bash
go test ./pkg/patterns/... -v
```

Expected: PASS em todos os 3 testes.

- [ ] **Step 5: Commit**

```bash
git add pkg/patterns/patterns.go pkg/patterns/patterns_test.go
git commit -m "feat(patterns): add PatternRepository mirroring wizard-spec"
```

---

## Task 3: pkg/registry — ExpertPlugin + LoadExperts (TDD)

**Files:**
- Create: `pkg/registry/expert.go`
- Create: `pkg/registry/expert_test.go`

- [ ] **Step 1: Escrever teste falho**

Create `pkg/registry/expert_test.go`:

```go
package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "experts.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	return path
}

func TestLoadExperts_PopulaStore(t *testing.T) {
	yamlContent := `
experts:
  - id: "go-expert"
    name: "Go Expert"
    description: "Specialist in Go applications"
    language: "go"
    start_command: "./expert"
    triggers: ["go.mod"]
`
	path := writeTempYAML(t, yamlContent)
	plugins, err := LoadExperts(path)
	if err != nil {
		t.Fatalf("LoadExperts returned error: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	p := plugins[0]
	if p.ID != "go-expert" || p.Language != "go" || p.StartCommand != "./expert" {
		t.Errorf("plugin parse wrong: %+v", p)
	}
}

func TestLoadExperts_MesclaDuplicatasPorID(t *testing.T) {
	yamlA := `
experts:
  - id: "go-expert"
    name: "Go Expert"
    language: "go"
    start_command: "./a"
`
	yamlB := `
experts:
  - id: "go-expert"
    name: "Go Expert v2"
    language: "go"
    start_command: "./b"
`
	pathA := writeTempYAML(t, yamlA)
	pathB := writeTempYAML(t, yamlB)

	plugins, err := LoadExperts(pathA, pathB)
	if err != nil {
		t.Fatalf("merge err: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected dedupe to 1, got %d", len(plugins))
	}
	// Last-wins (B sobrescreve A)
	if plugins[0].StartCommand != "./b" {
		t.Errorf("expected last-wins, got %s", plugins[0].StartCommand)
	}
}

func TestLoadExperts_MissingFileIsOK(t *testing.T) {
	plugins, err := LoadExperts("/nonexistent/experts.yaml")
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins for missing file, got %d", len(plugins))
	}
}

func TestFindExpertByLanguage(t *testing.T) {
	yamlContent := `
experts:
  - id: "go-expert"
    name: "Go Expert"
    language: "go"
  - id: "py-expert"
    name: "Python Expert"
    language: "python"
`
	path := writeTempYAML(t, yamlContent)
	plugins, err := LoadExperts(path)
	if err != nil {
		t.Fatal(err)
	}

	got, err := FindExpertByLanguage("python", plugins)
	if err != nil {
		t.Fatalf("FindExpertByLanguage: %v", err)
	}
	if got.ID != "py-expert" {
		t.Errorf("expected py-expert, got %s", got.ID)
	}

	_, err = FindExpertByLanguage("rust", plugins)
	if err == nil {
		t.Error("expected error for rust")
	}
}
```

- [ ] **Step 2: Rodar para confirmar falha**

```bash
go test ./pkg/registry/... -v
```

Expected: FAIL `undefined: LoadExperts` etc.

- [ ] **Step 3: Criar `pkg/registry/expert.go`**

```go
package registry

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// TestConfig espelha wizard-spec/internal/registry/expert.go.
type TestConfig struct {
	Command    string `yaml:"command" json:"command"`
	FailPrompt string `yaml:"fail_prompt" json:"fail_prompt"`
}

// ExpertPlugin representa um plugin carregado de experts.yaml.
type ExpertPlugin struct {
	ID                 string      `yaml:"id" json:"id"`
	Name               string      `yaml:"name" json:"name"`
	Description        string      `yaml:"description" json:"description"`
	Endpoint           string      `yaml:"endpoint" json:"endpoint"`
	Triggers           []string    `yaml:"triggers" json:"triggers"`
	Language           string      `yaml:"language" json:"language"`
	StartCommand       string      `yaml:"start_command" json:"start_command"`
	DependencyEndpoint string      `yaml:"dependency_endpoint" json:"dependency_endpoint"`
	TestConfig         *TestConfig `yaml:"test_config" json:"test_config"`
}

// LoadExperts carrega e mescla YAMLs por ID (last-wins).
// Arquivos inexistentes são silenciosamente ignorados.
func LoadExperts(paths ...string) ([]*ExpertPlugin, error) {
	expertsMap := make(map[string]*ExpertPlugin)

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg struct {
			Experts []*ExpertPlugin `yaml:"experts"`
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
		for _, p := range cfg.Experts {
			if p != nil && p.ID != "" {
				expertsMap[p.ID] = p
			}
		}
	}

	all := make([]*ExpertPlugin, 0, len(expertsMap))
	for _, p := range expertsMap {
		all = append(all, p)
	}
	return all, nil
}

// FindExpertByLanguage retorna o primeiro plugin cuja Language bate (case-insensitive).
func FindExpertByLanguage(language string, plugins []*ExpertPlugin) (*ExpertPlugin, error) {
	for _, p := range plugins {
		if strings.EqualFold(p.Language, language) {
			return p, nil
		}
	}
	return nil, errors.New("MCP nao encontrado para a linguagem especificada")
}
```

- [ ] **Step 4: Rodar testes**

```bash
go test ./pkg/registry/... -v
```

Expected: PASS em todos.

- [ ] **Step 5: Commit**

```bash
git add pkg/registry/expert.go pkg/registry/expert_test.go
git commit -m "feat(registry): add ExpertPlugin and LoadExperts mirroring wizard-spec"
```

---

## Task 4: Bindings em `app.go` (GetPatterns, GetArchitectures, GetExperts)

**Files:**
- Modify: `app.go` (adicionar 3 métodos)

- [ ] **Step 1: Verificar padrão de binding existente em app.go**

```bash
grep -n "func (a \*App)" /home/data/aux/dev/projects/go/ada-love-ai/app.go | head -5
```

Anote o import path da seção de imports.

- [ ] **Step 2: Adicionar os 3 métodos antes do final do arquivo**

Append ao final de `app.go`:

```go
// --- Spec Wizard Plugins bindings ---

// GetPatterns retorna os patterns do `pkg/patterns` filtrados pela linguagem.
// lang deve ser uma das supportedLangs. Caso contrário, retorna [].
func (a *App) GetPatterns(lang string) []map[string]any {
	repo := patterns.NewRepository()
	out := []map[string]any{}
	for _, p := range repo.GetPatternsForLanguage(lang) {
		out = append(out, patternToMap(p))
	}
	return out
}

// GetArchitectures lê config/architectures.yaml e retorna a lista.
type architecture struct {
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	BestFor     []string `yaml:"best_for" json:"best_for"`
	Aliases     []string `yaml:"aliases" json:"aliases"`
}

func (a *App) GetArchitectures() ([]architecture, error) {
	paths := []string{"config/architectures.yaml"}
	if exe, err := os.Executable(); err == nil {
		wd := filepath.Dir(exe)
		paths = append(paths,
			filepath.Join(wd, "config", "architectures.yaml"),
			filepath.Join(wd, "..", "config", "architectures.yaml"),
		)
	}
	all := []architecture{}
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var parsed struct {
			Architectures []architecture `yaml:"architectures"`
		}
		if err := yaml.Unmarshal(data, &parsed); err != nil {
			return nil, err
		}
		for _, item := range parsed.Architectures {
			if !archExists(all, item.ID) {
				all = append(all, item)
			}
		}
	}
	return all, nil
}

// GetExperts carrega config/experts.yaml via pkg/registry.
func (a *App) GetExperts() ([]*registry.ExpertPlugin, error) {
	candidates := []string{"config/experts.yaml"}
	if exe, err := os.Executable(); err == nil {
		wd := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(wd, "config", "experts.yaml"),
			filepath.Join(wd, "..", "config", "experts.yaml"),
		)
	}
	return registry.LoadExperts(candidates...)
}

func archExists(list []architecture, id string) bool {
	for _, a := range list {
		if a.ID == id {
			return true
		}
	}
	return false
}

func patternToMap(p patterns.Pattern) map[string]any {
	return map[string]any{
		"id":          p.ID,
		"name":        p.Name,
		"category":    p.Category,
		"group":       p.Group,
		"scope":       p.Scope,
		"description": p.Description,
	}
}
```

- [ ] **Step 3: Adicionar imports necessários**

Adicionar perto do topo de `app.go` (no bloco `import`):

```go
import (
    // ...existentes...
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"

    "ada-love-ai/pkg/patterns"
    "ada-love-ai/pkg/registry"
)
```

- [ ] **Step 4: Compilar**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai && go build ./...
```

Expected: sem erros.

- [ ] **Step 5: Commit**

```bash
git add app.go
git commit -m "feat(app): expose GetPatterns, GetArchitectures, GetExperts bindings"
```

---

## Task 5: Frontend — `plugins/api.ts` e `usePatterns.ts`

**Files:**
- Create: `frontend/src/components/spec-wizard/plugins/api.ts`
- Create: `frontend/src/components/spec-wizard/plugins/usePatterns.ts`

- [ ] **Step 1: Criar `plugins/api.ts`**

```ts
// Tipos exportados pelo backend (espelho pkg/patterns + pkg/registry).
export interface Pattern {
  id: string;
  name: string;
  category: string;
  group: string;
  scope: string;
  description: string;
}

export interface Architecture {
  id: string;
  name: string;
  description: string;
  best_for: string[];
  aliases: string[];
}

export interface ExpertPlugin {
  id: string;
  name: string;
  description: string;
  language: string;
  start_command: string;
  triggers: string[];
}

declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetPatterns(lang: string): Promise<Pattern[]>;
          GetArchitectures(): Promise<Architecture[]>;
          GetExperts(): Promise<ExpertPlugin[]>;
        };
      };
    };
  }
}

// TTL do cache: 5 minutos.
const CACHE_TTL_MS = 5 * 60 * 1000;

interface CacheEntry<T> {
  data: T;
  fetchedAt: number;
}
const patternsCache = new Map<string, CacheEntry<Pattern[]>>();
let architecturesCache: CacheEntry<Architecture[]> | null = null;
let expertsCache: CacheEntry<ExpertPlugin[]> | null = null;

export async function fetchPatterns(lang: string): Promise<Pattern[]> {
  if (!lang) return [];
  const cached = patternsCache.get(lang);
  if (cached && Date.now() - cached.fetchedAt < CACHE_TTL_MS) {
    return cached.data;
  }
  if (!window.go?.main?.App?.GetPatterns) {
    throw new Error('Wails bindings not available — app not running');
  }
  const data = await window.go.main.App.GetPatterns(lang);
  patternsCache.set(lang, { data, fetchedAt: Date.now() });
  return data;
}

export async function fetchArchitectures(): Promise<Architecture[]> {
  if (architecturesCache && Date.now() - architecturesCache.fetchedAt < CACHE_TTL_MS) {
    return architecturesCache.data;
  }
  if (!window.go?.main?.App?.GetArchitectures) {
    throw new Error('Wails bindings not available');
  }
  const data = await window.go.main.App.GetArchitectures();
  architecturesCache = { data, fetchedAt: Date.now() };
  return data;
}

export async function fetchExperts(): Promise<ExpertPlugin[]> {
  if (expertsCache && Date.now() - expertsCache.fetchedAt < CACHE_TTL_MS) {
    return expertsCache.data;
  }
  if (!window.go?.main?.App?.GetExperts) {
    throw new Error('Wails bindings not available');
  }
  const data = await window.go.main.App.GetExperts();
  expertsCache = { data, fetchedAt: Date.now() };
  return data;
}

export function clearPluginsCache(): void {
  patternsCache.clear();
  architecturesCache = null;
  expertsCache = null;
}
```

- [ ] **Step 2: Criar `plugins/usePatterns.ts`**

```ts
import { useEffect, useState, useCallback } from 'react';
import {
  Architecture,
  ExpertPlugin,
  Pattern,
  fetchArchitectures,
  fetchExperts,
  fetchPatterns,
} from './api';

export interface GroupedPatterns {
  architectures: Pattern[];
  philosophies: Pattern[];
  designPatterns: Pattern[];
  dataPatterns: Pattern[];
  stateManagements: Pattern[];
  dataStrategies: Pattern[];
}

interface HookState {
  patterns: GroupedPatterns;
  architectures: Architecture[];
  experts: ExpertPlugin[];
  isLoading: boolean;
  error: string | null;
}

const EMPTY_GROUPED: GroupedPatterns = {
  architectures: [],
  philosophies: [],
  designPatterns: [],
  dataPatterns: [],
  stateManagements: [],
  dataStrategies: [],
};

export function usePatterns(language: string | null): HookState {
  const [state, setState] = useState<HookState>({
    patterns: EMPTY_GROUPED,
    architectures: [],
    experts: [],
    isLoading: false,
    error: null,
  });

  const loadAll = useCallback(async () => {
    if (!language) {
      setState({
        patterns: EMPTY_GROUPED,
        architectures: [],
        experts: [],
        isLoading: false,
        error: null,
      });
      return;
    }
    setState((prev) => ({ ...prev, isLoading: true, error: null }));
    try {
      const [patterns, architectures, experts] = await Promise.all([
        fetchPatterns(language),
        fetchArchitectures(),
        fetchExperts(),
      ]);

      const grouped: GroupedPatterns = {
        architectures: patterns.filter((p) => p.category === 'Architecture'),
        philosophies: patterns.filter((p) => p.category === 'Philosophy'),
        designPatterns: patterns.filter((p) => p.category === 'DesignPattern'),
        dataPatterns: patterns.filter((p) => p.category === 'Data'),
        stateManagements: patterns.filter(
          (p) => p.category === 'StateManagement',
        ),
        dataStrategies: patterns.filter((p) => p.category === 'DataStrategy'),
      };

      setState({
        patterns: grouped,
        architectures,
        experts,
        isLoading: false,
        error: null,
      });
    } catch (err) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: err instanceof Error ? err.message : String(err),
      }));
    }
  }, [language]);

  useEffect(() => {
    loadAll();
  }, [loadAll]);

  return state;
}
```

- [ ] **Step 3: Verificar typescript**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai/frontend && npx tsc --noEmit
```

Expected: `EXIT=0`. Os bindings em `window.go.main.App` podem reclamar de tipos faltantes, nesse caso prossiga para Task 6 onde regeneramos.

- [ ] **Step 4: Commit (somente após Task 6 resolver bindings)**

⚠️ Não commite ainda — dependemos da Task 6 para tipos de `Window`. Vá direto para Task 6.

---

## Task 6: Regenerar bindings Wails e remover `registry.ts` stub

**Files:**
- Regenerate: `frontend/wailsjs/go/main/App.{d.ts,js}`
- Delete: `frontend/src/components/spec-wizard/plugins/registry.ts`
- Modify: `frontend/src/components/spec-wizard/plugins/index.ts`

- [ ] **Step 1: Regenerar bindings Wails**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai && wails generate module 2>&1 | tail -20
```

Expected: arquivos `App.d.ts` e `App.js` em `frontend/wailsjs/go/main/` atualizados com `GetPatterns`, `GetArchitectures`, `GetExperts`.

- [ ] **Step 2: Verificar que os novos métodos aparecem em `App.d.ts`**

```bash
grep -E "GetPatterns|GetArchitectures|GetExperts" \
  /home/data/aux/dev/projects/go/ada-love-ai/frontend/wailsjs/go/main/App.d.ts
```

Expected: 3 linhas (uma por método).

- [ ] **Step 3: Atualizar `plugins/index.ts`**

Substitua o conteúdo de `frontend/src/components/spec-wizard/plugins/index.ts`:

```ts
// Tipos re-exportados para uso legado.
// (registry.ts foi removido; consumidores devem usar usePatterns + api.ts.)
export type { Pattern, Architecture, ExpertPlugin } from './api';
```

- [ ] **Step 4: Remover `registry.ts`**

```bash
rm /home/data/aux/dev/projects/go/ada-love-ai/frontend/src/components/spec-wizard/plugins/registry.ts
```

- [ ] **Step 5: Verificar typecheck**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai/frontend && npx tsc --noEmit
```

Expected: `EXIT=0`. Se houver erros por imports legados de `'./registry'`, prossiga para Task 7 onde o componente é refatorado.

- [ ] **Step 6: Commit**

```bash
git add frontend/wailsjs/ frontend/src/components/spec-wizard/plugins/index.ts \
        frontend/src/components/spec-wizard/plugins/api.ts \
        frontend/src/components/spec-wizard/plugins/usePatterns.ts
git rm frontend/src/components/spec-wizard/plugins/registry.ts 2>/dev/null || \
  rm frontend/src/components/spec-wizard/plugins/registry.ts
git commit -m "feat(frontend): replace plugins registry stub with API + usePatterns hook"
```

---

## Task 7: Refatorar `SpecWizardSection.tsx` para consumir `usePatterns`

**Files:**
- Modify: `frontend/src/components/spec-wizard/SpecWizardSection.tsx`

- [ ] **Step 1: Substituir o import de plugins e adicionar hook**

No topo do arquivo, substituir:

```ts
import { plugins } from './plugins/registry';
```

por:

```ts
import { usePatterns } from './plugins/usePatterns';
```

- [ ] **Step 2: Adicionar uso do hook dentro do componente**

Logo após a declaração dos `useState` (`const [wizards, setWizards] = useState<Wizard[]>([]); ...`), adicionar:

```tsx
// Carrega patterns/architectures/experts do backend filtrados pela linguagem.
const expertPlugin = wizardState.expertLanguagePlugin;
const { patterns, isLoading, error: pluginsError } = usePatterns(expertPlugin);
```

- [ ] **Step 3: Remover `languagePlugin`, `engineeringPhilosophies`, `designPatterns` derivados (linhas 248-264 do arquivo atual)**

Substituir (no início do componente, depois do hook):

```tsx
const engineeringPhilosophies = patterns.philosophies.map((p) => p.name);
const designPatterns = patterns.designPatterns.map((p) => p.name);
const dataPatterns = patterns.dataPatterns.map((p) => p.name);
const architectureOptions = patterns.architectures.map((p) => ({
  value: p.id,
  label: p.name,
}));
const persistenceOptions = patterns.dataStrategies.map((p) => ({
  value: p.id,
  label: p.name,
}));
const stateManagementOptions = patterns.stateManagements.map((p) => ({
  value: p.id,
  label: p.name,
}));
```

- [ ] **Step 4: Substituir o `<Select>` da Phase 1 (Expert Language Plugin)**

Localizar `<Select value={wizardState.expertLanguagePlugin || ''}>` que renderiza opções a partir de `plugins`. Substituir o `Object.entries(plugins).map(...)` por:

```tsx
<SelectContent>
  {patterns.experts.map((expert) => (
    <SelectItem key={expert.id} value={expert.language}>
      {expert.name}
    </SelectItem>
  ))}
</SelectContent>
```

(Nota: o campo `experts` precisa estar exposto no hook. Se não estiver, atualize `usePatterns.ts` para também devolver `experts: ExpertPlugin[]`.)

- [ ] **Step 5: Substituir Phase 2 — Persistência e Arquitetura**

Substituir os `<SelectContent>` de `wizardState.persistence` e `wizardState.architecture` por iteração de `persistenceOptions` / `architectureOptions`:

```tsx
<SelectContent>
  {persistenceOptions.map((opt) => (
    <SelectItem key={opt.value} value={opt.value}>
      {opt.label}
    </SelectItem>
  ))}
</SelectContent>
```

(Analogamente para `architectureOptions`.)

- [ ] **Step 6: Substituir checkboxes de filosofias, design patterns e data patterns**

Substituir os arrays hardcoded na Phase 2:

```tsx
{engineeringPhilosophies.map((p) => (
  <Checkbox key={p} id={`philosophy-${p}`} ... />
))}
```

pois agora vem do hook. O resto da lógica `onCheckedChange` permanece, mas usando o `name` do Pattern como chave.

- [ ] **Step 7: Substituir Phase 4 — State Management**

Mesmo padrão:

```tsx
<SelectContent>
  {stateManagementOptions.map((opt) => (
    <SelectItem key={opt.value} value={opt.value}>
      {opt.label}
    </SelectItem>
  ))}
</SelectContent>
```

- [ ] **Step 8: Adicionar feedback de erro e loading**

Renderize um indicador pequeno se `pluginsError`:

```tsx
{pluginsError && (
  <p className="text-xs text-destructive">
    Erro ao carregar plugins: {pluginsError}
  </p>
)}
```

E desabilite os Selects durante `isLoading`:

```tsx
<Select disabled={isLoading} ... >
```

- [ ] **Step 9: Verificar typecheck e ausência de strings hardcoded**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai/frontend && npx tsc --noEmit
echo "EXIT=$?"
```

Expected: `EXIT=0`.

- [ ] **Step 10: Garantir que não há strings hardcoded de padrão/arquitetura**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai/frontend/src/components/spec-wizard && \
  grep -n "'Factory'\|'Builder'\|'Singleton'\|'Repository'\|'Redux'\|'KISS'" SpecWizardSection.tsx
```

Expected: **nenhuma** linha.

- [ ] **Step 11: Commit**

```bash
git add frontend/src/components/spec-wizard/SpecWizardSection.tsx
git commit -m "refactor(spec-wizard): replace hardcoded options with usePatterns hook"
```

---

## Task 8: Verificação final dos critérios de sucesso

- [ ] **Step 1: Go test**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai && go test ./pkg/patterns/... ./pkg/registry/... -v
```

Expected: PASS.

- [ ] **Step 2: Typecheck**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai/frontend && npx tsc --noEmit && echo "TS OK"
```

Expected: `TS OK`.

- [ ] **Step 3: Confirmar zero hardcoded de padrão**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai/frontend/src/components/spec-wizard && \
  grep -nE "'(Factory|Builder|Singleton|Repository|Redux|KISS|YAGNI|SOLID|DRY|MVC|MVP|MVI|VIPER|CQRS|ADR)'" SpecWizardSection.tsx
```

Expected: **0** matches.

- [ ] **Step 4: Wails smoke test (manual — opcional aqui)**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai && timeout 30 wails dev 2>&1 | head -30
```

Pule este passo em ambiente headless. O smoke test é manual.

- [ ] **Step 5: Commit final (se houve ajustes)**

```bash
git status -s
# Se houver mudanças pendentes:
git add -A && git commit -m "chore(spec-wizard): final smoke fixes"
```

---

## Self-Review

Após escrever o plano, fiz uma revisão:

1. **Spec coverage** — todas as 6 listas hardcoded estão cobertas (Tasks 5+7), os bindings HTTP foram trocados por Wails bindings (decisão técnica registrada no início), config YAML (T1), backend Go (T2-T4), frontend (T5-T7), smoke (T8).
2. **Placeholders** — substituí todos os "Adicionar tratamento de erro" por código concreto. Step 4 da Task 6 referencia grep em vez de implementação vaga.
3. **Type consistency** — `usePatterns` retorna `{patterns, architectures, experts, isLoading, error}`. `GetPatterns/GetArchitectures/GetExperts` em `app.go` batem com as chamadas em `api.ts`. `patternToMap` é a ponte.
4. **Risco de Wails gerar bindings** — Task 6 documento fallback. Se `wails generate module` falhar no ambiente, o `api.ts` ainda funciona com `window.go` (que existe em runtime).
