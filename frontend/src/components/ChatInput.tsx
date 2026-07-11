import { useState, useRef, useEffect, useCallback } from 'react';
import { useChat } from './ChatContext';
import { useSnackbar } from './Snackbar';
import { Button } from './ui/button';
import { Icon } from './Icon';
import { ModelPicker } from './ModelPicker';
import { SlashCommandMenu } from './SlashCommandMenu';
import { Dialog, DialogContent } from './ui/dialog';
import { cn } from '@/lib/utils';
import * as api from '../api';

const CHAT_MODES = [
  { id: 'ask', label: 'Ask', icon: 'Search', description: 'Asks questions without modifying the project' },
  { id: 'plan', label: 'Plan', icon: 'Layers', description: 'Plans changes without executing' },
  { id: 'auto', label: 'Auto', icon: 'Zap', description: 'Executes changes with confirmation' },
  { id: 'full', label: 'Full Access', icon: 'Eye', description: 'Executes everything without confirmation' },
] as const;

type ChatMode = (typeof CHAT_MODES)[number]['id'];

export function ChatInput() {
  const { sendMessage, loading, activeSessionId, activeSession, setSessionConfig, pendingApproval } = useChat();
  const { showSnackbar } = useSnackbar();
  const [draft, setDraft] = useState('');
  const [mode, setMode] = useState<ChatMode>('ask');
  const [expanded, setExpanded] = useState(false);
  const [thinking, setThinking] = useState(false);
  const [selectedModel, setSelectedModel] = useState<string | null>(null);
  const [initialized, setInitialized] = useState(false);
  const [commands, setCommands] = useState<api.backend.CommandInfo[]>([]);
  const [slashMenuVisible, setSlashMenuVisible] = useState(false);
  const [slashQuery, setSlashQuery] = useState('');
  const [executingCommand, setExecutingCommand] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Restore config from active session when switching chats
  useEffect(() => {
    if (activeSession) {
      console.log(`[ChatInput] Restoring config: model=${activeSession.model} mode=${activeSession.mode} thinking=${activeSession.thinking}`);
      setMode((activeSession.mode as ChatMode) || 'ask');
      setThinking(activeSession.thinking === 'high');
      setSelectedModel(activeSession.model || null);
      setInitialized(true);
    } else {
      setInitialized(true);
    }
  }, [activeSession?.id]);

  // Load slash commands on mount
  useEffect(() => {
    api.listCommands().then(setCommands);
  }, []);

  // Detect slash prefix and manage menu visibility
  const handleDraftChange = useCallback(
    (value: string) => {
      setDraft(value);
      const trimmed = value.trimStart();
      if (trimmed.startsWith('/') && !value.includes('\n')) {
        const query = trimmed.slice(1);
        setSlashQuery(query);
        setSlashMenuVisible(true);
      } else {
        setSlashMenuVisible(false);
        setSlashQuery('');
      }
    },
    [],
  );

  // Execute a slash command and show result in the chat
  const executeSlashCommand = useCallback(
    async (commandText: string) => {
      if (!activeSessionId || executingCommand) return;
      setExecutingCommand(true);
      setSlashMenuVisible(false);
      setSlashQuery('');
      setDraft('');

      try {
        await sendMessage(commandText, selectedModel ?? undefined, thinking ? 'high' : undefined, mode);
      } catch (e: any) {
        showSnackbar(e?.message || 'Erro ao executar comando', 'error');
      } finally {
        setExecutingCommand(false);
      }
    },
    [activeSessionId, executingCommand, mode, sendMessage, selectedModel, thinking, showSnackbar],
  );

  // Save config changes to backend
  const saveConfig = (newModel: string | null, newMode: string, newThinking: boolean) => {
    if (!activeSessionId) return;
    setSessionConfig(
      newModel ?? '',
      '',
      newMode,
      newThinking ? 'high' : '',
    );
  };

  const handleModeChange = (newMode: ChatMode) => {
    setMode(newMode);
    saveConfig(selectedModel, newMode, thinking);
  };

  const handleThinkingChange = () => {
    const newThinking = !thinking;
    setThinking(newThinking);
    saveConfig(selectedModel, mode, newThinking);
  };

  const handleModelChange = (newModel: string | null) => {
    setSelectedModel(newModel);
    saveConfig(newModel, mode, thinking);
  };

  const handleSubmit = async () => {
    if (!draft.trim() || loading || !activeSessionId) return;
    const text = draft.trim();
    const success = await sendMessage(text, selectedModel ?? undefined, thinking ? 'high' : undefined, mode);
    // Only clear input on success
    if (success) {
      setDraft('');
      if (expanded) setExpanded(false);
    }
  };

  const handleKeyDownInline = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (slashMenuVisible) {
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
        // Let the menu handle navigation — dispatch to the menu div
        const menuEl = document.querySelector('[data-slash-menu]');
        if (menuEl) {
          menuEl.dispatchEvent(new KeyboardEvent('keydown', { key: e.key, bubbles: true }));
        }
        e.preventDefault();
        return;
      }
      if (e.key === 'Enter') {
        e.preventDefault();
        const menuEl = document.querySelector('[data-slash-menu]');
        if (menuEl) {
          menuEl.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
        }
        return;
      }
      if (e.key === 'Escape') {
        e.preventDefault();
        setSlashMenuVisible(false);
        setSlashQuery('');
        return;
      }
    }
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  const handleKeyDownExpanded = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      handleSubmit();
    }
  };

  const currentMode = CHAT_MODES.find((m) => m.id === mode) ?? CHAT_MODES[0];

  return (
    <>
      <div className="shrink-0 border-t border-border bg-card px-3 pt-2 pb-2">
        {/* Textarea row — expand button sits left of textarea on the first line */}
        <div className="relative">
          <SlashCommandMenu
            visible={slashMenuVisible}
            query={slashQuery}
            commands={commands}
            onSelect={executeSlashCommand}
            onClose={() => {
              setSlashMenuVisible(false);
              setSlashQuery('');
            }}
          />
          <div className="flex items-start gap-1 mx-auto">
            <button
              type="button"
              className="toolbar-btn mt-1 shrink-0"
              title="Expand editor"
              onClick={() => setExpanded(true)}
            >
              <Icon name="Maximize2" className="w-3.5 h-3.5" />
            </button>

            <textarea
              ref={textareaRef}
              value={draft}
              onChange={(e) => handleDraftChange(e.target.value)}
              onKeyDown={handleKeyDownInline}
              placeholder={
                activeSessionId
                  ? 'Type your message...'
                  : 'Select a chat to start'
              }
              disabled={!activeSessionId || loading}
              rows={2}
              className="flex-1 min-h-[52px] max-h-[52px] resize-none bg-transparent px-2 py-1 text-[13px] leading-relaxed text-foreground placeholder:text-muted-foreground/50 focus:outline-none disabled:opacity-40"
              style={{ overflow: 'hidden' }}
            />
          </div>
        </div>

        {/* Bottom toolbar */}
        <div className="flex items-center gap-1 mx-auto mt-0.5">
          {/* Plus */}
          <button type="button" className="toolbar-btn" title="Attach file">
            <Icon name="Plus" className="w-4 h-4" />
          </button>

          {/* Mode selector */}
          <div className="flex items-center bg-muted rounded-md px-0.5 py-0.5 gap-0.5">
            {CHAT_MODES.map((m) => (
              <button
                key={m.id}
                type="button"
                onClick={() => handleModeChange(m.id)}
                className={cn(
                  'flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-medium transition-colors',
                  mode === m.id
                    ? m.id === 'full'
                      ? 'bg-red-500/20 text-red-500 shadow-sm'
                      : 'bg-card text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground',
                )}
                title={m.description}
              >
                <Icon name={m.icon} className="w-3 h-3" />
                <span className="hidden sm:inline">{m.label}</span>
              </button>
            ))}
          </div>

          {/* Thinking toggle */}
          <button
            type="button"
            onClick={handleThinkingChange}
            className={cn(
              'flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-medium transition-colors',
              thinking
                ? 'bg-accent text-accent-foreground'
                : 'text-muted-foreground hover:text-foreground',
            )}
            title={thinking ? 'Thinking: ligado' : 'Thinking: desligado'}
          >
            <Icon name="Brain" className="w-3 h-3" />
            <span className="hidden sm:inline">Think</span>
          </button>

          <div className="flex-1" />

          {/* Model picker */}
          <ModelPicker
            selectedModel={selectedModel}
            onSelect={handleModelChange}
            disabled={!activeSessionId}
          />

          {/* Send */}
          <Button
            size="icon"
            className={cn(
              'h-7 w-7 shrink-0 rounded-md transition-colors',
              loading && 'bg-red-500/20 text-red-400 hover:bg-red-500/30',
            )}
            disabled={!activeSessionId || loading || !draft.trim()}
            onClick={handleSubmit}
            title="Enviar"
          >
            <Icon name="Zap" className="w-3.5 h-3.5" />
          </Button>
        </div>
      </div>

      {/* Expanded editor dialog */}
      <Dialog open={expanded} onOpenChange={setExpanded}>
        <DialogContent hideClose className="max-w-4xl h-[70vh] p-0 gap-0 flex flex-col">
          <div className="flex items-center justify-between px-4 py-2 border-b border-border shrink-0">
            <div className="flex items-center gap-2">
              <Icon name={currentMode.icon} className="w-4 h-4 text-muted-foreground" />
              <span className="text-sm font-medium text-foreground">{currentMode.label}</span>
              <span className="text-[10px] text-muted-foreground">— {currentMode.description}</span>
            </div>
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0 text-muted-foreground"
                onClick={() => setExpanded(false)}
              >
                <Icon name="Minimize2" className="w-3.5 h-3.5" />
              </Button>
            </div>
          </div>

          <textarea
            value={draft}
            onChange={(e) => handleDraftChange(e.target.value)}
            onKeyDown={handleKeyDownExpanded}
            placeholder="Write your message..."
            disabled={!activeSessionId || loading}
            autoFocus
            className="flex-1 w-full min-h-0 resize-none bg-transparent px-4 py-3 text-sm leading-relaxed text-foreground placeholder:text-muted-foreground/50 focus:outline-none disabled:opacity-40"
          />

          <div className="flex items-center justify-between px-4 py-2 border-t border-border shrink-0">
            <div className="flex items-center gap-1">
              {CHAT_MODES.map((m) => (
                <button
                  key={m.id}
                  type="button"
                  onClick={() => setMode(m.id)}
                  className={cn(
                    'flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-medium transition-colors',
                    mode === m.id
                      ? m.id === 'full'
                        ? 'bg-red-500/20 text-red-500'
                        : 'bg-accent text-accent-foreground'
                      : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  <Icon name={m.icon} className="w-3 h-3" />
                  <span>{m.label}</span>
                </button>
              ))}

              <button
                type="button"
                onClick={handleThinkingChange}
                className={cn(
                  'flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-medium transition-colors ml-1',
                  thinking
                    ? 'bg-accent text-accent-foreground'
                    : 'text-muted-foreground hover:text-foreground',
                )}
              >
                <Icon name="Brain" className="w-3 h-3" />
                <span>Think</span>
              </button>
            </div>

            <div className="flex items-center gap-2">
              <ModelPicker
                selectedModel={selectedModel}
                onSelect={handleModelChange}
                disabled={!activeSessionId}
              />
              <Button
                size="icon"
                className={cn(
                  'h-7 w-7 shrink-0 rounded-md transition-colors',
                  loading && 'bg-red-500/20 text-red-400 hover:bg-red-500/30',
                )}
                disabled={!activeSessionId || loading || !draft.trim()}
                onClick={handleSubmit}
                title="Enviar (Ctrl+Enter)"
              >
                <Icon name="Zap" className="w-3.5 h-3.5" />
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
