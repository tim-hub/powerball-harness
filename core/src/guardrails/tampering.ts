/**
 * core/src/guardrails/tampering.ts
 * テスト改ざん検出エンジン
 *
 * posttooluse-tampering-detector.sh の全パターンを TypeScript に移植。
 * Write / Edit / MultiEdit ツールでテストファイルや CI 設定が変更された後、
 * 改ざんパターンを検出して警告を返す（ブロックはしない）。
 */

import type { HookInput, HookResult } from "../types.js";

// ============================================================
// ファイル種別判定
// ============================================================

const TEST_FILE_PATTERNS = [
  /\.test\.[jt]sx?$/,
  /\.spec\.[jt]sx?$/,
  /\.test\.py$/,
  /test_[^/]+\.py$/,
  /[^/]+_test\.py$/,
  /\.test\.go$/,
  /[^/]+_test\.go$/,
  /\/__tests__\//,
  /\/tests\//,
] as const;

const CONFIG_FILE_PATTERNS = [
  /(?:^|\/)\.eslintrc(?:\.[^/]+)?$/,
  /(?:^|\/)eslint\.config\.[^/]+$/,
  /(?:^|\/)\.prettierrc(?:\.[^/]+)?$/,
  /(?:^|\/)prettier\.config\.[^/]+$/,
  /(?:^|\/)tsconfig(?:\.[^/]+)?\.json$/,
  /(?:^|\/)biome\.json$/,
  /(?:^|\/)\.stylelintrc(?:\.[^/]+)?$/,
  /(?:^|\/)(?:jest|vitest)\.config\.[^/]+$/,
  /\.github\/workflows\/[^/]+\.ya?ml$/,
  /(?:^|\/)\.gitlab-ci\.ya?ml$/,
  /(?:^|\/)Jenkinsfile$/,
] as const;

function isTestFile(filePath: string): boolean {
  return TEST_FILE_PATTERNS.some((p) => p.test(filePath));
}

function isConfigFile(filePath: string): boolean {
  return CONFIG_FILE_PATTERNS.some((p) => p.test(filePath));
}

// ============================================================
// 改ざんパターン定義
// ============================================================

interface TamperingPattern {
  id: string;
  description: string;
  /** マッチさせるテキスト範囲のパターン */
  pattern: RegExp;
  /** テストファイルのみ適用（false = 設定ファイルにも適用） */
  testFileOnly: boolean;
}

