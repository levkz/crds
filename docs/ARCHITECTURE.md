# System Architecture

> Implementation status and known issues are tracked in `docs/status.md`.
> Planned work lives in `docs/roadmap.md`. This document describes how the
> system is designed and layered.

## Overview

CRDS is a terminal-based vocabulary learning application designed around a simple
architectural principle:

> **Vocabulary is immutable content. User progress is mutable state.**

Vocabulary is authored as human-readable YAML files and can be version-controlled,
shared, and edited manually.

User-specific data (progress, review history, scheduling, statistics,
preferences) is stored separately in SQLite.

This separation keeps learning content portable while allowing efficient tracking
of user progress.

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

The application is composed of independent subsystems communicating through
shared domain models. Each subsystem owns a single responsibility.

---

# Dependency Rules

Dependencies always point toward the center of the application.

```text
CLI
    ↓
Application
    ↓
Services
    ↓
Model
```

Infrastructure packages (parser, storage, UI) depend on the domain model. The
domain model must never depend on infrastructure. The dependency graph must stay
acyclic.

---

# Subsystems

| Subsystem | Responsibility | Status |
|---|---|---|
| CLI | Command-line interface and dispatch | Implemented |
| Parser | Loading and validating vocabulary files | Complete |
| Storage | Persisting user-specific state (SQLite) | Implemented |
| Quiz | Learning session orchestration | Not implemented (logic lives in UI screens) |
| Scheduler | Determining review order | Implemented (SM-2 in `internal/scheduler/`, queue in `internal/storage/`) |
| Search | Vocabulary lookup | Not implemented (logic lives in UI screens) |
| UI | Terminal presentation (Bubble Tea) | Implemented |
| Model | Shared domain objects | Complete |

See `docs/status.md` for precise implementation detail. Each subsystem has its
own documentation describing its internal design (see `docs/README.md`).

---

# Current Implementation

The codebase has two parallel implementations:

## 1. Legacy terminal quiz

- `cmd/legacy-quiz/main.go` reads `exercises/*.txt` files
- Simple terminal-based vocabulary drill
- No Bubble Tea, no UI framework

## 2. Modern UI application

- `cmd/crds/main.go` — Kong CLI + Bubble Tea
- Full terminal UI with navigation, theming, components

The modern UI is the active development focus.

---

# Composition Root

The application starts in `cmd/crds`. Startup responsibilities:

- loading configuration
- initializing dependencies
- opening the database
- constructing the application container
- invoking the CLI

Business logic should never exist in the application entry point.

`cmd/crds/main.go` pre-wires a fully populated `app.App` with `Store` (SQLite),
`StateStore`, `SharedDir`, and `DataDir`. When no subcommand is given, `CLI.Run()`
syncs decks and launches Bubble Tea. Three shell completion predictors are
registered: `"deck"` (from SQLite `Store.ListDecks()`), `"reserve"` (from the
default `reserve-copies/` directory), and `"term"` (from `Store.LoadDeck()` for
entry IDs).

---

# Domain-Centric Design

The domain model (`internal/model/`) is the center of the architecture.
Subsystems exchange domain objects rather than infrastructure-specific types:

- the parser returns `model.Deck`
- storage persists `model.Progress`
- the quiz consumes `model.Entry`

Database rows and YAML structures are implementation details of their
respective packages.

---

# Data Sources

CRDS intentionally separates content from state.

- **Vocabulary** — YAML files. Human editable, version controlled, shareable,
  portable. SQLite is **not** the source of truth for vocabulary.
- **User state** — SQLite database at `~/.local/share/crds/crds.db`. Progress,
  scheduling, review history, statistics, session history, typing quiz details.

The SQL stack is SQLite (`modernc.org/sqlite`) + goose (migrations) + sqlc
(type-safe queries). The DSN uses `_pragma` query parameters (not bare
key/value pairs) — see `internal/storage/store.go` and `docs/DATAMODEL.md`.

> **Do not** use `db.Exec("PRAGMA foreign_keys = ON")` after opening the
> connection. With `database/sql` connection pooling, the pragma only applies to
> the single connection it runs on. It must be set in the DSN via `_pragma` for
> every pooled connection.

---

# Extension Strategy

New functionality should be added by extending existing layers rather than
bypassing them:

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

Each package owns a single concern. Subsystem implementation details belong
inside their own directories.

```text
internal/
├── model/          Domain types (Deck, Entry, Progress, Review, Session)
├── parser/         YAML parsing + validation + normalization
├── config/         User configuration from ~/.config/crds/
├── app/            Composition root (Store, State, SharedDir, DataDir)
├── cli/            Kong commands
├── ui/             Terminal UI (Bubble Tea) — see internal/ui/docs/
├── storage/        SQLite persistence (goose + sqlc)
├── quiz/           [not implemented]
├── scheduler/      [not implemented]
└── search/         [not implemented]
```

Shared types belong in `internal/model/`.

---

# Testing Philosophy

Each subsystem should be independently testable. Prefer:

- table-driven tests
- deterministic behavior
- package-local `testdata/`
- minimal mocking

Favor testing observable behavior over implementation details.

---

# Documentation

Documentation is intentionally distributed. Each subsystem may carry a local
`CONTEXT.md`; project-level documentation lives in `docs/`.

- The canonical entry point and index is `docs/README.md`.
- Status and known issues: `docs/status.md`.
- Plans and backlog: `docs/roadmap.md`.

---

# Long-Term Vision

CRDS should remain:

- lightweight
- keyboard-first
- portable
- modular
- approachable for contributors

Architectural decisions should preserve these qualities as the project evolves.
