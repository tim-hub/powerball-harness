# Visual Effects Library

A collection of visual effect templates to add impact to videos.

---

## Color Palettes

### Cyberpunk / Neon (Recommended)

For impactful tech-oriented videos.

```tsx
const colors = {
  background: "#0A0A0F",  // Deep dark
  primary: "#00F5FF",     // Cyan
  secondary: "#FF00FF",   // Magenta
  accent: "#7B2FFF",      // Purple
  text: "#FFFFFF",
  glow: "rgba(0, 245, 255, 0.5)",
};
```

### Corporate / Professional

Calm tones for business use.

```tsx
const colors = {
  background: "#FFFFFF",
  primary: "#FF6B35",     // Orange
  secondary: "#004E89",   // Navy
  accent: "#2EC4B6",      // Teal
  text: "#1A1A2E",
};
```

---

## Effect Components

### GlitchText - Glitch Text Effect

RGB separation + random offset for cyberpunk-style text.

```tsx
import { useCurrentFrame, interpolate, random } from "remotion";

const GlitchText: React.FC<{
  text: string;
  fontSize?: number;
  startFrame?: number;
}> = ({ text, fontSize = 72, startFrame = 0 }) => {
  const frame = useCurrentFrame();
  const adjustedFrame = frame - startFrame;

  // Glitch intensity (decays over first 20 frames)
  const glitchIntensity = adjustedFrame < 20
    ? interpolate(adjustedFrame, [0, 20], [20, 0])
    : 0;
  const opacity = interpolate(adjustedFrame, [0, 15], [0, 1], {
    extrapolateRight: "clamp",
  });

  // Random offset
  const offsetX = glitchIntensity > 0
    ? (random(`x-${frame}`) - 0.5) * glitchIntensity
    : 0;
  const offsetY = glitchIntensity > 0
    ? (random(`y-${frame}`) - 0.5) * glitchIntensity * 0.5
    : 0;

  return (
    <div style={{ position: "relative", opacity }}>
      {/* Red channel (Magenta) */}
      <div
        style={{
          position: "absolute",
          fontSize,
          fontWeight: 800,
          color: "#FF00FF",
          transform: `translate(${offsetX - 3}px, ${offsetY}px)`,
          mixBlendMode: "screen",
          opacity: glitchIntensity > 0 ? 0.8 : 0,
        }}
      >
        {text}
      </div>
      {/* Blue channel (Cyan) */}
      <div
        style={{
          position: "absolute",
          fontSize,
          fontWeight: 800,
          color: "#00F5FF",
          transform: `translate(${offsetX + 3}px, ${offsetY}px)`,
          mixBlendMode: "screen",
          opacity: glitchIntensity > 0 ? 0.8 : 0,
        }}
      >
        {text}
      </div>
      {/* Main text */}
      <div
        style={{
          fontSize,
          fontWeight: 800,
          color: "#FFFFFF",
          textShadow: "0 0 20px rgba(0, 245, 255, 0.5)",
          transform: `translate(${offsetX}px, ${offsetY}px)`,
        }}
      >
        {text}
      </div>
    </div>
  );
};
```

**Usage**:
```tsx
<GlitchText text="Innovative Features" fontSize={64} startFrame={0} />
```

---

### Particles - Particle System

Floating and converging particle animation.

```tsx
import { useMemo } from "react";
import { useCurrentFrame, useVideoConfig, interpolate, random } from "remotion";

const Particles: React.FC<{
  count?: number;
  converge?: boolean;      // Converge to center
  convergeFrame?: number;  // Frame when convergence completes
}> = ({ count = 50, converge = false, convergeFrame = 100 }) => {
  const frame = useCurrentFrame();
  const { width, height } = useVideoConfig();

  // useMemo to fix particle initial positions (important!)
  const particles = useMemo(() => {
    return Array.from({ length: count }, (_, i) => ({
      id: i,
      startX: random(`px-${i}`) * width,
      startY: random(`py-${i}`) * height,
      speed: 0.5 + random(`speed-${i}`) * 2,
      size: 2 + random(`size-${i}`) * 4,
      hue: random(`hue-${i}`) > 0.5 ? "#00F5FF" : "#FF00FF",
    }));
  }, [count, width, height]);

  return (
    <div style={{ position: "absolute", inset: 0, overflow: "hidden" }}>
      {particles.map((p) => {
        const progress = converge
          ? interpolate(frame, [0, convergeFrame], [0, 1], {
              extrapolateRight: "clamp",
            })
          : 0;

        const targetX = width / 2;
        const targetY = height / 2;

        // Converge or float
        const x = converge
          ? interpolate(progress, [0, 1], [p.startX, targetX])
          : p.startX + Math.sin(frame * 0.02 * p.speed + p.id) * 30;
        const y = converge
          ? interpolate(progress, [0, 1], [p.startY, targetY])
          : p.startY + ((frame * p.speed * 0.5) % height);

        const opacity = converge
          ? interpolate(progress, [0, 0.8, 1], [0.8, 0.8, 0])
          : 0.6 + Math.sin(frame * 0.1 + p.id) * 0.4;

        return (
          <div
            key={p.id}
            style={{
              position: "absolute",
              left: x,
              top: y % height,
              width: p.size,
              height: p.size,
              borderRadius: "50%",
              backgroundColor: p.hue,
              boxShadow: `0 0 ${p.size * 2}px ${p.hue}`,
              opacity,
            }}
          />
        );
      })}
    </div>
  );
};
```

**Usage**:
```tsx
{/* Floating particles */}
<Particles count={80} />

{/* Converging particles (for CTA scenes) */}
<Particles count={100} converge convergeFrame={150} />
```

---

### ScanLine - Scan Line Effect

An analysis wave effect that sweeps across the screen.

