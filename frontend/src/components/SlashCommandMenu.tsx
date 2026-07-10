import { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { cn } from '@/lib/utils';
import { Icon } from './Icon';
import type { backend } from '../api';

interface SlashCommandMenuProps {
  visible: boolean;
  query: string;
  commands: backend.CommandInfo[];
  onSelect: (command: string) => void;
  onClose: () => void;
}

interface FlatItem {
  type: 'command' | 'sub';
  command: backend.CommandInfo;
  sub?: backend.SubCommandInfo;
  label: string;
  description: string;
  usage: string;
}

function buildFlatList(commands: backend.CommandInfo[], query: string): FlatItem[] {
  const q = query.toLowerCase();
  const items: FlatItem[] = [];

  for (const cmd of commands) {
    const aliases = cmd.aliases ?? [];
    const subCommands = cmd.sub_commands ?? [];
    const cmdMatch =
      !q ||
      cmd.name.toLowerCase().includes(q) ||
      cmd.description.toLowerCase().includes(q) ||
      aliases.some((a) => a.toLowerCase().includes(q));

    if (cmdMatch && subCommands.length === 0) {
      items.push({
        type: 'command',
        command: cmd,
        label: `/${cmd.name}`,
        description: cmd.description,
        usage: cmd.usage,
      });
      continue;
    }

    if (subCommands.length > 0) {
      for (const sc of subCommands) {
        const fullLabel = `/${cmd.name} ${sc.name}`;
        const scMatch =
          !q ||
          cmd.name.toLowerCase().includes(q) ||
          sc.name.toLowerCase().includes(q) ||
          sc.description.toLowerCase().includes(q) ||
          fullLabel.toLowerCase().includes(q);

        if (scMatch) {
          items.push({
            type: 'sub',
            command: cmd,
            sub: sc,
            label: fullLabel,
            description: sc.description,
            usage: sc.args_usage
              ? `/${cmd.name} ${sc.name} ${sc.args_usage}`
              : fullLabel,
          });
        }
      }
    }
  }

  return items;
}

export function SlashCommandMenu({
  visible,
  query,
  commands,
  onSelect,
  onClose,
}: SlashCommandMenuProps) {
  const [activeIndex, setActiveIndex] = useState(0);
  const listRef = useRef<HTMLDivElement>(null);

  const items = useMemo(() => buildFlatList(commands, query), [commands, query]);

  // Reset selection when items change
  useEffect(() => {
    setActiveIndex(0);
  }, [items.length, query]);

  // Scroll active item into view
  useEffect(() => {
    if (!listRef.current) return;
    const el = listRef.current.children[activeIndex] as HTMLElement | undefined;
    if (el) {
      el.scrollIntoView({ block: 'nearest' });
    }
  }, [activeIndex]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (!visible) return;

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setActiveIndex((i) => (i + 1) % Math.max(items.length, 1));
          break;
        case 'ArrowUp':
          e.preventDefault();
          setActiveIndex((i) => (i - 1 + items.length) % Math.max(items.length, 1));
          break;
        case 'Enter':
          if (items.length > 0) {
            e.preventDefault();
            const item = items[activeIndex];
            if (item) {
              onSelect(item.usage);
            }
          }
          break;
        case 'Escape':
          e.preventDefault();
          onClose();
          break;
      }
    },
    [visible, items, activeIndex, onSelect, onClose],
  );

  if (!visible || items.length === 0) return null;

  return (
    <div className="absolute bottom-full left-0 right-0 mb-1 z-[100]">
      <div
        ref={listRef}
        data-slash-menu
        className="mx-3 max-h-60 overflow-y-auto rounded-lg border border-border bg-popover shadow-lg"
        onKeyDown={handleKeyDown}
        tabIndex={-1}
      >
        {items.map((item, idx) => (
          <button
            key={`${item.command.name}-${item.sub?.name ?? 'root'}-${idx}`}
            type="button"
            className={cn(
              'flex w-full items-start gap-2 px-3 py-2 text-left text-[12px] transition-colors',
              idx === activeIndex
                ? 'bg-accent text-accent-foreground'
                : 'text-foreground hover:bg-accent/50',
            )}
            onMouseEnter={() => setActiveIndex(idx)}
            onMouseDown={(e) => {
              e.preventDefault();
              onSelect(item.usage);
            }}
          >
            <Icon
              name={item.type === 'sub' ? 'ChevronRight' : 'Zap'}
              className="w-3 h-3 mt-0.5 shrink-0 opacity-60"
            />
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="font-mono font-medium text-[12px]">
                  {item.label}
                </span>
                {item.type === 'sub' && (
                  <span className="text-[9px] text-muted-foreground bg-muted px-1 py-0.5 rounded">
                    {item.command.name}
                  </span>
                )}
              </div>
              <div className="text-[11px] text-muted-foreground truncate mt-0.5">
                {item.description}
              </div>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
