# Video Generator - Parallel Scene Generation Engine

Generates scenes in parallel using multi-agent architecture based on a scenario.

---

## Overview

This is the generation engine executed in Step 3 of `/generate-video`.
It receives the scenario from planner.md, generates each scene in parallel, and then integrates them.

## Input

Scenario from planner.md:
- Scene list (id, name, duration, template, content)
- Video settings (resolution, fps)

## Parallel Generation Architecture

```
Scenario (N scenes)
    |
    +--[Asset Generation Phase] <- NEW
    |   +-- Determine asset requirements for each scene
    |   +-- Generate images with Nano Banana Pro (2 images: 2 requests)
    |   +-- Claude performs quality assessment
    |   +-- OK -> Accept / NG -> Regenerate (max 3 times)
    |
    +--[Parallelism Determination]
    |   +-- Use min(scene count, 5) as parallel count
    |
    +--[Parallel Generation Phase]
    |   +-- Agent 1: Generate Scene 1
    |   +-- Agent 2: Generate Scene 2
    |   +-- Agent 3: Generate Scene 3
    |   +-- ... (max 5 parallel)
    |
    +--[Integration Phase]
    |   +-- Scene assembly
    |   +-- Add transitions
    |   +-- Audio sync (optional)
    |
    +--[Rendering Phase]
        +-- Final output (mp4/webm/gif)
```

---

## Asset Generation Phase (Nano Banana Pro)

Automatically generates required asset images before scene generation.

### Asset Requirement Determination

| Scene Type | Asset Required | Reason |
|------------|---------------|--------|
| intro | ✅ Required | Logo, title card |
| cta | ✅ Required | Action banner |
| architecture | ✅ Required | Concept diagram |
| ui-demo | ❌ Not required | Uses Playwright capture |
| changelog | ❌ Not required | Text-based |

### Determination Logic

```javascript
const needsGeneratedAsset = (scene) => {
  // Skip if existing assets are available
  if (scene.existingAssets?.length > 0) return false;

  // Skip Playwright capture targets
  if (scene.template === 'ui-demo') return false;

  // Skip text-based scenes
  if (scene.template === 'changelog') return false;

  // Everything else is a generation target
  return ['intro', 'cta', 'architecture', 'feature-highlight'].includes(scene.template);
};
```

### Generation Flow

```
For each scene:
    |
    +-- needsGeneratedAsset(scene) = false
    |   +-- Skip -> Next scene
    |
    +-- needsGeneratedAsset(scene) = true
        |
        +-- [Step 1] Generate prompt
        |   +-- Build prompt from scene info + brand info
        |
        +-- [Step 2] Generate images (2 images: 2 requests)
        |   +-- Call Nano Banana Pro API (generateContent x 2)
        |   +-- -> See image-generator.md
        |
        +-- [Step 3] Quality assessment
        |   +-- Claude evaluates and selects from 2 images
        |   +-- -> See image-quality-check.md
        |
        +-- [Step 4] Process results
            +-- Success -> out/assets/generated/{scene_name}.png
            +-- Failure -> Regenerate (max 3 times) or fallback
```

### Generated Image Storage

```
out/
+-- assets/
    +-- generated/
        +-- intro.png
        +-- cta.png
        +-- architecture.png
        +-- feature-highlight.png
```

### Integration into Scenes

Generated images are passed to scene generation agents:

```
Task:
  subagent_type: "video-scene-generator"
  prompt: |
    Scene information:
    - Name: intro
    - Template: intro
    - Generated image: out/assets/generated/intro.png  <- Added

    Use the generated image as a background or main element.
```

### Detailed Documentation

- [image-generator.md](./image-generator.md) - API calls, prompt design
- [image-quality-check.md](./image-quality-check.md) - Quality assessment logic

---

## Parallelism Determination Logic

| Scene Count | Parallel Count | Reason |
|-------------|---------------|--------|
| 1-2 | 1-2 | Overhead exceeds benefit |
| 3-4 | 3 | Optimal balance |
| 5+ | 5 | Beyond this, resource contention occurs |

