import { useState } from 'react';
import { Button } from '../ui/button';
import { Icon } from '../Icon';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '../ui/dialog';

interface AISuggestIconProps {
  fieldName: string;
  context: string;
  currentValue?: string;
  onApply: (value: string) => void;
}

export function AISuggestIcon({ fieldName, context, currentValue, onApply }: AISuggestIconProps) {
  const [open, setOpen] = useState(false);
  const [suggestion, setSuggestion] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchSuggestion = async () => {
    setLoading(true);
    setError(null);
    try {
      const app = window.go?.main?.App;
      if (!app?.SuggestFieldValue) {
        throw new Error('AI suggestions not available');
      }
      const result = await app.SuggestFieldValue(fieldName, context, currentValue || '');
      setSuggestion(result);
      setOpen(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Button
        variant="ghost"
        size="sm"
        className="h-6 w-6 p-0 text-muted-foreground hover:text-primary"
        onClick={fetchSuggestion}
        disabled={loading}
        title={`AI suggestion for ${fieldName}`}
      >
        {loading ? (
          <Icon name="Zap" size={14} className="animate-pulse" />
        ) : (
          <Icon name="Sparkles" size={14} />
        )}
      </Button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="text-sm">AI Suggestion for {fieldName}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            {currentValue && (
              <div className="space-y-1">
                <div className="text-xs text-muted-foreground">Current value:</div>
                <div className="text-sm p-2 bg-muted rounded">{currentValue}</div>
              </div>
            )}
            <div className="space-y-1">
              <div className="text-xs text-muted-foreground">Suggestion:</div>
              <div className="text-sm p-2 bg-primary/5 border border-primary/20 rounded whitespace-pre-wrap">
                {suggestion}
              </div>
            </div>
            {error && (
              <div className="text-xs text-destructive">{error}</div>
            )}
            <div className="flex justify-end gap-2">
              <Button variant="outline" size="sm" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button size="sm" onClick={() => { onApply(suggestion); setOpen(false); }}>
                Apply
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
