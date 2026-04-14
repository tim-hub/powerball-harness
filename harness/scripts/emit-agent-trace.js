#!/usr/bin/env node
/**
 * emit-agent-trace.js
 * PostToolUse hook for recording Agent Trace entries
 *
 * Usage (from hooks.json):
 *   "command": "node \"${CLAUDE_PLUGIN_ROOT}/scripts/emit-agent-trace.js\""
 *
 * Environment variables (set by Claude Code hooks):
 *   CLAUDE_TOOL_NAME - Name of the tool (Edit, Write, etc.)
 *   CLAUDE_TOOL_INPUT - JSON string of tool input
 *   CLAUDE_SESSION_ID - Current session ID
 *
 * OTel environment variables (optional):
 *   OTEL_EXPORTER_OTLP_ENDPOINT - If set, emit OTel Span JSON via HTTP POST to this endpoint
 *     Example: http://localhost:4318 (OTLP HTTP receiver default)
 *     The spans are posted to ${OTEL_EXPORTER_OTLP_ENDPOINT}/v1/traces
 *
 * Output:
 *   Appends JSONL record to .claude/state/agent-trace.jsonl
 *   When OTEL_EXPORTER_OTLP_ENDPOINT is set: also POSTs OTel Span JSON (non-blocking, 3s timeout)
 */

const fs = require('fs');
const os = require('os');
const path = require('path');
const { execFileSync, spawn } = require('child_process');
const crypto = require('crypto');

// Configuration
const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
const MAX_GENERATIONS = 3;
const TRACE_VERSION = '0.3.0';
const CACHE_TTL_MS = 60000; // 1 minute cache for project info

// In-memory cache for project metadata (persists across hook invocations within same process)
let projectCache = null;
let projectCacheTime = 0;

/**
 * Log error to stderr (non-blocking)
 */
function logError(context, error) {
  const msg = error instanceof Error ? error.message : String(error);
  process.stderr.write(`[agent-trace] ${context}: ${msg}\n`);
}

/**
 * Generate UUID v4
 */
function generateUUID() {
  return crypto.randomUUID();
}

/**
 * Get current timestamp in ISO8601 UTC format
 */
function getTimestamp() {
  return new Date().toISOString();
}

// Repo root cache (key: cwd, value: repo root)
let repoRootCache = null;
let repoRootCwd = null;

/**
 * Find repository root by walking up directory tree
 * Avoids git process spawn for better performance
 * Returns cwd if not in a git repo
 */
function findRepoRoot() {
  const cwd = process.cwd();

  // Return cached value if cwd hasn't changed
  if (repoRootCache && repoRootCwd === cwd) {
    return repoRootCache;
  }

  try {
    let dir = cwd;
    const fsRoot = path.parse(dir).root;

    while (dir !== fsRoot) {
      const gitDir = path.join(dir, '.git');
      if (fs.existsSync(gitDir)) {
        // Verify .git is a directory (not a file for submodules)
        const stat = fs.lstatSync(gitDir);
        if (stat.isDirectory() || stat.isFile()) {
          // .git file indicates submodule, still valid
          repoRootCache = dir;
          repoRootCwd = cwd;
          return dir;
        }
      }
      dir = path.dirname(dir);
    }
  } catch (err) {
    logError('findRepoRoot', err);
  }

  // Not in a git repo, use cwd
  repoRootCache = cwd;
  repoRootCwd = cwd;
  return cwd;
}

// VCS cache (persists within process, refreshed by TTL)
let vcsCache = null;
let vcsCacheTime = 0;
const VCS_CACHE_TTL_MS = 5000; // 5 seconds cache for VCS info

/**
 * Get VCS (Git) information with single git call for performance
 * Uses `git status --porcelain=2 -b -uno` to get branch, revision, and dirty status
 * -uno: Exclude untracked files for much faster execution on large repos
 * Returns null if not in a Git repository
 */
