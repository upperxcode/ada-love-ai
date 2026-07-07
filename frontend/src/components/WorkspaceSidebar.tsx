import { useState, useEffect } from 'react';
import { cn } from '@/lib/utils';
import { Icon } from './Icon';
import { Button } from './ui/button';
import {
  Popover,
  PopoverTrigger,
  PopoverContent,
} from './ui/popover';
import { LuPlus, LuCirclePlus } from 'react-icons/lu';
import * as api from '../api';

interface WorkspaceSidebarProps {
  selectedPath: string | null;
  onSelect: (workspace: api.backend.WorkspaceConfig) => void;
  onAddChat: (workspace: api.backend.WorkspaceConfig, worker: api.backend.WorkerConfig, summarized: boolean) => void;
}

export function WorkspaceSidebar({ selectedPath, onSelect, onAddChat }: WorkspaceSidebarProps) {
  const [workspaces, setWorkspaces] = useState<api.backend.WorkspaceConfig[]>([]);
  const [workers, setWorkers] = useState<api.backend.WorkerConfig[]>([]);
  const [selectedWorker, setSelectedWorker] = useState<string | null>(null);

  useEffect(() => {
    api.getWorkspaces().then(setWorkspaces).catch(() => setWorkspaces([]));
    api.getWorkers().then(setWorkers).catch(() => setWorkers([]));
  }, []);

  return (
    <div className="workspace-sidebar h-full flex flex-col border-r border-border bg-card">
      <div className="workspace-sidebar-header flex items-center justify-between px-3 py-2 border-b border-border">
        <span className="text-xs font-semibold text-foreground uppercase tracking-wider">
          Workspaces
        </span>
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="h-5 w-5 p-0"
              title="Adicionar chat por worker"
            >
              <Icon name="Plus" className="w-4 h-4" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-64 p-2" align="end" side="right">
            <div className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider px-2 py-1">
              Workers
            </div>
            <div className="mt-1 space-y-0.5">
              {workers.length === 0 && (
                <div className="px-2 py-2 text-[11px] text-muted-foreground">
                  Nenhum worker registrado.
                </div>
              )}
              {workers.map((worker) => {
                const active = selectedWorker === worker.name;
                return (
                  <button
                    key={worker.name}
                    type="button"
                    onClick={() => setSelectedWorker(worker.name)}
                    className={cn(
                      'w-full flex items-center justify-between gap-1 px-2 py-1.5 rounded text-left text-[11px] transition-colors',
                      active
                        ? 'bg-accent text-accent-foreground'
                        : 'text-muted-foreground hover:bg-muted hover:text-foreground',
                    )}
                    title={worker.name}
                  >
                    <span className="truncate">{worker.name}</span>
                    {active && (
                      <span className="flex items-center gap-0.5 shrink-0">
                        <button
                          type="button"
                          className="toolbar-btn"
                          title="Novo chat"
                          onClick={(e) => {
                            e.stopPropagation();
                            const ws = workspaces.find((w) => w.path === selectedPath);
                            if (ws) onAddChat(ws, worker, false);
                          }}
                        >
                          <LuPlus className="w-3.5 h-3.5" />
                        </button>
                        <button
                          type="button"
                          className="toolbar-btn"
                          title="Chat sumarizado"
                          onClick={(e) => {
                            e.stopPropagation();
                            const ws = workspaces.find((w) => w.path === selectedPath);
                            if (ws) onAddChat(ws, worker, true);
                          }}
                        >
                          <LuCirclePlus className="w-3.5 h-3.5" />
                        </button>
                      </span>
                    )}
                  </button>
                );
              })}
            </div>
          </PopoverContent>
        </Popover>
      </div>
      <div className="flex-1 overflow-y-auto p-1.5 space-y-0.5">
        {workspaces.length === 0 && (
          <div className="px-2 py-3 text-[11px] text-muted-foreground">
            Nenhum workspace criado.
          </div>
        )}
        {workspaces.map((ws) => {
          const active = selectedPath === ws.path;
          return (
            <button
              key={ws.path}
              type="button"
              onClick={() => onSelect(ws)}
              className={cn(
                'w-full flex items-center gap-2 px-2 py-1.5 rounded text-left text-[11px] leading-tight transition-colors',
                active
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:bg-muted hover:text-foreground',
              )}
              title={ws.title}
            >
              <span className="shrink-0 text-sm">
                {ws.icon || '📂'}
              </span>
              <span className="truncate">{ws.title || ws.path}</span>
              {!ws.enabled && (
                <span className="ml-auto text-[9px] opacity-60">off</span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
