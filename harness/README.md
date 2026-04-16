# Harness — Workflow Overview

Full lifecycle of a feature from project init to release, showing which skills and agents are active at each stage.

---

## Primary Workflow

```mermaid
flowchart LR
    START([User / New Project])

    %% ── SETUP ──────────────────────────────────────────────
    subgraph SETUP["① Setup"]
        direction TB
        HS["skill: harness-setup\n/harness-setup init"]
        SC["agent: scaffolder\n(tech-stack detection,\nCLAUDE.md, Plans.md)"]
        HS --> SC
    end

    %% ── PLANNING ────────────────────────────────────────────
    subgraph PLAN["② Planning"]
        direction TB
        HP["skill: harness-plan\n/harness-plan create|add"]
        SYNC0["skill: harness-sync\n(drift check, retrospective)"]
        MEM["skill: memory\n(decisions.md, patterns.md)"]
        HP --> MEM
    end

    %% ── WORK MODE SELECTION ─────────────────────────────────
    subgraph MODE["③ Execution Mode"]
        direction TB
        HW["skill: harness-work\n/harness-work all"]
        DEC{Task count?}
        HW --> DEC
    end

    %% ── SOLO ────────────────────────────────────────────────
    subgraph SOLO["Solo  (1 task)"]
        direction TB
        W1["agent: worker\nTDD → impl → self-check\ngit commit"]
        RV1["agent: reviewer\nstatic / runtime / browser\nverdict"]
        FX1{"APPROVE?"}
        W1 --> RV1 --> FX1
        FX1 -->|REQUEST_CHANGES\n≤ 3 retries| W1
    end

    %% ── PARALLEL ────────────────────────────────────────────
    subgraph PAR["Parallel  (2–3 tasks)"]
        direction TB
        W2A["agent: worker A\n(worktree)"]
        W2B["agent: worker B\n(worktree)"]
        RV2["agent: reviewer\nper-task verdict"]
        W2A & W2B --> RV2
    end

    %% ── BREEZING ────────────────────────────────────────────
    subgraph BREEZING["Breezing  (4+ tasks  /  team mode)"]
        direction TB
        LEAD["Lead (internal to harness-work)\nPhase A: decompose + sprint-contracts\nPhase B: delegate + fix-loop\nPhase C: cherry-pick + report"]
        WW["agent: worker ×1–3\n(parallel worktrees)\nTDD → impl → amend"]
        RVV["agent: reviewer\nindependent verdict"]
        LOOP{"APPROVE?"}
        LEAD -->|spawn| WW
        WW -->|result| RVV
        RVV --> LOOP
        LOOP -->|REQUEST_CHANGES\nSendMessage fix feedback\n≤ 3 retries| WW
        LOOP -->|APPROVE\ncherry-pick to main| LEAD
    end

    %% ── CI RECOVERY ─────────────────────────────────────────
    subgraph CI["CI / Error Recovery"]
        direction TB
        CISKILL["skill: ci\n(diagnose red build)"]
        CIFIXER["agent: ci-cd-fixer\n(3-strike escalation)"]
        CISKILL --> CIFIXER
    end

    %% ── REVIEW ──────────────────────────────────────────────
    subgraph REVIEW["④ Review Gate"]
        direction TB
        HR["skill: harness-review\n/harness-review code|plan|scope"]
        PROFILES["profiles:\n• static (diff)\n• runtime (tests)\n• browser (UI)\n• --dual (two reviewers)\n• --security"]
        HR --> PROFILES
    end

    %% ── RELEASE ─────────────────────────────────────────────
    subgraph RELEASE["⑤ Release"]
        direction TB
        CL["skill: writing-changelog\n(Before/After entries)"]
        REL["skill: harness-release\n/harness-release patch|minor|major"]
        STEPS["preflight → VERSION bump\n→ CHANGELOG finalize\n→ git tag → GitHub Release\n→ optional --announce"]
        CL --> REL --> STEPS
    end

    %% ── SESSION LAYER (always-on) ───────────────────────────
    subgraph SESSION["Session Layer  (always-on)"]
        direction LR
        SI["skill: session-init\n(env check on start)"]
        SS["skill: session-state\n(state machine)"]
        SM["skill: session-memory\n(auto-record logs)"]
        SC2["skill: session-control\n(resume / fork)"]
    end

    %% ── DOMAIN SKILLS (on-demand) ───────────────────────────
    subgraph DOMAIN["Domain Skills  (on-demand)"]
        direction TB
        AUTH["skill: auth\n(OAuth, RBAC,\nStripe, sessions)"]
        CRUD["skill: crud\n(endpoints, models,\nvalidation, tests)"]
        UI2["skill: ui\n(components,\nhero, forms, a11y)"]
        DEPLOY["skill: deploy\n(Vercel/Netlify,\nhealth, analytics)"]
    end

    %% ── UTILITY SKILLS ──────────────────────────────────────
    subgraph UTIL["Utility Skills  (on-demand)"]
        direction TB
        AB["skill: agent-browser\n(UI testing, scraping)"]
        CC["skill: cc-cursor-cc\n(Cursor ↔ CC handoff)"]
        GOG["skill: gogcli-ops\n(Drive, Sheets, Docs)"]
        NLM["skill: notebook-lm\n(doc export / slides)"]
    end

    %% ── MAIN FLOW ───────────────────────────────────────────
    START --> SETUP --> PLAN --> MODE

    DEC -->|1 task| SOLO
    DEC -->|2–3 tasks| PAR
    DEC -->|4+ tasks| BREEZING

    SOLO --> REVIEW
    PAR  --> REVIEW
    BREEZING --> REVIEW

    REVIEW --> SYNC0
    SYNC0  --> RELEASE

    %% CI recovery can fire after any work phase
    SOLO    -.->|CI red| CI
    PAR     -.->|CI red| CI
    BREEZING-.->|CI red| CI
    CI      -.->|fixed| REVIEW

    %% Domain and utility are invoked inside work phases
    SOLO -.->|feature work| DOMAIN
    PAR  -.->|feature work| DOMAIN
    BREEZING -.->|feature work| DOMAIN
    BREEZING -.->|cross-agent| UTIL

    %% Session layer runs throughout
    SESSION -.->|wraps entire lifecycle| SETUP
    SESSION -.->|wraps entire lifecycle| RELEASE
```

