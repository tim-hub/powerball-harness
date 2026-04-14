# Direction Guide

Defines the usage and best practices for the visual direction system in the generate-video skill.

---

## Overview

The direction system consists of the following 4 elements:

| Element | Role | Control |
|---------|------|---------|
| **transition** | Scene transitions | Fade, slide, zoom, cut |
| **emphasis** | Element highlighting | 3 levels of emphasis + sound effects |
| **background** | Background design | 5 background styles |
| **timing** | Timing adjustment | Wait times, audio offsets |

---

## Transition

### 4 Transition Types

| Type | Use Case | Visual Effect | Recommended Duration |
|------|----------|---------------|---------------------|
| **fade** | General-purpose transition | Smooth fade in/out | 500ms (15f) |
| **slideIn** | Moving to next topic | Directional slide (left/right/top/bottom) | 400ms (12f) |
| **zoom** | Drawing attention to detail | Zoom in/out | 600ms (18f) |
| **cut** | Instant transition | Cut (instantaneous) | 0ms |

### Usage Guidelines

#### fade
- **Recommended scenes**: General purpose, section start, calm transitions
- **Effect**: Visually gentle, doesn't draw too much attention
- **Examples**:
  - Intro → main explanation
  - Feature explanation → next feature explanation
  - Calm transition before CTA

```json
{
  "transition": {
    "type": "fade",
    "duration_ms": 500,
    "easing": "easeInOut"
  }
}
```

#### slideIn
- **Recommended scenes**: Topic change, comparison display, step progression
- **Effect**: Dynamic, creates anticipation for next content
- **direction**:
  - `right`: Sense of forward progress (next step)
  - `left`: Past reference (the "Before" in Before/After)
  - `top`: Important information appearing
  - `bottom`: Supplementary information added

```json
{
  "transition": {
    "type": "slideIn",
    "duration_ms": 400,
    "direction": "right",
    "easing": "easeOut"
  }
}
```

#### zoom
- **Recommended scenes**: Detail display, emphasis, impactful information
- **Effect**: Attention-drawing, impact
- **Examples**:
  - Displaying important numbers
  - Presenting core issues
  - Emphasizing differentiation points

```json
{
  "transition": {
    "type": "zoom",
    "duration_ms": 600,
    "easing": "easeInOut"
  }
}
```

#### cut
- **Recommended scenes**: Demo operations, fast-paced content, tension
- **Effect**: Instantaneous, speeds up tempo
- **Examples**:
  - Between UI operation steps
  - Fast demonstrations
  - Rhythmic feature introductions

```json
{
  "transition": {
    "type": "cut",
    "duration_ms": 0
  }
}
```

### Recommended Transitions by Funnel Stage

| Funnel Stage | Recommended Transitions | Reason |
|-------------|------------------------|--------|
| Awareness (LP/Ads) | fade, zoom | Calm, impact |
| Interest (Intro) | slideIn, fade | Dynamic, anticipation |
| Consideration (Feature Demo) | cut, slideIn | Tempo, efficiency |
| Conviction (Architecture) | fade, zoom | Detail, trust |
| Retention (Onboarding) | slideIn, cut | Step progression |

---

## Emphasis

### 3 Emphasis Levels

| Level | Use Case | Visual Effect | Recommended Sound |
|-------|----------|---------------|-------------------|
| **high** | Most important message | Large animation, bright color | whoosh, chime |
| **medium** | Important points | Medium animation, accent color | pop |
| **low** | Supplementary info | Subtle emphasis, light color | none, ding |

### Usage Guidelines

#### high (High Emphasis)
- **Recommended scenes**:
  - Hook (initial impact)
  - CTA (call to action)
  - Differentiator
  - Surprising results/numbers

- **Visual effects**:
  - Text size: extra large
  - Color: vivid (default: `#00F5FF` cyan)
  - Animation: scale 1.2, bounce
  - Sound: `whoosh` or `chime`

- **Examples**:
  - "3x faster" → high emphasis
  - "Try it free now" → high emphasis

