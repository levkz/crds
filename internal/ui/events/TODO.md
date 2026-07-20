# Events Package — Implementation Status

All event types are defined in `events.go`.

## Implemented

- [x] **Tick event** — `TickMsg time.Time` (moved from `app/tick.go`)
- [x] **Theme changed** — `ThemeSwitchMsg{Name string}` (moved from `ui/theme.go`)
- [x] **Notification events** — `ShowNotificationMsg{Text string}` and `HideNotificationMsg{}` (moved from `app/model.go`)

## Deferred

- [ ] Navigation event
- [ ] Resize event
- [ ] Focus event
- [ ] Blur event
- [ ] Session updated
- [ ] Review completed
