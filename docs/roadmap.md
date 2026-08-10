# Roadmap

**The single source of truth for planned work.** Actionable items are folded here
from the former `TODO.md` files, `ACTIONS_TODO.md`, and "Future work" sections.
Status for what exists lives in `docs/status.md`.

Feature proposals live in `docs/proposals/` and are archived here when implemented.

Legend: `[ ]` not started · `[~]` in progress · `[x]` done.

---

# Completed plans (archived)

## Accent input mappings + matching mode

- [x] `internal/mapping/` — input mappings (Babbel-style triggers): longest-suffix
      expansion, rune-safe, single pass per keystroke
- [x] Built-in French defaults; user files in `~/.config/crds/mappings/<lang>.yaml`;
      per-deck `input_mappings` field; precedence deck > user > built-in
- [x] Typing quiz applies the effective mapping as the user types
- [x] `internal/fuzzy` strict/approximate modes — `matching_mode: strict|approximate`
      in `~/.config/crds/config.yaml` (default `approximate`); accents stripped via
      NFD + combining-mark removal
- [x] Parse toggle (`ctrl+p`) in the typing quiz — toggles trigger expansion so a
      literal `e/` can be typed; old text is never re-parsed when toggling back on,
      while new text parses even mid-string (via `Mapping.ApplyAt` at the cursor)
- [x] `mappings/*.yaml` included in profile export/import
- [x] Docs: README, DECK_CREATION_GUIDE, status.md

## Scheduler: SM-2 spaced repetition

- [x] `internal/scheduler/` — pure Go SM-2 algorithm (ease, interval, lapse
      penalty, grade→interval mapping), no Bubble Tea dependency
- [x] Persist scheduling state with every answer (`RecordAnswer`/
      `RecordAnswerFull` transactional progress upsert)
- [x] Review queue (`DueForSelection`): unseen cards first, then due cards by
      due date; distinct across decks/tags
- [x] Due quiz mode (`QuizModeDue`) shared by Quiz and TypingQuiz, with the
      queue ordered by the review queue
- [x] Session snapshot — a running quiz keeps its queue; progress/due refreshes
      update state without reshuffling mid-session
- [x] Stats wiring — selection "Due Today" count and per-word yes/no, refreshed
      after every answer

## Statistics screen revamp (words + selection tabs)

- [x] Per-word tab: search-driven vocabulary lookup with per-word stats (total
      reviews, accuracy, confidence, mastered, last reviewed) and a per-word
      confidence graph
- [x] Selection tab: summary metrics for the selected decks/tags (reviewed
      today, accuracy, due today, current streak, total cards, mastered) with a
      confidence-over-time bar graph
- [x] Streak computed from review history (`internal/stats.Streak`)
- [x] Word-level + selection queries in `internal/storage` (sqlc) behind the
      `internal/stats.Provider` interface
- [x] `Statistics` keymap group (`tab` switch, `esc` back) with user overrides
- [x] Due Today from a real scheduler (`internal/scheduler/`) — see Scheduler above

## Former `PLAN.md` — original implementation plan

All phases complete. Superseded by this roadmap and `docs/status.md`.

- Phase 1 — renderer core utilities: done
- Phase 2 — layout adoption across screens: done
- Phase 3 — data pipeline (storage → app → screens): done
- Phase 4 — scheduler: done (see Scheduler above)
- Phase 5 — placeholder packages: see below

## Former `ACTIONS_TODO.md` — missing CLI commands

Mostly shipped with the CLI. Remaining items carried forward below.

---

# Backlog

## Core domain packages (highest priority)

- [ ] **`internal/quiz/`** — quiz session orchestration extracted from UI screens.
- [ ] **`internal/search/`** — vocabulary search extracted from UI screens.

## AI agent — `crds ai` (interpret / fill / add)

Detailed plan and context: `internal/ai/PLAN.md`.

- [x] `internal/ai/` package — OpenAI-compatible chat client, provider presets,
      config resolution, interpret/fill prompts, YAML output parsing
- [x] `ai:` config block (`provider`, `model`, `api_key`, `base_url`) + env
      overrides (`CRDS_AI_*`); default provider: Pollinations.AI (keyless)
- [x] CLI: `crds ai interpret [--deck <deck>]` — unstructured text → YAML entries
- [x] CLI: `crds ai fill <deck>` — structured YAML → completed YAML
- [x] CLI: `crds ai add <deck>` — interpret+fill with review, append to deck
- [x] Storage: `AppendEntries` bulk append with auto-ID on sync
- [x] CLI: `crds deck tag list <deck>` — list all tags in a deck (no term)
- [x] Docs: README CLI reference, `internal/ai/CONTEXT.md`, status.md update

## UI wiring (interactive components)

