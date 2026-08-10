# Mapping Context

> Per-package context: how this package works today. Status and plans live in
> `docs/status.md` and `docs/roadmap.md` (see `docs/README.md`).

## Purpose

The `mapping` package resolves and applies **input mappings**: trigger sequences
that expand into special characters while typing answers, e.g. `e/` → `é` (a
Babbel-style compose convention). It lets users and deck creators define their
own per-language mappings without touching app code.

## Layers

Mappings come from three layers, later layers winning for the same key:

| Layer | Source |
|-------|--------|
| Built-in defaults | Embedded in the binary (`builtin` map), keyed by language code. Ships French (`fr`). |
| User files | `~/.config/crds/mappings/<lang>.yaml`, one YAML map per language. |
| Deck field | `input_mappings` in the deck YAML. |

`Store.Resolve(lang, deck)` merges the three and returns a `Mapping`.

## Semantics

`Mapping.Apply(input)` replaces the **longest key that is a suffix** of the
input, single pass. This gives a compose-style experience: typing `e` then `/`
turns `mange/` into `mangé`. Matching is rune-based, so multi-byte triggers
(`ß` from `ss`) work. Keys are sorted longest-first so the longest match wins.
Empty keys, empty replacements, and self-identity pairs are dropped at build
time.

`Mapping.ApplyAt(input, end)` applies the same rule to a rune-indexed slice
`input[:end]` and returns the new end position, so callers can expand triggers
at the current cursor — including mid-string — and keep the cursor on the
inserted replacement. Text before the cursor is never re-scanned. A replacement
is not re-scanned for further triggers.

## Key files

| File | Responsibility |
|------|----------------|
| `mapping.go` | `Mapping`, `New`, `Apply`, `ApplyAt`, `Pairs`, `Store`, `LoadDir`, `Resolve`, `Languages`, built-in defaults |
| `mapping_test.go` | Suffix matching (whole input and at a cursor index), precedence, file loading, error cases |

## Dependencies

- `go.yaml.in/yaml/v3` — parsing user mapping files.
- Standard library otherwise. No dependency on other internal packages, so no
  import cycles.

## Integration

- `internal/config` exposes `MappingsDir()` and creates `~/.config/crds/mappings/`.
- `internal/ui/app` loads the store at startup (`mapping.LoadDir`) and passes it
  into `screens.NewTypingQuiz`.
- `internal/ui/screens.TypingQuizModel` resolves the effective mapping for the
  current deck in `SyncState` and applies it to input in `handleInput`.
- `internal/model` + `internal/parser` + `internal/storage` carry the deck-level
  `input_mappings` and `language` through to `ui.DeckData`.

## Notes for changes

- Accent-insensitive matching lives in `internal/fuzzy` (approximate mode), not
  here — mappings help typing, `fuzzy` handles forgiveness on grading.
- The typing quiz expands triggers via `ApplyAt` at the cursor; `ctrl+p` toggles
  parsing off so a literal trigger (e.g. a real `e/`) can be typed, then back on
  for newly typed text.
- Built-in defaults for more languages can be added to the `builtin` map.
