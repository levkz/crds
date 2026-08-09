# Theme Design Language

## Purpose

The theme package provides a semantic design system for the CRDS terminal UI. Every visual property is named by its purpose, not its appearance, so screens and components stay decoupled from specific color values, icon glyphs, or spacing numbers.

---

## Principles

### 1. Semantic over literal

A style is called `Primary`, not `Blue`. An icon is `Navigate`, not `ArrowRight`. A spacing value is `Md`, not `16`. This lets themes change the entire look of the application without touching screen code.

### 2. Minimal palette

Fifteen named colors. A subset drives the semantic styles (one role per
color); the remaining slots exist for custom themes and direct use. No
gradients, no opacity, no per-component overrides. If a new color is
needed, it should earn its place by serving a role that none of the
existing colors can fill.

### 3. Graceful degradation

All four icon sources (NerdFont, Emoji, Unicode, Fallback) cover the same semantics. If the terminal lacks NerdFont support, the app falls back to Emoji or Unicode without breaking layout. The Fallback set uses only ASCII.

### 4. Accessibility first

- Color should reinforce meaning, never replace it. A `Danger` style is distinct from `Success` by more than just hue.
- Foreground/background contrast for `Surface`, `Background`, and `Primary` must remain readable in both dark and light palettes.
- Semantic style colours follow a triadic scheme (e.g. blue Primary, cyan Secondary, orange Accent). Colours should reinforce meaning, never replace it.

---

## Color System

### Palette roles

| Role       | Dark | Light | Purpose                            |
| ---------- | ---- | ----- | ---------------------------------- |
| Blue       | 75   | 27    | Primary interactive elements       |
| Green      | 84   | 34    | Success states                     |
| Orange     | 215  | 208   | Warning states                     |
| Red        | 203  | 160   | Danger / destructive               |
| Gray       | 249  | 238   | Secondary / muted text             |
| White      | 255  | 235   | Primary foreground                 |
| Background | 233  | 255   | Canvas / page background           |
| Selection  | 27   | 39    | Selected / focused items           |
| Border     | 236  | 245   | Structural dividers                |
| Link       | 39   | 33    | Navigable / tappable elements      |
| Surface    | 236  | 250   | Elevated card / container surfaces |
| Magenta    | 177  | 171   | Extra palette slots for custom themes |
| Purple     | 140  | 98    | Extra palette slots for custom themes |
| Cyan       | 117  | 45    | Extra palette slots for custom themes |
| Yellow     | 220  | 208   | Extra palette slots for custom themes |

### Semantic styles

Built from the palette in `NewTheme()`:

| Style      | Color        | Attributes        | Use                                |
| ---------- | ------------ | ----------------- | ---------------------------------- |
| Primary    | Blue         | Foreground        | Links, active nav items            |
| Secondary  | Cyan         | Foreground        | Metadata, timestamps, help text    |
| Accent     | Orange       | Foreground        | Emphasis, highlighted terms        |
| Success    | Green        | Foreground        | Correct answers, confirmations     |
| Warning    | Orange       | Foreground        | Near-limit states, soft errors     |
| Danger     | Red          | Foreground        | Errors, destructive actions        |
| Muted      | Gray         | Foreground        | Disabled items, secondary info     |
| Header     | White        | Bold+Bg=Surface   | Section headers                    |
| Background | White        | Foreground        | Content on Background surface      |
| Surface    | Blue / Gray  | Bg=Surface+Fg     | Cards, panels, elevated containers |
| PrimaryBg  | White/Blue   | Bg=Blue+Fg=Back   | Primary background (tags, badges)  |
| SuccessBg  | White/Green  | Bg=Green+Fg=Back  | Success state background           |
| ErrorBg    | White/Red    | Bg=Red+Fg=Back    | Error state background             |
| WarningBg  | White/Orange | Bg=Orange+Fg=Back | Warning state background           |

Dark and light themes differ only in palette values. The style logic is identical — `NewTheme()` is the single source of truth.

### Referencing colors in Go code

Always prefer the **semantic style field** over the raw **palette key**:

| If you need…                            | Write this                                     | Not this                                 |
| --------------------------------------- | ---------------------------------------------- | ---------------------------------------- |
| Foreground color for an element         | `ui.Theme.Primary`                             | `ui.Theme.Palette.Blue`                  |
| Border color for an interactive element | `ui.Theme.Primary.GetForeground()`             | `ui.Theme.Palette.Blue`                  |
| Background for a selected item          | `Background(ui.Theme.Primary.GetForeground())` | `Background(ui.Theme.Palette.Selection)` |
| Border color for a structural panel     | `ui.Theme.Palette.Border`                      | —                                        |

**Why:** If a custom theme remaps `Primary` to use `Link` instead of `Blue`, anything written as `ui.Theme.Palette.Blue` stays blue and breaks the theme. Anything referencing `ui.Theme.Primary` follows the theme.

**Exceptions** — Direct palette access is acceptable when no semantic style exists:

- `Palette.Border` — structural dividers
- `Palette.Background` — canvas color for terminal background

### Configuring colors in YAML

Users configure colors in `config.yaml`. Every color field accepts **either** a palette key name or a direct ANSI 256‑color / hex value:

| Format      | Example     | Resolution                                     |
| ----------- | ----------- | ---------------------------------------------- |
| Palette key | `"blue"`    | Resolves to the palette's current `blue` value |
| ANSI code   | `"39"`      | Used as-is                                     |
| Hex         | `"#00afff"` | Used as-is (requires 24‑bit terminal)          |

