# Image Patterns Reference

Usage guide for the 4 image generation patterns (comparison, concept, flow, highlight).

---

## Overview

Defines image patterns optimized for video scenes. Each pattern is optimized for a specific purpose and works with AI image generation prompt templates.

### Pattern List

| Pattern | Use Case | Optimal Scene | Prompt Template |
|---------|----------|---------------|-----------------|
| **comparison** | Before/After, good/bad example contrast | Problem statement, improvement effect presentation | `templates/image-prompts/comparison.txt` |
| **concept** | Visualizing abstract concepts, hierarchies, relationships | Architecture explanation, concept description | `templates/image-prompts/concept.txt` |
| **flow** | Illustrating steps, processes, workflows | Demo procedures, processing flow | `templates/image-prompts/flow.txt` |
| **highlight** | Emphasizing key points, messages | Hook, CTA, conclusion | `templates/image-prompts/highlight.txt` |

---

## 1. Comparison Pattern {#comparison}

### Purpose

Visually contrast two states or choices such as Before/After or good/bad examples.

### Use Cases

| Scene | Example |
|-------|---------|
| **Problem statement** | Complexity of existing tools vs simplicity of this product |
| **Improvement effect** | Before (manual, slow) vs After (automated, fast) |
| **Feature comparison** | Traditional method vs new feature |
| **Release notes** | Old version vs new version |

### Visual Composition

```
+----------------------------------------------+
|                                              |
|  [Bad example/Before]  →  [Good example/After] |
|                                              |
|  ❌ Issue 1              ✅ Improvement 1     |
|  ❌ Issue 2              ✅ Improvement 2     |
|  ❌ Issue 3              ✅ Improvement 3     |
|                                              |
+----------------------------------------------+
```

### JSON Example

```json
{
  "type": "comparison",
  "topic": "Task management improvement",
  "style": "modern",
  "colorScheme": {
    "primary": "#3B82F6",
    "secondary": "#10B981",
    "background": "#1F2937"
  },
  "comparison": {
    "leftSide": {
      "label": "Before",
      "items": [
        "Manual spreadsheet management",
        "Frequent update omissions",
        "30 minutes to grasp status"
      ],
      "icon": "x",
      "sentiment": "negative"
    },
    "rightSide": {
      "label": "After",
      "items": [
        "Automatic dashboard updates",
        "Real-time sync",
        "Status visible at a glance"
      ],
      "icon": "check",
      "sentiment": "positive"
    },
    "divider": "arrow"
  }
}
```

### Prompt Generation Tips

- **Left side (Before/Bad example)**: Red tones, warning icons, cluttered appearance
- **Right side (After/Good example)**: Green tones, check icons, organized appearance
- **Divider**: Clear arrow or "VS" for visual separation
- **Text**: Short and specific (recommended 20 characters or less per item)

### Patterns to Avoid

| Avoid | Recommended |
|-------|-------------|
| Long text lists | Short keywords |
| Abstract descriptions | Specific numbers/results |
| Ambiguous evaluations | Clear contrast |
| Same icons on both sides | Different emotion icons |

---

## 2. Concept Pattern {#concept}

### Purpose

Visually represent abstract concepts, hierarchical structures, and relationships between elements.

### Use Cases

| Scene | Example |
|-------|---------|
| **Architecture explanation** | System architecture diagrams, layer structure |
| **Concept description** | Philosophy, design principles, value proposition illustration |
| **Relationships** | Dependencies between components |
| **Process overview** | Ecosystem, overall workflow |

### Visual Composition (Hierarchy Example)

```
        +-------------+
        |   Top level  |
        +------+------+
               |
      +--------+--------+
      |                 |
+-----v-----+     +-----v-----+
|  Level 1  |     |  Level 1  |
+-----------+     +-----+-----+
                        |
                  +-----v-----+
                  |  Level 2  |
                  +-----------+
```

### JSON Example

```json
{
  "type": "concept",
  "topic": "Microservices architecture",
  "style": "technical",
  "colorScheme": {
    "primary": "#6366F1",
    "secondary": "#8B5CF6",
    "background": "#0F172A"
  },
  "concept": {
    "elements": [
      {
        "id": "api-gateway",
        "label": "API Gateway",
        "description": "Entry point for all requests",
        "level": 0,
        "icon": "cloud",
        "emphasis": "high"
      },
      {
        "id": "auth-service",
        "label": "Auth Service",
        "level": 1,
        "parentId": "api-gateway",
        "icon": "server",
        "emphasis": "medium"
      },
      {
        "id": "data-service",
        "label": "Data Service",
        "level": 1,
        "parentId": "api-gateway",
        "icon": "database",
        "emphasis": "medium"
      }
    ],
    "relationships": [
      {
        "from": "api-gateway",
        "to": "auth-service",
        "label": "Auth check",
        "type": "flow"
      },
      {
        "from": "api-gateway",
        "to": "data-service",
        "label": "Data retrieval",
        "type": "flow"
      }
    ],
    "layout": "hierarchy"
  }
}
```

