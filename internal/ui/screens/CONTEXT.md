# Screens Context

> Per-package context: how this package works today. Status and plans live in
> `docs/status.md` and `docs/roadmap.md` (see `docs/README.md`).

## Purpose

The `screens` package implements all application screens. Each screen is a
self-contained model that owns its own state, handles keyboard input, and
renders its view using the `components` package.

Screens are registered into the navigation `Registry` by `app/app.go` and
switched between via `ui.NavigateToMsg`.

---

## Architecture

Every screen implements the `ui.Screen` interface:

```go
type Screen interface {
    Init() tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)
    View() string
}
```

The `Update` method returns `(ui.Screen, tea.Cmd)` — the screen value itself
satisfies the interface. Navigation commands are returned as `tea.Cmd` that
produce `ui.NavigateToMsg`:

```go
func (m HomeModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
    case "enter":
        return m, func() tea.Msg {
            return ui.NavigateToMsg{Screen: ui.QuizScreen}
        }
}
```

---

## Navigation

Screens **do not** import or reference the `app` or `navigation` packages.
Instead they emit typed messages that the root model handles:

| Message | Emitted by | Handled in |
|---|---|---|
| `ui.NavigateToMsg` | Home, any screen needing transition | `app/events.go` → `transitionTo()` |
| `events.ThemeSwitchMsg` | Settings | `app/events.go` → `theme.Switch()` |
| `ui.SetQuizModeMsg` | Quiz / TypingQuiz (mode cycle) | `app/events.go` → global `AppState.QuizMode` |
| `ui.RefreshStatsMsg` | Statistics (`OnEnter`) | `app/events.go` → `FetchStatsCmd()` |

Global keybindings (`esc`, `?`, `ctrl+c`) are handled in `app/events.go`
before messages reach screens. Screens only handle their own keys.

---

## Global State Sync

The root `Model` owns a single canonical snapshot, `ui.AppState` (defined in
`internal/ui/state.go`), holding the shared data screens render from:

- `Deck`, `DeckProgress` — merged selected decks + per-card progress
- `AllDecks`/`SelectedDecks`, `AllTags`/`SelectedTags`, `AllDeckTags`
- `QuizMode` — global, shared by both quiz screens
- `Stats` — latest stats summary

Screens that need this data implement `ui.StateSyncer`:

```go
type StateSyncer interface {
    SyncState(AppState) tea.Cmd
}
```

The root pushes the snapshot to the **active** screen at exactly two occasions:

1. **On entry** — `transitionTo`/`pushTo`/`popToPrevious` call
   `syncActiveScreen()` right after `OnEnter` (`app/lifecycle.go`), so a screen
   that just became visible receives the current state.
2. **On state change** — every mutation of `AppState` emits `ui.StateChangedMsg`;
   the root forwards it to the active screen via `syncActiveScreen()`
   (`app/events.go`). Transient messages (`SaveAnswerMsg`, ticks, notifications)
   do **not** trigger a sync.

`SyncState` must be **idempotent**: it should recompute derived state (sorting,
filtering) only when the incoming snapshot actually differs, so unrelated state
changes don't reset screen-local progress (e.g. a quiz's `cardIndex`).

Current implementers:

| Screen | Reads |
|---|---|
| `QuizModel` | `Deck`, `DeckProgress`, `QuizMode` → re-sort |
| `TypingQuizModel` | `Deck`, `DeckProgress`, `QuizMode` → re-sort |
| `DeckSelectModel` | `AllDecks`/`SelectedDecks`, `AllTags`/`SelectedTags`, `AllDeckTags` → rebuild columns |
| `SearchModel` | `Deck.Cards` → rebuild results |
| `StatisticsModel` | `Stats` → refresh summary |

`DetailModel` is the exception: it receives its entry via `NavigateToDetailMsg`
(`entrySetter.SetEntry`), a per-entry navigation payload, not from `AppState`.

---

## Screen Indexes

Defined in `ui/screen.go`:

```go
HomeScreen     ScreenIndex = iota
QuizScreen
SearchScreen
StatisticsScreen
SettingsScreen
DetailScreen
DecksScreen
TypingQuizScreen
PaletteScreen
```

---

## Current Screens

### HomeModel (`home.go`)

- **State**: `cursor`, `activities` list
- **Keys**: `up`/`k`, `down`/`j`, `enter`
- **Behavior**: Menu of activities (Study, Search, Statistics, Settings).
  Enter navigates to the selected screen.
- **Renders**: ASCII logo (from `assets/logo.txt`) + menu list (centered vertically) + footer (pinned to bottom)

### QuizModel (`quiz.go`)

- **State**: `cardIndex`, `revealed`, `cards`, `originalCards`, `cardProgress`,
  `mode`, `examplesPage`, `inverse`
- **Keys**: `enter` (reveal), `tab` (inverse), `m` (mode cycle),
  `a`/`h`/`o`/`e` or `1`/`2`/`3`/`4` (grade), `[`/`left` (prev example),
  `]`/`right` (next example)
