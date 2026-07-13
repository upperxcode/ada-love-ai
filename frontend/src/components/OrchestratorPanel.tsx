import { useChat, type OrchestratorSubTask } from './ChatContext';
import { Icon } from './Icon';

function agentIcon(agent: string): string {
  switch (agent) {
    case 'golang':
      return 'Terminal';
    case 'react':
      return 'Layout';
    case 'tester':
      return 'FlaskConical';
    default:
      return 'Bot';
  }
}

function agentLabel(agent: string): string {
  switch (agent) {
    case 'golang':
      return 'Go';
    case 'react':
      return 'React';
    case 'tester':
      return 'Tester';
    default:
      return agent || '...';
  }
}

function statusIcon(status: OrchestratorSubTask['status']): string {
  switch (status) {
    case 'completed':
      return 'CheckCircle2';
    case 'error':
      return 'XCircle';
    case 'started':
    default:
      return 'Loader';
  }
}

function statusColor(status: OrchestratorSubTask['status']): string {
  switch (status) {
    case 'completed':
      return 'text-green-500';
    case 'error':
      return 'text-red-500';
    case 'started':
    default:
      return 'text-blue-500 animate-spin';
  }
}

export function OrchestratorPanel() {
  const { orchestrator } = useChat();

  if (!orchestrator.active) return null;

  return (
    <div className="mx-6 mb-2 p-3 bg-primary/5 border border-primary/20 rounded-lg">
      {/* Header */}
      <div className="flex items-center gap-2 mb-2">
        <Icon name="Network" className="w-4 h-4 text-primary" />
        <span className="text-xs font-medium text-foreground">Orquestrador</span>
        {orchestrator.nextAgent && (
          <span className="text-[10px] px-1.5 py-0.5 rounded bg-primary/10 text-primary">
            {orchestrator.nextAgent}
          </span>
        )}
      </div>

      {/* Reasoning */}
      {orchestrator.reasoning && (
        <p className="text-[11px] text-muted-foreground mb-2 leading-relaxed">
          {orchestrator.reasoning}
        </p>
      )}

      {/* Sub-tasks */}
      {orchestrator.subTasks.length > 0 && (
        <div className="flex flex-col gap-1.5">
          {orchestrator.subTasks.map((st) => (
            <div key={st.id} className="flex items-center gap-2 text-[11px]">
              <Icon
                name={statusIcon(st.status)}
                className={`w-3.5 h-3.5 shrink-0 ${statusColor(st.status)}`}
              />
              <Icon
                name={agentIcon(st.agent)}
                className="w-3 h-3 text-muted-foreground shrink-0"
              />
              <span className="text-muted-foreground font-medium">
                {agentLabel(st.agent)}
              </span>
              <span className="text-foreground truncate flex-1">
                {st.task || `Sub-task ${st.id}`}
              </span>
              {st.error && (
                <span className="text-[10px] text-red-500 truncate max-w-[120px]">
                  {st.error}
                </span>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
