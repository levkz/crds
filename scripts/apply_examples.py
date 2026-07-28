#!/usr/bin/env python3
"""Apply example changes from combined.yaml back to the original deck files.

Usage:
    python3 scripts/apply_examples.py

Reads /tmp/decks/combined.yaml and applies the examples to each deck file in
/tmp/decks/, matched by deck_id and entry id. Only the examples field is
changed --- all other fields are preserved.

Entries without an id in either the combined file or the original deck are
skipped with a warning.
"""

import os
import sys
import yaml

DECKS_DIR = "/tmp/decks"
COMBINED = os.path.join(DECKS_DIR, "combined.yaml")


def apply():
    with open(COMBINED) as f:
        combined = yaml.safe_load(f)

    if not combined or "decks" not in combined:
        print("No decks found in combined.yaml", file=sys.stderr)
        sys.exit(1)

    for deck_entry in combined["decks"]:
        deck_id = deck_entry.get("deck_id", "")
        if not deck_id:
            continue

        filename = f"{deck_id}.yaml"
        path = os.path.join(DECKS_DIR, filename)
        if not os.path.exists(path):
            print(f"Warning: {filename} not found, skipping deck '{deck_id}'", file=sys.stderr)
            continue

        with open(path) as f:
            orig = yaml.safe_load(f)
        if not orig or "entries" not in orig:
            print(f"Warning: no entries in {filename}, skipping", file=sys.stderr)
            continue

        combined_entries = {}
        for e in deck_entry.get("entries", []):
            eid = e.get("id", "")
            if not eid:
                print(f"Warning: combined.yaml has an entry without id in deck '{deck_id}', skipping", file=sys.stderr)
                continue
            combined_entries[eid] = e

        changed = 0
        for orig_entry in orig["entries"]:
            if not isinstance(orig_entry, dict):
                continue
            eid = orig_entry.get("id", "")
            if not eid:
                term = orig_entry.get("term", "(unknown)")
                print(f"Warning: {filename}: entry '{term}' has no id, skipping", file=sys.stderr)
                continue
            if eid in combined_entries:
                combined_examples = combined_entries[eid].get("examples")
                if combined_examples is not None:
                    orig_entry["examples"] = combined_examples
                    changed += 1

        if changed > 0:
            with open(path, "w") as f:
                yaml.dump(orig, f, sort_keys=False, default_flow_style=False, allow_unicode=True)
            print(f"Updated {filename} ({changed} entries)")
        else:
            print(f"No changes in {filename}")


if __name__ == "__main__":
    apply()