const TAMPERING_PATTERNS: readonly TamperingPattern[] = [
  // --- テストスキップ ---
  {
    id: "T01:it-skip",
    description: "it.skip / describe.skip によるテストスキップ",
    pattern: /(?:it|test|describe|context)\.skip\s*\(/,
    testFileOnly: true,
  },
  {
    id: "T02:xit-xdescribe",
    description: "xit / xdescribe によるテスト無効化",
    pattern: /\b(?:xit|xtest|xdescribe)\s*\(/,
    testFileOnly: true,
  },
  {
    id: "T03:pytest-skip",
    description: "pytest.mark.skip によるテストスキップ",
    pattern: /@pytest\.mark\.(?:skip|xfail)\b/,
    testFileOnly: true,
  },
  {
    id: "T04:go-skip",
    description: "t.Skip() によるテストスキップ",
    pattern: /\bt\.Skip(?:f|Now)?\s*\(/,
    testFileOnly: true,
  },

  // --- アサーション削除・改ざん ---
  {
    id: "T05:expect-removed",
    description: "expect / assert が削除された可能性（コメントアウト）",
    pattern: /\/\/\s*expect\s*\(/,
    testFileOnly: true,
  },
  {
    id: "T06:assert-commented",
    description: "assert 呼び出しがコメントアウトされた",
    pattern: /\/\/\s*assert(?:Equal|NotEqual|True|False|Nil|Error)?\s*\(/,
    testFileOnly: true,
  },
  {
    id: "T07:todo-assert",
    description: "TODO コメントによってアサーションが置き換えられた",
    pattern: /\/\/\s*TODO.*assert|\/\/\s*TODO.*expect/i,
    testFileOnly: true,
  },

  // --- ESLint 無効化（テスト・CI 設定） ---
  {
    id: "T08:eslint-disable",
    description: "eslint-disable による lint ルール無効化",
    // // eslint-disable と /* eslint-disable */ 両形式に対応
    pattern: /(?:\/\/\s*eslint-disable(?:-next-line|-line)?(?:\s+[^\n]+)?$|\/\*\s*eslint-disable\b[^*]*\*\/)/m,
    testFileOnly: false,
  },

  // --- CI ワークフロー改ざん ---
  {
    id: "T09:ci-continue-on-error",
    description: "continue-on-error: true による CI 失敗無視",
    pattern: /continue-on-error\s*:\s*true/,
    testFileOnly: false,
  },
  {
    id: "T10:ci-if-always",
    description: "if: always() による CI ステップ強制実行",
    pattern: /if\s*:\s*always\s*\(\s*\)/,
    testFileOnly: false,
  },

  // --- ハードコード期待値 ---
  {
    id: "T11:hardcoded-answer",
    description: "テスト期待値のハードコード（辞書返し）",
    pattern: /answers?_for_tests?\s*=\s*\{/,
    testFileOnly: true,
  },
  {
    id: "T12:return-hardcoded",
    description: "テストケース値を直接 return するパターン",
    pattern:
      /return\s+(?:"[^"]*"|'[^']*'|\d+)\s*;\s*\/\/.*(?:test|spec|expect)/i,
    testFileOnly: true,
  },
];

// ============================================================
// 検出関数
// ============================================================

interface TamperingWarning {
  patternId: string;
  description: string;
  matchedText: string;
}

/**
 * テキスト（new_string または content）に対して改ざんパターンを検索する。
 */
function detectTampering(
  text: string,
  isTest: boolean
): TamperingWarning[] {
  const warnings: TamperingWarning[] = [];

  for (const p of TAMPERING_PATTERNS) {
    if (p.testFileOnly && !isTest) continue;

    const match = p.pattern.exec(text);
    if (match !== null) {
      warnings.push({
        patternId: p.id,
        description: p.description,
        matchedText: match[0].slice(0, 120),
      });
    }
  }

  return warnings;
}

/**
 * HookInput からファイルパスと変更テキストを抽出する。
 */
function extractTargets(
  input: HookInput
): { filePath: string; changedText: string } | null {
  const toolInput = input.tool_input;
  const filePath = toolInput["file_path"];
  if (typeof filePath !== "string" || filePath.length === 0) return null;

  // Write: content フィールド
  // Edit: new_string フィールド
  const changedText =
    typeof toolInput["content"] === "string"
      ? toolInput["content"]
      : typeof toolInput["new_string"] === "string"
        ? toolInput["new_string"]
        : null;

  if (changedText === null) return null;

  return { filePath, changedText };
}

// ============================================================
// エクスポート: PostToolUse エントリポイント
// ============================================================

/**
 * PostToolUse フックでテスト改ざんを検出し、警告を返す。
 * 改ざんを検出した場合でも decision は "approve"（ブロックしない）。
 * 警告は systemMessage として Claude に渡される。
 */
export function detectTestTampering(input: HookInput): HookResult {
  // Write / Edit / MultiEdit のみ対象
  if (!["Write", "Edit", "MultiEdit"].includes(input.tool_name)) {
    return { decision: "approve" };
  }

  const targets = extractTargets(input);
  if (targets === null) return { decision: "approve" };

  const { filePath, changedText } = targets;
  const isTest = isTestFile(filePath);
  const isConfig = isConfigFile(filePath);

  if (!isTest && !isConfig) return { decision: "approve" };

  const warnings = detectTampering(changedText, isTest);

  if (warnings.length === 0) return { decision: "approve" };

  const fileType = isTest ? "テストファイル" : "CI/設定ファイル";
  const warningLines = warnings
    .map((w) => `- [${w.patternId}] ${w.description}\n  検出箇所: ${w.matchedText}`)
    .join("\n");

  const systemMessage =
    `[Harness v3] テスト改ざん検出警告\n\n` +
    `${fileType} \`${filePath}\` に疑わしいパターンが検出されました:\n\n` +
    warningLines +
    `\n\n【確認してください】\n` +
    `この変更がテストを意図的に無効化したり、実装品質を下げるものでないかを確認してください。\n` +
    `改ざんと判断した場合は変更を元に戻してください。`;

  return {
    decision: "approve",
    systemMessage,
  };
}
