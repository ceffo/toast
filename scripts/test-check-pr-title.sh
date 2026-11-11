#!/usr/bin/env bash
# Test of check-pr-title.sh: valid titles pass, the others are rejected.
set -euo pipefail
check="$(dirname "$0")/check-pr-title.sh"
status=0

expect() { # expect <accept|reject> <title>
    if "$check" "$2" >/dev/null 2>&1; then got=accept; else got=reject; fi
    if [ "$got" != "$1" ]; then echo "FAIL: '$2' should be ${1}ed"; status=1; fi
}

expect accept 'fix(fs): count freed bytes per inode (#9)'
expect accept 'fix(deletion): erase a torrent again when a new read lists it (#102)'
expect accept 'feat!: x'
expect accept 'feat: add a thing'
expect accept 'refactor(tui)!: rename the model'
for t in feat fix test docs chore refactor perf ci build; do expect accept "$t: something"; done
expect reject 'Update stuff'
expect reject 'fix:'
expect reject 'fix: '
expect reject 'feat:no space'
expect reject 'feat(): empty scope'
expect reject 'feat(a b): spaced scope'
expect reject 'wip: unknown type'
expect reject 'Feat: capitalised type'
expect reject ' feat: leading space'
expect reject 'chore/release-version'
expect reject ''
[ "$status" = 0 ] && echo "check-pr-title: ok"
exit $status
