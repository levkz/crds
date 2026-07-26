# CLI Package Context

## Kong Wiring

The CLI uses [Kong v1.16](https://github.com/alecthomas/kong) for command-line parsing. Declarative struct tags define flags, arguments, and help text. Every command implements `Run(a *app.App) error`. The root struct also receives `*kong.Context` to detect subcommand dispatch.

### Root `CLI` struct (`root.go`)

```go
type CLI struct {
    Debug      bool
    Quiz       QuizCmd
    Sync       SyncCmd
    Stats      StatsCmd
    Search     SearchCmd
    Import     ImportCmd
    Export     ExportCmd
    Delete     DeleteCmd
    Reserve    ReserveCmd
    Revert     RevertCmd
    Edit       EditCmd
    Completion kongcompletion.Completion
}
```

Field name = command name. All subcommands have `cmd:""` tags. `CLI.Run(a, ctx)` starts the TUI when no subcommand is given, using `ctx.Selected()` to skip TUI launch after a subcommand has already run.

### Registration in `main.go`

Kong is initialised in `cmd/crds/main.go`. The `*app.App` is pre-wired with Store/State/SharedDir/DataDir before dispatch:

```go
a := &app.App{
    Store:     sqliteStore,
    State:     stateStore,
    SharedDir: sharedDir,
    DataDir:   dataDir,
}
parser, err := kong.New(&c, kong.Name("crds"), kong.Bind(a))
```

### How parsing dispatches

```
crds                        → CLI.Run(a, ctx)       → TUI
crds quiz --deck foo        → CLI.Quiz.Run(a)        → TUI with pre-selected deck
crds sync                   → CLI.Sync.Run(a)        → sync subcommand
crds export <deck>          → CLI.Export.Run(a)      → export subcommand
```

When a subcommand is matched, Kong's `RunNode` walks from the selected node up to the root calling every `Run()` it finds. Subcommand `Run()` executes first, then `CLI.Run()` receives `ctx.Selected() != nil` and returns immediately without launching the TUI.

---

## Paths

Constructed from `os.UserHomeDir()` in `main.go`:

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
| `DeleteDeck` | `deckID, deckDir` | Remove from DB (cascade), YAML, progress, reviews |
| `ListDecks` | `() ([]string, error)` | List deck IDs from SQLite |
| `SyncDecks` | `deckDir` | Re-sync all YAML files to SQLite cache |

### Entry-level operations

| Method | Parameters | Behaviour |
|--------|------------|-----------|
| `AddEntry` | `deckID, entry, deckDir` | Append entry to YAML + sync |
| `UpdateEntry` | `deckID, entryID, entry, deckDir` | Replace fields in-place (same ID) |

### Reserve operations

| Method | Parameters | Behaviour |
|--------|------------|-----------|
| `CreateReserve` | `sharedDir` | Backup DB + state.yaml + decks/ to `reserve-copies/` (default auto-name) |
| `CreateReserveTo` | `sharedDir, outputDir, name` | Same but with custom output dir and/or name; returns full path |
| `RevertReserve` | `sharedDir, reservePath` | Restore from backup, auto pre-backup, close/reopen DB |
| `ListReserves` | `sharedDir` (standalone) | Returns reserve archive paths, newest-first |

### Session and stats

| Method | Parameters | Behaviour |
|--------|------------|-----------|
| `EnsureSession` | `() (int64, error)` | Create or return current session |
| `Stats` | `() ui.Stats` | Today's aggregate |

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

Helpers for entries:

```go
entry, _ := editor.EditEntry(&existingEntry)  // marshal → edit → unmarshal
template := editor.EntryTemplate()             // blank YAML buffer
```

---

## Shell completion predictors

Two predictors are registered in `main.go`:

| Predictor | Type | Behaviour |
|-----------|------|-----------|
| `"deck"` | `*deckPredictor` | Lists deck IDs from SQLite `Store.ListDecks()` |
| `"reserve"` | `*reservePredictor` | Lists `.tar.gz` files from default `reserve-copies/` directory |

Used via `completion-predictor:"deck"` / `completion-predictor:"reserve"` on struct field tags.

---

## Adding a new command — step by step

1. **Create the file** `internal/cli/<name>.go` with a struct and `Run` method.
2. **Register** the struct as a field on `CLI` in `root.go` with `cmd:"" help:"..."` tag.
3. **Implement Run**: parse args, call store methods, print or edit.
4. **Add `completion-predictor:"deck"`** for any deck-name argument.
5. **If using the editor**: import `crds/internal/editor` and call `editor.Edit` or `editor.EditEntry`.
6. **Tests**: unit-test the Run logic if it contains non-trivial branching; otherwise test the store methods directly.

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

- `QuizCmd` has `--limit` and `--reverse` flags acknowledged with stderr warnings but not wired to the TUI
- Grade scale mismatch: Flashcard uses 0-3, Typing uses 1-3 (needs normalization)
- `scheduler/`, `search/`, `quiz/` implementations don't exist yet in the storage layer — only the UI and CLI wiring are done
