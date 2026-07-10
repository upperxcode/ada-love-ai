import { useState, useEffect, useMemo, useRef } from 'react';
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
  const [selectedProvider, setSelectedProvider] = useState<string>(
    selectedModel ? selectedModel.split('/')[0] : '',
  );
  const [providerOpen, setProviderOpen] = useState(false);
  const providerRef = useRef<HTMLDivElement>(null);
  const { getScore } = useModelHealth();

  useEffect(() => {
    if (!open) return;
    (async () => {
      try {
        const config = await (window as any).go?.main?.App?.GetAdaConfig?.();
        if (!config) return;
        const entries: ModelEntry[] = [];

        // 1. Models from model_list
        const list: any[] = config.model_list ?? [];
        for (const m of list) {
          if (m.enabled === false) continue;
          const provider = m.provider ?? 'local';
          const modelName = m.model_name ?? m.model ?? '';
          if (!modelName) continue;
          const key = `${provider}/${modelName}`;
          entries.push({ key, provider, modelName, displayName: modelName });
        }

        // 2. Models from providers config
        const providers: Record<string, any> = config.providers ?? {};
        for (const [providerName, providerCfg] of Object.entries(providers)) {
          const providerModels = providerCfg.models ?? {};
          for (const modelName of Object.keys(providerModels)) {
            const key = `${providerName}/${modelName}`;
            // Avoid duplicates with model_list
            if (!entries.some((e) => e.key === key)) {
              entries.push({ key, provider: providerName, modelName, displayName: modelName });
            }
          }
        }

        setModels(entries);

        // Derive provider from current selection if not already set
        const providerFromSelection = selectedModel ? selectedModel.split('/')[0] : '';
        if (providerFromSelection) {
          setSelectedProvider(providerFromSelection);
        } else if (entries.length > 0) {
          const providersSet = Array.from(new Set(entries.map((e) => e.provider))).sort();
          setSelectedProvider(providersSet[0]);
        }
      } catch {}
    })();
  }, [open]);

  const providers = useMemo(() => {
    const set = new Set(models.map((m) => m.provider));
    return Array.from(set).sort();
  }, [models]);

  // Close provider dropdown on outside click
  useEffect(() => {
    if (!providerOpen) return;
    const handler = (e: MouseEvent) => {
      if (providerRef.current && !providerRef.current.contains(e.target as Node)) {
        setProviderOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [providerOpen]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return models.filter((m) => {
      if (selectedProvider && m.provider !== selectedProvider) return false;
      if (q && !m.displayName.toLowerCase().includes(q) && !m.provider.toLowerCase().includes(q)) return false;
      return true;
    });
  }, [models, query, selectedProvider]);

  const listRef = useRef<HTMLDivElement>(null);

  // Scroll to the active model when list loads or provider changes
  useEffect(() => {
    if (!open || !selectedModel || !listRef.current || filtered.length === 0) return;
    const activeIndex = filtered.findIndex((m) => m.key === selectedModel);
    if (activeIndex < 0) return;
    const el = listRef.current.children[activeIndex] as HTMLElement;
    if (el) el.scrollIntoView({ block: 'nearest' });
  }, [open, selectedModel, filtered]);

  const displayLabel = selectedModel
    ? models.find((m) => m.key === selectedModel)?.displayName ?? selectedModel
    : 'Modelo';

  const displayProvider = selectedModel
    ? models.find((m) => m.key === selectedModel)?.provider ?? ''
    : '';

  return (
    <Popover open={open} onOpenChange={(o) => { setOpen(o); if (!o) { setQuery(''); } }}>
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

      <PopoverContent className="w-80 p-0" align="start" side="top">
        {/* Provider select (custom dropdown to avoid Portal conflict with Popover) */}
        <div className="p-2 border-b border-border" ref={providerRef}>
          <div className="relative">
            <button
              type="button"
              onClick={() => setProviderOpen(!providerOpen)}
              className="w-full flex items-center justify-between h-8 px-2.5 rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground text-xs font-medium"
            >
              <span>{selectedProvider || 'Select a provider'}</span>
              <Icon name="ChevronDown" className={cn('w-3 h-3 transition-transform', providerOpen && 'rotate-180')} />
            </button>
            {providerOpen && (
              <div className="absolute top-full left-0 right-0 z-10 mt-1 rounded-md border border-border bg-background shadow-md max-h-40 overflow-y-auto">
                {providers.map((p) => (
                  <button
                    key={p}
                    type="button"
                    className={cn(
                      'w-full flex items-center px-2.5 py-1.5 text-xs text-left transition-colors',
                      selectedProvider === p
                        ? 'bg-accent text-accent-foreground'
                        : 'text-foreground/80 hover:bg-muted',
                    )}
                    onClick={() => {
                      setSelectedProvider(p);
                      setQuery('');
                      setProviderOpen(false);
                    }}
                  >
                    {p}
                    {selectedProvider === p && (
                      <Icon name="Check" className="w-3 h-3 ml-auto shrink-0" />
                    )}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Search */}
        <div className="px-2 pb-2 border-b border-border">
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

        {/* Model list */}
        <div className="max-h-56 overflow-y-auto" ref={listRef}>
          {filtered.length === 0 ? (
            <div className="px-3 py-4 text-center text-[11px] text-muted-foreground">
              Nenhum modelo encontrado para este provider.
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
                  <span className="truncate">{model.displayName}</span>
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