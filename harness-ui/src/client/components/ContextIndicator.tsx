import type { ContextWindow } from '@shared/types';

interface ContextIndicatorProps {
  context?: ContextWindow | null;
}

// Unified thresholds to avoid drift between color and label functions
const THRESHOLDS = {
  good: 50,
  moderate: 70,
  high: 90,
} as const;

interface ContextStatus {
  color: string;
  label: string;
}

function getContextStatus(percentage: number): ContextStatus {
  // Clamp percentage to valid range
  const clamped = Math.max(0, Math.min(100, percentage));

  if (clamped < THRESHOLDS.good) {
    return { color: 'var(--status-running)', label: 'Good' }; // green
  }
  if (clamped < THRESHOLDS.moderate) {
    return { color: 'var(--status-waiting)', label: 'Moderate' }; // yellow
  }
  if (clamped < THRESHOLDS.high) {
    return { color: 'var(--status-blocked)', label: 'High' }; // red
  }
  return { color: 'var(--status-blocked)', label: 'Critical' }; // red
}

export function ContextIndicator({ context }: ContextIndicatorProps) {
  if (!context) {
    return (
      <div className="context-indicator" role="region" aria-label="Context usage indicator">
        <div className="context-label">Context: --</div>
      </div>
    );
  }

  // Clamp to valid range for safety
  const used_percentage = Math.max(0, Math.min(100, context.used_percentage));
  const { color, label } = getContextStatus(used_percentage);
  const percentDisplay = used_percentage.toFixed(0);

  return (
    <div className="context-indicator" role="region" aria-label="Context usage indicator">
      <div className="context-header">
        <span className="context-title">Context Usage</span>
        <span className="context-status" style={{ color }} aria-hidden="true">
          {label}
        </span>
      </div>
      <div className="context-bar">
        <div
          className="context-track"
          role="progressbar"
          aria-valuenow={used_percentage}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-label={`Context usage: ${percentDisplay}% - ${label}`}
        >
          <div
            className="context-fill"
            style={{
              width: `${used_percentage}%`,
              backgroundColor: color,
            }}
          />
        </div>
        <div className="context-percent" style={{ color }} aria-hidden="true">
          {percentDisplay}%
        </div>
      </div>
      {used_percentage >= THRESHOLDS.moderate && (
        <div
          className="context-warning"
          role="status"
          aria-live="polite"
        >
          {used_percentage >= THRESHOLDS.high
            ? 'Consider running /compact to free up context'
            : 'Context usage is getting high'}
        </div>
      )}
    </div>
  );
}
