# Video Planner - Scenario Planner

Automatically proposes scene composition from analysis results, then confirms and adjusts with the user.

---

## Overview

This is the scenario planner executed in Step 2 of `/generate-video`.
It receives output from analyzer.md and proposes the optimal scene composition.

> **Important**: Scene composition must follow the funnel-specific guidelines in [best-practices.md](best-practices.md)

## Input

Analysis results from analyzer.md:
- Project information (name, description)
- Detected feature list
- Recommended video type
- Recent changes

---

## Funnel-Specific Template Selection

### Step 0: Confirm Purpose (Required)

Confirm the video's purpose and select the appropriate template.

| Purpose (Funnel) | Video Type | Target Length | Composition Core |
|-------------------|------------|---------------|------------------|
| Awareness to Interest | LP/Ad Teaser | 30-90 sec | Pain → Result → CTA |
| Interest → Consideration | Intro Demo | 2-3 min | Complete 1 use case |
| Consideration → Conviction | Demo/Release Notes | 2-5 min | Preemptively address objections |
| Conviction → Decision | Walkthrough | 5-30 min | Real operations + evidence |
| Retention/Utilization | Onboarding | 30 sec - few min | Shortest path to Aha moment |

### 90-Second Teaser Template

**Use case**: LP/Ads, Awareness to Interest funnel

```
0:00-0:05 (150f)  → HookScene: Pain or desired outcome
0:05-0:15 (300f)  → ProblemPromise: Target user and promise
0:15-0:55 (1200f) → WorkflowDemo: Signature workflow
0:55-1:10 (450f)  → Differentiator: Basis for differentiation
1:10-1:30 (600f)  → CTA: Next step
```

### 3-Minute Intro Demo Template

**Use case**: Consideration, Interest → Consideration funnel

```
0:00-0:10 (300f)  → Hook: Conclusion + pain
0:10-0:30 (600f)  → UseCase: Use case declaration
0:30-2:20 (3300f) → Demo: Complete walkthrough on real screen
2:20-2:50 (900f)  → Objection: Address one common concern
2:50-3:00 (300f)  → CTA: Call to action
```

### 20-Minute Walkthrough Template

**Use case**: Decision, Conviction → Decision funnel

```
0:00-1:00   → Intro: Target audience and challenges
1:00-8:00   → BasicFlow: Basic flow
8:00-12:00  → Objections: Top 2 objections
12:00-15:00 → Security: Management/security
15:00-20:00 → CaseStudy+CTA: Success stories + CTA
```

## Scene Templates

### Common Scenes

| Scene | Recommended Duration | Content | Required |
|-------|---------------------|---------|----------|
| **Intro** | 3-5 sec | Logo + tagline + fade in | Yes |
| **CTA** | 3-5 sec | URL + contact + fade out | Yes |

### Product Demo Scenes

| Scene | Recommended Duration | Content |
|-------|---------------------|---------|
| **Feature Introduction** | 5-10 sec | Feature name + one-line description |
| **UI Demo** | 10-30 sec | Playwright capture |
| **Highlight** | 5-10 sec | Emphasize key characteristics |

### Architecture Explanation Scenes

| Scene | Recommended Duration | Content |
|-------|---------------------|---------|
| **Overview Diagram** | 5-10 sec | Overall architecture Mermaid diagram |
| **Detailed Explanation** | 10-20 sec | Zoom into each component |
| **Data Flow** | 10-15 sec | Sequence diagram animation |

### Release Notes Scenes

| Scene | Recommended Duration | Content |
|-------|---------------------|---------|
| **Version Display** | 3-5 sec | vX.Y.Z + release date |
| **Change List** | 5-15 sec | Added/Changed/Fixed animation |
| **Before/After** | 10-20 sec | Side-by-side UI change comparison |
| **New Feature Demo** | 10-30 sec | UI demo of added features |

---

## Scenario Generation Logic

### Step 1: Template Selection by Video Type

