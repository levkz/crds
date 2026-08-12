# CRDS — Terminal Flashcard Application

Keyboard-first vocabulary learning with flashcards and typing quizzes, entirely in your terminal.

Vocabulary belongs in YAML for easy manipulation. User progress belongs in SQLite.

## Features

- **Two quiz modes** — Flashcard grading (Again/Hard/Good/Easy) and Typing with fuzzy-matched auto-grading via Levenshtein distance
- **Accent input mappings** — Babbel-style triggers (`e/` → `é`) from per-language config files and per-deck fields, toggleable in the typing quiz; optional accent-insensitive matching
- **Full TUI** — 9 screens with stack-based navigation, theme system, and terminal-wide background fill
- **YAML decks** — Human-editable, version-controllable vocabulary files with auto-ID generation and variant expansion syntax
- **SQLite persistence** — Reviews, sessions, progress, typing details (via `modernc.org/sqlite` + goose + sqlc)
- **5 built-in themes** — default, dark, light, tokyonight, mocha; plus custom YAML themes via 18-field color palette with semantic overrides
- **CLI + TUI** — Manage decks, terms, tags, and backups from the command line or launch the interactive UI
- **Shell completion** — Deck names, entry IDs, and reserve file paths

## Installation

### Prerequisites

Go 1.25+

### Commands

| Action        | Command            |
|---------------|--------------------|
| Build         | `make build`       |
| Install       | `make install`     |
| Run           | `make run`         |
| Test all      | `make test`        |
| Test single   | `go test ./internal/parser/` |
| Lint          | `make lint`        |
| Tidy          | `make tidy`        |
| Legacy build  | `make legacy`      |

## Quick Start

1. Create a deck YAML file (see [Deck Creation Guide](docs/DECK_CREATION_GUIDE.md))
2. Import it: `crds deck import path/to/deck.yaml`
3. Launch the TUI: `crds`
4. Select your deck(s) in the deck selection screen
5. Choose Flashcard or Typing Quiz
6. Study, grade, repeat

## CLI Reference

### `crds` — Launch TUI

Synces all YAML decks from `~/.local/share/crds/decks/` to the SQLite cache, then opens the Bubble Tea terminal UI. This is the default command when no subcommand is given.

---

### Quiz & Statistics

#### `crds quiz [deck]`

Start a flashcard quiz. Optionally pre-select a single deck by name.

| Flag         | Description |
|--------------|-------------|
| `--limit N`  | Maximum number of cards (default 20, not yet wired in TUI) |
| `--reverse`  | Reverse quiz direction (not yet wired in TUI) |

#### `crds stats [--deck <deck>]`

Display today's learning statistics: reviewed count and accuracy percentage. With `--deck`, shows per-deck stats including total entries, reviewed today, and accuracy.

---

### Deck Management

#### `crds deck create <name> -F <from> -T <to> [--edit]`

Create a new empty deck as `<name>.yaml` with the minimum YAML: `id`, `name`, and both language fields, no entries.

| Flag         | Description |
|--------------|-------------|
| `-F <lang>`  | Source language for terms (e.g. `fr`) — required |
| `-T <lang>`  | Target language for translations (e.g. `en`) — required |
| `--edit`     | Open the new deck in `$EDITOR` after creating it (same flow as `crds deck edit <name>`) |

The deck is synced to the SQLite cache immediately, so it appears in `crds deck list` and the TUI. The name argument is used for both the deck `id` and `name:`.

#### `crds deck list`

List all decks with entry counts, language, and translation language. Outputs a formatted table. Shows "No decks found." if the cache is empty.

#### `crds deck import <path> [--replace]`

Import a deck from a YAML file (or directory of `.yaml` files).

| Flag         | Description |
|--------------|-------------|
| `--replace`  | Delete any existing deck with the same name before importing |

When `<path>` is a directory, iterates all `.yaml` files and imports each one, skipping non-YAML files and subdirectories.

#### `crds deck export <deck> [--all] [-o <path>]`

Export a deck to a YAML file (preserves original comments from the source file).

| Flag            | Description |
|-----------------|-------------|
| `--all`         | Export every deck |
| `-o <path>`     | Destination path (or directory with `--all`). Default: `<deck>.yaml` or current directory |

#### `crds deck delete <deck> [-f]`

Delete a deck from both the filesystem YAML and the SQLite cache. Cascades to remove progress, reviews, and session records.

| Flag  | Description |
|-------|-------------|
| `-f`  | Skip confirmation prompt |

Prompts `Delete deck "name"? [y/N]` unless `-f` is passed.

#### `crds deck search <query> [--deck <deck>] [--tags <tag>] [--color <mode>]`

Search for entries across all decks. Results are grouped by deck, sorted by term, and show tags, translations, and notes.

