# Actions TODO

## Status: Deferred

See CONTEXT.md for analysis.

The `keymap/` package already encodes user intent via `Binding.Match()`.
Adding an actions layer solves no concrete problem until one of these
triggers appears:

- [ ] Command palette (actions become the palette vocabulary)
- [ ] Mouse support (actions decouple click targets from keys)
- [ ] "Back" logic duplicated in 3+ screens (extract to shared action)
- [ ] Accessibility / alternative input methods

## Candidate actions

Screen-specific (low value as shared types):

- [ ] Reveal answer (Quiz only)
- [ ] Grade card (Quiz only)
- [ ] Open deck (Home only)
- [ ] Search (Search only)

Cross-screen (higher value if actions package exists):

- [ ] Navigate
- [ ] Back
- [ ] Confirm
- [ ] Cancel
- [ ] Close modal
- [ ] Show help
