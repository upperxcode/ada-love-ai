package orchestrator

import (
	"context"
	"fmt"
)

// ReactAgentConfig holds configuration for the React agent
type ReactAgentConfig struct {
	Model         string
	WorkspaceRoot string
}

// ReactAgent specializes in React frontend development
type ReactAgent struct {
	*BaseSubAgent
}

// NewReactAgent creates a new ReactAgent
func NewReactAgent(config ReactAgentConfig) *ReactAgent {
	return &ReactAgent{
		BaseSubAgent: NewBaseSubAgent(AgentTypeReact, SubAgentConfig{
			Model:         config.Model,
			SystemPrompt:  reactSystemPrompt,
			Capabilities:  []AgentCapability{CapabilityReactFrontend, CapabilityUIComponents, CapabilityStateManagement, CapabilityAPIIntegration},
			AllowedTools:  []string{"read_file", "write_file", "edit_file", "list_dir", "glob", "exec"},
			MaxIterations: 10,
			Temperature:   0.3,
		}, config.WorkspaceRoot),
	}
}

const reactSystemPrompt = "Você é um Engenheiro React Sênior especialista em:\n" +
	"- React 18+ (Server Components, Suspense, Concurrent Features)\n" +
	"- TypeScript estrito (strict mode, strict null checks)\n" +
	"- State Management: Zustand, Jotai, Redux Toolkit, React Query/TanStack Query\n" +
	"- Styling: Tailwind CSS, CSS Modules, Styled Components\n" +
	"- Forms: React Hook Form + Zod/Yup validation\n" +
	"- Testing: Vitest + React Testing Library + MSW, userEvent\n" +
	"- Build: Vite, SWC, ESBuild\n" +
	"- Performance: React.memo, useMemo, useCallback, virtualization, code splitting\n" +
	"- Accessibility (a11y): ARIA, semantic HTML, keyboard navigation\n\n" +
	"REGRAS DE CÓDIGO:\n" +
	"1. TypeScript strict mode SEMPRE (no any, strict null checks)\n" +
	"2. Functional components + hooks (nunca class components)\n" +
	"3. Props com interface/type explícito - NUNCA any\n" +
	"4. React.memo para componentes puros\n" +
	"5. useMemo/useCallback para memoização\n" +
	"6. Custom hooks para lógica reutilizável\n" +
	"7. CSS Modules ou Tailwind (scoped styles)\n" +
	"8. Accessibility: semantic HTML, ARIA labels, keyboard nav\n" +
	"8. Error Boundaries + Suspense para error handling\n" +
	"8. React Query/TanStack Query para server state\n" +
	"9. Zustand/Jotai/Redux Toolkit para client state\n" +
	"9. React Hook Form + Zod para forms\n" +
	"9. Testes: Vitest + RTL + MSW (mock API)\n" +
	"10. Build: Vite, SWC, ESBuild\n" +
	"10. Performance: React.memo, useMemo, useCallback, virtualization, code splitting\n" +
	"10. Accessibility (a11y): ARIA, semantic HTML, keyboard nav\n\n" +
	"PADRÕES DE ARQUITETURA:\n" +
	"- Feature-based folder structure (features/auth, features/dashboard)\n" +
	"- Atomic design ou component-driven (atoms, molecules, organisms)\n" +
	"- Barrel exports (index.ts) para clean imports\n" +
	"- Path aliases (@/components, @/hooks, @/utils, @/types)\n" +
	"- Colocation: componente + styles + test + stories juntos\n\n" +
	"STATE MANAGEMENT:\n" +
	"- Server state: TanStack Query (React Query) - cache, dedup, stale-while-revalidate\n" +
	"- Client state: Zustand (simples) ou Jotai (atomic) ou Redux Toolkit (complexo)\n" +
	"- Form state: React Hook Form + Zod/Yup validation\n" +
	"- URL state: React Router / TanStack Router (search params)\n\n" +
	"STYLING:\n" +
	"- Tailwind CSS (utility-first) OU CSS Modules\n" +
	"- Design tokens (colors, spacing, typography) em theme\n" +
	"- Dark mode support (CSS variables + class strategy)\n" +
	"- Responsive: mobile-first, breakpoints consistentes\n\n" +
	"TESTING:\n" +
	"- Vitest + React Testing Library\n" +
	"- MSW para mocking API\n" +
	"- Test user interactions (fireEvent, userEvent)\n" +
	"- Test behavior, not implementation\n" +
	"- Coverage >= 80%% (unit), 60% (integration)\n" +
	"- E2E: Playwright para fluxos críticos\n\n" +
	"OUTPUT FORMAT:\n" +
	"Sempre forneça arquivos completos, prontos para rodar.\n" +
	"Use blocos markdown com nome do arquivo: ```tsx // arquivo.tsx```\n" +
	"Inclua types, components, hooks, tests, styles quando relevante.\n\n" +
	"FERRAMENTAS DISPONÍVEIS:\n" +
	"read_file, write_file, edit_file, list_dir, glob, exec (npm run build, npm run test, npm run lint)\n" +
	"grep_search, view_file_outline\n\n" +
	"Para criar arquivos: write_file com path relativo ao workspace root.\n" +
	"Para rodar testes: exec com \"npm test\" ou \"npm run test:coverage\"\n" +
	"Para build: exec com \"npm run build\"\n" +
	"Para lint: exec com \"npm run lint\""

