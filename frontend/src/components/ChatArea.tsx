import { useEffect, useRef } from 'react';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import { useChat, type LocalMessage } from './ChatContext';
import { Icon } from './Icon';
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

export function ChatArea() {
  const { messages, loading, activeSessionId } = useChat();
  const scrollRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when messages change or loading finishes
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, loading]);

  return (
    <div ref={scrollRef} className="flex-1 min-h-0 overflow-y-auto px-6 py-3">
      {!activeSessionId ? (
        <div className="h-full flex flex-col items-center justify-center text-sm text-muted-foreground gap-2">
          <Icon name="MessageSquare" className="w-8 h-8 opacity-40" />
          <span>Select or create a chat to get started.</span>
        </div>
      ) : messages.length === 0 ? (
        <div className="h-full flex flex-col items-center justify-center text-sm text-muted-foreground gap-2">
          <Icon name="MessageSquare" className="w-8 h-8 opacity-40" />
          <span>No messages yet. Say hello!</span>
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
                elapsed={message.elapsed || 0}
                streaming={!!message.streaming}
              />
            ),
          )}
        </div>
      )}
    </div>
  );
}