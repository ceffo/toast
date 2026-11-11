#!/usr/bin/env bash
# Tests scripts/release-version.sh without the network or GitHub (part of `just check` and the CI `check` job):
#   - bump, table-driven: each commit type's level, a breaking change (minor plus a notice), docs/test/chore/ci only
#     and the no-release label (no release, with the reason), a mixed range takes the highest, --force, the first
#     release (no tag: 0.1.0), titles with a trailing (#n);
#   - next on a throwaway git repository with gh stubbed: the range after the last v* tag reachable from the
#     commit (first parent), each commit's PR labels, a BREAKING CHANGE footer, a newer tag that is not reachable
#     from the commit: ends quietly (superseded=true, no version) when the commit is inside it, refused when
#     history diverged, no tag at all (first release);
#   - resolve: a version tag's commit, an unknown version refused with its name.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
script="$root/scripts/release-version.sh"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
fails=0
fail() { echo "test-release-version: FAIL: $*" >&2; fails=$((fails + 1)); }

# c <title> [labels] [breaking footer 0|1] [pr]: one input line of bump.
n=0
c() { n=$((n + 1)); printf '%040d\t%s\t%s\t%s\t%s\n' "$n" "${4:-$n}" "${2:-}" "${3:-0}" "$1"; }

# bump_case <name> <last|none> <want release> <want version or -> <want bump> [notice substring] [--force] < commits
bump_case() {
    local name="$1" last="$2" want_release="$3" want_version="$4" want_bump="$5" want_notice="${6:-}" force="${7:-}"
    local out status=0 want_last="$2"
    [ "$last" != none ] || want_last=""
    : > "$tmp/out"
    # shellcheck disable=SC2086 # force is empty or one word
    out="$(GITHUB_OUTPUT="$tmp/out" "$script" bump "$last" $force 2>&1)" || status=$?
    [ "$status" = 0 ] || { fail "$name: exit $status: $out"; return; }
    grep -qx "release=$want_release" "$tmp/out" || fail "$name: release is not $want_release: $(cat "$tmp/out")"
    grep -qx "bump=$want_bump" "$tmp/out" || fail "$name: bump is not $want_bump: $(cat "$tmp/out")"
    grep -qx "last=$want_last" "$tmp/out" || fail "$name: last is not '$want_last': $(cat "$tmp/out")"
    if [ "$want_version" = - ]; then
        ! grep -q '^version=' "$tmp/out" || fail "$name: a version without a release: $(cat "$tmp/out")"
    else
        grep -qx "version=$want_version" "$tmp/out" || fail "$name: version is not $want_version: $(cat "$tmp/out")"
    fi
    [ -z "$want_notice" ] || [[ "$out" == *"$want_notice"* ]] || fail "$name: no notice '$want_notice': $out"
    ! grep -q '^::' "$tmp/out" || fail "$name: a notice in the outputs"
}

# --- bump: one commit of each kind ---
c 'feat: a panel' | bump_case "feat" v0.1.0 true 0.2.0 minor
c 'feat(tui): help overlay' | bump_case "feat with a scope" v0.1.0 true 0.2.0 minor
for t in 'fix: a bug' 'refactor(deletion): x' 'perf: faster' 'build: deps' 'revert: feat: x'; do
    c "$t" | bump_case "'$t'" v0.1.0 true 0.1.1 patch
done
for t in 'Update stuff' 'feat:no space' 'feat(: open' 'Merge pull request #3 from a/b' ''; do
    c "$t" | bump_case "unparseable '$t'" v0.1.0 true 0.1.1 patch
done
for t in 'docs: a page' 'test: more' 'chore: tidy' 'chore(deps): bump x' 'docs(adr): y' 'ci: a job' 'ci(release): x'; do
    c "$t" | bump_case "'$t' alone" v0.1.0 false - none "docs/test/chore/ci only since v0.1.0"
done

# --- bump: titles, the (#n) suffix changes nothing ---
c 'feat(candidate): rank by interest (#103)' | bump_case "feat with suffix" v0.1.0 true 0.2.0 minor
c 'fix(deletion): erase a torrent again when a new read lists it (#102)' | bump_case "fix with suffix" v0.1.0 true 0.1.1 patch
c 'docs(deletion): bring the guardrails up to date (#101)' | bump_case "docs with suffix" v0.1.0 false - none "docs/test/chore/ci only"
c 'refactor(pipeline): inputs in one place (#98)' | bump_case "refactor with suffix" v0.1.0 true 0.1.1 patch
c 'chore(workflow): developers run the tests (#100)' | bump_case "chore with suffix" v0.1.0 false - none
c 'fix(cmd)!: drop a flag (#5)' | bump_case "breaking with suffix" v0.1.0 true 0.2.0 minor "Major version?"

