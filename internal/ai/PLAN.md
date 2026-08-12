# AI Agent — Implementation Plan

> Working plan for the `crds ai` feature. This file lives in the package folder
> until the feature is implemented; once shipped, completion status folds into
> `docs/status.md` and the plan is archived in `docs/roadmap.md`. It describes
> what *will* be built — not the current state of the code.
>
> For the documentation taxonomy ("where a fact lives") see `docs/README.md`.

---

## Goal

Give the user an AI agent that turns a list of words (French → English, or any
deck language pair) into fully-formed vocabulary entries matching the CRDS YAML
format. Given an input word list with optional partial data (translations, tags,
examples, notes), the agent:

- fills in or creates missing data (example sentences with translations, notes,
  and tags),
- picks tags **only from the deck's existing tag list**,
- improves the term and translations using the CRDS variant syntax
  (see `docs/DECK_CREATION_GUIDE.md`), e.g.:
  - feminine noun → `(une/la) <word>` or `(une /l')` before vowel-initial words,
  - masculine noun → `(un/le) <word>` or `(un /l')`,
  - verb → English translation prepended with `(to)`,
  - conjugations compacted with `[er/e/ons]` style groups.

Two input modes:

1. **Unstructured** — the user writes whatever they want (one word per line,
   "manger, to eat", mixed notes). A "middleman" (interpreter) agent converts the
   free text into correct YAML entry structure.
2. **Structured** — the user passes correct (possibly partial) YAML entries
   directly to the filler agent.

---

## Non-goals

- No new YAML fields or schema changes. The agent produces the existing
  `model.Entry` shape (`id`, `term`, `translations`, `examples`, `tags`, `notes`).
- No TUI integration. This is a pure CLI feature; the `fillBackground()` ANSI
  gotcha (`app/view.go`) does not apply.
- No on-device fine-tuning, no chat loop, no interactive conversation with the
  model. One prompt → one response per agent call.
- No AI-generated grammar hints or auto-quizzes — those remain stretch ideas in
  `docs/roadmap.md`.

---

## User-facing behavior

```
crds ai interpret [--deck <deck>] [--text "..."] [--file -] [--dry-run]
    Unstructured text -> structured YAML entries, printed to stdout.

crds ai fill <deck> [--text "..."] [--file -] [--dry-run]
    Structured (partial) YAML entries -> completed YAML entries, printed to stdout.

crds ai add <deck> [--text "..."] [--file -]
    Auto-detect input mode, run interpret (if unstructured) then fill, print the
    result, and offer [a]ppend / [e]dit / [d]iscard before writing to the deck.
```

Input sources follow the existing CLI conventions in `internal/cli/`:
`--text` inline, `--file -` for stdin, or `$EDITOR` (via `internal/editor`)
when neither is given. `--dry-run` prints the assembled prompt + input without
calling the API — a zero-cost way to iterate on prompts.

---

## LLM providers

All supported providers expose the OpenAI Chat Completions wire format
(`POST {base_url}/chat/completions`), so a single stdlib HTTP client covers
every option. `base_url`, `model`, and `api_key` are configurable; each
provider has a baked-in preset.

| Provider | Base URL | Default model | API key | Notes |
|---|---|---|---|---|
| `pollinations` | `https://text.pollinations.ai/openai` | `openai` | — | **Default.** Keyless, anonymous. Rate limit ~1 req / 15 s. |
| `ollama` | `http://localhost:11434/v1` | `llama3.2` | — | Local, unlimited, private. |
| `openai` | `https://api.openai.com/v1` | `gpt-4o-mini` | required | |
| `gemini` | `https://generativelanguage.googleapis.com/v1beta/openai/` | `gemini-2.5-flash` | required | Google AI Studio key; ~1,500 req/day free tier. |
| `openrouter` | `https://openrouter.ai/api/v1` | `meta-llama/llama-3.3-70b-instruct:free` | required | Free models via `:free` suffix. |
| `groq` | `https://api.groq.com/openai/v1` | `llama-3.3-70b-versatile` | required | Very fast LPU inference. |
| `nvidia` | `https://integrate.api.nvidia.com/v1` | `meta/llama-3.3-70b-instruct` | required | 120+ open-weight models. |

**Free-tier landscape (context):** Ollama is the only *truly* keyless + private
option; Pollinations is the only hosted *no-signup* option. Free API keys (no
credit card) exist for Gemini, Groq, OpenRouter, NVIDIA NIM, Mistral, Cerebras,
Cloudflare Workers AI, and GitHub Models. The provider preset + env-override
design means users can switch between all of them without code changes.