### Layout Types

| Layout | Use Case | Visual Image |
|--------|----------|--------------|
| **hierarchy** | Hierarchical structure (org chart, dependencies) | Top-down tree |
| **radial** | Radial from center (ecosystem) | Main element at center, related elements around |
| **grid** | Parallel arrangement (category classification) | Matrix layout |
| **flow** | Processing flow (pipeline) | Left-to-right flow |
| **circular** | Cyclic process (lifecycle) | Circular ring |

### Prompt Generation Tips

- **Element count**: 2-10 (too many becomes hard to read)
- **Levels**: Maximum 3-4 levels
- **Icons**: Intuitively represent element characteristics
- **Relationships**: Express importance through arrow thickness and color

### Patterns to Avoid

| Avoid | Recommended |
|-------|-------------|
| 10+ elements | Keep to 7 or fewer |
| Complex relationship lines | Only key relationships |
| Long description text | Short labels + icons |
| Same-looking elements | Differentiate by emphasis level |

---

## 3. Flow Pattern {#flow}

### Purpose

Visualize procedures, processes, and workflows in chronological or step order.

### Use Cases

| Scene | Example |
|-------|---------|
| **Demo procedures** | Steps from setup to execution |
| **User flow** | Login → operation → completion flow |
| **Processing flow** | Data pipeline, CI/CD flow |
| **Onboarding** | First-time user journey |

### Visual Composition (Horizontal Example)

```
[1. Start] ──▶ [2. Input] ──▶ [3. Process] ──▶ [4. Complete]
   ⏱2min         ⏱1min         ⏱3sec         Instant
```

### JSON Example

```json
{
  "type": "flow",
  "topic": "Video generation flow",
  "style": "modern",
  "colorScheme": {
    "primary": "#F59E0B",
    "secondary": "#EF4444",
    "background": "#111827"
  },
  "flow": {
    "steps": [
      {
        "id": "analyze",
        "label": "Codebase analysis",
        "description": "Auto-detect project structure",
        "order": 1,
        "type": "start",
        "icon": "circle",
        "duration": "10 sec"
      },
      {
        "id": "plan",
        "label": "Scenario generation",
        "description": "Propose optimal video composition",
        "order": 2,
        "type": "process",
        "icon": "square",
        "duration": "20 sec"
      },
      {
        "id": "generate",
        "label": "Parallel generation",
        "description": "Create each scene simultaneously",
        "order": 3,
        "type": "parallel",
        "icon": "rounded",
        "duration": "2 min"
      },
      {
        "id": "render",
        "label": "Rendering",
        "description": "Output final video",
        "order": 4,
        "type": "end",
        "icon": "hexagon",
        "duration": "30 sec"
      }
    ],
    "direction": "horizontal",
    "arrowStyle": "solid",
    "showNumbers": true
  }
}
```

### Step Types

| Type | Use Case | Visual Representation |
|------|----------|----------------------|
| **start** | Flow start point | Circle icon, green |
| **process** | Normal processing step | Square, blue |
| **decision** | Conditional branch | Diamond, yellow |
| **parallel** | Parallel processing | Multiple icons, purple |
| **subprocess** | Sub-flow | Rounded square |
| **end** | Flow end point | Double circle, red |

### Prompt Generation Tips

- **Direction**: Horizontal is most readable (for English audiences)
- **Step count**: 2-10 steps (too many becomes complex)
- **Duration**: Showing time per step is practical
- **Numbers**: Indicate order explicitly (showNumbers: true)

### Patterns to Avoid

| Avoid | Recommended |
|-------|-------------|
| 10+ steps | Consolidate to 7 or fewer |
| Complex branching | Simplify to linear flow |
| Long step names | Concise verb + noun |
| Unclear ordering | Make explicit with order field |

---

## 4. Highlight Pattern {#highlight}

### Purpose

Emphasize a single message, keyword, or numeric value.

### Use Cases

| Scene | Example |
|-------|---------|
| **Hook (opening)** | "Tired of manual work?" |
| **CTA (call to action)** | "Try it now" |
| **Conclusion** | "3x faster, 10x simpler" |
| **Key metrics** | "95% time reduction" |

### Visual Composition

```
+----------------------------------------+
|                                        |
|                                        |
|        ⚡ 3x faster, 10x simpler ⚡    |
|                                        |
|    Development experience transformed  |
|           by automation                |
|                                        |
+----------------------------------------+
```

### JSON Example

