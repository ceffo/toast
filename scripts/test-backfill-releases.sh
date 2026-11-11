#!/usr/bin/env bash
# Tests scripts/backfill-releases.sh with `gh` stubbed (part of `just test-scripts` and the CI `check` job): the missing
# Releases are made with their CHANGELOG section as body, an existing one is left alone, a re-run changes nothing,
# a version without a section or a bad tag fails, "Latest" only goes to the highest version.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
script="$root/scripts/backfill-releases.sh"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
fails=0
fail() { echo "test-backfill-releases: FAIL: $*" >&2; fails=$((fails + 1)); }

mkdir "$tmp/bin"
cat > "$tmp/bin/gh" <<'STUB'
#!/usr/bin/env bash
set -euo pipefail
echo "$*" >> "$STATE/calls"
case "$1 $2" in
    "release list") ls "$STATE" | sed -n 's/^release-//p' ;;
    "release create")
        tag="$3"
        [ -e "$STATE/tag-$tag" ] || { echo "gh: tag $tag does not exist (--verify-tag)" >&2; exit 1; }
        : > "$STATE/release-$tag"; echo "$*" > "$STATE/args-$tag"
        while [ $# -gt 0 ]; do [ "$1" != --notes-file ] || cp "$2" "$STATE/body-$tag"; shift; done ;;
    *) echo "stub gh: unexpected $*" >&2; exit 99 ;;
esac
STUB
chmod +x "$tmp/bin/gh"

cat > "$tmp/CHANGELOG.md" <<'MD'
# Changelog

Pointer text.

## [0.3.0] - 2026-11-11

### Added

- third

## [0.2.1] - 2026-06-09

### Fixed

- second, with `code`

## [0.2.0] - 2026-06-09

### Changed

- first
- more

## [0.1.0] - 2026-06-05

Initial release.
MD

fresh() { rm -rf "$tmp/state"; mkdir "$tmp/state"; for t in v0.1.0 v0.2.0 v0.2.1 v0.3.0; do : > "$tmp/state/tag-$t"; done; : > "$tmp/state/release-v0.1.0"; }
run() { STATE="$tmp/state" PATH="$tmp/bin:$PATH" GITHUB_REPOSITORY=o/r CHANGELOG="$tmp/CHANGELOG.md" "$script" "$@"; }

fresh
out="$(run 2>&1)" || fail "backfill failed: $out"
for t in v0.2.0 v0.2.1 v0.3.0; do [ -e "$tmp/state/release-$t" ] || fail "no Release $t"; done
[ "$(cat "$tmp/state/body-v0.3.0")" = "$(printf '### Added\n\n- third')" ] || fail "v0.3.0 body: $(cat "$tmp/state/body-v0.3.0")"
[ "$(cat "$tmp/state/body-v0.2.1")" = "$(printf '### Fixed\n\n- second, with `code`')" ] || fail "v0.2.1 body"
[ "$(cat "$tmp/state/body-v0.2.0")" = "$(printf '### Changed\n\n- first\n- more')" ] || fail "v0.2.0 body (must stop at the next section)"
for t in v0.2.0 v0.2.1; do [[ "$(cat "$tmp/state/args-$t")" == *"--latest=false"* ]] || fail "$t may take Latest"; done
a="$(cat "$tmp/state/args-v0.3.0")"
[[ "$a" == *"--verify-tag"* && "$a" == *"--latest"* && "$a" != *"--latest=false"* && "$a" == *"--title v0.3.0"* ]] || fail "v0.3.0 args: $a"

# a re-run changes nothing
creates="$(grep -c 'release create' "$tmp/state/calls")"
out="$(run 2>&1)" || fail "re-run failed: $out"
[ "$(grep -c 'release create' "$tmp/state/calls")" = "$creates" ] || fail "the re-run created a Release"

# a Release that exists is left alone (its body is not touched)
fresh; : > "$tmp/state/release-v0.2.1"
run >/dev/null 2>&1 || fail "run with one existing Release failed"
[ ! -e "$tmp/state/body-v0.2.1" ] || fail "an existing Release was recreated"
[ -e "$tmp/state/release-v0.3.0" ] || fail "v0.3.0 not made next to an existing one"

# a newer Release exists: a backfilled one never takes Latest
fresh; : > "$tmp/state/release-v0.4.0"
run v0.3.0 >/dev/null 2>&1 || fail "backfill beside a newer Release failed"
[[ "$(cat "$tmp/state/args-v0.3.0")" == *"--latest=false"* ]] || fail "a backfill took Latest from v0.4.0"

# a tag that does not exist, a version with no section, a bad tag
fresh; rm "$tmp/state/tag-v0.2.0"
if run v0.2.0 >/dev/null 2>&1; then fail "a missing tag was accepted"; fi
fresh
out="$(run v0.9.9 2>&1)" && fail "a version without a CHANGELOG section was accepted"
[[ "$out" == *"no CHANGELOG section for 0.9.9"* ]] || fail "no message for the missing section: $out"
for bad in 0.2.0 v0.2 'v0.2.0; x'; do
    if run "$bad" >/dev/null 2>&1; then fail "bad tag '$bad' accepted"; fi
done

[ "$fails" = 0 ] || exit 1
echo "backfill-releases: ok"
