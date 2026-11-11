# Versioning and releases

CI releases toast on every merge to `main` that deserves one. Nobody runs a release step, picks a version or edits a
changelog. Decided 2026-11-11: version from conventional-commit PR titles, no release PR, no binary asset (a library:
Go consumers fetch the tag through the module proxy), toast stays 0.x.

Rejected: a release PR (release-please: a PR to merge on every release); a hand-run `scripts/release.sh` (removed);
a maintained `CHANGELOG.md` (frozen at v0.3.0, the notes are generated from PR titles).

## The PR title

The squash commit on `main` takes the PR title (repository setting: title = PR title, empty message), so the title is
`<type>(<scope>)?!?: <description>` with a type of `feat`, `fix`, `refactor`, `perf`, `test`, `docs`, `chore`, `ci` or
`build`. `.github/workflows/pr-title.yml` fails a PR whose title is not of that form (`scripts/check-pr-title.sh`). It is
not a required check.

## The rule

The range is the first-parent commits after the highest `vX.Y.Z` tag reachable from the merged commit.

| Commit | Level |
|---|---|
| `feat` | minor |
| `docs`, `test`, `chore`, `ci` | none |
| `fix`, `refactor`, `perf`, `build`, any other or unparseable title | patch |
| `!` after the type, or a `BREAKING CHANGE:` footer or text | minor, with a notice that a major may be due |

The highest level of the range wins. There is a release when at least one commit has a level and its PR does not carry
the label **`no-release`**; otherwise none, and the job stays green with a notice saying why. A PR labelled `no-release`
makes no release by itself, but its title still counts toward the level of the next release that happens.

While toast is 0.x a breaking change is a minor. A major (`1.0.0`, and any `v2` module path) is made by hand and is not
built here.

## The release job

The `release` job of `ci.yml` (push to `main`, after `check`) runs `scripts/release-version.sh next "$GITHUB_SHA"` on a
full-history checkout, then, with a release, `scripts/version-tag.sh X.Y.Z <sha> <previous tag>`: the annotated tag
`vX.Y.Z` on the commit and a GitHub Release with generated notes. `contents: write` is on this job only. Tags made with
the job's token start no workflow.

- Two quick merges never take the same number: the `release` concurrency group is never cancelled, and a run whose
  commit is already inside a newer release ends green (`superseded=true`) and moves no tag.
- A tag is never moved. Re-running the job on a released commit succeeds and changes nothing.
- A red `check` releases nothing; the next merge computes the version again.

## Backfill (one-off)

Tags `v0.2.0`, `v0.2.1`, `v0.3.0` predate CI releases. A manual run, `gh workflow run ci.yml -f backfill=true`, creates
their Releases with the matching `CHANGELOG.md` section as body (`scripts/backfill-releases.sh`). An existing Release is
left alone, so a re-run changes nothing.

## Checking

- `gh release list` shows every tag; `gh run list --workflow ci.yml` shows the `release` job (its notice names the
  version or the reason for none).
- Before merging, `scripts/release-version.sh bump <last tag|none> < commits` is the pure rule; its input format is in
  the script header.
- Tests: `just test-scripts` (part of `just check` and the CI `check` job).
