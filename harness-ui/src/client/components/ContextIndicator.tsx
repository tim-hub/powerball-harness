import type { ContextWindow } from '@shared/types';

interface ContextIndicatorProps {
  context?: ContextWindow | null;
}

function getContextColor(percentage: number): string {
  if (percentage < 50) return 'var(--status-running)'; // green
  if (percentage < 70) return 'var(--status-waiting)'; // yellow
  return 'var(--status-blocked)'; // red
}

function getContextLabel(percentage: number): string {
  if (percentage < 50) return 'Good';
  if (percentage < 70) return 'Moderate';
  if (percentage < 90) return 'High';
  return 'Critical';
}

export function ContextIndicator({ context }: ContextIndicatorProps) {
  if (!context) {
    return (
      <div className="context-indicator">
        <div className="context-label">Context: --</div>
      </div>
    );
  }

  const { used_percentage } = context;
  const color = getContextColor(used_percentage);
  const label = getContextLabel(used_percentage);

  return (
    <div className="context-indicator">
      <div className="context-header">
        <span className="context-title">Context Usage</span>
        <span className="context-status" style={{ color }}>
          {label}
        </span>
      </div>
      <div className="context-bar">
        <div className="context-track">
          <div
            className="context-fill"
            style={{
              width: `${used_percentage}%`,
              backgroundColor: color,
            }}
          />
        </div>
        <div className="context-percent" style={{ color }}>
          {used_percentage.toFixed(0)}%
        </div>
      </div>
      {used_percentage >= 70 && (
        <div className="context-warning">
          {used_percentage >= 90
            ? 'Consider running /compact to free up context'
            : 'Context usage is getting high'}
        </div>
      )}
    </div>
  );
}
