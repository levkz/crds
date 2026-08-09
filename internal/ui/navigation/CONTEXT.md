# Navigation Context

> Per-package context: how this package works today. Status and plans live in
> `docs/status.md` and `docs/roadmap.md` (see `docs/README.md`).

## Purpose

The `navigation` package manages screen transitions within the UI application.

It provides a stack-based navigation model with history, push/pop/replace
operations, and support for modal, overlay, and forward navigation.

---

## Current State

Implemented and tested. See `docs/status.md` for the test baseline.

### What's in place

- **`Manager`** — top-level struct holding navigation state. Exposes `Push`,
  `Pop`, `Replace`, `Forward`, `Reset`, `PushModal`, `DismissModal`,
  `ShowOverlay`, `HideOverlay`. Each method returns a typed event.
- **`Stack`** — ordered list with depth-limit support, `Push`/`Pop`/`Peek`/`All`/`SetLimit`.
- **`Registry`** — maps `ScreenIndex` → screen instance or lazy factory.
- **Events** — 9 typed event structs (`PushEvent`, `PopEvent`, `ReplaceEvent`,
  `ForwardEvent`, `ResetEvent`, `ModalPushEvent`, `ModalPopEvent`,
  `OverlayShownEvent`, `OverlayHiddenEvent`).
- **No Bubble Tea dependency** in production code — importers include it
  transitively through `internal/ui` only.
- **Tests are external** — in `tests/` subdirectory as `package navigation_test`.

---

## Relationship to `app/`

`internal/ui/app/` now delegates to the navigation package:

- `Model` has a `Navigator *nav.Manager` field instead of a raw
  `Current ScreenIndex`
- `transitionTo()` calls `m.Navigator.Replace(screen)` with lifecycle hooks
  (flat navigation, no history)
- `pushTo()` calls `m.Navigator.Push(screen)` with lifecycle hooks
  (stacked navigation, enables back via history stack)
- After entering a screen, `app/lifecycle.go` also calls `syncActiveScreen()`,
  which pushes the current `ui.AppState` snapshot to screens implementing
  `ui.StateSyncer` (Global State Sync protocol)
- `NavigateToMsg.Screen` is `ui.ScreenIndex`
- `forwardToScreen()` and `View()` read `m.Navigator.Current`
- App is initialized with `nav.New(HomeScreen)` in `app.New()`

### Navigation patterns

- **Flat** (`transitionTo` / `Replace`): Home → Decks, Home → Search,
  Decks → Home. No history entry; Esc returns to Home directly.
- **Stacked** (`pushTo` / `Push`): Search → Detail. Pushes Search onto
  the history stack; Esc from Detail pops back to Search.

### Not yet integrated from the navigation package

- `Push`/`Pop` — now used for Search → Detail navigation. Other transitions
  still use `Replace()` (flat transitions without history).
- Modal — `PushModal`/`DismissModal` available but unused.
- Overlay — `ShowOverlay`/`HideOverlay` available but app has its own
  `OverlayType` in `GlobalState`.

---

## Design Goals

- **Centralized** — Single source of truth for screen transitions
- **Stack-based** — History stack supports back/forward navigation
- **Testable** — Navigation logic is pure data, no UI dependency
- **Extensible** — Push, pop, replace, reset, modal, overlay variants
- **Event-driven** — Navigation changes emit typed events

---

## File Structure

```
navigation/
├── CONTEXT.md
├── manager.go          Manager struct + Push, Pop, Replace, Forward,
│                       Reset, SetRegistry, CurrentScreen, ShowOverlay,
│                       HideOverlay, SetMaxHistory, FullHistory, etc.
├── stack.go            Stack with Push/Pop/Peek/All/SetLimit
├── registry.go         Registry with Register/RegisterFactory/Get/Has/Remove
├── events.go           9 event types (push/pop/replace/forward/reset/
│                       modal/overlay)
└── tests/
    ├── helpers_test.go     Shared constants + mockScreen
    ├── stack_test.go       8 Stack tests
    ├── manager_test.go     14 Manager + CurrentScreen integration tests
    ├── registry_test.go    9 Registry tests
    ├── forward_test.go     10 Forward navigation tests
    ├── modal_test.go       11 Modal navigation tests
    ├── overlay_test.go     11 Overlay navigation tests
    └── history_test.go     11 History/depth-limit tests
```

---

## Dependencies

The `navigation` package:

- Depends on `internal/ui` for `ScreenIndex` and `Screen`
- Does NOT depend on screen implementations (`HomeModel`, `QuizModel`, etc.)
- Does NOT import Bubble Tea directly — only transitively via `internal/ui`
- Does NOT depend on `internal/ui/app` — it's the other way around

---

## Testing

All tests live in `tests/` as an external package (`navigation_test`).
This enforces black-box testing — only the public API is tested.

```go
mgr := nav.New(HomeScreen)
mgr.Push(QuizScreen)
mgr.Pop() // → HomeScreen
```

Run with:

```
go test ./internal/ui/navigation/tests/
```
