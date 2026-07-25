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
    height,
)
```

`fillBackground()` in `app/view.go` wraps each ANSI-reset-delimited segment with the theme's `Background` color, ensuring full-width background coverage across the entire terminal. This runs after screen composition and notification injection.

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

Decks

Multi-deck selection.

Toggle with space.

Select-all/deselect-all with a.

Enter confirms selection.

Quiz (Flash Cards)

Display cards.

Reveal answers.

Collect grades.

TypingQuiz

Display term.

User types translation.

Fuzzy-match answer and auto-grade.

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

All 29 components are implemented, split across two subpackages:

**`display/`** (20 stateless): Header, Footer, Card, Progress, List, Table,
Modal, Notification, Text, Label, Badge, Paragraph, Divider, Panel, Section,
Group, Window, StatusBar, ConfirmDialog, ErrorDialog

**`interactive/`** (9 stateful sub-models): TextInput, SearchInput, Checkbox,
RadioGroup, Select, MultiSelect, SelectableList, Tree, Spinner

Interactive components accept optional key config structs for configurable
vim-style keybindings. See `components/CONTEXT.md` for details.

---

# File Organization

Related packages outside `ui/`:

- **`internal/config/`** — User configuration from `~/.config/crds/`: directory creation, `config.yaml`, `keymaps.yaml`, `themes/*.yaml` discovery

```
ui/
├── screen.go           Screen interface + ScreenIndex type
├── theme.go            Semantic color theme (re-exports theme.Default)

├── app/                Root model (Bubble Tea entry point)
│   ├── app.go          New() + Run() + config/keymap/theme init
│   ├── model.go        Root Model struct, GlobalState, messages
│   ├── events.go       dispatchEvent + dispatchKeyEvent + forwardToScreen
│   ├── view.go         Root View() + help overlay using keymap.Registry
│   ├── update.go       Root Update()
│   ├── lifecycle.go    Lifecycle hooks, transitionTo, pushTo, popToPrevious
│   ├── commands.go     NavigateToMsg, Dispatcher, config update
│   ├── config.go       Config + DefaultConfig + ApplyYAML
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
│   ├── typing_quiz.go  TypingQuizModel — typing-based quiz with fuzzy matching
│   ├── decks.go        DecksModel — multi-deck selection with toggle/toggle-all
│   ├── search.go       SearchModel — two-phase: input mode (type + filter) + results mode (navigate + select)
│   ├── statistics.go   StatisticsModel — study metrics
│   ├── settings.go     SettingsModel — theme switching
│   └── detail.go       DetailModel — entry detail view

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
│   ├── presets.go      Dark/light/tokyonight presets
│   ├── store.go        Theme registry + switching
│   ├── DESIGN.md       Design language documentation
│   ├── theme_test.go   54 tests
│   └── testdata/       6 YAML fixtures

├── keymap/             Centralized keybinding definitions
│   ├── keymap.go       Binding, BindingList, Global, List, Quiz, Search,
│   │                   Registry, NamedBinding, KeymapConfig, ApplyDefaultOverrides
│   ├── keymap_test.go  16 tests
│   ├── CONTEXT.md      Package context
│   └── TODO.md         Progress tracker

├── layout/             Layout helpers
│   ├── page.go         Page(header, content, footer, height)
│   ├── column.go       Column(items...)
│   ├── row.go          Row(items...)
│   ├── center.go       Center(content, width, height)
│   ├── align.go        Align direction enum
│   ├── grid.go         Grid(items, columns, width)
│   ├── stack.go        Stack(items...)
│   ├── spacer.go       Spacer(n)
│   ├── layout_test.go  Tests
│   └── CONTEXT.md

├── events/             Centralized event types
│   ├── events.go       TickMsg, ThemeSwitchMsg, ShowNotificationMsg,
│   │                   HideNotificationMsg
│   └── TODO.md

├── renderer/           Text rendering utilities
│   ├── wrap.go         Wrap(content, width)
│   ├── width.go        VisibleWidth(s), Truncate(s, max)
│   ├── ansi.go         AnsiWidth(s) — strips ANSI for width calc
│   ├── renderer_test.go Tests
│   └── CONTEXT.md

├── actions/            Action types (empty)
├── animations/         Animation helpers (empty)
├── debug/              Debug utilities (empty)
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

# Current Status

## Complete

- **Theme** (`theme/`): Full implementation with 54+ tests, including 15-color palette, 10 semantic styles, typography, borders, 10-slot icons (4 icon sources), spacing scale, border roles, YAML loading with style overrides, and theme switching (built-in: default, dark, light, tokyonight)
- **Styles** (`styles/`): All 12 style definitions implemented with 60 tests — Header, Footer, SelectedItem, FocusedInput, Error, Warning, Success, Hint, MutedText, Card, Panel, Modal
- **Components** (`components/`): All 29 components implemented — 20 in `display/` (Header, Footer, Card, Progress, List, Table, Modal, Notification, Text, Label, Badge, Paragraph, Divider, Panel, Section, Group, Window, StatusBar, ConfirmDialog, ErrorDialog) and 9 in `interactive/` (TextInput, SearchInput, Checkbox, RadioGroup, Select, MultiSelect, SelectableList, Tree, Spinner)
- **Screens** (`screens/`): All 8 screens implemented — Home (activity menu), Quiz (flashcard), TypingQuiz (typing-based with fuzzy matching), Decks (multi-deck selection), Search (text input + results), Statistics (metrics display), Settings (theme switching), Detail (entry view)
- **Navigation** (`navigation/`): Complete with 82 black-box tests including Manager, stack, and registry
- **App** (`app/`): Root Bubble Tea model, events, lifecycle, commands, config, theme loading, overlay/notification system. `New()` wires keymap overrides (`keymaps.yaml`), user themes (`themes/`), and app config (`config.yaml`) from `~/.config/crds/`
- **Keymap** (`keymap/`): Centralized keybinding definitions. `Binding` with `Match()`, `BindingList` with `Help()`, per-screen structs (`Global`, `List`, `Quiz`, `Search`) with `Footer()`/`Revealed()`/`Unrevealed()`, `Registry` with `Bindings()`/`FindBinding()`, `KeymapConfig` + `ApplyDefaultOverrides()` for user-defined overrides. 16 tests.
- **Config** (`internal/config/`): User configuration directory (`~/.config/crds/`). Auto-creates directory tree with default files. Loads `config.yaml` (theme, animation, quiz limit), `keymaps.yaml` (keybinding overrides), and discovers `themes/*.yaml`. 13 tests.

## Placeholder Directories

- **actions** (empty): Action types
- **animations** (empty): Animation utilities
- **debug** (empty): Debug utilities
- **testdata** (empty): Test fixtures

---

# Known Issues

- QuizModel.Cards is pre-populated — no data loading wired yet
- Quiz progress always 0, grading returns no-op
- Statistics shows zeroes for all metrics (no scheduler wired)
- `~/.config/crds/` is created on `app.New()` but CLI commands are stubs, so it only triggers when the UI actually launches

---

# Future Work

Likely future additions:

- split layouts
- search highlighting
- progress graphs
- animations
- chord keybindings (e.g. `g` `g` for top of list)
- mouse support
- plugins
- command palette

The current architecture should support these without major refactoring.
