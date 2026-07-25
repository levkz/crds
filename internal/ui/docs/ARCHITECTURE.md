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
         ┌───────┼───────────┐
         │       │           │
    ┌────▼───┐ ┌─▼──────┐ ┌──▼──────────┐
    │ Events │ │ Keymap │ │ Navigation  │
    │ (4     │ │(Global │ │ Manager     │
    │ types) │ │ bindgs)│ │(Push/Pop)   │
    └────────┘ └────────┘ └──┬───────────┘
         │       │           │
         └───────┼───────────┘
                 │
          ┌──────┼──────────┐
          │      │          │
          ▼      ▼          ▼
       Home    Quiz       Search
      Screen  Screen      Screen
        │      │            │
        ▼      ▼            ▼
     Components (display/ + interactive/)
      ───────────────────────────────
      Header   Footer   Card    Modal
      List     Table    Badge   Panel
      Progress Text     Label   Window
      StatusBar  Divider  Section  Group
      Paragraph  Notification  ConfirmDialog
      ErrorDialog
      ───────────────────────────────
      TextInput  SearchInput  Checkbox
      RadioGroup  Select  MultiSelect
      SelectableList  Tree  Spinner
```

The root model delegates navigation to the `navigation` package and
keyboard handling to the global `keymap`. Application config is loaded
from `~/.config/crds/` by the `config` package at startup.

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
m.Navigator.Replace(screen)  // flat navigation (no history, fires OnLeave/OnEnter)
m.Navigator.Push(screen)     // stacked navigation (preserves screen in history, OnEnter only)
m.Navigator.Pop()            // back navigation (restores from history, OnEnter only)
```

Screens do **not** construct or reference other screens.

Instead they emit navigation events:

```
Search
  ↓ Enter on result
NavigateToDetailMsg
  ↓
Root Model → m.Navigator.Push(DetailScreen)
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

Screens may also implement the optional `Lifecycle` interface for
enter/leave hooks:

```go
type Lifecycle interface {
    OnEnter() tea.Cmd
    OnLeave() tea.Cmd
}
```

Currently implemented by `SearchModel` (OnEnter resets to input mode,
OnLeave clears query/results/mode) and `DecksModel` (OnLeave saves
selection). Lifecycle hooks fire on flat transitions (Replace) but not
on stacked navigation (Push/Pop) where screens are preserved in history.

Screens can implement `BackHandler` to intercept Esc before the global
handler applies default behavior:

```go
type BackHandler interface {
    HandleBack() bool
}
```

Return `true` if the screen consumed the event. Currently implemented by
`SearchModel` (returns to input mode when in results mode).

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

Components live in two subpackages:

- **`display/`** — 20 stateless render functions (Header, Footer, Card, Progress,
  List, Table, Modal, Notification, Text, Label, Badge, Paragraph, Divider,
  Panel, Section, Group, Window, StatusBar, ConfirmDialog, ErrorDialog)
- **`interactive/`** — 9 stateful Bubble Tea sub-models (TextInput, SearchInput,
  Checkbox, RadioGroup, Select, MultiSelect, SelectableList, Tree, Spinner)

All 29 components are implemented.

Display components receive data and return strings. Interactive components
own ephemeral UI state (cursor, focus, selection) but not business data.

Components should avoid hidden state. Interactive components use optional
key config structs (`NavigationKeys`, `TextInputKeys`, `CheckboxKeys`) for
configurable vim-style keybindings.

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

The `theme` package provides a complete design system:

- **Palette**: 15 named colors (Blue, Green, Orange, Red, Gray, White, Background, Selection, Border, Link, Surface, Magenta, Purple, Cyan, Yellow)
- **Semantic styles**: 10 styles (Primary, Secondary, Accent, Success, Warning, Danger, Muted, Header, Background, Surface)
- **Typography**: 6 text roles (Title, Subtitle, Body, Caption, Emphasis, Key)
- **Borders**: 5 styles (Normal, Rounded, Double, Thick, None)
- **Icons**: 4 sources × 10 semantic slots (NerdFont → Emoji → Unicode → ASCII), auto-detected at startup
- **Spacing**: 7-tier scale (Xxs → Xxl)
- **Border roles**: 6 semantic roles (Container, Card, Modal, Emphasis, Section, None)
- **YAML loading**: Custom theme files with palette, icons, and typography overrides
- **Store**: Multi-theme registry with runtime switching via Settings screen (built-in: default, dark, light, tokyonight)
- **Background fill**: `fillBackground()` in `app/view.go` re-wraps each ANSI-reset-delimited segment with `Background(p.Background)` so the theme background covers the entire terminal

Components use semantic styles instead of hardcoded colors. They should never use terminal color codes directly.

---

# State Ownership

```
Registry (ScreenIndex → Screen)

  ├── HomeScreen         → HomeModel         (activity menu)
  ├── QuizScreen         → QuizModel         (flashcard quiz)
  ├── TypingQuizScreen   → TypingQuizModel   (typing-based quiz)
  ├── DecksScreen        → DecksModel        (multi-deck selection)
  ├── SearchScreen       → SearchModel       (text search)
  ├── StatisticsScreen   → StatisticsModel   (metrics)
  ├── SettingsScreen     → SettingsModel     (theme switch)
  └── DetailScreen       → DetailModel       (entry view)
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