```json
{
  "emphasis": {
    "level": "high",
    "text": ["3x faster"],
    "sound": "whoosh",
    "color": "#00F5FF",
    "position": "center"
  }
}
```

#### medium (Medium Emphasis)
- **Recommended scenes**:
  - Key points of feature explanation
  - Workflow steps
  - Problem presentation
  - Solution

- **Visual effects**:
  - Text size: large
  - Color: accent (default: `#FFC700` gold)
  - Animation: scale 1.1, fade-in
  - Sound: `pop`

- **Examples**:
  - "Step 1: Setup" → medium emphasis
  - "Having this problem?" → medium emphasis

```json
{
  "emphasis": {
    "level": "medium",
    "text": ["Step 1: Setup"],
    "sound": "pop",
    "color": "#FFC700",
    "position": "top"
  }
}
```

#### low (Low Emphasis)
- **Recommended scenes**:
  - Supplementary information
  - Light introduction of additional features
  - Annotations
  - Links to detailed information

- **Visual effects**:
  - Text size: normal
  - Color: light (default: `#A8DADC` light blue)
  - Animation: fade-in only
  - Sound: `none` or `ding`

- **Examples**:
  - "*See documentation for details" → low emphasis
  - "Many more features" → low emphasis

```json
{
  "emphasis": {
    "level": "low",
    "text": ["*See documentation for details"],
    "sound": "none",
    "color": "#A8DADC",
    "position": "bottom"
  }
}
```

### Choosing Sound Effects

| Sound | Characteristics | Recommended Use |
|-------|----------------|-----------------|
| **whoosh** | Wind sound, dynamic | high emphasis, screen transitions |
| **chime** | Chime, beautiful tone | CTA, success display |
| **pop** | Pop, light and lively | medium emphasis, button display |
| **ding** | Small bell sound | low emphasis, light notification |
| **none** | Silent | Quiet information, continuous display |

### Recommended Emphasis Levels by Funnel Stage

| Funnel Stage | Primary Emphasis | Secondary Emphasis |
|-------------|-----------------|-------------------|
| Awareness (LP/Ads) | high (frequent) | medium (moderate) |
| Interest (Intro) | high (1-2 times) | medium (frequent) |
| Consideration (Feature Demo) | medium (primary) | low (supplementary) |
| Conviction (Architecture) | medium (moderate) | low (frequent) |
| Retention (Onboarding) | high (goals) | medium (steps) |

---

## Background

### 5 Background Styles

| Type | Visual Characteristics | Use Case | Color Examples |
|------|----------------------|----------|----------------|
| **cyberpunk** | Neon, grid, futuristic | Tech, cutting-edge appeal | `#0a0e27` + `#00f5ff` |
| **corporate** | Refined, trustworthy, professional | B2B, enterprise | `#1a1a2e` + `#16213e` |
| **minimal** | Simple, clean, focused | Explanation-focused, documentation | `#ffffff` + `#f0f0f0` |
| **gradient** | Colorful, dynamic, approachable | B2C, casual | `#667eea` → `#764ba2` |
| **particles** | Dynamic particles, energetic | Hook, CTA, impact | `#000000` + particles |

### Usage Guidelines

#### cyberpunk
- **Recommended scenes**:
  - Demonstrating technological advancement
  - Developer tools
  - AI/ML feature introduction
  - Architecture diagrams

- **Characteristics**:
  - Neon grid
  - Glitch effects
  - Blue/cyan color scheme

```json
{
  "background": {
    "type": "cyberpunk",
    "primaryColor": "#0a0e27",
    "secondaryColor": "#00f5ff",
    "opacity": 0.9
  }
}
```

#### corporate
- **Recommended scenes**:
  - B2B products
  - Enterprise features
  - Security/reliability appeal
  - Case studies/results

- **Characteristics**:
  - Dark blue tones
  - Clean gradients
  - Calm atmosphere

```json
{
  "background": {
    "type": "corporate",
    "primaryColor": "#1a1a2e",
    "secondaryColor": "#16213e",
    "opacity": 1
  }
}
```