**Implementation**:
```javascript
const parallelCount = Math.min(scenes.length, 5);
```

---

## Parallel JSON Generation via Task Tool

### New Generation Flow (JSON-schema driven)

```
Scenario (scenario.json)
    |
+---------------------------------------------+
|     Parallel Task Launch (each scene -> JSON)|
+---------------------------------------------+
| Agent 1 -> scenes/intro.json                 |
| Agent 2 -> scenes/auth-demo.json             |
| Agent 3 -> scenes/dashboard.json             |
| Agent 4 -> scenes/features.json              |
| Agent 5 -> scenes/cta.json                   |
+---------------------------------------------+
    |
+---------------------------------------------+
|         scenes/*.json -> Merge               |
+---------------------------------------------+
| - Sort by section_id + order                 |
| - Conflict detection (same scene_id = Crit.) |
| - Missing detection (section has no scenes)  |
+---------------------------------------------+
    |
video-script.json (all scenes integrated)
    |
Remotion rendering
```

### Scene Generation Agent Launch (JSON Output)

```
Launch Task tool for each scene:

Task:
  subagent_type: "video-scene-generator"
  run_in_background: true
  prompt: |
    Generate the JSON for the following scene according to scene.schema.json.

    Scene information:
    - scene_id: {scene.id}
    - section_id: {section.id}
    - order: {scene.order} (order within section)
    - type: {scene.type}
    - duration_ms: {scene.duration_ms}
    - content: {scene.content}

    Output path: out/video-{date}-{id}/scenes/{scene_id}.json

    Required fields:
    - scene_id, section_id, order, type, content
    - content.duration_ms (considering audio length + padding)
    - direction (transition, emphasis, background, timing)
    - assets (image/audio files used)

    Validation:
    ```bash
    node scripts/validate-scene.js out/video-{date}-{id}/scenes/{scene_id}.json
    ```

    Completion report:
    - File path
    - Validation result (PASS/FAIL)
    - Report any warnings
```

### Progress Monitoring

```
🎬 Parallel JSON generation in progress... (3/5 complete)

+-- [Agent 1] intro.json ✅ PASS
+-- [Agent 2] auth-demo.json ✅ PASS
+-- [Agent 3] dashboard.json ⏳ Generating...
+-- [Agent 4] features.json 🔜 Waiting
+-- [Agent 5] cta.json 🔜 Waiting
```

### Result Collection (JSON)

```
Collect results from each agent via TaskOutput:

Results:
  - scene_id: "intro"
    file: "out/video-20260202-001/scenes/intro.json"
    validation: "PASS"
    status: "success"

  - scene_id: "auth-demo"
    file: "out/video-20260202-001/scenes/auth-demo.json"
    validation: "PASS"
    status: "success"
    warnings: ["duration_ms may be shorter than audio length"]
```

### JSON Output Specification

**Output file**: `out/video-{date}-{id}/scenes/{scene_id}.json`

**Schema**: `schemas/scene.schema.json`

**Required fields**:
```json
{
  "scene_id": "intro",
  "section_id": "opening",
  "order": 0,
  "type": "intro",
  "content": {
    "title": "MyApp",
    "subtitle": "Task management made easy",
    "duration_ms": 5000
  },
  "direction": {
    "transition": {
      "in": "fade",
      "out": "fade",
      "duration_ms": 500
    },
    "emphasis": {
      "level": "high"
    },
    "background": {
      "type": "gradient",
      "value": "linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
    }
  },
  "assets": [
    {
      "type": "image",
      "source": "assets/generated/intro.png",
      "generated": true
    }
  ]
}
```

### Merge Phase

After all agents complete, run `scripts/merge-scenes.js`:

```bash
node scripts/merge-scenes.js out/video-20260202-001/
```

**Processing**:
1. Read `scenes/*.json`
2. Sort by `section_id` + `order`
3. Conflict detection (same `scene_id` -> Critical error)
4. Missing detection (section has no scenes -> Critical error)
5. Generate `video-script.json`

