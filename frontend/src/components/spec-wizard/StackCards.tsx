import { StackTemplate, StackConfig } from './plugins/api';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Icon } from '../Icon';

interface StackCardsProps {
  templates: StackTemplate[];
  selectedStacks: StackConfig[];
  onSelect: (stack: StackConfig) => void;
}

export function StackCards({ templates, selectedStacks, onSelect }: StackCardsProps) {
  if (templates.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        No stack templates available for this language.
      </p>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {templates.map((template) => {
        const isSelected = selectedStacks.some((s) => s.name === template.name);
        return (
          <Card
            key={template.id}
            className={`cursor-pointer transition-all hover:shadow-md ${
              isSelected ? 'border-primary bg-primary/5' : 'border-border'
            }`}
            onClick={() => {
              if (!isSelected) {
                onSelect({
                  name: template.name,
                  example: template.libraries
                    .map((l) => `${l.name}: ${l.usage_example}`)
                    .join('\n'),
                });
              }
            }}
          >
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-semibold flex items-center gap-2">
                <Icon name="Layers" size={16} />
                {template.name}
                {isSelected && (
                  <span className="text-xs text-primary ml-auto flex items-center gap-1">
                    <Icon name="Check" size={12} /> Selected
                  </span>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="space-y-2">
                {template.libraries.map((lib) => (
                  <div key={lib.name} className="text-xs">
                    <span className="font-mono font-medium">{lib.name}</span>
                    {lib.mandatory && (
                      <span className="text-destructive ml-1">*</span>
                    )}
                    <pre className="mt-1 p-2 bg-muted rounded text-[10px] overflow-x-auto">
                      {lib.usage_example}
                    </pre>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
