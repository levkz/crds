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
    Storage
   (DeckStore)
        │
        ▼
Application Models
        │
        ├────────► Quiz Engine
        │
        └────────► ProgressStore (in-memory, future SQLite)
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

IDs are generated automatically when missing.

Format:

```text
<language>_<normalized_word>

Examples

fr_bonjour

fr_au_revoir

fr_s_il_vous_plait
```

If duplicate terms exist:

```text
fr_bonjour

fr_bonjour_2

fr_bonjour_3
```

Normalization rules:

- lowercase
- remove accents
- replace whitespace with `_`
- replace apostrophes with `_`
- remove remaining punctuation
- collapse repeated `_`

IDs should remain stable after generation.

The `sync` command is responsible for generating missing IDs.

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

## SQLite (planned)

SQLite will store only application state when implemented.

Vocabulary content remains in YAML.

Planned storage includes:

- review progress
- spaced repetition scheduling
- review history
- statistics
- session history
- cached search data

## Current implementation

An in-memory `ProgressStore` (`internal/storage/progress_store.go`) tracks progress and provides stats within a session. `DeckStore` (`internal/storage/deck_store.go`) reads YAML decks from `~/.local/share/crds/decks/`. No data is persisted to disk yet.

The YAML files remain the source of truth.

---

# Progress Model

```go
type Progress struct {
    DeckID string
    EntryID string

    Ease float64
    Interval int

    Due time.Time

    Correct int
    Incorrect int
}
```

Each progress record references an entry using:

- Deck ID
- Entry ID

This avoids conflicts between different decks.

---

# Review History

Every completed review is stored.

```go
type Review struct {
    DeckID string
    EntryID string

    Grade int

    ReviewedAt time.Time
}
```

This enables:

- statistics
- graphs
- streaks
- future scheduler improvements

---

# Session

A session represents one quiz.

```go
type Session struct {
    StartedAt time.Time

    FinishedAt time.Time

    Reviewed int

    Correct int

    Incorrect int
}
```

---

# Synchronization

The `sync` command synchronizes YAML decks with SQLite.

Responsibilities:

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
