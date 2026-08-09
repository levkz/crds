# Documentation Map

This is the single entry point for all project documentation. It defines the
documentation taxonomy and indexes every document in the repository.

---

# Taxonomy

Documents are classified by **lifetime and volatility**. A document's type tells
you when it must be updated — and where a fact belongs.

| Type | Purpose | When to update | Volatility |
|---|---|---|---|
| Map | Index + rules for where facts live (this file) | Rarely | Low |
| Reference | How the system *is*: architecture, data model, per-package details | In the same commit as the code it describes | Medium |
| Decision | *Why* things are the way they are: principles, trade-offs, rationale | Rarely | Low |
| Guide | How to *use or do* things | When the UX or file format changes | Medium |
| Status | What is implemented, known issues, test counts | Every feature or fix | High |
| Roadmap | What comes next, priorities | Planning sessions | High |
| Proposal | Feature specs before/during implementation | Until implemented, then archived | High → dead |

**Rules:**

1. **One fact, one home.** Status, schemas, screen lists, and known issues are
   never restated — other documents link to their home.
2. **Status lives only in `docs/status.md`.** No other document carries
   implementation status, test counts, or known issues.
3. **Plans live only in `docs/roadmap.md`.** Reference and decision documents
   contain no "future work" or aspirational content.
4. **Proposals are archived when done.** When a `docs/proposals/*.md` feature is
   implemented, mark it done and move any residual work into `docs/roadmap.md`.
5. **Docs change with the code.** Update documentation in the same commit as the
   code change it describes.
6. **Never document a fact twice.** If you catch yourself copying a section,
   link to it instead.

---

# Where a fact lives

| Fact | Home |
|---|---|
| What's implemented / known issues / test counts | `docs/status.md` |
| What's planned / priorities | `docs/roadmap.md` |
| System layering, dependency rules, package org | `docs/ARCHITECTURE.md` |
| YAML deck format, SQLite schema, grade scale | `docs/DATAMODEL.md` |
| Why things are built this way | `docs/DESIGN.md` |
| How to create/manage decks | `docs/DECK_CREATION_GUIDE.md` |
| User-facing features, CLI reference | `README.md` |
| Agent/contributor workflow, conventions, commands | `AGENTS.md` |
| A single package's internals | that package's `CONTEXT.md` |

---

# Index

## Project-level (`docs/`)

| Document | Type | Purpose |
|---|---|---|
| `docs/README.md` | Map | This file |
| `docs/status.md` | Status | Implementation status, known issues, test baseline |
| `docs/roadmap.md` | Roadmap | Backlog and priorities |
| `docs/ARCHITECTURE.md` | Reference | System architecture, layers, dependency rules |
| `docs/DATAMODEL.md` | Reference | Deck format, SQLite schema, grade scale |
| `docs/DESIGN.md` | Decision | Design decisions and rationale |
| `docs/DECK_CREATION_GUIDE.md` | Guide | Creating and managing vocabulary decks |

## Feature proposals (`docs/proposals/`)

| Document | Type | Purpose | Status |
|---|---|---|---|
| `docs/proposals/tag_architecture.md` | Proposal | Deck↔tag cross-filtering | Storage done; UI wiring deferred |
| `docs/proposals/deck_selection_screen.md` | Proposal | Split-column deck+tag selection | Implemented |
| `docs/proposals/search_screen_revamp.md` | Proposal | Three-column search | Not started |

## Root

| Document | Type | Purpose |
|---|---|---|
| `README.md` | Guide | User-facing overview, quick start, CLI reference |
| `AGENTS.md` | Guide | Instructions for AI coding agents and contributors |

## UI (`internal/ui/docs/`)

| Document | Type | Purpose |
|---|---|---|
| `internal/ui/docs/ARCHITECTURE.md` | Reference | How the UI layer is built (Bubble Tea, navigation, screens, state) |
| `internal/ui/docs/STYLE_GUIDE.md` | Decision | Visual language of the terminal UI |

## Per-package context

Each package keeps a local `CONTEXT.md` describing its internals. It uses a fixed
skeleton: Purpose / Responsibilities / Key files / Dependencies / Integration /
Notes for changes. It contains **no status, test counts, known issues, or plans**
— those live in `docs/status.md` and `docs/roadmap.md`.

| Document | Purpose |
|---|---|
| `internal/cli/CONTEXT.md` | Kong command wiring, store methods, editor flow |
| `internal/ui/components/CONTEXT.md` | Component taxonomy, state model, API conventions |
| `internal/ui/screens/CONTEXT.md` | Screen responsibilities, state sync, messages |
| `internal/ui/keymap/CONTEXT.md` | Keybinding design, footer generation, overrides |
| `internal/ui/layout/CONTEXT.md` | Layout primitives and width propagation |
| `internal/ui/navigation/CONTEXT.md` | Navigation manager, registry, stack |
| `internal/ui/renderer/CONTEXT.md` | Rendering utilities (width, ANSI, wrapping) |
| `internal/ui/styles/CONTEXT.md` | Semantic style definitions |
| `internal/ui/theme/CONTEXT.md` | Design system: palette, typography, icons, theming |
| `internal/ui/theme/DESIGN.md` | Theme design language (colors, icons, spacing, borders) — decision doc |
| `internal/ui/actions/CONTEXT.md` | Why the actions layer is deferred |
