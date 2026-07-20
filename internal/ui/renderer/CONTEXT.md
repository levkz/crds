# Renderer Context

## Purpose

The `renderer` package provides rendering utilities for terminal applications.
It complements the UI layer by handling technical rendering concerns:
- Terminal width/height calculations
- Text wrapping and measurement
- ANSI escape sequence processing
- Unicode width handling for multi-byte characters
- Viewport management and scrolling
- Performance-optimized rendering helpers

These utilities sit below the UI layer (composing views) but above the raw terminal interface.

---

## Current state

The `renderer` package is currently a placeholder – a directory with only `TODO.md` present. All rendering responsibilities are currently handled elsewhere:

- **Bubble Tea** – Text rendering, terminal event handling
- **lipgloss** – Styling, borders, padding, text positioning
- **Components** – Domain-specific rendering helpers
- **Screens** – View composition using `strings.Builder`

No dedicated rendering package exists yet.

---

## Dependencies

Assuming the `renderer` package will be utilized:
- **github.com/charmbracelet/lipgloss** – ANSI formatting, width/height measurement
- **github.com/mattn/go-runewidth** – Unicode character width detection
- **github.com/charmbracelet/lipgl** (if exists) – Terminal-aware rendering
- **github.com/charmbracelet/bubbletea** – Viewport management (theoretically)

---

## Integration points

### With UI layer

The renderer would integrate with the `ui/` package at these points:

- **Screen rendering** – `ui.Screen.View()` methods could delegate to renderer utilities
- **Text measurement** – Determining text width for dynamic layout
- **Character counting** – Accurate counting of displayed characters (i.e., actual width calculation for large view states)
- **Cursor positioning** – Precise placement of cursor overlay/manipulation
- **Viewport management** – Handling truncated content
- **Line height calculations** – Terminal-aware vertical spacing

### With components

Rendering utilities would support component rendering:

- **Text measurement** – `MeasureText(text) -> width`
- **Character counting** – `CountVisibleChars(text) -> count`
- **Line breaking** – `WrapText(text, maxWidth) -> []string`
- **ANSI handling** – Strip/color handling for terminal output

---

## Testing

Rendering utilities should be tested deterministically:

- **Character width** – Fixed tests for multi-byte Unicode characters
- **Text wrapping** – Consistent output regardless of terminal width
- **ANSI handling** – Predictable escape sequence stripping
- **Viewport calculations** – Edge cases and boundary conditions

Tests should mock terminal width/height without actual Bubble Tea dependency.

---

## API design (proposed)

```go
package renderer

// Character width
import "github.com/mattn/go-runewidth"
func VisibleWidth(s string) int
func VisibleCount(s string) int

// Text wrapping
func Wrap(text string, maxWidth int) []string
func LineWidth(s string) int

// ANSI handling
func StripANSI(s string) string
func CountANSISequences(s string) int

// Viewport
func Truncate(text string, maxWidth int) string
func Fit(text string, maxWidth int) string

// Terminal state
type TerminalState struct {
    Width  int
    Height int
}
func (ts *TerminalState) Update(msg tea.WindowSizeMsg) { ... }

// Layout helpers
func MaxLineWidth(text string) int
func TextDimensions(text string) (width, height int)
```

---

## Known issues

- **No dedicated package** – All rendering is scattered across existing packages
- **Character width bugs** – Many packages handle Unicode incorrectly
- **ANSI complexity** – Escape sequences for colors, styles, cursor control
- **Performance concerns** – Text measurement and wrapping should be cacheable

---

## Related directories

```
ui/
├── docs/              UI documentation
├── screen.go         Screen interface
├── theme/            Design system (colors, icons, typography)
├── styles/           Semantic style definitions (Header, Footer, Card, etc.)
├── components/       Reusable components (Header, Footer, Card, etc.)
├── screens/          Screen implementations
├── keymap/           Centralized keyboard handling
├── navigation/       Navigation stack management
├── app/              Root Bubble Tea model
├── layout/           Structural primitives (Column, Row, Grid, Spacer, etc.) ← Just completed
└── renderer/         [This package - to be implemented]
```

The renderer would complement these existing packages, focusing on terminal-specific rendering utilities that enhance existing functionality without duplicating responsibilities.

---

## Best practices

### Performance

- Cache frequently called calculations (text width, line counts)
- Minimize allocations in hot paths
- Use incremental updates for dynamic content

### Correctness

- Use standards-compliant Unicode width handling
- Handle ANSI escape sequences correctly
- Account for terminal size changes

### Testability

- Mock terminal state without actual terminal dependency
- Isolate ANSI processing from layout logic
- Provide test utilities for edge cases

---

## Historical context

Rendering has evolved in this codebase:

- **Early days** – Manual string concatenation with `"\\n"`
- **lipgloss integration** – Semantic styling with `Header(60)`, `Card(60)`, etc.
- **Layout primitives** – `Column()`, `Page()`, `Center()` helpers
- **Responsive design** – Width propagation from root model

The next step is to consolidate terminal-aware rendering logic that these pieces currently lack.

---

## Future development

### Phase 1: Core utilities

Implement building blocks:
- Character width/measurements
- Text wrapping and truncation
- ANSI escape sequence stripping

### Phase 2: Viewport management

Add terminal-awareness:
- Terminal size tracking
- Scrolling and viewport offsets
- Line and page navigation

### Phase 3: Advanced rendering

Enhance existing functionality:
- Efficient re-rendering logic
- Component-aware character counting
- Performance profiling utilities

### Phase 4: Integration

Enhance UI layer:
- Screen rendering optimization
- Component measurement utilities
- Viewport-aware layout calculations

---

## Success criteria

The `renderer` package is complete when it:

1. **Replaces** hardcoded width calculations with accurate terminal measurements
2. **Supports** Unicode correctly in all rendering contexts
3. **Automates** text wrapping where appropriate (e.g., long entries in detail view)
4. **Handles** ANSI sequences correctly for dynamic styling
5. **Works** with any terminal size (testing success criteria)
6. **Integrates** cleanly with existing UI, components, and screens
7. **Is fully testable** without Bubble Tea dependencies
8. **Improves performance** for rendering-heavy operations
9. **Is documented** with examples and edge case handling
10. **Follows** the codebase's style and architecture

Minimal viable implementation focuses on Phase 1 (core utilities) that immediately improve existing rendering correctness and reduce duplication.

---

## Dependencies

`github.com/mattn/go-runewidth` is the recommended approach for Unicode width handling. The `github.com/charmbracelet/lipgloss` package already uses this for width calculations in styles.

---

## Testing approach

```bash
go test ./internal/ui/renderer/ -v -count=1
```

## Goal state

Rendering should become a first-class concern in the UI architecture, addressing terminal-specific challenges that currently require workaround patterns in screens and components.

The goal is to reduce manual character counting, improve text layout correctness, and make terminal-aware rendering more consistent across the application.

---

## Summary

The `renderer` package aims to solve terminal-specific rendering challenges that the current architecture currently handles inconsistently or manually. By consolidating these concerns, the UI layer can become more focused on composition and interaction, while the renderer handles the technical details of terminal output.

The implementation should follow the codebase's philosophy of:
- Small, focused functions
- Deterministic testing
- Early returns and error handling
- Avoiding premature optimization
- Leveraging dependency injection where appropriate