Bubble Tea → tea.KeyMsg, tea.WindowSizeMsg

↓

Root Model → dispatchEvent()
│              ├─ events.TickMsg → re-arm TickCmd()
│              ├─ events.ThemeSwitchMsg → theme.Switch()
│              ├─ events.ShowNotificationMsg → show notification
│              ├─ events.HideNotificationMsg → hide notification
│              ├─ tea.WindowSizeMsg → resize handler
│              ├─ tea.KeyMsg → dispatchKeyEvent()
│              │                  ├─ keymap.DefaultGlobal → global action
│              │                  │   (Back checks BackHandler first)
│              │                  └─ forwardToScreen() → screen.Update()
│              │                                           │
│              │                                     screen handles locally
│              │                                     via keymap.Default*
│              │
│              └─ forwardToScreen() (pass-through)
│

Command

↓

Application

↓

Updated State

↓

Render
```

Screen events are defined in `internal/ui/screen.go`:
`NavigateToMsg`, `SaveAnswerMsg`, `TypeAnswerMsg`, `TypingGradeMsg`,
`DeckSelectionChangedMsg`, `NavigateToDetailMsg`, `DataLoadedMsg`,
`DataErrorMsg`, `SavedMsg`. Additional centralized event types live
in the `events/` package (4 types: `TickMsg`, `ThemeSwitchMsg`,
`ShowNotificationMsg`, `HideNotificationMsg`). The UI should
communicate through events rather than direct mutation.

Keypresses are matched against the centralized `keymap` package:
global keys in the root model, screen-local keys in each screen's
`Update()` method. This ensures a single source of truth for all
keybindings.

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
- **Theme** (`theme/`): Complete design system with 54 tests:
  - 11-color palette (Blue, Green, Orange, Red, Gray, White, Background, Selection, Border, Link, Surface)
  - 10 semantic styles (Primary, Secondary, Accent, Success, Warning, Danger, Muted, Header, Background, Surface)
  - Typography system (Title, Subtitle, Body, Caption, Emphasis, Key)
  - Border styles (Normal, Rounded, Double, Thick, None)
  - 10-slot icon set with 4 sources: NerdFont → Emoji → Unicode → ASCII
  - Environment auto-detection (CRDS_ICON_SOURCE, NerdFont, Emoji, Unicode)
  - Spacing scale (7 tiers), border roles (6 semantic roles)
  - YAML theme loading with 6 test fixtures + style/typography overrides
  - Theme store with runtime switching
- **Styles** (`styles/`): 12 style definitions with 60 tests (Header, Footer, SelectedItem, FocusedInput, Error, Warning, Success, Hint, MutedText, Card, Panel, Modal)
- **Components** (`components/`): 29 components across `display/` (20 stateless) and `interactive/` (9 stateful) — all implemented
- **Screens** (`screens/`): All 7 screens — Home (activity menu), Quiz (flashcard), TypingQuiz (typing-based), Search (two-phase: input + results), Statistics (metrics), Settings (theme switch), Detail (entry view)
- **Keymap** (`keymap/`): Centralized keybinding definitions with `Binding.Match()`, `BindingList.Help()`, per-screen keymap structs with `Footer()` methods, `Registry` with `Bindings()`/`FindBinding()`, `KeymapConfig` for user overrides, and 16 tests. All screens and the root model use `keymap.Default*` instead of hardcoded strings.
- **Config** (`internal/config/`): User configuration from `~/.config/crds/` — directory auto-creation, `config.yaml` loading, `keymaps.yaml` loading with `keymap.ApplyDefaultOverrides()`, theme discovery from `themes/*.yaml`. 13 tests.
- **Events** (`events/`): 4 centralized event types (`TickMsg`, `ThemeSwitchMsg`, `ShowNotificationMsg`, `HideNotificationMsg`)
- **Layout** (`layout/`): Layout helpers (Page, Column, Row, Center, Align, Stack, Grid, Spacer)
- **Renderer** (`renderer/`): Custom renderer utilities (Wrap, Truncate, AnsiWidth, VisibleWidth)

## Placeholder Directories

- **actions** (empty): Action types
- **animations** (empty): Animation utilities
- **debug** (empty): Debug utilities
- **testdata** (empty): Test fixtures

---

# Future Extensions

The architecture should support:

- Vim keybindings (partial — `"k"`/`"j"` already defined in `keymap.DefaultList`)
- Chord bindings (e.g. `g` then `g` for top of list)
- mouse mode
- plugin widgets
- split views
- audio
- inline images
- AI hints

These additions should not require redesigning existing screens.
