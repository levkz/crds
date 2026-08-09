# Tag Architecture

> Status: **partially implemented** — storage layer (`entry_tags` tables and
> queries) done; deck-level aggregation and deck/tag filtering UI deferred.
> See `docs/status.md` and `docs/roadmap.md`.

## Current State

Tags are stored per-entry in the `entry_tags` table (composite PK `entry_id, tag`).
There is no dedicated `tags` table, no `deck_tags` linking table, and no way to:

- List all unique tags across the entire corpus
- List tags belonging to a deck without scanning all entries
- Filter decks by tags
- Compute the intersection of tags when decks are selected (or vice-versa)

The `syncDeck()` function in `internal/storage/sync.go` inserts tags one at a
time via `InsertEntryTag` but never aggregates them into a deck-level structure.

## Goals

1. Support efficient deck↔tag cross-filtering — selecting a deck narrows the
   tag list and selecting a tag narrows the deck list, both at minimal
   computational cost.
2. Persist unique tags in the DB so queries can return them without
   scanning every entry.
3. Keep the sync path clean — one pass through entries updates all tag
   relationships.

## Schema Changes

### New table: `tags`

Stores every unique tag string. Referenced by `deck_tags`.

```sql
CREATE TABLE tags (
    tag TEXT PRIMARY KEY
);
```

This is a lightweight dictionary table. Tags are short strings like
`"greeting"`, `"A1"`, `"verb"`.

### New table: `deck_tags`

Links a deck to all tags that appear among its entries. This is the core
of efficient cross-filtering.

```sql
CREATE TABLE deck_tags (
    deck_id TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    tag TEXT NOT NULL REFERENCES tags(tag) ON DELETE CASCADE,
    PRIMARY KEY (deck_id, tag)
);

CREATE INDEX idx_deck_tags_tag ON deck_tags(tag);
```

The composite PK allows fast lookups in both directions:
- `SELECT tag FROM deck_tags WHERE deck_id = ?` → tags for a deck
- `SELECT deck_id FROM deck_tags WHERE tag = ?` → decks that have a tag

### Why not derive everything from `entry_tags`?

The query for "all tags for deck X" without `deck_tags` is:

```sql
SELECT DISTINCT tag FROM entry_tags
JOIN entries ON entries.id = entry_tags.entry_id
WHERE entries.deck_id = ?
```

This works but scans `entries` and aggregates distinct tags across potentially
thousands of entries every time the user navigates the deck selection screen.
With `deck_tags`, it's a single PK lookup.

The same applies in reverse: "all decks that have tag Y" without `deck_tags`
requires joining through `entries` and `entry_tags`. With `deck_tags`, it's
a direct index scan on `idx_deck_tags_tag`.

## New sqlc Queries

Add to `queries/entry_tags.sql` (or a new `queries/tags.sql`):

```sql
-- name: InsertTag :exec
INSERT OR IGNORE INTO tags (tag) VALUES (?);

-- name: InsertDeckTag :exec
INSERT OR IGNORE INTO deck_tags (deck_id, tag) VALUES (?, ?);

-- name: DeleteDeckTags :exec
DELETE FROM deck_tags WHERE deck_id = ?;

-- name: ListDeckTags :many
SELECT tag FROM deck_tags WHERE deck_id = ? ORDER BY tag;

-- name: ListAllTags :many
SELECT tag FROM tags ORDER BY tag;

-- name: ListDecksByTag :many
SELECT deck_id FROM deck_tags WHERE tag = ? ORDER BY deck_id;

-- name: ListDeckTagsByDecks :many
SELECT deck_id, tag FROM deck_tags WHERE deck_id IN (?) ORDER BY deck_id, tag;
-- Note: sqlc handles IN with a variadic slice via `IN (/* SLICE:deck_ids*/)` syntax.
```

## Sync Changes (`sync.go`)

During `syncDeck()`, after processing all entries but before `UpsertSyncState`,
add:

