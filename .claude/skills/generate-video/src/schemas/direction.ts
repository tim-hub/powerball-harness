/**
 * @file direction.ts
 * @description Auto-generated Zod schema for Direction Schema
 * @version 1.0.0
 * @generated This file is auto-generated from schemas/direction.schema.json
 *           All $ref references are resolved during generation.
 *           DO NOT EDIT MANUALLY - run `npm run generate:schemas` instead
 */

import { z } from 'zod';

/**
 * Schema for the direction system. Defines per-scene direction parameters (transitions, emphasis, backgrounds, timing).
 */
export const DirectionSchema = z.object({ "scene_id": z.string().regex(new RegExp("^[a-zA-Z0-9_-]+$")).describe("Target scene ID (must match a scene in the scenario)"), "transition": z.object({ "type": z.enum(["fade","slide_in","zoom","cut"]).describe("Transition type"), "duration_ms": z.number().int().gte(0).lte(2000).describe("Transition duration (milliseconds)").default(500), "easing": z.enum(["linear","easeIn","easeOut","easeInOut"]).describe("Easing function").default("easeInOut"), "direction": z.enum(["left","right","top","bottom"]).describe("Slide direction (only for slide_in)").default("right") }).strict().describe("Transition settings for this scene"), "emphasis": z.object({ "level": z.enum(["high","medium","low"]).describe("Emphasis level"), "text": z.array(z.string()).describe("Text elements to emphasize (keywords/phrases)").default([]), "sound": z.enum(["none","pop","whoosh","chime","ding"]).describe("Sound effect type").default("none"), "color": z.string().regex(new RegExp("^#[0-9A-Fa-f]{6}$")).describe("Emphasis color (HEX format)").default("#00F5FF"), "position": z.enum(["center","top","bottom","left","right"]).describe("Emphasis element position").default("center") }).strict().describe("Emphasis expression within the scene"), "background": z.object({ "type": z.enum(["cyberpunk","corporate","minimal","gradient","particles"]).describe("Background type"), "primary_color": z.string().regex(new RegExp("^#[0-9A-Fa-f]{6}$")).describe("Primary color (HEX format)").optional(), "secondary_color": z.string().regex(new RegExp("^#[0-9A-Fa-f]{6}$")).describe("Secondary color (HEX format)").optional(), "opacity": z.number().gte(0).lte(1).describe("Background opacity").default(1), "blur": z.boolean().describe("Whether to apply blur to the background").default(false) }).strict().describe("Background settings"), "timing": z.object({ "delay_before_ms": z.number().int().gte(0).describe("Wait time before scene start (milliseconds)").default(0), "delay_after_ms": z.number().int().gte(0).describe("Wait time after scene end (milliseconds)").default(0), "audio_start_offset_ms": z.number().int().describe("Audio start offset (milliseconds, positive = delay, negative = advance)").default(1000) }).strict().describe("Timing adjustments within the scene") }).strict().describe("Schema for the direction system. Defines per-scene direction parameters (transitions, emphasis, backgrounds, timing).")

/**
 * Inferred TypeScript type from Zod schema
 */
export type Direction = z.infer<typeof DirectionSchema>;

/**
 * Schema metadata
 */
export const DirectionMeta = {
  version: '1.0.0',
  title: 'Direction Schema',
  description: 'Schema for the direction system. Defines per-scene direction parameters (transitions, emphasis, backgrounds, timing).',
} as const;