**Output**: `out/video-20260202-001/video-script.json`

**Format**:
```json
{
  "scenes": [
    { "scene_id": "intro", "section_id": "opening", "order": 0, ... },
    { "scene_id": "hook", "section_id": "opening", "order": 1, ... },
    { "scene_id": "demo", "section_id": "main", "order": 0, ... }
  ],
  "metadata": {
    "total_duration_ms": 180000,
    "scene_count": 12,
    "generated_at": "2026-02-02T12:34:56Z"
  }
}
```

---

## Scene Generation Templates

### intro Template

```tsx
// remotion/src/scenes/intro.tsx
import { AbsoluteFill, useCurrentFrame, interpolate } from "remotion";
import { FadeIn } from "../components/FadeIn";

export const IntroScene: React.FC<{
  title: string;
  tagline: string;
}> = ({ title, tagline }) => {
  const frame = useCurrentFrame();
  const opacity = interpolate(frame, [0, 30], [0, 1]);

  return (
    <AbsoluteFill style={{ backgroundColor: "#000", opacity }}>
      <FadeIn durationInFrames={30}>
        <h1>{title}</h1>
        <p>{tagline}</p>
      </FadeIn>
    </AbsoluteFill>
  );
};

export const DURATION = 150; // 5 seconds @ 30fps
```

### ui-demo Template (Playwright Integration)

```tsx
// remotion/src/scenes/ui-demo.tsx
import { AbsoluteFill, Img, Sequence } from "remotion";

export const UIDemoScene: React.FC<{
  screenshots: string[];
  duration: number;
}> = ({ screenshots, duration }) => {
  const framePerScreenshot = Math.floor(duration / screenshots.length);

  return (
    <AbsoluteFill>
      {screenshots.map((src, i) => (
        <Sequence from={i * framePerScreenshot} durationInFrames={framePerScreenshot}>
          <Img src={src} style={{ width: "100%", height: "100%" }} />
        </Sequence>
      ))}
    </AbsoluteFill>
  );
};
```

### cta Template

```tsx
// remotion/src/scenes/cta.tsx
import { AbsoluteFill, useCurrentFrame, interpolate } from "remotion";

export const CTAScene: React.FC<{
  url: string;
  text: string;
}> = ({ url, text }) => {
  const frame = useCurrentFrame();
  const scale = interpolate(frame, [0, 15], [0.8, 1], {
    extrapolateRight: "clamp",
  });

  return (
    <AbsoluteFill style={{ backgroundColor: "#1a1a1a" }}>
      <div style={{ transform: `scale(${scale})` }}>
        <h2>{text}</h2>
        <p>{url}</p>
      </div>
    </AbsoluteFill>
  );
};

export const DURATION = 150; // 5 seconds @ 30fps
```

---

## Audio Sync Rules (Important)

When generating videos with narration, strictly follow these rules.

### 1. Pre-check Audio File Duration

```bash
# Check the duration of each audio file
for f in public/audio/*.wav; do
  name=$(basename "$f" .wav)
  dur=$(ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "$f")
  frames=$(echo "$dur * 30" | bc | cut -d. -f1)
  echo "$name: ${dur}s = ${frames} frames"
done
```

### 2. Scene Duration Formula

```
Scene duration = 1s wait (30f) + audio duration + pre-transition padding (20f+)
```

| Element | Frames | Description |
|---------|--------|-------------|
| 1s wait | 30f | Wait for visual settling after scene start before audio begins |
| Audio duration | Variable | Pre-check with ffprobe |
| Padding | 20f+ | Ensure audio ends before transition starts |

### 3. Audio Start Timing

```
Audio start = Scene start frame + 30 frames (1s wait)
```

### 4. Scene Start Frame Calculation (When Using TransitionSeries)

```
Scene start frame = Previous scene start + Previous scene duration - Transition duration
```

