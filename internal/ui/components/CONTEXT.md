# Components Context

This document provides implementation context for future development sessions.

---

## Purpose

The `components` package provides reusable UI rendering primitives for all
screens. Each component encapsulates a visual pattern — a flashcard, a list,
a progress bar — and exposes it as a simple function or lightweight struct.

Components sit between `styles/` (lipgloss wrappers) and `screens/` (Bubble
Tea models), assembling styled primitives into meaningful visual units.

Every component should be:
- **Reusable** — importable by any screen without modification
- **Stateless when practical** — display components are pure functions
- **Theme-aware** — always use `styles/` and `ui.Theme`, never hardcode colors
- **Responsive** — accept `width int` (and `height` where relevant)
- **Ignorant of screens** — no imports from `app/`, `screens/`, or `navigation/`

---

## Design Principles

1. **Functions over structs** — If a component has no state, it should be a
   function. Structs are only used when state (cursor, focus, selection) is
   required.

2. **No business state** — Interactive components own only ephemeral UI state:
   cursor position, selection, scroll offset, focus, viewport, blink timers.
   Business state (search query, selected deck, quiz answer) belongs to the
   parent screen and is passed in.

3. **Delegation to styles** — All visual decoration (colors, borders, padding)
   comes from `styles/`. Components compose and assemble; they do not style.

4. **Whitespace over borders** — Follow the UI philosophy: prefer spacing to
   delineate rather than drawing boxes.

5. **Width propagation** — Every render function accepts `width int` (or
   `height int` where relevant). Hardcoded widths are banned.

6. **One component per file** — Named `<name>.go`. Test files are optional
   inline or in a standalone test file.

---

## Component Taxonomy

Components are organized into six categories plus structural helpers:

### Display (stateless)
| Component | Status | File | Purpose |
|---|---|---|---|---|
| Text | ✅ | `display/text.go` | Muted single-line text |
| Label | ✅ | `display/label.go` | Field label / heading |
| Paragraph | ✅ | `display/paragraph.go` | Word-wrapped text block (uses `renderer.Wrap`) |
| Divider | ✅ | `display/divider.go` | Horizontal separator line |
| Badge | ✅ | `display/badge.go` | Status indicator (success/error/warning/info variants) |

### Containers (stateless)
| Component | Status | File | Purpose |
|---|---|---|---|
| Header | ✅ | `display/header.go` | Page title bar |
| Footer | ✅ | `display/footer.go` | Shortcut bar |
| Card | ✅ | `display/card.go` | Flashcard face (front/back/notes) |
| Panel | ✅ | `display/panel.go` | Bordered container (wraps `styles.Panel`) |
| Section | ✅ | `display/section.go` | Labeled group with title + divider |
| Group | ✅ | `display/group.go` | Titled container without border |
| Window | ✅ | `display/window.go` | Resizable container with header/content/footer |

### Lists (stateless + hybrid)
| Component | Status | File | Purpose |
|---|---|---|---|
| List | ✅ | `display/list.go` | Flat item list with single selection |
| Selectable list | ✅ | `interactive/selectable_list.go` | Rich list with multi-select (hybrid state) |
| Tree | ✅ | `interactive/tree.go` | Hierarchical/nested list with expand/collapse |
| Table | ✅ | `display/table.go` | Grid with headers and columns (auto column widths) |

### Inputs (hybrid state)
| Component | Status | File | Purpose |
|---|---|---|---|
| Text input | ✅ | `interactive/text_input.go` | Single-line text field with cursor |
| Search input | ✅ | `interactive/search_input.go` | Text field with live query change emission |
| Checkbox | ✅ | `interactive/checkbox.go` | Boolean toggle with focus state |
| Radio group | ✅ | `interactive/radio_group.go` | Single selection from options |
| Select | ✅ | `interactive/select.go` | Dropdown/picker with expand/collapse |
| Multi-select | ✅ | `interactive/multi_select.go` | Multiple selection with checkboxes |

### Feedback (stateless + hybrid)
| Component | Status | File | Purpose |
|---|---|---|---|
| Progress bar | ✅ | `display/progress.go` | Numeric progress display |
| Spinner | ✅ | `interactive/spinner.go` | Loading / async indicator with animation frames |
| Notification | ✅ | `display/notification.go` | Ephemeral feedback (merged with Toast) |
| Status bar | ✅ | `display/status_bar.go` | Persistent bottom status line (left/right aligned) |

### Dialogs (stateless)
| Component | Status | File | Purpose |
|---|---|---|---|
| Modal | ✅ | `display/modal.go` | Overlay dialog container |
| Confirmation dialog | ✅ | `display/confirm_dialog.go` | Yes/no confirm prompt |
| Error dialog | ✅ | `display/error_dialog.go` | Error message display with styling |

**Total: 29 components — all 29 implemented.**