#### minimal
- **Recommended scenes**:
  - Focusing on content
  - Complex diagrams/code display
  - Onboarding
  - Documentation-style explanations

- **Characteristics**:
  - White/gray tones
  - Simple
  - Visibility-focused

```json
{
  "background": {
    "type": "minimal",
    "primaryColor": "#ffffff",
    "secondaryColor": "#f0f0f0",
    "opacity": 1
  }
}
```

#### gradient
- **Recommended scenes**:
  - B2C products
  - Approachability appeal
  - Intro/CTA
  - Casual tone

- **Characteristics**:
  - Colorful gradients
  - Soft impression
  - Visually enjoyable

```json
{
  "background": {
    "type": "gradient",
    "primaryColor": "#667eea",
    "secondaryColor": "#764ba2",
    "opacity": 0.95
  }
}
```

#### particles
- **Recommended scenes**:
  - Hook (opening impact)
  - CTA (call to action)
  - Important turning points
  - Energetic impression

- **Characteristics**:
  - Dynamic particles
  - Sense of energy
  - Attention-drawing

```json
{
  "background": {
    "type": "particles",
    "primaryColor": "#000000",
    "secondaryColor": "#00f5ff",
    "opacity": 0.8
  }
}
```

### Recommended Backgrounds by Funnel Stage

| Funnel Stage | Recommended Background | Reason |
|-------------|----------------------|--------|
| Awareness (LP/Ads) | particles, gradient | Visual impact |
| Interest (Intro) | gradient, cyberpunk | Approachability, cutting-edge |
| Consideration (Feature Demo) | minimal, corporate | Focus, trust |
| Conviction (Architecture) | corporate, cyberpunk | Professional |
| Retention (Onboarding) | minimal, gradient | Simple, friendly |

---

## Timing

### Timing Parameters

| Parameter | Use Case | Recommended Value |
|-----------|----------|-------------------|
| **delay_before** | Wait before scene start | 0-15f (0-500ms) |
| **delay_after** | Wait after scene end | 0-30f (0-1000ms) |
| **audio_start_offset** | Audio start offset | 30f (1000ms, standard) |

### Usage Guidelines

#### delay_before (Pre-start Wait)
- **Use cases**:
  - Visual settling after transition
  - Lingering from previous scene
  - Pause to draw attention

- **Recommended values**:
  - `0f`: When transition is sufficient
  - `5-10f`: Light pause
  - `15f`: Firm pause

```json
{
  "timing": {
    "delay_before": 10
  }
}
```

#### delay_after (Post-end Wait)
- **Use cases**:
  - Lingering after audio ends
  - Ensuring CTA display time
  - Ensuring reading time

- **Recommended values**:
  - `0f`: Move to next immediately
  - `15-20f`: Standard lingering
  - `30f`: Ensure thorough reading

```json
{
  "timing": {
    "delay_after": 20
  }
}
```

#### audio_start_offset (Audio Start Offset)
- **Use cases**:
  - Wait after scene display before audio starts
  - Audio starts after visual settling

- **Recommended values**:
  - `30f` (1000ms): Standard (recommended)
  - `15f` (500ms): Fast-paced
  - `45f` (1500ms): Relaxed

```json
{
  "timing": {
    "audio_start_offset": 30
  }
}
```

### Important Audio Sync Rules

> **Important**: Strictly follow these rules for narrated videos

1. **Scene length formula**:
   ```
   duration_ms = audio_start_offset + audio_length + delay_after
   ```

2. **Pre-check audio length**:
   ```bash
   ffprobe -v error -show_entries format=duration \
     -of default=noprint_wrappers=1:nokey=1 audio/scene.wav
   ```

3. **Coordination with transitions**:
   ```
   Scene start = Previous scene start + Previous scene length - Transition length
   Audio start = Scene start + audio_start_offset
   ```

4. **Ensure margins**:
   - Audio must finish before transition starts
   - Ensure at least `delay_after: 20f`

---

## Best Practices

### 1. Direction Combinations by Funnel Stage