| Flag            | Description |
|-----------------|-------------|
| `--deck <d>`    | Limit search to specific deck(s), repeatable |
| `--tags <t>`    | Filter by tags (AND logic), repeatable |
| `--color <mode>`| Highlight matches: `auto` (default, TTY only), `always`, `never`. Respects `GREP_COLORS` / `GREP_COLOR` env vars |

Output format:
```
=== Deck Name (deck_id) — 2 match(es) ===
  term [tag1,tag2]  → translation1, translation2
         notes: example note
```

#### `crds deck edit <deck>`

Open the deck's full YAML file in `$EDITOR` (fallback: nano → vim → vi). After the editor exits:

1. **Parses and validates** the edited YAML
2. **If parse fails** — prompts:
   - `[d]iscard` — discard changes, exit
   - `[c]ontinue` — re-open editor
   - `[s]ave anyway` — write raw bytes (sync will skip broken file)
3. **If parse succeeds** — detects changes:
   - **Same term, changed ID** — for each: prompt `[m]igrate` stats or `[c]reate` as new entry
     - Migrate: calls `Store.ReplaceEntryID()` to update all progress/review references
     - Create: leaves old entry orphaned
   - **Deleted entries** — prompt `[c]lear all` cache, `[r]evert all` deletions, or `[r]eview each` individually
4. Writes final YAML to disk and applies any pending ID migrations

---

### Term Management

#### `crds deck term add <deck> [-t <term>] [-f <file>] [--translations <csv>] [--examples <csv>] [--tags <tag>...]`

Add a new term to a deck. Three input modes, tried in order:

| Mode      | Trigger      | Behaviour |
|-----------|-------------|-----------|
| File      | `-f <path>` | Parse YAML entry from file (use `-f -` for stdin) |
| Inline    | `-t <term>` | Use `--translations` (comma-separated), `--examples` (comma-separated), and `--tags` (repeatable) |
| Editor    | (no flags)  | Open `$EDITOR` with a blank YAML entry template |

Appends the entry to the deck's YAML file and syncs to SQLite.

#### `crds deck term edit <deck> <id> [-t <term>] [-f <file>]`

Edit an existing term. Three input modes:

| Mode      | Trigger      | Behaviour |
|-----------|-------------|-----------|
| File      | `-f <path>` | Replace full entry from YAML (preserves original ID) |
| Inline    | `-t <term>` | Replace term text only, in-place |
| Editor    | (no flags)  | Open `$EDITOR` with current entry pre-filled as YAML |

#### `crds deck term rm <deck> <id> [-f]`

Remove a term from its deck. Deletes from YAML file, syncs SQLite, and cleans up progress and review history.

| Flag  | Description |
|-------|-------------|
| `-f`  | Skip confirmation prompt |

---

### Tag Management

#### `crds deck tag add <deck> <id> <tag> [<tag>...]`

Add one or more tags to an entry. Tags are positional arguments, space-separated.

#### `crds deck tag rm <deck> <id> <tag> [<tag>...]`

Remove specific tags from an entry.

#### `crds deck tag list <deck> [<id>]`

List all tags on an entry, sorted alphabetically. Omit the entry ID to list all tags used across the whole deck.

---

### State Management (Backups)

#### `crds state reserve [-o <dir>] [-n <name>]`

Create a backup/reserve copy. Archives the SQLite database, `state.yaml`, and all deck YAML files into a `.tar.gz` archive.

| Flag        | Description |
|-------------|-------------|
| `-o <dir>`  | Output directory (default: `~/.local/share/crds/reserve-copies/`) |
| `-n <name>` | Archive name (`.tar.gz` auto-appended). Default: auto-generated timestamp |

Returns the full path to the created archive.

#### `crds state revert --latest | -f <archive>`

Restore from a reserve copy. Automatically creates a pre-revert backup before restoring.

| Flag           | Description |
|----------------|-------------|
| `--latest`     | Use the most recent reserve in the default directory |
| `-f <archive>` | Path to a specific reserve archive (supports tab completion) |

Closes and reopens the database connection during restore, and runs any pending migrations on the restored DB.

#### `crds state sync [-w]`

Re-sync all YAML deck files to the SQLite cache. Compares file mtimes and only processes changed files.

| Flag  | Description |
|-------|-------------|
| `-w`  | Write auto-generated entry IDs back into YAML files |

This is called automatically on TUI startup (`crds` or `crds quiz`), but can be run explicitly when needed.

---

### Profile Migration

#### `crds profile export [-o <dir>] [-n <name>]`

Export your full profile for device migration. Packages the following into a single `crds-profile.tar.gz`:

- SQLite database (`crds.db`)
- `state.yaml`
- All deck YAML files
- Config (`~/.config/crds/config.yaml`)
- Keymaps (`~/.config/crds/keymaps.yaml`)
- Custom themes (`~/.config/crds/themes/`)
- Input mappings (`~/.config/crds/mappings/`)

Auto-increments the archive name on filename collision.

| Flag        | Description |
|-------------|-------------|
| `-o <dir>`  | Output directory (default: current directory) |
| `-n <name>` | Archive name (`.tar.gz` auto-appended). Default: `crds-profile` |

#### `crds profile import <file>`

Import a profile from another device. Creates a pre-import backup of the current state, then:

1. Extracts the archive (shared dir files → `~/.local/share/crds/`, config subtree → `~/.config/crds/`)
2. Closes the current DB connection
3. Reopens the restored DB
4. Runs any pending migrations
5. Syncs decks

---

### AI Agent

#### `crds ai interpret [--deck <deck>] [-t <text>] [-f <file>] [--minimal|--full] [-F <from>] [-T <to>] [--msg <text>] [--dry-run]`

Convert unstructured text (words, phrases, or `term = translation` lines) into
YAML entries. With `--deck`, the deck's language pair is used and sample
entries seed the prompt. Prints the proposed YAML without writing anything.

| Flag | Description |
|------|-------------|
| `--minimal` | Bare `term` + `translations` only (the default). |
| `--full` | Full entries: at least 4 example uses (each a source-language sentence + target-language translation), a `notes` field, and tags. Structural tags (noun, verb, adjective, gender, ...) and a CEFR proficiency tag (A1–C2) are always added; theme tags come from the deck's tag list when `--deck` is given, otherwise a concise theme tag is chosen (e.g. `greetings`). |
| `-F, --translate-from <lang>` | Source language for terms/examples (overrides the deck). |
| `-T, --translate-to <lang>` | Target language for translations (overrides the deck). |
| `--msg <text>` | Pass an extra instruction to the model, e.g. `--msg "use formal register"`. |
| `--dry-run` | Print the prompt instead of calling the API. |

`--minimal` and `--full` are mutually exclusive.

#### `crds ai fill <deck> [-t <text>] [-f <file>] [-F <from>] [-T <to>] [--msg <text>] [--dry-run]`

Complete partial YAML entries (e.g. just a `term` + translations) into full
entries: at least 4 language-appropriate example sentences, a `notes` field,
and tags. Structural tags (noun, verb, adjective, gender, ...) and a CEFR
proficiency tag (A1–C2) are always added; theme tags are chosen only from the
deck's existing tag list. `-F`/`-T` override the deck's language pair; `--msg`
passes an extra instruction to the model. Prints the completed YAML; nothing
is written.

#### `crds ai add <deck> [-t <text>] [-f <file>] [-F <from>] [-T <to>] [--msg <text>]`

Interpret words (or YAML) and fill them out in one step, then let you review
before appending: `[a]ppend`, `[e]dit` (re-open in `$EDITOR`), or `[d]iscard`.
`-F`/`-T` override the deck's language pair; `--msg` passes an extra
instruction to the model in both steps. Appends go through the full
parser/validation chain via the storage `AppendEntries` and are synced with
auto-generated IDs.

**AI configuration.** The default provider is Pollinations.AI (keyless,
model `openai`). Configure via `~/.config/crds/config.yaml`:

```yaml
ai:
  provider: pollinations   # pollinations | ollama | openai | gemini | openrouter | groq | nvidia
  model: openai            # e.g. llama3.2 for ollama
  api_key: ""              # not needed for pollinations/ollama
  base_url: ""             # default per provider
```

Environment variables override the file: `CRDS_AI_PROVIDER`, `CRDS_AI_MODEL`,
`CRDS_AI_API_KEY`, `CRDS_AI_BASE_URL`. See `internal/ai/PLAN.md` for the full
provider preset list.

---

### Shell Completion

#### `crds completion install [--shell bash|zsh|fish]`

Install shell completion. Tab-completion predictors are registered for:

- **Deck names** — from SQLite `Store.ListDecks()`
- **Entry IDs** — per-deck, based on the deck argument before the cursor
- **Reserve paths** — from `~/.local/share/crds/reserve-copies/`

## Deck Format

See the full [Deck Creation Guide](docs/DECK_CREATION_GUIDE.md) for complete documentation on:

- YAML deck structure and required fields
- Auto-generated entry IDs (and how to override them)
- Variant expansion syntax (parens `()` for optional parts, brackets `[]` for required alternatives — ideal for conjugations and articles)
- Validation rules (deck fields, terms, translations, duplicate detection)

Minimal deck example:

```yaml
id: french_a1
name: French A1
language: fr
translation_language: en

entries:
  - id: bonjour
    term: bonjour
    translations:
      - text: hello
```

