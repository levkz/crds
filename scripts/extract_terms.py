#!/usr/bin/env python3
"""Extract term, id, and examples from all deck YAML files into a single YAML.

Usage:
    python3 scripts/extract_terms.py

Output:
    /tmp/decks/combined.yaml

Skips entries without an explicit `id` field (can't be safely matched back).
"""

import os
import sys
import yaml

DECKS_DIR = "/tmp/decks"
OUTPUT = os.path.join(DECKS_DIR, "combined.yaml")


def extract():
    decks = []
    for filename in sorted(os.listdir(DECKS_DIR)):
        if not filename.endswith(".yaml") or filename == os.path.basename(OUTPUT):
            continue
        path = os.path.join(DECKS_DIR, filename)
        with open(path) as f:
            deck = yaml.safe_load(f)
        if not deck or "entries" not in deck:
            continue

        deck_id = deck.get("id", "")
        deck_name = deck.get("name", "")
        entries = []
        for entry in deck["entries"]:
            if not isinstance(entry, dict):
                continue
            eid = entry.get("id", "")
            term = entry.get("term", "")
            if not eid:
                print(f"Warning: {filename}: entry '{term}' has no id, skipping", file=sys.stderr)
                continue
            raw_examples = entry.get("examples") or []
            examples = []
            for ex in raw_examples:
                if isinstance(ex, dict):
                    item = {"text": ex.get("text", "")}
                    if "translation" in ex:
                        item["translation"] = ex.get("translation", "")
                    examples.append(item)
            entries.append({"id": eid, "term": term, "examples": examples})

        decks.append({"deck_id": deck_id, "deck_name": deck_name, "entries": entries})

    combined = {"decks": decks}
    with open(OUTPUT, "w") as f:
        yaml.dump(combined, f, sort_keys=False, default_flow_style=False, allow_unicode=True)
    print(f"Wrote {OUTPUT} with {len(decks)} decks")


if __name__ == "__main__":
    extract()