```go
// Collect all unique tags for this deck
tagSet := make(map[string]bool)
for _, entry := range deck.Entries {
    for _, tag := range entry.Tags {
        tagSet[tag] = true
    }
}

// 1. Ensure every tag exists in the tags dictionary
for tag := range tagSet {
    if err := s.queries.InsertTag(ctx, tag); err != nil {
        return fmt.Errorf("insert tag %q: %w", tag, err)
    }
}

// 2. Bulk-replace deck_tags for this deck
if err := s.queries.DeleteDeckTags(ctx, deck.ID); err != nil {
    return fmt.Errorf("delete deck tags: %w", err)
}
for tag := range tagSet {
    if err := s.queries.InsertDeckTag(ctx, db.InsertDeckTagParams{
        DeckID: deck.ID,
        Tag:    tag,
    }); err != nil {
        return fmt.Errorf("insert deck tag: %w", err)
    }
}
```

The `entry_tags` rows (per-entry) remain untouched — they are still needed
for per-entry tag display in quiz and search results. `deck_tags` is a
summary layer on top.

## New Store Methods

```go
package storage

// ListAllTags returns every unique tag across all decks.
func (s *Store) ListAllTags() ([]string, error)

// ListDeckTags returns all tags belonging to a single deck.
func (s *Store) ListDeckTags(deckID string) ([]string, error)

// ListDecksByTag returns deck IDs that have the given tag.
func (s *Store) ListDecksByTag(tag string) ([]string, error)

// FilterDecksByTags returns deck IDs that have ALL of the given tags (AND logic).
func (s *Store) FilterDecksByTags(tags []string) ([]string, error)

// FilterTagsByDecks returns tags that belong to ALL of the given decks (intersection).
func (s *Store) FilterTagsByDecks(deckIDs []string) ([]string, error)
```

### Cross-filtering implementation

`FilterDecksByTags`:
```sql
SELECT deck_id FROM deck_tags
WHERE tag IN (?)
GROUP BY deck_id
HAVING COUNT(DISTINCT tag) = ?
```
Returns decks that have every requested tag (AND). For OR semantics, remove
the HAVING clause.

`FilterTagsByDecks`:
```sql
SELECT tag FROM deck_tags
WHERE deck_id IN (?)
GROUP BY tag
HAVING COUNT(DISTINCT deck_id) = ?
```
Returns tags that appear in every selected deck (intersection). For union,
remove the HAVING clause.

## Migration

New goose migration, e.g. `20260729100000_add_tags_tables.sql`:

```sql
-- +goose Up
CREATE TABLE tags (
    tag TEXT PRIMARY KEY
);

CREATE TABLE deck_tags (
    deck_id TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    tag TEXT NOT NULL REFERENCES tags(tag) ON DELETE CASCADE,
    PRIMARY KEY (deck_id, tag)
);

CREATE INDEX idx_deck_tags_tag ON deck_tags(tag);

-- Populate tags from existing entry_tags
INSERT OR IGNORE INTO tags (tag) SELECT DISTINCT tag FROM entry_tags;

-- Populate deck_tags from existing entries + entry_tags
INSERT OR IGNORE INTO deck_tags (deck_id, tag)
SELECT DISTINCT e.deck_id, et.tag
FROM entry_tags et
JOIN entries e ON e.id = et.entry_id;

-- +goose Down
DROP TABLE IF EXISTS deck_tags;
DROP TABLE IF EXISTS tags;
```

## UI Integration Points

The `DecksScreen` calls the new store methods through the `TagProvider`
interface to populate its tag column.

Tag data flows:
- On app init: `ListAllTags()` populates the tag column in deck selection
- When a deck is toggled: `FilterTagsByDecks(selectedDecks)` narrows the tag list
- When a tag is toggled: `FilterDecksByTags(selectedTags)` narrows the deck list
- Search screen: `FilterTagsByDecks(matchedDeckIDs)` narrows the tag column
  after a text query produces a subset of decks

## File Checklist

- [x] `internal/storage/migrations/20260729100000_add_tags_tables.sql` — new migration
- [x] `internal/storage/queries/tags.sql` — new sqlc queries
- [x] `internal/storage/db/` — regenerate via `sqlc generate`
- [x] `internal/storage/sync.go` — update `syncDeck()` to populate `tags` and `deck_tags`
- [x] `internal/storage/store.go` — new `Store` methods for tag queries
- [ ] `internal/ui/app/dependencies.go` — add `TagProvider` interface if needed (deferred — UI integration not yet wired)
- [x] `internal/storage/store_test.go` — test new methods
