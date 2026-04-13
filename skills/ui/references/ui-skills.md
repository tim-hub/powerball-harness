---
name: ui-skills-summary
description: "UI Skills constraint set summary (implementation quality first)"
---

# UI Skills Summary

A constraint set to prevent common failure points in UI implementation.

## Stack
- MUST: Use Tailwind CSS default values (exceptions only for existing customizations or explicit requests)
- MUST: Use `motion/react` if JavaScript animations are needed
- SHOULD: Use `tw-animate-css` for Tailwind enter/minor animations
- MUST: Use `cn` (`clsx` + `tailwind-merge`) for class control

## Components
- MUST: Use accessible primitives for keyboard/focus behavior
- MUST: Prefer existing primitives
- NEVER: Mix primitives on the same interaction surface
- SHOULD: Prefer Base UI when compatible
- MUST: Add `aria-label` to icon-only buttons
- NEVER: Hand-implement keyboard/focus behavior (unless explicitly requested)

## Interaction
- MUST: Use AlertDialog for destructive operations
- SHOULD: Use structural skeletons for loading states
- NEVER: Use `h-screen`; use `h-dvh` instead
- MUST: Consider `safe-area-inset` for fixed elements
- MUST: Display errors near the point of interaction
- NEVER: Block paste on input/textarea

## Animation
- NEVER: Add animations unless explicitly requested
- MUST: Only animate `transform` / `opacity`
- NEVER: Animate `width/height/top/left/margin/padding`
- SHOULD: Animate `background/color` only for small, localized UI
- SHOULD: Use `ease-out` for entrances
- NEVER: Exceed 200ms for feedback
- MUST: Pause loops when offscreen
- SHOULD: Respect `prefers-reduced-motion`
- NEVER: Use custom easing unless explicitly requested
- SHOULD: Avoid animations for large images/full-screen elements

## Typography
- MUST: Use `text-balance` for headings
- MUST: Use `text-pretty` for body text
- MUST: Use `tabular-nums` for numbers
- SHOULD: Use `truncate` or `line-clamp` for dense UI
- NEVER: Change `tracking-*` without explicit request

## Layout
- MUST: Use a fixed `z-index` scale (avoid arbitrary `z-*`)
- SHOULD: Use `size-*` for squares

## Performance
- NEVER: Animate large `blur()` / `backdrop-filter`
- NEVER: Apply `will-change` permanently
- NEVER: Put in `useEffect` what can be computed during render

## Design
- NEVER: Use gradients unless explicitly requested
- NEVER: Use purple/multicolor gradients
- NEVER: Use glow for primary affordances
- SHOULD: Use Tailwind's default shadow scale
- MUST: Show one "next step" for empty states
- SHOULD: Limit accent colors to one
- SHOULD: Prefer existing theme/tokens over new colors

## Sources
- https://www.ui-skills.com/
- https://agent-skills.xyz/skills/baptistearno-typebot-io-ui-skills
