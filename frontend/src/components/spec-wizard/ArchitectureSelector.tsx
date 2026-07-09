import { Pattern } from './plugins/api';
import { Icon } from '../Icon';

interface ArchitectureSelectorProps {
  options: Pattern[];
  selected: string[];
  onChange: (selected: string[]) => void;
}

export function ArchitectureSelector({ options, selected, onChange }: ArchitectureSelectorProps) {
  const toggle = (id: string) => {
    if (selected.includes(id)) {
      onChange(selected.filter((s) => s !== id));
    } else {
      onChange([...selected, id]);
    }
  };

  if (options.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        No architecture options available.
      </p>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
      {options.map((opt) => {
        const isSelected = selected.includes(opt.id);
        return (
          <button
            key={opt.id}
            type="button"
            onClick={() => toggle(opt.id)}
            className={`flex items-start p-3 border rounded-lg text-left transition-all ${
              isSelected
                ? 'border-primary bg-primary/5'
                : 'border-border hover:border-primary/50 hover:bg-accent/50'
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
            <div className={`ml-2 mt-0.5 flex-shrink-0 ${isSelected ? 'text-primary' : 'text-muted-foreground'}`}>
              {isSelected ? (
                <Icon name="Check" size={16} />
              ) : (
                <span className="block w-4 h-4 border border-current rounded-sm" />
              )}
            </div>
          </button>
        );
      })}
    </div>
  );
}
