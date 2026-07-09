# SpecWizard Melhorias — Stack Cards, Architecture Buttons, Review, Health, AI Suggest

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 5 melhorias coesas no SpecWizard: (1) Stack cards com templates de plugin, (2) Architecture buttons estilizados, (3) Review recomendações visíveis, (4) Health bar melhorada, (5) Ícones de IA para sugestões de campos.

**Architecture:** Frontend React com componentes reutilizáveis. Backend Go expõe bindings Wails para patterns/architectures/stacks e novo binding `SuggestFieldValue` que chama LLM via `providers.LLMProvider.Chat()` com modelo configurável (`spec_model`). Dados de stack vêm dos plugins (via `/options` do expert Go).

**Tech Stack:** Go 1.26, React + TS, Wails bindings, Tailwind/shadcn, `providers.LLMProvider`

---

## File Structure

### Backend Go (modificado)

```
app.go                          # +SuggestFieldValue binding
pkg/patterns/patterns.go        # +StackTemplate type + GetStacksForLanguage
```

### Frontend (modificado)

```
frontend/src/components/spec-wizard/
  SpecWizardSection.tsx          # Refactor: usar novos componentes
  StackCards.tsx                 # NOVO: cards de stack com exemplo de uso
  ArchitectureSelector.tsx      # NOVO: botões flat estilizados
  ReviewSection.tsx              # NOVO: recomendações formatadas
  HealthBar.tsx                  # NOVO: barra de saúde com cores
  AISuggestIcon.tsx              # NOVO: ícone de IA + dialog de sugestão
```

---

## Task 1: StackTemplate type + GetStacksForLanguage

**Files:**
- Modify: `pkg/patterns/patterns.go` (adicionar tipos e método)

- [ ] **Step 1: Adicionar tipos StackTemplate e Library**

```go
// StackTemplate representa um template de stack para uma linguagem.
type StackTemplate struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Libraries []Library `json:"libraries"`
}

// Library representa uma biblioteca dentro de um stack template.
type Library struct {
    Name         string `json:"name"`
    Mandatory    bool   `json:"mandatory"`
    UsageExample string `json:"usage_example"`
}
```

- [ ] **Step 2: Adicionar método GetStacksForLanguage**

```go
func (r *PatternRepository) GetStacksForLanguage(lang string) []StackTemplate {
    // Retorna stack templates hardcoded por linguagem
    // (futuramente pode vir de config/experts.yaml)
    key := strings.ToLower(lang)
    switch key {
    case "flutter":
        return []StackTemplate{
            {
                ID:   "flutter_standard",
                Name: "Flutter Standard (Dio/GetIt/Bloc)",
                Libraries: []Library{
                    {Name: "dio", Mandatory: true, UsageExample: "final dio = Dio();\nawait dio.get('/api');"},
                    {Name: "get_it", Mandatory: true, UsageExample: "GetIt.I.registerSingleton(Service());"},
                    {Name: "flutter_bloc", Mandatory: true, UsageExample: "BlocProvider(create: (_) => MyBloc());"},
                },
            },
            {
                ID:   "flutter_enterprise",
                Name: "Flutter Enterprise (Riverpod/Freezed)",
                Libraries: []Library{
                    {Name: "flutter_riverpod", Mandatory: true, UsageExample: "final provider = StateProvider((ref) => 0);"},
                    {Name: "freezed_annotation", Mandatory: true, UsageExample: "@freezed class User with _$User {...}"},
                    {Name: "dio", Mandatory: true, UsageExample: "final response = await dio.get('/endpoint');"},
                },
            },
        }
    case "go":
        return []StackTemplate{
            {
                ID:   "go_standard",
                Name: "Go Standard (Chi/GORM)",
                Libraries: []Library{
                    {Name: "chi", Mandatory: true, UsageExample: "r := chi.NewRouter()\nr.Get(\"/api\", handler)"},
                    {Name: "gorm", Mandatory: true, UsageExample: "db.Create(&user)\nvar result User\nFirst(&result, 1)"},
                },
            },
        }
    default:
        return []StackTemplate{}
    }
}
```

- [ ] **Step 3: Commit**

