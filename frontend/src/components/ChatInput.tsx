import { useState, useRef, useEffect } from 'react';
import { useChat } from './ChatContext';
import { Button } from './ui/button';
import { Icon } from './Icon';
import { ModelPicker } from './ModelPicker';
import { Dialog, DialogContent } from './ui/dialog';
import { cn } from '@/lib/utils';

const CHAT_MODES = [
  { id: 'ask', label: 'Ask', icon: 'Search', description: 'Faz perguntas sem modificar o projeto' },
  { id: 'plan', label: 'Plan', icon: 'Layers', description: 'Planeja mudanças sem executar' },
  { id: 'auto', label: 'Auto', icon: 'Zap', description: 'Executa mudanças com confirmação' },
  { id: 'full', label: 'Full Access', icon: 'Eye', description: 'Executa tudo sem confirmação' },
] as const;

type ChatMode = (typeof CHAT_MODES)[number]['id'];

export function ChatInput() {
  const { sendMessage, loading, activeSessionId } = useChat();
  const [draft, setDraft] = useState('');
  const [mode, setMode] = useState<ChatMode>('ask');
  const [expanded, setExpanded] = useState(false);
  const [thinking, setThinking] = useState(false);
  const [selectedModel, setSelectedModel] = useState<string | null>(() => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('ada:selectedModel');
  });
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (selectedModel) {
      localStorage.setItem('ada:selectedModel', selectedModel);
    } else {
      localStorage.removeItem('ada:selectedModel');
    }
  }, [selectedModel]);

  useEffect(() => {
    localStorage.setItem('ada:thinking', thinking ? '1' : '0');
  }, [thinking]);

  const handleSubmit = () => {
    if (!draft.trim() || loading || !activeSessionId) return;
    console.log('[ChatInput] handleSubmit', { selectedModel, thinking, mode, draft: draft.trim().substring(0, 50) });
    sendMessage(draft.trim(), selectedModel ?? undefined, thinking ? 'high' : undefined, mode);
    setDraft('');
    if (expanded) setExpanded(false);
  };

  const handleKeyDownInline = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
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
        <div className="flex items-start gap-1 mx-auto">
          <button
            type="button"
            className="toolbar-btn mt-1 shrink-0"
            title="Expandir editor"
            onClick={() => setExpanded(true)}
          >
            <Icon name="Maximize2" className="w-3.5 h-3.5" />
          </button>

          <textarea
            ref={textareaRef}
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={handleKeyDownInline}
            placeholder={
              activeSessionId
                ? 'Digite sua mensagem…'
                : 'Selecione um chat para digitar'
            }
            disabled={!activeSessionId || loading}
            rows={2}
            className="flex-1 min-h-[52px] max-h-[52px] resize-none bg-transparent px-2 py-1 text-[13px] leading-relaxed text-foreground placeholder:text-muted-foreground/50 focus:outline-none disabled:opacity-40"
            style={{ overflow: 'hidden' }}
          />
        </div>

        {/* Bottom toolbar */}
        <div className="flex items-center gap-1 mx-auto mt-0.5">
          {/* Plus */}
          <button type="button" className="toolbar-btn" title="Anexar arquivo">
            <Icon name="Plus" className="w-4 h-4" />
          </button>

          {/* Mode selector */}
          <div className="flex items-center bg-muted rounded-md px-0.5 py-0.5 gap-0.5">
            {CHAT_MODES.map((m) => (
              <button
                key={m.id}
                type="button"
                onClick={() => setMode(m.id)}
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
            onClick={() => setThinking(!thinking)}
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
            onSelect={setSelectedModel}
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
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={handleKeyDownExpanded}
            placeholder="Escreva sua mensagem…"
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
                onClick={() => setThinking(!thinking)}
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
                onSelect={setSelectedModel}
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
