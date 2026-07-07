import { useState, useEffect } from 'react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Switch } from './ui/switch';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from './ui/dialog';
import { Card, CardContent } from './ui/card';
import { Separator } from './ui/separator';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select';
import { Combobox } from './ui/combobox';
import { BaseCard } from './BaseCard';
import { EditDialog } from './EditDialog';
import { ExpandableEditor } from './ExpandableEditor';
import { IconPicker } from './IconPicker';
import { Icon } from './Icon';
import * as api from '../api';
import { useTheme } from '../lib/theme';

const categoryColors: Record<string, string> = {
  'File System': '#3b82f6',
  Git: '#f59e0b',
  'Code Search': '#10b981',
  Testing: '#8b5cf6',
  Build: '#ec4899',
  Web: '#22c55e',
  Media: '#eab308',
  MCP: '#a855f7',
  Communication: '#8b5cf6',
  Shell: '#ef4444',
  'Scheduled Tasks': '#14b8a6',
  Memory: '#f43f5e',
  Knowledge: '#06b6d4',
  Hardware: '#06b6d4',
  Skills: '#a855f7',
  Agent: '#f97316',
  Other: '#6b7280',
};

const settingsSections = [
  { id: 'general', icon: 'Settings', label: 'General' },
  { id: 'workspace', icon: 'Folder', label: 'Workspace' },
  { id: 'agents', icon: 'User', label: 'Agents' },
  { id: 'skills', icon: 'Brain', label: 'Skills' },
  { id: 'tools', icon: 'Wrench', label: 'Tools' },
  { id: 'models', icon: 'Cpu', label: 'Models' },
];

