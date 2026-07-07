import { useState } from 'react';
import { Button } from './ui/button';
import { Icon } from './Icon';
import { Dialog, DialogContent } from './ui/dialog';

interface ExpandableEditorProps {
  value: string;
  onChange: (v: string) => void;
  label: string;
}

export function ExpandableEditor({
  value,
  onChange,
  label,
}: ExpandableEditorProps) {
  const [expanded, setExpanded] = useState(false);
  const [preview, setPreview] = useState(false);

  const editor = (
    <textarea
      className={`w-full rounded-md border border-border bg-transparent px-3 py-2 text-sm resize-none ${
        expanded ? 'h-full min-h-[300px]' : 'h-16'
      }`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  );

  return (
    <>
      <div className="flex items-center justify-between">
        <label className="text-xs text-muted-foreground">{label}</label>
        <Button
          variant="ghost"
          size="sm"
          className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground"
          onClick={() => setExpanded(true)}
        >
          <Icon name="Maximize2" className="w-3 h-3" />
        </Button>
      </div>
      {editor}

      <Dialog open={expanded} onOpenChange={setExpanded}>
        <DialogContent className="max-w-4xl h-[80vh] p-0 gap-0 flex flex-col">
          <div className="flex items-center justify-between px-6 py-3 border-b border-border shrink-0">
            <h2 className="text-sm font-semibold text-foreground">{label}</h2>
            <div className="flex items-center gap-1 mr-10">
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 text-muted-foreground"
                onClick={() => setPreview(!preview)}
              >
                {preview ? (
                  <Icon name="FileText" className="w-4 h-4" />
                ) : (
                  <Icon name="Search" className="w-4 h-4" />
                )}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 text-muted-foreground"
                onClick={() => setExpanded(false)}
              >
                <Icon name="Minimize2" className="w-4 h-4" />
              </Button>
            </div>
          </div>
          <div className="flex-1 p-6 overflow-auto">
            {preview ? (
              <div className="text-sm whitespace-pre-wrap text-foreground">
                {value}
              </div>
            ) : (
              <textarea
                className="w-full h-full min-h-[300px] rounded-md border border-border bg-transparent px-3 py-2 text-sm resize-none"
                value={value}
                onChange={(e) => onChange(e.target.value)}
              />
            )}
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
