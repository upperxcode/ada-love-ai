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
  elapsed?: number;
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
  activeSession: api.backend.ChatSession | null;
  activeSessionPath: string | null;
  activeSessionWorker: string | null;
  activeWorkspacePath: string | null;
  loadSessions: (workspacePath: string) => Promise<void>;
  loadAllSessions: (workspacePaths: string[]) => Promise<void>;
  createSession: (
    workspace: api.backend.WorkspaceConfig,
    worker: api.backend.WorkerConfig,
    summarized?: boolean,
  ) => Promise<api.backend.ChatSession | null>;
  selectSession: (sessionId: string | null) => void;
  sendMessage: (text: string, modelKey?: string, thinkingLevel?: string, mode?: string) => Promise<boolean>;
  answerQuestion: (answer: string) => Promise<void>;
  answerApproval: (approved: boolean, reason?: string) => Promise<void>;
  stopGeneration: () => Promise<void>;
  deleteSession: (sessionId: string) => Promise<void>;
  renameSession: (sessionId: string, newTitle: string) => Promise<api.backend.ChatSession | null>;
  togglePinSession: (sessionId: string) => Promise<void>;
  setSessionConfig: (model: string, provider: string, mode: string, thinking: string) => Promise<void>;
  deleteMessage: (messageId: string) => void;
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
  const deletedIdsRef = useRef(new Set<string>());
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
  const clearedSessionRef = useRef<Set<string>>(new Set());

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
          // Skip reload if this session was just cleared via /clear
          if (clearedSessionRef.current.has(sid)) {
            clearedSessionRef.current.delete(sid);
            console.log(`[Chat] turnEnd: skipping reload — session ${sid} was just cleared`);
            return;
          }
          console.log(`[Chat] turnEnd: reloading sessions for workspace="${sess.workspace_id}"`);
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

    unsubs.push(
      api.onChatEvent('chat:cleared', (payload: any) => {
        const { session_id } = payload || {};
        console.log(`[Chat] chat:cleared received — payload=`, payload, `activeSessionID=${activeSessionRef.current?.id}`);
        if (!session_id) return;
        if (activeSessionRef.current?.id !== session_id) return;
        console.log(`[Chat] chat:cleared — clearing messages for session ${session_id}`);
        setMessages([]);
        setLoading(false);
        setStage('');
        // Track that this session was just cleared so turnEnd won't reload from DB
        clearedSessionRef.current.add(session_id);
      }),
    );

    return () => {
      unsubs.forEach((u) => u());
    };
  }, []);

  const loadSessions = useCallback(async (workspacePath: string) => {
    console.log(`[Chat] loadSessions("${workspacePath}") — fetching...`);
    const list = await api.getSessions(workspacePath);
    console.log(`[Chat] loadSessions("${workspacePath}") — got ${list.length} sessions`);
    for (const s of list) {
      console.log(`[Chat]   session="${s.id}" title="${s.title}" worker="${s.worker_name}" messages=${s.messages.length}`);
    }
    // Merge: remove old sessions for THIS workspace, keep others, add fresh ones
    setSessions((prev) => {
      const otherSessions = prev.filter((s) => s.workspace_id !== workspacePath);
      return [...otherSessions, ...list];
    });
    // Only change active session if the current one no longer exists
    setActiveSessionId((current) => {
      if (!current) return list[0]?.id ?? null;
      return list.some((s) => s.id === current) ? current : null;
    });
  }, []);

  // Load sessions for ALL workspaces at once
  const loadAllSessions = useCallback(async (workspacePaths: string[]) => {
    console.log(`[Chat] loadAllSessions: loading ${workspacePaths.length} workspaces...`);
    const results = await Promise.all(
      workspacePaths.map(async (path) => {
        const list = await api.getSessions(path);
        console.log(`[Chat] loadAllSessions: "${path}" → ${list.length} sessions`);
        return list;
      }),
    );
    const allSessions = results.flat();
    console.log(`[Chat] loadAllSessions: total ${allSessions.length} sessions across all workspaces`);
    for (const s of allSessions) {
      console.log(`[Chat]   session="${s.id}" workspace="${s.workspace_id}" title="${s.title}" worker="${s.worker_name}" messages=${s.messages.length}`);
    }
    setSessions(allSessions);
  }, []);
  const loadSessionsRef = useRef(loadSessions);
  loadSessionsRef.current = loadSessions;

  const selectSession = useCallback((sessionId: string | null) => {
    console.log(`[Chat] selectSession("${sessionId}")`);
    setActiveSessionId(sessionId);
  }, []);

  // Sync messages when active session changes or sessions are reloaded from backend
  useEffect(() => {
    const newMessages = mapMessages(activeSession);
    setMessages((prev) => {
      return newMessages
        .filter((m) => !deletedIdsRef.current.has(m.id));
    });
  }, [activeSessionId, activeSession]);

  const createSession = useCallback(
    async (
      workspace: api.backend.WorkspaceConfig,
      worker: api.backend.WorkerConfig,
      summarized = false,
    ): Promise<api.backend.ChatSession | null> => {
      let raw: api.backend.ChatSession | null = null;
      console.log(`[Chat] createSession("${workspace.path}", "${worker.name}", summarized=${summarized})`);
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
      console.log(`[Chat] createSession: created session="${session.id}" title="${session.title}" workspace="${session.workspace_id}" worker="${session.worker_name}"`);
      setSessions((prev) => [session, ...prev]);
      setActiveSessionId(session.id);
      return session;
    },
    [activeSessionId],
  );

  const sendMessage = useCallback(
    async (text: string, modelKey?: string, thinkingLevel?: string, mode?: string): Promise<boolean> => {
      if (!activeSessionId || !text.trim()) return false;
      setLoading(true);
      setStage('');

      // Add user message to chat immediately (before calling backend so streaming works)
      const localUserMsg: LocalMessage = {
        id: `user-${Date.now()}`,
        role: 'user',
        content: text.trim(),
      };
      setMessages((prev) => [...prev, localUserMsg]);

      try {
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
            const hasStreaming = prev.some((m) => m.role === 'assistant' && m.streaming);
            if (hasStreaming) {
              return prev.map((m) => (m.streaming ? { ...m, streaming: false } : m));
            }
            return [...prev, { id: `assistant-${Date.now()}`, role: 'assistant' as const, content: response }];
          });
        }
        setLoading(false);
        if (modelKey) recordSuccess(modelKey);
        return true;
      } catch (e: any) {
        setLoading(false);
        if (modelKey) recordFailure(modelKey);
        showSnackbar(e?.message || 'Erro ao enviar mensagem', 'error');
        return false;
      }
    },
    [activeSessionId, recordSuccess, recordFailure, showSnackbar],
  );

  const deleteSession = useCallback(async (sessionId: string) => {
    await api.deleteSession(sessionId);
    setSessions((prev) => prev.filter((s) => s.id !== sessionId));
    if (activeSessionId === sessionId) {
      setActiveSessionId(null);
    }
  }, [activeSessionId]);

  const renameSession = useCallback(async (sessionId: string, newTitle: string): Promise<api.backend.ChatSession | null> => {
      try {
        const updatedSession = await api.renameSession(sessionId, newTitle);
        if (updatedSession) {
          // Use the title from the backend (may have been modified for uniqueness)
          setSessions((prev) =>
            prev.map((s) => (s.id === sessionId ? { ...s, title: updatedSession?.title ?? s.title } : s)),
          );
        }
        return updatedSession;
      } catch (error) {
        console.error('[renameSession] Error:', error);
        return null;
      }
    }, []);

  const togglePinSession = useCallback(async (sessionId: string) => {
    await api.togglePinSession(sessionId);
    setSessions((prev) =>
      prev.map((s) => (s.id === sessionId ? { ...s, pinned: !s.pinned } : s)),
    );
  }, []);

  const setSessionConfig = useCallback(async (model: string, provider: string, mode: string, thinking: string) => {
    if (!activeSessionId) return;
    console.log(`[Chat] setSessionConfig: session=${activeSessionId} model=${model} mode=${mode} thinking=${thinking}`);
    await api.setSessionConfig(activeSessionId, model, provider, mode, thinking);
    setSessions((prev) =>
      prev.map((s) =>
        s.id === activeSessionId ? { ...s, model, provider, mode, thinking } : s,
      ),
    );
  }, [activeSessionId]);

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

  const deleteMessage = useCallback((messageId: string) => {
    deletedIdsRef.current.add(messageId);
    setMessages((prev) => prev.filter((m) => m.id !== messageId));
  }, []);

  const value: ChatContextValue = {
    sessions,
    activeSessionId,
    activeSession,
    activeSessionPath,
    activeSessionWorker,
    activeWorkspacePath,
    messages,
    loading,
    stage,
    pendingQuestion,
    pendingApproval,
    loadSessions,
    loadAllSessions,
    createSession,
    selectSession,
    sendMessage,
    answerQuestion,
    answerApproval,
    stopGeneration,
    deleteSession,
    renameSession,
    togglePinSession,
    setSessionConfig,
    deleteMessage,
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
