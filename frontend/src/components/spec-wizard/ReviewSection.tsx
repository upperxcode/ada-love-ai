import { Icon } from '../Icon';

interface ReviewItem {
  title: string;
  value: string;
}

interface ReviewSectionProps {
  recommendations: string[];
  healthScore: number;
  items: ReviewItem[];
}

export function ReviewSection({ recommendations, healthScore, items }: ReviewSectionProps) {
  return (
    <div className="space-y-6">
      {/* Summary */}
      <div className="grid grid-cols-2 gap-4">
        {items.map((item) => (
          <div key={item.title} className="space-y-1">
            <div className="text-xs text-muted-foreground">{item.title}</div>
            <div className="text-sm font-medium truncate">{item.value || '—'}</div>
          </div>
        ))}
      </div>

      {/* Recommendations */}
      {recommendations.length > 0 && (
        <div className="bg-accent/30 border border-accent rounded-lg p-4">
          <h4 className="text-sm font-semibold mb-2 flex items-center gap-2">
            <Icon name="Sparkles" size={16} className="text-amber-500" />
            Architecture Recommendations
          </h4>
          <ul className="space-y-2">
            {recommendations.map((rec, i) => (
              <li key={i} className="flex items-start gap-2 text-sm">
                <span className="text-amber-500 mt-0.5">🔹</span>
                <span>{rec}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
