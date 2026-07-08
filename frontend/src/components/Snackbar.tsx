import { createContext, useContext, useState, useCallback, useRef, useEffect } from 'react';
import { cn } from '@/lib/utils';
import { Icon } from './Icon';

interface SnackbarItem {
  id: number;
  message: string;
  type: 'error' | 'warning' | 'success' | 'info';
}

interface SnackbarContextValue {
  showSnackbar: (message: string, type?: SnackbarItem['type']) => void;
}

const SnackbarContext = createContext<SnackbarContextValue | null>(null);

export function useSnackbar() {
  const ctx = useContext(SnackbarContext);
  if (!ctx) throw new Error('useSnackbar must be used inside SnackbarProvider');
  return ctx;
}

const ICON_MAP: Record<SnackbarItem['type'], string> = {
  error: 'X',
  warning: 'Zap',
  success: 'Check',
  info: 'Zap',
};

const COLOR_MAP: Record<SnackbarItem['type'], string> = {
  error: 'border-red-500/50 bg-red-950/80 text-red-200',
  warning: 'border-yellow-500/50 bg-yellow-950/80 text-yellow-200',
  success: 'border-green-500/50 bg-green-950/80 text-green-200',
  info: 'border-border bg-card text-foreground',
};

function SnackbarItem({ item, onDismiss }: { item: SnackbarItem; onDismiss: () => void }) {
  const [visible, setVisible] = useState(false);
  const [exiting, setExiting] = useState(false);

  useEffect(() => {
    requestAnimationFrame(() => setVisible(true));
    const timer = setTimeout(() => {
      setExiting(true);
      setTimeout(onDismiss, 300);
    }, 5000);
    return () => clearTimeout(timer);
  }, [onDismiss]);

  return (
    <div
      className={cn(
        'flex items-center gap-2 px-3 py-2 rounded-lg border text-[12px] leading-snug shadow-lg max-w-sm transition-all duration-300',
        COLOR_MAP[item.type],
        visible && !exiting
          ? 'opacity-100 translate-y-0'
          : 'opacity-0 translate-y-2 pointer-events-none',
      )}
    >
      <Icon name={ICON_MAP[item.type]} className="w-3.5 h-3.5 shrink-0" />
      <span className="flex-1 truncate">{item.message}</span>
      <button
        type="button"
        onClick={() => {
          setExiting(true);
          setTimeout(onDismiss, 300);
        }}
        className="shrink-0 opacity-60 hover:opacity-100 transition-opacity"
      >
        <Icon name="X" className="w-3 h-3" />
      </button>
    </div>
  );
}

export function SnackbarProvider({ children }: { children: React.ReactNode }) {
  const [items, setItems] = useState<SnackbarItem[]>([]);
  const nextId = useRef(0);

  const showSnackbar = useCallback(
    (message: string, type: SnackbarItem['type'] = 'error') => {
      const id = nextId.current++;
      setItems((prev) => [...prev, { id, message, type }]);
    },
    [],
  );

  const dismiss = useCallback((id: number) => {
    setItems((prev) => prev.filter((item) => item.id !== id));
  }, []);

  return (
    <SnackbarContext.Provider value={{ showSnackbar }}>
      {children}
      {/* Snackbar container */}
      <div className="fixed bottom-24 right-4 z-[200] flex flex-col gap-2 pointer-events-none">
        <div className="pointer-events-auto">
          {items.map((item) => (
            <SnackbarItem key={item.id} item={item} onDismiss={() => dismiss(item.id)} />
          ))}
        </div>
      </div>
    </SnackbarContext.Provider>
  );
}