**Example (with 15-frame transitions)**:
```
hook:       0
problem:    175 - 15 = 160
solution:   160 + 415 - 15 = 560
workPlan:   560 + 340 - 15 = 885
...
```

### 5. Implementation Template

```tsx
const SCENE_DURATIONS = {
  hook: 175,      // 30 + 121(audio) + 24(padding)
  problem: 415,   // 30 + 360(audio) + 25(padding)
  solution: 340,  // 30 + 286(audio) + 24(padding)
  // ...
};
const TRANSITION = 15;

// Scene start frames (cumulative calculation)
// hook:0, problem:160, solution:560, ...

const audioTimings = {
  hook: 30,       // scene 0 + 30
  problem: 190,   // scene 160 + 30
  solution: 590,  // scene 560 + 30
  // ...
};
```

### 6. Common Issues and Solutions

| Issue | Cause | Solution |
|-------|-------|----------|
| Audio overlapping | Next audio starts before previous ends | Check audio duration and adjust scene duration |
| Slide change out of sync with audio | TransitionSeries overlap not considered | Scene start = prev scene start + prev scene duration - transition duration |
| Audio cut off mid-playback | Scene duration < audio duration | Adjust scene duration to audio duration + padding |
| Long silence | Audio starts too late | Standardize at scene start + 30f |

---

## Integration Phase

### Scene Assembly

```tsx
// remotion/src/FullVideo.tsx
import { Composition, Series } from "remotion";
import { IntroScene } from "./scenes/intro";
import { UIDemoScene } from "./scenes/ui-demo";
import { CTAScene } from "./scenes/cta";

export const FullVideo: React.FC = () => {
  return (
    <Series>
      <Series.Sequence durationInFrames={150}>
        <IntroScene title="MyApp" tagline="Task management made easy" />
      </Series.Sequence>
      <Series.Sequence durationInFrames={450}>
        <UIDemoScene screenshots={[...]} duration={450} />
      </Series.Sequence>
      <Series.Sequence durationInFrames={150}>
        <CTAScene url="https://myapp.com" text="Try it now" />
      </Series.Sequence>
    </Series>
  );
};
```

### Adding Transitions

```tsx
// Transition component
import { TransitionSeries, linearTiming } from "@remotion/transitions";
import { fade } from "@remotion/transitions/fade";

<TransitionSeries>
  <TransitionSeries.Sequence durationInFrames={150}>
    <IntroScene {...} />
  </TransitionSeries.Sequence>
  <TransitionSeries.Transition
    presentation={fade()}
    timing={linearTiming({ durationInFrames: 15 })}
  />
  <TransitionSeries.Sequence durationInFrames={450}>
    <UIDemoScene {...} />
  </TransitionSeries.Sequence>
</TransitionSeries>
```

---

## Rendering Phase

### Command Execution

```bash
# MP4 rendering
npx remotion render remotion/index.ts FullVideo out/video.mp4

# GIF rendering (for short videos)
npx remotion render remotion/index.ts FullVideo out/video.gif

# WebM rendering (for web)
npx remotion render remotion/index.ts FullVideo out/video.webm --codec=vp8
```

### Output Options

| Format | Recommended Use | Options |
|--------|----------------|---------|
| MP4 | General purpose, social media | `--codec=h264` |
| WebM | Web embedding | `--codec=vp8` |
| GIF | Short loops | Recommended under 15 seconds |

---

## Completion Report

```markdown
✅ **Video generation complete**

📁 **Output files**:
- `out/video.mp4` (45s, 1080p, 12.3MB)

📊 **Generation statistics**:
| Item | Value |
|------|-------|
| Scene count | 4 |
| Parallel agent count | 3 |
| Generation time | 45s |
| Rendering time | 30s |

🎬 **Preview**:
- Studio: `npm run remotion` -> http://localhost:3000
- File: `open out/video.mp4`
```

---

## Error Handling

### Scene Generation Failure

```
⚠️ Scene generation error

Scene "auth-demo" generation failed.
Cause: Playwright capture failed - application is not running

Actions:
1. Start the application: `npm run dev`
2. Regenerate: "Regenerate auth-demo"
3. Skip: "Skip this scene"
```

