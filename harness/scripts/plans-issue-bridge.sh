#!/usr/bin/env bash
# plans-issue-bridge.sh
# Generate a dry-run payload from plans.json for GitHub Issue integration.
# SSOT: .claude/harness/plans.json

set -euo pipefail

# Source plans.json helpers (for default path resolution)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "${SCRIPT_DIR}/config-utils.sh" ]; then
  # shellcheck source=./config-utils.sh
  source "${SCRIPT_DIR}/config-utils.sh"
fi

usage() {
  cat >&2 <<'EOF'
Usage: scripts/plans-issue-bridge.sh [--plans PATH] [--format json|markdown] [--team-mode] [--output PATH]
  --plans PATH   Path to plans.json (default: .claude/harness/plans.json)
EOF
  exit 1
}

# Default plans.json path (SSOT)
if declare -F get_plans_json_path >/dev/null 2>&1; then
  PLANS_FILE="$(get_plans_json_path)"
else
  PLANS_FILE=".claude/harness/plans.json"
fi
PLANS_FILE_DEFAULT="$PLANS_FILE"
FORMAT="json"
TEAM_MODE="false"
OUTPUT_FILE=""

while [ $# -gt 0 ]; do
  case "$1" in
    --plans)
      PLANS_FILE="${2:-}"
      shift 2
      ;;
    --format)
      FORMAT="${2:-}"
      shift 2
      ;;
    --team-mode)
      TEAM_MODE="true"
      shift
      ;;
    --no-team-mode)
      TEAM_MODE="false"
      shift
      ;;
    --output)
      OUTPUT_FILE="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      ;;
    *)
      if [ -z "$PLANS_FILE" ] || [ "$PLANS_FILE" = "$PLANS_FILE_DEFAULT" ]; then
        PLANS_FILE="$1"
        shift
      else
        usage
      fi
      ;;
  esac
done

if [ ! -f "$PLANS_FILE" ]; then
  echo "plans.json not found: $PLANS_FILE" >&2
  exit 2
fi

node - "$PLANS_FILE" "$FORMAT" "$TEAM_MODE" "$OUTPUT_FILE" <<'NODE'
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

process.stdout.on('error', (error) => {
  if (error.code === 'EPIPE') process.exit(0);
  throw error;
});

const [plansFile, format, teamModeFlag, outputFile] = process.argv.slice(2);
const teamMode = teamModeFlag === 'true';

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, 'utf8'));
}

function repoNameFrom(plansPath) {
  try {
    const gitRoot = execSync('git rev-parse --show-toplevel', {
      cwd: path.dirname(plansPath),
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    }).trim();
    if (gitRoot) return path.basename(gitRoot);
  } catch {
    // fall back to the directory name that contains plans.json
  }
  return path.basename(path.resolve(path.dirname(plansPath)));
}

function normalizeDepends(value) {
  if (Array.isArray(value)) {
    return value.map((item) => String(item).trim()).filter(Boolean);
  }
  if (!value || value === '-') return [];
  return String(value).split(',').map((item) => item.trim()).filter(Boolean);
}

// Parse the plans.json SSOT shape:
//   { "phases": [ { "id", "title", "status", "tasks": [ { "id", "name", "dod",
//     "depends": [...], "status", ... } ] } ] }
function parsePlans(plansData) {
  const phases = [];
  const tasks = [];

  const rawPhases = Array.isArray(plansData && plansData.phases) ? plansData.phases : [];

  for (const rawPhase of rawPhases) {
    const currentPhase = {
      id: String(rawPhase.id ?? '').trim(),
      title: String(rawPhase.title ?? '').trim(),
    };
    const phaseEntry = { ...currentPhase, tasks: [] };
    phases.push(phaseEntry);

    const rawTasks = Array.isArray(rawPhase.tasks) ? rawPhase.tasks : [];
    for (const rawTask of rawTasks) {
      const dependsOn = normalizeDepends(rawTask.depends);
      const task = {
        id: String(rawTask.id ?? '').trim(),
        content: String(rawTask.name ?? rawTask.content ?? ''),
        dod: String(rawTask.dod ?? ''),
        depends: dependsOn.join(', '),
        depends_on: dependsOn,
        status: String(rawTask.status ?? ''),
        phase: { ...currentPhase },
      };

      tasks.push(task);
      phaseEntry.tasks.push(task);
    }
  }

  return { phases, tasks };
}

function countStatuses(tasks) {
  const counts = {};
  for (const task of tasks) {
    counts[task.status] = (counts[task.status] || 0) + 1;
  }
  return counts;
}

function escapeCell(value) {
  return String(value ?? '').replace(/\|/g, '\\|');
}

