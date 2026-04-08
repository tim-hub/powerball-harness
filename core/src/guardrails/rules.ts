/**
 * core/src/guardrails/rules.ts
 * Harness v3 宣言的ガードルールテーブル
 *
 * pretooluse-guard.sh の全ルールを TypeScript 型安全な宣言的テーブルとして移植。
 * 各 GuardRule は条件 (toolPattern + evaluate) とアクション (HookResult) のペア。
 */

import type { GuardRule, HookResult, RuleContext } from "../types.js";

// ============================================================
// ヘルパー関数
// ============================================================

/** ファイルパスが保護されたパスに該当するか判定 */
function isProtectedPath(filePath: string): boolean {
  const protected_patterns = [
    /^\.git\//,
    /\/\.git\//,
    /^\.env$/,
    /\/\.env$/,
    /\.env\./,
    /id_rsa/,
    /id_ed25519/,
    /id_ecdsa/,
    /id_dsa/,
    /\.pem$/,
    /\.key$/,
    /\.p12$/,
    /\.pfx$/,
    /authorized_keys/,
    /known_hosts/,
  ];
  return protected_patterns.some((p) => p.test(filePath));
}

/** ファイルパスがプロジェクトルート配下にあるか判定 */
function isUnderProjectRoot(filePath: string, projectRoot: string): boolean {
  const root = projectRoot.endsWith("/") ? projectRoot : `${projectRoot}/`;
  return filePath.startsWith(root) || filePath === projectRoot;
}

/** Bash コマンド文字列から危険な rm -rf パターンを検出 */
function hasDangerousRmRf(command: string): boolean {
  // -rf または -fr フラグを含む rm コマンドを検出
  // 注意: rm -f（-r なし）は対象外
  if (/\brm\s+(?:[^\s]*\s+)*-(?=[^-]*r)[rf]+\b/.test(command)) return true;
  if (/\brm\s+--recursive\b/.test(command)) return true;
  return false;
}

/** git push --force パターンを検出 */
function hasForcePush(command: string): boolean {
  return /\bgit\s+push\b.*--force(?:-with-lease)?\b/.test(command) ||
    /\bgit\s+push\b.*-f\b/.test(command);
}

/** sudo の使用を検出 */
function hasSudo(command: string): boolean {
  return /(?:^|\s)sudo\s/.test(command);
}

