import React, { useState, useEffect } from 'react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../ui/select';
import { Card, CardContent } from '../ui/card';
import { BaseCard } from '../BaseCard';
import { EditDialog } from '../EditDialog';
import { Icon } from '../Icon';
import { Checkbox } from '../ui/checkbox';
import { Label } from '../ui/label';
import { usePatterns } from './plugins/usePatterns';
import { fetchStacks, StackTemplate } from './plugins/api';
import { StackCards } from './StackCards';
import { PatternSelector } from './PatternSelector';
import { ReviewSection } from './ReviewSection';
import { HealthBar } from './HealthBar';
import { AISuggestIcon } from './AISuggestIcon';

interface Wizard {
  id: string;
  name: string;
  step: number;
  expertLanguagePlugin: string | null;
  description?: string;
  prd?: string;
  functionalRequirements?: string[];
  nonFunctionalRequirements?: string[];
  persistence?: string;
  architecture?: string;
  engineeringPhilosophies?: string[];
  designPatterns?: string[];
  dataPatterns?: string[];
  stackConfig?: Array<{ name: string; example: string }>;
  business?: {
    stateManagement?: string;
    apiContract?: string;
    customizationDetails?: string;
    finalAdjustments?: string;
    architectureRecommendations?: string;
  };
  architectureHealth?: number;
  color?: string;
  icon?: string;
}

// Simple UUID generator (no external dependency)
const generateId = (): string => {
  return 'wizard-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
};

// Fases do wizard
const phases = [
  { id: 1, name: 'Basic Info', icon: 'Info' },
  { id: 2, name: 'Architecture', icon: 'Layers' },
  { id: 3, name: 'Stack', icon: 'Package' },
  { id: 4, name: 'Business', icon: 'Briefcase' },
  { id: 5, name: 'Review', icon: 'Check' },
];

