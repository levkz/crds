# Keymap Context

> Per-package context: how this package works today. Status and plans live in
> `docs/status.md` and `docs/roadmap.md` (see `docs/README.md`).

## Purpose

The `keymap` package centralizes keybinding definitions so screens and the
root model never hardcode key strings. Each keymap is a group of `Binding`
values — each holding the keys that trigger it and a help label.

---

## Current keyboard handling (with keymap)

Keys flow through two layers:

### Layer 1 — Global (app/events.go)

```go
func (m Model) dispatchKeyEvent(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    case keymap.DefaultGlobal.Quit.Match(msg):
        return m, tea.Sequence(m.ShutdownCmd(), tea.Quit)
    case keymap.DefaultGlobal.Help.Match(msg):
        return m.WithOverlay(HelpOverlay), nil
    case keymap.DefaultGlobal.Back.Match(msg):
        // dismiss overlay → go back → home
    }
    return m.forwardToScreen(msg)
}
```

All three global keys (`ctrl+c`, `?`, `esc`) are defined in `keymap.DefaultGlobal`.

### Layer 2 — Screen-local key dispatch

Each screen's `Update` uses `keymap.Default*` instead of hardcoded strings:

| Screen     | Keymap                       |
|------------|------------------------------|
| Home       | `keymap.DefaultList`         |
| Quiz       | `keymap.DefaultQuiz`         |
| TypingQuiz | `keymap.DefaultTypingQuiz`   |
| Search     | `keymap.DefaultSearch`       |
| Decks      | `keymap.DefaultDecks`        |
| Settings   | `keymap.DefaultList`         |
| Statistics | (no input)                   |
| Detail     | (no input)                   |

### Layer 3 — Footer help text

Every screen derives its footer from keymap methods:

```go
// Home
components.Footer(keymap.DefaultList.Footer() + " · " + keymap.DefaultGlobal.Help.Help)

// Quiz (unrevealed)
components.Footer(keymap.DefaultQuiz.Unrevealed())

// Quiz (revealed)
components.Footer(keymap.DefaultQuiz.Revealed())

// TypingQuiz (unrevealed)
components.Footer(keymap.DefaultTypingQuiz.Footer() + " · " + keymap.DefaultGlobal.Back.Help)

// TypingQuiz (revealed)
components.Footer("enter/down next · esc back")

// Search
components.Footer(keymap.DefaultSearch.Footer() + " · " + keymap.DefaultGlobal.Back.Help)

// Decks
components.Footer(keymap.DefaultDecks.Footer() + " · " + keymap.DefaultGlobal.Back.Help)

// Settings
components.Footer(keymap.DefaultList.Footer() + " · " + keymap.DefaultGlobal.Back.Help)
```

No hardcoded footer strings remain.

---

## Keymap package design

### Binding

```go
type Binding struct {
    Keys []string   // e.g. {"up", "k"}
    Help string     // e.g. "↑ navigate" — shown in footer
}

func (b Binding) Match(msg tea.KeyMsg) bool { ... }
```

Bindings hold all alternate keys (e.g. `up` + `k`) for the same action.
`Match()` is the primary way to test a keypress against a binding.

### BindingList

```go
type BindingList []Binding

func (bl BindingList) Help() string { ... }
```

Joins each binding's `Help` field with ` · `, skipping empty ones.

### Predefined keymaps

| Variable          | Fields                                                   |
|-------------------|----------------------------------------------------------|
| `DefaultGlobal`   | `Quit` (`ctrl+c`), `Help` (`?`), `Back` (`esc`)         |
| `DefaultList`     | `Up` (`up`/`k`), `Down` (`down`/`j`), `Select` (`enter`)|
| `DefaultQuiz`     | `Reveal` (`enter`), `Again`/`Hard`/`Good`/`Easy` (`1`–`4`), `Inverse` (`tab`) |
| `DefaultTypingQuiz` | `Submit` (`enter`), `Reveal` (`ctrl+r`), `Inverse` (`tab`), `PrevExample` (`[`/`left`), `NextExample` (`]`/`right`), `ModeCycle` (`ctrl+t`), `ToggleParse` (`ctrl+p`) |
| `DefaultDecks`    | embeds `List` + `Toggle` (`space`), `ToggleAll` (`a`)    |
| `DefaultSearch`   | embeds `List` + `Open` (`enter`), `DeleteChar` (`backspace`) |