- **Behavior**: Displays a flashcard. On enter, reveals the answer and shows
  the grade menu (`[a]gain`, `[h]ard`, `[o]kay`, `[e]asy`). Grading advances
  to the next card. When all cards are done, shows "Quiz complete!" with
  `[enter] restart` and `[esc] back` options. Tab toggles inverse mode —
  shows translations as the question and the term as the answer.
- **Data**: receives deck, progress, and mode via `SyncState` (see Global State
  Sync); re-sorts cards only when that data actually changes. Mode cycling
  (`m`) emits `ui.SetQuizModeMsg` so both quiz screens stay in sync.
- **Renders**: Term (centered, vertically padded at height/4) + (if revealed)
  correct answer + centered grade menu + bottom section (notes, tags as PrimaryBg pills,
  paginated examples in single/two-column layout, 8-char side padding) +
  progress "card N/M" + footer shortcuts.

### SearchModel (`search.go`)

- **State**: `query`, `cursor`, `scrollOffset`, `results`, `cards`, `mode` (input/results)
- **Keys**: `j`/`k`/arrows, `enter`, `backspace`, printable chars, `esc`
- **Behavior**: Two-phase search. In **input mode**, all printable keys
  (including j/k) type into the query, results filter live. Enter switches
  to **results mode** if results exist. In results mode, j/k/arrows navigate,
  Enter opens the selected result's detail, Esc returns to input mode.
  Implements `Lifecycle` — `OnEnter` resets to input mode, `OnLeave` clears
  query/results/cursor/mode when navigating away.
- **Data**: receives the current deck's cards via `SyncState`; re-filters
  results when a query is active.
- **Scrolling**: Uses `RenderListClipped` with `scrollOffset` to clip
  results to available terminal height. `adjustScroll()` keeps cursor
  visible when navigating. `↑`/`↓` indicators show when content is clipped.
- **Renders**: Header + FocusedInput styled input (with/without cursor) +
  RenderListClipped (clipped to fit terminal height) + footer (dynamic per mode)

### StatisticsModel (`statistics.go`)

- **State**: `summary` (from `SyncState`)
- **Keys**: None (information only)
- **Behavior**: Displays 6 study metrics in `styles.Panel` containers:
  Reviewed Today, Accuracy, Due Today, Current Streak, Total Cards, Mastered.
  `OnEnter` emits `ui.RefreshStatsMsg`; the root fetches stats and pushes them
  back via `SyncState`.
- **Renders**: Header + Panel grid + footer
- **Known issue**: Due Today and Current Streak show 0 / "—" — no scheduler
  wired yet. Tracked in `docs/status.md`.

### SettingsModel (`settings.go`)

- **State**: `cursor`, `themes` (from `theme.Names()`)
- **Keys**: `up`/`k`, `down`/`j`, `enter`
- **Behavior**: Theme selector. Enter emits `ui.ThemeSwitchMsg`.
- **Renders**: Header + theme list with `ui.Theme.Icons.Navigate` marker
  and `ui.Theme.Icons.Check` for active theme + footer

### DetailModel (`detail.go`)

- **State**: `Term`, `Translations`, `Examples`, `Notes`
- **Keys**: None (esc handled globally — pops back to Search)
- **Behavior**: Displays vocabulary entry details using `styles.Card` containers
  for translations, examples, and notes sections. Receives entry data from
  Search via `NavigateToDetailMsg`. Uses stacked navigation (`pushTo`) so
  Esc returns to Search.
- **Renders**: Header + Primary term + section(s) + footer

### DeckSelectModel (`deck_select.go`)

- **State**: Two columns (decks + tags), each with `cursor`, `items`, `selected` map,
  `searchActive`, `searchQuery`, `scrollOffset`. Cross-filtering: OR logic between
  columns (any selected deck shows all its tags, any selected tag shows all its decks).
- **Keys**: `up`/`k`, `down`/`j`, `tab` (next column), `shift+tab` (prev column),
  `space` (toggle), `a` (toggle all), `s` (search toggle), `enter` (confirm),
  `esc` (clear search / dismiss via BackHandler), `backspace` (search delete char)
- **Behavior**: Split-column selection with per-column search. Toggle items independently
  in each column. Search filters items in real time via substring match. BackHandler
  intercepts Esc: clears search query (non-empty) or dismisses screen (empty query).
  Enter confirms selection and emits `DeckSelectionChangedMsg` with both `Selected`
  and `SelectedTags`. Columns are rebuilt from `AppState` via `SyncState`.
- **Renders**: Header + two centered columns with secondary-color divider + scrollable
  lists with selected items highlighted via `Secondary.Render("✓ " + name)` + footer

### TypingQuizModel (`typing_quiz.go`)

- **State**: `CardIndex`, `Input` (text input), `Feedback`, `Progress`, `Cards`,
  `Inverse` (mode toggle)
- **Keys**: `enter` (submit answer), `ctrl+r` (reveal without grading),
  `tab` (toggle inverse mode), `m` (mode cycle)
