# UI Context

This document provides implementation context for future development sessions.

---

# Philosophy

The UI is intentionally conservative.

The goal is not to impress with visuals.

The goal is to disappear and let the user focus on learning.

Whitespace is preferred over borders.

Typography is preferred over decoration.

Colors communicate meaning.

---

# Libraries

Primary library:

- Bubble Tea

Supporting libraries:

- Lip Gloss

Do not reinvent functionality already provided by Bubble Tea unless there is a
strong reason.

---

# Navigation

Navigation is centralized in the `navigation` package.

A `Manager` holds the current screen index, a history stack,
a forward stack, and a modal stack. It exposes `Push`, `Pop`,
`Replace`, `Forward`, `PushModal`, `DismissModal`, `ShowOverlay`,
`HideOverlay`, and `Reset`.

A `Registry` maps `ScreenIndex` → `Screen` instances, decoupling
screen identity from concrete types. The root model delegates to
`Manager` and retrieves the active screen via `CurrentScreen()`.

Screens should emit events instead of changing the active screen directly.

---

# Components

Components should remain:

- small
- reusable
- stateless whenever practical

Prefer functions over structs unless state is required.

Good:

```
Header(title)
Footer(keys)
Progress(value)
```

Avoid components that know about application models.

---

# Rendering

Rendering should be declarative.

Prefer:

```
Page(
    Header(...),
    Content(...),
    Footer(...),
)
```

instead of manually concatenating strings throughout the codebase.

---

# Styling

Never hardcode colors.

Always use semantic theme values.

Example:

Theme.Success

instead of

```
lipgloss.Color("42")
```

inside widgets.

---

# Keyboard

Everything should be usable without a mouse.

Every visible action should have a keyboard shortcut.

The footer documents the shortcuts for the current screen.

---

# Screen Responsibilities

Home

Choose an activity.

Quiz

Display cards.

Reveal answers.

Collect grades.

Search

Search entries.

Statistics

Display study metrics.

Settings

Configure the application.

Detail

Display a vocabulary entry.

---

# Root Responsibilities

The root model owns:

- screen selection
- overlays
- notifications
- global shortcuts
- terminal resizing

---

# Components

Current implemented components (8):

Header
Footer
Card
Progress
List
Modal
Notification
Text

Future components may include:

Table
Tabs
Input
Viewport
Tree

---

# File Organization

```
ui/
├── screen.go           Screen interface + ScreenIndex type
├── theme.go            Semantic color theme (re-exports theme.Default)

├── app/                Root model (Bubble Tea entry point)
│   ├── app.go          New() + Run()
│   ├── model.go        Root Model struct, GlobalState, messages
│   ├── events.go       dispatchEvent + dispatchKeyEvent + forwardToScreen
│   ├── view.go         Root View()
│   ├── update.go       Root Update()
│   ├── lifecycle.go    Lifecycle hooks, transitionTo, popToPrevious
│   ├── commands.go     NavigateToMsg, Dispatcher, config update
│   ├── config.go       Config + DefaultConfig
│   ├── dependencies.go DeckProvider, ProgressRecorder interfaces
│   └── tick.go         Tick loop

├── navigation/         Centralized navigation
│   ├── manager.go      Manager (Push, Pop, Replace, Forward, Reset, ...)
│   ├── stack.go        History stack with depth limit
│   ├── registry.go     Registry (ScreenIndex → Screen)
│   ├── events.go       9 event types (Push, Pop, Replace, Forward, ...)
│   └── tests/          82 black-box tests

├── screens/            Screen implementations
│   ├── home.go         HomeModel — activity menu
│   ├── quiz.go         QuizModel — flashcard quiz
│   ├── search.go       SearchModel — text input + results display
│   ├── statistics.go   StatisticsModel — study metrics
│   ├── settings.go     SettingsModel — theme switching
│   └── detail.go       DetailModel — entry detail view

├── components/         Reusable components
│   ├── header.go       Header component
│   ├── footer.go       Footer component
│   ├── card.go         Card component
│   ├── progress.go     Progress bar component
│   ├── list.go         List component (selected-item aware)
│   ├── modal.go        Modal/dialog component
│   ├── notification.go Notification component
│   └── text.go         Text component

├── styles/             Semantic style definitions
│   ├── header.go       Header(width int)
│   ├── footer.go       Footer(width int)
│   ├── selected_item.go SelectedItem()
│   ├── focused_input.go FocusedInput()
│   ├── error.go        Error()
│   ├── warning.go      Warning()
│   ├── success.go      Success()
│   ├── hint.go         Hint()
│   ├── muted_text.go   MutedText()
│   ├── card.go         Card(width int)
│   ├── panel.go        Panel(width int)
│   ├── modal.go        Modal(width, height int)
│   └── styles_test.go  60 tests

├── theme/              Design system (colors, typography, icons, borders)
│   ├── palette.go      11-color Palette + DefaultPalette
│   ├── theme.go        Theme struct + NewTheme() + BorderFor()
│   ├── typography.go   6 text role styles
│   ├── borders.go      5 border styles
│   ├── icons.go        4 icon sets × 10 semantic slots
│   ├── spacing.go      7-tier Spacing scale
│   ├── border_role.go  BorderRole enum
│   ├── detect.go       Icon source auto-detection
│   ├── nerdfont.go     NerdFont detection
│   ├── config.go       YAML loading + style overrides
│   ├── presets.go      Dark/light presets
│   ├── store.go        Theme registry + switching
│   ├── DESIGN.md       Design language documentation
│   ├── theme_test.go   54 tests
│   └── testdata/       6 YAML fixtures

├── layout/             Layout helpers (empty)
├── events/             Event types (empty)
├── keymap/             Keybinding types (empty)
├── actions/            Action types (empty)
├── debug/              Debug utilities (empty)
├── animations/         Animation helpers (empty)
├── renderer/           Custom renderer (empty)
└── testdata/           Test fixtures (empty)
```