# --- bump: a breaking change is a minor, with a notice that a major may be due ---
c 'fix!: drop a flag' | bump_case "fix!" v0.4.2 true 0.5.0 minor "Major version?"
c 'feat(cmd)!: new shape' | bump_case "feat(scope)!" v0.4.2 true 0.5.0 minor "Major version?"
c 'fix: x' '' 1 | bump_case "BREAKING CHANGE footer" v0.4.2 true 0.5.0 minor "Major version?"
c 'refactor: BREAKING CHANGE in the title' | bump_case "BREAKING CHANGE title" v0.4.2 true 0.5.0 minor "Major version?"
c 'docs!: breaking docs' | bump_case "docs! is a release" v0.4.2 true 0.5.0 minor "Major version?"
c 'fix!: x' | bump_case "breaking on 1.x stays a minor" v1.4.2 true 1.5.0 minor "Major version?"
out="$(c 'feat: x' | GITHUB_OUTPUT=/dev/null "$script" bump v0.4.2)"
[[ "$out" != *"Major version?"* ]] || fail "a plain feat gets the major notice"

# --- bump: ranges ---
{ c 'docs: a'; c 'fix: b'; c 'feat: c'; c 'test: d'; } | bump_case "mixed range takes the highest" v0.4.2 true 0.5.0 minor
{ c 'feat: c'; c 'fix: b'; c 'docs: a'; } | bump_case "a feat earlier in the range still counts" v0.4.2 true 0.5.0 minor
{ c 'docs: a'; c 'fix: b'; c 'chore: c'; } | bump_case "fix among docs" v0.4.2 true 0.4.3 patch
{ c 'feat: no release' 'no-release' 0 41; c 'docs: x'; c 'fix: y'; } \
    | bump_case "a labelled feat still sets the level of a releasing range" v0.4.2 true 0.5.0 minor
{ c 'docs: a'; c 'chore: b'; c 'test: c'; } | bump_case "docs/test/chore/ci only" v0.4.2 false - none "docs/test/chore/ci only"

# --- bump: the no-release label ---
c 'feat: x' 'no-release' 0 77 | bump_case "no-release label" v0.4.2 false - none "no-release label on PR #77"
c 'fix: x' 'ui,no-release,deps' 0 78 | bump_case "no-release among labels" v0.4.2 false - none "no-release label on PR #78"
c 'fix: x' 'no-release-later' 0 79 | bump_case "a label that only starts with no-release" v0.4.2 true 0.4.3 patch
{ c 'docs: a'; c 'feat: b' 'no-release' 0 80; c 'fix: c' 'no-release' 0 81; } \
    | bump_case "every releasing commit labelled" v0.4.2 false - none "no-release label on PR #80 #81"

# --- bump: nothing since the last version; --force ---
bump_case "empty range" v0.4.2 false - none "no commit since v0.4.2" </dev/null
bump_case "forced, empty range" v0.4.2 true 0.4.3 patch "Forced release" --force </dev/null
c 'docs: a' | bump_case "forced, docs only" v0.4.2 true 0.4.3 patch "Forced release" --force
c 'feat: a' 'no-release' | bump_case "forced, labelled feat" v0.4.2 true 0.5.0 minor "Forced release" --force
c 'feat: a' | bump_case "forced when the rules release anyway" v0.4.2 true 0.5.0 minor "" --force

# --- bump: no tag yet: the first release is 0.1.0, whatever the level; the rules still decide whether it happens ---
c 'feat: a' | bump_case "no tag, feat" none true 0.1.0 minor "after the start"
c 'fix: a' | bump_case "no tag, fix" none true 0.1.0 minor
c 'fix!: a' | bump_case "no tag, breaking" none true 0.1.0 minor "Major version?"
{ c 'docs: a'; c 'fix: b'; } | bump_case "no tag, docs then fix" none true 0.1.0 minor
c 'docs: a' | bump_case "no tag, docs only: no release" none false - none "docs/test/chore/ci only since the start"
c 'feat: a' 'no-release' 0 9 | bump_case "no tag, no-release label" none false - none "no-release label on PR #9"
bump_case "no tag, empty range" none false - none "no commit since the start" </dev/null
c 'docs: a' | bump_case "no tag, forced docs only" none true 0.1.0 minor "Forced release" --force

# --- bump: version arithmetic ---
c 'fix: a' | bump_case "patch 9 to 10" v0.1.9 true 0.1.10 patch
c 'feat: a' | bump_case "minor 9 to 10, patch reset" v0.9.9 true 0.10.0 minor
c 'feat: a' | bump_case "from 0.1.0" v0.1.0 true 0.2.0 minor