```
Select base template based on recommended video type:
    |
    +-- LP/Ad Teaser (30-90 sec)
    |   +-- Hook → ProblemPromise → WorkflowDemo → Differentiator → CTA
    |
    +-- Intro Demo (2-3 min)
    |   +-- Hook → UseCase Declaration → Real Screen Demo → Objection → CTA
    |
    +-- Release Notes (1-3 min)
    |   +-- Hook → Version → Before/After → New Feature Demo → CTA
    |
    +-- Architecture Explanation (5-30 min)
    |   +-- Intro → Overview Diagram → Detailed Explanation x N → Data Flow → Management/Security → CTA
    |
    +-- Onboarding (30 sec - few min)
        +-- Welcome → Quick Win → Next Steps
```

**Key Principles**:
- Don't show logo or company intro for too long at the start (prevent drop-off)
- Place CTAs not only at the end but also in the middle
- Tell a "pain → solution" story, not a feature list

### Step 2: Generate Scenes from Detected Features

```python
# Pseudocode
for feature in detected_features:
    if feature.type == "auth":
        add_scene("Auth Flow Demo", duration=15, source="playwright")
    elif feature.type == "dashboard":
        add_scene("Dashboard Introduction", duration=20, source="playwright")
    elif feature.type == "api":
        add_scene("API Overview", duration=10, source="mermaid")
```

### Step 3: Duration Optimization

| Video Length | Recommended Use | Scene Count Guide |
|-------------|-----------------|-------------------|
| 15 sec | Social media ads | 3-4 |
| 30 sec | Short videos | 5-6 |
| 60 sec | Standard demo | 8-10 |
| 2-3 min | Detailed explanation | 15-20 |

---

## User Confirmation Flow

### Proposal Display

```markdown
🎬 Scenario Plan

**Video Type**: Product Demo
**Total Duration**: 45 sec

| # | Scene | Duration | Content | Source |
|---|-------|----------|---------|--------|
| 1 | Intro | 5 sec | MyApp - Task management made easy | Template |
| 2 | Auth Flow | 15 sec | Login screen demo | Playwright |
| 3 | Dashboard | 20 sec | Main feature introduction | Playwright |
| 4 | CTA | 5 sec | myapp.com | Template |

Is this composition acceptable?
1. OK, start generation
2. I want to edit
3. Cancel
```

### AskUserQuestion Implementation

```
AskUserQuestion:
  question: "Would you like to generate the video with this scenario?"
  header: "Scenario Confirmation"
  options:
    - label: "OK, start generation"
      description: "Generate the video with this scene composition"
    - label: "I want to edit"
      description: "Add/remove/modify scenes"
    - label: "Cancel"
      description: "Cancel video generation"
```

### Edit Mode

When the user selects "I want to edit":

```markdown
📝 Scenario Editor

You can edit with the following commands:

- **Add**: "Add a demo of Feature X"
- **Remove**: "Remove scene 2"
- **Modify**: "Shorten intro to 3 seconds"
- **Swap**: "Swap scenes 2 and 3"
- **Done**: "This is fine"

What would you like to edit?
```

---

## Output Format

planner.md output (input to generator.md):

```yaml
video:
  type: "product-demo"
  total_duration: 45
  resolution: "1080p"
  fps: 30

scenes:
  - id: 1
    name: "intro"
    duration: 5
    template: "intro"
    content:
      title: "MyApp"
      tagline: "Task management made easy"
      logo: "public/logo.svg"

  - id: 2
    name: "auth-demo"
    duration: 15
    template: "ui-demo"
    source: "playwright"
    content:
      url: "http://localhost:3000/login"
      actions:
        - click: "[data-testid=email-input]"
        - type: "user@example.com"
        - click: "[data-testid=login-button]"

  - id: 3
    name: "dashboard"
    duration: 20
    template: "ui-demo"
    source: "playwright"
    content:
      url: "http://localhost:3000/dashboard"
      actions:
        - wait: 1000
        - scroll: "down"

  - id: 4
    name: "cta"
    duration: 5
    template: "cta"
    content:
      url: "https://myapp.com"
      text: "Try it now"
```

---

## Notes

- If there are too many scenes, lower-priority ones are automatically excluded from the proposal
- Users can manually add scenes
- Scenes with Playwright source require the app to be running
