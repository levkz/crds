# UI Architecture

> Package guide: how the `ui` package is structured and wired. Status and
> plans live in `docs/status.md` and `docs/roadmap.md` (see `docs/README.md`).

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

## Philosophy

The UI is intentionally conservative.

The goal is not to impress with visuals.

The goal is to disappear and let the user focus on learning.

Whitespace is preferred over borders.

Typography is preferred over decoration.

Colors communicate meaning.

## Libraries

Primary library:

- Bubble Tea

Supporting libraries:

- Lip Gloss

Do not reinvent functionality already provided by Bubble Tea unless there is a
strong reason.

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
OnLeave clears query/results/mode). `DeckSelectModel` uses `BackHandler`
instead of lifecycle. Lifecycle hooks fire on flat transitions (Replace)
but not on stacked navigation (Push/Pop) where screens are preserved.

Screens can implement `BackHandler` to intercept Esc before the global
handler applies default behavior:

```go
type BackHandler interface {
    HandleBack() bool
}
```

Return `true` if the screen consumed the event. Currently implemented by
`SearchModel` (returns to input mode when in results mode) and
`DeckSelectModel` (clears search query if non-empty, deactivates search
if empty query).

A screen should never modify global application state.

---

# Responsibilities

Root model:

- navigation
- window resizing
- global shortcuts
- overlays
- event dispatching
- the canonical `AppState` snapshot (`Model.State`) and its propagation to
  screens via the `ui.StateSyncer` protocol

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

Screen views are composed with `layout.Page()` and `layout.Column()`
instead of manual string concatenation.

---

# Styling

Styling is centralized.

The `theme` package provides a complete design system:

- **Palette**: 15 named colors + 3 semantic overrides (Primary, Secondary, Accent)
- **Semantic styles**: 14 styles (Primary, Secondary, Accent, Success, Warning, Danger, Muted, Header, Background, Surface, PrimaryBg, SuccessBg, ErrorBg, WarningBg)
- **Typography**: 6 text roles (Title, Subtitle, Body, Caption, Emphasis, Key)
- **Borders**: 5 styles (Normal, Rounded, Double, Thick, None)
- **Icons**: 4 sources × 10 semantic slots (NerdFont → Emoji → Unicode → ASCII), auto-detected at startup
- **Spacing**: 7-tier scale (Xxs → Xxl)
- **Border roles**: 6 semantic roles (Container, Card, Modal, Emphasis, Section, None)
- **YAML loading**: Custom theme files with palette, icons, and typography overrides
- **Store**: Multi-theme registry with runtime switching via Settings screen (built-in: default, dark, light, tokyonight)

Components use semantic styles instead of hardcoded colors. They should never use terminal color codes directly.

---

# Rendering

Rendering should be declarative.

Prefer:

```
Page(
    Header(...),
    Content(...),
    Footer(...),
    height,
)
```

instead of manually concatenating strings throughout the codebase.

`fillBackground()` in `app/view.go` wraps each ANSI-reset-delimited segment
with the theme's `Background` color, ensuring full-width background coverage
across the entire terminal. This runs after screen composition and
notification injection.

The `renderer` package provides Unicode-aware width measurement, ANSI
stripping, and text wrapping/truncation for terminal-aware layout. See
`renderer/CONTEXT.md`.

---

# State Ownership

```
Registry (ScreenIndex → Screen)

  ├── HomeScreen         → HomeModel         (activity menu)
  ├── QuizScreen         → QuizModel         (flashcard quiz)
  ├── TypingQuizScreen   → TypingQuizModel   (typing-based quiz)
  ├── DecksScreen        → DeckSelectModel   (split-column deck + tag selection)
  ├── SearchScreen       → SearchModel       (text search)
  ├── StatisticsScreen   → StatisticsModel   (metrics)
  ├── SettingsScreen     → SettingsModel     (theme switch)
  ├── DetailScreen       → DetailModel       (entry view)
  └── PaletteScreen      → PaletteModel      (theme palette test)
```

Screens are stored in a `navigation.Registry` and managed by the
`navigation.Manager`. The root model references the current screen
through `m.Navigator.CurrentScreen()` rather than holding concrete fields.

Screens own only their own *local* (ephemeral) state — cursor positions,
queries, reveal state, card index. Shared data lives in the root model's
`Model.State ui.AppState` snapshot (see `internal/ui/state.go`):

