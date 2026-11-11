#!/usr/bin/env bash
# Tags a release and creates its GitHub Release (docs/versioning.md): the annotated tag vX.Y.Z on the commit,
# through the GitHub API with the workflow's token, then the Release with notes generated from the PR titles since
# the previous tag. Ported from spare's. `gh` with GH_TOKEN (contents: write) and GITHUB_REPOSITORY. The release
# job runs it only after `check` passed, so a red main takes no version.
#
#   scripts/version-tag.sh <X.Y.Z> <full commit sha> [<previous tag vX.Y.Z>]
#
# No asset is attached to the Release (a library: the tag is the artifact, the Go proxy serves it). A tag made
# with the workflow's token starts no workflow. An existing tag is never moved: one on this commit counts as made
# (a retry that lost the answer, or a re-run), one on another commit fails. A Release that exists is left alone, so
# a re-run on the same commit is a success that changes nothing.
set -euo pipefail

usage="usage: version-tag.sh <X.Y.Z> <full commit sha> [<previous tag vX.Y.Z>]"
version="${1:?$usage}"
sha="${2:?$usage}"
prev="${3:-}"
[ $# -le 3 ] || { echo "version-tag: $usage" >&2; exit 2; }
[[ "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || { echo "version-tag: '$version' is not MAJOR.MINOR.PATCH" >&2; exit 2; }
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || { echo "version-tag: the commit must be a full sha, got '$sha'" >&2; exit 2; }
[ -z "$prev" ] || [[ "$prev" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || { echo "version-tag: the previous tag must be vMAJOR.MINOR.PATCH, got '$prev'" >&2; exit 2; }
repo="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is not set}"
tag="v$version"

# A transient GitHub API error (502, a secondary-rate-limit 403) must not leave main red after a passed check:
# every call is retried (TAG_ATTEMPTS, default 5, TAG_DELAY seconds apart, doubling), and every step is safe to
# repeat.
attempts="${TAG_ATTEMPTS:-5}"
delay="${TAG_DELAY:-5}"
retry() {
    local n=1 wait="$delay"
    until "$@"; do
        [ "$n" -lt "$attempts" ] || return 1
        echo "version-tag: attempt $n/$attempts failed, retrying in ${wait}s" >&2
        sleep "$wait"
        n=$((n + 1)); wait=$((wait * 2))
    done
}

# The commit the tag $tag points at, empty when there is no such tag (a 404 is not an error to retry).
tag_commit() {
    local out type obj
    out="$(gh api "repos/$repo/git/ref/tags/$tag" 2>&1)" || { [[ "$out" == *"404"* || "$out" == *"Not Found"* ]] && return 0; echo "$out" >&2; return 1; }
    type="$(jq -r .object.type <<<"$out")"; obj="$(jq -r .object.sha <<<"$out")"
    if [ "$type" = tag ]; then
        out="$(gh api "repos/$repo/git/tags/$obj")" || return 1
        jq -r .object.sha <<<"$out"
    else
        echo "$obj"
    fi
}

make_tag() {
    local existing object
    existing="$(tag_commit)" || return 1
    if [ -n "$existing" ]; then
        [ "$existing" = "$sha" ] && return 0
        echo "version-tag: $tag exists already on $existing, not on $sha; a tag is never moved" >&2
        exit 1
    fi
    object="$(jq -n --arg tag "$tag" --arg sha "$sha" --arg date "$(date -u +%Y-%m-%dT%H:%M:%SZ)" '{
            tag: $tag, object: $sha, type: "commit", message: "Release \($tag)",
            tagger: {name: "github-actions[bot]", email: "41898282+github-actions[bot]@users.noreply.github.com", date: $date}}' \
        | gh api -X POST "repos/$repo/git/tags" --input - | jq -r .sha)" || return 1
    [[ "$object" =~ ^[0-9a-f]{40}$ ]] || return 1
    jq -n --arg ref "refs/tags/$tag" --arg sha "$object" '{ref: $ref, sha: $sha}' \
        | gh api -X POST "repos/$repo/git/refs" --input - >/dev/null
}

# The Release: skipped when one exists for the tag. --verify-tag makes gh refuse to create the tag itself.
make_release() {
    if gh release view "$tag" --repo "$repo" >/dev/null 2>&1; then return 0; fi
    local args=(--repo "$repo" --title "$tag" --generate-notes --verify-tag)
    [ -z "$prev" ] || args+=(--notes-start-tag "$prev")
    gh release create "$tag" "${args[@]}"
}

retry make_tag || { echo "version-tag: GitHub refused to create $tag after $attempts attempts (the workflow token may not write the repository)" >&2; exit 1; }
retry make_release || { echo "version-tag: $tag is tagged but its GitHub Release could not be created after $attempts attempts; re-run the job" >&2; exit 1; }
echo "::notice title=Release::$tag tagged on $sha and released"
