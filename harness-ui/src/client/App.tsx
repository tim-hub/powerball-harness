import { useState, useCallback, useEffect } from 'react';
import type { Page, PlansData, ContextWindow } from '@shared/types';
import { useWebSocket } from './hooks/useWebSocket';
import { usePlans } from './hooks/usePlans';
import { useProjects } from './hooks/useProjects';
import { Dashboard } from './components/Dashboard';
import { WorkPage } from './components/WorkPage';
import { SettingsPage } from './components/SettingsPage';
import { ProjectSelector } from './components/ProjectSelector';

const COMMON_TERMINAL_PROJECT_ID = '__common__';
const CONTEXT_POLL_INTERVAL = 10000; // 10 seconds

export function App() {
  const [page, setPage] = useState<Page>('dashboard');
  const [focusedTerminal, setFocusedTerminal] = useState<string | null>(null);
  const [commonTerminalId, setCommonTerminalId] = useState<string | null>(null);
  const [commonTerminalOutput, setCommonTerminalOutput] = useState<string>('');
  const [contextWindow, setContextWindow] = useState<ContextWindow | null>(null);

  const { projects, activeProject, activateProject, addProject, removeProject } = useProjects();

  // Fetch context window data periodically (with Page Visibility API optimization)
  useEffect(() => {
    let intervalId: ReturnType<typeof setInterval> | null = null;
    let isFetching = false;

    const fetchContext = async () => {
      // Prevent overlapping requests
      if (isFetching) return;
      isFetching = true;

      try {
        const res = await fetch('/api/context');
        if (res.ok) {
          const data = await res.json();
          if (data && typeof data.used_percentage === 'number') {
            setContextWindow(data);
          }
        }
      } catch {
        // Silently ignore errors
      } finally {
        isFetching = false;
      }
    };

    const startPolling = () => {
      if (!intervalId) {
        intervalId = setInterval(fetchContext, CONTEXT_POLL_INTERVAL);
      }
    };

    const stopPolling = () => {
      if (intervalId) {
        clearInterval(intervalId);
        intervalId = null;
      }
    };

    const handleVisibilityChange = () => {
      if (document.hidden) {
        stopPolling();
      } else {
        fetchContext(); // Fetch immediately when becoming visible
        startPolling();
      }
    };

    // Initial fetch and start polling if visible
    fetchContext();
    if (!document.hidden) {
      startPolling();
    }

    document.addEventListener('visibilitychange', handleVisibilityChange);

    return () => {
      stopPolling();
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, []);
  const { plans, updatePlans } = usePlans({ projectId: activeProject?.id });

  const handlePlansUpdate = useCallback(
    (data: PlansData) => {
      updatePlans(data);
    },
    [updatePlans]
  );

  const handleLogChunk = useCallback(
    (sessionId: string, data: string) => {
      if (sessionId === commonTerminalId) {
        setCommonTerminalOutput(data);
      }
    },
    [commonTerminalId]
  );

  const { connected, sessions, sendInput, createSession, destroySession, resizeTerminal } =
    useWebSocket({
      onPlansUpdate: handlePlansUpdate,
      onLogChunk: handleLogChunk,
    });

  const commonSession = sessions.find((s) => s.projectId === COMMON_TERMINAL_PROJECT_ID);

  // Auto-create common terminal session on mount (always create if not exists)
  useEffect(() => {
    if (connected && !commonTerminalId) {
      // Check if common terminal already exists
      const existingCommon = sessions.find((s) => s.projectId === COMMON_TERMINAL_PROJECT_ID);
      if (existingCommon) {
        setCommonTerminalId(existingCommon.id);
      } else {
        // Always create common terminal session if not exists
        createSession(COMMON_TERMINAL_PROJECT_ID);
      }
    }
  }, [connected, commonTerminalId, sessions, createSession]);

  // Track common terminal ID when session is created
  useEffect(() => {
    const commonSession = sessions.find((s) => s.projectId === COMMON_TERMINAL_PROJECT_ID);
    if (commonSession && commonTerminalId !== commonSession.id) {
      setCommonTerminalId(commonSession.id);
    }
  }, [sessions, commonTerminalId]);

  const handleCommonTerminalInput = useCallback(
    (data: string) => {
      if (commonTerminalId) {
        sendInput(commonTerminalId, data);
      }
    },
    [commonTerminalId, sendInput]
  );

  const handleCommandSend = useCallback(
    (sessionId: string, data: string) => {
      if (!sessionId) return;
      sendInput(sessionId, data);
      setFocusedTerminal(sessionId);
      setPage('work');
    },
    [sendInput]
  );

  const handleTerminalFocus = useCallback((sessionId: string) => {
    setFocusedTerminal(sessionId);
    setPage('work');
  }, []);

  // Wrapper for createSession that includes active project ID
  const handleCreateSession = useCallback(
    (worktreePath?: string) => {
      const projectId = activeProject?.id || 'default';
      createSession(projectId, worktreePath);
    },
    [createSession, activeProject]
  );

  // Filter sessions by active project
  const projectSessions = sessions.filter(
    (s) => !activeProject || s.projectId === activeProject.id
  );

  return (
    <div className="app">
      <ProjectSelector
        projects={projects}
        activeProjectId={activeProject?.id || null}
        onActivate={activateProject}
        onAdd={addProject}
        onRemove={removeProject}
      />

      <nav className="nav">
        <button
          className={`nav-btn ${page === 'dashboard' ? 'active' : ''}`}
          onClick={() => setPage('dashboard')}
        >
          Dashboard
        </button>
        <button
          className={`nav-btn ${page === 'work' ? 'active' : ''}`}
          onClick={() => setPage('work')}
        >
          Work
        </button>
        <button
          className={`nav-btn ${page === 'settings' ? 'active' : ''}`}
          onClick={() => setPage('settings')}
        >
          Settings
        </button>
        <span style={{ marginLeft: 'auto', color: connected ? '#22c55e' : '#ef4444' }}>
          {connected ? 'Connected' : 'Disconnected'}
        </span>
      </nav>

      <main className="main">
        {page === 'dashboard' && (
          <Dashboard
            plans={plans}
            sessions={projectSessions}
            worktreePaths={activeProject?.worktreePaths}
            contextWindow={contextWindow}
            onTerminalFocus={handleTerminalFocus}
            onCreateSession={handleCreateSession}
            onSendInput={handleCommandSend}
            commonTerminalId={commonTerminalId || undefined}
            onCommonTerminalInput={handleCommonTerminalInput}
            commonTerminalLogs={commonSession?.logs}
            commonTerminalOutput={commonTerminalOutput}
          />
        )}
        {page === 'work' && (
          <WorkPage
            plans={plans}
            sessions={projectSessions}
            worktreePaths={activeProject?.worktreePaths}
            focusedTerminal={focusedTerminal}
            onTerminalFocus={setFocusedTerminal}
            onSendInput={sendInput}
            onCreateSession={handleCreateSession}
            onDestroySession={destroySession}
            onResize={resizeTerminal}
          />
        )}
        {page === 'settings' && <SettingsPage />}
      </main>
    </div>
  );
}
