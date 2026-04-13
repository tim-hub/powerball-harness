#!/usr/bin/env node
/**
 * pre-compact-save.js
 * PreCompact hook for saving critical session context before compaction
 *
 * Usage (from hooks.json):
 *   "command": "node \"${CLAUDE_PLUGIN_ROOT}/scripts/hook-handlers/pre-compact-save.js\""
 *
 * Environment variables:
 *   CLAUDE_SESSION_ID - Current session ID
 *
 * Output:
 *   Saves handoff-artifact.json and precompact-snapshot.json to .claude/state/
 */

const fs = require('fs');
const path = require('path');

// Configuration
const ARTIFACT_VERSION = '2.0.0';
const LEGACY_SNAPSHOT_VERSION = '1.0.0';
const GIT_TIMEOUT_MS = 5000;
const HANDOFF_ARTIFACT_FILENAME = 'handoff-artifact.json';
const LEGACY_SNAPSHOT_FILENAME = 'precompact-snapshot.json';

/**
 * Log to stderr (non-blocking)
 * @param {string} message - Message to log
 */
function log(message) {
  process.stderr.write(`[pre-compact-save] ${message}\n`);
}

/**
 * Find repository root by walking up directory tree
 * @returns {string} Repository root path or cwd if not found
 */
function findRepoRoot() {
  let dir = process.cwd();
  const fsRoot = path.parse(dir).root;

  while (dir !== fsRoot) {
    if (fs.existsSync(path.join(dir, '.git'))) {
      return dir;
    }
    dir = path.dirname(dir);
  }
  return process.cwd();
}

/**
 * Get current timestamp in ISO8601 UTC format
 * @returns {string} ISO8601 timestamp
 */
function getTimestamp() {
  return new Date().toISOString();
}

/**
 * Read JSON from file if available.
 * @param {string} filePath - JSON file path
 * @returns {object|null} Parsed JSON or null
 */
function readJsonFile(filePath) {
  try {
    if (!fs.existsSync(filePath)) {
      return null;
    }
    return JSON.parse(fs.readFileSync(filePath, 'utf8'));
  } catch (err) {
    log(`Error reading ${path.basename(filePath)}: ${err.message}`);
    return null;
  }
}

/**
 * Read Plans.md and extract WIP tasks
 * @param {string} repoRoot - Repository root path
 * @returns {string[]} Array of WIP task descriptions
 */
function getPlanRows(repoRoot) {
  const plansPath = path.join(repoRoot, 'Plans.md');
  const planRows = [];

  try {
    if (!fs.existsSync(plansPath)) {
      return planRows;
    }

    const content = fs.readFileSync(plansPath, 'utf8');
    const lines = content.split('\n');

    for (const line of lines) {
      if (!line.includes('|')) {
        continue;
      }

      // Replace escaped `\|` with a placeholder before parsing, then restore after
      const PIPE_PH = '\x00PIPE\x00';
      const escaped = line.replace(/\\\|/g, PIPE_PH);
      const rawCells = escaped.split('|').map((cell) => cell.trim());
      const firstCell = rawCells[0] === '' ? 1 : 0;
      const lastCell = rawCells[rawCells.length - 1] === '' ? rawCells.length - 1 : rawCells.length;
      const cells = rawCells.slice(firstCell, lastCell);

      if (cells.length < 5) {
        continue;
      }

      const restore = (s) => s.replace(new RegExp(PIPE_PH, 'g'), '|');
      // Split from the right by fixed column count (handles rows where title or DoD contains `|`)
      const taskId = restore(cells[0]);
      const status = restore(cells[cells.length - 1]);
      const depends = restore(cells[cells.length - 2]);
      const middleParts = cells.slice(1, cells.length - 2);
      const title = restore(middleParts.length > 0 ? middleParts[0] : '');
      const dod = restore(middleParts.length > 1 ? middleParts.slice(1).join('|') : '');
      if (!taskId || taskId === 'Task' || /---+/.test(taskId)) {
        continue;
      }

      const normalizedStatus = status || '';
      const isTodo = /`?cc:TODO`?/i.test(normalizedStatus);
      const isWip = /`?cc:WIP`?/i.test(normalizedStatus) || /\[in_progress\]/i.test(normalizedStatus);
      const isBlocked = /`?cc:blocked`?/i.test(normalizedStatus) || /\[blocked\]/i.test(normalizedStatus);

      if (!isTodo && !isWip && !isBlocked) {
        continue;
      }

      planRows.push({
        taskId: taskId.trim(),
        title: title.trim(),
        dod: dod.trim(),
        depends: depends.trim(),
        status: status.trim(),
        tags: {
          todo: isTodo,
          wip: isWip,
          blocked: isBlocked
        }
      });
    }
  } catch (err) {
    log(`Error reading Plans.md: ${err.message}`);
  }

  return planRows;
}

