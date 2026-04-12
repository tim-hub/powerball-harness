#!/usr/bin/env node
/**
 * generate-agent-telemetry.js
 * Aggregate per-agent telemetry from statusline, trace, and usage artifacts.
 *
 * Inputs (default under .claude/state):
 * - statusline-telemetry.jsonl
 * - agent-trace.jsonl
 * - harness-usage.json
 * - session.events.jsonl
 *
 * Output:
 * - JSON report to stdout, or to --output when provided
 */

const fs = require('fs');
const path = require('path');

function parseArgs(argv) {
  const args = {
    stateDir: process.env.HARNESS_STATE_DIR || path.join(process.cwd(), '.claude', 'state'),
    output: ''
  };

  for (let i = 0; i < argv.length; i += 1) {
    const token = argv[i];
    if (token === '--state-dir' && argv[i + 1]) {
      args.stateDir = argv[i + 1];
      i += 1;
    } else if (token === '--output' && argv[i + 1]) {
      args.output = argv[i + 1];
      i += 1;
    }
  }

  return args;
}

function readJson(filePath) {
  try {
    if (!fs.existsSync(filePath)) {
      return null;
    }
    return JSON.parse(fs.readFileSync(filePath, 'utf8'));
  } catch {
    return null;
  }
}

function readJsonLines(filePath) {
  try {
    if (!fs.existsSync(filePath)) {
      return [];
    }

    return fs.readFileSync(filePath, 'utf8')
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
      .map((line) => {
        try {
          return JSON.parse(line);
        } catch {
          return null;
        }
      })
      .filter(Boolean);
  } catch {
    return [];
  }
}

