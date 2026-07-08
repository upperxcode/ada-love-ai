import { useState, useEffect, useRef } from 'react';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import { useChat, type LocalMessage } from './ChatContext';
import { cn } from '@/lib/utils';
import { Icon } from './Icon';
import { TypingLoader } from './TypingLoader';
import { CopyButton } from './CopyButton';

function formatTime(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}m ${s}s`;
}

function UserBubble({ message }: { message: LocalMessage }) {
  return (
    <div className="flex w-full justify-end">
      <div className="max-w-[70%] rounded-xl bg-primary text-primary-foreground px-4 py-2.5 text-[13px] leading-relaxed">
        {message.content}
      </div>
    </div>
  );
}

function AssistantBubble({
  message,
  elapsed,
  streaming,
}: {
  message: LocalMessage;
  elapsed: number;
  streaming: boolean;
}) {
  return (
    <div className="flex w-full justify-start group">
      <div className="w-full">
        {/* Markdown content */}
        <div className="prose prose-sm max-w-none text-[13px] leading-relaxed text-foreground prose-headings:text-foreground prose-p:text-foreground prose-li:text-foreground prose-strong:text-foreground prose-code:text-foreground prose-pre:bg-muted prose-pre:text-foreground prose-a:text-blue-400 prose-a:no-underline hover:prose-a:underline">
          <Markdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeHighlight]}>{message.content}</Markdown>
        </div>

        {/* Footer: timer + copy */}
        {!streaming && (
          <div className="flex items-center gap-1.5 mt-1.5">
            <Icon name="Zap" className="w-3 h-3 text-muted-foreground/40" />
            <span className="text-[10px] text-muted-foreground/50 tabular-nums">
              {formatTime(elapsed)}
            </span>
            <CopyButton text={message.content} />
          </div>
        )}
      </div>
    </div>
  );
}

function Timer({ isRunning }: { isRunning: boolean }) {
  const [seconds, setSeconds] = useState(0);
  const startRef = useRef(Date.now());

  useEffect(() => {
    if (isRunning) {
      startRef.current = Date.now();
      setSeconds(0);
      const interval = setInterval(() => {
        setSeconds(Math.floor((Date.now() - startRef.current) / 1000));
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [isRunning]);

  if (!isRunning) return null;

  return (
    <div className="flex items-center gap-1.5 ml-1">
      <Icon name="Zap" className="w-3 h-3 text-muted-foreground/40" />
      <span className="text-[10px] text-muted-foreground/50 tabular-nums">
        {formatTime(seconds)}
      </span>
    </div>
  );
}

function stageLabel(stage: string): string {
  if (!stage) return '';
  if (stage === 'thinking') return 'Thinking';
  if (stage.startsWith('tool:')) {
    const tool = stage.replace('tool:', '');
    return `Tool: ${tool}`;
  }
  if (stage.startsWith('subagent:')) {
    const label = stage.replace('subagent:', '');
    return `SubAgent${label ? `: ${label}` : ''}`;
  }
  if (stage === 'writing') return 'Writing';
  return stage;
}

function stageIcon(stage: string): string {
  if (stage === 'thinking') return 'Brain';
  if (stage.startsWith('tool:')) return 'Wrench';
  if (stage.startsWith('subagent:')) return 'GitBranch';
  if (stage === 'writing') return 'PenLine';
  return 'Loader';
}

export function ChatArea() {
  const { messages, loading, activeSessionId, stage, pendingQuestion, answerQuestion, pendingApproval, answerApproval, stopGeneration } = useChat();
  const scrollRef = useRef<HTMLDivElement>(null);
  const [elapsedAtEnd, setElapsedAtEnd] = useState(0);
  const [questionAnswer, setQuestionAnswer] = useState('');

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, loading]);

  // Track elapsed time during streaming
  const startTimeRef = useRef(Date.now());
  useEffect(() => {
    if (loading) {
      startTimeRef.current = Date.now();
    } else if (!loading && messages.length > 0) {
      setElapsedAtEnd(Math.floor((Date.now() - startTimeRef.current) / 1000));
    }
  }, [loading, messages.length]);

  const lastMsg = messages[messages.length - 1];
  const isStreamingAssistant = loading && lastMsg?.role === 'assistant' && lastMsg.streaming;

  return (
    <div ref={scrollRef} className="flex-1 min-h-0 overflow-y-auto px-6 py-3">
      {!activeSessionId ? (
        <div className="h-full flex flex-col items-center justify-center text-sm text-muted-foreground gap-2">
          <Icon name="MessageSquare" className="w-8 h-8 opacity-40" />
          <span>Selecione ou crie um chat para começar.</span>
        </div>
      ) : messages.length === 0 ? (
        <div className="h-full flex flex-col items-center justify-center text-sm text-muted-foreground gap-2">
          <Icon name="MessageSquare" className="w-8 h-8 opacity-40" />
          <span>Nenhuma mensagem ainda. Diga olá!</span>
        </div>
      ) : (
        <div className="flex flex-col gap-5">
          {messages.map((message) =>
            message.role === 'user' ? (
              <UserBubble key={message.id} message={message} />
            ) : (
              <AssistantBubble
                key={message.id}
                message={message}
                elapsed={elapsedAtEnd}
                streaming={!!message.streaming}
              />
            ),
          )}

          {/* Loading state */}
          {loading && lastMsg?.role === 'user' && (
            <div className="flex items-center gap-3 py-2">
              <TypingLoader />
              {stage && (
                <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                  <Icon name={stageIcon(stage)} className="w-3 h-3" />
                  <span>{stageLabel(stage)}</span>
                </div>
              )}
              <Timer isRunning={true} />
              <button
                type="button"
                className="ml-1 flex items-center gap-1 px-2 py-1 rounded-md text-[10px] text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                onClick={() => stopGeneration()}
                title="Parar geração"
              >
                <Icon name="Square" className="w-3 h-3" />
                <span>Stop</span>
              </button>
            </div>
          )}

          {/* Streaming: show loader under partial content */}
          {isStreamingAssistant && (
            <div className="flex items-center gap-3 py-1 ml-0">
              <TypingLoader />
              {stage && (
                <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                  <Icon name={stageIcon(stage)} className="w-3 h-3" />
                  <span>{stageLabel(stage)}</span>
                </div>
              )}
            </div>
          )}

          {/* Pending question from AI */}
          {pendingQuestion && (
            <div className="flex flex-col gap-2 py-2 px-3 rounded-lg border border-border bg-muted/50 max-w-[85%]">
              <div className="flex items-start gap-2">
                <Icon name="HelpCircle" className="w-4 h-4 text-primary mt-0.5 shrink-0" />
                <span className="text-[13px] text-foreground">{pendingQuestion.question}</span>
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  value={questionAnswer}
                  onChange={(e) => setQuestionAnswer(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && questionAnswer.trim()) {
                      answerQuestion(questionAnswer);
                      setQuestionAnswer('');
                    }
                  }}
                  autoFocus
                  placeholder="Digite sua resposta…"
                  className="flex-1 bg-transparent text-[13px] text-foreground placeholder:text-muted-foreground/50 focus:outline-none px-2 py-1.5 rounded border border-border"
                />
                <button
                  type="button"
                  className="shrink-0 px-3 py-1.5 rounded-md bg-primary text-primary-foreground text-[12px] font-medium hover:bg-primary/90 transition-colors disabled:opacity-40"
                  disabled={!questionAnswer.trim()}
                  onClick={() => {
                    answerQuestion(questionAnswer);
                    setQuestionAnswer('');
                  }}
                >
                  Responder
                </button>
              </div>
            </div>
          )}

          {/* Pending tool approval */}
          {pendingApproval && (
            <div className="flex flex-col gap-2 py-2 px-3 rounded-lg border border-yellow-500/30 bg-yellow-500/5 max-w-[85%]">
              <div className="flex items-start gap-2">
                <Icon name="ShieldAlert" className="w-4 h-4 text-yellow-500 mt-0.5 shrink-0" />
                <div className="flex flex-col gap-0.5">
                  <span className="text-[12px] font-medium text-foreground">
                    Permitir execução: <code className="text-[11px] text-primary">{pendingApproval.tool}</code>
                  </span>
                  {pendingApproval.args && (
                    <span className="text-[11px] text-muted-foreground font-mono break-all">
                      {pendingApproval.args}
                    </span>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-2 mt-1">
                <button
                  type="button"
                  className="px-3 py-1.5 rounded-md bg-green-600 text-white text-[12px] font-medium hover:bg-green-700 transition-colors"
                  onClick={() => answerApproval(true)}
                >
                  Aprovar
                </button>
                <button
                  type="button"
                  className="px-3 py-1.5 rounded-md bg-red-600 text-white text-[12px] font-medium hover:bg-red-700 transition-colors"
                  onClick={() => answerApproval(false, 'Denied by user')}
                >
                  Negar
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