```json
{
  "type": "highlight",
  "topic": "Product value emphasis",
  "style": "gradient",
  "colorScheme": {
    "primary": "#EC4899",
    "accent": "#8B5CF6",
    "background": "#18181B"
  },
  "highlight": {
    "mainText": "95% time reduction",
    "subText": "Development teams freed from manual work",
    "icon": "rocket",
    "position": "center",
    "effect": "glow",
    "fontSize": "xlarge",
    "emphasis": "high"
  }
}
```

### Effect Types

| Effect | Use Case | Visual Representation |
|--------|----------|----------------------|
| **glow** | Radiant emphasis (CTA, conclusion) | Glow effect |
| **shadow** | Subtle emphasis (Hook) | Drop shadow |
| **gradient** | Modern impression | Gradient background |
| **outline** | Sharp impression | Outline only |
| **none** | Minimal | No decoration |

### Icons and Emotions

| Icon | Emotion/Meaning | Use Case |
|------|-----------------|----------|
| **star** | Excellence, quality | Feature introduction, ratings |
| **check** | Completion, success | Implementation effect, results |
| **alert** | Attention | Problem statement, warnings |
| **trophy** | Achievement, victory | Results, accomplishments |
| **rocket** | Speed, innovation | Performance, new features |
| **fire** | Popularity, trending | Trends, attention |
| **bolt** | Instant, power | Speed, efficiency |

### Prompt Generation Tips

- **Brevity is key**: Main text ideally 10 characters or less
- **Numbers**: Specific numbers are more persuasive ("95%", "3x")
- **Contrast**: Pair two values like "faster, simpler"
- **Emotion**: Amplify emotion with icon + effect

### Patterns to Avoid

| Avoid | Recommended |
|-------|-------------|
| Long text (20+ characters) | Short catchcopy |
| Multiple claims | Focus on one |
| Plain design | Stand out with effects |
| Small font | xlarge recommended |

---

## Pattern Selection Guide

### Recommended Patterns by Scene Type

| Scene Type | Primary | Secondary | Purpose |
|------------|---------|-----------|---------|
| **Hook** | highlight | comparison | Strong first impression |
| **Problem** | comparison | concept | Clearly show current issues |
| **Solution** | concept | flow | Mechanism of the solution |
| **Demo** | flow | comparison | Visualize procedures |
| **Differentiator** | comparison | concept | Differentiation points |
| **CTA** | highlight | - | Call to action |

### Usage Frequency by Funnel

| Pattern | Awareness/Interest | Consideration | Conviction | Retention |
|---------|-------------------|---------------|------------|-----------|
| **comparison** | ★★★ | ★★★ | ★★☆ | ★☆☆ |
| **concept** | ★☆☆ | ★★★ | ★★★ | ★★☆ |
| **flow** | ★★☆ | ★★★ | ★★☆ | ★★★ |
| **highlight** | ★★★ | ★★☆ | ★★★ | ★☆☆ |

### Combining Multiple Patterns

**Example for 90-second teaser (LP/ads)**:

| Time | Scene | Pattern | Content |
|------|-------|---------|---------|
| 0-5 sec | Hook | **highlight** | "Tired of manual work?" |
| 5-15 sec | Problem | **comparison** | Before (manual) vs After (automated) |
| 15-55 sec | Solution | **flow** | 3 steps: Setup → Execute → Complete |
| 55-70 sec | Proof | **concept** | Architecture robustness |
| 70-90 sec | CTA | **highlight** | "Start free now" |

---

## Implementation Notes

### 1. JSON Schema Validation

- **Required**: `type`, `topic` fields are required
- **oneOf**: Pattern-specific field required based on type (e.g., type="comparison" requires comparison field)
- **Validation**: Verify with `scripts/validate-visual-pattern.js`

### 2. Integration with Prompt Templates

- **Templates**: Use `templates/image-prompts/{type}.txt`
- **Placeholders**: Replace `{{topic}}`, `{{items}}`, `{{style}}` etc. with JSON values
- **Generation**: `references/image-generator.md` handles actual generation

### 3. Image Quality Check

- **Auto assessment**: Quality evaluation via `references/image-quality-check.md`
- **Retry**: Up to 3 regeneration attempts on failure
- **Determinism**: Save seed values for reproducibility

### 4. Asset Management

- **Output**: `out/video-{id}/assets/generated/`
- **Manifest**: Recorded in `assets.manifest.schema.json`
- **Hash**: SHA-256 for tamper detection

---

## Related Documents

- [visual-patterns.schema.json](../schemas/visual-patterns.schema.json) - JSON Schema definitions
- [image-generator.md](./image-generator.md) - AI image generation implementation
- [image-quality-check.md](./image-quality-check.md) - Quality assessment logic
- [templates/image-prompts/](../templates/image-prompts/) - Prompt templates
- [best-practices.md](./best-practices.md) - Overall video best practices

---

**Created**: 2026-02-02
**Target Phase**: Phase 6 - Image Generation Patterns
**Maintenance**: Update when schema changes