function getVcsInfo() {
  const now = Date.now();

  // Return cached if still valid (dirty status may change but branch/revision rarely do)
  if (vcsCache && (now - vcsCacheTime) < VCS_CACHE_TTL_MS) {
    return vcsCache;
  }

  try {
    // Single git call to get all VCS info
    // -uno: Skip untracked files for performance
    const output = execFileSync('git', [
      'status', '--porcelain=2', '-b', '-uno'
    ], {
      encoding: 'utf8',
      stdio: ['pipe', 'pipe', 'pipe']
    }).trim();

    const lines = output.split('\n');
    let revision = '';
    let branch = '';
    let dirty = false;

    for (const line of lines) {
      if (line.startsWith('# branch.oid ')) {
        revision = line.slice(13);
      } else if (line.startsWith('# branch.head ')) {
        branch = line.slice(14);
      } else if (line && !line.startsWith('#')) {
        // Any non-header line indicates modified files (untracked excluded by -uno)
        dirty = true;
      }
    }

    if (!revision || !branch) {
      return null;
    }

    vcsCache = { revision, branch, dirty };
    vcsCacheTime = now;
    return vcsCache;
  } catch (err) {
    logError('getVcsInfo', err);
    return null;
  }
}

/**
 * Detect project type from file patterns
 * Uses existsSync with early termination for better performance
 * Avoids reading entire directory on large monorepos
 */
function detectProjectType(repoRoot) {
  // Priority-ordered list: more specific frameworks first, generic last
  const checks = [
    ['next.config.js', 'nextjs'],
    ['next.config.ts', 'nextjs'],
    ['next.config.mjs', 'nextjs'],
    ['nuxt.config.js', 'nuxt'],
    ['nuxt.config.ts', 'nuxt'],
    ['svelte.config.js', 'svelte'],
    ['astro.config.mjs', 'astro'],
    ['astro.config.ts', 'astro'],
    ['Cargo.toml', 'rust'],
    ['go.mod', 'go'],
    ['pyproject.toml', 'python'],
    ['setup.py', 'python'],
    ['requirements.txt', 'python'],
    ['Gemfile', 'ruby'],
    ['composer.json', 'php'],
    ['package.json', 'node'],
  ];

  try {
    for (const [file, type] of checks) {
      if (fs.existsSync(path.join(repoRoot, file))) {
        return type;
      }
    }
  } catch (err) {
    logError('detectProjectType', err);
  }

  return 'unknown';
}

/**
 * Get project name from package.json or directory name
 */
function getProjectName(repoRoot) {
  try {
    const pkgPath = path.join(repoRoot, 'package.json');
    if (fs.existsSync(pkgPath)) {
      const pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf8'));
      if (pkg.name) return pkg.name;
    }
  } catch (err) {
    logError('getProjectName', err);
  }
  return path.basename(repoRoot);
}

/**
 * Get cached project metadata or compute it
 */
function getProjectMetadata(repoRoot) {
  const now = Date.now();

  // Return cached if still valid
  if (projectCache && (now - projectCacheTime) < CACHE_TTL_MS) {
    return projectCache;
  }

  // Compute and cache
  projectCache = {
    project: getProjectName(repoRoot),
    projectType: detectProjectType(repoRoot)
  };
  projectCacheTime = now;

  return projectCache;
}

// Attribution cache
let attributionCache = null;
let attributionCacheTime = 0;

/**
 * Get attribution information from marketplace.json
 * v0.2.0: Added for tracking AI-generated code provenance
 */
function getAttribution() {
  const now = Date.now();

  // Return cached if still valid
  if (attributionCache && (now - attributionCacheTime) < CACHE_TTL_MS) {
    return attributionCache;
  }

  try {
    const pluginRoot = process.env.CLAUDE_PLUGIN_ROOT;
    if (!pluginRoot) {
      return null;
    }

    const pluginJsonPath = path.join(pluginRoot, 'marketplace.json');
    if (!fs.existsSync(pluginJsonPath)) {
      return null;
    }

    const pluginJson = JSON.parse(fs.readFileSync(pluginJsonPath, 'utf8'));
    const pluginEntry = (pluginJson.plugins && pluginJson.plugins[0]) || {};

    attributionCache = {
      plugin: pluginEntry.name || pluginJson.name || 'unknown',
      version: pluginJson.version || 'unknown',
      license: pluginEntry.license || null,
      author: pluginEntry.author || pluginJson.owner || null
    };
    attributionCacheTime = now;

    return attributionCache;
  } catch (err) {
    logError('getAttribution', err);
    return null;
  }
}

/**
 * Extract metrics from Task tool result
 * v0.3.0: Added for Claude Code 2.1.30+ Task tool metrics
 */
