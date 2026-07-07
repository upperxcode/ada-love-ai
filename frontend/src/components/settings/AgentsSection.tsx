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
import { EditDialog } from '../EditDialog';
import { ExpandableEditor } from '../ExpandableEditor';
import { IconPicker } from '../IconPicker';
import { Icon } from '../Icon';
import AgentCard from './AgentCard';
import * as api from '../../api';

function AgentsSection() {
  const [agents, setAgents] = useState<api.backend.AgentConfig[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [availableProviders, setAvailableProviders] = useState<string[]>([]);
  const [showEdit, setShowEdit] = useState(false);
  const [editing, setEditing] = useState<api.backend.AgentConfig | null>(null);
  const [A, setA] = useState<api.backend.AgentConfig>(
    new api.backend.AgentConfig(),
  );
  const [showNewCategory, setShowNewCategory] = useState(false);
  const [newCategory, setNewCategory] = useState('');

  const load = async () => {
    setAgents(await api.getAgents());
    setCategories(await api.getAgentCategories());
    setAvailableProviders(await api.listChatProviders());
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

  const handleAddCategory = async () => {
    const name = newCategory.trim();
    if (!name) return;
    const updated = categories.includes(name)
      ? categories
      : [...categories, name];
    setCategories(updated);
    setA({ ...A, category: name });
    await api.setAgentCategories(updated);
    setNewCategory('');
    setShowNewCategory(false);
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
          color: '#6366f1',
        }),
      );
    }
    setShowEdit(true);
  };

  const groupedAgents = agents.reduce(
    (acc, agent) => {
      const cat = agent.category || 'Other';
      if (!acc[cat]) acc[cat] = [];
      acc[cat].push(agent);
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
            Manage your AI agents.
          </p>
        </div>
        <Button size="sm" onClick={() => openEdit()} className="gap-1">
          <Icon name="Plus" size={16} /> New Agent
        </Button>
      </div>

      <div className="agents-grid">
        {Object.entries(groupedAgents).map(([category, categoryAgents]) => (
          <div key={category} className="col-span-full">
            <h4 className="text-sm font-medium text-muted-foreground mb-2">
              {category}
            </h4>
            <div className="agents-grid">
              {categoryAgents.map((agent) => (
                <AgentCard
                  key={agent.name}
                  agent={agent}
                  onEdit={openEdit}
                  onDelete={handleDelete}
                />
              ))}
            </div>
          </div>
        ))}
      </div>

      {agents.length === 0 && (
        <Card>
          <CardContent className="py-8">
            <p className="text-sm text-muted-foreground text-center">
              No agents yet.
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
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Persona</label>
            <ExpandableEditor
              label="Persona"
              value={A.persona}
              onChange={(v) => setA({ ...A, persona: v })}
            />
          </div>
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
              />
            </div>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Category</label>
            <div className="flex items-center gap-2">
              <Select
                value={A.category}
                onValueChange={(v) => setA({ ...A, category: v })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select category" />
                </SelectTrigger>
                <SelectContent>
                  {categories.map((c) => (
                    <SelectItem key={c} value={c}>
                      {c}
                    </SelectItem>
                  ))}
                  <SelectItem value="Other">Other</SelectItem>
                </SelectContent>
              </Select>
              <Button
                type="button"
                size="sm"
                variant="outline"
                className="shrink-0 px-2"
                onClick={() => setShowNewCategory(true)}
                title="Add category"
              >
                <Icon name="Plus" size={16} />
              </Button>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1">
              <div
                className="w-10 h-10 rounded border cursor-pointer"
                style={{ backgroundColor: A.color || '#6366f1' }}
                onClick={() =>
                  document.getElementById('agent-pick-color')?.click()
                }
              />
              <input
                id="agent-pick-color"
                type="color"
                value={A.color || '#6366f1'}
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

      <Dialog open={showNewCategory} onOpenChange={setShowNewCategory}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New Category</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <Input
              value={newCategory}
              onChange={(e) => setNewCategory(e.target.value)}
              placeholder="Category name"
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleAddCategory();
              }}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowNewCategory(false)}>
              Cancel
            </Button>
            <Button onClick={handleAddCategory}>Add</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default AgentsSection;
