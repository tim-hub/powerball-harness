# Optional Briefs and Skill Manifest

`harness-plan create` attaches a brief only when needed. Briefs do not replace Plans.md; they are supplementary materials that briefly define implementation assumptions.

## Design Brief

Create a `design brief` for tasks involving UI.

Minimum required content:

- What you want to achieve
- Who will use it
- Important screen states
- Visual and interaction constraints
- Completion criteria

## Contract Brief

Create a `contract brief` for tasks involving APIs.

Minimum required content:

- What is received / returned
- Input validation conditions
- Failure behavior
- External dependencies
- Completion criteria

## Skill Manifest

`scripts/generate-skill-manifest.sh` converts `SKILL.md` frontmatter across the repo into stable JSON.

Use cases:

- Skill surface auditing
- Cross-mirror comparison
- Input for automated docs generation

Output includes:

- `name`
- `description`
- `do_not_use_for`
- `allowed_tools`
- `argument_hint`
- `effort`
- `user_invocable`
- `surface`
- `related_surfaces`

`related_surfaces` also includes mirror information such as `skills-v3`, `skills`, `codex/.codex/skills`, `opencode/skills`.

## Example

```bash
scripts/generate-skill-manifest.sh --output .claude/state/skill-manifest.json
```