/** Bash token の前後クォートを除去する */
function normalizeGitToken(token: string): string {
  return token.replace(/^['"]|['"]$/g, "");
}

/** `--no-verify` / `--no-gpg-sign` の使用を検出 */
function hasDangerousGitBypassFlag(command: string): boolean {
  return /(?:^|\s)--no-verify(?:\s|$)/.test(command) ||
    /(?:^|\s)--no-gpg-sign(?:\s|$)/.test(command);
}

/** protected branch への `git reset --hard` を検出 */
function hasProtectedBranchResetHard(command: string): boolean {
  const tokens = command.trim().split(/\s+/).map(normalizeGitToken);
  const resetIndex = tokens.indexOf("reset");
  if (resetIndex === -1) return false;
  if (!tokens.includes("--hard")) return false;

  const isProtectedBranchRef = (ref: string): boolean =>
    /^(?:origin\/|upstream\/)?(?:refs\/heads\/)?(?:main|master)(?:[~^]\d+)?$/.test(normalizeGitToken(ref));

  return tokens.slice(resetIndex + 1).some((token) => !token.startsWith("-") && isProtectedBranchRef(token));
}

/** protected branch への direct push を検出 */
function hasDirectPushToProtectedBranch(command: string): boolean {
  if (!/\bgit\s+push\b/.test(command)) return false;

  const tokens = command.trim().split(/\s+/);
  const pushIndex = tokens.indexOf("push");
  if (pushIndex === -1) return false;

  const args = tokens.slice(pushIndex + 1).filter((token) => !token.startsWith("-"));
  if (args.length === 0) return false;

  const isProtectedBranchRef = (ref: string): boolean =>
    /^(?:origin\/|upstream\/)?(?:refs\/heads\/)?(?:main|master)(?:[~^]\d+)?$/.test(normalizeGitToken(ref));

  for (const arg of args) {
    if (isProtectedBranchRef(arg)) return true;

    const refspecParts = arg.split(":");
    if (refspecParts.length === 2 && typeof refspecParts[1] === "string" && isProtectedBranchRef(refspecParts[1])) {
      return true;
    }
  }

  return false;
}

/** 重要ファイルへの書き込みを警告対象として検出 */
function isProtectedReviewPath(filePath: string): boolean {
  const protected_patterns = [
    /(?:^|\/)package\.json$/,
    /(?:^|\/)Dockerfile$/,
    /(?:^|\/)docker-compose\.yml$/,
    /(?:^|\/)\.github\/workflows\/[^/]+$/,
    /(?:^|\/)schema\.prisma$/,
    /(?:^|\/)wrangler\.toml$/,
    /(?:^|\/)index\.html$/,
  ];
  return protected_patterns.some((p) => p.test(filePath));
}

// ============================================================
// ガードルールテーブル
// ============================================================

export const GUARD_RULES: readonly GuardRule[] = [
  // ------------------------------------------------------------------
  // R01: sudo ブロック（Bash）
  // ------------------------------------------------------------------
  {
    id: "R01:no-sudo",
    toolPattern: /^Bash$/,
    evaluate(ctx: RuleContext): HookResult | null {
      const command = ctx.input.tool_input["command"];
      if (typeof command !== "string") return null;
      if (!hasSudo(command)) return null;
      return {
        decision: "deny",
        reason: "sudo の使用は禁止されています。必要な場合はユーザーに手動実行を依頼してください。",
      };
    },
  },

  // ------------------------------------------------------------------
  // R02: 保護パスへの書き込みブロック（Write / Edit / Bash）
  // ------------------------------------------------------------------
  {
    id: "R02:no-write-protected-paths",
    toolPattern: /^(?:Write|Edit|MultiEdit)$/,
    evaluate(ctx: RuleContext): HookResult | null {
      const filePath = ctx.input.tool_input["file_path"];
      if (typeof filePath !== "string") return null;
      if (!isProtectedPath(filePath)) return null;
      return {
        decision: "deny",
        reason: `保護されたパスへの書き込みは禁止されています: ${filePath}`,
      };
    },
  },

  // ------------------------------------------------------------------
  // R03: Bash での保護パスへの書き込みブロック（echo redirect / tee 等）
  // ------------------------------------------------------------------
  {
    id: "R03:no-bash-write-protected-paths",
    toolPattern: /^Bash$/,
    evaluate(ctx: RuleContext): HookResult | null {
      const command = ctx.input.tool_input["command"];
      if (typeof command !== "string") return null;
      // echo > .env, tee .git/config 等を検出
      // '>>' / '>' の後にスペースを挟んで保護パスが続くパターンも検出
      const writePatterns = [
        /(?:>>?|tee)\s+\S*\.env\b/,
        /(?:>>?|tee)\s+\S*\.env\./,
        /(?:>>?|tee)\s+\S*\.git\//,
        /(?:>>?|tee)\s+\S*id_rsa\b/,
        /(?:>>?|tee)\s+\S*id_ed25519\b/,
        /(?:>>?|tee)\s+\S*\.pem\b/,
        /(?:>>?|tee)\s+\S*\.key\b/,
      ];
      if (!writePatterns.some((p) => p.test(command))) return null;
      return {
        decision: "deny",
        reason: "保護されたファイルへのシェル書き込みは禁止されています。",
      };
    },
  },

  // ------------------------------------------------------------------
  // R04: プロジェクト外への書き込み確認（work モード時はスキップ）
  // ------------------------------------------------------------------
  {
    id: "R04:confirm-write-outside-project",
    toolPattern: /^(?:Write|Edit|MultiEdit)$/,
    evaluate(ctx: RuleContext): HookResult | null {
      const filePath = ctx.input.tool_input["file_path"];
      if (typeof filePath !== "string") return null;
      // 相対パスはプロジェクト内とみなす
      if (!filePath.startsWith("/")) return null;
      if (isUnderProjectRoot(filePath, ctx.projectRoot)) return null;
      // work モード時は確認をスキップ
      if (ctx.workMode) return null;
      return {
        decision: "ask",
        reason: `プロジェクトルート外への書き込みです: ${filePath}\n許可しますか？`,
      };
    },
  },

  // ------------------------------------------------------------------
  // R05: rm -rf 確認（work モードでバイパス可）
  // ------------------------------------------------------------------
  {
    id: "R05:confirm-rm-rf",
    toolPattern: /^Bash$/,
    evaluate(ctx: RuleContext): HookResult | null {
      const command = ctx.input.tool_input["command"];
      if (typeof command !== "string") return null;
      if (!hasDangerousRmRf(command)) return null;
      // work モードでバイパスが許可されている場合はスキップ
      if (ctx.workMode) return null;
      return {
        decision: "ask",
        reason: `危険な削除コマンドを検出しました:\n${command}\n実行しますか？`,
      };
    },
  },

  // ------------------------------------------------------------------
  // R06: git push --force ブロック（work モード時も例外なし）
  // ------------------------------------------------------------------
  {
    id: "R06:no-force-push",
    toolPattern: /^Bash$/,
    evaluate(ctx: RuleContext): HookResult | null {
      const command = ctx.input.tool_input["command"];
      if (typeof command !== "string") return null;
      if (!hasForcePush(command)) return null;
      return {
        decision: "deny",
        reason: "git push --force は禁止されています。履歴を破壊する操作は許可されません。",
      };
    },
  },

  // ------------------------------------------------------------------
  // R07: Codex モード時の Write/Edit ブロック
  // Claude は PM 役 — 実装は Codex Worker に委譲
  // ------------------------------------------------------------------
  {
    id: "R07:codex-mode-no-write",
    toolPattern: /^(?:Write|Edit|MultiEdit)$/,
    evaluate(ctx: RuleContext): HookResult | null {
      // Write / Edit / MultiEdit のみ対象（Bash は除外）
      if (!["Write", "Edit", "MultiEdit"].includes(ctx.input.tool_name)) {
        return null;
      }
      if (!ctx.codexMode) return null;
      return {
        decision: "deny",
        reason: "Codex モード中は Claude が直接ファイルを書き込めません。実装は Codex Worker (codex exec) に委譲してください。",
      };
    },
  },

  // ------------------------------------------------------------------
  // R08: Breezing ロールガード — reviewer は Write/Edit 不可
  // ------------------------------------------------------------------
  {
    id: "R08:breezing-reviewer-no-write",
    toolPattern: /^(?:Write|Edit|MultiEdit|Bash)$/,
    evaluate(ctx: RuleContext): HookResult | null {
      if (ctx.breezingRole !== "reviewer") return null;
      // Bash は読み取り専用コマンドのみ許可（ブロックはスクリプト側で判断）
      if (ctx.input.tool_name === "Bash") {
        const command = ctx.input.tool_input["command"];
        if (typeof command !== "string") return null;
        // git commit / git push / rm / mv 等を禁止
        const prohibited = [
          /\bgit\s+(?:commit|push|reset|checkout|merge|rebase)\b/,
          /\brm\s+/,
          /\bmv\s+/,
          /\bcp\s+.*-r\b/,
        ];
        if (!prohibited.some((p) => p.test(command))) return null;
      }
      return {
        decision: "deny",
        reason: `Breezing reviewer ロールはファイル書き込みおよびデータ変更コマンドを実行できません。`,
      };
    },
  },

  // ------------------------------------------------------------------
  // R09: 機密情報を含むファイルへのアクセス制限（Read のみ警告）
  // ------------------------------------------------------------------
  {
    id: "R09:warn-secret-file-read",
    toolPattern: /^Read$/,
    evaluate(ctx: RuleContext): HookResult | null {
      const filePath = ctx.input.tool_input["file_path"];
      if (typeof filePath !== "string") return null;
      const secretPatterns = [/\.env$/, /id_rsa$/, /\.pem$/, /\.key$/, /secrets?\//];
      if (!secretPatterns.some((p) => p.test(filePath))) return null;
      return {
        decision: "approve",
        systemMessage: `警告: 機密情報が含まれる可能性のあるファイルを読み取っています: ${filePath}`,
      };
    },
  },

  // ------------------------------------------------------------------
  // R10: Bash での `--no-verify` / `--no-gpg-sign` ブロック
  // ------------------------------------------------------------------
  {
    id: "R10:no-git-bypass-flags",
    toolPattern: /^Bash$/,
    evaluate(ctx: RuleContext): HookResult | null {
      const command = ctx.input.tool_input["command"];
      if (typeof command !== "string") return null;
      if (!hasDangerousGitBypassFlag(command)) return null;
      return {
        decision: "deny",
        reason: "--no-verify / --no-gpg-sign の使用は禁止されています。フックや署名検証を迂回しないでください。",
      };
    },
  },

  // ------------------------------------------------------------------
  // R11: protected branch への `git reset --hard` ブロック
  // ------------------------------------------------------------------
  {
    id: "R11:no-reset-hard-protected-branch",
    toolPattern: /^Bash$/,
    evaluate(ctx: RuleContext): HookResult | null {
      const command = ctx.input.tool_input["command"];
      if (typeof command !== "string") return null;
      if (!hasProtectedBranchResetHard(command)) return null;
      return {
        decision: "deny",
        reason: "protected branch への git reset --hard は禁止されています。履歴を壊さない方法を使ってください。",
      };
    },
  },

  // ------------------------------------------------------------------
  // R12: protected branch への direct push 警告
  // ------------------------------------------------------------------
  {
    id: "R12:deny-direct-push-protected-branch",
    toolPattern: /^Bash$/,
    evaluate(ctx: RuleContext): HookResult | null {
      const command = ctx.input.tool_input["command"];
      if (typeof command !== "string") return null;
      if (!hasDirectPushToProtectedBranch(command)) return null;
      return {
        decision: "deny",
        reason: "main/master への直接 push は禁止されています。feature branch 経由で PR を作成してください。",
      };
    },
  },

  // ------------------------------------------------------------------
  // R13: 重要ファイルの変更警告（Write / Edit / MultiEdit）
  // ------------------------------------------------------------------
  {
    id: "R13:warn-protected-review-paths",
    toolPattern: /^(?:Write|Edit|MultiEdit)$/,
    evaluate(ctx: RuleContext): HookResult | null {
      const filePath = ctx.input.tool_input["file_path"];
      if (typeof filePath !== "string") return null;
      if (!isProtectedReviewPath(filePath)) return null;
      return {
        decision: "approve",
        systemMessage: `警告: 重要ファイルへの変更を検出しました: ${filePath}`,
      };
    },
  },
];

/**
 * 全ルールを順番に評価し、最初にマッチしたルールの HookResult を返す。
 * どのルールもマッチしない場合は approve を返す。
 */
export function evaluateRules(ctx: RuleContext): HookResult {
  const toolName = ctx.input.tool_name;

  for (const rule of GUARD_RULES) {
    if (!rule.toolPattern.test(toolName)) continue;
    const result = rule.evaluate(ctx);
    if (result !== null) return result;
  }

  return { decision: "approve" };
}