Each struct also has footer helpers:
- `Global.Footer()`, `List.Footer()`
- `Quiz.Unrevealed()`, `Quiz.Revealed()`
- `TypingQuiz.Footer()`
- `Decks.Footer()`
- `Search.Footer()`

### Registry

```go
type Registry struct {
    Global     Global
    List       List
    Quiz       Quiz
    TypingQuiz TypingQuiz
    Decks      Decks
    Search     Search
}

func (r Registry) Bindings() []NamedBinding    // all bindings for display
func (r Registry) FindBinding(key string) *NamedBinding  // lookup by key
```

`NamedBinding` wraps `Binding` with `Group` and `Action` metadata:

```go
type NamedBinding struct {
    Group   string   // e.g. "Global", "List"
    Action  string   // e.g. "Quit", "Up"
    Binding Binding
}
```

`DefaultRegistry` is the pre-configured instance.
The help overlay in `app/view.go` uses `DefaultRegistry.Bindings()` to render the keyboard shortcuts reference.

### Usage from a screen

```go
import "crds/internal/ui/keymap"

func (m HomeModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case keymap.DefaultList.Up.Match(msg):
            m.cursor--
        case keymap.DefaultList.Down.Match(msg):
            m.cursor++
        case keymap.DefaultList.Select.Match(msg):
            // navigate
        }
    }
}
```

### User-defined overrides

`KeymapConfig` mirrors the keymap struct hierarchy with optional override fields:

```go
type BindingOverride struct {
    Keys []string `yaml:"keys"`
    Help *string  `yaml:"help,omitempty"`
}

type KeymapConfig struct {
    Global     *struct { Quit *BindingOverride ... } `yaml:"global,omitempty"`
    List       *struct { Up   *BindingOverride ... } `yaml:"list,omitempty"`
    Quiz       *struct { Reveal, Inverse *BindingOverride ... } `yaml:"quiz,omitempty"`
    TypingQuiz *struct { Submit, Reveal, Inverse *BindingOverride ... } `yaml:"typing_quiz,omitempty"`
    Decks      *struct { Toggle, ToggleAll *BindingOverride ... } `yaml:"decks,omitempty"`
    Search     *struct { Open *BindingOverride ... } `yaml:"search,omitempty"`
}
```

Call `keymap.ApplyDefaultOverrides(cfg)` to apply overrides to all `Default*`
vars. This is called from `app.New()` after loading `~/.config/crds/keymaps.yaml`.

---

## Footer rendering pattern

Footer strings are generated by keymap methods, not hardcoded:

```go
keymap.DefaultList.Footer()          // "↑ navigate · ↓ navigate · enter select"
keymap.DefaultGlobal.Help.Help       // "? help"
keymap.DefaultGlobal.Back.Help       // "esc back"
keymap.DefaultQuiz.Unrevealed()      // "enter reveal · tab inverse"
keymap.DefaultQuiz.Revealed()        // "1 again · 2 hard · 3 good · 4 easy · tab inverse"
keymap.DefaultTypingQuiz.Footer()    // "enter submit · ctrl+r reveal · tab inverse"
keymap.DefaultDecks.Footer()         // "↑ navigate · ↓ navigate · space toggle · a toggle all · enter select"
```

Screens compose them with ` · ` to build complete footers. `ToggleParse` is a
typing-quiz-only binding: `ctrl+p` switches accent-trigger expansion on and off.

---

## What keymap enables

1. **Single source of truth** — Change a key in one place, every screen updates
2. **Configurable keys** — User overrides from `~/.config/crds/keymaps.yaml`
3. **Consistent footers** — Derived automatically from keymap definitions
4. **Vim bindings** — `"k"` for up, `"j"` for down already in `DefaultList`
5. **Discoverability** — `Registry.Bindings()` iterates all bindings for help overlay
6. **Type safety** — Compile-time check that referenced keymaps exist

---

## Integration

- **`internal/config/`** — `LoadKeymapConfig()` reads `keymaps.yaml` and returns a `*KeymapConfig`
- **`internal/ui/app/app.go`** — `New()` calls `config.LoadKeymapConfig()` then `keymap.ApplyDefaultOverrides()`
- **`internal/ui/app/view.go`** — `renderHelpOverlay()` uses `keymap.DefaultRegistry.Bindings()` to show all shortcuts grouped by category
- **`internal/ui/screens/*.go`** — All 8 screens use `keymap.Default*` for key dispatch and footer generation

## Notes for changes

- Chord bindings (e.g. `g` then `g` for top of list), mouse bindings, and
  per-screen keymap overrides are planned — tracked in `docs/roadmap.md`.