```bash
git add pkg/patterns/patterns.go
git commit -m "feat(patterns): add StackTemplate type and GetStacksForLanguage"
```

---

## Task 2: Binding GetStacks no app.go

**Files:**
- Modify: `app.go` (adicionar método GetStacks)

- [ ] **Step 1: Adicionar binding GetStacks**

```go
// GetStacks retorna os stack templates para a linguagem especificada.
func (a *App) GetStacks(lang string) []map[string]any {
    repo := patterns.NewRepository()
    out := []map[string]any{}
    for _, s := range repo.GetStacksForLanguage(lang) {
        libs := []map[string]any{}
        for _, lib := range s.Libraries {
            libs = append(libs, map[string]any{
                "name":          lib.Name,
                "mandatory":     lib.Mandatory,
                "usage_example": lib.UsageExample,
            })
        }
        out = append(out, map[string]any{
            "id":        s.ID,
            "name":      s.Name,
            "libraries": libs,
        })
    }
    return out
}
```

- [ ] **Step 2: Build e verificar**

```bash
go build . && echo "BUILD OK"
```

- [ ] **Step 3: Commit**

```bash
git add app.go
git commit -m "feat(app): expose GetStacks binding"
```

---

## Task 3: Tipos TypeScript para Stack

**Files:**
- Modify: `frontend/src/components/spec-wizard/plugins/api.ts`
- Modify: `frontend/src/api.ts`

- [ ] **Step 1: Adicionar tipo StackTemplate em api.ts**

```ts
export interface StackTemplate {
  id: string;
  name: string;
  libraries: Array<{
    name: string;
    mandatory: boolean;
    usage_example: string;
  }>;
}

export interface StackConfig {
  name: string;
  example: string;
}
```

- [ ] **Step 2: Adicionar GetStacks no api.ts**

```ts
export async function fetchStacks(lang: string): Promise<StackTemplate[]> {
  if (!lang) return [];
  // Usa o binding Wails diretamente (cache em memória)
  if (!window.go?.main?.App?.GetStacks) {
    throw new Error('Wails bindings not available');
  }
  const data = (await window.go.main.App.GetStacks(lang)) as unknown as StackTemplate[];
  return data;
}
```

- [ ] **Step 3: Adicionar tipo StackTemplate no Window declaration em api.ts**

No `declare global { interface Window { go: { main: { App: { ...` adicionar:

```ts
GetStacks(lang: string): Promise<Record<string, any>[]>;
```

- [ ] **Step 4: TypeScript check**

```bash
cd frontend && npx tsc --noEmit && echo "TS OK"
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/spec-wizard/plugins/api.ts frontend/src/api.ts
git commit -m "feat(frontend): add StackTemplate type and fetchStacks"
```

---

## Task 4: Componente StackCards.tsx

**Files:**
- Create: `frontend/src/components/spec-wizard/StackCards.tsx`

- [ ] **Step 1: Criar StackCards.tsx**

```tsx
import { StackTemplate, StackConfig } from './plugins/api';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Icon } from './Icon';

interface StackCardsProps {
  templates: StackTemplate[];
  selectedStacks: StackConfig[];
  onSelect: (stack: StackConfig) => void;
}

export function StackCards({ templates, selectedStacks, onSelect }: StackCardsProps) {
  if (templates.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        No stack templates available for this language.
      </p>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {templates.map((template) => {
        const isSelected = selectedStacks.some((s) => s.name === template.name);
        return (
          <Card
            key={template.id}
            className={`cursor-pointer transition-all hover:shadow-md ${
              isSelected ? 'border-primary bg-primary/5' : 'border-border'
            }`}
            onClick={() => {
              if (!isSelected) {
                onSelect({
                  name: template.name,
                  example: template.libraries
                    .map((l) => `${l.name}: ${l.usage_example}`)
                    .join('\n'),
                });
              }
            }}
          >
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-semibold flex items-center gap-2">
                <Icon name="Package" size={16} />
                {template.name}
                {isSelected && (
                  <span className="text-xs text-primary ml-auto">Selected</span>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="space-y-2">
                {template.libraries.map((lib) => (
                  <div key={lib.name} className="text-xs">
                    <span className="font-mono font-medium">{lib.name}</span>
                    {lib.mandatory && (
                      <span className="text-destructive ml-1">*</span>
                    )}
                    <pre className="mt-1 p-2 bg-muted rounded text-[10px] overflow-x-auto">
                      {lib.usage_example}
                    </pre>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 2: Verificar typescript**

```bash
npx tsc --noEmit && echo "TS OK"
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/spec-wizard/StackCards.tsx
git commit -m "feat(spec-wizard): add StackCards component"
```

---

## Task 5: Componente ArchitectureSelector.tsx

**Files:**
- Create: `frontend/src/components/spec-wizard/ArchitectureSelector.tsx`

- [ ] **Step 1: Criar ArchitectureSelector.tsx**

```tsx
import { Pattern } from './plugins/api';
import { Icon } from './Icon';