---

## Execution Mode Decision

```mermaid
flowchart LR
    HW["harness-work all"]
    HW --> D{task count\nor flag}
    D -->|1 task\nor --solo| SOLO["SOLO\nWorker → Reviewer"]
    D -->|2–3 tasks\nor --parallel N| PAR["PARALLEL\nWorker A + B → Reviewer"]
    D -->|4+ tasks\nor --breezing| BR["BREEZING\nLead → Workers → Reviewer\n(worktree isolation)"]
    D -->|--codex| COD["CODEX\nDelegate to Codex CLI\n(via codex-plugin-cc)"]
```

---

## Breezing Fix Loop (most common review cycle)

```mermaid
sequenceDiagram
    participant Lead
    participant Worker
    participant Reviewer

    Lead->>Worker: spawn(worktree, task + sprint-contract)
    Worker->>Worker: TDD → implement → self-check → commit
    Worker-->>Lead: done (commit hash)
    Lead->>Reviewer: review(diff, sprint-contract)
    Reviewer-->>Lead: verdict

    alt APPROVE
        Lead->>Lead: cherry-pick to main
        Lead->>Lead: Plans.md → cc:Done [hash]
    else REQUEST_CHANGES (≤ 3 retries)
        Lead->>Worker: SendMessage(critical/major issues)
        Worker->>Worker: fix → git commit --amend
        Worker-->>Lead: done (new hash)
        Lead->>Reviewer: re-review
    end
```

---

## Memory & Session Architecture

