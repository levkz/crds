# UI Design

## Purpose

CRDS is a terminal-first flashcard application focused on efficient, distraction-free learning.

The interface should help users spend as little time navigating as possible and as much time studying as possible.

The UI should remain simple enough to disappear into the background while still providing enough structure to make long study sessions enjoyable.

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

---

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

---

## Keyboard First

Every feature should be accessible without a mouse.

Navigation should be efficient enough that experienced users rarely remove their hands from the keyboard.

Every action should have a discoverable keyboard shortcut.

---

## Progressive Disclosure

Only reveal information when it becomes useful.

Example:

During a quiz, display only the front of the card.

After revealing the answer, display:

- translations
- notes
- grading options

Do not overwhelm users with unnecessary information.

---

## Consistency

Similar actions should behave similarly throughout the application.

Examples:

- Esc always goes back or closes the current overlay.
- Enter confirms the primary action.
- Arrow keys navigate lists.
- Help is always available through `?`.

Users should not have to relearn controls between screens.

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

---

## Quiz

The primary experience.

The quiz should minimize distractions and keep the learner focused on the current card.

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

Only one card should receive attention at a time.

---

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

---

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

---

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

---

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

# Keyboard

Global shortcuts:

```text
Ctrl+C  Quit
Esc     Back
?       Help
```

Quiz:

```text
Enter   Reveal

1       Again
2       Hard
3       Good
4       Easy

q       Quit Session
```

Search:

```text
↑ ↓     Navigate
Enter   Open Entry
Esc     Back
/       Focus Search
```

Individual screens may define additional shortcuts when appropriate.

---

# Feedback

The interface should provide immediate feedback for user actions.

Examples:

- grading a card
- completing an import
- exporting a deck
- saving settings

Feedback should be brief and unobtrusive.

---

# Accessibility

The application should remain usable in a wide variety of terminal environments.

Support should include:

- monochrome terminals
- reduced motion
- terminals without emoji support
- high-contrast themes where possible

Color should reinforce meaning, never replace it.

---

# Future Features

The design should accommodate future additions without changing existing workflows.

Potential additions include:

- Vim keybindings
- mouse support
- pronunciation playback
- inline images
- split-screen dictionary
- configurable themes
- plugins
- AI-generated hints
- progress graphs

Future features should integrate naturally with the existing interaction model.

---

# Guiding Principle

The interface exists to support learning, not to draw attention to itself.

Every design decision should make studying feel faster, simpler, and more enjoyable.
