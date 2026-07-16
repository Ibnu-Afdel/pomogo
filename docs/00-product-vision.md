# PomoGo Product Vision

PomoGo is a beautiful, lightweight terminal deep-work companion for developers.

It is not a Pomodoro app with decorations. Pomodoro is the internal timing engine; the
product is the feeling of being in a calm, intentional focus block that fits naturally into a
developer's terminal setup.

## Product Frame

- **Quick Focus:** a classic 25 minute focus session where the 5 minute break continues as part
  of the cycle. Use this when the user wants to start working immediately.
- **Deep Focus:** a 1-4 hour block where PomoGo classifies the time behind the scenes into work
  and break segments. The user sees the whole block countdown and does not need to think about
  individual Pomodoro intervals.
- **Aesthetic goal:** screenshot-worthy, terminal-native, and restrained. Add more themes and
  layouts, but never turn the screen into junk.

## Original Notes

I actually think you're aiming at something more interesting than "another Pomodoro app."

The screenshots you shared tell me exactly what attracts people:

* The timer is **beautiful enough to screenshot.**
* It communicates **"I'm deep in work"** without being noisy.
* It becomes part of the developer aesthetic (terminal rice, Omarchy, Neovim, Ghostty, etc.).

If you make **a TUI that people actually want to leave open all day**, you've got something.

---

# The problem with most Pomodoro apps

Most of them optimize for productivity.

Developers often optimize for:

* flow
* aesthetics
* habit
* sharing screenshots
* seeing progress

Nobody wants to stare at

```
25:00

Press S to start.
```

for 8 hours.

Instead they want something like

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

        PROJECT
      Fawz OS

      TASK
Implement AI Context

      18:32

██████████░░░░░░░░░░░

Session 2 / 8

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

Minimal.
Beautiful.
Screenshot worthy.

---

# The "4 Hour Focus" trend

The screenshots you posted are basically this.

People don't actually sit 4 hours continuously.

Instead they do

```
4 hour block

↓

Pomodoro internally

25 work
5 break

repeat

↓

Display

3h 12m remaining
```

Exactly.

The user never thinks about pomodoros.

They think about

> "I'm in my Deep Work block."

Internally

```
240 minutes

↓

25
5
25
5
25
5
...
```

This is much more psychologically appealing.

---

# I'd make two modes

## Quick Focus

```
25 min
```

Perfect when you just want to work.

---

## Deep Focus

Choose

```
1 hour

2 hour

3 hour

4 hour
```

Internally

```
25/5

or

50/10
```

User never sees the complexity.

They only see

```
Deep Focus

01:53:21 remaining
```

---

# Random beautiful themes

This is genius.

Every launch

Randomly pick

```
Nord

Tokyo Night

Catppuccin

Gruvbox

Dracula

Everforest

Rose Pine

Kanagawa

Carbon

Night Owl
```

BubbleTea makes this easy.

People LOVE randomness.

---

Even better

```
Theme of the day

July 15

Rose Pine
```

Tomorrow

```
Gruvbox
```

---

# Different layouts

Instead of only colors

Random layout too.

Example

Classic

```
PROJECT

Fawz

TASK

Build Auth

18:22

███████░░░░░░░░
```

---

Minimal

```
18:22

Build Auth

███████████
```

---

Centered

```
          18:22

     Deep Work

     Fawz

██████████████████
```

---

Retro

```
╔════════════════════╗

   DEEP WORK

01:22:34

██████▒▒▒▒▒▒▒

╚════════════════════╝
```

---

ASCII

```
██████

██████

01:12

██████
```

Imagine screenshots.

---

# Tiny animated backgrounds

Don't make it distracting.

Instead

```
.
..
...
```

stars

```
✦

✧

✦
```

Rain

```
│

│

╲
```

Snow

```
•

•

•
```

Terminal particles.

Almost free CPU.

---

# Session titles

Instead of

```
Focus
```

Display

```
Building

Learning

Writing

Reading

Debugging

Designing

Researching
```

