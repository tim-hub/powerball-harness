# Slide Generator - Nano Banana Pro Slide Image Generation

Automatically generates project introduction slide images using Nano Banana Pro (Google DeepMind).

---

## Overview

This is the image generation logic executed in Step 3 of `/generate-slide`.
It generates 2 images for each of the 3 design patterns, then selects the best one after quality checks.

## Prerequisites

- `GOOGLE_AI_API_KEY` environment variable is set
- Nano Banana Pro (Gemini 3 Pro Image Preview) is enabled in Google AI Studio

---

## API Specification

> **Common Specification**: Uses the same Nano Banana Pro API as `generate-video/references/image-generator.md`.

### Endpoint

```
POST https://generativelanguage.googleapis.com/v1beta/models/gemini-3-pro-image-preview:generateContent
```

### Authentication

```bash
x-goog-api-key: ${GOOGLE_AI_API_KEY}
```

### Request Format

```json
{
  "contents": [{
    "parts": [
      {"text": "<slide prompt here>"}
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

### Response Format

```json
{
  "candidates": [{
    "content": {
      "parts": [
        {"text": "Description of the generated slide..."},
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

---

## Default Settings

| Setting | Value | Description |
|---------|-------|-------------|
| Model | `gemini-3-pro-image-preview` | Pro quality (recommended) |
| Aspect ratio | `16:9` | Presentation standard |
| Resolution | `2K` | 2048px, standard quality |
| responseModalities | `["TEXT", "IMAGE"]` | Text description + image |

### Aspect Ratio Options

| Ratio | Use Case |
|-------|----------|
| `16:9` | Presentations, screens (recommended) |
| `4:3` | Traditional presentations |
| `1:1` | Social media posts, icons |

---

## 3 Design Patterns

### Pattern A: Minimalist

**Concept**: Whitespace and typography-driven. Refined impression.

**Prompt Template**:

```
Create a minimalist project introduction slide for "{project_name}".

Project description: {project_description}
Key features: {features}

Design style:
- Clean whitespace-dominant layout
- Typography-driven hierarchy with bold project name
- Subtle accent color: {accent_color}
- {tone} aesthetic
- No cluttered elements, elegant simplicity
- Professional presentation quality, 2K resolution

Important: This is a single slide image, not a deck. Focus on clear visual hierarchy with the project name prominent and key value proposition visible.
```

**Visual Layout**:
```
+------------------------------------------+
|                                          |
|                                          |
|        PROJECT NAME                      |
|        _______________                   |
|                                          |
|        One-line description              |
|                                          |
|        * Feature 1                       |
|        * Feature 2                       |
|        * Feature 3                       |
|                                          |
+------------------------------------------+
```

### Pattern B: Infographic

**Concept**: Data and flow visualization. Information-rich but well-organized.

**Prompt Template**:

```
Create an infographic-style project introduction slide for "{project_name}".

Project description: {project_description}
Key features: {features}
Tech stack: {tech_stack}

Design style:
- Data visualization and structured layout
- Icons and visual elements for each feature
- Flow or architecture diagram elements
- Metrics and key numbers highlighted
- {tone} color palette with {accent_color} accents
- Professional infographic quality, 2K resolution

Important: This is a single slide image. Organize information visually with icons, sections, and clear data hierarchy. Make the project's value immediately understandable through visual structure.
```

**Visual Layout**:
```
+------------------------------------------+
|  PROJECT NAME          [icon] [icon]     |
|  ================                        |
|                                          |
|  [Feature 1]    [Feature 2]    [Feat 3]  |
|  +----------+   +----------+   +------+  |
|  | icon     |   | icon     |   | icon |  |
|  | detail   |   | detail   |   | det  |  |
|  +----------+   +----------+   +------+  |
|                                          |
|  Tech: [TS] [Node] [React]    v1.0      |
+------------------------------------------+
```

### Pattern C: Hero Visual

**Concept**: Bold visuals and catchcopy for maximum impact.

**Prompt Template**:

```
Create a hero-style project introduction slide for "{project_name}".

Project description: {project_description}
Key value: {key_value_proposition}

Design style:
- Bold, impactful hero image as background
- Large catchy headline text
- Dramatic visual composition
- {tone} mood with cinematic lighting
- Strong visual metaphor representing the project's purpose
- Professional marketing quality, 2K resolution

Important: This is a single slide image. Prioritize visual impact and emotional resonance. The project name and core value should be immediately visible with a compelling visual backdrop.
```

**Visual Layout**:
```
+------------------------------------------+
|                                          |
|    ==============================        |
|    ||  PROJECT NAME            ||        |
|    ||                          ||        |
|    ||  "Catchy tagline here"   ||        |
|    ||                          ||        |
|    ==============================        |
|                                          |
|         [ Bold Visual BG ]               |
|                                          |
+------------------------------------------+
```

---

## Prompt Structure

### Basic Structure

```
[Project overview] + [Design style] + [Quality specification] + [Constraints]
```

### Tone-Specific Modifiers

| Tone | Modifiers |
|------|-----------|
| Tech | `dark theme, code-inspired, terminal aesthetic, neon accents` |
| Casual | `bright colors, friendly, playful, approachable` |
| Corporate | `formal, trustworthy, blue tones, clean lines, business` |
| Creative | `bold, artistic, gradient, unconventional layout` |

### Quality Enhancement Keywords

| Keyword | Effect |
|---------|--------|
| `professional presentation quality` | Presentation quality |
| `clean design` | Reduces unnecessary elements |
| `2K resolution` | High resolution |
| `clear visual hierarchy` | Visual hierarchy |
| `modern aesthetic` | Modern design |

### Prompts to Avoid

| NG Pattern | Reason |
|------------|--------|
| Vague instructions | "A nice-looking slide" → unstable results |
| Overly complex | Too many elements degrades quality |
| Long text specification | AI-generated text has unstable quality. Keep to keywords |
| Copyrighted materials | Brand logos etc. cannot be generated |

---

## Bash Examples

### Environment Variable Check

```bash
test -n "$GOOGLE_AI_API_KEY" && echo "GOOGLE_AI_API_KEY is set" || { echo "GOOGLE_AI_API_KEY is not set"; exit 1; }
```

### Create Output Directory

```bash
mkdir -p out/slides/selected
```

### Image Generation with curl

```bash
PROMPT='Create a minimalist project introduction slide for "My Project". Clean whitespace-dominant layout, typography-driven, professional presentation quality, 2K resolution.'

curl -s -X POST \
  "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-pro-image-preview:generateContent" \
  -H "x-goog-api-key: ${GOOGLE_AI_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "{
    \"contents\": [{
      \"parts\": [
        {\"text\": \"${PROMPT}\"}
      ]
    }],
    \"generationConfig\": {
      \"responseModalities\": [\"TEXT\", \"IMAGE\"],
      \"imageConfig\": {
        \"aspectRatio\": \"16:9\",
        \"imageSize\": \"2K\"
      }
    }
  }" \
  -o /tmp/slide_response.json

# Base64 decode and save as PNG
cat /tmp/slide_response.json | jq -r '.candidates[0].content.parts[] | select(.inline_data) | .inline_data.data' | head -1 | base64 -d > out/slides/minimalist_1.png
```

> **Note**: One image is generated per request. If 2 images are needed, execute 2 requests.

### Parallel Generation (6 images at once)

```bash
mkdir -p out/slides/selected

generate_slide() {
  local pattern=$1
  local index=$2
  local prompt=$3
  local aspect_ratio=${4:-"16:9"}
  local image_size=${5:-"2K"}

  curl -s -X POST \
    "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-pro-image-preview:generateContent" \
    -H "x-goog-api-key: ${GOOGLE_AI_API_KEY}" \
    -H "Content-Type: application/json" \
    -d "{
      \"contents\": [{
        \"parts\": [
          {\"text\": \"${prompt}\"}
        ]
      }],
      \"generationConfig\": {
        \"responseModalities\": [\"TEXT\", \"IMAGE\"],
        \"imageConfig\": {
          \"aspectRatio\": \"${aspect_ratio}\",
          \"imageSize\": \"${image_size}\"
        }
      }
    }" \
    -o "/tmp/slide_${pattern}_${index}.json"

  # Base64 decode
  cat "/tmp/slide_${pattern}_${index}.json" \
    | jq -r '.candidates[0].content.parts[] | select(.inline_data) | .inline_data.data' \
    | head -1 \
    | base64 -d > "out/slides/${pattern}_${index}.png"
}

# Parallel execution (background jobs)
generate_slide "minimalist" "1" "$MINIMALIST_PROMPT" &
generate_slide "minimalist" "2" "$MINIMALIST_PROMPT" &
generate_slide "infographic" "1" "$INFOGRAPHIC_PROMPT" &
generate_slide "infographic" "2" "$INFOGRAPHIC_PROMPT" &
generate_slide "hero" "1" "$HERO_PROMPT" &
generate_slide "hero" "2" "$HERO_PROMPT" &
wait

echo "Generation of 6 images complete"
```

---

## Regeneration Prompt Improvement Strategy

### Improvement Per Attempt

| Attempt | Improvement Strategy |
|---------|---------------------|
| 1st | Generate with initial prompt |
| 2nd | Reflect quality feedback and adjust prompt (add specific modifiers) |
| 3rd | Significantly change style, add more specific composition instructions |

### Improvement by Problem Category

| Problem | Prompt Addition |
|---------|-----------------|
| Text unreadable | Add `no text elements, text-free design` |
| Layout broken | Add `balanced composition, grid-based layout` |
| Insufficient information | Explicitly specify feature names/numbers in prompt |
| Low professionalism | Add `corporate quality, polished, refined` |
| Colors don't match | Specify specific HEX color codes |

---

## Error Handling

### API Errors

| Error Code | Cause | Action |
|------------|-------|--------|
| `400` | Invalid prompt | Check and correct prompt content |
| `401` | Authentication failure | Verify API key |
| `429` | Rate limit | Wait 60 seconds and retry |
| `500` | Server error | Wait 30 seconds and retry |

### jq Parse Errors

If the response does not contain image data:

```bash
# Check response
cat /tmp/slide_response.json | jq '.candidates[0].content.parts | length'

# Check error message
cat /tmp/slide_response.json | jq '.error'
```

---

## Cost Estimate

### Per Execution

```
Base: 6 images x ~$0.06 = ~$0.36 (2K resolution)
Max (all patterns retry 3 times): 18 images x ~$0.06 = ~$1.08
```

### Cost by Resolution

| Resolution | Per Image | 6 Images (Base) | 18 Images (Max) |
|------------|-----------|-----------------|-----------------|
| `1K` | ~$0.02 | ~$0.12 | ~$0.36 |
| `2K` | ~$0.06 | ~$0.36 | ~$1.08 |
| `4K` | ~$0.12 | ~$0.72 | ~$2.16 |

---

## Related Documents

- [slide-quality-check.md](./slide-quality-check.md) — Quality assessment logic
- [generate-video/references/image-generator.md](../../generate-video/references/image-generator.md) — Common API specification (details)
- [generate-video/references/image-quality-check.md](../../generate-video/references/image-quality-check.md) — Video image quality check (reference)
