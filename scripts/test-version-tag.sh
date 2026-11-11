#!/usr/bin/env bash
# Tests scripts/version-tag.sh with `gh` stubbed (no network; part of `just check` and the CI `check` job): the tag
# and the Release are made, a re-run on the same commit is a no-op success, a tag on another commit is refused and
# never moved, a transient API error is retried, bad arguments are refused, no asset is attached.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
script="$root/scripts/version-tag.sh"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
fails=0
fail() { echo "test-version-tag: FAIL: $*" >&2; fails=$((fails + 1)); }

mkdir "$tmp/bin"
cat > "$tmp/bin/gh" <<'STUB'
#!/usr/bin/env bash
# A tiny GitHub: refs, tag objects and releases as files under $STATE.
set -euo pipefail
echo "$*" >> "$STATE/calls"
if [ -f "$STATE/flaky" ]; then n="$(cat "$STATE/flaky")"; if [ "$n" -gt 0 ]; then echo $((n - 1)) > "$STATE/flaky"; echo "gh: HTTP 502" >&2; exit 1; fi; fi
case "$1 $2" in
    "api -X")
        path="$4"; body="$(cat)"
        case "$path" in
            */git/tags)
                obj="$(printf '%s' "$body" | jq -r '.tag + .object' | sha1sum | cut -c1-40)"
                printf '%s' "$body" | jq -r .object > "$STATE/tagobj-$obj"
                printf '{"sha":"%s"}\n' "$obj" ;;
            */git/refs)
                ref="$(jq -r .ref <<<"$body")"; name="${ref#refs/tags/}"
                [ ! -e "$STATE/ref-$name" ] || { echo "gh: Reference already exists (HTTP 422)" >&2; exit 1; }
                jq -r .sha <<<"$body" > "$STATE/ref-$name" ;;
        esac ;;
    "api repos"*)
        case "$2" in
            */git/ref/tags/*)
                name="${2##*/}"
                [ -e "$STATE/ref-$name" ] || { echo "gh: Not Found (HTTP 404)" >&2; exit 1; }
                printf '{"object":{"type":"tag","sha":"%s"}}\n' "$(cat "$STATE/ref-$name")" ;;
            */git/tags/*)
                printf '{"object":{"sha":"%s"}}\n' "$(cat "$STATE/tagobj-${2##*/}")" ;;
        esac ;;
    "release view") [ -e "$STATE/release-$3" ] ;;
    "release create")
        tag="$3"
        [ -e "$STATE/ref-$tag" ] || { echo "gh: tag $tag does not exist (--verify-tag)" >&2; exit 1; }
        : > "$STATE/release-$tag"; echo "$*" > "$STATE/release-args-$tag" ;;
    *) echo "stub gh: unexpected $*" >&2; exit 99 ;;
esac
STUB
chmod +x "$tmp/bin/gh"

a=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
b=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
fresh() { rm -rf "$tmp/state"; mkdir "$tmp/state"; }
run() { STATE="$tmp/state" PATH="$tmp/bin:$PATH" GITHUB_REPOSITORY=o/r TAG_DELAY=0 "$script" "$@"; }

# creates the tag and the Release, with notes since the previous tag and no asset
fresh
out="$(run 0.2.0 "$a" v0.1.0 2>&1)" || fail "first run failed: $out"
[ -e "$tmp/state/ref-v0.2.0" ] || fail "no tag ref created"
[ "$(cat "$tmp/state/tagobj-$(cat "$tmp/state/ref-v0.2.0")")" = "$a" ] || fail "the tag does not point at the commit"
args="$(cat "$tmp/state/release-args-v0.2.0" 2>/dev/null || true)"
for want in "release create v0.2.0" "--generate-notes" "--verify-tag" "--notes-start-tag v0.1.0"; do
    [[ "$args" == *"$want"* ]] || fail "release create lacks '$want': $args"
done
[[ "$args" != *.tar* && "$args" != *.zip* ]] || fail "an asset seems attached: $args"
[[ "$out" == *"::notice"*"v0.2.0"* ]] || fail "no notice: $out"

# the first release has no previous tag: no --notes-start-tag
fresh
run 0.1.0 "$a" >/dev/null 2>&1 || fail "first-ever release failed"
[[ "$(cat "$tmp/state/release-args-v0.1.0")" != *notes-start-tag* ]] || fail "--notes-start-tag with no previous tag"

# a re-run on the same commit succeeds and creates nothing more
fresh
run 0.2.0 "$a" v0.1.0 >/dev/null 2>&1
creates="$(grep -c 'release create\|git/refs\|git/tags$' "$tmp/state/calls")"
run 0.2.0 "$a" v0.1.0 >/dev/null 2>&1 || fail "re-run on the same commit failed"
[ "$(grep -c 'release create\|git/refs\|git/tags$' "$tmp/state/calls")" = "$creates" ] || fail "the re-run created something"

# a tag already on this commit but no Release (the first run died between them): the Release is made
fresh
run 0.2.0 "$a" >/dev/null 2>&1
rm "$tmp/state/release-v0.2.0"
run 0.2.0 "$a" >/dev/null 2>&1 || fail "completing a half-done release failed"
[ -e "$tmp/state/release-v0.2.0" ] || fail "the Release was not completed"

# a tag on another commit is refused and left where it is
fresh
run 0.2.0 "$a" >/dev/null 2>&1
before="$(cat "$tmp/state/ref-v0.2.0")"
if out="$(run 0.2.0 "$b" 2>&1)"; then fail "a tag on another commit was accepted"; fi
[[ "$out" == *"never moved"* ]] || fail "no 'never moved' message: $out"
[ "$(cat "$tmp/state/ref-v0.2.0")" = "$before" ] || fail "the tag moved"

# a transient error is retried
fresh; echo 2 > "$tmp/state/flaky"
run 0.2.0 "$a" >/dev/null 2>&1 || fail "a transient error was not retried"
[ -e "$tmp/state/release-v0.2.0" ] || fail "no Release after the retry"

# a permanent error fails after TAG_ATTEMPTS
fresh; echo 99 > "$tmp/state/flaky"
if TAG_ATTEMPTS=2 run 0.2.0 "$a" >/dev/null 2>&1; then fail "a permanent error succeeded"; fi

# bad arguments
fresh
for bad in "0.2 $a" "v0.2.0 $a" "0.2.0 abc" "0.2.0 $a 0.1.0" ""; do
    # shellcheck disable=SC2086 # word splitting is the point
    if run $bad >/dev/null 2>&1; then fail "arguments '$bad' accepted"; fi
done
[ ! -s "$tmp/state/calls" ] || fail "bad arguments reached GitHub"

[ "$fails" = 0 ] || exit 1
echo "version-tag: ok (tag and Release made, re-run no-op, tag on another commit refused, transient errors retried)"
