## `app/`

Responsible for the lifecycle of the UI application.

### Status

All core app scaffolding is in place. Remaining work is per-screen implementation
(`internal/ui/screens/` and `internal/ui/components/`).

## Application Model

### Potential gaps to consider:

- Model initialization beyond basic screen structs
- State validation and invariants
- Model cleanup/shutdown hooks (partially implemented in ShutdownCmd)
- State serialization (if needed for persistence)
- Integration with domain models (currently uses stubs)
- Model testing (no tests yet)

Recommendation:

The application model appears substantially complete for the current architecture. The main remaining consideration is whether to expand the screen model stubs to be production-ready or keep them as placeholders for full implementation.

### Checklist

- [x] **Application model** — Root `Model` struct with `GlobalState`, `Config`,
      `ScreenIndex` navigation, all screen model instances, and terminal dimensions.
      Defined in `model.go`.

- [x] **Bubble Tea model** — `Init`, `Update`, `View` methods on `Model` satisfying
      `tea.Model`. Init starts the tick loop; Update delegates to `dispatchEvent`;
      View renders the active screen (or overlay/notification).

- [x] **Initialization** — `New(deps Dependencies)` creates the root model with
      defaults and injected dependencies. `Run(deps Dependencies)` starts the
      Bubble Tea program in the alternate screen buffer (`app.go`).

- [x] **Shutdown** — `ShutdownMsg` + `ShutdownCmd()` for graceful cleanup. Ctrl+C
      sequences shutdown before `tea.Quit` (`lifecycle.go`, `events.go`).

- [x] **Global state** — `GlobalState` struct with `Overlay`, `Notification`,
      `Loading` fields. `WithOverlay`, `WithNotification`, `WithLoading` helpers
      return immutable copies (`model.go`).

- [x] **Tick/update loop** — `TickMsg` fired every 1 second via `tea.Tick`,
      re-armed on each tick. Started in `Init()` (`tick.go`).

- [x] **Command dispatch** — `Dispatcher` struct holding injected dependencies
      (`DeckProvider`, `ProgressRecorder`). Command constructors
      (`ListDecksCmd`, `LoadDeckCmd`, `RecordAnswerCmd`) return typed
      `tea.Cmd` values. `Cmd`/`Dispatch` helpers in `commands.go`.

- [x] **Event dispatch** — Central `dispatchEvent(msg)` router in `events.go`.
      Root-level events (window resize, overlays, notifications, navigation,
      command results, tick) handled directly. Everything else forwarded to the
      active screen via `forwardToScreen`. Overlay blocks screen input.

- [x] **Screen lifecycle** — `Lifecycle` interface with `OnEnter`/`OnLeave`
      hooks. `transitionTo(screen)` calls `OnLeave` on the current screen,
      switches, then calls `OnEnter` on the new screen. `NavigateToMsg` and
      `esc` key both use `transitionTo` (`lifecycle.go`).

- [x] **Configuration** — `Config` struct with keybinding overrides,
      animation toggle, and quiz defaults. `DefaultConfig()` provides
      sensible values. `ConfigUpdatedMsg` for runtime changes (`config.go`).

- [x] **Dependency injection** — `Dependencies` struct bundles `DeckProvider`
      and `ProgressRecorder` interfaces. Injected through `New(deps)` and
      wired into the `Dispatcher`. Domain types (`DeckData`, `CardData`) are
      defined in `dependencies.go` to keep the UI decoupled from storage/parser.
