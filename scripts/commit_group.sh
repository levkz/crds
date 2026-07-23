#!/usr/bin/env zsh
# scripts/commit_group.sh — shared utility for commit-group scripts.
#
# Source this from your commit script, then call commit() for each group:
#
#   #!/usr/bin/env zsh
#   set -euo pipefail
#   source "$(dirname "$0")/commit_group.sh"
#
#   commit "feat(x): add foo" foo.go foo_test.go
#   commit "feat(y): add bar" bar.go

if [[ "${COMMIT_GROUP_LOADED:-}" == "1" ]]; then
    return 0
fi
COMMIT_GROUP_LOADED=1

SCRIPT_NAME="${SCRIPT_NAME:-$(basename "$0")}"

if [[ "$#" -gt 0 ]]; then
    echo "Usage: source $SCRIPT_NAME in your commit script, then call commit()"
    echo
    echo "  source \"\$(dirname \"\$0\")/commit_group.sh\""
    echo "  commit \"msg\" file1.go file2.go"
    exit 1
fi

export DRY_RUN
if [[ "${COMMIT_EXECUTE:-}" != "1" ]]; then
    DRY_RUN=true
else
    DRY_RUN=false
fi

if $DRY_RUN; then
    echo "=== DRY RUN — run with COMMIT_EXECUTE=1 to actually commit ==="
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

if $DRY_RUN; then
    cleanup() {
        echo "=== Dry run complete. Re-run with COMMIT_EXECUTE=1 to commit. ==="
    }
    trap cleanup EXIT
fi