import { Pattern } from './plugins/api';
import { Icon } from '../Icon';

interface PatternSelectorProps {
  options: Pattern[];
  selected: string[];
  onChange: (selected: string[]) => void;
  title: string;
}

export function PatternSelector({ options, selected, onChange, title }: PatternSelectorProps) {
  const toggle = (id: string) => {
    if (selected.includes(id)) {
      onChange(selected.filter((s) => s !== id));
    } else {
      onChange([...selected, id]);
    }
  };

  if (options.length === 0) {
    return null;
  }

  return (
    <div className="space-y-2">
      <label className="text-sm font-medium">{title}</label>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
        {options.map((opt) => {
          const isSelected = selected.includes(opt.id || opt.name);
          return (
            <button
              key={opt.id || opt.name}
              type="button"
              onClick={() => toggle(opt.id || opt.name)}
              className={`flex items-start p-3 border rounded-lg text-left transition-all ${
                isSelected
                  ? 'border-primary bg-primary/5'
                  : 'border-transparent hover:border-primary/50 hover:bg-accent/50'
              }`}
            >
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium truncate">{opt.name}</div>
                {opt.description && (
                  <div className="text-xs text-muted-foreground line-clamp-2 mt-0.5">
                    {opt.description}
                  </div>
                )}
              </div>
              {isSelected && (
                <div className="ml-2 mt-0.5 flex-shrink-0 text-primary">
                  <Icon name="Check" size={16} />
                </div>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