---

## Current Components

| Component | Signature | Type | Depends on |
|---|---|---|---|
| `Text` | `func Text(content string) string` | stateless | `styles.MutedText()` |
| `Label` | `func Label(text string) string` | stateless | `styles.MutedText()` |
| `Paragraph` | `func Paragraph(content string, width int) string` | stateless | `renderer.Wrap()` |
| `Divider` | `func Divider(width int) string` | stateless | `styles.MutedText()` |
| `Badge` | `func Badge(text string, variant BadgeVariant) string` | stateless | `styles.Success/Error/Warning/Hint()`, `ui.Theme.Icons` |
| `Header` | `func Header(title string, width int) string` | stateless | `styles.Header(width)` |
| `Footer` | `func Footer(keys string, width int) string` | stateless | `styles.Footer(width)` |
| `Card` | `func RenderCard(c Card, revealed bool, width int) string` | stateless | `styles.Card(width)`, `renderer.Wrap()` |
| `Panel` | `func Panel(content string, width int) string` | stateless | `styles.Panel(width)` |
| `Section` | `func Section(title, content string, width int) string` | stateless | `styles.MutedText()`, `renderer.VisibleWidth` |
| `Group` | `func Group(title, content string, width int) string` | stateless | `styles.Hint()` |
| `Window` | `func Window(title, content, footer string, width int) string` | stateless | `Header()`, `Footer()` |
| `RenderList` | `func RenderList(items []string, selected int, width int) string` | stateless | `styles.SelectedItem()`, `renderer.Truncate()`, `ui.Theme.Icons` |
| `RenderListClipped` | `func RenderListClipped(items []string, selected int, offset int, maxItems int, width int) string` | stateless | `RenderList` + scroll indicators |
| `SelectableListModel` | `func NewSelectableList(multi bool) SelectableListModel` | hybrid | `styles.SelectedItem()`, `renderer.Truncate()`, `ui.Theme.Icons` |
| `TreeModel` | `func NewTree() TreeModel` | hybrid | `styles.SelectedItem()`, `renderer.Truncate()`, `ui.Theme.Icons` |
| `Table` | `func Table(headers []string, rows [][]string, width int) string` | stateless | `styles.MutedText()`, `renderer.Truncate()` |
| `TextInputModel` | `func NewTextInput() TextInputModel` | hybrid | `styles.FocusedInput()` |
| `SearchInputModel` | `func NewSearchInput() SearchInputModel` | hybrid | extends `TextInputModel` |
| `CheckboxModel` | `func NewCheckbox() CheckboxModel` | hybrid | `styles.MutedText()/SelectedItem()`, `ui.Theme.Icons` |
| `RadioGroupModel` | `func NewRadioGroup() RadioGroupModel` | hybrid | `styles.SelectedItem()`, `renderer.Truncate()`, `ui.Theme.Icons` |
| `SelectModel` | `func NewSelect() SelectModel` | hybrid | `styles.FocusedInput()/MutedText()` |
| `MultiSelectModel` | `func NewMultiSelect() MultiSelectModel` | hybrid | `styles.FocusedInput()/MutedText()`, `ui.Theme.Icons` |
| `ProgressBar` | `func ProgressBar(progress int) string` | stateless | `styles.MutedText()` |
| `SpinnerModel` | `func NewSpinner() SpinnerModel` | hybrid | Bubble Tea, built-in Braille frames |
| `RenderNotification` | `func RenderNotification(text string) string` | stateless | `styles.Hint()` |
| `StatusBar` | `func StatusBar(left, right string, width int) string` | stateless | `styles.MutedText()`, `renderer.VisibleWidth` |
| `RenderModal` | `func RenderModal(title, content string, width, height int) string` | stateless | `styles.Modal(width, height)` |
| `ConfirmDialog` | `func ConfirmDialog(title, message, confirm, cancel string, width, height int) string` | stateless | `RenderModal()` |
| `ErrorDialog` | `func ErrorDialog(title, message string, width, height int) string` | stateless | `RenderModal()`, `styles.Error()` |

All stateless components are pure functions. They accept data and return
render strings. The `Card` type is a data struct (`Front`, `Back`, `Notes`),
not a stateful model — the component itself is the function `RenderCard`.

Interactive components own ephemeral UI state (cursor, focus, selection,
animation frame) but not business state. Their `View()` methods or setters
accept business data from the parent screen.

---

## State Model

Components follow a three-tier state model:

### 1. Stateless (Display, Feedback, Dialogs, Lists)
Pure functions. Accept input, return a string. No struct, no methods, no
`tea.Cmd`. Used for Text, Label, Card, Modal, Divider, Badge, Panel, etc.

```go
func Badge(text string, variant BadgeVariant) string
```

