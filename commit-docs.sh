#!/usr/bin/env zsh
# commit-docs.sh — commit the remaining groups not yet committed.
# Run without args to preview; run with --execute to proceed.

set -euo pipefail

if [[ "$#" -eq 0 ]]; then
    DRY_RUN=true
elif [[ "$1" == "--execute" ]]; then
    DRY_RUN=false
else
    echo "Usage: $0 [--execute]"
    exit 1
fi

if $DRY_RUN; then
    echo "=== DRY RUN — pass --execute to actually commit ==="
    echo
fi

commit() {
    local msg="$1"
    shift
    if $DRY_RUN; then
        echo "Would commit: $msg"
        echo "  files: $*"
        echo
    else
        echo "--- $msg ---"
        git add "$@"
        git commit -m "$msg"
        echo
    fi
}

# Remaining groups (not yet committed by commit.sh):

# 7. CLI wiring — already staged
if ! $DRY_RUN; then
    git commit -m "feat(cli): wire SQLite store and state into startup"
    echo
fi

# 8. Legacy exercise conversion script
commit "feat(scripts): add legacy exercise to YAML converter" \
    scripts/

# 9. Documentation updates (including AGENTS.md and this script)
commit "docs: update project and subsystem documentation" \
    AGENTS.md \
    docs/ \
    internal/ui/docs/ \
    internal/ui/screens/CONTEXT.md \
    commit-docs.sh

if $DRY_RUN; then
    echo "=== Dry run complete. Run '$0 --execute' to commit. ==="
fi