Full-featured entry:

```yaml
entries:
  - id: bonjour
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
```

Variant syntax example (expands to 5 conjugations, auto-ID picks shortest):

```yaml
  - term: mang[er/ez/e/ons/ent]
    translations:
      - text: eat
```

## Quiz Modes

### Flashcard Quiz

See the term, press a key to reveal the answer. Grade your recall on a 4-point scale:

| Grade  | Key | Meaning       |
|--------|-----|---------------|
| Again  | `a` | Didn't recall |
| Hard   | `h` | Recalled with difficulty |
| Good   | `g` | Recalled correctly |
| Easy   | `e` | Recalled instantly |

Answer side displays tags, usage examples (single or two-column layout with pagination), and notes. Progress is recorded to SQLite on each answer.

### Typing Quiz

Type the translation directly. The answer is auto-graded using Levenshtein distance:

| Score                | Grade    |
|----------------------|----------|
| Exact match          | Good (2) |
| Similarity ≥ 0.7     | Hard (1) |
| Below threshold      | Again (0) |

On reveal, shows the correct answer alongside your input. Tags and examples are displayed. This mode does not use a grade menu — grading is automatic.

### Accent Input Mappings & Matching Mode

Accented characters can be typed with Babbel-style trigger combos, e.g. `e/` → `é`,
`` e` `` → `è`. Mappings are resolved per deck from three layers (later wins):

1. Built-in defaults (French ships with the binary)
2. User files: `~/.config/crds/mappings/<lang>.yaml`
3. Deck field: `input_mappings` (see [Deck Creation Guide](docs/DECK_CREATION_GUIDE.md))

Triggers expand as you type at the cursor position. Press `ctrl+p` to toggle
parsing — with it off you can type a literal `e/`, and turning it back on parses
newly typed text (even mid-string) without re-scanning what's already in the input.
While parsing is on and the answer isn't submitted yet, the active triggers are
shown as a legend (e.g. `e/→é`) above the status bar.

The matching mode in `~/.config/crds/config.yaml` controls how typed answers are graded:

| Mode           | Behavior                                        |
|----------------|-------------------------------------------------|
| `approximate` (default) | Accents ignored — `cafe` matches `café` |
| `strict`       | Accents matter — `cafe` does not match `café`    |

## Configuration

Location: `~/.config/crds/`

| File / Dir                | Purpose |
|---------------------------|---------|
| `config.yaml`             | Theme, animation enabled, default quiz limit, matching mode |
| `keymaps.yaml`            | Keybinding overrides applied via `keymap.ApplyDefaultOverrides()` |
| `themes/*.yaml`           | Custom themes with named palette references or direct ANSI/hex values |
| `mappings/*.yaml`         | Per-language input mappings (accent triggers) |

### Built-in Themes

Default, dark, light, tokyonight (hex values from [folke/tokyonight.nvim](https://github.com/folke/tokyonight.nvim)), mocha. The palette has 18 fields: 15 named colors plus 3 semantic overrides (Primary, Secondary, Accent).

## Data Locations

| What                       | Path |
|----------------------------|------|
| Deck YAML files            | `~/.local/share/crds/decks/` |
| SQLite database            | `~/.local/share/crds/crds.db` |
| Reserve backups            | `~/.local/share/crds/reserve-copies/` |
| Selected decks state       | `~/.local/share/crds/state.yaml` |
| Config directory           | `~/.config/crds/` |
| Custom themes              | `~/.config/crds/themes/` |
| Input mappings             | `~/.config/crds/mappings/` |

## Architecture

**Stack:** Go + [Kong](https://github.com/alecthomas/kong) (CLI) + [Bubble Tea](https://github.com/charmbracelet/bubbletea) (TUI) + `modernc.org/sqlite` (pure Go SQLite, no CGo) + [goose](https://github.com/pressly/goose) (migrations) + [sqlc](https://sqlc.dev/) (type-safe query generation)

**Core principle:** Vocabulary content lives in YAML files (portable, version-controllable, human-editable). User progress lives in SQLite (reviews, sessions, spaced repetition state, typing details). The SQLite cache is rebuilt from YAML on startup via mtime-based incremental sync.

## Development

```bash
make test     # go test ./...
make lint     # golangci-lint run
make tidy     # go mod tidy
make build    # go build -o crds ./cmd/crds/
make legacy   # go build -o crds-legacy ./cmd/legacy-quiz/
make docs-check  # verify documentation links and single-source rules
```

## Documentation

All project documentation is indexed in [`docs/README.md`](docs/README.md),
which defines the documentation taxonomy and where each fact lives. For status
and known issues see [`docs/status.md`](docs/status.md); for planned work see
[`docs/roadmap.md`](docs/roadmap.md).

## License

MIT
