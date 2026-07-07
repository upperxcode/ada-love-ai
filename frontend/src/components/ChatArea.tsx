import { useChat, type LocalMessage } from './ChatContext';
import { cn } from '@/lib/utils';
import { Icon } from './Icon';

function MessageBubble({ message }: { message: LocalMessage }) {
  const isUser = message.role === 'user';
  return (
    <div
      className={cn(
        'flex w-full',
        isUser ? 'justify-end' : 'justify-start',
      )}
    >
      <div
        className={cn(
          'max-w-[80%] rounded-lg px-3 py-2 text-sm leading-relaxed',
          isUser
            ? 'bg-primary text-primary-foreground'
            : 'bg-muted text-foreground',
          message.error && 'bg-destructive text-destructive-foreground',
        )}
      >
        <div className="whitespace-pre-wrap">{message.content}</div>
        {message.streaming && (
          <span className="inline-block w-1.5 h-1.5 ml-1 rounded-full bg-current animate-pulse" />
        )}
      </div>
    </div>
  );
}

export function ChatArea() {
  const { messages, loading, activeSessionId } = useChat();

  return (
    <div className="flex-1 min-h-0 overflow-y-auto p-4">
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
        <div className="flex flex-col gap-3 max-w-3xl mx-auto">
          {messages.map((message) => (
            <MessageBubble key={message.id} message={message} />
          ))}
          {loading && messages[messages.length - 1]?.role === 'user' && (
            <div className="flex justify-start">
              <div className="bg-muted text-foreground rounded-lg px-3 py-2 text-sm">
                Pensando
                <span className="inline-block w-1 h-1 ml-1 rounded-full bg-current animate-pulse" />
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
