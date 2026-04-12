# Claude harness Introductory X Post Collection

> **Note**: This file is excluded via `.gitignore` (local use only)

Each post includes a corresponding slide number.
Slides can be batch-generated with `notebooklm-slides-for-x-posts.yaml`.

---

## Post List

| ID | Target | Type | Slide |
|----|-----------|--------|---------|
| E1 | Engineers | Problem statement | slide_E1 |
| E2 | Engineers | Technical details | slide_E2 |
| E3 | Engineers | Before/After | slide_E3 |
| V1 | Vibe coders | Empathy | slide_V1 |
| V2 | Vibe coders | Concrete examples | slide_V2 |
| V3 | Vibe coders | Reassurance | slide_V3 |
| N1 | Interested newcomers | Lowering barriers | slide_N1 |
| N2 | Interested newcomers | Addressing concerns | slide_N2 |
| N3 | Interested newcomers | Call to action | slide_N3 |
| R1 | Returning users | Empathy | slide_R1 |
| R2 | Returning users | Improvements | slide_R2 |
| R3 | Returning users | Encouragement | slide_R3 |
| G1 | General | Overview | slide_G1 |
| G2 | General | Value proposition | slide_G2 |

---

## For Engineers (E1-E3)

### E1: Problem Statement

```
[Post text]

Solved the problem of AI "fixing" test failures by modifying the tests themselves.

Claude harness's 3-layer defense:
Rules - Always active
Skills - Auto-triggered based on context
Hooks - Technically blocked

No more broken tests.

github.com/Chachamaru127/claude-code-harness

#ClaudeCode #AIDev #QualityAssurance
```

**Attachment**: slide_E1 (3-layer defense diagram)
**Timing**: Weekdays 19:00-21:00
**Hook**: Draw empathy through problem statement

---

### E2: Technical Details

```
[Post text]

Are you using Claude Code's Hooks?

PreToolUse -> Validate before tool execution
PostToolUse -> Process after execution

Claude harness uses these to:
- Block test tampering
- Auto-run quality checks
- Auto-remind about skills

Details:
github.com/Chachamaru127/claude-code-harness

#ClaudeCode #Hooks #AIDev
```

**Attachment**: slide_E2 (Hooks flow diagram)
**Timing**: Weekdays 12:00-13:00
**Hook**: Stimulate technical curiosity

---

### E3: Before/After

```
[Post text]

Here's what Sonnet 4 generated as "implementation":

def slugify(text):
    answers = {"HelloWorld": "hello-world"}
    return answers.get(text, "")

Tests passed, but it's just hardcoded values.

Claude harness detects and blocks this kind of "hollow implementation" too.

#ClaudeCode #TestDrivenDevelopment
```

**Attachment**: slide_E3 (actual code example + block screen)
**Timing**: Weekdays 19:00-21:00
**Hook**: Draw recognition through real examples

---

## For Vibe Coders (V1-V3)

### V1: Empathy

```
[Post text]

"I have an app idea, but coding is..."

Don't worry. With Claude harness, you just talk in your language.

You: "I want to make a reservation management app"
Claude: "I'll create a plan. A screen for entering names and dates, sound good?"

AI writes all the code.

#VibeCoder #NoCoding #AIDev
```

**Attachment**: slide_V1 (conversation image)
**Timing**: Weekends 10:00-12:00
**Hook**: Lead with empathy

---

### V2: Concrete Examples

```
[Post text]

Apps you can build with zero programming knowledge:

- Reservation management app
- Expense tracker app
- ToDo list
- Portfolio site

With Claude harness, just say "I want to make something like this."

Plan -> Work -> Review
Three steps from idea to reality.

#VibeCoder #AIDev
```

**Attachment**: slide_V2 (app examples list)
**Timing**: Weekends 10:00-12:00
**Hook**: Expand imagination with concrete examples

---

### V3: Reassurance

```
[Post text]

"AI might generate weird code..."

Claude harness has quality guardrails:

Auto-stops suspicious code
Remembers previous conversations
AI helps when you're stuck

Safe even for beginners.

#VibeCoder #AIDev #BeginnersWelcome
```

**Attachment**: slide_V3 (guardrail diagram)
**Timing**: Weekdays 12:00-13:00
**Hook**: Proactively address concerns

---

## For Interested But Not Yet Using (N1-N3)

### N1: Lowering Barriers

```
[Post text]

Claude Code + harness, ready in 5 minutes.

1. Install Claude Code (2 min)
2. Add harness plugin (1 min)
3. /harness-init (2 min)

Only 3 commands to remember.

Easier than you think.

#ClaudeCode #AIDev
```

