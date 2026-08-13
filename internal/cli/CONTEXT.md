# CLI Package Context

> Per-package context: how this package works today. Status and plans live in
> `docs/status.md` and `docs/roadmap.md` (see `docs/README.md`).

## Kong Wiring

The CLI uses [Kong v1.16](https://github.com/alecthomas/kong) for command-line parsing. Declarative struct tags define flags, arguments, and help text. Every command implements `Run(a *app.App) error`. The root struct also receives `*kong.Context` to detect subcommand dispatch.

### Root `CLI` struct (`root.go`)

```go
type CLI struct {
    Debug      bool
    Quiz       QuizCmd
    Stats      StatsCmd
    Deck       DeckCmd      `cmd:"" help:"Deck operations."`
    State      StateCmd     `cmd:"" help:"State management."`
    Profile    ProfileCmd   `cmd:"" help:"Profile operations."`
    Completion kongcompletion.Completion
}
```

Flat commands (Quiz, Stats, Completion) remain at root. Deck, state, theme, and profile operations are grouped under `DeckCmd`, `StateCmd`, `ThemeCmd`, and `ProfileCmd`.

### Command groups (`deck.go`, `state.go`, `term.go`, `profile.go`)

```go
type DeckCmd struct {
    Create CreateCmd   `cmd:"" help:"Create a new empty deck."`
    List   ListCmd     `cmd:"" help:"List all decks with entry counts."`
    Import ImportCmd   `cmd:"" help:"Import a deck from a YAML file."`
    Export ExportCmd   `cmd:"" help:"Export a deck to a YAML file."`
    Delete DeleteCmd   `cmd:"" help:"Delete a deck."`
    Search SearchCmd   `cmd:"" help:"Search vocabulary across decks."`
    Edit   EditDeckCmd `cmd:"" help:"Edit a deck's full YAML file."`
    Term   TermCmd     `cmd:"" help:"Manage individual terms in a deck."`
    Tag    TagCmd      `cmd:"" help:"Manage tags on terms."`
}

type TermCmd struct {
    Add  TermAddCmd  `cmd:"" help:"Add a new term."`
    Rm   TermRmCmd   `cmd:"" help:"Remove a term."`
    Edit TermEditCmd `cmd:"" help:"Edit a term."`
}

type TagCmd struct {
    Add  TagAddCmd  `cmd:"" help:"Add tags to a term."`
    Rm   TagRmCmd   `cmd:"" help:"Remove tags from a term."`
    List TagListCmd `cmd:"" help:"List tags on a term."`
}

type StateCmd struct {
    Reserve ReserveCmd `cmd:"" help:"Create a backup/reserve copy."`
    Revert  RevertCmd  `cmd:"" help:"Revert from a reserve copy."`
    Sync    SyncCmd    `cmd:"" help:"Synchronize decks."`
}

type ProfileCmd struct {
    Export ProfileExportCmd `cmd:"" help:"Export full profile for device migration."`
    Import ProfileImportCmd `cmd:"" help:"Import a profile from another device."`
}

type ThemeCmd struct {
    Add    ThemeAddCmd    `cmd:"" help:"Create a new theme."`
    Delete ThemeDeleteCmd `cmd:"" help:"Delete a theme."`
    Edit   ThemeEditCmd   `cmd:"" help:"Edit a theme by opening its YAML file."`
    List   ThemeListCmd   `cmd:"" help:"List all user themes."`
}

```

Each subcommand accepts its own positional args (e.g. `ExportCmd.Deck`, `TermEditCmd.Deck` / `TermEditCmd.TermID`).

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
crds                              → CLI.Run(a, ctx)           → TUI
crds quiz --deck foo              → CLI.Quiz.Run(a)           → TUI with pre-selected deck
crds state sync                   → StateCmd.Sync.Run(a)      → sync
crds deck list                    → DeckCmd.List.Run(a)       → list decks
crds deck create <name> -F <from> -T <to> [--edit] → DeckCmd.Create.Run(a) → create empty deck
crds deck import <file> --replace → DeckCmd.Import.Run(a)     → import (with replace)
crds deck export <deck>           → DeckCmd.Export.Run(a)     → export
crds deck export --all            → DeckCmd.Export.Run(a)     → export all
crds deck edit <deck>             → DeckCmd.Edit.Run(a)       → full deck edit
crds deck search <query> --tags --color auto → DeckCmd.Search.Run(a) → search with filters
crds deck term add <deck> -t ...  → DeckCmd.Term.Add.Run(a)  → add entry inline
crds deck term edit <deck> <id>   → DeckCmd.Term.Edit.Run(a) → edit entry
crds deck term rm <deck> <id> -f  → DeckCmd.Term.Rm.Run(a)   → remove entry (force)
crds deck tag add <deck> <id> ... → DeckCmd.Tag.Add.Run(a)   → add tags
crds deck tag rm <deck> <id> ...  → DeckCmd.Tag.Rm.Run(a)    → remove tags
crds deck tag list <deck> <id>    → DeckCmd.Tag.List.Run(a)  → list tags
crds stats --deck <deck>          → CLI.Stats.Run(a)          → per-deck stats
crds profile export               → ProfileCmd.Export.Run(a)  → profile export
crds profile import <file>        → ProfileCmd.Import.Run(a)  → profile import
crds theme add <name> -p dark     → ThemeCmd.Add.Run(a)       → create theme from preset
crds theme delete <name> -f       → ThemeCmd.Delete.Run(a)    → delete theme
crds theme edit <name>            → ThemeCmd.Edit.Run(a)      → edit theme (seeds built-in copy)
crds theme list                   → ThemeCmd.List.Run(a)      → list themes
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
| `ConfigDir` | `~/.config/crds/` |
| `ThemesDir` | `~/.config/crds/themes/` |

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
| `LoadDeck` | `id (ui.DeckData, error)` | Full deck data from cache for search/display |

### Entry-level operations

| Method | Parameters | Behaviour |
|--------|------------|-----------|
| `AddEntry` | `deckID, entry, deckDir` | Append entry to YAML + sync |
| `UpdateEntry` | `deckID, entryID, entry, deckDir` | Replace fields in-place (same ID) |
| `RemoveEntry` | `deckID, entryID, deckDir` | Delete from YAML, sync, clean progress + reviews |
| `ReplaceEntryID` | `deckID, oldID, newID, deckDir` | Migrate ID in YAML + progress + reviews |

### Reserve operations

| Method | Parameters | Behaviour |
|--------|------------|-----------|
| `CreateReserve` | `sharedDir` | Backup DB + state.yaml + decks/ to `reserve-copies/` (default auto-name) |
| `CreateReserveTo` | `sharedDir, outputDir, name` | Same but with custom output dir and/or name; returns full path |
| `RevertReserve` | `sharedDir, reservePath` | Restore from backup, auto pre-backup, close/reopen DB |
| `ListReserves` | `sharedDir` (standalone) | Returns reserve archive paths, newest-first |

### Profile operations

| Method | Parameters | Behaviour |
|--------|------------|-----------|
| `CreateProfile` | `sharedDir, configDir, outputDir, name` | Pack DB + state + decks + config/keymaps/themes into `crds-profile.tar.gz` (auto-increment on collision) |
| `ImportProfile` | `sharedDir, configDir, profilePath` | Pre-backup, close DB, extract (sharedDir files → sharedDir, config/ subtree → configDir), reopen DB, migrate, sync |

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

### `crds deck edit <deck>` — full deck edit flow

Opens a **copy** of the full `deck.yaml` in `$EDITOR`. After the editor exits, the flow is:

1. Try `parser.Parse()` on the edited YAML
   - **Parse/validate fails** → prompt: `[d]iscard, [c]ontinue editing, [s]ave anyway`
     - discard → no changes
     - continue → re-open editor
     - save anyway → writes raw bytes to disk (sync will skip broken file)
2. **Parse succeeds** → diff against original entries:
   a. **Same term, changed ID** → for each: prompt `[m]igrate stats` / `[c]reate new entry`
      - migrate → calls `Store.ReplaceEntryID(deck, oldID, newID, deckDir)`
      - create new → leaves as-is (old entry falls out, stats orphaned)
   b. **Deleted entries** → after ID migrations resolved: prompt `[c]lear all cache` / `[r]evert all` / `[r]eview each`
      - clear all → `Store.RemoveEntry()` for each deleted ID
      - revert all → re-adds deleted entries from original
      - review each → per entry: clear / revert / skip
3. Write final deck YAML to disk → sync

---

## Shell completion predictors

Three predictors are registered in `main.go`:

| Predictor | Type | Behaviour |
|-----------|------|-----------|
| `"deck"` | `*deckPredictor` | Lists deck IDs from SQLite `Store.ListDecks()` |
| `"reserve"` | `*reservePredictor` | Lists `.tar.gz` files from default `reserve-copies/` directory |
| `"term"` | `*entryPredictor` | Lists entry IDs for the deck typed before the cursor (from `Store.LoadDeck()`) |
| `"theme"` | `*themePredictor` | Lists theme names from `config.DiscoverThemeFiles()` |
| `"preset"` | `*presetPredictor` | Lists built-in preset names via `theme.BuiltinNames()` (used by `theme add -p`) |

Used via `completion-predictor:"deck"` / `completion-predictor:"reserve"` / `completion-predictor:"term"` on struct field tags.

---

## Adding a new command — step by step

1. **Create the file** `internal/cli/<name>.go` with a struct and `Run` method.
2. **Register** the struct as a field on the appropriate group (`DeckCmd`, `StateCmd`, `ThemeCmd`, `TermCmd`, or root `CLI`) in `*.go` with `cmd:"" help:"..."` tag.
3. **Implement Run**: parse args, call store methods, print or edit.
4. **Add `completion-predictor:"deck"`** for any deck-name argument, `completion-predictor:"term"` for entry IDs.
5. **If using the editor**: import `crds/internal/editor` and call `editor.Edit` or `editor.EditEntry`.
 6. **Tests**: unit-test the Run logic if it contains non-trivial branching; otherwise test the store methods directly.

## Interactive prompts (`prompt.go`, `deck_resolve.go`)

The AI deck flow adds interactive line input:

- `promptReadLine(prompt, completions)` is the shared seam (replaceable in
  tests). Real behaviour: `ergochat/readline` with tab-completion over
  `completions` when stdin is a TTY, else a plain buffered scan. Non-TTY reads
  share one `bufio.Reader` keyed to the current `os.Stdin`, so consecutive
  prompts (and `reviewAndAppend`) never lose input to buffering.
- `promptYesNo(prompt)` loops on `y/n` (empty = yes).
- `resolveDeck(a, client, raw, msg, from, to)` resolves an omitted deck for
  `ai add`/`ai fill`: asks the `ai.SuggestDeck` agent, confirms a match, and on
  decline offers create-with-proposed-name / manual name / pick-existing
  (`selectDeckID`, accepts number, id, or name; tab completes ids). Returns
  `errAborted` when the user cancels — callers return nil for that sentinel.

### Example command template

```go
package cli

import (
    "crds/internal/app"
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
| `completion-predictor:"deck"` | `Deck string \`completion-predictor:"deck"\`` | Deck ID completion |
| `completion-predictor:"term"` | `TermID string \`completion-predictor:"term"\`` | Entry ID completion |
