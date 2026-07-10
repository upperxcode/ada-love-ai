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
import { Card, CardContent } from '../ui/card';
import { BaseCard } from '../BaseCard';
import { EditDialog } from '../EditDialog';
import { Icon } from '../Icon';
import * as api from '../../api';

interface MCPServerEntry extends api.backend.MCPServerUI {
  name: string;
}

function MCPSection() {
  const [servers, setServers] = useState<MCPServerEntry[]>([]);
  const [showEdit, setShowEdit] = useState(false);
  const [editing, setEditing] = useState<MCPServerEntry | null>(null);
  const [M, setM] = useState<MCPServerEntry>({
    name: '',
    command: '',
    args: [],
    env: {},
    url: '',
    enabled: true,
    icon: '🔌',
    color: '#6b7280',
  });
  const [showRepoDialog, setShowRepoDialog] = useState(false);
  const [repoUrl, setRepoUrl] = useState('');
  const [installing, setInstalling] = useState(false);

  useEffect(() => {
    loadMCPServers();
  }, []);

  const loadMCPServers = async () => {
    try {
      const config = await api.getAdaConfig();
      if (config?.mcp_servers) {
        const serverList = Object.entries(config.mcp_servers).map(
          ([name, cfg]) => ({
            name,
            command: cfg.command || '',
            args: cfg.args || [],
            env: cfg.env || {},
            url: cfg.url || '',
            enabled: cfg.enabled ?? true,
            icon: cfg.icon || '🔌',
            color: cfg.color || '#6b7280',
          } as MCPServerEntry),
        );
        setServers(serverList);
      }
    } catch (e) {
      console.error('[MCPSection] load error:', e);
    }
  };

  const handleSave = async () => {
    const updated = [...servers];
    if (editing) {
      const idx = updated.findIndex((s) => s.name === editing.name);
      if (idx >= 0) updated[idx] = M;
    } else {
      updated.push(M);
    }
    setServers(updated);
    await saveMCPServers(updated);
    setShowEdit(false);
    setEditing(null);
  };

  const handleDelete = async (name: string) => {
    const updated = servers.filter((s) => s.name !== name);
    setServers(updated);
    await saveMCPServers(updated);
  };

  const saveMCPServers = async (serverList: MCPServerEntry[]) => {
    try {
      const config = await api.getAdaConfig();
      if (config) {
        const mcpServers: Record<string, api.backend.MCPServerUI> = {};
        serverList.forEach((s) => {
          mcpServers[s.name] = {
            command: s.command,
            args: s.args,
            env: s.env,
            url: s.url,
            enabled: s.enabled,
            icon: s.icon,
            color: s.color,
          };
        });
        const newConfig = {
          ...config,
          mcp_servers: mcpServers,
        };
        await api.setAdaConfig(newConfig);
      }
    } catch (e) {
      console.error('[MCPSection] save error:', e);
    }
  };

  const handleInstallFromRepo = async () => {
    if (!repoUrl.trim()) return;
    setInstalling(true);
    try {
      await api.installMCPFromRepo(repoUrl);
      await loadMCPServers();
      setRepoUrl('');
      setShowRepoDialog(false);
    } catch (e) {
      console.error('[MCPSection] install error:', e);
    } finally {
      setInstalling(false);
    }
  };

  const openEdit = (server: MCPServerEntry | null = null) => {
    if (server) {
      setEditing(server);
      setM({ ...server });
    } else {
      setEditing(null);
      setM({
        name: '',
        command: '',
        args: [],
        env: {},
        url: '',
        enabled: true,
        icon: '🔌',
        color: '#6b7280',
      });
    }
    setShowEdit(true);
  };

  const addEnvVar = () => {
    setM({ ...M, env: { ...M.env, '': '' } });
  };

  const updateEnvKey = (oldKey: string, newKey: string) => {
    const newEnv = { ...M.env };
    if (oldKey !== newKey) {
      delete newEnv[oldKey];
    }
    newEnv[newKey] = newEnv[newKey] ?? '';
    setM({ ...M, env: newEnv });
  };

  const updateEnvValue = (key: string, value: string) => {
    setM({ ...M, env: { ...M.env, [key]: value } });
  };

  const removeEnvVar = (key: string) => {
    const newEnv = { ...M.env };
    delete newEnv[key];
    setM({ ...M, env: newEnv });
  };

  const addArg = () => {
    setM({ ...M, args: [...M.args, ''] });
  };

  const updateArg = (index: number, value: string) => {
    const newArgs = [...M.args];
    newArgs[index] = value;
    setM({ ...M, args: newArgs });
  };

  const removeArg = (index: number) => {
    setM({ ...M, args: M.args.filter((_, i) => i !== index) });
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-foreground">MCP Servers</h3>
          <p className="text-sm text-muted-foreground">
            Configure Model Context Protocol servers to extend tool
            capabilities.
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={() => setShowRepoDialog(true)}
          >
            <Icon name="Download" size={16} />
            Install from Repo
          </Button>
          <Button size="sm" onClick={() => openEdit()}>
            <Icon name="Plus" size={16} />
            New Server
          </Button>
        </div>
      </div>

      <div className="agents-grid">
        {servers.map((server) => {
          const color = server.color || '#6b7280';
          const icon = server.icon || (server.url ? '🌐' : '⚡');
          return (
            <BaseCard
              key={server.name}
              color={color}
              headerLeft={
                <span className="text-xs text-white opacity-90">
                  {server.url ? 'WebSocket' : 'CLI'}
                </span>
              }
              headerRight={
                <div className="flex gap-1">
                  <button
                    className="base-card-btn"
                    onClick={() => openEdit(server)}
                    title="Edit"
                  >
                    <Icon name="Edit" className="w-3 h-3" />
                  </button>
                  <button
                    className="base-card-btn"
                    onClick={() => handleDelete(server.name)}
                    title="Delete"
                  >
                    <Icon name="Trash2" className="w-3 h-3" />
                  </button>
                </div>
              }
              icon={icon}
              title={server.name}
            >
              <div className="text-xs text-muted-foreground truncate">
                {server.url || server.command || 'CLI Server'}
              </div>
              <div className="flex items-center justify-between mt-2">
                <span className="text-xs text-muted-foreground">Enabled</span>
                <Switch
                  checked={server.enabled}
onCheckedChange={(checked) => {
                    const updated = servers.map((s) =>
                      s.name === server.name ? { ...s, enabled: checked } : s,
                    );
                    setServers(updated);
                    saveMCPServers(updated);
                  }}
                />
              </div>
            </BaseCard>
          );
        })}
      </div>

      {servers.length === 0 && (
        <Card>
          <CardContent className="py-8">
            <p className="text-sm text-muted-foreground text-center">
              No MCP servers configured. Add one to enable MCP tools.
            </p>
          </CardContent>
        </Card>
      )}

      <EditDialog
        open={showEdit}
        onOpenChange={setShowEdit}
        title={editing ? 'Edit MCP Server' : 'New MCP Server'}
        onSave={handleSave}
        color={M.color || '#6b7280'}
        icon={M.icon || '🔌'}
        onColorChange={(color) => setM({ ...M, color })}
        onIconChange={(icon) => setM({ ...M, icon })}
      >
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Name</label>
            <Input
              value={M.name}
              onChange={(e) => setM({ ...M, name: e.target.value })}
              placeholder="e.g., my-mcp-server"
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Connection Type</label>
            <div className="flex gap-2">
              <Button
                type="button"
                variant={M.url ? 'default' : 'outline'}
                size="sm"
                onClick={() => setM({ ...M, url: '', command: '' })}
              >
                WebSocket/URL
              </Button>
              <Button
                type="button"
                variant={M.command ? 'default' : 'outline'}
                size="sm"
                onClick={() => setM({ ...M, command: '', url: '' })}
              >
                CLI Command
              </Button>
            </div>
          </div>

          {M.url ? (
            <div className="space-y-2">
              <label className="text-sm font-medium">URL</label>
              <Input
                value={M.url}
                onChange={(e) => setM({ ...M, url: e.target.value })}
                placeholder="ws://localhost:8080"
              />
            </div>
          ) : (
            <>
              <div className="space-y-2">
                <label className="text-sm font-medium">Command</label>
                <Input
                  value={M.command}
                  onChange={(e) => setM({ ...M, command: e.target.value })}
                  placeholder="npx my-mcp-server"
                />
              </div>

              {/* Args */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium">Arguments</label>
                  <Button type="button" size="sm" variant="outline" onClick={addArg}>
                    <Icon name="Plus" size={14} /> Add Arg
                  </Button>
                </div>
                {M.args.map((arg, index) => (
                  <div key={index} className="flex items-center gap-2">
                    <Input
                      value={arg}
                      onChange={(e) => updateArg(index, e.target.value)}
                      placeholder={`--arg-${index + 1}`}
                      className="flex-1"
                    />
                    <button
                      type="button"
                      className="base-card-btn"
                      onClick={() => removeArg(index)}
                      title="Remove"
                    >
                      <Icon name="X" className="w-3 h-3" />
                    </button>
                  </div>
                ))}
              </div>

              {/* Env vars */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium">Environment Variables</label>
                  <Button type="button" size="sm" variant="outline" onClick={addEnvVar}>
                    <Icon name="Plus" size={14} /> Add Env Var
                  </Button>
                </div>
                {Object.entries(M.env).map(([key, value], index) => (
                  <div key={index} className="flex items-center gap-2">
                    <Input
                      value={key}
                      onChange={(e) => updateEnvKey(key, e.target.value)}
                      placeholder="KEY"
                      className="w-32 font-mono text-xs"
                    />
                    <Input
                      value={value}
                      onChange={(e) => updateEnvValue(key, e.target.value)}
                      placeholder="value"
                      className="flex-1"
                    />
                    <button
                      type="button"
                      className="base-card-btn"
                      onClick={() => removeEnvVar(key)}
                      title="Remove"
                    >
                      <Icon name="X" className="w-3 h-3" />
                    </button>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </EditDialog>

      <Dialog open={showRepoDialog} onOpenChange={setShowRepoDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Install MCP Server from Repository</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <Input
              value={repoUrl}
              onChange={(e) => setRepoUrl(e.target.value)}
              placeholder="https://github.com/user/mcp-server"
            />
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowRepoDialog(false)}
              disabled={installing}
            >
              Cancel
            </Button>
            <Button onClick={handleInstallFromRepo} disabled={installing}>
              {installing ? 'Installing...' : 'Install'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default MCPSection;