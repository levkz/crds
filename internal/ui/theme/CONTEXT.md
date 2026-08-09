# Theme Context

> Per-package context: how this package works today. Status and plans live in
> `docs/status.md` and `docs/roadmap.md` (see `docs/README.md`).

## Purpose

The `theme` package provides a centralized visual design system for the
terminal UI. It defines every visual primitive — colors, text roles,
borders, icons — so that screens and components never hardcode colors
or characters.

---

## Current State

Implemented and tested. See `docs/status.md` for the test baseline.

### What's in place

- **Theme** `theme.go` — Root struct with 14 semantic styles
  (`Primary`, `Secondary`, `Accent`, `Success`, `Warning`, `Danger`,
  `Muted`, `Header`, `Background`, `Surface`,
  `PrimaryBg`, `SuccessBg`, `ErrorBg`, `WarningBg`), a palette, typography,
  borders, icons, spacing, and border roles.

- **Palette** `palette.go` — `Palette` with 15 named colors
  (`Blue`, `Green`, `Orange`, `Red`, `Gray`, `White`,
  `Background`, `Selection`, `Border`, `Link`, `Surface`,
  `Magenta`, `Purple`, `Cyan`, `Yellow`) plus 3 semantic
  override slots (`Primary`, `Secondary`, `Accent`).
  `DefaultPalette` uses ANSI 256‑color values (`"39"`, `"42"`, etc.).

- **Typography** `typography.go` — 6 text role styles built from the
  palette: `Title`, `Subtitle`, `Body`, `Caption`, `Emphasis`, `Key`.

- **Borders** `borders.go` — 5 lipgloss border styles:
  `Normal`, `Rounded`, `Double`, `Thick`, `None`.

- **Icons** `icons.go` — 4 named icon sets with 10 semantic slots,
  selected by detection priority:

  | Source     | Check | Cross | Up | Down | Bullet | Selected | Navigate | Highlight | Close |
  |------------|-------|-------|----|------|--------|----------|----------|-----------|-------|
  | NerdFont   |     |     |  |    |      |        |        |         |     |
  | Emoji      | ✅    | ❌    | ⬆ | ⬇   | ⭕     | ⭕      | ➡       | ⭐        | ❌    |
  | Unicode    | ✓     | ✗     | ▲ | ▼   | •      | •       | ▶       | ★         | ✗     |
  | Fallback   | [x]   | [ ]   | ^ | v   | *      | *       | >       | *         | [ ]   |

  An `IconSource` enum (`Auto`, `NerdFont`, `Emoji`, `Unicode`,
  `Fallback`) controls which set is active via `IconsFromSource()`.

- **Auto-detection** `detect.go` — `DetectIconSource()` resolves in
  priority order:

  1. `CRDS_ICON_SOURCE` env override
  2. `NerdFontSupported()` — checks `CRDS_NERD_FONT` env + `TERM`
  3. `EmojiSupported()` — checks `COLORTERM` + Unicode
  4. `UnicodeSupported()` — checks `LC_ALL` / `LC_CTYPE` / `LANG`
  5. `FallbackIcons` — ASCII safe

- **NerdFont** `nerdfont.go` — `NerdFontSupported()` detects via
  `CRDS_NERD_FONT=1` or well-known TERM values
  (xterm-256color, kitty, alacritty, wezterm, foot, tmux, etc.).

- **YAML Loading** `config.go` — `Config` struct with `yaml` tags.
  `LoadTheme(path)` reads a YAML file and builds a full `Theme`.
  `ParseTheme(data)` works from raw bytes. Unset fields inherit
  from `DefaultPalette`. Invalid icon values return an error.
  `ConfigTypography` + `ConfigTextRole` allow per-role color, bold,
  and italic overrides referencing palette color names or direct values.
  `ConfigPalette.Primary`/`Secondary`/`Accent` override semantic style
  colors by palette key name or direct ANSI/hex value.
  `paletteColor()` resolves palette key names then falls back to
  `resolveDirectColor()` for ANSI (`"39"`) and hex (`"#00afff"`) values.

- **Spacing** `spacing.go` — `Spacing` struct with 7-tier scale
  (`Xxs=2`, `Xs=4`, `Sm=8`, `Md=16`, `Lg=24`, `Xl=32`, `Xxl=48`).

- **Border roles** `border_role.go` — `BorderRole` enum with 6 semantic
  roles (`Container`→Normal, `Card`→Rounded, `Modal`→Rounded,
  `Emphasis`→Double, `Section`→Thick, `None`→Hidden). Convenience
  method `Theme.BorderFor(role)` for centralized border selection.

- **Design language** `DESIGN.md` — Documents the full design system:
  color principles, semantic style roles, icon semantics, spacing scale,
  border role usage, theme switching, and YAML configuration format.

- **Built-in presets** `presets.go` — `DarkPalette`/`DarkTheme()`,
  `LightPalette`/`LightTheme()`, and `TokyonightPalette`/`TokyonightTheme()`
  presets for dark, light, and TokyoNight backgrounds.

- **Store & Switching** `store.go` — `Store` is a named‑theme registry:
  `Register(name, Theme)`, `Switch(name)` (returns the Theme),
  `Current()`, `CurrentName()`, `Names()`, `Get()`, `Unregister()`,
  `Has()`, `Len()`. Pre-registers `"default"`, `"dark"`, `"light"`, `"tokyonight"`,
  `"mocha"`.
  Package‑level functions delegate to `DefaultStore`.

  Callers switch by:
  ```go
  th, err := theme.Switch("dark")
  ui.SetTheme(th)   // updates global ui.Theme
  ```

