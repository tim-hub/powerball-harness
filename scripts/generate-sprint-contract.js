#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

function usage() {
  console.error('Usage: scripts/generate-sprint-contract.js <task-id> [plans-file] [output-file]');
  process.exit(1);
}

const taskId = process.argv[2];
if (!taskId) usage();

const repoRoot = process.cwd();
const plansFile = process.argv[3] ? path.resolve(process.argv[3]) : path.join(repoRoot, 'Plans.md');
const defaultOut = path.join(repoRoot, '.claude', 'state', 'contracts', `${taskId}.sprint-contract.json`);
const outputFile = process.argv[4] ? path.resolve(process.argv[4]) : defaultOut;

if (!fs.existsSync(plansFile)) {
  console.error(`Plans.md not found: ${plansFile}`);
  process.exit(2);
}

function parseTaskRow(markdown, targetTaskId) {
  const lines = markdown.split('\n');
  // エスケープされた pipe `\|` をプレースホルダに置換してパース後に復元する
  const PIPE_PLACEHOLDER = '\x00PIPE\x00';
  const escapePipes = (s) => s.replace(/\\\|/g, PIPE_PLACEHOLDER);
  const restorePipes = (s) => s.replace(new RegExp(PIPE_PLACEHOLDER, 'g'), '|');

  // ヘッダー行のカラム位置を検出して正確な分割に使う
  // Plans.md の想定カラム: | Task | 内容 | DoD | Depends | Status |
  // ヘッダーの区切り線 (|---|---|...) から物理カラム数を判定
  let headerColCount = 5; // デフォルト
  for (const line of lines) {
    const trimmed = line.trim();
    if (/^\|[\s-]+\|/.test(trimmed)) {
      const sepCols = trimmed.split('|').filter((c) => c.trim().length > 0);
      if (sepCols.length >= 5) headerColCount = sepCols.length;
      break;
    }
  }

  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed.startsWith('|')) continue;
    if (/^\|[\s-]+\|/.test(trimmed)) continue; // 区切り行をスキップ

    // エスケープされた `\|` を保護してから split
    const inner = escapePipes(trimmed.replace(/^\|/, '').replace(/\|$/, ''));
    const parts = inner.split('|');
    if (parts.length < 5) continue;

    const taskId = parts[0].trim();
    if (taskId !== targetTaskId) continue;

    // 右端2カラム（depends, status）は固定。`|` を含まない短い値。
    const status = parts[parts.length - 1].trim();
    const depends = parts[parts.length - 2].trim();

    // 残りの parts[1..n-3] を title と dod に分割。
    // ヘッダーカラム数から middle のカラム数（title + dod = headerColCount - 3）を判定
    const middleParts = parts.slice(1, parts.length - 2);
    const expectedMiddle = headerColCount - 3; // title + dod のカラム数（通常 2）

    let title, dod;
    if (expectedMiddle <= 1 || middleParts.length <= 1) {
      // カラム数が少ない or 余分な `|` がない → 単純分割
      title = middleParts[0] ? middleParts[0].trim() : '';
      dod = '';
    } else {
      // middleParts を半分に分割: 前半 = title, 後半 = dod
      // ただし正規テーブルでは `|` が1つだけ title/dod を区切るので
      // 余分な `|` は全て title 側に帰属させる（title の方が `|` を含みやすい）
      // dod は末尾1セル固定
      dod = middleParts[middleParts.length - 1].trim();
      title = middleParts.slice(0, middleParts.length - 1).join('|').trim();
    }
    return {
      taskId: restorePipes(taskId),
      title: restorePipes(title),
      dod: restorePipes(dod),
      depends: restorePipes(depends),
      status: restorePipes(status),
    };
  }
  return null;
}

function toList(value) {
  if (!value || value === '-') return [];
  return value.split(',').map((item) => item.trim()).filter(Boolean);
}

function detectProfile(task) {
  const text = `${task.title} ${task.dod}`.toLowerCase();
  const hasUiRubricHints = (
    /\bui-rubric\b|\bdesign\b|styling|aesthetic|visual polish|design-heavy|design quality|originality|craft|functionality|デザイン|見た目品質|意匠|質感|デザイン品質/.test(text) ||
    (/\bui\b/.test(text) && /(design|styling|aesthetic|layout|visual|polish|デザイン|見た目)/.test(text)) ||
    (/\blayout\b/.test(text) && /(design|styling|aesthetic|visual|polish|デザイン|見た目)/.test(text))
  );
  if (hasUiRubricHints) {
    return 'ui-rubric';
  }
  if (/(browser|chrome|playwright|\bui\b|layout|responsive|スクリーンショット|画面|web アプリ|webアプリ)/.test(text)) {
    return 'browser';
  }
  if (/(runtime|typecheck|lint|test|api|probe|integration|e2e|検証コマンド)/.test(text)) {
    return 'runtime';
  }
  return 'static';
}

// profile ごとの default max_iterations
const PROFILE_MAX_ITERATIONS = {
  static: 3,
  runtime: 3,
  browser: 5,
  'ui-rubric': 10,
};

const DEFAULT_UI_RUBRIC_TARGET = Object.freeze({
  design: 6,
  originality: 6,
  craft: 6,
  functionality: 6,
});

function detectMaxIterations(profile, task) {
  const profileDefaults = { static: 3, runtime: 3, browser: 5, 'ui-rubric': 10 };
  const defaultValue = profileDefaults[profile] ?? 3;

  // HTML コメント形式のマーカーのみを受け付ける:
  //   <!-- max_iterations: 15 -->
  // Markdown として表示されないため、例示テキストと区別可能（自己参照バグ防止）。
  // 素のテキスト「max_iterations: 15」は意図的に無視する。
  const text = `${task.title}\n${task.dod}`;
  const match = text.match(/<!--\s*max_iterations:\s*(\d+)\s*-->/i);
  if (match) {
    const value = parseInt(match[1], 10);
    if (value >= 1 && value <= 30) {
      return value;
    }
    process.stderr.write(
      `[warn] max_iterations=${value} out of range (1-30), falling back to default ${defaultValue}\n`
    );
  }
  return defaultValue;
}

