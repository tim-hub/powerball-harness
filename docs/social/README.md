# Social Content

This directory stores authoritative content for external communications on X and other platforms.

## What Goes Here

- Post text
- Thread manuscripts
- Image generation prompts
- Alt text
- Post notes

Supplementary notes:
- In-progress `x-article` packages go to `out/social/<slug>/`
- These contain `article.md`, images, quality reports, and API responses
- `docs/social/` is used for authoritative content that has been promoted for publication

## What Does NOT Go Here

- Generated images
- Candidate renderings
- Derivative comparison images
- Temporary output

Those go to `out/social/`.

Legacy directories `out/x-post/` and `out/x-promo/` remain for compatibility, but new social artifacts should go to `out/social/`.

## Naming Conventions

- Update introductions: `claude-code-<version>-harness-update.md`
- General announcements: `harness-<topic>-x-post.md`
- Topic names are preferred over sequential numbering

## Operational Rules

- When editing post text before publishing, update the manuscript in this directory
- Only add post-generation artifacts to `out/social/`
- When in doubt, refer to `docs/content-layout.md` as the authoritative layout rules
