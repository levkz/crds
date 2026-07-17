# Navigation Context

## Purpose

The `navigation` package manages screen transitions within the UI application.

It provides a stack-based navigation model with history, push/pop/replace
operations, and support for modal, overlay, and forward navigation.

---

## Current State

All 14 original TODO items are implemented and tested (82 tests).

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
- `NavigateToMsg.Screen` is `ui.ScreenIndex`
- `forwardToScreen()` and `View()` read `m.Navigator.Current`
- App is initialized with `nav.New(HomeScreen)` in `app.New()`

### Not yet integrated from the navigation package

- `Registry` — screen stubs in `app/screens.go` don't implement `ui.Screen`
  yet (their `Update` returns `(Model, tea.Cmd)`, not just `tea.Cmd`).
  Once they implement `ui.Screen`, the switch in `forwardToScreen` and
  `View()` can be replaced with `m.Navigator.CurrentScreen()`.
- `Push`/`Pop` — app currently uses `Replace()` (flat transitions without
  history). Switching to `Push`/`Pop` enables back/forward navigation.
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
├── TODO.md
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

---

## Future Work

### Near-term (within `app/` integration)

- Refactor screen stubs in `app/screens.go` to implement `ui.Screen`
- Use `nav.Registry` + `mgr.CurrentScreen()` to replace the `forwardToScreen`
  and `View` switches
- Wire `Push`/`Pop` for history-based navigation (back button)
- Replace `NavigateToMsg` with direct Manager calls

### Longer-term (new navigation features)

- **Forward navigation** — already implemented in the navigation package;
  `app/` could expose a "forward" keybinding
- **Stacked overlays** — currently single-overlay; could support stacking
- **Navigation middleware** — hooks that fire on every transition (logging,
  analytics, validation)
- **Deep linking** — parse a path string into a sequence of Push/Pop calls