### Rendering Failure

```
⚠️ Rendering error

Cause: Out of memory

Actions:
1. Reduce parallel count: `--concurrency 2`
2. Lower resolution: Retry at 720p
3. Split scenes: Break long scenes into shorter ones
```

---

## BGM Support

### Implementation

Add `bgmPath` and `bgmVolume` properties to the composition:

```tsx
export const VideoComposition: React.FC<{
  enableAudio?: boolean;
  volume?: number;
  bgmPath?: string;      // BGM file path (relative to staticFile)
  bgmVolume?: number;    // BGM volume (0.0-1.0)
}> = ({ enableAudio = true, volume = 1, bgmPath, bgmVolume = 0.25 }) => {
  return (
    <AbsoluteFill>
      {/* Scene content */}

      {/* BGM (lower volume than narration) */}
      {enableAudio && bgmPath && (
        <Audio src={staticFile(bgmPath)} volume={bgmVolume} />
      )}
    </AbsoluteFill>
  );
};
```

### BGM Volume Guidelines

| Narration Present | Recommended bgmVolume |
|-------------------|----------------------|
| Yes | 0.20 - 0.30 |
| No | 0.50 - 0.80 |

### Royalty-Free BGM Sources

- [DOVA-SYNDROME](https://dova-s.jp/) - Japanese, free
- [Amachamusic](https://amachamusic.chagasi.com/) - Japanese, free
- [Pixabay Music](https://pixabay.com/music/) - English, free

---

## Subtitle Support

### Implementation

```tsx
// Font embedding (Base64 recommended)
const FontStyle: React.FC = () => (
  <style>
    {`
      @font-face {
        font-family: 'CustomFont';
        src: url('${FONT_DATA_URL}') format('opentype');
        font-weight: normal;
        font-style: normal;
      }
    `}
  </style>
);

// Subtitle component
const Subtitle: React.FC<{ text: string }> = ({ text }) => {
  const frame = useCurrentFrame();
  const opacity = interpolate(frame, [0, 10], [0, 1], {
    extrapolateRight: "clamp",
  });

  return (
    <>
      <FontStyle />
      <div
        style={{
          position: "absolute",
          bottom: 80,
          left: 0,
          right: 0,
          display: "flex",
          justifyContent: "center",
          padding: "0 60px",
        }}
      >
        <div
          style={{
            fontFamily: "'CustomFont', sans-serif",
            fontSize: 32,
            color: "#FFFFFF",
            backgroundColor: "rgba(0, 0, 0, 0.8)",
            padding: "14px 28px",
            borderRadius: 8,
            textAlign: "center",
            maxWidth: 1000,
            lineHeight: 1.5,
            opacity,
          }}
        >
          {text}
        </div>
      </div>
    </>
  );
};
```

### Subtitle Timing Rules

| Item | Value |
|------|-------|
| Subtitle start | Same timing as audio start |
| Subtitle duration | Audio duration + 10f (padding) |

### Font Embedding (Base64)

To reliably load custom fonts, use Base64 embedding:

```typescript
// src/utils/custom-font.ts
import fs from "fs";
import path from "path";

// Base64 encode at build time
const fontPath = path.join(__dirname, "../../public/font/MyFont.otf");
const fontBuffer = fs.readFileSync(fontPath);
export const FONT_DATA_URL = `data:font/otf;base64,${fontBuffer.toString("base64")}`;
```

### Subtitle Data Structure

```tsx
const SUBTITLES = [
  { id: "hook", text: "Subtitle text", start: 30, duration: 120 },
  { id: "problem", text: "Next subtitle", start: 175, duration: 178 },
  // ...
];

// Usage
{SUBTITLES.map((sub) => (
  <Sequence key={sub.id} from={sub.start} durationInFrames={sub.duration}>
    <Subtitle text={sub.text} />
  </Sequence>
))}
```

---

## Notes

- Parallel generation is only effective for independent scenes
- Playwright capture requires the application to be running beforehand
- Split rendering is recommended for long videos (3+ minutes)
- Keep BGM volume low enough that narration remains audible
- Use Base64 embedding for reliable custom font loading

---

## Phase 10: Future Extension (Character Dialogue Videos)

### Overview

The current video generation supports **single narration** format, but the design allows future extension to **character dialogue videos** as follows:

| Current | After Phase 10 Extension |
|---------|-------------------------|
| Single narrator | Multi-character dialogue |
| Static slides + audio | Character display + dialogue effects |
| TTS: single voice only | TTS: per-character voice |

### Use Case Examples

```
[Intro video example]

Narrator:  "Today we'll introduce a new feature"
User:      "What can this do?"
AI Guide:  "Let me explain briefly"
```

```
[Technical explanation video example]

Interviewer: "What are the characteristics of this architecture?"
Expert:      "We focused on scalability"
Reviewer:    "Let's look at the specific numbers"
```

### Extension Points (Design Only)

#### 1. Character Definition (`schemas/character.schema.json`)

**Already implemented** schema that defines the following:

```json
{
  "character_id": "narrator",
  "name": "Narrator",
  "role": "narrator",
  "voice": {
    "provider": "google-cloud-tts",
    "voice_id": "ja-JP-Neural2-B",
    "language": "ja",
    "speed": 1.1,
    "style": "professional"
  },
  "appearance": {
    "type": "avatar",
    "position": "left"
  }
}
```

**Extension areas**:
- `voice`: TTS settings (provider, voice ID, speed, style)
- `appearance`: Visual settings (avatar, icon, position)
- `dialogue_style`: Dialogue effects (speech bubble style, animations)
- `personality`: Personality traits (for future AI dialogue generation)

#### 2. Dialogue Scene Definition (Future Specification)

**dialogue.json** structure (implementation is Phase 10+):

```json
{
  "scene_id": "intro-dialogue",
  "type": "dialogue",
  "content": {
    "duration_ms": 15000,
    "exchanges": [
      {
        "character_id": "user",
        "text": "What can this feature do?",
        "timing_ms": 0,
        "duration_ms": 3000,
        "emotion": "curious"
      },
      {
        "character_id": "guide",
        "text": "Let me explain. First...",
        "timing_ms": 3500,
        "duration_ms": 5000,
        "emotion": "friendly"
      },
      {
        "character_id": "narrator",
        "text": "Let's look at the actual screen",
        "timing_ms": 9000,
        "duration_ms": 3000,
        "emotion": "neutral"
      }
    ]
  },
  "characters": [
    {
      "$ref": "characters/user.json"
    },
    {
      "$ref": "characters/guide.json"
    },
    {
      "$ref": "characters/narrator.json"
    }
  ],
  "direction": {
    "layout": "split-screen",
    "transition_between_speakers": "highlight"
  }
}
```

#### 3. TTS Integration Extension

**Current (single voice)**:
```javascript
// Play a single audio file
<Audio src={staticFile('narration.wav')} />
```

**After Phase 10 extension (per-character voice)**:
```javascript
// Call TTS per character
async function generateDialogue(exchanges, characters) {
  const audioFiles = await Promise.all(
    exchanges.map(async (exchange) => {
      const character = characters.find(c => c.character_id === exchange.character_id);

      // Call TTS API (branch based on provider)
      const audioBuffer = await ttsProvider.synthesize({
        text: exchange.text,
        voiceId: character.voice.voice_id,
        speed: character.voice.speed,
        emotion: exchange.emotion,
      });

      return {
        character_id: exchange.character_id,
        audio: audioBuffer,
        timing_ms: exchange.timing_ms,
        duration_ms: exchange.duration_ms,
      };
    })
  );

  return audioFiles;
}
```

**TTS Provider Integration**:

| Provider | API Call Example |
|----------|----------------|
| Google Cloud TTS | `textToSpeech.synthesizeSpeech({ voice, input })` |
| ElevenLabs | `elevenlabs.textToSpeech({ voiceId, text })` |
| OpenAI TTS | `openai.audio.speech.create({ voice, input })` |
| AWS Polly | `polly.synthesizeSpeech({ VoiceId, Text })` |

#### 4. Visual Effect Extensions

**Character Display (Remotion component example)**:

```tsx
// Future implementation: DialogueScene.tsx
const DialogueScene: React.FC<{
  exchanges: Exchange[];
  characters: Character[];
}> = ({ exchanges, characters }) => {
  const frame = useCurrentFrame();

  return (
    <AbsoluteFill>
      {/* Background */}
      <Background />

      {/* Character display */}
      <CharacterDisplay
        characters={characters}
        activeCharacterId={getCurrentSpeaker(frame, exchanges)}
      />

      {/* Dialogue text (speech bubble) */}
      <DialogueBubble
        exchange={getCurrentExchange(frame, exchanges)}
      />

      {/* Audio playback */}
      {exchanges.map((ex, i) => (
        <Sequence from={ex.timing_ms / 33.33} durationInFrames={ex.duration_ms / 33.33}>
          <Audio src={staticFile(`dialogue/${ex.character_id}_${i}.wav`)} />
        </Sequence>
      ))}
    </AbsoluteFill>
  );
};
```

**Animation examples**:
- Highlight the speaking character
- Make non-speaking characters semi-transparent
- Speech bubbles fade in/out
- Character avatar lip sync (optional)

#### 5. Implementation Roadmap (Phase 10+)

| Phase | Implementation | Priority |
|-------|---------------|----------|
| **Phase 10.1** | `character.schema.json` implementation | ✅ Complete |
| **Phase 10.2** | TTS provider integration (Google Cloud TTS) | High |
| **Phase 10.3** | `DialogueScene` Remotion component | High |
| **Phase 10.4** | `dialogue.json` schema definition | Medium |
| **Phase 10.5** | Character display UI (avatar/icon) | Medium |
| **Phase 10.6** | Speech bubble animations | Low |
| **Phase 10.7** | Multiple TTS provider support (ElevenLabs, OpenAI) | Low |
| **Phase 10.8** | AI dialogue generation (based on personality) | Future |

#### 6. Maintaining Compatibility

The extension is designed to **maintain backward compatibility**:

```
Existing video-script.json (single narration)
    | Works as-is
New dialogue.json (dialogue format)
    | Added as a new scene type
Both can coexist
```

**Addition to scene.schema.json**:
```json
{
  "type": {
    "enum": [
      "intro",
      "ui-demo",
      "dialogue",  // <- Added in Phase 10
      "..."
    ]
  }
}
```

#### 7. Reference Implementations

Examples from existing projects:
- **Manim Community**: Character animation
- **Remotion Templates**: Dialogue format templates
- **Google Cloud TTS**: Multi-language, multi-voice support

---

### Phase 10 Implementation Checklist

Verify the following when implementing in the future:

- [ ] `character.schema.json` is valid (already completed in Phase 10.1)
- [ ] TTS API key is configured (Google Cloud TTS recommended)
- [ ] `dialogue.json` schema is defined
- [ ] `DialogueScene.tsx` Remotion component is implemented
- [ ] Character audio file naming conventions are unified
- [ ] Speech bubble styles maintain brand consistency
- [ ] Coexistence testing with existing scenes (intro, ui-demo, etc.)
- [ ] Performance: optimize simultaneous rendering of multiple audio tracks

---

### Summary (Phase 10)

**Current**: Supports single narration videos
**Phase 10 Design**: Extension points for character dialogue videos are clearly defined
**Implemented**: `character.schema.json` (character definitions)
**Not Implemented**: TTS integration, dialogue scenes, visual effects (future implementation)

This design enables the following in the future:
- Multi-character dialogue format videos
- Per-character voice styles
- Visual character display and dialogue effects
- AI-powered dialogue generation (based on personality settings)
