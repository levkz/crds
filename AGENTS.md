# AGENTS.md

Instructions for AI coding agents and human contributors.

For implementation status, known issues, and the roadmap, see `docs/README.md`
(the documentation map) → `docs/status.md` and `docs/roadmap.md`. This file only
contains workflow, conventions, and project-specific gotchas.

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
| Docs link check | `make docs-check` |

## Commit workflow

After each feature is implemented and tests pass (`make test` + `make build`),
commit manually with logical grouping:

1. `git add <files>` for each logical change
2. `git commit -m "scope: description"` with a short subject line and a blank line
   followed by a brief body explaining what and why
3. Repeat for each independent concern (feature, refactor, tests, docs)

Group related files together. Separate unrelated changes into distinct commits.
Keep the subject line under 72 characters. Use imperative mood.

If something needs fixing after a commit, use `git reset --soft HEAD~1`, fix, test,
and re-commit.

## Documentation conventions

- Update documentation in the **same commit** as the code change it describes.
- Status, known issues, and test counts live **only** in `docs/status.md`.
- Plans live **only** in `docs/roadmap.md`.
- Per-package `CONTEXT.md` files describe the package as it is: no status, no
  plans, no proposals.
- Never restate a fact that has a home elsewhere — link to it. See
  `docs/README.md` for the taxonomy and "where a fact lives" table.

## Parser specifics

- Uses `go.yaml.in/yaml/v3` (NOT the standard `gopkg.in/yaml/v3`)
- `Normalize()` trims whitespace on all fields in-place
- `assignIDs()` generates IDs for entries missing them: expands the term via `ExpandText`, picks the shortest variant, sanitises
- `Validate()` checks: deck id/name/language required, entry id required (auto-filled), no duplicate IDs, term required, ≥1 translation
- Test fixtures in `internal/parser/testdata/` (12 YAML files)
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

## Gotchas

- **Background fill ANSI nesting:** `fillBackground()` must split at every
  `\033[0m` (full reset) and re-wrap each segment — otherwise inner resets
  destroy outer backgrounds. This is handled in `app/view.go` but any new
  ANSI-producing code must account for it.
- **SQLite pragmas:** `PRAGMA foreign_keys = ON` must be set via the DSN
  `_pragma` parameters, not `db.Exec()` — pooled connections ignore
  `db.Exec` pragmas (see `docs/DATAMODEL.md`).

## Style

- Prefer table-driven tests with `testdata/` fixtures
- One responsibility per package; no circular deps
- Explicit over implicit; small functions; early returns
- Avoid unnecessary interfaces and global state
- Never hardcode colors — use semantic theme values
- Keyboard first — every visible action has a shortcut

## Other

- Vocabulary files in `exercises/*.txt` (legacy `=>` format) and in YAML (new parser format)
- `.opencodeignore` excludes `.git`, `go.sum`, `exercises/`