- **Backward compat** — `ui/theme.go` re-exports `theme.Default` as
  `ui.Theme` and provides `ui.SetTheme()`. All existing
  `ui.Theme.Primary` / `ui.Theme.Muted` calls continue to work.

---

## Integration into `app/`

The theme package is wired into the app at startup and at runtime:

- **Config file loading** — `app.New()` accepts a `Config` with optional
  `ThemePath`. If set, the theme is loaded via `theme.LoadTheme()` and
  registered as `"_loaded"`, then switched to. Errors are logged but
  non-fatal (falls back to default).

- **Runtime switching** — The Settings screen lists all registered theme
  names with keyboard navigation (↑/↓/j/k/enter) and a theme‑aware icon
  indicator for the current theme. On selection, it emits
  `ThemeSwitchMsg`, which the root model handles by calling
  `theme.Switch(name)` + `ui.SetTheme(th)`.

- **Live styles** — The `styles/` package wraps `ui.Theme` in stateless
  functions called at render time, so every render re-reads `ui.Theme`.

---

## Relationship to `ui/`

The `ui` package re-exports the `Default` theme:

```go
// internal/ui/theme.go
var Theme = theme.Default
func SetTheme(t theme.Theme) { Theme = t }
```

Callers (`app/view.go`, `components/*.go`) use `ui.Theme.Primary`,
`ui.Theme.Muted`, etc. When the theme is switched, `ui.SetTheme(th)`
updates the global so all callers see the new styles.

Components that call style functions at render time (the current pattern)
automatically pick up the switched theme.

---

## File Structure

```
theme/
├── CONTEXT.md
├── DESIGN.md          Design language documentation
├── theme.go          Theme struct, NewTheme, NewTerminalTheme,
│                     WithIconSource, WithFallbackIcons, Default,
│                     BorderFor method
├── palette.go        Palette struct + DefaultPalette (15 colors + 3 semantic overrides)
├── typography.go     Typography struct (Title, Subtitle, Body,
│                     Caption, Emphasis, Key) + NewTypography
├── borders.go        Borders struct (Normal, Rounded, Double,
│                     Thick, None) + DefaultBorders
├── icons.go          IconSource enum, Icons struct (10 slots),
│                     4 named icon sets, IconsFromSource, DefaultIcons
├── spacing.go        Spacing struct + DefaultSpacing (7-tier scale)
├── border_role.go    BorderRole enum + defaultBorderForRole map
├── detect.go         UnicodeSupported, EmojiSupported,
│                     DetectIconSource, DetectedIcons
├── nerdfont.go       NerdFontSupported (env + TERM detection)
├── config.go         Config + ConfigPalette (15 colors + 3 style
│                     overrides) + ConfigTypography + ConfigTextRole
│                     YAML structs, LoadTheme, ParseTheme,
│                     Config.Build, paletteColor, resolveDirectColor,
│                     applyTextRole
├── presets.go        DarkPalette, LightPalette, TokyonightPalette,
│                     MochaPalette, DarkTheme, LightTheme,
│                     TokyonightTheme, MochaTheme
├── store.go          Store (Register, Switch, Current, Names,
│                     etc.), DefaultStore, package-level convs
└── testdata/
    ├── full.yaml         15 palette colors + 3 style overrides +
    │                     nerdfont icons + typography
    ├── partial.yaml      2 palette fields, no icons
    ├── icons_only.yaml   Only icons: emoji
    ├── typography.yaml   Only typography overrides (color/bold/italic)
    ├── invalid_icons.yaml  Unknown icon source "nope"
    └── malformed.yaml    Broken YAML
```

---

## Dependencies

- `github.com/charmbracelet/lipgloss` — `lipgloss.Style`,
  `lipgloss.Color`, `lipgloss.Border`, `lipgloss.HiddenBorder()`, etc.
- `go.yaml.in/yaml/v3` — YAML parsing in `config.go`

No dependency on Bubble Tea, terminal detection, or the `ui/app` package.

---

## Testing

All tests are in-package (`package theme`), table-driven where practical.
See `docs/status.md` for the test baseline.

Run:

```
go test ./internal/ui/theme/ -v -count=1
```

Coverage areas: palette values (18 fields: 15 colors + 3 semantic overrides), style rendering (14 styles),
custom palette construction, typography styles, border edge characters,
border rendering, spacing defaults, border role defaults and Theme.BorderFor(),
Unicode env detection (LC_ALL / LC_CTYPE / LANG), emoji env detection
(COLORTERM), NerdFont TERM matching + override, icon source priority chain,
override via CRDS_ICON_SOURCE, icons from source, semantic icon defaults
across all 4 sources, store creation/register/switch/names/get/unregister,
convenience functions, YAML loading (full with style overrides, partial,
icons-only, typography, empty, invalid, malformed, missing file),
unknown color error in typography config, dark/light preset themes
(surface values), built-in theme registration, theme switching with custom
palette, theme border role resolution.

Planned enhancements (CLI `--theme` flag, more `ConfigTextRole` fields,
plugin theme loading) are tracked in `docs/roadmap.md`.
