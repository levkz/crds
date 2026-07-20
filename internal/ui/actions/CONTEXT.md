# Actions Context

## Purpose

The `actions` package would define high-level user intentions as typed
constants, decoupling "what key was pressed" from "what the user wants to do."

Proposed flow:

```
Key press → keymap matches → Action constant → screen handles action
```

Compared to the current flow:

```
Key press → keymap matches → screen handles directly
```

## Current status

Not implemented. The directory contains only a TODO.md with candidate
action types.

## Why it does not exist yet

The current architecture handles this concern at two layers:

1. **`keymap/`** — centralizes "what does this key mean?" via `Binding.Match()`.
   Each screen already checks `keymap.DefaultQuiz.Reveal.Match(msg)` which
   encodes the same intent as an `ActionReveal` constant.

2. **`screens/`** — each screen's `Update()` method handles matched keys
   directly, mutating local state and returning commands.

Adding an actions package would insert an intermediate representation
between keymap matching and screen handling. In the current codebase,
this indirection solves no concrete problem — actions are almost always
screen-specific (e.g., "Reveal answer" only matters in Quiz, "Search"
only matters in Search).

## When it would be worth implementing

- **Command palette** — if a command palette is added (listed in future
  work), actions become the vocabulary the palette exposes. Each action
  maps to a screen transition or state mutation, and the palette
  displays them to the user.

- **Mouse support** — if mouse input is added, actions decouple click
  targets from keyboard shortcuts. Both input methods produce the same
  action constants.

- **Cross-screen action sharing** — if multiple screens need identical
  behavior for an action (e.g., "Back" does the same thing everywhere),
  a shared action type avoids duplicating the logic.

- **Accessibility** — screen readers or alternative input methods benefit
  from a well-defined set of actions rather than raw key bindings.

## Relationship to `events/`

The `events/` package defines messages that cross package boundaries
(TickMsg, ThemeSwitchMsg, notifications). Actions would represent
within-screen intent — a different concern. The two packages should
not overlap.

## Candidate actions

From TODO.md:

- Navigate
- Back
- Confirm
- Cancel
- Search
- Reveal answer
- Grade card
- Open deck
- Close modal
- Show help

Most of these are screen-specific. Only Navigate, Back, Confirm, Cancel,
and Show help are candidates for cross-screen reuse.

## Recommendation

Defer implementation until a concrete trigger appears:

| Trigger | Action |
|---|---|
| Command palette added | Implement actions package |
| Mouse support added | Implement actions package |
| "Back" logic duplicated in 3+ screens | Extract to shared action |

Until then, the keymap + screen Update() pattern is sufficient and
avoids premature abstraction.
