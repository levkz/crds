# Search Screen Revamp

> Status: **not started**. See `docs/roadmap.md`.

## Current State

`SearchModel` in `internal/ui/screens/search.go` (416 lines) is a single-column
screen with two phases:

- **Input mode**: user types a query, results filter live
- **Results mode**: user navigates results via `j`/`k`, Enter opens detail view

The screen receives `[]ui.CardData` via `SyncState(AppState)` (the current
deck's cards) and filters it in
memory. There is no deck or tag awareness — results are a flat list of
entries matching the text query.

## Goals

1. Three-column layout: decks (left), results (center), tags (right).
2. Arrow keys `h`/`l` or `←`/`→` to move between columns.
3. Selecting an item in any column highlights it and filters the other two
   columns.
4. Multiple selections possible in decks and tags columns.
5. Selecting a result opens the detail view (current behavior).
6. Overflow wraps around (cycling through columns/lists).

## Screen Layout

```
┌──────────────────────────────────────────────────────┐
│  header: "search"                                    │
├───────────┬──────────────────────────┬────────────────┤
│  Decks    │  Results                 │  Tags          │
│           │                          │                │
│  ☑ fr_a1  │  ▸ bonjour → hello      │  ☑ greeting    │
│  ☐ es_a1  │  ▸ au revoir → goodbye  │  ☐ verb        │
│  ☐ de_a1  │  ▸ merci → thank you    │  ☐ food        │
│  ☐ ja_a1  │  ▸ s'il vous plaît ...  │  ☑ A1           │
│           │                          │                │
│           │                          │                │
├───────────┴──────────────────────────┴────────────────┤
│  footer: keyboard shortcuts                          │
└──────────────────────────────────────────────────────┘
```

- Left column (decks): narrow, ~20% width
- Center column (results): primary content, ~55% width
- Right column (tags): narrow, ~25% width
- Vertical dividers between columns
- Each column has its own scroll state, cursor position, and selection state

## Column Details

### Decks Column (left)

- Shows all decks that have entries matching the current filters
- Initially shows all decks (when no query or tags are active)
- Checkboxes: user can select/deselect decks
- Selecting a deck refilters the results center column and the tags right column
- If a tag is selected in the right column, only decks with that tag are shown
- If a text query is active, only decks containing matching entries are shown

### Results Column (center)

- Shows entries that match all active filters (text query + selected decks +
  selected tags)
- Each result line: `term → translation1, translation2`
- Selected/focused result highlights (current behavior)
- `Enter` on a result opens the detail view (`NavigateToDetailMsg`)
- Scrolling: `j`/`k` within this column only when the column is active
- If no filters are active: show "Type to search vocabulary" (current behavior)

### Tags Column (right)

- Shows all tags present among the currently visible results
- Initially shows all tags (when no query or deck filters are active)
- Checkboxes: user can select/deselect tags
- Selecting a tag refilters the results column and the decks column
- If a deck is selected in the left column, only tags from that deck are shown
- If a text query is active, only tags from matching entries are shown

## Keyboard Navigation

| Key | Action |
|-----|--------|
| `h` / `←` | Move focus to the left column |
| `l` / `→` | Move focus to the right column |
| `j` / `↓` | Cursor down in the active column |
| `k` / `↑` | Cursor up in the active column |
| `enter` | In results column: open detail. In decks/tags: toggle selection |
| `space` | Toggle selection of current item in decks or tags column |
| `esc` | Clear search input / go back |
| All printable chars | Type into the search input (always visible at top) |

### Column cycling (overflow wrapping)

When the user presses `→` on the rightmost column, focus wraps to the left
(decks) column. When pressing `←` on the leftmost column, focus wraps to the
right (tags) column.

Similarly, within a list, pressing `j` on the last item wraps to the first,
and `k` on the first wraps to the last.

## Active Column Indicator

- The active column has a highlighted border or brighter header
- The focused item in the active column uses the `Selection` background color
- Items in inactive columns use muted/secondary styling
- The search input only accepts text when the center column is active (results)

Actually — reconsidering: the search input should be a global text input at the
top that always accepts input regardless of which column is active. The query
filters across all three columns. This simplifies the UX significantly.

## Data Flow

### Input: text query

The text query at the top filters all three columns simultaneously:
- Decks: only decks containing entries matching the query
- Results: entries matching the query (existing behavior)
- Tags: only tags from entries matching the query

### Input: deck selection

When a deck is toggled:
- Results: filtered to only entries in the selected deck(s)
- Tags: filtered to only tags from the selected deck(s) ∩ matching entries
- The text query (if any) further narrows both

### Input: tag selection

When a tag is toggled:
- Results: filtered to only entries with the selected tag(s)
- Decks: filtered to only decks containing entries with the selected tag(s)
- The text query (if any) further narrows both

### Combination logic (AND)

All active filters combine with AND logic:
```
visibleEntries = allCards
    .filter(card → query matches card)
    .filter(card → card.deckID in selectedDecks)
    .filter(card → card.tags contains all selectedTags)

visibleDecks = allDecks
    .filter(deck → deck has visible entries)
    // Also: if tags selected, deck must have those tags

visibleTags = allTags
    .filter(tag → tag appears in visible entries)
    // Also: if decks selected, tag must appear in those decks
```

## Implementation Strategy

### New Model

Replace the existing `SearchModel` with a new one that holds:

```go
type SearchModel struct {
    query        string
    activeColumn int // 0=decks, 1=results, 2=tags

    // Decks column
    allDecks      []string
    deckCursor    int
    deckScroll    int
    deckSelected  map[string]bool

    // Results column (center)
    allCards      []ui.CardData
    resultCursor  int
    resultScroll  int
    results       []searchEntry  // filtered view

    // Tags column
    allTags       []string
    tagCursor     int
    tagScroll     int
    tagSelected   map[string]bool

    width         int
    height        int
}
```

The `filterResults()`, `filterDecks()`, `filterTags()` methods recompute each
column's visible items based on the current state of all three filters.

### Rendering

Divide the available width into three columns with dividers:

```go
func (m *SearchModel) renderDecksColumn(width int) string { ... }
func (m *SearchModel) renderResultsColumn(width int) string { ... }
func (m *SearchModel) renderTagsColumn(width int) string { ... }
```

Combine horizontally:

```go
func (m *SearchModel) View() string {
    deckWidth := m.width * 20 / 100
    resultWidth := m.width * 55 / 100
    tagWidth := m.width - deckWidth - resultWidth - 2 // 2 for dividers

    deckCol := m.renderDecksColumn(deckWidth)
    resultCol := m.renderResultsColumn(resultWidth)
    tagCol := m.renderTagsColumn(tagWidth)

    body := lipgloss.JoinHorizontal(
        lipgloss.Top,
        deckCol,
        divider(),
        resultCol,
        divider(),
        tagCol,
    )

    return layout.Page(header, body, footer, m.height)
}
```

### Divider

Use `ui.Theme.Palette.Border` for the vertical divider style.

### Search Input

Keep a single search input at the top of the screen (above the columns),
same as the current implementation. It always accepts input.

When the user types, `filterResults()` is called (extended to also recompute
decks and tags).

### Active Column Highlighting

- The active column's header is rendered with `ui.Theme.Primary`
- Inactive columns' headers are rendered with `ui.Theme.Muted`
- The focused item in the active column uses selection background
- When results column is active, `j`/`k` scroll results
- When decks column is active, `j`/`k` scroll decks
- When tags column is active, `j`/`k` scroll tags

## Scrolling

Each column tracks its own `cursor` and `scrollOffset`. `adjustScroll()` is
called per-column when its cursor moves.

## Overflow Wrapping

Within each column:
- `↑`/`k` on first item → wraps to last item
- `↓`/`j` on last item → wraps to first item

Between columns:
- `←`/`h` on leftmost column → wraps to rightmost column
- `→`/`l` on rightmost column → wraps to leftmost column

## Backwards Compatibility

- `SyncState(AppState)` — still needed, now also populates `allCards`,
  `allDecks`, `allTags` from the global snapshot
- `OnEnter()` — still resets to initial state
- `OnLeave()` — still clears everything

## Keymap Changes

Extend `Search` keymap in `keymap.go` to support column navigation:

```go
type Search struct {
    Open       Binding
    DeleteChar Binding
    LeftColumn  Binding
    RightColumn Binding
    List                    // Up, Down, Select (embedded)
}
```

Default bindings:
- `LeftColumn`: `left`, `h`
- `RightColumn`: `right`, `l`

## File Checklist

- [ ] `internal/ui/screens/search.go` — complete rewrite to three-column model (~500-600 lines)
- [ ] `internal/ui/keymap/keymap.go` — add `LeftColumn`/`RightColumn` to `Search` group
- [ ] `internal/ui/screens/detail.go` — no changes expected (detail is downstream)
- [ ] `internal/ui/app/events.go` — `SyncState` already receives the current
      deck's cards; add per-deck/tag data to `AppState` if the revamp needs it