function GeneralSection() {
  const {
    themes,
    currentTheme,
    setTheme,
    iconThemes,
    currentIconTheme,
    setIconTheme,
    iconSets,
    currentIconSet,
    setIconSet,
  } = useTheme();

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-foreground">General</h3>
        <p className="text-sm text-muted-foreground">
          Theme and application settings.
        </p>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium text-foreground">Theme</h4>
        <div className="flex flex-wrap gap-3">
          {themes.map((t) => (
            <button
              key={t.id}
              className={`flex flex-col gap-2 p-3 rounded-lg border text-sm transition-colors w-48 ${
                currentTheme === t.id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setTheme(t.id)}
            >
              <div className="flex items-center justify-between">
                <span className="font-medium">{t.name}</span>
                {t.author && (
                  <span className="text-xs text-muted-foreground">
                    {t.author}
                  </span>
                )}
              </div>
              <div className="flex gap-1.5">
                <span
                  className="w-5 h-5 rounded-sm"
                  style={{ backgroundColor: t.light?.primary }}
                />
                <span
                  className="w-5 h-5 rounded-sm"
                  style={{ backgroundColor: t.light?.secondary }}
                />
                <span
                  className="w-5 h-5 rounded-sm"
                  style={{ backgroundColor: t.light?.accent }}
                />
                <span
                  className="w-5 h-5 rounded-sm"
                  style={{ backgroundColor: t.light?.muted }}
                />
                <span
                  className="w-5 h-5 rounded-sm"
                  style={{ backgroundColor: t.light?.destructive }}
                />
              </div>
              <div className="flex gap-1.5 mt-1">
                <span
                  className="px-2 py-0.5 text-[10px] rounded-sm text-white"
                  style={{ backgroundColor: t.light?.primary }}
                >
                  Button
                </span>
                <span
                  className="px-2 py-0.5 text-[10px] rounded-sm"
                  style={{
                    backgroundColor: t.light?.secondary,
                    color: t.light?.['secondary-foreground'],
                  }}
                >
                  Button
                </span>
                <span
                  className="px-2 py-0.5 text-[10px] rounded-sm border"
                  style={{
                    borderColor: t.light?.border,
                    color: t.light?.['secondary-foreground'],
                  }}
                >
                  Outline
                </span>
              </div>
            </button>
          ))}
        </div>
        <p className="text-xs text-muted-foreground">
          Install:{' '}
          <code className="bg-muted px-1 rounded">
            npm run theme:add &lt;tweakcn-url&gt; "Theme Name"
          </code>
        </p>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium text-foreground">Icon Style</h4>
        <div className="flex flex-wrap gap-2">
          {iconThemes.map((t) => (
            <button
              key={t.id}
              className={`px-3 py-1.5 rounded-lg border text-xs transition-colors ${
                currentIconTheme === t.id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setIconTheme(t.id)}
            >
              {t.name}
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium text-foreground">Icon Set</h4>
        <div className="flex flex-wrap gap-2">
          {iconSets.map((t) => (
            <button
              key={t.id}
              className={`px-3 py-1.5 rounded-lg border text-xs transition-colors ${
                currentIconSet === t.id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setIconSet(t.id)}
            >
              {t.name}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
function WorkspacesSection() {
  const [workspaces, setWorkspaces] = useState<api.backend.WorkspaceConfig[]>(
    [],
  );
  const [showAdd, setShowAdd] = useState(false);
  const [showEdit, setShowEdit] = useState(false);
  const [editing, setEditing] = useState<api.backend.WorkspaceConfig | null>(null);
  const [showAddTool, setShowAddTool] = useState(false);
  const [selectedFilterProfileId, setSelectedFilterProfileId] = useState<number | null>(null);
  const [availableTools, setAvailableTools] = useState<
    api.backend.ToolUIInfo[]
  >([]);
  const [availableProfiles, setAvailableProfiles] = useState<
    api.backend.ToolProfile[]
  >([]);

  useEffect(() => {
    api.getAvailableTools().then(setAvailableTools);
    api.getToolProfiles().then(setAvailableProfiles);
  }, []);
  const [knownTools, setKnownTools] = useState<api.backend.ToolUIInfo[]>([]);
  const [knownSkills, setKnownSkills] = useState<string[]>([]);
  const [E, setE] = useState({
    title: '',
    description: '',
    color: '#3b82f6',
    icon: '📂',
    personality: '',
    folders: [] as string[],
    knowledge: [] as string[],
    workspace_agents: [] as string[],
    skills: [] as string[],
    tools: [] as string[],
    enabled: true,
    maxPromptSend: 0,
    commitChanges: true,
    maxContextLength: 0,
  });
  const [A, setA] = useState({ title: '', personality: '' });
  const [selectedField, setSelectedField] = useState('folders');
  const fieldMap: Record<
    string,
    { state: string[]; set: (v: string[]) => void }
  > = {
    folders: {
      state: E.folders,
      set: (v) => setE((prev) => ({ ...prev, folders: v })),
    },
    knowledge: {
      state: E.knowledge,
      set: (v) => setE((prev) => ({ ...prev, knowledge: v })),
    },
    workspace_agents: {
      state: E.workspace_agents,
      set: (v) => setE((prev) => ({ ...prev, workspace_agents: v })),
    },
    skills: {
      state: E.skills,
      set: (v) => setE((prev) => ({ ...prev, skills: v })),
    },
    tools: {
      state: E.tools,
      set: (v) => setE((prev) => ({ ...prev, tools: v })),
    },
  };

  const computeTokens = (ws: api.backend.WorkspaceConfig) => {
    let t = 2000;
    for (const f of ws.folders || []) t += f.length * 10;
    for (const a of ws.workspace_agents || []) t += 500;
    for (const s of ws.skills || []) t += 300;
    return t.toLocaleString();
  };

  const load = () => {
    api.getWorkspaces().then(setWorkspaces);
    api.getAvailableTools().then(setKnownTools);
  };
  useEffect(() => {
    load();
    api.getAgentCategories().then(setKnownSkills);
  }, []);

  const handleAdd = async () => {
    if (!A.title.trim()) return;
    await api.addWorkspace(A.title.trim(), '', A.personality.trim());
    setShowAdd(false);
    setA({ title: '', personality: '' });
    load();
  };

  const handleDelete = (title: string) => api.deleteWorkspace(title).then(load);
  const handleToggle = (title: string) => api.toggleWorkspace(title).then(load);

  const toggleArrayItem = (arr: string[], item: string) =>
    arr.includes(item) ? arr.filter((x) => x !== item) : [...arr, item];

  const openEdit = (ws: api.backend.WorkspaceConfig) => {
    setEditing(ws);
    setE({
      title: ws.title,
      description: ws.description,
      color: ws.color || '#3b82f6',
      icon: ws.icon || '📂',
      personality: ws.personality,
      folders: ws.folders || [],
      knowledge: ws.knowledge || [],
      workspace_agents: ws.workspace_agents || [],
      skills: ws.skills || [],
      tools: ws.tools || [],
      enabled: ws.enabled,
      maxPromptSend: ws.max_prompt_send || 0,
      commitChanges: ws.commit_changes !== false,
      maxContextLength: ws.max_context_length || 0,
    });
    setShowEdit(true);
  };

  const handleSaveEdit = async () => {
    if (!editing) return;
    await api.updateWorkspace(editing.title, {
      ...editing,
      title: E.title,
      description: E.description,
      color: E.color,
      icon: E.icon,
      personality: E.personality,
      folders: E.folders,
      knowledge: E.knowledge,
      workspace_agents: E.workspace_agents,
      skills: E.skills,
      tools: E.tools,
      enabled: E.enabled,
      max_prompt_send: E.maxPromptSend,
      commit_changes: E.commitChanges,
      max_context_length: E.maxContextLength,
    });
    setShowEdit(false);
    setEditing(null);
    load();
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-foreground">Workspaces</h3>
          <p className="text-sm text-muted-foreground">
            Manage your workspaces.
          </p>
        </div>
        <Button size="sm" onClick={() => setShowAdd(true)} className="gap-1">
          <Icon name="Plus" size={16} /> New Workspace
        </Button>
      </div>

      <div className="agents-grid">
        {workspaces.map((ws) => (
          <BaseCard
            key={ws.title}
            color={ws.enabled ? ws.color || '#3b82f6' : '#6b7280'}
            inactive={!ws.enabled}
            headerLeft={
              ws.enabled ? (
                <span className="text-xs text-green-400">✔ Active</span>
              ) : null
            }
            headerRight={
              <>
                <button
                  className="base-card-btn"
                  onClick={() => openEdit(ws)}
                  title="Edit"
                >
                  <Icon name="Edit" className="w-3 h-3" />
                </button>
                <button
                  className="base-card-btn"
                  onClick={() => handleToggle(ws.title)}
                  title="Toggle"
                >
                  <Icon name={ws.enabled ? 'Check' : 'X'} className="w-3 h-3" />
                </button>
                <button
                  className="base-card-btn"
                  onClick={() => handleDelete(ws.title)}
                  title="Delete"
                >
                  <Icon name="Trash2" className="w-3 h-3" />
                </button>
              </>
            }
            icon={ws.icon || '📂'}
            title={ws.title}
          >
            <div className="base-card-desc">
              {ws.description || 'No description'}
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              {ws.folders?.length || 0} folders ·{' '}
              {ws.workspace_agents?.length || 0} agents ·{' '}
              {ws.skills?.length || 0} skills
            </div>
          </BaseCard>
        ))}
        {workspaces.length === 0 && (
          <div className="col-span-full text-center py-12 text-muted-foreground text-sm">
            No workspaces yet.
          </div>
        )}
      </div>

      <Dialog open={showAdd} onOpenChange={setShowAdd}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New Workspace</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Title</label>
              <Input
                value={A.title}
                onChange={(e) => setA({ ...A, title: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Personality</label>
              <Input
                value={A.personality}
                onChange={(e) => setA({ ...A, personality: e.target.value })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAdd(false)}>
              Cancel
            </Button>
            <Button onClick={handleAdd}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <EditDialog
        open={showEdit}
        onOpenChange={setShowEdit}
        title="Edit Workspace"
        description={editing ? `${computeTokens(editing)} tokens` : undefined}
        onSave={handleSaveEdit}
      >
        <div className="grid grid-cols-2 gap-x-4 gap-y-3">
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">Title</label>
            <Input
              value={E.title}
              onChange={(e) => setE({ ...E, title: e.target.value })}
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">Description</label>
            <Input
              value={E.description}
              onChange={(e) => setE({ ...E, description: e.target.value })}
            />
          </div>

          <div className="col-span-2 flex items-center gap-3 flex-wrap">
            <div className="flex items-center gap-1">
              <div
                className="w-7 h-7 rounded border cursor-pointer"
                style={{ backgroundColor: E.color }}
                onClick={() => document.getElementById('pick-color')?.click()}
              />
              <input
                id="pick-color"
                type="color"
                value={E.color}
                onChange={(e) => setE({ ...E, color: e.target.value })}
                className="w-0 h-0 opacity-0"
              />
            </div>
            <IconPicker
              value={E.icon}
              onChange={(v) => setE({ ...E, icon: v })}
            />
            <Separator orientation="vertical" className="h-6" />
            <div className="flex items-center gap-1.5">
              <label className="text-xs text-muted-foreground whitespace-nowrap">
                Max Prompt
              </label>
              <Input
                type="number"
                value={E.maxPromptSend}
                onChange={(e) =>
                  setE({ ...E, maxPromptSend: parseInt(e.target.value) || 0 })
                }
                className="h-7 w-16"
              />
            </div>
            <div className="flex items-center gap-1.5">
              <label className="text-xs text-muted-foreground whitespace-nowrap">
                Max Context
              </label>
              <Input
                type="number"
                value={E.maxContextLength}
                onChange={(e) =>
                  setE({
                    ...E,
                    maxContextLength: parseInt(e.target.value) || 0,
                  })
                }
                className="h-7 w-16"
              />
            </div>
            <div className="flex items-center gap-1.5">
              <Switch
                checked={E.commitChanges}
                onCheckedChange={(c) => setE({ ...E, commitChanges: c })}
              />
              <label className="text-xs text-muted-foreground cursor-pointer">
                Commit
              </label>
            </div>
          </div>

          <div className="col-span-2">
            <ExpandableEditor
              label="Personality"
              value={E.personality}
              onChange={(v) => setE({ ...E, personality: v })}
            />
          </div>

          <div className="col-span-2 flex gap-0 min-h-[180px]">
            <div className="w-28 shrink-0 flex flex-col gap-0 pt-0">
              {[
                'folders',
                'knowledge',
                'workspace_agents',
                'skills',
                'tools',
              ].map((key) => (
                <div
                  key={key}
                  className={`flex items-center gap-1 px-2 py-1 rounded text-xs transition-colors cursor-pointer ${
                    selectedField === key
                      ? 'bg-primary/10 text-primary font-medium'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                  }`}
                  onClick={() => setSelectedField(key)}
                >
                  <span className="grow text-left">
                    {
                      {
                        folders: 'Folders',
                        knowledge: 'Knowledge',
                        workspace_agents: 'Agents',
                        skills: 'Skills',
                        tools: 'Tools',
                      }[key]
                    }
                  </span>
                  {selectedField === key && (
                    <button
                      className="shrink-0 flex items-center justify-center w-4 h-4 rounded hover:bg-muted-foreground/20"
                      onClick={async (e) => {
                        e.stopPropagation();
                        console.log('Plus button clicked for:', key);
                        if (key === 'folders') {
                          const dir = await api.openDirectoryDialog();
                          if (dir) {
                            const f = fieldMap[key];
                            if (!f.state.includes(dir))
                              f.set([...f.state, dir]);
                          }
                        } else if (key === 'tools') {
                          setShowAddTool(true);
                        }
                      }}
                      title={
                        key === 'folders'
                          ? 'Add folder'
                          : key === 'knowledge'
                            ? 'Add file'
                            : 'Add'
                      }
                    >
                      <Icon name="Plus" className="w-3 h-3" />
                    </button>
                  )}
                </div>
              ))}
            </div>
            <div className="flex-1 min-h-[28px] max-h-[160px] overflow-y-auto p-1.5 rounded border border-border">
              <div className="flex flex-wrap gap-1">
                {fieldMap[selectedField]?.state.map((item: string) => (
                  <span
                    key={item}
                    className="flex items-center gap-1 px-2 py-0.5 text-xs rounded bg-muted text-muted-foreground whitespace-nowrap"
                  >
                    {item}
                    <button
                      onClick={() =>
                        fieldMap[selectedField].set(
                          fieldMap[selectedField].state.filter(
                            (x: string) => x !== item,
                          ),
                        )
                      }
                      className="text-destructive hover:text-destructive/80"
                    >
                      ✕
                    </button>
                  </span>
                ))}
              </div>
            </div>
          </div>
        </div>
      </EditDialog>

      <Dialog open={showAddTool} onOpenChange={(open) => {
        setShowAddTool(open);
        if (!open) setSelectedFilterProfileId(null);
      }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Tools</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div>
              <h4 className="text-sm font-medium mb-2">Profiles (Filter)</h4>
              <div className="grid grid-cols-2 gap-2">
                {availableProfiles.map((p) => (
                  <button
                    key={p.id}
                    className={`text-xs p-2 rounded border text-left ${
                      selectedFilterProfileId === p.id
                        ? 'bg-primary/10 border-primary'
                        : 'hover:bg-muted'
                    }`}
                    onClick={() => {
                      setSelectedFilterProfileId(
                        selectedFilterProfileId === p.id ? null : p.id
                      );
                    }}
                  >
                    {p.name} ({p.tools.length} tools)
                  </button>
                ))}
              </div>
            </div>
            {selectedFilterProfileId && (
              <Button
                size="sm"
                className="w-full"
                onClick={() => {
                  const profile = availableProfiles.find(
                    (p) => p.id === selectedFilterProfileId
                  );
                  if (profile) {
                    const newTools = [...new Set([...E.tools, ...profile.tools])];
                    setE({ ...E, tools: newTools });
                  }
                }}
              >
                Add All Tools from Profile
              </Button>
            )}
            <div>
              <h4 className="text-sm font-medium mb-2">Tools</h4>
              <div className="grid grid-cols-2 gap-2 max-h-60 overflow-y-auto">
                {(() => {
                  const selectedProfile = availableProfiles.find(
                    (p) => p.id === selectedFilterProfileId
                  );
                  const filteredTools = selectedProfile
                    ? availableTools.filter((t) =>
                      selectedProfile.tools.includes(t.name)
                    )
                    : availableTools;
                  return filteredTools.map((t) => (
                    <button
                      key={t.name}
                      className={`text-xs p-2 rounded border text-left ${
                        E.tools.includes(t.name)
                          ? 'bg-primary/10 border-primary'
                          : 'hover:bg-muted'
                      }`}
                      onClick={() => {
                        if (!E.tools.includes(t.name)) {
                          setE({ ...E, tools: [...E.tools, t.name] });
                        }
                      }}
                    >
                      {t.name}
                    </button>
                  ));
                })()}
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function AgentsSection() {
  const [agents, setAgents] = useState<api.backend.AgentConfig[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [availableProviders, setAvailableProviders] = useState<string[]>([]);
  const [showEdit, setShowEdit] = useState(false);
  const [editing, setEditing] = useState<api.backend.AgentConfig | null>(null);
  const [A, setA] = useState<api.backend.AgentConfig>(
    new api.backend.AgentConfig()
  );

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
        a.name === editing.name ? A : a
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
      setA(new api.backend.AgentConfig({
        icon: '🤖',
        color: '#6366f1',
      }));
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
    {} as Record<string, api.backend.AgentConfig[]>
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
                <BaseCard
                  key={agent.name}
                  color={agent.color || '#6366f1'}
                  headerLeft={null}
                  headerRight={
                    <div className="flex gap-1">
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
                    </div>
                  }
                  icon={agent.icon || '🤖'}
                  title={agent.name}
                  small
                >
                  <div className="base-card-desc">
                    {agent.provider} / {agent.model}
                  </div>
                </BaseCard>
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
          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Category</label>
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
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Icon</label>
              <IconPicker
                value={A.icon || '🤖'}
                onChange={(v) => setA({ ...A, icon: v })}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Color</label>
              <div className="flex items-center gap-2">
                <div
                  className="w-8 h-8 rounded border border-border cursor-pointer"
                  style={{ backgroundColor: A.color || '#6366f1' }}
                />
                <Input
                  type="color"
                  value={A.color || '#6366f1'}
                  onChange={(e) => setA({ ...A, color: e.target.value })}
                  className="w-full h-8"
                />
              </div>
            </div>
          </div>
        </div>
      </EditDialog>
    </div>
  );
}

function SkillsSection() {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-foreground">Skills</h3>
      <p className="text-sm text-muted-foreground">
        Manage skills and capabilities.
      </p>
    </div>
  );
}

function ToolsSection() {
  const [tools, setTools] = useState<api.backend.ToolUIInfo[]>([]);
  const [profiles, setProfiles] = useState<api.backend.ToolProfile[]>([]);
  const [selectedProfileID, setSelectedProfileID] = useState<number | null>(
    null,
  );
  const [showNewProfileDialog, setShowNewProfileDialog] = useState(false);
  const [newProfileName, setNewProfileName] = useState('');
  const [newProfileColor, setNewProfileColor] = useState('#6b7280');
  const [newProfileIcon, setNewProfileIcon] = useState('🔧');
  const [filterType, setFilterType] = useState<'all' | 'active' | 'inactive'>('all');
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    api.getAvailableTools().then(setTools);
    api.getToolProfiles().then((profiles) => {
      setProfiles(profiles);
      const defaultProfile = profiles.find((p) => p.name === 'Default');
      if (defaultProfile) setSelectedProfileID(defaultProfile.id);
    });
  }, []);

  const handleToggle = async (toolName: string, enabled: boolean) => {
    const profile = profiles.find((p) => p.id === selectedProfileID);
    if (profile && profile.name !== 'Default') {
      await api.toggleProfileTool(profile.id, toolName, enabled);
      setProfiles((prev) =>
        prev.map((p) => {
          if (p.id === profile.id) {
            return {
              ...p,
              tools: enabled
                ? [...p.tools, toolName]
                : p.tools.filter((t) => t !== toolName),
            };
          }
          return p;
        }),
      );
    } else {
      await api.toggleTool(toolName, enabled);
      setTools((prev) =>
        prev.map((t) => (t.name === toolName ? { ...t, enabled } : t)),
      );
    }
  };

  const handleCreateProfile = async () => {
    const profile = await api.createToolProfile(
      newProfileName,
      newProfileColor,
      newProfileIcon,
    );
    if (profile) {
      setProfiles((prev) => [...prev, profile]);
      setSelectedProfileID(profile.id);
      setShowNewProfileDialog(false);
      setNewProfileName('');
      setNewProfileColor('#6b7280');
      setNewProfileIcon('🔧');
    }
  };

  const handleDeleteProfile = async (id: number) => {
    const success = await api.deleteToolProfile(id);
    if (success) {
      setProfiles((prev) => prev.filter((p) => p.id !== id));
      if (selectedProfileID === id) {
        const defaultProfile = profiles.find((p) => p.name === 'Default');
        setSelectedProfileID(defaultProfile?.id || null);
      }
    }
  };

  const selectedProfile = profiles.find((p) => p.id === selectedProfileID);
  const isProfileActive = selectedProfile && selectedProfile.name !== 'Default';

  // Apply all filters
  const filteredTools = tools.filter((tool) => {
    // Search filter
    if (searchQuery.trim() !== '') {
      const query = searchQuery.toLowerCase();
      if (!tool.name.toLowerCase().includes(query)) {
        return false;
      }
    }

    // Status filter
    const isEnabled = isProfileActive
      ? (selectedProfile?.tools.includes(tool.name) ?? false)
      : tool.enabled;

    if (filterType === 'active' && !isEnabled) return false;
    if (filterType === 'inactive' && isEnabled) return false;

    return true;
  });

  const groupedTools = filteredTools.reduce(
    (acc, tool) => {
      const cat = tool.category || 'Other';
      if (!acc[cat]) acc[cat] = [];
      acc[cat].push(tool);
      return acc;
    },
    {} as Record<string, api.backend.ToolUIInfo[]>,
  );

  return (
    <div className="space-y-6">
      <div>
        <div className="flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-foreground">Tools</h3>
          </div>
          <div className="flex items-center gap-3">
            <Input
              placeholder="Search tools by name..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="max-w-sm"
            />
            <Select
              value={filterType}
              onValueChange={(value: 'all' | 'active' | 'inactive') =>
                setFilterType(value)
              }
            >
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="Filter" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="inactive">Inactive</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <p className="text-sm text-muted-foreground">
            Configure agent tools for the active workspace.
          </p>
        </div>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Profiles</h4>
        <div className="profiles-grid">
          {profiles.map((profile) => (
            <BaseCard
              key={profile.id}
              color={
                selectedProfileID === profile.id ? profile.color : '#4b5563'
              }
              headerLeft={
                <span className="text-xs text-white opacity-90">Profile</span>
              }
              headerRight={
                profile.name !== 'Default' ? (
                  <button
                    className="base-card-btn"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDeleteProfile(profile.id);
                    }}
                  >
                    <Icon name="Trash2" className="w-3 h-3" />
                  </button>
                ) : null
              }
              icon={profile.icon}
              title={profile.name}
              selected={selectedProfileID === profile.id}
              onClick={() => setSelectedProfileID(profile.id)}
              className="cursor-pointer"
              small
            >
              <div className="base-card-desc">{profile.tools.length} tools</div>
            </BaseCard>
          ))}
          <BaseCard
            color="#374151"
            headerLeft={
              <span className="text-xs text-white opacity-90">Action</span>
            }
            icon="+"
            title="Add Profile"
            onClick={() => setShowNewProfileDialog(true)}
            className="cursor-pointer"
            small
          >
            <div className="base-card-desc">Create new profile</div>
          </BaseCard>
        </div>
      </div>

      {Object.entries(groupedTools).map(([category, categoryTools]) => {
        return (
          <div key={category}>
            <h4 className="text-sm font-medium text-foreground mb-2">
              {category}
            </h4>
            <div className="tools-grid">
              {categoryTools.map((tool) => {
                const isEnabled = isProfileActive
                  ? (selectedProfile?.tools.includes(tool.name) ?? false)
                  : tool.enabled;
                return (
                  <BaseCard
                    key={tool.name}
                    color={isEnabled ? categoryColors[category] : '#6b7280'}
                    headerLeft={
                      <span className="text-xs text-white opacity-90">
                        {category}
                      </span>
                    }
                    headerRight={
                      <Switch
                        checked={isEnabled}
                        onCheckedChange={(checked) =>
                          handleToggle(tool.name, checked)
                        }
                        className="data-[state=checked]:bg-white/30"
                      />
                    }
                    icon="🔧"
                    title={tool.name}
                  >
                    <div className="base-card-desc">{tool.description}</div>
                  </BaseCard>
                );
              })}
            </div>
          </div>
        );
      })}

      {tools.length === 0 && (
        <Card>
          <CardContent className="py-8">
            <p className="text-sm text-muted-foreground text-center">
              No tools available
            </p>
          </CardContent>
        </Card>
      )}

      <Dialog
        open={showNewProfileDialog}
        onOpenChange={setShowNewProfileDialog}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Profile</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="grid grid-cols-4 items-center gap-4">
              <span className="text-right text-sm">Name</span>
              <Input
                value={newProfileName}
                onChange={(e) => setNewProfileName(e.target.value)}
                className="col-span-3"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <span className="text-right text-sm">Color</span>
              <div className="col-span-3 flex items-center gap-2">
                <div
                  className="w-8 h-8 rounded border border-border cursor-pointer"
                  style={{ backgroundColor: newProfileColor }}
                />
                <Input
                  type="color"
                  value={newProfileColor}
                  onChange={(e) => setNewProfileColor(e.target.value)}
                  className="w-16 h-8"
                />
              </div>
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <span className="text-right text-sm">Icon</span>
              <IconPicker value={newProfileIcon} onChange={setNewProfileIcon} />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowNewProfileDialog(false)}
            >
              Cancel
            </Button>
            <Button onClick={handleCreateProfile}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

const knownApiUrls = [
  { name: 'OpenAI', url: 'https://api.openai.com/v1' },
  { name: 'Cloudflare', url: 'https://api.cloudflare.com/client/v4/accounts/{account_id}/ai/v1' },
  { name: 'OpenRouter', url: 'https://openrouter.ai/api/v1' },
  { name: 'Anthropic', url: 'https://api.anthropic.com/v1' },
  { name: 'Google Gemini', url: 'https://generativelanguage.googleapis.com/v1beta' },
  { name: 'Ollama', url: 'http://localhost:11434/v1' },
];

function ModelsSection() {
  const [adaConfig, setAdaConfig] = useState<api.backend.AdaConfig | null>(null);
  const [showProviderDialog, setShowProviderDialog] = useState(false);
  const [editingProvider, setEditingProvider] = useState<{ name: string; config: api.backend.ProviderConfig } | null>(null);
  const [collapsed, setCollapsed] = useState<{ apiUrl: boolean; apiKeys: boolean; models: boolean }>({
    apiUrl: false,
    apiKeys: false,
    models: true,
  });
  const [newModelForm, setNewModelForm] = useState({ name: '', type: '' as api.backend.ModelType | '' });
  const [providerForm, setProviderForm] = useState({
    name: '',
    api_url: '',
    api_keys: [] as api.backend.ProviderApiKey[],
    type_connection: '',
    models: {} as Record<string, api.backend.ModelSettings>,
  });

  // Fetch models dialog state
  const [showFetchModelsDialog, setShowFetchModelsDialog] = useState(false);
  const [fetchedModels, setFetchedModels] = useState<api.backend.ProviderModel[]>([]);
  const [selectedModelsToAdd, setSelectedModelsToAdd] = useState<string[]>([]);
  // Multi-select filter: set of active type/capability filters. Model is shown if
  // it matches ALL active filters (pure AND). Each filter restricts the list;
  // toggling a filter off relaxes that restriction. Empty set = show all.
  const [activeFilters, setActiveFilters] = useState<Set<string>>(new Set());
  const [fetchingModels, setFetchingModels] = useState(false);

  const toggleFilter = (filter: string) => {
    setActiveFilters(prev => {
      const next = new Set(prev);
      if (next.has(filter)) {
        next.delete(filter);
      } else {
        next.add(filter);
      }
      return next;
    });
  };

  // Returns true if a fetched model passes the current active filter set.
  // Pure AND: every active filter must match. Empty set = show all.
  // Each of free/thinking/tools/embedding/vision is an independent capability:
  // clicking a filter adds a restriction, un-clicking removes it.
  const matchesFilter = (model: api.backend.ProviderModel) => {
    if (activeFilters.size === 0) return true;

    const isFree = !!model.free;
    const isThinking = !!model.thinking;
    const isTools = model.tools !== false;
    const isEmbedding = !!model.embedding;
    const isVision = !!model.vision;

    if (activeFilters.has('free') && !isFree) return false;
    if (activeFilters.has('thinking') && !isThinking) return false;
    if (activeFilters.has('tools') && !isTools) return false;
    if (activeFilters.has('embedding') && !isEmbedding) return false;
    if (activeFilters.has('vision') && !isVision) return false;

    return true;
  };

  // API key validation
  const [validatingApiKey, setValidatingApiKey] = useState<number | null>(null);
  const [apiKeyValidationResults, setApiKeyValidationResults] = useState<Record<number, api.backend.ProviderTestResult>>({});

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    const cfg = await api.getAdaConfig();
    setAdaConfig(cfg || new api.backend.AdaConfig());
  };

  const saveConfig = async () => {
    if (adaConfig) {
      await api.setAdaConfig(adaConfig);
    }
  };

  const handleOpenProvider = (name: string, config: api.backend.ProviderConfig) => {
    setEditingProvider({ name, config });
    setProviderForm({
      name,
      api_url: config.api_url,
      api_keys: config.api_keys || [],
      type_connection: config.type_connection,
      models: config.models || {},
    });
    setShowProviderDialog(true);
  };

  const handleAddModel = () => {
    if (!newModelForm.name || !newModelForm.type) return;
    setProviderForm({
      ...providerForm,
      models: {
        ...providerForm.models,
        [newModelForm.name]: new api.backend.ModelSettings({ type: newModelForm.type }),
      },
    });
    setNewModelForm({ name: '', type: '' });
  };

  const handleRemoveModel = (modelName: string) => {
    const newModels = { ...providerForm.models };
    delete newModels[modelName];
    setProviderForm({ ...providerForm, models: newModels });
  };

  const handleAddApiKey = () => {
    setProviderForm({
      ...providerForm,
      api_keys: [...providerForm.api_keys, { key: '', user_key: '' }],
    });
  };

  const handleUpdateApiKey = (index: number, field: 'key' | 'user_key', value: string) => {
    const newKeys = [...providerForm.api_keys];
    newKeys[index] = { ...newKeys[index], [field]: value };
    setProviderForm({ ...providerForm, api_keys: newKeys });
  };

  const handleRemoveApiKey = (index: number) => {
    const newKeys = providerForm.api_keys.filter((_, i) => i !== index);
    setProviderForm({ ...providerForm, api_keys: newKeys });
    // Clear validation results for this index
    setApiKeyValidationResults(prev => {
      const newResults = { ...prev };
      delete newResults[index];
      return newResults;
    });
  };

  const handleValidateApiKey = async (index: number) => {
    const apiKeyEntry = providerForm.api_keys[index];
    if (!apiKeyEntry || !apiKeyEntry.key || !providerForm.api_url || !providerForm.type_connection) return;

    setValidatingApiKey(index);
    const result = await api.testProviderConnection(
      providerForm.name,
      apiKeyEntry.key,
      providerForm.api_url,
      providerForm.type_connection,
    );
    setApiKeyValidationResults(prev => ({ ...prev, [index]: result }));
    setValidatingApiKey(null);
  };

  const handleFetchModels = async () => {
    // Use first API key if available
    const firstApiKey = providerForm.api_keys.find(k => k.key.trim() !== '');
    if (!firstApiKey || !providerForm.api_url || !providerForm.type_connection || !providerForm.name) return;

    setFetchingModels(true);
    const models = await api.fetchProviderModels(
      providerForm.name,
      firstApiKey.key,
      providerForm.api_url,
      providerForm.type_connection,
    );
    setFetchedModels(models);
    setSelectedModelsToAdd([]);
    setShowFetchModelsDialog(true);
    setFetchingModels(false);
  };

  const handleAddSelectedModels = () => {
    const newModels = { ...providerForm.models };
    selectedModelsToAdd.forEach(modelName => {
      const model = fetchedModels.find(m => m.name === modelName);
      if (model) {
        newModels[modelName] = new api.backend.ModelSettings({
          vision: model.vision,
          embedding: model.embedding,
          tools: model.tools,
          free: model.free,
          thinking: model.thinking,
        });
      }
    });
    setProviderForm({ ...providerForm, models: newModels });
    setShowFetchModelsDialog(false);
    setSelectedModelsToAdd([]);
    setFetchedModels([]);
  };

  const handleSaveProvider = async () => {
    if (!adaConfig || !providerForm.name) return;

    // Create a proper AdaConfig instance
    const newConfig = new api.backend.AdaConfig({
      ...adaConfig,
      providers: {
        ...(adaConfig.providers || {}),
        [providerForm.name]: new api.backend.ProviderConfig({
          api_url: providerForm.api_url,
          api_keys: providerForm.api_keys.filter((k) => k.key.trim() !== ''),
          type_connection: providerForm.type_connection,
          models: providerForm.models,
        }),
      },
    });

    setAdaConfig(newConfig);
    await api.setAdaConfig(newConfig);
    setShowProviderDialog(false);
    setEditingProvider(null);
  };

  const handleDeleteProvider = async (name: string) => {
    if (!adaConfig) return;

    // Create proper AdaConfig instance
    const newProviders = { ...(adaConfig.providers || {}) };
    delete newProviders[name];

    const newConfig = new api.backend.AdaConfig({
      ...adaConfig,
      providers: newProviders,
    });

    setAdaConfig(newConfig);
    await api.setAdaConfig(newConfig);
  };

  // Modelos disponíveis para os selects de Embedding/Image, derivados dos
  // modelos cadastrados no provider escolhido e filtrados por capability.
  const embeddingModels = (() => {
    const provider = adaConfig?.embedding_provider;
    const models = provider ? adaConfig?.providers?.[provider]?.models : undefined;
    if (!models) return [];
    return Object.entries(models)
      .filter(([, s]) => s?.embedding)
      .map(([name]) => ({ value: name, label: name }));
  })();

  const imageModels = (() => {
    const provider = adaConfig?.image_provider;
    const models = provider ? adaConfig?.providers?.[provider]?.models : undefined;
    if (!models) return [];
    return Object.entries(models)
      .filter(([, s]) => s?.vision)
      .map(([name]) => ({ value: name, label: name }));
  })();

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-foreground">Models</h3>
        <p className="text-sm text-muted-foreground">
          Configure providers, API keys, and models.
        </p>
      </div>

      <div className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Embedding Provider</label>
            <Select
              value={adaConfig?.embedding_provider || ''}
              onValueChange={(v) => {
                if (adaConfig) {
                  const newCfg = new api.backend.AdaConfig({
                    ...adaConfig,
                    embedding_provider: v,
                  });
                  setAdaConfig(newCfg);
                }
              }}
              onCloseAutoFocus={() => saveConfig()}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select provider" />
              </SelectTrigger>
              <SelectContent>
                {adaConfig &&
                  Object.keys(adaConfig.providers || {}).map((name) => (
                    <SelectItem key={name} value={name}>
                      {name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Embedding Model</label>
            <Combobox
              value={adaConfig?.embedding_model || ''}
              onValueChange={(v) => {
                if (adaConfig) {
                  const newCfg = new api.backend.AdaConfig({
                    ...adaConfig,
                    embedding_model: v,
                  });
                  setAdaConfig(newCfg);
                  saveConfig();
                }
              }}
              options={embeddingModels}
              placeholder="e.g., text-embedding-3-small"
              emptyText="Nenhum modelo de embedding neste provider — digite um valor custom"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Image Provider</label>
            <Select
              value={adaConfig?.image_provider || ''}
              onValueChange={(v) => {
                if (adaConfig) {
                  const newCfg = new api.backend.AdaConfig({
                    ...adaConfig,
                    image_provider: v,
                  });
                  setAdaConfig(newCfg);
                }
              }}
              onCloseAutoFocus={() => saveConfig()}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select provider" />
              </SelectTrigger>
              <SelectContent>
                {adaConfig &&
                  Object.keys(adaConfig.providers || {}).map((name) => (
                    <SelectItem key={name} value={name}>
                      {name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Image Model</label>
            <Combobox
              value={adaConfig?.image_model || ''}
              onValueChange={(v) => {
                if (adaConfig) {
                  const newCfg = new api.backend.AdaConfig({
                    ...adaConfig,
                    image_model: v,
                  });
                  setAdaConfig(newCfg);
                  saveConfig();
                }
              }}
              options={imageModels}
              placeholder="e.g., dall-e-3"
              emptyText="Nenhum modelo com visão neste provider — digite um valor custom"
            />
          </div>
        </div>
      </div>

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h4 className="text-sm font-medium text-foreground">Providers</h4>
          <Button
            size="sm"
            onClick={() => {
              setEditingProvider(null);
              setProviderForm({ name: '', api_url: '', api_keys: [], type_connection: '', models: {} });
              setShowProviderDialog(true);
            }}
          >
            <Icon name="Plus" size={16} />
            Add Provider
          </Button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {adaConfig &&
            Object.entries(adaConfig.providers || {}).map(([name, config]) => (
              <BaseCard
                key={name}
                color="#6b7280"
                headerLeft={
                  <span className="text-xs text-white opacity-90">Provider</span>
                }
                headerRight={
                  <div className="flex gap-1">
                    <button
                      className="base-card-btn"
                      onClick={() => handleOpenProvider(name, config)}
                    >
                      <Icon name="Edit" className="w-3 h-3" />
                    </button>
                    <button
                      className="base-card-btn"
                      onClick={() => handleDeleteProvider(name)}
                    >
                      <Icon name="Trash2" className="w-3 h-3" />
                    </button>
                  </div>
                }
                icon="🔌"
                title={name}
                small
              >
                <div className="text-xs text-muted-foreground truncate">
                  {config.api_url || 'No API URL'}
                </div>
              </BaseCard>
            ))}
        </div>
      </div>

      <EditDialog
        open={showProviderDialog}
        onOpenChange={setShowProviderDialog}
        title={editingProvider ? 'Edit Provider' : 'New Provider'}
        onSave={handleSaveProvider}
      >
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Name</label>
            <Input
              value={providerForm.name}
              onChange={(e) => setProviderForm({ ...providerForm, name: e.target.value })}
              placeholder="e.g., OpenAI"
              disabled={!!editingProvider}
            />
          </div>
          <div className="space-y-2">
            <div
              className="flex items-center justify-between cursor-pointer"
              onClick={() => setCollapsed({ ...collapsed, apiUrl: !collapsed.apiUrl })}
            >
              <label className="text-sm font-medium">API URL</label>
              <Icon
                name={collapsed.apiUrl ? "ChevronDown" : "ChevronRight"}
                size={16}
                className="text-muted-foreground"
              />
            </div>
            {!collapsed.apiUrl && (
              <div className="space-y-1">
                <Select
                  value={knownApiUrls.find((u) => u.url === providerForm.api_url) ? providerForm.api_url : 'custom'}
                  onValueChange={(v) => {
                    if (v !== 'custom') {
                      setProviderForm({ ...providerForm, api_url: v });
                    }
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select known API" />
                  </SelectTrigger>
                  <SelectContent>
                    {knownApiUrls.map((u) => (
                      <SelectItem key={u.url} value={u.url}>
                        {u.name}
                      </SelectItem>
                    ))}
                    <SelectItem value="custom">Custom URL</SelectItem>
                  </SelectContent>
                </Select>
                <Input
                  value={providerForm.api_url}
                  onChange={(e) => setProviderForm({ ...providerForm, api_url: e.target.value })}
                  placeholder="Enter custom API URL"
                />
              </div>
            )}
          </div>
          <div className="space-y-2">
            <div
              className="flex items-center justify-between cursor-pointer"
              onClick={() => setCollapsed({ ...collapsed, apiKeys: !collapsed.apiKeys })}
            >
              <div className="flex items-center gap-2">
                <label className="text-sm font-medium">API Keys</label>
                {providerForm.api_keys.length > 0 && (
                  <span className="text-xs text-muted-foreground">({providerForm.api_keys.length})</span>
                )}
              </div>
              <Icon
                name={collapsed.apiKeys ? "ChevronDown" : "ChevronRight"}
                size={16}
                className="text-muted-foreground"
              />
            </div>
            {!collapsed.apiKeys && (
              <div className="space-y-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleAddApiKey}
                  className="w-full"
                >
                  <Icon name="Plus" size={14} className="mr-1" /> Add API Key
                </Button>
                {providerForm.api_keys.map((apiKey, index) => (
                  <div key={index} className="space-y-1 p-2 border border-input rounded-md">
                    <div className="flex gap-2">
                      <Input
                        type="password"
                        value={apiKey.key}
                        onChange={(e) => handleUpdateApiKey(index, 'key', e.target.value)}
                        placeholder="Valor da chave ou nome de variável (ex: OPENROUTER_API_KEY)"
                        className="flex-1"
                      />
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleValidateApiKey(index)}
                        disabled={validatingApiKey === index}
                      >
                        {validatingApiKey === index ? (
                          <span className="flex items-center gap-1">
                            <span className="animate-spin h-4 w-4 border-2 border-current border-t-transparent rounded-full"></span>
                          </span>
                        ) : (
                          <span>Validate</span>
                        )}
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleRemoveApiKey(index)}
                      >
                        <Icon name="Trash2" size={14} />
                      </Button>
                    </div>
                    <Input
                      value={apiKey.user_key}
                      onChange={(e) => handleUpdateApiKey(index, 'user_key', e.target.value)}
                      placeholder="User Key (label or Cloudflare user ID)"
                    />
                    {apiKeyValidationResults[index] && (
                      <div
                        className={`text-xs p-1 rounded ${
                          apiKeyValidationResults[index].success ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                        }`}
                      >
                        {apiKeyValidationResults[index].success ? '✓ ' : '✗ '}
                        {apiKeyValidationResults[index].message}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
          <div className="space-y-2">
            <div
              className="flex items-center justify-between cursor-pointer"
              onClick={() => setCollapsed({ ...collapsed, models: !collapsed.models })}
            >
              <div className="flex items-center gap-2">
                <label className="text-sm font-medium">Models</label>
                {Object.keys(providerForm.models).length > 0 && (
                  <span className="text-xs text-muted-foreground">({Object.keys(providerForm.models).length})</span>
                )}
              </div>
              <Icon
                name={collapsed.models ? "ChevronDown" : "ChevronRight"}
                size={16}
                className="text-muted-foreground"
              />
            </div>
            {!collapsed.models && (
              <div className="space-y-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleFetchModels}
                  disabled={fetchingModels || providerForm.api_keys.filter(k => k.key.trim() !== '').length === 0}
                  className="w-full"
                >
                  {fetchingModels ? (
                    <span className="flex items-center gap-1">
                      <span className="animate-spin h-4 w-4 border-2 border-current border-t-transparent rounded-full"></span>
                      Fetching...
                    </span>
                  ) : (
                    <span>
                      <Icon name="Plus" size={14} className="mr-1" /> Fetch Models
                    </span>
                  )}
                </Button>
                <div className="flex gap-2">
                  <Input
                    value={newModelForm.name}
                    onChange={(e) => setNewModelForm({ ...newModelForm, name: e.target.value })}
                    placeholder="Model name"
                    className="flex-1"
                  />
                  <Select
                    value={newModelForm.type}
                    onValueChange={(v) => setNewModelForm({ ...newModelForm, type: v as api.backend.ModelType })}
                  >
                    <SelectTrigger className="w-32">
                      <SelectValue placeholder="Type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="free">Free</SelectItem>
                      <SelectItem value="thinking">Thinking</SelectItem>
                      <SelectItem value="tools">Tools</SelectItem>
                    </SelectContent>
                  </Select>
                  <Button
                    size="sm"
                    onClick={handleAddModel}
                    disabled={!newModelForm.name || !newModelForm.type}
                  >
                    <Icon name="Plus" size={14} />
                  </Button>
                </div>
                <div className="space-y-1">
                  {Object.entries(providerForm.models).map(([modelName, settings]) => (
                    <div
                      key={modelName}
                      className="flex items-center justify-between p-2 border border-input rounded-md"
                    >
                      <div className="flex items-center gap-2 min-w-0">
                        <span className="text-sm font-medium truncate">{modelName}</span>
                        {/* Clickable capability toggles */}
                        <div className="flex items-center gap-1 shrink-0">
                          <button
                            type="button"
                            className={`inline-flex items-center justify-center h-5 w-5 rounded transition-colors ${
                              settings?.free
                                ? 'bg-green-500/15 text-green-600 dark:text-green-400'
                                : 'text-muted-foreground/30 hover:text-muted-foreground/60'
                            }`}
                            title="Free"
                            onClick={(e) => {
                              e.stopPropagation();
                              const updated = new api.backend.ModelSettings({
                                ...settings,
                                free: !settings?.free,
                              });
                              setProviderForm({
                                ...providerForm,
                                models: { ...providerForm.models, [modelName]: updated },
                              });
                            }}
                          >
                            <Icon name="MessageSquare" size={12} />
                          </button>
                          <button
                            type="button"
                            className={`inline-flex items-center justify-center h-5 w-5 rounded transition-colors ${
                              settings?.thinking
                                ? 'bg-purple-500/15 text-purple-600 dark:text-purple-400'
                                : 'text-muted-foreground/30 hover:text-muted-foreground/60'
                            }`}
                            title="Thinking"
                            onClick={(e) => {
                              e.stopPropagation();
                              const updated = new api.backend.ModelSettings({
                                ...settings,
                                thinking: !settings?.thinking,
                              });
                              setProviderForm({
                                ...providerForm,
                                models: { ...providerForm.models, [modelName]: updated },
                              });
                            }}
                          >
                            <Icon name="Sparkles" size={12} />
                          </button>
                          <button
                            type="button"
                            className={`inline-flex items-center justify-center h-5 w-5 rounded transition-colors ${
                              settings?.tools !== false
                                ? 'bg-blue-500/15 text-blue-600 dark:text-blue-400'
                                : 'text-muted-foreground/30 hover:text-muted-foreground/60'
                            }`}
                            title="Tools"
                            onClick={(e) => {
                              e.stopPropagation();
                              const updated = new api.backend.ModelSettings({
                                ...settings,
                                tools: settings?.tools === false ? true : false,
                              });
                              setProviderForm({
                                ...providerForm,
                                models: { ...providerForm.models, [modelName]: updated },
                              });
                            }}
                          >
                            <Icon name="Wrench" size={12} />
                          </button>
                          <button
                            type="button"
                            className={`inline-flex items-center justify-center h-5 w-5 rounded transition-colors ${
                              settings?.embedding
                                ? 'bg-orange-500/15 text-orange-600 dark:text-orange-400'
                                : 'text-muted-foreground/30 hover:text-muted-foreground/60'
                            }`}
                            title="Embedding"
                            onClick={(e) => {
                              e.stopPropagation();
                              const updated = new api.backend.ModelSettings({
                                ...settings,
                                embedding: !settings?.embedding,
                              });
                              setProviderForm({
                                ...providerForm,
                                models: { ...providerForm.models, [modelName]: updated },
                              });
                            }}
                          >
                            <Icon name="Network" size={12} />
                          </button>
                          <button
                            type="button"
                            className={`inline-flex items-center justify-center h-5 w-5 rounded transition-colors ${
                              settings?.vision
                                ? 'bg-cyan-500/15 text-cyan-600 dark:text-cyan-400'
                                : 'text-muted-foreground/30 hover:text-muted-foreground/60'
                            }`}
                            title="Vision"
                            onClick={(e) => {
                              e.stopPropagation();
                              const updated = new api.backend.ModelSettings({
                                ...settings,
                                vision: !settings?.vision,
                              });
                              setProviderForm({
                                ...providerForm,
                                models: { ...providerForm.models, [modelName]: updated },
                              });
                            }}
                          >
                            <Icon name="Eye" size={12} />
                          </button>
                        </div>
                      </div>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleRemoveModel(modelName)}
                      >
                        <Icon name="Trash2" size={14} />
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Connection Type</label>
            <Select
              value={providerForm.type_connection}
              onValueChange={(v) => setProviderForm({ ...providerForm, type_connection: v })}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="openai">OpenAI Compatible</SelectItem>
                <SelectItem value="anthropic">Anthropic</SelectItem>
                <SelectItem value="gemini">Google Gemini</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </EditDialog>

      {/* Fetch Models Dialog */}
      <Dialog open={showFetchModelsDialog} onOpenChange={setShowFetchModelsDialog}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Select Models to Add</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            {/* Filter Checkboxes */}
            <div className="flex flex-wrap gap-2">
              {([
                { key: 'free', label: 'Free', icon: 'MessageSquare' },
                { key: 'thinking', label: 'Thinking', icon: 'Sparkles' },
                { key: 'tools', label: 'Tools', icon: 'Wrench' },
                { key: 'embedding', label: 'Embedding', icon: 'Network' },
                { key: 'vision', label: 'Vision', icon: 'Eye' },
              ] as const).map(({ key, label, icon }) => (
                <button
                  key={key}
                  type="button"
                  onClick={() => toggleFilter(key)}
                  className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md border text-xs transition-colors ${
                    activeFilters.has(key)
                      ? 'border-primary bg-primary/10 text-primary'
                      : 'border-border text-muted-foreground hover:border-primary/50 hover:text-foreground'
                  }`}
                >
                  <Icon name={icon} size={12} />
                  <span>{label}</span>
                </button>
              ))}
            </div>

            {/* Models List — fixed height */}
            <div className="h-80 overflow-y-auto space-y-1">
              {fetchedModels
                .filter(matchesFilter)
                .filter(model => !providerForm.models[model.name]) // Exclude already added models
                .map(model => {
                  const isSelected = selectedModelsToAdd.includes(model.name);
                  const isFree = !!model.free;
                  const isThinking = !!model.thinking;
                  const isTools = model.tools !== false;
                  const isEmbedding = !!model.embedding;
                  const isVision = !!model.vision;
                  return (
                    <div
                      key={model.name}
                      className={`flex items-center gap-2.5 p-2 border rounded-md cursor-pointer transition-colors ${
                        isSelected
                          ? 'border-primary bg-primary/5'
                          : 'border-input hover:bg-accent'
                      }`}
                      onClick={() => {
                        setSelectedModelsToAdd(prev => {
                          if (prev.includes(model.name)) {
                            return prev.filter(n => n !== model.name);
                          } else {
                            return [...prev, model.name];
                          }
                        });
                      }}
                    >
                      {/* Themed checkbox */}
                      <span
                        className={`inline-flex items-center justify-center h-4 w-4 rounded border transition-colors shrink-0 ${
                          isSelected
                            ? 'bg-primary border-primary text-primary-foreground'
                            : 'border-muted-foreground/40'
                        }`}
                      >
                        {isSelected && <Icon name="Check" size={10} />}
                      </span>
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium truncate">{model.name}</div>
                      </div>
                      {/* Capability icons — show all applicable */}
                      <div className="flex items-center gap-1 text-muted-foreground shrink-0">
                        {isFree && (
                          <span title="Free"><Icon name="MessageSquare" size={14} /></span>
                        )}
                        {isThinking && (
                          <span title="Thinking"><Icon name="Sparkles" size={14} /></span>
                        )}
                        {isTools && (
                          <span title="Tools"><Icon name="Wrench" size={14} /></span>
                        )}
                        {isEmbedding && (
                          <span title="Embedding"><Icon name="Network" size={14} /></span>
                        )}
                        {isVision && (
                          <span title="Vision"><Icon name="Eye" size={14} /></span>
                        )}
                      </div>
                    </div>
                  );
                })}
              {fetchedModels.filter(matchesFilter).filter(m => !providerForm.models[m.name]).length === 0 && (
                <div className="flex items-center justify-center h-full text-sm text-muted-foreground">
                  No models available for the selected filter.
                </div>
              )}
            </div>
          </div>
          <DialogFooter className="sm:justify-between">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                const visible = fetchedModels
                  .filter(matchesFilter)
                  .filter(m => !providerForm.models[m.name]);
                const allSelected = visible.length > 0 && visible.every(m => selectedModelsToAdd.includes(m.name));
                if (allSelected) {
                  setSelectedModelsToAdd([]);
                } else {
                  setSelectedModelsToAdd(visible.map(m => m.name));
                }
              }}
            >
              {(() => {
                const visible = fetchedModels.filter(matchesFilter).filter(m => !providerForm.models[m.name]);
                const allSelected = visible.length > 0 && visible.every(m => selectedModelsToAdd.includes(m.name));
                return allSelected ? 'Uncheck All' : 'Check All';
              })()}
            </Button>
            <div className="flex gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  setShowFetchModelsDialog(false);
                  setSelectedModelsToAdd([]);
                setFetchedModels([]);
              }}
            >
              Cancel
            </Button>
            <Button onClick={handleAddSelectedModels} disabled={selectedModelsToAdd.length === 0}>
              Add Selected ({selectedModelsToAdd.length})
            </Button>
            </div>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default function SettingsPage() {
  const [activeSection, setActiveSection] = useState('general');

  return (
    <div className="fixed inset-0 z-50 flex bg-background">
      <div className="settings-sidebar">
        <div className="settings-sidebar-header">
          <span className="text-sm font-semibold text-foreground">
            Settings
          </span>
        </div>
        <div className="settings-sidebar-nav">
          {settingsSections.map((section) => (
            <button
              key={section.id}
              className={`settings-nav-item ${activeSection === section.id ? 'active' : ''}`}
              onClick={() => setActiveSection(section.id)}
            >
              <span className="settings-nav-icon">
                <Icon name={section.icon} className="w-4 h-4" />
              </span>
              <span>{section.label}</span>
            </button>
          ))}
        </div>
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto">
        <div className="settings-content">
          {activeSection === 'general' && <GeneralSection />}
          {activeSection === 'workspace' && <WorkspacesSection />}
          {activeSection === 'agents' && <AgentsSection />}
          {activeSection === 'skills' && <SkillsSection />}
          {activeSection === 'tools' && <ToolsSection />}
          {activeSection === 'models' && <ModelsSection />}
        </div>
      </div>
    </div>
  );
}
