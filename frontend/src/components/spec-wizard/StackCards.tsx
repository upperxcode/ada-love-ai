import { StackTemplate, StackConfig } from './plugins/api';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Icon } from '../Icon';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../ui/select';

interface StackCardsProps {
  templates: StackTemplate[];
  selectedStack: StackConfig | null;
  onSelect: (stack: StackConfig | null) => void;
  manualStacks: StackConfig[];
  onAddManual: (stack: StackConfig) => void;
  onRemoveManual: (index: number) => void;
}

export function StackCards({
  templates,
  selectedStack,
  onSelect,
  manualStacks,
  onAddManual,
  onRemoveManual,
}: StackCardsProps) {
  return (
    <div className="space-y-4">
      {/* Combo Select */}
      <div className="space-y-2">
        <label className="text-sm font-medium">Select Stack Template</label>
        <Select
          value={selectedStack?.name || ''}
          onValueChange={(value) => {
            const template = templates.find((t) => t.name === value);
            if (template) {
              onSelect({
                name: template.name,
                example: template.libraries
                  .map((l) => `${l.name}: ${l.usage_example}`)
                  .join('\n'),
              });
            }
          }}
        >
          <SelectTrigger>
            <SelectValue placeholder="Choose a stack template..." />
          </SelectTrigger>
          <SelectContent>
            {templates.map((template) => (
              <SelectItem key={template.id} value={template.name}>
                {template.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Selected Stack Card */}
      {selectedStack && (
        <div className="space-y-2">
          <label className="text-sm font-medium">Selected Stack</label>
          {templates
            .filter((t) => t.name === selectedStack.name)
            .map((template) => (
              <Card key={template.id} className="border-primary bg-primary/5">
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-semibold flex items-center gap-2">
                    <Icon name="Layers" size={16} />
                    {template.name}
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
            ))}
        </div>
      )}

      {/* Manual Stacks */}
      {manualStacks.length > 0 && (
        <div className="space-y-2">
          <label className="text-sm font-medium">Manual Stacks</label>
          {manualStacks.map((stack, idx) => (
            <Card key={idx} className="border-border">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-semibold flex items-center justify-between">
                  <span className="flex items-center gap-2">
                    <Icon name="Layers" size={16} />
                    {stack.name || `Manual Stack ${idx + 1}`}
                  </span>
                  <button
                    onClick={() => onRemoveManual(idx)}
                    className="text-muted-foreground hover:text-destructive"
                  >
                    <Icon name="X" size={14} />
                  </button>
                </CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                <pre className="p-2 bg-muted rounded text-[10px] overflow-x-auto">
                  {stack.example}
                </pre>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