Keep responsibilities separate.

Avoid large files.

Most implementation work happens in `app/`, `navigation/`, and `screens/`.

---

# Code Style

Prefer:

small methods

small structs

small components

Avoid deeply nested switch statements.

Prefer early returns.

Keep rendering code readable.

---

# Performance

Rendering should avoid unnecessary allocations.

Reuse strings.Builder where appropriate.

Do not prematurely optimize.

Bubble Tea already performs efficient redraws.

---

# Testing

Business logic should be testable without Bubble Tea.

Rendering helpers should be deterministic.

Navigation should be testable through emitted events.

---

# Current Status

## Complete

- **Theme** (`theme/`): Full implementation with 54 tests, including 11-color palette, 10 semantic styles, typography, borders, 10-slot icons (4 icon sources), spacing scale, border roles, YAML loading with style overrides, and theme switching
- **Styles** (`styles/`): All 12 style definitions implemented with 60 tests — Header, Footer, SelectedItem, FocusedInput, Error, Warning, Success, Hint, MutedText, Card, Panel, Modal
- **Components** (`components/`): All 8 components implemented — Header, Footer, Card, Progress, List, Modal, Notification, Text
- **Screens** (`screens/`): All 6 screens implemented — Home (activity menu), Quiz (flashcard), Search (text input + results), Statistics (metrics display), Settings (theme switching), Detail (entry view)
- **Navigation** (`navigation/`): Complete with 82 black-box tests including Manager, stack, and registry
- **App** (`app/`): Root Bubble Tea model, events, lifecycle, commands, config, theme loading, overlay/notification system

## Placeholder Directories

- **actions** (empty): Action types
- **keymap** (empty): Keyboard bindings
- **events** (empty): Event types
- **layout** (empty): Layout helpers
- **animations** (empty): Animation utilities
- **debug** (empty): Debug utilities
- **renderer** (empty): Custom renderer
- **testdata** (empty): Test fixtures

---

# Known Issues

- QuizModel.Cards is pre-populated — no data loading wired yet
- Quiz progress always 0, grading returns no-op
- Search currently shows placeholder results (not wired to real data)
- Statistics shows zeroes for all metrics (no scheduler wired)
- Detail shows "Select an entry" until data is passed
- Screens use hardcoded width 60 — no terminal resize awareness yet

---

# Future Work

Likely future additions:

- split layouts
- search highlighting
- progress graphs
- animations
- configurable themes
- mouse support
- plugins
- command palette

The current architecture should support these without major refactoring.
