#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

function usage() {
  console.error('Usage: scripts/generate-sprint-contract.sh <task-id> [plans-file] [output-file]');
  process.exit(1);
}

const taskId = process.argv[2];
if (!taskId) usage();

const repoRoot = spawnSync('git', ['rev-parse', '--show-toplevel'], { encoding: 'utf8' }).stdout.trim();
const plansFile = process.argv[3] ? path.resolve(process.argv[3]) : path.join(repoRoot, '.claude', 'harness', 'plans.json');
const defaultOut = path.join(repoRoot, '.claude', 'state', 'contracts', `${taskId}.sprint-contract.json`);
const outputFile = process.argv[4] ? path.resolve(process.argv[4]) : defaultOut;

if (!fs.existsSync(plansFile)) {
  console.error(`plans.json not found: ${plansFile}`);
  process.exit(2);
}

// Look up a task by id in plans.json. Returns the shape the rest of this script
// expects: { taskId, title, dod, depends (comma string), status }.
// qualityMarkers are folded into the title text as [marker] tokens so the
// existing text-based profile / risk-flag detectors keep working unchanged.
function findTaskInPlans(plansJsonText, targetTaskId) {
  let data;
  try {
    data = JSON.parse(plansJsonText);
  } catch (e) {
    console.error(`failed to parse plans.json: ${e.message}`);
    process.exit(2);
  }
  for (const phase of data.phases || []) {
    for (const task of phase.tasks || []) {
      if (task.id !== targetTaskId) continue;
      const markers = Array.isArray(task.qualityMarkers) ? task.qualityMarkers : [];
      const markerText = markers.map((m) => `[${m}]`).join(' ');
      const name = task.name || '';
      return {
        taskId: task.id,
        title: markerText ? `${name} ${markerText}` : name,
        dod: task.dod || '',
        depends: Array.isArray(task.depends) ? task.depends.join(',') : '',
        status: task.status || '',
      };
    }
  }
  return null;
}

function toList(value) {
  if (!value || value === '-') return [];
  return value.split(',').map((item) => item.trim()).filter(Boolean);
}

function detectProfile(task) {
  const text = `${task.title} ${task.dod}`.toLowerCase();
  if (/(browser|chrome|playwright|\bui\b|layout|responsive|screenshot|screen|web app|webapp)/.test(text)) {
    return 'browser';
  }
  if (/(runtime|typecheck|lint|test|api|probe|integration|e2e|validation command)/.test(text)) {
    return 'runtime';
  }
  return 'static';
}

function detectBrowserMode(task) {
  const text = `${task.title} ${task.dod}`.toLowerCase();
  if (/(browser_mode\s*:\s*exploratory|\bexploratory\b|exploratory mode|exploratory)/.test(text)) {
    return 'exploratory';
  }
  if (/(browser_mode\s*:\s*scripted|\bscripted\b|scripted|fixed)/.test(text)) {
    return 'scripted';
  }
  return 'scripted';
}

function detectRiskFlags(task) {
  const text = `${task.title} ${task.dod}`.toLowerCase();
  const flags = [];
  if (/\[needs-spike\]/.test(task.title) || /\[needs-spike\]/.test(task.dod)) flags.push('needs-spike');
  if (/(security|auth|permission|secret|guardrail)/.test(text)) flags.push('security-sensitive');
  if (/(migration|schema|state|resume|session|artifact)/.test(text)) flags.push('state-migration');
  if (/(browser|ui|layout|responsive|playwright|chrome|screen|layout)/.test(text)) flags.push('ux-regression');
  return [...new Set(flags)];
}

function hasCommand(command) {
  const result = spawnSync('bash', ['-lc', `command -v ${JSON.stringify(command)} >/dev/null 2>&1`], {
    stdio: 'ignore',
  });
  return result.status === 0;
}

function hasPlaywrightBasis(root) {
  if (process.env.HARNESS_BROWSER_REVIEW_DISABLE_PLAYWRIGHT) {
    return false;
  }
  const packageJsonPath = path.join(root, 'package.json');
  if (fs.existsSync(packageJsonPath)) {
    try {
      const pkg = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
      const scripts = pkg.scripts || {};
      const deps = { ...(pkg.dependencies || {}), ...(pkg.devDependencies || {}) };
      if (scripts['test:e2e'] || deps.playwright || deps['@playwright/test']) {
        return true;
      }
    } catch {
      // ignore parse failures here; runtime command generation reports them separately
    }
  }

  // Do not rely on global CLI (`command -v playwright`).
  // Detect only via repo-based check (package.json deps) for deterministic results.
  return false;
}

function hasAgentBrowser() {
  if (process.env.HARNESS_BROWSER_REVIEW_DISABLE_AGENT_BROWSER) {
    return false;
  }
  // Do not rely on global CLI (prevents environment-dependent contract changes)
  // agent-browser is detected at generate-browser-review-artifact.sh execution time
  return false;
}

function detectExplicitBrowserRoute(task) {
  const text = `${task.title}\n${task.dod}`;
  const match = text.match(/(?:browser_)?route\s*:\s*(playwright|agent-browser|chrome-devtools)/i);
  return match ? match[1].toLowerCase() : null;
}

