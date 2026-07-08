# Plano de Implementação - Spec Wizard Plugin System

## FASE 1: Estrutura de Pastas e Plugins

### 1.1 Criar diretório de plugins
```bash
mkdir -p /home/data/aux/dev/projects/go/ada-love-ai/frontend/src/components/spec-wizard/plugins
```

### 1.2 Criar interface do plugin (index.ts)
```typescript
export interface LanguagePlugin {
  name: string;
  engineeringPhilosophies: string[];
  designPatterns: string[];
  dataPatterns?: string[];
}

export interface PluginRegistry {
  [language: string]: LanguagePlugin;
}
```

### 1.3 Criar registry de plugins (registry.ts)
```typescript
import { PluginRegistry } from './index';

export const plugins: PluginRegistry = {
  'go': {
    name: 'Go',
    engineeringPhilosophies: ['KISS', 'YAGNI'],
    designPatterns: ['Factory', 'Builder', 'Singleton']
  },
  'java': {
    name: 'Java',
    engineeringPhilosophies: ['SOLID', 'DRY'],
    designPatterns: ['Factory', 'Builder', 'Singleton', 'Observer', 'Strategy']
  },
  'python': {
    name: 'Python',
    engineeringPhilosophies: ['DRY', 'KISS'],
    designPatterns: ['Factory', 'Builder', 'Singleton', 'Adapter']
  }
};
```

## FASE 2: Modificar SpecWizardSection.tsx

### 2.1 Importar plugins
```typescript
import { plugins } from './plugins/registry';
```

### 2.2 Modificar lógica para usar plugins
```typescript
const expertLanguagePlugin = specWizardState.expertLanguagePlugin;
const languagePlugin = expertLanguagePlugin ? plugins[expertLanguagePlugin] : null;
```

### 2.3 Modificar listagens para usar plugins
```typescript
const engineeringPhilosophies = languagePlugin?.engineeringPhilosophies || [...];
const designPatterns = languagePlugin?.designPatterns || [...];
```

## FASE 3: Modificar SpecWizardConfig.tsx

### 3.1 Adicionar persistência para múltiplos wizards
- loadWizards(): carregar do localStorage
- saveWizards(): salvar no localStorage

### 3.2 Modificar funções CRUD
- addWizard(): salvar automaticamente
- deleteWizard(): salvar automaticamente
- editWizard(): resetar step para 0

## FASE 4: Testes
- ✅ Criar múltiplos wizards
- ✅ Alternar entre wizards
- ✅ Editar nomes
- ✅ Deletar wizards
- ✅ Seleção de linguagem
- ✅ Persistência localStorage

## RESUMO DAS MUDANÇAS

### Arquivos criados:
1. frontend/src/components/spec-wizard/plugins/index.ts
2. frontend/src/components/spec-wizard/plugins/registry.ts

### Arquivos modificados:
1. frontend/src/components/spec-wizard/SpecWizardSection.tsx
2. frontend/src/components/spec-wizard/SpecWizardConfig.tsx

### Funcionalidades:
- ✅ Sistema de plugins por linguagem
- ✅ Múltiplos Spec Wizards
- ✅ CRUD completo
- ✅ Persistência localStorage
- ✅ Plugin-dependent fields
