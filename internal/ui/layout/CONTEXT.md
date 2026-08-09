# Layout Context

> Per-package context: how this package works today. Status and plans live in
> `docs/status.md` and `docs/roadmap.md` (see `docs/README.md`).

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
`detail.go`). This has since been replaced with `layout.Page()` and
`layout.Column()`.

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

## Implemented primitives

Row, Column, Stack, Grid, Spacer (VSpace/HSpace), Alignment (Center,
AlignLeft, AlignRight), Centering, and Responsive sizing (SetSize on Screen +
Page wrapper). Padding, Margin, and Borders are handled by `styles/` via
`ui.Theme.BorderFor`.

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