function detectBrowserRoute(task, root, browserMode) {
  // Only bake explicit route specifications from the task into the contract.
  // Otherwise return null and let generate-browser-review-artifact.sh
  // determine the route at runtime (eliminates environment dependency from contract).
  const explicitRoute = detectExplicitBrowserRoute(task);
  if (explicitRoute) {
    return explicitRoute;
  }

  // exploratory mode resolves at runtime (agent-browser preferred), so don't bake it in
  if (browserMode === 'exploratory') return null;

  // scripted mode: bake in repo-based detection result
  // (package.json deps are environment-independent → deterministic)
  if (hasPlaywrightBasis(root)) return 'playwright';

  // Still unresolved — return null (resolved at artifact generation time)
  return null;
}

function pickRuntimeCommands(root) {
  const commands = [];
  const packageJsonPath = path.join(root, 'package.json');
  if (fs.existsSync(packageJsonPath)) {
    try {
      const pkg = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
      const scripts = pkg.scripts || {};
      // Suppress watch mode with CI=true (Jest/Vitest compatible)
      if (scripts.test) commands.push({ label: 'package-test', command: 'CI=true npm test' });
      if (scripts.lint) commands.push({ label: 'package-lint', command: 'npm run lint' });
      if (scripts.typecheck) commands.push({ label: 'package-typecheck', command: 'npm run typecheck' });
      if (scripts['test:e2e']) commands.push({ label: 'package-e2e', command: 'npm run test:e2e' });
    } catch (error) {
      // On parse failure, exit 1 explicitly (prevent runtime gate bypass)
      commands.push({ label: 'package-parse-error', command: `echo "ERROR: package.json parse failed: ${error.message.replace(/"/g, '\\"')}" >&2; exit 1` });
    }
  }

  const fallbackChecks = [
    { marker: 'pnpm-lock.yaml', label: 'pnpm-test', command: 'pnpm test' },
    { marker: 'bun.lock', label: 'bun-test', command: 'bun test' },
    { marker: 'go.mod', label: 'go-test', command: 'go test ./...' },
    { marker: 'Cargo.toml', label: 'cargo-test', command: 'cargo test' },
  ];

  for (const check of fallbackChecks) {
    if (commands.length > 0) break;
    if (fs.existsSync(path.join(root, check.marker))) {
      commands.push({ label: check.label, command: check.command });
    }
  }

  // shell-repo fallback: use validate-plugin.sh / check-consistency.sh if no package.json etc.
  if (commands.length === 0) {
    const shellFallbacks = [
      { path: 'tests/validate-plugin.sh', label: 'validate-plugin', command: './tests/validate-plugin.sh' },
      { path: 'local-scripts/check-consistency.sh', label: 'check-consistency', command: './local-scripts/check-consistency.sh' },
    ];
    for (const fb of shellFallbacks) {
      if (fs.existsSync(path.join(root, fb.path))) {
        commands.push({ label: fb.label, command: fb.command });
      }
    }
  }

  return commands;
}

const plansJsonText = fs.readFileSync(plansFile, 'utf8');
const row = findTaskInPlans(plansJsonText, taskId);
if (!row) {
  console.error(`Task not found in plans.json: ${taskId}`);
  process.exit(3);
}

const reviewerProfile = detectProfile(row);
const browserMode = reviewerProfile === 'browser' ? detectBrowserMode(row) : null;
const browserRoute = reviewerProfile === 'browser' ? detectBrowserRoute(row, repoRoot, browserMode) : null;
const runtimeValidation = reviewerProfile === 'runtime' ? pickRuntimeCommands(repoRoot) : [];
const riskFlags = detectRiskFlags(row);

const contract = {
  schema_version: 'sprint-contract.v1',
  generated_at: new Date().toISOString(),
  source: {
    plans_file: path.relative(repoRoot, plansFile) || 'plans.json',
    task_id: row.taskId,
  },
  task: {
    id: row.taskId,
    title: row.title,
    definition_of_done: row.dod,
    depends_on: toList(row.depends),
    status_at_generation: row.status,
  },
  contract: {
    checks: [
      {
        id: 'dod-primary',
        source: 'plans.json.dod',
        description: row.dod,
      },
    ],
    non_goals: [],
    runtime_validation: runtimeValidation,
    browser_validation: reviewerProfile === 'browser'
      ? [
          {
            id: 'browser-smoke',
            description: row.dod,
            required_artifacts: browserMode === 'exploratory'
              ? ['snapshot', 'ui-flow-log']
              : ['trace', 'screenshot', 'ui-flow-log'],
          },
        ]
      : [],
    risk_flags: riskFlags,
  },
  review: {
    status: 'draft',
    reviewer_profile: reviewerProfile,
    browser_mode: browserMode,
    route: browserRoute,
    reviewer_notes: [],
    approved_at: null,
    gaps: [],
    followups: [],
  },
};

fs.mkdirSync(path.dirname(outputFile), { recursive: true });
fs.writeFileSync(outputFile, `${JSON.stringify(contract, null, 2)}\n`);
console.log(outputFile);
