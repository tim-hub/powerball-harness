# SaaS Video Best Practices

A collection of best practices for SaaS introduction videos.
Provides guidelines for selecting the optimal composition based on video purpose and funnel stage.

---

## Core Principles

### 1. Decide "Whose Pain to Solve" Before the Screen

Design as a device that creates the shortest path from viewer's pain to confidence in the solution, not a feature introduction.

**Risk**: If the target is vague, it won't resonate with anyone

### 2. Cut Unnecessary Ceremonies in the First Few Seconds

Showing logos or intro effects too long at the beginning increases drop-off rate — get to the point immediately.
Prioritize speed of sound and topic start.

### 3. Length Depends on Purpose

Engagement rate generally drops with longer videos, but optimal length depends on purpose.

| Length | Use Case |
|--------|----------|
| 1-2 min | Low early drop-off |
| 5-10 min | Longer explanations |

**Conclusion**: For long videos, opening design is critical

### 4. Don't Only Place CTA at the End

Place CTAs in the middle too, to account for mid-video drop-off.

### 5. Watchable Video is Hygiene

**Priority**: Audio quality > Screen readability > Tempo > Visuals

**Important**: Bad audio causes immediate drop-off

### 6. Subtitles/Transcripts Are Required

Subtitles are a mandatory requirement — not just auto-generated, corrections are needed too.

### 7. Chapters for Long Videos

Video viewers want to skip around, so add chapters (table of contents).

---

## Video Type Comparison by Funnel Stage

| Purpose (Funnel) | Video Type | Target Length | Composition Core | Primary KPI |
|-------------------|------------|---------------|------------------|-------------|
| Awareness to Interest | Ultra-short Teaser | 30-90 sec | Pain → Result (future) → "More here" | View retention/CTR |
| Interest → Consideration | Intro (Overview Demo) | 2-3 min | Complete 1 use case fastest | Demo signups/Trial starts |
| Consideration → Conviction | Demo (Short Sales) | 2-5 min | Preemptively address objection seeds | Deal rate/Reply rate |
| Conviction → Decision | Walkthrough / Webinar | 5-30 min | Real operations + evidence | CVR/Inquiries |
| Expansion/Efficiency | Hybrid Demo | Recording + Q&A | Standardize with recording → optimize individually with live | Close rate/Labor reduction |
| Retention/Utilization | Onboarding | 30 sec - few min (segmented) | Shortest quick win → Aha | Activation/Retention |
| Support | How-to/Troubleshoot | 1-5 min | One purpose per video | Inquiry reduction/Self-resolution |

---

## Category Guides

### LP/Ads: Short Introduction (30-90 seconds)

**Purpose**: Awareness to Interest funnel

**Composition Outline**:
- 0-5 sec: Pain or desired outcome
- 5-20 sec: Target user and promise
- 20-60 sec: Signature workflow x 1
- 60-90 sec: Next step

**Pitfall to Avoid**: Feature listing, abstract word overload

### Consideration: 2-3 Minute Intro Demo

Complete one use case, avoiding jargon and breaking things down.

**Recommendations**:
- Keep opening short, get to the point
- Choose the single most impactful case
- Place mid-video CTA too

### Sales: 2-5 Minute Demo

Short sales video to move buyers to the next step.

**Strategy**:
- Preemptively address top 3 objections
- 1 video = 1 industry/role

### Decision: 15-30 Minute Walkthrough

Long-form video to drive deeper understanding and decisions. Chapter design is required.

**Key Points**:
- Basic flow
- Objection handling
- Management/security explanation

### Onboarding

Design to get users to their first success (Aha moment).

**Guidelines**:
- Don't teach everything in one video
- Role-specific branching videos

### Support/Help Videos

Don't make video the only answer — combine with text.

**Required Elements**:
- Video + step-by-step text
- Subtitles + full transcript

---

## Templates

### 90-Second Impactful Intro Template

**Use case**: LP/Ads, Awareness to Interest funnel

| Time | Content | Frames (30fps) |
|------|---------|-----------------|
| 0:00-0:05 | Pain or desired outcome | 0-150 |
| 0:05-0:15 | Target user and promise | 150-450 |
| 0:15-0:55 | Signature workflow | 450-1650 |
| 0:55-1:10 | Basis for differentiation | 1650-2100 |
| 1:10-1:30 | CTA | 2100-2700 |

**Total**: 90 seconds = 2700 frames

### 3-Minute Intro Demo Template

**Use case**: Consideration, Interest → Consideration funnel

