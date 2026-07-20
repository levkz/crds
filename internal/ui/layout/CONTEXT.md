# Layout Context

## Purpose

The `layout` package provides structural primitives for composing screen
views. It sits between `styles/` (lipgloss style functions) and
`components/` (domain-specific rendering helpers), filling gaps that
neither currently addresses: spacing, alignment, page structure, and
width awareness.

The implementation achieves:
- **Page wrapper** – eliminate `strings.Builder` + `"\\n\\n"` pattern across all screens
- **Column/Row/Grid/Stack** – vertical/horizontal composition utilities
- **Spacer/VSpace/HSpace** – spacing helpers consuming `theme.Spacing` scale
- **Center/AlignLeft/AlignRight** – text alignment utilities
- **Width propagation** – screen size tracking from `tea.WindowSizeMsg` → all components
- **Responsive styling** – all components accept `width`/`height` parameters

---

## Current layout situation

Before implementation, the codebase used `strings.Builder` + raw `"\\n\\n"`
concatenation in every screen. The pattern was duplicated across all 6
screens (`home.go`, `quiz.go`, `search.go`, `settings.go`, `statistics.go`,
`detail.go`).

### Hardcoded width `60` everywhere

Previously, the literal `60` appeared in:

| File | Usage |
|---|---|
| `components/header.go:6` | `styles.Header(60).Render(title)` |
| `components/footer.go:6` | `styles.Footer(60).Render(keys)` |
| `components/card.go:16` | `styles.Card(60)` |
| `components/modal.go:6` | `styles.Modal(40, 10)` |
| `screens/statistics.go:42` | `styles.Panel(60)` |
| `screens/detail.go:41,48,55` | `styles.Card(60)` (three times) |

The root `app/model.go` tracked `Width int` and `Height int` from
`tea.WindowSizeMsg`, but these were **never passed to screens**.

### Theme spacing was unused

`theme.Spacing` defined a 7-tier scale (`Xxs=2`, `Xs=4`, `Sm=8`, `Md=16`,
`Lg=24`, `Xl=32`, `Xxl=48`) but was completely unused – every screen
used raw `"\\n\\n"` instead.

### Footer separator was inlined

Three screens concatenated keymap footers with `" · "`:
```go
keymap.DefaultList.Footer() + " · " + keymap.DefaultGlobal.Help.Help
```

### List indentation duplicated

`components/list.go` and `screens/settings.go` each implemented the same
selected/unselected item marking pattern using `"  "` and
`ui.Theme.Icons.Navigate + " "`.

---

## Current implementation

### Layout primitives

Available in `internal/ui/layout/`:

| Function | Purpose |
|---|---|
| `Page(header, body, footer)` | Horizontal composition (Header, Content, Footer) with spacing |
| `Column(items...)` | Vertical composition using `\\n\\n` separators |
| `Row(items...)` | Horizontal composition using `lipgloss.JoinHorizontal` |
| `Grid(items, cols)` | Multi-column layout with automatic wrapping |
| `Stack(layers...)` | Center-overlap layering (for modals/overlays) |
| `VSpace(n) / HSpace(n)` | Vertical/horizontal spacing helper |
| `Center(text, width)` | Center text within given width |
| `AlignLeft(text, width) / AlignRight(text, width)` | Text alignment utilities |

All functions are deterministic and use table-driven tests without Bubble Tea dependencies.

### Width propagation (responsive sizing)

**Screen interface change:**
```go
type Screen interface {
    Init() tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)
    View() string
    SetSize(w, h int) // added
}
```

**All 6 screens:**
- Store `width`/`height` (default 60/24)
- Implement `SetSize(w, h int)`
- Pass `m.width` to components instead of hardcoded `60`
- Components (`Header`, `Footer`, `RenderCard`, `RenderModal`) accept `width`/`height`

**Root model:**
```go
case tea.WindowSizeMsg:
    m.Width = msg.Width
    m.Height = msg.Height
    if screen, ok := m.Navigator.CurrentScreen(); ok {
        screen.SetSize(msg.Width, msg.Height)
    }
```

### Architectural relationship

```
theme/  ──>  styles/  ──>  components/  ──>  screens/
                    \\               ^
                     v              |
                  layout/  ────────┘
                  (structural
                   helpers)
```

- `layout/` depends on `styles/` and `ui.Theme` for visual rendering
- `layout/` does **not** duplicate `components/` – components provide domain-specific rendering
- `layout/` does **not** duplicate `styles/` – style functions return `lipgloss.Style` values

---

## Current TODO.md status

| Item | Status |
|---|---|
| `Row` | ✅ Implemented |
| `Column` | ✅ Implemented |
| `Stack` | ✅ Implemented |
| `Grid` | ✅ Implemented |
| `Spacer` | ✅ Implemented (VSpace/HSpace) |
| `Padding` | ✅ Handled by `styles/` functions |
| `Margin` | ✅ Handled by `styles/` functions |
| `Alignment` | ✅ Implemented (Center, AlignLeft, AlignRight) |
| `Borders` | ✅ Handled by `styles/` via `ui.Theme.BorderFor` |
| `Centering` | ✅ Implemented (Center) |
| `Responsive sizing` | ✅ Implemented (SetSize on Screen + Page wrapper) |

---

## Dependencies

- `github.com/charmbracelet/lipgloss` – `JoinHorizontal`, `Place`, `Width`, `Height`
- `crds/internal/ui/theme` – `Theme`, `Spacing`
- `crds/internal/ui/styles` – style functions
- `crds/internal/ui` – `Screen` interface (with `SetSize`)

---

## Files

```
layout/
├── CONTEXT.md      This file
├── TODO.md         Planned items
├── page.go         Page(header, body, footer)
├── column.go       Column(items ...string)
├── row.go          Row(items ...string)
├── grid.go         Grid(items []string, cols int)
├── stack.go        Stack(layers ...string)
├── spacer.go       VSpace(n), HSpace(n)
├── center.go       Center(text string, width int)
├── align.go        AlignLeft(text string, width int), AlignRight(text string, width int)
└── layout_test.go  Tests
```

---

## Testing

All layout functions are deterministic – same input → same output:

```go
func TestPage(t *testing.T) {
    got := Page("header", "body", "footer")
    want := "header\\n\\nbody\\n\\nfooter"
    if got != want {
        t.Errorf("Page() = %q, want %q", got, want)
    }
}
```

Tests avoid Bubble Tea dependency – run at string-returning function level.

---

## CSS-inspired convenience

For future ease of use, these higher-level helpers could be added:

```go
// Apply semantic spacing levels
func Spacing(level theme.SpacingLevel, items ...string) string {
    return Column(VSpace(int(level)), items...)
}

// Compose reusable screen templates
func ActivityScreen(title, list, footer string) string {
    return Page(components.Header(title, m.width), list, components.Footer(footer, m.width))
}
```

---

## Future extensions

Potential enhancements within this architecture:
- **Mouse mode support** – via lipgloss input handling
- **Chord keybindings** – integrate with `keymap` package
- **Plugin widgets** – layout could be extended by plugins
- **Split views** – Grid improvements for sidebar layouts
