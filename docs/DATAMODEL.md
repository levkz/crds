# Data Model

**This is the canonical home for the vocabulary file format and the SQLite
schema.** Status of the storage layer lives in `docs/status.md`; planned schema
work lives in `docs/roadmap.md`.

---

# Design Goals

The data model is built around the following principles:

- Vocabulary decks should be easy to edit by hand.
- Decks should be easy to share via Git or other version control systems.
- User progress should never modify the original deck files.
- The application should support future features such as spaced repetition,
  tags, audio, and AI-generated content without requiring changes to the file
  format.

The architecture therefore separates **content** from **state**:

```text
Vocabulary Deck (.yaml)
        │
        ▼
    Parser
        │
        ▼
    Store (SQLite cache + state)
        │
        ▼
    Application Models
```

> **Vocabulary belongs in YAML. User progress belongs in SQLite.**

---

# Vocabulary Decks

Vocabulary is stored in YAML files. Each file represents a single deck.

```yaml
id: french_a1
name: French A1
language: fr
translation_language: en

entries:
  - id: fr_bonjour
    term: bonjour

    translations:
      - text: hello
      - text: good morning

    examples:
      - text: Bonjour, Marie.
        translation: Hello, Marie.

    tags:
      - greeting
      - A1

    notes: Common greeting used throughout the day.

  - id: fr_manger
    term: manger

    translations:
      - text: eat
```

---

# Entry IDs

Each entry has a stable identifier. IDs are generated automatically when missing
(see `docs/DECK_CREATION_GUIDE.md` for details).

Generation rules:

1. The term is expanded into all its variants (see variant syntax below).
2. The shortest variant is selected.
3. It is sanitised: lowercased, spaces/apostrophes/hyphens become `_`, letters
   and digits are kept.

Examples:

```text
bonjour          → bonjour
mang[er/ez/e/ons/ent] → mange
(un)necessary    → necessary
s'il vous plaît → s_il_vous_plaît
[une/la] baguette → la_baguette
```

If the generated ID collides with an existing ID, a numeric suffix is appended:
`bonjour`, `bonjour_2`, `bonjour_3`.

---

# Application Models

The application uses separate models for vocabulary content and user progress.

## Deck

```go
type Deck struct {
    ID                  string
    Name                string
    Language            string
    TranslationLanguage string
    Entries             []Entry
}
```

## Entry

```go
type Entry struct {
    ID           string
    Term         string
    Translations []Translation
    Examples     []Example
    Tags         []string
    Notes        string
}
```

## Translation

```go
type Translation struct {
    Text string
}
```

The model intentionally remains simple. Additional metadata can be added in the
future without changing existing decks.

## Example

```go
type Example struct {
    Text        string
    Translation string
}
```

---

# Storage

## Technology Stack

| Component | Choice | Purpose |
|-----------|--------|---------|
| Database | SQLite | Local, embedded database |
| Driver | `modernc.org/sqlite` | Pure Go SQLite driver (no CGo) |
| Migrations | goose | Schema versioning and migrations |
| Query builder | sqlc | Type-safe Go code from SQL queries |
| Interface | `database/sql` | Standard library database interface |

- **SQLite**: lightweight, serverless, single-file database perfect for a
  terminal app.
- **modernc.org/sqlite**: pure Go implementation avoids CGo build complexity.
- **goose**: simple, proven migration tool with SQL-based migrations.
- **sqlc**: generates type-safe Go code from SQL, eliminating manual query
  strings and ensuring compile-time safety.

## Database Location

```text
~/.local/share/crds/crds.db
```

The database file is created automatically on first run.

## DSN Format

The connection string uses `_pragma` query parameters — **not** bare key/value
pairs. The `modernc.org/sqlite` driver ignores unrecognized parameters silently,
so `_journal_mode=WAL&_foreign_keys=on` appears to work but does nothing.

```text
crds.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)
```

Each `_pragma` value is prepended with `PRAGMA ` and executed as raw SQL on every
new connection drawn from the pool. The parenthesized form avoids URL-encoding
`=` as `%3D`.