/**
 * Return WIP task titles for legacy compatibility.
 * @param {Array<object>} planRows - Parsed plan rows
 * @returns {string[]} WIP task titles
 */
function getWipTasks(planRows) {
  return planRows.map((row) => row.title).filter(Boolean);
}

/**
 * Get recently modified files from git with timeout
 * @param {string} repoRoot - Repository root path
 * @returns {string[]} Array of recently modified file paths
 */
function getRecentEdits(repoRoot) {
  const recentEdits = [];

  try {
    const { execFileSync } = require('child_process');

    // Get files modified in working tree (staged + unstaged + untracked)
    // Reflects the current working state rather than HEAD~5
    const staged = execFileSync('git', ['diff', '--name-only', '--cached'], {
      cwd: repoRoot, encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'], timeout: GIT_TIMEOUT_MS
    }).trim();
    const unstaged = execFileSync('git', ['diff', '--name-only'], {
      cwd: repoRoot, encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'], timeout: GIT_TIMEOUT_MS
    }).trim();
    const untracked = execFileSync('git', ['ls-files', '--others', '--exclude-standard'], {
      cwd: repoRoot, encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'], timeout: GIT_TIMEOUT_MS
    }).trim();

    const allFiles = [staged, unstaged, untracked].filter(Boolean).join('\n');
    if (allFiles) {
      const unique = [...new Set(allFiles.split('\n'))].slice(0, 20);
      recentEdits.push(...unique);
    }
  } catch (err) {
    // git command may fail if no commits or shallow repo
    // Fallback: try unstaged changes only
    try {
      const { execFileSync } = require('child_process');
      const output = execFileSync('git', ['diff', '--name-only'], {
        cwd: repoRoot,
        encoding: 'utf8',
        stdio: ['pipe', 'pipe', 'pipe'],
        timeout: GIT_TIMEOUT_MS
      }).trim();

      if (output) {
        recentEdits.push(...output.split('\n').slice(0, 20));
      }
    } catch {
      log(`Error getting recent edits: ${err.message}`);
    }
  }

  return recentEdits;
}

/**
 * Get session metrics from state
 * @param {string} repoRoot - Repository root path
 * @returns {object|null} Session metrics or null
 */
function getSessionMetrics(repoRoot) {
  const metricsPath = path.join(repoRoot, '.claude', 'state', 'session-metrics.json');

  return readJsonFile(metricsPath);
}

/**
 * Read work state (work-active.json / ultrawork-active.json) if present
 * @param {string} repoRoot - Repository root path
 * @returns {object|null} Active work state
 */
function getWorkState(repoRoot) {
  const stateDir = path.join(repoRoot, '.claude', 'state');
  const candidates = [
    path.join(stateDir, 'work-active.json'),
    path.join(stateDir, 'ultrawork-active.json')
  ];

  for (const candidate of candidates) {
    const parsed = readJsonFile(candidate);
    if (parsed) {
      return parsed;
    }
  }

  return null;
}

/**
 * Read the current session state if present
 * @param {string} repoRoot - Repository root path
 * @returns {object|null} Session state
 */
function getSessionState(repoRoot) {
  const sessionPath = path.join(repoRoot, '.claude', 'state', 'session.json');
  return readJsonFile(sessionPath);
}

