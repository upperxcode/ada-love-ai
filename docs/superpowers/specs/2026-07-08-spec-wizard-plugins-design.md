# Design: Sistema de Plugins para o SpecWizard (espelho wizard-spec)

**Data**: 2026-07-08
**Status**: Aprovado para escrita, pendente revisão do usuário
**Autor**: sessão de brainstorming

---

## 1. Contexto

O componente `SpecWizardSection.tsx` em
`frontend/src/components/spec-wizard/` tem hoje **5 conjuntos de opções
hardcoded** que não respeitam o sistema de plugins do projeto:

1. **Engineering Philosophies** (linha 252–256) — fallback para
   `['KISS', 'YAGNI', 'SOLID', 'DRY']` quando o plugin não fornece.
2. **Design Patterns** (linha 258–263) — fallback hardcoded.
3. **Data Patterns** (linha 644–650) — lista fixa
   `['DTO', 'Entity', 'Repository', 'Active Record', 'DAO']`, ignorando o plugin.
4. **Architectures** (linha 530–535) — lista fixa sem ler YAML.
5. **Persistence** (linha 515–519) — lista fixa (`custom/remote/sql/nosql/mixed`).
6. **State Management** (linha 788–794) — lista fixa
   (`redux/context/mobx/zustand/recoil/custom`).

O frontend tem um registry stub (`plugins/registry.ts`) com **3 linguagens** e
sem filtragem por escopo (`mobile`/`web`/`backend`/`global`).

O projeto de referência `/home/data/aux/dev/projetos/go/wizard-spec/spec-wizard/`
já tem o sistema completo de plugins em Go com:

- `internal/patterns/repository.go` — `PatternRepository` com categorias
  (architectures, philosophies, designPatterns, dataPatterns, stateManagements,
  dataStrategies), filtragem por `Scope` e 6 linguagens suportadas.
- `internal/registry/expert.go` — `ExpertPlugin` carregado de YAML.
- `config/experts.yaml` e `config/architectures.yaml` — fontes externas.

## 2. Decisões Tomadas

- **Origem dos plugins**: replicar estrutura Go e servir via backend HTTP.
- **Escopo**: tudo (architecture + patterns + experts).
- **Formato de configs**: espelho direto do original (compatibilidade 1:1).

## 3. Arquitetura

### 3.1 Backend Go — `pkg/patterns/`

Espelha `wizard-spec/internal/patterns/repository.go`.

```
pkg/patterns/
  patterns.go           # Pattern struct + PatternRepository
  patterns_test.go      # unit tests para filtragem por Scope
```

```go
type Pattern struct {
    ID               string   `json:"id"`
    Name             string   `json:"name"`
    Category         string   `json:"category"`
    Group            string   `json:"group"`
    Scope            string   `json:"scope"` // mobile/web/backend/global (vazio = global)
    Description      string   `json:"description"`
    IncompatibleWith []string `json:"incompatibleWith"`
    GoldenRules      []string `json:"goldenRules"`
}

type PatternRepository struct {
    Store map[string][]Pattern
}

func NewRepository() *PatternRepository
func (r *PatternRepository) GetPatternsForLanguage(lang string) []Pattern
func (r *PatternRepository) GetMultiplePatternRules(lang string, ids []string) map[string][]string
```

Constantes internas (espelhadas do original):

- `supportedLangs = []string{"flutter", "dart", "go", "python", "javascript", "typescript"}`
- Categorias: `Architecture`, `Philosophy`, `DesignPattern`, `Data`,
  `StateManagement`, `DataStrategy`.

### 3.2 Backend Go — `pkg/registry/`

Espelha `wizard-spec/internal/registry/expert.go`.

```
pkg/registry/
  expert.go             # ExpertPlugin + LoadExperts + FindExpertByLanguage
  expert_test.go
```

```go
type TestConfig struct {
    Command    string `yaml:"command" json:"command"`
    FailPrompt string `yaml:"fail_prompt" json:"fail_prompt"`
}

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

func LoadExperts(paths ...string) ([]*ExpertPlugin, error)
func FindExpertByLanguage(language string, plugins []*ExpertPlugin) (*ExpertPlugin, error)
```

### 3.3 Configs versionadas

```
config/
  experts.yaml         # copiado do wizard-spec
  architectures.yaml   # copiado do wizard-spec
```

Lidos em runtime por `LoadExperts("config/experts.yaml")`.

### 3.4 Camada HTTP

Adicionado ao `app.go`:

| Método | Rota                                   | Descrição                                         |
| ------ | -------------------------------------- | ------------------------------------------------- |
| GET    | `/api/patterns?lang=<lang>`            | Patterns filtrados por linguagem (todas categorias) |
| GET    | `/api/patterns/categories?lang=<lang>` | Agrupado por categoria                            |
| GET    | `/api/architectures`                   | Lê `config/architectures.yaml`                    |
| GET    | `/api/experts`                         | Lista `ExpertPlugin`s carregados                  |
| POST   | `/api/experts/{id}/start`              | (Fora de escopo desta entrega)    |

