# CLI Package Context

## Kong Wiring

The CLI uses [Kong v1.16](https://github.com/alecthomas/kong) for command-line parsing. Declarative struct tags define flags, arguments, and help text. Every command implements `Run(a *app.App) error`.

### Root `CLI` struct (`root.go`)

```go
type CLI struct {
    Debug      bool                      `help:"Enable debug output."`
    Quiz       QuizCmd                   `cmd:"" help:"Start a quiz."`
    Sync       SyncCmd                   `cmd:"" help:"Synchronize decks."`
    Stats      StatsCmd                  `cmd:"" help:"Show learning statistics."`
    Search     SearchCmd                 `cmd:"" help:"Search vocabulary."`
    Completion kongcompletion.Completion `cmd:"" help:"Install shell completion."`
}
```

Field name = command name unless overridden. CLI commands are registered as fields; the base `Run(a *app.App)` starts the TUI when no subcommand is given.

### Registration in `main.go`

Kong is initialised in `cmd/crds/main.go`:

```go
var c cli.CLI
parser, err := kong.New(
    &c,
    kong.Name("crds"),
    kong.Description("Terminal flashcard application."),
    kong.Bind(&app.App{}),
)
```

Kong binds `*app.App` into the context so every `Run(a *app.App)` receives it.

### How parsing dispatches

```
crds                  → CLI.Run(a)            → TUI
crds quiz --deck foo  → CLI.Quiz.Run(a)       → quiz subcommand
crds sync             → CLI.Sync.Run(a)       → sync subcommand
```

---

## Current state of CLI commands

| Command | File | Struct | Implementation |
|---------|------|--------|----------------|
| `quiz` | `quiz.go` | `QuizCmd` | Stub — prints args |
| `sync` | `sync.go` | `SyncCmd` | Stub — prints args |
| `stats` | `stats.go` | `StatsCmd` | Stub — no Run method |
| `search` | `search.go` | `SearchCmd` | Stub — no Run method |

All `Run()` methods need to be implemented. Some have no `Run()` yet (Kong will call it via the method set; missing Run panics).

---

## The Store access problem

Currently Kong only binds `*app.App` which is an empty struct:

```go
type App struct {
    // empty
}
```

CLI commands **cannot access** the SQLite `*storage.Store` because it's created inside `CLI.Run()` (for the TUI path), not before command dispatch.

**To wire CLI commands, add Store fields to `app.App`** and initialise them in `main.go`:

```go
type App struct {
    Store     *storage.Store
    State     *storage.StateStore
    SharedDir string
    DataDir   string
}
```

Then in `main.go`:

```go
sharedDir := filepath.Join(home, ".local", "share", "crds")
dataDir := filepath.Join(sharedDir, "decks")
os.MkdirAll(dataDir, 0755)

sqliteStore, _ := storage.NewStore(filepath.Join(sharedDir, "crds.db"))
stateStore := storage.NewStateStore(sharedDir)

a := &app.App{
    Store:     sqliteStore,
    State:     stateStore,
    SharedDir: sharedDir,
    DataDir:   dataDir,
}

parser, err := kong.New(&c, kong.Name("crds"), kong.Bind(a))
```

Now every command's `Run(a *app.App)` has direct access to `a.Store`, `a.State`, `a.SharedDir`, and `a.DataDir`.

The TUI path (`CLI.Run`) can reuse the same pre-wired App instead of creating its own store.

---

## Paths

Constructed from `os.UserHomeDir()` in `app.App`:

| Variable | Path |
|----------|------|
| `SharedDir` | `~/.local/share/crds/` |
| `DataDir` | `~/.local/share/crds/decks/` |
| `DBPath` | `~/.local/share/crds/crds.db` |
| `ReserveDir` | `~/.local/share/crds/reserve-copies/` |
| `StatePath` | `~/.local/share/crds/state.yaml` |

---

## Available Store methods for CLI wiring

All methods on `*storage.Store`. All paths use `deckDir` (the directory containing `.yaml` files) and/or `sharedDir` (the parent `~/.local/share/crds/`).

### Deck-level operations

| Method | Parameters | Behaviour |
|--------|------------|-----------|
| `ImportDeck` | `srcPath, deckDir` | Parse YAML, copy to deckDir, sync to DB |
| `ExportDeck` | `deckID, dstPath, deckDir` | Copy source YAML file to dstPath (preserves comments) |
| `ExportDeckFromCache` | `deckID, dstPath` | Reconstruct from SQLite as canonical YAML (no comments) |
| `RenameDeck` | `deckID, newName, deckDir` | Rename in YAML + DB |
| `ChangeDeckID` | `deckID, newID, deckDir` | Change ID in YAML, DB, cascade to entries/progress/reviews/sync_state |
| `DeleteDeck` | `deckID, deckDir` | Remove from DB (cascade), YAML, progress, reviews |
| `ListDecks` | `() ([]string, error)` | List deck IDs from SQLite |
| `SyncDecks` | `deckDir` | Re-sync all YAML files to SQLite cache |

### Entry-level operations

| Method | Parameters | Behaviour |
|--------|------------|-----------|
| `AddEntry` | `deckID, entry, deckDir` | Append entry to YAML + sync |
| `UpdateEntry` | `deckID, entryID, entry, deckDir` | Replace fields in-place (same ID) |
| `ReplaceEntryID` | `deckID, oldID, newID, deckDir` | Change ID, migrate progress/reviews |
| `RemoveEntry` | `deckID, entryID, deckDir` | Delete from YAML + DB, clean progress/reviews |

### Reserve operations

| Method | Parameters | Behaviour |
|--------|------------|-----------|
| `CreateReserve` | `sharedDir` | Backup DB + state.yaml + decks/ to `reserve-copies/` |
| `RevertReserve` | `sharedDir, reservePath` | Restore from backup, auto pre-backup, close/reopen DB |

### Session and stats

| Method | Parameters | Behaviour |
|--------|------------|-----------|
| `EnsureSession` | `() (int64, error)` | Create or return current session |
| `ResetSession` | `() error` | Close current session |
| `RecordAnswer` | `deckID, cardID, grade, reverse` | Log a review |
| `RecordAnswerFull` | `sessionID, deckID, entryID, grade, reverse, userInput, correctAnswer, similarity` | Log with typing detail |
| `Stats` | `() ui.Stats` | Today's aggregate |
| `GetReviewsByEntry` | `entryID, limit` | Last N reviews |
| `GetWeakTypingEntries` | `deckID, limit` | Weakest typed answers |

---

## Editor integration

Package `internal/editor/` provides:

```go
func Edit(content string) (string, error)
```

Opens `$EDITOR` (fallback: nano, vim, vi) with a temp file. CLI commands should:

1. Prepare YAML content (marshalled entry or template)
2. Call `editor.Edit(content)`
3. Parse the returned YAML into the target struct
4. Call the appropriate `Store` method

Helper for entries:

```go
entry, _ := editor.EditEntry(&existingEntry)  // marshal → edit → unmarshal
template := editor.EntryTemplate()             // blank YAML buffer
```

---

## Completing a deck name in shell

A `deckPredictor` is registered in `main.go` for the `"deck"` completion tag:

```go
type deckPredictor struct{ store *storage.DeckStore }
```

Kong completion uses `completion-predictor:"deck"` on arg/flag tags:

```go
Deck string `arg:"" completion-predictor:"deck"`
```

The predictor currently uses the legacy `DeckStore` (filesystem only). When the SQLite `Store` is wired via `App`, this should switch to `a.Store.ListDecks()`.

---

## Adding a new command — step by step

1. **Create the file** `internal/cli/<name>.go` with a struct and `Run` method.
2. **Register** the struct as a field on `CLI` in `root.go` with `cmd:"" help:"..."` tag.
3. **Wire App** in `main.go` so `a.Store`, `a.SharedDir`, `a.DataDir` are available.
4. **Implement Run**: parse args, call store methods, print or edit.
5. **Add `completion-predictor:"deck"`** for any deck-name argument.
6. **If using the editor**: import `crds/internal/editor` and call `editor.Edit` or `editor.EditEntry`.
7. **Tests**: unit-test the Run logic if it contains non-trivial branching; otherwise test the store methods directly.

### Example command template

```go
package cli

import (
    "crds/internal/app"
    "crds/internal/editor"
    "crds/internal/model"
)

type ImportCmd struct {
    Src string `arg:"" required:"" help:"Path to the YAML file to import."`
}

func (c *ImportCmd) Run(a *app.App) error {
    return a.Store.ImportDeck(c.Src, a.DataDir)
}
```

---

## Flag/arg tag patterns used in the codebase

| Tag | Example | Meaning |
|-----|---------|---------|
| `arg:""` | `Deck string \`arg:""\`` | Positional argument |
| `optional:""` | `Deck string \`arg:"" optional:""\`` | Optional positional |
| `required:""` | `Query string \`arg:"" required:""\`` | Required positional |
| `help:"..."` | `Deck string \`help:"Deck name."\`` | Help text |
| `short:"n"` | `Limit int \`short:"n"\`` | Short flag (`-n`) |
| `default:"20"` | `Limit int \`default:"20"\`` | Default value |
| `completion-predictor:"deck"` | `Deck string \`completion-predictor:"deck"\`` | Shell completion |

---

## Known limitations

- `app.App` is empty — must be extended with Store/State/SharedDir/DataDir fields before CLI commands can work
- `CLI.Run()` (TUI path) currently creates its own Store — should reuse the one from App
- `DeckStore` (legacy, filesystem-only) is used for completion prediction; should switch to SQLite `Store.ListDecks()`
- `goose.SetNopLogger()` is called in `CLI.Run()` only, not for CLI command paths — need to suppress goose output in CLI commands too