# --- bump: refused input ---
for last in 1.2.3 v1.2 v01.2.3 v1.2.3-rc1 ''; do
    : | "$script" bump "$last" >/dev/null 2>&1 && fail "bump accepted the last version '$last'"
done
: | "$script" bump v1.2.3 --other >/dev/null 2>&1 && fail "bump accepted an unknown flag"
: | "$script" bump v1.2.3 --major >/dev/null 2>&1 && fail "bump accepted --major (not built)"

# --- next: a throwaway repository, gh stubbed ---
stubs="$tmp/bin"
mkdir -p "$stubs" "$tmp/prs"
# gh api repos/o/r/commits/<sha>/pulls --jq <expr>: answers $tmp/prs/<sha> (else []), through the real jq.
cat > "$stubs/gh" <<STUB
#!/usr/bin/env bash
echo "gh \$*" >> "$tmp/gh-calls"
[ "\$1" = api ] || exit 1
[ ! -f "$tmp/gh-fails" ] || exit 1
sha="\$(sed -n 's|^repos/o/r/commits/\([0-9a-f]*\)/pulls\$|\1|p' <<<"\$2")"
[ -n "\$sha" ] || exit 1
f="$tmp/prs/\$sha"; [ -f "\$f" ] || f="$tmp/prs/none"
jq -r "\$4" "\$f"
STUB
echo '[]' > "$tmp/prs/none"
chmod +x "$stubs/gh"

repo="$tmp/repo"
git init -q -b main "$repo"
cd "$repo"
git config user.email t@example.com
git config user.name test
git config commit.gpgsign false
git config tag.gpgsign false
commit() { git commit -q --allow-empty -m "$1" ${2:+-m "$2"}; git rev-parse HEAD; }
next() { : > "$tmp/out"; : > "$tmp/gh-calls"; status=0
    out="$(PATH="$stubs:$PATH" GITHUB_REPOSITORY=o/r GITHUB_OUTPUT="$tmp/out" "$script" next "$@" 2>&1)" || status=$?; }

# No tag yet: the first release is 0.1.0 and counts every commit.
a1="$(commit 'docs: the first page (#1)')"
a2="$(commit 'feat(cmd): the first thing (#2)')"
next "$a1"
[ "$status" = 0 ] && [ "$(cat "$tmp/out")" = "$(printf 'release=false\nbump=none\nlast=')" ] && [[ "$out" == *"docs/test/chore/ci only since the start"* ]] \
    || fail "next, no tag, docs only: $status $out / $(cat "$tmp/out")"
next "$a2"
[ "$status" = 0 ] && [ "$(cat "$tmp/out")" = "$(printf 'release=true\nbump=minor\nlast=\nversion=0.1.0')" ] || fail "next, no tag: $status $out / $(cat "$tmp/out")"
[ "$(grep -c 'gh api repos/o/r/commits/' "$tmp/gh-calls")" = 2 ] || fail "next, no tag: not one PR lookup per commit: $(cat "$tmp/gh-calls")"
next "$a1" --force
grep -qx 'version=0.1.0' "$tmp/out" || fail "next, no tag, --force: $(cat "$tmp/out")"

git tag -a v0.1.0 -m r
git tag -a v0.0.9 -m r HEAD       # an older version on the same commit
git tag nonsemver-v9 HEAD
c1="$(commit 'docs: a page (#10)')"
c2="$(commit 'fix: a bug (#12)')"
c3="$(commit 'feat: a panel (#13)' 'BREAKING CHANGE: the config shape')"
c4="$(commit 'chore: tidy (#14)')"
# c2 is the merge commit of PR #12 (labelled no-release); another merged PR also lists it.
printf '[{"number": 9, "merged_at": "x", "merge_commit_sha": "other", "labels": [{"name": "ui"}]},
         {"number": 12, "merged_at": "x", "merge_commit_sha": "%s", "labels": [{"name": "no-release"}, {"name": "ui"}]},
         {"number": 5, "merged_at": null, "merge_commit_sha": "%s", "labels": []}]' "$c2" "$c2" > "$tmp/prs/$c2"
printf '[{"number": 13, "merged_at": "x", "merge_commit_sha": "%s", "labels": []}]' "$c3" > "$tmp/prs/$c3"

next "$c4"
[ "$status" = 0 ] || fail "next: exit $status: $out"
[ "$(cat "$tmp/out")" = "$(printf 'release=true\nbump=minor\nlast=v0.1.0\nversion=0.2.0')" ] || fail "next c4: outputs $(cat "$tmp/out")"
[[ "$out" == *"Major version?"* && "$out" == *"${c3:0:7}"* ]] || fail "next: the BREAKING CHANGE footer gave no notice: $out"
[ "$(grep -c 'gh api repos/o/r/commits/' "$tmp/gh-calls")" = 4 ] || fail "next: not one PR lookup per commit since v0.1.0: $(cat "$tmp/gh-calls")"

