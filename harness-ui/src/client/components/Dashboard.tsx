import { useMemo, useRef, useCallback, useEffect } from 'react';
import type { PlansData, PTYSession, ContextWindow } from '@shared/types';
import { PlansBoard } from './PlansBoard';
import { TerminalCards } from './TerminalCards';
import { ProgressSummary } from './ProgressSummary';
import { ContextIndicator } from './ContextIndicator';
import { CommandBar } from './CommandBar';
import { Terminal, TerminalHandle } from './Terminal';

type ProjectState = 'active' | 'idle' | 'stopped';

interface DashboardProps {
  plans: PlansData;
  sessions: PTYSession[];
  worktreePaths?: string[];
  contextWindow?: ContextWindow | null;
  onTerminalFocus: (sessionId: string) => void;
  onCreateSession: (worktreePath?: string) => void;
  onSendInput: (sessionId: string, data: string) => void;
  commonTerminalId?: string;
  onCommonTerminalInput?: (data: string) => void;
  commonTerminalLogs?: string[];
  commonTerminalOutput?: string;
}

function getProjectState(sessions: PTYSession[]): ProjectState {
  if (sessions.length === 0) return 'stopped';
  const hasRunning = sessions.some((s) => s.status === 'RUNNING');
  const hasWaiting = sessions.some((s) => s.status === 'WAITING');
  if (hasRunning || hasWaiting) return 'active';
  return 'idle';
}

export function Dashboard({
  plans,
  sessions,
  worktreePaths,
  contextWindow,
  onTerminalFocus,
  onCreateSession,
  onSendInput,
  commonTerminalId,
  onCommonTerminalInput,
  commonTerminalLogs,
  commonTerminalOutput,
}: DashboardProps) {
  const commonTerminalRef = useRef<TerminalHandle>(null);
  const commonLogIndexRef = useRef(0);
  const projectState = useMemo(() => getProjectState(sessions), [sessions]);

  useEffect(() => {
    commonLogIndexRef.current = 0;
  }, [commonTerminalId]);

  const hydrateCommonLogs = useCallback(
    (reset = false) => {
      if (!commonTerminalId || !commonTerminalRef.current || !commonTerminalLogs?.length) return;
      if (reset) {
        commonTerminalRef.current.reset();
        commonLogIndexRef.current = 0;
      }
      const start = commonLogIndexRef.current;
      if (commonTerminalLogs.length > start) {
        const chunk = commonTerminalLogs.slice(start).join('');
        if (chunk) {
          commonTerminalRef.current.write(chunk);
        }
        commonLogIndexRef.current = commonTerminalLogs.length;
      }
    },
    [commonTerminalId, commonTerminalLogs]
  );

  useEffect(() => {
    hydrateCommonLogs();
  }, [hydrateCommonLogs]);

  useEffect(() => {
    if (!commonTerminalId || !commonTerminalOutput || !commonTerminalRef.current) return;
    commonTerminalRef.current.write(commonTerminalOutput);
  }, [commonTerminalId, commonTerminalOutput]);

  useEffect(() => {
    const handleVisibility = () => {
      if (document.visibilityState !== 'visible') return;
      if (commonTerminalRef.current) {
        commonTerminalRef.current.fit();
      }
      hydrateCommonLogs(true);
    };
    document.addEventListener('visibilitychange', handleVisibility);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibility);
    };
  }, [hydrateCommonLogs]);

  const handleCommonTerminalInput = useCallback(
    (data: string) => {
      if (onCommonTerminalInput) {
        onCommonTerminalInput(data);
      }
    },
    [onCommonTerminalInput]
  );

  const handleResize = useCallback(() => {
    // No-op for now, can add resize handling later
  }, []);

  return (
    <div className="dashboard">
      <div className="project-status-bar">
        <div className={`project-state project-state-${projectState}`}>
          <span className="state-dot" />
          <span className="state-label">
            {projectState === 'active' && 'Active'}
            {projectState === 'idle' && 'Idle'}
            {projectState === 'stopped' && 'Stopped'}
          </span>
        </div>
        <div className="session-count">
          {sessions.length} terminal{sessions.length !== 1 ? 's' : ''}
        </div>
      </div>

      <ProgressSummary summary={plans.summary} />

      <ContextIndicator context={contextWindow} />

      <CommandBar sessions={sessions} onSendCommand={onSendInput} />

      <div className="common-terminal-panel">
        <div className="panel-header">
          <h3>Common Terminal (Wall-打ち)</h3>
          <span className="terminal-hint">Interactive session for quick commands</span>
        </div>
        <div className="common-terminal-body">
          {commonTerminalId ? (
            <Terminal
              ref={commonTerminalRef}
              sessionId={commonTerminalId}
              onInput={handleCommonTerminalInput}
              onResize={handleResize}
            />
          ) : (
            <div className="terminal-placeholder">
              No common terminal available. Create a session to start.
            </div>
          )}
        </div>
      </div>

      <PlansBoard sections={plans.sections} />

      <TerminalCards
        sessions={sessions}
        worktreePaths={worktreePaths}
        onFocus={onTerminalFocus}
        onCreateSession={onCreateSession}
      />
    </div>
  );
}
