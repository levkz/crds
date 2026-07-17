Responsible for moving between screens.

- [x] Screen manager (`manager.go`)
- [x] Screen registry (`registry.go` — `Register`, `RegisterFactory`, `Get`, `Has`, `Remove`, `Len`)
- [x] Navigation stack (`stack.go`)
- [x] Push screen
- [x] Pop screen
- [x] Replace screen
- [x] Back navigation (via `Pop()`)
- [x] Forward navigation (`Forward()`, `CanGoForward()` — forward stack + `ForwardEvent`)
- [x] Modal navigation (`PushModal()`, `DismissModal()`, `IsModalActive()`, `ModalDepth()` — separate modal stack + `ModalPushEvent`/`ModalPopEvent`)
- [x] Overlay navigation (`ShowOverlay()`, `HideOverlay()`, `IsOverlayActive()` — non-fullscreen overlay, separate from screens/stacks + `OverlayShownEvent`/`OverlayHiddenEvent`)
- [x] Navigation events (`events.go` — `PushEvent`, `PopEvent`, `ReplaceEvent`, `ResetEvent`)
- [x] Navigation history (`SetMaxHistory(n)`, `FullHistory()`, `HistoryDepth()` — depth-limited back stack + full breadcrumb trail; `Stack.SetLimit(n)` for dynamic limits)