function parsePositiveInt(value, fallback) {
  const parsed = Number.parseInt(String(value ?? ''), 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

function getContextResetPolicy() {
  return {
    mode: process.env.HARNESS_CONTEXT_RESET_MODE || 'auto',
    dryRun: /^(1|true|yes|on)$/i.test(process.env.HARNESS_CONTEXT_RESET_DRY_RUN || ''),
    thresholds: {
      wipTasks: parsePositiveInt(process.env.HARNESS_CONTEXT_RESET_WIP_THRESHOLD, 4),
      blockedTasks: parsePositiveInt(process.env.HARNESS_CONTEXT_RESET_BLOCKED_THRESHOLD, 1),
      recentEdits: parsePositiveInt(process.env.HARNESS_CONTEXT_RESET_RECENT_EDITS_THRESHOLD, 8),
      failedChecks: parsePositiveInt(process.env.HARNESS_CONTEXT_RESET_FAILED_CHECKS_THRESHOLD, 1),
      sessionAgeMinutes: parsePositiveInt(process.env.HARNESS_CONTEXT_RESET_AGE_MINUTES, 120)
    }
  };
}

function buildContextResetRecommendation(planRows, recentEdits, workState, metrics, sessionState) {
  const policy = getContextResetPolicy();
  const wipCount = planRows.filter((row) => row.tags?.wip).length;
  const blockedCount = planRows.filter((row) => row.tags?.blocked).length;
  const failureCount =
    workState?.failed_checks?.length ||
    workState?.failedChecks?.length ||
    workState?.failures?.length ||
    metrics?.failed_checks?.length ||
    metrics?.failedChecks?.length ||
    metrics?.failures?.length ||
    metrics?.failure_count ||
    metrics?.failed_count ||
    0;
  let sessionAgeMinutes = null;
  if (sessionState && typeof sessionState === 'object' && sessionState.started_at) {
    const startedAt = Date.parse(sessionState.started_at);
    if (Number.isFinite(startedAt)) {
      sessionAgeMinutes = Math.max(0, Math.floor((Date.now() - startedAt) / 60000));
    }
  }

  const candidates = [
    {
      key: 'wip_tasks',
      label: 'WIP task count',
      actual: wipCount,
      threshold: policy.thresholds.wipTasks
    },
    {
      key: 'blocked_tasks',
      label: 'blocked task count',
      actual: blockedCount,
      threshold: policy.thresholds.blockedTasks
    },
    {
      key: 'recent_edits',
      label: 'recent edit count',
      actual: recentEdits.length,
      threshold: policy.thresholds.recentEdits
    },
    {
      key: 'failed_checks',
      label: 'failed check count',
      actual: failureCount,
      threshold: policy.thresholds.failedChecks
    },
    {
      key: 'session_age_minutes',
      label: 'session age (minutes)',
      actual: sessionAgeMinutes ?? 0,
      threshold: policy.thresholds.sessionAgeMinutes
    }
  ];

  const reasons = [];
  for (const candidate of candidates) {
    if (candidate.actual >= candidate.threshold) {
      reasons.push(`${candidate.actual} ${candidate.label} exceed threshold ${candidate.threshold}`);
      candidate.triggered = true;
    } else {
      candidate.triggered = false;
    }
  }

  const recommended = reasons.length > 0;
  const summary = recommended
    ? `Context reset recommended (${policy.mode}${policy.dryRun ? ', dry-run' : ''}): ${reasons.slice(0, 4).join('; ')}`
    : `Context reset not required (${policy.mode}${policy.dryRun ? ', dry-run' : ''})`;

  return {
    policy,
    recommended,
    summary,
    reasons,
    candidates,
    counters: {
      wipTasks: wipCount,
      blockedTasks: blockedCount,
      recentEdits: recentEdits.length,
      failedChecks: failureCount,
      sessionAgeMinutes
    }
  };
}

function buildContinuityContext(sessionState, nextAction) {
  const effortHint = process.env.HARNESS_EFFORT_DEFAULT || 'medium';
  const activeSkill = sessionState?.active_skill || sessionState?.activeSkill || null;
  const resumeToken = sessionState?.resume_token || null;
  const pluginFirst = true;

  const summary = [
    `plugin-first workflow: ${pluginFirst ? 'enabled' : 'disabled'}`,
    `resume-aware effort continuity: ${effortHint}`,
    activeSkill ? `active_skill=${activeSkill}` : null,
    resumeToken ? 'resume_token present' : null,
    nextAction?.taskId ? `next_task=${nextAction.taskId}` : null
  ].filter(Boolean).join('; ');

  return {
    plugin_first_workflow: pluginFirst,
    resume_aware_effort_continuity: true,
    effort_hint: effortHint,
    active_skill: activeSkill,
    summary
  };
}

/**
 * Pick the highest-priority next action from plan rows.
 * @param {Array<object>} planRows - Parsed plan rows
 * @returns {object|null} Next action candidate
 */
function pickNextAction(planRows) {
  const priorityRows = planRows.filter((row) => row.tags?.wip || row.tags?.todo || row.tags?.blocked);
  if (priorityRows.length === 0) {
    return null;
  }

  const preferred = priorityRows.find((row) => row.tags?.wip)
    || priorityRows.find((row) => row.tags?.todo)
    || priorityRows[0];

  return {
    taskId: preferred.taskId,
    task: preferred.title,
    dod: preferred.dod,
    depends: preferred.depends,
    status: preferred.status,
    source: 'Plans.md',
    priority: preferred.tags?.blocked ? 'blocked' : preferred.tags?.wip ? 'high' : 'normal',
    summary: `Continue ${preferred.taskId} ${preferred.title}`.trim()
  };
}

/**
 * Build structured open risks from the current state.
 * @param {Array<object>} planRows - Parsed plan rows
 * @param {Array<string>} recentEdits - Recently modified files
 * @param {object|null} workState - Active work state
 * @param {object|null} metrics - Session metrics
 * @returns {Array<object>} Risk entries
 */
function buildOpenRisks(planRows, recentEdits, workState, metrics) {
  const risks = [];
  const activeCount = planRows.length;
  const blockedCount = planRows.filter((row) => row.tags?.blocked).length;
  const wipCount = planRows.filter((row) => row.tags?.wip).length;

  if (wipCount > 0) {
    risks.push({
      severity: 'medium',
      kind: 'continuity',
      summary: `${wipCount} WIP task(s) remain in Plans.md`,
      detail: planRows.filter((row) => row.tags?.wip).slice(0, 5).map((row) => `${row.taskId} ${row.title}`).join('; ')
    });
  }

  if (blockedCount > 0) {
    risks.push({
      severity: 'high',
      kind: 'dependency',
      summary: `${blockedCount} blocked task(s) need attention before finish`,
      detail: planRows.filter((row) => row.tags?.blocked).slice(0, 5).map((row) => `${row.taskId} ${row.title}`).join('; ')
    });
  }

  if (recentEdits.length > 0) {
    risks.push({
      severity: 'medium',
      kind: 'verification',
      summary: `${recentEdits.length} recent edit(s) should be re-validated after resume`,
      detail: recentEdits.slice(0, 5).join(', ')
    });
  }

  if (workState && typeof workState === 'object') {
    const reviewStatus = workState.review_status || workState.reviewStatus;
    if (reviewStatus === 'failed') {
      risks.push({
        severity: 'high',
        kind: 'review',
        summary: 'work review_status is failed',
        detail: workState.last_failure || workState.failure_reason || workState.reason || 'The active work state needs repair before completion.'
      });
    } else if (reviewStatus && reviewStatus !== 'passed') {
      risks.push({
        severity: 'medium',
        kind: 'review',
        summary: `work review_status is ${reviewStatus}`,
        detail: 'Independent review is still required before finalizing the work.'
      });
    }
  }

  const failureCount =
    metrics?.failed_checks?.length ||
    metrics?.failedChecks?.length ||
    metrics?.failures?.length ||
    metrics?.failure_count ||
    metrics?.failed_count ||
    0;

  if (failureCount > 0) {
    risks.push({
      severity: 'high',
      kind: 'quality',
      summary: `${failureCount} recorded failed check(s) in session metrics`,
      detail: 'Review the latest validation results before resuming work.'
    });
  }

  if (activeCount > 0 && risks.length === 0) {
    risks.push({
      severity: 'low',
      kind: 'continuity',
      summary: 'Open plan items still exist and should be re-read after compaction',
      detail: `${activeCount} plan row(s) captured from Plans.md`
    });
  }

  return risks.slice(0, 8);
}

/**
 * Normalize failed checks from known active state fields.
 * @param {object|null} workState - Active work state
 * @param {object|null} metrics - Session metrics
 * @returns {Array<object>} Failed check entries
 */
function buildFailedChecks(workState, metrics) {
  const failedChecks = [];

  const pushEntries = (source, value) => {
    if (!value) {
      return;
    }
    const values = Array.isArray(value) ? value : [value];
    for (const item of values) {
      if (!item) {
        continue;
      }
      if (typeof item === 'string') {
        failedChecks.push({
          source,
          check: item,
          status: 'failed'
        });
        continue;
      }
      if (typeof item === 'object') {
        failedChecks.push({
          source,
          check: item.check || item.name || item.type || 'unknown',
          status: item.status || 'failed',
          detail: item.detail || item.message || item.reason || item.description || ''
        });
      }
    }
  };

  if (workState && typeof workState === 'object') {
    pushEntries('work-active.json', workState.failed_checks || workState.failedChecks || workState.failures || workState.checks_failed);

    const reviewStatus = workState.review_status || workState.reviewStatus;
    if (reviewStatus === 'failed' && failedChecks.length === 0) {
      failedChecks.push({
        source: 'work-active.json',
        check: 'review_status',
        status: 'failed',
        detail: workState.last_failure || workState.failure_reason || 'Active work state reported a failed review.'
      });
    }
  }

  if (metrics && typeof metrics === 'object') {
    pushEntries('session-metrics.json', metrics.failed_checks || metrics.failedChecks || metrics.failures);
  }

  return failedChecks.slice(0, 8);
}

/**
 * Build a concise decision log for the handoff artifact.
 * @param {string} timestamp - Current timestamp
 * @param {object|null} nextAction - Next action entry
 * @param {object|null} workState - Active work state
 * @returns {Array<object>} Decision log entries
 */
function buildDecisionLog(timestamp, nextAction, workState) {
  const logEntries = [
    {
      timestamp,
      actor: 'pre-compact-save',
      decision: 'canonical_handoff_artifact_written',
      rationale: 'Persist a stable JSON artifact in .claude/state for long-running session handoff.'
    },
    {
      timestamp,
      actor: 'pre-compact-save',
      decision: 'legacy_snapshot_mirrored',
      rationale: 'Keep precompact-snapshot.json for backward compatibility with older hooks.'
    }
  ];

  if (nextAction) {
    logEntries.push({
      timestamp,
      actor: 'pre-compact-save',
      decision: 'next_action_selected',
      rationale: `${nextAction.summary}${nextAction.source ? ` (source: ${nextAction.source})` : ''}`
    });
  }

  if (workState && typeof workState === 'object') {
    const reviewStatus = workState.review_status || workState.reviewStatus;
    if (reviewStatus) {
      logEntries.push({
        timestamp,
        actor: 'pre-compact-save',
        decision: 'active_work_status_captured',
        rationale: `work review_status=${reviewStatus}`
      });
    }
  }

  return logEntries.slice(0, 6);
}

/**
 * Build the structured handoff artifact payload.
 * @param {string} repoRoot - Repository root path
 * @param {string} sessionId - Current session ID
 * @param {string} timestamp - Current timestamp
 * @returns {object} Structured handoff artifact
 */
function buildHandoffArtifact(repoRoot, sessionId, timestamp) {
  const planRows = getPlanRows(repoRoot);
  const wipTasks = getWipTasks(planRows);
  const recentEdits = getRecentEdits(repoRoot);
  const metrics = getSessionMetrics(repoRoot);
  const workState = getWorkState(repoRoot);
  const sessionState = getSessionState(repoRoot);
  const nextAction = pickNextAction(planRows);
  const openRisks = buildOpenRisks(planRows, recentEdits, workState, metrics);
  const failedChecks = buildFailedChecks(workState, metrics);
  const decisionLog = buildDecisionLog(timestamp, nextAction, workState);
  const contextReset = buildContextResetRecommendation(planRows, recentEdits, workState, metrics, sessionState);
  const continuity = buildContinuityContext(sessionState, nextAction);
  const activePlanCount = planRows.length;
  const wipCount = planRows.filter((row) => row.tags?.wip).length;
  const blockedCount = planRows.filter((row) => row.tags?.blocked).length;
  const summaryParts = [];

  if (wipCount > 0) {
    summaryParts.push(`${wipCount} WIP`);
  }
  if (blockedCount > 0) {
    summaryParts.push(`${blockedCount} blocked`);
  }
  if (recentEdits.length > 0) {
    summaryParts.push(`${recentEdits.length} recent edit(s)`);
  }

  const previousStateSummary = summaryParts.length > 0
    ? `Before compaction: ${summaryParts.join(', ')}`
    : 'Before compaction: no active WIP tasks detected';

  return {
    version: ARTIFACT_VERSION,
    legacy_version: LEGACY_SNAPSHOT_VERSION,
    artifactType: 'structured-handoff',
    timestamp,
    sessionId,
    previous_state: {
      summary: previousStateSummary,
      session_state: sessionState
        ? {
            state: sessionState.state || 'unknown',
            resumed_at: sessionState.resumed_at || null,
            active_skill: sessionState.active_skill || null,
            review_status: workState?.review_status || workState?.reviewStatus || null
          }
        : null,
      plan_counts: {
        total: activePlanCount,
        wip: wipCount,
        blocked: blockedCount,
        recent_edits: recentEdits.length
      }
    },
    next_action: nextAction || {
      summary: 'Re-read Plans.md and determine the next task',
      taskId: null,
      task: null,
      dod: null,
      depends: null,
      status: null,
      source: 'fallback',
      priority: 'normal'
    },
    open_risks: openRisks,
    failed_checks: failedChecks,
    decision_log: decisionLog,
    context_reset: contextReset,
    continuity,
    planItems: planRows,
    wipTasks,
    recentEdits,
    metrics
  };
}

/**
 * Main function
 */
function main() {
  const sessionId = process.env.CLAUDE_SESSION_ID || '';
  const repoRoot = findRepoRoot();
  const timestamp = getTimestamp();

  const claudeDir = path.join(repoRoot, '.claude');
  const stateDir = path.join(claudeDir, 'state');
  const artifactPath = path.join(stateDir, HANDOFF_ARTIFACT_FILENAME);
  const snapshotPath = path.join(stateDir, LEGACY_SNAPSHOT_FILENAME);

  try {
    // Resolve repo root for security checks
    const resolvedRepo = fs.realpathSync(repoRoot);

    // Security: Verify .claude is not a symlink pointing outside repo
    if (fs.existsSync(claudeDir)) {
      const claudeDirStat = fs.lstatSync(claudeDir);
      if (claudeDirStat.isSymbolicLink()) {
        const resolvedClaudeDir = fs.realpathSync(claudeDir);
        if (!resolvedClaudeDir.startsWith(resolvedRepo + path.sep) &&
            resolvedClaudeDir !== resolvedRepo) {
          log('.claude symlink points outside repo');
          console.log(JSON.stringify({ continue: true, message: 'Skipped: security check failed' }));
          return;
        }
      }
    }

    // Ensure state directory exists with restricted permissions
    if (fs.existsSync(stateDir)) {
      // Security: Verify stateDir is not a symlink
      const stateDirStat = fs.lstatSync(stateDir);
      if (stateDirStat.isSymbolicLink()) {
        log('stateDir is a symlink, refusing to write');
        console.log(JSON.stringify({ continue: true, message: 'Skipped: stateDir is symlink' }));
        return;
      }
      // Ensure permissions are restricted
      fs.chmodSync(stateDir, 0o700);
    } else {
      fs.mkdirSync(stateDir, { recursive: true, mode: 0o700 });
    }

    // Security: Verify artifact path is not a symlink
    if (fs.existsSync(artifactPath)) {
      const artifactStat = fs.lstatSync(artifactPath);
      if (artifactStat.isSymbolicLink()) {
        log('artifactPath is a symlink, refusing to write');
        console.log(JSON.stringify({ continue: true, message: 'Skipped: artifactPath is symlink' }));
        return;
      }
    }

    // Security: Verify snapshot path is not a symlink
    if (fs.existsSync(snapshotPath)) {
      const snapshotStat = fs.lstatSync(snapshotPath);
      if (snapshotStat.isSymbolicLink()) {
        log('snapshotPath is a symlink, refusing to write');
        console.log(JSON.stringify({ continue: true, message: 'Skipped: snapshotPath is symlink' }));
        return;
      }
    }

    const artifact = buildHandoffArtifact(repoRoot, sessionId, timestamp);
    const snapshot = {
      ...artifact,
      version: LEGACY_SNAPSHOT_VERSION,
      artifactType: 'precompact-snapshot',
      wipTasks: artifact.wipTasks,
      recentEdits: artifact.recentEdits,
      metrics: artifact.metrics,
      context_reset: artifact.context_reset,
      continuity: artifact.continuity
    };

    // Save the canonical structured handoff artifact and the legacy snapshot alias.
    fs.writeFileSync(artifactPath, JSON.stringify(artifact, null, 2), { mode: 0o600 });
    fs.writeFileSync(snapshotPath, JSON.stringify(snapshot, null, 2), { mode: 0o600 });

    // Output for hook feedback
    const result = {
      continue: true,
      message: `Saved structured handoff artifact: ${artifact.wipTasks.length} WIP tasks, ${artifact.recentEdits.length} recent edits`
    };

    console.log(JSON.stringify(result));
  } catch (err) {
    log(`Error saving snapshot: ${err.message}`);
    // Don't block compaction on errors
    console.log(JSON.stringify({ continue: true, message: `Error: ${err.message}` }));
  }
}

main();
