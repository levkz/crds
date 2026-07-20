## Done

- [x] `Matches(tea.KeyMsg) bool` helper on Binding
- [x] `BindingList` with `Help() string` for footer generation
- [x] Global shortcuts — `dispatchKeyEvent` uses `keymap.DefaultGlobal`
- [x] Screen-local shortcuts — Home, Quiz, Search, Settings use `keymap.DefaultList/Quiz/Search`
- [x] Focus navigation — Search uses `keymap.DefaultSearch.FocusToggle`
- [x] Action mapping — Quiz uses `keymap.DefaultQuiz.Again/Hard/Good/Easy`
- [x] Footer auto-generation — all screens derive footers from keymap
- [x] Removed hardcoded `KeyQuit`/`KeyHelp` from `app.Config`
- [x] `Registry` with `Bindings()` and `FindBinding()` for central lookup
- [x] Help overlay renders all keybindings via `Registry.Bindings()`
- [x] `KeymapConfig` + `BindingOverride` structs for user-defined overrides
- [x] `ApplyDefaultOverrides()` to apply overrides to `Default*` vars
- [x] `internal/config/` loads `~/.config/crds/keymaps.yaml` and wires into `app.New()`
- [x] Vim bindings — `"k"`/`"j"` for up/down defined in `DefaultList`

## Future

- [ ] Chord bindings — e.g. `g` then `g` for top of list
- [ ] Mouse binding support
- [ ] Per-screen keymap overrides in config