```go
type AppState struct {
    Deck          *DeckData
    DeckProgress  map[string]stats.EntryProgress
    AllDecks      []string
    SelectedDecks []string
    AllTags       []string
    SelectedTags  []string
    AllDeckTags   map[string][]string
    QuizMode      QuizMode
    Stats         *stats.Summary
}
```

Screens that render from `AppState` implement `ui.StateSyncer` and receive the
snapshot via `SyncState(AppState) tea.Cmd`. The root pushes it to the active
screen at two occasions: when the screen becomes visible (after `OnEnter` in
`transitionTo`/`pushTo`/`popToPrevious`) and whenever the snapshot changes
(`ui.StateChangedMsg` → `syncActiveScreen()`). `SyncState` must be idempotent —
it recomputes derived state only when the incoming data differs.

Example:

Quiz owns:

- current card
- reveal state
- examples page
- inverse mode flag

Quiz does NOT own:

- the deck or per-card progress (read from `AppState` via `SyncState`)
- the global quiz mode (`AppState.QuizMode`, cycled via `ui.SetQuizModeMsg`)
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
│              ├─ ui.StateChangedMsg → syncActiveScreen() → active screen's SyncState()
│              ├─ ui.SetQuizModeMsg → AppState.QuizMode + StateChangedMsg
│              ├─ ui.RefreshStatsMsg → FetchStatsCmd()
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
`DeckSelectionChangedMsg`, `NavigateToDetailMsg`, `StateChangedMsg`,
`SetQuizModeMsg`, `RefreshStatsMsg`. Additional centralized event types live
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

Use `layout.Page()` and `layout.Column()` for view composition
instead of manual `strings.Builder` concatenation.

Do not prematurely optimize.

Bubble Tea already performs efficient redraws.

---

# Testing

Business logic should be testable without Bubble Tea.

Rendering helpers should be deterministic.

Navigation should be testable through emitted events.

---

# File Organization

Related packages outside `ui/`:

- **`internal/config/`** — User configuration from `~/.config/crds/`: directory creation, `config.yaml`, `keymaps.yaml`, `themes/*.yaml` discovery

