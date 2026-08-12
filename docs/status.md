# Project Status

**This is the single source of truth for implementation status and known issues.**
Every other document links here instead of restating status. Update this file in the
same commit as the code change it describes.

For what comes next, see `docs/roadmap.md`.

---

# Test Baseline

Ran `go test ./...` from the repo root. Results are stable and current:

| Package | Tests | Notes |
|---|---|---|
| `internal/model/` | 3 | Domain types |
| `internal/parser/` | 5 | 13 YAML fixtures in `testdata/` |
| `internal/config/` | 17 | User configuration |
| `internal/fuzzy/` | 12 | Levenshtein matching, accent stripping, strict/approximate modes |
| `internal/mapping/` | 12 | Input mappings: suffix expansion at a cursor index, pairs listing, layer precedence, file loading |
| `internal/stats/` | 5 | Stats aggregation, streak, word stats |
| `internal/scheduler/` | 7 | SM-2 spaced-repetition algorithm |
| `internal/storage/` | 88 | SQLite Store (goose + sqlc) |
| `internal/cli/` | 28 | Kong command wiring |
| `internal/ai/` | 54 | AI agent: providers, config, prompts, parsing |
| `internal/ui/` | 7 | Quiz modes, card sorting |
| `internal/ui/app/` | 5 | Config apply, deck merging, session finalize on shutdown |
| `internal/ui/theme/` | 61 | Design system, 6 fixtures |
| `internal/ui/styles/` | 14 | Semantic styles |
| `internal/ui/layout/` | 10 | Layout primitives |
| `internal/ui/renderer/` | 9 | Rendering utilities |
| `internal/ui/keymap/` | 18 | Keybindings |
| `internal/ui/navigation/tests/` | 74 | Black-box navigation tests |
| `internal/ui/app/tests/` | 8 | State-sync protocol, quiz-mode persistence |
| `internal/ui/components/display/` | 5 | Graph bar chart |
| `internal/ui/screens/` | 24 | Statistics logic (metric grid fits content width), search detail (tags + examples carried through to Detail), typing quiz input mappings (incl. parse toggle + legend) + grading |

Total: **466 test functions**, all passing.

---

# Implementation Status

## Complete

| Package | Description | Tests |
|---|---|---|
| `internal/model/` | Domain types: Deck, Entry, Progress, Review, Session | 3 |
| `internal/parser/` | YAML parsing, validation, normalization, auto-ID generation | 5 (13 fixtures) |
| `internal/config/` | User config from `~/.config/crds/`: `config.yaml`, `keymaps.yaml`, `themes/*.yaml`, `mappings/*.yaml` | 17 |
| `internal/fuzzy/` | Levenshtein-based fuzzy matching for typed answers, accent stripping, strict/approximate modes | 12 |
| `internal/mapping/` | Input mappings (Babbel-style triggers): suffix expansion at a cursor index (`ApplyAt`), pairs listing, built-in/user/deck layer precedence, `~/.config/crds/mappings/<lang>.yaml` loading | 11 |
| `internal/stats/` | Stats aggregation for statistics screen, word-level stats, streak from review history | 5 |
| `internal/scheduler/` | SM-2 spaced-repetition algorithm: ease, interval, lapse penalty, grade→interval mapping | 7 |
| `internal/editor/` | `$EDITOR`/nano/vim invocation with YAML buffer handling | — |
| `internal/cli/` | Kong commands: quiz, stats, deck (create/list/import/export/delete/search/edit), term (add/rm/edit), tag (add/rm/list), state (reserve/revert/sync), profile (export/import), ai (interpret/fill/add, incl. `--full`/`--minimal`/`--msg`/`--translate-from`/`--translate-to`) | 28 |
| `internal/ai/` | Agent: 7 provider presets (pollinations, ollama, openai, gemini, openrouter, groq, nvidia), OpenAI-compatible client, prompts (minimal/full interpret, fill, structural+proficiency+theme `tagRules`, `--msg` passthrough), YAML parsing (examples validated for both languages), interpret/fill agents | 54 |
| `internal/ui/theme/` | Design system: 18-field palette (15 colors + 3 semantic overrides), typography, icons, borders, spacing, 5 built-in themes, YAML loading, store | 61 |
| `internal/ui/styles/` | Semantic style definitions | 14 |
| `internal/ui/components/` | 29 components (20 display + 9 interactive) | — |
| `internal/ui/components/display/` | Graph bar chart, confidence coloring | 5 |
| `internal/ui/navigation/` | Stack-based navigation: push/pop/replace/forward/modal/overlay | 74 |
| `internal/ui/keymap/` | Centralized keybinding definitions with user overrides | 18 |
| `internal/ui/layout/` | Layout primitives: Page, Column, Row, Grid, Stack, Spacer, Center, Align | 10 |
| `internal/ui/renderer/` | Terminal rendering: width, ANSI, wrapping | 9 |
| `internal/ui/app/` | Root Bubble Tea model, event dispatch, lifecycle (shutdown finalizes the active session), commands, `ui.AppState` sync | 5 |
| `internal/ui/screens/` | Typing quiz applies per-language input mappings while typing (toggleable with `ctrl+p`, with an on-screen trigger legend); grading respects strict/approximate matching. Statistics metric panels account for the border frame so the grid fits the content width. Search/Detail carry each card's tags and examples into the quiz-style bottom section. | 24 |