### 2. Data container (hybrid data-only)
A struct holds immutable data (like `Card{Front, Back, Notes}`) but the
rendering function takes that data and returns a string. No mutations, no
`Update()`.

```go
type Card struct {
    Front string
    Back  []string
    Notes string
}

func RenderCard(c Card, revealed bool, width int) string
```

### 3. Interactive (Inputs, Selectable list, Spinner)
Struct with ephemeral UI state. Follows Bubble Tea's model pattern but
without owning business data. Business data is passed into `View()` or set
via a method.

```go
type TextInputModel struct {
    // Component owns:
    cursor  int
    focused bool

    // Screen owns (passed in):
    // value string  ← NOT stored here
}

func NewTextInput(...) TextInputModel
func (m *TextInputModel) Init() tea.Cmd
func (m *TextInputModel) Update(msg tea.Msg) (TextInputModel, tea.Cmd)
func (m *TextInputModel) View(value string, width int) string
```

The `View()` method accepts business state as parameters, keeping the
component reusable across screens.

### State ownership summary

| Owned by component | Owned by screen |
|---|---|
| Cursor position | Search query |
| Selection index | Selected deck |
| Scroll offset | Quiz answer |
| Focus state | Filters |
| Viewport | User data |
| Blink/animation timers | Application state |

---

## API Conventions

### Naming

- **Stateless components**: `func ComponentName(args...) string`
  - e.g., `func Text(content string) string`
  - e.g., `func RenderCard(c Card, revealed bool, width int) string`
- **Interactive components**: `type ComponentNameModel struct` with methods:
  - `NewComponentName(keys ...KeyConfigType) ComponentNameModel`
  - `(m *ComponentNameModel) Init() tea.Cmd`
  - `(m *ComponentNameModel) Update(tea.Msg) (ComponentNameModel, tea.Cmd)`
  - `(m *ComponentNameModel) View(businessArgs..., width int) string`

### Parameters

- Always accept `width int` for responsive rendering
- Accept `height int` where vertical space matters (Modal, Window, Table)
- Group related data into small structs (like `Card`) only when multiple
  fields are naturally cohesive

### Key configuration

Interactive components accept optional key config structs via variadic
parameter in their constructor:

```go
func NewTextInput(keys ...TextInputKeys) TextInputModel
func NewRadioGroup(keys ...NavigationKeys) RadioGroupModel
func NewCheckbox(keys ...CheckboxKeys) CheckboxModel
```

When no keys are passed, `DefaultTextInputKeys` / `DefaultNavigationKeys` /
`DefaultCheckboxKeys` are used. To customize:

```go
input := components.NewTextInput()
input := components.NewTextInput(components.TextInputKeys{
    Left:  []string{"left", "h"},
    Right: []string{"right", "l"},
})
```

The three key config structs are:

| Struct | Used by | Defaults |
|---|---|---|
| `NavigationKeys` | RadioGroup, Select, MultiSelect, SelectableList, Tree | up/k, down/j, home/g, end/G, enter, esc, space, right/l, left/h |
| `TextInputKeys` | TextInput, SearchInput | left, right, home, end, backspace, delete |
| `CheckboxKeys` | Checkbox | space, enter |

The `keyIn(msg, keys)` helper matches a `tea.KeyMsg` against the configured
key strings. See `input_keys.go` for all defaults and the helper function.

### What components must NOT do

- Import `app/`, `screens/`, or `navigation/`
- Access global application state
- Own business data
- Emit commands that modify persistence
- Construct or reference other screens

---

## Integration Points

### Dependencies

| Package | Used by | Purpose |
|---|---|---|
| `styles/` | All components | Visual styling (colors, borders, padding) |
| `renderer/` | Card, List, Paragraph | Text wrapping, truncation, width measurement |
| `ui` (theme re-exports) | List | Icon constants, theme access |
| `theme/` | (indirect via `ui`) | Theming on render |
| `github.com/charmbracelet/bubbletea` | Interactive components | `tea.Model`, `tea.Cmd`, `tea.Msg` |

### Consumers

| Consumer | Components used |
|---|---|
| `screens/home.go` | Header, Footer, List, Text, Notification |
| `screens/quiz.go` | Header, Footer, Card, ProgressBar, Text, Notification |
| `screens/search.go` | Header, Footer, List, Text, Notification, Modal |
| `screens/statistics.go` | Header, Footer, Text, ProgressBar, Notification |
| `screens/settings.go` | Header, Footer, List, Text, Notification |
| `screens/detail.go` | Header, Footer, Text, Notification |
| `app/view.go` | Text (help overlay) |