- [ ] Wire `SearchInputModel` into `screens/search.go`
- [ ] Wire `SelectableListModel` into `screens/settings.go`
- [ ] Wire `StatusBar` into `app/view.go`
- [ ] Wire `SpinnerModel` into async operations
- [ ] Wire `ConfirmDialog` into `screens/quiz.go` (confirm quit before abandoning)
- [ ] Wire `RadioGroupModel` / `SelectModel` / `MultiSelectModel` into settings forms
- [ ] Wire `TreeModel` into a deck/category browser
- [ ] Wire `Table` into a leaderboard screen (Statistics uses graph + metric grid)
- [ ] Add tests for the 20 display + 9 interactive components
- [ ] `FocusGroup` component — manage tab-order between interactive components on one screen
- [ ] Error boundaries — `RenderError` component or pattern for render panics
- [ ] Responsive breakpoints — compact mode at narrow terminal widths
- [ ] Viewport component — scrollable wrapper for overflow content (Statistics word list, Detail)

*(source: `components/CONTEXT.md` "Not yet wired" + Suggestions)*

## Keymap

- [ ] Chord bindings (e.g. `g` then `g` for top of list)
- [ ] Mouse binding support
- [ ] Per-screen keymap overrides in user config

*(source: `keymap/CONTEXT.md` Future work)*

## Renderer

- [ ] `TerminalState` struct with `Width`/`Height` and `Update(tea.WindowSizeMsg)`
- [ ] Scrolling helpers (offset, line/page navigation)
- [ ] Viewport-aware truncation
- [ ] Efficient re-rendering / diffing
- [ ] Component-aware character counting
- [ ] Performance profiling utilities

*(source: `renderer/CONTEXT.md` Future work)*

## Events package

- [ ] Navigation event
- [ ] Resize event
- [ ] Focus / blur events
- [ ] Session updated / review completed events

*(source: `docs/status.md` "Not Implemented" — events placeholder)*

## Actions package

Deferred until a concrete trigger appears:

- [ ] Command palette (actions become the palette vocabulary)
- [ ] Mouse support (actions decouple click targets from keys)
- [ ] "Back" logic duplicated in 3+ screens (extract to shared action)
- [ ] Accessibility / alternative input methods

*(source: `actions/CONTEXT.md`)*

## CLI enhancements

- [ ] `crds deck search --fuzzy-find` — interactive fzf picker
- [ ] `crds stats --tag <tag>` / `--level <level>` filters
- [ ] Wire `quiz --limit` and `quiz --reverse` into the TUI

*(sources: former `ACTIONS_TODO.md`, `internal/cli/CONTEXT.md`)*

## Theme

- [ ] `--theme` CLI flag → `app.Config.ThemePath` for custom YAML themes
- [ ] Expose underline, strikethrough, background, padding in `ConfigTextRole`
- [ ] Plugin themes from `~/.config/crds/themes/`
- [ ] Icon-aware style markers (`SelectedItemIcon()`, etc.)
- [ ] `FocusedInput(active bool)` variant for unfocused state
- [ ] Style composition helpers if duplication emerges

*(sources: `theme/CONTEXT.md` Future work, `styles/CONTEXT.md`)*

## Navigation

- [ ] Forward keybinding exposed in the app
- [ ] Stacked overlays
- [ ] Navigation middleware (logging, analytics, validation)
- [ ] Deep linking — parse a path string into Push/Pop sequences

*(source: `navigation/CONTEXT.md` Future work)*

## Placeholder packages

- [ ] `internal/ui/animations/` — spinner, progress, toast animations, transitions
- [ ] `internal/ui/debug/` — debug overlay, layout inspector, event logger, FPS counter
- [ ] `internal/ui/testdata/` — reusable fixtures (sample decks, sessions, stats, app state)

*(sources: `docs/status.md` "Not Implemented" — placeholder packages)*

## Stretch / product ideas (from `docs/DESIGN.md`)

- [ ] Audio pronunciation
- [ ] Images attached to vocabulary
- [ ] AI-generated example sentences / grammar hints / auto-quizzes — example-
      sentence generation ships with the AI agent (see "AI agent" section above);
      grammar hints and auto-quizzes remain open
- [ ] TUI dashboard — reviews due, streak, accuracy, recently learned
- [ ] Instant fuzzy search
- [ ] Tag taxonomy (food, travel, verbs, adjectives, A1–B2)
- [ ] Statistics — avg response time, weakest words (daily reviews, streak,
      accuracy, mastered shipped in the 2026 statistics revamp; see "Completed plans")
- [ ] `crds sync --prune` — remove stale decks from the SQLite cache

## Data model extensions (from `docs/DATAMODEL.md`)

The YAML/data model is designed to be extensible without architectural change.
Candidates:

- [ ] Entry fields: pronunciation, gender, plural forms, irregular conjugations
- [ ] Difficulty / CEFR levels (A1–C2) on entries
- [ ] Synonyms / antonyms / etymology
- [ ] Grammar-topic references (links entries to grammar rules)

## Input mappings & matching (follow-ups to the shipped feature)

- [ ] Per-deck `matching_mode` override (deck wins over the user setting)
- [ ] Built-in defaults for more languages (es, de, pt, pl, vi, ...)
- [ ] Handle non-decomposing letters in approximate mode (`ß`→`ss`, `ø`→`o`, `å`→`a`, ...) as optional language-aware folds
- [ ] On-screen hint showing the active mapping triggers for the current deck
