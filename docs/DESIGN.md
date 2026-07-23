# Vocabulary Trainer Design Document

## Goals

The application is a terminal-based vocabulary trainer focused on long-term language learning.

Primary goals:

- Learn vocabulary efficiently.
- Support multiple languages.
- Be fast to use entirely from the keyboard.
- Store learning progress.
- Be easily extensible.
- Remain a single executable with minimal dependencies.

---

# Technology Stack

Language

- Go 1.25+

CLI

- Kong

Terminal UI

- Bubble Tea
- Bubbles
- Lip Gloss

Configuration

- Viper

Persistence

- SQLite (via `modernc.org/sqlite`)
- goose (migrations)
- sqlc (type-safe query generation)
- `database/sql` (standard library interface)

Testing

- Go testing package

---

# Project Layout

```text
vocab/

├── cmd/
│   ├── crds/            Main app (Kong CLI + Bubble Tea TUI)
│   └── legacy-quiz/     Legacy terminal quiz (exercises/*.txt)
│
├── internal/
│   ├── cli/
│   │
│   ├── model/
│   │
│   ├── parser/
│   │
│   ├── storage/
│   │
│   ├── quiz/            [not yet implemented]
│   │
│   ├── scheduler/       [not yet implemented]
│   │
│   ├── search/          [not yet implemented]
│   │
│   ├── config/
│   │
│   └── ui/
│
├── exercises/
│
├── docs/
│
├── Makefile
├── go.mod
└── go.sum
```

---

# Responsibilities

## cli

Defines every command using Kong.

Example

```
vocab quiz

vocab review

vocab add

vocab search maison

vocab stats

vocab import french.txt
```

The CLI package should only parse arguments.

Business logic belongs elsewhere.

---

## parser

Responsible for parsing vocabulary files.

Responsibilities

- read exercise files
- ignore comments
- parse descriptions
- support multiple accepted translations
- validation
- duplicate detection

Example input

```
bonjour => hello / good morning => common greeting

manger => eat

chat => cat
```

Produces

```go
Entry
```

objects.

---

## quiz

Contains the learning engine.

Responsibilities

- start quiz
- validate answers
- normalize text
- scoring
- multiple quiz modes
- statistics
- reveal answers
- keyboard shortcuts

Future quiz modes

- French → English

- English → French

- Multiple choice

- Type answer

- Timed

- Listening

- Review mistakes

---

## scheduler

Responsible for deciding what should be shown next.

Initially

Simple repeat-until-known algorithm.

Eventually

Spaced repetition.

Possible algorithms

- Leitner
- SM-2
- FSRS

This package should not know anything about Bubble Tea.

It only returns cards that should be reviewed.

---

## storage

Persistence layer for decks and progress.

Current implementation:
- `DeckStore` reads YAML decks from `~/.local/share/crds/decks/` (legacy)
- `Store` (SQLite) persists reviews, sessions, typing details, progress, and deck cache
- `StateStore` persists selected decks to YAML file

On startup, `SyncDecks()` syncs YAML decks into the SQLite cache using mtime checks.
The SQLite `Store` then serves as the `DeckProvider`, `ProgressRecorder`, and `StatsProvider`.

SQLite stack:
- `modernc.org/sqlite` for pure Go SQLite driver
- `goose` for schema migrations
- `sqlc` for type-safe query generation
- `database/sql` standard library interface

Database location: `~/.local/share/crds/crds.db`

Schema includes:
- `sessions` — quiz session tracking
- `reviews` — individual answer records
- `typing_details` — typing quiz specifics (user input, similarity)
- `progress` — per-entry spaced repetition state
- `decks` — deck metadata cache
- `entries` — entry cache with translations, examples, tags
- `sync_state` — file mtime tracking for incremental sync

---

## importer

Imports external formats.

Examples

- txt
- csv
- json
- anki
- quizlet

---

## exporter

Exports vocabulary.

Possible formats

- txt

- csv

- json

- anki

