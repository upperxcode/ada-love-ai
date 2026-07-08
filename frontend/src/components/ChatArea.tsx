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

export function ChatArea() {
  const { messages, loading, activeSessionId } = useChat();
  const scrollRef = useRef<HTMLDivElement>(null);
  const [elapsedAtEnd, setElapsedAtEnd] = useState(0);

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
              <Timer isRunning={true} />
            </div>
          )}

          {/* Streaming: show loader under partial content */}
          {isStreamingAssistant && (
            <div className="flex items-center gap-3 py-1 ml-0">
              <TypingLoader />
            </div>
          )}
        </div>
      )}
    </div>
  );
}
