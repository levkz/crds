# UI Style Guide

This document defines the visual language of the terminal UI.

Its purpose is consistency rather than creativity.

Every screen should look like it belongs to the same application.

---

# Design Principles

The UI should feel:

- calm
- focused
- lightweight
- predictable

The application is a learning tool.

The interface should never compete with the content.

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
