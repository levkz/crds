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
- Prompt construction for the three agent modes:
  - **Interpret (minimal)**: unstructured text (words, phrases, `term = translation`) → bare YAML entries.
  - **Interpret (full)**: same input → complete entries (≥4 example uses, notes, deck-constrained tags) in one call.
  - **Fill**: partial YAML entries → completed entries, constrained by deck context.
- A shared `termConventions` block (CRDS variant notation) used by both the
  fill prompt and the full-effort interpret prompt.
- A shared `tagRules` block: structural tags (noun, verb, adjective, gender,
  verb class) and one CEFR proficiency tag (A1–C2) are always allowed; theme
  tags come from the deck allowlist (`allowed theme tags:`), or a concise
  model-chosen theme tag when no deck is supplied.
- An optional `msg` passthrough on every agent call: an extra instruction from
  `--msg` appended to the user prompt.
- Language pair is filled from the deck, then overridden by `-F/--translate-from`
  and `-T/--translate-to` (flags win), so the same prompts work with or
  without a deck.

## Key files

| File | Purpose |
|---|---|
| `config.go` | `Config` struct, 7 provider presets, `Resolve()` precedence |
| `client.go` | `Client` struct + `NewClient`, raw HTTP POST, error wrapping |
| `prompts.go` | `InterpretMessages`, `InterpretFullMessages`, `FillMessages`, deck-context block, `termConventions` |
| `agents.go` | `Interpret`, `InterpretFull`, `Fill`, `IsStructuredInput`, `LanguageContext`, `DeckContext` |
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