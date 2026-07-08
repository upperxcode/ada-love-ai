import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  useRef,
} from 'react';
import * as api from '../api';
import { useModelHealth } from '@/lib/modelHealth';
import { useSnackbar } from './Snackbar';

export interface LocalMessage {
  id: string;
  role: 'user' | 'assistant' | 'system' | 'tool';
  content: string;
  streaming?: boolean;
  error?: boolean;
}

export interface ChatState {
  sessions: api.backend.ChatSession[];
  activeSessionId: string | null;
  messages: LocalMessage[];
  loading: boolean;
  stage: string;
}

interface ChatContextValue extends ChatState {
  activeSessionPath: string | null;
  activeSessionWorker: string | null;
  activeWorkspacePath: string | null;
  loadSessions: (workspacePath: string) => Promise<void>;
  createSession: (
    workspace: api.backend.WorkspaceConfig,
    worker: api.backend.WorkerConfig,
    summarized?: boolean,
  ) => Promise<api.backend.ChatSession | null>;
  selectSession: (sessionId: string | null) => void;
  sendMessage: (text: string, modelKey?: string, thinkingLevel?: string) => Promise<void>;
  deleteSession: (sessionId: string) => Promise<void>;
  renameSession: (sessionId: string, newTitle: string) => Promise<void>;
  togglePinSession: (sessionId: string) => Promise<void>;
}

const ChatContext = createContext<ChatContextValue | null>(null);

function mapMessages(session: api.backend.ChatSession | null): LocalMessage[] {
  if (!session) return [];
  return session.messages.map((m, idx) => ({
    id: `${session.id}-msg-${idx}`,
    role: (m.role as any) || 'assistant',
    content: m.content || '',
  }));
}

