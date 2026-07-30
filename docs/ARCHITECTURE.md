# ARCHITECTURE.md

# System Architecture

## Overview

CRDS is a terminal-based vocabulary learning application designed around a simple architectural principle:

> **Vocabulary is immutable content. User progress is mutable state.**

Vocabulary is authored as human-readable YAML files and can be version-controlled, shared, and edited manually.

User-specific data (progress, review history, scheduling, statistics, preferences) is stored separately in SQLite.

This separation keeps learning content portable while allowing efficient tracking of user progress.

---

# Architectural Goals

The architecture prioritizes:

- clear separation of responsibilities
- maintainability
- testability
- explicit dependencies
- incremental evolution
- idiomatic Go

When trade-offs arise, prefer simplicity over abstraction.

---

# System Layers

```text
                 cmd/crds
                     │
                     ▼
                 CLI (Kong)
                     │
                     ▼
              Application (App)
                     │
      ┌──────────────┼──────────────┐
      ▼              ▼              ▼
   Quiz         Scheduler      Search
      │              │              │
      └──────────────┼──────────────┘
                     ▼
               Domain Model
            (internal/model)
          ┌──────────┴──────────┐
          ▼                     ▼
      Parser                Storage
       (YAML)               (SQLite)
```

The application is composed of independent subsystems communicating through shared domain models.

Each subsystem owns a single responsibility.

---

# Dependency Rules

Dependencies always point toward the center of the application.

Allowed:

```text
CLI
    ↓
Application
    ↓
Services
    ↓
Model
```

Infrastructure packages (parser, storage, UI) depend on the domain model.

The domain model must never depend on infrastructure.

The dependency graph should remain acyclic.

---

# Subsystems

The project is divided into several independent subsystems.

| Subsystem | Responsibility | Status |
|---|---|---|
| CLI | Command-line interface and command dispatch | Fully implemented (3 command groups + top-level Quiz/Stats, Kong dispatch, TUI launch) |
| Parser | Loading and validating vocabulary files | Complete |
| Storage | Persisting user-specific state | Fully implemented via `Store` (SQLite, goose, sqlc) |
| Quiz | Learning session orchestration | Not implemented |
| Scheduler | Determining review order | Not implemented |
| Search | Vocabulary lookup | Not implemented |
| UI | Terminal presentation | Full UI with background fill, theme switching, 8 screens |
| Model | Shared domain objects | Complete |

Each subsystem has its own documentation describing its internal design.

---

# Current Implementation

## What exists

The codebase has two parallel implementations:

### 1. Legacy terminal quiz

- `cmd/legacy-quiz/main.go` reads `exercises/*.txt` files
- Simple terminal-based vocabulary drill
- No Bubble Tea, no UI framework

### 2. Modern UI application

- `cmd/crds/main.go` — Kong CLI + Bubble Tea
- Full terminal UI with navigation, theming, components

The modern UI is the active development focus.

---

## UI layer (`internal/ui/`)

The UI is the most developed subsystem. It follows Bubble Tea's Elm-style architecture.

### Complete packages

| Package | Description | Tests |
|---|---|---|
| `app/` | Root model, event dispatch, lifecycle, commands | — |
| `navigation/` | Stack-based manager with push/pop/replace/forward/modal | 82 tests |
| `keymap/` | Centralized keybinding definitions, user overrides | 16 tests |
| `theme/` | Design system: 18-field palette (15 colors + 3 semantic overrides), typography, icons, borders, spacing, 4 built-in themes | 68+ tests |
| `styles/` | 12 semantic style definitions | 60 tests |
| `components/` | 29 components (display + interactive) | — |
| `layout/` | Layout primitives: Page, Column, Row, Grid, Stack, Spacer, Center, Align | Tests |
| `renderer/` | Terminal rendering: width calculation, ANSI handling, text wrapping | — |
| `screens/` | 9 screens: Home, Quiz, TypingQuiz, DeckSelect, Search, Statistics, Settings, Detail, Palette | — |

### UI architecture

```text
            Bubble Tea
               │
               ▼
            tea.Msg
               │
               ▼
            Root Model (app/)
               │
         ┌─────┼──────────────┐
         │     │              │
         ▼     ▼              ▼
      Keymap  Navigation    Config
      (keymap/) (navigation/) (internal/config/)
         │     │              │
         │     ▼              │
         │  Registry → Screen │
         └─────┼──────────────┘
               │
      ┌────────┼──────────┐
      │        │          │
      ▼        ▼          ▼
    Home     Quiz      Search
    Screen   Screen     Screen
```

Screens are registered in a `navigation.Registry` and managed by `navigation.Manager`. The root model delegates to `Manager` and retrieves the active screen via `CurrentScreen()`.

Screens emit navigation events instead of changing the active screen directly. This prevents coupling between screens.

### Screen status

