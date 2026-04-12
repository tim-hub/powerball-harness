# Claude Code Harness v3.10 — X (Twitter) Post Thread

> **Posting time**: Weekdays 19:00-21:00 (golden hour after engineers get home)
> **Target**: Japanese developers interested in AI-assisted development

---

### 1/6

Claude Code Harness, which autonomously operates Claude Code in a "Plan -> Work -> Review" flow, has reached v3.10.

All features from Claude Code 2.1.50 through 2.1.74 have been organized, consolidating 50+ entries into a single Feature Table.

"How should I use this feature?" -- the answer is right here.

github.com/Chachamaru127/claude-code-harness

#ClaudeCode #AIDevTools #Harness

---

### 2/6

Auto Mode Support

Auto Mode started as a Claude Code Research Preview.

Harness has organized the gradual migration from bypassPermissions to Auto Mode.

- Shipped default maintains bypassPermissions
- --auto-mode is opt-in for gradual adoption
- Multi-layer defense combined with Hooks for safe autonomous operation

A realistic bridge to "safe AI autonomous operation."

#ClaudeCode #AutoMode #AIDevTools

---

### 3/6

Agent Teams Evolution

Added matchers to SubagentStart / SubagentStop. You can now individually track the startup and shutdown of Worker, Reviewer, Scaffolder, and Video Generator.

Additionally:
- Automatic task dependency management
- Following official best practice of 5-6 tasks/teammate
- Declarative permissionMode in agent definitions

Team visibility and control taken to the next level.

#ClaudeCode #AgentTeams #Harness

---

### 4/6

Major Developer Experience Improvements

Status Line -- Always displays context usage, cost, and git status. Red warning above 90%.

Checkpointing -- `/rewind` to roll back to any point within a session. Escape from debugging dead ends.

Sandboxing -- OS-level filesystem/network isolation. A complementary layer to bypassPermissions.

Subtle but used every day.

#ClaudeCode #DevEx #AIDevTools

---

### 5/6

Complete Feature Table

How Harness utilizes features from Claude Code 2.1.50 through 2.1.74, consolidated into one table.

50+ entries, over 1,000 lines.

Agent Memory, Worktree, Hooks, Chrome Integration, LSP Integration, 1M Context...

A "Claude Code Feature Dictionary" for developers.

docs/CLAUDE-feature-table.md

#ClaudeCode #Harness #AIDevTools

---

### 6/6

Harness is Self-Referential

This plugin uses Harness itself to improve itself.

`/breezing` launches Agent Teams -> Worker implements -> Reviewer reviews -> writing Harness's own code.

Dogfooding taken to the extreme.

Want to try it?
github.com/Chachamaru127/claude-code-harness

Stars are appreciated!

#ClaudeCode #AIDevTools #Harness

---

## Image Prompt Candidates (for Nano Banana Pro / Gemini image generation)

### Post 1

```
A sleek infographic showing the "Plan -> Work -> Review" cycle as three interconnected nodes in a circular flow diagram. Dark background with blue and purple gradient accents. Text "v3.10" prominently displayed. "50+ Features" badge in the corner. Modern tech aesthetic, clean lines, minimal design. Labels for each node: Plan, Work, Review. 16:9 aspect ratio.
```

### Post 2

```
A diagram showing a migration path from "bypassPermissions" to "Auto Mode" with a bridge metaphor. Left side labeled "Current: bypassPermissions" in blue, right side "Future: Auto Mode" in green, bridge in the middle with safety guardrails. Shield icons representing Hooks defense layers. Dark tech background. 16:9 aspect ratio.
```

### Post 3

```
Four agent icons (Worker, Reviewer, Scaffolder, Video Generator) arranged around a central monitoring dashboard with tracking lines for each role. Each agent has a distinct color: Worker in blue, Reviewer in green, Scaffolder in orange, Video Generator in magenta. Metrics and status indicators floating around each agent. "SubagentStart/Stop" label. Dark background, modern UI style. 16:9 aspect ratio.
```

### Post 4

```
A developer's terminal screen showing three feature panels: Left panel shows a status bar with context usage meter (gradient from green to yellow to red at 90%), cost counter, and git branch. Center panel shows a timeline with rewind points marked as checkpoints. Right panel shows a sandbox container with lock icon. Dark theme terminal aesthetic. "DevEx" header text. 16:9 aspect ratio.
```

### Post 5

```
A large reference book or encyclopedia open to a two-page spread, with a dense feature table visible on the pages. "50+" floating above the book as a badge. Small icons representing different features scattered around: gears, terminal, brain, shield, git branch. Title "Claude Code Feature Dictionary" on the cover. Blue and purple color scheme, dark background. 16:9 aspect ratio.
```

### Post 6

```
An ouroboros (snake eating its own tail) reimagined as a code loop: a plugin icon that points to itself with arrows labeled "improve". The cycle shows: "Harness writes code" -> "Code improves Harness" -> repeat. GitHub star icon in the corner with a sparkle effect. "Self-Referential" label. Minimal, modern design with dark background and accent colors. 16:9 aspect ratio.
```
