---
name: docs-notebooklm-slide-yaml
description: "Generates two design instruction YAML variations (from different angles) for use in NotebookLM's [Customize Slide Deck] feature."
allowed-tools: ["Read", "Write"]
---

# NotebookLM Slide Design YAML (2 Variations)

## Purpose

For NotebookLM's "Slide Deck" generation, **explicitly specify design direction in YAML to improve quality and reproducibility**.

This skill outputs **two YAML variations with different interpretive approaches** so the same source can produce different visual results.

References (usage/examples):
- `https://note.com/yoshifujidesign/n/nd9c8db0b55b8`
- `https://note.com/yoshifujidesign/n/n7412bccb5762`

---

## Usage (NotebookLM side)

After loading source materials in NotebookLM, go to
[Slide Deck] -> [Customize Slide Deck] -> [Describe the slides you want to create]
and **paste the entire YAML output from this skill**, then click [Generate].

---

## Input (Confirm with User)

If possible, confirm the following (up to 5 questions) first. If no answer is given, proceed with the defaults in parentheses.

1. Purpose (e.g., internal sharing / sales / recruiting / investor presentation) (internal sharing)
2. Target audience (e.g., executives / developers / non-engineers) (mixed, including non-engineers)
3. Tone (e.g., trustworthy / innovative / friendly) (trustworthy)
4. Brand elements (logo / brand colors / font specifications) (none specified)
5. Photo/chart ratio (photo-heavy / data-heavy) (data-heavy)

---

## Output (Required Format)

Always output the following **2 variations**.

- Variation A: **Prioritizes readability, trustworthiness, and data communication** (minimal/corporate)
- Variation B: **Emphasizes emotion, storytelling, and the aesthetics of whitespace** (editorial/lifestyle)

Both must be in **YAML**, ready to paste directly into NotebookLM.

Output format:

- A single YAML block (code block) directly under `### Variation A`
- A single YAML block (code block) directly under `### Variation B`

---

## Generation Rules

- Include specific colors in **HEX** format (e.g., `#FFFFFF`)
- Include some "don'ts" (e.g., no excessive borders, no excessive shadows, limit number of colors)
- Explicitly document constraints that **improve reproducibility**, such as chapter structure/navigation/whitespace/grids
- Write as a **structural template** that works even when the content (source) is unknown

---

## Output Template (Example Skeleton)

### Variation A (Minimal/Corporate)

(Adjust as needed based on purpose/audience/brand colors)

```yaml
# presentation_design_spec_minimal_corporate_jp.yaml
# Style: Modern Minimal / Corporate
# Purpose: Prioritize readability and data communication (trustworthiness)

Overall Design Settings:
  Tone: "Trustworthy, professional, clean, logical"
  Color Palette:
    Base: "#FFFFFF (white) or #F5F5F7 (very light gray)"
    Text Color: "#111111 (near black) or #333333 (charcoal)"
    Accent: "Use brand color if available. Otherwise adopt #0052CC (blue)"
    Emphasis: "#FFB020 (amber) - limited to important numbers and warnings"
    Color Limit: "Maximum 3 colors + gray scale"
  Typography:
    Headings: "Bold gothic (sans-serif). Short, impactful words."
    Body: "Highly readable gothic. Wide line spacing."
    Numbers: "Large sans-serif bold. Units (%, etc.) smaller."
  Common Layout Rules:
    Grid: "12-column equivalent alignment. Wide left/right margins."
    Navigation: "Small '01. SECTION' number + chapter title in upper left."
    Whitespace: "No overcrowding. One message per slide."
    Charts: "Minimize rules/grid lines. Accent color only for emphasis points."
  Prohibited:
    - "No excessive drop shadows"
    - "No overuse of decorative borders"
    - "No excessive gradients"

Layout Variations (Catalog):
  - Type: "Cover"
    Design: "Large whitespace. Title left-aligned and large, subtitle small. One thin horizontal rule."
  - Type: "Chapter Divider"
    Design: "Large chapter title. Minimal solid-color background."
  - Type: "Conclusion (TL;DR)"
    Design: "3 conclusion points + one large number/metric."
  - Type: "2-Column (Problem vs Solution)"
    Design: "Split in two with a vertical line. Left: problem, Right: solution. Headings bold, body thin."
  - Type: "Data/Chart"
    Design: "Simple bar/line representation. Short annotations."
  - Type: "Process"
    Design: "Horizontal steps (01-05). Minimal circle icons."
  - Type: "Timeline"
    Design: "Vertical line + large year + small description."
  - Type: "Summary"
    Design: "3 next actions. Concise checklist style."
```

### Variation B (Editorial/Story)

(Adjust as needed based on purpose/audience/photo ratio)

```yaml
# presentation_design_spec_editorial_story_jp.yaml
# Style: Modern Editorial / Lifestyle Magazine
# Purpose: Tell a story through whitespace and photography (emotion/persuasion)

Overall Design Settings:
  Tone: "Calm, intellectual, emotional, organic, refined"
  Visual Identity:
    Background Color: "#F3F0EB (sand beige) or #EBEBEB (warm gray)"
    Text Color: "#333333 (charcoal) - avoid pure black"
    Accent Color: "#E07A5F (terracotta) or #708D81 (olive)"
    Image Style:
      Characteristics: "Natural light, film grain, low saturation"
      Shape: "Rounded rectangles or arch shapes"
  Typography:
    Headings: "Elegant serif/Mincho typeface. Subtly emphasize key words."
    Body: "Readable gothic/sans-serif, slightly smaller."
    Numbers: "Serif numerals for a calm appearance."
  Common Layout Rules:
    Whitespace: "Whitespace is top priority. Reduce text density."
    Photos: "One large photo per slide as the default."
    Rules: "Borders are prohibited. Separation by whitespace or ultra-thin lines only."
  Design Rules:
    - "Cool tones and neon are prohibited (unified warm tones)"
    - "Excessive icons are prohibited (convey through photos and typography)"

Layout Variations (Catalog):
  - Type: "Cover (Magazine Cover)"
    Design: "Photo full bleed on right, title on left. Separated by thin horizontal rule."
  - Type: "Quote"
    Design: "Short message centered. Large faint quotation marks in background."
  - Type: "Story (1->2->3)"
    Design: "Vertical steps. Connected by dotted lines. Small circular photo beside each step."
  - Type: "Comparison"
    Design: "Two photos side by side + captions for each. Muted colors."
  - Type: "Data"
    Design: "Ultra-thin line chart + descriptive paragraph for context (adding narrative)."
  - Type: "Summary"
    Design: "3 next actions. Close with ample whitespace."
```