function extractTaskMetrics(toolResult) {
  try {
    if (!toolResult) return null;

    const result = typeof toolResult === 'string' ? JSON.parse(toolResult) : toolResult;

    if (result.metrics) {
      return {
        tokenCount: result.metrics.tokenCount ?? null,
        toolUses: result.metrics.toolUses ?? null,
        duration: result.metrics.duration ?? null
      };
    }
  } catch (err) {
    logError('extractTaskMetrics', err);
  }

  return null;
}

/**
 * Normalize an agent/subagent name to a harness role.
 */
function normalizeAgentRole(name) {
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

/**
 * Validate that a path is within the repository
 * Security: Prevents recording paths outside the repo
 * Handles both existing and non-existing paths safely
 */
function isPathWithinRepo(filePath, repoRoot) {
  try {
    // Reject paths with '..' to prevent traversal
    if (filePath.includes('..')) {
      logError('isPathWithinRepo', `Path contains '..': ${filePath}`);
      return false;
    }

    // Resolve to absolute path
    const absolutePath = path.isAbsolute(filePath)
      ? filePath
      : path.resolve(repoRoot, filePath);

    // Resolve repo root (must succeed)
    const resolvedRepo = fs.realpathSync(repoRoot);

    // Use realpath to resolve symlinks
    let resolvedPath;
    try {
      resolvedPath = fs.realpathSync(absolutePath);
    } catch {
      // File doesn't exist yet (Write operation)
      // Security: Resolve parent directory if it exists, then append basename
      const parentDir = path.dirname(absolutePath);
      const baseName = path.basename(absolutePath);

      try {
        // Check if parent is a symlink pointing outside repo
        const resolvedParent = fs.realpathSync(parentDir);
        if (!resolvedParent.startsWith(resolvedRepo + path.sep) &&
            resolvedParent !== resolvedRepo) {
          logError('isPathWithinRepo', `Parent dir outside repo: ${parentDir}`);
          return false;
        }
        resolvedPath = path.join(resolvedParent, baseName);
      } catch {
        // Parent doesn't exist either - use normalized path but be strict
        resolvedPath = path.normalize(absolutePath);
        // Additional check: normalized path should not escape repo
        const relativePath = path.relative(resolvedRepo, resolvedPath);
        if (relativePath.startsWith('..') || path.isAbsolute(relativePath)) {
          logError('isPathWithinRepo', `Normalized path escapes repo: ${resolvedPath}`);
          return false;
        }
      }
    }

    // Check if path starts with repo root (with separator to avoid /repo vs /repo2)
    return resolvedPath.startsWith(resolvedRepo + path.sep) ||
           resolvedPath === resolvedRepo;
  } catch (err) {
    logError('isPathWithinRepo', err);
    return false;
  }
}

/**
 * Parse tool input to extract file information
 */
function parseToolInput(toolName, toolInput, repoRoot) {
  const files = [];

  try {
    const input = typeof toolInput === 'string' ? JSON.parse(toolInput) : toolInput;

    if (toolName === 'Edit') {
      if (input.file_path) {
        // Security: Validate path is within repo
        if (isPathWithinRepo(input.file_path, repoRoot)) {
          files.push({
            path: input.file_path,
            action: 'modify',
            range: 'unknown'
          });
        } else {
          logError('parseToolInput', `Path outside repo: ${input.file_path}`);
        }
      }
    } else if (toolName === 'Write') {
      if (input.file_path) {
        // Security: Validate path is within repo
        if (isPathWithinRepo(input.file_path, repoRoot)) {
          // Use absolute path for exists check to avoid cwd dependency
          const absolutePath = path.isAbsolute(input.file_path)
            ? input.file_path
            : path.resolve(repoRoot, input.file_path);
          const exists = fs.existsSync(absolutePath);
          files.push({
            path: input.file_path,
            action: exists ? 'modify' : 'create',
            range: 'unknown'
          });
        } else {
          logError('parseToolInput', `Path outside repo: ${input.file_path}`);
        }
      }
    }
  } catch (err) {
    logError('parseToolInput', err);
  }

  return files;
}

/**
 * Make path relative to repo root
 */
function makeRelativePath(filePath, repoRoot) {
  if (path.isAbsolute(filePath)) {
    return path.relative(repoRoot, filePath);
  }
  return filePath;
}

/**
 * Rotate trace file if it exceeds size limit
 * Uses lock file to prevent concurrent rotation
 */
function rotateIfNeeded(tracePath) {
  // Skip rotation if file doesn't exist
  if (!fs.existsSync(tracePath)) return;

  try {
    const stats = fs.statSync(tracePath);
    if (stats.size < MAX_FILE_SIZE) return;

    // Use lock file to prevent concurrent rotation
    const lockPath = `${tracePath}.lock`;
    let lockFd;
    try {
      // O_CREAT | O_EXCL equivalent: fails if lock exists
      lockFd = fs.openSync(lockPath, 'wx');
    } catch {
      // Lock exists, another process is rotating - skip rotation
      return;
    }

    try {
      // Re-check size after acquiring lock (another process may have rotated)
      if (!fs.existsSync(tracePath)) {
        return;
      }
      const currentStats = fs.statSync(tracePath);
      if (currentStats.size < MAX_FILE_SIZE) {
        return;
      }

      for (let i = MAX_GENERATIONS - 1; i >= 1; i--) {
        const oldPath = `${tracePath}.${i}`;
        const newPath = `${tracePath}.${i + 1}`;
        if (fs.existsSync(oldPath)) {
          if (i === MAX_GENERATIONS - 1) {
            fs.unlinkSync(oldPath);
          } else {
            fs.renameSync(oldPath, newPath);
          }
        }
      }
      fs.renameSync(tracePath, `${tracePath}.1`);
    } finally {
      // Release lock
      fs.closeSync(lockFd);
      try { fs.unlinkSync(lockPath); } catch { /* ignore */ }
    }
  } catch (err) {
    logError('rotateIfNeeded', err);
  }
}

/**
 * Convert a JSONL trace record to OTel Span JSON (OTLP HTTP format).
 *
 * Time handling:
 *   - record.timestamp is ISO8601 (end of the tool call)
 *   - OTel requires startTimeUnixNano / endTimeUnixNano in nanoseconds (as strings to avoid
 *     JS integer precision loss on 64-bit values).
 *   - We use record.timestamp as both start and end time (single-point span) because the
 *     hook only fires at PostToolUse and we don't have a reliable start time.
 *
 * traceId / spanId:
 *   - OTel traceId: 32 hex chars (128-bit), spanId: 16 hex chars (64-bit).
 *   - We derive them from record.id (UUID v4) to keep them deterministic per record.
 */
function buildOtelSpanJson(record, serviceVersion) {
  const endMs = new Date(record.timestamp).getTime();
  const endNano = String(endMs) + '000000'; // ms → ns as string

  // Derive traceId (32 hex) and spanId (16 hex) from record UUID
  const uuidHex = record.id.replace(/-/g, '');              // 32 hex chars
  const traceId = uuidHex;                                   // 32 hex = 128-bit trace ID
  const spanId = uuidHex.slice(0, 16);                       // first 16 hex = 64-bit span ID

  // Build span attributes from available record fields
  const attributes = [];

  const taskId = record.metadata && record.metadata.taskId;
  if (taskId) {
    attributes.push({ key: 'task.id', value: { stringValue: String(taskId) } });
  }

  const agentRole = record.metadata && record.metadata.agentRole;
  if (agentRole) {
    attributes.push({ key: 'agent.type', value: { stringValue: String(agentRole) } });
  }

  const effort = record.metadata && record.metadata.effort;
  if (effort) {
    attributes.push({ key: 'effort', value: { stringValue: String(effort) } });
  }

  // tool name is always present
  attributes.push({ key: 'tool.name', value: { stringValue: record.tool } });

  if (record.vcs && record.vcs.branch) {
    attributes.push({ key: 'vcs.branch', value: { stringValue: record.vcs.branch } });
  }

  if (record.metadata && record.metadata.sessionId) {
    attributes.push({ key: 'session.id', value: { stringValue: record.metadata.sessionId } });
  }

  const spanName = agentRole
    ? `harness.${agentRole}`
    : `harness.${record.tool.toLowerCase()}`;

  return {
    resourceSpans: [{
      resource: {
        attributes: [
          { key: 'service.name',    value: { stringValue: 'claude-code-harness' } },
          { key: 'service.version', value: { stringValue: serviceVersion } }
        ]
      },
      scopeSpans: [{
        scope: { name: 'harness.agent' },
        spans: [{
          traceId,
          spanId,
          name: spanName,
          kind: 1,              // SPAN_KIND_INTERNAL
          startTimeUnixNano: endNano,
          endTimeUnixNano:   endNano,
          attributes
        }]
      }]
    }]
  };
}

/**
 * Emit OTel Span JSON to OTLP HTTP endpoint via curl (non-blocking, 3s timeout).
 * Fires and forgets: failures are logged to stderr but do not block the hook.
 *
 * @param {string} otlpEndpoint - Base URL, e.g. "http://localhost:4318"
 * @param {object} record       - JSONL trace record
 * @param {string} serviceVersion
 */
function emitOtelSpan(otlpEndpoint, record, serviceVersion) {
  try {
    const spanJson = buildOtelSpanJson(record, serviceVersion);
    const body = JSON.stringify(spanJson);

    // POST to /v1/traces (OTLP HTTP traces endpoint)
    const url = otlpEndpoint.replace(/\/$/, '') + '/v1/traces';

    // Spawn curl as a fully detached fire-and-forget child process.
    // All stdio is ignored so no pipe listeners keep the Node event loop alive.
    // --silent --max-time 3: hard timeout of 3 seconds, no progress output.
    // --fail: return exit code 22 on HTTP errors (4xx/5xx).
    // --write-out: append status to a log file so failures are diagnosable.
    const otelLogFile = path.join(os.tmpdir(), 'harness-otel-export.log');
    const child = spawn('sh', [
      '-c',
      `HTTP_CODE=$(curl --silent --output /dev/null --write-out "%{http_code}" --max-time 3 --request POST --header "Content-Type: application/json" --data @- "${url}" <<'EOFBODY'\n${body}\nEOFBODY\n); ` +
      `if [ "$HTTP_CODE" -lt 200 ] || [ "$HTTP_CODE" -ge 300 ] 2>/dev/null; then ` +
      `printf "[%s] otel export failed: HTTP %s -> %s\\n" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$HTTP_CODE" "${url}" >> "${otelLogFile}"; fi`
    ], {
      stdio: ['ignore', 'ignore', 'ignore'],
      detached: true
    });

    child.on('error', (err) => {
      // sh/curl not found or spawn failed — log to stderr (non-blocking)
      process.stderr.write(`[agent-trace] otel spawn failed: ${err.message}\n`);
    });

    // Unref so this process can exit immediately without waiting for curl
    child.unref();
  } catch (err) {
    logError('emitOtelSpan', err);
  }
}

/**
 * Read service version from marketplace.json (or VERSION file as fallback).
 * Returns '0.0.0' on failure.
 */
function readServiceVersion() {
  try {
    const pluginRoot = process.env.CLAUDE_PLUGIN_ROOT;
    if (pluginRoot) {
      const pluginJsonPath = path.join(pluginRoot, 'marketplace.json');
      if (fs.existsSync(pluginJsonPath)) {
        const pkg = JSON.parse(fs.readFileSync(pluginJsonPath, 'utf8'));
        if (pkg.version) return String(pkg.version);
      }
      const versionPath = path.join(pluginRoot, 'VERSION');
      if (fs.existsSync(versionPath)) {
        return fs.readFileSync(versionPath, 'utf8').trim();
      }
    }
  } catch {
    // fall through to default
  }
  return '0.0.0';
}

/**
 * Main function
 */
function main() {
  const toolName = process.env.CLAUDE_TOOL_NAME;
  const toolInput = process.env.CLAUDE_TOOL_INPUT;
  const toolResult = process.env.CLAUDE_TOOL_RESULT;
  const sessionId = process.env.CLAUDE_SESSION_ID || '';

  if (!toolName || !['Edit', 'Write', 'Task'].includes(toolName)) {
    process.exit(0);
  }

  const repoRoot = findRepoRoot();
  const files = parseToolInput(toolName, toolInput, repoRoot);

  if (files.length === 0 && toolName !== 'Task') {
    process.exit(0);
  }

  files.forEach(f => {
    f.path = makeRelativePath(f.path, repoRoot);
  });

  const record = {
    version: TRACE_VERSION,
    id: generateUUID(),
    timestamp: getTimestamp(),
    tool: toolName,
    files: files
  };

  const vcs = getVcsInfo();
  if (vcs) {
    record.vcs = vcs;
  }

  const metadata = getProjectMetadata(repoRoot);
  record.metadata = { ...metadata };
  if (sessionId) {
    record.metadata.sessionId = sessionId;
  }

  // v0.2.0: Add attribution for AI-generated code tracking
  const attribution = getAttribution();
  if (attribution) {
    record.attribution = attribution;
  }

  // v0.3.0: Add Task tool metrics (Claude Code 2.1.30+)
  if (toolName === 'Task') {
    const metrics = extractTaskMetrics(toolResult);
    if (metrics) {
      record.metrics = metrics;
    }

    // Extract taskId from tool input if available
    try {
      const input = typeof toolInput === 'string' ? JSON.parse(toolInput) : toolInput;
      if (input.task_id) {
        record.metadata.taskId = input.task_id;
      }
      if (input.subagent_type) {
        record.metadata.subagentType = input.subagent_type;
        record.metadata.agentRole = normalizeAgentRole(input.subagent_type);
      } else if (input.agent_name) {
        record.metadata.agentRole = normalizeAgentRole(input.agent_name);
      }
    } catch {
      // Ignore parsing errors
    }
  }

  const stateDir = path.join(repoRoot, '.claude', 'state');
  const tracePath = path.join(stateDir, 'agent-trace.jsonl');

  try {
    const claudeDir = path.join(repoRoot, '.claude');
    const resolvedRepo = fs.realpathSync(repoRoot);

    // Security: Verify .claude is not a symlink pointing outside repo
    if (fs.existsSync(claudeDir)) {
      const claudeDirStat = fs.lstatSync(claudeDir);
      if (claudeDirStat.isSymbolicLink()) {
        const resolvedClaudeDir = fs.realpathSync(claudeDir);
        if (!resolvedClaudeDir.startsWith(resolvedRepo + path.sep) &&
            resolvedClaudeDir !== resolvedRepo) {
          logError('main', '.claude symlink points outside repo');
          process.exit(0);
        }
      }
    }

    // Security: Verify stateDir BEFORE creating/modifying
    if (fs.existsSync(stateDir)) {
      // Existing stateDir: verify it's within repo BEFORE chmod
      const stateDirLstat = fs.lstatSync(stateDir);
      if (stateDirLstat.isSymbolicLink()) {
        logError('main', 'stateDir is a symlink, refusing to modify');
        process.exit(0);
      }
      const resolvedStateDir = fs.realpathSync(stateDir);
      if (!resolvedStateDir.startsWith(resolvedRepo + path.sep) &&
          resolvedStateDir !== resolvedRepo) {
        logError('main', 'stateDir resolves outside repo');
        process.exit(0);
      }
      // Now safe to chmod
      fs.chmodSync(stateDir, 0o700);
    } else {
      // Create directory with restricted permissions (owner only)
      fs.mkdirSync(stateDir, { recursive: true, mode: 0o700 });
    }

    // Security: Verify tracePath is not a symlink and is a regular file
    if (fs.existsSync(tracePath)) {
      const tracePathStat = fs.lstatSync(tracePath);
      if (tracePathStat.isSymbolicLink()) {
        logError('main', 'tracePath is a symlink, refusing to write');
        process.exit(0);
      }
      if (!tracePathStat.isFile()) {
        logError('main', 'tracePath is not a regular file');
        process.exit(0);
      }
    }

    rotateIfNeeded(tracePath);

    // Security: Write with restricted permissions (owner read/write only)
    const fd = fs.openSync(tracePath, 'a', 0o600);
    // Verify opened file is regular file (post-open check)
    const fdStat = fs.fstatSync(fd);
    if (!fdStat.isFile()) {
      fs.closeSync(fd);
      logError('main', 'opened fd is not a regular file');
      process.exit(0);
    }
    fs.fchmodSync(fd, 0o600);
    fs.writeSync(fd, JSON.stringify(record) + '\n');
    fs.closeSync(fd);
  } catch (err) {
    logError('main', err);
  }

  // OTel Span JSON export (optional, non-blocking).
  // Only fires when OTEL_EXPORTER_OTLP_ENDPOINT is configured.
  // Failures are silently logged to stderr; they never block the hook.
  const otlpEndpoint = process.env.OTEL_EXPORTER_OTLP_ENDPOINT;
  if (otlpEndpoint) {
    const serviceVersion = readServiceVersion();
    emitOtelSpan(otlpEndpoint, record, serviceVersion);
  }

  process.exit(0);
}

main();