**Escopo MVP**: apenas os 3 GETs (`/api/patterns`, `/api/architectures`,
`/api/experts`). `POST /start` está explicitamente fora de escopo (ver
seção 7).

### 3.5 Frontend

```
frontend/src/components/spec-wizard/
  plugins/
    index.ts            # tipos LanguagePlugin, etc (manter)
    registry.ts         # REMOVER — substituído por api.ts
    api.ts              # cliente HTTP (fetch)
    usePatterns.ts      # hook React
  SpecWizardSection.tsx # substituir hardcoded por hook
```

#### `plugins/api.ts`

```ts
export async function fetchPatterns(lang: string): Promise<Pattern[]>;
export async function fetchArchitectures(): Promise<Architecture[]>;
export async function fetchExperts(): Promise<ExpertPlugin[]>;
```

Cache em memória (TTL 5 min) por linguagem para evitar refetch a cada
abertura do dialog.

#### `plugins/usePatterns.ts`

```ts
export function usePatterns(language: string): {
  architectures: Pattern[];
  philosophies: Pattern[];
  designPatterns: Pattern[];
  dataPatterns: Pattern[];
  stateManagements: Pattern[];
  dataStrategies: Pattern[];
  isLoading: boolean;
  error: string | null;
};
```

#### `SpecWizardSection.tsx`

Remove hardcoded. Consome `usePatterns(wizardState.expertLanguagePlugin)`.
Se `language === null`, renderiza `[]` (Selects ficam vazios no estado neutro).
Se erro de rede, exibe toast e mantém Selects vazios (sem fallback hardcoded
que mascararia o problema).

### 3.6 Erro handling

| Camada       | Comportamento                                                          |
| ------------ | ----------------------------------------------------------------------- |
| Backend      | Resposta 200 com payload válido sempre; erros 5xx só se YAML inválido. |
| Frontend API | `try/catch` → `error: string` populado.                                 |
| Frontend UI  | Select vazio + `<p className="text-xs text-destructive">` se erro.      |

Sem fallback para listas hardcoded: isso reintroduziria o problema que estamos
resolvendo.

## 4. Data Flow

```
[SpecWizardSection]
   └─ usePatterns(language) ──fetch──> /api/patterns?lang=go
                                          │
                                          ├─ pkg/patterns/repository.go
                                          │     └─ filtragem por Scope
                                          └─ JSON response
   └─ <Select> populado dinamicamente
```

```
[/api/experts] ──> pkg/registry/expert.go ──> LoadExperts("config/experts.yaml")
                  ──> JSON response
```

## 5. Testing

- **Unit (Go)**:
  - `pkg/patterns/patterns_test.go` — para cada `supportedLang`, verifica que
    `mobile` só retorna items com `Scope == ""` ou `"mobile"`, etc.
  - `pkg/registry/expert_test.go` — `LoadExperts` mescla YAMLs duplicados.
- **Integration (Go)**: rota `/api/patterns?lang=go` retorna o set esperado.
- **Manual**: abrir SpecWizard, mudar linguagem, verificar que cada Select
  troca suas opções.

## 6. Compatibilidade 1:1 com wizard-spec

Mapeamento direto:

| wizard-spec                                | ada-love-ai (este design)        |
| ------------------------------------------ | ------------------------------- |
| `internal/patterns/repository.go`          | `pkg/patterns/patterns.go`      |
| `internal/registry/expert.go`              | `pkg/registry/expert.go`        |
| `config/experts.yaml`                      | `config/experts.yaml`           |
| `config/architectures.yaml`                | `config/architectures.yaml`     |
| MCPServer (binário em `plugins/{lang}/`)   | Fora de escopo (esta entrega)   |

## 7. Fora de escopo (próximas specs)

- Configuração de modelo LLM "ajudante" nessas telas (mencionado pelo usuário).
- Auto-detecção de plugin via `triggers:` (`pubspec.yaml`, `go.mod`).
- `POST /api/experts/{id}/start` para iniciar binários.
- Substituição do plugin registry stub do frontend por carga dinâmica
  vinda de `/api/experts`.

## 8. Riscos

- **YAML grande**: `architectures.yaml` (~7KB) é carregado em cada request.
  Mitigar com cache em memória (sync.Once ou mapa estático).
- **Frontend sem backend rodando**: desenvolvimento local pode falhar.
  Mitigar com mensagem clara no `error`.

## 9. Critérios de sucesso

1. `SpecWizardSection.tsx` não contém nenhuma string hardcoded de filosofia
   /padrão/arquitetura/persistência/state management.
2. Trocar linguagem no `<Select>` da Phase 1 re-popula todas as opções
   subsequentes via API.
3. `go test ./pkg/patterns/...` passa com cobertura > 80% da função de filtragem.
4. TypeScript compila sem erros (`npx tsc --noEmit`).
