# Renderer Context

> Per-package context: how this package works today. Status and plans live in
> `docs/status.md` and `docs/roadmap.md` (see `docs/README.md`).

## Purpose

The `renderer` package provides terminal-aware text rendering utilities:
Unicode-aware width measurement, ANSI escape sequence handling, and word
wrapping/truncation. It sits below the UI composition layer (screens,
components) and above the raw terminal, filling gaps that lipgloss does not
cover directly.

## Responsibilities

- **Width measurement** — visible character width of strings with ANSI
  sequences and multi-byte/CJK/wide characters stripped (via `go-runewidth`)
- **ANSI processing** — strip escape sequences and count them
- **Wrapping/truncation** — wrap text at word boundaries (with hard-wrapping
  of overlong words), truncate or fit to a width with an ellipsis

## Key files

```
renderer/
├── CONTEXT.md       This file
├── width.go         VisibleWidth, LineWidth, MaxLineWidth, TextDimensions
├── ansi.go          StripANSI, CountANSISequences
├── wrap.go          Wrap, Truncate, Fit, hardWrap
└── renderer_test.go Table-driven tests
```

## API

- `VisibleWidth(s string) int` — width of `s` after stripping ANSI, using
  `runewidth.StringWidth`
- `LineWidth(s string) int` — alias of `VisibleWidth`
- `MaxLineWidth(text string) int` — max visible width across newline-split lines
- `TextDimensions(text string) (width, height int)` — max line width + line count
- `StripANSI(s string) string` — removes CSI sequences, two-char C1 controls,
  and bare `\x1b` escapes
- `CountANSISequences(s string) int` — counts CSI and C1 control sequences
- `Wrap(text string, maxWidth int) []string` — word-wrap; overlong words are
  hard-wrapped; `maxWidth < 1` is treated as 1
- `Truncate(text string, maxWidth int) string` — cut to `maxWidth` with "…"
- `Fit(text string, maxWidth int) string` — cut to `maxWidth` with no ellipsis

## Dependencies

- `github.com/mattn/go-runewidth` — Unicode character width detection
- stdlib `strings` — splitting, building

No Bubble Tea or lipgloss dependency.

## Integration

- Used by screens and view composition where accurate display width or text
  fitting is needed (e.g. wrapping long entry content, fitting rows).
- `StripANSI` underpins `VisibleWidth`, so every width computation is
  ANSI-safe.

## Notes for changes

- Keep functions pure and deterministic; add table-driven cases in
  `renderer_test.go` alongside new behavior.
- Do not introduce terminal-size state here — `TerminalState`-style tracking
  is a roadmap item, not implemented.