| Screen | Status | Notes |
|---|---|---|---|
| Home | Functional | 6-item activity menu. |
| Decks | Functional | Split-column deck+tag selection. Space toggle, a toggle-all, s search, tab/shift+tab column switch, enter confirms both selections. BackHandler for Esc handling. Persists decks and tags via StateStore. |
| Quiz | Functional | Real keyboard handling. Receives deck data, grades cards, dispatches progress. |
| TypingQuiz | Functional | Real keyboard handling. Text input for typed answers. Fuzzy-matched auto-grading. |
| Search | Functional | Real text input. Filters loaded deck data. Navigates to detail with stacked back-navigation. Clears state on leave via Lifecycle. |
| Statistics | Partial | Receives real stats from ProgressStore. Shows metrics. |
| Settings | Functional | Real theme switching. Most complete screen. |
| Detail | Functional | Receives entry data from Search via NavigateToDetailMsg. Uses stacked navigation (pushTo) for back to Search. |

---

## Parser layer (`internal/parser/`)

Fully functional with tests.

- `ParseFile(path)` and `Parse(data)` produce `*model.Deck`
- Pipeline: Unmarshal → Normalize (trim whitespace) → Validate (required fields, duplicate checks)
- Uses `go.yaml.in/yaml/v3` (NOT the standard `gopkg.in/yaml.v3`)
- 12 test fixtures in `internal/parser/testdata/`

### Known issues

- `duplicate_terms` test expects error but `validate.go` only checks duplicate IDs
- `auto_ids.yaml` fixture exists but auto-ID generation is not wired to validation

---

## Model layer (`internal/model/`)

Domain types with no methods or business logic:

| Type | Purpose |
|---|---|
| `Deck` | Vocabulary deck with entries |
| `Entry` | Single vocabulary item with translations, examples, tags |
| `Translation` | Text translation of an entry |
| `Example` | Example sentence with translation |
| `Progress` | Per-entry learning progress (ease, interval, due date) |
| `Review` | Single review record (grade, timestamp) |
| `Session` | Quiz session summary |

Note: `TypingDetail` exists only in sqlc-generated code (`internal/storage/db/models.go`), not in the domain model package.

---

## Config layer (`internal/config/`)

User configuration from `~/.config/crds/`:

- Auto-creates directory tree with default files
- Loads `config.yaml` (theme, animation, quiz limit)
- Loads `keymaps.yaml` (keybinding overrides via `keymap.ApplyDefaultOverrides()`)
- Discovers `themes/*.yaml` for custom themes
- 13 tests

---

# Composition Root

The application starts in `cmd/crds`.

Startup responsibilities include:

- loading configuration
- initializing dependencies
- opening the database
- constructing the application container
- invoking the CLI

Business logic should never exist in the application entry point.

**Current state:** `cmd/crds/main.go` pre-wires a fully populated `app.App` with `Store` (SQLite), `StateStore`, `SharedDir`, and `DataDir`. All CLI commands are fully implemented across 3 groups (DeckCmd, TermCmd, StateCmd) plus top-level Quiz/Stats. When no subcommand is given, `CLI.Run()` syncs decks and launches Bubble Tea. Three shell completion predictors are registered: `"deck"` (from SQLite `Store.ListDecks()`), `"reserve"` (from the default `reserve-copies/` directory), and `"term"` (from `Store.LoadDeck()` for entry IDs).

---

# Domain-Centric Design

The domain model is the center of the architecture.

Subsystems exchange domain objects rather than infrastructure-specific types.

For example:

- the parser returns `model.Deck`
- storage persists `model.Progress`
- the quiz consumes `model.Entry`

Database rows and YAML structures are implementation details of their respective packages.

---

# Data Sources

CRDS intentionally separates content from state.

## Vocabulary

Stored as YAML.

Properties:

- human editable
- version controlled
- shareable
- portable

## User State

Stored in SQLite.

Properties:

- progress
- scheduling
- review history
- statistics
- session history
- typing quiz details

SQLite is **not** the source of truth for vocabulary.

### SQL Stack

| Component | Choice | Purpose |
|-----------|--------|---------|
| Database | SQLite | Local, embedded database |
| Driver | `modernc.org/sqlite` | Pure Go SQLite driver (no CGo) |
| Migrations | goose | Schema versioning and migrations |
| Query builder | sqlc | Type-safe Go code from SQL queries |
| Interface | `database/sql` | Standard library database interface |

### Database Location

```text
~/.local/share/crds/crds.db
```

### SQLite DSN Format

The connection string uses `_pragma` query parameters — **not** bare key/value pairs. The `modernc.org/sqlite` driver ignores unrecognized parameters silently, so `_journal_mode=WAL&_foreign_keys=on` appears to work but does nothing.

Correct format (`internal/storage/store.go`):

```text
crds.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)
```

Each `_pragma` value is prepended with `PRAGMA ` and executed as raw SQL on every new connection drawn from the pool. The parenthesized form avoids URL-encoding `=` as `%3D`. An equivalent equals-form exists (`_pragma=foreign_keys=on`) but is harder to read in DSN strings.

