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
import { EditDialog } from '../EditDialog';
import { IconPicker } from '../IconPicker';
import { Icon } from '../Icon';
import * as api from '../../api';

interface MCPServerConfig {
  name: string;
  command: string;
  args: string[];
  url: string;
  enabled: boolean;
  icon: string;
}

function MCPSection() {
  const [servers, setServers] = useState<MCPServerConfig[]>([]);
  const [showEdit, setShowEdit] = useState(false);
  const [editing, setEditing] = useState<MCPServerConfig | null>(null);
  const [M, setM] = useState<MCPServerConfig>({
    name: '',
    command: '',
    args: [],
    url: '',
    enabled: true,
    icon: '🔌',
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
      if (config?.tools?.mcp?.servers) {
        const serverList = Object.entries(config.tools.mcp.servers).map(
          ([name, cfg]: [string, any]) => ({
            name,
            command: cfg.command || '',
            args: cfg.args || [],
            url: cfg.url || '',
            enabled: cfg.enabled ?? true,
            icon: cfg.icon || '🔌',
          }),
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

  const saveMCPServers = async (serverList: MCPServerConfig[]) => {
    try {
      const config = await api.getAdaConfig();
      if (config) {
        const mcpServers: Record<string, any> = {};
        serverList.forEach((s) => {
          mcpServers[s.name] = {
            command: s.command,
            args: s.args,
            url: s.url,
            enabled: s.enabled,
            icon: s.icon,
          };
        });
        const newConfig = {
          ...config,
          tools: {
            ...config.tools,
            mcp: {
              ...config.tools?.mcp,
              servers: mcpServers,
            },
          },
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

  const openEdit = (server: MCPServerConfig | null = null) => {
    if (server) {
      setEditing(server);
      setM(server);
    } else {
      setEditing(null);
      setM({
        name: '',
        command: '',
        args: [],
        url: '',
        enabled: true,
        icon: '🔌',
      });
    }
    setShowEdit(true);
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

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {servers.map((server) => (
          <Card key={server.name} className="relative group">
            <CardContent className="p-4">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-lg">{server.url ? '🌐' : '⚡'}</span>
                  <div>
                    <div className="text-sm font-medium">{server.name}</div>
                    <div className="text-xs text-muted-foreground">
                      {server.url || server.command || 'CLI Server'}
                    </div>
                  </div>
                </div>
                <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
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
              </div>
              <div className="flex items-center justify-between mt-3">
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
            </CardContent>
          </Card>
        ))}
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
                onClick={() => setM({ ...M, url: '', command: '' })}
              >
                WebSocket/URL
              </Button>
              <Button
                type="button"
                variant={M.command ? 'default' : 'outline'}
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
            <div className="space-y-2">
              <label className="text-sm font-medium">Command</label>
              <Input
                value={M.command}
                onChange={(e) => setM({ ...M, command: e.target.value })}
                placeholder="npx my-mcp-server"
              />
            </div>
          )}

          <div className="flex items-center gap-2">
            <span className="text-2xl">{M.icon || '🔌'}</span>
            <IconPicker
              value={M.icon || '🔌'}
              onChange={(v) => setM({ ...M, icon: v })}
            />
          </div>
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