---

## config

Loads configuration.

Examples

```
language = "fr"

theme = "dark"

scheduler = "fsrs"

showDescriptions = true
```

Uses Viper.

---

## ui

Contains Bubble Tea models.

Responsible for

- quiz screen
- menus
- search
- statistics
- progress bars
- confirmation dialogs

Should not contain quiz logic.

---

## util

Small reusable helpers.

Examples

- string normalization
- terminal helpers
- file utilities

---

# Domain Model

## Entry

```go
type Entry struct {
    ID           int64
    Term         string
    Answers      []string
    Description  string
}
```

---

## Card

Represents a learnable item.

```go
type Card struct {
    Entry

    Ease float64

    Interval int

    Due time.Time

    Repetitions int
}
```

---

## Session

Represents one quiz.

```go
type Session struct {
    Started time.Time

    Reviewed int

    Correct int

    Incorrect int

    Duration time.Duration
}
```

---

# Commands

```
vocab quiz

vocab quiz french

vocab review

vocab add

vocab edit

vocab remove

vocab search

vocab list

vocab stats

vocab import

vocab export

vocab config
```

---

# Storage

## SQL Stack

| Component | Choice | Purpose |
|-----------|--------|---------|
| Database | SQLite | Local, embedded database |
| Driver | `modernc.org/sqlite` | Pure Go SQLite driver (no CGo) |
| Migrations | goose | Schema versioning and migrations |
| Query builder | sqlc | Type-safe Go code from SQL queries |
| Interface | `database/sql` | Standard library database interface |

## Schema

```sql
-- Quiz sessions
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME,
    reviewed INTEGER NOT NULL DEFAULT 0,
    correct INTEGER NOT NULL DEFAULT 0,
    incorrect INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0
);

-- Individual review records
CREATE TABLE reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    deck_id TEXT NOT NULL,
    entry_id TEXT NOT NULL,
    grade INTEGER NOT NULL,
    reviewed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Typing quiz details
CREATE TABLE typing_details (
    review_id INTEGER PRIMARY KEY REFERENCES reviews(id) ON DELETE CASCADE,
    user_input TEXT NOT NULL,
    correct_answer TEXT NOT NULL,
    similarity REAL NOT NULL
);

-- Per-entry progress
CREATE TABLE progress (
    deck_id TEXT NOT NULL,
    entry_id TEXT NOT NULL,
    ease REAL NOT NULL DEFAULT 2.5,
    interval INTEGER NOT NULL DEFAULT 0,
    due DATETIME,
    correct INTEGER NOT NULL DEFAULT 0,
    incorrect INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (deck_id, entry_id)
);
```

## Grade Values

| Quiz Type | Scale | Meaning |
|-----------|-------|---------|
| Flashcard | 0-3 | Again(0), Hard(1), Good(2), Easy(3) |
| Typing | 1-3 | Again(1), Hard(2), Good(3) |

Future support

- multiple decks
- categories
- language pairs

---

# Future Features

## Spaced repetition

FSRS scheduler.

---

## Audio

Play pronunciation.

---

## Images

Attach images to vocabulary.

---

## AI

Generate example sentences.

Explain difficult grammar.

Generate quizzes automatically.

---

## TUI Dashboard

Home screen showing

- reviews due
- streak
- accuracy
- recently learned
- progress

---

## Search

Instant fuzzy search.

---

## Tags

Examples

```
food

travel

verbs

adjectives

A2

B1

B2
```

---

## Statistics

Daily reviews

Learning streak

Accuracy

Average response time

Cards due

Mastered cards

Weakest words

---

# Design Principles

- Keep business logic independent of the terminal UI.
- Use interfaces around storage to simplify testing.
- Keep parsing, scheduling, persistence, and presentation separate.
- Prefer composition over inheritance.
- Favor small packages with clear responsibilities.
- Make it easy to add new quiz modes without changing existing code.
- Ensure every feature can be tested without requiring interactive terminal input.