Chosen automatically or manually.

Looks much nicer.

---

# Project emojis/icons

```
🐹 Go

🦀 Rust

🐍 Python

🐘 PHP

⚛ React

🐳 Docker

☁ Cloud

📚 Reading
```

Tiny detail.

Huge personality.

---

# Tiny stats

Instead of analytics dashboards

Just

```
Today

3h 42m

███████
```

---

Week

```
Mon ████

Tue ███████

Wed ██

Thu █████

Fri ███
```

ASCII.

---

Lifetime

```
482h

917 sessions

48 day streak
```

Enough.

---

# Focus level

Very cool.

```
Deep

■■■■■■■

7/10
```

Based on

* pauses
* breaks skipped
* distractions

No AI.

Simple heuristics.

---

# Git integration

Optional.

When inside Git repo

Automatically

```
Project

Fawz

Branch

feature/timer

Commit

a91df2
```

Very developer-ish.

---

# Terminal presence

If tmux

```
Session

backend
```

Current pane title.

---

# Spotify integration (optional)

Display

```
♪

Hans Zimmer
```

Very low priority.

---

# Screenshot mode

This is the killer feature.

Press

```
S
```

Automatically

* hide controls
* center timer
* show clean layout

Example

```
━━━━━━━━━━━━━━━━━━━━━━

Deep Work

03:12:18

Building Fawz

━━━━━━━━━━━━━━━━━━━━━━
```

Nothing else.

Perfect for screenshots.

---

# Session recap

When finished

```
Deep Work Complete

4h

8 focus sessions

7 breaks

0 pauses

Amazing.
```

Simple dopamine.

---

# Wallpapers

Since it's TUI

You can't.

But you CAN fake atmosphere.

Example

```
☾

Stars

Clouds

Mountains

Forest

Rain

Ocean
```

ASCII decorations.

---

# Resource usage

Keep it tiny.

Avoid

* telemetry
* internet
* Electron
* heavy animations

Target

```
<15 MB RAM

<1% CPU
```

People on Arch will appreciate that.

---

# Open source attraction

What gets GitHub stars isn't more features.

It's identity.

People should immediately recognize a screenshot.

Like:

```
        ╭──────────────────────────╮

             DEEP WORK

             03:18:22

        Building Fawz

█████████████░░░░░░░░░

      Session 6 of 8

╰──────────────────────────╯
```

One glance and someone says:

> "What terminal app is that?"

---

# What I would NOT add

Avoid turning it into a task manager or productivity suite.

Skip:

* ❌ Kanban boards
* ❌ Calendar sync
* ❌ Notes editor
* ❌ AI coach
* ❌ Team collaboration
* ❌ Cloud accounts
* ❌ User authentication
* ❌ Complex graphs

Keep it focused on one promise:

> **The most beautiful, lightweight terminal companion for deep work.**

---

## A roadmap I'd genuinely build

### v0.1 — Core

* Deep Focus mode (1–4 hours with Pomodoro internally)
* Classic 25/5 mode
* SQLite persistence
* Projects & tasks
* Daily/weekly stats
* Configurable durations

### v0.2 — Personality

* 10+ built-in themes
* 4–6 random layouts
* Screenshot mode
* Subtle ASCII ambient effects (rain, stars, snow)
* Theme/layout rotation on launch

### v0.3 — Developer features

* Git repository detection
* Branch display
* tmux integration
* Streaks and lifetime statistics
* Session recap

### v1.0 — Polish

* Plugin/theme system
* Export session history (JSON/CSV)
* AUR package, Homebrew formula, Scoop
* Beautiful README with GIFs and screenshots

One last thought: I wouldn't market it as a Pomodoro app at all.

I'd market it as:

> **A beautiful terminal deep-work companion for developers.**

That framing matches what developers are already sharing on X and GitHub—less "productivity timer," more "this is part of my coding setup." Given you're already using Go + Bubble Tea, that niche is surprisingly underserved, and it aligns well with the lightweight, polished aesthetic you're after.
