# Renderer TODO

## Phase 1 — Core utilities (✅ complete)

- [x] `VisibleWidth(s string) int` — display width (ANSI-aware via `go-runewidth`)
- [x] `LineWidth(s string) int` — alias for VisibleWidth
- [x] `MaxLineWidth(text string) int` — widest line in multi-line text
- [x] `TextDimensions(text string) (width, height int)` — terminal cell dimensions
- [x] `StripANSI(s string) string` — strip CSI + C1 control codes
- [x] `CountANSISequences(s string) int` — count valid escape sequences
- [x] `Wrap(text string, maxWidth int) []string` — word-wrap with hard breaks
- [x] `Truncate(text string, maxWidth int) string` — fit width + trailing ellipsis
- [x] `Fit(text string, maxWidth int) string` — hard truncation

## Phase 2 — Viewport management

- [ ] `TerminalState` struct with `Width`/`Height` and `Update(tea.WindowSizeMsg)`
- [ ] Scrolling helpers (offset, line/page navigation)
- [ ] Viewport-aware truncation

## Phase 3 — Advanced rendering

- [ ] Efficient re-rendering / diffing
- [ ] Component-aware character counting
- [ ] Performance profiling utilities

## Phase 4 — UI layer integration

- [ ] Screen rendering optimization via renderer utilities
- [ ] Component measurement helpers
- [ ] Viewport-aware layout calculations