| Time | Content |
|------|---------|
| 0:00-0:10 | Conclusion + pain |
| 0:10-0:30 | Use case declaration |
| 0:30-2:20 | Complete walkthrough on real screen |
| 2:20-2:50 | Address one common concern |
| 2:50-3:00 | CTA |

### 20-Minute Decision Walkthrough

**Use case**: Conviction → Decision funnel

| Time | Content |
|------|---------|
| 0:00-1:00 | Target audience and challenges |
| 1:00-8:00 | Basic flow |
| 8:00-12:00 | Top 2 objections |
| 12:00-15:00 | Management/security |
| 15:00-20:00 | Success stories + CTA |

---

## Production Checklist

### Pre-Recording

- [ ] Script: Cut down what to say to the minimum
- [ ] Demo environment: Notifications off, personal info removed
- [ ] Screen: Zoom so UI text is readable

### During Recording

- [ ] Audio: Clarity is top priority
- [ ] Lighting/appearance: Minimum acceptable
- [ ] Tempo: Cut all moments of hesitation

### Publishing

- [ ] Subtitles: Always quality-check
- [ ] Transcript: For search and skimming
- [ ] Chapters required for long videos

---

## Common Failure Patterns

| Failure | Impact |
|---------|--------|
| Unclear target audience | Resonates with no one |
| Long logo/company intro | Early drop-off |
| Feature overload | Loses focus |
| No/sloppy subtitles | Accessibility decline |
| Just placing the video and done | Can't measure effectiveness |
| CTA only at the end | Missed by mid-video drop-offs |

---

## Recommended 3-Video Set

The basic set for fastest results:

1. **90-second teaser** - Awareness acquisition
2. **3-minute intro demo** - Consideration promotion
3. **15-25 minute decision walkthrough** - Close support

**Result**: Fills the information gap from acquisition → consideration → decision

---

## Usage in Harness

### Automatic Video Type Detection

| Harness Detection Condition | Recommended Video Type | Template |
|-----------------------------|----------------------|----------|
| New project | LP/Ad short-form | 90-second teaser |
| UI change detected | Intro demo | 3-minute template |
| CHANGELOG updated | Release notes | Before/After focused |
| Large structural change | Architecture explanation | Walkthrough |

### Scene Composition Guide

#### Short-form (30-90 seconds)

```
HookScene (3s) → ProblemPromise (7s) → WorkflowDemo (40-60s) → Differentiator (10s) → CTA (10s)
```

#### Intro Demo (2-3 minutes)

```
Hook (10s) → UseCase Declaration (20s) → Real Screen Demo (110s) → Concern Resolution (30s) → CTA (10s)
```

#### Walkthrough (15-30 minutes)

```
Target & Challenges (1min) → Basic Flow (7min) → Objection Handling (4min) → Management/Security (3min) → Success Stories + CTA (5min)
```

---

## Remotion Implementation Rules

### Animation

| Rule | Reason |
|------|--------|
| `useCurrentFrame()` required | CSS animations prohibited, use Remotion's frame control |
| `spring({ damping: 200 })` | Smooth motion |
| `interpolate()` + `extrapolateRight: 'clamp'` | Prevent value runaway |
| Typewriter effect uses `text.slice(0, charCount)` | Per-character opacity changes not recommended |

### Audio

| Rule | Reason |
|------|--------|
| `Audio` imported from `@remotion/media` | `Html5Audio` is deprecated |
| Audio start = scene start + 30f (1-second wait) | Audio starts after visual settling on slide change |
| Scene length = 30f + audio length + 20f margin | Audio finishes before transition |
| Pre-check audio length with `ffprobe` | Prerequisite for scene design |

### TransitionSeries

| Rule | Reason |
|------|--------|
| Transition length recommended at 15f (0.5 seconds) | Natural switching |
| Scene start = previous scene start + previous scene length - transition length | Account for overlap |
| Audio must finish before transition starts | Audio during transition feels unnatural |

### Scene Length Calculation Example

```
Audio length check (at 30fps):
  hook: 4.0 sec = 121 frames
  problem: 12.0 sec = 360 frames

Scene length calculation:
  hook: 30 (wait) + 121 (audio) + 24 (margin) = 175 frames
  problem: 30 (wait) + 360 (audio) + 25 (margin) = 415 frames

Scene start frame (transition 15f):
  hook: 0
  problem: 175 - 15 = 160

Audio start timing:
  hook: 0 + 30 = 30
  problem: 160 + 30 = 190
```

---

## References

- [planner.md](planner.md) - Scenario planning
- [generator.md](generator.md) - Parallel scene generation
- [analyzer.md](analyzer.md) - Codebase analysis
