#!/usr/bin/env bash
# Creates the GitHub Releases that are missing for tags made before CI released (docs/versioning.md), once, from a
# manual `workflow_dispatch` run with `backfill=true`. The body of each Release is that version's section of
# CHANGELOG.md. `gh` with GH_TOKEN (contents: write) and GITHUB_REPOSITORY, run from the repository root.
#
#   scripts/backfill-releases.sh [<tag vX.Y.Z>...]     (default: v0.2.0 v0.2.1 v0.3.0)
#
# A Release that exists is left alone, so a re-run changes nothing. A tag that does not exist or a version with no
# CHANGELOG section fails the run (nothing is invented). Tags are never made or moved here: --verify-tag. A backfilled
# Release is marked "Latest" only when its version is the highest of all the Releases and of this run, so a backfill never steals the
# badge from a newer one.
set -euo pipefail

repo="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is not set}"
changelog="${CHANGELOG:-CHANGELOG.md}"
semver='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'
[ $# -gt 0 ] || set -- v0.2.0 v0.2.1 v0.3.0
for t in "$@"; do [[ "$t" =~ $semver ]] || { echo "backfill: '$t' is not vMAJOR.MINOR.PATCH" >&2; exit 2; }; done

# section <X.Y.Z>: the CHANGELOG text under `## [X.Y.Z]` up to the next `## [`, outer blank lines trimmed.
section() {
    awk -v h="## [$1]" '
        index($0, h) == 1 { on = 1; next }
        on && /^## \[/ { exit }
        on { print }' "$changelog" | sed -e :a -e '/^\n*$/{$d;N;ba' -e '}' | sed '/./,$!d'
}

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
existing="$(gh release list --repo "$repo" --limit 1000 --json tagName --jq '.[].tagName')"

for tag in "$@"; do
    if grep -qxF "$tag" <<<"$existing"; then
        echo "::notice title=Backfill::$tag already has a Release, left alone"
        continue
    fi
    body="$(section "${tag#v}")"
    [ -n "$body" ] || { echo "backfill: no CHANGELOG section for ${tag#v}; no Release made for $tag" >&2; exit 1; }
    printf '%s\n' "$body" > "$tmp/notes"
    top="$(printf '%s\n' "$existing" "$@" | grep -E "$semver" | sort -V | tail -n 1)"
    latest=--latest=false
    [ "$top" != "$tag" ] || latest=--latest
    gh release create "$tag" --repo "$repo" --title "$tag" --notes-file "$tmp/notes" --verify-tag "$latest"
    echo "::notice title=Backfill::$tag released with its CHANGELOG section"
done
