import { useState, useEffect, useMemo } from 'react';
import { cn } from '@/lib/utils';
import { Icon } from './Icon';
import {
  Popover,
  PopoverTrigger,
  PopoverContent,
} from './ui/popover';
import {
  useModelHealth,
  getHealthColor,
} from '@/lib/modelHealth';

interface ModelPickerProps {
  selectedModel: string | null;
  onSelect: (modelKey: string) => void;
  disabled?: boolean;
}

interface ModelEntry {
  key: string;       // provider/model_name
  provider: string;
  modelName: string;
  displayName: string;
}

export function ModelPicker({ selectedModel, onSelect, disabled }: ModelPickerProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [models, setModels] = useState<ModelEntry[]>([]);
  const [selectedProvider, setSelectedProvider] = useState<string | null>(null);
  const { getScore } = useModelHealth();

  useEffect(() => {
    if (!open) return;
    (async () => {
      try {
        const config = await (window as any).go?.main?.App?.GetAdaConfig?.();
        if (!config) return;
        const list: any[] = config.model_list ?? [];
        const entries: ModelEntry[] = list
          .filter((m: any) => m.enabled ?? true)
          .map((m: any) => {
            const provider = m.provider ?? 'local';
            const modelName = m.model_name ?? m.model ?? '';
            const key = `${provider}/${modelName}`;
            return {
              key,
              provider,
              modelName,
              displayName: m.model_name ?? m.model ?? key,
            };
          })
          .filter((e: ModelEntry) => e.modelName);
        setModels(entries);
      } catch {}
    })();
  }, [open]);

  const providers = useMemo(() => {
    const set = new Set(models.map((m) => m.provider));
    return Array.from(set).sort();
  }, [models]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return models.filter((m) => {
      if (selectedProvider && m.provider !== selectedProvider) return false;
      if (q && !m.displayName.toLowerCase().includes(q) && !m.provider.toLowerCase().includes(q)) return false;
      return true;
    });
  }, [models, query, selectedProvider]);

  const displayLabel = selectedModel
    ? models.find((m) => m.key === selectedModel)?.displayName ?? selectedModel
    : 'Modelo';

  const displayProvider = selectedModel
    ? models.find((m) => m.key === selectedModel)?.provider ?? ''
    : '';

  return (
    <Popover open={open} onOpenChange={(o) => { setOpen(o); if (!o) { setQuery(''); setSelectedProvider(null); } }}>
      <PopoverTrigger asChild>
        <button
          type="button"
          disabled={disabled}
          className={cn(
            'flex items-center gap-1.5 px-2 py-1 rounded text-[11px] transition-colors',
            'text-muted-foreground hover:text-foreground hover:bg-muted',
            disabled && 'opacity-40 pointer-events-none',
          )}
          title="Selecionar modelo"
        >
          <Icon name="Cpu" className="w-3.5 h-3.5 shrink-0" />
          <span className="truncate max-w-[140px]">
            {selectedModel ? (
              <>
                <span style={{ color: getHealthColor(getScore(selectedModel)) }}>●</span>{' '}
                {displayProvider}/{displayLabel}
              </>
            ) : (
              'Modelo'
            )}
          </span>
          <Icon name="ChevronDown" className="w-3 h-3 shrink-0 opacity-50" />
        </button>
      </PopoverTrigger>

      <PopoverContent className="w-72 p-0" align="start" side="top">
        {/* Search */}
        <div className="p-2 border-b border-border">
          <div className="relative">
            <Icon
              name="Search"
              className="w-3 h-3 absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
            />
            <input
              autoFocus
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Filtrar modelos…"
              className="w-full pl-7 pr-2 py-1.5 rounded bg-background border border-border text-[11px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
            />
          </div>
        </div>

        {/* Provider filter chips */}
        {providers.length > 1 && (
          <div className="flex flex-wrap gap-1 p-2 border-b border-border">
            <button
              type="button"
              onClick={() => setSelectedProvider(null)}
              className={cn(
                'px-2 py-0.5 rounded text-[10px] transition-colors',
                selectedProvider === null
                  ? 'bg-accent text-accent-foreground'
                  : 'bg-muted text-muted-foreground hover:text-foreground',
              )}
            >
              Todos
            </button>
            {providers.map((p) => (
              <button
                key={p}
                type="button"
                onClick={() => setSelectedProvider(selectedProvider === p ? null : p)}
                className={cn(
                  'px-2 py-0.5 rounded text-[10px] transition-colors',
                  selectedProvider === p
                    ? 'bg-accent text-accent-foreground'
                    : 'bg-muted text-muted-foreground hover:text-foreground',
                )}
              >
                {p}
              </button>
            ))}
          </div>
        )}

        {/* Model list */}
        <div className="max-h-56 overflow-y-auto">
          {filtered.length === 0 ? (
            <div className="px-3 py-4 text-center text-[11px] text-muted-foreground">
              Nenhum modelo encontrado.
            </div>
          ) : (
            filtered.map((model) => {
              const score = getScore(model.key);
              const active = selectedModel === model.key;
              return (
                <button
                  key={model.key}
                  type="button"
                  onClick={() => {
                    onSelect(model.key);
                    setOpen(false);
                  }}
                  className={cn(
                    'w-full flex items-center gap-2 px-3 py-1.5 text-left text-[11px] transition-colors border-b border-border/30 last:border-b-0',
                    active
                      ? 'bg-accent text-accent-foreground'
                      : 'text-foreground/80 hover:bg-muted',
                  )}
                >
                  <span
                    className="w-2 h-2 rounded-full shrink-0"
                    style={{ backgroundColor: getHealthColor(score) }}
                  />
                  <span className="truncate">
                    <span className="text-muted-foreground">{model.provider}/</span>
                    {model.displayName}
                  </span>
                  {active && (
                    <Icon name="Check" className="w-3 h-3 shrink-0 ml-auto" />
                  )}
                </button>
              );
            })
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}
