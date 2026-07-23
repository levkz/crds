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
- Grade scale mismatch: Flashcard uses 0-3, Typing uses 1-3 (needs normalization)

## Architecture (current vs docs)

Docs describe an aspirational layered architecture. Reality is partial:

- **model/** — domain types (Deck, Entry, Progress, Review, Session). `TypingDetail` only exists in sqlc-generated code.
- **parser/** — YAML parsing + validation + normalization (has tests)
- **cli/** — Kong command stubs (quiz, sync, stats, search) — most `Run()` methods only print
- **app/** — empty composition root struct (but `internal/ui/app/` has real UI scaffolding)
- **ui/** — Bubble Tea scaffolding: navigation/ fully implemented, app/ wired, events/ with 4 centralized event types, most other subdirectories still empty
- **storage/** — SQLite fully implemented via `Store` (goose + sqlc). On startup, `SyncDecks()` caches YAML decks in SQLite. `Store` implements `DeckProvider`, `ProgressRecorder`, and `StatsProvider`. `DeckStore` (legacy filesystem) remains but is not wired.

SQL stack: SQLite (`modernc.org/sqlite`) + goose (migrations) + sqlc (type-safe queries)

Most work ahead: wiring CLI commands → parser → storage → quiz/scheduler logic.

## Commit workflow

After each feature is implemented and tests pass (`make test` + `make build`), commit
using `commit.sh` (run `./commit.sh --execute`). It groups changed files by feature
into 9 logical commits. If something needs fixing, reset the last commit with
`git reset --soft HEAD~1`, fix, test, and re-run `./commit.sh --execute` (the script
stages and commits incrementally — already-committed groups are skipped).

The shared boilerplate lives in `scripts/commit_group.sh` — source it in new commit
scripts to get the `commit()` helper and dry-run handling. Feed `COMMIT_EXECUTE=1`
as the environment variable (or set it via a `--execute` flag in your wrapper).

## Parser specifics

- Uses `go.yaml.in/yaml/v3` (NOT the standard `gopkg.in/yaml.v3`)
- `Normalize()` trims whitespace on all fields in-place
- `Validate()` checks: deck id/name/language required, entry id required, no duplicate IDs, term required, ≥1 translation
- Test fixtures in `internal/parser/testdata/` (12 YAML files)
- `testdata/auto_ids.yaml` tests entries without explicit IDs (not yet wired to validation)

## Style

- Prefer table-driven tests with `testdata/` fixtures
- One responsibility per package; no circular deps
- Explicit over implicit; small functions; early returns
- Avoid unnecessary interfaces and global state

## Other

- Vocabulary files in `exercises/*.txt` (legacy `=>` format) and in YAML (new parser format)
- `.opencodeignore` excludes `.git`, `go.sum`, `exercises/`