```mermaid
flowchart LR
    subgraph L0["Layer 0 · Agent Trace (auto)"]
        AT[".claude/state/agent-trace.jsonl\ntool execution history"]
    end
    subgraph L1["Layer 1 · Project SSOT (skill: memory)"]
        DEC2["decisions.md\n(why decisions were made)"]
        PAT["patterns.md\n(reusable implementation patterns)"]
        SL["session-log.md\n(per-session notes)"]
    end
    subgraph L2["Layer 2 · Unified DB (MCP: harness-mem)"]
        DB["~/.harness-mem/harness-mem.db\nshared: Claude + Codex + OpenCode"]
    end

    AT -->|promoted by session-memory| L1
    L1 -->|sync via /memory sync-across| L2
    L2 -->|recalled via /memory search| L1
```

---

## Skill Catalog by Lifecycle Phase

| Phase | Skill | Trigger |
|-------|-------|---------|
| **Setup** | `harness-setup` | `/harness-setup init` |
| **Planning** | `harness-plan` | `/harness-plan create\|add\|update` |
| **Planning** | `harness-sync` | `/harness-sync` or "where am I?" |
| **Implementation** | `harness-work` | `/harness-work all\|N\|--breezing` |
| **Implementation** | `breezing` | `/breezing all` (alias for team mode) |
| **Implementation** | `auth` | building login / OAuth / payments |
| **Implementation** | `crud` | building data endpoints |
| **Implementation** | `ui` | building components / pages |
| **Implementation** | `deploy` | shipping to Vercel / Netlify |
| **CI Recovery** | `ci` | red build or "diagnose CI" |
| **Review** | `harness-review` | `/harness-review code\|plan\|scope` |
| **Release** | `writing-changelog` | before release, drafting entries |
| **Release** | `harness-release` | `/harness-release patch\|minor\|major` |
| **Memory** | `memory` | `/memory ssot\|sync\|search\|record` |
| **Session** | `session-init` | auto — every session start |
| **Session** | `session-control` | auto — resume / fork flags |
| **Session** | `session-memory` | auto — session end recording |
| **Utility** | `agent-browser` | UI testing, web scraping |
| **Utility** | `cc-cursor-cc` | Cursor ↔ Claude Code handoff |
| **Utility** | `gogcli-ops` | Google Workspace read/write |
| **Utility** | `notebook-lm` | doc export, slide generation |
| **Guidance** | `workflow-guide` | "how does this work?" |
| **Guidance** | `vibecoder-guide` | non-technical orientation |
| **Guidance** | `principles` | coding guidelines reference |

---

## Agent Roles

| Agent | Role | Permissions |
|-------|------|-------------|
| **scaffolder** | Project init, tech-stack detection, state updates | Read / Write / Edit |
| **worker** | TDD, implementation, self-check, git commit | Read / Write / Edit / Bash |
| **reviewer** | Independent verdict against sprint-contract | Read / Grep / Glob only |
| **ci-cd-fixer** | CI failure diagnosis and fix with 3-strike escalation | Read / Write / Edit / Bash |
| **Lead** *(internal)* | Orchestrate phases A→B→C in breezing mode | Spawns Worker + Reviewer |

> See `agents/` for full agent definitions and [hooks/README.md](hooks/README.md) for the hook event map.

---

## Hooks

Hooks are the always-on automation layer — they fire on Claude Code events (PreToolUse, PostToolUse, SessionStart, Stop, etc.) and invoke Go binary handlers or shell scripts without any user action required.

```
Claude Code Event → hooks.json matcher → Go binary (bin/harness) → handler script
```

Key hook groups at a glance:

| Event | What fires |
|-------|-----------|
| **PreToolUse** `Write\|Edit` | Guardrail check, inbox scan, secrets agent |
| **PostToolUse** `Write\|Edit\|Task` | Memory bridge, trace, auto-test, quality-pack, plans-watcher |
| **PostToolUse** `Bash` | Commit cleanup, async CI status check |
| **PermissionRequest** | File-modification guard, test/build validation |
| **SessionStart** | Env check, memory bridge init |
| **Stop / SessionEnd** | Session summary, WIP-task gate, memory finalise |
| **UserPromptSubmit** | Policy injection, command tracking, breezing signal |
| **Pre/PostCompact** | State save, context re-injection |

For the full event map and per-hook script references, see **[hooks/README.md](hooks/README.md)**.