- **Behavior**: Displays a term, user types the translation. Uses
  `fuzzy.LevenshteinMatcher` to grade the typed answer. `ctrl+r` reveals the
  correct answer (records `Again`, grade 1). Tab toggles inverse mode —
  shows translations as the prompt and expects the term as the answer;
  term variants use the same `()`/`[]` expansion syntax as translations.
  When all cards are done, shows "Quiz complete!" with `[enter] restart`
  and `[esc] back` options. Receives deck, progress, and mode via `SyncState`;
  mode cycling (`m`) emits `ui.SetQuizModeMsg`.
- **Renders**: Header + term (centered, vertically padded, at height/5) + correct-answer
  slot (always present; empty placeholder when unrevealed, "Correct: ..." text when revealed) +
  centered input with cursor (fixed position — does not shift on reveal) + bottom section after
  reveal (notes, tags as PrimaryBg pills, paginated examples in single/two-column layout,
  all sections with 8-char side padding) + progress "card N/M" in footer + footer shortcuts.
  Header shows "(inverse)" suffix in inverse mode.

### PaletteModel (`palette.go`)

- **State**: `scrollOffset`, `width`, `height`
- **Keys**: `↑`/`k` (scroll up), `↓`/`j` (scroll down), `esc` (back to Home)
- **Behavior**: Scrollable dev/debug screen showing every visual property of the
  current theme live. Displays 7 sections: theme info, 15 palette colors as
  colored pill swatches with ANSI/hex values and purpose descriptions, 14
  semantic styles with sample text, 6 typography roles, 10 icon slots with
  current glyphs, 5 border styles as rendered boxes, and 7 spacing tiers as
  bar charts. Updates instantly when theme is switched via Settings.
- **Renders**: Custom scrollable layout — builds full content, applies scroll
  offset, renders visible slice + pinned footer.

---

## Common Patterns

### Keyboard handling

```
up/k → cursor up
down/j → cursor down
enter → confirm/select
```

Every visible action should have a keyboard shortcut, documented in the
footer via `components.Footer(...)`.

### Rendering

Screens use layout primitives for consistent structure:

```go
return layout.Page(
    components.Header("Title", m.width),
    layout.Column(/* body sections */),
    components.Footer(keys, m.width),
)
```

`layout.Page()` handles header/body/footer with `"\n\n"` separators.
`layout.Column()` composes body sections with `"\n\n"` separators.

**Exception:** `HomeModel` does its own vertical centering — it computes
`topPad`/`bottomPad` to center the logo + menu body above a pinned footer,
so the footer stays at the terminal bottom while the body floats vertically.

Prefer `components.*` over raw `ui.Theme.*` calls for consistency.

### Screen lifecycle

Screens can optionally implement the `app.Lifecycle` interface:

```go
type Lifecycle interface {
    OnEnter() tea.Cmd
    OnLeave() tea.Cmd
}
```

This is checked by `app/lifecycle.go` — screens that don't implement it
get no-op lifecycle hooks.

### BackHandler

Screens can implement `ui.BackHandler` to intercept Back (Esc) before
the global handler applies default behavior:

```go
type BackHandler interface {
    HandleBack() bool
}
```

Return `true` if the screen consumed the event. Currently implemented by
`SearchModel` (returns to input mode when in results mode) and
`DeckSelectModel` (clears search query if non-empty, deactivates search
if empty query — returns `true` when query cleared, `false` otherwise).

---

## Dependencies

- `github.com/charmbracelet/bubbletea` — `tea.Model`, `tea.Cmd`, `tea.Msg`
- `crds/internal/ui` — `ui.Screen`, `ui.ScreenIndex`, `ui.NavigateToMsg`,
  `ui.Theme`
- `crds/internal/ui/events` — `events.ThemeSwitchMsg`
- `crds/internal/ui/components` — `Header`, `Footer`, `RenderCard`,
  `ProgressBar`, `RenderList`, `Text`
- `crds/internal/ui/layout` — `Page`, `Column`
- `crds/internal/ui/theme` — `theme.Names()`, `theme.CurrentName()`

Screens do NOT depend on:
- `crds/internal/ui/app`
- `crds/internal/ui/navigation`
- `crds/internal/parser`

---

## File Organization

```
screens/
├── CONTEXT.md
├── home.go          HomeModel — activity menu
├── quiz.go          QuizModel — flashcard quiz
├── quiz_shared.go   Shared quiz rendering (tags, examples, bottom section, pagination)
├── search.go        SearchModel — text input + results
├── statistics.go    StatisticsModel — study metrics
├── settings.go      SettingsModel — theme switching
├── detail.go        DetailModel — entry detail
├── deck_select.go   DeckSelectModel — split-column deck+tag selection
├── typing_quiz.go   TypingQuizModel — typing quiz
└── palette.go       PaletteModel — theme palette test
```

## Notes for changes

- Home could show recent decks or quick-start options once deck loading is
  richer.
- Screens should support dynamic sizing via `SetSize(w, h int)` rather than
  relying on hardcoded defaults.
- Search filters `Deck.Cards` in memory via `SyncState`; debouncing and
  async loading are potential future refinements.
- Quiz and TypingQuiz are candidates for `OnEnter` lifecycle hooks (e.g.
  reset per-visit state).

These ideas are tracked in `docs/roadmap.md`.
