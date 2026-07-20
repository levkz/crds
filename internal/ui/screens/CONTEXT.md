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
```

---

## Current Screens

### HomeModel (`home.go`)

- **State**: `cursor`, `activities` list
- **Keys**: `up`/`k`, `down`/`j`, `enter`
- **Behavior**: Menu of activities (Study, Search, Statistics, Settings).
  Enter navigates to the selected screen.
- **Renders**: Header + component list + footer

### QuizModel (`quiz.go`)

- **State**: `CardIndex`, `Revealed`, `Progress`, `Cards`
- **Keys**: `enter` (reveal), `1`–`4` (grade)
- **Behavior**: Displays a flashcard. On enter, reveals the answer. Grading
  keys (`1`–`4`) record difficulty via `grade()` helper.
- **Renders**: Header + card + progress bar + footer with grading options
- **Known issues**: `Cards` is pre-populated — no data loading wired yet.
  Progress always 0. Grading returns no-op command.

### SearchModel (`search.go`)

- **State**: `query`, `cursor`, `results`, `focused`
- **Keys**: `up`/`k`, `down`/`j`, `enter`, `backspace`, `tab`, printable chars
- **Behavior**: Text input with live query filtering. Results via `components.RenderList`.
  Enter opens the selected result's detail screen.
- **Renders**: Header + FocusedInput styled input + RenderList + footer
- **Known issue**: Results are placeholder text — no real search wired yet.

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
- **Keys**: None (esc handled globally)
- **Behavior**: Displays vocabulary entry details using `styles.Card` containers
  for translations, examples, and notes sections.
- **Renders**: Header + Primary term + section(s) + footer
- **Known issue**: No data is passed to the model — shows "Select an entry"
  until wired with real data.

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

Every screen follows the same layout:

```
Header()
\n\n
Content
\n\n
Footer()
```

Use `strings.Builder` to construct views. Prefer `components.*` over
raw `ui.Theme.*` calls for consistency.

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

---

## Dependencies

- `github.com/charmbracelet/bubbletea` — `tea.Model`, `tea.Cmd`, `tea.Msg`
- `crds/internal/ui` — `ui.Screen`, `ui.ScreenIndex`, `ui.NavigateToMsg`,
  `ui.Theme`
- `crds/internal/ui/events` — `events.ThemeSwitchMsg`
- `crds/internal/ui/components` — `Header`, `Footer`, `RenderCard`,
  `ProgressBar`, `RenderList`, `Text`
- `crds/internal/ui/theme` — `theme.Names()`, `theme.CurrentName()`

Screens do NOT depend on:
- `crds/internal/ui/app`
- `crds/internal/ui/navigation`
- `crds/internal/parser`, `crds/internal/model`

---

## File Organization

```
screens/
├── CONTEXT.md
├── home.go          HomeModel — activity menu
├── quiz.go          QuizModel — flashcard quiz
├── search.go        SearchModel — text input + results
├── statistics.go    StatisticsModel — study metrics
├── settings.go      SettingsModel — theme switching
└── detail.go        DetailModel — entry detail
```

---

## Suggestions

1. **Data loading** — Screens like Quiz should accept data at construction
   time or load it via commands returned from `Init()`. Currently `Cards` is
   prepopulated and always empty.

2. **Lifecycle hooks** — Use `app.Lifecycle` for screens that need to
   (re)load data on enter or save on leave. Search and Quiz are candidates.

3. **Home improvements** — After wiring deck loading, show recent decks or
   quick-start options. Consider `components.Text` for descriptions below
   each activity.

4. **Keyboard documentation** — Footer strings are currently inline. Consider
   a `keymap` package for centralized shortcut definitions.

5. **Width awareness** — Screens currently use `components.Header(60)`,
   `components.Footer(60)`, etc. These widths should come from the actual
   terminal width (`m.Width` in the root model). Pass width to screens or
   define a `SetSize(w, h int)` on the `Screen` interface.

6. **Event-driven search** — When search is wired to real data, consider
   debouncing the query input and using a channel or command to load results
   asynchronously.

7. **Statistics refresh** — Statistics could refresh on `OnEnter()` lifecycle
   hook to show fresh data each time the screen is visited.

---

## TODOs

- [ ] Wire quiz data loading — `QuizModel.Cards` should come from a deck via
      `app.LoadDeckCmd` or similar
- [ ] Wire search to real data — `filterResults()` should query actual entries
- [ ] Wire statistics to real metrics — pull from a scheduler/progress store
- [ ] Wire detail view — accept an entry ID and populate `DetailModel` fields
- [ ] Add `Width`/`Height` awareness — allow screens to adapt to terminal size
- [ ] Add `Lifecycle` implementation to QuizModel (load deck on enter)
- [ ] Centralize keybinding strings in a `keymap` package
