# AGENTS.md

## Entry points

- `cmd/crds/main.go` — main app (Kong CLI + Bubble Tea TUI)
- `cmd/legacy-quiz/main.go` — legacy terminal quiz, reads `exercises/*.txt`

## Module

`crds` (Go 1.25.1). Run everything from repo root.

## Commands

| Action | Command |
|--------|---------|
| Build | `make build` |
| Install | `make install` |
| Run | `make run` |
| All tests | `make test` |
| Single pkg | `go test ./internal/parser/` |
| Lint | `make lint` (requires `golangci-lint`) |
| Tidy | `make tidy` |
| Build legacy quiz | `make legacy` |

## Known issues

- `scheduler/`, `search/`, `quiz/` implementations don't exist yet — those are aspirational (docs/ outruns code)
- CLI commands are all stubs — see `internal/cli/CONTEXT.md` for the wiring guide
- `app.App` is empty — must be extended with Store/State/SharedDir/DataDir before CLI commands can work
- Background fill ANSI nesting: `fillBackground()` must split at every `\033[0m` (full reset) and re-wrap each segment — otherwise inner resets destroy outer backgrounds. This is handled in `app/view.go` but any new ANSI-producing code must account for it.

## Architecture (current vs docs)

Docs describe an aspirational layered architecture. Reality is partial:

- **model/** — domain types (Deck, Entry, Progress, Review, Session). `TypingDetail` only exists in sqlc-generated code.
- **parser/** — YAML parsing + validation + normalization (has tests)
- **cli/** — Kong command stubs — see `internal/cli/CONTEXT.md` for the full wiring guide
- **app/** — empty composition root struct (but `internal/ui/app/` has real UI scaffolding)
- **editor/** — `$EDITOR`/nano/vim invocation + YAML buffer handling for entry editing
- **ui/** — Full Bubble Tea UI: navigation/ fully implemented, app/ wired, 8 screens, theme system with 18-field palette (15 colors + 3 semantic overrides) and 4 built-in themes (dark, light, tokyonight, default), background fill across entire terminal, inline notifications
- **storage/** — SQLite fully implemented via `Store` (goose + sqlc). On startup, `SyncDecks()` caches YAML decks in SQLite. `Store` implements `DeckProvider`, `ProgressRecorder`, and `StatsProvider`. Deck+entry CRUD, reserve/backup, revert all implemented. `DeckStore` (legacy filesystem) remains but is not wired.

SQL stack: SQLite (`modernc.org/sqlite`) + goose (migrations) + sqlc (type-safe queries)

Most work ahead: wiring CLI commands (`internal/cli/CONTEXT.md`) → quiz/scheduler logic.

## Commit workflow

After each feature is implemented and tests pass (`make test` + `make build`), commit
manually with logical grouping:

1. `git add <files>` for each logical change
2. `git commit -m "scope: description"` with a short subject line and a blank line
   followed by a brief body explaining what and why
3. Repeat for each independent concern (feature, refactor, tests, docs)

Group related files together. Separate unrelated changes into distinct commits.
Keep the subject line under 72 characters. Use imperative mood.

If something needs fixing after a commit, use `git reset --soft HEAD~1`, fix, test,
and re-commit.

## Parser specifics

- Uses `go.yaml.in/yaml/v3` (NOT the standard `gopkg.in/yaml/v3`)
- `Normalize()` trims whitespace on all fields in-place
- `assignIDs()` generates IDs for entries missing them: expands the term via `ExpandText`, picks the shortest variant, sanitises
- `Validate()` checks: deck id/name/language required, entry id required (auto-filled), no duplicate IDs, term required, ≥1 translation
- Test fixtures in `internal/parser/testdata/` (13 YAML files)
- `testdata/auto_ids.yaml` tests entries without explicit IDs

## Theme specifics

- `Palette` struct has 15 named colors + 3 semantic overrides (Primary, Secondary, Accent)
- 4 built-in themes: default (ANSI 256), dark, light, tokyonight (hex values from folke/tokyonight.nvim)
- `theme.Store` pre-registers "default", "dark", "light", "tokyonight"
- Config supports custom themes via YAML with named palette references or direct ANSI/hex values
- `fillBackground()` wraps every line with the theme background — handles ANSI reset codes by splitting and re-wrapping segments

## Quiz rendering

- Shared rendering functions in `internal/ui/screens/quiz_shared.go`: `renderQuizBottomSection`, `renderQuizTags`, `renderQuizExamplesBlock`, `renderQuizExamplesSingleCol`, `renderQuizExamplesTwoCol`, `renderQuizExampleCell`, `quizExamplesPerPage`
- Both `QuizModel` and `TypingQuizModel` use these shared functions, computing `topBodyLines` from their own `renderTopBody()`
- `renderTopBody()` is model-specific (flashcard includes grade menu, typing doesn't) and stays in each file
- `renderFooter()` is model-specific (different key references) and stays in each file

## Grade scale

Unified 0-3 scale across both quiz types:
- `GradeAgain=0`, `GradeHard=1`, `GradeGood=2`, `GradeEasy=3`
- Type defined in `internal/ui/grade.go` as `ui.Grade`
- Flashcard uses `ui.Grade*` constants directly
- Typing quiz uses `fuzzy.*` constants internally; converts to `ui.Grade` at the `SaveAnswerMsg` boundary via `ui.Grade(fuzzyGrade)`
- `ProgressStore.Stats()` considers `Grade >= ui.GradeGood` as correct
- `SaveAnswerMsg.Grade` is typed as `ui.Grade`; cast to `int` at the storage boundary in `commands.go`

## Style

- Prefer table-driven tests with `testdata/` fixtures
- One responsibility per package; no circular deps
- Explicit over implicit; small functions; early returns
- Avoid unnecessary interfaces and global state

## Other

- Vocabulary files in `exercises/*.txt` (legacy `=>` format) and in YAML (new parser format)
- `.opencodeignore` excludes `.git`, `go.sum`, `exercises/`
