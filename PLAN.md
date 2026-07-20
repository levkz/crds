# Implementation Plan

## Current state summary

| Layer | Status |
|---|---|
| `app/`, `navigation/`, `keymap/` | ~90% complete, well-tested |
| `theme/`, `styles/` | ~85-90% complete, tested |
| `components/` | ~70% — 8 components, functional |
| `layout/` | ~80% — 9 primitives exist, adopted by all screens |
| `screens/` | ~30% — Settings works; Home/Quiz/Search have UI but no data; Statistics/Detail are placeholders |
| `renderer/` | ~50% — core utilities implemented |
| `parser/` (outside ui) | Functional with tests, not wired to UI |
| `model/` (outside ui) | Domain types exist (Deck, Entry, Progress), no storage/scheduler |

---

## Phase 1: `renderer/` core utilities (standalone, no deps on UI) — COMPLETE

**Files created:**
- `internal/ui/renderer/width.go` — `VisibleWidth(s)`, `LineWidth(s)`, `MaxLineWidth(text)`, `TextDimensions(text)` using `go-runewidth`
- `internal/ui/renderer/ansi.go` — `StripANSI(s)`, `CountANSISequences(s)`
- `internal/ui/renderer/wrap.go` — `Wrap(text, maxWidth)`, `Truncate(text, maxWidth)`, `Fit(text, maxWidth)`
- `internal/ui/renderer/renderer_test.go` — table-driven tests for all the above

**Status:** Implemented by another agent.

---

## Phase 2: Wire `layout/` into screens — COMPLETE

Replaced manual `strings.Builder` concatenation in all 6 screens with `layout.Page()`, `layout.Column()`.

**Files modified:**
- `internal/ui/screens/home.go` — `Page`
- `internal/ui/screens/quiz.go` — `Page` + `Column`
- `internal/ui/screens/search.go` — `Page` + `Column`
- `internal/ui/screens/statistics.go` — `Page` + `Column`
- `internal/ui/screens/settings.go` — `Page` + `Column`
- `internal/ui/screens/detail.go` — `Page` + `Column`

**Status:** Complete. All screens use declarative layout primitives with consistent `"\n\n"` spacing. Build, tests, and vet pass.

---

## Phase 3: Data pipeline — parser → app → screens

### 3a: Storage layer (`internal/storage/`)

**New package:** `internal/storage/`

- `storage.go` — `Store` interface: `ListDecks()`, `LoadDeck(id)`, `GetEntries(deckID)`, `GetProgress(entryID)`, `SaveProgress(Review)`
- `file.go` — Filesystem implementation (reads YAML from a data directory, caches in memory)
- `storage_test.go` — tests with fixtures

**Depends on:** `internal/parser/`, `internal/model/`

### 3b: Wire `DeckProvider` + `ProgressRecorder` into app

**Files to modify:**
- `internal/ui/app/dependencies.go` — keep interfaces as-is
- `internal/ui/app/app.go` — accept a `Dependencies` struct in `New()`, store on root Model
- `internal/ui/app/commands.go` — `ListDecksCmd`/`LoadDeckCmd` call real `DeckProvider`; `RecordAnswerCmd` calls real `ProgressRecorder`
- `cmd/crds/main.go` — construct `storage.Store`, pass as `Dependencies` to `app.New()`

### 3c: Wire screens to data

- **Quiz:** `Init()` → emit `LoadDeckCmd`; handle `DataLoadedMsg` → populate `Cards`; `grade()` → advance index + emit `RecordAnswerCmd`
- **Search:** Accept `[]Entry` from loaded deck; `filterResults()` → real substring match against terms/translations
- **Statistics:** Accept progress data; populate metrics from real `Review` records
- **Detail:** Accept `Entry` reference; render real fields
- **Home:** `Init()` → emit `ListDecksCmd`; render deck list instead of hardcoded activities

---

## Phase 4: Scheduler (`internal/scheduler/`)

**New package:** `internal/scheduler/`

- `scheduler.go` — Spaced repetition logic (SM-2 or simpler interval-based)
- `scheduler_test.go` — tests

**Depends on:** `internal/model/` (Progress, Review types)

**Used by:** Quiz screen's `grade()` to compute next review date and select due cards.

---

## Phase 5: Remaining placeholder packages (optional, lower priority)

| Package | Purpose | When |
|---|---|---|
| `actions/` | Typed action constants (if event routing needs them) | When screens emit more complex actions |
| `events/` | Shared event types (currently in `app/commands.go`) | When events cross package boundaries |
| `debug/` | Debug overlay, FPS counter | When needed for development |
| `animations/` | Transitions, loading spinners | Polish phase |

---

## Execution order

```
Phase 1 (renderer)        — COMPLETE
    ↓
Phase 2 (layout adoption) — COMPLETE
    ↓
Phase 3a (storage)        — New package, depends on parser/model
    ↓
Phase 3b (wire app)       — Connects storage → app interfaces
    ↓
Phase 3c (wire screens)   — Connects app → screens (Quiz/Search/Stats/Detail)
    ↓
Phase 4 (scheduler)       — Adds spaced repetition to Quiz
    ↓
Phase 5 (polish)          — Optional extras
```
