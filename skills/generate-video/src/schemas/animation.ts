/**
 * @file animation.ts
 * @description Auto-generated Zod schema for Animation Schema
 * @version 1.0.0
 * @generated This file is auto-generated from schemas/animation.schema.json
 *           All $ref references are resolved during generation.
 *           DO NOT EDIT MANUALLY - run `npm run generate:schemas` instead
 */

import { z } from 'zod';

/**
 * Schema for animation settings. Defines transitions, easing, and spring physics parameters.
 */
export const AnimationSchema = z.object({ "type": z.enum(["fade","slide_in","zoom","cut","spring","rotate","scale"]).describe("Animation type"), "duration_ms": z.number().int().gte(0).lte(10000).describe("Animation duration (milliseconds)"), "easing": z.enum(["linear","easeIn","easeOut","easeInOut","easeInQuad","easeOutQuad","easeInOutQuad","easeInCubic","easeOutCubic","easeInOutCubic"]).describe("Easing function (ignored for spring type)").default("easeInOut"), "spring": z.object({ "damping": z.number().gte(1).lte(500).describe("Damping (decay coefficient). Higher values stop faster").default(200), "stiffness": z.number().gte(1).lte(500).describe("Stiffness (spring rigidity). Higher values produce stronger bounce").default(100), "mass": z.number().gte(0.1).lte(10).describe("Mass. Higher values produce heavier motion").default(1), "overshoot_clamping": z.boolean().describe("Whether to suppress overshoot (overshooting)").default(false) }).strict().describe("Spring physics parameters (only applied when type is spring)").optional(), "delay_ms": z.number().int().gte(0).describe("Delay before animation start (milliseconds)").default(0), "from": z.object({ "opacity": z.number().gte(0).lte(1).optional(), "x": z.number().describe("X-coordinate offset (pixels)").optional(), "y": z.number().describe("Y-coordinate offset (pixels)").optional(), "scale": z.number().gte(0).describe("Scale value").optional(), "rotate": z.number().describe("Rotation angle (degrees)").optional() }).strict().describe("Values at animation start").optional(), "to": z.object({ "opacity": z.number().gte(0).lte(1).optional(), "x": z.number().describe("X-coordinate offset (pixels)").optional(), "y": z.number().describe("Y-coordinate offset (pixels)").optional(), "scale": z.number().gte(0).describe("Scale value").optional(), "rotate": z.number().describe("Rotation angle (degrees)").optional() }).strict().describe("Values at animation end").optional(), "interpolate": z.object({ "input_range": z.array(z.number()).min(2).describe("Input range (array of frame numbers)").optional(), "output_range": z.array(z.number()).min(2).describe("Output range (array of values)").optional(), "extrapolate_left": z.enum(["clamp","extend","identity"]).describe("Left extrapolation method (out of range)").default("clamp"), "extrapolate_right": z.enum(["clamp","extend","identity"]).describe("Right extrapolation method (out of range)").default("clamp") }).strict().describe("Detailed settings for the Remotion interpolate function").optional(), "loop": z.object({ "enabled": z.boolean().describe("Whether to enable looping").default(false), "count": z.number().int().gte(0).describe("Loop count (0 for infinite loop)").default(0), "reverse": z.boolean().describe("Whether to use ping-pong loop (forward then reverse)").default(false) }).strict().describe("Loop settings (optional)").optional() }).strict().describe("Schema for animation settings. Defines transitions, easing, and spring physics parameters.")

/**
 * Inferred TypeScript type from Zod schema
 */
export type Animation = z.infer<typeof AnimationSchema>;

/**
 * Schema metadata
 */
export const AnimationMeta = {
  version: '1.0.0',
  title: 'Animation Schema',
  description: 'Schema for animation settings. Defines transitions, easing, and spring physics parameters.',
} as const;