**Attachment**: slide_N1 (3-step diagram)
**Timing**: Weekdays 19:00-21:00
**Hook**: Lower barriers with numbers

---

### N2: Addressing Concerns

```
[Post text]

Curious about Claude Code, but...

"Seems hard" -> One command to install
"It's in English" -> Works fully in Japanese
"Expensive" -> From $20/month. Plugin is free
"Where to start" -> Just /harness-init

Take one step, and AI will guide you.

#ClaudeCode #AIDev
```

**Attachment**: slide_N2 (Q&A format diagram)
**Timing**: Weekdays 12:00-13:00
**Hook**: List and address concerns

---

### N3: Call to Action

```
[Post text]

Want to try Claude Code today?

Recommended first projects:
- Self-introduction page (30 min)
- Counter app (1 hour)
- ToDo list (2 hours)

Errors are okay.
Just say "got an error" and AI fixes it.

#ClaudeCode #StartToday
```

**Attachment**: slide_N3 (recommended projects list)
**Timing**: Weekends 10:00-12:00
**Hook**: Prompt specific action

---

## For Returning Users (R1-R3)

### R1: Empathy

```
[Post text]

"I tried Claude Code before, but it didn't work out..."

We understand. But v2.6 changed things significantly:

Before: Figure out what to do yourself
Now: Skills auto-guide you

Before: Explain from scratch every time
Now: Claude-mem remembers

Want to try again?

#ClaudeCode #TryAgain
```

**Attachment**: slide_R1 (Before/After comparison)
**Timing**: Weekdays 19:00-21:00
**Hook**: Acknowledge past experience before proposing

---

### R2: Improvements

```
[Post text]

Claude harness v2.6 improvements:

Skills Gate - Auto-launch skills
Claude-mem - Memory across sessions
session-init - AI tells you what to do
/harness-update - Update while preserving settings

If you gave up before, now's a good time.

#ClaudeCode #AIDev
```

**Attachment**: slide_R2 (improvements list)
**Timing**: Weekdays 12:00-13:00
**Hook**: List concrete improvements

---

### R3: Encouragement

```
[Post text]

"Maybe I'll fail at AI development again..."

This time is different.

Lost? -> /sync-status guides you
Error? -> troubleshoot skill auto-launches
Forgot? -> Claude-mem remembers

Use your past experience. This time it'll work.

#ClaudeCode #TryAgain
```

**Attachment**: slide_R3 (support features diagram)
**Timing**: Weekends 10:00-12:00
**Hook**: Acknowledge anxiety and encourage

---

## General (G1-G2)

### G1: Overview

```
[Post text]

What is Claude harness?

A plugin that autonomously operates Claude Code
in the Plan -> Work -> Review pattern.

- Auto-creates plans
- Auto-executes implementation
- Auto-checks quality
- Learns across sessions

If you're doing it all solo, this is it.

github.com/Chachamaru127/claude-code-harness

#ClaudeCode #AIDev
```

**Attachment**: slide_G1 (overall diagram)
**Timing**: Weekdays 19:00-21:00
**Hook**: Concisely convey the big picture

---

### G2: Value Proposition

```
[Post text]

With Claude harness:

Time: Planning done in 5 minutes
Quality: Test tampering auto-blocked
Learning: Past mistakes not repeated
Continuity: Context maintained across sessions

Solo development productivity transformed.

#ClaudeCode #Productivity
```

**Attachment**: slide_G2 (4 value points diagram)
**Timing**: Weekdays 12:00-13:00
**Hook**: List concrete benefits

---

## Example Posting Schedule

### Week 1

| Day | Post | Target |
|------|------|-----------|
| Mon | G1 | General (overview) |
| Tue | E1 | Engineers |
| Wed | V1 | Vibe coders |
| Thu | N1 | Interested newcomers |
| Fri | R1 | Returning users |
| Sat | V2 | Vibe coders |
| Sun | N3 | Interested newcomers |

### Week 2

| Day | Post | Target |
|------|------|-----------|
| Mon | G2 | General (value) |
| Tue | E2 | Engineers |
| Wed | V3 | Vibe coders |
| Thu | N2 | Interested newcomers |
| Fri | R2 | Returning users |
| Sat | E3 | Engineers |
| Sun | R3 | Returning users |

---

## Posting Guidelines

1. **Always attach an image** (use the corresponding slide)
2. **Consider deleting if no likes within 10 minutes**
3. **Paste URLs directly without shortening**
4. **Use 2-3 hashtags**
5. **Include profile keywords**