function renderTable(rows) {
  const header = ['Task', 'Content', 'DoD', 'Depends', 'Status'];
  const body = [header, ...rows.map((row) => [
    row.id,
    row.content,
    row.dod,
    row.depends,
    row.status,
  ])];
  return [
    `| ${body[0].map(escapeCell).join(' | ')} |`,
    `|${body[0].map(() => '---').join('|')}|`,
    ...body.slice(1).map((row) => `| ${row.map(escapeCell).join(' | ')} |`),
  ].join('\n');
}

function buildReport(plansPath, teamModeEnabled) {
  const repoName = repoNameFrom(plansPath);
  const plansData = readJson(plansPath);
  const parsed = parsePlans(plansData);
  const summary = {
    phase_count: parsed.phases.length,
    task_count: parsed.tasks.length,
    status_counts: countStatuses(parsed.tasks),
  };

  const subIssues = parsed.tasks.map((task) => ({
    title: `${task.id} ${task.content}`,
    body: [
      `Task: ${task.id}`,
      `Phase: ${task.phase ? `${task.phase.id}: ${task.phase.title}` : 'unassigned'}`,
      `DoD: ${task.dod}`,
      `Depends: ${task.depends}`,
      `Status: ${task.status}`,
      '',
      'This is a dry-run payload. plans.json remains the source of truth.',
    ].join('\n'),
    labels: ['harness-plan', `phase-${(task.phase && String(task.phase.id).split('.')[0]) || 'unassigned'}`],
    depends_on: task.depends_on,
    phase: task.phase,
    task_id: task.id,
    status: task.status,
  }));

  const trackingIssue = {
    title: `plans.json tracking issue: ${repoName}`,
    body: [
      '# plans.json tracking issue dry-run',
      '',
      `- Repo: ${repoName}`,
      `- Plans file: ${path.relative(process.cwd(), plansPath)}`,
      `- Team mode: ${teamModeEnabled ? 'enabled' : 'disabled (opt-in required)'}`,
      `- Sub-issues: ${subIssues.length}`,
      '',
      '## Phase overview',
      ...parsed.phases.map((phase) => `- ${phase.id}: ${phase.title} (${phase.tasks.length} tasks)`),
      '',
      '## Task snapshot',
      renderTable(parsed.tasks),
      '',
      'plans.json is the source of truth. This payload is preview-only and does not mutate the plan.',
    ].join('\n'),
    labels: ['harness-plan', 'plans-sync', teamModeEnabled ? 'team-mode' : 'solo-preview'],
    state: 'dry-run',
  };

  return {
    schema_version: 'plans-issue-bridge.v1',
    generated_at: new Date().toISOString(),
    source: {
      repo_name: repoName,
      plans_file: path.resolve(plansPath),
    },
    team_mode: {
      enabled: teamModeEnabled,
      opt_in_required: true,
    },
    tracking_issue: trackingIssue,
    phases: parsed.phases,
    sub_issues: subIssues,
    summary,
  };
}

function renderMarkdown(report) {
  const lines = [
    '# plans.json issue bridge dry-run',
    '',
    `- Repo: ${report.source.repo_name}`,
    `- Plans file: ${report.source.plans_file}`,
    `- Team mode: ${report.team_mode.enabled ? 'enabled' : 'disabled (opt-in required)'}`,
    `- Tracking issue: ${report.tracking_issue.title}`,
    `- Sub-issues: ${report.sub_issues.length}`,
    '',
    '## Summary',
    `- Phases: ${report.summary.phase_count}`,
    `- Tasks: ${report.summary.task_count}`,
    `- Status counts: ${Object.entries(report.summary.status_counts).map(([k, v]) => `${k}=${v}`).join(', ') || 'none'}`,
    '',
    '## Tracking issue body',
    report.tracking_issue.body,
    '',
    '## Sub-issues',
  ];

  for (const subIssue of report.sub_issues) {
    lines.push(`- ${subIssue.title}`);
    lines.push(`  - Depends: ${subIssue.depends_on.length > 0 ? subIssue.depends_on.join(', ') : '-'}`);
    lines.push(`  - Status: ${subIssue.status}`);
  }

  lines.push('');
  lines.push('plans.json remains the source of truth.');
  return `${lines.join('\n')}\n`;
}

const report = buildReport(plansFile, teamMode);
const output = format === 'markdown'
  ? renderMarkdown(report)
  : `${JSON.stringify(report, null, 2)}\n`;

if (outputFile) {
  fs.mkdirSync(path.dirname(outputFile), { recursive: true });
  fs.writeFileSync(outputFile, output);
}

process.stdout.write(output);
NODE
