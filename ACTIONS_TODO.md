# ACTIONS_TODO — Missing CLI Commands

Commands that would have saved time during the deck migration and classification workflow.

## 1. Deck listing

```bash
crds deck list                       # all decks with entry counts
crds deck list --tags noun --tags B1 # filter by tags
```

Output columns: `deck_id | name | entries | levels | languages`

## 2. Entry management

All subcommands accept `-t <text>` for inline input or `-f <file>` for piped/file input.
Without either flag, open `$EDITOR` (current behavior of `deck edit`).

```bash
# Add
crds deck term add <deck> -t "le bras" --translations "arm" --examples "Il muscle le bras." --tags noun,masculin,A1
crds deck term add <deck> -f entry.yaml        # read from YAML file
cat entry.yaml | crds deck term add <deck> -f - # pipe from stdin

# Edit
crds deck term edit <deck> <term_id> -t "le bras" --translations "arm,Arm (anatomy)"
crds deck term edit <deck> <term_id> -f updated.yaml

# Delete
crds deck term delete <deck> <term_id>
crds deck term delete <deck> <term_id> --yes    # skip confirmation
```

## 3. Tag management

```bash
# Add tags to an entry
crds deck tag add <deck> <term_id> noun masculin A1

# Remove tags from an entry
crds deck tag remove <deck> <term_id> noun

# List tags on an entry
crds deck tag list <deck> <term_id>

# Bulk: add tags to all entries matching a filter
crds deck tag add --deck <deck> --filter "term starts with le" noun masculin
```

## 4. Search

```bash
crds deck search <query>                                  # basic search across all decks
crds deck search <query> --deck <deck>                    # limit to one deck
crds deck search <query> --deck <d1> --deck <d2>          # multiple decks

# Object type filter
crds deck search --object-type entry <query>              # find entries (default)
crds deck search --object-type tag <query>                # find tags by name

# Tag filter (entries must have ALL listed tags)
crds deck search --tags noun --tags A1                    # nouns at A1 level
crds deck search --tags feminin --deck body-parts-a1      # feminine nouns in a deck

# Deck filter (search only in listed decks)
crds deck search --deck emotions-b1 --deck sleep-b1 "tired"

# Fuzzy find mode (requires fzf)
crds deck search --fuzzy-find                             # interactive fzf picker
crds deck search --fuzzy-find --tags verb --deck advanced-verbs-b2

# Output formats
crds deck search "bras" --format table                    # default: grouped by deck
crds deck search "bras" --format yaml                     # raw YAML output
crds deck search "bras" --format json                     # JSON output
```

### Search output (table format, default)

Entries grouped by deck:
```
=== body-parts-a1 (Body Parts, A1) — 2 matches ===
  le bras          noun,masc,A1  → arm
  la jambe         noun,fem,A1   → leg

=== body-parts-a2 (Body Parts, A2) — 1 match ===
  le genou         noun,masc,A2  → knee
```

Tags listing (`--object-type tag`):
```
=== tag: noun ===
  body-parts-a1:   8 entries
  emotions-b1:     3 entries
  connectors-b1:   2 entries
  total:          13 decks, 115 entries

=== tag: feminin ===
  body-parts-a1:   4 entries
  health-wellness: 2 entries
  total:           6 decks, 54 entries
```

## 5. Stats enhancements

```bash
crds stats                              # overview (existing)
crds stats --deck <deck>                # per-deck stats
crds stats --tag noun                   # stats filtered by tag
crds stats --level B1 --deck emotions-b1  # combined filters
crds stats --object-type entry --tags verb --decks advanced-verbs-b2
```

## 6. Import/export enhancements

```bash
crds deck import <file> --replace       # overwrite existing deck
crds deck import <dir>                  # import all .yaml files from directory
crds deck export --all --format yaml    # export all decks
crds deck export --deck <deck> -f out.yaml  # export to file
```

## Priority

| Command | Impact | Effort |
|---------|--------|--------|
| `deck list` | High — can't see what's in the DB | Low |
| `deck search --tags --deck` | High — core workflow for classification | Medium |
| `deck term add -t/-f` | High — enables scripting/piping | Low |
| `deck tag add/remove/list` | High — tag management is essential | Low |
| `deck search --fuzzy-find` | Medium — nice UX for interactive use | Medium |
| `deck import --replace` | Medium — avoids delete+import cycle | Low |
| `stats --deck --tag` | Low — useful but not blocking | Low |