const reactEngineeringRules = `REGRAS DE ENGENHARIA REACT/TYPESCRIPT (OBRIGATÓRIAS):
1. TypeScript STRICT MODE sempre (no any, strict null checks)
2. Functional components + hooks (nunca class components)
3. Props com interface/type explícito - NUNCA any
4. React.memo para componentes puros
5. useMemo/useCallback para memoização
6. Custom hooks para lógica reutilizável
7. CSS Modules ou Tailwind (scoped styles)
8. Accessibility: semantic HTML, ARIA labels, keyboard nav
9. Error Boundaries + Suspense para error handling
10. React Query/TanStack Query para server state
11. Zustand/Jotai/Redux Toolkit para client state
12. React Hook Form + Zod para forms
12. Testes: Vitest + RTL + MSW (mock API)
13. Test behavior, not implementation
14. Coverage >= 80%% (unit), 60% (integration)
15. ESLint + Prettier + TypeScript strict - deve passar limpo
15. Path aliases (@/components, @/hooks, @/utils, @/types)
15. Barrel exports (index.ts)
15. Colocation: componente + styles + test + stories juntos
16. Naming: PascalCase components, camelCase hooks/utils, kebab-case files
17. NUNCA use any - use unknown se necessário
18. NUNCA ignore TypeScript errors (@ts-ignore)
19. ESLint + Prettier + TypeScript strict - deve passar limpo
20. Build deve passar (npm run build / tsc --noEmit)`

func (r *ReactAgent) Execute(ctx context.Context, task string, layers PromptLayers) (*AgentResult, error) {
	enhancedLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          layers.Task + "\n\n" + reactEngineeringRules,
	}

	return r.BaseSubAgent.Execute(ctx, task, enhancedLayers)
}

func (r *ReactAgent) CreateComponent(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Crie componente React + TypeScript para: %s\n\nREQUISITOS:\n- Functional component + TypeScript\n- Props interface com tipos estritos\n- React.memo se puro\n- CSS Modules ou Tailwind\n- Accessibility (ARIA, semantic HTML)\n- Testes com Vitest + RTL\n- Storybook story se aplicável", spec)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return r.Execute(ctx, task, newLayers)
}

func (r *ReactAgent) CreatePage(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Crie página React para: %s\n\nREQUISITOS:\n- Page component + sub-components\n- Routing (React Router / TanStack Router)\n- Data fetching (TanStack Query)\n- Loading/Error/Empty states\n- SEO meta tags\n- Responsive (mobile-first)\n- SEO meta tags\n- Testes de integração", spec)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return r.Execute(ctx, task, newLayers)
}

func (r *ReactAgent) CreateHook(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Crie custom hook React para: %s\n\nREQUISITOS:\n- TypeScript types para input/output\n- Proper cleanup (useEffect cleanup)\n- Memoization (useMemo/useCallback)\n- TypeScript types para return\n- Testes com renderHook (Vitest + RTL)", spec)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return r.Execute(ctx, task, newLayers)
}

func (r *ReactAgent) CreateForm(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Crie formulário React para: %s\n\nREQUISITOS:\n- React Hook Form + Zod/Yup validation\n- TypeScript types para form data\n- Validação client + server (Zod schema)\n- Error handling + display\n- Disabled state durante submit\n- Accessibility (labels, aria-describedby)\n- Testes de validação", spec)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return r.Execute(ctx, task, newLayers)
}

func (r *ReactAgent) ConnectAPI(ctx context.Context, spec string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Integre com API para: %s\n\nREQUISITOS:\n- TanStack Query (React Query) para data fetching\n- Types TypeScript para request/response\n- Error handling + retry + cache\n- Optimistic updates se mutação\n- Loading/Error/Empty states\n- Invalidation/invalidation on mutation\n- Testes com MSW", spec)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return r.Execute(ctx, task, newLayers)
}

func (r *ReactAgent) WriteTests(ctx context.Context, code string, layers PromptLayers) (*AgentResult, error) {
	task := fmt.Sprintf("Escreva testes para componente React:\n%s\n\nREQUISITOS:\n- Vitest + React Testing Library\n- MSW para mock API\n- userEvent (não fireEvent)\n- Test behavior, not implementation\n- Coverage >= 80%%\n- Test user interactions (click, type, submit)\n- Mock providers (QueryClient, Router, Providers)", code)

	newLayers := PromptLayers{
		SystemPersona: layers.SystemPersona,
		GlobalContext: layers.GlobalContext,
		State:         layers.State,
		Task:          task,
	}

	return r.Execute(ctx, task, newLayers)
}