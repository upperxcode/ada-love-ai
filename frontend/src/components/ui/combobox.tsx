import * as React from 'react';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';

export interface ComboboxOption {
  value: string;
  label?: string;
}

export interface ComboboxProps {
  value: string;
  onValueChange: (value: string) => void;
  options: ComboboxOption[];
  placeholder?: string;
  /** Texto exibido quando não há opções que casem com o texto digitado. */
  emptyText?: string;
  disabled?: boolean;
  className?: string;
}

/**
 * Combobox baseado em Popover + Input, sem cmdk.
 *
 * O campo é editável: o usuário pode digitar um valor custom (não presente em
 * `options`) ou escolher uma das sugestões filtradas. A lista é filtrada
 * insensivelmente a caixa conforme o texto é digitado.
 */
export const Combobox = React.forwardRef<HTMLInputElement, ComboboxProps>(
  (
    {
      value,
      onValueChange,
      options,
      placeholder,
      emptyText = 'Nenhum resultado — digite um valor custom',
      disabled,
      className,
    },
    ref,
  ) => {
    const [open, setOpen] = React.useState(false);
    const [query, setQuery] = React.useState('');

    // Sincroniza o texto do input com o valor controlado quando fechado e o
    // query não foi tocado pelo usuário. Isso evita sobrescrever o que o
    // usuário está digitando enquanto o popover está aberto.
    React.useEffect(() => {
      if (!open) setQuery('');
    }, [open]);

    const inputText = open ? query : value;
    const filtered = React.useMemo(() => {
      if (!query.trim()) return options;
      const q = query.toLowerCase();
      return options.filter((opt) => {
        const label = (opt.label ?? opt.value).toLowerCase();
        return label.includes(q);
      });
    }, [options, query]);

    const handleSelect = (opt: ComboboxOption) => {
      onValueChange(opt.value);
      setOpen(false);
    };

    return (
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Input
            ref={ref}
            value={inputText}
            disabled={disabled}
            placeholder={placeholder}
            className={className}
            onChange={(e) => {
              // Propaga o texto digitado imediatamente para o valor controlado
              // (permite valores custom), e mantém o query para filtrar.
              const next = e.target.value;
              setQuery(next);
              onValueChange(next);
              if (!open) setOpen(true);
            }}
            onFocus={() => setOpen(true)}
          />
        </PopoverTrigger>
        <PopoverContent
          align="start"
          className="w-[var(--radix-popover-trigger-width)] min-w-72 p-0"
        >
          {filtered.length === 0 ? (
            <div className="px-3 py-2 text-sm text-muted-foreground">{emptyText}</div>
          ) : (
            <ScrollArea className="max-h-60">
              <div className="p-1">
                {filtered.map((opt) => {
                  const selected = opt.value === value;
                  return (
                    <button
                      key={opt.value}
                      type="button"
                      onMouseDown={(e) => {
                        // Previne que o blur do input feche o popover antes do
                        // clique registrar.
                        e.preventDefault();
                      }}
                      onClick={() => handleSelect(opt)}
                      className={cn(
                        'flex w-full items-center rounded-sm px-2 py-1.5 text-left text-sm outline-none hover:bg-accent hover:text-accent-foreground',
                        selected && 'bg-accent text-accent-foreground',
                      )}
                    >
                      {opt.label ?? opt.value}
                    </button>
                  );
                })}
              </div>
            </ScrollArea>
          )}
        </PopoverContent>
      </Popover>
    );
  },
);
Combobox.displayName = 'Combobox';
