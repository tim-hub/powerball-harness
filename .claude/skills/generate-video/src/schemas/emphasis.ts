/**
 * @file emphasis.ts
 * @description Auto-generated Zod schema for Emphasis Schema
 * @version 1.0.0
 * @generated This file is auto-generated from schemas/emphasis.schema.json
 *           All $ref references are resolved during generation.
 *           DO NOT EDIT MANUALLY - run `npm run generate:schemas` instead
 */

import { z } from 'zod';

/**
 * Schema for emphasis expressions. Defines text emphasis, sound effects, colors, and positioning.
 */
export const EmphasisSchema = z.object({ "level": z.enum(["high","medium","low"]).describe("Emphasis level"), "text": z.array(z.object({ "content": z.string().describe("Text content to emphasize"), "start_ms": z.number().int().gte(0).describe("Emphasis start time (relative position within the scene, in milliseconds)").optional(), "duration_ms": z.number().int().gte(1).describe("Emphasis display duration (milliseconds)").default(1000), "style": z.enum(["bold","glitch","underline","highlight","glow"]).describe("Text style").default("bold") }).strict()).describe("Text elements to emphasize (keywords/phrases)").default([]), "sound": z.object({ "type": z.enum(["none","pop","whoosh","chime","ding"]).describe("Sound effect type").default("none"), "volume": z.number().gte(0).lte(1).describe("Volume (0.0-1.0)").default(0.5), "timing": z.enum(["start","end","peak"]).describe("Sound effect timing").default("start"), "trigger_ms": z.number().int().gte(0).describe("Sound effect trigger time (relative position within the scene, in milliseconds)").default(0) }).strict().describe("Sound effect settings").optional(), "color": z.object({ "primary": z.string().regex(new RegExp("^#[0-9A-Fa-f]{6}$")).describe("Primary color (HEX format)").default("#00F5FF"), "secondary": z.string().regex(new RegExp("^#[0-9A-Fa-f]{6}$")).describe("Secondary color (HEX format, for gradients)").optional(), "glow": z.boolean().describe("Whether to apply glow effect").default(true), "glow_intensity": z.number().gte(0).lte(100).describe("Glow intensity (blur radius, pixels)").default(20) }).strict().describe("Emphasis color settings").optional(), "position": z.object({ "alignment": z.enum(["center","top","bottom","left","right","top_left","top_right","bottom_left","bottom_right"]).describe("Alignment position").default("center"), "offset": z.object({ "x": z.number().describe("X-axis offset").default(0), "y": z.number().describe("Y-axis offset").default(0) }).strict().describe("Position offset (pixels)").optional(), "padding": z.number().gte(0).describe("Padding from screen edges (pixels)").default(40) }).strict().describe("Emphasis element position settings").optional(), "animation": z.object({ "entry": z.enum(["none","fade_in","slide_in","zoom_in","bounce"]).describe("Entry animation").default("fade_in"), "exit": z.enum(["none","fade_out","slide_out","zoom_out"]).describe("Exit animation").default("fade_out"), "duration_ms": z.number().int().gte(1).lte(2000).describe("Animation duration (milliseconds)").default(500), "pulse": z.boolean().describe("Whether to enable pulse effect (blink/scale)").default(false), "pulse_speed": z.number().gte(0.1).lte(10).describe("Pulse speed (cycles per second)").default(1) }).strict().describe("Emphasis display animation settings").optional(), "background": z.object({ "enabled": z.boolean().describe("Whether to show background").default(false), "color": z.string().describe("Background color (HEX format or RGBA)").default("rgba(0, 0, 0, 0.8)"), "border_radius": z.number().gte(0).describe("Border radius (pixels)").default(8), "padding": z.object({ "top": z.number().gte(0).default(16), "right": z.number().gte(0).default(32), "bottom": z.number().gte(0).default(16), "left": z.number().gte(0).default(32) }).strict().describe("Padding within background (pixels)").optional(), "border": z.object({ "enabled": z.boolean().default(false), "width": z.number().gte(0).default(2), "color": z.string().regex(new RegExp("^#[0-9A-Fa-f]{6}$")).describe("Border color (HEX format)").default("#00F5FF") }).strict().describe("Border settings").optional() }).strict().describe("Emphasis element background settings (box display, etc.)").optional(), "typography": z.object({ "font_size": z.number().gte(12).lte(200).describe("Font size (pixels)").default(48), "font_weight": z.union([z.literal(100), z.literal(200), z.literal(300), z.literal(400), z.literal(500), z.literal(600), z.literal(700), z.literal(800), z.literal(900), z.literal("normal"), z.literal("bold")]).describe("Font weight").default(700), "font_family": z.string().describe("Font family").default("sans-serif"), "line_height": z.number().gte(0.5).lte(3).describe("Line height (multiplier)").default(1.5), "letter_spacing": z.number().describe("Letter spacing (pixels)").default(0), "text_transform": z.enum(["none","uppercase","lowercase","capitalize"]).describe("Text transform").default("none") }).strict().describe("Typography settings").optional() }).strict().describe("Schema for emphasis expressions. Defines text emphasis, sound effects, colors, and positioning.")

/**
 * Inferred TypeScript type from Zod schema
 */
export type Emphasis = z.infer<typeof EmphasisSchema>;

/**
 * Schema metadata
 */
export const EmphasisMeta = {
  version: '1.0.0',
  title: 'Emphasis Schema',
  description: 'Schema for emphasis expressions. Defines text emphasis, sound effects, colors, and positioning.',
} as const;
