# Scheduler Context

> Per-package context: how this package works today. Status and plans live in
> `docs/status.md` and `docs/roadmap.md` (see `docs/README.md`).

## Purpose

The `scheduler` package is the spaced-repetition service. It decides when each
card is due for review next. It is deliberately decoupled from the terminal UI,
storage, and configuration — it is a pure function of a progress record and a
grade.

## Algorithm

SM-2 (ease × interval), an Anki-style simplification:

- New cards start at ease 2.5, interval 0, due immediately.
- First review intervals by grade: Again → 0 d, Hard → 1 d, Good → 1 d, Easy → 4 d.
- Subsequent reviews: Again resets to 0 d and ease −0.2; Hard → `round(interval×1.2)`
  and ease −0.15; Good → `round(interval×ease)`; Easy → `round(interval×ease×1.3)`
  and ease +0.15. Ease never goes below 1.3.
- A review with grade `>= Good` increments `Correct`; `Again`/`Hard` increment
  `Incorrect`.
- Forward and reverse directions are tracked as separate `progress` rows.

## Key files

| File | Responsibility |
|------|----------------|
| `scheduler.go` | `Update`, `IsCorrect`, SM-2 constants |
| `scheduler_test.go` | Table-driven SM-2 transitions |

## Dependencies

- `crds/internal/model` — `model.Progress` in, `model.Progress` out.

Notably there is **no** dependency on `internal/ui` (which pulls in Bubble Tea);
grades are passed as plain ints (`GradeAgain` … `GradeEasy`).

## Integration

Storage (`internal/storage`) loads the current progress row for a
`(deck, entry, reverse)` key, calls `scheduler.Update` with the review grade,
and upserts the result. The same scheduler rule shapes "due today" stats and
the quiz's due-mode ordering.

## Notes for changes

- Swapping in another algorithm (FSRS, Leitner) only requires re-implementing
  `Update`; callers are unchanged as long as `model.Progress` semantics hold.
- Ease change order matters for the Easy path: ease is adjusted before the new
  interval is computed.