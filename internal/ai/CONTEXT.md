# AI Agent Context

> Per-package context: how this package works today. Status and plans live in
> `docs/status.md` and `docs/roadmap.md` (see `docs/README.md`).

## Purpose

Provide a keyless-friendly LLM backend for the `crds ai` commands (interpret,
fill, add). The package never writes to storage itself — the CLI layer parses
and validates everything it returns before any append happens.

## Responsibilities

- Provider presets and config resolution (env > `config.yaml` `ai:` block > preset).
- An OpenAI-compatible chat client (`POST /v1/chat/completions`) with a single
  `Complete(ctx, system, user)` call.
- Prompt construction for the two agent modes:
  - **Interpret**: unstructured text (words, phrases, `term = translation`) → YAML entries.
  - **Fill**: partial YAML entries → completed entries (examples, notes, tags,
    CRDS variant notation), constrained by deck context.
- Parsing agent output into `[]model.Entry`, applying CRDS conventions
  (translation-ish "term" direction, no duplicates, ≥1 translation).

## Key files

| File | Purpose |
|---|---|
| `config.go` | `Config` struct, 7 provider presets, `Resolve()` precedence |
| `client.go` | `Client` struct + `NewClient`, raw HTTP POST, error wrapping |
| `prompts.go` | `InterpretMessages`, `FillMessages`, deck-context block |
| `agents.go` | `Interpret`, `Fill`, `IsStructuredInput`, `LanguageContext`, `DeckContext` |
| `parse.go` | `ParseEntries` — YAML→`[]model.Entry` with validation |
| `PLAN.md` | Implementation plan (needed if modified, keep in sync with roadmap) |

## Dependencies

- `crds/internal/model` for `Entry`.
- Standard library only (`net/http`, `encoding/json`, `os`, `strings`, `gopkg.in/yaml.v3` re-exported via `crds/internal/parser`).
- YAML comes from `go.yaml.in/yaml/v3` via the parser package — do not add a second YAML dependency.

## Integration

- `internal/config/config.go` exposes an `ai:` block (`AIConfig` on `ConfigYAML`).
- `internal/cli/ai.go` wires the three commands. The `resolveAIClient` variable
  is the test seam: tests override it with a fake `ai.Client`.
- `internal/storage/store.go` provides `ListDeckTags(deckID)` and
  `AppendEntries(deckID, entries, deckDir)` used by the CLI layer.

## Notes for changes

- `Resolve()` must be called before `NewClient` to apply preset defaults; the
  CLI does this inside `resolveAIClient()`.
- Agent output is arbitrary text from a model — always parse and validate
  before using. `ParseEntries` enforces the deck schema invariants (term,
  translations, no duplicate IDs).
- Prompts are exercises in wording; the fill prompt must never invent a new ID
  (`Do not set the id field`) so appends can auto-ID on sync.
- Keep `presets` in `config.go` and the `crds ai --help` text in sync.