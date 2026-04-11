/**
 * core/src/guardrails/rules.ts
 * Harness v3 declarative guard rule table
 *
 * All rules from pretooluse-guard.sh ported as a type-safe declarative table.
 * Each GuardRule is a pair of condition (toolPattern + evaluate) and action (HookResult).
 */
// ============================================================
// Helper functions
// ============================================================
/** Check if a file path matches a protected path pattern */
function isProtectedPath(filePath) {
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
/** Check if a file path is under the project root */
function isUnderProjectRoot(filePath, projectRoot) {
    const root = projectRoot.endsWith("/") ? projectRoot : `${projectRoot}/`;
    return filePath.startsWith(root) || filePath === projectRoot;
}
/** Detect dangerous rm -rf patterns in a Bash command string */
function hasDangerousRmRf(command) {
    // Detect rm commands with -rf or -fr flags
    // Note: rm -f (without -r) is not targeted
    if (/\brm\s+(?:[^\s]*\s+)*-(?=[^-]*r)[rf]+\b/.test(command))
        return true;
    if (/\brm\s+--recursive\b/.test(command))
        return true;
    return false;
}
/** Detect git push --force patterns */
function hasForcePush(command) {
    return /\bgit\s+push\b.*--force(?:-with-lease)?\b/.test(command) ||
        /\bgit\s+push\b.*-f\b/.test(command);
}
/** Detect sudo usage */
function hasSudo(command) {
    return /(?:^|\s)sudo\s/.test(command);
}
/** Strip surrounding quotes from a Bash token */
function normalizeGitToken(token) {
    return token.replace(/^['"]|['"]$/g, "");
}
/** Detect usage of `--no-verify` / `--no-gpg-sign` */
function hasDangerousGitBypassFlag(command) {
    return /(?:^|\s)--no-verify(?:\s|$)/.test(command) ||
        /(?:^|\s)--no-gpg-sign(?:\s|$)/.test(command);
}
/** Detect `git reset --hard` targeting a protected branch */
function hasProtectedBranchResetHard(command) {
    const tokens = command.trim().split(/\s+/).map(normalizeGitToken);
    const resetIndex = tokens.indexOf("reset");
    if (resetIndex === -1)
        return false;
    if (!tokens.includes("--hard"))
        return false;
    const isProtectedBranchRef = (ref) => /^(?:origin\/|upstream\/)?(?:refs\/heads\/)?(?:main|master)(?:[~^]\d+)?$/.test(normalizeGitToken(ref));
    return tokens.slice(resetIndex + 1).some((token) => !token.startsWith("-") && isProtectedBranchRef(token));
}
/** Detect direct push to a protected branch */
function hasDirectPushToProtectedBranch(command) {
    if (!/\bgit\s+push\b/.test(command))
        return false;
    const tokens = command.trim().split(/\s+/);
    const pushIndex = tokens.indexOf("push");
    if (pushIndex === -1)
        return false;
    const args = tokens.slice(pushIndex + 1).filter((token) => !token.startsWith("-"));
    if (args.length === 0)
        return false;
    const isProtectedBranchRef = (ref) => /^(?:origin\/|upstream\/)?(?:refs\/heads\/)?(?:main|master)(?:[~^]\d+)?$/.test(normalizeGitToken(ref));
    for (const arg of args) {
        if (isProtectedBranchRef(arg))
            return true;
        const refspecParts = arg.split(":");
        if (refspecParts.length === 2 && typeof refspecParts[1] === "string" && isProtectedBranchRef(refspecParts[1])) {
            return true;
        }
    }
    return false;
}
/** Detect writes to important files that warrant review warnings */
function isProtectedReviewPath(filePath) {
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
// Guard rule table
// ============================================================
export const GUARD_RULES = [
    // ------------------------------------------------------------------
    // R01: Block sudo (Bash)
    // ------------------------------------------------------------------
    {
        id: "R01:no-sudo",
        toolPattern: /^Bash$/,
        evaluate(ctx) {
            const command = ctx.input.tool_input["command"];
            if (typeof command !== "string")
                return null;
            if (!hasSudo(command))
                return null;
            return {
                decision: "deny",
                reason: "Use of sudo is prohibited. Please ask the user to run manually if needed.",
            };
        },
    },
    // ------------------------------------------------------------------
    // R02: Block writes to protected paths (Write / Edit)
    // ------------------------------------------------------------------
    {
        id: "R02:no-write-protected-paths",
        toolPattern: /^(?:Write|Edit|MultiEdit)$/,
        evaluate(ctx) {
            const filePath = ctx.input.tool_input["file_path"];
            if (typeof filePath !== "string")
                return null;
            if (!isProtectedPath(filePath))
                return null;
            return {
                decision: "deny",
                reason: `Writing to protected paths is prohibited: ${filePath}`,
            };
        },
    },
    // ------------------------------------------------------------------
    // R03: Block Bash writes to protected paths (echo redirect / tee, etc.)
    // ------------------------------------------------------------------
    {
        id: "R03:no-bash-write-protected-paths",
        toolPattern: /^Bash$/,
        evaluate(ctx) {
            const command = ctx.input.tool_input["command"];
            if (typeof command !== "string")
                return null;
            // Detect echo > .env, tee .git/config, etc.
            // Also detect '>>' / '>' followed by a space and a protected path
            const writePatterns = [
                /(?:>>?|tee)\s+\S*\.env\b/,
                /(?:>>?|tee)\s+\S*\.env\./,
                /(?:>>?|tee)\s+\S*\.git\//,
                /(?:>>?|tee)\s+\S*id_rsa\b/,
                /(?:>>?|tee)\s+\S*id_ed25519\b/,
                /(?:>>?|tee)\s+\S*\.pem\b/,
                /(?:>>?|tee)\s+\S*\.key\b/,
            ];
            if (!writePatterns.some((p) => p.test(command)))
                return null;
            return {
                decision: "deny",
                reason: "Shell writes to protected files are prohibited.",
            };
        },
    },
    // ------------------------------------------------------------------
    // R04: Confirm writes outside project root (skipped in work mode)
    // ------------------------------------------------------------------
    {
        id: "R04:confirm-write-outside-project",
        toolPattern: /^(?:Write|Edit|MultiEdit)$/,
        evaluate(ctx) {
            const filePath = ctx.input.tool_input["file_path"];
            if (typeof filePath !== "string")
                return null;
            // Relative paths are considered within the project
            if (!filePath.startsWith("/"))
                return null;
            if (isUnderProjectRoot(filePath, ctx.projectRoot))
                return null;
            // Skip confirmation in work mode
            if (ctx.workMode)
                return null;
            return {
                decision: "ask",
                reason: `Writing outside the project root: ${filePath}\nAllow?`,
            };
        },
    },
    // ------------------------------------------------------------------
    // R05: Confirm rm -rf (bypassable in work mode)
    // ------------------------------------------------------------------
    {
        id: "R05:confirm-rm-rf",
        toolPattern: /^Bash$/,
        evaluate(ctx) {
            const command = ctx.input.tool_input["command"];
            if (typeof command !== "string")
                return null;
            if (!hasDangerousRmRf(command))
                return null;
            // Skip if bypass is allowed in work mode
            if (ctx.workMode)
                return null;
            return {
                decision: "ask",
                reason: `Dangerous delete command detected:\n${command}\nProceed?`,
            };
        },
    },
    // ------------------------------------------------------------------
    // R06: Block git push --force (no exceptions even in work mode)
    // ------------------------------------------------------------------
    {
        id: "R06:no-force-push",
        toolPattern: /^Bash$/,
        evaluate(ctx) {
            const command = ctx.input.tool_input["command"];
            if (typeof command !== "string")
                return null;
            if (!hasForcePush(command))
                return null;
            return {
                decision: "deny",
                reason: "git push --force is prohibited. History-destroying operations are not allowed.",
            };
        },
    },
    // ------------------------------------------------------------------
    // R07: Block Write/Edit in Codex mode
    // Claude acts as PM — implementation is delegated to Codex Worker
    // ------------------------------------------------------------------
    {
        id: "R07:codex-mode-no-write",
        toolPattern: /^(?:Write|Edit|MultiEdit)$/,
        evaluate(ctx) {
            // Only target Write / Edit / MultiEdit (Bash is excluded)
            if (!["Write", "Edit", "MultiEdit"].includes(ctx.input.tool_name)) {
                return null;
            }
            if (!ctx.codexMode)
                return null;
            return {
                decision: "deny",
                reason: "Claude cannot write files directly in Codex mode. Delegate implementation to the Codex Worker (codex exec).",
            };
        },
    },
    // ------------------------------------------------------------------
    // R08: Breezing role guard — reviewer cannot Write/Edit
    // ------------------------------------------------------------------
    {
        id: "R08:breezing-reviewer-no-write",
        toolPattern: /^(?:Write|Edit|MultiEdit|Bash)$/,
        evaluate(ctx) {
            if (ctx.breezingRole !== "reviewer")
                return null;
            // For Bash, only allow read-only commands (blocking is determined by the script)
            if (ctx.input.tool_name === "Bash") {
                const command = ctx.input.tool_input["command"];
                if (typeof command !== "string")
                    return null;
                // Prohibit git commit / git push / rm / mv, etc.
                const prohibited = [
                    /\bgit\s+(?:commit|push|reset|checkout|merge|rebase)\b/,
                    /\brm\s+/,
                    /\bmv\s+/,
                    /\bcp\s+.*-r\b/,
                ];
                if (!prohibited.some((p) => p.test(command)))
                    return null;
            }
            return {
                decision: "deny",
                reason: `The breezing reviewer role cannot perform file writes or data modification commands.`,
            };
        },
    },
    // ------------------------------------------------------------------
    // R09: Restrict access to files containing sensitive data (Read warning only)
    // ------------------------------------------------------------------
    {
        id: "R09:warn-secret-file-read",
        toolPattern: /^Read$/,
        evaluate(ctx) {
            const filePath = ctx.input.tool_input["file_path"];
            if (typeof filePath !== "string")
                return null;
            const secretPatterns = [/\.env$/, /id_rsa$/, /\.pem$/, /\.key$/, /secrets?\//];
            if (!secretPatterns.some((p) => p.test(filePath)))
                return null;
            return {
                decision: "approve",
                systemMessage: `Warning: Reading a file that may contain sensitive data: ${filePath}`,
            };
        },
    },
    // ------------------------------------------------------------------
    // R10: Block `--no-verify` / `--no-gpg-sign` in Bash
    // ------------------------------------------------------------------
    {
        id: "R10:no-git-bypass-flags",
        toolPattern: /^Bash$/,
        evaluate(ctx) {
            const command = ctx.input.tool_input["command"];
            if (typeof command !== "string")
                return null;
            if (!hasDangerousGitBypassFlag(command))
                return null;
            return {
                decision: "deny",
                reason: "Use of --no-verify / --no-gpg-sign is prohibited. Do not bypass hooks or signature verification.",
            };
        },
    },
    // ------------------------------------------------------------------
    // R11: Block `git reset --hard` to protected branches
    // ------------------------------------------------------------------
    {
        id: "R11:no-reset-hard-protected-branch",
        toolPattern: /^Bash$/,
        evaluate(ctx) {
            const command = ctx.input.tool_input["command"];
            if (typeof command !== "string")
                return null;
            if (!hasProtectedBranchResetHard(command))
                return null;
            return {
                decision: "deny",
                reason: "git reset --hard to a protected branch is prohibited. Use a non-destructive method.",
            };
        },
    },
    // ------------------------------------------------------------------
    // R12: Warn on direct push to protected branches
    // ------------------------------------------------------------------
    {
        id: "R12:warn-direct-push-protected-branch",
        toolPattern: /^Bash$/,
        evaluate(ctx) {
            const command = ctx.input.tool_input["command"];
            if (typeof command !== "string")
                return null;
            if (!hasDirectPushToProtectedBranch(command))
                return null;
            return {
                decision: "approve",
                systemMessage: "Warning: Direct push to main/master detected. Using feature branches is recommended.",
            };
        },
    },
    // ------------------------------------------------------------------
    // R13: Warn on important file changes (Write / Edit / MultiEdit)
    // ------------------------------------------------------------------
    {
        id: "R13:warn-protected-review-paths",
        toolPattern: /^(?:Write|Edit|MultiEdit)$/,
        evaluate(ctx) {
            const filePath = ctx.input.tool_input["file_path"];
            if (typeof filePath !== "string")
                return null;
            if (!isProtectedReviewPath(filePath))
                return null;
            return {
                decision: "approve",
                systemMessage: `Warning: Change detected to an important file: ${filePath}`,
            };
        },
    },
];
/**
 * Evaluate all rules in order and return the HookResult of the first match.
 * Returns approve if no rules match.
 */
export function evaluateRules(ctx) {
    const toolName = ctx.input.tool_name;
    for (const rule of GUARD_RULES) {
        if (!rule.toolPattern.test(toolName))
            continue;
        const result = rule.evaluate(ctx);
        if (result !== null)
            return result;
    }
    return { decision: "approve" };
}
//# sourceMappingURL=rules.js.map