function toNumber(value) {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

function normalizeRole(name) {
  const value = String(name || '').trim().toLowerCase();
  if (!value) {
    return 'unknown';
  }
  if (value.includes('review')) {
    return 'reviewer';
  }
  if (value.includes('lead') || value.includes('planner')) {
    return 'lead';
  }
  if (value.includes('worker') || value.includes('impl')) {
    return 'worker';
  }
  return value;
}

function createBucket(role) {
  return {
    role,
    statusline_count: 0,
    trace_count: 0,
    usage_count: 0,
    token_count: 0,
    duration_ms: 0,
    cost_usd: 0,
    retry_count: 0,
    artifact_count: 0,
    statusline_samples: [],
    trace_samples: [],
    agent_names: []
  };
}

function accumulate(map, role, mutator) {
  const normalized = normalizeRole(role);
  if (!map[normalized]) {
    map[normalized] = createBucket(normalized);
  }
  mutator(map[normalized]);
}

function mergeAgentNames(bucket, name) {
  const value = String(name || '').trim();
  if (!value) {
    return;
  }
  if (!bucket.agent_names.includes(value)) {
    bucket.agent_names.push(value);
  }
}

function incrementRetryCount(bucket, value) {
  const amount = toNumber(value);
  if (amount > 0) {
    bucket.retry_count += amount;
  }
}

function extractEventRole(entry) {
  return normalizeRole(
    entry?.role ||
    entry?.agent_role ||
    entry?.agent ||
    entry?.data?.role ||
    entry?.data?.agent_role ||
    entry?.data?.agent ||
    entry?.data?.subagent_type ||
    entry?.data?.subagentType
  );
}

function findWorktreeStateDirs(repoRoot) {
  // In Breezing, Worker/Reviewer run in separate git worktrees,
  // so telemetry is distributed across each worktree's .claude/state.
  // List all worktrees with git worktree list and collect state directories.
  const dirs = [];
  try {
    const { execFileSync } = require('child_process');
    const output = execFileSync('git', ['worktree', 'list', '--porcelain'], {
      cwd: repoRoot, encoding: 'utf8', timeout: 5000
    });
    for (const line of output.split('\n')) {
      if (line.startsWith('worktree ')) {
        const wtPath = line.replace('worktree ', '').trim();
        const wtState = path.join(wtPath, '.claude', 'state');
        if (fs.existsSync(wtState)) {
          dirs.push(wtState);
        }
      }
    }
  } catch {
    // Ignore if git worktree list fails
  }
  return dirs;
}

function buildReport(stateDir) {
  const statuslineFile = path.join(stateDir, 'statusline-telemetry.jsonl');
  const traceFile = path.join(stateDir, 'agent-trace.jsonl');
  const usageFile = path.join(stateDir, 'harness-usage.json');
  const eventsFile = path.join(stateDir, 'session.events.jsonl');

  // Also aggregate distributed worktree telemetry
  const repoRoot = path.resolve(stateDir, '..', '..');
  const extraStateDirs = findWorktreeStateDirs(repoRoot).filter((d) => d !== stateDir);

  const buckets = {};
  const roleSet = new Set(['worker', 'reviewer', 'lead']);

  // Combine main and worktree statuslines
  const allStatuslineFiles = [statuslineFile, ...extraStateDirs.map((d) => path.join(d, 'statusline-telemetry.jsonl'))];
  const statuslines = allStatuslineFiles.flatMap((f) => readJsonLines(f));
  for (const entry of statuslines) {
    const role = entry?.agent_name || entry?.role || 'unknown';
    const bucketRole = normalizeRole(role);
    if (!roleSet.has(bucketRole)) {
      continue;
    }
    accumulate(buckets, bucketRole, (bucket) => {
      bucket.statusline_count += 1;
      bucket.duration_ms += toNumber(entry.duration_ms);
      bucket.cost_usd += toNumber(entry.cost_usd);
      bucket.artifact_count += 1;
      mergeAgentNames(bucket, entry.agent_name);
      bucket.statusline_samples.push({
        timestamp: entry.timestamp || null,
        model: entry.model || null,
        context_used_percentage: entry.context_used_percentage ?? null
      });
    });
  }

  const allTraceFiles = [traceFile, ...extraStateDirs.map((d) => path.join(d, 'agent-trace.jsonl'))];
  const traces = allTraceFiles.flatMap((f) => readJsonLines(f));
  for (const record of traces) {
    if (record?.tool !== 'Task') {
      continue;
    }
    const role = normalizeRole(
      record?.metadata?.agentRole ||
      record?.metadata?.subagentType ||
      record?.metadata?.subagent_type ||
      record?.metadata?.role
    );
    if (!roleSet.has(role)) {
      continue;
    }
    accumulate(buckets, role, (bucket) => {
      const metrics = record.metrics || {};
      bucket.trace_count += 1;
      bucket.token_count += toNumber(metrics.tokenCount);
      // Duration may already be counted in statusline,
      // so skip duration addition from trace when statusline_count > 0
      // (statusline has finer sampling granularity and is more accurate)
      if (bucket.statusline_count === 0) {
        bucket.duration_ms += toNumber(metrics.duration);
      }
      bucket.artifact_count += Math.max(1, Array.isArray(record.files) ? record.files.length : 0);
      mergeAgentNames(bucket, record?.metadata?.subagentType || record?.metadata?.agentRole);
      bucket.trace_samples.push({
        timestamp: record.timestamp || null,
        taskId: record?.metadata?.taskId || null,
        tokenCount: metrics.tokenCount ?? null,
        duration: metrics.duration ?? null
      });
    });
  }

  // Combine main and worktree usage
  const allUsageFiles = [usageFile, ...extraStateDirs.map((d) => path.join(d, 'harness-usage.json'))];
  const mergedUsages = allUsageFiles.map((f) => readJson(f)).filter(Boolean);
  const usage = mergedUsages[0] || {};
  // Merge worktree usage into the first usage
  for (let i = 1; i < mergedUsages.length; i++) {
    const extra = mergedUsages[i];
    if (extra?.agents) {
      usage.agents = usage.agents || {};
      for (const [name, summary] of Object.entries(extra.agents)) {
        if (!usage.agents[name]) { usage.agents[name] = summary; }
        else {
          usage.agents[name].count = (usage.agents[name].count || 0) + (summary?.count || 0);
          usage.agents[name].retryCount = (usage.agents[name].retryCount || 0) + (summary?.retryCount || 0);
        }
      }
    }
  }
  const usageRoles = usage?.roles;
  if (Array.isArray(usageRoles) && usageRoles.length > 0) {
    for (const [role, summary] of usageRoles) {
      const normalized = normalizeRole(role);
      if (!roleSet.has(normalized)) {
        continue;
      }
      accumulate(buckets, normalized, (bucket) => {
        bucket.usage_count += toNumber(summary?.count);
        incrementRetryCount(bucket, summary?.retryCount);
        mergeAgentNames(bucket, summary?.agentNames?.join(', '));
      });
    }
  } else if (usage?.agents && typeof usage.agents === 'object') {
    for (const [name, summary] of Object.entries(usage.agents)) {
      const normalized = normalizeRole(name);
      if (!roleSet.has(normalized)) {
        continue;
      }
      accumulate(buckets, normalized, (bucket) => {
        bucket.usage_count += toNumber(summary?.count);
        incrementRetryCount(bucket, summary?.retryCount);
        mergeAgentNames(bucket, name);
      });
    }
  }

  // Combine main and worktree events
  const allEventsFiles = [eventsFile, ...extraStateDirs.map((d) => path.join(d, 'session.events.jsonl'))];
  const events = allEventsFiles.flatMap((f) => readJsonLines(f));
  let retryEvents = 0;
  for (const entry of events) {
    const type = String(entry?.type || '').toLowerCase();
    if (!type.includes('retry')) {
      continue;
    }
    retryEvents += 1;
    const role = extractEventRole(entry);
    if (!roleSet.has(role)) {
      continue;
    }
    accumulate(buckets, role, (bucket) => {
      incrementRetryCount(bucket, entry?.count ?? entry?.data?.count ?? 1);
      mergeAgentNames(bucket, entry?.agent || entry?.data?.agent || entry?.data?.subagentType || role);
    });
  }

  const roles = {};
  const totals = {
    statusline_count: 0,
    trace_count: 0,
    usage_count: 0,
    token_count: 0,
    duration_ms: 0,
    cost_usd: 0,
    retry_count: 0,
    session_retry_events: retryEvents,
    artifact_count: 0
  };

  for (const role of ['worker', 'reviewer', 'lead']) {
    const bucket = buckets[role] || createBucket(role);
    totals.statusline_count += bucket.statusline_count;
    totals.trace_count += bucket.trace_count;
    totals.usage_count += bucket.usage_count;
    totals.token_count += bucket.token_count;
    totals.duration_ms += bucket.duration_ms;
    totals.cost_usd += bucket.cost_usd;
    totals.retry_count += bucket.retry_count;
    totals.artifact_count += bucket.artifact_count;
    roles[role] = bucket;
  }

  return {
    version: '1.0.0',
    generated_at: new Date().toISOString(),
    state_dir: stateDir,
    sources: {
      statusline_telemetry: path.relative(process.cwd(), statuslineFile),
      agent_trace: path.relative(process.cwd(), traceFile),
      harness_usage: path.relative(process.cwd(), usageFile),
      session_events: path.relative(process.cwd(), eventsFile)
    },
    roles,
    totals
  };
}

function main() {
  const args = parseArgs(process.argv.slice(2));
  const report = buildReport(args.stateDir);
  const output = JSON.stringify(report, null, 2);

  if (args.output) {
    fs.mkdirSync(path.dirname(args.output), { recursive: true });
    fs.writeFileSync(args.output, output + '\n', 'utf8');
  } else {
    process.stdout.write(output + '\n');
  }
}

main();
