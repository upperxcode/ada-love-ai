import { useState } from 'react';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import { useChat } from './ChatContext';
import { Icon } from './Icon';
import { CopyButton } from './CopyButton';

// looksLikeMarkdown detects whether the command output should be rendered as
// rich markdown. Command reports (e.g. /health) are plain text with ASCII
// boxes and emoji — feeding those to react-markdown mangles the layout, so we
// fall back to a monospaced <pre> for anything that isn't structured markdown.
function looksLikeMarkdown(text: string): boolean {
  if (!text) return false;
  // ASCII box-drawing characters => plain/text report, keep as <pre>.
  if (/[━┃┏┓┗┛┣┫┳┻╋│─]/.test(text)) return false;
  // Typical markdown signals.
  return /(^|\n)#{1,6} |\n\*\s|\n-\s|\n\d+\.\s|```|\[.+\]\(.+\)|\*\*|\[!.*?\]/.test(text);
}

export function CommandResultPanel() {
  const { commandResult, setCommandResult } = useChat();
  const [collapsed, setCollapsed] = useState(false);

  if (!commandResult) return null;

  const close = () => setCommandResult(null);
  const toggle = () => setCollapsed((c) => !c);
  const isMarkdown = looksLikeMarkdown(commandResult.output);

  return (
    <div className="pointer-events-none absolute inset-0 z-[60] p-[10px]">
      <div className="pointer-events-auto ml-auto flex max-h-full w-full max-w-2xl flex-col overflow-hidden rounded-xl border border-border bg-popover shadow-2xl">
        {/* Header */}
        <div className="flex shrink-0 items-center gap-2 border-b border-border px-3 py-2">
          <Icon name="Terminal" className="h-4 w-4 text-muted-foreground" />
          <span className="font-mono text-xs font-medium text-foreground">
            /{commandResult.command}
          </span>
          <span className="text-[10px] text-muted-foreground/60">result</span>
          <span className="flex-1" />
          <button
            type="button"
            onClick={toggle}
            className="toolbar-btn"
            title={collapsed ? 'Expand' : 'Collapse'}
          >
            <Icon name={collapsed ? 'ChevronUp' : 'ChevronDown'} className="h-3.5 w-3.5" />
          </button>
          <button
            type="button"
            onClick={close}
            className="toolbar-btn"
            title="Close"
          >
            <Icon name="X" className="h-3.5 w-3.5" />
          </button>
        </div>

        {/* Body */}
        {!collapsed && (
          <div className="flex min-h-0 flex-1 flex-col">
            <div className="min-h-0 flex-1 overflow-y-auto px-4 py-3">
              {isMarkdown ? (
                <div className="prose prose-sm max-w-none text-[13px] leading-relaxed text-foreground prose-headings:text-foreground prose-p:text-foreground prose-li:text-foreground prose-strong:text-foreground prose-code:text-foreground prose-pre:bg-muted prose-pre:text-foreground prose-a:text-blue-400 prose-a:no-underline hover:prose-a:underline">
                  <Markdown
                    remarkPlugins={[remarkGfm]}
                    rehypePlugins={[rehypeHighlight]}
                  >
                    {commandResult.output}
                  </Markdown>
                </div>
              ) : (
                <pre className="whitespace-pre-wrap break-words font-mono text-[12px] leading-relaxed text-foreground">
                  {commandResult.output}
                </pre>
              )}
            </div>
            <div className="flex shrink-0 items-center justify-end gap-1.5 border-t border-border px-3 py-1.5">
              <CopyButton text={commandResult.output} />
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
