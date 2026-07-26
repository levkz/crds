# Data Model

This document describes how vocabulary content and user progress are represented within the application.

## Design Goals

The data model is built around the following principles:

- Vocabulary decks should be easy to edit by hand.
- Decks should be easy to share via Git or other version control systems.
- User progress should never modify the original deck files.
- The application should support future features such as spaced repetition, tags, audio, and AI-generated content without requiring changes to the file format.

The architecture therefore separates **content** from **state**.

```text
Vocabulary Deck (.yaml)
        │
        ▼
    Parser
        │
        ▼
    DeckStore (filesystem)
        │
        ▼
    Application Models
        │
        ├────────► Quiz Engine
        │
        └────────► SQLiteStore (goose + sqlc)
                       │
                       ▼
                   SQLite DB
```

---

# Vocabulary Decks

Vocabulary is stored in YAML files.

Each file represents a single deck.

Example:

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

Each entry has a stable identifier.

IDs are generated automatically when missing (see `DECK_CREATION_GUIDE.md` for details).

Generation rules:

1. The term is expanded into all its variants (see variant syntax below).
2. The shortest variant is selected.
3. It is sanitised: lowercased, spaces/apostrophes/hyphens become `_`, letters and digits are kept.

Examples:

```text
bonjour          → bonjour
mang[er/ez/e/ons/ent] → mange
(un)necessary    → necessary
s'il vous plaît → s_il_vous_plaît
[une/la] baguette → la_baguette
```

If the generated ID collides with an existing ID, a numeric suffix is appended: `bonjour`, `bonjour_2`, `bonjour_3`.

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

---

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

---

## Translation

```go
type Translation struct {
    Text string
}
```

The model intentionally remains simple.

Additional metadata can be added in the future without changing existing decks.

---

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

### Why this stack?

- **SQLite**: Lightweight, serverless, single-file database perfect for a terminal app
- **modernc.org/sqlite**: Pure Go implementation avoids CGo build complexity
- **goose**: Simple, proven migration tool with SQL-based migrations
- **sqlc**: Generates type-safe Go code from SQL queries, eliminating manual query strings and ensuring compile-time safety

## Database Location

```text
~/.local/share/crds/crds.db
```

The database file is created automatically on first run.

## Schema

### Sessions

Tracks quiz sessions.

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

Records individual answers during quiz sessions.

```sql
CREATE TABLE reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    deck_id TEXT NOT NULL,
    entry_id TEXT NOT NULL,
    grade INTEGER NOT NULL,
    reviewed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Grade values:

| Quiz Type | Values | Meaning |
|-----------|--------|---------|
| Flashcard | 0-3 | Again(0), Hard(1), Good(2), Easy(3) |
| Typing | 1-3 | Again(1), Hard(2), Good(3) |

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

Fields:

- `user_input`: What the user typed
- `correct_answer`: Expected translation
- `similarity`: Levenshtein similarity score (0.0-1.0)

### Progress

Per-entry learning progress for spaced repetition.

```sql
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

## Current implementation

Three storage backends coexist:

| Backend | What It Stores | Persistence |
|---------|---------------|-------------|
| `DeckStore` | Deck YAML files | Filesystem (`~/.local/share/crds/decks/`) |
| `Store` (SQLite) | Reviews, sessions, typing details, progress | SQLite (`crds.db`) |
| `StateStore` | Selected decks list | YAML file (`state.yaml`) |

The in-memory `ProgressStore` is a legacy predecessor. The SQLite `Store` has fully replaced it for the wired path in `cli/root.go`. However, `ProgressStore` remains in the codebase with its tests passing.

The YAML files remain the source of truth for vocabulary.

---

# Progress Model

Domain model (plain struct, no DB fields):

```go
type Progress struct {
    DeckID   string
    EntryID  string
    Ease     float64
    Interval int
    Due      time.Time
    Correct  int
    Incorrect int
}
```

sqlc-generated model (with DB tags):

```go
type Progress struct {
    DeckID    string
    EntryID   string
    Ease      float64
    Interval  int64
    Due       sql.NullTime
    Correct   int64
    Incorrect int64
}
```

Each progress record references an entry using:

- Deck ID
- Entry ID

This avoids conflicts between different decks.

---

# Review History

Every completed review is stored.

Domain model (plain struct):

```go
type Review struct {
    DeckID     string
    EntryID    string
    ReviewedAt time.Time
    Grade      int
}
```

sqlc-generated model (with DB tags):

```go
type Review struct {
    ID         int64
    SessionID  int64
    DeckID     string
    EntryID    string
    Grade      int64
    ReviewedAt time.Time
}
```

Grade values:

| Quiz Type | Scale | Meaning |
|-----------|-------|---------|
| Flashcard | 0-3 | Again(0), Hard(1), Good(2), Easy(3) |
| Typing | 1-3 | Again(1), Hard(2), Good(3) |

---

# Typing Details

For typing quizzes, additional data is captured.

sqlc-generated model (no domain model equivalent):

```go
type TypingDetail struct {
    ReviewID      int64
    UserInput     string
    CorrectAnswer string
    Similarity    float64
}
```

This enables:

- analyzing common mistakes
- retraining based on similarity scores
- displaying weak words in statistics

---

# Session

A session represents one quiz.

Domain model (plain struct):

```go
type Session struct {
    Started  time.Time
    Reviewed int
    Correct  int
    Incorrect int
    Duration time.Duration
}
```

