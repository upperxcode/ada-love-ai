import { useState, useEffect } from 'react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../ui/select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '../ui/dialog';
import { Card, CardContent } from '../ui/card';
import { BaseCard } from '../BaseCard';
import { EditDialog } from '../EditDialog';
import { ExpandableEditor } from '../ExpandableEditor';
import { IconPicker } from '../IconPicker';
import { Icon } from '../Icon';
import * as api from '../../api';

const agentTypes = [
  { value: 'executor', label: 'Executor', icon: '⚙️', description: 'Executes tasks directly using tools' },
  { value: 'delegator', label: 'Delegator', icon: '🔀', description: 'Delegates tasks to other agents/models' },
  { value: 'reviewer', label: 'Reviewer', icon: '🔍', description: 'Reviews and validates work from others' },
  { value: 'researcher', label: 'Researcher', icon: '🔬', description: 'Gathers and synthesizes information' },
];

const typeColors: Record<string, string> = {
  executor: '#3b82f6',
  delegator: '#8b5cf6',
  reviewer: '#f59e0b',
  researcher: '#10b981',
};

function AgentsSection() {
  const [agents, setAgents] = useState<api.backend.AgentConfig[]>([]);
  const [availableProviders, setAvailableProviders] = useState<string[]>([]);
  const [showEdit, setShowEdit] = useState(false);
  const [editing, setEditing] = useState<api.backend.AgentConfig | null>(null);
  const [A, setA] = useState<api.backend.AgentConfig>(
    new api.backend.AgentConfig(),
  );

  const load = async () => {
    try {
      setAgents(await api.getAgents() || []);
      setAvailableProviders(await api.listChatProviders() || []);
    } catch (e) {
      console.error('[AgentsSection] load error:', e);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const handleSave = async () => {
    let updatedAgents = [...agents];
    if (editing) {
      updatedAgents = updatedAgents.map((a) =>
        a.name === editing.name ? A : a,
      );
    } else {
      updatedAgents = [...updatedAgents, A];
    }
    setAgents(updatedAgents);
    await api.setAgents(updatedAgents);
    setShowEdit(false);
    setEditing(null);
  };

  const handleDelete = async (name: string) => {
    const updatedAgents = agents.filter((a) => a.name !== name);
    setAgents(updatedAgents);
    await api.setAgents(updatedAgents);
  };

  const openEdit = (agent: api.backend.AgentConfig | null = null) => {
    if (agent) {
      setEditing(agent);
      setA(agent);
    } else {
      setEditing(null);
      setA(
        new api.backend.AgentConfig({
          icon: '🤖',
          color: '#3b82f6',
          type: 'executor',
          temperature: 0.7,
        }),
      );
    }
    setShowEdit(true);
  };

  const groupedAgents = agents.reduce(
    (acc, agent) => {
      const type = agent.type || 'executor';
      if (!acc[type]) acc[type] = [];
      acc[type].push(agent);
      return acc;
    },
    {} as Record<string, api.backend.AgentConfig[]>,
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-foreground">Agents</h3>
          <p className="text-sm text-muted-foreground">
            Specialized models that execute and/or delegate tasks to other models.
          </p>
        </div>
        <Button size="sm" onClick={() => openEdit()} className="gap-1">
          <Icon name="Plus" size={16} /> New Agent
        </Button>
      </div>

      {agentTypes.map(({ value: type, label, icon }) => {
        const typeAgents = groupedAgents[type] || [];
        if (typeAgents.length === 0) return null;
        return (
          <div key={type}>
            <h4 className="text-sm font-medium text-muted-foreground mb-2 flex items-center gap-2">
              <span>{icon}</span>
              {label}
              <span className="text-xs">({typeAgents.length})</span>
            </h4>
            <div className="agents-grid">
              {typeAgents.map((agent) => {
                const typeMeta = agentTypes.find((t) => t.value === agent.type);
                const color = agent.color || typeColors[agent.type] || '#6b7280';
                return (
                  <BaseCard
                    key={agent.name}
                    color={color}
                    headerLeft={
                      <span className="text-xs text-white opacity-90">
                        {typeMeta?.label || agent.type}
                      </span>
                    }
                    headerRight={
                      <>
                        <button
                          className="base-card-btn"
                          onClick={() => openEdit(agent)}
                          title="Edit"
                        >
                          <Icon name="Edit" className="w-3 h-3" />
                        </button>
                        <button
                          className="base-card-btn"
                          onClick={() => handleDelete(agent.name)}
                          title="Delete"
                        >
                          <Icon name="Trash2" className="w-3 h-3" />
                        </button>
                      </>
                    }
                    icon={agent.icon || typeMeta?.icon || '🤖'}
                    title={agent.name}
                  >
                    {agent.description && (
                      <div className="base-card-desc">{agent.description}</div>
                    )}
                    <div className="flex flex-wrap gap-1 justify-center mt-1">
                      {agent.provider && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                          {agent.provider}
                        </span>
                      )}
                      {agent.model && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                          {agent.model}
                        </span>
                      )}
                      {agent.type === 'delegator' && agent.delegates && agent.delegates.length > 0 && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded bg-purple-500/15 text-purple-500">
                          → {agent.delegates.length} delegates
                        </span>
                      )}
                      {agent.max_iterations > 0 && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                          max {agent.max_iterations} iters
                        </span>
                      )}
                    </div>
                  </BaseCard>
                );
              })}
            </div>
          </div>
        );
      })}

      {agents.length === 0 && (
        <Card>
          <CardContent className="py-8">
            <p className="text-sm text-muted-foreground text-center">
              No agents yet. Create specialized task executors, delegators, reviewers, or researchers.
            </p>
          </CardContent>
        </Card>
      )}

      <EditDialog
        open={showEdit}
        onOpenChange={setShowEdit}
        title={editing ? 'Edit Agent' : 'New Agent'}
        onSave={handleSave}
      >
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Name</label>
            <Input
              value={A.name}
              onChange={(e) => setA({ ...A, name: e.target.value })}
              placeholder="e.g., Code Executor, Task Reviewer"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Description</label>
            <Input
              value={A.description}
              onChange={(e) => setA({ ...A, description: e.target.value })}
              placeholder="What this agent does"
            />
          </div>

          {/* Agent Type */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Type</label>
            <div className="grid grid-cols-2 gap-2">
              {agentTypes.map(({ value, label, icon, description }) => (
                <button
                  key={value}
                  type="button"
                  className={`flex items-start gap-2 p-2 rounded border text-left transition-colors ${
                    A.type === value
                      ? 'border-primary bg-primary/10'
                      : 'border-border hover:border-primary/50'
                  }`}
                  onClick={() => setA({ ...A, type: value, color: typeColors[value] })}
                >
                  <span className="text-base mt-0.5">{icon}</span>
                  <div className="min-w-0">
                    <div className="text-sm font-medium">{label}</div>
                    <div className="text-xs text-muted-foreground">{description}</div>
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Model Configuration */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Provider</label>
              <Select
                value={A.provider}
                onValueChange={(v) => setA({ ...A, provider: v })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select provider" />
                </SelectTrigger>
                <SelectContent>
                  {availableProviders.map((p) => (
                    <SelectItem key={p} value={p}>
                      {p}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Model</label>
              <Input
                value={A.model}
                onChange={(e) => setA({ ...A, model: e.target.value })}
                placeholder="e.g., gpt-4o, claude-3-opus"
              />
            </div>
          </div>

          {/* Execution Configuration */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Max Iterations</label>
              <Input
                type="number"
                value={A.max_iterations}
                onChange={(e) =>
                  setA({ ...A, max_iterations: parseInt(e.target.value) || 0 })
                }
                placeholder="0 = unlimited"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Temperature</label>
              <Input
                type="number"
                step="0.1"
                min="0"
                max="2"
                value={A.temperature}
                onChange={(e) =>
                  setA({ ...A, temperature: parseFloat(e.target.value) || 0 })
                }
                placeholder="0.7"
              />
            </div>
          </div>

          {/* Delegates (for delegator type) */}
          {A.type === 'delegator' && (
            <div className="space-y-2">
              <label className="text-sm font-medium">Delegates</label>
              <p className="text-xs text-muted-foreground">
                Names of agents this one can delegate tasks to (comma-separated).
              </p>
              <Input
                value={A.delegates.join(', ')}
                onChange={(e) =>
                  setA({
                    ...A,
                    delegates: e.target.value
                      .split(',')
                      .map((s) => s.trim())
                      .filter(Boolean),
                  })
                }
                placeholder="e.g., Code Executor, Researcher"
              />
            </div>
          )}

          {/* System Prompt */}
          <div className="space-y-2">
            <label className="text-sm font-medium">System Prompt</label>
            <ExpandableEditor
              label="System Prompt"
              value={A.system_prompt}
              onChange={(v) => setA({ ...A, system_prompt: v })}
            />
          </div>

          {/* Appearance */}
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1">
              <div
                className="w-10 h-10 rounded border cursor-pointer"
                style={{ backgroundColor: A.color || '#3b82f6' }}
                onClick={() =>
                  document.getElementById('agent-pick-color')?.click()
                }
              />
              <input
                id="agent-pick-color"
                type="color"
                value={A.color || '#3b82f6'}
                onChange={(e) => setA({ ...A, color: e.target.value })}
                className="w-0 h-0 opacity-0"
              />
            </div>
            <IconPicker
              value={A.icon || '🤖'}
              onChange={(v) => setA({ ...A, icon: v })}
            />
          </div>
        </div>
      </EditDialog>
    </div>
  );
}

export default AgentsSection;