```yaml
theme:
  palette:
    blue: "39" # direct ANSI value
    red: "160" # direct ANSI value
  typography:
    title:
      color: "blue" # resolves to palette.blue ("39")
    caption:
      color: "240" # direct ANSI value, bypasses palette
```

A palette key name (`"blue"`) always resolves through the palette. A direct value (`"240"`, `"#00afff"`) bypasses the palette entirely. This lets users build themes that reference a shared palette while also making one-off overrides without adding new palette slots.

---

## Icon System

### Semantic naming

| Icon        | Purpose                                    |
| ----------- | ------------------------------------------ |
| `Check`     | Correct answer, confirmed, active state    |
| `Cross`     | Incorrect answer, error, inactive          |
| `ArrowUp`   | Navigate up, increase                      |
| `ArrowDown` | Navigate down, decrease                    |
| `ArrowLeft` | Go back, previous                          |
| `Bullet`    | List item marker, decorative separator     |
| `Selected`  | Active menu item, current selection marker |
| `Navigate`  | Forward navigation indicator               |
| `Highlight` | Featured item, important marker            |
| `Close`     | Dismiss modal, remove item, clear          |

### Source matrix

| Source   | Selected | Navigate | Highlight | Close |
| -------- | -------- | -------- | --------- | ----- |
| NerdFont |         |         |          |      |
| Emoji    | ⭕       | ➡        | ⭐        | ❌    |
| Unicode  | •        | ▶        | ★         | ✗     |
| Fallback | \*       | >        | \*        | [ ]   |

The remaining six icons (`Check`, `Cross`, `ArrowUp`, `ArrowDown`, `ArrowLeft`, `Bullet`) come from the original design and serve consistent roles across all sources.

### Fallback strategy

1. `DetectedIcons()` — auto-detect terminal capabilities
2. `WithIconSource(s)` — explicit override via config
3. `WithFallbackIcons()` — guaranteed ASCII output

Detection runs once at startup. The icon set is immutable for the lifetime of the theme.

---

## Spacing System

### Scale

| Token | PX (cols) | Use                          |
| ----- | --------- | ---------------------------- |
| Xxs   | 2         | Tiny gutter, badge padding   |
| Xs    | 4         | Tight spacing, inline gaps   |
| Sm    | 8         | Element padding, small gaps  |
| Md    | 16        | Default padding, card margin |
| Lg    | 24        | Section spacing              |
| Xl    | 32        | Major section breaks         |
| Xxl   | 48        | Page margins, empty states   |

### Usage in components

- Dialogs use `Lg` padding around content, `Md` between sections.
- Cards use `Md` padding inside, `Sm` between lines.
- Lists use `Xs` between items, `Sm` for indentation.
- Screens use `Xl` vertical margin between header, content, and footer.

Spacing is always measured in terminal columns (`int`). No fractional units.

---

## Border System

### Roles

| Role        | Border style     | Use case                           |
| ----------- | ---------------- | ---------------------------------- |
| `Container` | Normal (`┌┐└┘`)  | Page-level containers, main panels |
| `Card`      | Rounded (`╭╮╰╯`) | Elevated cards, list items         |
| `Modal`     | Rounded (`╭╮╰╯`) | Dialog boxes, overlays             |
| `Emphasis`  | Double (`╔╗╚╝`)  | Important informational blocks     |
| `Section`   | Thick (`┏┓┗┛`)   | Section dividers, nested groups    |
| `None`      | Hidden           | Clean edges, embedded content      |

### Application

Use `Theme.BorderFor(role)` rather than selecting a border directly. This keeps the mapping centralized and allows themes to override it. For custom borders, use `Theme.Borders.*` directly.

---

## Theme Switching

Themes are registered by name and stored in `Store`. Built-in themes:

| Name       | Palette              |
| ---------- | -------------------- |
| default    | `DefaultPalette`     |
| dark       | `DarkPalette`        |
| light      | `LightPalette`       |
| tokyonight | `TokyonightPalette`  |
| mocha      | `MochaPalette`       |

Custom themes can be registered from YAML files via `store.RegisterPath()`.

---

## YAML Configuration

### Palette colors

The config palette maps directly to the `Palette` struct:

```yaml
theme:
  palette:
    blue: "75"
    green: "84"
    orange: "215"
    red: "203"
    gray: "242"
    white: "255"
    background: "233"
    selection: "27"
    border: "236"
    link: "39"
    surface: "236"
```

### Style color overrides

`primary`, `secondary`, and `accent` override which palette color each semantic style uses. The value can be a **palette key name** or a **direct color value**:

```yaml
theme:
  palette:
    primary: "green" # Primary uses palette green instead of blue
    secondary: "gray" # Secondary uses palette gray instead of cyan
    accent: "link" # Accent uses palette link instead of orange
```

This lets you remap styles without changing the underlying palette:

```yaml
theme:
  palette:
    primary: "39" # direct ANSI value, bypasses palette entirely
    accent: "#ff0000" # direct hex value (24‑bit terminal required)
```

Unknown palette keys and style names are silently ignored. Missing values use the default.

---

## Evolution

This design language is intentionally minimal. Future additions (motion tokens, elevation shadows, animation curves) should follow the same pattern: semantic naming, single source of truth, graceful fallback.