export function ChatProvider({ children }: { children: React.ReactNode }) {
  const [sessions, setSessions] = useState<api.backend.ChatSession[]>([]);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(() => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('ada:lastChatId');
  });
  const [messages, setMessages] = useState<LocalMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const [stage, setStage] = useState('');
  const { recordSuccess, recordFailure } = useModelHealth();
  const { showSnackbar } = useSnackbar();
  const recordFailureRef = useRef(recordFailure);
  recordFailureRef.current = recordFailure;

  const activeSession = sessions.find((s) => s.id === activeSessionId) || null;
  const activeSessionRef = useRef(activeSession);
  activeSessionRef.current = activeSession;
  const activeSessionPath = activeSession?.workspace_id ?? null;
  const activeSessionWorker = activeSession?.worker_name ?? null;

  // Persist active chat
  useEffect(() => {
    if (activeSessionId) {
      localStorage.setItem('ada:lastChatId', activeSessionId);
    } else {
      localStorage.removeItem('ada:lastChatId');
    }
  }, [activeSessionId]);

  // Subscribe to backend streaming events
  useEffect(() => {
    const unsubs: Array<() => void> = [];

    unsubs.push(
      api.onChatEvent('chat:delta', (payload: any) => {
        const { session_id, content } = payload || {};
        if (!session_id || !content) return;
        if (activeSessionRef.current?.id !== session_id) return;
        setMessages((prev) => {
          const last = prev[prev.length - 1];
          if (last && last.role === 'assistant' && last.streaming) {
            const updated = [...prev];
            updated[updated.length - 1] = {
              ...last,
              content: last.content + content,
            };
            return updated;
          }
          return [
            ...prev,
            {
              id: `${session_id}-stream-${Date.now()}`,
              role: 'assistant',
              content,
              streaming: true,
            },
          ];
        });
      }),
    );

    unsubs.push(
      api.onChatEvent('chat:turnEnd', () => {
        setLoading(false);
        setMessages((prev) =>
          prev.map((m) => (m.streaming ? { ...m, streaming: false } : m)),
        );
      }),
    );

    unsubs.push(
      api.onChatEvent('chat:error', (payload: any) => {
        setLoading(false);
        const msg = payload?.message || 'Erro desconhecido';
        showSnackbar(msg, 'error');
      }),
    );

    return () => {
      unsubs.forEach((u) => u());
    };
  }, []);

  const loadSessions = useCallback(async (workspacePath: string) => {
    const list = await api.getSessions(workspacePath);
    setSessions(list);
    setActiveSessionId((current) => {
      const stillExists = list.some((s) => s.id === current);
      return stillExists ? current : (list[0]?.id ?? null);
    });
  }, []);

  const selectSession = useCallback((sessionId: string | null) => {
    setActiveSessionId(sessionId);
  }, []);

  // Sync messages when active session changes
  useEffect(() => {
    setMessages(mapMessages(activeSession));
  }, [activeSessionId, activeSession?.id]);

  const createSession = useCallback(
    async (
      workspace: api.backend.WorkspaceConfig,
      worker: api.backend.WorkerConfig,
      summarized = false,
    ): Promise<api.backend.ChatSession | null> => {
      let raw: api.backend.ChatSession | null = null;
      if (summarized && activeSessionId) {
        raw = await api.createSummarizedSession(
          workspace.path,
          worker.name,
          activeSessionId,
        );
      } else {
        raw = await api.createSession(workspace.path, worker.name);
      }
      if (!raw) {
        showSnackbar('Não foi possível criar o chat.', 'error');
        return null;
      }
      const session = new api.backend.ChatSession(raw);
      setSessions((prev) => [session, ...prev]);
      setActiveSessionId(session.id);
      return session;
    },
    [activeSessionId],
  );

  const sendMessage = useCallback(
    async (text: string, modelKey?: string, thinkingLevel?: string) => {
      if (!activeSessionId || !text.trim()) return;
      setLoading(true);

      const localUserMsg: LocalMessage = {
        id: `user-${Date.now()}`,
        role: 'user',
        content: text.trim(),
      };
      setMessages((prev) => [...prev, localUserMsg]);

      try {
        console.log('[Chat] sendMessage', { sessionId: activeSessionId, modelKey, thinkingLevel });
        const response = await api.sendMessage(
          activeSessionId,
          text.trim(),
          modelKey ?? '',
          thinkingLevel ?? '',
        );
        // If we didn't get a streaming response via events, add the response directly.
        if (response) {
          setMessages((prev) => {
            // Avoid duplicates if streaming already added it
            const hasStreaming = prev.some(
              (m) => m.role === 'assistant' && m.streaming,
            );
            if (hasStreaming) {
              return prev.map((m) =>
                m.streaming ? { ...m, streaming: false } : m,
              );
            }
            return [
              ...prev,
              {
                id: `assistant-${Date.now()}`,
                role: 'assistant' as const,
                content: response,
              },
            ];
          });
        }
        setLoading(false);
        if (modelKey) recordSuccess(modelKey);
      } catch (e: any) {
        setLoading(false);
        if (modelKey) recordFailure(modelKey);
        const msg = e?.message || 'Erro ao enviar mensagem';
        showSnackbar(msg, 'error');
      }
      // O streaming finaliza via chat:turnEnd; a resposta também vem no retorno,
      // mas confiamos nos eventos para montar o texto progressivamente.
    },
    [activeSessionId, recordSuccess, recordFailure],
  );

  const deleteSession = useCallback(async (sessionId: string) => {
    await api.deleteSession(sessionId);
    setSessions((prev) => prev.filter((s) => s.id !== sessionId));
    if (activeSessionId === sessionId) {
      setActiveSessionId(null);
    }
  }, [activeSessionId]);

  const renameSession = useCallback(
    async (sessionId: string, newTitle: string) => {
      await api.renameSession(sessionId, newTitle);
      setSessions((prev) =>
        prev.map((s) => (s.id === sessionId ? { ...s, title: newTitle } : s)),
      );
    },
    [],
  );

  const togglePinSession = useCallback(async (sessionId: string) => {
    await api.togglePinSession(sessionId);
    setSessions((prev) =>
      prev.map((s) => (s.id === sessionId ? { ...s, pinned: !s.pinned } : s)),
    );
  }, []);

  const value: ChatContextValue = {
    sessions,
    activeSessionId,
    activeSessionPath,
    activeSessionWorker,
    messages,
    loading,
    loadSessions,
    createSession,
    selectSession,
    sendMessage,
    deleteSession,
    renameSession,
    togglePinSession,
  };

  return <ChatContext.Provider value={value}>{children}</ChatContext.Provider>;
}

export function useChat() {
  const ctx = useContext(ChatContext);
  if (!ctx) {
    throw new Error('useChat must be used inside ChatProvider');
  }
  return ctx;
}
