# Components TODO

## Done — All 29 Components Implemented

### Display (`display/`)
- [x] `Text` — muted single-line text
- [x] `Label` — field label / heading
- [x] `Paragraph` — word-wrapped text block using `renderer.Wrap()`
- [x] `Divider` — horizontal separator line
- [x] `Badge` — status indicator (success/error/warning/info variants)

### Containers (`display/`)
- [x] `Header` — page title bar
- [x] `Footer` — shortcut bar
- [x] `Card` — flashcard face with front/back/notes
- [x] `Panel` — bordered container wrapping `styles.Panel(width)`
- [x] `Section` — labeled group with title + divider
- [x] `Group` — titled container without border
- [x] `Window` — resizable container composing Header/Content/Footer

### Lists (`display/` + `interactive/`)
- [x] `RenderList` — flat item list with single selection (`display/list.go`)
- [x] `SelectableList` — rich list with multi-select, cursor, selection state (`interactive/selectable_list.go`)
- [x] `Tree` — hierarchical list with expand/collapse (`interactive/tree.go`)
- [x] `Table` — grid with headers and auto-sized columns (`display/table.go`)

### Inputs (`interactive/`)
- [x] `TextInput` — single-line text field with cursor navigation
- [x] `SearchInput` — text field with query change emission (extends TextInput)
- [x] `Checkbox` — boolean toggle with focus state
- [x] `RadioGroup` — single selection from options
- [x] `Select` — dropdown/picker with expand/collapse
- [x] `MultiSelect` — dropdown with checkboxes

### Feedback (`display/` + `interactive/`)
- [x] `ProgressBar` — numeric progress display (`display/progress.go`)
- [x] `Spinner` — loading indicator with Braille animation frames (`interactive/spinner.go`)
- [x] `RenderNotification` — ephemeral feedback (`display/notification.go`)
- [x] `StatusBar` — persistent bottom status line (`display/status_bar.go`)

### Dialogs (`display/`)
- [x] `RenderModal` — overlay dialog container
- [x] `ConfirmDialog` — yes/no prompt with confirm/cancel labels
- [x] `ErrorDialog` — error message display with Error styling

---

## Next Steps

1. **Wire interactive components into screens** — SearchInputModel → search.go,
   SelectableListModel → settings.go, StatusBar → app/view.go, SpinnerModel →
   all screens, ConfirmDialog → quiz.go, TreeModel → deck browser
2. **Add tests** for the 20 display and 9 interactive components
3. **FocusGroup component** — manage tab-order between multiple interactive
   components on one screen
