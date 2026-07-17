# UI Architecture

## Purpose

The `ui` package is responsible for the entire terminal user interface.

It is intentionally isolated from business logic.

The UI should:

- display application state
- collect user input
- emit commands/events
- never directly manipulate persistence

The scheduler, repositories, parser, importer and database remain outside this
package.

---

# Design Goals

The UI should be:

- keyboard first
- responsive
- minimal
- accessible
- composable
- easy to extend

Every screen should have one responsibility.

---

# Architecture

The UI follows Bubble Tea's Elm-style architecture.

```
            Bubble Tea

               │

               ▼

            tea.Msg

               │

               ▼

            Root Model

               │

               ▼

      Navigation Manager
      (Push/Pop/Replace)

               │

               ▼

      Registry → Screen

               │

      ┌─────────┼──────────┐
      │         │          │
      ▼         ▼          ▼

    Home      Quiz      Search
    Screen     Screen     Screen

      │         │          │

      ▼         ▼          ▼

            Components

     Header
     Footer
     Card
     Progress
     List
     Modal
     Notification
     Text
```

The root model delegates navigation to the `navigation` package.

Each screen owns only its own state.

---

# Navigation

Navigation is centralized in the `navigation` package.

The active screen is represented by a `ScreenIndex` and a `Manager`
holds the current index, a history stack, modal stack, and optional
forward stack. It exposes `Push`, `Pop`, `Replace`, `PushModal`,
`DismissModal`, `ShowOverlay`, `HideOverlay`, and `Reset`.

A `Registry` maps `ScreenIndex` → `Screen` instances, decoupling
screen identity from their concrete types.

The root model delegates to `Manager`:

```go
m.Navigator.Replace(screen)
m.Navigator.Push(screen)
m.Navigator.Pop()
```

Screens do **not** construct or reference other screens.

Instead they emit navigation events:

```
Quiz
  ↓ Esc
Navigate(Home)
  ↓
Root Model → m.Navigator.Pop()
```

This prevents coupling between screens.

---

# Screen Lifecycle

Every screen implements:

```go
type Screen interface {
    Init() tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)  // returns updated self
    View() string
}
```

A screen should never modify global application state.

---

# Responsibilities

Root model:

- navigation
- window resizing
- global shortcuts
- overlays
- event dispatching

Screen:

- local state
- keyboard handling
- rendering

Component:

- reusable rendering
- no business logic

---

# Components

Components should be reusable.

Current implemented components:

- Header
- Footer
- Card
- Progress

Planned but not yet implemented:

- List
- Modal
- Notification
- Text

Future components may include:

- Table
- Tabs
- Input
- Viewport
- Tree

Components should receive data and return strings.

Avoid hidden state whenever possible.

---

# Layout

Every screen follows the same structure.

```
Header

Content

Footer
```

The footer always displays currently available shortcuts.

---

# Styling

Styling is centralized.

The `theme` package provides complete semantic styling:

- Theme has 6 semantic styles: Primary, Success, Warning, Danger, Muted, Header
- Pallette with 10 colors: Blue, Green, Orange, Red, Gray, White, Background, Selection, Border, Link
- Typography: Title, Subtitle, Body, Caption, Emphasis, Key
- Borders: Normal, Rounded, Double, Thick, None
- Icons: NerdFont support (primary), Emoji fallback, Unicode fallback, ASCII fallback

Components use semantic styles instead of hardcoded colors.

Example:

Primary

Success

Warning

Danger

Muted

Background

Components should never use terminal color codes directly.

---

# State Ownership

```
Registry (ScreenIndex → Screen)

  ├── HomeScreen    → HomeModel (stubbed)
  ├── QuizScreen    → QuizModel (implemented)
  ├── SearchScreen  → SearchModel (stubbed)
  ├── StatisticsScreen → StatisticsModel (stubbed)
  ├── SettingsScreen   → SettingsModel (stubbed)
  └── DetailScreen     → DetailModel (stubbed)
```

Screens are stored in a `navigation.Registry` and managed by the
`navigation.Manager`. The root model references the current screen
through `m.Navigator.CurrentScreen()` rather than holding concrete fields.

Every screen owns only its own fields.

Example:

Quiz owns:

- current card
- reveal state
- progress
- selection

Quiz does NOT own:

- scheduler
- database
- deck storage

---

# Event Flow

```
Keyboard

↓

Bubble Tea

↓

Screen

↓

Command

↓

Application

↓

Updated State

↓

Render
```

The UI should communicate through events rather than direct mutation.

---

# Accessibility

Support:

- monochrome terminals
- no emoji (fallback available)
- reduced motion (optional animations)
- screen readers where possible

Animations should remain optional.

---

# Current Implementation Status

## Complete

- **Navigation** (`navigation/`): Full Manager, stack, Registry with 82 black-box tests
- **Theme** (`theme/`): Complete design system with 82 tests:
  - 10-color palette
  - 6 semantic styles (Primary, Success, Warning, Danger, Muted, Header)
  - Typography system (Title, Subtitle, Body, Caption, Emphasis, Key)
  - Border styles (Normal, Rounded, Double, Thick, None)
  - Icon source priority: NerdFont → Emoji → Unicode → Fallback
  - Environment auto-detection (CRDS_ICON_SOURCE, NerdFont, Emoji, Unicode)
  - YAML theme loading with 5 test fixtures
  - Theme store and switching
- **Components**: Header, Footer, Card, Progress - all functional
- **Screens**: Quiz - full implementation

## In Progress

- **Styles** (`styles/`): Empty directory awaiting style definitions for all 14 shared Lip Gloss styles
- **Missing Components**: List, Modal, Notification, Text (functions needed)
- **Empty Screens**: Home, Search, Statistics, Settings, Detail (stubbed)

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

# Future Extensions

The architecture should support:

- Vim keybindings
- mouse mode
- themes (complete and functional)
- plugin widgets
- split views
- audio
- inline images
- AI hints

These additions should not require redesigning existing screens.