```
ui/
├── screen.go           Screen interface + ScreenIndex type
├── theme.go            Semantic color theme (re-exports theme.Default)
├── state.go            AppState snapshot + StateSyncer protocol + StateChangedMsg
├── quiz_mode.go        QuizMode enum + parse/next
├── sorter.go           SortCards (mode-aware card ordering)

├── app/                Root model (Bubble Tea entry point)
│   ├── app.go          New() + Run() + config/keymap/theme init
│   ├── model.go        Root Model struct, GlobalState + State (AppState), messages
│   ├── events.go       dispatchEvent + dispatchKeyEvent + forwardToScreen + syncActiveScreen
│   ├── view.go         Root View() + help overlay using keymap.Registry
│   ├── update.go       Root Update()
│   ├── lifecycle.go    Lifecycle hooks, transitionTo, pushTo, popToPrevious (entry state sync)
│   ├── commands.go     NavigateToMsg, Dispatcher, config update
│   ├── config.go       Config + DefaultConfig + ApplyYAML
│   ├── dependencies.go DeckProvider, ProgressRecorder interfaces
│   ├── tick.go         Tick loop
│   └── tests/          state sync protocol tests

├── navigation/         Centralized navigation
│   ├── manager.go      Manager (Push, Pop, Replace, Forward, Reset, ...)
│   ├── stack.go        History stack with depth limit
│   ├── registry.go     Registry (ScreenIndex → Screen)
│   ├── events.go       9 event types (Push, Pop, Replace, Forward, ...)
│   └── tests/          black-box navigation tests

├── screens/            Screen implementations
│   ├── home.go         HomeModel — activity menu
│   ├── quiz.go         QuizModel — flashcard quiz
│   ├── typing_quiz.go  TypingQuizModel — typing-based quiz with fuzzy matching
│   ├── deck_select.go  DeckSelectModel — split-column deck+tag selection with search
│   ├── search.go       SearchModel — two-phase: input mode (type + filter) + results mode (navigate + select)
│   ├── statistics.go   StatisticsModel — study metrics
│   ├── settings.go     SettingsModel — theme switching
│   ├── detail.go       DetailModel — entry detail view
│   └── palette.go      PaletteModel — theme palette test

├── components/         Reusable components (29 total)
│   ├── display/        20 stateless render functions
│   │   ├── text.go           Text(content)
│   │   ├── label.go          Label(text)
│   │   ├── paragraph.go      Paragraph(content, width)
│   │   ├── divider.go        Divider(width)
│   │   ├── badge.go          Badge(text, variant)
│   │   ├── header.go         Header(title, width)
│   │   ├── footer.go         Footer(keys, width)
│   │   ├── card.go           Card struct + RenderCard(c, revealed, width)
│   │   ├── panel.go          Panel(content, width)
│   │   ├── section.go        Section(title, content, width)
│   │   ├── group.go          Group(title, content, width)
│   │   ├── window.go         Window(title, content, footer, width)
│   │   ├── list.go           RenderList + RenderListClipped (scrollable)
│   │   ├── table.go          Table(headers, rows, width)
│   │   ├── progress.go       ProgressBar(progress)
│   │   ├── notification.go   RenderNotification(text)
│   │   ├── status_bar.go     StatusBar(left, right, width)
│   │   ├── modal.go          RenderModal(title, content, width, height)
│   │   ├── confirm_dialog.go ConfirmDialog(title, msg, confirm, cancel, w, h)
│   │   └── error_dialog.go   ErrorDialog(title, msg, width, height)
│   └── interactive/    9 stateful Bubble Tea sub-models
│       ├── input_keys.go     Key config structs + keyIn() helper
│       ├── text_input.go     TextInputModel (cursor, focus)
│       ├── search_input.go   SearchInputModel (extends TextInput)
│       ├── checkbox.go       CheckboxModel (toggle, focus)
│       ├── radio_group.go    RadioGroupModel (single select)
│       ├── select.go         SelectModel (dropdown)
│       ├── multi_select.go   MultiSelectModel (checkbox dropdown)
│       ├── selectable_list.go SelectableListModel (multi-select)
│       ├── tree.go           TreeModel (expand/collapse)
│       └── spinner.go        SpinnerModel (animation frames)

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
│   ├── primary_bg.go   PrimaryBg()
│   ├── success_bg.go   SuccessBg()
│   ├── error_bg.go     ErrorBg()
│   ├── warning_bg.go   WarningBg()
│   └── styles_test.go  Tests

├── theme/              Design system (colors, typography, icons, borders)
│   ├── palette.go      18-field Palette (15 colors + Primary/Secondary/Accent) + DefaultPalette
│   ├── theme.go        Theme struct + NewTheme() + BorderFor()
│   ├── typography.go   6 text role styles
│   ├── borders.go      5 border styles
│   ├── icons.go        4 icon sets × 10 semantic slots
│   ├── spacing.go      7-tier Spacing scale
│   ├── border_role.go  BorderRole enum
│   ├── detect.go       Icon source auto-detection
│   ├── nerdfont.go     NerdFont detection
│   ├── config.go       YAML loading + style overrides
│   ├── presets.go      Dark/light/tokyonight presets
│   ├── store.go        Theme registry + switching
│   ├── DESIGN.md       Design language documentation
│   ├── theme_test.go   Tests
│   └── testdata/       6 YAML fixtures

├── keymap/             Centralized keybinding definitions
│   ├── keymap.go       Binding, BindingList, Global, List, Quiz, TypingQuiz,
│   │                   Decks, Search, Registry, NamedBinding, KeymapConfig,
│   │                   ApplyDefaultOverrides
│   └── keymap_test.go  Tests

├── layout/             Layout helpers
│   ├── page.go         Page(header, content, footer, height)
│   ├── column.go       Column(items...)
│   ├── row.go          Row(items...)
│   ├── center.go       Center(content, width, height)
│   ├── align.go        Align direction enum
│   ├── grid.go         Grid(items, columns, width)
│   ├── stack.go        Stack(items...)
│   ├── spacer.go       Spacer(n)
│   └── layout_test.go  Tests

├── events/             Centralized event types
│   └── events.go       TickMsg, ThemeSwitchMsg, ShowNotificationMsg,
│                       HideNotificationMsg

├── renderer/           Text rendering utilities
│   ├── width.go        VisibleWidth, LineWidth, MaxLineWidth, TextDimensions
│   ├── ansi.go         StripANSI, CountANSISequences
│   ├── wrap.go         Wrap, Truncate, Fit
│   └── renderer_test.go Tests

├── actions/            Action types (empty)
├── animations/         Animation helpers (empty)
├── debug/              Debug utilities (empty)
└── testdata/           Test fixtures (empty)
```

Keep responsibilities separate.

Avoid large files.

Most implementation work happens in `app/`, `navigation/`, and `screens/`.

Each package carries a `CONTEXT.md` describing how it works today. See
`docs/README.md` for the taxonomy.
