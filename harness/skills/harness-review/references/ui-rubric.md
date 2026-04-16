# UI Rubric Reference

4-axis design quality scoring for `harness-review --ui-rubric`.
Each axis is scored from 0 to 10. Scores are informational and do not affect the
APPROVE/REQUEST_CHANGES verdict unless Functionality drops to 3 or below.

---

## Axis 1: Design Quality (0–10)

**Definition**: How well the UI communicates information through visual structure — hierarchy,
spacing, color, and typography consistency.

### Scoring Guide

| Score | Meaning |
|-------|---------|
| 9–10  | Exceptional hierarchy; every element has clear weight; spacing system is consistent throughout |
| 7–8   | Strong fundamentals; minor inconsistencies in spacing or type scale that don't impede comprehension |
| 5–6   | Acceptable but flat; hierarchy relies on text alone; spacing feels ad-hoc |
| 3–4   | Competing visual weights; hard to know where to look first; color used inconsistently |
| 1–2   | No discernible hierarchy; elements fight for attention |
| 0     | Broken layout; elements overlap or are invisible |

### What to Look For

- Visual hierarchy: Is there a clear primary → secondary → tertiary reading order?
- Spacing: Are margins and padding drawn from a consistent scale (e.g., 4px grid)?
- Color consistency: Do interactive elements share a color treatment? Are text contrast ratios accessible?
- Typography: Are heading levels, weights, and sizes used semantically?

### Example Observations

- "Strong visual hierarchy — primary action is clearly dominant at 9/10 scale relative to secondary actions."
- "Spacing inconsistency: some cards use 16px padding, others 20px — no obvious grid system."
- "Low contrast on disabled state text (#aaa on white = 2.3:1, fails WCAG AA)."

---

## Axis 2: Originality (0–10)

**Definition**: The degree to which the design makes deliberate, distinctive creative choices rather
than defaulting to template or boilerplate patterns.

### Scoring Guide

| Score | Meaning |
|-------|---------|
| 9–10  | Distinctive voice; custom illustrations, micro-interactions, or layout patterns that feel intentional and unique |
| 7–8   | Clear design language with some custom touches; not generic even if using a component library |
| 5–6   | Competent but recognizable as a framework default; relies heavily on library presets |
| 3–4   | Looks like a direct clone of a starter template with minimal customization |
| 1–2   | Bare library defaults; no visible design decision beyond "add the component" |
| 0     | Placeholder or wireframe-level; no design intent visible |

### What to Look For

- Does the color palette feel chosen for this product, or is it a library default?
- Are there any custom illustrations, icons, or animations that signal design investment?
- Layout: Is the grid/layout system adapted to the content, or is it a generic 12-column grid applied mechanically?
- Micro-interactions: Are hover states, transitions, and focus styles considered?

### Example Observations

- "Custom empty-state illustration signals care beyond a generic 'No data' text message."
- "Button styles are Tailwind defaults (blue-500, rounded) — no brand adaptation visible."
- "Modal uses a standard shadcn/ui Dialog with no customization — score reflects competent use of tooling."

---

## Axis 3: Craft (0–10)

**Definition**: Implementation quality — attention to pixel-level detail, code cleanliness of
the UI layer, and polish of interactive states.

### Scoring Guide

| Score | Meaning |
|-------|---------|
| 9–10  | Pixel-perfect; all interactive states (hover, focus, active, disabled, loading, error) are handled; transitions are smooth |
| 7–8   | Most states handled; minor rough edges (e.g., missing loading state on one button) |
| 5–6   | Happy path polished; edge-case states (empty, error, loading) feel like afterthoughts |
| 3–4   | Visible layout bugs at non-standard viewport sizes; focus styles missing; inconsistent icon sizing |
| 1–2   | Multiple broken interactive states; layout collapses at mobile breakpoints |
| 0     | Non-functional or visually broken in the provided diff |

### What to Look For

- Are all button/input states (hover, focus, active, disabled, loading) implemented?
- Does the layout hold at 320px, 768px, and 1440px widths?
- Are icon sizes consistent throughout?
- Are transitions/animations using CSS variables or design tokens, or are they hardcoded magic numbers?
- Accessibility: are focus rings visible? Are interactive elements reachable via keyboard?

### Example Observations

- "All 6 interactive states implemented with smooth 150ms transitions — craft is exceptional."
- "Loading spinner present on form submit but missing on table row delete action."
- "Text truncation breaks at 320px — line overflows container by ~8px."

---

## Axis 4: Functionality (0–10)

**Definition**: How completely the UI implements its intended features — covering happy paths,
error states, empty states, and accessibility requirements.

### Scoring Guide

| Score | Meaning |
|-------|---------|
| 9–10  | All specified user flows complete; error and empty states handled; accessible to screen reader and keyboard users |
| 7–8   | Primary flows complete; 1–2 non-critical edge cases (e.g., pagination at 0 items) missing |
| 5–6   | Core feature works; error states missing or generic; no keyboard navigation |
| 3–4   | Feature is partially complete; core flow broken in at least one scenario — **verdict impact threshold** |
| 1–2   | Feature is barely functional; multiple broken flows |
| 0     | Non-functional — the described feature does not work |

### Verdict Impact Rule

A Functionality score of **3 or below** escalates the verdict to REQUEST_CHANGES, even if
no other critical/major issues exist. A partially-complete feature that ships breaks the user contract.

### What to Look For

- Does the primary user flow complete end-to-end without errors?
- Is there an empty state when the list/table has no data?
- Are error messages specific and actionable, or generic ("Something went wrong")?
- Can the feature be navigated entirely by keyboard?
- Are form fields labeled for screen readers (aria-label or associated `<label>`)?

### Example Observations

- "All CRUD operations functional; empty state ('No items yet') present; keyboard navigable — full marks."
- "Form submits successfully but shows no confirmation — user has no feedback that the action worked."
- "Functionality score 3: the modal close button is non-functional; users cannot dismiss without pressing Escape."

---

## Aggregate Scoring

After scoring all four axes, summarize results in the review output:

```json
{
  "ui_rubric": {
    "design_quality": { "score": 8, "observations": ["..."] },
    "originality":    { "score": 6, "observations": ["..."] },
    "craft":          { "score": 9, "observations": ["..."] },
    "functionality":  { "score": 7, "observations": ["..."] }
  }
}
```

**Aggregate interpretation:**

| Average Score | Interpretation |
|---------------|----------------|
| 8.0–10.0      | Excellent — ready to ship with pride |
| 6.0–7.9       | Good — minor improvements recommended |
| 4.0–5.9       | Acceptable — notable gaps, follow-up tasks advised |
| Below 4.0     | Needs work — significant design or functional issues |

The aggregate is informational. Only the Functionality axis (score ≤ 3) has a hard verdict impact.
