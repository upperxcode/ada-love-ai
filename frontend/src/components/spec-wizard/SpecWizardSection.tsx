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
import { plugins } from './plugins/registry';

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

  // Plugin system
  const languagePlugin = wizardState.expertLanguagePlugin
    ? plugins[wizardState.expertLanguagePlugin]
    : null;

  const engineeringPhilosophies = languagePlugin?.engineeringPhilosophies || [
    'KISS',
    'YAGNI',
    'SOLID',
    'DRY',
  ];
  const designPatterns = languagePlugin?.designPatterns || [
    'Factory',
    'Builder',
    'Singleton',
    'Observer',
    'Strategy',
  ];

  useEffect(() => {
    loadWizards();
  }, []);

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
                    {plugins[wizard.expertLanguagePlugin]?.name ||
                      wizard.expertLanguagePlugin}
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
                    {Object.entries(plugins).map(([key, plugin]) => (
                      <SelectItem key={key} value={key}>
                        {plugin.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  PRD (Escopo do Problema)
                </label>
                <textarea
                  value={wizardState.prd}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                    updateWizardState('prd', e.target.value)
                  }
                  placeholder="Problem scope and objectives..."
                  rows={3}
                  className="w-full px-3 py-2 border rounded-md bg-background text-foreground resize-none"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Requisitos Funcionais
                </label>
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
                  placeholder="O que o sistema faz..."
                  rows={3}
                  className="w-full px-3 py-2 border rounded-md bg-background text-foreground resize-none"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Requisitos Não Funcionais
                </label>
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
                  placeholder="Rápido, escalável, seguro..."
                  rows={3}
                  className="w-full px-3 py-2 border rounded-md bg-background text-foreground resize-none"
                />
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
                    <SelectItem value="custom">Custom</SelectItem>
                    <SelectItem value="remote">Remote Only</SelectItem>
                    <SelectItem value="sql">SQL</SelectItem>
                    <SelectItem value="nosql">NoSQL</SelectItem>
                    <SelectItem value="mixed">Mixed</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Arquitetura</label>
                <Select
                  value={wizardState.architecture || ''}
                  onValueChange={(value) =>
                    updateWizardState('architecture', value)
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select architecture..." />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="custom">Custom</SelectItem>
                    <SelectItem value="flat">Flat/Standard</SelectItem>
                    <SelectItem value="clean">Clean Architecture</SelectItem>
                    <SelectItem value="crud">CRUD</SelectItem>
                    <SelectItem value="event-sourcing">
                      Event Sourcing
                    </SelectItem>
                    <SelectItem value="cqrs">CQRS</SelectItem>
                    <SelectItem value="mvc">MVC</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {languagePlugin && (
                <>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">
                      Filosofia de Engenharia:
                    </label>
                    <div className="flex flex-wrap gap-2">
                      {engineeringPhilosophies.map((p) => (
                        <div key={p} className="flex items-center space-x-2">
                          <Checkbox
                            id={`philosophy-${p}`}
                            checked={
                              wizardState.engineeringPhilosophies?.includes(
                                p,
                              ) || false
                            }
                            onCheckedChange={(checked: boolean | undefined) => {
                              const newPhils = checked
                                ? [
                                    ...(wizardState.engineeringPhilosophies ||
                                      []),
                                    p,
                                  ]
                                : (
                                    wizardState.engineeringPhilosophies || []
                                  ).filter((x) => x !== p);
                              updateWizardState(
                                'engineeringPhilosophies',
                                newPhils,
                              );
                            }}
                          />
                          <Label
                            htmlFor={`philosophy-${p}`}
                            className="text-sm cursor-pointer"
                          >
                            {p}
                          </Label>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="space-y-2">
                    <label className="text-sm font-medium">
                      Padrões de Design:
                    </label>
                    <div className="flex flex-wrap gap-2">
                      {designPatterns.map((p) => (
                        <div key={p} className="flex items-center space-x-2">
                          <Checkbox
                            id={`pattern-${p}`}
                            checked={
                              wizardState.designPatterns?.includes(p) || false
                            }
                            onCheckedChange={(checked: boolean | undefined) => {
                              const newPatterns = checked
                                ? [...(wizardState.designPatterns || []), p]
                                : (wizardState.designPatterns || []).filter(
                                    (x) => x !== p,
                                  );
                              updateWizardState('designPatterns', newPatterns);
                            }}
                          />
                          <Label
                            htmlFor={`pattern-${p}`}
                            className="text-sm cursor-pointer"
                          >
                            {p}
                          </Label>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="space-y-2">
                    <label className="text-sm font-medium">
                      Padrões de Dados:
                    </label>
                    <div className="flex flex-wrap gap-2">
                      {[
                        'DTO',
                        'Entity',
                        'Repository',
                        'Active Record',
                        'DAO',
                      ].map((p) => (
                        <div key={p} className="flex items-center space-x-2">
                          <Checkbox
                            id={`data-${p}`}
                            checked={
                              wizardState.dataPatterns?.includes(p) || false
                            }
                            onCheckedChange={(checked: boolean | undefined) => {
                              const newPatterns = checked
                                ? [...(wizardState.dataPatterns || []), p]
                                : (wizardState.dataPatterns || []).filter(
                                    (x) => x !== p,
                                  );
                              updateWizardState('dataPatterns', newPatterns);
                            }}
                          />
                          <Label
                            htmlFor={`data-${p}`}
                            className="text-sm cursor-pointer"
                          >
                            {p}
                          </Label>
                        </div>
                      ))}
                    </div>
                  </div>
                </>
              )}
            </div>
          )}

          {currentPhase === 3 && (
            <div className="space-y-4 flex flex-col items-stretch justify-start">
              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Stack Configurável
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
                    <SelectItem value="redux">Redux</SelectItem>
                    <SelectItem value="context">React Context</SelectItem>
                    <SelectItem value="mobx">MobX</SelectItem>
                    <SelectItem value="zustand">Zustand</SelectItem>
                    <SelectItem value="recoil">Recoil</SelectItem>
                    <SelectItem value="custom">Custom</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Contrato de API / Comunicação
                </label>
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
                  className="w-full px-3 py-2 border rounded-md bg-background text-foreground resize-none"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Detalhes de Customização e Sutilezas
                </label>
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
                  className="w-full px-3 py-2 border rounded-md bg-background text-foreground resize-none"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Ajustes Finais e Advisor
                </label>
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
                  className="w-full px-3 py-2 border rounded-md bg-background text-foreground resize-none"
                />
              </div>
            </div>
          )}

          {currentPhase === 5 && (
            <div className="space-y-4 flex flex-col items-stretch justify-start">
              <div className="space-y-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium">
                    Recomendações da Arquitetura
                  </label>
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
                    className="w-full px-3 py-2 border rounded-md bg-background text-foreground resize-none"
                  />
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-medium">
                    Saúde da Arquitetura
                  </label>
                  <div className="flex items-center gap-4">
                    <div className="flex-1">
                      <div className="w-full bg-muted rounded-full h-2.5">
                        <div
                          className="bg-green-500 h-2.5 rounded-full transition-all duration-300"
                          style={{ width: `${calculateHealth()}%` }}
                        />
                      </div>
                      <div className="flex justify-between text-xs text-muted-foreground mt-1">
                        <span>Complexidade Alta</span>
                        <span className="font-medium text-foreground">
                          {calculateHealth()}%
                        </span>
                        <span>Saúde Alta</span>
                      </div>
                    </div>
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
