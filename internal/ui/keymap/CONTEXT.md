# Keymap Context

## Purpose

The `keymap` package centralizes keybinding definitions so screens and the
root model never hardcode key strings. Each keymap is a group of `Binding`
values — each holding the keys that trigger it and a help label.

---

## Current keyboard handling (without keymap)

Keys are currently handled in three layers with no central definition:

### Layer 1 — Global (app/events.go)

```go
func (m Model) dispatchKeyEvent(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch msg.String() {
    case m.Config.KeyQuit:   // "ctrl+c"
    case m.Config.KeyHelp:   // "?"
    case "esc":
        // dismiss overlay → go back → home
    }
    return m.forwardToScreen(msg)
}
```

Only `KeyQuit` and `KeyHelp` are configurable (via `app.Config`).
`esc` is hardcoded.

### Layer 2 — Screen-local key dispatch

Each screen's `Update` switches on `tea.KeyMsg`:

| Screen     | Keys                                          |
|------------|-----------------------------------------------|
| Home       | `up`/`k`, `down`/`j`, `enter`                |
| Quiz       | `enter`, `1`, `2`, `3`, `4`                  |
| Search     | `up`/`k`, `down`/`j`, `enter`, `backspace`, `tab`, printable chars |
| Settings   | `up`/`k`, `down`/`j`, `enter`                |
| Statistics | — (no input)                                  |
| Detail     | — (no input)                                  |

Pattern `"up", "k"` / `"down", "j"` repeats across three screens.

### Layer 3 — Footer help text

Every screen constructs a help string and passes it to `components.Footer()`:

- Home: `"↑/↓ navigate · enter select · ? help"`
- Quiz: `"Enter Reveal"` / `"1 Again   2 Hard   3 Good   4 Easy"`
- Search: `"type to search · ↑/↓ navigate · enter open · esc back"`
- Settings: `"↑/↓ navigate · enter select · esc back"`

These strings are hardcoded inline in each screen's `View()` method.

---

## Keymap package design

### Binding

```go
type Binding struct {
    Keys []string   // e.g. {"up", "k"}
    Help string     // e.g. "↑" — used in footer
}
```

Bindings hold all alternate keys (e.g. `up` + `k`) for the same action.

### Predefined keymaps

| Variable          | Fields                                                   |
|-------------------|----------------------------------------------------------|
| `DefaultGlobal`   | `Quit` (`ctrl+c`), `Help` (`?`), `Back` (`esc`)         |
| `DefaultList`     | `Up` (`up`/`k`), `Down` (`down`/`j`), `Select` (`enter`)|
| `DefaultQuiz`     | `Reveal` (`enter`), `Again`/`Hard`/`Good`/`Easy` (`1`–`4`) |
| `DefaultSearch`   | embeds `List` + `FocusToggle` (`tab`), `Select` (`enter`), `DeleteChar` (`backspace`) |

### Usage from a screen

```go
import "crds/internal/ui/keymap"

func (m HomeModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case keymap.DefaultList.Up.Keys[0],
             keymap.DefaultList.Up.Keys[1]:
        }
    }
}
```

Or with a helper:

```go
if msg.In(keymap.DefaultList.Up) { ... }
```

---

## Footer rendering pattern

Each keymap can produce a help string for the current screen:

```go
func (km Global) Footer() string {
    return BindingList{
        km.Quit, km.Help,
    }.Help()
}
```

Screens combine keymaps:

```go
footer := keymap.DefaultList.Footer() + " · " + keymap.DefaultGlobal.Help.Help
```

---

## What keymap enables

1. **Single source of truth** — Change `"k"` to `"K"` in one place
2. **Configurable keys** — Load user key overrides from config
3. **Consistent footers** — Derive from keymap instead of hardcoding strings
4. **Vim bindings** — Add `"j"` for down, `"k"` for up centrally
5. **User-defined bindings** — One config section to remap any action

---

## Future work

- [ ] Add `Matches(msg tea.KeyMsg) bool` helper on Binding
- [ ] Add `BindingList` with `Help() string` for automatic footer generation
- [ ] Wire keymaps into screens — replace inline `case "up", "k":` with
      `case keymap.DefaultList.Up.Matches(msg):`
- [ ] Replace `app.Config.KeyQuit`/`KeyHelp` with keymap.Global reference
- [ ] Add `KeymapConfig` struct for user overrides in YAML
- [ ] Support chord bindings (e.g. `g` then `g` for top of list)
