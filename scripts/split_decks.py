#!/usr/bin/env python3
"""Split combined.yaml into one file per deck under /tmp/decks/parsed/.

Usage:
    python3 scripts/split_decks.py
"""

import os
import sys
import yaml

COMBINED = "/tmp/decks/combined.yaml"
OUTPUT_DIR = "/tmp/decks/parsed"


def split():
    with open(COMBINED) as f:
        combined = yaml.safe_load(f)

    if not combined or "decks" not in combined:
        print("No decks found in combined.yaml", file=sys.stderr)
        sys.exit(1)

    os.makedirs(OUTPUT_DIR, exist_ok=True)

    for deck_entry in combined["decks"]:
        deck_id = deck_entry.get("deck_id", "")
        if not deck_id:
            print("Warning: skipping deck without deck_id", file=sys.stderr)
            continue

        out_path = os.path.join(OUTPUT_DIR, f"{deck_id}.yaml")
        with open(out_path, "w") as f:
            yaml.dump(deck_entry, f, sort_keys=False, default_flow_style=False, allow_unicode=True)
        entry_count = len(deck_entry.get("entries", []))
        print(f"Wrote {out_path} ({entry_count} entries)")


if __name__ == "__main__":
    split()