next "$c2"
[ "$(cat "$tmp/out")" = "$(printf 'release=false\nbump=none\nlast=v0.1.0')" ] || fail "next c2 (docs + labelled fix): $(cat "$tmp/out")"
[[ "$out" == *"no-release label on PR #12"* ]] || fail "next c2: the labelled PR is not named: $out"

next "$c1"
[[ "$(cat "$tmp/out")" == "release=false"* ]] && [[ "$out" == *"docs/test/chore/ci only since v0.1.0"* ]] || fail "next c1: $out"
next "$c1" --force
grep -qx 'version=0.1.1' "$tmp/out" || fail "next c1 --force: $(cat "$tmp/out")"

# A new version tag moves the range; a merged side branch counts by its merge commit only (first parent).
git tag -a v0.2.0 -m r "$c4"
git switch -q -c side "$c4"
commit 'feat: on the side' >/dev/null
git switch -q main
git merge -q --no-ff side -m 'fix: merge the side branch (#15)'
m="$(git rev-parse HEAD)"
next "$m"
[ "$(cat "$tmp/out")" = "$(printf 'release=true\nbump=patch\nlast=v0.2.0\nversion=0.2.1')" ] || fail "next after v0.2.0 (first parent): $(cat "$tmp/out")"
next "$c4"
[[ "$out" == *"no commit since v0.2.0"* ]] || fail "next on a tagged commit: $out"

# A newer tag on a commit that is not before this one: refused, nothing output.
git switch -q -c elsewhere "$c4"
commit 'fix: never merged' >/dev/null
git tag -a v0.2.1 -m r
git switch -q main
next "$m"
[ "$status" != 0 ] && [[ "$out" == *"a newer release exists: v0.2.1"* ]] && [ ! -s "$tmp/out" ] || fail "next: a newer tag elsewhere is accepted: $out"
git tag -d v0.2.1 >/dev/null

# An older run after a newer release: m (a fix, 0.2.1) lost the race, n (a feat) was tagged v0.3.0 and contains m.
# Running m again ends quietly (exit 0, release=false, superseded=true, a notice saying why), --force too: it never
# outputs a version, so nothing is tagged, released, signed or deployed behind v0.3.0.
n="$(commit 'feat: newer (#16)')"
git tag -a v0.3.0 -m r "$n"
for flag in "" --force; do
    next "$m" $flag
    [ "$status" = 0 ] && [[ "$out" == *"already contained in the newer release v0.3.0"* ]] \
        || fail "next $flag: an older run is not quiet: $status $out"
    [ "$(cat "$tmp/out")" = "$(printf 'release=false\nbump=none\nlast=v0.3.0\nsuperseded=true')" ] \
        || fail "next $flag: a superseded run outputs: $(cat "$tmp/out")"
    grep -q '^version=' "$tmp/out" && fail "next $flag: a superseded run outputs a version"
done
next "$n"
[ "$status" = 0 ] && [[ "$out" == *"no commit since v0.3.0"* ]] || fail "next on the newest tagged commit: $out"

touch "$tmp/gh-fails"
next "$n"
[ "$status" = 0 ] || fail "next with nothing to look up needs no GitHub: $out"
e="$(commit 'fix: a bug (#18)')"
next "$e"
[ "$status" != 0 ] && [[ "$out" == *"could not read the pull requests"* ]] && [ ! -s "$tmp/out" ] || fail "next: a GitHub failure is not a failure: $out"
rm -f "$tmp/gh-fails"
next deadbeef
[ "$status" != 0 ] && [[ "$out" == *"unknown commit"* ]] || fail "next: an unknown commit: $out"
next "$c1" --other
[ "$status" != 0 ] || fail "next: an unknown flag is accepted"

# --- resolve ---
[ "$("$script" resolve 0.2.0)" = "$c4" ] || fail "resolve 0.2.0 is not its commit"
[ "$("$script" resolve v0.1.0)" = "$(git rev-parse 'v0.1.0^{commit}')" ] || fail "resolve v0.1.0"
status=0; out="$("$script" resolve 9.9.9 2>&1)" || status=$?
[ "$status" != 0 ] && [[ "$out" == *"unknown version v9.9.9"* ]] || fail "resolve of an unknown version: $out"
for v in 1.2 abc 'v1.2.3; rm -rf /' ''; do
    "$script" resolve "$v" >/dev/null 2>&1 && fail "resolve accepted '$v'"
done
cd "$root"

[ "$fails" = 0 ] || exit 1
echo "release-version: ok"
