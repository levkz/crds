# Design Decisions

This document records **why** CRDS is built the way it is. It is a decision log,
not a reference: for how the system works, see `docs/ARCHITECTURE.md` and
`docs/DATAMODEL.md`; for status and plans, see `docs/status.md` and
`docs/roadmap.md`.

---

# Goals

CRDS is a terminal-based vocabulary trainer focused on long-term language learning.

- Learn vocabulary efficiently.
- Support multiple languages.
- Be fast to use entirely from the keyboard.
- Store learning progress.
- Be easily extensible.
- Remain a single executable with minimal dependencies.

---

# Technology Stack

| Concern | Choice | Rationale |
|---|---|---|
| Language | Go 1.25+ | Single static binary, fast startup, small footprint |
| CLI | Kong | Declarative struct-tag parsing, shell completion support |
| Terminal UI | Bubble Tea + Lip Gloss | Elm-style architecture; Lip Gloss for styling |
| Configuration | Hand-rolled YAML config | Small surface; avoids a heavy dependency for ~3 files |
| Persistence | SQLite (`modernc.org/sqlite`) | Embedded, pure Go, no CGo |
| Migrations | goose | SQL-based schema versioning |
| Queries | sqlc | Type-safe Go from SQL; compile-time safety |
| Testing | Go testing package | No external framework needed |

Decisions worth noting:

- **`modernc.org/sqlite` over CGo drivers** — avoids CGo build complexity and keeps
  cross-compilation trivial.
- **goose over hand-rolled migrations** — proven, simple, SQL-first.
- **sqlc over an ORM** — type-safe queries without reflection or a runtime layer.
- **No Viper** — configuration is three small YAML files; a full config library
  is disproportionate. (Earlier drafts proposed Viper; the hand-rolled loader in
  `internal/config/` is the chosen approach.)

---

# Responsibilities

Each package owns a single responsibility:

## cli

Defines every command using Kong. The CLI package should only parse arguments.
Business logic belongs elsewhere.

## parser

Parses vocabulary files: reads YAML decks, validates structure, normalizes
whitespace, generates missing IDs, detects duplicates. Produces `model.Deck`.

## quiz

Learning engine. Starts quizzes, validates answers, normalizes text,
scores, reveals answers. Must not depend on Bubble Tea. Not yet built —
see `docs/status.md` and `docs/roadmap.md`.

## scheduler

Decides what should be shown next. Initially a simple repeat-until-known
algorithm; eventually spaced repetition (Leitner, SM-2, or FSRS). Must not know
anything about Bubble Tea. Only returns cards that should be reviewed.

## storage

Persistence layer for decks and progress. The SQLite `Store` persists reviews,
sessions, typing details, progress, and the deck cache; on startup `SyncDecks()`
syncs YAML decks into the cache using mtime checks. Legacy filesystem
(`DeckStore`) and in-memory (`ProgressStore`) backends remain but are not wired.

## importer / exporter

Imports and exports external formats (txt, csv, json, anki, quizlet). Not built —
see `docs/status.md` and `docs/roadmap.md`.

## config

Loads user configuration. Kept minimal — see the stack table above.

## ui

Contains Bubble Tea models. Should not contain quiz logic — it displays state,
collects input, and emits commands/events.

---

# Domain Model

The domain model is intentionally simple and free of business logic. It is
documented fully in `docs/DATAMODEL.md`. Key decision: **content and state are
separate types** — `Deck`/`Entry` describe vocabulary, `Progress`/`Review`/
`Session` describe learning state, and the two never mix.

---

# Design Principles

- Keep business logic independent of the terminal UI.
- Use interfaces around storage to simplify testing.
- Keep parsing, scheduling, persistence, and presentation separate.
- Prefer composition over inheritance.
- Favor small packages with clear responsibilities.
- Make it easy to add new quiz modes without changing existing code.
- Ensure every feature can be tested without requiring interactive terminal input.
- Prefer simplicity over abstraction when trade-offs arise.
