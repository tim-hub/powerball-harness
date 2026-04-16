# Go Guardrail Engine

How the Go binary (`bin/harness`) evaluates every Claude Code tool invocation in real time.

---

## End-to-End Flow

```mermaid
flowchart LR
    CC(["Claude Code\ntool invocation"])

    CC -->|stdin JSON| ENTRY["bin/harness hook &lt;event&gt;\nmain.go::runHook()"]

    ENTRY --> PARSE["hook.ReadInput()\nParse HookInput\n{sessionID, toolName,\ntoolInput, CWD, ...}"]

    PARSE -->|parse error\nor missing toolName| SAFE["hook.SafeResult()\n→ approve\n(fail-open)"]

    PARSE -->|ok| ROUTE{route by\nhook type}

    ROUTE -->|pre-tool| PRE["runPreTool()\n→ EvaluatePreTool()"]
    ROUTE -->|post-tool| POST["runPostTool()\n→ EvaluatePostTool()"]
    ROUTE -->|permission| PERM["runPermission()\n→ EvaluatePermission()"]
    ROUTE -->|session-*\nnotification\nworktree-*\netc.| EVT["Event handlers\n(no guardrail eval)"]

    PRE --> CTX["BuildContext()\n① env vars override\n  HARNESS_WORK_MODE\n  HARNESS_CODEX_MODE\n  HARNESS_BREEZING_ROLE\n② SQLite work_states\n   (fallback, best-effort)\n③ defaults"]

    CTX --> RULES["EvaluateRules()\nIterate Rules[]\nfirst match wins"]

    RULES --> FMT["FormatPreToolResult()"]
    POST  --> FMT2["PostToolOutput"]
    PERM  --> FMT3["PermissionOutput"]

    FMT  --> OUT["hook.WriteJSON(stdout)\nos.Exit(code)"]
    FMT2 --> OUT
    FMT3 --> OUT
    SAFE --> OUT
    EVT  --> OUT

    OUT --> CC2(["Claude Code\ndecision"])
```

---

## Rule Engine (Pre-Tool)

```mermaid
flowchart LR
    INPUT["RuleContext\n{toolName, toolInput,\nworkMode, codexMode,\nbreezingRole, projectRoot}"]

    INPUT --> LOOP["for each rule in Rules[]"]

    LOOP -->|toolPattern\ndoes not match| NEXT["next rule"]
    NEXT --> LOOP

    LOOP -->|toolPattern matches| EVAL["rule.Evaluate(ctx)"]

    EVAL -->|nil = no match| NEXT

    EVAL -->|non-nil result| RESULT{decision}

    RESULT -->|Deny| D["exit 2\npermissionDecision: deny\n+ reason"]
    RESULT -->|Ask| A["exit 0\npermissionDecision: ask\n+ reason"]
    RESULT -->|Approve + warn| W["exit 0\npermissionDecision: allow\n+ additionalContext"]
    RESULT -->|Approve| OK["exit 0\n(empty stdout)"]

    LOOP -->|all rules exhausted| OK
```

---

## Rules R01 – R13

```mermaid
flowchart LR
    subgraph BASH["Bash tool"]
        R01["R01 no-sudo\nDetects: sudo keyword\n→ DENY"]
        R03["R03 no-bash-write-protected\nDetects: shell redirection\nto .env, keys, certs\n→ DENY"]
        R05["R05 confirm-rm-rf\nDetects: rm -rf / rm -r -f\n→ ASK  (bypass: workMode)"]
        R06["R06 no-force-push\nDetects: git push --force/-f\n→ DENY  (no bypass)"]
        R08B["R08 breezing-reviewer-no-write\nDetects: reviewer role +\ngit commit/push/reset/merge\nor rm/mv/cp -r\n→ DENY"]
        R10["R10 no-git-bypass-flags\nDetects: --no-verify\n--no-gpg-sign\n→ DENY"]
        R11["R11 no-reset-hard-protected\nDetects: git reset --hard\nmain or master\n→ DENY"]
        R12["R12 no-direct-push-protected\nDetects: git push to\nmain or master\n→ DENY"]
    end

    subgraph WRITE["Write / Edit / MultiEdit"]
        R02["R02 no-write-protected-paths\nDetects: .env*, .git/,\nSSH keys, certs, .husky/\n→ DENY"]
        R04["R04 confirm-write-outside-project\nDetects: absolute path\nnot under projectRoot\n→ ASK  (bypass: workMode)"]
        R07["R07 codex-mode-no-write\nDetects: CodexMode == true\n→ DENY"]
        R08W["R08 breezing-reviewer-no-write\nDetects: reviewer role\n→ DENY"]
        R13["R13 warn-protected-review-paths\nDetects: package.json,\nDockerfile, workflows,\nschema.prisma, etc.\n→ APPROVE + warn"]
    end

    subgraph READ["Read tool"]
        R09["R09 warn-secret-file-read\nDetects: .env, id_rsa,\n*.pem, *.key, secrets/\n→ APPROVE + warn"]
    end
```