---

## Config schema

`~/.config/crds/config.yaml` gains an `ai:` block (`internal/config/config.go`,
`ConfigYAML`):

```yaml
ai:
  provider: pollinations   # pollinations|ollama|openai|gemini|openrouter|groq|nvidia
  model: ""                # optional; overrides the provider default
  api_key: ""              # ignored for pollinations/ollama
  base_url: ""             # optional; overrides the provider default entirely
```

Resolution precedence: **env var > config.yaml > provider preset**.

Env vars: `CRDS_AI_PROVIDER`, `CRDS_AI_MODEL`, `CRDS_AI_API_KEY`,
`CRDS_AI_BASE_URL`.

The API key must never be logged, printed, or committed.

---

## Architecture

### Package layout (`internal/ai/`)

| File | Responsibility |
|---|---|
| `config.go` | Resolve `ai.Config` (provider presets, config.yaml, env overrides). Preset table lives here. |
| `client.go` | OpenAI-compatible chat client: `Complete(ctx, system, user) (string, error)` over stdlib `net/http`. Parses `choices[0].message.content`, surfaces HTTP/JSON errors. |
| `prompts.go` | `InterpretPrompt(...)` and `FillPrompt(...)` builders. The only place with model-facing instructions. |
| `parse.go` | Extract YAML from the model reply (strip markdown fences), unmarshal into `[]model.Entry`, validate. |
| `agents.go` | `Interpret(...)` / `Fill(...)` orchestrators: prompt → client → parse. |

### Client seam

```go
// A Provider is anything that can answer a chat completion. HTTP is behind
// this so CLI tests can inject a fake and never hit the network.
type Client interface {
    Complete(ctx context.Context, system, user string) (string, error)
}
```

`config.go` returns the concrete `*client` (or a constructor func); tests and
`--dry-run` substitute a fake.

### Prompt design

- **Interpreter** (`InterpretPrompt`): takes the deck's `language` /
  `translation_language` (when a deck is given) plus the raw free text.
  Output contract: a YAML list of entries, each with at least `term` and as much
  of `translations`/`tags`/`notes` as the input implied. Strict YAML, no prose.
- **Filler** (`FillPrompt`): takes the deck metadata, the partial entries, the
  **existing tag list** (from `Store.ListDeckTags`), and 2–3 sample entries from
  the deck for style. Instructs:
  - keep every field the user provided (fix only obvious typos),
  - add 2–3 example sentences in the source language, each with a translation,
  - choose tags **exclusively from the provided allowlist**,
  - apply the variant-syntax conventions (see Goal) — French gender articles
    `(une/la)`, `(un/le)`, verbs `(to) …`, conjugations `[er/e/ons]`,
  - output only the YAML list (no markdown fences, no commentary).
  The prompt is in prose and lives in Go, so it is testable and versionable.

### Output validation

The model reply is parsed defensively:

1. Trim; strip a single ` ```yaml … ``` ` fence if present.
2. `yaml.Unmarshal` into `[]model.Entry` (same `go.yaml.in/yaml/v3` used by the
   parser).
3. Reject empty/duplicate terms; require ≥1 translation per entry.
4. Leave `id` empty — `parser.assignIDs()` fills it during deck sync, exactly as
   it does for hand-written decks with missing IDs.

A wrong reply is a hard error with a clear message, never a silently corrupted
deck. `crds ai add` writes only after the existing parse/validate chain in
`Store.AppendEntries` succeeds.

### Deck-suggestion agent

`crds ai add` and `crds ai fill` accept an omitted deck: the CLI resolves it
through a fourth `SuggestDeck` agent in `agents.go`:

- **Prompt** (`SuggestDeckMessages` in `prompts.go`): passes the existing deck
  list (id, name, `language -> translation_language`) plus the raw input and
  demands a strict JSON reply:
  - `{"deck": "<id>"}` when one existing deck clearly fits, or
  - `{"deck": null, "proposed": {"name": ..., "from": ..., "to": ...}}` when
    none fits but a new deck is sensible.
- **Parsing** (`ParseSuggestion` in `parse.go`): a deck id not present in the
  given list is ignored (the model must never invent ids); a malformed reply
  degrades to "no match", never an error — a wrong guess is a UX prompt, not a
  data-write risk. Only transport failures propagate from the agent.
