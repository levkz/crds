# Roadmap

**The single source of truth for planned work.** Actionable items are folded here
from the former `TODO.md` files, `ACTIONS_TODO.md`, and "Future work" sections.
Status for what exists lives in `docs/status.md`.

Feature proposals live in `docs/proposals/` and are archived here when implemented.

Legend: `[ ]` not started · `[~]` in progress · `[x]` done.

---

# Completed plans (archived)

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
- [ ] Due Today from a real scheduler (`internal/scheduler/`) — still open

## Former `PLAN.md` — original implementation plan

All phases complete. Superseded by this roadmap and `docs/status.md`.

- Phase 1 — renderer core utilities: done
- Phase 2 — layout adoption across screens: done
- Phase 3 — data pipeline (storage → app → screens): done
- Phase 4 — scheduler: not built (still open, see below)
- Phase 5 — placeholder packages: see below

## Former `ACTIONS_TODO.md` — missing CLI commands

Mostly shipped with the CLI. Remaining items carried forward below.

---

# Backlog

## Core domain packages (highest priority)

- [ ] **`internal/scheduler/`** — spaced repetition (SM-2, Leitner, or FSRS). Decides due cards and next review dates. Must not depend on Bubble Tea. Unblocks Statistics due/streak and per-answer snapshot refresh. *(source: former `PLAN.md` Phase 4, `docs/status.md` "Not Implemented" + known issues)*
- [ ] **`internal/quiz/`** — quiz session orchestration extracted from UI screens.
- [ ] **`internal/search/`** — vocabulary search extracted from UI screens.

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
- [ ] AI-generated example sentences / grammar hints / auto-quizzes
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
