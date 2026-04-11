# Video Analyzer - Codebase Analysis Engine

Automatically analyzes projects and extracts information needed for video generation.

---

## Overview

This is the analysis engine executed in Step 1 of `/generate-video`.
It parses the codebase and project assets to determine the optimal video composition.

## Analysis Items

### 1. Framework Detection

| Detection Target | Detection Method |
|------------------|-----------------|
| Next.js | Presence of `next.config.*` |
| React | `package.json` dependencies |
| Vue | `vue.config.*` or `nuxt.config.*` |
| Svelte | `svelte.config.*` |
| Express/Fastify | `package.json` dependencies |

**Execution Commands**:
```bash
# Extract dependencies from package.json
cat package.json | jq '.dependencies, .devDependencies'

# Check for config file existence
ls -la *.config.* 2>/dev/null
```

### 2. Key Feature Detection

| Feature | Detection Pattern |
|---------|-------------------|
| Auth | `auth/`, `login/`, `@clerk`, `@auth0`, `supabase` |
| Payments | `payment/`, `billing/`, `stripe`, `@stripe` |
| Dashboard | `dashboard/`, `admin/`, `analytics` |
| API | `api/`, `routes/`, `trpc`, `graphql` |
| DB | `prisma/`, `drizzle/`, `@supabase` |

**Execution Commands**:
```bash
# Infer features from directory structure
find src app -type d -name "auth" -o -name "login" -o -name "dashboard" 2>/dev/null

# Infer features from packages
grep -E "clerk|stripe|supabase|prisma" package.json
```

### 3. UI Component Detection

| Item | Detection Method |
|------|-----------------|
| Page count | Count of `app/**/page.tsx` or `pages/**/*.tsx` |
| Component count | Count of `components/**/*.tsx` |
| UI library | Detection of `shadcn`, `radix`, `chakra`, `mui` |

**Execution Commands**:
```bash
# Count pages
find . -name "page.tsx" -o -name "page.jsx" 2>/dev/null | wc -l

# Count components
find . -path "*/components/*" -name "*.tsx" 2>/dev/null | wc -l
```

### 4. Project Asset Analysis

| Asset | Use |
|-------|-----|
| `package.json` | Project name, description |
| `README.md` | Project overview, tagline |
| `Plans.md` | Completed tasks (for release notes) |
| `CHANGELOG.md` | Changes (for release notes) |
| `.claude/memory/decisions.md` | Technical decisions (for architecture explanation) |

**Execution Commands**:
```bash
# Extract project info
cat package.json | jq '{name, description, version}'

# Extract first paragraph of README
head -20 README.md
```

---

## Automatic Video Type Detection

### Detection Logic

```
Determine video type from analysis results:
    |
    +-- CHANGELOG recently updated (within 7 days)
    |   +-- → Release notes video
    |
    +-- Large structural changes (new directory additions, etc.)
    |   +-- → Architecture explanation
    |
    +-- Many UI changes (component additions/modifications)
    |   +-- → Product demo
    |
    +-- Matches multiple conditions
        +-- → Composite video (confirm with user)
```

### Detection Criteria

| Type | Condition |
|------|-----------|
| **Release Notes** | `git log --since="7 days ago"` contains tags/releases |
| **Architecture** | New `src/*/` directories, major refactoring |
| **Product Demo** | UI component additions/modifications |
| **Default** | Product demo (most general-purpose) |

---

## Output Format

Analysis results output in the following format:

```yaml
project:
  name: "MyAwesomeApp"
  description: "Task management made easy"
  version: "1.2.0"

framework:
  primary: "Next.js"
  ui_library: "shadcn/ui"

features:
  - name: "Auth"
    type: "auth"
    path: "src/app/(auth)/"
    provider: "Clerk"
  - name: "Dashboard"
    type: "dashboard"
    path: "src/app/dashboard/"
  - name: "API"
    type: "api"
    path: "src/app/api/"

stats:
  pages: 12
  components: 45
  api_routes: 8

recent_changes:
  changelog_updated: true
  last_release: "2026-01-20"
  major_changes:
    - "Auth flow added"
    - "Dashboard improvements"

recommended_video_type: "release-notes"
confidence: 0.85
```

---

## Execution Example

```
📊 Analyzing project...

✅ Analysis complete

| Item | Result |
|------|--------|
| Project name | MyAwesomeApp |
| Framework | Next.js 14 |
| UI library | shadcn/ui |
| Pages | 12 |
| Components | 45 |

🔍 Detected features:
- Auth (Clerk)
- Dashboard
- API (8 endpoints)

📋 Recent changes:
- v1.2.0 release (3 days ago)
- Auth flow added
- Dashboard improvements

🎬 Recommended video type: Release notes video
   Reason: Recent release with major feature additions
```

---

## Notes

- Analysis is non-destructive (does not modify files)
- Completes in seconds even for large projects
- Undetected features can be added manually (in planner.md)