function detectBrowserMode(task) {
  const text = `${task.title} ${task.dod}`.toLowerCase();
  if (/(browser_mode\s*:\s*exploratory|\bexploratory\b|探索モード|探索的)/.test(text)) {
    return 'exploratory';
  }
  if (/(browser_mode\s*:\s*scripted|\bscripted\b|定型|決め打ち)/.test(text)) {
    return 'scripted';
  }
  return 'scripted';
}

function detectRiskFlags(task) {
  const text = `${task.title} ${task.dod}`.toLowerCase();
  const flags = [];
  if (/\[needs-spike\]/.test(task.title) || /\[needs-spike\]/.test(task.dod)) flags.push('needs-spike');
  if (/(security|auth|permission|secret|guardrail|セキュリティ|権限)/.test(text)) flags.push('security-sensitive');
  if (/(migration|schema|state|resume|session|artifact|マイグレーション|セッション|再開)/.test(text)) flags.push('state-migration');
  if (/(browser|ui|layout|responsive|playwright|chrome|画面|レイアウト)/.test(text)) flags.push('ux-regression');
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

  // グローバル CLI (`command -v playwright`) には依存しない。
  // repo-based 検出（package.json の deps）のみで判定する。
  return false;
}

function hasAgentBrowser() {
  if (process.env.HARNESS_BROWSER_REVIEW_DISABLE_AGENT_BROWSER) {
    return false;
  }
  // グローバル CLI に依存しない（環境差で contract が変わることを防止）
  // agent-browser は generate-browser-review-artifact.sh 実行時に検出する
  return false;
}

function detectExplicitBrowserRoute(task) {
  const text = `${task.title}\n${task.dod}`;
  const match = text.match(/(?:browser_)?route\s*:\s*(playwright|agent-browser|chrome-devtools)/i);
  return match ? match[1].toLowerCase() : null;
}

function detectBrowserRoute(task, root, browserMode) {
  // タスクに明示的な route 指定がある場合のみ contract に焼き込む。
  // それ以外は null を返し、generate-browser-review-artifact.sh が
  // 実行時の環境で route を決定する（contract の環境依存を排除）。
  const explicitRoute = detectExplicitBrowserRoute(task);
  if (explicitRoute) {
    return explicitRoute;
  }

  // exploratory モードは実行時に agent-browser 優先で解決するため焼き込まない
  if (browserMode === 'exploratory') return null;

  // scripted モードでは repo-based の検出結果を contract に焼き込む
  // （package.json の deps は環境非依存 → deterministic）
  if (hasPlaywrightBasis(root)) return 'playwright';

  // それでも解決できなければ null（artifact 生成時に実行環境で解決）
  return null;
}

function pickRuntimeCommands(root) {
  const commands = [];
  const packageJsonPath = path.join(root, 'package.json');
  if (fs.existsSync(packageJsonPath)) {
    try {
      const pkg = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
      const scripts = pkg.scripts || {};
      // CI=true で watch mode を抑制（Jest/Vitest 互換）
      if (scripts.test) commands.push({ label: 'package-test', command: 'CI=true npm test' });
      if (scripts.lint) commands.push({ label: 'package-lint', command: 'npm run lint' });
      if (scripts.typecheck) commands.push({ label: 'package-typecheck', command: 'npm run typecheck' });
      if (scripts['test:e2e']) commands.push({ label: 'package-e2e', command: 'npm run test:e2e' });
    } catch (error) {
      // パース失敗時は exit 1 で明示的に失敗させる（runtime gate すり抜け防止）
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

  // shell-repo fallback: package.json 等がなくても validate-plugin.sh / check-consistency.sh があれば使う
  if (commands.length === 0) {
    const shellFallbacks = [
      { path: 'tests/validate-plugin.sh', label: 'validate-plugin', command: './tests/validate-plugin.sh' },
      { path: 'scripts/ci/check-consistency.sh', label: 'check-consistency', command: './scripts/ci/check-consistency.sh' },
    ];
    for (const fb of shellFallbacks) {
      if (fs.existsSync(path.join(root, fb.path))) {
        commands.push({ label: fb.label, command: fb.command });
      }
    }
  }

  return commands;
}

const markdown = fs.readFileSync(plansFile, 'utf8');
const row = parseTaskRow(markdown, taskId);
if (!row) {
  console.error(`Task row not found in Plans.md: ${taskId}`);
  process.exit(3);
}

const reviewerProfile = detectProfile(row);
const browserMode = reviewerProfile === 'browser' ? detectBrowserMode(row) : null;
const browserRoute = reviewerProfile === 'browser' ? detectBrowserRoute(row, repoRoot, browserMode) : null;
const runtimeValidation = reviewerProfile === 'runtime' ? pickRuntimeCommands(repoRoot) : [];
const riskFlags = detectRiskFlags(row);
const maxIterations = detectMaxIterations(reviewerProfile, row);
const rubricTarget = reviewerProfile === 'ui-rubric'
  ? { ...DEFAULT_UI_RUBRIC_TARGET }
  : null;

const contract = {
  schema_version: 'sprint-contract.v1',
  generated_at: new Date().toISOString(),
  source: {
    plans_file: path.relative(repoRoot, plansFile) || 'Plans.md',
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
        source: 'Plans.md.DoD',
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
    max_iterations: maxIterations,
    ...(rubricTarget ? { rubric_target: rubricTarget } : {}),
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
