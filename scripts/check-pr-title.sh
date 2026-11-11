#!/usr/bin/env bash
# Fail unless the pull request title (argument 1) is a conventional-commit line:
#   <type>(<scope>)?!?: <description>
# The squash commit on main takes the PR title, and the release reads the version bump from it
# (docs/versioning.md).
set -euo pipefail
title="${1-}"
types='feat|fix|refactor|perf|test|docs|chore|ci|build'
if [[ "$title" =~ ^($types)(\([^\)[:space:]]+\))?\!?:\ [^[:space:]] ]]; then
    exit 0
fi
cat >&2 <<MSG
The pull request title is not a conventional-commit line:
  $title

Expected:  <type>(<scope>)?!?: <description>
  types:   ${types//|/ }
  example: feat(model): add WithFoo option
  example: fix!: reject a zero width
The title becomes the squash commit on main and drives the release version (feat: minor, fix and
the others: patch, docs/test/chore/ci: no release, "!": minor with a notice, toast stays 0.x).
MSG
exit 1
