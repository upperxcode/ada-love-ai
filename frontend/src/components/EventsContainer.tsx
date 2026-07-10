import { useState, useEffect, useRef } from 'react';
import { useChat } from './ChatContext';
import { Icon } from './Icon';

function Timer({ isRunning }: { isRunning: boolean }) {
  const [seconds, setSeconds] = useState(0);
  const startRef = useRef(Date.now());

  useEffect(() => {
    if (isRunning) {
      startRef.current = Date.now();
      setSeconds(0);
      const interval = setInterval(() => {
        setSeconds(Math.floor((Date.now() - startRef.current) / 1000));
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [isRunning]);

  if (!isRunning) return null;

  return (
    <span className="text-[10px] text-muted-foreground/50 tabular-nums">
      {seconds < 60 ? `${seconds}s` : `${Math.floor(seconds / 60)}m ${seconds % 60}s`}
    </span>
  );
}

function stageLabel(stage: string): string {
  if (!stage) return '';
  if (stage === 'thinking') return 'Thinking';
  if (stage.startsWith('tool:')) {
    const tool = stage.replace('tool:', '');
    return `Tool: ${tool}`;
  }
  if (stage.startsWith('subagent:')) {
    const label = stage.replace('subagent:', '');
    return `SubAgent${label ? `: ${label}` : ''}`;
  }
  if (stage.startsWith('agent:')) {
    const label = stage.replace('agent:', '');
    return `Agent${label ? `: ${label}` : ''}`;
  }
  if (stage === 'agent_done') return 'Agent done';
  if (stage === 'agent_error') return 'Agent error';
  if (stage === 'writing') return 'Writing';
  return stage;
}

function stageIcon(stage: string): string {
  if (stage === 'thinking') return 'Brain';
  if (stage.startsWith('tool:')) return 'Wrench';
  if (stage.startsWith('subagent:')) return 'GitBranch';
  if (stage.startsWith('agent:')) return 'Bot';
  if (stage === 'agent_done') return 'CheckCircle';
  if (stage === 'agent_error') return 'AlertTriangle';
  if (stage === 'writing') return 'PenLine';
  return 'Loader';
}

export function EventsContainer() {
  const { loading, stage, pendingApproval, answerApproval, stopGeneration } = useChat();
  
  const hasApproval = !!pendingApproval;
  const hasLoading = loading;
  const hasStage = !!stage;
  const hasActivity = hasApproval || hasLoading || hasStage;

  if (!hasActivity) return null;

  return (
    <div className="mb-2 px-3 py-2 bg-background/95 backdrop-blur-sm border border-border/50 shadow-md">
      {/* Permission dialog */}
      {pendingApproval && (
        <div className="mb-2 p-2 bg-amber-500/10 border border-amber-500/30 rounded">
          <div className="flex items-start gap-1.5">
            <Icon name="ShieldAlert" className="w-3.5 h-3.5 text-amber-500 mt-0.5 shrink-0" />
            <div className="flex-1">
              <span className="text-xs font-medium text-foreground mb-0.5 block">
                Allow tool?
              </span>
              <span className="text-xs text-muted-foreground">
                <strong>{pendingApproval.tool}</strong> requires your permission to continue.
              </span>
            </div>
          </div>
          <div className="flex items-center gap-1.5 mt-1.5">
            <button
              type="button"
              className="px-2.5 py-1 rounded-md bg-green-600 text-white text-[11px] font-medium hover:bg-green-700 transition-colors"
              onClick={() => answerApproval(true)}
            >
              Yes, allow
            </button>
            <button
              type="button"
              className="px-2.5 py-1 rounded-md bg-red-600 text-white text-[11px] font-medium hover:bg-red-700 transition-colors"
              onClick={() => answerApproval(false, 'Denied by user')}
            >
              No, deny
            </button>
          </div>
        </div>
      )}

      {/* Loading status */}
      {hasLoading && (
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <span className="animate-pulse text-xs">•</span>
            
            {stage && (
              <div className="flex items-center gap-1.5">
                <Icon name={stageIcon(stage)} className="w-3.5 h-3.5" />
                <span className="text-xs text-muted-foreground">{stageLabel(stage)}</span>
              </div>
            )}
            
            <Timer isRunning={true} />
          </div>
          
          <button
            type="button"
            className="flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
            onClick={stopGeneration}
            title="Cancel generation"
          >
            <Icon name="Square" className="w-2.5 h-2.5" />
            <span>Stop</span>
          </button>
        </div>
      )}
    </div>
  );
}