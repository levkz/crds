# Styles Context

This document provides implementation context for future development sessions.

---

# Purpose

The `styles` package provides shared Lip Gloss style definitions for the terminal UI.

Every style in this package is:

- **Small** - One function, one return value
- **Stateless** - No internal state, calls ui.Theme on each render
- **Reusable** - Imported by components throughout the application
- **Semantic** - Uses ui.Theme.Primary, ui.Theme.Success, etc. instead of hardcoded colors

The goal is consistency across the entire UI - all visual primitives are defined in one place.

---

# Philosophy

1. **Never hardcode colors** - Always use semantic theme values:

```go
ui.Theme.Primary  // Instead of lipgloss.Color("#0000ff")
ui.Theme.Success  // Instead of lipgloss.Color("#00ff00")
```

2. **One responsibility per file** - Each function has one purpose
3. **Early returns** - Simple logic flow
4. **Reuse strings.Builder** - Avoid unnecessary allocations
5. **Keyboard first** - Every style is part of a keyboard-navigable UI

---

# Current Status

**All 12 style definitions are implemented** (60 tests):

| Style | Function | Source |
|---|---|---|
| Header | `Header(width int)` | `header.go` |
| Footer | `Footer(width int)` | `footer.go` |
| Selected item | `SelectedItem()` | `selected_item.go` |
| Focused input | `FocusedInput()` | `focused_input.go` |
| Error | `Error()` | `error.go` |
| Warning | `Warning()` | `warning.go` |
| Success | `Success()` | `success.go` |
| Hint | `Hint()` | `hint.go` |
| Muted text | `MutedText()` | `muted_text.go` |
| Card | `Card(width int)` | `card.go` |
| Panel | `Panel(width int)` | `panel.go` |
| Modal | `Modal(width, height int)` | `modal.go` |

**Dependencies:**

- `github.com/charmbracelet/lipgloss`
- `crds/internal/ui` (theme re-exports)
- `crds/internal/ui/theme` (BorderRole constants)

No dependency on Bubble Tea, state management, or application logic.

---

# Design Goals

- **Responsive** - Styles must adapt to theme changes
- **Consistent** - Same visual language across all components
- **Efficient** - Minimal allocations, fast rendering
- **Maintainable** - Clear, simple functions
- **Testable** - Deterministic rendering without terminal dependency

---

# Style Reference

### Header(width int)
- Uses `ui.Theme.Header` (bold)
- Sets width and horizontal padding
- Used by `components.Header()`

### Footer(width int)
- Uses `ui.Theme.Muted` (gray foreground)
- Sets width and horizontal padding
- Used by `components.Footer()`

### SelectedItem()
- Uses `ui.Theme.Primary` (blue foreground) with `Background(Selection)`
- No parameters — chain `.Width(n)` if needed
- Used by `components.RenderList()`

### FocusedInput()
- Uses `lipgloss.NewStyle()` with `ui.Theme.BorderFor(BorderRoleCard)`, primary border foreground
- Border color from `ui.Theme.Primary.GetForeground()` — follows theme switches
- Padding for text input appearance
- Used by `screens/search.go`

### Error()
- Returns `ui.Theme.Danger` (red foreground)
- Pass-through — chain additional properties as needed

### Warning()
- Returns `ui.Theme.Warning` (orange foreground)
- Pass-through — chain additional properties as needed

### Success()
- Returns `ui.Theme.Success` (green foreground)
- Pass-through — chain additional properties as needed

### Hint()
- Returns `ui.Theme.Muted` with `Italic(true)` (gray italic)
- Used for help text, hints, secondary information

### MutedText()
- Returns `ui.Theme.Muted` (gray foreground)
- Used by `components.Text()` and `components.ProgressBar()`

### Card(width int)
- Uses `ui.Theme.Primary` (blue foreground) with width and padding
- Used by `components.RenderCard()`

### Panel(width int)
- Uses `lipgloss.NewStyle()` with `ui.Theme.BorderFor(BorderRoleContainer)`, border color from palette
- For grouping related content — sections, groups, containers

### Modal(width, height int)
- Uses `lipgloss.NewStyle()` with `ui.Theme.BorderFor(BorderRoleModal)`, primary border foreground
- Border color from `ui.Theme.Primary.GetForeground()` — follows theme switches
- Width, height, and padding for overlay dialogs
- Used by `components.RenderModal()`

---

# File Organization

```
styles/
├── CONTEXT.md
├── TODO.md
├── header.go          Header(width int)
├── footer.go          Footer(width int)
├── selected_item.go   SelectedItem()
├── focused_input.go   FocusedInput()
├── error.go           Error()
├── warning.go         Warning()
├── success.go         Success()
├── hint.go            Hint()
├── muted_text.go      MutedText()
├── card.go            Card(width int)
├── panel.go           Panel(width int)
├── modal.go           Modal(width, height int)
└── styles_test.go     60 tests
```

---

# Testing

All tests are in-package (`package styles`). 60 tests total:

- Each style has 3–5 sub-tests: render, parameter verification, color/border checks, chaining
- `TestAllStylesRender` — smoke test that all 12 styles render without panic
- `TestThemeSwitchUpdatesStyles` — verifies live theme switching propagates

Run:

```
go test ./internal/ui/styles/ -v -count=1
```

Coverage areas:

- Color application (semantic values)
- Width/height settings
- Padding and margins
- Bold/italic settings
- Border rendering
- Theme switching integration

---

# Integration into Application

Styles are used by components in `internal/ui/components/`:

```go
// components/header.go
func Header(title string) string {
    return styles.Header(60).Render(title)
}
```

When the theme is switched at runtime (via Settings screen), all styles
automatically update because they call `ui.Theme` on every render.

Currently wired consumers:

| Consumer | Style Used |
|---|---|
| `components.Header()` | `styles.Header(60)` |
| `components.Footer()` | `styles.Footer(60)` |
| `components.RenderCard()` | `styles.Card(60)` |
| `components.ProgressBar()` | `styles.MutedText()` |
| `components.RenderList()` | `styles.SelectedItem()` |
| `components.RenderModal()` | `styles.Modal(40, 10)` |
| `components.RenderNotification()` | `styles.Hint()` |
| `components.Text()` | `styles.MutedText()` |

---

# Suggestions

1. **Icon-aware markers** — Some styles (SelectedItem, Error, Success) could
   expose icons from `ui.Theme.Icons` to respect NerdFont/Emoji/Unicode.

2. **Width parameter consistency** — All "full-width" styles (Header, Footer,
   Card, Panel) now take `width int`. If screens later need dynamic widths
   from terminal size, the parameter stays the same — callers pass
   `t.Width` or similar.

3. **FocusedInput variants** — Consider `FocusedInput(active bool)` to toggle
   between focused/unfocused border colors.

4. **Style composition helpers** — If patterns like "bold + primary" or
   "italic + muted" appear frequently, consider helper functions that
   compose base styles.

---

## TODOs

- [ ] Add icon-aware marker helpers — e.g., `SelectedItemIcon()` that returns
      `ui.Theme.Icons.Bullet` or similar for the selected item marker
- [ ] Add `FocusedInput(active bool)` variant for unfocused state
- [ ] Add style composition functions if duplication emerges across consumers
- [ ] Consider `testdata/` directory if parameterized test fixtures become
      useful (currently not needed — all styles are pure functions)