| Parameter | Effect |
|-----------|--------|
| `_pragma=foreign_keys(1)` | Enables `ON DELETE CASCADE` and other FK constraints |
| `_pragma=journal_mode(WAL)` | Write-ahead logging for concurrent reads |
| `_pragma=busy_timeout(5000)` | Wait up to 5 s before returning SQLITE_BUSY |

**Why this matters:** Without `PRAGMA foreign_keys = ON`, `ON DELETE CASCADE` in the schema is silently ignored. Deleting a row in `entries` does **not** cascade to `translations`, `examples`, or `entry_tags`, leaving orphaned rows that accumulate on every re-sync.

> **Do not** use `db.Exec("PRAGMA foreign_keys = ON")` after opening the connection. With `database/sql` connection pooling, the pragma only applies to the single connection it runs on — not to connections later drawn from the pool.

### Storage Layer Structure

```text
internal/storage/
├── migrations/          # goose SQL migrations
│   ├── 20260716121051_init.sql
│   └── 20260721000000_add_deck_cache.sql
├── queries/             # sqlc query definitions
│   ├── sessions.sql
│   ├── reviews.sql
│   ├── typing_details.sql
│   ├── progress.sql
│   ├── decks.sql
│   ├── entries.sql
│   ├── translations.sql
│   ├── examples.sql
│   ├── entry_tags.sql
│   └── sync_state.sql
├── sqlc.yaml            # sqlc configuration
├── db/                  # generated by sqlc (12 files)
├── store.go             # manual wrapper: session mgmt, RecordAnswer, Stats
├── sync.go              # deck sync from YAML + DB-backed ListDecks/LoadDeck
├── deck_store.go        # filesystem YAML deck loading (legacy)
├── progress_store.go    # legacy in-memory progress store
├── state.go             # YAML-based app state
├── converter.go         # model.Entry -> ui.CardData
└── schema.sql           # standalone schema reference
```

**Current state:** SQLite is fully implemented via `Store` in `store.go`. On startup, `SyncDecks` syncs YAML decks into the `decks`/`entries`/`translations`/`examples`/`entry_tags` cache tables. `Store` implements `DeckProvider`, `ProgressRecorder`, `StatsProvider`, and `SessionManager`. The in-memory `ProgressStore` is a legacy predecessor that remains but is not wired.

---

# Extension Strategy

New functionality should be added by extending existing layers rather than bypassing them.

Typical flow:

```text
CLI
    ↓
Application
    ↓
Service
    ↓
Model
    ↓
Infrastructure
```

Avoid introducing shortcuts between unrelated packages.

---

# Package Organization

Each package should own a single concern.

Subsystem implementation details belong inside their own directories.

```text
internal/
├── model/          Domain types (Deck, Entry, Progress, Review, Session, TypingDetail)
├── parser/         YAML parsing + validation + normalization
├── config/         User configuration from ~/.config/crds/
├── app/            Composition root (Store, State, SharedDir, DataDir)
├── cli/            Kong commands (3 groups: DeckCmd, TermCmd, StateCmd) + CLI tests
├── ui/             Terminal UI (Bubble Tea)
│   ├── app/        Root model, event dispatch, lifecycle
│   ├── screens/    Screen implementations
│   ├── components/ Reusable UI components
│   ├── theme/      Design system
│   ├── styles/     Semantic style definitions
│   ├── keymap/     Centralized keybindings
│   ├── navigation/ Stack-based navigation
│   ├── layout/     Layout primitives
│   ├── renderer/   Terminal rendering utilities
│   └── docs/       UI documentation
├── storage/        SQLite persistence (goose + sqlc)
│   ├── migrations/ # SQL migrations
│   ├── queries/    # sqlc query definitions
│   └── *.go        # generated + manual code
├── quiz/           [not implemented]
├── scheduler/      [not implemented]
└── search/         [not implemented]
```

Shared types belong in:

```text
internal/model/
```

---

# Testing Philosophy

Each subsystem should be independently testable.

Prefer:

- table-driven tests
- deterministic behavior
- package-local `testdata/`
- minimal mocking

Favor testing observable behavior over implementation details.

---

# Documentation

Documentation is intentionally distributed.

Under each subsystem there may be a `docs/` directory with `CONTEXT.md` and `ARCHITECTURE.md`.

This document describes the overall system.

Additional documentation is located alongside the relevant subsystem.

Examples:

```text
internal/ui/docs/
docs/
```

---

# Known Issues

- `duplicate_terms` test expects error but validation only checks duplicate IDs
- CLI commands fully implemented: quiz, stats (--deck), deck (list, import/--replace/dir, export/--all, delete, search/--deck/--tags, edit, term-add/edit/rm with -t/-f, tag-add/rm/list), state (reserve/revert/sync), profile (export/import)
- Deck selection screen exists but empty selection means no quiz (must pick at least one deck)
- Grade scale mismatch: Flashcard uses 0-3, Typing uses 1-3 (needs normalization)

---

# Long-Term Vision

CRDS should remain:

- lightweight
- keyboard-first
- portable
- modular
- approachable for contributors

Architectural decisions should preserve these qualities as the project evolves.
