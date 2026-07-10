import { useState, useEffect, useMemo } from 'react';
import { cn } from '@/lib/utils';
import { Icon } from './Icon';
import { Button } from './ui/button';
import { LuPlus, LuCirclePlus } from 'react-icons/lu';
import {
  Popover,
  PopoverTrigger,
  PopoverContent,
} from './ui/popover';
import { useChat } from './ChatContext';
import * as api from '../api';

interface WorkspaceTreeProps {
  workspaces: api.backend.WorkspaceConfig[];
  workers: api.backend.WorkerConfig[];
  onAddChat: (
    workspace: api.backend.WorkspaceConfig,
    worker: api.backend.WorkerConfig,
    summarized: boolean,
  ) => void;
  onWorkspacesChanged: () => void;
}

function relativeTime(iso: string): string {
  if (!iso) return '';
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return '';
  const now = Date.now();
  const diff = now - date.getTime();
  const min = Math.floor(diff / 60000);
  if (min < 1) return 'agora';
  if (min < 60) return `${min}m`;
  const hours = Math.floor(min / 60);
  if (hours < 24) return `${hours}h`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d`;
  const weeks = Math.floor(days / 7);
  if (weeks < 4) return `${weeks}sem`;
  return date.toLocaleDateString();
}

function ChatRow({ session }: { session: api.backend.ChatSession }) {
  const {
    activeSessionId,
    selectSession,
    deleteSession,
    renameSession,
    togglePinSession,
  } = useChat();
  const [renaming, setRenaming] = useState(false);
  const [draft, setDraft] = useState(session.title);
  const active = activeSessionId === session.id;
  const summarized = session.parent_session_id !== '' || session.title.toLowerCase().startsWith('resumo');

  return (
    <div
      className={cn(
        'group flex items-center gap-2 px-3 py-2 cursor-pointer text-left transition-colors border-b border-border/40',
        active
          ? 'bg-accent text-accent-foreground'
          : 'text-foreground/80 hover:bg-muted',
      )}
      onClick={() => !renaming && selectSession(session.id)}
    >
      <Icon
        name={summarized ? 'FileText' : 'MessageSquare'}
        className={cn(
          'w-4 h-4 shrink-0',
          active ? 'text-accent-foreground' : 'text-muted-foreground',
        )}
      />
      <div className="flex-1 min-w-0">
        {renaming ? (
          <input
            autoFocus
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onBlur={() => {
              if (draft.trim()) renameSession(session.id, draft.trim());
              setRenaming(false);
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                if (draft.trim()) renameSession(session.id, draft.trim());
                setRenaming(false);
              }
              if (e.key === 'Escape') setRenaming(false);
            }}
            className="w-full min-w-0 bg-transparent outline-none border-b border-current text-[12px]"
            onClick={(e) => e.stopPropagation()}
          />
        ) : (
          <div className="flex items-center justify-between gap-2">
            <span className="truncate text-[12px] font-medium leading-tight">
              {session.pinned && <span className="mr-1 opacity-70">★</span>}
              {session.title}
            </span>
            <span className="text-[10px] text-muted-foreground shrink-0 tabular-nums">
              {relativeTime(session.updated_at)}
            </span>
          </div>
        )}
        {session.worker_name && !renaming && (
          <div className="truncate text-[10px] text-muted-foreground leading-tight mt-0.5">
            {session.worker_name}
          </div>
        )}
      </div>

      {!renaming && (
        <div className="hidden group-hover:flex items-center gap-0.5 shrink-0">
          <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5 p-0"
            title="Fixar"
            onClick={(e) => {
              e.stopPropagation();
              togglePinSession(session.id);
            }}
          >
            <Icon
              name="Star"
              className={cn('w-3 h-3', session.pinned && 'fill-current')}
            />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5 p-0"
            title="Renomear"
            onClick={(e) => {
              e.stopPropagation();
              setRenaming(true);
            }}
          >
            <Icon name="Edit" className="w-3 h-3" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5 p-0"
            title="Excluir"
            onClick={(e) => {
              e.stopPropagation();
              deleteSession(session.id);
            }}
          >
            <Icon name="Trash2" className="w-3 h-3" />
          </Button>
        </div>
      )}
    </div>
  );
}

function WorkerNode({
  worker,
  workspace,
  sessions,
  selectedWorker,
  onSelectWorker,
  onAddChat,
}: {
  worker: api.backend.WorkerConfig;
  workspace: api.backend.WorkspaceConfig;
  sessions: api.backend.ChatSession[];
  selectedWorker: string | null;
  onSelectWorker: (name: string | null) => void;
  onAddChat: (
    workspace: api.backend.WorkspaceConfig,
    worker: api.backend.WorkerConfig,
    summarized: boolean,
  ) => void;
}) {
  const [open, setOpen] = useState(false);
  const active = selectedWorker === worker.name;

  return (
    <div className="ml-3 border-l border-border/50">
      <div
        className={cn(
          'group flex items-center justify-between gap-1 px-2 py-1.5 cursor-pointer text-[11px] transition-colors',
          active
            ? 'text-foreground'
            : 'text-muted-foreground hover:text-foreground',
        )}
        onClick={() => onSelectWorker(active ? null : worker.name)}
      >
        <div className="flex items-center gap-1.5 min-w-0">
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              setOpen(!open);
            }}
            className="shrink-0 p-0.5 rounded hover:bg-black/10 dark:hover:bg-white/10"
          >
            <Icon
              name="ChevronRight"
              className={cn('w-3 h-3 transition-transform', open && 'rotate-90')}
            />
          </button>
          <span className="truncate font-medium">{worker.name}</span>
        </div>

        {active && (
          <div className="flex items-center gap-0.5 shrink-0 worker-plus-actions">
            <button
              type="button"
              className="toolbar-btn"
              title="Novo chat"
              onClick={(e) => {
                e.stopPropagation();
                onAddChat(workspace, worker, false);
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
                onAddChat(workspace, worker, true);
              }}
            >
              <LuCirclePlus className="w-3.5 h-3.5" />
            </button>
          </div>
        )}
      </div>

      {open && (
        <div className="ml-1 space-y-0.5">
          {sessions.length > 0 ? (
            sessions
              .slice()
              .sort((a, b) => {
                if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
                return a.title.localeCompare(b.title, 'pt-BR', { sensitivity: 'base' });
              })
              .map((session) => (
                <div key={session.id} className="-ml-3 border-l border-border/50">
                  <div className="ml-3">
                    <ChatRow session={session} />
                  </div>
                </div>
              ))
          ) : (
            <div className="px-2 py-1 text-[10px] text-muted-foreground">
              Nenhum chat.
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function AddWorkerPopover({
  workspace,
  globalWorkers,
  onSaved,
}: {
  workspace: api.backend.WorkspaceConfig;
  globalWorkers: api.backend.WorkerConfig[];
  onSaved: () => void;
}) {
  const [open, setOpen] = useState(false);

  const handleAdd = async (worker: api.backend.WorkerConfig) => {
    const currentWorkers = workspace.workers ?? [];
    const updatedWorkers = [...currentWorkers, worker];
    await api.updateWorkspace(workspace.title, {
      ...workspace,
      workers: updatedWorkers,
    });
    setOpen(false);
    onSaved();
  };

  // Filtra workers que ainda não estão no workspace
  const available = globalWorkers.filter(
    (gw) => !(workspace.workers ?? []).some((w) => w.name === gw.name),
  );

  if (available.length === 0) return null;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="toolbar-btn"
          title="Adicionar worker ao workspace"
          onClick={(e) => e.stopPropagation()}
        >
          <Icon name="Plus" className="w-3 h-3" />
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-56 p-2" align="end" side="right">
        <div className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider px-2 py-1">
          Adicionar worker
        </div>
        <div className="mt-1 space-y-0.5">
          {available.map((worker) => (
            <button
              key={worker.name}
              type="button"
              onClick={() => handleAdd(worker)}
              className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-left text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            >
              <span>{worker.icon || '🤖'}</span>
              <span className="truncate">{worker.name}</span>
            </button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  );
}

function WorkspacePlusButton({
  workspace,
  onAddChat,
}: {
  workspace: api.backend.WorkspaceConfig;
  onAddChat: (
    workspace: api.backend.WorkspaceConfig,
    worker: api.backend.WorkerConfig,
    summarized: boolean,
  ) => void;
}) {
  const [selectedWorker, setSelectedWorker] = useState<string | null>(null);
  const [open, setOpen] = useState(false);
  const workers = workspace.workers ?? [];

  const handleAdd = (worker: api.backend.WorkerConfig, summarized: boolean) => {
    onAddChat(workspace, worker, summarized);
    setOpen(false);
    setSelectedWorker(null);
  };

  // Se não tem workers, mostra mensagem
  if (workers.length === 0) {
    return (
      <span className="text-[9px] text-muted-foreground px-1" title="Adicione workers ao workspace nas configurações">
        +Workers
      </span>
    );
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="h-5 w-5 p-0"
          title="Adicionar chat"
          onClick={(e) => e.stopPropagation()}
        >
          <Icon name="Plus" className="w-3.5 h-3.5" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-60 p-2" align="end" side="right">
        <div className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider px-2 py-1">
          Workers neste workspace
        </div>
        <div className="mt-1 space-y-0.5">
          {workers.filter((w) => w.name).map((worker) => {
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
              >
                <span className="truncate">{worker.icon || '🤖'} {worker.name}</span>
                {active && (
                  <span className="flex gap-0.5">
                    <button
                      type="button"
                      className="toolbar-btn"
                      title="Novo chat"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleAdd(worker, false);
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
                        handleAdd(worker, true);
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
  );
}

function WorkspaceNode({
  workspace,
  workers,
  globalWorkers,
  sessions,
  expanded,
  selectedWorkspace,
  selectedWorker,
  onToggle,
  onSelect,
  onSelectWorker,
  onAddChat,
  onWorkspacesChanged,
}: {
  workspace: api.backend.WorkspaceConfig;
  workers: api.backend.WorkerConfig[];
  globalWorkers: api.backend.WorkerConfig[];
  sessions: api.backend.ChatSession[];
  expanded: boolean;
  selectedWorkspace: string | null;
  selectedWorker: string | null;
  onToggle: (e: React.MouseEvent) => void;
  onSelect: () => void;
  onSelectWorker: (name: string | null) => void;
  onAddChat: (
    workspace: api.backend.WorkspaceConfig,
    worker: api.backend.WorkerConfig,
    summarized: boolean,
  ) => void;
  onWorkspacesChanged: () => void;
}) {
  const active = selectedWorkspace === workspace.path || selectedWorkspace === (workspace.path || workspace.title);

  return (
    <div className="border-b border-border/40">
      <div
        className={cn(
          'flex items-center gap-1.5 px-3 py-2.5 text-left text-[12px] font-semibold transition-colors cursor-pointer',
          active
            ? 'text-foreground'
            : 'text-foreground/80 hover:bg-muted',
        )}
        onClick={onSelect}
      >
        <button
          type="button"
          onClick={onToggle}
          className="shrink-0 p-0.5 rounded hover:bg-black/10 dark:hover:bg-white/10"
        >
          <Icon
            name="ChevronRight"
            className={cn('w-3.5 h-3.5 transition-transform', expanded && 'rotate-90')}
          />
        </button>
        <span className="text-base shrink-0">{workspace.icon || '📂'}</span>
        <span className="flex-1 truncate">{workspace.title || workspace.path}</span>
        {active && (
          <div className="flex items-center gap-0.5 shrink-0">
            <WorkspacePlusButton workspace={workspace} onAddChat={onAddChat} />
            <AddWorkerPopover workspace={workspace} globalWorkers={globalWorkers} onSaved={onWorkspacesChanged} />
          </div>
        )}
        {!workspace.enabled && !active && (
          <span className="text-[9px] opacity-60">off</span>
        )}
      </div>

      {expanded && (
        <div className="pb-2 space-y-0.5">
          {workers.filter((w) => w.name).length === 0 && (
            <div className="px-5 py-2 text-[10px] text-muted-foreground">
              Nenhum worker neste workspace.
            </div>
          )}
          {workers.filter((w) => w.name).map((worker) => (
            <WorkerNode
              key={`${workspace.path || workspace.title}:${worker.name}`}
              worker={worker}
              workspace={workspace}
              sessions={sessions.filter((s) => s.worker_name === worker.name)}
              selectedWorker={selectedWorker}
              onSelectWorker={onSelectWorker}
              onAddChat={onAddChat}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function WorkspaceTree({
  workspaces,
  workers,
  onAddChat,
  onWorkspacesChanged,
}: WorkspaceTreeProps) {
  const {
    sessions,
    activeSessionPath,
    activeSessionWorker,
    loadSessions,
  } = useChat();

  const [selectedWorkspace, setSelectedWorkspace] = useState<string | null>(
    activeSessionPath,
  );
  const [expandedWorkspaces, setExpandedWorkspaces] = useState<Set<string>>(
    () => new Set(),
  );
  const [selectedWorker, setSelectedWorker] = useState<string | null>(
    activeSessionWorker,
  );
  const [query, setQuery] = useState('');

  // Expand ancestors of active chat and load its sessions
  useEffect(() => {
    if (activeSessionPath && activeSessionWorker) {
      setSelectedWorkspace(activeSessionPath);
      setSelectedWorker(activeSessionWorker);
      setExpandedWorkspaces((prev) => new Set([...prev, activeSessionPath]));
      loadSessions(activeSessionPath);
    }
  }, [activeSessionPath, activeSessionWorker, loadSessions]);

  // Load sessions for all workspaces when searching (to populate flat list)
  const isSearching = query.trim().length > 0;
  useEffect(() => {
    if (isSearching) {
      workspaces.forEach((ws) => loadSessions(ws.path || ws.title));
    }
  }, [isSearching, workspaces, loadSessions]);

  const toggleWorkspace = (path: string) => {
    setExpandedWorkspaces((prev) => {
      const next = new Set(prev);
      if (next.has(path)) next.delete(path);
      else next.add(path);
      return next;
    });
  };

  const handleSelectWorkspace = (path: string) => {
    setSelectedWorkspace(path);
  };

  const handleSelectWorker = (path: string, name: string | null) => {
    setSelectedWorkspace(path);
    setExpandedWorkspaces((prev) => {
      const next = new Set(prev);
      next.add(path);
      return next;
    });
    setSelectedWorker(name);
  };

  const handleAddChat = (
    workspace: api.backend.WorkspaceConfig,
    worker: api.backend.WorkerConfig,
    summarized: boolean,
  ) => {
    const wsPath = workspace.path || workspace.title;
    setSelectedWorkspace(wsPath);
    setExpandedWorkspaces((prev) => new Set([...prev, wsPath]));
    setSelectedWorker(worker.name);
    onAddChat(workspace, worker, summarized);
  };

  const filteredSessions = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return [];
    return sessions
      .filter((s) => s.title.toLowerCase().includes(q))
      .sort((a, b) => {
        if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
        return a.title.localeCompare(b.title, 'pt-BR', { sensitivity: 'base' });
      });
  }, [sessions, query]);

  return (
    <div className="flex flex-col h-full bg-sidebar">
      {/* Search header */}
      <div className="shrink-0 p-2 border-b border-border">
        <div className="relative">
          <Icon
            name="Search"
            className="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
          />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Buscar chats…"
            className="w-full pl-8 pr-7 py-1.5 rounded-md bg-background border border-border text-[12px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
          />
          {query && (
            <button
              type="button"
              onClick={() => setQuery('')}
              className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              title="Limpar"
            >
              <Icon name="X" className="w-3 h-3" />
            </button>
          )}
        </div>
      </div>

      {/* Body */}
      {isSearching ? (
        <div className="flex-1 min-h-0 overflow-y-auto">
          {filteredSessions.length === 0 ? (
            <div className="px-4 py-6 text-center text-[11px] text-muted-foreground">
              Nenhum chat encontrado.
            </div>
          ) : (
            <>
              {filteredSessions.some((s) => s.pinned) && (
                <div className="px-3 pt-2 pb-1 text-[10px] font-semibold text-muted-foreground uppercase tracking-wider">
                  ★ Pinned
                </div>
              )}
              {filteredSessions
                .filter((s) => s.pinned)
                .map((session) => (
                  <ChatRow key={session.id} session={session} />
                ))}
              {filteredSessions.some((s) => !s.pinned) && (
                <div className="px-3 pt-2 pb-1 text-[10px] font-semibold text-muted-foreground uppercase tracking-wider">
                  Conversas
                </div>
              )}
              {filteredSessions
                .filter((s) => !s.pinned)
                .map((session) => (
                  <ChatRow key={session.id} session={session} />
                ))}
            </>
          )}
        </div>
      ) : (
        <div className="flex-1 min-h-0 overflow-y-auto">
          {workspaces.length === 0 && (
            <div className="px-3 py-3 text-[11px] text-muted-foreground">
              Nenhum workspace criado.
            </div>
          )}
          {workspaces.map((ws) => (
            <WorkspaceNode
              key={ws.path || ws.title}
              workspace={ws}
              workers={ws.workers ?? []}
              globalWorkers={workers}
              sessions={sessions.filter((s) => s.workspace_id === (ws.path || ws.title))}
              expanded={expandedWorkspaces.has(ws.path || ws.title)}
              selectedWorkspace={selectedWorkspace}
              selectedWorker={selectedWorker}
              onToggle={(e) => {
                e.stopPropagation();
                toggleWorkspace(ws.path || ws.title);
              }}
              onSelect={() => handleSelectWorkspace(ws.path || ws.title)}
              onSelectWorker={(name) => handleSelectWorker(ws.path || ws.title, name)}
              onAddChat={handleAddChat}
              onWorkspacesChanged={onWorkspacesChanged}
            />
          ))}
        </div>
      )}
    </div>
  );
}
