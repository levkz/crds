#!/usr/bin/env bash
#
# check-docs.sh — verify documentation hygiene.
#
# 1. Every path referenced in docs/README.md resolves to a real file.
# 2. Status-only content (Known Issues, test counts) lives only in docs/status.md.
# 3. CONTEXT.md files use the fixed skeleton and carry the pointer header.
#
# Exit code 0 = all good; 1 = problems found.

set -u

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fail=0

note()  { printf '  ok: %s\n' "$*"; }
warn()  { printf '  FAIL: %s\n' "$*" >&2; fail=1; }

echo "== Referenced paths in docs/README.md =="
# Extract backticked relative paths (doc.md / docs/*.md / internal/**/CONTEXT.md).
# Keep paths that contain a directory component, plus root-level AGENTS.md/README.md.
# Bare prose references like `CONTEXT.md` (generic term) are not file paths.
mapfile -t refs < <(grep -oE '`[./]?([A-Za-z0-9_./-]+\.(md|sh))`' \
  "$ROOT/docs/README.md" | tr -d '`' | sort -u | \
  grep -E '(/|^AGENTS\.md$|^README\.md$)')

if [ ${#refs[@]} -eq 0 ]; then
  warn "no backticked paths found in docs/README.md"
else
  for ref in "${refs[@]}"; do
    if [ -f "$ROOT/$ref" ]; then
      note "$ref"
    else
      warn "docs/README.md references missing file: $ref"
    fi
  done
fi

echo "== Status-only content must live only in docs/status.md =="
while IFS= read -r file; do
  case "$file" in
    "$ROOT/docs/status.md") continue ;;
  esac
  # 'Known Issues' / 'Current Status' / 'Future Work' / 'Known Limitations'
  # sections (any casing), and inline test counts ("(74 tests)", "60+ tests",
  # "8 Stack tests"), must not appear outside docs/status.md.
  if grep -Eqi '^#+ +(Known Issues|Current Status|Future (Work|Extensions)|Known Limitations)' "$file"; then
    warn "$(basename "$file") has status/future section: $(grep -Ei '^#+ +(Known Issues|Current Status|Future (Work|Extensions)|Known Limitations)' "$file" | tr '\n' ' ')"
  fi
  if grep -Eq '\([0-9]+\+? tests?\)' "$file"; then
    warn "$(basename "$file") contains inline test counts"
  fi
  if grep -Eq '[0-9]+\+? (tests?|[A-Za-z][A-Za-z-]* tests?)\b' "$file"; then
    warn "$(basename "$file") contains bare test counts"
  fi
done < <(find "$ROOT/docs" "$ROOT/internal" -name '*.md' -not -path '*/testdata/*'; echo "$ROOT/README.md"; echo "$ROOT/AGENTS.md")

echo "== CONTEXT.md skeleton and header =="
while IFS= read -r file; do
  if ! grep -q '^> Per-package context: how this package works today. Status and plans live in' "$file"; then
    warn "$(basename "$file") missing pointer header (see docs/README.md taxonomy)"
  fi
  if grep -Eqi '^#+ +(Known Issues|Current Status|Current status|Future (Work|Extensions)|Known Limitations|TODOs|Suggestions)' "$file"; then
    warn "$(basename "$file") has status/plan section; fold into docs/status.md or docs/roadmap.md"
  fi
done < <(find "$ROOT/internal" -name 'CONTEXT.md')

if [ "$fail" -eq 0 ]; then
  echo "All checks passed."
fi
exit "$fail"
