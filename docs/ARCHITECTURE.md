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
| CLI | Command-line interface and command dispatch | Stubs only |
| Parser | Loading and validating vocabulary files | Complete |
| Storage | Persisting user-specific state | Partially implemented (DeckStore + in-memory ProgressStore) |
| Quiz | Learning session orchestration | Not implemented |
| Scheduler | Determining review order | Not implemented |
| Search | Vocabulary lookup | Not implemented |
| UI | Terminal presentation | Wired to real data |
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
| `theme/` | Design system: palette, typography, icons, borders, spacing | 54 tests |
| `styles/` | 12 semantic style definitions | 60 tests |
| `components/` | 29 components (display + interactive) | — |
| `layout/` | Layout primitives: Page, Column, Row, Grid, Stack, Spacer, Center, Align | Tests |
| `renderer/` | Terminal rendering: width calculation, ANSI handling, text wrapping | — |
| `screens/` | 6 screens: Home, Quiz, Search, Statistics, Settings, Detail | — |

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
|---|---|---|
| Home | Functional | Static 4-item menu. Auto-loads first deck on init. |
| Quiz | Functional | Real keyboard handling. Receives deck data, grades cards, dispatches progress. |
| Search | Partial | Real text input. Filters loaded deck data. Navigates to detail. |
| Statistics | Partial | Receives real stats from ProgressStore. Shows metrics. |
| Settings | Functional | Real theme switching. Most complete screen. |
| Detail | Partial | Receives entry data from search. Shows term, translations, examples. |

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

**Current state:** `cmd/crds/main.go` uses Kong to route to `cli.CLI` commands. Most commands are stubs. The TUI is launched via `cli.CLI.Run()`, which creates `DeckStore` (filesystem YAML reader) and `ProgressStore` (in-memory progress), wires them as `DeckProvider`/`ProgressRecorder`/`StatsProvider`, and starts Bubble Tea. A `deck` completion predictor reads deck names from the filesystem for tab completion.

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
- preferences

SQLite is **not** the source of truth for vocabulary.

**Current state:** SQLite is not implemented. The `migrations/20260716121051_init.sql` file is a goose placeholder with no real schema. The `internal/storage/` package contains `DeckStore` (filesystem YAML reader) and `ProgressStore` (in-memory progress tracking) as a lightweight alternative until SQLite is wired.

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
├── model/          Domain types (Deck, Entry, Progress, Review, Session)
├── parser/         YAML parsing + validation + normalization
├── config/         User configuration from ~/.config/crds/
├── app/            Empty composition root (aspirational)
├── cli/            Kong command stubs
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
├── storage/        DeckStore + in-memory ProgressStore
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
- Screens use hardcoded width 60 — no terminal resize awareness
- Progress is in-memory only — not persisted to disk
- CLI commands are stubs — only the TUI launches
- No deck selection screen — auto-loads the first available deck

---

# Long-Term Vision

CRDS should remain:

- lightweight
- keyboard-first
- portable
- modular
- approachable for contributors

Architectural decisions should preserve these qualities as the project evolves.
