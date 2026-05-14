# Path Conventions for SKILL.md and Shell Scripts

All paths in skills and scripts must be clearly anchored to one of three tiers.
Using the wrong tier causes fragile CWD-dependent paths or incorrect root resolution.

## The Three Tiers

| Tier | What it covers | Anchor in SKILL.md | Anchor in shell scripts |
|------|---------------|-------------------|------------------------|
| **skill-local** | Files inside this skill's own directory (`scripts/`, `references/`) | `${CLAUDE_SKILL_DIR}/...` | `"$(dirname "${BASH_SOURCE[0]}")/..."` |
| **plugin-local** | Files elsewhere in the harness plugin (`docs/`, `harness/scripts/`, other skills) | `${CLAUDE_SKILL_DIR}/../../...` | Derive `PLUGIN_ROOT` from script location |
| **project-root** | Files in the user's repo (`CHANGELOG.md`, `Plans.md`, `VERSION`, `local-scripts/`) | Plain relative paths or describe in prose | `git rev-parse --show-toplevel` |

## Rules

### In SKILL.md

**Skill-local references and scripts** — always use `${CLAUDE_SKILL_DIR}`:

```bash
# ✅ skill-local
bash "${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh"
```
```markdown
<!-- ✅ skill-local link -->
See [phases.md](${CLAUDE_SKILL_DIR}/references/phases.md)
```

**Plugin-local files** — use `${CLAUDE_SKILL_DIR}/../../` (skills are always exactly two levels below plugin root: `skills/<name>/`):

```bash
# ✅ plugin-local
bash "${CLAUDE_SKILL_DIR}/../../scripts/validate-release-notes.sh"
```
```markdown
<!-- ✅ plugin-local link -->
See [release docs](${CLAUDE_SKILL_DIR}/../../docs/release-preflight.md)
```

**Project-root files** — use plain names or annotate with `(project-root)` / `(repo-root)`:

```bash
# ✅ project-root — CHANGELOG.md is in the user's repo root, not the plugin
cat CHANGELOG.md

# ✅ project-root — dev-only scripts at repo root
bash local-scripts/check-consistency.sh  # (repo-root local-scripts/)
```

**Never use bare relative paths for skill-local or plugin-local scripts in bash code blocks:**

```bash
# ❌ fragile — CWD-dependent, breaks when Claude's working directory differs
bash skills/harness-release/scripts/release-preflight.sh
bash ./skills/harness-release/scripts/sync-version.sh bump
```

### In Shell Scripts

**Skill-local** — resolve relative to the script itself:

```bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/utils.sh"  # skill-local: utility in same scripts/ dir
```

**Plugin-local** — derive plugin root from script location (scripts live at `skills/<name>/scripts/`, so `../../../` reaches plugin root):

```bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "${SCRIPT_DIR}/../../../" && pwd)"  # plugin-local: harness plugin root
source "${PLUGIN_ROOT}/scripts/lib/common.sh"
```

**Project-root** — always use `git rev-parse --show-toplevel`:

```bash
PROJECT_ROOT="$(git rev-parse --show-toplevel)"  # project-root: user's git repo
CHANGELOG="${PROJECT_ROOT}/CHANGELOG.md"          # project-root: CHANGELOG lives at repo root
```

Never derive the user's project root from the script's own location — the plugin may be installed at an arbitrary path relative to the user's repo.

## Naming Convention for Comments

Tag the tier in code comments so reviewers can verify correctness at a glance:

```bash
PLUGIN_ROOT="..."       # plugin-local: harness plugin root
GIT_ROOT="..."          # project-root: user's git repository root
source "lib/utils.sh"   # skill-local: utility in this skill's scripts/
```

## Depth Assumption

Skills are always at `harness/skills/<name>/SKILL.md` — exactly **two levels** below the plugin root. The `${CLAUDE_SKILL_DIR}/../../` traversal relies on this. If the directory structure ever nests skills deeper, all plugin-local links in SKILL.md files must be updated.