sqlc-generated model (with DB tags):

```go
type Session struct {
    ID         int64
    StartedAt  time.Time
    FinishedAt sql.NullTime
    Reviewed   int64
    Correct    int64
    Incorrect  int64
    DurationMs int64
}
```

---

# Synchronization

Decks are automatically synced from YAML to SQLite on application startup.

## Sync on startup

`Store.SyncDecks()` is called in `cli/root.go` during startup:

1. Lists all `.yaml` files in `~/.local/share/crds/decks/`
2. For each file, checks its mtime against the `sync_state` table
3. If the file is new or has changed since last sync:
   - Parses the YAML via `parser.ParseFile()`
   - Upserts deck metadata into `decks` table
   - Deletes and re-inserts entries, translations, examples, and tags
   - Updates `sync_state` with the file's mtime
4. Unchanged files are skipped

## Stale decks

Decks removed from the filesystem are not deleted from the database.
They remain in the cache until the user decides to remove them.

Future releases may add a `sync` CLI command with cleanup flags:

```
crds sync          # sync all decks
crds sync --prune  # sync and remove stale decks from cache
```

## Responsibilities

- parse every deck
- validate structure
- generate missing IDs
- detect duplicate IDs
- detect duplicate terms
- rebuild search cache
- update metadata

Vocabulary data is never generated from SQLite.

SQLite is considered a cache and state store.

---

# sqlc Queries

SQL queries are organized by domain in `internal/storage/queries/`:

```text
internal/storage/
├── migrations/
│   ├── 20260716121051_init.sql
│   └── 20260721000000_add_deck_cache.sql
├── queries/
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
├── sqlc.yaml
├── db/            (generated by sqlc)
│   ├── db.go
│   ├── models.go
│   ├── sessions.sql.go
│   ├── reviews.sql.go
│   ├── progress.sql.go
│   ├── typing_details.sql.go
│   ├── decks.sql.go
│   ├── entries.sql.go
│   ├── translations.sql.go
│   ├── examples.sql.go
│   ├── entry_tags.sql.go
│   └── sync_state.sql.go
├── store.go       (manual wrapper)
├── sync.go        (deck sync + cache queries)
└── schema.sql     (standalone schema reference)
```

Query files define SQL operations that sqlc compiles to type-safe Go functions.

Example query file (`queries/sessions.sql`):

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

After running `sqlc generate`, these become Go methods on a `Queries` struct:

```go
func (q *Queries) CreateSession(ctx context.Context) (db.Session, error)
func (q *Queries) FinishSession(ctx context.Context, arg FinishSessionParams) error
```

### Available Queries

| Query | Type | Description |
|-------|------|-------------|
| `CreateSession` | :one | Insert new session, return all fields |
| `FinishSession` | :exec | Update session with review counts and duration |
| `GetSession` | :one | Get session by ID |
| `CreateReview` | :one | Insert new review, return all fields |
| `GetReviewsBySession` | :many | Get all reviews for a session |
| `GetReviewsByEntry` | :many | Get last N reviews for an entry |
| `GetTodayStats` | :one | Aggregate today's reviews and correct count |
| `GetTodayStatsByDeck` | :one | Same as above, filtered by deck |
| `GetWeakTypingEntries` | :many | Get weakest typed answers for a deck |
| `CreateTypingDetail` | :exec | Insert typing detail |
| `GetTypingDetailByReview` | :one | Get typing detail by review ID |
| `GetProgress` | :one | Get progress by deck and entry |
| `UpsertProgress` | :exec | Insert or update progress |
| `GetDueCards` | :many | Get cards due for review in a deck |
| `UpsertDeck` | :exec | Insert or update a deck cache entry |
| `ListDeckNames` | :many | List all cached deck IDs and names |
| `GetDeck` | :one | Get deck metadata by ID |
| `UpsertEntry` | :exec | Insert or update an entry cache entry |
| `ListEntriesByDeck` | :many | List all entries for a deck |
| `GetEntry` | :one | Get entry by ID |
| `DeleteEntriesByDeck` | :exec | Remove all entries for a deck |
| `InsertTranslation` | :exec | Insert a translation for an entry |
| `DeleteTranslationsByEntry` | :exec | Remove all translations for an entry |
| `GetTranslationsByEntry` | :many | List translations for an entry |
| `InsertExample` | :exec | Insert an example for an entry |
| `DeleteExamplesByEntry` | :exec | Remove all examples for an entry |
| `GetExamplesByEntry` | :many | List examples for an entry |
| `InsertEntryTag` | :exec | Add a tag to an entry |
| `DeleteTagsByEntry` | :exec | Remove all tags for an entry |
| `GetTagsByEntry` | :many | List tags for an entry |
| `GetLastModified` | :one | Get last sync mtime for a file |
| `UpsertSyncState` | :exec | Record a file's sync mtime |

---

# Future Extensions

The data model is intentionally extensible.

Possible future additions include:

- pronunciation
- gender
- plural forms
- irregular conjugations
- audio
- images
- difficulty levels
- CEFR levels
- synonyms
- antonyms
- etymology
- AI-generated examples
- references to grammar topics

Because YAML supports nested structures, these features can be added without changing the overall architecture.

---

# Guiding Principle

The project follows a simple rule:

> **Vocabulary belongs in YAML. User progress belongs in SQLite.**

This keeps decks portable, version-controllable, and easy to share while allowing the application to efficiently track each user's individual learning progress.