#### 90-Second LP/Ad Teaser (Awareness to Interest)
```json
{
  "hook": {
    "transition": { "type": "zoom", "duration_ms": 600 },
    "emphasis": { "level": "high", "sound": "whoosh" },
    "background": { "type": "particles" },
    "timing": { "delay_before": 10, "delay_after": 20 }
  },
  "problem": {
    "transition": { "type": "slideIn", "direction": "right", "duration_ms": 400 },
    "emphasis": { "level": "medium", "sound": "pop" },
    "background": { "type": "gradient" },
    "timing": { "delay_before": 0, "delay_after": 15 }
  },
  "cta": {
    "transition": { "type": "zoom", "duration_ms": 600 },
    "emphasis": { "level": "high", "sound": "chime" },
    "background": { "type": "particles" },
    "timing": { "delay_before": 15, "delay_after": 30 }
  }
}
```

#### 3-Minute Intro Demo (Interest → Consideration)
```json
{
  "intro": {
    "transition": { "type": "fade", "duration_ms": 500 },
    "emphasis": { "level": "high", "sound": "whoosh" },
    "background": { "type": "gradient" },
    "timing": { "delay_before": 0, "delay_after": 20 }
  },
  "demo": {
    "transition": { "type": "cut", "duration_ms": 0 },
    "emphasis": { "level": "medium", "sound": "pop" },
    "background": { "type": "minimal" },
    "timing": { "delay_before": 0, "delay_after": 10 }
  },
  "cta": {
    "transition": { "type": "fade", "duration_ms": 500 },
    "emphasis": { "level": "high", "sound": "chime" },
    "background": { "type": "gradient" },
    "timing": { "delay_before": 10, "delay_after": 30 }
  }
}
```

### 2. Appropriate Sound Effect Usage

**Rules**:
- Maximum **5-7 sound effects** per video
- Reduce sound effects in consecutive scenes (diminishing returns from habituation)
- Always attach sound effects to high emphasis
- Medium emphasis is selective
- Low emphasis is basically silent

### 3. Background Consistency

**Rules**:
- Maximum **2-3 background types** per video
- Unify within sections (same background within a section)
- Only Hook/CTA may use special backgrounds (particles)

### 4. Transition Rhythm

**Rules**:
- Don't use the same transition more than 3 times consecutively
- Combine fast transitions (cut) with varied pacing (fade/zoom)
- Recommend fade or zoom at section start

### 5. Emphasis Level Distribution

**Rules (for 90-second video)**:
- high: 2-3 times (Hook, Differentiator, CTA)
- medium: 5-8 times (key messages)
- low: as needed (supplementary information)

---

## Design Checklist

Verify the following when designing scene direction:

### Transition
- [ ] Selected transition appropriate for scene purpose
- [ ] Avoided consecutive use of same transition
- [ ] duration_ms is appropriate (fade: 500ms, slideIn: 400ms, zoom: 600ms)

### Emphasis
- [ ] Emphasis level is appropriate (high: most important only)
- [ ] Sound effect usage count is appropriate (5-7 total)
- [ ] Keywords to emphasize specified in text array

### Background
- [ ] Background type matches funnel stage
- [ ] Background unified within section
- [ ] primaryColor and secondaryColor specified

### Timing
- [ ] audio_start_offset is 30f (standard)
- [ ] Scene length = audio_start + audio length + delay_after
- [ ] Audio finishes before transition starts

### Overall Balance
- [ ] Sound effects used 5-7 times or less
- [ ] Background types limited to 2-3
- [ ] high emphasis limited to 2-3 times

---

## Related Documents

- [generator.md](./generator.md) - Parallel generation flow
- [visual-effects.md](./visual-effects.md) - Visual effects library
- [schemas/direction.schema.json](../schemas/direction.schema.json) - Direction schema definition
- [schemas/emphasis.schema.json](../schemas/emphasis.schema.json) - Emphasis schema definition
- [schemas/animation.schema.json](../schemas/animation.schema.json) - Animation schema definition
