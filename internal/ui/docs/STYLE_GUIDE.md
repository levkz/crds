# UI Style Guide

> Package guide: the visual language and interaction design of the terminal
> UI. Status and plans live in `docs/status.md` and `docs/roadmap.md` (see
> `docs/README.md`).

This document defines the visual language of the terminal UI.

Its purpose is consistency rather than creativity.

Every screen should look like it belongs to the same application.

---

# Design Goals

The interface should be:

- Keyboard-first
- Fast
- Responsive
- Accessible
- Predictable
- Pleasant to use
- Easy to learn
- Easy to extend

The application should feel lightweight regardless of the size of the user's collection.

The UI should feel:

- calm
- focused
- lightweight
- predictable

The application is a learning tool.

The interface should never compete with the content.

---

# Design Principles

## One Screen, One Purpose

Every screen exists to perform a single task.

Examples:

- Home
- Quiz
- Search
- Statistics
- Settings

Avoid combining unrelated functionality into the same screen.

## Minimize Visual Noise

Only display information relevant to the current task.

Prefer:

- whitespace
- indentation
- typography
- subtle color

Avoid:

- decorative borders
- unnecessary tables
- large banners
- excessive icons
- excessive colors

## Keyboard First

Every feature should be accessible without a mouse.

Navigation should be efficient enough that experienced users rarely remove their hands from the keyboard.

Every action should have a discoverable keyboard shortcut.

The footer documents the shortcuts for the current screen.

## Progressive Disclosure

Only reveal information when it becomes useful.

Example:

During a quiz, display only the front of the card.

After revealing the answer, display:

- translations
- notes
- grading options

Do not overwhelm users with unnecessary information.

## Consistency

Similar actions should behave similarly throughout the application.

Examples:

- Esc always goes back or closes the current overlay.
- Enter confirms the primary action.
- Arrow keys navigate lists.
- Help is always available through `?`.

Users should not have to relearn controls between screens.

---

# Visual Hierarchy

Visual hierarchy should come from:

- whitespace
- alignment
- typography
- semantic colors

Avoid using decorative elements.

Good:

```
French A1

bonjour

Press Enter to reveal.
```

Avoid:

```
==============================
===== FRENCH A1 ============
==============================
```

---

# Whitespace

Whitespace is preferred over borders.

Use blank lines to separate sections.

Example:

```
Header

Content

Content

Footer
```

Avoid stacking elements tightly together.

---

# Borders

Borders should be rare.

Use borders only when they communicate structure.

Examples:

- modal dialogs
- temporary overlays
- confirmation windows

Avoid surrounding every component with a box.

Bad:

```
+------------------+

Card

+------------------+
```

Good:

```
bonjour

hello
```

---

# Alignment

Headers should align to the left.

Lists should align to the left.

Statistics should be visually grouped.

Avoid centered text except:

- splash screen
- loading screen
- empty states

---

# Typography

Headers

- bold

Body

- normal

Muted information

- muted color

Keyboard shortcuts

- muted color

Never rely on ALL CAPS.

---

# Colors

Colors communicate meaning.

Never decorate.

Semantic colors:

Primary

Current selection

Success

Correct answer

Warning

Needs attention

Danger

Incorrect answer

Muted

Secondary information

Background

Terminal background

Do not invent additional colors without updating the theme.

---

# Selection

Selection should be obvious.

Example:

```
> Quiz

  Search

  Statistics
```

Only one primary selection should exist.

Avoid multiple highlighted areas.

---

# Keyboard Shortcuts

The footer always displays available shortcuts.

Example:

```
Enter Reveal

Esc Back

? Help
```

Use concise wording.

Prefer:

```
Esc Back
```

instead of

```
Press Escape to return to the previous screen.
```

Global shortcuts:

```
Ctrl+C  Quit
Esc     Back
?       Help
```

Quiz:

```
Enter   Reveal

1       Again
2       Hard
3       Good
4       Easy

q       Quit Session
```

Search:

```
↑ ↓     Navigate
Enter   Open Entry
Esc     Back
/       Focus Search
```

Individual screens may define additional shortcuts when appropriate.

---

# Lists

Lists should be simple.

Example:

```
> French

  Spanish

  German
```

Avoid:

- numbering unless meaningful
- decorative bullets
- excessive indentation

---

# Cards

Cards are the primary UI element.

Before reveal:

```
bonjour
```

After reveal:

```
bonjour

hello

good morning

Common greeting.
```

Cards should breathe.

Avoid compressing information.

---

# Statistics

Statistics should emphasize values.

Example:

```
Reviewed Today

52

Accuracy

91%
```

Avoid tables unless comparing many values.

---

# Progress

Progress bars should remain subtle.

Example:

```
██████████░░░░░░░░
```

Progress should never dominate the screen.

---

# Notifications

Notifications appear briefly.

Success

```
✓ Card saved
```

Warning

```
! Import completed with warnings
```

Error

```
✗ Database unavailable
```

Notifications should disappear automatically.

Feedback for user actions (grading a card, completing an import, saving
settings) should be brief and unobtrusive.

---

# Modals

Modals interrupt the workflow.

Use them sparingly.

Good:

```
Delete deck?

Enter Confirm

Esc Cancel
```

Avoid placing unrelated information inside modals.

---

# Empty States

Always explain what the user can do.

Bad:

```
No cards.
```

Better:

```
No cards due.

Come back later or study another deck.
```

---

# Loading States

Prefer lightweight indicators.

Example:

```
Loading…

██████░░░░░░
```

Avoid spinners that dominate attention.

---

# Animation

Animations should be subtle.

Allowed:

- progress updates
- notification fade
- loading indicator

Avoid:

- screen transitions
- bouncing elements
- animated menus

Respect reduced-motion preferences.

---

# Accessibility

The UI must remain usable:

- without color
- without emoji
- in monochrome terminals
- with large fonts

Every visual cue should have a textual equivalent.

Example:

Good:

```
✓ Correct
```

Better:

```
Correct
```

Color should reinforce information, not provide it.

---

# Screen Layout

Every screen follows the same structure.

```
Header

Content

Footer
```

Header

- title
- optional subtitle

Content

- screen-specific information

Footer

- available shortcuts

---

# Screens

## Home

The entry point of the application.

Responsibilities:

- start studying
- search vocabulary
- view statistics
- open settings
- quit the application

Example:

```text
CRDS

Choose an action

> Quiz

  Search

  Statistics

  Settings

  Quit
```

## Quiz

The primary experience.

The quiz should minimize distractions and keep the learner focused on the current card.

Only one card should receive attention at a time.

Before revealing:

```text
French A1

bonjour

Press Enter to reveal.
```

After revealing:

```text
French A1

bonjour

hello

good morning

Common greeting.

1 Again
2 Hard
3 Good
4 Easy
```

## Search

Allows users to quickly find vocabulary.

Searching should update results immediately as the query changes.

Example:

```text
Search

> bonj

bonjour

bonne nuit

bonsoir
```

Selecting an entry opens its detail page.

## Entry Detail

Displays complete information about a vocabulary entry.

Possible sections include:

- translations
- examples
- notes
- tags
- pronunciation
- related entries

Example:

```text
bonjour

Translations

• hello

• good morning

Examples

Bonjour Marie.

Hello Marie.

Tags

greeting

A1
```

## Statistics

Provides insight into learning progress.

Typical information includes:

- reviewed today
- cards due
- review accuracy
- streak
- retention

Example:

```text
French A1

Reviewed Today

52

Accuracy

91%

Due Today

18

Current Streak

12
```

Statistics should prioritize clarity over density.

## Settings

Provides access to application configuration.

Possible sections:

- themes
- scheduler
- language
- database
- import
- export
- animations
- accessibility

Settings should be organized into logical groups.

---

# Navigation

The application should behave predictably.

Users should always know:

- where they are
- how to go back
- how to quit
- how to access help

Navigation should avoid unnecessary intermediate screens.

---

# Spacing Rules

Between sections:

1 blank line

Between title and content:

1 blank line

Between content and footer:

1 blank line

Do not insert multiple consecutive blank lines unless centering content vertically.

---

# Consistency

When adding a new component, ask:

- Does it match existing spacing?
- Does it use semantic colors?
- Does it introduce unnecessary decoration?
- Does it support keyboard navigation?
- Can it work without color?
- Would it still look correct in a monochrome terminal?

If the answer to any of these is "no", reconsider the design before implementing it.

---

# Guiding Principle

The best UI is one the user stops noticing.

The learner's attention should remain on the vocabulary, not the interface.

The interface exists to support learning, not to draw attention to itself.

Every design decision should make studying feel faster, simpler, and more enjoyable.