- **CLI flow** (`internal/cli/deck_resolve.go`): a suggested deck is confirmed
  with the user (`[y/N]`); on decline (or no match) the user can create a deck
  with the proposed name, type a new name, pick an existing deck (readline
  tab-completion over deck ids), or abort.

---

## Storage changes

New method in `internal/storage/store.go`:

```go
// AppendEntries appends entries to a deck's YAML file and syncs once.
// Empty IDs are auto-assigned during sync. Rejects duplicate terms via parse.
func (s *Store) AppendEntries(deckID string, entries []model.Entry, deckDir string) error
```

One parse → append → marshal → `syncDeck`, rather than looping `AddEntry`
(re-parses and re-syncs per entry). Duplicate-term validation is inherited from
`parser.Validate`.

## Tag listing

`TagListCmd.TermID` becomes optional (`internal/cli/tag.go`):

- `crds deck tag list <deck>` → all tags in the deck (`Store.ListDeckTags`).
- `crds deck tag list <deck> <id>` → tags on a single term (current behavior).

`Store.ListDeckTags` already exists (see `docs/proposals/tag_architecture.md`);
this is CLI wiring only.

---

## Tests

| Package | Coverage |
|---|---|
| `internal/ai` | client JSON round-trip against `httptest.Server` (success, HTTP error, malformed body); config resolution (preset, config.yaml, env override precedence, missing key for keyed providers); prompt content (tag allowlist present, variant-syntax rules, deck languages); parse (plain YAML, fenced YAML, garbage → error). |
| `internal/cli` | `interpret`/`fill`/`add` wiring with a fake `ai.Client`; auto-detect (structured vs unstructured input); tag-list command with and without term. |
| `internal/storage` | `AppendEntries`: append, auto-ID assignment on sync, duplicate-term rejection. |
| `internal/config` | `ai:` block parsing + env override. |

Verify with `make test`, `make build`, `make lint`.

---

## Docs

- `docs/roadmap.md` — "AI agent" section (this plan is referenced, not restated).
- `docs/README.md` — map entry for `internal/ai/PLAN.md`.
- `README.md` — CLI reference for `ai interpret/fill/add`, the `ai:` config
  block, and the extended `deck tag list`.
- `docs/status.md` — package row + test counts, **after** the code lands.
- `internal/ai/CONTEXT.md` — per-package context describing the package as it is,
  created when the package ships.

---

## Gotchas

- **Pollinations anonymous rate limit (~1 req / 15 s):** `crds ai add` makes two
  calls (interpret + fill), so it can take ~30 s+. Point `provider` at a keyed
  free tier (e.g. `groq` or `gemini`) to avoid the delay.
- **Deck `language` / `translation_language` matter:** the prompts must always
  receive them; `ui.DeckData` drops them, so the CLI must read the deck YAML via
  `parser.ParseFile` (as `TermEditCmd` already does) rather than `LoadDeck`.
- **Model replies are unreliable by nature:** never write to disk unvalidated.
  The existing parse/validate path is the gatekeeper.
- **Do not hardcode model behavior** in the Go code; all instructions live in
  the prompt builders so they can be tuned without code changes.
- **API key hygiene:** keys are read from config/env only, never echoed, and the
  file format comment in `config.yaml` must say so.

---

## Implementation order

1. `internal/ai`: `config.go` (presets + resolution) and `client.go` (+ tests).
2. `internal/ai`: `prompts.go` and `parse.go` (+ tests).
3. `internal/ai`: `agents.go` orchestrators.
4. `internal/config`: `ai:` block in `ConfigYAML` (+ tests).
5. `internal/storage`: `AppendEntries` (+ tests).
6. `internal/cli`: `ai.go` commands (`interpret`/`fill`/`add`) + tag-list change
   (+ tests).
7. Docs: README, status.md, `internal/ai/CONTEXT.md`; archive this plan in
   `docs/roadmap.md`.
8. Verify: `make test && make build && make lint && make docs-check`.

## Commit grouping

1. `ai: add OpenAI-compatible client and provider config` — `internal/ai`
   `config.go`/`client.go`, `internal/config` `ai:` block.
2. `ai: add interpret/fill prompts and output parsing` — `prompts.go`,
   `parse.go`, `agents.go`.
3. `cli: add ai interpret/fill/add commands`
4. `storage: add AppendEntries`
5. `cli: support deck-wide tag list`
6. `docs: AI agent feature` — README, roadmap, status, CONTEXT, PLAN archive.
