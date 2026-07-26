# Deck Creation Guide

This guide covers how to create and manage vocabulary decks for CRDS.

---

## Where Decks Live

CRDS loads YAML deck files from:

```
~/.local/share/crds/decks/
```

Each `.yaml` file in this directory is treated as a single deck. You can also specify a custom path via the `sync` command.

Decks are plain YAML — version-control them with Git, share them, edit them in any text editor.

---

## Minimal Deck

```yaml
id: french_a1
name: French A1
language: fr
translation_language: en

entries:
  - id: bonjour
    term: bonjour
    translations:
      - text: hello
```

---

## Deck-Level Fields

| Field | Required | Description |
|---|---|---|
| `id` | yes | Unique identifier for the deck |
| `name` | yes | Human-readable name shown in the UI |
| `language` | yes | Source language code (e.g. `fr`, `de`, `es`) |
| `translation_language` | yes | Target language code (e.g. `en`) |

---

## Entry Fields

| Field | Required | Description |
|---|---|---|
| `id` | no | Unique identifier. Auto-generated if omitted (see below). |
| `term` | yes | The word or phrase to learn |
| `translations` | yes | At least one translation (`text` field) |
| `examples` | no | Usage examples with `text` and optional `translation` |
| `tags` | no | List of categorisation tags |
| `notes` | no | Free-form notes shown on the answer side |

### Examples with examples

```yaml
  - id: bonjour
    term: bonjour
    translations:
      - text: hello
      - text: good morning
    examples:
      - text: Bonjour, Marie.
        translation: Hello, Marie.
    tags:
      - greeting
      - A1
    notes: Common greeting used throughout the day.
```

---

## Auto-Generated IDs

If you omit `id` on an entry, CRDS generates one automatically from the term.

### How it works

1. The term is expanded into all its variants (see Variant Syntax below).
2. The shortest variant is selected.
3. It is sanitised into a valid identifier: lowercased, spaces/apostrophes/hyphens become `_`, letters and digits are kept.

### Examples

| Term | Auto-Generated ID |
|---|---|
| `bonjour` | `bonjour` |
| `mang[er/ez/e/ons/ent]` | `mange` (shortest of 5 variants) |
| `(un)necessary` | `necessary` |
| `s'il vous plaît` | `s_il_vous_plaît` |
| `[une/la] baguette` | `la_baguette` |

If the generated ID collides with an existing ID (explicit or auto-generated), a numeric suffix is appended: `bonjour`, `bonjour_2`, `bonjour_3`, etc.

An entry with an empty term, or one whose term produces only non-alphanumeric characters, falls back to `entry` as the base name.

---

## Variant Syntax (parens and brackets)

Both terms and translations support expansion syntax for compactly representing multiple forms.

| Syntax | Meaning | Example | Produces |
|---|---|---|---|
| `(text)` | Optional group | `(to) eat` | `eat`, `to eat` |
| `(a/b/c)` | Optional alternatives | `(he/she) eats` | `eats`, `he eats`, `she eats` |
| `[text]` | Required group | `[he] eats` | `he eats` |
| `[a/b]` | Required alternatives | `[he/she] eats` | `he eats`, `she eats` |
| `(a)(b)` | Adjacent groups | `(a)(b)` | `""`, `b`, `a`, `ab` |

### Conjugation example

```yaml
  - term: mang(er)
    translations:
      - text: (to) eat
```

Expands the term to `mang` and `manger`, and the translation to `eat` and `to eat`.

### Required-article example

```yaml
  - term: [le/la] livre
    translations:
      - text: book
```

Expands the term to `le livre` and `la livre` — both required variants because `[]` was used.

### Mixed example

```yaml
  - term: mang[er/ez/e/ons/ent]
    translations:
      - text: eat
```

The term expands to `manger`, `mangez`, `mange`, `mangeons`, `mangent`. The auto-ID picks `mange` (the shortest).

---

## Validation Rules

When a deck is parsed, CRDS checks:

- Deck-level fields (`id`, `name`, `language`, `translation_language`) are non-empty
- Every entry has a non-empty `term`
- Every entry has at least one `translation`
- No duplicate entry IDs (including auto-generated ones)
- No duplicate terms within a deck
- All YAML is well-formed

If any check fails, parsing stops and an error is reported. No partial data is loaded.

---

## Tips

- **One file per deck.** Keeps things organised and makes it easy to share individual decks.
- **Use consistent ID conventions.** Even though IDs can be auto-generated, explicit IDs like `fr_bonjour` make references stable across edits.
- **Prefer `()` for optional parts** of a term and `[]` for required alternatives. This is especially useful for verbs with conjugations or nouns with articles.
- **Keep terms concise.** The variant syntax is meant for common spelling variations, not for encoding entire paradigms.
- **Run `make test` after editing a deck file** to catch structural issues before loading the app.