| Parameter | Effect |
|-----------|--------|
| `_pragma=foreign_keys(1)` | Enables `ON DELETE CASCADE` and other FK constraints |
| `_pragma=journal_mode(WAL)` | Write-ahead logging for concurrent reads |
| `_pragma=busy_timeout(5000)` | Wait up to 5 s before returning `SQLITE_BUSY` |

**Why this matters:** without `PRAGMA foreign_keys = ON`, `ON DELETE CASCADE` is
silently ignored. Deleting a row in `entries` does **not** cascade to
`translations`, `examples`, or `entry_tags`, leaving orphaned rows that
accumulate on every re-sync.

> **Do not** use `db.Exec("PRAGMA foreign_keys = ON")` after opening the
> connection. With `database/sql` connection pooling, the pragma only applies to
> the single connection it runs on.

## Schema

Migrations live in `internal/storage/migrations/` (goose):

```text
20260716121051_init.sql
20260721000000_add_deck_cache.sql
20260726120000_add_reverse.sql
20260729100000_add_tags_tables.sql
```

### Sessions

Tracks quiz sessions. A session is created on the first answer and reused
until it is reset — on a deck/tag selection change and on app shutdown
(`ShutdownCmd`), so an abandoned (unfinished) session still persists its
aggregates instead of staying open with zeros. `ResetSession` counts the
session's reviews and writes `finished_at`, `reviewed`, `correct`,
`incorrect`, and `duration_ms`.

```sql
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME,
    reviewed INTEGER NOT NULL DEFAULT 0,
    correct INTEGER NOT NULL DEFAULT 0,
    incorrect INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0
);
```

### Reviews

Records individual answers during quiz sessions. `grade` uses the unified
0-3 scale (`Again=0`, `Hard=1`, `Good=2`, `Easy=3`); a review counts as
**correct when `grade >= 2`** (`GradeGood` or better) everywhere —
`ResetSession`, `Store.Stats`, and every SQL `correct_reviews`/`incorrect_reviews`
aggregate use the same threshold.

```sql
CREATE TABLE reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    deck_id TEXT NOT NULL,
    entry_id TEXT NOT NULL,
    grade INTEGER NOT NULL,
    reverse INTEGER NOT NULL DEFAULT 0,
    reviewed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Typing Details

Stores typing-specific answer data for analytics and retraining.

```sql
CREATE TABLE typing_details (
    review_id INTEGER PRIMARY KEY REFERENCES reviews(id) ON DELETE CASCADE,
    user_input TEXT NOT NULL,
    correct_answer TEXT NOT NULL,
    similarity REAL NOT NULL
);
```

- `user_input`: what the user typed
- `correct_answer`: expected translation
- `similarity`: Levenshtein similarity score (0.0–1.0)

### Progress

Per-entry learning progress for spaced repetition.

```sql
CREATE TABLE progress (
    deck_id TEXT NOT NULL,
    entry_id TEXT NOT NULL,
    reverse INTEGER NOT NULL DEFAULT 0,
    ease REAL NOT NULL DEFAULT 2.5,
    interval INTEGER NOT NULL DEFAULT 0,
    due DATETIME,
    correct INTEGER NOT NULL DEFAULT 0,
    incorrect INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (deck_id, entry_id, reverse)
);
```

### Decks and entries (cache)

YAML decks are synced into a cache: `decks` (metadata — including `language`
and the `input_mappings` trigger map serialized as JSON), `entries`, and child
`translations`, `examples`, `entry_tags`. The cache is rebuilt on sync; YAML
files remain the source of truth.

### Tags and deck_tags

For efficient deck↔tag cross-filtering (see `docs/proposals/tag_architecture.md`):

```sql
CREATE TABLE tags (
    tag TEXT PRIMARY KEY
);

CREATE TABLE deck_tags (
    deck_id TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    tag TEXT NOT NULL REFERENCES tags(tag) ON DELETE CASCADE,
    PRIMARY KEY (deck_id, tag)
);

