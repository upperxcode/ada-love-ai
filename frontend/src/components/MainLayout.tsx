import { useState, useEffect } from 'react';
import { WorkspaceTree } from './WorkspaceTree';
import { ChatArea } from './ChatArea';
import { ChatInput } from './ChatInput';
import { EventsContainer } from './EventsContainer';
import { OrchestratorPanel } from './OrchestratorPanel';
import { CommandResultPanel } from './CommandResultPanel';
import { Icon } from './Icon';
import { useTheme } from '../lib/theme';
import * as api from '../api';
import { useChat, ChatProvider } from './ChatContext';

function ChatLayoutContent({ onOpenSettings }: { onOpenSettings: () => void }) {
  const [workspaces, setWorkspaces] = useState<api.backend.WorkspaceConfig[]>([]);
  const [workers, setWorkers] = useState<api.backend.WorkerConfig[]>([]);
  const [sidebarVisible, setSidebarVisible] = useState(true);
  const { dark, setDark } = useTheme();
  const { createSession, loadAllSessions } = useChat();

  useEffect(() => {
    console.log('[MainLayout] Mounting: loading workspaces & workers...');
    api.getWorkspaces().then((ws) => {
      console.log(`[MainLayout] Loaded ${ws.length} workspaces:`, ws.map((w) => w.title || w.path));
      setWorkspaces(ws);
      // Eagerly load sessions for ALL workspaces using their backend path (empty string is valid for orphaned sessions)
      const paths = ws.map((w) => w.path);
      // Deduplicate to avoid redundant calls
      const uniquePaths = [...new Set(paths)];
      console.log(`[MainLayout] Eager loading sessions for ${uniquePaths.length} unique workspace paths...`);
      loadAllSessions(uniquePaths);
    }).catch((e) => {
      console.error('[MainLayout] Failed to load workspaces:', e);
      setWorkspaces([]);
    });
    api.getWorkers().then((w) => {
      console.log(`[MainLayout] Loaded ${w.length} workers:`, w.map((wk) => wk.name));
      setWorkers(w);
    }).catch((e) => {
      console.error('[MainLayout] Failed to load workers:', e);
      setWorkers([]);
    });
  }, [loadAllSessions]);

  return (
    <div className="flex flex-col h-screen w-screen overflow-hidden bg-background">
      <div className="flex flex-1 min-h-0">
        {sidebarVisible && (
          <div className="workspace-sidebar shrink-0 h-full flex flex-col">
            <WorkspaceTree
              workspaces={workspaces}
              worker_names={workers.map((w) => w.name)}
              workers={workers}
              onAddChat={createSession}
              onWorkspacesChanged={() => {
                api.getWorkspaces().then(setWorkspaces).catch(() => {});
              }}
            />
          </div>
        )}

        <div className="relative flex flex-col flex-1 min-w-0 min-h-0 border-l border-border">
          {/* Chat area - histórico de conversas */}
          <ChatArea />
          {/* Orchestrator panel - progresso do multi-agent */}
          <OrchestratorPanel />
          {/* Events container - notificações de status */}
          <EventsContainer />
          {/* Input area */}
          <ChatInput />
          {/* Floating panel for slash command results (e.g. /health) */}
          <CommandResultPanel />
        </div>
      </div>

      <div className="h-[22px] shrink-0 w-full border-t border-border bg-card flex items-center px-1.5 gap-1">
        <button
          type="button"
          onClick={() => setSidebarVisible(!sidebarVisible)}
          className="toolbar-btn"
          title={sidebarVisible ? 'Hide panel' : 'Show panel'}
        >
          <Icon name={sidebarVisible ? 'PanelLeftClose' : 'PanelLeft'} className="w-4 h-4" />
        </button>

        <span className="flex-1" />

        <button
          type="button"
          onClick={() => setDark(!dark)}
          className="toolbar-btn"
          title={dark ? 'Light mode' : 'Dark mode'}
        >
          <Icon name={dark ? 'Sun' : 'Moon'} className="w-4 h-4" />
        </button>

        <span className="w-px h-3 bg-border mx-0.5" />

        <button
          type="button"
          onClick={onOpenSettings}
          className="toolbar-btn"
          title="Settings"
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
