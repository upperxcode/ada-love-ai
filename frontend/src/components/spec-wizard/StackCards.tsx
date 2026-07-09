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

interface StackLibraryCardProps {
  name: string;
  mandatory: boolean;
  usageExample: string;
  onToggleMandatory: () => void;
}

function StackLibraryCard({ name, mandatory, usageExample, onToggleMandatory }: StackLibraryCardProps) {
  return (
    <div className="border rounded-lg p-3 space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-sm font-mono font-medium">{name}</span>
        <button
          type="button"
          onClick={onToggleMandatory}
          className={`text-lg transition-all hover:scale-110 ${
            mandatory ? 'opacity-100' : 'opacity-30'
          }`}
          title={mandatory ? 'Mandatory (click to remove)' : 'Optional (click to make mandatory)'}
        >
          ❗
        </button>
      </div>
      <pre className="text-xs text-muted-foreground bg-muted rounded p-2 overflow-x-auto">
        {usageExample}
      </pre>
    </div>
  );
}

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
  const [libraries, setLibraries] = useState<Record<string, boolean>>({});

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

  const toggleMandatory = (libName: string) => {
    setLibraries((prev) => ({
      ...prev,
      [libName]: prev[libName] === undefined ? false : !prev[libName],
    }));
  };

  const getMandatory = (libName: string, defaultVal: boolean) => {
    return libraries[libName] ?? defaultVal;
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
              // Initialize mandatory state from template
              const initState: Record<string, boolean> = {};
              template.libraries.forEach((l) => {
                initState[l.name] = l.mandatory;
              });
              setLibraries(initState);
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
          <CardContent className="pt-0 space-y-2">
            {selectedTemplate.libraries.map((lib) => (
              <StackLibraryCard
                key={lib.name}
                name={lib.name}
                mandatory={getMandatory(lib.name, lib.mandatory)}
                usageExample={lib.usage_example}
                onToggleMandatory={() => toggleMandatory(lib.name)}
              />
            ))}
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