## Partially Implemented

| Package | Status |
|---|---|
| `internal/storage/` | `Store` (SQLite) fully implemented: deck+entry CRUD, tags/`deck_tags`, `ListDeckTags` (deck-wide tag list), `AppendEntries` (batch append), reserve/backup, revert, profile export/import, sync. Reserve copies and profile exports bundle the config subtree (`config.yaml`, `keymaps.yaml`, `themes/*.yaml`, `mappings/*.yaml`) under a `config/` prefix, and revert/import restore it, so per-language input mappings are covered by both migration paths. SM-2 scheduling persisted with every answer via `RecordAnswer`/`RecordAnswerFull`; due queue (`DueForSelection`) and due-today counts wired. Review stats treat `grade >= GradeGood` (2) as correct consistently, in SQL and Go. Sessions are finalized (reviewed/correct/incorrect/duration written) on deck-switch and on app shutdown, so abandoned sessions still persist their aggregates. Legacy `DeckStore` and `ProgressStore` remain but are not wired. |
| `internal/app/` | Composition root with Store/State/SharedDir/DataDir, pre-wired before Kong dispatch. |
| `internal/ui/screens/` | 9 screens implemented and functional: Home, Quiz, TypingQuiz, DeckSelect, Search, Statistics, Settings, Detail, Palette (dev screen). Statistics has words/selection tabs: per-word search with per-word stats, and a selection summary with a confidence-over-time graph. Due Today is wired for both the selection summary (count) and per-word detail (yes/no). |
| `internal/ui/components/` | All 29 built; the 9 interactive components are not yet wired into screens. |

## Not Implemented

| Package | Status |
|---|---|
| `internal/quiz/` | Does not exist. Quiz logic lives in UI screens only. |
| `internal/search/` | Does not exist. Search lives in UI screens only. || `internal/ui/actions/` | Empty placeholder. |
| `internal/ui/events/` | 4 basic event types defined; rest deferred. |
| `internal/ui/animations/` | Empty placeholder. |
| `internal/ui/debug/` | Empty placeholder. |
| `internal/ui/testdata/` | Empty placeholder. |

---

# Known Issues

- Deck selection requires at least one deck — empty selection means no quiz.
- `crds quiz --limit` and `--reverse` flags are acknowledged with warnings but not wired to the TUI.
- `crds deck edit` prompts are interactive and cannot be tested automatically.
- Quiz progress in the global `AppState` snapshot is refreshed after each answer, but a running quiz freezes its queue (session snapshot) to avoid reshuffling mid-session.
- Due-mode reviews are drawn from the selection's review queue (unseen first, then due by date); the "Due Today" stat counts only due entries, so it can differ from the length of the due-mode queue when unseen cards are pending.
- Interactive components are built but not yet wired into screens (`search.go`, `settings.go` still handle input inline).
- Approximate matching strips combining diacritics only; letters that don't NFD-decompose (`ß`, `ø`, `ł`, `å`) are unaffected — use input mappings for those.
- `internal/editor/` has no tests yet.
- Legacy `ProgressStore` and `DeckStore` remain in the codebase but are not wired.
