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
  pendingQuestion: { question: string } | null;
  pendingApproval: { id: string; tool: string; args: string } | null;
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
  sendMessage: (text: string, modelKey?: string, thinkingLevel?: string, mode?: string) => Promise<void>;
  answerQuestion: (answer: string) => Promise<void>;
  answerApproval: (approved: boolean, reason?: string) => Promise<void>;
  stopGeneration: () => Promise<void>;
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
  const [pendingQuestion, setPendingQuestion] = useState<{ question: string } | null>(null);
  const [pendingApproval, setPendingApproval] = useState<{ id: string; tool: string; args: string } | null>(null);
  const { recordSuccess, recordFailure } = useModelHealth();
  const { showSnackbar } = useSnackbar();
  const recordFailureRef = useRef(recordFailure);
  recordFailureRef.current = recordFailure;

  const activeSession = sessions.find((s) => s.id === activeSessionId) || null;
  const activeSessionRef = useRef(activeSession);
  activeSessionRef.current = activeSession;
  const activeSessionPath = activeSession?.workspace_id ?? null;
  const activeSessionWorker = activeSession?.worker_name ?? null;
  const activeWorkspacePath = activeSession?.workspace_id ?? null;

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
      api.onChatEvent('chat:status', (payload: any) => {
        const { session_id, stage: newStage } = payload || {};
        if (!session_id || !newStage) return;
        if (activeSessionRef.current?.id !== session_id) return;
        setStage(newStage);
      }),
    );

    unsubs.push(
      api.onChatEvent('chat:turnEnd', (payload: any) => {
        setLoading(false);
        setStage('');
        setMessages((prev) =>
          prev.map((m) => (m.streaming ? { ...m, streaming: false } : m)),
        );
        // Reload sessions from backend to sync with DB
        const sid = payload?.session_id;
        const sess = activeSessionRef.current;
        if (sid && sess && sess.id === sid && sess.workspace_id) {
          loadSessionsRef.current(sess.workspace_id);
        }
      }),
    );

    unsubs.push(
      api.onChatEvent('chat:error', (payload: any) => {
        setLoading(false);
        setStage('');
        const msg = payload?.message || 'Erro desconhecido';
        showSnackbar(msg, 'error');
      }),
    );

    unsubs.push(
      api.onChatEvent('chat:question', (payload: any) => {
        const { session_id, question } = payload || {};
        if (!session_id || !question) return;
        if (activeSessionRef.current?.id !== session_id) return;
        setPendingQuestion({ question });
      }),
    );

    unsubs.push(
      api.onChatEvent('chat:questionAnswered', () => {
        setPendingQuestion(null);
      }),
    );

    unsubs.push(
      api.onChatEvent('chat:toolApproval', (payload: any) => {
        const { id, session_id, tool, args } = payload || {};
        if (!id || !session_id || !tool) return;
        if (activeSessionRef.current?.id !== session_id) return;
        setPendingApproval({ id, tool, args: args || '' });
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
  const loadSessionsRef = useRef(loadSessions);
  loadSessionsRef.current = loadSessions;

  const selectSession = useCallback((sessionId: string | null) => {
    setActiveSessionId(sessionId);
  }, []);

  // Sync messages when active session changes or sessions are reloaded from backend
  useEffect(() => {
    setMessages(mapMessages(activeSession));
  }, [activeSessionId, activeSession]);

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
    async (text: string, modelKey?: string, thinkingLevel?: string, mode?: string) => {
      if (!activeSessionId || !text.trim()) return;
      setLoading(true);
      setStage('');
      setPendingQuestion(null);

      const localUserMsg: LocalMessage = {
        id: `user-${Date.now()}`,
        role: 'user',
        content: text.trim(),
      };
      setMessages((prev) => [...prev, localUserMsg]);

      try {
        console.log('[Chat] sendMessage', { sessionId: activeSessionId, modelKey, thinkingLevel, mode });
        const response = await api.sendMessage(
          activeSessionId,
          text.trim(),
          modelKey ?? '',
          thinkingLevel ?? '',
          mode ?? '',
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

  const answerQuestion = useCallback(async (answer: string) => {
    if (!activeSessionId || !answer.trim()) return;
    await api.answerQuestion(activeSessionId, answer.trim());
    setPendingQuestion(null);
  }, [activeSessionId]);

  const answerApproval = useCallback(async (approved: boolean, reason: string = '') => {
    if (!pendingApproval) return;
    await api.answerApproval(pendingApproval.id, approved, reason);
    setPendingApproval(null);
  }, [pendingApproval]);

  const stopGeneration = useCallback(async () => {
    if (!activeSessionId) return;
    await api.stopGeneration(activeSessionId);
    setLoading(false);
    setStage('');
    setPendingQuestion(null);
    setPendingApproval(null);
    setMessages((prev) =>
      prev.map((m) => (m.streaming ? { ...m, streaming: false } : m)),
    );
  }, [activeSessionId]);

  const value: ChatContextValue = {
    sessions,
    activeSessionId,
    activeSessionPath,
    activeSessionWorker,
    activeWorkspacePath,
    messages,
    loading,
    stage,
    pendingQuestion,
    pendingApproval,
    loadSessions,
    createSession,
    selectSession,
    sendMessage,
    answerQuestion,
    answerApproval,
    stopGeneration,
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