### Not yet wired into screens
- **SearchInputModel** → `screens/search.go` (currently has inline input handling)
- **SelectableListModel** → `screens/settings.go` (theme selection)
- **StatusBar** → `app/view.go` (global status)
- **SpinnerModel** → all screens (async operations)
- **ConfirmDialog** → `screens/quiz.go` (confirm quit before abandoning quiz)
- **RadioGroupModel** → theme/picker use cases
- **SelectModel** / **MultiSelectModel** → settings forms
- **TreeModel** → deck/category browser
- **Table** → statistics/leaderboard screens

### Dependencies between components
- **Section** → `styles.MutedText()` + `renderer.VisibleWidth()` (calculates divider fill)
- **Group** → `styles.Hint()` (title rendered as hint text)
- **Window** → `Header()` + `Footer()` (composes them internally)
- **Search input** → `TextInputModel` (extends with query change emission)
- **Confirm dialog** → `RenderModal()` (adds confirm/cancel labels)
- **Error dialog** → `RenderModal()` + `styles.Error()` (adds error styling)

---

## File Organization

```
components/
├── CONTEXT.md
├── TODO.md
├── display/              Stateless render functions
│   ├── text.go           Text(content)
│   ├── label.go          Label(text)
│   ├── paragraph.go      Paragraph(content, width)
│   ├── divider.go        Divider(width)
│   ├── badge.go          Badge(text, variant)
│   ├── header.go         Header(title, width)
│   ├── footer.go         Footer(keys, width)
│   ├── card.go           Card struct + RenderCard(c, revealed, width)
│   ├── panel.go          Panel(content, width)
│   ├── section.go        Section(title, content, width)
│   ├── group.go          Group(title, content, width)
│   ├── window.go         Window(title, content, footer, width)
│   ├── list.go           RenderList(items, selected, width) + RenderListClipped(items, selected, offset, maxItems, width)
│   ├── table.go          Table(headers, rows, width)
│   ├── progress.go       ProgressBar(progress)
│   ├── notification.go   RenderNotification(text)
│   ├── status_bar.go     StatusBar(left, right, width)
│   ├── modal.go          RenderModal(title, content, width, height)
│   ├── confirm_dialog.go ConfirmDialog(title, message, confirm, cancel, width, height)
│   └── error_dialog.go   ErrorDialog(title, message, width, height)
└── interactive/           Stateful models (Bubble Tea sub-models)
    ├── input_keys.go     Key config structs + keyIn() helper
    ├── text_input.go     TextInputModel (cursor, focus)
    ├── search_input.go   SearchInputModel (extends TextInput)
    ├── checkbox.go       CheckboxModel (toggle, focus)
    ├── radio_group.go    RadioGroupModel (single select)
    ├── select.go         SelectModel (dropdown)
    ├── multi_select.go   MultiSelectModel (checkbox dropdown)
    ├── selectable_list.go SelectableListModel (multi-select)
    ├── tree.go           TreeModel (expand/collapse)
    └── spinner.go        SpinnerModel (animation)
```

Screens import display components via an import alias for brevity:

```go
import components "crds/internal/ui/components/display"
```

---

## Testing

### Stateless components
- In-package tests (`package components`)
- Table-driven with expected string output
- No Bubble Tea dependency — pure function tests
- Verify: correct output given inputs, width responsiveness, edge cases
  (empty strings, very narrow width, special characters)

### Interactive components
- In-package tests with minimal Bubble Tea
- Test `Init()` returns expected command
- Test `Update()` handles all relevant `tea.KeyMsg` values → state transitions
- Test `View()` renders correctly after state changes
- Test focus/blur behavior, cursor movement, boundary conditions

### General patterns
- Use `testdata/` for multi-line expected output fixtures
- Use `go-cmp` or `reflect.DeepEqual` for struct comparisons
- Aim for coverage: normal, empty, max-length, edge widths
- Run:
  ```
  go test ./internal/ui/components/display/ -v -count=1
  go test ./internal/ui/components/interactive/ -v -count=1
  ```

---

## Suggestions

1. **Error boundaries** — Consider a `RenderError` component or pattern for
   gracefully handling component render panics in development

2. **Responsive breakpoints** — If screens pass terminal width, components
   could adapt behavior (e.g., compact mode at < 60 columns)

3. **Animation support** — Spinner and ProgressBar could accept an
   animation tick channel from the parent for frame-driven updates

4. **Composite components** — Pre-built combinations like
   `ConfirmDialog(title, message, onConfirm, onCancel)` could reduce
   boilerplate in screens

5. **Viewport component** — A scrollable viewport wrapper would be useful
   for screens with overflow content (Statistics, Detail)

6. **Keyboard navigation composability** — Consider a `FocusGroup` component
   that manages tab-order between multiple interactive components on one
   screen (e.g., Search with input + list)

7. **Wiring components to screens** — Interactive components (TextInput,
   SelectableList, etc.) are built and ready but not yet wired into any
   screen. Screens like `search.go` and `settings.go` still handle input
   inline. Wiring them to use the new components is the next step.
