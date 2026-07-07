import { useState, useEffect } from 'react';
import { WorkspaceSidebar } from './WorkspaceSidebar';
import { Icon } from './Icon';
import { Button } from './ui/button';
import { cn } from '@/lib/utils';
import * as api from '../api';

interface ChatItem {
  id: string;
  title: string;
  worker: string;
  summarized: boolean;
}

interface MainLayoutProps {
  onOpenSettings: () => void;
}

export function MainLayout({ onOpenSettings }: MainLayoutProps) {
  const [selectedWorkspace, setSelectedWorkspace] = useState<api.backend.WorkspaceConfig | null>(null);
  const [activeChatId, setActiveChatId] = useState<string | null>(null);
  const [chats, setChats] = useState<ChatItem[]>([]);
  const [draft, setDraft] = useState('');

  // Reseta os chats locais ao trocar de workspace.
  useEffect(() => {
    setChats([]);
    setActiveChatId(null);
  }, [selectedWorkspace?.path]);

  const handleAddChat = (
    workspace: api.backend.WorkspaceConfig,
    worker: api.backend.WorkerConfig,
    summarized: boolean,
  ) => {
    const id = `${workspace.path}::${worker.name}::${Date.now()}`;
    const title = summarized
      ? `Resumo • ${worker.name}`
      : `Chat • ${worker.name}`;
    const newChat: ChatItem = { id, title, worker: worker.name, summarized };
    setChats((prev) => [...prev, newChat]);
    setActiveChatId(id);
  };

  const activeChat = chats.find((c) => c.id === activeChatId) || null;

  return (
    <div className="flex flex-col h-screen w-screen overflow-hidden bg-background">
      {/* Área principal acima da toolbar */}
      <div className="flex flex-1 min-h-0">
        <div className="workspace-sidebar shrink-0 h-full flex flex-col">
          <WorkspaceSidebar
            selectedPath={selectedWorkspace?.path ?? null}
            onSelect={setSelectedWorkspace}
            onAddChat={handleAddChat}
          />

          {/* Lista de chats do workspace selecionado */}
          <div className="flex-1 min-h-0 overflow-y-auto border-t border-border p-1.5 space-y-0.5">
            {selectedWorkspace && chats.length === 0 && (
              <div className="px-2 py-2 text-[10px] text-muted-foreground">
                Use o + acima para adicionar um chat.
              </div>
            )}
            {selectedWorkspace && chats.map((chat) => (
              <button
                key={chat.id}
                type="button"
                onClick={() => setActiveChatId(chat.id)}
                className={cn(
                  'w-full flex items-center gap-2 px-2 py-1 rounded text-left text-[11px] leading-tight transition-colors',
                  activeChatId === chat.id
                    ? 'bg-accent text-accent-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground',
                )}
                title={chat.title}
              >
                <Icon name={chat.summarized ? 'FileText' : 'MessageSquare'} className="w-3 h-3 shrink-0" />
                <span className="truncate">{chat.title}</span>
              </button>
            ))}
          </div>
        </div>

        <div className="flex flex-col flex-1 min-w-0 min-h-0 border-l border-border">
          {/* Área de mensagens */}
          <div className="flex-1 min-h-0 overflow-y-auto p-4">
            {activeChat ? (
              <div className="h-full flex flex-col items-center justify-center text-center text-sm text-muted-foreground gap-2">
                <Icon name={activeChat.summarized ? 'FileText' : 'MessageSquare'} className="w-8 h-8 opacity-40" />
                <span className="font-medium text-foreground">{activeChat.title}</span>
                <span className="max-w-md">
                  {activeChat.summarized
                    ? 'Chat sumarizado — espaço reservado para o resumo da conversa atual.'
                    : 'Espaço reservado para a conversa.'}
                </span>
              </div>
            ) : selectedWorkspace ? (
              <div className="h-full flex flex-col items-center justify-center text-center text-sm text-muted-foreground gap-2">
                <span className="text-3xl">{selectedWorkspace.icon || '📂'}</span>
                <span className="font-medium text-foreground">{selectedWorkspace.title}</span>
                <span className="max-w-md">Selecione ou crie um chat no painel à esquerda.</span>
              </div>
            ) : (
              <div className="h-full flex flex-col items-center justify-center text-sm text-muted-foreground gap-2">
                <Icon name="MessageSquare" className="w-8 h-8 opacity-40" />
                <span>Selecione um workspace para começar.</span>
              </div>
            )}
          </div>

          {/* Área de input */}
          <div className="shrink-0 border-t border-border bg-card p-3">
            <div className="flex items-end gap-2 max-w-4xl mx-auto">
              <textarea
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                placeholder={activeChat ? 'Mensagem...' : 'Selecione um chat para digitar'}
                disabled={!activeChat}
                rows={1}
                className="flex-1 min-h-[40px] max-h-32 resize-none rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:opacity-50"
              />
              <Button
                size="icon"
                className="h-10 w-10 shrink-0"
                disabled={!activeChat}
                onClick={() => {
                  if (!draft.trim()) return;
                  setDraft('');
                }}
              >
                <Icon name="Plus" className="w-4 h-4 rotate-45" />
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Toolbar inferior de 22px */}
      <div className="h-[22px] shrink-0 w-full border-t border-border bg-card flex items-center px-1.5 gap-1">
        <button
          type="button"
          onClick={onOpenSettings}
          className="toolbar-btn"
          title="Configurações"
        >
          <Icon name="Settings" className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}
