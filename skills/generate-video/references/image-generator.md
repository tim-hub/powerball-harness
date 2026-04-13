# Image Generator - Nano Banana Pro Automatic Image Generation

Automatically generates high-quality images for video scenes using Nano Banana Pro (Google DeepMind).

---

## Overview

Automatically executed when asset images are determined to be needed during the scene generation phase of `/generate-video`.
Implements a quality assurance loop: generate 2 images → Claude assesses quality → regenerate if NG.

## Prerequisites

- `GOOGLE_AI_API_KEY` environment variable is set
- Nano Banana Pro (Gemini 3 Pro Image Preview) is enabled in Google AI Studio

---

## API Specification

> **Official Documentation**: [Nano Banana image generation | Gemini API](https://ai.google.dev/gemini-api/docs/image-generation)

### Endpoint

```
POST https://generativelanguage.googleapis.com/v1beta/models/gemini-3-pro-image-preview:generateContent
```

### Model Selection

| Model | Use Case | Max Resolution |
|-------|----------|----------------|
| `gemini-3-pro-image-preview` | Pro quality (recommended) | 4K |
| `gemini-2.5-flash-image` | Fast, low cost | 1024px |

### Authentication

```bash
# x-goog-api-key header (Gemini API standard method)
x-goog-api-key: ${GOOGLE_AI_API_KEY}
```

> **Note**: Gemini API uses the `x-goog-api-key` header. Query parameter method (`?key=...`) is also available, but header method is recommended.

### Request Format

```json
{
  "contents": [{
    "parts": [
      {"text": "A modern SaaS dashboard interface with clean design, showing analytics charts and user metrics, professional UI mockup, light theme"}
    ]
  }],
  "generationConfig": {
    "responseModalities": ["TEXT", "IMAGE"],
    "imageConfig": {
      "aspectRatio": "16:9",
      "imageSize": "2K"
    }
  }
}
```

> **Note**: You can specify `["TEXT", "IMAGE"]` or `["IMAGE"]` for `responseModalities`. This flow specifies both to also obtain text descriptions for quality assessment.

### Response Format

```json
{
  "candidates": [{
    "content": {
      "parts": [
        {"text": "Here is the generated image of a modern SaaS dashboard..."},
        {
          "inline_data": {
            "mime_type": "image/png",
            "data": "iVBORw0KGgoAAAANS..."
          }
        }
      ]
    }
  }]
}
```

> **Note**: REST API uses snake_case (`inline_data`, `mime_type`). SDK uses camelCase (`inlineData`, `mimeType`).

---

## Resolution Options

| Setting | Resolution | Use Case | Cost Estimate |
|---------|-----------|----------|---------------|
| `1K` | 1024x1024 | Preview, testing | ~$0.02/image |
| `2K` | 2048x2048 | Standard quality | ~$0.06/image |
| `4K` | 4096x4096 | High quality, professional | ~$0.12/image |

### Aspect Ratio

| Ratio | Use Case |
|-------|----------|
| `16:9` | Video scenes (recommended) |
| `1:1` | Icons, logos |
| `9:16` | Vertical video |
| `4:3` | Presentation materials |

---

## Prompt Design Guidelines

### Basic Structure

```
[Subject] + [Style] + [Quality specification] + [Constraints]
```

### Prompt Templates by Scene Type

#### Intro/Title Scene

```
Professional product logo and title card for "{product_name}",
modern minimalist design, clean typography,
{brand_color} accent color, dark background,
cinematic quality, 4K render
```

#### UI Demo Scene (Supplementary Image)

```
Modern web application interface showing {feature_description},
clean UI design, light theme, subtle shadows,
professional SaaS aesthetic, mockup style,
no text labels, focus on visual hierarchy
```

#### CTA Scene

```
Call-to-action banner for {product_name},
action-oriented design, prominent button,
{brand_color} gradient, professional marketing style,
clear visual hierarchy, engaging composition
```

#### Architecture/Concept Diagram

```
Technical architecture diagram showing {concept},
isometric illustration style, modern tech aesthetic,
clear visual flow, connected components,
professional documentation quality, clean lines
```

### Prompt Quality Improvement Tips

| Addition | Effect |
|----------|--------|
| `professional quality` | Overall quality improvement |
| `clean design` | Reduces unnecessary elements |
| `modern aesthetic` | Modern design |
| `cinematic lighting` | Dramatic lighting |
| `4K render` | High resolution |
| `no text` | No text (for adding later) |

### Prompts to Avoid

| NG Pattern | Reason |
|------------|--------|
| Vague instructions | "A nice-looking image" → unstable results |
| Overly complex | Too many elements degrades quality |
| Text specification | AI-generated text has unstable quality |
| Copyrighted materials | Brand logos etc. cannot be generated |

---

## Execution Flow

```
Scene Generation Phase
    |
    +-- [Step 1] Asset Need Assessment
    |   +-- Check scene type, existing asset availability
    |       +-- Assets exist → skip
    |       +-- No assets → go to Step 2
    |
    +-- [Step 2] Prompt Generation
    |   +-- Build prompt from scene information
    |   +-- Reflect brand information (colors, style)
    |   +-- Apply template
    |
    +-- [Step 3] Image Generation (2 images in parallel)
    |   +-- Nano Banana Pro API call (2 parallel requests)
    |       generateContent x 2 (simultaneous requests to reduce latency)
    |
    +-- [Step 4] Quality Assessment
    |   +-- → See image-quality-check.md
    |
    +-- [Step 5] Result Processing
    |   +-- Success → save image, incorporate into scene
    |   +-- Failure → go to Step 6
    |
    +-- [Step 6] Regeneration Loop (max 3 times)
        +-- Prompt improvement (suggested by Claude)
        +-- Return to Step 3
```

---

## Bash Example

### API Call with curl

```bash
# Environment variable check (verify key is set)
test -n "$GOOGLE_AI_API_KEY" && echo "GOOGLE_AI_API_KEY is set" || echo "GOOGLE_AI_API_KEY is not set"

# Image generation request
curl -X POST \
  "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-pro-image-preview:generateContent" \
  -H "x-goog-api-key: ${GOOGLE_AI_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{
      "parts": [
        {"text": "Modern SaaS dashboard interface, clean design, light theme, professional UI"}
      ]
    }],
    "generationConfig": {
      "responseModalities": ["TEXT", "IMAGE"],
      "imageConfig": {
        "aspectRatio": "16:9",
        "imageSize": "2K"
      }
    }
  }' \
  -o response.json

# Base64 decode and save (extract image data from parts array)
cat response.json | jq -r '.candidates[0].content.parts[] | select(.inline_data) | .inline_data.data' | head -1 | base64 -d > out/assets/generated/image_1.png
```

> **Note**: One image is generated per request. If 2 images are needed, execute 2 requests.

### Image Save Location

```
out/
└── assets/
    └── generated/
        ├── intro_1.png
        ├── intro_2.png
        ├── cta_1.png
        └── cta_2.png
```

---

## Regeneration Loop Control

### Maximum Attempts

```
max_attempts = 3
```

### Prompt Improvement on Regeneration

Claude improves the prompt on each attempt:

| Attempt | Improvement Strategy |
|---------|---------------------|
| 1st | Generate with initial prompt |
| 2nd | Reflect quality feedback and adjust prompt |
| 3rd | Add more specific instructions, change style |

### Improvement Prompt Generation

```
The previous image was rejected for the following reasons:
- {rejection_reason}

Improvement suggestions:
1. {improvement_1}
2. {improvement_2}

New prompt:
{improved_prompt}
```

### Fallback After 3 Failures

```
⚠️ Image generation failed 3 times

Scene: {scene_name}
Last error: {last_error}

Options:
1. "Continue" → Proceed with placeholder image
2. "Skip" → Generate this scene without image
3. "Manual" → User provides image
```

---

## Error Handling

### API Errors

| Error Code | Cause | Action |
|------------|-------|--------|
| `400` | Invalid prompt | Check prompt content |
| `401` | Authentication failure | Verify API key |
| `429` | Rate limit | Wait 60 seconds and retry |
| `500` | Server error | Wait 30 seconds and retry |

### Content Policy Violation

```
⚠️ Content Policy Violation

The prompt violates Google's policies.
Please remove/modify the following:
- {violation_reason}

Would you like to attempt auto-correction? (y/n)
```

### Environment Variable Not Set

```
⚠️ GOOGLE_AI_API_KEY is not set

Setup instructions:
1. Get an API key from Google AI Studio
   https://ai.google.dev/aistudio

2. Set as environment variable
   export GOOGLE_AI_API_KEY="your-api-key"

3. Or add to .env.local
   GOOGLE_AI_API_KEY=your-api-key
```

---

## Cost Estimate

### Cost Per Scene

```
Base: 2 images x $0.12 = $0.24
Max (3 regenerations): 6 images x $0.12 = $0.72
```

### Cost Per Video Estimate

| Video Type | Scene Count | Image Generation Scenes | Cost Estimate |
|-----------|-------------|-------------------------|---------------|
| 90-sec teaser | 5 | 2-3 | $0.48-$0.72 |
| 3-min demo | 8 | 3-4 | $0.72-$0.96 |
| 5-min architecture | 12 | 4-6 | $0.96-$1.44 |

---

## Related Documents

- [image-quality-check.md](./image-quality-check.md) - Quality assessment logic
- [generator.md](./generator.md) - Parallel scene generation engine
- [planner.md](./planner.md) - Scenario planner
