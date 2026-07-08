import { useState } from 'react';
import { Button } from './ui/button';
import { Icon } from './Icon';

export function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {}
  };

  return (
    <Button
      variant="ghost"
      size="icon"
      className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100 transition-opacity"
      onClick={handleCopy}
      title={copied ? 'Copiado!' : 'Copiar resposta'}
    >
      <Icon
        name={copied ? 'Check' : 'Copy'}
        className="w-3 h-3"
      />
    </Button>
  );
}
