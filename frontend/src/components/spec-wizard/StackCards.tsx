import { useState } from 'react';
import { StackTemplate, StackConfig } from './plugins/api';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
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
  onAddManual: (stack: StackConfig) => void;
}

export function StackCards({
  templates,
  selectedStack,
  onSelect,
  onAddManual,
}: StackCardsProps) {
  const [manualName, setManualName] = useState('');
  const [manualExample, setManualExample] = useState('');

  // Find the selected template to display its card
  const selectedTemplate = templates.find((t) => t.name === selectedStack?.name);

  const handleAddManual = () => {
    if (manualName.trim()) {
      onAddManual({
        name: manualName.trim(),
        example: manualExample.trim(),
      });
      setManualName('');
      setManualExample('');
    }
  };

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

      {/* Selected Template Card */}
      {selectedTemplate && (
        <Card className="border-primary bg-primary/5">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-semibold flex items-center gap-2">
              <Icon name="Layers" size={16} />
              {selectedTemplate.name}
              <span className="text-xs text-primary ml-auto flex items-center gap-1">
                <Icon name="Check" size={12} /> Active
              </span>
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            <div className="space-y-2">
              {selectedTemplate.libraries.map((lib) => (
                <div key={lib.name} className="flex items-start gap-2 text-xs">
                  <div className="flex items-center gap-1 min-w-0 flex-1">
                    <span className="font-mono font-medium">{lib.name}</span>
                    {lib.mandatory && (
                      <span
                        className="inline-flex items-center gap-0.5 px-1 py-0.5 rounded bg-amber-100 text-amber-700 text-[10px] font-medium"
                        title="Mandatory: AI will use this library"
                      >
                        <Icon name="Zap" size={10} />
                        required
                      </span>
                    )}
                  </div>
                </div>
              ))}
              {/* Show usage examples collapsed */}
              <div className="mt-2 space-y-1">
                {selectedTemplate.libraries.map((lib) => (
                  <pre
                    key={`example-${lib.name}`}
                    className="p-2 bg-muted rounded text-[10px] overflow-x-auto"
                  >
                    <span className="text-muted-foreground">{lib.name}:</span>{' '}
                    {lib.usage_example}
                  </pre>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Manual Stack Input */}
      <div className="space-y-2">
        <label className="text-sm font-medium">Add Manual Stack</label>
        <div className="flex gap-2">
          <Input
            value={manualName}
            onChange={(e) => setManualName(e.target.value)}
            placeholder="Stack name"
            className="flex-1"
          />
          <Input
            value={manualExample}
            onChange={(e) => setManualExample(e.target.value)}
            placeholder="Example usage (optional)"
            className="flex-1"
          />
          <Button
            variant="outline"
            size="sm"
            onClick={handleAddManual}
            disabled={!manualName.trim()}
          >
            <Icon name="Plus" size={16} /> Add
          </Button>
        </div>
      </div>
    </div>
  );
}
