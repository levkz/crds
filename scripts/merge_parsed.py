#!/usr/bin/env python3
"""Merge enriched examples from /tmp/decks/parsed/<name>.yaml back into /tmp/decks/<name>.yaml.

Usage:
    python3 scripts/merge_parsed.py <deck-name>   # e.g. health-emergency-b1
    python3 scripts/merge_parsed.py all            # all decks
"""

import os, re, sys, yaml

PARSED_DIR = "/tmp/decks/parsed"
ORIG_DIR   = "/tmp/decks"

def extract_examples_from_raw(text):
    """Parse examples from raw YAML text by extracting blocks under each entry.
    Returns dict: entry_id -> [{'text': ..., 'translation': ...}]
    """
    examples_by_id = {}
    current_id = None
    in_examples = False
    current_examples = []

    for line in text.splitlines():
        id_match = re.match(r'^-\s+id:\s+(.+)$', line)
        if id_match:
            if current_id and current_examples:
                examples_by_id[current_id] = current_examples
            current_id = id_match.group(1).strip().strip("'\"")
            current_examples = []
            in_examples = False
            continue

        if current_id is None:
            continue

        if re.match(r'^\s{2}examples:\s*$', line):
            in_examples = True
            continue

        if in_examples:
            ex_match = re.match(r'^\s{2}-\s+text:\s+(.+)$', line)
            if not ex_match:
                ex_match = re.match(r'^-\s+text:\s+(.+)$', line)
            if ex_match:
                txt = ex_match.group(1).strip().strip("'\"")
                current_examples.append({'text': txt, 'translation': ''})
                continue

            trans_match = re.match(r'^\s{4}translation:\s+(.+)$', line)
            if not trans_match:
                trans_match = re.match(r'^\s{2}translation:\s+(.+)$', line)
            if trans_match:
                tr = trans_match.group(1).strip().strip("'\"")
                if current_examples:
                    current_examples[-1]['translation'] = tr
                continue

            if line.strip() == '':
                continue

    if current_id and current_examples:
        examples_by_id[current_id] = current_examples

    return examples_by_id


def merge_deck(deck_name):
    parsed_path = os.path.join(PARSED_DIR, f"{deck_name}.yaml")
    orig_path   = os.path.join(ORIG_DIR,   f"{deck_name}.yaml")

    if not os.path.exists(parsed_path):
        print(f"  SKIP {deck_name}: no parsed file")
        return False
    if not os.path.exists(orig_path):
        print(f"  SKIP {deck_name}: no original file")
        return False

    with open(parsed_path) as f:
        raw_parsed = f.read()
    with open(orig_path) as f:
        orig = yaml.safe_load(f)

    parsed_examples = extract_examples_from_raw(raw_parsed)

    changed = 0
    for entry in orig.get("entries", []):
        eid = entry.get("id")
        if eid not in parsed_examples:
            continue
        new_examples = parsed_examples[eid]
        if new_examples and new_examples != entry.get("examples", []):
            entry["examples"] = new_examples
            changed += 1

    with open(orig_path, "w") as f:
        yaml.dump(orig, f, allow_unicode=True, default_flow_style=False, sort_keys=False)

    print(f"  OK   {deck_name}: {changed} entries updated")
    return True


def main():
    if len(sys.argv) < 2:
        print(__doc__.strip())
        sys.exit(1)

    arg = sys.argv[1]
    if arg == "all":
        names = sorted(f.removesuffix(".yaml") for f in os.listdir(PARSED_DIR) if f.endswith(".yaml"))
        for name in names:
            merge_deck(name)
    else:
        merge_deck(arg)


if __name__ == "__main__":
    main()