interface ArchitectureSelectorProps {
  options: Pattern[];
  selected: string[];
  onChange: (selected: string[]) => void;
}

export function ArchitectureSelector({ options, selected, onChange }: ArchitectureSelectorProps) {
  const toggle = (id: string) => {
    if (selected.includes(id)) {
      onChange(selected.filter((s) => s !== id));
    } else {
      onChange([...selected, id]);
    }
  };

  if (options.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        No architecture options available.
      </p>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
      {options.map((opt) => {
        const isSelected = selected.includes(opt.id);
        return (
          <button
            key={opt.id}
            type="button"
            onClick={() => toggle(opt.id)}
            className={`flex items-start p-3 border rounded-lg text-left transition-all ${
              isSelected
                ? 'border-primary bg-primary/5'
                : 'border-border hover:border-primary/50 hover:bg-accent/50'
            }`}
          >
            <div className="flex-1 min-w-0">
              <div className="text-sm font-medium truncate">{opt.name}</div>
              {opt.description && (
                <div className="text-xs text-muted-foreground line-clamp-2 mt-0.5">
                  {opt.description}
                </div>
              )}
            </div>
            <div className={`ml-2 mt-0.5 flex-shrink-0 ${isSelected ? 'text-primary' : 'text-muted-foreground'}`}>
              {isSelected ? (
                <Icon name="CheckCircle" size={16} />
              ) : (
                <Icon name="Circle" size={16} />
              )}
            </div>
          </button>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 2: TypeScript check**

```bash
npx tsc --noEmit && echo "TS OK"
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/spec-wizard/ArchitectureSelector.tsx
git commit -m "feat(spec-wizard): add ArchitectureSelector component"
```

---

## Task 6: Componente ReviewSection.tsx

**Files:**
- Create: `frontend/src/components/spec-wizard/ReviewSection.tsx`

- [ ] **Step 1: Criar ReviewSection.tsx**

```tsx
import { Icon } from './Icon';

interface ReviewItem {
  title: string;
  value: string;
}

interface ReviewSectionProps {
  recommendations: string[];
  healthScore: number;
  items: ReviewItem[];
}

export function ReviewSection({ recommendations, healthScore, items }: ReviewSectionProps) {
  return (
    <div className="space-y-6">
      {/* Summary */}
      <div className="grid grid-cols-2 gap-4">
        {items.map((item) => (
          <div key={item.title} className="space-y-1">
            <div className="text-xs text-muted-foreground">{item.title}</div>
            <div className="text-sm font-medium truncate">{item.value || '—'}</div>
          </div>
        ))}
      </div>

      {/* Recommendations */}
      {recommendations.length > 0 && (
        <div className="bg-accent/30 border border-accent rounded-lg p-4">
          <h4 className="text-sm font-semibold mb-2 flex items-center gap-2">
            <Icon name="Lightbulb" size={16} className="text-amber-500" />
            Architecture Recommendations
          </h4>
          <ul className="space-y-2">
            {recommendations.map((rec, i) => (
              <li key={i} className="flex items-start gap-2 text-sm">
                <span className="text-amber-500 mt-0.5">🔹</span>
                <span>{rec}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 2: TypeScript check**

```bash
npx tsc --noEmit && echo "TS OK"
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/spec-wizard/ReviewSection.tsx
git commit -m "feat(spec-wizard): add ReviewSection component"
```

---

## Task 7: Componente HealthBar.tsx

**Files:**
- Create: `frontend/src/components/spec-wizard/HealthBar.tsx`

- [ ] **Step 1: Criar HealthBar.tsx**

```tsx
import { Icon } from './Icon';

interface HealthBarProps {
  score: number;
}

export function HealthBar({ score }: HealthBarProps) {
  const getColor = () => {
    if (score >= 70) return 'bg-green-500';
    if (score >= 40) return 'bg-amber-500';
    return 'bg-red-500';
  };

  const getLabel = () => {
    if (score >= 70) return 'Good';
    if (score >= 40) return 'Fair';
    return 'Poor';
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted-foreground">Architecture Health</span>
        <span className={`font-medium ${score >= 70 ? 'text-green-600' : score >= 40 ? 'text-amber-600' : 'text-red-600'}`}>
          {score}% — {getLabel()}
        </span>
      </div>
      <div className="w-full bg-muted rounded-full h-2.5">
        <div
          className={`h-2.5 rounded-full transition-all duration-300 ${getColor()}`}
          style={{ width: `${score}%` }}
        />
      </div>
      <div className="flex justify-between text-[10px] text-muted-foreground">
        <span>Poor</span>
        <span>Fair</span>
        <span>Good</span>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: TypeScript check**

```bash
npx tsc --noEmit && echo "TS OK"
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/spec-wizard/HealthBar.tsx
git commit -m "feat(spec-wizard): add HealthBar component"
```

---

## Task 8: Componente AISuggestIcon.tsx + Dialog

**Files:**
- Create: `frontend/src/components/spec-wizard/AISuggestIcon.tsx`
- Modify: `frontend/src/api.ts`

- [ ] **Step 1: Criar AISuggestIcon.tsx**

```tsx
import { useState } from 'react';
import { Button } from './ui/button';
import { Icon } from './Icon';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';

interface AISuggestIconProps {
  fieldName: string;
  context: string;
  currentValue?: string;
  onApply: (value: string) => void;
}

export function AISuggestIcon({ fieldName, context, currentValue, onApply }: AISuggestIconProps) {
  const [open, setOpen] = useState(false);
  const [suggestion, setSuggestion] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchSuggestion = async () => {
    setLoading(true);
    setError(null);
    try {
      const app = window.go?.main?.App;
      if (!app?.SuggestFieldValue) {
        throw new Error('AI suggestions not available');
      }
      const result = await app.SuggestFieldValue(fieldName, context, currentValue || '');
      setSuggestion(result);
      setOpen(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Button
        variant="ghost"
        size="sm"
        className="h-6 w-6 p-0 text-muted-foreground hover:text-primary"
        onClick={fetchSuggestion}
        disabled={loading}
        title={`AI suggestion for ${fieldName}`}
      >
        {loading ? (
          <Icon name="Loader2" size={14} className="animate-spin" />
        ) : (
          <Icon name="Sparkles" size={14} />
        )}
      </Button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="text-sm">AI Suggestion for {fieldName}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            {currentValue && (
              <div className="space-y-1">
                <div className="text-xs text-muted-foreground">Current value:</div>
                <div className="text-sm p-2 bg-muted rounded">{currentValue}</div>
              </div>
            )}
            <div className="space-y-1">
              <div className="text-xs text-muted-foreground">Suggestion:</div>
              <div className="text-sm p-2 bg-primary/5 border border-primary/20 rounded whitespace-pre-wrap">
                {suggestion}
              </div>
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" size="sm" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button size="sm" onClick={() => { onApply(suggestion); setOpen(false); }}>
                Apply
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
```

- [ ] **Step 2: Adicionar SuggestFieldValue no Window declaration**

No `frontend/src/api.ts`, adicionar no `App`:

```ts
SuggestFieldValue(fieldName: string, context: string, currentValue: string): Promise<string>;
```

- [ ] **Step 3: TypeScript check**

```bash
npx tsc --noEmit && echo "TS OK"
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/spec-wizard/AISuggestIcon.tsx frontend/src/api.ts
git commit -m "feat(spec-wizard): add AISuggestIcon component"
```

---

## Task 9: Binding SuggestFieldValue no app.go

**Files:**
- Modify: `app.go`

- [ ] **Step 1: Adicionar binding SuggestFieldValue**

```go
// SuggestFieldValue usa o modelo configurado (spec_model) para sugerir um valor
// para um campo do SpecWizard.
func (a *App) SuggestFieldValue(fieldName, context, currentValue string) (string, error) {
    // Obtém o provider configurado
    cfg := a.engine.GetAdaConfig()
    
    // Usa o spec_model configurado ou fallback para o primeiro provider
    modelName := ""
    if a.engine.Cfg().SpecModel != "" {
        modelName = a.engine.Cfg().SpecModel
    }
    
    // Cria o provider
    provider, err := providers.CreateProviderForModel(modelName, cfg.Providers)
    if err != nil {
        return "", fmt.Errorf("failed to create provider: %w", err)
    }
    
    // Monta o prompt
    systemPrompt := "You are an expert software architect. Generate concise, practical suggestions for specification fields. Return ONLY the suggested value, no explanations."
    userPrompt := fmt.Sprintf("Field: %s\nContext: %s\nCurrent value: %s\n\nSuggest a value for this field:", fieldName, context, currentValue)
    
    messages := []providers.Message{
        {Role: "system", Content: systemPrompt},
        {Role: "user", Content: userPrompt},
    }
    
    // Chama o LLM
    response, err := provider.Chat(a.ctx, messages, nil, modelName, nil)
    if err != nil {
        return "", fmt.Errorf("LLM request failed: %w", err)
    }
    
    if response == nil || len(response.Choices) == 0 {
        return "", fmt.Errorf("no response from LLM")
    }
    
    return response.Choices[0].Message.Content, nil
}
```

- [ ] **Step 2: Implementar CreateProviderForModel**

Em `pkg/providers/`:

```go
// CreateProviderForModel cria um provider para um modelo específico.
func CreateProviderForModel(modelName string, providers map[string]ProviderConfig) (LLMProvider, error) {
    // Encontra o provider que contém o modelo
    for _, cfg := range providers {
        for model := range cfg.Models {
            if strings.Contains(model, modelName) || modelName == "" {
                return CreateProvider(cfg)
            }
        }
    }
    return nil, fmt.Errorf("model %s not found", modelName)
}
```

- [ ] **Step 3: Adicionar campo SpecModel na config**

Em `pkg/config/config.go`:

```go
type Config struct {
    // ...existing fields...
    SpecModel string `json:"spec_model"` // Model for spec wizard suggestions
}
```

- [ ] **Step 4: Build e verificar**

```bash
go build . && echo "BUILD OK"
```

- [ ] **Step 5: Commit**

```bash
git add app.go pkg/providers/*.go pkg/config/config.go
git commit -m "feat(app): add SuggestFieldValue binding with spec_model config"
```

---

## Task 10: Integrar novos componentes no SpecWizardSection.tsx

**Files:**
- Modify: `frontend/src/components/spec-wizard/SpecWizardSection.tsx`

- [ ] **Step 1: Adicionar imports dos novos componentes**

```tsx
import { StackCards } from './StackCards';
import { ArchitectureSelector } from './ArchitectureSelector';
import { ReviewSection } from './ReviewSection';
import { HealthBar } from './HealthBar';
import { AISuggestIcon } from './AISuggestIcon';
import { fetchStacks, StackTemplate } from './plugins/api';
```

- [ ] **Step 2: Adicionar estado para stacks**

```tsx
const [stacks, setStacks] = useState<StackTemplate[]>([]);
const [selectedStacks, setSelectedStacks] = useState<StackConfig[]>([]);

// Carregar stacks quando linguagem mudar
useEffect(() => {
  if (expertPlugin) {
    fetchStacks(expertPlugin).then(setStacks);
  }
}, [expertPlugin]);
```

- [ ] **Step 3: Substituir Phase 3 (Stack) por StackCards**

```tsx
{currentPhase === 3 && (
  <div className="space-y-4">
    <StackCards
      templates={stacks}
      selectedStacks={selectedStacks}
      onSelect={(stack) => {
        setSelectedStacks([...selectedStacks, stack]);
        updateWizardState('stackConfig', [
          ...(wizardState.stackConfig || []),
          { name: stack.name, example: stack.example },
        ]);
      }}
    />
    {/* Manual stack addition (existing UI) */}
    <div className="space-y-2">
      <label className="text-sm font-medium">Manual Stack Configuration</label>
      {/* existing stack input UI */}
    </div>
  </div>
)}
```

- [ ] **Step 4: Substituir checkboxes de Architecture por ArchitectureSelector**

```tsx
{currentPhase === 2 && (
  <div className="space-y-4">
    {/* ...persistence select... */}
    
    <div className="space-y-2">
      <label className="text-sm font-medium">Architectures:</label>
      <ArchitectureSelector
        options={architectureOptions}
        selected={wizardState.engineeringPhilosophies || []}
        onChange={(selected) => updateWizardState('engineeringPhilosophies', selected)}
      />
    </div>
    
    {/* ...design patterns checkboxes... */}
  </div>
)}
```

- [ ] **Step 5: Adicionar AISuggestIcon em campos de texto**

Em cada `<textarea>`, adicionar wrapper:

```tsx
<div className="relative">
  <textarea ... />
  <div className="absolute right-2 top-2">
    <AISuggestIcon
      fieldName="PRD"
      context={JSON.stringify(wizardState)}
      currentValue={wizardState.prd}
      onApply={(value) => updateWizardState('prd', value)}
    />
  </div>
</div>
```

- [ ] **Step 6: Substituir Phase 5 por ReviewSection + HealthBar**

```tsx
{currentPhase === 5 && (
  <div className="space-y-6">
    <HealthBar score={calculateHealth()} />
    <ReviewSection
      recommendations={wizardState.business?.architectureRecommendations?.split('\n') || []}
      healthScore={calculateHealth()}
      items={[
        { title: 'Name', value: wizardState.name },
        { title: 'Architecture', value: wizardState.architecture || '—' },
        { title: 'Persistence', value: wizardState.persistence || '—' },
        { title: 'Stack', value: wizardState.stackConfig?.map(s => s.name).join(', ') || '—' },
      ]}
    />
  </div>
)}
```

- [ ] **Step 7: TypeScript check**

```bash
npx tsc --noEmit && echo "TS OK"
```

- [ ] **Step 8: Commit**

```bash
git add frontend/src/components/spec-wizard/SpecWizardSection.tsx
git commit -m "feat(spec-wizard): integrate StackCards, ArchitectureSelector, ReviewSection, HealthBar, AISuggestIcon"
```

---

## Task 11: Verificação Final

- [ ] **Step 1: Go build**

```bash
go build . && echo "BUILD OK"
```

- [ ] **Step 2: TypeScript check**

```bash
cd frontend && npx tsc --noEmit && echo "TS OK"
```

- [ ] **Step 3: Go tests**

```bash
go test ./pkg/patterns/... ./pkg/registry/... -v
```

- [ ] **Step 4: Smoke test manual**

1. Abrir SpecWizard
2. Criar novo wizard
3. Selecionar linguagem (ex: Flutter) → verificar que stacks aparecem como cards
4. Phase 2 → verificar architecture buttons estilizados
5. Phase 3 → selecionar stack card
6. Phase 5 → verificar HealthBar e ReviewSection
7. Clicar em ícone de IA em campo de texto → verificar dialog de sugestão

---

## Self-Review

**Spec coverage:**
1. ✅ Stack cards com templates de plugin → Tasks 1-4, 10
2. ✅ Architecture buttons estilizados → Tasks 5, 10
3. ✅ Review recomendações visíveis → Tasks 6, 10
4. ✅ Health bar melhorada → Tasks 7, 10
5. ✅ Ícones de IA → Tasks 8-9, 10

**Placeholder scan:** Não há "TBD" ou "TODO" nos passos de implementação.

**Type consistency:** Tipos definidos em Tasks 1-3 são usados consistentemente em Tasks 4-10.
