# Deck Selection Screen

> Status: **implemented** — `internal/ui/screens/deck_select.go`. See
> `docs/status.md` and `docs/roadmap.md`.

## Current State

Single unified screen in `internal/ui/screens/deck_select.go` (~500 lines).

Replaces both the old full-screen `DecksModel` (`decks.go`) and the overlay
approach (`overlay_decks.go`). Split layout with two columns: decks (left)
and tags (right), each with inline search, toggle, and scrolling.

## Screen Layout

```
┌─────────────────────────────────────────────┐
│  header: "Deck Selection"                   │
├────────────────────┬────────────────────────┤
│  Decks             │  Tags                  │
│  ┌──────────────┐  │  ┌──────────────────┐  │
│  │ search: ____ │  │  │ search: ________ │  │
│  └──────────────┘  │  └──────────────────┘  │
│                    │                        │
│  ☐ french_a1      │  ☐ greeting            │
│  ☑ spanish_a1     │  ☑ A1                  │
│  ☑ german_a1      │  ☐ verb                │
│  ☐ japanese_a1    │  ☑ food                │
│  ...              │  ...                    │
│                    │                        │
├────────────────────┴────────────────────────┤
│  footer: keyboard shortcuts                 │
└─────────────────────────────────────────────┘
```

- Columns are centered with 4-char margin from left/right edges
- Vertical divider rendered in `ui.Theme.Secondary`
- Each column has its own scroll state, cursor, and search field
- Screen height is stable — body fills `height-4` lines regardless of content

## Behaviors

### Column State

Each column (`columnState`) manages:
- `cursor` — current item index in filtered list
- `items` — list of names (deck IDs or tag names)
- `selected` — map of name → bool
- `searchActive` — whether search input is shown
- `searchQuery` — current filter text
- `scrollOffset` — for scrollable list viewport

### Cross-Filtering

Both columns filter using **OR** logic for selections:
- `filteredDecks()` shows decks that are (a) selected OR match any selected tag
- `filteredTags()` shows tags that are (a) selected OR appear in any selected deck

This means selecting any deck shows all tags across all selected decks, and
selecting any tag shows all decks that have any selected tag. Selections are
always visible regardless of filters.

### Search

- Press `s` to toggle search input in the active column
- Typing filters items in real time (case-insensitive substring match)
- `Enter` confirms the search (deactivates search, cursor goes to filtered list)
- `Esc` with non-empty query clears the query and resets cursor
- `Esc` with empty query deactivates search (letting global handler navigate back)
- `Backspace` removes last character of the query

### Keyboard Navigation

| Key | Action |
|-----|--------|
| `↑`/`k` | Cursor up in active column |
| `↓`/`j` | Cursor down in active column |
| `tab` | Move focus to next column |
| `shift+tab` | Move focus to previous column |
| `s` | Toggle search input in active column |
| `space` | Toggle selection of current item |
| `a` | Toggle all in active column |
| `enter` | In search: confirm query. In list: confirm selection |
| `esc` | Via `BackHandler`: clear search / dismiss screen |

### BackHandler

`DeckSelectModel` implements `ui.BackHandler` to intercept Esc before the
global handler. If the active column has search active:
- Non-empty query → clears the query, stays on screen (returns `true`)
- Empty query → deactivates search, returns `false` to let global handler
  navigate back/dismiss the screen

Outside search, `HandleBack()` returns `false` — the global handler navigates
back or to Home.

### Scrolling

Each column independently scrolls when its cursor exceeds the visible area.
`adjustScroll()` keeps the cursor in view, reserving 2 lines for top/bottom
scroll indicators.

### Selection Highlights

- Selected items: `ui.Theme.Secondary.Render("✓ " + name)`
- Toggled items use secondary color for immediate visual feedback

## Data Flow

1. On app startup, `app/events.go` loads decks, tags, and deck→tag mapping
2. The data lands in the global `ui.AppState` snapshot; the root pushes it to
   `DeckSelectModel` via `SyncState(AppState)` (on entry and on state change),
   and the screen rebuilds its two columns
3. User interacts (toggle, search, switch columns)
4. On Enter → emits `ui.DeckSelectionChangedMsg{Selected: deckIDs, SelectedTags: tags}`
5. `app/events.go` → if quiz is mid-progress with recorded answers, shows confirmation
   dialog; otherwise applies immediately via `handleDeckSelectionWithTags()`
6. After applying, navigates back to the previous screen (via `popToPrevious`),
   or to Home if no navigation history

## Registration

Registered in `app/app.go` as:
```go
reg.Register(ui.DecksScreen, screens.NewDeckSelect())
```

Replaces the old `screens.NewDecks()` at the same slot (`ui.DecksScreen`).

## Navigation Integration

- **From Home**: "deck selection" menu item emits `NavigateToMsg{Screen: ui.DecksScreen}`
- **Global shortcut**: `ctrl+f` navigates via `pushTo` (preserves navigation stack)
- **Return**: After confirming selection, user returns to the screen they came from
- **Mid-quiz confirmation**: If the quiz is in progress (answered cards not all done),
  a confirmation dialog ("Already answered cards will be recorded. Are you sure?")
  must be accepted before the selection takes effect
- **Completed quiz**: No confirmation — changing decks after the quiz is done is free

## State Persistence

- `storage.State` now includes `SelectedTags []string` alongside `SelectedDecks`
- Both are restored on startup via `StateStore.Load()`
- Tags saved via `SaveStateCmd` which accepts variadic `selectedTags`

## Keymap

New keymap group `DeckSelect` in `keymap.go`:

```go
type DeckSelect struct {
    List                   // Up, Down, Select (embedded)
    Toggle                 Binding  // space
    ToggleAll              Binding  // a
    SearchToggle           Binding  // s
    NextColumn             Binding  // tab
    PrevColumn             Binding  // shift+tab
}
```

Global keymap includes `DeckSelect Binding` (default: `ctrl+f`).

## Files Changed

| File | Change |
|------|--------|
| `internal/ui/screens/deck_select.go` | NEW — unified screen |
| `internal/ui/screens/decks.go` | DELETED |
| `internal/ui/app/overlay_decks.go` | DELETED |
| `internal/ui/app/model.go` | Removed `DeckOverlay`, added tag/deck-tag fields, `ConfirmOverlay` |
| `internal/ui/app/events.go` | Removed overlay handling, wired tag loading, added `handleDeckSelectionWithTags` |
| `internal/ui/app/commands.go` | Added `ListAllTagsCmd`, `LoadAllDeckTagsCmd`, updated `SaveStateCmd` |
| `internal/ui/app/app.go` | Registered `DeckSelect`, wired `TagProvider` |
| `internal/ui/app/dependencies.go` | Added `TagProvider` interface |
| `internal/ui/app/view.go` | Replaced `DeckSelectionOverlay` with `ConfirmOverlay` |
| `internal/ui/screens/home.go` | Changed to emit `NavigateToMsg{DecksScreen}` instead of `ShowDeckSelectionMsg` |
| `internal/ui/keymap/keymap.go` | Added `DeckSelect` group, global `DeckSelect` binding, config overrides |
| `internal/ui/keymap/keymap_test.go` | Updated `Global` struct for test |
| `internal/ui/screen.go` | Updated `DeckSelectionChangedMsg` with `SelectedTags`, removed `NoScreen`/`ShowDeckSelectionMsg` |
| `internal/storage/state.go` | Added `SelectedTags` field |
