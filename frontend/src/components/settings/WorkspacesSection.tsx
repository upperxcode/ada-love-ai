import { useState, useEffect } from 'react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Switch } from '../ui/switch';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '../ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../ui/select';
import { BaseCard } from '../BaseCard';
import { EditDialog } from '../EditDialog';
import { ExpandableEditor } from '../ExpandableEditor';
import { Icon } from '../Icon';
import * as api from '../../api';

function WorkspacesSection() {
  const [workspaces, setWorkspaces] = useState<api.backend.WorkspaceConfig[]>(
    [],
  );
  const [showAdd, setShowAdd] = useState(false);
  const [showEdit, setShowEdit] = useState(false);
  const [editing, setEditing] = useState<api.backend.WorkspaceConfig | null>(
    null,
  );
  const [showAddTool, setShowAddTool] = useState(false);
  const [selectedFilterProfileId, setSelectedFilterProfileId] = useState<
    number | null
  >(null);
  const [availableTools, setAvailableTools] = useState<
    api.backend.ToolUIInfo[]
  >([]);
  const [availableProfiles, setAvailableProfiles] = useState<
    api.backend.ToolProfile[]
  >([]);
  const [specWizards, setSpecWizards] = useState<Array<{ id: string; name: string }>>([]);
  const [skillInput, setSkillInput] = useState('');

  useEffect(() => {
    api.getAvailableTools().then(setAvailableTools);
    api.getToolProfiles().then(setAvailableProfiles);
    const saved = localStorage.getItem('spec-wizards');
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        setSpecWizards(parsed.map((w: any) => ({ id: w.id, name: w.name })));
      } catch {}
    }
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
    workers: [] as string[],
    skills: [] as string[],
    tools: [] as string[],
    enabled: true,
    maxPromptSend: 0,
    commitChanges: true,
    maxContextLength: 0,
    specWizard: '',
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
    workers: {
      state: E.workers,
      set: (v) => setE((prev) => ({ ...prev, workers: v })),
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
    for (const a of ws.workers || []) t += 500;
    for (const s of ws.skills || []) t += 300;
    return t.toLocaleString();
  };

  const load = () => {
    api.getWorkspaces().then(setWorkspaces).catch(() => setWorkspaces([]));
    api.getAvailableTools().then(setKnownTools).catch(() => setKnownTools([]));
  };
  useEffect(() => {
    load();
    api.getWorkerCategories().then(setKnownSkills).catch(() => {});
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
      workers: (ws.workers || []).map((w) => typeof w === 'string' ? w : (w as any).name || ''),
      skills: ws.skills || [],
      tools: ws.tools || [],
      enabled: ws.enabled,
      maxPromptSend: ws.max_prompt_send || 0,
      commitChanges: ws.commit_changes !== false,
      maxContextLength: ws.max_context_length || 0,
      specWizard: ws.spec_wizard || '',
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
      workers: E.workers.map((name) => ({ name } as api.backend.WorkerConfig)),
      skills: E.skills,
      tools: E.tools,
      enabled: E.enabled,
      max_prompt_send: E.maxPromptSend,
      commit_changes: E.commitChanges,
      max_context_length: E.maxContextLength,
      spec_wizard: E.specWizard,
    });
    setShowEdit(false);
    setEditing(null);
    load();
  };

  const handleFieldAdd = async (key: string) => {
    if (key === 'folders') {
      const dir = await api.openDirectoryDialog();
      if (dir) {
        const f = fieldMap[key];
        if (!f.state.includes(dir)) f.set([...f.state, dir]);
      }
    } else if (key === 'knowledge') {
      const file = await api.openFileDialog();
      if (file) {
        const f = fieldMap[key];
        if (!f.state.includes(file)) f.set([...f.state, file]);
      }
    } else if (key === 'skills') {
      if (skillInput.trim()) {
        const f = fieldMap[key];
        if (!f.state.includes(skillInput.trim())) {
          f.set([...f.state, skillInput.trim()]);
        }
        setSkillInput('');
      }
    } else if (key === 'tools') {
      setShowAddTool(true);
    }
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
              {ws.workers?.length || 0} agents ·{' '}
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
        color={E.color}
        icon={E.icon}
        onColorChange={(color) => setE({ ...E, color })}
        onIconChange={(icon) => setE({ ...E, icon })}
      >
        <div className="space-y-4">
          {/* Row 1: Title + Description */}
          <div className="grid grid-cols-2 gap-4">
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
          </div>

          {/* Row 2: Max Prompt, Max Context, Commit */}
          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-1">
              <label className="text-xs text-muted-foreground">Max Prompt</label>
              <Input
                type="number"
                value={E.maxPromptSend}
                onChange={(e) =>
                  setE({ ...E, maxPromptSend: parseInt(e.target.value) || 0 })
                }
                className="h-8"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs text-muted-foreground">Max Context</label>
              <Input
                type="number"
                value={E.maxContextLength}
                onChange={(e) =>
                  setE({
                    ...E,
                    maxContextLength: parseInt(e.target.value) || 0,
                  })
                }
                className="h-8"
              />
            </div>
            <div className="space-y-1 flex items-end">
              <div className="flex items-center gap-1.5 h-8">
                <Switch
                  checked={E.commitChanges}
                  onCheckedChange={(c) => setE({ ...E, commitChanges: c })}
                />
                <label className="text-xs text-muted-foreground cursor-pointer">
                  Commit
                </label>
              </div>
            </div>
          </div>

          {/* Spec Wizard */}
          <div>
            <label className="text-xs font-medium mb-1 block">Spec Wizard</label>
            <div className="flex gap-2">
              <Select
                value={E.specWizard || ''}
                onValueChange={(v) => setE({ ...E, specWizard: v })}
              >
                <SelectTrigger className="h-8 text-xs flex-1">
                  <SelectValue placeholder="Link a Spec Wizard..." />
                </SelectTrigger>
                <SelectContent>
                  {specWizards.map((sw) => (
                    <SelectItem key={sw.id} value={sw.id}>
                      {sw.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {E.specWizard && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 px-2 text-xs"
                  onClick={() => setE({ ...E, specWizard: '' })}
                >
                  Clear
                </Button>
              )}
            </div>
          </div>

          {/* Personality */}
          <div>
            <ExpandableEditor
              label="Personality"
              value={E.personality}
              onChange={(v) => setE({ ...E, personality: v })}
            />
          </div>

          {/* Tabs: Folders, Knowledge, Agents, Skills, Tools */}
          <div className="flex gap-0 min-h-[180px]">
            <div className="w-28 shrink-0 flex flex-col gap-0 pt-0">
              {[
                { key: 'folders', label: 'Folders' },
                { key: 'knowledge', label: 'Knowledge' },
                { key: 'workers', label: 'Agents' },
                { key: 'skills', label: 'Skills' },
                { key: 'tools', label: 'Tools' },
              ].map(({ key, label }) => (
                <div
                  key={key}
                  className={`flex items-center gap-1 px-2 py-1 rounded text-xs transition-colors cursor-pointer ${
                    selectedField === key
                      ? 'bg-primary/10 text-primary font-medium'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                  }`}
                  onClick={() => setSelectedField(key)}
                >
                  <span className="grow text-left">{label}</span>
                  {selectedField === key && (
                    <button
                      className="shrink-0 flex items-center justify-center w-4 h-4 rounded hover:bg-muted-foreground/20"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleFieldAdd(key);
                      }}
                      title={
                        key === 'folders'
                          ? 'Add folder'
                          : key === 'knowledge'
                            ? 'Add file'
                            : key === 'skills'
                              ? 'Add skill'
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
              {/* Skills input */}
              {selectedField === 'skills' && (
                <div className="flex gap-1 mb-1.5">
                  <Input
                    value={skillInput}
                    onChange={(e) => setSkillInput(e.target.value)}
                    placeholder="Type skill name..."
                    className="h-6 text-xs flex-1"
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        e.preventDefault();
                        handleFieldAdd('skills');
                      }
                    }}
                  />
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 px-1.5"
                    onClick={() => handleFieldAdd('skills')}
                    disabled={!skillInput.trim()}
                  >
                    <Icon name="Plus" size={12} />
                  </Button>
                </div>
              )}
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

      <Dialog
        open={showAddTool}
        onOpenChange={(open) => {
          setShowAddTool(open);
          if (!open) setSelectedFilterProfileId(null);
        }}
      >
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
                        selectedFilterProfileId === p.id ? null : p.id,
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
                    (p) => p.id === selectedFilterProfileId,
                  );
                  if (profile) {
                    const newTools = [
                      ...new Set([...E.tools, ...profile.tools]),
                    ];
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
                    (p) => p.id === selectedFilterProfileId,
                  );
                  const filteredTools = selectedProfile
                    ? availableTools.filter((t) =>
                        selectedProfile.tools.includes(t.name),
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

export default WorkspacesSection;
