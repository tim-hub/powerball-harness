# Claude harness Architecture

## 1. Overview

`claude-code-harness` is a modular, autonomous development framework designed to maximize Claude Code's capabilities. Its central design philosophy is to support the systematic **Plan → Work → Review** development cycle through three main extensions: **Skills**, **Rules**, and **Hooks**.

## 2. Three-Layer Architecture

This plugin adopts the following three-layer architecture to improve reusability and maintainability.

```mermaid
graph TD
    subgraph Profile Layer
        A[profiles/claude-worker.yaml]
    end
    subgraph Workflow Layer
        B[init.yaml, plan.yaml, work.yaml, review.yaml]
    end
    subgraph Skill Layer
        C[30+ SKILL.md files]
    end

    A -- references --> B;
    B -- uses --> C;
```

- **Skill Layer**: Self-contained knowledge units defined as `SKILL.md` files. Each contains specific procedures and knowledge for executing particular tasks (e.g., security review, code implementation).
- **Workflow Layer**: Defined as `*.yaml` files, orchestrating **Skills** for specific development phases (e.g., `/work`). Manages step ordering, conditional branching, error handling, and more.
- **Profile Layer**: Defines the overall plugin behavior. Specifies which workflows map to which commands, which Skill categories are permitted, and so on.

## 3. Directory Structure

```
claude-code-harness/
├── .claude-plugin/         # Plugin metadata
│   ├── plugin.json
│   └── hooks.json
├── skills/                 # Skill definitions (SKILL.md + references/)
│   ├── impl/               # Implementation skills
│   ├── harness-review/     # Review skills
│   ├── verify/             # Verification skills
│   ├── planning/           # Planning skills
│   ├── setup/              # Setup skills
│   ├── ci/                 # CI/CD related skills
│   └── ...                 # 30+ other skills
├── agents/                 # Sub-agent definitions (Markdown)
├── hooks/                  # Hooks definitions (hooks.json)
├── scripts/                # Automation shell scripts
├── docs/                   # Documentation
└── templates/              # Various templates
```

## 4. Key Components

### 4.1. Skills

Each skill explicitly declares a `description` (when to use it) and `allowed-tools` (permitted tools), supporting autonomous discovery and safe execution by Claude.

### 4.2. Rules

Configuration files strictly defined in `claude-code-harness.config.schema.json` enforce safety (`dry-run` mode) and path restrictions (`protected` paths).

### 4.3. Hooks

Defined in `hooks.json`, hooks automatically execute scripts at critical points in the development process.
- **SessionStart**: Environment checks at session startup
- **PostToolUse**: Automatic testing and change tracking after file edits
- **Stop**: Summary generation at session end

### 4.4. Parallel Processing

The `/harness-review` command launches multiple `code-reviewer` sub-agents simultaneously, running security, performance, and quality reviews in parallel to significantly reduce feedback time.