```tsx
const ScanLine: React.FC<{ speed?: number }> = ({ speed = 1 }) => {
  const frame = useCurrentFrame();
  const { height } = useVideoConfig();
  const y = (frame * speed * 5) % (height + 100);

  return (
    <div
      style={{
        position: "absolute",
        left: 0,
        right: 0,
        top: y - 50,
        height: 100,
        background: `linear-gradient(180deg, transparent, #00F5FF40, transparent)`,
        boxShadow: "0 0 60px #00F5FF",
      }}
    />
  );
};
```

**Usage**:
```tsx
{/* Analysis effect */}
{frame < 60 && <ScanLine speed={3} />}
```

---

### ProgressBar - Progress Bar

Visualizes progress of parallel processing.

```tsx
const ProgressBar: React.FC<{ progress: number; label: string }> = ({
  progress,
  label,
}) => {
  return (
    <div style={{ width: 400, marginBottom: 16 }}>
      <div
        style={{
          fontSize: 18,
          color: "#FFFFFF",
          marginBottom: 8,
          fontFamily: "monospace",
        }}
      >
        {label}
      </div>
      <div
        style={{
          height: 8,
          background: "rgba(255,255,255,0.1)",
          borderRadius: 4,
          overflow: "hidden",
        }}
      >
        <div
          style={{
            width: `${progress * 100}%`,
            height: "100%",
            background: "linear-gradient(90deg, #00F5FF, #FF00FF)",
            boxShadow: "0 0 20px #00F5FF",
            borderRadius: 4,
          }}
        />
      </div>
    </div>
  );
};
```

**Usage**:
```tsx
const agents = [
  { name: "Agent 1: Intro", progress: Math.min(1, frame / 150) },
  { name: "Agent 2: Demo", progress: Math.min(1, (frame - 30) / 180) },
  { name: "Agent 3: CTA", progress: Math.min(1, (frame - 60) / 120) },
];

{agents.map((agent) => (
  <ProgressBar key={agent.name} progress={agent.progress} label={agent.name} />
))}
```

---

### 3D Parallax - Parallax Effect

3D card display with depth.

```tsx
const ParallaxCard: React.FC<{
  children: React.ReactNode;
  delay: number;
  color: string;
}> = ({ children, delay, color }) => {
  const frame = useCurrentFrame();

  const opacity = interpolate(frame, [delay, delay + 30], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const z = interpolate(frame, [delay, delay + 30], [-100, 0], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const rotateY = interpolate(frame, [delay, delay + 30], [45, 0], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

  return (
    <div
      style={{
        width: 280,
        height: 160,
        background: `linear-gradient(135deg, ${color}30, ${color}10)`,
        border: `2px solid ${color}`,
        borderRadius: 16,
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        opacity,
        transform: `translateZ(${z}px) rotateY(${rotateY}deg)`,
        boxShadow: `0 0 40px ${color}40`,
      }}
    >
      {children}
    </div>
  );
};
```

**Usage**:
```tsx
<div style={{ display: "flex", gap: 40, perspective: 1000 }}>
  <ParallaxCard delay={30} color="#00F5FF">LP/Ads</ParallaxCard>
  <ParallaxCard delay={70} color="#FF00FF">Intro Demo</ParallaxCard>
  <ParallaxCard delay={110} color="#7B2FFF">Release Notes</ParallaxCard>
</div>
```

---

## Combination Examples

### Impact-Focused Hook Scene

```tsx
const HookScene: React.FC = () => {
  const frame = useCurrentFrame();

  return (
    <AbsoluteFill style={{ background: "#0A0A0F" }}>
      <Particles count={80} />
      <div
        style={{
          position: "absolute",
          inset: 0,
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <GlitchText text="Videos auto-generated" fontSize={64} startFrame={0} />
        <div style={{ height: 20 }} />
        <GlitchText text="from code" fontSize={64} startFrame={15} />
      </div>
      {frame < 30 && <ScanLine speed={3} />}
    </AbsoluteFill>
  );
};
```

### CTA Scene (Particle Convergence)

```tsx
const CTAScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const logoScale = spring({ frame: frame - 60, fps, config: { damping: 200 } });
  const pulse = Math.sin(frame / 10) * 0.03 + 1;

  return (
    <AbsoluteFill style={{ background: "#0A0A0F" }}>
      <Particles count={100} converge convergeFrame={150} />
      <div
        style={{
          position: "absolute",
          inset: 0,
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <div
          style={{
            opacity: interpolate(frame, [60, 90], [0, 1], {
              extrapolateRight: "clamp",
            }),
            transform: `scale(${Math.max(0, logoScale)})`,
          }}
        >
          <Img src={staticFile("logo.png")} style={{ width: 120, height: 120 }} />
        </div>
        <div
          style={{
            marginTop: 40,
            padding: "16px 48px",
            background: "linear-gradient(90deg, #00F5FF, #FF00FF)",
            borderRadius: 12,
            fontSize: 24,
            fontWeight: 700,
            color: "#0A0A0F",
            transform: `scale(${pulse})`,
            boxShadow: "0 0 40px rgba(0, 245, 255, 0.6)",
          }}
        >
          Try it now
        </div>
      </div>
    </AbsoluteFill>
  );
};
```

---

## Notes

| Item | Rule |
|------|------|
| `random()` | Seed argument required (same value per frame) |
| `useMemo` | Must memoize large object collections like particles |
| `interpolate` | Use `extrapolateRight: "clamp"` to prevent value runaway |
| `spring` | Use `config: { damping: 200 }` for smooth animation |
| CSS animations | Prohibited, use Remotion's `useCurrentFrame()` |

---

## References

- [generator.md](generator.md) - Parallel generation engine
- [best-practices.md](best-practices.md) - Video production best practices
