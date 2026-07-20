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

- `duplicate_terms` test expects an error but `validate.go` only checks duplicate IDs, not duplicate terms → test fails
- `scheduler/`, `search/`, `quiz/` implementations don't exist yet — those are aspirational (docs/ outruns code)
- `migrations/20260716121051_init.sql` is a goose placeholder (no real schema)

## Architecture (current vs docs)

Docs describe an aspirational layered architecture. Reality is partial:

- **model/** — domain types (Deck, Entry, Progress, Review, Session)
- **parser/** — YAML parsing + validation + normalization (has tests)
- **cli/** — Kong command stubs (quiz, sync, stats, search) — most `Run()` methods only print
- **app/** — empty composition root struct (but `internal/ui/app/` has real UI scaffolding)
- **ui/** — Bubble Tea scaffolding: navigation/ fully implemented, app/ wired, events/ with 4 centralized event types, most other subdirectories still empty

Most work ahead: wiring CLI commands → parser → storage → quiz/scheduler logic.

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