---

## Post-Tool Pipeline (Tampering + Security Scan)

```mermaid
flowchart LR
    IN["PostToolUse\ntoolName + content"]

    IN --> CHK{Write / Edit\n/ MultiEdit?}
    CHK -->|No| APASS["Approve\n(exit 0)"]

    CHK -->|Yes| TAMP["detectTampering(content)\nTest / CI file?"]

    TAMP --> T["Tampering patterns T01-T12\nT01: it.skip / describe.skip\nT02: xit / xdescribe\nT03: pytest.mark.skip\nT04: t.Skip()\nT05: expect() commented out\nT06: assert() commented out\nT07: TODO instead of assertion\nT08: eslint-disable\nT09: continue-on-error: true\nT10: if: always()\nT11: hardcoded answer dict\nT12: hardcoded return value"]

    IN --> SEC["detectSecurityRisks(content)\nAll Write/Edit files"]

    SEC --> S["Security patterns\nprocess.env secrets leak\neval with user input\nshell injection via template\ninnerHTML with user data\nhardcoded password/api_key"]

    T & S --> AGG["Aggregate warnings"]

    AGG -->|warnings| WARN["Approve + systemMessage\n(advisory, not blocking)"]
    AGG -->|none| APASS2["Approve\n(exit 0)"]
```

---

## Permission Handler

```mermaid
flowchart LR
    P["PermissionRequest\ntoolName + command"]

    P --> PT{tool type}

    PT -->|Write / Edit\n/ MultiEdit| ALLOW["makePermissionAllow()\nbehavior: allow"]

    PT -->|Bash| SAFE2{isSafeCommand?}

    SAFE2 -->|yes| ALLOW

    SAFE2 -->|no| NIL["nil output\n(CC prompts user)"]

    PT -->|anything else| NIL

    subgraph SAFE_CMDS["Safe command allowlist"]
        SC1["git status / diff / log\ngit branch / show / ls-files\ngit rev-parse"]
        SC2["npm test / pnpm test\nyarn test / bun test"]
        SC3["pytest / python -m pytest\ngo test / cargo test"]
    end

    subgraph UNSAFE["Auto-reject if command has"]
        U1["; & | < > backtick dollar\n(shell specials)"]
        U2["Multiline / backslash escapes"]
        U3["Unknown env-var prefixes"]
    end

    SAFE2 -.->|checked against| SAFE_CMDS
    SAFE2 -.->|rejected by| UNSAFE
```

---

## State Resolution (Context Building)

```mermaid
flowchart LR
    ENV["① Env vars\nHARNESS_WORK_MODE\nHARNESS_CODEX_MODE\nHARNESS_BREEZING_ROLE"]
    DB["② SQLite\nwork_states table\n(best-effort, silent on error)"]
    DEF["③ Defaults\nworkMode=false\ncodexMode=false\nbreezingRole=''"]

    ENV -->|present| CTX2["RuleContext"]
    ENV -->|absent| DB
    DB  -->|row found| CTX2
    DB  -->|miss / error| DEF
    DEF --> CTX2

    subgraph SQLite["SQLite Schema"]
        T1["sessions — lifecycle"]
        T2["work_states — mode flags"]
        T3["signals — inter-session msgs"]
        T4["task_failures — error tracking"]
        T5["agent_states — subagent lifecycle"]
    end
```

---

## Fail-Safe Design

```mermaid
flowchart LR
    ERR["Infrastructure error"]

    ERR --> E1["Empty stdin"]
    ERR --> E2["JSON parse error"]
    ERR --> E3["Missing toolName"]
    ERR --> E4["Symlink loop on\nprotected path check"]
    ERR --> E5["SQLite unavailable\n(context build)"]

    E1 & E2 & E3 --> FS["hook.SafeResult()\n→ approve\n(fail-open)"]
    E4 --> FD["→ deny\n(fail-safe: unknown path = block)"]
    E5 --> FD2["→ use defaults\n(silent, no block)"]
```

> **Principle**: Hook infrastructure failures never lock users out. Only explicit rule matches deny. Symlink errors are the one exception — an unresolvable path is treated as protected.

---

## Key Implementation Details

| Property | Value |
|----------|-------|
| **Latency target** | < 5 ms per hook invocation |
| **Rule evaluation** | Short-circuit — first match wins |
| **Regex compilation** | Pre-compiled at package `init()` |
| **Whitespace normalization** | Applied before all regex (bypass defence) |
| **Protected paths** | `.env*`, `.git/`, SSH keys, certs, `.husky/` |
| **Protected branches** | `main`, `master` (and `origin/`, `upstream/` prefixes) |
| **Binary size** | ~2.5 MB (darwin/arm64, stripped) |
| **External deps** | stdlib + `google/uuid` only |
| **DB access** | Best-effort SQLite; silently ignored on error |
| **Exit codes** | `0` = allow, `2` = deny |

> See `internal/guardrail/` for rule source and `internal/hook/codec.go` for the stdin/stdout codec.
