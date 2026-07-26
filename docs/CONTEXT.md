# CONTEXT.md

# Project

CRDS is a keyboard-first terminal application for learning vocabulary using flashcards and spaced repetition.

The project emphasizes:

- simplicity
- maintainability
- fast terminal workflows
- clean architecture
- human-editable vocabulary

Vocabulary is stored as YAML.

User progress is stored separately in SQLite.

---

# Purpose

CRDS is intended to be:

- a portable learning tool
- easy to extend
- pleasant to use every day
- approachable for contributors

It is **not** intended to become a general note-taking or knowledge management application.

---

# Module

Module name: `crds`

Go version: 1.25.1

Run everything from repo root.

---

# Entry Points

| Entry point | Purpose |
|---|---|
| `cmd/crds/main.go` | Main application. Kong CLI + Bubble Tea TUI. |
| `cmd/legacy-quiz/main.go` | Legacy terminal quiz. Reads `exercises/*.txt`. |

---

# Commands

| Action | Command |
|---|---|
| Build | `make build` |
| Install | `make install` |
| Run | `make run` |
| All tests | `make test` |
| Single package | `go test ./internal/parser/` |
| Lint | `make lint` (requires `golangci-lint`) |
| Tidy | `make tidy` |
| Build legacy quiz | `make legacy` |

---

# Documentation Index

## Project Documentation

| Document | Purpose |
|---|---|---|
| `AGENTS.md` | Instructions for AI coding agents and contributors |
| `docs/ARCHITECTURE.md` | High-level system architecture |
| `docs/DATAMODEL.md` | Vocabulary and persistence model |
| `docs/DECK_CREATION_GUIDE.md` | How to create and manage vocabulary decks |
| `docs/CONTEXT.md` | This document |
| `docs/DESIGN.md` | Design decisions |

---

## Subsystem Documentation

| Subsystem | Documentation | Status |
|---|---|---|
| Parser | — | No dedicated docs yet |
| Storage | — | SQLite fully implemented via `Store` (goose + sqlc). `DeckStore` reads YAML decks. `StateStore` persists selected decks. |
| Quiz | — | Package does not exist yet |
| Scheduler | — | Package does not exist yet |
| Search | — | Package does not exist yet |
| UI | `internal/ui/docs/ARCHITECTURE.md`, `internal/ui/docs/CONTEXT.md` | Comprehensive |

---

# Repository Organization

```text
cmd/
    Application entry points

internal/
    Application source code

docs/
    Project-level design decisions and documentation

exercises/
    Legacy vocabulary files (.txt format)
```

Most implementation work happens inside `internal/`.

---

# Current Implementation Status

## Complete

| Package | Description | Tests |
|---|---|---|
| `internal/model/` | Domain types: Deck, Entry, Progress, Review, Session | — |
| `internal/parser/` | YAML parsing, validation, normalization, auto-ID generation | 13 test fixtures in `testdata/` |
| `internal/config/` | User configuration from `~/.config/crds/` | 13 tests |
| `internal/ui/theme/` | Design system: 15-color palette, typography, icons, borders, spacing, 4 built-in themes | 68+ tests |
| `internal/ui/styles/` | 12 semantic style definitions | 60 tests |
| `internal/ui/components/` | 29 components (display + interactive) | — |
| `internal/ui/navigation/` | Stack-based navigation with push/pop/replace/forward/modal | 82 tests |
| `internal/ui/keymap/` | Centralized keybinding definitions with user overrides | 16 tests |
| `internal/ui/layout/` | Layout primitives: Page, Column, Row, Grid, Stack, Spacer, Center, Align | 258 lines of tests |
| `internal/ui/renderer/` | Terminal rendering utilities: width, ANSI, wrapping | Implemented |
| `internal/ui/app/` | Root Bubble Tea model, event dispatch, lifecycle, commands | — |
| `internal/ui/screens/` | 8 screens: Home, Quiz, Typing Quiz, Search, Statistics, Settings, Detail, Decks | — |
| `internal/fuzzy/` | Levenshtein-based fuzzy string matching for grading typed answers | 8 tests |

## Partially Implemented

| Package | Status |
|---|---|
| `internal/storage/` | SQLite fully implemented via `Store` (goose + sqlc). `DeckStore` reads YAML decks. `StateStore` persists selected decks. Legacy `ProgressStore` remains but is not wired. |
| `internal/cli/` | Kong command stubs. Most `Run()` methods only print. |
| `internal/app/` | Empty composition root struct. Aspirational. |

## Not Implemented

| Package | Status |
|---|---|
| `internal/quiz/` | Does not exist. Quiz logic is in UI screens only. |
| `internal/scheduler/` | Does not exist. No spaced repetition. |
| `internal/search/` | Does not exist. Search is implemented in UI screens only. |
| `internal/ui/actions/` | Empty. Deferred until command palette or mouse support. |
| `internal/ui/events/` | 4 basic event types. Mostly deferred. |
| `internal/ui/animations/` | Empty placeholder. |
| `internal/ui/debug/` | Empty placeholder. |

---

# Known Issues

- Deck selection screen exists but empty selection means no quiz (must pick at least one deck)
- CLI commands (quiz, sync, stats, search) are stubs — only the TUI launches
- Grade scale mismatch: Flashcard uses 0-3, Typing uses 1-3 (needs normalization)

---

# Architecture Overview

The system is organized into independent subsystems.

```text
CLI
    ↓
Application
    ↓
Services
    ↓
Model
    ↓
Infrastructure
```

See `docs/ARCHITECTURE.md` for a complete description.

---

# Data Overview

Two kinds of data exist:

Vocabulary

- YAML
- version controlled
- human editable

User state

- SQLite (via `modernc.org/sqlite`)
- goose (migrations)
- sqlc (type-safe query generation)
- local
- generated

See `docs/DATAMODEL.md` for details.

---

# Development Workflow

When implementing a feature:

1. Read the relevant subsystem documentation.
2. Keep responsibilities separated.
3. Write or update tests.
4. Update documentation if architecture changes.

---

# Style

- Prefer table-driven tests with `testdata/` fixtures
- One responsibility per package; no circular dependencies
- Explicit over implicit; small functions; early returns
- Avoid unnecessary interfaces and global state
- Never hardcode colors — use semantic theme values
- Keyboard first — every visible action has a shortcut

---

# For Contributors and AI Agents

Before modifying a subsystem:

1. Read this document.
2. Read `docs/ARCHITECTURE.md`.
3. Read the subsystem's documentation (if it exists).
4. Follow the style guidelines.
5. Update documentation when introducing architectural changes.

The goal is to keep the project understandable, consistent, and easy to extend.
