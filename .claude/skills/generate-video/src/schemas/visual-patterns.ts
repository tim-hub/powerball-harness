/**
 * @file visual-patterns.ts
 * @description Auto-generated Zod schema for Visual Patterns Schema
 * @version 1.0.0
 * @generated This file is auto-generated from schemas/visual-patterns.schema.json
 *           All $ref references are resolved during generation.
 *           DO NOT EDIT MANUALLY - run `npm run generate:schemas` instead
 */

import { z } from 'zod';

/**
 * Schema for image generation patterns. Defines 4 pattern types: comparison, concept, flow, and highlight.
 */
export const VisualPatternsSchema = z.object({ "type": z.enum(["comparison","concept","flow","highlight"]).describe("Image pattern type"), "topic": z.string().min(1).max(200).describe("Image subject/theme"), "style": z.enum(["minimalist","technical","modern","gradient","flat","3d"]).describe("Visual style (optional)").default("modern"), "color_scheme": z.object({ "primary": z.string().regex(new RegExp("^#[0-9A-Fa-f]{6}$")).describe("Primary color (e.g., #3B82F6)").optional(), "secondary": z.string().regex(new RegExp("^#[0-9A-Fa-f]{6}$")).describe("Secondary color (e.g., #10B981)").optional(), "accent": z.string().regex(new RegExp("^#[0-9A-Fa-f]{6}$")).describe("Accent color (e.g., #F59E0B)").optional(), "background": z.string().regex(new RegExp("^#[0-9A-Fa-f]{6}$")).describe("Background color (e.g., #1F2937)").default("#1F2937") }).strict().describe("Color scheme settings").optional(), "comparison": z.object({ "left_side": z.object({ "label": z.string().min(1).max(50).describe("Label (e.g., Before, Bad example)"), "items": z.array(z.string().min(1).max(100)).min(1).max(5).describe("List of display items"), "icon": z.enum(["x","warning","sad","slow","confused","broken"]).describe("Icon type (e.g., x, warning, sad)").optional(), "sentiment": z.enum(["negative","neutral","caution"]).describe("Sentiment/impression").default("negative") }).strict().describe("Left side (Before / Bad example)"), "right_side": z.object({ "label": z.string().min(1).max(50).describe("Label (e.g., After, Good example)"), "items": z.array(z.string().min(1).max(100)).min(1).max(5).describe("List of display items"), "icon": z.enum(["check","star","happy","fast","clear","solid"]).describe("Icon type (e.g., check, star, happy)").optional(), "sentiment": z.enum(["positive","neutral","success"]).describe("Sentiment/impression").default("positive") }).strict().describe("Right side (After / Good example)"), "divider": z.enum(["arrow","vs","line","gradient"]).describe("Divider type").default("arrow") }).strict().describe("Settings for the comparison pattern").optional(), "concept": z.object({ "elements": z.array(z.object({ "id": z.string().regex(new RegExp("^[a-z0-9_-]+$")).describe("Unique ID of the element"), "label": z.string().min(1).max(50).describe("Element label"), "description": z.string().max(200).describe("Element description (optional)").optional(), "level": z.number().int().gte(0).lte(5).describe("Hierarchy level (0 = top level)").default(0), "parent_id": z.string().regex(new RegExp("^[a-z0-9_-]+$")).describe("Parent element ID (for expressing hierarchy)").optional(), "icon": z.enum(["box","circle","diamond","hexagon","cloud","gear","database","server","user","code"]).describe("Icon type").optional(), "emphasis": z.enum(["high","medium","low"]).describe("Emphasis level").default("medium") }).strict()).min(2).max(10).describe("List of concept elements"), "relationships": z.array(z.object({ "from": z.string().regex(new RegExp("^[a-z0-9_-]+$")).describe("Source element ID of the relationship"), "to": z.string().regex(new RegExp("^[a-z0-9_-]+$")).describe("Target element ID of the relationship"), "label": z.string().max(30).describe("Relationship label (e.g., contains, generates, depends on)").optional(), "type": z.enum(["hierarchy","flow","dependency","association"]).describe("Relationship type").default("association"), "bidirectional": z.boolean().describe("Whether the relationship is bidirectional").default(false) }).strict()).describe("Relationships between elements").optional(), "layout": z.enum(["hierarchy","radial","grid","flow","circular"]).describe("Layout type").default("hierarchy") }).strict().describe("Settings for the concept diagram pattern").optional(), "flow": z.object({ "steps": z.array(z.object({ "id": z.string().regex(new RegExp("^[a-z0-9_-]+$")).describe("Unique ID of the step"), "label": z.string().min(1).max(50).describe("Step label"), "description": z.string().max(150).describe("Detailed description of the step").optional(), "order": z.number().int().gte(1).lte(20).describe("Step order (1-based)").optional(), "type": z.enum(["start","process","decision","end","parallel","subprocess"]).describe("Step type").default("process"), "icon": z.enum(["circle","square","diamond","rounded","hexagon"]).describe("Icon type").optional(), "duration": z.string().max(20).describe("Estimated duration (e.g., 2 min, instant)").optional() }).strict()).min(2).max(10).describe("Flow steps"), "direction": z.enum(["horizontal","vertical","zigzag"]).describe("Flow direction").default("horizontal"), "arrow_style": z.enum(["solid","dashed","dotted","thick","animated"]).describe("Arrow style").default("solid"), "show_numbers": z.boolean().describe("Whether to show step numbers").default(true) }).strict().describe("Settings for the flow diagram pattern").optional(), "highlight": z.object({ "main_text": z.string().min(1).max(100).describe("Main text (content to emphasize)"), "sub_text": z.string().max(150).describe("Sub text (supplementary description)").optional(), "icon": z.enum(["star","check","alert","info","trophy","rocket","fire","bolt","heart","none"]).describe("Icon type").default("none"), "position": z.enum(["center","top","bottom","left","right"]).describe("Text position").default("center"), "effect": z.enum(["glow","shadow","gradient","outline","none"]).describe("Visual effect").default("glow"), "font_size": z.enum(["small","medium","large","xlarge"]).describe("Font size").default("large"), "emphasis": z.enum(["high","medium","low"]).describe("Emphasis level").default("high") }).strict().describe("Settings for the highlight pattern").optional(), "dimensions": z.object({ "width": z.number().int().gte(256).lte(2048).describe("Width (pixels)").default(1920), "height": z.number().int().gte(256).lte(2048).describe("Height (pixels)").default(1080), "aspect_ratio": z.enum(["16:9","4:3","1:1","9:16"]).describe("Aspect ratio").default("16:9") }).strict().describe("Image size settings").optional(), "generation": z.object({ "seed": z.number().int().gte(0).describe("Deterministic seed value (for reproducibility)").optional(), "quality": z.enum(["draft","standard","high"]).describe("Image quality").default("standard"), "retries": z.number().int().gte(0).lte(5).describe("Maximum retry count on quality failure").default(3) }).strict().describe("Generation settings").optional(), "metadata": z.object({ "scene_id": z.string().describe("Associated scene ID").optional(), "purpose": z.string().max(50).describe("Image purpose (e.g., intro, demo, cta)").optional(), "tags": z.array(z.string().max(30)).max(10).describe("Tags").optional() }).catchall(z.unknown()).describe("Metadata (additional information)").optional() }).strict().and(z.unknown().superRefine((x, ctx) => {
    const schemas = [z.object({ "type": z.literal("comparison").optional() }), z.object({ "type": z.literal("concept").optional() }), z.object({ "type": z.literal("flow").optional() }), z.object({ "type": z.literal("highlight").optional() })];
    const errors = schemas.reduce<z.ZodError[]>(
      (errors, schema) =>
        ((result) =>
          result.error ? [...errors, result.error] : errors)(
          schema.safeParse(x),
        ),
      [],
    );
    if (schemas.length - errors.length !== 1) {
      ctx.addIssue({
        path: ctx.path,
        code: "invalid_union",
        unionErrors: errors,
        message: "Invalid input: Should pass single schema",
      });
    }
  })).describe("Schema for image generation patterns. Defines 4 pattern types: comparison, concept, flow, and highlight.")

/**
 * Inferred TypeScript type from Zod schema
 */
export type VisualPatterns = z.infer<typeof VisualPatternsSchema>;

/**
 * Schema metadata
 */
export const VisualPatternsMeta = {
  version: '1.0.0',
  title: 'Visual Patterns Schema',
  description: 'Schema for image generation patterns. Defines 4 pattern types: comparison, concept, flow, and highlight.',
} as const;