function SpecWizardSection() {
  const [wizards, setWizards] = useState<Wizard[]>([]);
  const [showEdit, setShowEdit] = useState(false);
  const [editing, setEditing] = useState<Wizard | null>(null);
  const [currentPhase, setCurrentPhase] = useState(1);
  const [wizardState, setWizardState] = useState<
    Omit<Wizard, 'id' | 'step' | 'architectureHealth'>
  >({
    name: '',
    expertLanguagePlugin: null,
    description: '',
    prd: '',
    functionalRequirements: [],
    nonFunctionalRequirements: [],
    persistence: '',
    architecture: '',
    engineeringPhilosophies: [],
    designPatterns: [],
    dataPatterns: [],
    stackConfig: [],
    business: {
      stateManagement: '',
      apiContract: '',
      customizationDetails: '',
      finalAdjustments: '',
      architectureRecommendations: '',
    },
    color: '#3b82f6',
    icon: '📝',
  });

  // Load wizards from localStorage
  const loadWizards = (): void => {
    const saved = localStorage.getItem('spec-wizards');
    if (saved) {
      setWizards(JSON.parse(saved));
    }
  };

  // Save wizards to localStorage
  const saveWizards = (wizardsToSave: Wizard[]): void => {
    localStorage.setItem('spec-wizards', JSON.stringify(wizardsToSave));
  };

  // Handle save
  const handleSave = () => {
    let updatedWizards = [...wizards];
    if (editing) {
      updatedWizards = updatedWizards.map((w) =>
        w.id === editing.id
          ? {
              ...wizardState,
              id: w.id,
              step: 0,
              architectureHealth: calculateHealth(),
            }
          : w,
      );
    } else {
      updatedWizards = [
        ...updatedWizards,
        {
          ...wizardState,
          id: generateId(),
          step: 0,
          architectureHealth: calculateHealth(),
        },
      ];
    }
    setWizards(updatedWizards);
    saveWizards(updatedWizards);
    setShowEdit(false);
    setEditing(null);
    setWizardState({
      name: '',
      expertLanguagePlugin: null,
      description: '',
      prd: '',
      functionalRequirements: [],
      nonFunctionalRequirements: [],
      persistence: '',
      architecture: '',
      engineeringPhilosophies: [],
      designPatterns: [],
      dataPatterns: [],
      stackConfig: [],
      business: {
        stateManagement: '',
        apiContract: '',
        customizationDetails: '',
        finalAdjustments: '',
        architectureRecommendations: '',
      },
      color: '#3b82f6',
      icon: '📝',
    });
    setCurrentPhase(1);
  };

  // Handle delete
  const handleDelete = (id: string) => {
    const updatedWizards = wizards.filter((w) => w.id !== id);
    setWizards(updatedWizards);
    saveWizards(updatedWizards);
  };

  // Open edit dialog
  const openEdit = (wizard: Wizard | null = null) => {
    if (wizard) {
      setEditing(wizard);
      setWizardState({
        name: wizard.name,
        expertLanguagePlugin: wizard.expertLanguagePlugin || null,
        description: wizard.description || '',
        prd: wizard.prd || '',
        functionalRequirements: wizard.functionalRequirements || [],
        nonFunctionalRequirements: wizard.nonFunctionalRequirements || [],
        persistence: wizard.persistence || '',
        architecture: wizard.architecture || '',
        engineeringPhilosophies: wizard.engineeringPhilosophies || [],
        designPatterns: wizard.designPatterns || [],
        dataPatterns: wizard.dataPatterns || [],
        stackConfig: wizard.stackConfig || [],
        business: wizard.business || {
          stateManagement: '',
          apiContract: '',
          customizationDetails: '',
          finalAdjustments: '',
          architectureRecommendations: '',
        },
        color: wizard.color || '#3b82f6',
        icon: wizard.icon || '📝',
      });
    } else {
      setEditing(null);
      setWizardState({
        name: '',
        expertLanguagePlugin: null,
        description: '',
        prd: '',
        functionalRequirements: [],
        nonFunctionalRequirements: [],
        persistence: '',
        architecture: '',
        engineeringPhilosophies: [],
        designPatterns: [],
        dataPatterns: [],
        stackConfig: [],
        business: {
          stateManagement: '',
          apiContract: '',
          customizationDetails: '',
          finalAdjustments: '',
          architectureRecommendations: '',
        },
        color: '#3b82f6',
        icon: '📝',
      });
    }
    setCurrentPhase(1);
    setShowEdit(true);
  };

  // Update wizard state
  const updateWizardState = (field: keyof typeof wizardState, value: any) => {
    setWizardState((prev) => ({ ...prev, [field]: value }));
  };

  // Calculate architecture health (simple algorithm)
  const calculateHealth = (): number => {
    let score = 100;
    // Penalty for too many patterns
    const totalPatterns =
      (wizardState.engineeringPhilosophies?.length || 0) +
      (wizardState.designPatterns?.length || 0) +
      (wizardState.dataPatterns?.length || 0);
    if (totalPatterns > 5) score -= (totalPatterns - 5) * 5;

    // Bonus for reasonable architecture choices
    if (wizardState.architecture && wizardState.architecture !== 'custom')
      score += 5;
    if (wizardState.persistence && wizardState.persistence !== 'custom')
      score += 5;

    return Math.max(0, Math.min(100, score));
  };

  // Plugin system — consumido do backend via hook (sem hardcoded).
  const expertPlugin = wizardState.expertLanguagePlugin;
  const { patterns, experts, isLoading, error: pluginsError } =
    usePatterns(expertPlugin);

  const [stacks, setStacks] = useState<StackTemplate[]>([]);
  const [selectedStack, setSelectedStack] = useState<{name: string; example: string} | null>(null);
  const [manualStacks, setManualStacks] = useState<Array<{name: string; example: string}>>([]);

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

  useEffect(() => {
    loadWizards();
  }, []);

  // Carregar stacks quando linguagem mudar
  useEffect(() => {
    if (expertPlugin) {
      fetchStacks(expertPlugin).then(setStacks).catch(() => setStacks([]));
    } else {
      setStacks([]);
    }
  }, [expertPlugin]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-foreground">
            Spec Wizards
          </h3>
          <p className="text-sm text-muted-foreground">
            Specification wizards for generating project specifications.
          </p>
        </div>
        <Button size="sm" onClick={() => openEdit()} className="gap-1">
          <Icon name="Plus" size={16} /> New Spec Wizard
        </Button>
      </div>

      {/* Wizard List */}
      {wizards.length === 0 ? (
        <Card>
          <CardContent className="py-8">
            <p className="text-sm text-muted-foreground text-center">
              No spec wizards yet. Create your first one above!
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4">
          {wizards.map((wizard) => (
            <BaseCard
              key={wizard.id}
              color={wizard.color}
              icon={wizard.icon}
              headerLeft={
                <span className="text-xs text-white opacity-90">
                  Spec Wizard
                </span>
              }
              headerRight={
                <>
                  <button
                    className="base-card-btn"
                    onClick={() => openEdit(wizard)}
                    title="Edit"
                  >
                    <Icon name="Edit" className="w-3 h-3" />
                  </button>
                  <button
                    className="base-card-btn"
                    onClick={() => handleDelete(wizard.id)}
                    title="Delete"
                  >
                    <Icon name="Trash2" className="w-3 h-3" />
                  </button>
                </>
              }
              title={wizard.name}
            >
              {wizard.description && (
                <div className="base-card-desc">{wizard.description}</div>
              )}
              {wizard.expertLanguagePlugin && (
                <div className="flex flex-wrap gap-1 justify-center mt-1">
                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-secondary text-secondary-foreground">
                    {wizard.expertLanguagePlugin}
                  </span>
                </div>
              )}
            </BaseCard>
          ))}
        </div>
      )}

      {/* Edit Dialog */}
      <EditDialog
        open={showEdit}
        onOpenChange={setShowEdit}
        onSave={handleSave}
        title={editing ? 'Edit Spec Wizard' : 'New Spec Wizard'}
        showBack={true}
        showNext={currentPhase < 5}
        onBack={() => setCurrentPhase(currentPhase > 1 ? currentPhase - 1 : 1)}
        onNext={() => setCurrentPhase(Math.min(5, currentPhase + 1))}
        color={wizardState.color}
        icon={wizardState.icon}
        onColorChange={(color) => updateWizardState('color', color)}
        onIconChange={(icon) => updateWizardState('icon', icon)}
      >
        <div className="space-y-4 flex flex-col items-stretch justify-start">
          {pluginsError && (
            <p className="text-xs text-destructive">
              Erro ao carregar plugins: {pluginsError}
            </p>
          )}
          {/* Phase Tabs with connected dots */}
          <div className="flex items-center justify-start mb-6">
            <div className="flex items-center">
              {phases.map((phase, index) => (
                <React.Fragment key={phase.id}>
                  <div className="flex flex-col items-center">
                    <button
                      className={`flex items-center justify-center w-8 h-8 rounded-full transition-all ${
                        currentPhase === phase.id
                          ? 'bg-primary text-primary-foreground w-10 h-10'
                          : currentPhase > phase.id
                            ? 'bg-primary/50 text-primary'
                            : 'bg-muted text-muted-foreground'
                      }`}
                      onClick={() => setCurrentPhase(phase.id)}
                    >
                      <Icon name={phase.icon as any} size={14} />
                    </button>
                    <span className="text-xs text-muted-foreground mt-1">
                      {phase.name}
                    </span>
                  </div>
                  {index < phases.length - 1 && (
                    <div
                      className="w-12 h-0.5 bg-border mx-2"
                      style={{
                        background:
                          currentPhase > phase.id ? '#3b82f6' : '#94a3b8',
                      }}
                    />
                  )}
                </React.Fragment>
              ))}
            </div>
          </div>

          {/* Phase Content */}
          {currentPhase === 1 && (
            <div className="space-y-4 flex flex-col items-stretch justify-start">
              <div className="space-y-2">
                <label className="text-sm font-medium">Name</label>
                <Input
                  value={wizardState.name}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    updateWizardState('name', e.target.value)
                  }
                  placeholder="e.g., API Specification"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Description</label>
                <Input
                  value={wizardState.description}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    updateWizardState('description', e.target.value)
                  }
                  placeholder="What this spec wizard generates"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Expert Language Plugin:
                </label>
                <Select
                  value={wizardState.expertLanguagePlugin || ''}
                  onValueChange={(value) =>
                    updateWizardState('expertLanguagePlugin', value || null)
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select a language..." />
                  </SelectTrigger>
                  <SelectContent>
                    {experts.map((expert) => (
                      <SelectItem key={expert.id} value={expert.language}>
                        {expert.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  PRD (Problem Scope)
                </label>
                <div className="relative">
                  <textarea
                    value={wizardState.prd}
                    onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                      updateWizardState('prd', e.target.value)
                    }
                    placeholder="Problem scope and objectives..."
                    rows={3}
                    className="w-full px-3 py-2 pr-8 border rounded-md bg-background text-foreground resize-none"
                  />
                  <div className="absolute right-1 top-1">
                    <AISuggestIcon
                      fieldName="PRD"
                      context={JSON.stringify({ language: wizardState.expertLanguagePlugin })}
                      currentValue={wizardState.prd}
                      onApply={(value) => updateWizardState('prd', value)}
                    />
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Functional Requirements
                </label>
                <div className="relative">
                  <textarea
                    value={wizardState.functionalRequirements?.join('\n') || ''}
                    onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                      updateWizardState(
                        'functionalRequirements',
                        e.target.value
                          .split('\n')
                          .filter((l: string) => l.trim()),
                      )
                    }
                    placeholder="What the system does..."
                    rows={3}
                    className="w-full px-3 py-2 pr-8 border rounded-md bg-background text-foreground resize-none"
                  />
                  <div className="absolute right-1 top-1">
                    <AISuggestIcon
                      fieldName="Functional Requirements"
                      context={JSON.stringify({ language: wizardState.expertLanguagePlugin, prd: wizardState.prd })}
                      currentValue={wizardState.functionalRequirements?.join('\n')}
                      onApply={(value) => updateWizardState('functionalRequirements', value.split('\n').filter(Boolean))}
                    />
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Non-Functional Requirements
                </label>
                <div className="relative">
                  <textarea
                    value={
                      wizardState.nonFunctionalRequirements?.join('\n') || ''
                    }
                    onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                      updateWizardState(
                        'nonFunctionalRequirements',
                        e.target.value
                          .split('\n')
                          .filter((l: string) => l.trim()),
                      )
                    }
                    placeholder="Fast, scalable, secure..."
                    rows={3}
                    className="w-full px-3 py-2 pr-8 border rounded-md bg-background text-foreground resize-none"
                  />
                  <div className="absolute right-1 top-1">
                    <AISuggestIcon
                      fieldName="Non-Functional Requirements"
                      context={JSON.stringify({ language: wizardState.expertLanguagePlugin })}
                      currentValue={wizardState.nonFunctionalRequirements?.join('\n')}
                      onApply={(value) => updateWizardState('nonFunctionalRequirements', value.split('\n').filter(Boolean))}
                    />
                  </div>
                </div>
              </div>
            </div>
          )}

          {currentPhase === 2 && (
            <div className="space-y-4 flex flex-col items-stretch justify-start">
              <div className="space-y-2">
                <label className="text-sm font-medium">Persistência</label>
                <Select
                  value={wizardState.persistence || ''}
                  onValueChange={(value) =>
                    updateWizardState('persistence', value)
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select persistence strategy..." />
                  </SelectTrigger>
                  <SelectContent>
                    {persistenceOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Architectures:</label>
                <PatternSelector
                  options={patterns.architectures}
                  selected={wizardState.engineeringPhilosophies || []}
                  onChange={(selected) => updateWizardState('engineeringPhilosophies', selected)}
                  title=""
                />
              </div>

              <PatternSelector
                options={patterns.philosophies.map(p => ({...p, name: p.name, description: `Engineering philosophy: ${p.name}`}))}
                selected={wizardState.engineeringPhilosophies || []}
                onChange={(selected) => updateWizardState('engineeringPhilosophies', selected)}
                title="Engineering Philosophy"
              />

              <PatternSelector
                options={patterns.designPatterns.map(p => ({...p, name: p.name, description: `Design pattern: ${p.name}`}))}
                selected={wizardState.designPatterns || []}
                onChange={(selected) => updateWizardState('designPatterns', selected)}
                title="Design Patterns"
              />

              <PatternSelector
                options={patterns.dataPatterns.map(p => ({...p, name: p.name, description: `Data pattern: ${p.name}`}))}
                selected={wizardState.dataPatterns || []}
                onChange={(selected) => updateWizardState('dataPatterns', selected)}
                title="Data Patterns"
              />
            </div>
          )}

          {currentPhase === 3 && (
            <div className="space-y-4 flex flex-col items-stretch justify-start">
              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Stack Templates
                </label>
                <StackCards
                  templates={stacks}
                  selectedStack={selectedStack}
                  onSelect={(stack) => {
                    setSelectedStack(stack);
                    if (stack) {
                      updateWizardState('stackConfig', [
                        { name: stack.name, example: stack.example },
                      ]);
                    } else {
                      updateWizardState('stackConfig', []);
                    }
                  }}
                  manualStacks={manualStacks}
                  onAddManual={(stack) => {
                    setManualStacks([...manualStacks, stack]);
                    updateWizardState('stackConfig', [
                      ...(wizardState.stackConfig || []),
                      stack,
                    ]);
                  }}
                  onRemoveManual={(idx) => {
                    const newManual = manualStacks.filter((_, i) => i !== idx);
                    setManualStacks(newManual);
                    const newStackConfig = (wizardState.stackConfig || []).filter(
                      (_, i) => i !== idx + (selectedStack ? 1 : 0),
                    );
                    updateWizardState('stackConfig', newStackConfig);
                  }}
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Manual Stack Configuration
                </label>
                <div className="space-y-2">
                  {wizardState.stackConfig?.map((item, idx) => (
                    <div key={idx} className="flex gap-2">
                      <Input
                        value={item.name}
                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                          const newStack = [...(wizardState.stackConfig || [])];
                          newStack[idx] = { ...item, name: e.target.value };
                          updateWizardState('stackConfig', newStack);
                        }}
                        placeholder="Library name"
                      />
                      <Input
                        value={item.example}
                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                          const newStack = [...(wizardState.stackConfig || [])];
                          newStack[idx] = { ...item, example: e.target.value };
                          updateWizardState('stackConfig', newStack);
                        }}
                        placeholder="Example usage"
                      />
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          const newStack = (
                            wizardState.stackConfig || []
                          ).filter((_, i) => i !== idx);
                          updateWizardState('stackConfig', newStack);
                        }}
                      >
                        <Icon name="X" size={16} />
                      </Button>
                    </div>
                  ))}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      const newStack = [
                        ...(wizardState.stackConfig || []),
                        { name: '', example: '' },
                      ];
                      updateWizardState('stackConfig', newStack);
                    }}
                  >
                    <Icon name="Plus" size={16} /> Add Stack Item
                  </Button>
                </div>
              </div>
            </div>
          )}

          {currentPhase === 4 && (
            <div className="space-y-4 flex flex-col items-stretch justify-start">
              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Gerenciamento de Estado
                </label>
                <Select
                  value={wizardState.business?.stateManagement || ''}
                  onValueChange={(value) =>
                    updateWizardState('business', {
                      ...wizardState.business,
                      stateManagement: value,
                    })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select state management..." />
                  </SelectTrigger>
                  <SelectContent>
                    {stateManagementOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  API Contract / Communication
                </label>
                <div className="relative">
                  <textarea
                    value={wizardState.business?.apiContract || ''}
                    onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                      updateWizardState('business', {
                        ...wizardState.business,
                        apiContract: e.target.value,
                      })
                    }
                    placeholder="API contract details..."
                    rows={3}
                    className="w-full px-3 py-2 pr-8 border rounded-md bg-background text-foreground resize-none"
                  />
                  <div className="absolute right-1 top-1">
                    <AISuggestIcon
                      fieldName="API Contract"
                      context={JSON.stringify({ language: wizardState.expertLanguagePlugin, architecture: wizardState.architecture })}
                      currentValue={wizardState.business?.apiContract}
                      onApply={(value) => updateWizardState('business', { ...wizardState.business, apiContract: value })}
                    />
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Customization Details
                </label>
                <div className="relative">
                  <textarea
                    value={wizardState.business?.customizationDetails || ''}
                    onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                      updateWizardState('business', {
                        ...wizardState.business,
                        customizationDetails: e.target.value,
                      })
                    }
                    placeholder="Customization details..."
                    rows={3}
                    className="w-full px-3 py-2 pr-8 border rounded-md bg-background text-foreground resize-none"
                  />
                  <div className="absolute right-1 top-1">
                    <AISuggestIcon
                      fieldName="Customization Details"
                      context={JSON.stringify({ language: wizardState.expertLanguagePlugin })}
                      currentValue={wizardState.business?.customizationDetails}
                      onApply={(value) => updateWizardState('business', { ...wizardState.business, customizationDetails: value })}
                    />
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Final Adjustments
                </label>
                <div className="relative">
                  <textarea
                    value={wizardState.business?.finalAdjustments || ''}
                    onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                      updateWizardState('business', {
                        ...wizardState.business,
                        finalAdjustments: e.target.value,
                      })
                    }
                    placeholder="Final adjustments and advisor notes..."
                    rows={3}
                    className="w-full px-3 py-2 pr-8 border rounded-md bg-background text-foreground resize-none"
                  />
                  <div className="absolute right-1 top-1">
                    <AISuggestIcon
                      fieldName="Final Adjustments"
                      context={JSON.stringify({ language: wizardState.expertLanguagePlugin })}
                      currentValue={wizardState.business?.finalAdjustments}
                      onApply={(value) => updateWizardState('business', { ...wizardState.business, finalAdjustments: value })}
                    />
                  </div>
                </div>
              </div>
            </div>
          )}

          {currentPhase === 5 && (
            <div className="space-y-4 flex flex-col items-stretch justify-start">
              <HealthBar score={calculateHealth()} />
              <ReviewSection
                recommendations={
                  wizardState.business?.architectureRecommendations?.split('\n').filter(Boolean) || []
                }
                healthScore={calculateHealth()}
                items={[
                  { title: 'Name', value: wizardState.name || '—' },
                  { title: 'Architecture', value: wizardState.architecture || '—' },
                  { title: 'Persistence', value: wizardState.persistence || '—' },
                  { title: 'Stack', value: wizardState.stackConfig?.map(s => s.name).join(', ') || '—' },
                ]}
              />
              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Architecture Recommendations
                </label>
                <div className="relative">
                  <textarea
                    value={
                      wizardState.business?.architectureRecommendations || ''
                    }
                    onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                      updateWizardState('business', {
                        ...wizardState.business,
                        architectureRecommendations: e.target.value,
                      })
                    }
                    placeholder="Architecture recommendations based on your choices..."
                    rows={3}
                    className="w-full px-3 py-2 pr-8 border rounded-md bg-background text-foreground resize-none"
                  />
                  <div className="absolute right-1 top-1">
                    <AISuggestIcon
                      fieldName="Architecture Recommendations"
                      context={JSON.stringify({
                        language: wizardState.expertLanguagePlugin,
                        architecture: wizardState.architecture,
                        persistence: wizardState.persistence,
                        stack: wizardState.stackConfig?.map(s => s.name).join(', '),
                      })}
                      currentValue={wizardState.business?.architectureRecommendations}
                      onApply={(value) => updateWizardState('business', { ...wizardState.business, architectureRecommendations: value })}
                    />
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </EditDialog>
    </div>
  );
}

export default SpecWizardSection;
