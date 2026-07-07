import { BaseCard } from '../BaseCard';
import { Icon } from '../Icon';
import * as api from '../../api';

interface AgentCardProps {
  agent: api.backend.AgentConfig;
  onEdit?: (agent: api.backend.AgentConfig) => void;
  onDelete?: (name: string) => void;
}

function AgentCard({ agent, onEdit, onDelete }: AgentCardProps) {
  return (
    <BaseCard
      color={agent.color || '#6366f1'}
      headerLeft={
        agent.category ? (
          <span className="text-xs text-white opacity-90">{agent.category}</span>
        ) : null
      }
      headerRight={
        <div className="flex gap-1">
          {onEdit && (
            <button
              className="base-card-btn"
              onClick={(e) => {
                e.stopPropagation();
                onEdit(agent);
              }}
              title="Edit"
            >
              <Icon name="Edit" className="w-3 h-3" />
            </button>
          )}
          {onDelete && (
            <button
              className="base-card-btn"
              onClick={(e) => {
                e.stopPropagation();
                onDelete(agent.name);
              }}
              title="Delete"
            >
              <Icon name="Trash2" className="w-3 h-3" />
            </button>
          )}
        </div>
      }
      icon={agent.icon || '🤖'}
      title={agent.name}
    >
      <div className="base-card-desc">{agent.persona || 'No persona'}</div>
      <div className="mt-1 flex flex-col gap-0.5 text-xs text-muted-foreground">
        <span className="truncate">
          <span className="opacity-70">Provider:</span> {agent.provider || '—'}
        </span>
        <span className="truncate">
          <span className="opacity-70">Model:</span> {agent.model || '—'}
        </span>
      </div>
    </BaseCard>
  );
}

export default AgentCard;
