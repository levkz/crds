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

| Subsystem | Responsibility                              |
| --------- | ------------------------------------------- |
| CLI       | Command-line interface and command dispatch |
| Parser    | Loading and validating vocabulary files     |
| Storage   | Persisting user-specific state              |
| Quiz      | Learning session orchestration              |
| Scheduler | Determining review order                    |
| Search    | Vocabulary lookup                           |
| UI        | Terminal presentation                       |
| Model     | Shared domain objects                       |

Each subsystem has its own `ARCHITECTURE.md` describing its internal design.

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

Examples:

```text
internal/parser/
internal/storage/
internal/quiz/
internal/ui/
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
under each susbsystem there's a `docs/` directory with `CONTEXT.md` and `ARCHITECTURE.md` (other .md files can possibly be there)

This document describes the overall system.

Additional documentation is located alongside the relevant subsystem.

Examples:

```text
internal/parser/docs/
internal/storage/docs/
internal/ui/docs/
```

---

# Long-Term Vision

CRDS should remain:

- lightweight
- keyboard-first
- portable
- modular
- approachable for contributors

Architectural decisions should preserve these qualities as the project evolves.
