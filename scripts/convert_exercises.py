#!/usr/bin/env python3
"""Convert legacy exercise .txt files to deck.yaml format.

Usage:
    python3 scripts/convert_exercises.py exercises/autre.txt
    python3 scripts/convert_exercises.py exercises/*.txt -o decks/

Legacy format per line:
    #SECTION_NAME
    term => translation1 / translation2 => example sentence
"""

import argparse
import os
import re
import sys
import unicodedata


def slugify(text: str) -> str:
    text = text.lower().strip()
    text = unicodedata.normalize("NFKD", text)
    text = text.encode("ascii", "ignore").decode("ascii")
    text = re.sub(r"[^\w\s-]", "", text)
    text = re.sub(r"[\s_]+", "_", text)
    return text.strip("_")


def parse_exercise(path: str) -> list[dict]:
    entries = []
    current_tag = None

    with open(path, encoding="utf-8") as f:
        for line in f:
            line = line.rstrip("\n")
            if not line.strip():
                continue

            if line.startswith("#"):
                current_tag = line.lstrip("#").strip()
                continue

            parts = [p.strip() for p in line.split("=>")]
            if len(parts) < 2:
                continue

            term = parts[0]
            translations_str = parts[1]
            example = parts[2] if len(parts) > 2 else ""

            translations = [t.strip() for t in translations_str.split("/") if t.strip()]

            entries.append({
                "term": term,
                "translations": translations,
                "example": example,
                "tag": current_tag,
            })

    return entries


def yaml_str(s: str) -> str:
    """Quote a string for YAML if it contains special chars."""
    if not s:
        return '""'
    if re.search(r"[:{}\[\],&*?|>!%@`#\-']", s) or s.startswith(" ") or s.endswith(" "):
        escaped = s.replace('"', '\\"')
        return f'"{escaped}"'
    return s


def build_deck(path: str, deck_id: str | None = None, name: str | None = None) -> str:
    basename = os.path.splitext(os.path.basename(path))[0]
    entries_raw = parse_exercise(path)

    deck_id = deck_id or slugify(basename)
    deck_name = name or basename.replace("_", " ").replace("-", " ").title()

    lines = [
        f"id: {yaml_str(deck_id)}",
        f"name: {yaml_str(deck_name)}",
        "language: fr",
        "translation_language: en",
        "",
        "entries:",
    ]

    seen_ids: dict[str, int] = {}
    for e in entries_raw:
        sid = slugify(e["term"])
        if sid in seen_ids:
            seen_ids[sid] += 1
            sid = f"{sid}_{seen_ids[sid]}"
        else:
            seen_ids[sid] = 0

        entry_id = f"{deck_id}_{sid}"
        lines.append(f"  - id: {yaml_str(entry_id)}")
        lines.append(f"    term: {yaml_str(e['term'])}")
        lines.append("    translations:")
        for t in e["translations"]:
            lines.append(f"      - text: {yaml_str(t)}")

        if e["example"]:
            lines.append("    examples:")
            lines.append(f"      - text: {yaml_str(e['example'])}")
            lines.append('        translation: ""')

        if e["tag"]:
            lines.append("    tags:")
            lines.append(f"      - {yaml_str(slugify(e['tag']))}")

        lines.append("")

    return "\n".join(lines) + "\n"


def main():
    parser = argparse.ArgumentParser(description="Convert legacy exercise .txt to deck.yaml")
    parser.add_argument("files", nargs="+", help="Exercise .txt files to convert")
    parser.add_argument("-o", "--output-dir", default=".", help="Output directory (default: current dir)")
    parser.add_argument("--id", help="Override deck ID (only with single file)")
    parser.add_argument("--name", help="Override deck name (only with single file)")
    args = parser.parse_args()

    if (args.id or args.name) and len(args.files) > 1:
        print("error: --id and --name can only be used with a single file", file=sys.stderr)
        sys.exit(1)

    os.makedirs(args.output_dir, exist_ok=True)

    for path in args.files:
        deck_yaml = build_deck(path, deck_id=args.id, name=args.name)
        # extract id from generated yaml for filename
        first_line = deck_yaml.split("\n")[0]
        deck_id_val = first_line.split(":", 1)[1].strip().strip('"')
        out_name = slugify(deck_id_val) + ".yaml"
        out_path = os.path.join(args.output_dir, out_name)

        with open(out_path, "w", encoding="utf-8") as f:
            f.write(deck_yaml)

        entry_count = deck_yaml.count("\n  - id:")
        print(f"{path} -> {out_path} ({entry_count} entries)")


if __name__ == "__main__":
    main()
