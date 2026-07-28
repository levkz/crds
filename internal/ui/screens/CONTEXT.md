# Screens Context

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

Global keybindings (`esc`, `?`, `ctrl+c`) are handled in `app/events.go`
before messages reach screens. Screens only handle their own keys.

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

- **State**: `cardIndex`, `revealed`, `cards`, `examplesPage`, `inverse`
- **Keys**: `enter` (reveal), `tab` (inverse), `a`/`h`/`o`/`e` or `1`/`2`/`3`/`4` (grade), `[`/`left` (prev example), `]`/`right` (next example)
- **Behavior**: Displays a flashcard. On enter, reveals the answer and shows
  the grade menu (`[a]gain`, `[h]ard`, `[o]kay`, `[e]asy`). Grading advances
  to the next card; navigating to Statistics when deck finishes. Tab toggles
  inverse mode — shows translations as the question and the term as the answer.
- **Renders**: Term (centered, vertically padded at height/4) + (if revealed)
  correct answer + centered grade menu + bottom section (notes, tags as PrimaryBg pills,
  paginated examples in single/two-column layout, 8-char side padding) +
  progress "card N/M" + footer shortcuts.
- **Known issues**: `Cards` is pre-populated — no data loading wired yet.

### SearchModel (`search.go`)

- **State**: `query`, `cursor`, `scrollOffset`, `results`, `cards`, `mode` (input/results)
- **Keys**: `j`/`k`/arrows, `enter`, `backspace`, printable chars, `esc`
- **Behavior**: Two-phase search. In **input mode**, all printable keys
  (including j/k) type into the query, results filter live. Enter switches
  to **results mode** if results exist. In results mode, j/k/arrows navigate,
  Enter opens the selected result's detail, Esc returns to input mode.
  Implements `Lifecycle` — `OnEnter` resets to input mode, `OnLeave` clears
  query/results/cursor/mode when navigating away.
- **Scrolling**: Uses `RenderListClipped` with `scrollOffset` to clip
  results to available terminal height. `adjustScroll()` keeps cursor
  visible when navigating. `↑`/`↓` indicators show when content is clipped.
- **Renders**: Header + FocusedInput styled input (with/without cursor) +
  RenderListClipped (clipped to fit terminal height) + footer (dynamic per mode)

### StatisticsModel (`statistics.go`)

- **State**: None (static display)
- **Keys**: None (information only)
- **Behavior**: Displays 6 study metrics in `styles.Panel` containers:
  Reviewed Today, Accuracy, Due Today, Current Streak, Total Cards, Mastered.
- **Renders**: Header + Panel grid + footer
- **Known issue**: All metrics show zero / "—" — no scheduler wired yet.

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

### DecksModel (`decks.go`)

- **State**: `cursor`, `decks` list, `selected` set
- **Keys**: `up`/`k`, `down`/`j`, `space` (toggle), `a` (toggle all), `enter` (confirm)
- **Behavior**: Multi-deck selection with checkmarks. Space toggles current item,
  'a' toggles all. Enter emits `DeckSelectionChangedMsg` and navigates to Home.
  Implements `Lifecycle` — `OnLeave` saves the current selection whenever the
  user leaves the screen (via Enter or Esc).
- **Renders**: Header + checkmarked list + footer

### TypingQuizModel (`typing_quiz.go`)

- **State**: `CardIndex`, `Input` (text input), `Feedback`, `Progress`, `Cards`,
  `Inverse` (mode toggle)
- **Keys**: `enter` (submit answer), `ctrl+r` (reveal without grading),
  `tab` (toggle inverse mode)
- **Behavior**: Displays a term, user types the translation. Uses
  `fuzzy.LevenshteinMatcher` to grade the typed answer. `ctrl+r` reveals the
  correct answer (records `Again`, grade 1). Tab toggles inverse mode —
  shows translations as the prompt and expects the term as the answer;
  term variants use the same `()`/`[]` expansion syntax as translations.
- **Renders**: Header + term (centered, vertically padded) + centered input with cursor +
  bottom section after reveal (notes, tags as PrimaryBg pills, paginated examples in single/two-column layout,
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
`SearchModel` (returns to input mode when in results mode).

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
├── decks.go         DecksModel — deck selection
├── typing_quiz.go   TypingQuizModel — typing quiz
└── palette.go       PaletteModel — theme palette test
```

---

## Suggestions

1. **Data loading** — Screens like Quiz should accept data at construction
   time or load it via commands returned from `Init()`. Currently `Cards` is
   prepopulated and always empty.

2. **Lifecycle hooks** — Search and Decks now implement `app.Lifecycle`.
   Quiz and TypingQuiz are candidates for `OnEnter` (load deck data) and
   Statistics for `OnEnter` (refresh metrics).

3. **Home improvements** — After wiring deck loading, show recent decks or
   quick-start options. Consider `components.Text` for descriptions below
   each activity.

4. **Width awareness** — `SetSize(w, h int)` is now called on every screen
   during `transitionTo()` via the `Lifecycle` interface. Screens should
   support dynamic sizing rather than relying on hardcoded defaults.

5. **Event-driven search** — When search is wired to real data, consider
   debouncing the query input and using a channel or command to load results
   asynchronously.

6. **Statistics refresh** — Statistics could refresh on `OnEnter()` lifecycle
   hook to show fresh data each time the screen is visited.

---

## TODOs

- [ ] Wire quiz data loading — `QuizModel.Cards` should come from a deck via
      `app.LoadDeckCmd` or similar
- [ ] Wire statistics to real metrics — pull from a scheduler/progress store
- [ ] Add `Width`/`Height` awareness — allow screens to adapt to terminal size
- [ ] Add `Lifecycle` implementation to QuizModel (load deck on enter)