CREATE INDEX idx_deck_tags_tag ON deck_tags(tag);
```

### Sync state

Tracks file mtimes so unchanged decks are skipped on resync:

```sql
CREATE TABLE sync_state (
    file TEXT PRIMARY KEY,
    mtime DATETIME NOT NULL
);
```

---

# Grade Scale

Both quiz types share a single unified 0–3 scale, defined as `ui.Grade` in
`internal/ui/grade.go`:

| Grade | Meaning |
|-------|---------|
| 0 (`GradeAgain`) | Again |
| 1 (`GradeHard`) | Hard |
| 2 (`GradeGood`) | Good |
| 3 (`GradeEasy`) | Easy |

The typing quiz converts its internal fuzzy grade to `ui.Grade` at the
`SaveAnswerMsg` boundary. `ProgressStore.Stats()` treats `Grade >= GradeGood` as
correct. `SaveAnswerMsg.Grade` is typed as `ui.Grade` and cast to `int` at the
storage boundary.

---

# Progress Model

```go
type Progress struct {
    DeckID   string
    EntryID  string
    Reverse  bool
    Ease     float64
    Interval int
    Due      time.Time
    Correct  int
    Incorrect int
}
```

Each progress record references an entry by deck ID + entry ID, avoiding
conflicts between different decks.

### Scheduling

Every answer persists the updated scheduling state (ease, interval, due,
outcomes) in the same transaction as the review (`internal/storage.Store` →
`persistAnswer`). The next due date is computed by `internal/scheduler` (SM-2).
The review queue for a deck/tag selection (`DueForSelection`) is: unseen cards
first (deck order), then due cards ordered by due date; entries across decks
are deduplicated.

---

# Review History

Every completed review is stored.

```go
type Review struct {
    DeckID     string
    EntryID    string
    ReviewedAt time.Time
    Grade      int
    Reverse    bool
}
```

---

# Typing Details

```go
type TypingDetail struct {
    ReviewID      int64
    UserInput     string
    CorrectAnswer string
    Similarity    float64
}
```

This enables analyzing common mistakes, retraining based on similarity scores,
and displaying weak words in statistics. `TypingDetail` exists only in
sqlc-generated code (`internal/storage/db/models.go`), not in the domain model.

---

# Session

```go
type Session struct {
    Started  time.Time
    Reviewed int
    Correct  int
    Incorrect int
    Duration time.Duration
}
```

---

# Synchronization

`Store.SyncDecks()` is called at startup:

1. Lists all `.yaml` files in `~/.local/share/crds/decks/`.
2. For each file, checks its mtime against the `sync_state` table.
3. If the file is new or changed:
   - Parses the YAML via `parser.ParseFile()`
   - Upserts deck metadata into `decks`
   - Deletes and re-inserts entries, translations, examples, and tags
   - Updates `sync_state` with the file's mtime
4. Unchanged files are skipped.

Decks removed from the filesystem are **not** deleted from the database; they
remain in the cache until the user decides to remove them. A `crds sync --prune`
flag is planned (see `docs/roadmap.md`).

---

# sqlc Queries

SQL queries are organized by domain in `internal/storage/queries/`:

```text
internal/storage/queries/
├── sessions.sql
├── reviews.sql
├── typing_details.sql
├── progress.sql
├── decks.sql
├── entries.sql
├── translations.sql
├── examples.sql
├── entry_tags.sql
├── tags.sql
└── sync_state.sql
```

Query files define SQL operations that sqlc compiles to type-safe Go functions.
Generated code lives in `internal/storage/db/` (regenerate with `sqlc generate`).
`store.go` wraps the generated queries with the public `Store` API.

Example (`queries/sessions.sql`):

```sql
-- name: CreateSession :one
INSERT INTO sessions (started_at) VALUES (CURRENT_TIMESTAMP)
RETURNING id;

-- name: FinishSession :exec
UPDATE sessions
SET finished_at = CURRENT_TIMESTAMP,
    reviewed = $2,
    correct = $3,
    incorrect = $4,
    duration_ms = $5
WHERE id = $1;
```

---

The data model is intentionally extensible. Because YAML supports nested
structures, new entry fields can be added without changing the overall
architecture. Candidate extensions are tracked in `docs/roadmap.md`.
