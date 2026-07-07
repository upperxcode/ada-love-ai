import { useState, useEffect } from 'react';
import { WorkspaceTree } from './WorkspaceTree';
import { ChatArea } from './ChatArea';
import { ChatInput } from './ChatInput';
import { Icon } from './Icon';
import * as api from '../api';
import { useChat, ChatProvider } from './ChatContext';

function ChatLayoutContent({ onOpenSettings }: { onOpenSettings: () => void }) {
  const [workspaces, setWorkspaces] = useState<api.backend.WorkspaceConfig[]>([]);
  const [workers, setWorkers] = useState<api.backend.WorkerConfig[]>([]);
  const { createSession } = useChat();

  useEffect(() => {
    api.getWorkspaces().then(setWorkspaces).catch(() => setWorkspaces([]));
    api.getWorkers().then(setWorkers).catch(() => setWorkers([]));
  }, []);

  return (
    <div className="flex flex-col h-screen w-screen overflow-hidden bg-background">
      <div className="flex flex-1 min-h-0">
        <div className="workspace-sidebar shrink-0 h-full flex flex-col">
          <WorkspaceTree
            workspaces={workspaces}
            workers={workers}
            onAddChat={createSession}
          />
        </div>

        <div className="flex flex-col flex-1 min-w-0 min-h-0 border-l border-border">
          <ChatArea />
          <ChatInput />
        </div>
      </div>

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

interface MainLayoutProps {
  onOpenSettings: () => void;
}

export function MainLayout({ onOpenSettings }: MainLayoutProps) {
  return (
    <ChatProvider>
      <ChatLayoutContent onOpenSettings={onOpenSettings} />
    </ChatProvider>
  );
}
