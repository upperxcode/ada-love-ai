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
import { Card, CardContent } from '../ui/card';
import { BaseCard } from '../BaseCard';
import { EditDialog } from '../EditDialog';
import { ExpandableEditor } from '../ExpandableEditor';
import { Icon } from '../Icon';
import * as api from '../../api';

function WorkersSection() {
  const [workers, setWorkers] = useState<api.backend.WorkerConfig[]>([]);
  const [connections, setConnections] = useState<api.backend.ConnectionDefinition[]>([]);
  const [showEdit, setShowEdit] = useState(false);
  const [editing, setEditing] = useState<api.backend.WorkerConfig | null>(null);
  const [W, setW] = useState<api.backend.WorkerConfig>(
    new api.backend.WorkerConfig(),
  );
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<api.backend.ConnectionTestResult | null>(null);
  const [showConnectionsList, setShowConnectionsList] = useState(false);

  const load = async () => {
    try {
      setWorkers(await api.getWorkers() || []);
      setConnections(await api.getPredefinedConnections() || []);
    } catch (e) {
      console.error('[WorkersSection] load error:', e);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const handleSave = async () => {
    let updatedWorkers = [...workers];
    if (editing) {
      updatedWorkers = updatedWorkers.map((w) =>
        w.name === editing.name ? W : w,
      );
    } else {
      updatedWorkers = [...updatedWorkers, W];
    }
    setWorkers(updatedWorkers);
    await api.setWorkers(updatedWorkers);
    setShowEdit(false);
    setEditing(null);
  };

  const handleDelete = async (name: string) => {
    const updatedWorkers = workers.filter((w) => w.name !== name);
    setWorkers(updatedWorkers);
    await api.setWorkers(updatedWorkers);
  };

  const openEdit = (worker: api.backend.WorkerConfig | null = null) => {
    if (worker) {
      setEditing(worker);
      setW(worker);
    } else {
      setEditing(null);
      setW(
        new api.backend.WorkerConfig({
          icon: '🤖',
          color: '#3b82f6',
          connection_type: 'ada',
          connection_name: 'Ada',
        }),
      );
    }
    setShowEdit(false);
    setTimeout(() => setShowEdit(true), 0);
    setTestResult(null);
  };

  const handleTest = async () => {
    setTesting(true);
    setTestResult(null);
    const result = await api.testConnection(
      W.connection_type,
      W.connection_name,
      W.connection_config,
    );
    setTestResult(result);
    setTesting(false);
  };

  const selectedConnection = connections.find(
    (c) => c.name === W.connection_name && c.type === W.connection_type,
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-foreground">Workers</h3>
          <p className="text-sm text-muted-foreground">
            Chat workers within each workspace. Each worker connects via a channel (Ada, CLI, REST, MCP).
          </p>
        </div>
        <Button size="sm" onClick={() => openEdit()} className="gap-1">
          <Icon name="Plus" size={16} /> New Worker
        </Button>
      </div>

      <div className="agents-grid">
        {workers.map((worker) => {
          const color = worker.color || '#3b82f6';
          const icon = worker.icon || (worker.connection_name === 'Ada' ? '🤖' : worker.connection_name === 'Crush' ? '💎' : worker.connection_name === 'OpenCode' ? '🔓' : worker.connection_name === 'Aider' ? '🤝' : '⚡');
          return (
            <BaseCard
              key={worker.name}
              color={color}
              headerLeft={
                <span className="text-xs text-white opacity-90">
                  {worker.connection_name || worker.connection_type}
                </span>
              }
              headerRight={
                <div className="flex gap-1">
                  <button
                    className="base-card-btn"
                    onClick={() => openEdit(worker)}
                    title="Edit"
                  >
                    <Icon name="Edit" className="w-3 h-3" />
                  </button>
                  <button
                    className="base-card-btn"
                    onClick={() => handleDelete(worker.name)}
                    title="Delete"
                  >
                    <Icon name="Trash2" className="w-3 h-3" />
                  </button>
                </div>
              }
              icon={icon}
              title={worker.name}
            >
              <div className="flex flex-wrap gap-1 justify-center">
                {worker.language && (
                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-primary/10 text-primary">
                    {worker.language}
                  </span>
                )}
              </div>
              {worker.persona && (
                <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                  {worker.persona}
                </p>
              )}
              <div className="flex flex-wrap gap-1 mt-2">
                {(['inherit_folders', 'inherit_knowledge', 'inherit_skills', 'inherit_tools', 'inherit_persona'] as const).map((key) => {
                  const label = key.replace('inherit_', '');
                  return (
                    <span
                      key={key}
                      className={`text-[10px] px-1.5 py-0.5 rounded ${
                        worker[key]
                          ? 'bg-primary/10 text-primary'
                          : 'bg-muted text-muted-foreground'
                      }`}
                    >
                      {worker[key] ? '↔' : '✕'} {label.charAt(0).toUpperCase() + label.slice(1)}
                    </span>
                  );
                })}
              </div>
            </BaseCard>
          );
        })}
      </div>

      {workers.length === 0 && (
        <Card>
          <CardContent className="py-8">
            <p className="text-sm text-muted-foreground text-center">
              No workers yet. Create your first worker (e.g., Ada) to get started.
            </p>
          </CardContent>
        </Card>
      )}

      <EditDialog
        open={showEdit}
        onOpenChange={setShowEdit}
        title={editing ? 'Edit Worker' : 'New Worker'}
        onSave={handleSave}
        color={W.color || '#3b82f6'}
        icon={W.icon || '🤖'}
        onColorChange={(color) => setW({ ...W, color })}
        onIconChange={(icon) => setW({ ...W, icon })}
      >
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Name</label>
            <Input
              value={W.name}
              onChange={(e) => setW({ ...W, name: e.target.value })}
              placeholder="e.g., Ada, Crush, OpenCode"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Persona</label>
            <ExpandableEditor
              label="Persona"
              value={W.persona}
              onChange={(v) => setW({ ...W, persona: v })}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Response Language</label>
            <p className="text-xs text-muted-foreground">
              The AI will prefer English internally, but respond to you in this language.
            </p>
            <Select
              value={W.language}
              onValueChange={(v) => setW({ ...W, language: v })}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select language" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="pt-BR">Português (BR)</SelectItem>
                <SelectItem value="en">English</SelectItem>
                <SelectItem value="es">Español</SelectItem>
                <SelectItem value="fr">Français</SelectItem>
                <SelectItem value="de">Deutsch</SelectItem>
                <SelectItem value="it">Italiano</SelectItem>
                <SelectItem value="ja">日本語</SelectItem>
                <SelectItem value="zh">中文</SelectItem>
                <SelectItem value="ko">한국어</SelectItem>
                <SelectItem value="ru">Русский</SelectItem>
                <SelectItem value="ar">العربية</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Connection Settings */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Connection</label>
            <p className="text-xs text-muted-foreground">
              How this worker communicates. The connection provides the LLM backend.
            </p>
            <div className="flex items-center gap-2 p-3 rounded border border-border bg-muted/30">
              <span className="text-lg">{selectedConnection?.icon || '⚡'}</span>
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium">
                  {W.connection_name || 'Not configured'}
                </div>
                <div className="text-xs text-muted-foreground truncate">
                  {selectedConnection?.description || W.connection_type}
                </div>
              </div>
              <Button
                size="sm"
                variant="outline"
                onClick={() => setShowConnectionsList(true)}
              >
                Change
              </Button>
            </div>
          </div>

          {/* Connection Config (only when not Ada) */}
          {W.connection_type !== 'ada' && (
            <div className="space-y-2">
              <label className="text-sm font-medium">Connection Config</label>
              <Input
                value={W.connection_config}
                onChange={(e) => setW({ ...W, connection_config: e.target.value })}
                placeholder={W.connection_type === 'cli' ? 'e.g., /usr/bin/crush' : 'e.g., http://localhost:8080'}
              />
              <div className="flex items-center gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleTest}
                  disabled={testing}
                >
                  {testing ? (
                    <span className="flex items-center gap-1">
                      <span className="animate-spin h-3 w-3 border-2 border-current border-t-transparent rounded-full"></span>
                      Testing...
                    </span>
                  ) : (
                    'Test Connection'
                  )}
                </Button>
                {testResult && (
                  <span className={`text-xs ${testResult.success ? 'text-green-500' : 'text-red-500'}`}>
                    {testResult.success ? '✓' : '✗'} {testResult.message}
                    {testResult.latency_ms > 0 && ` (${testResult.latency_ms}ms)`}
                  </span>
                )}
              </div>
            </div>
          )}

          {/* Inheritance Section */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Workspace Inheritance</label>
            <p className="text-xs text-muted-foreground">
              Choose which workspace properties this worker inherits.
            </p>
            <div className="grid grid-cols-2 gap-2">
              {([
                { key: 'inherit_folders' as const, label: 'Folders', icon: 'Folder' },
                { key: 'inherit_knowledge' as const, label: 'Knowledge', icon: 'BookOpen' },
                { key: 'inherit_skills' as const, label: 'Skills', icon: 'Brain' },
                { key: 'inherit_tools' as const, label: 'Tools', icon: 'Wrench' },
                { key: 'inherit_persona' as const, label: 'Persona', icon: 'User' },
              ]).map(({ key, label, icon }) => (
                <div
                  key={key}
                  className="flex items-center justify-between p-2 rounded border border-border"
                >
                  <div className="flex items-center gap-2">
                    <Icon name={icon} size={14} className="text-muted-foreground" />
                    <span className="text-sm">{label}</span>
                  </div>
                  <Switch
                    checked={W[key]}
                    onCheckedChange={(checked) =>
                      setW({ ...W, [key]: checked })
                    }
                  />
                </div>
              ))}
            </div>
          </div>
        </div>
      </EditDialog>

      {/* Connections List Dialog */}
      <Dialog open={showConnectionsList} onOpenChange={setShowConnectionsList}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Choose Connection</DialogTitle>
          </DialogHeader>
          <div className="space-y-2 py-2">
            {connections.map((conn) => (
              <button
                key={`${conn.type}-${conn.name}`}
                type="button"
                className={`w-full flex items-center gap-3 p-3 rounded border text-left transition-colors ${
                  W.connection_name === conn.name && W.connection_type === conn.type
                    ? 'border-primary bg-primary/10'
                    : 'border-border hover:border-primary/50'
                }`}
                onClick={() => {
                  setW({
                    ...W,
                    connection_type: conn.type,
                    connection_name: conn.name,
                    connection_config: conn.command ? JSON.stringify({ command: conn.command }) : '',
                  });
                  setShowConnectionsList(false);
                }}
              >
                <span className="text-xl">{conn.icon}</span>
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium">{conn.name}</div>
                  <div className="text-xs text-muted-foreground">{conn.description}</div>
                </div>
                {W.connection_name === conn.name && W.connection_type === conn.type && (
                  <Icon name="Check" size={16} className="text-primary shrink-0" />
                )}
              </button>
            ))}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default WorkersSection;
