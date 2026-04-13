# Content Layout

This repository uses the following rules to separate `docs/` and `out/`.

## Basic Rules

- `docs/`: Authoritative content for human readers. Manually edited text, publication drafts, prompts, design notes, and reference materials go here
- `out/`: Output destination for tools and generation processes. Images, candidate drafts, exports, and derivative comparison materials go here

When in doubt, use these criteria:

1. Is this text that someone will read and reuse later?
   - Yes: `docs/`
2. Is this a regeneratable artifact?
   - Yes: `out/`

## Social / X Operations

- `docs/social/`
  - Authoritative post text
  - Image generation prompts
  - Alt text
  - Post notes, structure drafts
- `out/social/`
  - Generated images
  - Candidate cards
  - Tool-generated drafts
  - Temporary comparison artifacts

In other words, `docs/social/` is "what to post" and `out/social/` is "what was generated."

Legacy directories `out/x-post/`, `out/x-posts/`, `out/x-promo/`, and `out/x-release/` also exist from past operations. These will remain for now, but **all new social outputs should go to `out/social/`**.

## Slides / Media Operations

- `docs/slides/`
  - Slide manuscripts, specs, YAML sources
- `out/slides/`
  - Exported images, selected images, quality reports

## Rules for Additions

- When saving new post text, first save it to `docs/social/`
- Save results from image generation or export runs to `out/social/`
- Do not add generated images to `docs/`
- Do not add authoritative descriptions to `out/`

## Current Organization Policy

- X post source manuscripts are being consolidated into `docs/social/`
- Existing `out/social/` content is maintained as generated artifacts
- Future additions follow the same rules
